package transport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileSigStore is a file-backed implementation of SigStoreBackend.
// It keeps an in-memory map for fast reads and flushes to disk on every
// Put/Delete via atomic write (temp file + os.Rename).
type FileSigStore struct {
	mu   sync.RWMutex
	data map[string][]byte
	path string
}

// NewFileSigStore creates a FileSigStore that persists to the given path.
// If the file exists, its contents are loaded. If it does not exist, the
// store starts empty and the file is created on the first write.
func NewFileSigStore(path string) (*FileSigStore, error) {
	fs := &FileSigStore{
		data: make(map[string][]byte),
		path: path,
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
	fs.data[jti] = stored
	_ = fs.flushLocked()
}

func (fs *FileSigStore) Get(jti string) ([]byte, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	sig, ok := fs.data[jti]
	return sig, ok
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
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("file_store: writing temp file: %w", err)
	}
	if err := os.Rename(tmp, fs.path); err != nil {
		return fmt.Errorf("file_store: renaming temp file: %w", err)
	}
	return nil
}
