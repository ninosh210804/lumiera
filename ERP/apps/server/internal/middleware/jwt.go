package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
	"github.com/ninosh210804/lumiera/apps/server/internal/token"
)

type contextKey string

const claimsKey contextKey = "claims"

// Authenticate extracts and validates the Bearer JWT from the Authorization header.
// On success, domain.Claims are injected into the request context.
func Authenticate(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				Error(w, domain.NewUnauthorized("missing or malformed Authorization header"))
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := token.Parse(secret, tokenStr)
			if err != nil {
				Error(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFrom retrieves the authenticated claims from the request context.
// Returns nil, false if the context has no claims (unauthenticated request).
func ClaimsFrom(ctx context.Context) (*domain.Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*domain.Claims)
	return c, ok && c != nil
}
