package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// RefreshTokenRecord represents one issued refresh token.
type RefreshTokenRecord struct {
	Token    string
	Subject  string
	TenantID string
	ClientID string
	Scope    string
	// FamilyID links every refresh token descended from a single
	// authorization_code exchange. Rotating a refresh token creates a new
	// record in the same family; replaying an already-used token revokes
	// the whole family (RevokeFamily), since that pattern only happens if a
	// token was stolen and both the legitimate holder and the thief tried
	// to use it.
	FamilyID  string
	ExpiresAt time.Time
	Used      bool
}

// RefreshTokenStore manages refresh token issuance, rotation, and
// reuse/theft detection. Unlike RevocationStore (a blocklist of otherwise
// generally-valid tokens), this is a positive record store: a refresh token
// is only valid if a matching, unused, unexpired, unrevoked record exists.
type RefreshTokenStore interface {
	Create(rec *RefreshTokenRecord) error
	Get(token string) (*RefreshTokenRecord, error)
	MarkUsed(token string) error
	RevokeFamily(familyID string) error
	Purge()
}

// MemoryRefreshTokenStore is an in-memory RefreshTokenStore. Suitable for
// single-node deployments; a Redis-backed implementation (mirroring
// AuthzStore/RedisAuthzStore's split) is a natural follow-up for HA.
type MemoryRefreshTokenStore struct {
	mu      sync.Mutex
	tokens  map[string]*RefreshTokenRecord // token -> record
	revoked map[string]bool                // familyID -> revoked
}

// NewMemoryRefreshTokenStore creates an empty MemoryRefreshTokenStore.
func NewMemoryRefreshTokenStore() *MemoryRefreshTokenStore {
	return &MemoryRefreshTokenStore{
		tokens:  make(map[string]*RefreshTokenRecord),
		revoked: make(map[string]bool),
	}
}

func (s *MemoryRefreshTokenStore) Create(rec *RefreshTokenRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[rec.Token] = rec
	return nil
}

func (s *MemoryRefreshTokenStore) Get(token string) (*RefreshTokenRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.tokens[token]
	if !ok {
		return nil, errors.New("refresh token not found")
	}
	if s.revoked[rec.FamilyID] {
		return nil, errors.New("refresh token family revoked")
	}
	return rec, nil
}

func (s *MemoryRefreshTokenStore) MarkUsed(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.tokens[token]
	if !ok {
		return errors.New("refresh token not found")
	}
	rec.Used = true
	return nil
}

func (s *MemoryRefreshTokenStore) RevokeFamily(familyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[familyID] = true
	return nil
}

// Purge removes expired token records, bounding memory growth. Revoked
// family markers are intentionally left in place — they're only created on
// theft-detection events, so they're expected to stay rare, and dropping one
// prematurely would let a revoked family's remaining tokens work again.
func (s *MemoryRefreshTokenStore) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for token, rec := range s.tokens {
		if now.After(rec.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}

// generateRefreshToken returns a random, URL-safe refresh token value.
func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "rt_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

// generateFamilyID returns a random family identifier for a new refresh-token lineage.
func generateFamilyID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
