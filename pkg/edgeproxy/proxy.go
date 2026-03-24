package edgeproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

// EdgeProxy acts as an API Gateway that offloads ML-DSA-44 post-quantum
// signature verification. It validates DilithiAuth "Jumbo JWTs", strips
// the large post-quantum signature to save internal bandwidth, and issues
// a lightweight classical JWT to upstream microservices.
type EdgeProxy struct {
	// Verifier validates incoming DilithiAuth tokens (must have PQC keys)
	verifier *verifier.Verifier

	// InternalSecret is a symmetric key used to sign the stripped tokens
	// for internal microservices (e.g. HS256).
	internalSecret []byte

	// Upstream is the target microservice to reverse-proxy requests to.
	upstream *url.URL
}

// NewEdgeProxy creates a new EdgeProxy.
func NewEdgeProxy(v *verifier.Verifier, internalSecret []byte, upstreamURL string) (*EdgeProxy, error) {
	u, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}

	return &EdgeProxy{
		verifier:       v,
		internalSecret: internalSecret,
		upstream:       u,
	}, nil
}

// Handler returns an HTTP handler that proxies requests, offloading PQC verification.
func (p *EdgeProxy) Handler() http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(p.upstream)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 1. Verify the Jumbo PQC Token (this checks BOTH Ed25519 and ML-DSA-44)
		claims, err := p.verifier.VerifyJWT(tokenStr)
		if err != nil {
			http.Error(w, "invalid PQC token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 2. Strip the PQC signature and issue a lightweight internal token.
		// We issue a standard HS256 JWT containing only the standard claims.
		internalClaims := jwt.RegisteredClaims{
			Issuer:    "edge-proxy",
			Subject:   claims.Subject,
			ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(claims.IssuedAt),
			ID:        claims.TokenID,
		}

		internalToken := jwt.NewWithClaims(jwt.SigningMethodHS256, internalClaims)
		internalTokenStr, err := internalToken.SignedString(p.internalSecret)
		if err != nil {
			http.Error(w, "failed to sign internal token", http.StatusInternalServerError)
			return
		}

		// 3. Replace the Authorization header with the new lightweight token
		r.Header.Set("Authorization", "Bearer "+internalTokenStr)

		// Optional: Pass the subject directly as a header for even simpler backends
		r.Header.Set("X-User-Subject", claims.Subject)

		// 4. Forward to the upstream service
		proxy.ServeHTTP(w, r)
	})
}

// StripPQCMiddleware is a middleware version for use in existing Go routers
// instead of a full reverse proxy. It verifies the PQC token, strips it,
// signs a new internal token, and injects it into the request context/headers.
func (p *EdgeProxy) StripPQCMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			// Verify PQC
			if claims, err := p.verifier.VerifyJWT(tokenStr); err == nil {
				// Strip and Re-sign
				internalClaims := jwt.RegisteredClaims{
					Issuer:    "edge-proxy",
					Subject:   claims.Subject,
					ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt),
					IssuedAt:  jwt.NewNumericDate(claims.IssuedAt),
					ID:        claims.TokenID,
				}
				internalToken := jwt.NewWithClaims(jwt.SigningMethodHS256, internalClaims)
				if internalTokenStr, err := internalToken.SignedString(p.internalSecret); err == nil {
					// Mutate the request for downstream handlers
					r.Header.Set("Authorization", "Bearer "+internalTokenStr)
					ctx := context.WithValue(r.Context(), "user_subject", claims.Subject)
					r = r.WithContext(ctx)
				}
			} else {
				http.Error(w, "invalid PQC token: "+err.Error(), http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
