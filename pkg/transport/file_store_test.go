package transport_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
)

func TestFileSigStore_PutGetDeleteLen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sigstore.json")
	fs, err := transport.NewFileSigStore(path)
	if err != nil {
		t.Fatalf("NewFileSigStore() error = %v", err)
	}

	if fs.Len() != 0 {
		t.Fatalf("initial Len = %d, want 0", fs.Len())
	}

	fs.Put("jti-1", []byte("sig-1"))
	fs.Put("jti-2", []byte("sig-2"))

	if fs.Len() != 2 {
		t.Errorf("Len = %d, want 2", fs.Len())
	}

	got, ok := fs.Get("jti-1")
	if !ok {
		t.Fatal("Get(jti-1) = false, want true")
	}
	if string(got) != "sig-1" {
		t.Errorf("Get(jti-1) = %q, want %q", got, "sig-1")
	}

	fs.Delete("jti-1")
	if _, ok := fs.Get("jti-1"); ok {
		t.Error("Get(jti-1) after Delete = true, want false")
	}
	if fs.Len() != 1 {
		t.Errorf("Len after Delete = %d, want 1", fs.Len())
	}
}

func TestFileSigStore_Persistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sigstore.json")

	// First instance: write data.
	fs1, err := transport.NewFileSigStore(path)
	if err != nil {
		t.Fatalf("NewFileSigStore() error = %v", err)
	}
	fs1.Put("jti-persist", []byte("persistent-sig"))

	// Second instance: should see the data.
	fs2, err := transport.NewFileSigStore(path)
	if err != nil {
		t.Fatalf("NewFileSigStore() error = %v", err)
	}
	got, ok := fs2.Get("jti-persist")
	if !ok {
		t.Fatal("second instance: Get(jti-persist) = false, want true")
	}
	if string(got) != "persistent-sig" {
		t.Errorf("second instance: Get(jti-persist) = %q, want %q", got, "persistent-sig")
	}
}

func TestFileSigStore_EmptyFileBootstrap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sigstore.json")

	// File does not exist yet — should bootstrap cleanly.
	fs, err := transport.NewFileSigStore(path)
	if err != nil {
		t.Fatalf("NewFileSigStore() error = %v", err)
	}
	if fs.Len() != 0 {
		t.Errorf("Len = %d, want 0", fs.Len())
	}

	// File should not exist until first write.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to not exist before first write")
	}

	fs.Put("jti-boot", []byte("boot-sig"))

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist after first write: %v", err)
	}
}

func TestFileSigStore_WithTTL_ExpiresAndPurges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sigstore.json")
	fs, err := transport.NewFileSigStoreWithTTL(path, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("NewFileSigStoreWithTTL() error = %v", err)
	}
	fs.Put("jti-1", []byte("sig-1"))

	if _, ok := fs.Get("jti-1"); !ok {
		t.Fatal("entry should be readable immediately after Put")
	}

	time.Sleep(20 * time.Millisecond)
	if _, ok := fs.Get("jti-1"); ok {
		t.Error("expired entry should no longer be returned by Get, even before Purge runs")
	}
	if fs.Len() != 1 {
		t.Fatalf("Len() = %d, want 1 before Purge", fs.Len())
	}

	fs.Purge()
	if fs.Len() != 0 {
		t.Errorf("Len() = %d, want 0 after Purge", fs.Len())
	}

	// Purge's flush must also be reflected on disk.
	fs2, err := transport.NewFileSigStoreWithTTL(path, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("NewFileSigStoreWithTTL() error = %v", err)
	}
	if fs2.Len() != 0 {
		t.Errorf("reloaded Len() = %d, want 0 (Purge should have flushed to disk)", fs2.Len())
	}
}

func TestNewSigStoreWithBackend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sigstore.json")
	fb, err := transport.NewFileSigStore(path)
	if err != nil {
		t.Fatalf("NewFileSigStore() error = %v", err)
	}

	store := transport.NewSigStoreWithBackend(fb)
	store.Put("jti-backend", []byte("backend-sig"))

	got, ok := store.Get("jti-backend")
	if !ok {
		t.Fatal("Get(jti-backend) = false, want true")
	}
	if string(got) != "backend-sig" {
		t.Errorf("Get(jti-backend) = %q, want %q", got, "backend-sig")
	}

	// Verify persistence through the file backend.
	fs2, err := transport.NewFileSigStore(path)
	if err != nil {
		t.Fatalf("NewFileSigStore() error = %v", err)
	}
	got2, ok := fs2.Get("jti-backend")
	if !ok {
		t.Fatal("persistence: Get(jti-backend) = false, want true")
	}
	if string(got2) != "backend-sig" {
		t.Errorf("persistence: Get(jti-backend) = %q, want %q", got2, "backend-sig")
	}
}
