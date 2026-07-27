package transport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// fileSigEntry is the on-disk/in-memory representation of one signature.
// ExpiresAt is the zero value when the entry never expires (ttl == 0).
type fileSigEntry struct {
	Sig       []byte    `json:"sig"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// FileSigStore is a file-backed implementation of SigStoreBackend.
// It keeps an in-memory map for fast reads and flushes to disk on every
// Put/Delete via atomic write (temp file + os.Rename).
//
// Entries expire after ttl (set via NewFileSigStoreWithTTL), mirroring
// MemorySigStore, so the backing file and in-memory map don't grow forever.
type FileSigStore struct {
	mu   sync.RWMutex
	data map[string]fileSigEntry
	path string
	ttl  time.Duration // 0 = entries never expire
}

// NewFileSigStore creates a FileSigStore that persists to the given path.
// If the file exists, its contents are loaded. If it does not exist, the
// store starts empty and the file is created on the first write. Entries
// never expire; prefer NewFileSigStoreWithTTL in production.
func NewFileSigStore(path string) (*FileSigStore, error) {
	return NewFileSigStoreWithTTL(path, 0)
}

// NewFileSigStoreWithTTL is like NewFileSigStore but entries expire ttl
// after being Put (0 disables expiry).
func NewFileSigStoreWithTTL(path string, ttl time.Duration) (*FileSigStore, error) {
	fs := &FileSigStore{
		data: make(map[string]fileSigEntry),
		path: path,
		ttl:  ttl,
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fs, nil
		}
		return nil, fmt.Errorf("file_store: reading %s: %w", path, err)
	}

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &fs.data); err != nil {
			return nil, fmt.Errorf("file_store: parsing %s: %w", path, err)
		}
	}

	return fs, nil
}

func (fs *FileSigStore) Put(jti string, sig []byte) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	stored := make([]byte, len(sig))
	copy(stored, sig)
	entry := fileSigEntry{Sig: stored}
	if fs.ttl > 0 {
		entry.ExpiresAt = time.Now().Add(fs.ttl)
	}
	fs.data[jti] = entry
	_ = fs.flushLocked()
}

func (fs *FileSigStore) Get(jti string) ([]byte, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	entry, ok := fs.data[jti]
	if !ok {
		return nil, false
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Sig, true
}

func (fs *FileSigStore) Delete(jti string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.data, jti)
	_ = fs.flushLocked()
}

func (fs *FileSigStore) Len() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return len(fs.data)
}

// Purge removes expired entries and flushes the result to disk once,
// reclaiming both memory and disk space. Call it periodically (e.g. hourly)
// from a background goroutine.
func (fs *FileSigStore) Purge() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	now := time.Now()
	changed := false
	for jti, entry := range fs.data {
		if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
			delete(fs.data, jti)
			changed = true
		}
	}
	if changed {
		_ = fs.flushLocked()
	}
}

// flushLocked writes the current map to disk atomically. Caller must hold mu.
func (fs *FileSigStore) flushLocked() error {
	raw, err := json.Marshal(fs.data)
	if err != nil {
		return fmt.Errorf("file_store: marshalling: %w", err)
	}

	dir := filepath.Dir(fs.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("file_store: creating directory: %w", err)
	}

	tmp := fs.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("file_store: writing temp file: %w", err)
	}
	if err := os.Rename(tmp, fs.path); err != nil {
		return fmt.Errorf("file_store: renaming temp file: %w", err)
	}
	return nil
}
