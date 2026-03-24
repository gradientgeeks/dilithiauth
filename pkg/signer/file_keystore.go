package signer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileKeyStore implements KeyStore using a local directory.
type FileKeyStore struct {
	dir string
}

// NewFileKeyStore creates a new FileKeyStore.
func NewFileKeyStore(dir string) (*FileKeyStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("signer: creating key dir %s: %w", dir, err)
	}
	return &FileKeyStore{dir: dir}, nil
}

func (s *FileKeyStore) Save(kp *HybridKeyPair) error {
	data, err := MarshalKeyPair(kp)
	if err != nil {
		return fmt.Errorf("signer: marshalling key %s: %w", kp.Kid, err)
	}
	path := filepath.Join(s.dir, kp.Kid+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("signer: writing key %s: %w", kp.Kid, err)
	}
	return nil
}

func (s *FileKeyStore) Load(kid string) (*HybridKeyPair, error) {
	path := filepath.Join(s.dir, kid+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("signer: reading key %s: %w", kid, err)
	}
	return UnmarshalKeyPair(data)
}

func (s *FileKeyStore) LoadAll() ([]*HybridKeyPair, error) {
	matches, err := filepath.Glob(filepath.Join(s.dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("signer: listing key files: %w", err)
	}

	// Sort to ensure consistent ordering
	sort.Strings(matches)

	var keys []*HybridKeyPair
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("signer: reading %s: %w", path, err)
		}
		kp, err := UnmarshalKeyPair(data)
		if err != nil {
			return nil, fmt.Errorf("signer: loading key from %s: %w", path, err)
		}
		keys = append(keys, kp)
	}
	return keys, nil
}

func (s *FileKeyStore) SetActiveKid(kid string) error {
	path := filepath.Join(s.dir, "active_kid.txt")
	return os.WriteFile(path, []byte(kid), 0o600)
}

func (s *FileKeyStore) GetActiveKid() (string, error) {
	path := filepath.Join(s.dir, "active_kid.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, fallback to the last loaded key if available
		if os.IsNotExist(err) {
			keys, err := s.LoadAll()
			if err == nil && len(keys) > 0 {
				kid := keys[len(keys)-1].Kid
				_ = s.SetActiveKid(kid) // persist it
				return kid, nil
			}
		}
		return "", fmt.Errorf("signer: reading active kid: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}
