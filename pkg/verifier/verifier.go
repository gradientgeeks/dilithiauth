// Package verifier provides token verification for DilithiAuth tokens,
// supporting thin JWTs, full JWTs, and CWTs with hybrid Ed25519+ML-DSA-44
// signature verification.
package verifier

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cbor "github.com/fxamacker/cbor/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
)

// VerifiedClaims contains the verified claims from a token.
type VerifiedClaims struct {
	Issuer    string    `json:"iss"`
	Subject   string    `json:"sub"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	TokenID   string    `json:"jti"`
}

// RevocationChecker allows the verifier to check if a token has been revoked.
type RevocationChecker interface {
	IsRevoked(jti string) bool
}

// Verifier handles verification of DilithiAuth tokens.
type Verifier struct {
	KeyManager *signer.KeyManager
	KeyPair    *signer.HybridKeyPair
	SigStore   *transport.SigStore
	Revocation RevocationChecker // optional; nil skips revocation checking
}

// NewVerifier creates a Verifier with a KeyManager and SigStore.
func NewVerifier(km *signer.KeyManager, store *transport.SigStore) *Verifier {
	return &Verifier{KeyManager: km, SigStore: store}
}

// NewVerifierWithKeyPair creates a Verifier with a single key pair and SigStore.
func NewVerifierWithKeyPair(kp *signer.HybridKeyPair, store *transport.SigStore) *Verifier {
	return &Verifier{KeyPair: kp, SigStore: store}
}

// VerifyJWT auto-detects thin vs full JWT and verifies accordingly.
func (v *Verifier) VerifyJWT(tokenStr string) (*VerifiedClaims, error) {
	// Peek at claims to determine token type without full verification.
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"EdDSA"}))
	token, _, err := parser.ParseUnverified(tokenStr, &oauth.HybridClaims{})
	if err != nil {
		return nil, fmt.Errorf("verifier: parsing token: %w", err)
	}

	claims, ok := token.Claims.(*oauth.HybridClaims)
	if !ok {
		return nil, errors.New("verifier: unexpected claims type")
	}

	if claims.PQCSig != "" {
		return v.VerifyJWTFull(tokenStr)
	}
	return v.VerifyJWTThin(tokenStr)
}

// VerifyJWTThin verifies a thin JWT: Ed25519 via jwt-go, then ML-DSA-44 via
// SigStore lookup using pqc_ref.
func (v *Verifier) VerifyJWTThin(tokenStr string) (*VerifiedClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &oauth.HybridClaims{},
		v.jwtKeyFunc,
		jwt.WithValidMethods([]string{"EdDSA"}),
	)
	if err != nil {
		return nil, fmt.Errorf("verifier: Ed25519 verification failed: %w", err)
	}

	claims, ok := parsed.Claims.(*oauth.HybridClaims)
	if !ok {
		return nil, errors.New("verifier: unexpected claims type")
	}

	if claims.PQCRef == "" {
		return nil, errors.New("verifier: thin token missing pqc_ref")
	}

	// Look up PQC signature from SigStore by pqc_ref.
	pqcSig, ok := v.SigStore.Get(claims.PQCRef)
	if !ok {
		return nil, errors.New("verifier: PQC signature not found in SigStore")
	}

	// Verify the pqc_ref matches the signature.
	refBytes := sha256.Sum256(pqcSig)
	computedRef := base64.RawURLEncoding.EncodeToString(refBytes[:])
	if computedRef != claims.PQCRef {
		return nil, errors.New("verifier: pqc_ref mismatch")
	}

	// Reconstruct the payload that was signed: json.Marshal(RegisteredClaims).
	payloadBytes, err := json.Marshal(claims.RegisteredClaims)
	if err != nil {
		return nil, fmt.Errorf("verifier: marshalling claims: %w", err)
	}

	// Verify ML-DSA-44.
	kp := v.keyForKid(parsed.Header["kid"].(string))
	if kp == nil {
		return nil, errors.New("verifier: key not found for kid")
	}
	hybridSig := &signer.HybridSignature{MLDSASig: pqcSig}
	if !kp.VerifyMLDSA(payloadBytes, hybridSig) {
		return nil, errors.New("verifier: ML-DSA-44 verification failed")
	}

	if v.Revocation != nil && v.Revocation.IsRevoked(claims.ID) {
		return nil, errors.New("verifier: token has been revoked")
	}

	return v.claimsFromRegistered(&claims.RegisteredClaims), nil
}

// VerifyJWTFull verifies a full/jumbo JWT with embedded pqc_sig.
func (v *Verifier) VerifyJWTFull(tokenStr string) (*VerifiedClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &oauth.HybridClaims{},
		v.jwtKeyFunc,
		jwt.WithValidMethods([]string{"EdDSA"}),
	)
	if err != nil {
		return nil, fmt.Errorf("verifier: Ed25519 verification failed: %w", err)
	}

	claims, ok := parsed.Claims.(*oauth.HybridClaims)
	if !ok {
		return nil, errors.New("verifier: unexpected claims type")
	}

	if claims.PQCSig == "" {
		return nil, errors.New("verifier: full token missing pqc_sig")
	}

	pqcSig, err := base64.RawURLEncoding.DecodeString(claims.PQCSig)
	if err != nil {
		return nil, fmt.Errorf("verifier: decoding pqc_sig: %w", err)
	}

	// Reconstruct the payload: json.Marshal(RegisteredClaims).
	payloadBytes, err := json.Marshal(claims.RegisteredClaims)
	if err != nil {
		return nil, fmt.Errorf("verifier: marshalling claims: %w", err)
	}

	kp := v.keyForKid(parsed.Header["kid"].(string))
	if kp == nil {
		return nil, errors.New("verifier: key not found for kid")
	}
	hybridSig := &signer.HybridSignature{MLDSASig: pqcSig}
	if !kp.VerifyMLDSA(payloadBytes, hybridSig) {
		return nil, errors.New("verifier: ML-DSA-44 verification failed")
	}

	if v.Revocation != nil && v.Revocation.IsRevoked(claims.ID) {
		return nil, errors.New("verifier: token has been revoked")
	}

	return v.claimsFromRegistered(&claims.RegisteredClaims), nil
}

// VerifyCWT verifies a CBOR Web Token with Ed25519 envelope + PQC sidecar.
func (v *Verifier) VerifyCWT(token []byte) (*VerifiedClaims, error) {
	// Decode envelope: [claims_bytes, ed25519_sig].
	var envelope []cbor.RawMessage
	if err := cbor.Unmarshal(token, &envelope); err != nil {
		return nil, fmt.Errorf("verifier: decoding CWT envelope: %w", err)
	}
	if len(envelope) != 2 {
		return nil, fmt.Errorf("verifier: CWT envelope has %d elements, want 2", len(envelope))
	}

	var claimsBytes []byte
	if err := cbor.Unmarshal(envelope[0], &claimsBytes); err != nil {
		return nil, fmt.Errorf("verifier: decoding claims bytes: %w", err)
	}

	var edSig []byte
	if err := cbor.Unmarshal(envelope[1], &edSig); err != nil {
		return nil, fmt.Errorf("verifier: decoding Ed25519 sig: %w", err)
	}

	// Decode claims from the envelope (these include pqc_ref).
	var claims oauth.CWTClaims
	if err := cbor.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("verifier: decoding CWT claims: %w", err)
	}

	// Reconstruct base claims WITHOUT pqc_ref — this is what was actually
	// signed by both Ed25519 and ML-DSA-44 during issuance.
	baseClaims := oauth.CWTClaims{
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		Audience:  claims.Audience,
		ExpiresAt: claims.ExpiresAt,
		IssuedAt:  claims.IssuedAt,
		JWTID:     claims.JWTID,
	}
	baseBytes, err := cbor.Marshal(baseClaims)
	if err != nil {
		return nil, fmt.Errorf("verifier: marshalling base CWT claims: %w", err)
	}

	// Verify Ed25519 over the base claims bytes (without pqc_ref).
	kp := v.activeKeyPair()
	if kp == nil {
		return nil, errors.New("verifier: no key available")
	}
	if !ed25519.Verify(kp.Ed25519Public, baseBytes, edSig) {
		return nil, errors.New("verifier: Ed25519 CWT verification failed")
	}

	// Check expiry.
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("verifier: CWT expired")
	}

	// Look up PQC signature from SigStore by JWTID.
	pqcSig, ok := v.SigStore.Get(claims.JWTID)
	if !ok {
		return nil, errors.New("verifier: PQC signature not found in SigStore for CWT")
	}

	// Verify pqc_ref.
	refBytes := sha256.Sum256(pqcSig)
	computedRef := base64.RawURLEncoding.EncodeToString(refBytes[:])
	if computedRef != claims.PQCRef {
		return nil, errors.New("verifier: CWT pqc_ref mismatch")
	}

	// Verify ML-DSA-44 over the same base claims bytes.
	hybridSig := &signer.HybridSignature{MLDSASig: pqcSig}
	if !kp.VerifyMLDSA(baseBytes, hybridSig) {
		return nil, errors.New("verifier: ML-DSA-44 CWT verification failed")
	}

	if v.Revocation != nil && v.Revocation.IsRevoked(claims.JWTID) {
		return nil, errors.New("verifier: token has been revoked")
	}

	return &VerifiedClaims{
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		IssuedAt:  time.Unix(claims.IssuedAt, 0),
		TokenID:   claims.JWTID,
	}, nil
}

func (v *Verifier) jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	kidRaw, ok := token.Header["kid"]
	if !ok {
		return nil, errors.New("verifier: missing kid in JWT header")
	}
	kid, ok := kidRaw.(string)
	if !ok {
		return nil, errors.New("verifier: kid is not a string")
	}
	kp := v.keyForKid(kid)
	if kp == nil {
		return nil, fmt.Errorf("verifier: unknown kid %q", kid)
	}
	return kp.Ed25519Public, nil
}

func (v *Verifier) keyForKid(kid string) *signer.HybridKeyPair {
	if v.KeyManager != nil {
		return v.KeyManager.KeyByKid(kid)
	}
	if v.KeyPair != nil && v.KeyPair.Kid == kid {
		return v.KeyPair
	}
	return nil
}

func (v *Verifier) activeKeyPair() *signer.HybridKeyPair {
	if v.KeyManager != nil {
		return v.KeyManager.ActiveKey()
	}
	return v.KeyPair
}

func (v *Verifier) claimsFromRegistered(rc *jwt.RegisteredClaims) *VerifiedClaims {
	vc := &VerifiedClaims{
		Issuer:  rc.Issuer,
		Subject: rc.Subject,
		TokenID: rc.ID,
	}
	if rc.ExpiresAt != nil {
		vc.ExpiresAt = rc.ExpiresAt.Time
	}
	if rc.IssuedAt != nil {
		vc.IssuedAt = rc.IssuedAt.Time
	}
	return vc
}
