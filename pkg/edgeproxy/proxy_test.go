package edgeproxy_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/edgeproxy"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

func TestEdgeProxy_OffloadPQC(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	provider := oauth.NewProvider("https://dilithiauth.gradientgeeks.tech", kp)
	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	internalSecret := []byte("secret_internal_key_for_microservices_12345")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("upstream received invalid Authorization header: %s", authHeader)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return internalSecret, nil
		})
		if err != nil || !token.Valid {
			t.Errorf("upstream failed to verify internal token: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		sub, _ := token.Claims.GetSubject()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","user":"` + sub + `"}`))
	}))
	defer upstream.Close()

	proxy, err := edgeproxy.NewEdgeProxy(v, internalSecret, upstream.URL)
	if err != nil {
		t.Fatalf("failed to create edge proxy: %v", err)
	}

	proxyServer := httptest.NewServer(proxy.Handler())
	defer proxyServer.Close()

	jumboToken, err := provider.IssueJWTFull("default", "user_oracle_cloud_123", time.Hour)
	if err != nil {
		t.Fatalf("failed to issue jumbo JWT: %v", err)
	}

	req, err := http.NewRequest("GET", proxyServer.URL+"/api/resource", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+jumboToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send request to proxy: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from proxy, got %d", resp.StatusCode)
	}
}

func TestEdgeProxy_AsymmetricEd25519ReSigning(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	provider := oauth.NewProvider("https://dilithiauth.gradientgeeks.tech", kp)
	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	// Generate asymmetric Ed25519 key for zero-trust proxy re-signing
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate Ed25519 key: %v", err)
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Downstream microservice verifies using Ed25519 public key!
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return pubKey, nil
		})
		if err != nil || !token.Valid {
			t.Errorf("upstream failed to verify Ed25519 token: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		sub, _ := token.Claims.GetSubject()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","user":"` + sub + `"}`))
	}))
	defer upstream.Close()

	proxy, err := edgeproxy.NewEdgeProxyAsymmetric(v, privKey, upstream.URL)
	if err != nil {
		t.Fatalf("failed to create asymmetric edge proxy: %v", err)
	}

	proxyServer := httptest.NewServer(proxy.Handler())
	defer proxyServer.Close()

	jumboToken, err := provider.IssueJWTFull("default", "user_zero_trust_456", time.Hour)
	if err != nil {
		t.Fatalf("failed to issue jumbo JWT: %v", err)
	}

	req, err := http.NewRequest("GET", proxyServer.URL+"/api/resource", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+jumboToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send request to proxy: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from proxy, got %d", resp.StatusCode)
	}
}
