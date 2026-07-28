package signer_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/cloudflare/circl/sign/mldsa/mldsa44"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
)

func TestParametersVerification(t *testing.T) {
	// Verify exact cryptographic parameters for ML-DSA-44 and Ed25519
	if mldsa44.PublicKeySize != 1312 {
		t.Errorf("ML-DSA-44 Public Key size = %d, want 1312", mldsa44.PublicKeySize)
	}
	if mldsa44.PrivateKeySize != 2560 {
		t.Errorf("ML-DSA-44 Private Key size = %d, want 2560", mldsa44.PrivateKeySize)
	}
	if mldsa44.SignatureSize != 2420 {
		t.Errorf("ML-DSA-44 Signature size = %d, want 2420", mldsa44.SignatureSize)
	}
	if ed25519.PublicKeySize != 32 {
		t.Errorf("Ed25519 Public Key size = %d, want 32", ed25519.PublicKeySize)
	}
	if ed25519.PrivateKeySize != 64 {
		t.Errorf("Ed25519 Private Key size = %d, want 64", ed25519.PrivateKeySize)
	}
	if ed25519.SignatureSize != 64 {
		t.Errorf("Ed25519 Signature size = %d, want 64", ed25519.SignatureSize)
	}
}

func BenchmarkSign_Ed25519(b *testing.B) {
	_, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	msg := []byte("Benchmark message payload for DilithiAuth Ed25519 signing")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ed25519.Sign(edPriv, msg)
	}
}

func BenchmarkSign_MLDSA44(b *testing.B) {
	_, mlPriv, err := mldsa44.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	msg := []byte("Benchmark message payload for DilithiAuth ML-DSA-44 signing")
	sig := make([]byte, mldsa44.SignatureSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mldsa44.SignTo(mlPriv, msg, nil, false, sig)
	}
}

func BenchmarkSign_Hybrid(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key pair: %v", err)
	}
	msg := []byte("Benchmark message payload for DilithiAuth Hybrid signing")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := kp.Sign(msg)
		if err != nil {
			b.Fatalf("Sign error: %v", err)
		}
	}
}

func BenchmarkVerify_Ed25519(b *testing.B) {
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	msg := []byte("Benchmark message payload for DilithiAuth Ed25519 verification")
	sig := ed25519.Sign(edPriv, msg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ed25519.Verify(edPub, msg, sig)
	}
}

func BenchmarkVerify_MLDSA44(b *testing.B) {
	mlPub, mlPriv, err := mldsa44.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	msg := []byte("Benchmark message payload for DilithiAuth ML-DSA-44 verification")
	sig := make([]byte, mldsa44.SignatureSize)
	if err := mldsa44.SignTo(mlPriv, msg, nil, false, sig); err != nil {
		b.Fatalf("failed to sign: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mldsa44.Verify(mlPub, msg, nil, sig)
	}
}

func BenchmarkVerify_Hybrid(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key pair: %v", err)
	}
	msg := []byte("Benchmark message payload for DilithiAuth Hybrid verification")
	sig, err := kp.Sign(msg)
	if err != nil {
		b.Fatalf("failed to sign: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !kp.VerifyBoth(msg, sig) {
			b.Fatal("verify failed")
		}
	}
}
