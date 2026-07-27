package bff_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/bff"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

func TestBFFServer_LoginAndSessionFlow(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	cfg := bff.Config{
		IssuerURL:     "https://dilithiauthapis.gradientgeeks.tech",
		ClientID:      "da_developer_portal",
		RedirectURI:   "https://dilithiauth.gradientgeeks.tech/auth/callback",
		FrontendURL:   "https://dilithiauth.gradientgeeks.tech",
		SessionSecret: []byte("01234567890123456789012345678901"),
	}

	server, err := bff.NewServer(cfg, v)
	if err != nil {
		t.Fatalf("failed to create BFF server: %v", err)
	}

	handler := server.Router()

	t.Run("HandleLogin Initiates PKCE Redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/login", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusFound {
			t.Fatalf("expected 302 Found, got %d", rec.Code)
		}

		location := rec.Header().Get("Location")
		if !strings.HasPrefix(location, "https://dilithiauthapis.gradientgeeks.tech/authorize") {
			t.Errorf("unexpected location: %s", location)
		}
		if !strings.Contains(location, "code_challenge=") || !strings.Contains(location, "code_challenge_method=S256") {
			t.Errorf("location missing PKCE params: %s", location)
		}

		cookies := rec.Result().Cookies()
		var foundPKCE bool
		for _, c := range cookies {
			if c.Name == "dilithiauth_bff_pkce" {
				foundPKCE = true
				if !c.HttpOnly {
					t.Errorf("PKCE cookie must be HttpOnly")
				}
			}
		}
		if !foundPKCE {
			t.Errorf("PKCE session cookie missing")
		}
	})

	t.Run("HandleMe Returns Unauthorized Without Cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("HandleLogout Clears Session", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var res map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&res)
		if res["status"] != "logged_out" {
			t.Errorf("expected status logged_out, got %s", res["status"])
		}
	})
}
