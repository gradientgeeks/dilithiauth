package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
)

func TestNewPQCServer(t *testing.T) {
	srv := transport.NewPQCServer(":8080", http.DefaultServeMux)
	if srv.MaxHeaderBytes != transport.MaxHeaderBytesForPQC {
		t.Errorf("MaxHeaderBytes = %d, want %d", srv.MaxHeaderBytes, transport.MaxHeaderBytesForPQC)
	}
	if srv.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", srv.Addr, ":8080")
	}
}

func TestMaxHeaderBytesForPQC(t *testing.T) {
	// 16 KiB = 16384 bytes – must be > 4096 (standard limit) and <= 65536.
	if transport.MaxHeaderBytesForPQC <= 4096 {
		t.Errorf("MaxHeaderBytesForPQC = %d, must be > 4096", transport.MaxHeaderBytesForPQC)
	}
	if transport.MaxHeaderBytesForPQC > 1<<16 {
		t.Errorf("MaxHeaderBytesForPQC = %d, sanity upper bound exceeded", transport.MaxHeaderBytesForPQC)
	}
}

func TestJumboHeaderMiddleware(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := transport.JumboHeaderMiddleware(inner)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Simulate a large Authorization header.
	req.Header.Set("Authorization", "Bearer "+string(make([]byte, 3000)))

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("inner handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// --- SigStore tests ---

func TestSigStorePutGet(t *testing.T) {
	store := transport.NewSigStore()
	sig := []byte("fake-mldsa44-signature-bytes")
	store.Put("jti-123", sig)

	got, ok := store.Get("jti-123")
	if !ok {
		t.Fatal("Get() = false, want true")
	}
	if string(got) != string(sig) {
		t.Errorf("Get() = %q, want %q", got, sig)
	}
}

func TestSigStore_MutationSafety(t *testing.T) {
	store := transport.NewSigStore()
	original := []byte{1, 2, 3}
	store.Put("jti-mut", original)

	// Mutate the original slice after storing.
	original[0] = 99

	got, _ := store.Get("jti-mut")
	if got[0] == 99 {
		t.Error("SigStore did not copy the slice – mutation affected stored value")
	}
}

func TestSigStoreDelete(t *testing.T) {
	store := transport.NewSigStore()
	store.Put("jti-del", []byte("sig"))
	store.Delete("jti-del")

	if _, ok := store.Get("jti-del"); ok {
		t.Error("Get() after Delete() = true, want false")
	}
}

func TestSigStoreLen(t *testing.T) {
	store := transport.NewSigStore()
	if store.Len() != 0 {
		t.Errorf("initial Len = %d, want 0", store.Len())
	}
	store.Put("a", []byte("sig-a"))
	store.Put("b", []byte("sig-b"))
	if store.Len() != 2 {
		t.Errorf("Len after 2 puts = %d, want 2", store.Len())
	}
}

func TestSigStoreGetMiss(t *testing.T) {
	store := transport.NewSigStore()
	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("Get() on empty store = true, want false")
	}
}

// --- Context helpers ---

func TestWithSigStore(t *testing.T) {
	store := transport.NewSigStore()
	ctx := transport.WithSigStore(context.Background(), store)

	got := transport.SigStoreFromContext(ctx)
	if got != store {
		t.Error("SigStoreFromContext returned different store than was injected")
	}
}

func TestSigStoreFromContext_Empty(t *testing.T) {
	got := transport.SigStoreFromContext(context.Background())
	if got != nil {
		t.Error("SigStoreFromContext on empty context should return nil")
	}
}

func TestSigStoreMiddleware(t *testing.T) {
	store := transport.NewSigStore()
	var captured *transport.SigStore

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = transport.SigStoreFromContext(r.Context())
	})
	handler := transport.SigStoreMiddleware(store)(inner)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if captured != store {
		t.Error("middleware did not inject SigStore into request context")
	}
}
