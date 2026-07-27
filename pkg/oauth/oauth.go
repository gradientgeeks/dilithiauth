// Package oauth provides the OIDC Discovery endpoint and token-issuance
// helpers for DilithiAuth. It supports both JWT and CBOR Web Tokens (CWT)
// as well as a "Thin Token + Signature Reference" strategy to mitigate the
// token-bloat problem introduced by the large ML-DSA-44 signatures.
package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cbor "github.com/fxamacker/cbor/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
)

// DiscoveryDocument is the OpenID Connect discovery metadata published at
// /.well-known/openid-configuration (RFC 8414).
type DiscoveryDocument struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	JWKsURI                          string   `json:"jwks_uri"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	PQCHybridEnabled                 bool     `json:"pqc_hybrid_enabled"`
}

// JWKSDocument holds the public keys published at the JWKS URI.
type JWKSDocument struct {
	Keys []JWK `json:"keys"`
}

// JWK is a minimal JSON Web Key representation that embeds both the
// classical Ed25519 and post-quantum ML-DSA-44 public-key bytes.
type JWK struct {
	Kid        string `json:"kid"`
	Kty        string `json:"kty"`
	Crv        string `json:"crv,omitempty"`
	X          string `json:"x,omitempty"`           // Ed25519 public key (base64url)
	MLDSA44Pub string `json:"mldsa44_pub,omitempty"` // ML-DSA-44 public key (base64url)
	Use        string `json:"use"`
	Alg        string `json:"alg"`
}

// Provider holds the configuration and key material for a DilithiAuth IdP.
type Provider struct {
	Issuer     string
	KeyPair    *signer.HybridKeyPair
	KeyManager *signer.KeyManager
}

// NewProvider creates a Provider from an issuer URL and a pre-generated key pair.
func NewProvider(issuer string, kp *signer.HybridKeyPair) *Provider {
	return &Provider{Issuer: issuer, KeyPair: kp}
}

// NewProviderWithKeyManager creates a Provider that uses a KeyManager for
// key rotation support. The active key from the manager is used for signing.
func NewProviderWithKeyManager(issuer string, km *signer.KeyManager) *Provider {
	return &Provider{Issuer: issuer, KeyManager: km}
}

// activeKeyPair returns the key pair to use for signing. If a KeyManager is
// configured, it returns the active key; otherwise it falls back to KeyPair.
func (p *Provider) activeKeyPair() *signer.HybridKeyPair {
	if p.KeyManager != nil {
		return p.KeyManager.ActiveKey()
	}
	return p.KeyPair
}

// DiscoveryDocument returns the OpenID Connect discovery metadata.
func (p *Provider) DiscoveryDocument() *DiscoveryDocument {
	return &DiscoveryDocument{
		Issuer:                 p.Issuer,
		AuthorizationEndpoint:  p.Issuer + "/authorize",
		TokenEndpoint:          p.Issuer + "/token",
		JWKsURI:                p.Issuer + "/.well-known/jwks.json",
		ResponseTypesSupported: []string{"code"},
		SubjectTypesSupported:  []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{
			"RS256",   // legacy / interop
			"EdDSA",   // Ed25519 classical
			"MLDSA44", // post-quantum FIPS 204
		},
		PQCHybridEnabled: true,
	}
}

// JWKSDocument returns the JSON Web Key Set containing the provider's
// public keys (both Ed25519 and ML-DSA-44 under the same kid).
func (p *Provider) JWKSDocument() *JWKSDocument {
	var infos []*signer.PublicKeyInfo
	if p.KeyManager != nil {
		infos = p.KeyManager.AllPublicKeys()
	} else {
		infos = []*signer.PublicKeyInfo{p.KeyPair.PublicKeyInfo()}
	}

	keys := make([]JWK, len(infos))
	for i, info := range infos {
		keys[i] = JWK{
			Kid:        info.Kid,
			Kty:        "OKP",
			Crv:        "Ed25519",
			X:          base64.RawURLEncoding.EncodeToString(info.Ed25519Public),
			MLDSA44Pub: base64.RawURLEncoding.EncodeToString(info.MLDSAPublic),
			Use:        "sig",
			Alg:        "EdDSA+MLDSA44",
		}
	}
	return &JWKSDocument{Keys: keys}
}

// DiscoveryHandler returns an http.HandlerFunc that serves the OIDC
// discovery document at /.well-known/openid-configuration.
func (p *Provider) DiscoveryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doc := p.DiscoveryDocument()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// JWKSHandler returns an http.HandlerFunc that serves the JWKS document.
func (p *Provider) JWKSHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doc := p.JWKSDocument()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// -------------------------------------------------------------------
// JWT issuance
// -------------------------------------------------------------------

// HybridClaims extends the standard jwt.RegisteredClaims with the PQC
// reference claim used by the "Thin Token" optimisation strategy.
type HybridClaims struct {
	jwt.RegisteredClaims
	// PQCRef is a SHA-256 hash of the full ML-DSA-44 signature. High-security
	// services can use this as a cache key to fetch the full signature from a
	// sidecar store (e.g. Redis).
	PQCRef string `json:"pqc_ref,omitempty"`
	// PQCSig embeds the full ML-DSA-44 signature directly (base64url). Only
	// used in "Jumbo Token" mode. Set to empty string for "Thin Token" mode.
	PQCSig string `json:"pqc_sig,omitempty"`
	// TenantID scopes this token to the tenant it was issued under. A missing
	// value is treated as tenant.DefaultTenantID by verifiers for backward
	// compatibility with tokens issued before this claim existed.
	TenantID string `json:"tid,omitempty"`
	// Scope lists the space-delimited OAuth scopes granted to this token
	// (e.g. "openid profile email"), set from the consent grant at issuance.
	Scope string `json:"scope,omitempty"`
}

// ThinTokenResult is returned by IssueJWTThin. The full ML-DSA-44 signature
// bytes are NOT embedded in the token but returned separately so the caller
// can store them in a sidecar (e.g. Redis keyed by Claims.ID / jti).
type ThinTokenResult struct {
	// Token is the signed JWT string (contains only Ed25519 sig + pqc_ref).
	Token string
	// PQCSig is the full ML-DSA-44 signature that must be stored separately.
	PQCSig []byte
	// PQCRef is SHA-256(PQCSig) encoded as base64url, embedded in the token.
	PQCRef string
}

// IssueJWTThin creates a "Thin" JWT that contains only the Ed25519 signature
// plus a pqc_ref pointing to the ML-DSA-44 signature stored in a sidecar.
// This keeps the token small (< 500 bytes) while still providing a quantum-safe
// audit trail.
func (p *Provider) IssueJWTThin(tenantID, subject string, ttl time.Duration) (*ThinTokenResult, error) {
	return p.IssueJWTThinWithScope(tenantID, subject, "", ttl)
}

// IssueJWTThinWithScope is IssueJWTThin with an explicit granted-scope claim
// (space-delimited, e.g. "openid profile email") recorded on the token.
func (p *Provider) IssueJWTThinWithScope(tenantID, subject, scope string, ttl time.Duration) (*ThinTokenResult, error) {
	now := time.Now()
	jti, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("oauth: generating jti: %w", err)
	}

	claims := HybridClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.Issuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        jti,
		},
		TenantID: tenantID,
		Scope:    scope,
	}

	// Sign the payload with ML-DSA-44 first so we can compute the reference.
	payloadBytes, err := json.Marshal(claims.RegisteredClaims)
	if err != nil {
		return nil, fmt.Errorf("oauth: marshalling claims: %w", err)
	}

	kp := p.activeKeyPair()
	hybridSig, err := kp.Sign(payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("oauth: hybrid signing: %w", err)
	}

	// Compute the PQC reference hash.
	refBytes := sha256.Sum256(hybridSig.MLDSASig)
	pqcRef := base64.RawURLEncoding.EncodeToString(refBytes[:])

	claims.PQCRef = pqcRef

	// Sign the JWT using the Ed25519 private key via jwt-go.
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = kp.Kid
	tokenStr, err := token.SignedString(kp.Ed25519Private)
	if err != nil {
		return nil, fmt.Errorf("oauth: signing JWT: %w", err)
	}

	return &ThinTokenResult{
		Token:  tokenStr,
		PQCSig: hybridSig.MLDSASig,
		PQCRef: pqcRef,
	}, nil
}

// IssueJWTFull creates a "Jumbo" JWT that embeds the full ML-DSA-44 signature
// directly in the pqc_sig claim. The resulting token is large (~3-4 KB) but
// self-contained: no sidecar lookup is required.
func (p *Provider) IssueJWTFull(tenantID, subject string, ttl time.Duration) (string, error) {
	return p.IssueJWTFullWithScope(tenantID, subject, "", ttl)
}

// IssueJWTFullWithScope is IssueJWTFull with an explicit granted-scope claim.
func (p *Provider) IssueJWTFullWithScope(tenantID, subject, scope string, ttl time.Duration) (string, error) {
	now := time.Now()
	jti, err := generateID()
	if err != nil {
		return "", fmt.Errorf("oauth: generating jti: %w", err)
	}

	payloadBytes, err := json.Marshal(jwt.RegisteredClaims{
		Issuer:    p.Issuer,
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		ID:        jti,
	})
	if err != nil {
		return "", fmt.Errorf("oauth: marshalling claims: %w", err)
	}

	kp := p.activeKeyPair()
	hybridSig, err := kp.Sign(payloadBytes)
	if err != nil {
		return "", fmt.Errorf("oauth: hybrid signing: %w", err)
	}

	claims := HybridClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.Issuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        jti,
		},
		PQCSig:   base64.RawURLEncoding.EncodeToString(hybridSig.MLDSASig),
		TenantID: tenantID,
		Scope:    scope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = kp.Kid
	tokenStr, err := token.SignedString(kp.Ed25519Private)
	if err != nil {
		return "", fmt.Errorf("oauth: signing JWT: %w", err)
	}

	return tokenStr, nil
}

// -------------------------------------------------------------------
// CWT (CBOR Web Token) issuance – RFC 8392
// -------------------------------------------------------------------

// CWTClaims mirrors the standard CWT claim keys defined in RFC 8392.
// Using integer keys instead of string keys saves significant space in the
// binary CBOR encoding.
type CWTClaims struct {
	Issuer    string `cbor:"1,keyasint"`            // iss
	Subject   string `cbor:"2,keyasint"`            // sub
	Audience  string `cbor:"3,keyasint,omitempty"`  // aud
	ExpiresAt int64  `cbor:"4,keyasint"`            // exp
	IssuedAt  int64  `cbor:"6,keyasint"`            // iat
	JWTID     string `cbor:"7,keyasint"`            // cti
	PQCRef    string `cbor:"8,keyasint,omitempty"`  // private: pqc_ref
	TenantID  string `cbor:"9,keyasint,omitempty"`  // private: tid — see HybridClaims.TenantID
	Scope     string `cbor:"10,keyasint,omitempty"` // private: scope — see HybridClaims.Scope
}

// CWTResult contains the binary CBOR token and supporting metadata.
type CWTResult struct {
	// Token is the raw CBOR-encoded CWT byte slice. It is typically 20-30%
	// smaller than the equivalent Base64url-encoded JWT.
	Token []byte
	// PQCSig is the ML-DSA-44 signature. Store in sidecar keyed by JWTID.
	PQCSig []byte
	// PQCRef is SHA-256(PQCSig) base64url-encoded, embedded in the token.
	PQCRef string
	// JWTID is the unique token identifier (use as cache key for PQCSig).
	JWTID string
}

// IssueCWT issues a CBOR Web Token (RFC 8392) signed with Ed25519. The
// ML-DSA-44 signature is stored separately (thin-token approach) to keep the
// binary token below the infrastructure header limits.
func (p *Provider) IssueCWT(tenantID, subject string, ttl time.Duration) (*CWTResult, error) {
	return p.IssueCWTWithScope(tenantID, subject, "", ttl)
}

// IssueCWTWithScope is IssueCWT with an explicit granted-scope claim.
func (p *Provider) IssueCWTWithScope(tenantID, subject, scope string, ttl time.Duration) (*CWTResult, error) {
	now := time.Now()
	jti, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("oauth: generating jti: %w", err)
	}

	baseClaims := CWTClaims{
		Issuer:    p.Issuer,
		Subject:   subject,
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
		JWTID:     jti,
		TenantID:  tenantID,
		Scope:     scope,
	}

	// Compute ML-DSA-44 signature over the CBOR-encoded base claims.
	claimsBytes, err := cbor.Marshal(baseClaims)
	if err != nil {
		return nil, fmt.Errorf("oauth: marshalling CWT claims: %w", err)
	}

	kp := p.activeKeyPair()
	hybridSig, err := kp.Sign(claimsBytes)
	if err != nil {
		return nil, fmt.Errorf("oauth: hybrid signing CWT: %w", err)
	}

	refBytes := sha256.Sum256(hybridSig.MLDSASig)
	pqcRef := base64.RawURLEncoding.EncodeToString(refBytes[:])

	baseClaims.PQCRef = pqcRef

	// Re-marshal claims now that pqc_ref is set.
	tokenBytes, err := cbor.Marshal(baseClaims)
	if err != nil {
		return nil, fmt.Errorf("oauth: marshalling final CWT: %w", err)
	}

	// Prepend the Ed25519 signature as a CBOR-tagged COSE_Sign1-like envelope.
	// For simplicity we use a two-element array: [claims_bytes, ed25519_sig].
	edSig := hybridSig.Ed25519Sig
	envelope := []interface{}{tokenBytes, edSig}
	envelopeBytes, err := cbor.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("oauth: marshalling CWT envelope: %w", err)
	}

	return &CWTResult{
		Token:  envelopeBytes,
		PQCSig: hybridSig.MLDSASig,
		PQCRef: pqcRef,
		JWTID:  jti,
	}, nil
}

// generateID returns a 16-byte random ID encoded as base64url.
func generateID() (string, error) {
	buf := make([]byte, 16)
	if _, err := cryptoRead(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
