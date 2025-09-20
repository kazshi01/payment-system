package httpi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/kazshi01/payment-system/internal/auth"
)

type AuthHandler struct {
	OIDC *auth.OIDC
}

func NewAuthHandler(ctx context.Context) (*AuthHandler, error) {
	oidcClient, err := auth.NewOIDC(ctx, auth.OIDCConfig{
		Issuer:      os.Getenv("OIDC_ISSUER"),
		ClientID:    os.Getenv("OIDC_CLIENT_ID"),
		RedirectURL: os.Getenv("OIDC_REDIRECT_URL"),
		Scopes:      []string{"profile", "email"},
	})
	if err != nil {
		return nil, err
	}
	return &AuthHandler{OIDC: oidcClient}, nil
}

const (
	cookieStateKey    = "oidc_state"
	cookieVerifierKey = "oidc_verifier"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, _ := auth.RandBase64URL(24)
	verifier, _ := auth.RandBase64URL(32)
	challenge := auth.CodeChallenge(verifier)

	// Cookie に state / verifier を短時間だけ保存（Todo: 本番では不要）
	setTmpCookie(w, cookieStateKey, state, 10*time.Minute)
	setTmpCookie(w, cookieVerifierKey, verifier, 10*time.Minute)

	// auth URL に PKCE パラメータを付与
	authURL := h.OIDC.Config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		http.Error(w, "missing code/state", http.StatusBadRequest)
		return
	}

	// CSRF 対策：state 照合
	cState, _ := r.Cookie(cookieStateKey)
	if cState == nil || cState.Value != state {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	// PKCE: code_verifier を添えてトークン交換
	cVerifier, _ := r.Cookie(cookieVerifierKey)
	if cVerifier == nil || cVerifier.Value == "" {
		http.Error(w, "missing verifier", http.StatusBadRequest)
		return
	}

	tok, err := h.OIDC.Config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", cVerifier.Value),
	)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	rt, _ := tok.Extra("refresh_token").(string)
	if rt != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "rt",
			Value:    rt,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   60 * 60, // 1h
		})

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// アクセストークンを検証 & クレームを読む（任意）
	idTokenRaw, _ := tok.Extra("id_token").(string)
	verifier := h.OIDC.Provider.Verifier(&oidc.Config{
		ClientID: h.OIDC.Config.ClientID,
	})
	var claims map[string]any
	if idTokenRaw != "" {
		if idt, err := verifier.Verify(ctx, idTokenRaw); err == nil {
			_ = idt.Claims(&claims)
		}
	}

	resp := map[string]any{
		"access_token": tok.AccessToken,
		"expires_in":   int64(time.Until(tok.Expiry).Seconds()),
		"token_type":   tok.TokenType,
		"id_token":     idTokenRaw,
		"claims":       claims,
	}
	WriteJSON(w, http.StatusCreated, resp)
}

func setTmpCookie(w http.ResponseWriter, name, val string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

// POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	c, err := r.Cookie("rt")
	if err != nil || c.Value == "" {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	ts := h.OIDC.Config.TokenSource(r.Context(), &oauth2.Token{
		RefreshToken: c.Value,
	})
	tok, err := ts.Token()
	if err != nil {
		http.Error(w, "refresh failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	newRT := tok.RefreshToken
	if newRT == "" {
		newRT = c.Value
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "rt",
		Value:    newRT,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // 本番は有効
		MaxAge: 60 * 60,
	})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": tok.AccessToken,
		"expires_in":   int64(time.Until(tok.Expiry).Seconds()),
		"token_type":   tok.TokenType,
	})
}

// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ログアウト時には refresh token を渡す
	c, err := r.Cookie("rt")
	if err != nil || c.Value == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	issuer := os.Getenv("OIDC_ISSUER")
	logoutURL := issuer + "/protocol/openid-connect/logout"

	form := url.Values{}
	form.Set("client_id", h.OIDC.Config.ClientID)
	form.Set("refresh_token", c.Value)

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, logoutURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if resp, err := http.DefaultClient.Do(req); err == nil && resp != nil {
		_ = resp.Body.Close()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "rt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // 本番は有効
		MaxAge: 0, // delete
	})
	w.WriteHeader(http.StatusNoContent)
}
