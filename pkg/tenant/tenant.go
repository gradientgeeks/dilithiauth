// Package tenant provides tenant ID extraction and context propagation
// for multi-tenant deployments of DilithiAuth.
//
// Tenant resolution priority (highest first):
//  1. X-Tenant-Id HTTP header — only honored from a trusted network when
//     using NewMiddleware(trustedCIDRs); Middleware honors it unconditionally
//     and should only be used in tests/single-tenant deployments.
//  2. Subdomain of the Host header (e.g. acme.auth.dilithiauth.com → "acme")
//  3. DefaultTenantID ("default") for single-tenant / dev deployments
package tenant

import (
	"context"
	"net"
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
// request context, honoring the X-Tenant-Id header from any caller.
//
// Deprecated: this trusts X-Tenant-Id unconditionally, which lets any
// external caller claim to be acting for an arbitrary tenant. Prefer
// NewMiddleware with an explicit trusted-CIDR list in production; this is
// kept for tests and single-tenant/dev deployments where no such
// restriction is needed.
func Middleware(next http.Handler) http.Handler {
	return NewMiddleware(nil)(next)
}

// NewMiddleware returns tenant-resolution middleware that only honors the
// X-Tenant-Id header when the request's remote address falls within one of
// trustedCIDRs (e.g. an internal service network). All other callers are
// resolved via subdomain (or DefaultTenantID), regardless of any
// X-Tenant-Id header they send — this closes a tenant-spoofing vector where
// an external client could otherwise act as an arbitrary tenant simply by
// setting the header. Pass a nil/empty list to trust no callers (i.e.
// X-Tenant-Id is never honored, only subdomain/default resolution applies).
func NewMiddleware(trustedCIDRs []string) func(http.Handler) http.Handler {
	nets := parseCIDRs(trustedCIDRs)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := DefaultTenantID

			// Priority 1: explicit header, but only from a trusted network
			// (internal service-to-service calls, tests, M2M).
			if h := r.Header.Get("X-Tenant-Id"); h != "" && fromTrustedNetwork(r, nets) {
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
					if sub != "auth" && sub != "www" && sub != "api" && sub != "localhost" && sub != "dilithiauthapis" && sub != "dilithiauth" {
						tenantID = sanitize(sub)
					}
				}
			}

			next.ServeHTTP(w, r.WithContext(WithTenant(r.Context(), tenantID)))
		})
	}
}

func parseCIDRs(cidrs []string) []*net.IPNet {
	var nets []*net.IPNet
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, n, err := net.ParseCIDR(c); err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}

func fromTrustedNetwork(r *http.Request, nets []*net.IPNet) bool {
	if len(nets) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
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
