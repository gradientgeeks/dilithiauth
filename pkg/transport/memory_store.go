package transport

import "sync"

// MemorySigStore is an in-memory implementation of SigStoreBackend.
// It stores ML-DSA-44 signatures in a thread-safe map.
type MemorySigStore struct {
	mu    sync.RWMutex
	store map[string][]byte
}

// NewMemorySigStore creates an empty in-memory signature store.
func NewMemorySigStore() *MemorySigStore {
	return &MemorySigStore{store: make(map[string][]byte)}
}

func (m *MemorySigStore) Put(jti string, sig []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := make([]byte, len(sig))
	copy(stored, sig)
	m.store[jti] = stored
}

func (m *MemorySigStore) Get(jti string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sig, ok := m.store[jti]
	return sig, ok
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
