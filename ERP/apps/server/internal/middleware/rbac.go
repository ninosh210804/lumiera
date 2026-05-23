package middleware

import (
	"net/http"

	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

// RequireAuth rejects unauthenticated requests (no valid claims in context).
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := ClaimsFrom(r.Context()); !ok {
			Error(w, domain.NewUnauthorized("authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole allows only users whose role matches one of the given roles.
// Authenticate middleware must run first.
func RequireRole(roles ...domain.RoleType) func(http.Handler) http.Handler {
	set := make(map[domain.RoleType]struct{}, len(roles))
	for _, r := range roles {
		set[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFrom(r.Context())
			if !ok {
				Error(w, domain.NewUnauthorized("authentication required"))
				return
			}
			if _, allowed := set[claims.Role]; !allowed {
				Error(w, domain.ErrInsufficientPerm)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
