package verifier

import (
	"context"
	"net/http"
	"strings"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
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

			// Reject a token issued under one tenant if the request resolved
			// to a different tenant (e.g. via subdomain or a trusted
			// X-Tenant-Id header) — otherwise a token minted for "alice" in
			// tenant A would be indistinguishable from tenant B's "alice".
			requestTenant := tenant.FromContext(r.Context())
			if claims.TenantID != "" && claims.TenantID != requestTenant {
				http.Error(w, "invalid token: tenant mismatch", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
