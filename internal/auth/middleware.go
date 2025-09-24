package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type Config struct {
	Issuer   string
	Audience string
}

type ctxKey string

const ClaimsKey ctxKey = "claims"

func Middleware(cfg Config) (func(http.Handler) http.Handler, error) {
	provider, err := oidc.NewProvider(context.Background(), cfg.Issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID:          cfg.Audience,
		SkipClientIDCheck: false,
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			raw := strings.TrimPrefix(authz, "Bearer ")

			idt, err := verifier.Verify(r.Context(), raw)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// claims を取り出して context に保存
			var claims map[string]any
			_ = idt.Claims(&claims)

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}, nil
}

// 取り出しヘルパ
func UserIDFrom(ctx context.Context) (string, bool) {
	if m, ok := ctx.Value(ClaimsKey).(map[string]any); ok {
		if s, ok := m["sub"].(string); ok {
			return s, true
		}
	}
	return "", false
}

func HasRealmRole(ctx context.Context, role string) bool {
	m, ok := ctx.Value(ClaimsKey).(map[string]any)
	if !ok {
		return false
	}
	ra, ok := m["realm_access"].(map[string]any)
	if !ok {
		return false
	}
	roles, ok := ra["roles"].([]any)
	if !ok {
		return false
	}
	for _, v := range roles {
		if s, ok := v.(string); ok && s == role {
			return true
		}
	}
	return false
}

func IsAdmin(ctx context.Context) bool {
	return HasRealmRole(ctx, "payment_admin")
}
