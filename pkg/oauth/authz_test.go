package oauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
)

func newAuthzStore() *oauth.AuthzStore {
	s := oauth.NewAuthzStore()
	s.AddUser(&oauth.DemoUser{Username: "alice", Password: "password", Subject: "alice-sub"})
	s.AddClient(&oauth.DemoClient{ClientID: "demo-client", RedirectURI: "http://localhost:8080/callback", Name: "Demo"})
	return s
}

// --- AuthzStore tests ---

func TestAuthzStore_AuthenticateUser(t *testing.T) {
	s := newAuthzStore()

	u, err := s.AuthenticateUser("alice", "password")
	if err != nil {
		t.Fatalf("AuthenticateUser() error = %v", err)
	}
	if u.Subject != "alice-sub" {
		t.Errorf("Subject = %q, want %q", u.Subject, "alice-sub")
	}

	_, err = s.AuthenticateUser("alice", "wrong")
	if err == nil {
		t.Error("AuthenticateUser() with wrong password should fail")
	}

	_, err = s.AuthenticateUser("nobody", "password")
	if err == nil {
		t.Error("AuthenticateUser() with unknown user should fail")
	}
}

func TestAuthzStore_ValidateClient(t *testing.T) {
	s := newAuthzStore()

	c, err := s.ValidateClient("demo-client", "http://localhost:8080/callback")
	if err != nil {
		t.Fatalf("ValidateClient() error = %v", err)
	}
	if c.Name != "Demo" {
		t.Errorf("Name = %q, want %q", c.Name, "Demo")
	}

	_, err = s.ValidateClient("demo-client", "http://evil.com/callback")
	if err == nil {
		t.Error("ValidateClient() with wrong redirect_uri should fail")
	}

	_, err = s.ValidateClient("unknown", "http://localhost:8080/callback")
	if err == nil {
		t.Error("ValidateClient() with unknown client should fail")
	}
}

func TestAuthzStore_CreateRedeemCode(t *testing.T) {
	s := newAuthzStore()

	code, err := s.CreateCode("alice-sub", "demo-client", "http://localhost:8080/callback", "openid")
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}

	ac, err := s.RedeemCode(code, "demo-client", "http://localhost:8080/callback")
	if err != nil {
		t.Fatalf("RedeemCode() error = %v", err)
	}
	if ac.Subject != "alice-sub" {
		t.Errorf("Subject = %q, want %q", ac.Subject, "alice-sub")
	}
}

func TestAuthzStore_RedeemCode_AlreadyUsed(t *testing.T) {
	s := newAuthzStore()
	code, _ := s.CreateCode("alice-sub", "demo-client", "http://localhost:8080/callback", "")
	s.RedeemCode(code, "demo-client", "http://localhost:8080/callback")

	_, err := s.RedeemCode(code, "demo-client", "http://localhost:8080/callback")
	if err == nil {
		t.Error("RedeemCode() should fail on reuse")
	}
}

func TestAuthzStore_RedeemCode_WrongClient(t *testing.T) {
	s := newAuthzStore()
	code, _ := s.CreateCode("alice-sub", "demo-client", "http://localhost:8080/callback", "")

	_, err := s.RedeemCode(code, "wrong-client", "http://localhost:8080/callback")
	if err == nil {
		t.Error("RedeemCode() with wrong client should fail")
	}
}

func TestAuthzStore_RedeemCode_InvalidCode(t *testing.T) {
	s := newAuthzStore()
	_, err := s.RedeemCode("nonexistent", "demo-client", "http://localhost:8080/callback")
	if err == nil {
		t.Error("RedeemCode() with invalid code should fail")
	}
}

// --- Handler tests ---

func newHandlers(t *testing.T) (*oauth.AuthzHandlers, *transport.SigStore) {
	t.Helper()
	p := newProvider(t)
	store := transport.NewSigStore()
	authzStore := newAuthzStore()
	h := &oauth.AuthzHandlers{Provider: p, Store: authzStore, SigStore: store}
	return h, store
}

func TestAuthorizeGET_ReturnsForm(t *testing.T) {
	h, _ := newHandlers(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		"/authorize?client_id=demo-client&redirect_uri=http://localhost:8080/callback&state=xyz&scope=openid", nil)
	h.AuthorizeGET(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "demo-client") {
		t.Error("form should contain client_id")
	}
	if !strings.Contains(body, "xyz") {
		t.Error("form should contain state")
	}
}

func TestAuthorizeGET_MissingParams(t *testing.T) {
	h, _ := newHandlers(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/authorize", nil)
	h.AuthorizeGET(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthorizePOST_RedirectsWithCode(t *testing.T) {
	h, _ := newHandlers(t)

	form := url.Values{
		"username":     {"alice"},
		"password":     {"password"},
		"client_id":    {"demo-client"},
		"redirect_uri": {"http://localhost:8080/callback"},
		"state":        {"xyz"},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "http://localhost:8080/callback?code=") {
		t.Errorf("Location = %q, want prefix with callback?code=", loc)
	}
	if !strings.Contains(loc, "state=xyz") {
		t.Error("redirect should include state parameter")
	}
}

func TestAuthorizePOST_BadCredentials(t *testing.T) {
	h, _ := newHandlers(t)

	form := url.Values{
		"username":     {"alice"},
		"password":     {"wrong"},
		"client_id":    {"demo-client"},
		"redirect_uri": {"http://localhost:8080/callback"},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTokenExchange_ValidCode(t *testing.T) {
	h, _ := newHandlers(t)

	code, err := h.Store.CreateCode("alice-sub", "demo-client", "http://localhost:8080/callback", "openid")
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}

	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"client_id":    {"demo-client"},
		"redirect_uri": {"http://localhost:8080/callback"},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["access_token"] == "" {
		t.Error("access_token is empty")
	}
	if resp["token_type"] != "Bearer" {
		t.Errorf("token_type = %q, want %q", resp["token_type"], "Bearer")
	}
}

func TestTokenExchange_InvalidCode(t *testing.T) {
	h, _ := newHandlers(t)

	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {"invalid-code"},
		"client_id":    {"demo-client"},
		"redirect_uri": {"http://localhost:8080/callback"},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTokenExchange_UnsupportedGrant(t *testing.T) {
	h, _ := newHandlers(t)

	form := url.Values{
		"grant_type": {"client_credentials"},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// --- End-to-end flow test ---

func TestAuthzFlow_EndToEnd(t *testing.T) {
	h, _ := newHandlers(t)

	// Step 1: GET /authorize → login form.
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet,
		"/authorize?client_id=demo-client&redirect_uri=http://localhost:8080/callback&response_type=code&state=xyz", nil)
	h.AuthorizeGET(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("GET /authorize status = %d", rec1.Code)
	}

	// Step 2: POST /authorize → redirect with code.
	form := url.Values{
		"username":     {"alice"},
		"password":     {"password"},
		"client_id":    {"demo-client"},
		"redirect_uri": {"http://localhost:8080/callback"},
		"state":        {"xyz"},
	}
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec2, req2)
	if rec2.Code != http.StatusFound {
		t.Fatalf("POST /authorize status = %d", rec2.Code)
	}

	// Extract code from redirect URL.
	loc, _ := url.Parse(rec2.Header().Get("Location"))
	code := loc.Query().Get("code")
	if code == "" {
		t.Fatal("no code in redirect URL")
	}

	// Step 3: POST /token → exchange code for token.
	tokenForm := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"client_id":    {"demo-client"},
		"redirect_uri": {"http://localhost:8080/callback"},
	}
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(tokenForm.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("POST /token status = %d, body = %s", rec3.Code, rec3.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rec3.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding token response: %v", err)
	}
	if resp["access_token"] == "" {
		t.Error("access_token is empty")
	}

	// Step 4: Replay same code → should fail.
	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(tokenForm.Encode()))
	req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec4, req4)
	if rec4.Code != http.StatusBadRequest {
		t.Errorf("replayed code should fail, got status = %d", rec4.Code)
	}
}

// Suppress unused warning for time import.
var _ = time.Now
