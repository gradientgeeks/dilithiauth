package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
)

func TestGetSocialAuthURL(t *testing.T) {
	cfg := SocialConfig{
		ClientID:    "test_client_id",
		RedirectURI: "https://myapp.com/callback",
	}

	googleURL, err := GetSocialAuthURL(ProviderGoogle, cfg, "state123")
	if err != nil {
		t.Fatalf("GetSocialAuthURL Google error: %v", err)
	}
	if googleURL == "" || !strings.Contains(googleURL, "accounts.google.com") {
		t.Errorf("Google auth URL is invalid: %s", googleURL)
	}

	githubURL, err := GetSocialAuthURL(ProviderGitHub, cfg, "state123")
	if err != nil {
		t.Fatalf("GetSocialAuthURL GitHub error: %v", err)
	}
	if githubURL == "" || !strings.Contains(githubURL, "github.com") {
		t.Errorf("GitHub auth URL is invalid: %s", githubURL)
	}
}

func TestExchangeSocialCode_Google_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if r.Method != "POST" {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "mock_google_token",
			})
		case "/userinfo":
			if r.Header.Get("Authorization") != "Bearer mock_google_token" {
				t.Errorf("unexpected authorization header: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":    "google_user_123",
				"email": "alice@gmail.com",
				"name":  "Alice Smith",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := SocialConfig{
		ClientID:     "g_client",
		ClientSecret: "g_secret",
		RedirectURI:  "http://localhost/callback",
		TokenURL:     ts.URL + "/token",
		UserInfoURL:  ts.URL + "/userinfo",
	}

	user, err := ExchangeSocialCode(context.Background(), ProviderGoogle, cfg, "mock_code")
	if err != nil {
		t.Fatalf("ExchangeSocialCode failed: %v", err)
	}

	if user.ID != "google_user_123" {
		t.Errorf("expected ID google_user_123, got %s", user.ID)
	}
	if user.Email != "alice@gmail.com" {
		t.Errorf("expected email alice@gmail.com, got %s", user.Email)
	}
	if user.Name != "Alice Smith" {
		t.Errorf("expected name Alice Smith, got %s", user.Name)
	}
	if user.Provider != ProviderGoogle {
		t.Errorf("expected provider google, got %s", user.Provider)
	}
}

func TestExchangeSocialCode_GitHub_PrivateEmail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "mock_github_token",
			})
		case "/user":
			if r.Header.Get("Authorization") != "Bearer mock_github_token" {
				t.Errorf("unexpected authorization header: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    45678,
				"email": nil,
				"name":  "Bob GitHub",
				"login": "bobgh",
			})
		case "/user/emails":
			if r.Header.Get("Authorization") != "Bearer mock_github_token" {
				t.Errorf("unexpected authorization header: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{"email": "bob_secondary@example.com", "primary": false, "verified": true},
				{"email": "bob_primary@example.com", "primary": true, "verified": true},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := SocialConfig{
		ClientID:     "gh_client",
		ClientSecret: "gh_secret",
		RedirectURI:  "http://localhost/callback",
		TokenURL:     ts.URL + "/token",
		UserInfoURL:  ts.URL + "/user",
	}

	user, err := ExchangeSocialCode(context.Background(), ProviderGitHub, cfg, "mock_code")
	if err != nil {
		t.Fatalf("ExchangeSocialCode failed: %v", err)
	}

	if user.ID != "45678" {
		t.Errorf("expected ID 45678, got %s", user.ID)
	}
	if user.Email != "bob_primary@example.com" {
		t.Errorf("expected email bob_primary@example.com, got %s", user.Email)
	}
	if user.Name != "Bob GitHub" {
		t.Errorf("expected name Bob GitHub, got %s", user.Name)
	}
	if user.Provider != ProviderGitHub {
		t.Errorf("expected provider github, got %s", user.Provider)
	}
}

func TestSocialLoginGET(t *testing.T) {
	store := NewAuthzStore()
	client := &OAuthClient{
		ClientID:    "test_app",
		RedirectURI: "https://client.example.com/callback",
		Name:        "Test App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("test_consent_secret_32bytes_long!!"),
		GoogleConfig: SocialConfig{
			ClientID:     "google_client_123",
			ClientSecret: "google_secret_456",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
		},
	}

	r := chi.NewRouter()
	r.Get("/auth/social/{provider}", h.SocialLoginGET)

	// Test 1: Unconfigured provider (GitHub) -> 400 Bad Request
	reqUnconf := httptest.NewRequest("GET", "/auth/social/github?client_id=test_app&redirect_uri=https://client.example.com/callback", nil)
	recUnconf := httptest.NewRecorder()
	r.ServeHTTP(recUnconf, reqUnconf)

	if recUnconf.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for unconfigured provider, got %d", recUnconf.Code)
	}

	// Test 2: Configured provider (Google) -> 302 Redirect to Google Auth URL with signed state
	reqGoogle := httptest.NewRequest("GET", "/auth/social/google?client_id=test_app&redirect_uri=https://client.example.com/callback&state=client_state_789", nil)
	recGoogle := httptest.NewRecorder()
	r.ServeHTTP(recGoogle, reqGoogle)

	if recGoogle.Code != http.StatusFound {
		t.Fatalf("expected status 302, got %d", recGoogle.Code)
	}

	loc := recGoogle.Header().Get("Location")
	if !strings.Contains(loc, "accounts.google.com") {
		t.Errorf("expected redirect to accounts.google.com, got %s", loc)
	}
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("failed to parse redirect location: %v", err)
	}
	signedState := u.Query().Get("state")
	if signedState == "" {
		t.Fatalf("expected signed state query parameter in location")
	}

	// Verify state claims can be parsed
	claims, err := h.parseSocialState(signedState)
	if err != nil {
		t.Fatalf("failed to parse signed state: %v", err)
	}
	if claims.ClientID != "test_app" {
		t.Errorf("expected client_id test_app in state claims, got %s", claims.ClientID)
	}
	if claims.ClientState != "client_state_789" {
		t.Errorf("expected client_state client_state_789 in state claims, got %s", claims.ClientState)
	}
	if claims.Provider != "google" {
		t.Errorf("expected provider google in state claims, got %s", claims.Provider)
	}
}

func TestSocialCallbackGET(t *testing.T) {
	// Mock Google OAuth server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "mock_access_token",
			})
		case "/userinfo":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":    "g_user_99",
				"email": "charlie@example.com",
				"name":  "Charlie Brown",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	store := NewAuthzStore()
	client := &OAuthClient{
		ClientID:    "demo_client",
		RedirectURI: "https://client.example.com/oauth/callback",
		Name:        "Demo Client App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("test_consent_secret_32bytes_long!!"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_client",
			ClientSecret: "g_secret",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
			TokenURL:     ts.URL + "/token",
			UserInfoURL:  ts.URL + "/userinfo",
		},
	}

	r := chi.NewRouter()
	r.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	// Generate signed state
	signedState, err := h.signSocialState(socialStateClaims{
		ClientID:            "demo_client",
		RedirectURI:         "https://client.example.com/oauth/callback",
		ClientState:         "my_custom_state_123",
		Scope:               "openid profile email",
		CodeChallenge:       "",
		CodeChallengeMethod: "",
		Provider:            "google",
	})
	if err != nil {
		t.Fatalf("signSocialState failed: %v", err)
	}

	callbackURL := "/auth/social/google/callback?code=valid_social_code&state=" + url.QueryEscape(signedState)
	req := httptest.NewRequest("GET", callbackURL, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect on callback, got status %d body: %s", rec.Code, rec.Body.String())
	}

	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://client.example.com/oauth/callback?") {
		t.Errorf("expected redirect to client redirect_uri, got %s", loc)
	}
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("failed to parse redirect URI: %v", err)
	}

	idpCode := u.Query().Get("code")
	if idpCode == "" {
		t.Errorf("expected authorization code in redirect query")
	}
	retState := u.Query().Get("state")
	if retState != "my_custom_state_123" {
		t.Errorf("expected client state my_custom_state_123, got %s", retState)
	}

	// Verify user was provisioned in store
	user, err := store.GetUserByEmail(tenant.DefaultTenantID, "charlie@example.com")
	if err != nil {
		t.Fatalf("expected provisioned user in store, got error: %v", err)
	}
	if user.Email != "charlie@example.com" {
		t.Errorf("expected user email charlie@example.com, got %s", user.Email)
	}
	if !strings.HasPrefix(user.Subject, "usr_") {
		t.Errorf("expected subject prefix usr_, got %s", user.Subject)
	}
}

func TestSocialCallbackGET_ProviderError(t *testing.T) {
	store := NewAuthzStore()
	client := &OAuthClient{
		ClientID:    "prov_err_client",
		RedirectURI: "https://app.example.com/callback",
		Name:        "Provider Error Test App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("test_consent_secret_32bytes_long!!"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_client",
			ClientSecret: "g_secret",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
		},
	}

	r := chi.NewRouter()
	r.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	signedState, err := h.signSocialState(socialStateClaims{
		ClientID:    "prov_err_client",
		RedirectURI: "https://app.example.com/callback",
		ClientState: "state_abc_123",
		Provider:    "google",
	})
	if err != nil {
		t.Fatalf("failed to sign state: %v", err)
	}

	// Case 1: Valid state with provider error -> 302 Redirect to client redirect_uri
	reqURL := "/auth/social/google/callback?error=access_denied&error_description=User+cancelled&state=" + url.QueryEscape(signedState)
	req := httptest.NewRequest("GET", reqURL, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://app.example.com/callback?") {
		t.Errorf("unexpected redirect location: %s", loc)
	}

	u, _ := url.Parse(loc)
	if u.Query().Get("error") != "access_denied" {
		t.Errorf("expected error=access_denied, got %s", u.Query().Get("error"))
	}
	if u.Query().Get("error_description") != "User cancelled" {
		t.Errorf("expected error_description 'User cancelled', got %s", u.Query().Get("error_description"))
	}
	if u.Query().Get("state") != "state_abc_123" {
		t.Errorf("expected state 'state_abc_123', got %s", u.Query().Get("state"))
	}

	// Case 2: Invalid state with provider error -> 400 Bad Request
	reqInvalidState := httptest.NewRequest("GET", "/auth/social/google/callback?error=access_denied&state=invalid_state_jwt", nil)
	recInvalidState := httptest.NewRecorder()
	r.ServeHTTP(recInvalidState, reqInvalidState)

	if recInvalidState.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid state with provider error, got %d", recInvalidState.Code)
	}
}

func TestSocialCallbackGET_UsernameCollisionFallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "mock_tok"})
		case "/userinfo":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":    "g_user_collision",
				"email": "social_jane@example.com",
				"name":  "Jane Doe",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	store := NewAuthzStore()
	// Pre-register user with username "Jane Doe" but different email
	_ = store.RegisterUser(tenant.DefaultTenantID, &User{
		Username: "Jane Doe",
		Email:    "local_jane@example.com",
		Password: "Password123!",
		Subject:  "usr_local_jane",
	})

	client := &OAuthClient{
		ClientID:    "collision_app",
		RedirectURI: "https://app.example.com/cb",
		Name:        "Collision Test App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("test_consent_secret_32bytes_long!!"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_client",
			ClientSecret: "g_secret",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
			TokenURL:     ts.URL + "/token",
			UserInfoURL:  ts.URL + "/userinfo",
		},
	}

	r := chi.NewRouter()
	r.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	signedState, _ := h.signSocialState(socialStateClaims{
		ClientID:    "collision_app",
		RedirectURI: "https://app.example.com/cb",
		Provider:    "google",
	})

	req := httptest.NewRequest("GET", "/auth/social/google/callback?code=mock_code&state="+url.QueryEscape(signedState), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect on username collision fallback, got status %d body: %s", rec.Code, rec.Body.String())
	}

	// Verify new user was provisioned with fallback username
	user, err := store.GetUserByEmail(tenant.DefaultTenantID, "social_jane@example.com")
	if err != nil {
		t.Fatalf("expected provisioned user in store, got error: %v", err)
	}
	if !strings.HasPrefix(user.Username, "Jane Doe_") {
		t.Errorf("expected fallback username starting with 'Jane Doe_', got %s", user.Username)
	}
}

func TestSocialCallbackGET_TenantIDValidation(t *testing.T) {
	store := NewAuthzStore()
	client := &OAuthClient{
		ClientID:    "tenant_client",
		RedirectURI: "https://app.example.com/cb",
		Name:        "Tenant Test App",
	}
	store.AddClient("tenant-b", client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("test_consent_secret_32bytes_long!!"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_client",
			ClientSecret: "g_secret",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
		},
	}

	r := chi.NewRouter()
	r.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	// State signed for tenant-a
	signedState, _ := h.signSocialState(socialStateClaims{
		ClientID:    "tenant_client",
		RedirectURI: "https://app.example.com/cb",
		Provider:    "google",
		TenantID:    "tenant-a",
	})

	// Callback request evaluated under tenant-b context
	req := httptest.NewRequest("GET", "/auth/social/google/callback?code=mock_code&state="+url.QueryEscape(signedState), nil)
	req = req.WithContext(tenant.WithTenant(req.Context(), "tenant-b"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on tenant mismatch, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "tenant mismatch") {
		t.Errorf("expected error message to mention 'tenant mismatch', got: %s", rec.Body.String())
	}
}
