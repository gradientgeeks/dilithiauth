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
	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
)

// VerifiedClaims contains the verified claims from a token.
type VerifiedClaims struct {
	Issuer    string    `json:"iss"`
	Subject   string    `json:"sub"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	TokenID   string    `json:"jti"`
	// TenantID is the tenant this token was issued under (oauth.HybridClaims.TenantID
	// / oauth.CWTClaims.TenantID). Tokens issued before this claim existed have
	// it defaulted to tenant.DefaultTenantID by the issuing/verifying code.
	TenantID string `json:"tid,omitempty"`
	// Scope lists the space-delimited OAuth scopes granted to this token.
	Scope string `json:"scope,omitempty"`
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
	// Issuer, when set, is required to exactly match a token's iss claim.
	// Left empty, issuer validation is skipped (e.g. in tests that don't
	// care about it) — production wiring in cmd/server/main.go sets this.
	Issuer string
}

// NewVerifier creates a Verifier with a KeyManager and SigStore.
func NewVerifier(km *signer.KeyManager, store *transport.SigStore) *Verifier {
	return &Verifier{KeyManager: km, SigStore: store}
}

// NewVerifierWithKeyPair creates a Verifier with a single key pair and SigStore.
func NewVerifierWithKeyPair(kp *signer.HybridKeyPair, store *transport.SigStore) *Verifier {
	return &Verifier{KeyPair: kp, SigStore: store}
}

// VerifyJWT auto-detects thin vs full JWT and verifies both Ed25519 and ML-DSA-44
// in a single parse pass to eliminate double-parse overhead.
func (v *Verifier) VerifyJWT(tokenStr string) (*VerifiedClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &oauth.HybridClaims{},
		v.jwtKeyFunc,
		v.parseOptions()...,
	)
	if err != nil {
		return nil, fmt.Errorf("verifier: Ed25519 verification failed: %w", err)
	}

	claims, ok := parsed.Claims.(*oauth.HybridClaims)
	if !ok {
		return nil, errors.New("verifier: unexpected claims type")
	}

	kid, ok := parsed.Header["kid"].(string)
	if !ok || kid == "" {
		return nil, errors.New("verifier: missing or invalid kid in header")
	}

	kp := v.keyForKid(kid)
	if kp == nil {
		return nil, fmt.Errorf("verifier: key not found for kid %q", kid)
	}

	payloadBytes, err := json.Marshal(claims.RegisteredClaims)
	if err != nil {
		return nil, fmt.Errorf("verifier: marshalling claims: %w", err)
	}

	var pqcSig []byte
	if claims.PQCSig != "" {
		// Full Jumbo JWT
		pqcSig, err = base64.RawURLEncoding.DecodeString(claims.PQCSig)
		if err != nil {
			return nil, fmt.Errorf("verifier: decoding pqc_sig: %w", err)
		}
	} else if claims.PQCRef != "" {
		// Thin JWT via SigStore lookup
		if v.SigStore == nil {
			return nil, errors.New("verifier: SigStore is required to verify thin JWTs")
		}
		sig, found := v.SigStore.Get(claims.PQCRef)
		if !found {
			return nil, errors.New("verifier: PQC signature not found in SigStore")
		}
		refBytes := sha256.Sum256(sig)
		if base64.RawURLEncoding.EncodeToString(refBytes[:]) != claims.PQCRef {
			return nil, errors.New("verifier: pqc_ref mismatch")
		}
		pqcSig = sig
	} else {
		return nil, errors.New("verifier: missing PQC signature or reference")
	}

	hybridSig := &signer.HybridSignature{MLDSASig: pqcSig}
	if !kp.VerifyMLDSA(payloadBytes, hybridSig) {
		return nil, errors.New("verifier: ML-DSA-44 verification failed")
	}

	if v.Revocation != nil && v.Revocation.IsRevoked(claims.ID) {
		return nil, errors.New("verifier: token has been revoked")
	}

	return v.claimsFromHybrid(claims), nil
}

// VerifyFastPath performs lightweight Ed25519-only verification for zero-latency
// requirements (skips ML-DSA-44 PQC signature verification).
func (v *Verifier) VerifyFastPath(tokenStr string) (*VerifiedClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &oauth.HybridClaims{},
		v.jwtKeyFunc,
		v.parseOptions()...,
	)
	if err != nil {
		return nil, fmt.Errorf("verifier: Ed25519 verification failed: %w", err)
	}

	claims, ok := parsed.Claims.(*oauth.HybridClaims)
	if !ok {
		return nil, errors.New("verifier: unexpected claims type")
	}

	if v.Revocation != nil && v.Revocation.IsRevoked(claims.ID) {
		return nil, errors.New("verifier: token has been revoked")
	}

	return v.claimsFromHybrid(claims), nil
}

// VerifyJWTAsync verifies Ed25519 synchronously for zero response latency,
// then offloads ML-DSA-44 verification to a background goroutine.
func (v *Verifier) VerifyJWTAsync(tokenStr string, onPQCError func(err error)) (*VerifiedClaims, error) {
	claims, err := v.VerifyFastPath(tokenStr)
	if err != nil {
		return nil, err
	}

	go func() {
		_, fullErr := v.VerifyJWT(tokenStr)
		if fullErr != nil && onPQCError != nil {
			onPQCError(fullErr)
		}
	}()

	return claims, nil
}

// VerifyJWTThin verifies a thin JWT (kept for backward compatibility).
func (v *Verifier) VerifyJWTThin(tokenStr string) (*VerifiedClaims, error) {
	return v.VerifyJWT(tokenStr)
}

// VerifyJWTFull verifies a full/jumbo JWT (kept for backward compatibility).
func (v *Verifier) VerifyJWTFull(tokenStr string) (*VerifiedClaims, error) {
	return v.VerifyJWT(tokenStr)
}

// VerifyCWT verifies a CBOR Web Token with Ed25519 envelope + PQC sidecar.
func (v *Verifier) VerifyCWT(token []byte) (*VerifiedClaims, error) {
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

	var claims oauth.CWTClaims
	if err := cbor.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("verifier: decoding CWT claims: %w", err)
	}

	baseClaims := oauth.CWTClaims{
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		Audience:  claims.Audience,
		ExpiresAt: claims.ExpiresAt,
		IssuedAt:  claims.IssuedAt,
		JWTID:     claims.JWTID,
		TenantID:  claims.TenantID,
		Scope:     claims.Scope,
	}
	baseBytes, err := cbor.Marshal(baseClaims)
	if err != nil {
		return nil, fmt.Errorf("verifier: marshalling base CWT claims: %w", err)
	}

	kp := v.activeKeyPair()
	if kp == nil {
		return nil, errors.New("verifier: no key available")
	}
	if !ed25519.Verify(kp.Ed25519Public, baseBytes, edSig) {
		return nil, errors.New("verifier: Ed25519 CWT verification failed")
	}

	if v.Issuer != "" && claims.Issuer != v.Issuer {
		return nil, errors.New("verifier: CWT issuer mismatch")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("verifier: CWT expired")
	}

	if v.SigStore == nil {
		return nil, errors.New("verifier: SigStore required for CWT PQC verification")
	}

	pqcSig, ok := v.SigStore.Get(claims.JWTID)
	if !ok {
		return nil, errors.New("verifier: PQC signature not found in SigStore for CWT")
	}

	refBytes := sha256.Sum256(pqcSig)
	computedRef := base64.RawURLEncoding.EncodeToString(refBytes[:])
	if computedRef != claims.PQCRef {
		return nil, errors.New("verifier: CWT pqc_ref mismatch")
	}

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
		TenantID:  normalizeTenantID(claims.TenantID),
		Scope:     claims.Scope,
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

// parseOptions returns the jwt.ParserOption set applied to every ParseWithClaims
// call, restricting accepted algorithms and, when v.Issuer is configured,
// requiring an exact issuer match.
func (v *Verifier) parseOptions() []jwt.ParserOption {
	opts := []jwt.ParserOption{jwt.WithValidMethods([]string{"EdDSA"})}
	if v.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(v.Issuer))
	}
	return opts
}

func (v *Verifier) claimsFromHybrid(claims *oauth.HybridClaims) *VerifiedClaims {
	rc := &claims.RegisteredClaims
	vc := &VerifiedClaims{
		Issuer:   rc.Issuer,
		Subject:  rc.Subject,
		TokenID:  rc.ID,
		TenantID: normalizeTenantID(claims.TenantID),
		Scope:    claims.Scope,
	}
	if rc.ExpiresAt != nil {
		vc.ExpiresAt = rc.ExpiresAt.Time
	}
	if rc.IssuedAt != nil {
		vc.IssuedAt = rc.IssuedAt.Time
	}
	return vc
}

// normalizeTenantID treats a missing tenant claim (tokens issued before it
// existed) as tenant.DefaultTenantID, matching how tokens are issued today
// for single-tenant/default deployments.
func normalizeTenantID(tenantID string) string {
	if tenantID == "" {
		return tenant.DefaultTenantID
	}
	return tenantID
}
