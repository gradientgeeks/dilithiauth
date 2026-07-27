package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/ratelimit"
)

func TestLimiter_AllowsUpToBurstThenBlocks(t *testing.T) {
	l := ratelimit.NewLimiter(0, 3) // rate=0 so no mid-test refill races
	for i := 0; i < 3; i++ {
		if !l.Allow("a") {
			t.Fatalf("request %d should be allowed within burst", i)
		}
	}
	if l.Allow("a") {
		t.Error("request exceeding burst should be blocked")
	}
}

func TestLimiter_KeysAreIndependent(t *testing.T) {
	l := ratelimit.NewLimiter(0, 1)
	if !l.Allow("a") {
		t.Fatal("first request for key a should be allowed")
	}
	if !l.Allow("b") {
		t.Error("key b should have its own independent budget")
	}
}

func TestMiddleware_BlocksOverLimit(t *testing.T) {
	l := ratelimit.NewLimiter(0, 1)
	mw := ratelimit.Middleware(l, ratelimit.ClientIPKey())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "203.0.113.5:12345"

	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", rec1.Code, http.StatusOK)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}
