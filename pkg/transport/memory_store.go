package transport

import (
	"sync"
	"time"
)

type sigEntry struct {
	sig       []byte
	expiresAt time.Time // zero value = never expires
}

// MemorySigStore is an in-memory implementation of SigStoreBackend.
// It stores ML-DSA-44 signatures in a thread-safe map.
//
// Entries expire after ttl (set via NewMemorySigStoreWithTTL) so the store
// doesn't grow unboundedly for the life of the process — a thin/CWT token is
// only ever verified while its access token is still valid, so signatures
// don't need to outlive that window. Purge() should be called periodically
// to actually reclaim memory from expired entries; Get() also treats an
// expired entry as a miss even before it's purged.
type MemorySigStore struct {
	mu    sync.RWMutex
	store map[string]sigEntry
	ttl   time.Duration // 0 = entries never expire
}

// NewMemorySigStore creates an empty in-memory signature store whose entries
// never expire. Prefer NewMemorySigStoreWithTTL in production.
func NewMemorySigStore() *MemorySigStore {
	return &MemorySigStore{store: make(map[string]sigEntry)}
}

// NewMemorySigStoreWithTTL creates an in-memory signature store whose
// entries expire ttl after being Put, bounding the store's memory growth.
func NewMemorySigStoreWithTTL(ttl time.Duration) *MemorySigStore {
	return &MemorySigStore{store: make(map[string]sigEntry), ttl: ttl}
}

func (m *MemorySigStore) Put(jti string, sig []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := make([]byte, len(sig))
	copy(stored, sig)
	entry := sigEntry{sig: stored}
	if m.ttl > 0 {
		entry.expiresAt = time.Now().Add(m.ttl)
	}
	m.store[jti] = entry
}

func (m *MemorySigStore) Get(jti string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.store[jti]
	if !ok {
		return nil, false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.sig, true
}

func (m *MemorySigStore) Delete(jti string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, jti)
}

func (m *MemorySigStore) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.store)
}

// Purge removes entries whose ttl has elapsed, freeing memory. Call it
// periodically (e.g. hourly) from a background goroutine.
func (m *MemorySigStore) Purge() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for jti, entry := range m.store {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(m.store, jti)
		}
	}
}
