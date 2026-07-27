package edgeproxy

import (
	"context"
	"crypto/ed25519"
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
// a lightweight classical JWT (HS256 or Ed25519) to upstream microservices.
type EdgeProxy struct {
	// Verifier validates incoming DilithiAuth tokens (must have PQC keys)
	verifier *verifier.Verifier

	// InternalSecret is a symmetric key used to sign the stripped tokens (e.g. HS256).
	internalSecret []byte

	// InternalEdKey is an asymmetric Ed25519 private key used to sign stripped tokens
	// for zero-trust microservice environments.
	internalEdKey ed25519.PrivateKey

	// Upstream is the target microservice to reverse-proxy requests to.
	upstream *url.URL

	// ExcludedPaths defines routes that bypass PQC token checking (e.g. /healthz, /public).
	ExcludedPaths []string
}

// NewEdgeProxy creates a new EdgeProxy with symmetric (HS256) internal re-signing.
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

// NewEdgeProxyAsymmetric creates an EdgeProxy with asymmetric (Ed25519) internal re-signing.
func NewEdgeProxyAsymmetric(v *verifier.Verifier, edKey ed25519.PrivateKey, upstreamURL string) (*EdgeProxy, error) {
	u, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}

	return &EdgeProxy{
		verifier:      v,
		internalEdKey: edKey,
		upstream:      u,
	}, nil
}

func (p *EdgeProxy) isExcluded(path string) bool {
	for _, exp := range p.ExcludedPaths {
		if strings.HasPrefix(path, exp) {
			return true
		}
	}
	return false
}

// Handler returns an HTTP handler that proxies requests, offloading PQC verification.
func (p *EdgeProxy) Handler() http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(p.upstream)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.isExcluded(r.URL.Path) {
			proxy.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 1. Verify the Jumbo PQC Token (checks BOTH Ed25519 and ML-DSA-44)
		claims, err := p.verifier.VerifyJWT(tokenStr)
		if err != nil {
			http.Error(w, "invalid PQC token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 2. Strip the PQC signature and issue a lightweight internal token.
		internalClaims := jwt.RegisteredClaims{
			Issuer:    "edge-proxy",
			Subject:   claims.Subject,
			ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(claims.IssuedAt),
			ID:        claims.TokenID,
		}

		var internalTokenStr string
		if p.internalEdKey != nil {
			// Asymmetric Ed25519 Re-Signing (Zero-Trust)
			token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, internalClaims)
			internalTokenStr, err = token.SignedString(p.internalEdKey)
		} else {
			// Symmetric HS256 Re-Signing
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, internalClaims)
			internalTokenStr, err = token.SignedString(p.internalSecret)
		}

		if err != nil {
			http.Error(w, "failed to sign internal token", http.StatusInternalServerError)
			return
		}

		// 3. Replace the Authorization header with the lightweight internal token
		r.Header.Set("Authorization", "Bearer "+internalTokenStr)
		r.Header.Set("X-User-Subject", claims.Subject)
		r.Header.Set("X-Token-ID", claims.TokenID)
		r.Header.Set("X-Tenant-Id", claims.TenantID)

		// 4. Forward to upstream microservice
		proxy.ServeHTTP(w, r)
	})
}

// StripPQCMiddleware is a middleware version for use in Go routers.
func (p *EdgeProxy) StripPQCMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.isExcluded(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			if claims, err := p.verifier.VerifyJWT(tokenStr); err == nil {
				internalClaims := jwt.RegisteredClaims{
					Issuer:    "edge-proxy",
					Subject:   claims.Subject,
					ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt),
					IssuedAt:  jwt.NewNumericDate(claims.IssuedAt),
					ID:        claims.TokenID,
				}

				var internalTokenStr string
				if p.internalEdKey != nil {
					token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, internalClaims)
					internalTokenStr, err = token.SignedString(p.internalEdKey)
				} else {
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, internalClaims)
					internalTokenStr, err = token.SignedString(p.internalSecret)
				}

				if err == nil {
					r.Header.Set("Authorization", "Bearer "+internalTokenStr)
					r.Header.Set("X-User-Subject", claims.Subject)
					r.Header.Set("X-Token-ID", claims.TokenID)
					r.Header.Set("X-Tenant-Id", claims.TenantID)
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
