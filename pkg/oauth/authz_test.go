package oauth_test

import (
	"bytes"
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

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/developer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

// testPKCEVerifier/testPKCEChallenge are a matched S256 PKCE pair used by
// tests exercising the public "demo-client" fixture, which (having no
// client_secret) is now required to use PKCE on every authorization request.
const testPKCEVerifier = "test-code-verifier-1234567890-abcdefghij"

func testPKCEChallenge() string {
	h := sha256.Sum256([]byte(testPKCEVerifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

const testTenant = tenant.DefaultTenantID

func newAuthzStore() oauth.AuthzStore {
	s := oauth.NewAuthzStore()
	s.AddUser(testTenant, &oauth.User{Username: "alice", Password: "password", Subject: "alice-sub"})
	s.AddClient(testTenant, &oauth.OAuthClient{ClientID: "demo-client", RedirectURI: "http://localhost:8080/callback", Name: "Demo"})
	return s
}

// --- AuthzStore tests ---

func TestAuthzStore_AuthenticateUser(t *testing.T) {
	s := newAuthzStore()

	u, err := s.AuthenticateUser(testTenant, "alice", "password")
	if err != nil {
		t.Fatalf("AuthenticateUser() error = %v", err)
	}
	if u.Subject != "alice-sub" {
		t.Errorf("Subject = %q, want %q", u.Subject, "alice-sub")
	}

	_, err = s.AuthenticateUser(testTenant, "alice", "wrong")
	if err == nil {
		t.Error("AuthenticateUser() with wrong password should fail")
	}

	_, err = s.AuthenticateUser(testTenant, "nobody", "password")
	if err == nil {
		t.Error("AuthenticateUser() with unknown user should fail")
	}
}

func TestAuthzStore_AuthenticateUser_TenantIsolation(t *testing.T) {
	s := newAuthzStore()
	// alice exists in testTenant but not in "other-tenant"
	_, err := s.AuthenticateUser("other-tenant", "alice", "password")
	if err == nil {
		t.Error("AuthenticateUser() should fail for unknown tenant")
	}
}

func TestAuthzStore_GetUserBySubject(t *testing.T) {
	s := newAuthzStore()
	u, err := s.GetUserBySubject(testTenant, "alice-sub")
	if err != nil {
		t.Fatalf("GetUserBySubject() error = %v", err)
	}
	if u.Username != "alice" {
		t.Errorf("Username = %q, want %q", u.Username, "alice")
	}

	if _, err := s.GetUserBySubject(testTenant, "nonexistent-sub"); err == nil {
		t.Error("GetUserBySubject() with unknown subject should fail")
	}
	if _, err := s.GetUserBySubject("other-tenant", "alice-sub"); err == nil {
		t.Error("GetUserBySubject() should fail for the wrong tenant")
	}
}

func TestAuthzStore_GetUserByEmail(t *testing.T) {
	s := oauth.NewAuthzStore()
	if err := s.RegisterUser(testTenant, &oauth.User{Username: "carol", Email: "carol@example.com", Password: "supersecret123"}); err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	u, err := s.GetUserByEmail(testTenant, "carol@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}
	if u.Username != "carol" {
		t.Errorf("Username = %q, want %q", u.Username, "carol")
	}

	if _, err := s.GetUserByEmail(testTenant, "nobody@example.com"); err == nil {
		t.Error("GetUserByEmail() with unknown email should fail")
	}
}

func TestAuthzStore_SetPassword(t *testing.T) {
	s := oauth.NewAuthzStore()
	if err := s.RegisterUser(testTenant, &oauth.User{Username: "dave", Email: "dave@example.com", Password: "old-password1"}); err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte("new-password1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	if err := s.SetPassword(testTenant, "dave", string(newHash)); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	if _, err := s.AuthenticateUser(testTenant, "dave", "old-password1"); err == nil {
		t.Error("old password should no longer authenticate after SetPassword")
	}
	if _, err := s.AuthenticateUser(testTenant, "dave", "new-password1"); err != nil {
		t.Errorf("new password should authenticate after SetPassword: %v", err)
	}

	if err := s.SetPassword(testTenant, "nonexistent-user", string(newHash)); err == nil {
		t.Error("SetPassword() for unknown user should fail")
	}
}

func TestAuthzStore_ValidateClient(t *testing.T) {
	s := newAuthzStore()

	c, err := s.ValidateClient(testTenant, "demo-client", "http://localhost:8080/callback")
	if err != nil {
		t.Fatalf("ValidateClient() error = %v", err)
	}
	if c.Name != "Demo" {
		t.Errorf("Name = %q, want %q", c.Name, "Demo")
	}

	_, err = s.ValidateClient(testTenant, "demo-client", "http://evil.com/callback")
	if err == nil {
		t.Error("ValidateClient() with wrong redirect_uri should fail")
	}

	_, err = s.ValidateClient(testTenant, "unknown", "http://localhost:8080/callback")
	if err == nil {
		t.Error("ValidateClient() with unknown client should fail")
	}
}

func TestAuthzStore_CreateRedeemCode(t *testing.T) {
	s := newAuthzStore()

	code, err := s.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "openid", "", "")
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}

	ac, err := s.RedeemCode(testTenant, code, "demo-client", "http://localhost:8080/callback", "")
	if err != nil {
		t.Fatalf("RedeemCode() error = %v", err)
	}
	if ac.Subject != "alice-sub" {
		t.Errorf("Subject = %q, want %q", ac.Subject, "alice-sub")
	}
}

func TestAuthzStore_RedeemCode_AlreadyUsed(t *testing.T) {
	s := newAuthzStore()
	code, _ := s.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "", "", "")
	s.RedeemCode(testTenant, code, "demo-client", "http://localhost:8080/callback", "")

	_, err := s.RedeemCode(testTenant, code, "demo-client", "http://localhost:8080/callback", "")
	if err == nil {
		t.Error("RedeemCode() should fail on reuse")
	}
}

func TestAuthzStore_RedeemCode_WrongClient(t *testing.T) {
	s := newAuthzStore()
	code, _ := s.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "", "", "")

	_, err := s.RedeemCode(testTenant, code, "wrong-client", "http://localhost:8080/callback", "")
	if err == nil {
		t.Error("RedeemCode() with wrong client should fail")
	}
}

func TestAuthzStore_RedeemCode_InvalidCode(t *testing.T) {
	s := newAuthzStore()
	_, err := s.RedeemCode(testTenant, "nonexistent", "demo-client", "http://localhost:8080/callback", "")
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

// newHandlersWithVerifier is like newHandlers but also wires up a real
// Verifier + RevocationStore, needed to exercise /revoke and /verify/jwt.
func newHandlersWithVerifier(t *testing.T) (*oauth.AuthzHandlers, *transport.SigStore) {
	t.Helper()
	h, store := newHandlers(t)
	v := verifier.NewVerifierWithKeyPair(h.Provider.KeyPair, store)
	rev := oauth.NewMemoryRevocationStore()
	v.Revocation = rev
	h.Revocation = rev
	h.Verifier = &verifier.OAuthAdapter{Verifier: v}
	return h, store
}

// injectTenant injects the default tenant ID into a request context.
func injectTenant(r *http.Request) *http.Request {
	return r.WithContext(tenant.WithTenant(r.Context(), testTenant))
}

func TestAuthorizeGET_ReturnsForm(t *testing.T) {
	h, _ := newHandlers(t)

	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodGet,
		"/authorize?client_id=demo-client&redirect_uri=http://localhost:8080/callback&state=xyz&scope=openid", nil))
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
	req := injectTenant(httptest.NewRequest(http.MethodGet, "/authorize", nil))
	h.AuthorizeGET(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthorizePOST_RedirectsWithCode(t *testing.T) {
	h, _ := newHandlers(t)

	form := url.Values{
		"username":       {"alice"},
		"password":       {"password"},
		"client_id":      {"demo-client"},
		"redirect_uri":   {"http://localhost:8080/callback"},
		"state":          {"xyz"},
		"code_challenge": {testPKCEChallenge()},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d, body/location=%v", rec.Code, http.StatusFound, rec.Header().Get("Location"))
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
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	// Bad credentials redirect back to the login form (303 See Other).
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

// demo-client has no client_secret (ClientSecretHash == ""), making it a
// public client. Public clients must use PKCE — omitting code_challenge
// must be rejected rather than silently issuing a code interceptable
// without proof of possession.
// --- Scope validation & consent screen tests ---

func newHandlersWithConsent(t *testing.T) *oauth.AuthzHandlers {
	t.Helper()
	h, _ := newHandlers(t)
	h.GrantStore = oauth.NewMemoryGrantStore()
	h.ConsentSecret = bytes.Repeat([]byte{0x02}, 32)
	return h
}

func authorizePOSTForm(scope string) url.Values {
	return url.Values{
		"username":       {"alice"},
		"password":       {"password"},
		"client_id":      {"demo-client"},
		"redirect_uri":   {"http://localhost:8080/callback"},
		"state":          {"xyz"},
		"scope":          {scope},
		"code_challenge": {testPKCEChallenge()},
	}
}

func TestAuthorizePOST_InvalidScope_Rejected(t *testing.T) {
	h, _ := newHandlers(t)

	form := authorizePOSTForm("openid not-a-real-scope")
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if !strings.Contains(rec.Header().Get("Location"), "error=invalid_scope") {
		t.Errorf("Location = %q, want an invalid_scope error", rec.Header().Get("Location"))
	}
}

func TestAuthorizePOST_NoGrantStore_SkipsConsentEntirely(t *testing.T) {
	h, _ := newHandlers(t) // GrantStore left nil

	form := authorizePOSTForm("openid email")
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d (no GrantStore configured, consent should be skipped)", rec.Code, http.StatusFound)
	}
}

func TestAuthorizePOST_FirstTime_ShowsConsentScreen(t *testing.T) {
	h := newHandlersWithConsent(t)

	form := authorizePOSTForm("openid email")
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (consent screen), body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Demo") {
		t.Error("consent page should show the client's app name")
	}
	if !strings.Contains(rec.Body.String(), `name="ticket"`) {
		t.Error("consent page should carry a signed ticket hidden field")
	}
}

func TestConsentPOST_Allow_IssuesCodeAndSavesGrant(t *testing.T) {
	h := newHandlersWithConsent(t)

	// Step 1: reach the consent screen and extract the ticket.
	form := authorizePOSTForm("openid email")
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected consent screen, status = %d", rec.Code)
	}
	ticket := extractTicket(t, rec.Body.String())

	// Step 2: approve.
	consentForm := url.Values{"ticket": {ticket}, "decision": {"allow"}}
	consentRec := httptest.NewRecorder()
	consentReq := injectTenant(httptest.NewRequest(http.MethodPost, "/consent", strings.NewReader(consentForm.Encode())))
	consentReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ConsentPOST(consentRec, consentReq)

	if consentRec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d, body=%s", consentRec.Code, http.StatusFound, consentRec.Body.String())
	}
	loc := consentRec.Header().Get("Location")
	if !strings.HasPrefix(loc, "http://localhost:8080/callback?code=") {
		t.Errorf("Location = %q, want a code redirect", loc)
	}

	grant, err := h.GrantStore.Get(testTenant, "alice-sub", "demo-client")
	if err != nil || grant == nil {
		t.Fatalf("Get() grant error = %v, grant = %v", err, grant)
	}
	if !strings.Contains(grant.Scope, "openid") || !strings.Contains(grant.Scope, "email") {
		t.Errorf("grant.Scope = %q, want it to contain openid and email", grant.Scope)
	}
}

func TestConsentPOST_Deny_RedirectsWithAccessDenied(t *testing.T) {
	h := newHandlersWithConsent(t)

	form := authorizePOSTForm("openid")
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)
	ticket := extractTicket(t, rec.Body.String())

	consentForm := url.Values{"ticket": {ticket}, "decision": {"deny"}}
	consentRec := httptest.NewRecorder()
	consentReq := injectTenant(httptest.NewRequest(http.MethodPost, "/consent", strings.NewReader(consentForm.Encode())))
	consentReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ConsentPOST(consentRec, consentReq)

	if consentRec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", consentRec.Code, http.StatusFound)
	}
	loc := consentRec.Header().Get("Location")
	if !strings.HasPrefix(loc, "http://localhost:8080/callback?error=access_denied") {
		t.Errorf("Location = %q, want an access_denied redirect to the client's redirect_uri", loc)
	}

	if grant, _ := h.GrantStore.Get(testTenant, "alice-sub", "demo-client"); grant != nil {
		t.Error("denying consent must not save a grant")
	}
}

func TestAuthorizePOST_RepeatLogin_SkipsConsentWhenScopeAlreadyGranted(t *testing.T) {
	h := newHandlersWithConsent(t)

	// First login: consent, then approve.
	form := authorizePOSTForm("openid email")
	rec1 := httptest.NewRecorder()
	req1 := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec1, req1)
	ticket := extractTicket(t, rec1.Body.String())

	consentForm := url.Values{"ticket": {ticket}, "decision": {"allow"}}
	consentRec := httptest.NewRecorder()
	consentReq := injectTenant(httptest.NewRequest(http.MethodPost, "/consent", strings.NewReader(consentForm.Encode())))
	consentReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ConsentPOST(consentRec, consentReq)
	if consentRec.Code != http.StatusFound {
		t.Fatalf("first consent approval status = %d", consentRec.Code)
	}

	// Second login, same scope: must skip straight to a code redirect.
	rec2 := httptest.NewRecorder()
	req2 := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec2, req2)

	if rec2.Code != http.StatusFound {
		t.Fatalf("repeat login status = %d, want %d (consent should be skipped)", rec2.Code, http.StatusFound)
	}
}

func TestConsentPOST_ForgedTicket_Rejected(t *testing.T) {
	h := newHandlersWithConsent(t)

	consentForm := url.Values{"ticket": {"not-a-valid-jwt"}, "decision": {"allow"}}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/consent", strings.NewReader(consentForm.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ConsentPOST(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// extractTicket pulls the hidden "ticket" field's value out of the consent
// page's rendered HTML.
func extractTicket(t *testing.T, body string) string {
	t.Helper()
	const marker = `name="ticket" value="`
	i := strings.Index(body, marker)
	if i == -1 {
		t.Fatalf("ticket field not found in consent page body: %s", body)
	}
	rest := body[i+len(marker):]
	j := strings.Index(rest, `"`)
	if j == -1 {
		t.Fatalf("malformed ticket field in consent page body")
	}
	return rest[:j]
}

func TestAuthorizePOST_PublicClientWithoutPKCE_Rejected(t *testing.T) {
	h, _ := newHandlers(t)

	form := url.Values{
		"username":     {"alice"},
		"password":     {"password"},
		"client_id":    {"demo-client"},
		"redirect_uri": {"http://localhost:8080/callback"},
		// no code_challenge
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.AuthorizePOST(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d (redirect back to login with error)", rec.Code, http.StatusSeeOther)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "error=") {
		t.Errorf("Location = %q, want an error query parameter", loc)
	}
}

func TestRegisterPOST_UsesStoreByDefault(t *testing.T) {
	h, _ := newHandlers(t)

	form := url.Values{
		"username":       {"erin"},
		"email":          {"erin@example.com"},
		"password":       {"supersecret123"},
		"client_id":      {"demo-client"},
		"redirect_uri":   {"http://localhost:8080/callback"},
		"code_challenge": {testPKCEChallenge()},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.RegisterPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d, location=%v", rec.Code, http.StatusFound, rec.Header().Get("Location"))
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "http://localhost:8080/callback?code=") {
		t.Errorf("Location = %q, want prefix with callback?code=", loc)
	}

	// Registered directly into h.Store since no UserRegistrar was set.
	if _, err := h.Store.AuthenticateUser(testTenant, "erin", "supersecret123"); err != nil {
		t.Errorf("registered user should authenticate against h.Store: %v", err)
	}
}

// fakeUserRegistrar mimics *userstore.PostgresUserStore's RegisterForSubject
// adapter for tests, without needing a real database.
type fakeUserRegistrar struct {
	registered map[string]string // username -> subject
}

func (f *fakeUserRegistrar) RegisterForSubject(_ context.Context, _, username, _, _ string) (string, error) {
	subject := "usr_" + username
	if f.registered == nil {
		f.registered = make(map[string]string)
	}
	f.registered[username] = subject
	return subject, nil
}

// RegisterPOST must use UserRegistrar (e.g. Postgres) when configured,
// instead of always falling back to the in-memory/Redis Store — otherwise
// registration and authentication would silently target different backends.
func TestRegisterPOST_UsesUserRegistrarWhenSet(t *testing.T) {
	h, _ := newHandlers(t)
	registrar := &fakeUserRegistrar{}
	h.UserRegistrar = registrar

	form := url.Values{
		"username":       {"frank"},
		"email":          {"frank@example.com"},
		"password":       {"supersecret123"},
		"client_id":      {"demo-client"},
		"redirect_uri":   {"http://localhost:8080/callback"},
		"code_challenge": {testPKCEChallenge()},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.RegisterPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d, location=%v", rec.Code, http.StatusFound, rec.Header().Get("Location"))
	}
	if registrar.registered["frank"] != "usr_frank" {
		t.Fatalf("RegisterForSubject was not called via UserRegistrar; registered = %v", registrar.registered)
	}

	// The auth code must carry the subject returned by UserRegistrar
	// ("usr_frank"), not the raw username — otherwise the issued token's
	// subject wouldn't match what UserStore.AuthenticateForSubject/
	// GetBySubject would look up.
	loc := rec.Header().Get("Location")
	code := loc[strings.Index(loc, "code=")+len("code="):]
	ac, err := h.Store.RedeemCode(testTenant, code, "demo-client", "http://localhost:8080/callback", testPKCEVerifier)
	if err != nil {
		t.Fatalf("RedeemCode() error = %v", err)
	}
	if ac.Subject != "usr_frank" {
		t.Errorf("AuthCode.Subject = %q, want %q", ac.Subject, "usr_frank")
	}

	// h.Store must NOT have registered "frank" — registration went to the registrar only.
	if _, err := h.Store.AuthenticateUser(testTenant, "frank", "supersecret123"); err == nil {
		t.Error("h.Store should not have a record for a user registered via UserRegistrar")
	}
}

// --- Password reset tests ---

func newHandlersWithOTP(t *testing.T) *oauth.AuthzHandlers {
	t.Helper()
	h, _ := newHandlers(t)
	h.OTPStore = developer.NewOTPStore(nil) // nil mailer logs to stdout in tests
	return h
}

func TestForgotPasswordPOST_NoEnumerationSignal(t *testing.T) {
	h := newHandlersWithOTP(t)
	if err := h.Store.RegisterUser(testTenant, &oauth.User{Username: "grace", Email: "grace@example.com", Password: "supersecret123"}); err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	form := url.Values{"email": {"grace@example.com"}}
	recExisting := httptest.NewRecorder()
	reqExisting := injectTenant(httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode())))
	reqExisting.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ForgotPasswordPOST(recExisting, reqExisting)

	form2 := url.Values{"email": {"nobody@example.com"}}
	recMissing := httptest.NewRecorder()
	reqMissing := injectTenant(httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form2.Encode())))
	reqMissing.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ForgotPasswordPOST(recMissing, reqMissing)

	if recExisting.Code != recMissing.Code {
		t.Errorf("status differs between existing (%d) and nonexistent (%d) email — enumeration signal", recExisting.Code, recMissing.Code)
	}
	if recExisting.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", recExisting.Code, http.StatusSeeOther)
	}
}

func TestResetPasswordPOST_ValidCode_ChangesPassword(t *testing.T) {
	h := newHandlersWithOTP(t)
	if err := h.Store.RegisterUser(testTenant, &oauth.User{Username: "henry", Email: "henry@example.com", Password: "old-password1"}); err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	code, _, err := h.OTPStore.GenerateCode("henry@example.com")
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	form := url.Values{
		"email":    {"henry@example.com"},
		"code":     {code},
		"password": {"new-password1"},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ResetPasswordPOST(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d, location=%v", rec.Code, http.StatusSeeOther, rec.Header().Get("Location"))
	}
	if !strings.Contains(rec.Header().Get("Location"), "/authorize?") {
		t.Errorf("Location = %q, want redirect to /authorize", rec.Header().Get("Location"))
	}

	if _, err := h.Store.AuthenticateUser(testTenant, "henry", "old-password1"); err == nil {
		t.Error("old password should no longer authenticate")
	}
	if _, err := h.Store.AuthenticateUser(testTenant, "henry", "new-password1"); err != nil {
		t.Errorf("new password should authenticate: %v", err)
	}
}

func TestResetPasswordPOST_InvalidCode_Rejected(t *testing.T) {
	h := newHandlersWithOTP(t)
	if err := h.Store.RegisterUser(testTenant, &oauth.User{Username: "iris", Email: "iris@example.com", Password: "old-password1"}); err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	form := url.Values{
		"email":    {"iris@example.com"},
		"code":     {"000000"},
		"password": {"new-password1"},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ResetPasswordPOST(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if !strings.Contains(rec.Header().Get("Location"), "/reset-password?") {
		t.Errorf("Location = %q, want redirect back to /reset-password on failure", rec.Header().Get("Location"))
	}
	if _, err := h.Store.AuthenticateUser(testTenant, "iris", "old-password1"); err != nil {
		t.Error("password must not change when the OTP code is invalid")
	}
}

// A public client's code, even if somehow created without a PKCE challenge,
// must still be rejected at token exchange time (defense in depth).
func TestTokenExchange_PublicClientCodeWithoutPKCE_Rejected(t *testing.T) {
	h, _ := newHandlers(t)

	code, err := h.Store.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "openid", "", "")
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
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTokenExchange_ValidCode(t *testing.T) {
	h, _ := newHandlers(t)

	code, err := h.Store.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "openid", testPKCEChallenge(), "S256")
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {"demo-client"},
		"redirect_uri":  {"http://localhost:8080/callback"},
		"code_verifier": {testPKCEVerifier},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
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

// The access token issued by TokenExchange must carry the granted scope as
// a claim, so resource servers (and /userinfo) can gate claims by it.
func TestTokenExchange_AccessTokenCarriesScopeClaim(t *testing.T) {
	h, _ := newHandlers(t)

	code, err := h.Store.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "openid email", testPKCEChallenge(), "S256")
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {"demo-client"},
		"redirect_uri":  {"http://localhost:8080/callback"},
		"code_verifier": {testPKCEVerifier},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	parser := jwt.NewParser()
	tok, _, err := parser.ParseUnverified(resp["access_token"], &oauth.HybridClaims{})
	if err != nil {
		t.Fatalf("parsing access token: %v", err)
	}
	claims, ok := tok.Claims.(*oauth.HybridClaims)
	if !ok {
		t.Fatal("unexpected claims type")
	}
	if claims.Scope != "openid email" {
		t.Errorf("Scope claim = %q, want %q", claims.Scope, "openid email")
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
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
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
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// --- Refresh token tests ---

func newHandlersWithRefresh(t *testing.T) *oauth.AuthzHandlers {
	t.Helper()
	h, _ := newHandlers(t)
	h.RefreshTokens = oauth.NewMemoryRefreshTokenStore()
	return h
}

func exchangeAuthCode(t *testing.T, h *oauth.AuthzHandlers) map[string]string {
	t.Helper()
	// offline_access is required for a refresh token to be issued at all
	// (see issueTokens) — these tests are specifically exercising refresh
	// token behavior, so the code must carry that scope.
	code, err := h.Store.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "openid offline_access", testPKCEChallenge(), "S256")
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {"demo-client"},
		"redirect_uri":  {"http://localhost:8080/callback"},
		"code_verifier": {testPKCEVerifier},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authorization_code exchange status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	return resp
}

func refreshTokenRequest(t *testing.T, h *oauth.AuthzHandlers, refreshToken string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {"demo-client"},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)
	return rec
}

func TestTokenExchange_AuthCode_IssuesRefreshTokenWhenConfigured(t *testing.T) {
	h := newHandlersWithRefresh(t)
	resp := exchangeAuthCode(t, h)
	if resp["refresh_token"] == "" {
		t.Error("refresh_token should be issued when RefreshTokens is configured")
	}
}

func TestTokenExchange_AuthCode_NoRefreshTokenWhenNotConfigured(t *testing.T) {
	h, _ := newHandlers(t) // RefreshTokens left nil
	resp := exchangeAuthCode(t, h)
	if _, ok := resp["refresh_token"]; ok {
		t.Error("refresh_token should not be issued when RefreshTokens is not configured")
	}
}

// A refresh token must only be issued when the client requested
// offline_access — otherwise every login would silently grant long-lived
// refresh capability regardless of consent.
func TestTokenExchange_AuthCode_NoRefreshTokenWithoutOfflineAccessScope(t *testing.T) {
	h := newHandlersWithRefresh(t)

	code, err := h.Store.CreateCode(testTenant, "alice-sub", "demo-client", "http://localhost:8080/callback", "openid", testPKCEChallenge(), "S256")
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {"demo-client"},
		"redirect_uri":  {"http://localhost:8080/callback"},
		"code_verifier": {testPKCEVerifier},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if _, ok := resp["refresh_token"]; ok {
		t.Error("refresh_token should not be issued without an offline_access scope request")
	}
}

func TestTokenExchange_RefreshToken_RotatesToken(t *testing.T) {
	h := newHandlersWithRefresh(t)
	first := exchangeAuthCode(t, h)

	rec := refreshTokenRequest(t, h, first["refresh_token"])
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var second map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&second); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if second["access_token"] == "" || second["refresh_token"] == "" {
		t.Fatal("refresh exchange should return a new access_token and refresh_token")
	}
	if second["refresh_token"] == first["refresh_token"] {
		t.Error("refresh token must be rotated (a new value), not reused")
	}

	// The new refresh token must work.
	rec2 := refreshTokenRequest(t, h, second["refresh_token"])
	if rec2.Code != http.StatusOK {
		t.Errorf("second refresh with the rotated token should succeed, status = %d, body=%s", rec2.Code, rec2.Body.String())
	}
}

// Replaying an already-used refresh token must be treated as theft and
// revoke the entire token family, not just reject the one reused token.
func TestTokenExchange_RefreshToken_ReuseRevokesFamily(t *testing.T) {
	h := newHandlersWithRefresh(t)
	first := exchangeAuthCode(t, h)

	rec := refreshTokenRequest(t, h, first["refresh_token"])
	if rec.Code != http.StatusOK {
		t.Fatalf("first refresh status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var second map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&second)

	// Replay the already-used first refresh token.
	recReplay := refreshTokenRequest(t, h, first["refresh_token"])
	if recReplay.Code != http.StatusBadRequest {
		t.Errorf("replaying a used refresh token status = %d, want %d", recReplay.Code, http.StatusBadRequest)
	}

	// The legitimately-rotated second token must now ALSO be rejected,
	// since reuse revokes the whole family.
	recSecondAfterTheft := refreshTokenRequest(t, h, second["refresh_token"])
	if recSecondAfterTheft.Code != http.StatusBadRequest {
		t.Errorf("second token after family revocation status = %d, want %d", recSecondAfterTheft.Code, http.StatusBadRequest)
	}
}

func TestTokenExchange_RefreshToken_WrongClientRejected(t *testing.T) {
	h := newHandlersWithRefresh(t)
	first := exchangeAuthCode(t, h)

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {first["refresh_token"]},
		"client_id":     {"wrong-client"},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec, req)

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 401 or 400 for an unknown client", rec.Code)
	}
}

func TestTokenExchange_RefreshTokenGrant_DisabledWhenNotConfigured(t *testing.T) {
	h, _ := newHandlers(t) // RefreshTokens left nil

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"rt_whatever"},
		"client_id":     {"demo-client"},
	}
	rec := httptest.NewRecorder()
	req := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode())))
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
	req1 := injectTenant(httptest.NewRequest(http.MethodGet,
		"/authorize?client_id=demo-client&redirect_uri=http://localhost:8080/callback&response_type=code&state=xyz", nil))
	h.AuthorizeGET(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("GET /authorize status = %d", rec1.Code)
	}

	// Step 2: POST /authorize → redirect with code.
	form := url.Values{
		"username":       {"alice"},
		"password":       {"password"},
		"client_id":      {"demo-client"},
		"redirect_uri":   {"http://localhost:8080/callback"},
		"state":          {"xyz"},
		"code_challenge": {testPKCEChallenge()},
	}
	rec2 := httptest.NewRecorder()
	req2 := injectTenant(httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(form.Encode())))
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
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {"demo-client"},
		"redirect_uri":  {"http://localhost:8080/callback"},
		"code_verifier": {testPKCEVerifier},
	}
	rec3 := httptest.NewRecorder()
	req3 := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(tokenForm.Encode())))
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
	req4 := injectTenant(httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(tokenForm.Encode())))
	req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.TokenExchange(rec4, req4)
	if rec4.Code != http.StatusBadRequest {
		t.Errorf("replayed code should fail, got status = %d", rec4.Code)
	}
}

// --- VerifyJWTPOST tests ---

func TestVerifyJWTPOST_ValidToken_Accepted(t *testing.T) {
	h, store := newHandlersWithVerifier(t)

	result, err := h.Provider.IssueJWTThin(testTenant, "alice-sub", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	body, _ := json.Marshal(map[string]string{"token": result.Token})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verify/jwt", bytes.NewReader(body))
	h.VerifyJWTPOST(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["valid"] != true {
		t.Errorf("valid = %v, want true", resp["valid"])
	}
	if resp["subject"] != "alice-sub" {
		t.Errorf("subject = %v, want alice-sub", resp["subject"])
	}
}

func TestVerifyJWTPOST_ForgedToken_Rejected(t *testing.T) {
	h, store := newHandlersWithVerifier(t)

	result, err := h.Provider.IssueJWTThin(testTenant, "alice-sub", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	// Tamper with the signature portion of the JWT.
	forged := tamperJWTSignature(result.Token)

	body, _ := json.Marshal(map[string]string{"token": forged})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verify/jwt", bytes.NewReader(body))
	h.VerifyJWTPOST(rec, req)

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["valid"] != false {
		t.Errorf("valid = %v, want false for a forged token", resp["valid"])
	}
}

func TestVerifyJWTPOST_NoVerifierConfigured(t *testing.T) {
	h, _ := newHandlers(t) // no Verifier set

	body, _ := json.Marshal(map[string]string{"token": "irrelevant"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verify/jwt", bytes.NewReader(body))
	h.VerifyJWTPOST(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

// --- RevokeToken tests ---

func TestRevokeToken_ForgedToken_NotRevoked(t *testing.T) {
	h, store := newHandlersWithVerifier(t)

	result, err := h.Provider.IssueJWTThin(testTenant, "alice-sub", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)
	forged := tamperJWTSignature(result.Token)

	form := url.Values{"token": {forged}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.RevokeToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (RFC 7009 always returns 200)", rec.Code, http.StatusOK)
	}

	// The real token's jti must still be usable — verify it wasn't revoked
	// as a side effect of processing the forged token.
	claims, err := h.Verifier.VerifyJWT(result.Token)
	if err != nil {
		t.Fatalf("original token should still verify: %v", err)
	}
	if h.Revocation.IsRevoked(claims.TokenID) {
		t.Error("original token's jti should not have been revoked by a forged request")
	}
}

func TestRevokeToken_ValidToken_Revoked(t *testing.T) {
	h, store := newHandlersWithVerifier(t)

	result, err := h.Provider.IssueJWTThin(testTenant, "alice-sub", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWTThin() error = %v", err)
	}
	store.Put(result.PQCRef, result.PQCSig)

	form := url.Values{"token": {result.Token}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.RevokeToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// VerifyJWT itself checks revocation, so a revoked token must now fail.
	if _, verr := h.Verifier.VerifyJWT(result.Token); verr == nil {
		t.Error("revoked token should no longer verify successfully")
	}
}

// tamperJWTSignature flips a character in the middle of a JWT's signature
// segment, invalidating it while keeping the token structurally well-formed.
// (The very last base64url character of an Ed25519 sig only encodes 2
// meaningful bits, so mutating it can be a no-op after re-decoding; a
// mid-segment byte has no such ambiguity.)
func tamperJWTSignature(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(parts[2]) < 2 {
		return token + "x"
	}
	sig := []byte(parts[2])
	mid := len(sig) / 2
	if sig[mid] == 'A' {
		sig[mid] = 'B'
	} else {
		sig[mid] = 'A'
	}
	parts[2] = string(sig)
	return strings.Join(parts, ".")
}

// Suppress unused warning for time import.
var _ = time.Now
