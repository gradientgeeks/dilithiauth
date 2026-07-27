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
		next.ServeHTTP(w, r)
	})
}

// -------------------------------------------------------------------
// Signature Reference sidecar store
// -------------------------------------------------------------------

// SigStoreBackend is the interface that signature store backends must implement.
type SigStoreBackend interface {
	Put(jti string, sig []byte)
	Get(jti string) ([]byte, bool)
	Delete(jti string)
	Len() int
}

// SigStore is a thin wrapper that delegates to a SigStoreBackend. It
// implements the "Signature Reference" strategy: thin tokens carry only a
// pqc_ref hash, while the full signature lives in this sidecar keyed by jti.
type SigStore struct {
	backend SigStoreBackend
}

// NewSigStore creates a SigStore backed by an in-memory MemorySigStore.
func NewSigStore() *SigStore {
	return &SigStore{backend: NewMemorySigStore()}
}

// NewSigStoreWithBackend creates a SigStore backed by the given backend.
func NewSigStoreWithBackend(b SigStoreBackend) *SigStore {
	return &SigStore{backend: b}
}

// Put stores the ML-DSA-44 signature bytes under the given token ID.
func (s *SigStore) Put(jti string, sig []byte) {
	s.backend.Put(jti, sig)
}

// Get retrieves the ML-DSA-44 signature for the given token ID.
// Returns (sig, true) on a cache hit and (nil, false) on a miss.
func (s *SigStore) Get(jti string) ([]byte, bool) {
	return s.backend.Get(jti)
}

// Delete removes the signature entry for the given token ID.
func (s *SigStore) Delete(jti string) {
	s.backend.Delete(jti)
}

// Len returns the number of entries currently in the store.
func (s *SigStore) Len() int {
	return s.backend.Len()
}

// sigStorePurger is implemented by backends that support purging expired
// entries (MemorySigStore, FileSigStore, PostgresSigStore). RedisSigStore
// doesn't need it — Redis expires keys natively via WithTTL.
type sigStorePurger interface {
	Purge()
}

// Purge removes expired entries from the backend, if it supports purging.
// A no-op for backends without an expiry concept. Call periodically (e.g.
// hourly) from a background goroutine to bound memory/disk/DB growth.
func (s *SigStore) Purge() {
	if p, ok := s.backend.(sigStorePurger); ok {
		p.Purge()
	}
}

// sigStorePinger is implemented by backends with an external dependency to
// check (PostgresSigStore, RedisSigStore). In-memory/file backends have
// nothing to ping, so they're simply not asserted to this interface.
type sigStorePinger interface {
	Ping(ctx context.Context) error
}

// Ping reports whether the backend's external dependency (if any) is
// reachable, for use by readiness checks (e.g. an HTTP /readyz handler). It
// returns nil for backends with no external dependency (in-memory, file).
func (s *SigStore) Ping(ctx context.Context) error {
	if p, ok := s.backend.(sigStorePinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

// sigStoreCloser is implemented by backends holding an external connection
// that should be released on shutdown (PostgresSigStore via *sql.DB,
// RedisSigStore via *redis.Client). In-memory/file backends have nothing to
// close, so they're simply not asserted to this interface.
type sigStoreCloser interface {
	Close() error
}

// Close releases the backend's underlying connection, if any, for use during
// graceful shutdown. A no-op for backends with no external connection
// (in-memory, file).
func (s *SigStore) Close() error {
	if c, ok := s.backend.(sigStoreCloser); ok {
		return c.Close()
	}
	return nil
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
