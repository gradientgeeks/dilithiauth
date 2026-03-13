package verifier

import (
	"context"
	"net/http"
	"strings"
)

type claimsContextKey struct{}

// ClaimsFromContext retrieves VerifiedClaims from the request context.
func ClaimsFromContext(ctx context.Context) *VerifiedClaims {
	v, _ := ctx.Value(claimsContextKey{}).(*VerifiedClaims)
	return v
}

// Middleware returns HTTP middleware that extracts a Bearer token from the
// Authorization header, verifies it, and injects VerifiedClaims into the
// request context. Returns 401 on failure.
func (v *Verifier) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			claims, err := v.VerifyJWT(tokenStr)
			if err != nil {
				http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
