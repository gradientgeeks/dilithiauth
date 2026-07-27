package tenant_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
)

func TestNewMiddleware_HonorsHeaderFromTrustedNetwork(t *testing.T) {
	mw := tenant.NewMiddleware([]string{"10.0.0.0/8"})
	var got string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = tenant.FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "acme")
	req.RemoteAddr = "10.1.2.3:5555"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "acme" {
		t.Errorf("tenant = %q, want %q", got, "acme")
	}
}

// A caller outside the trusted network must not be able to claim an
// arbitrary tenant via the X-Tenant-Id header — this is the tenant-spoofing
// vector the trusted-CIDR restriction exists to close.
func TestNewMiddleware_IgnoresHeaderFromUntrustedNetwork(t *testing.T) {
	mw := tenant.NewMiddleware([]string{"10.0.0.0/8"})
	var got string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = tenant.FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "acme")
	req.Host = "localhost"              // excluded from subdomain resolution
	req.RemoteAddr = "203.0.113.5:5555" // outside 10.0.0.0/8
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got == "acme" {
		t.Error("X-Tenant-Id from an untrusted network must not be honored")
	}
	if got != tenant.DefaultTenantID {
		t.Errorf("tenant = %q, want fallback to %q", got, tenant.DefaultTenantID)
	}
}

func TestNewMiddleware_EmptyTrustedCIDRs_NeverHonorsHeader(t *testing.T) {
	mw := tenant.NewMiddleware(nil)
	var got string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = tenant.FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "acme")
	req.Host = "localhost" // excluded from subdomain resolution
	req.RemoteAddr = "127.0.0.1:5555"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != tenant.DefaultTenantID {
		t.Errorf("tenant = %q, want %q (no trusted CIDRs configured)", got, tenant.DefaultTenantID)
	}
}

func TestNewMiddleware_SubdomainFallback(t *testing.T) {
	mw := tenant.NewMiddleware(nil)
	var got string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = tenant.FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "acme.auth.dilithiauth.com"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "acme" {
		t.Errorf("tenant = %q, want %q", got, "acme")
	}
}
