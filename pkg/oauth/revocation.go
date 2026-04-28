package oauth

import (
	"sync"
	"time"
)

// RevocationStore tracks revoked token JTIs until their natural expiry.
type RevocationStore interface {
	// Revoke marks the token with the given JTI as revoked until expiry.
	Revoke(jti string, expiry time.Time) error
	// IsRevoked returns true if the token has been explicitly revoked and has
	// not yet reached its natural expiry.
	IsRevoked(jti string) bool
}

// MemoryRevocationStore is an in-memory RevocationStore.
// Suitable for single-node deployments; use a Redis/DB-backed store for HA.
type MemoryRevocationStore struct {
	mu      sync.RWMutex
	revoked map[string]time.Time // jti → token expiry
}

// NewMemoryRevocationStore creates an empty MemoryRevocationStore.
func NewMemoryRevocationStore() *MemoryRevocationStore {
	return &MemoryRevocationStore{revoked: make(map[string]time.Time)}
}

func (s *MemoryRevocationStore) Revoke(jti string, expiry time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[jti] = expiry
	return nil
}

func (s *MemoryRevocationStore) IsRevoked(jti string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exp, ok := s.revoked[jti]
	if !ok {
		return false
	}
	// Once the token would have expired anyway, it no longer occupies the list
	// after Purge runs — but IsRevoked still returns true until then.
	return time.Now().Before(exp)
}

// Purge removes entries whose token expiry has passed, freeing memory.
// Call it periodically (e.g. every hour) in a background goroutine.
func (s *MemoryRevocationStore) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for jti, exp := range s.revoked {
		if now.After(exp) {
			delete(s.revoked, jti)
		}
	}
}
