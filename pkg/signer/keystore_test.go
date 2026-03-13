package signer_test

import (
	"path/filepath"
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
)

func TestMarshalUnmarshalKeyPair(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}

	data, err := signer.MarshalKeyPair(kp)
	if err != nil {
		t.Fatalf("MarshalKeyPair() error = %v", err)
	}

	restored, err := signer.UnmarshalKeyPair(data)
	if err != nil {
		t.Fatalf("UnmarshalKeyPair() error = %v", err)
	}

	if restored.Kid != kp.Kid {
		t.Errorf("Kid = %q, want %q", restored.Kid, kp.Kid)
	}

	// Verify that the restored key pair can still sign and verify.
	msg := []byte("round-trip test message")
	sig, err := restored.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() with restored key error = %v", err)
	}
	if !restored.VerifyBoth(msg, sig) {
		t.Error("VerifyBoth() with restored key = false, want true")
	}
	// Cross-verify: original key pair should verify signature from restored key.
	if !kp.VerifyBoth(msg, sig) {
		t.Error("original VerifyBoth() of restored sig = false, want true")
	}
}

func TestKeyManager_Creation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "keys")
	km, err := signer.NewKeyManager(dir)
	if err != nil {
		t.Fatalf("NewKeyManager() error = %v", err)
	}

	active := km.ActiveKey()
	if active == nil {
		t.Fatal("ActiveKey() = nil, want non-nil")
	}

	all := km.AllPublicKeys()
	if len(all) != 1 {
		t.Fatalf("AllPublicKeys() len = %d, want 1", len(all))
	}
	if all[0].Kid != active.Kid {
		t.Errorf("AllPublicKeys()[0].Kid = %q, want %q", all[0].Kid, active.Kid)
	}
}

func TestKeyManager_Rotate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "keys")
	km, err := signer.NewKeyManager(dir)
	if err != nil {
		t.Fatalf("NewKeyManager() error = %v", err)
	}

	oldKid := km.ActiveKey().Kid

	newKP, err := km.Rotate()
	if err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}

	if newKP.Kid == oldKid {
		t.Error("Rotate() returned same kid as before")
	}

	active := km.ActiveKey()
	if active.Kid != newKP.Kid {
		t.Errorf("ActiveKey().Kid = %q, want %q", active.Kid, newKP.Kid)
	}

	all := km.AllPublicKeys()
	if len(all) != 2 {
		t.Fatalf("AllPublicKeys() len = %d, want 2", len(all))
	}
}

func TestKeyManager_KeyByKid(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "keys")
	km, err := signer.NewKeyManager(dir)
	if err != nil {
		t.Fatalf("NewKeyManager() error = %v", err)
	}

	firstKid := km.ActiveKey().Kid
	km.Rotate()

	found := km.KeyByKid(firstKid)
	if found == nil {
		t.Fatalf("KeyByKid(%q) = nil", firstKid)
	}
	if found.Kid != firstKid {
		t.Errorf("KeyByKid().Kid = %q, want %q", found.Kid, firstKid)
	}

	notFound := km.KeyByKid("nonexistent")
	if notFound != nil {
		t.Error("KeyByKid(nonexistent) should return nil")
	}
}

func TestKeyManager_Persistence(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "keys")

	// First manager: generate and rotate.
	km1, err := signer.NewKeyManager(dir)
	if err != nil {
		t.Fatalf("NewKeyManager() error = %v", err)
	}
	firstKid := km1.ActiveKey().Kid
	km1.Rotate()
	secondKid := km1.ActiveKey().Kid

	// Second manager from same directory: should see both keys.
	km2, err := signer.NewKeyManager(dir)
	if err != nil {
		t.Fatalf("NewKeyManager() error = %v", err)
	}

	all := km2.AllPublicKeys()
	if len(all) != 2 {
		t.Fatalf("AllPublicKeys() len = %d, want 2", len(all))
	}

	if km2.KeyByKid(firstKid) == nil {
		t.Errorf("second manager: KeyByKid(%q) = nil", firstKid)
	}
	if km2.KeyByKid(secondKid) == nil {
		t.Errorf("second manager: KeyByKid(%q) = nil", secondKid)
	}
}
