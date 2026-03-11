// Package transport provides HTTP middleware and server configuration helpers
// for handling "Jumbo Tokens" – large PQC-signed tokens that exceed the
// default 4 KB header limit of most web infrastructure.
//
// It also implements the "Signature Reference" sidecar strategy that keeps
// tokens small by storing the heavy ML-DSA-44 signature out-of-band.
package transport

import (
	"context"
	"net/http"
	"sync"
)

// MaxHeaderBytesForPQC is the recommended MaxHeaderBytes value for a Go
// net/http server that needs to accept PQC-sized tokens in Authorization
// headers.  16 KiB safely accommodates a hybrid JWT (~3.5–4 KB) while
// staying conservative for non-PQC deployments.
const MaxHeaderBytesForPQC = 1 << 14 // 16 KiB

// NewPQCServer returns a pre-configured *http.Server with MaxHeaderBytes set
// to MaxHeaderBytesForPQC.  All other settings are left at their zero values
// so the caller can further configure Addr, Handler, TLSConfig, etc.
func NewPQCServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        handler,
		MaxHeaderBytes: MaxHeaderBytesForPQC,
	}
}

// -------------------------------------------------------------------
// Jumbo-header middleware
// -------------------------------------------------------------------

// JumboHeaderMiddleware is HTTP middleware that simply allows large headers to
// reach the handler.  The actual size enforcement happens at the server level
// via MaxHeaderBytes; this middleware exists to document the intent and to
// provide a hook for adding header-size metrics in the future.
func JumboHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No action needed – the net/http server already enforces MaxHeaderBytes.
		// This middleware documents the flow and can be extended for logging.
		next.ServeHTTP(w, r)
	})
}

// -------------------------------------------------------------------
// Signature Reference sidecar store
// -------------------------------------------------------------------

// SigStore is a thread-safe in-memory store for ML-DSA-44 signature bytes
// keyed by token ID (jti / cti).  In production, replace this with a Redis
// client or another distributed cache so the store is shared across instances.
//
// The store implements the "Signature Reference" strategy described in the
// DilithiAuth design document: thin tokens carry only a pqc_ref hash, while
// the full signature lives in this sidecar keyed by jti.
type SigStore struct {
	mu    sync.RWMutex
	store map[string][]byte
}

// NewSigStore creates an empty SigStore.
func NewSigStore() *SigStore {
	return &SigStore{store: make(map[string][]byte)}
}

// Put stores the ML-DSA-44 signature bytes under the given token ID.
func (s *SigStore) Put(jti string, sig []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Copy to avoid the caller mutating the stored slice.
	stored := make([]byte, len(sig))
	copy(stored, sig)
	s.store[jti] = stored
}

// Get retrieves the ML-DSA-44 signature for the given token ID.
// Returns (sig, true) on a cache hit and (nil, false) on a miss.
func (s *SigStore) Get(jti string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sig, ok := s.store[jti]
	return sig, ok
}

// Delete removes the signature entry for the given token ID.
// Call this when a token is revoked or has expired.
func (s *SigStore) Delete(jti string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, jti)
}

// Len returns the number of entries currently in the store.
func (s *SigStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.store)
}

// -------------------------------------------------------------------
// Context helpers
// -------------------------------------------------------------------

type contextKey string

const sigStoreKey contextKey = "sig_store"

// WithSigStore attaches the given SigStore to the request context so that
// downstream handlers can retrieve the full ML-DSA-44 signature when needed.
func WithSigStore(ctx context.Context, store *SigStore) context.Context {
	return context.WithValue(ctx, sigStoreKey, store)
}

// SigStoreFromContext retrieves the SigStore previously attached by
// WithSigStore.  Returns nil if no store is present.
func SigStoreFromContext(ctx context.Context) *SigStore {
	v, _ := ctx.Value(sigStoreKey).(*SigStore)
	return v
}

// SigStoreMiddleware attaches the given SigStore to every request context so
// that handlers can perform sidecar lookups without an explicit dependency.
func SigStoreMiddleware(store *SigStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(WithSigStore(r.Context(), store))
			next.ServeHTTP(w, r)
		})
	}
}
