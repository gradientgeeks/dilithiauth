// Package tenant provides tenant ID extraction and context propagation
// for multi-tenant deployments of DilithiAuth.
//
// Tenant resolution priority (highest first):
//  1. X-Tenant-Id HTTP header
//  2. Subdomain of the Host header (e.g. acme.auth.dilithiauth.com → "acme")
//  3. DefaultTenantID ("default") for single-tenant / dev deployments
package tenant

import (
	"context"
	"net/http"
	"strings"
)

// DefaultTenantID is used when no tenant is specified — preserves backward
// compatibility with the existing single-tenant cmd/server binary.
const DefaultTenantID = "default"

type contextKey struct{}

// WithTenant attaches a tenant ID to ctx.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, contextKey{}, tenantID)
}

// FromContext returns the tenant ID stored in ctx, or DefaultTenantID.
func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKey{}).(string); ok && v != "" {
		return v
	}
	return DefaultTenantID
}

// Middleware resolves the tenant for every request and injects it into the
// request context. Attach it early in the middleware chain.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := DefaultTenantID

		// Priority 1: explicit header (useful for M2M, tests, and internal calls)
		if h := r.Header.Get("X-Tenant-Id"); h != "" {
			tenantID = sanitize(h)
		} else {
			// Priority 2: leftmost subdomain, skipping common non-tenant labels
			host := r.Host
			if colon := strings.LastIndex(host, ":"); colon != -1 {
				host = host[:colon] // strip :port
			}
			parts := strings.SplitN(host, ".", 2)
			if len(parts) == 2 {
				sub := parts[0]
				if sub != "auth" && sub != "www" && sub != "api" && sub != "localhost" {
					tenantID = sanitize(sub)
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(WithTenant(r.Context(), tenantID)))
	})
}

// sanitize keeps only alphanumerics and hyphens, limiting to 63 chars
// (DNS label limit) to prevent key-injection attacks in Redis prefixes.
func sanitize(s string) string {
	var b strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' {
			b.WriteRune(c)
		}
		if b.Len() == 63 {
			break
		}
	}
	if b.Len() == 0 {
		return DefaultTenantID
	}
	return b.String()
}
