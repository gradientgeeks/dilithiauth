package developer

import (
	"sync"
	"time"
)

// SessionRevocationStore tracks revoked session JTIs until their natural
// expiry. Sessions are stateless JWTs, so without this a "logged out"
// session remains fully valid until its 24h MaxAge elapses.
type SessionRevocationStore interface {
	// Revoke marks the session with the given jti as revoked until expiry.
	Revoke(jti string, expiry time.Time) error
	// IsRevoked returns true if the session has been explicitly revoked and
	// has not yet reached its natural expiry.
	IsRevoked(jti string) bool
}

// MemorySessionRevocationStore is an in-memory SessionRevocationStore.
// Suitable for single-node deployments; use a Redis-backed store for HA.
type MemorySessionRevocationStore struct {
	mu      sync.RWMutex
	revoked map[string]time.Time // jti -> session expiry
}

// NewMemorySessionRevocationStore creates an empty MemorySessionRevocationStore.
func NewMemorySessionRevocationStore() *MemorySessionRevocationStore {
	return &MemorySessionRevocationStore{revoked: make(map[string]time.Time)}
}

func (s *MemorySessionRevocationStore) Revoke(jti string, expiry time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[jti] = expiry
	return nil
}

func (s *MemorySessionRevocationStore) IsRevoked(jti string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exp, ok := s.revoked[jti]
	if !ok {
		return false
	}
	return time.Now().Before(exp)
}

// Purge removes entries whose session expiry has passed, freeing memory.
// Call it periodically (e.g. every hour) in a background goroutine.
func (s *MemorySessionRevocationStore) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for jti, exp := range s.revoked {
		if now.After(exp) {
			delete(s.revoked, jti)
		}
	}
}
