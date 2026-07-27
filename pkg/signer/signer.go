// Package signer implements a hybrid PQC+classical signature scheme using
// Cloudflare CIRCL (ML-DSA-44) and the standard crypto/ed25519 library.
// A single composite key pair combines Ed25519 (fast, classical) with
// ML-DSA-44 (NIST FIPS 204, quantum-safe) so that a single kid maps to both.
package signer

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa44"
)

// HybridKeyPair holds both classical (Ed25519) and post-quantum (ML-DSA-44)
// key material under a shared kid.
type HybridKeyPair struct {
	// Kid is the Key ID shared by both sub-keys.
	Kid string

	// Ed25519 classical key pair.
	Ed25519Public  ed25519.PublicKey
	Ed25519Private ed25519.PrivateKey

	// ML-DSA-44 post-quantum key pair (NIST FIPS 204).
	MLDSAPublic  *mldsa44.PublicKey
	MLDSAPrivate *mldsa44.PrivateKey
}

// HybridSignature holds both the classical and post-quantum signatures for a
// given message. Consumers can verify either or both depending on security
// policy.
type HybridSignature struct {
	Ed25519Sig []byte `json:"ed25519_sig"`
	MLDSASig   []byte `json:"mldsa44_sig"`
}

// PublicKeyInfo contains the serialisable public-key material that should be
// published in the JWKS endpoint.
type PublicKeyInfo struct {
	Kid           string `json:"kid"`
	Ed25519Public []byte `json:"ed25519_pub"`
	MLDSAPublic   []byte `json:"mldsa44_pub"`
}

// GenerateHybridKeyPair generates a fresh HybridKeyPair. The kid is derived
// from a base64url encoding of 8 random bytes so it is short yet unique enough
// for a single-node IdP.
func GenerateHybridKeyPair() (*HybridKeyPair, error) {
	// Generate kid
	kidBytes := make([]byte, 8)
	if _, err := rand.Read(kidBytes); err != nil {
		return nil, fmt.Errorf("signer: generating kid: %w", err)
	}
	kid := base64.RawURLEncoding.EncodeToString(kidBytes)

	// Generate Ed25519 key pair
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("signer: generating Ed25519 key: %w", err)
	}

	// Generate ML-DSA-44 key pair (NIST FIPS 204)
	mlPub, mlPriv, err := mldsa44.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("signer: generating ML-DSA-44 key: %w", err)
	}

	return &HybridKeyPair{
		Kid:            kid,
		Ed25519Public:  edPub,
		Ed25519Private: edPriv,
		MLDSAPublic:    mlPub,
		MLDSAPrivate:   mlPriv,
	}, nil
}

// Sign produces a HybridSignature over msg using both sub-keys.
// Both signatures are always produced so the resulting token is a genuine
// "dual-signed" artefact.
func (kp *HybridKeyPair) Sign(msg []byte) (*HybridSignature, error) {
	if kp == nil {
		return nil, errors.New("signer: nil key pair")
	}

	// Ed25519 signature (64 bytes)
	edSig := ed25519.Sign(kp.Ed25519Private, msg)

	// ML-DSA-44 signature (~2420 bytes)
	mlSig := make([]byte, mldsa44.SignatureSize)
	if err := mldsa44.SignTo(kp.MLDSAPrivate, msg, nil, false, mlSig); err != nil {
		return nil, fmt.Errorf("signer: ML-DSA-44 signing: %w", err)
	}

	return &HybridSignature{
		Ed25519Sig: edSig,
		MLDSASig:   mlSig,
	}, nil
}

// VerifyEd25519 verifies only the classical Ed25519 component of a
// HybridSignature. Suitable for high-throughput microservices that do not
// require quantum security for every request.
func (kp *HybridKeyPair) VerifyEd25519(msg []byte, sig *HybridSignature) bool {
	if kp == nil || sig == nil {
		return false
	}
	return ed25519.Verify(kp.Ed25519Public, msg, sig.Ed25519Sig)
}

// VerifyMLDSA verifies only the post-quantum ML-DSA-44 component.
// Suitable for "Quantum-Required" high-security services.
func (kp *HybridKeyPair) VerifyMLDSA(msg []byte, sig *HybridSignature) bool {
	if kp == nil || sig == nil {
		return false
	}
	return mldsa44.Verify(kp.MLDSAPublic, msg, nil, sig.MLDSASig)
}

// VerifyBoth verifies that BOTH the Ed25519 and ML-DSA-44 signatures are valid.
// This is the strictest verification mode.
func (kp *HybridKeyPair) VerifyBoth(msg []byte, sig *HybridSignature) bool {
	return kp.VerifyEd25519(msg, sig) && kp.VerifyMLDSA(msg, sig)
}

// PublicKeyInfo returns the serialisable public-key material.
func (kp *HybridKeyPair) PublicKeyInfo() *PublicKeyInfo {
	return &PublicKeyInfo{
		Kid:           kp.Kid,
		Ed25519Public: []byte(kp.Ed25519Public),
		MLDSAPublic:   kp.MLDSAPublic.Bytes(),
	}
}

// MarshalHybridSignature serialises a HybridSignature to JSON.
func MarshalHybridSignature(sig *HybridSignature) ([]byte, error) {
	return json.Marshal(sig)
}

// UnmarshalHybridSignature deserialises a HybridSignature from JSON.
func UnmarshalHybridSignature(data []byte) (*HybridSignature, error) {
	var sig HybridSignature
	if err := json.Unmarshal(data, &sig); err != nil {
		return nil, fmt.Errorf("signer: unmarshalling hybrid signature: %w", err)
	}
	return &sig, nil
}
