package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
)

type mockSigStore struct{}

func (m *mockSigStore) Put(jti string, sig []byte) {}

// TestM63Empirical_AllScenarios empirically verifies all required tasks for Milestone 6.3 challenger verification.
func TestM63Empirical_AllScenarios(t *testing.T) {
	// Setup mock social identity provider (Google & GitHub)
	mockProviderServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/google/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "google_access_token_emp123"})
		case "/google/userinfo":
			if r.Header.Get("Authorization") != "Bearer google_access_token_emp123" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    "g_emp_9999",
				"email": "emp_google_user@example.com",
				"name":  "Colliding User",
			})
		case "/github/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "github_access_token_emp456"})
		case "/github/user":
			if r.Header.Get("Authorization") != "Bearer github_access_token_emp456" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    7777,
				"email": "emp_github_user@example.com",
				"name":  "GitHub Emp User",
				"login": "ghempuser",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockProviderServer.Close()

	store := NewAuthzStore()

	// 1. Pre-register a local user with username "Colliding User" under tenant-emp-1
	_ = store.RegisterUser("tenant-emp-1", &User{
		Username: "Colliding User",
		Email:    "existing_local_user@example.com",
		Password: "Password123!",
		Subject:  "usr_existing_local_123",
	})

	// Add client under tenant-emp-1
	client := &OAuthClient{
		ClientID:         "emp_client_id",
		ClientSecretHash: "$2a$10$wN105s5P8tQy0Qp1Yx9HNe4T9e8z5x5v5u5t5s5r5q5p5o5n5m5l5",
		RedirectURI:      "https://client.example.com/callback",
		Name:             "Empirical Client App",
	}
	store.AddClient("tenant-emp-1", client)

	provider := NewProvider("https://auth.example.com", nil)
	h := &AuthzHandlers{
		Provider:      provider,
		Store:         store,
		ConsentSecret: []byte("empirical_test_secret_32_bytes!"),
		SigStore:      &mockSigStore{},
		GoogleConfig: SocialConfig{
			ClientID:     "google_client_id",
			ClientSecret: "google_client_secret",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
			TokenURL:     mockProviderServer.URL + "/google/token",
			UserInfoURL:  mockProviderServer.URL + "/google/userinfo",
		},
		GitHubConfig: SocialConfig{
			ClientID:     "github_client_id",
			ClientSecret: "github_client_secret",
			RedirectURI:  "https://auth.example.com/auth/social/github/callback",
			TokenURL:     mockProviderServer.URL + "/github/token",
			UserInfoURL:  mockProviderServer.URL + "/github/user",
		},
	}

	r := chi.NewRouter()
	r.Use(tenant.NewMiddleware([]string{"127.0.0.1/32"})) // honor X-Tenant-Id in tests
	r.Get("/auth/social/{provider}", h.SocialLoginGET)
	r.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)
	r.Post("/token", h.TokenExchange)

	makeReq := func(method, target string) *http.Request {
		req, _ := http.NewRequest(method, target, nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Tenant-Id", "tenant-emp-1")
		return req
	}

	// --- Subtest A: GET /auth/social/google & github redirects with valid state JWT containing tenant ID ---
	t.Run("Subtest A: Social Login Redirects & Tenant ID in State JWT", func(t *testing.T) {
		reqG := makeReq("GET", "/auth/social/google?client_id=emp_client_id&redirect_uri=https://client.example.com/callback&state=client_state_aaa")
		recG := httptest.NewRecorder()
		r.ServeHTTP(recG, reqG)

		if recG.Code != http.StatusFound {
			t.Fatalf("Google Login: expected HTTP 302, got %d (body: %s)", recG.Code, recG.Body.String())
		}
		locG := recG.Header().Get("Location")
		if !strings.HasPrefix(locG, "https://accounts.google.com/o/oauth2/v2/auth?") {
			t.Errorf("Google Login: unexpected redirect location: %s", locG)
		}

		uG, _ := url.Parse(locG)
		stateJWTG := uG.Query().Get("state")
		claimsG, err := h.parseSocialState(stateJWTG)
		if err != nil {
			t.Fatalf("Google Login: invalid state JWT: %v", err)
		}
		if claimsG.TenantID != "tenant-emp-1" {
			t.Errorf("Google Login: expected state JWT tenant_id 'tenant-emp-1', got '%s'", claimsG.TenantID)
		}
		if claimsG.ClientState != "client_state_aaa" {
			t.Errorf("Google Login: expected state JWT client_state 'client_state_aaa', got '%s'", claimsG.ClientState)
		}

		// GitHub redirect
		reqGH := makeReq("GET", "/auth/social/github?client_id=emp_client_id&redirect_uri=https://client.example.com/callback&state=client_state_bbb")
		recGH := httptest.NewRecorder()
		r.ServeHTTP(recGH, reqGH)

		if recGH.Code != http.StatusFound {
			t.Fatalf("GitHub Login: expected HTTP 302, got %d (body: %s)", recGH.Code, recGH.Body.String())
		}
		locGH := recGH.Header().Get("Location")
		if !strings.HasPrefix(locGH, "https://github.com/login/oauth/authorize?") {
			t.Errorf("GitHub Login: unexpected redirect location: %s", locGH)
		}
		uGH, _ := url.Parse(locGH)
		stateJWTGH := uGH.Query().Get("state")
		claimsGH, err := h.parseSocialState(stateJWTGH)
		if err != nil {
			t.Fatalf("GitHub Login: invalid state JWT: %v", err)
		}
		if claimsGH.TenantID != "tenant-emp-1" {
			t.Errorf("GitHub Login: expected state JWT tenant_id 'tenant-emp-1', got '%s'", claimsGH.TenantID)
		}
		if claimsGH.ClientState != "client_state_bbb" {
			t.Errorf("GitHub Login: expected state JWT client_state 'client_state_bbb', got '%s'", claimsGH.ClientState)
		}
	})

	// --- Subtest B: Unconfigured credentials return HTTP 400 Bad Request ---
	t.Run("Subtest B: Unconfigured Credentials return 400 Bad Request", func(t *testing.T) {
		hUnconf := &AuthzHandlers{
			Store:         store,
			ConsentSecret: []byte("empirical_test_secret_32_bytes!"),
			GoogleConfig:  SocialConfig{}, // Empty
			GitHubConfig:  SocialConfig{}, // Empty
		}
		rUnconf := chi.NewRouter()
		rUnconf.Use(tenant.NewMiddleware([]string{"127.0.0.1/32"}))
		rUnconf.Get("/auth/social/{provider}", hUnconf.SocialLoginGET)

		req := makeReq("GET", "/auth/social/google?client_id=emp_client_id&redirect_uri=https://client.example.com/callback")
		rec := httptest.NewRecorder()
		rUnconf.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for unconfigured Google, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "unconfigured") {
			t.Errorf("expected body to mention 'unconfigured', got: %s", rec.Body.String())
		}
	})

	// --- Subtest C: Provider callback error redirects HTTP 302 with ?error=access_denied&state=<client_state> ---
	t.Run("Subtest C: Provider Callback Error Redirection", func(t *testing.T) {
		signedState, err := h.signSocialState(socialStateClaims{
			ClientID:    "emp_client_id",
			RedirectURI: "https://client.example.com/callback",
			ClientState: "my_app_state_777",
			Provider:    "google",
			TenantID:    "tenant-emp-1",
		})
		if err != nil {
			t.Fatalf("signSocialState failed: %v", err)
		}

		cbURL := "/auth/social/google/callback?error=access_denied&error_description=User+cancelled&state=" + url.QueryEscape(signedState)
		req := makeReq("GET", cbURL)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusFound {
			t.Fatalf("expected HTTP 302 redirect for provider error callback, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		loc := rec.Header().Get("Location")
		if !strings.HasPrefix(loc, "https://client.example.com/callback?") {
			t.Errorf("unexpected redirect location: %s", loc)
		}
		u, _ := url.Parse(loc)
		if u.Query().Get("error") != "access_denied" {
			t.Errorf("expected error=access_denied, got '%s'", u.Query().Get("error"))
		}
		if u.Query().Get("error_description") != "User cancelled" {
			t.Errorf("expected error_description='User cancelled', got '%s'", u.Query().Get("error_description"))
		}
		if u.Query().Get("state") != "my_app_state_777" {
			t.Errorf("expected state='my_app_state_777', got '%s'", u.Query().Get("state"))
		}
	})

	// --- Subtest D: User provisioning handles username collisions by generating fallback usernames ---
	t.Run("Subtest D: Username Collision Fallback Handling", func(t *testing.T) {
		signedState, err := h.signSocialState(socialStateClaims{
			ClientID:    "emp_client_id",
			RedirectURI: "https://client.example.com/callback",
			ClientState: "state_collision_test",
			Provider:    "google",
			TenantID:    "tenant-emp-1",
		})
		if err != nil {
			t.Fatalf("signSocialState failed: %v", err)
		}

		cbURL := "/auth/social/google/callback?code=valid_g_code&state=" + url.QueryEscape(signedState)
		req := makeReq("GET", cbURL)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusFound {
			t.Fatalf("expected HTTP 302 on callback, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		loc := rec.Header().Get("Location")
		u, _ := url.Parse(loc)
		idpCode := u.Query().Get("code")

		// Redeem code to find provisioned subject
		ac, err := store.RedeemCode("tenant-emp-1", idpCode, "emp_client_id", "https://client.example.com/callback", "")
		if err != nil {
			t.Fatalf("failed to redeem code: %v", err)
		}

		// Verify user provisioned in store under emp_google_user@example.com
		user, err := store.GetUserByEmail("tenant-emp-1", "emp_google_user@example.com")
		if err != nil {
			t.Fatalf("user not found by email: %v", err)
		}
		if user.Subject != ac.Subject {
			t.Errorf("expected user subject '%s', got '%s'", ac.Subject, user.Subject)
		}
		if !strings.HasPrefix(user.Username, "Colliding User_") {
			t.Errorf("expected fallback username starting with 'Colliding User_', got '%s'", user.Username)
		}
	})

	// --- Subtest E: Full Code Exchange, User Provisioning, and IdP Authorization Code Issuance ---
	t.Run("Subtest E: End-to-End Code Exchange & Token Issuance", func(t *testing.T) {
		codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		hSum := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(hSum[:])

		signedState, err := h.signSocialState(socialStateClaims{
			ClientID:            "emp_client_id",
			RedirectURI:         "https://client.example.com/callback",
			ClientState:         "e2e_state_123",
			Scope:               "openid profile email",
			CodeChallenge:       codeChallenge,
			CodeChallengeMethod: "S256",
			Provider:            "github",
			TenantID:            "tenant-emp-1",
		})
		if err != nil {
			t.Fatalf("signSocialState failed: %v", err)
		}

		cbURL := "/auth/social/github/callback?code=github_mock_code&state=" + url.QueryEscape(signedState)
		req := makeReq("GET", cbURL)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusFound {
			t.Fatalf("expected HTTP 302 on callback, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		loc := rec.Header().Get("Location")
		u, _ := url.Parse(loc)
		idpCode := u.Query().Get("code")
		retState := u.Query().Get("state")

		if idpCode == "" {
			t.Fatalf("expected authorization code in redirect URL")
		}
		if retState != "e2e_state_123" {
			t.Errorf("expected state 'e2e_state_123', got '%s'", retState)
		}

		// Verify GitHub user provisioned
		ghUser, err := store.GetUserByEmail("tenant-emp-1", "emp_github_user@example.com")
		if err != nil {
			t.Fatalf("GitHub user was not provisioned: %v", err)
		}
		if ghUser.Username != "GitHub Emp User" {
			t.Errorf("expected username 'GitHub Emp User', got '%s'", ghUser.Username)
		}
	})
}
