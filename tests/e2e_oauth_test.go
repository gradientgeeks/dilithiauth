package tests

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

func TestE2E_FullOAuth2PostQuantumFlow(t *testing.T) {
	issuer := "https://dilithiauth.gradientgeeks.tech"

	// 1. Initialize Hybrid Key Pair & IdP Provider
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("e2e: failed to generate hybrid keys: %v", err)
	}

	provider := oauth.NewProvider(issuer, kp)
	authzStore := oauth.NewAuthzStore()
	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	// Add seed user
	authzStore.AddUser("default", &oauth.User{
		Username: "alice",
		Password: "SuperSecretPassword123!",
		Subject:  "sub_alice_12345",
	})

	// Add registered client app
	clientID := "da_gradientgeeks_app"
	clientSecret := "dacs_super_secret_key"
	redirectURI := "https://client.gradientgeeks.tech/callback"

	// Hash client secret using bcrypt
	secretHashBytes, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("e2e: failed to hash client secret: %v", err)
	}

	authzStore.AddClient("default", &oauth.OAuthClient{
		ClientID:         clientID,
		ClientSecretHash: string(secretHashBytes),
		RedirectURI:      redirectURI,
		Name:             "GradientGeeks Test Client",
	})

	// 2. Setup HTTP Server Router
	r := chi.NewRouter()

	// OIDC Endpoints
	r.Get("/.well-known/openid-configuration", provider.DiscoveryHandler())
	r.Get("/.well-known/jwks.json", provider.JWKSHandler())

	// OAuth 2.0 Handlers
	oauthHandlers := &oauth.AuthzHandlers{
		Provider: provider,
		Store:    authzStore,
		SigStore: sigStore,
	}
	r.Get("/authorize", oauthHandlers.AuthorizeGET)
	r.Post("/authorize", oauthHandlers.AuthorizePOST)
	r.Post("/token", oauthHandlers.TokenExchange)

	// Direct token issuance endpoints for testing full/jumbo and CWT tokens
	r.Post("/token/full", func(w http.ResponseWriter, r *http.Request) {
		sub := r.URL.Query().Get("sub")
		if sub == "" {
			sub = "sub_alice_12345"
		}
		tokStr, err := provider.IssueJWTFull("default", sub, time.Hour)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": tokStr,
			"token_type":   "Bearer",
		})
	})

	// Protected Endpoint
	r.With(v.Middleware()).Get("/protected/whoami", func(w http.ResponseWriter, r *http.Request) {
		claims := verifier.ClaimsFromContext(r.Context())
		if claims == nil {
			http.Error(w, "missing claims", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claims)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := ts.Client()

	// Step 1: Verify OIDC Discovery
	t.Run("OIDC Discovery Document", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/.well-known/openid-configuration")
		if err != nil {
			t.Fatalf("failed to fetch discovery: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var doc oauth.DiscoveryDocument
		if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			t.Fatalf("failed to decode discovery: %v", err)
		}

		if doc.Issuer != issuer {
			t.Errorf("expected issuer %s, got %s", issuer, doc.Issuer)
		}
		if !doc.PQCHybridEnabled {
			t.Errorf("expected PQCHybridEnabled = true")
		}
	})

	// Step 2: Verify JWKS Endpoint
	t.Run("JWKS Key Material", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/.well-known/jwks.json")
		if err != nil {
			t.Fatalf("failed to fetch JWKS: %v", err)
		}
		defer resp.Body.Close()

		var jwks oauth.JWKSDocument
		if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
			t.Fatalf("failed to decode jwks: %v", err)
		}

		if len(jwks.Keys) == 0 {
			t.Fatalf("expected at least 1 key in JWKS")
		}
		if jwks.Keys[0].MLDSA44Pub == "" {
			t.Errorf("expected ML-DSA-44 public key in JWKS")
		}
	})

	// Step 3: OAuth 2.0 Authorization Code Flow with PKCE
	t.Run("Authorization Code Flow with PKCE and Thin JWT", func(t *testing.T) {
		// PKCE Code Verifier & Challenge (S256)
		codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		hash := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

		// A. User submits login form at /authorize
		formData := url.Values{
			"username":              {"alice"},
			"password":              {"SuperSecretPassword123!"},
			"client_id":             {clientID},
			"redirect_uri":          {redirectURI},
			"scope":                 {"openid profile"},
			"code_challenge":        {codeChallenge},
			"code_challenge_method": {"S256"},
		}

		// Disable auto-redirect to capture 302 Location header
		noRedirectClient := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err := noRedirectClient.PostForm(ts.URL+"/authorize", formData)
		if err != nil {
			t.Fatalf("failed to submit authorization form: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusFound {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 302 redirect, got %d: %s", resp.StatusCode, string(body))
		}

		location := resp.Header.Get("Location")
		u, err := url.Parse(location)
		if err != nil {
			t.Fatalf("failed to parse redirect location: %v", err)
		}

		code := u.Query().Get("code")
		if code == "" {
			t.Fatalf("authorization code missing from redirect URL: %s", location)
		}

		// B. Exchange code for access token at /token
		tokenFormData := url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {clientID},
			"client_secret": {clientSecret},
			"code":          {code},
			"redirect_uri":  {redirectURI},
			"code_verifier": {codeVerifier},
		}

		tokenResp, err := client.PostForm(ts.URL+"/token", tokenFormData)
		if err != nil {
			t.Fatalf("failed to request token: %v", err)
		}
		defer tokenResp.Body.Close()

		if tokenResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(tokenResp.Body)
			t.Fatalf("token exchange failed with status %d: %s", tokenResp.StatusCode, string(body))
		}

		var tokenResult map[string]interface{}
		if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
			t.Fatalf("failed to decode token response: %v", err)
		}

		accessToken, ok := tokenResult["access_token"].(string)
		if !ok || accessToken == "" {
			t.Fatalf("access_token missing from response")
		}

		// C. Access Protected Resource using the issued Post-Quantum Access Token
		req, _ := http.NewRequest("GET", ts.URL+"/protected/whoami", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		protResp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to call protected endpoint: %v", err)
		}
		defer protResp.Body.Close()

		if protResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(protResp.Body)
			t.Fatalf("protected endpoint returned %d: %s", protResp.StatusCode, string(body))
		}

		var claims verifier.VerifiedClaims
		if err := json.NewDecoder(protResp.Body).Decode(&claims); err != nil {
			t.Fatalf("failed to decode claims: %v", err)
		}

		if claims.Subject != "sub_alice_12345" && claims.Subject != "alice" {
			t.Errorf("expected subject alice, got %s", claims.Subject)
		}
	})

	// Step 4: Verify Jumbo JWT (Self-Contained Post-Quantum Token)
	t.Run("Jumbo JWT Self-Contained Verification", func(t *testing.T) {
		resp, err := client.Post(ts.URL+"/token/full?sub=sub_alice_12345", "application/json", nil)
		if err != nil {
			t.Fatalf("failed to request jumbo token: %v", err)
		}
		defer resp.Body.Close()

		var res map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode jumbo token response: %v", err)
		}

		jumboToken := res["access_token"]
		req, _ := http.NewRequest("GET", ts.URL+"/protected/whoami", nil)
		req.Header.Set("Authorization", "Bearer "+jumboToken)

		protResp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to call protected endpoint with jumbo token: %v", err)
		}
		defer protResp.Body.Close()

		if protResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(protResp.Body)
			t.Fatalf("protected endpoint with jumbo token returned %d: %s", protResp.StatusCode, string(body))
		}
	})
}
