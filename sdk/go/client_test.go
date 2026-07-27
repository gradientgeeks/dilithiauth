package dilithiauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGeneratePKCEPair(t *testing.T) {
	verifier, challenge, err := GeneratePKCEPair()
	if err != nil {
		t.Fatalf("GeneratePKCEPair returned error: %v", err)
	}

	if len(verifier) == 0 {
		t.Error("verifier is empty")
	}
	if len(challenge) == 0 {
		t.Error("challenge is empty")
	}
}

func TestGetAuthorizationURL(t *testing.T) {
	client, err := NewClient(Config{
		Domain:   "https://auth.example.com",
		ClientID: "test-client-id",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	urlStr := client.GetAuthorizationURL("https://myapp.com/callback", "state123", "challenge123")
	if urlStr == "" {
		t.Error("authorization URL is empty")
	}
}

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/oauth/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "dummy_token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{
		Domain:   server.URL,
		ClientID: "test-client-id",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	resp, err := client.ExchangeCode(context.Background(), "auth_code", "https://myapp.com/callback", "verifier123")
	if err != nil {
		t.Fatalf("ExchangeCode error: %v", err)
	}

	if resp.AccessToken != "dummy_token" {
		t.Errorf("expected access_token 'dummy_token', got '%s'", resp.AccessToken)
	}
}

func TestParseClaims(t *testing.T) {
	headerBytes, _ := json.Marshal(map[string]string{"alg": "Ed25519+ML-DSA-44", "typ": "JWT"})
	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"iss":     "https://auth.example.com",
		"sub":     "alice@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"pqc_ref": "sha256:abc123",
	})

	tokenStr := base64.RawURLEncoding.EncodeToString(headerBytes) + "." +
		base64.RawURLEncoding.EncodeToString(payloadBytes) + "." +
		"dummy_signature"

	claims, err := ParseClaims(tokenStr)
	if err != nil {
		t.Fatalf("ParseClaims error: %v", err)
	}

	if claims.Subject != "alice@example.com" {
		t.Errorf("expected subject 'alice@example.com', got '%s'", claims.Subject)
	}
	if claims.PQCRef != "sha256:abc123" {
		t.Errorf("expected pqc_ref 'sha256:abc123', got '%s'", claims.PQCRef)
	}
}
