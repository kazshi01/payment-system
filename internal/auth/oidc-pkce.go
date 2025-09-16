package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDC struct {
	Provider *oidc.Provider
	Config   *oauth2.Config
}

type OIDCConfig struct {
	Issuer      string
	ClientID    string
	RedirectURL string
	Scopes      []string
}

func NewOIDC(ctx context.Context, c OIDCConfig) (*OIDC, error) {
	provider, err := oidc.NewProvider(ctx, c.Issuer)
	if err != nil {
		return nil, err
	}
	conf := &oauth2.Config{
		ClientID:    c.ClientID,
		RedirectURL: c.RedirectURL,
		Endpoint:    provider.Endpoint(),
		Scopes:      append([]string{"openid"}, c.Scopes...),
	}
	return &OIDC{Provider: provider, Config: conf}, nil
}

func RandBase64URL(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func CodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func AddQuery(u string, kv map[string]string) string {
	uu, _ := url.Parse(u)
	q := uu.Query()
	for k, v := range kv {
		q.Set(k, v)
	}
	uu.RawQuery = q.Encode()
	return uu.String()
}
