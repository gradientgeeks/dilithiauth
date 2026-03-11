package signer_test

import (
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
)

func TestGenerateHybridKeyPair(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}
	if kp.Kid == "" {
		t.Error("expected non-empty kid")
	}
	if len(kp.Ed25519Public) == 0 {
		t.Error("expected non-empty Ed25519 public key")
	}
	if kp.MLDSAPublic == nil {
		t.Error("expected non-nil ML-DSA-44 public key")
	}
}

func TestHybridSign_VerifyBoth(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}

	msg := []byte("DilithiAuth test message")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if !kp.VerifyBoth(msg, sig) {
		t.Error("VerifyBoth() = false, want true")
	}
}

func TestHybridSign_VerifyEd25519(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}

	msg := []byte("DilithiAuth classical verification test")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if !kp.VerifyEd25519(msg, sig) {
		t.Error("VerifyEd25519() = false, want true")
	}
}

func TestHybridSign_VerifyMLDSA(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}

	msg := []byte("DilithiAuth quantum verification test")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if !kp.VerifyMLDSA(msg, sig) {
		t.Error("VerifyMLDSA() = false, want true")
	}
}

func TestVerify_TamperedMessage(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}

	msg := []byte("original message")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	tampered := []byte("tampered message")
	if kp.VerifyEd25519(tampered, sig) {
		t.Error("VerifyEd25519() = true on tampered message, want false")
	}
	if kp.VerifyMLDSA(tampered, sig) {
		t.Error("VerifyMLDSA() = true on tampered message, want false")
	}
}

func TestMarshalUnmarshalHybridSignature(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}

	msg := []byte("marshal/unmarshal test")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	data, err := signer.MarshalHybridSignature(sig)
	if err != nil {
		t.Fatalf("MarshalHybridSignature() error = %v", err)
	}

	restored, err := signer.UnmarshalHybridSignature(data)
	if err != nil {
		t.Fatalf("UnmarshalHybridSignature() error = %v", err)
	}

	if !kp.VerifyBoth(msg, restored) {
		t.Error("VerifyBoth() after round-trip = false, want true")
	}
}

func TestPublicKeyInfo(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}

	info := kp.PublicKeyInfo()
	if info.Kid != kp.Kid {
		t.Errorf("PublicKeyInfo.Kid = %q, want %q", info.Kid, kp.Kid)
	}
	if len(info.Ed25519Public) != 32 {
		t.Errorf("Ed25519Public size = %d, want 32", len(info.Ed25519Public))
	}
	if len(info.MLDSAPublic) == 0 {
		t.Error("MLDSAPublic is empty")
	}
}
