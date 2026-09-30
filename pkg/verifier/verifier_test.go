package verifier_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

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

// TestVerifyJWT_MLDSASignatureTamperedIsRejected isolates the ML-DSA-44
// (post-quantum) verification step from the surrounding Ed25519 + SigStore
// hash-check layers. A black-box HTTP test can't produce this case: since
// pqc_ref is itself covered by the Ed25519 signature, mutating the JWT
// externally always trips Ed25519 first. Here, with direct access to the
// signing key, we build a token whose Ed25519 signature is valid and whose
// SigStore hash check passes (the corrupted signature's own SHA-256 is used
// as pqc_ref), so the only thing that can fail is the ML-DSA-44 check
// itself — proving it actually runs and isn't short-circuited by the
// classical checks around it.
func TestVerifyJWT_MLDSASignatureTamperedIsRejected(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}
	store := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, store)

	now := time.Now()
	claims := oauth.HybridClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://auth.dilithiauth.io",
			Subject:   "alice",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			ID:        "tampered-mldsa-test",
		},
		TenantID: "default",
	}

	payloadBytes, err := json.Marshal(claims.RegisteredClaims)
	if err != nil {
		t.Fatalf("marshalling claims: %v", err)
	}

	hybridSig, err := kp.Sign(payloadBytes)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Corrupt the ML-DSA-44 signature only. Ed25519Sig is untouched, and
	// isn't part of what's re-verified below (VerifyJWT re-derives it via
	// jwt-go against the token's own bytes) — the point is that pqc_ref
	// is recomputed FROM the corrupted bytes, so the hash check the
	// verifier does before calling VerifyMLDSA still passes.
	corrupted := append([]byte(nil), hybridSig.MLDSASig...)
	corrupted[0] ^= 0xFF

	refBytes := sha256.Sum256(corrupted)
	pqcRef := base64.RawURLEncoding.EncodeToString(refBytes[:])
	claims.PQCRef = pqcRef
	store.Put(pqcRef, corrupted)

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = kp.Kid
	tokenStr, err := token.SignedString(kp.Ed25519Private)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = v.VerifyJWT(tokenStr)
	if err == nil {
		t.Fatal("VerifyJWT() should fail when the ML-DSA-44 signature is tampered, even though Ed25519 and the pqc_ref hash check both pass")
	}
	if !strings.Contains(err.Error(), "ML-DSA-44 verification failed") {
		t.Errorf("VerifyJWT() error = %q, want it to specifically report ML-DSA-44 failure (not Ed25519 or hash-check failure)", err.Error())
	}
}
