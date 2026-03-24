package signer

import (
	"fmt"
	"sync"
	"time"
)

// KeyManager manages multiple HybridKeyPairs with support for key rotation
// and persistence via a KeyStore interface.
type KeyManager struct {
	mu           sync.RWMutex
	keys         map[string]*HybridKeyPair
	activeKid    string
	store        KeyStore
	lastRefresh  time.Time
	refreshDelay time.Duration
}

// NewKeyManager creates a KeyManager that persists keys using a KeyStore.
func NewKeyManagerWithStore(store KeyStore) (*KeyManager, error) {
	km := &KeyManager{
		keys:         make(map[string]*HybridKeyPair),
		store:        store,
		refreshDelay: 10 * time.Second,
	}

	if err := km.Refresh(); err != nil {
		// If no keys exist yet, initialize with a new key
		kp, genErr := GenerateHybridKeyPair()
		if genErr != nil {
			return nil, fmt.Errorf("signer: generating initial key: %w", genErr)
		}
		if saveErr := km.store.Save(kp); saveErr != nil {
			return nil, saveErr
		}
		if setErr := km.store.SetActiveKid(kp.Kid); setErr != nil {
			return nil, setErr
		}
		km.keys[kp.Kid] = kp
		km.activeKid = kp.Kid
	}

	return km, nil
}

// NewKeyManager is kept for backward compatibility and defaults to FileKeyStore.
func NewKeyManager(dir string) (*KeyManager, error) {
	store, err := NewFileKeyStore(dir)
	if err != nil {
		return nil, err
	}
	return NewKeyManagerWithStore(store)
}

// Refresh reloads the key state from the store.
func (km *KeyManager) Refresh() error {
	km.mu.Lock()
	defer km.mu.Unlock()

	keys, err := km.store.LoadAll()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return fmt.Errorf("signer: no keys found in store")
	}

	km.keys = make(map[string]*HybridKeyPair)
	for _, k := range keys {
		km.keys[k.Kid] = k
	}

	activeKid, err := km.store.GetActiveKid()
	if err != nil {
		return err
	}
	km.activeKid = activeKid
	km.lastRefresh = time.Now()

	return nil
}

func (km *KeyManager) maybeRefresh() {
	km.mu.RLock()
	needsRefresh := time.Since(km.lastRefresh) > km.refreshDelay
	km.mu.RUnlock()

	if needsRefresh {
		// Ignore error, fallback to cached state
		_ = km.Refresh()
	}
}

// ActiveKey returns the currently active key pair used for signing.
func (km *KeyManager) ActiveKey() *HybridKeyPair {
	km.maybeRefresh()
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.keys[km.activeKid]
}

// KeyByKid returns the key pair with the given kid, or nil if not found.
func (km *KeyManager) KeyByKid(kid string) *HybridKeyPair {
	km.maybeRefresh()
	km.mu.RLock()
	kp, ok := km.keys[kid]
	km.mu.RUnlock()

	if !ok {
		// Fallback to fetch explicitly if missing
		loadedKP, err := km.store.Load(kid)
		if err == nil && loadedKP != nil {
			km.mu.Lock()
			km.keys[kid] = loadedKP
			km.mu.Unlock()
			return loadedKP
		}
	}

	return kp
}

// AllPublicKeys returns PublicKeyInfo for all managed keys.
func (km *KeyManager) AllPublicKeys() []*PublicKeyInfo {
	km.maybeRefresh()
	km.mu.RLock()
	defer km.mu.RUnlock()

	var infos []*PublicKeyInfo
	for _, kp := range km.keys {
		infos = append(infos, kp.PublicKeyInfo())
	}
	return infos
}

// Rotate generates a new key pair, persists it, and makes it active.
func (km *KeyManager) Rotate() (*HybridKeyPair, error) {
	kp, err := GenerateHybridKeyPair()
	if err != nil {
		return nil, fmt.Errorf("signer: generating rotated key: %w", err)
	}

	if err := km.store.Save(kp); err != nil {
		return nil, err
	}

	if err := km.store.SetActiveKid(kp.Kid); err != nil {
		return nil, err
	}

	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys[kp.Kid] = kp
	km.activeKid = kp.Kid
	km.lastRefresh = time.Now()

	return kp, nil
}
