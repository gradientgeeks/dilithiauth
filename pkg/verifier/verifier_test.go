package verifier_test

import (
	"testing"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

func setup(t *testing.T) (*oauth.Provider, *verifier.Verifier, *transport.SigStore) {
	t.Helper()
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}
	store := transport.NewSigStore()
	p := oauth.NewProvider("https://auth.dilithiauth.io", kp)
	v := verifier.NewVerifierWithKeyPair(kp, store)
	return p, v, store
}

func TestVerifyJWTThin_HappyPath(t *testing.T) {
	p, v, store := setup(t)

	result, err := p.IssueJWTThin("default", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	claims, err := v.VerifyJWTThin(result.Token)
	if err != nil {
		t.Fatalf("VerifyJWTThin() error = %v", err)
	}
	if claims.Subject != "alice" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "alice")
	}
}

func TestVerifyJWTThin_MissingSigStore(t *testing.T) {
	p, v, _ := setup(t)

	result, err := p.IssueJWTThin("default", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	// Intentionally NOT storing the PQC signature.

	_, err = v.VerifyJWTThin(result.Token)
	if err == nil {
		t.Fatal("VerifyJWTThin() should fail when PQC sig not in store")
	}
}

func TestVerifyJWTFull_HappyPath(t *testing.T) {
	p, v, _ := setup(t)

	tokenStr, err := p.IssueJWTFull("default", "bob", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTFull() error = %v", err)
	}

	claims, err := v.VerifyJWTFull(tokenStr)
	if err != nil {
		t.Fatalf("VerifyJWTFull() error = %v", err)
	}
	if claims.Subject != "bob" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "bob")
	}
}

func TestVerifyCWT_HappyPath(t *testing.T) {
	p, v, store := setup(t)

	result, err := p.IssueCWT("default", "charlie", time.Hour)
	if err != nil {
		t.Fatalf("IssueCWT() error = %v", err)
	}
	store.Put(result.JWTID, result.PQCSig)

	claims, err := v.VerifyCWT(result.Token)
	if err != nil {
		t.Fatalf("VerifyCWT() error = %v", err)
	}
	if claims.Subject != "charlie" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "charlie")
	}
}

func TestVerifyJWT_AutoDetection(t *testing.T) {
	p, v, store := setup(t)

	// Thin token → auto-detect.
	thinResult, err := p.IssueJWTThin("default", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(thinResult.PQCRef, thinResult.PQCSig)

	claims, err := v.VerifyJWT(thinResult.Token)
	if err != nil {
		t.Fatalf("VerifyJWT(thin) error = %v", err)
	}
	if claims.Subject != "alice" {
		t.Errorf("thin: Subject = %q, want %q", claims.Subject, "alice")
	}

	// Full token → auto-detect.
	fullToken, err := p.IssueJWTFull("default", "bob", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTFull() error = %v", err)
	}

	claims, err = v.VerifyJWT(fullToken)
	if err != nil {
		t.Fatalf("VerifyJWT(full) error = %v", err)
	}
	if claims.Subject != "bob" {
		t.Errorf("full: Subject = %q, want %q", claims.Subject, "bob")
	}
}

func TestVerifyJWT_WrongKey(t *testing.T) {
	p, _, store := setup(t)

	result, err := p.IssueJWTThin("default", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	// Create verifier with a different key pair.
	otherKP, _ := signer.GenerateHybridKeyPair()
	wrongV := verifier.NewVerifierWithKeyPair(otherKP, store)

	_, err = wrongV.VerifyJWT(result.Token)
	if err == nil {
		t.Fatal("VerifyJWT() should fail with wrong key")
	}
}

func TestVerifyJWT_ExpiredToken(t *testing.T) {
	p, v, store := setup(t)

	// Issue token with 0 TTL (already expired).
	result, err := p.IssueJWTThin("default", "alice", -time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	_, err = v.VerifyJWT(result.Token)
	if err == nil {
		t.Fatal("VerifyJWT() should fail for expired token")
	}
}

func TestVerifyJWT_IssuerMismatch(t *testing.T) {
	p, v, store := setup(t)
	v.Issuer = "https://wrong-issuer.example.com" // provider issues as https://auth.dilithiauth.io

	result, err := p.IssueJWTThin("default", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	if _, err := v.VerifyJWT(result.Token); err == nil {
		t.Error("VerifyJWT() should fail when the token's issuer doesn't match Verifier.Issuer")
	}
}

func TestVerifyJWT_IssuerMatch(t *testing.T) {
	p, v, store := setup(t)
	v.Issuer = "https://auth.dilithiauth.io" // matches setup()'s provider issuer

	result, err := p.IssueJWTThin("default", "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	if _, err := v.VerifyJWT(result.Token); err != nil {
		t.Errorf("VerifyJWT() error = %v, want success for a matching issuer", err)
	}
}
