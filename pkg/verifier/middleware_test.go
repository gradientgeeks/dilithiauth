package verifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

func TestMiddleware_ValidToken(t *testing.T) {
	p, v, store := setup(t)

	result, err := p.IssueJWTThin("alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	var captured *verifier.VerifiedClaims
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = verifier.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := v.Middleware()(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+result.Token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if captured == nil {
		t.Fatal("claims not injected into context")
	}
	if captured.Subject != "alice" {
		t.Errorf("Subject = %q, want %q", captured.Subject, "alice")
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	_, v, _ := setup(t)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	handler := v.Middleware()(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	_, v, _ := setup(t)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	handler := v.Middleware()(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMiddleware_FullToken(t *testing.T) {
	p, v, _ := setup(t)

	tokenStr, err := p.IssueJWTFull("bob", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTFull() error = %v", err)
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := verifier.ClaimsFromContext(r.Context())
		_ = json.NewEncoder(w).Encode(claims)
	})

	handler := v.Middleware()(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
