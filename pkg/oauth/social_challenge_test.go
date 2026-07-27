package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
)

// TestSocialAuth_Redirects_And_Unconfigured tests task requirement 2a & 2b:
// - GET /auth/social/google and GET /auth/social/github respond with HTTP 302 redirects containing valid state JWTs when credentials are provided.
// - Unconfigured credentials return HTTP 400 Bad Request.
func TestSocialAuth_Redirects_And_Unconfigured(t *testing.T) {
	store := NewAuthzStore()
	client := &OAuthClient{
		ClientID:    "valid_client",
		RedirectURI: "https://app.example.com/callback",
		Name:        "Valid App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	// Case A: Google configured, GitHub unconfigured
	hGoogleOnly := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("secret_key_32_bytes_long_1234567"),
		GoogleConfig: SocialConfig{
			ClientID:     "google_client_id_1",
			ClientSecret: "google_client_secret_1",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
		},
		// GitHubConfig is empty (unconfigured)
	}

	router := chi.NewRouter()
	router.Get("/auth/social/{provider}", hGoogleOnly.SocialLoginGET)

	// 1. Google request with credentials -> 302 redirect
	reqGoogle := httptest.NewRequest("GET", "/auth/social/google?client_id=valid_client&redirect_uri=https://app.example.com/callback&state=client_state_123", nil)
	recGoogle := httptest.NewRecorder()
	router.ServeHTTP(recGoogle, reqGoogle)

	if recGoogle.Code != http.StatusFound {
		t.Fatalf("Google: expected 302 redirect, got %d", recGoogle.Code)
	}

	locGoogle := recGoogle.Header().Get("Location")
	if !strings.HasPrefix(locGoogle, "https://accounts.google.com/o/oauth2/v2/auth?") {
		t.Errorf("Google: unexpected redirect location: %s", locGoogle)
	}

	uGoogle, err := url.Parse(locGoogle)
	if err != nil {
		t.Fatalf("Google: failed to parse location URL: %v", err)
	}

	stateJWT := uGoogle.Query().Get("state")
	if stateJWT == "" {
		t.Fatalf("Google: state parameter missing from redirect URL")
	}

	// Validate the state JWT
	claimsGoogle, err := hGoogleOnly.parseSocialState(stateJWT)
	if err != nil {
		t.Fatalf("Google: invalid state JWT: %v", err)
	}
	if claimsGoogle.ClientID != "valid_client" || claimsGoogle.ClientState != "client_state_123" || claimsGoogle.Provider != "google" {
		t.Errorf("Google: state JWT claims mismatch: %+v", claimsGoogle)
	}

	// 2. GitHub request unconfigured -> 400 Bad Request
	reqGitHubUnconf := httptest.NewRequest("GET", "/auth/social/github?client_id=valid_client&redirect_uri=https://app.example.com/callback", nil)
	recGitHubUnconf := httptest.NewRecorder()
	router.ServeHTTP(recGitHubUnconf, reqGitHubUnconf)

	if recGitHubUnconf.Code != http.StatusBadRequest {
		t.Errorf("GitHub unconfigured: expected 400 Bad Request, got %d", recGitHubUnconf.Code)
	}
	if !strings.Contains(recGitHubUnconf.Body.String(), "unconfigured") {
		t.Errorf("GitHub unconfigured: expected error msg mentioning unconfigured, got: %s", recGitHubUnconf.Body.String())
	}

	// Case B: GitHub configured, Google unconfigured
	hGitHubOnly := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("secret_key_32_bytes_long_1234567"),
		GitHubConfig: SocialConfig{
			ClientID:     "github_client_id_2",
			ClientSecret: "github_client_secret_2",
			RedirectURI:  "https://auth.example.com/auth/social/github/callback",
		},
	}

	routerB := chi.NewRouter()
	routerB.Get("/auth/social/{provider}", hGitHubOnly.SocialLoginGET)

	// 3. GitHub request with credentials -> 302 redirect
	reqGitHub := httptest.NewRequest("GET", "/auth/social/github?client_id=valid_client&redirect_uri=https://app.example.com/callback&state=client_state_456", nil)
	recGitHub := httptest.NewRecorder()
	routerB.ServeHTTP(recGitHub, reqGitHub)

	if recGitHub.Code != http.StatusFound {
		t.Fatalf("GitHub: expected 302 redirect, got %d", recGitHub.Code)
	}

	locGitHub := recGitHub.Header().Get("Location")
	if !strings.HasPrefix(locGitHub, "https://github.com/login/oauth/authorize?") {
		t.Errorf("GitHub: unexpected redirect location: %s", locGitHub)
	}

	uGitHub, err := url.Parse(locGitHub)
	if err != nil {
		t.Fatalf("GitHub: failed to parse location URL: %v", err)
	}

	ghStateJWT := uGitHub.Query().Get("state")
	claimsGitHub, err := hGitHubOnly.parseSocialState(ghStateJWT)
	if err != nil {
		t.Fatalf("GitHub: invalid state JWT: %v", err)
	}
	if claimsGitHub.ClientID != "valid_client" || claimsGitHub.ClientState != "client_state_456" || claimsGitHub.Provider != "github" {
		t.Errorf("GitHub: state JWT claims mismatch: %+v", claimsGitHub)
	}

	// 4. Google request unconfigured -> 400 Bad Request
	reqGoogleUnconf := httptest.NewRequest("GET", "/auth/social/google?client_id=valid_client&redirect_uri=https://app.example.com/callback", nil)
	recGoogleUnconf := httptest.NewRecorder()
	routerB.ServeHTTP(recGoogleUnconf, reqGoogleUnconf)

	if recGoogleUnconf.Code != http.StatusBadRequest {
		t.Errorf("Google unconfigured: expected 400 Bad Request, got %d", recGoogleUnconf.Code)
	}
}

// TestSocialAuth_CallbackProcessing tests task requirement 2c:
// - Callback processing validates state JWT signature, handles mock code exchange, provisions user, and redirects HTTP 302 to client redirect_uri with authorization code.
func TestSocialAuth_CallbackProcessing(t *testing.T) {
	// Mock Social Identity Provider Server
	mockProviderServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/google/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "google_tok_123"})
		case "/google/userinfo":
			if r.Header.Get("Authorization") != "Bearer google_tok_123" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    "g_1001",
				"email": "user_google@example.com",
				"name":  "Google User",
			})
		case "/github/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "github_tok_456"})
		case "/github/user":
			if r.Header.Get("Authorization") != "Bearer github_tok_456" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    5002,
				"email": "user_github@example.com",
				"name":  "GitHub User",
				"login": "ghuser",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockProviderServer.Close()

	store := NewAuthzStore()
	client := &OAuthClient{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		Name:        "My Client App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("super_secret_social_key_32bytes"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_cid",
			ClientSecret: "g_sec",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
			TokenURL:     mockProviderServer.URL + "/google/token",
			UserInfoURL:  mockProviderServer.URL + "/google/userinfo",
		},
		GitHubConfig: SocialConfig{
			ClientID:     "gh_cid",
			ClientSecret: "gh_sec",
			RedirectURI:  "https://auth.example.com/auth/social/github/callback",
			TokenURL:     mockProviderServer.URL + "/github/token",
			UserInfoURL:  mockProviderServer.URL + "/github/user",
		},
	}

	router := chi.NewRouter()
	router.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	// 1. Successful Google Callback
	googleState, err := h.signSocialState(socialStateClaims{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		ClientState: "custom_app_state_abc",
		Scope:       "openid email",
		Provider:    "google",
	})
	if err != nil {
		t.Fatalf("failed to sign Google social state: %v", err)
	}

	reqGoogleCB := httptest.NewRequest("GET", "/auth/social/google/callback?code=mock_g_code&state="+url.QueryEscape(googleState), nil)
	recGoogleCB := httptest.NewRecorder()
	router.ServeHTTP(recGoogleCB, reqGoogleCB)

	if recGoogleCB.Code != http.StatusFound {
		t.Fatalf("Google callback: expected 302 redirect, got %d. Body: %s", recGoogleCB.Code, recGoogleCB.Body.String())
	}

	locGoogleCB := recGoogleCB.Header().Get("Location")
	if !strings.HasPrefix(locGoogleCB, "https://myclient.com/oauth2/callback?") {
		t.Errorf("Google callback: unexpected redirect location: %s", locGoogleCB)
	}

	uGCB, _ := url.Parse(locGoogleCB)
	codeGoogle := uGCB.Query().Get("code")
	stateGoogle := uGCB.Query().Get("state")

	if codeGoogle == "" {
		t.Errorf("Google callback: authorization code missing in redirect query")
	}
	if stateGoogle != "custom_app_state_abc" {
		t.Errorf("Google callback: expected custom_app_state_abc, got %s", stateGoogle)
	}

	// Verify user provisioned
	uGoogle, err := store.GetUserByEmail(tenant.DefaultTenantID, "user_google@example.com")
	if err != nil {
		t.Fatalf("Google callback: user not provisioned in store: %v", err)
	}
	if !strings.HasPrefix(uGoogle.Subject, "usr_") {
		t.Errorf("Google callback: expected subject prefix usr_, got %s", uGoogle.Subject)
	}

	// 2. Successful GitHub Callback
	githubState, err := h.signSocialState(socialStateClaims{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		ClientState: "custom_app_state_xyz",
		Scope:       "openid profile email",
		Provider:    "github",
	})
	if err != nil {
		t.Fatalf("failed to sign GitHub social state: %v", err)
	}

	reqGitHubCB := httptest.NewRequest("GET", "/auth/social/github/callback?code=mock_gh_code&state="+url.QueryEscape(githubState), nil)
	recGitHubCB := httptest.NewRecorder()
	router.ServeHTTP(recGitHubCB, reqGitHubCB)

	if recGitHubCB.Code != http.StatusFound {
		t.Fatalf("GitHub callback: expected 302 redirect, got %d. Body: %s", recGitHubCB.Code, recGitHubCB.Body.String())
	}

	locGitHubCB := recGitHubCB.Header().Get("Location")
	if !strings.HasPrefix(locGitHubCB, "https://myclient.com/oauth2/callback?") {
		t.Errorf("GitHub callback: unexpected redirect location: %s", locGitHubCB)
	}

	uGHCB, _ := url.Parse(locGitHubCB)
	codeGitHub := uGHCB.Query().Get("code")
	stateGitHub := uGHCB.Query().Get("state")

	if codeGitHub == "" {
		t.Errorf("GitHub callback: authorization code missing in redirect query")
	}
	if stateGitHub != "custom_app_state_xyz" {
		t.Errorf("GitHub callback: expected custom_app_state_xyz, got %s", stateGitHub)
	}

	// Verify user provisioned
	uGitHub, err := store.GetUserByEmail(tenant.DefaultTenantID, "user_github@example.com")
	if err != nil {
		t.Fatalf("GitHub callback: user not provisioned in store: %v", err)
	}
	if !strings.HasPrefix(uGitHub.Subject, "usr_") {
		t.Errorf("GitHub callback: expected subject prefix usr_, got %s", uGitHub.Subject)
	}
}

// TestSocialAuth_StateValidation_And_EdgeCases tests adversarial scenarios for state JWT validation.
func TestSocialAuth_StateValidation_And_EdgeCases(t *testing.T) {
	store := NewAuthzStore()
	client := &OAuthClient{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		Name:        "My Client App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("super_secret_social_key_32bytes"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_cid",
			ClientSecret: "g_sec",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
		},
	}

	router := chi.NewRouter()
	router.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	// 1. Missing state or code -> 400 Bad Request
	reqMissingState := httptest.NewRequest("GET", "/auth/social/google/callback?code=some_code", nil)
	recMissingState := httptest.NewRecorder()
	router.ServeHTTP(recMissingState, reqMissingState)
	if recMissingState.Code != http.StatusBadRequest {
		t.Errorf("missing state: expected 400, got %d", recMissingState.Code)
	}

	reqMissingCode := httptest.NewRequest("GET", "/auth/social/google/callback?state=some_state", nil)
	recMissingCode := httptest.NewRecorder()
	router.ServeHTTP(recMissingCode, reqMissingCode)
	if recMissingCode.Code != http.StatusBadRequest {
		t.Errorf("missing code: expected 400, got %d", recMissingCode.Code)
	}

	// 2. Tampered state JWT signature -> 400 Bad Request
	validState, _ := h.signSocialState(socialStateClaims{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		Provider:    "google",
	})
	tamperedState := validState + "tampered"

	reqTampered := httptest.NewRequest("GET", "/auth/social/google/callback?code=some_code&state="+url.QueryEscape(tamperedState), nil)
	recTampered := httptest.NewRecorder()
	router.ServeHTTP(recTampered, reqTampered)
	if recTampered.Code != http.StatusBadRequest {
		t.Errorf("tampered state: expected 400, got %d", recTampered.Code)
	}

	// 3. Expired state JWT -> 400 Bad Request
	expiredClaims := socialStateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-20 * time.Minute)),
		},
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		Provider:    "google",
	}
	expiredTok := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredStateStr, _ := expiredTok.SignedString(h.socialSecret())

	reqExpired := httptest.NewRequest("GET", "/auth/social/google/callback?code=some_code&state="+url.QueryEscape(expiredStateStr), nil)
	recExpired := httptest.NewRecorder()
	router.ServeHTTP(recExpired, reqExpired)
	if recExpired.Code != http.StatusBadRequest {
		t.Errorf("expired state: expected 400, got %d", recExpired.Code)
	}

	// 4. Provider mismatch in state -> 400 Bad Request
	googleStateForMismatch, _ := h.signSocialState(socialStateClaims{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		Provider:    "google",
	})

	// Presented to GitHub callback!
	reqMismatch := httptest.NewRequest("GET", "/auth/social/github/callback?code=some_code&state="+url.QueryEscape(googleStateForMismatch), nil)
	recMismatch := httptest.NewRecorder()
	router.ServeHTTP(recMismatch, reqMismatch)
	if recMismatch.Code != http.StatusBadRequest {
		t.Errorf("provider mismatch: expected 400, got %d", recMismatch.Code)
	}
	if !strings.Contains(recMismatch.Body.String(), "provider mismatch") {
		t.Errorf("provider mismatch: expected error msg mentioning provider mismatch, got %s", recMismatch.Body.String())
	}
}

// TestSocialAuth_ExistingUser_Linking tests linking social login to an existing registered user.
func TestSocialAuth_ExistingUser_Linking(t *testing.T) {
	mockProviderServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "tok_existing"})
		case "/userinfo":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    "g_9999",
				"email": "existing_user@example.com",
				"name":  "Existing User",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockProviderServer.Close()

	store := NewAuthzStore()
	// Pre-register user with email existing_user@example.com and subject sub_existing_123
	err := store.RegisterUser(tenant.DefaultTenantID, &User{
		Username: "existing_user",
		Email:    "existing_user@example.com",
		Password: "Password123!",
		Subject:  "sub_existing_123",
	})
	if err != nil {
		t.Fatalf("failed to register pre-existing user: %v", err)
	}

	client := &OAuthClient{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		Name:        "My Client App",
	}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("super_secret_social_key_32bytes"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_cid",
			ClientSecret: "g_sec",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
			TokenURL:     mockProviderServer.URL + "/token",
			UserInfoURL:  mockProviderServer.URL + "/userinfo",
		},
	}

	router := chi.NewRouter()
	router.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	stateStr, _ := h.signSocialState(socialStateClaims{
		ClientID:    "app_client",
		RedirectURI: "https://myclient.com/oauth2/callback",
		ClientState: "state_link",
		Provider:    "google",
	})

	req := httptest.NewRequest("GET", "/auth/social/google/callback?code=mock_code&state="+url.QueryEscape(stateStr), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("existing user callback: expected 302 redirect, got %d", rec.Code)
	}

	loc := rec.Header().Get("Location")
	u, _ := url.Parse(loc)
	idpCode := u.Query().Get("code")

	// Redeem code to verify the subject returned is sub_existing_123
	ac, err := store.RedeemCode(tenant.DefaultTenantID, idpCode, "app_client", "https://myclient.com/oauth2/callback", "")
	if err != nil {
		t.Fatalf("failed to redeem code issued for existing social user: %v", err)
	}

	if ac.Subject != "sub_existing_123" {
		t.Errorf("expected code subject sub_existing_123, got %s", ac.Subject)
	}
}

// TestEmpirical_1_GitHubPrivateEmailFallback verifies GET /user/emails fallback mechanisms.
func TestEmpirical_1_GitHubPrivateEmailFallback(t *testing.T) {
	t.Run("Primary email selected from /user/emails when /user email is null", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/token":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "gh_tok_1"})
			case "/user":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": 8801, "email": nil, "name": "Dev One"})
			case "/user/emails":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode([]map[string]interface{}{
					{"email": "secondary@example.com", "primary": false},
					{"email": "primary@example.com", "primary": true},
				})
			default:
				http.NotFound(w, r)
			}
		}))
		defer ts.Close()

		cfg := SocialConfig{ClientID: "gh_id", ClientSecret: "gh_sec", TokenURL: ts.URL + "/token", UserInfoURL: ts.URL + "/user"}
		u, err := ExchangeSocialCode(context.Background(), ProviderGitHub, cfg, "code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Email != "primary@example.com" {
			t.Errorf("expected primary@example.com, got %s", u.Email)
		}
	})

	t.Run("First secondary email selected when no email is primary: true", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/token":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "gh_tok_2"})
			case "/user":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": 8802, "email": ""})
			case "/user/emails":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode([]map[string]interface{}{
					{"email": "first_sec@example.com", "primary": false},
					{"email": "second_sec@example.com", "primary": false},
				})
			default:
				http.NotFound(w, r)
			}
		}))
		defer ts.Close()

		cfg := SocialConfig{ClientID: "gh_id", ClientSecret: "gh_sec", TokenURL: ts.URL + "/token", UserInfoURL: ts.URL + "/user"}
		u, err := ExchangeSocialCode(context.Background(), ProviderGitHub, cfg, "code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Email != "first_sec@example.com" {
			t.Errorf("expected first_sec@example.com, got %s", u.Email)
		}
	})
}

// TestEmpirical_2_GitHubLoginHandleFallback verifies login handle fallback when name is empty.
func TestEmpirical_2_GitHubLoginHandleFallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "gh_tok_3"})
		case "/user":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    9901,
				"name":  nil,
				"login": "octocat_handle",
				"email": "octo@example.com",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := SocialConfig{ClientID: "gh_id", ClientSecret: "gh_sec", TokenURL: ts.URL + "/token", UserInfoURL: ts.URL + "/user"}
	u, err := ExchangeSocialCode(context.Background(), ProviderGitHub, cfg, "code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "octocat_handle" {
		t.Errorf("expected name to fall back to login 'octocat_handle', got %s", u.Name)
	}
}

// TestEmpirical_3_UserProvisioningAndUsernameCollision verifies user provisioning and uncovers username collision edge case.
func TestEmpirical_3_UserProvisioningAndUsernameCollision(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "tok_prov"})
		case "/userinfo":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    "g_prov_1",
				"email": "new_social_user@example.com",
				"name":  "CollidingUsername",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	store := NewAuthzStore()

	// Pre-create local user with username "CollidingUsername" but DIFFERENT email
	store.AddUser(tenant.DefaultTenantID, &User{
		Username: "CollidingUsername",
		Email:    "existing_local@example.com",
		Password: "password",
		Subject:  "usr_local_existing",
	})

	client := &OAuthClient{ClientID: "app_prov", RedirectURI: "https://app.com/cb", Name: "Prov App"}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("secret_key_32_bytes_long_1234567"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_id",
			ClientSecret: "g_sec",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
			TokenURL:     ts.URL + "/token",
			UserInfoURL:  ts.URL + "/userinfo",
		},
	}

	router := chi.NewRouter()
	router.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	stateStr, _ := h.signSocialState(socialStateClaims{
		ClientID: "app_prov", RedirectURI: "https://app.com/cb", Provider: "google",
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/social/google/callback?code=code123&state="+url.QueryEscape(stateStr), nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", rec.Code)
	}

	loc := rec.Header().Get("Location")
	u, _ := url.Parse(loc)
	code := u.Query().Get("code")

	ac, err := store.RedeemCode(tenant.DefaultTenantID, code, "app_prov", "https://app.com/cb", "")
	if err != nil {
		t.Fatalf("failed to redeem code: %v", err)
	}

	// Code was issued for ac.Subject
	// Look up user by subject in store
	userInStore, lookupErr := store.GetUserBySubject(tenant.DefaultTenantID, ac.Subject)
	if lookupErr != nil {
		t.Fatalf("expected user with subject %s to be stored in AuthzStore after fallback registration, got error: %v", ac.Subject, lookupErr)
	}
	if !strings.HasPrefix(userInStore.Username, "CollidingUsername_") {
		t.Errorf("expected fallback username with suffix, got %s", userInStore.Username)
	}
}

// TestEmpirical_4_PKCEPropagationAndVerification verifies PKCE challenge parameter propagation and redemption.
func TestEmpirical_4_PKCEPropagationAndVerification(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "tok_pkce"})
		case "/userinfo":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    "g_pkce_1",
				"email": "pkce_user@example.com",
				"name":  "PKCE User",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	store := NewAuthzStore()
	client := &OAuthClient{ClientID: "pkce_client", RedirectURI: "https://pkce.example.com/cb", Name: "PKCE Client"}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("secret_key_32_bytes_long_1234567"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_id",
			ClientSecret: "g_sec",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
			TokenURL:     ts.URL + "/token",
			UserInfoURL:  ts.URL + "/userinfo",
		},
	}

	router := chi.NewRouter()
	router.Get("/auth/social/{provider}", h.SocialLoginGET)
	router.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	hSum := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hSum[:])

	// 1. Social login GET with PKCE
	reqLogin := httptest.NewRequest("GET", "/auth/social/google?client_id=pkce_client&redirect_uri=https://pkce.example.com/cb&code_challenge="+codeChallenge+"&code_challenge_method=S256", nil)
	recLogin := httptest.NewRecorder()
	router.ServeHTTP(recLogin, reqLogin)

	if recLogin.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", recLogin.Code)
	}

	locLogin := recLogin.Header().Get("Location")
	uLogin, _ := url.Parse(locLogin)
	signedState := uLogin.Query().Get("state")

	// 2. Callback from provider
	reqCB := httptest.NewRequest("GET", "/auth/social/google/callback?code=mock_code&state="+url.QueryEscape(signedState), nil)
	recCB := httptest.NewRecorder()
	router.ServeHTTP(recCB, reqCB)

	if recCB.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", recCB.Code)
	}

	locCB := recCB.Header().Get("Location")
	uCB, _ := url.Parse(locCB)
	idpCode := uCB.Query().Get("code")

	// 3. Redeem code with wrong verifier -> fail
	_, errWrong := store.RedeemCode(tenant.DefaultTenantID, idpCode, "pkce_client", "https://pkce.example.com/cb", "wrong_verifier")
	if errWrong == nil {
		t.Errorf("expected RedeemCode with wrong verifier to fail, but succeeded")
	}

	// 4. Redeem code with correct verifier -> pass
	ac, errRight := store.RedeemCode(tenant.DefaultTenantID, idpCode, "pkce_client", "https://pkce.example.com/cb", codeVerifier)
	if errRight != nil {
		t.Errorf("expected RedeemCode with correct verifier to succeed, got %v", errRight)
	}
	if ac.CodeChallenge != codeChallenge {
		t.Errorf("expected code challenge %s, got %s", codeChallenge, ac.CodeChallenge)
	}
}

// TestEmpirical_5_ProviderErrorCallbackPropagation uncovers bug where provider error callback is not propagated to client.
func TestEmpirical_5_ProviderErrorCallbackPropagation(t *testing.T) {
	store := NewAuthzStore()
	client := &OAuthClient{ClientID: "err_client", RedirectURI: "https://client.example.com/cb", Name: "Err Client"}
	store.AddClient(tenant.DefaultTenantID, client)

	h := &AuthzHandlers{
		Store:         store,
		ConsentSecret: []byte("secret_key_32_bytes_long_1234567"),
		GoogleConfig: SocialConfig{
			ClientID:     "g_id",
			ClientSecret: "g_sec",
			RedirectURI:  "https://auth.example.com/auth/social/google/callback",
		},
	}

	router := chi.NewRouter()
	router.Get("/auth/social/{provider}/callback", h.SocialCallbackGET)

	signedState, _ := h.signSocialState(socialStateClaims{
		ClientID: "err_client", RedirectURI: "https://client.example.com/cb", ClientState: "app_state_999", Provider: "google",
	})

	// Provider callback with error=access_denied (no code parameter)
	errURL := "/auth/social/google/callback?error=access_denied&error_description=User+denied+access&state=" + url.QueryEscape(signedState)
	req := httptest.NewRequest("GET", errURL, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect for provider error propagation, got status %d, body: %s", rec.Code, rec.Body.String())
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://client.example.com/cb?") {
		t.Errorf("expected redirect location starting with https://client.example.com/cb?, got %s", loc)
	}
	u, _ := url.Parse(loc)
	if u.Query().Get("error") != "access_denied" {
		t.Errorf("expected error=access_denied, got %s", u.Query().Get("error"))
	}
	if u.Query().Get("error_description") != "User denied access" {
		t.Errorf("expected error_description 'User denied access', got %s", u.Query().Get("error_description"))
	}
	if u.Query().Get("state") != "app_state_999" {
		t.Errorf("expected state 'app_state_999', got %s", u.Query().Get("state"))
	}
}

