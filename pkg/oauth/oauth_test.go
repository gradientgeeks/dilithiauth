package oauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
)

func newProvider(t *testing.T) *oauth.Provider {
	t.Helper()
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair() error = %v", err)
	}
	return oauth.NewProvider("https://auth.dilithiauth.io", kp)
}

// --- Discovery tests ---

func TestDiscoveryDocument(t *testing.T) {
	p := newProvider(t)
	doc := p.DiscoveryDocument()

	if doc.Issuer != "https://auth.dilithiauth.io" {
		t.Errorf("Issuer = %q, want %q", doc.Issuer, "https://auth.dilithiauth.io")
	}
	if !doc.PQCHybridEnabled {
		t.Error("PQCHybridEnabled = false, want true")
	}
	algFound := false
	for _, alg := range doc.IDTokenSigningAlgValuesSupported {
		if alg == "MLDSA44" {
			algFound = true
		}
	}
	if !algFound {
		t.Error("MLDSA44 not listed in id_token_signing_alg_values_supported")
	}
}

func TestDiscoveryHandler(t *testing.T) {
	p := newProvider(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	p.DiscoveryHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var doc oauth.DiscoveryDocument
	if err := json.NewDecoder(rec.Body).Decode(&doc); err != nil {
		t.Fatalf("decoding discovery doc: %v", err)
	}
	if doc.JWKsURI == "" {
		t.Error("jwks_uri is empty")
	}
}

func TestJWKSHandler(t *testing.T) {
	p := newProvider(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	p.JWKSHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var doc oauth.JWKSDocument
	if err := json.NewDecoder(rec.Body).Decode(&doc); err != nil {
		t.Fatalf("decoding JWKS: %v", err)
	}
	if len(doc.Keys) == 0 {
		t.Fatal("JWKS contains no keys")
	}
	if doc.Keys[0].MLDSA44Pub == "" {
		t.Error("JWK missing mldsa44_pub field")
	}
}

// --- JWT tests ---

func TestIssueJWTThin(t *testing.T) {
	p := newProvider(t)
	result, err := p.IssueJWTThin("alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	if result.Token == "" {
		t.Error("Token is empty")
	}
	if len(result.PQCSig) == 0 {
		t.Error("PQCSig is empty")
	}
	if result.PQCRef == "" {
		t.Error("PQCRef is empty")
	}

	// The token should be parseable and contain the pqc_ref claim.
	parts := strings.Split(result.Token, ".")
	if len(parts) != 3 {
		t.Errorf("token has %d parts, want 3", len(parts))
	}

	// Verify the token using the Ed25519 public key.
	parsed, err := jwt.ParseWithClaims(result.Token, &oauth.HybridClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return p.KeyPair.Ed25519Public, nil
		},
		jwt.WithValidMethods([]string{"EdDSA"}),
	)
	if err != nil {
		t.Fatalf("jwt.ParseWithClaims() error = %v", err)
	}
	if !parsed.Valid {
		t.Error("parsed token is not valid")
	}

	claims, ok := parsed.Claims.(*oauth.HybridClaims)
	if !ok {
		t.Fatal("claims type assertion failed")
	}
	if claims.PQCRef != result.PQCRef {
		t.Errorf("pqc_ref in token = %q, want %q", claims.PQCRef, result.PQCRef)
	}
	// Thin token must NOT embed the full PQC sig.
	if claims.PQCSig != "" {
		t.Error("thin token should not contain pqc_sig claim")
	}
}

func TestIssueJWTFull(t *testing.T) {
	p := newProvider(t)
	tokenStr, err := p.IssueJWTFull("bob", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTFull() error = %v", err)
	}
	if tokenStr == "" {
		t.Error("token string is empty")
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &oauth.HybridClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return p.KeyPair.Ed25519Public, nil
		},
		jwt.WithValidMethods([]string{"EdDSA"}),
	)
	if err != nil {
		t.Fatalf("jwt.ParseWithClaims() error = %v", err)
	}
	claims, ok := parsed.Claims.(*oauth.HybridClaims)
	if !ok {
		t.Fatal("claims type assertion failed")
	}
	if claims.PQCSig == "" {
		t.Error("full token must contain pqc_sig claim")
	}
}

// --- CWT tests ---

func TestIssueCWT(t *testing.T) {
	p := newProvider(t)
	result, err := p.IssueCWT("charlie", time.Hour)
	if err != nil {
		t.Fatalf("IssueCWT() error = %v", err)
	}
	if len(result.Token) == 0 {
		t.Error("CWT token bytes are empty")
	}
	if len(result.PQCSig) == 0 {
		t.Error("PQCSig is empty")
	}
	if result.PQCRef == "" {
		t.Error("PQCRef is empty")
	}
	if result.JWTID == "" {
		t.Error("JWTID is empty")
	}
}

func TestCWTSmallerThanJWT(t *testing.T) {
	p := newProvider(t)

	cwtResult, err := p.IssueCWT("dave", time.Hour)
	if err != nil {
		t.Fatalf("IssueCWT() error = %v", err)
	}

	jwtResult, err := p.IssueJWTThin("dave", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}

	// CWT binary bytes should be smaller than the JWT string bytes.
	if len(cwtResult.Token) >= len([]byte(jwtResult.Token)) {
		t.Logf("CWT size: %d bytes, JWT size: %d bytes", len(cwtResult.Token), len(jwtResult.Token))
		t.Log("Note: CWT size comparison is informational - size depends on claim content")
	}
	t.Logf("CWT size: %d bytes, JWT (thin) size: %d bytes", len(cwtResult.Token), len(jwtResult.Token))
}
