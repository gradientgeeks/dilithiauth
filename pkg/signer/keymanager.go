package signer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// KeyManager manages multiple HybridKeyPairs with support for key rotation
// and persistence to a directory of JSON files.
type KeyManager struct {
	mu        sync.RWMutex
	keys      []*HybridKeyPair
	activeKid string
	dir       string
}

// NewKeyManager creates a KeyManager that persists keys to dir. If the
// directory contains existing *.json key files, they are loaded. If no keys
// are found, a new key pair is generated.
func NewKeyManager(dir string) (*KeyManager, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("signer: creating key dir %s: %w", dir, err)
	}

	km := &KeyManager{dir: dir}

	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("signer: listing key files: %w", err)
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("signer: reading %s: %w", path, err)
		}
		kp, err := UnmarshalKeyPair(data)
		if err != nil {
			return nil, fmt.Errorf("signer: loading key from %s: %w", path, err)
		}
		km.keys = append(km.keys, kp)
	}

	if len(km.keys) == 0 {
		kp, err := GenerateHybridKeyPair()
		if err != nil {
			return nil, fmt.Errorf("signer: generating initial key: %w", err)
		}
		if err := km.saveKey(kp); err != nil {
			return nil, err
		}
		km.keys = append(km.keys, kp)
	}

	// The last key loaded (or the freshly generated one) is the active key.
	km.activeKid = km.keys[len(km.keys)-1].Kid

	return km, nil
}

// ActiveKey returns the currently active key pair used for signing.
func (km *KeyManager) ActiveKey() *HybridKeyPair {
	km.mu.RLock()
	defer km.mu.RUnlock()
	for _, kp := range km.keys {
		if kp.Kid == km.activeKid {
			return kp
		}
	}
	return nil
}

// KeyByKid returns the key pair with the given kid, or nil if not found.
func (km *KeyManager) KeyByKid(kid string) *HybridKeyPair {
	km.mu.RLock()
	defer km.mu.RUnlock()
	for _, kp := range km.keys {
		if kp.Kid == kid {
			return kp
		}
	}
	return nil
}

// AllPublicKeys returns PublicKeyInfo for all managed keys.
func (km *KeyManager) AllPublicKeys() []*PublicKeyInfo {
	km.mu.RLock()
	defer km.mu.RUnlock()
	infos := make([]*PublicKeyInfo, len(km.keys))
	for i, kp := range km.keys {
		infos[i] = kp.PublicKeyInfo()
	}
	return infos
}

// Rotate generates a new key pair, persists it, and makes it active.
func (km *KeyManager) Rotate() (*HybridKeyPair, error) {
	kp, err := GenerateHybridKeyPair()
	if err != nil {
		return nil, fmt.Errorf("signer: generating rotated key: %w", err)
	}
	if err := km.saveKey(kp); err != nil {
		return nil, err
	}

	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys = append(km.keys, kp)
	km.activeKid = kp.Kid
	return kp, nil
}

func (km *KeyManager) saveKey(kp *HybridKeyPair) error {
	data, err := MarshalKeyPair(kp)
	if err != nil {
		return fmt.Errorf("signer: marshalling key %s: %w", kp.Kid, err)
	}
	path := filepath.Join(km.dir, kp.Kid+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("signer: writing key %s: %w", kp.Kid, err)
	}
	return nil
}
