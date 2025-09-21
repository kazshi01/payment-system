package httpi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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

	// CSRF: state
	cState, _ := r.Cookie(cookieStateKey)
	if cState == nil || cState.Value != state {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	// PKCE: code_verifier
	cVerifier, _ := r.Cookie(cookieVerifierKey)
	if cVerifier == nil || cVerifier.Value == "" {
		http.Error(w, "missing verifier", http.StatusBadRequest)
		return
	}

	// トークン交換（PKCE）
	tok, err := h.OIDC.Config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", cVerifier.Value),
	)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 一時Cookie（state / verifier）は使い終わったので削除
	http.SetCookie(w, &http.Cookie{
		Name:     cookieStateKey,
		Value:    "",
		Path:     "/",
		MaxAge:   0,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     cookieVerifierKey,
		Value:    "",
		Path:     "/",
		MaxAge:   0,
		SameSite: http.SameSiteLaxMode,
	})

	// access_token を可視Cookieに保存（本番は HttpOnly + BFF推奨）
	ttl := int(time.Until(tok.Expiry).Seconds())
	if ttl < 1 {
		ttl = 60 // 最低1分は保持（学習用）
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "at",
		Value:    tok.AccessToken, // ← ここは tok.AccessToken を使う
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   ttl,
		// Secure: true, // 本番は true（HTTPS）
	})

	// refresh_token は HttpOnly で保存（Keycloakはローテーションあり）
	if rt := tok.RefreshToken; rt != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "rt",
			Value:    rt,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   60 * 60, // 本番はIdP設定に合わせる／offline_accessならもっと長い
			// Secure: true,
		})
	}

	// キャッシュさせない & リダイレクト
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

// GET/POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Cookie 取得
	var rt string
	if c, err := r.Cookie("rt"); err == nil {
		rt = c.Value
	}

	// Keycloak 側のセッションも終了（refresh_token がある場合のみ）
	if rt != "" {
		issuer := os.Getenv("OIDC_ISSUER")
		logoutURL := issuer + "/protocol/openid-connect/logout"

		form := url.Values{}
		form.Set("client_id", h.OIDC.Config.ClientID)
		form.Set("refresh_token", rt)

		req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, logoutURL, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if resp, err := http.DefaultClient.Do(req); err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}

	// ローカル Cookie を削除（at / rt 両方）
	http.SetCookie(w, &http.Cookie{
		Name:     "at",
		Value:    "",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   0,
		// Secure: true, // 本番
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "rt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   0,
		// Secure: true, // 本番
	})

	// キャッシュさせず、ログイン画面へ 303 リダイレクト
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther) // 303
}
