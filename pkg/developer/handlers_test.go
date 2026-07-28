package developer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/developer"
)

func newTestHandlers() *developer.Handlers {
	return &developer.Handlers{
		Store:             developer.NewMemoryDevStore(),
		OTPStore:          developer.NewOTPStore(nil), // nil mailer logs to stdout in tests
		SessionSecret:     bytes.Repeat([]byte{0x01}, 32),
		SessionRevocation: developer.NewMemorySessionRevocationStore(),
	}
}

// Register must never leak the OTP verification code in its JSON response —
// doing so lets anyone complete "email verification" without ever reading
// the victim's inbox.
func TestRegister_DoesNotLeakVerificationCode(t *testing.T) {
	h := newTestHandlers()

	body, _ := json.Marshal(map[string]string{
		"email":    "alice@example.com",
		"name":     "Alice",
		"password": "supersecret123",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"verification_code"`) {
		t.Errorf("response must not contain a verification_code field, got: %s", rec.Body.String())
	}
}

// ResendCode must never leak the OTP verification code either — it is an
// unauthenticated endpoint that only requires knowing the victim's email.
func TestResendCode_DoesNotLeakVerificationCode(t *testing.T) {
	h := newTestHandlers()

	regBody, _ := json.Marshal(map[string]string{
		"email":    "bob@example.com",
		"name":     "Bob",
		"password": "supersecret123",
	})
	regRec := httptest.NewRecorder()
	h.Register(regRec, httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody)))
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", regRec.Code, http.StatusCreated)
	}

	body, _ := json.Marshal(map[string]string{"email": "bob@example.com"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/resend-code", bytes.NewReader(body))
	h.ResendCode(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"verification_code"`) {
		t.Errorf("response must not contain a verification_code field, got: %s", rec.Body.String())
	}
}

// VerifyEmail must issue a session bound to the developer's real ID, not a
// fabricated one — otherwise a developer who verifies email gets a
// different DeveloperID than one who later logs in via /login, and
// ListApps/CreateApp/etc (all keyed on DeveloperID) silently diverge.
func TestVerifyEmail_SessionUsesRealDeveloperID(t *testing.T) {
	h := newTestHandlers()

	dev, err := h.Store.RegisterDeveloper(context.Background(), "erin@example.com", "Erin", "supersecret123")
	if err != nil {
		t.Fatalf("RegisterDeveloper() error = %v", err)
	}
	code, _, err := h.OTPStore.GenerateCode(dev.Email)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	body, _ := json.Marshal(map[string]string{"email": dev.Email, "code": code})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verify-email", bytes.NewReader(body))
	h.VerifyEmail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("VerifyEmail response set no cookies")
	}
	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	for _, c := range cookies {
		meReq.AddCookie(c)
	}
	meRec := httptest.NewRecorder()
	h.Router().ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("/me status = %d, want %d, body=%s", meRec.Code, http.StatusOK, meRec.Body.String())
	}

	var me struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(meRec.Body).Decode(&me); err != nil {
		t.Fatalf("decoding /me response: %v", err)
	}
	if me.ID != dev.ID {
		t.Errorf("session DeveloperID = %q, want the real developer ID %q", me.ID, dev.ID)
	}
}

// CreateApp (and the other state-changing app endpoints) must reject
// requests missing a valid X-CSRF-Token header, and accept requests that
// echo back the token returned by /me — otherwise SameSite=Lax is the only
// CSRF defense for these JSON endpoints.
func TestCreateApp_RequiresValidCSRFToken(t *testing.T) {
	h := newTestHandlers()
	router := h.Router()

	if _, err := h.Store.RegisterDeveloper(context.Background(), "frank@example.com", "Frank", "supersecret123"); err != nil {
		t.Fatalf("RegisterDeveloper() error = %v", err)
	}
	loginBody, _ := json.Marshal(map[string]string{"email": "frank@example.com", "password": "supersecret123"})
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody)))
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginRec.Code, http.StatusOK)
	}
	cookies := loginRec.Result().Cookies()

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	for _, c := range cookies {
		meReq.AddCookie(c)
	}
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	var me struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.NewDecoder(meRec.Body).Decode(&me); err != nil {
		t.Fatalf("decoding /me response: %v", err)
	}
	if me.CSRFToken == "" {
		t.Fatal("/me did not return a csrf_token")
	}

	appBody, _ := json.Marshal(map[string]any{
		"name":          "Test App",
		"redirect_uris": []string{"https://example.com/callback"},
	})

	// Without the CSRF header: must be rejected.
	noTokenReq := httptest.NewRequest(http.MethodPost, "/apps", bytes.NewReader(appBody))
	for _, c := range cookies {
		noTokenReq.AddCookie(c)
	}
	noTokenRec := httptest.NewRecorder()
	router.ServeHTTP(noTokenRec, noTokenReq)
	if noTokenRec.Code != http.StatusForbidden {
		t.Errorf("CreateApp without CSRF token status = %d, want %d", noTokenRec.Code, http.StatusForbidden)
	}

	// With the correct CSRF header: must succeed.
	withTokenReq := httptest.NewRequest(http.MethodPost, "/apps", bytes.NewReader(appBody))
	for _, c := range cookies {
		withTokenReq.AddCookie(c)
	}
	withTokenReq.Header.Set("X-CSRF-Token", me.CSRFToken)
	withTokenRec := httptest.NewRecorder()
	router.ServeHTTP(withTokenRec, withTokenReq)
	if withTokenRec.Code != http.StatusCreated {
		t.Errorf("CreateApp with valid CSRF token status = %d, want %d, body=%s", withTokenRec.Code, http.StatusCreated, withTokenRec.Body.String())
	}
}

// A BFF-issued session cookie (dilithiauth_bff_session) is a different JWT
// shape than the developer portal's own da_sess cookie, but must still carry
// and enforce a CSRF token through the requireSession fallback path — this
// guards against the BFF login flow (the one the production frontend
// actually uses) silently bypassing CSRF checks because its claims lack a
// "csrf" field.
func TestCreateApp_RequiresValidCSRFToken_ViaBFFSessionCookie(t *testing.T) {
	h := newTestHandlers()
	router := h.Router()

	// The BFF session carries an end-user Subject (e.g. "usr_...") which is
	// never a valid developer_accounts.id — requireSession must resolve (or
	// just-in-time provision) a real developer account via the verified
	// Email claim instead of using Subject directly as DeveloperID, or
	// CreateApp would violate the oauth_apps_developer_id_fkey constraint
	// against Postgres (the in-memory store used here doesn't enforce that,
	// so this only surfaces in production without this coverage).
	bffClaims := struct {
		jwt.RegisteredClaims
		Subject string `json:"sub"`
		Email   string `json:"email,omitempty"`
		CSRF    string `json:"csrf,omitempty"`
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Subject: "usr_bff1234567890ab",
		Email:   "bff-user@example.com",
		CSRF:    "bff-csrf-token-value",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, bffClaims)
	signed, err := tok.SignedString(h.SessionSecret)
	if err != nil {
		t.Fatalf("signing BFF-style cookie: %v", err)
	}
	bffCookie := &http.Cookie{Name: "dilithiauth_bff_session", Value: signed}

	appBody, _ := json.Marshal(map[string]any{
		"name":          "BFF App",
		"redirect_uris": []string{"https://example.com/callback"},
	})

	// Missing the CSRF header: must be rejected even though the session itself is valid.
	noTokenReq := httptest.NewRequest(http.MethodPost, "/apps", bytes.NewReader(appBody))
	noTokenReq.AddCookie(bffCookie)
	noTokenRec := httptest.NewRecorder()
	router.ServeHTTP(noTokenRec, noTokenReq)
	if noTokenRec.Code != http.StatusForbidden {
		t.Errorf("CreateApp via BFF cookie without CSRF header status = %d, want %d", noTokenRec.Code, http.StatusForbidden)
	}

	// With the correct CSRF header: must succeed.
	withTokenReq := httptest.NewRequest(http.MethodPost, "/apps", bytes.NewReader(appBody))
	withTokenReq.AddCookie(bffCookie)
	withTokenReq.Header.Set("X-CSRF-Token", "bff-csrf-token-value")
	withTokenRec := httptest.NewRecorder()
	router.ServeHTTP(withTokenRec, withTokenReq)
	if withTokenRec.Code != http.StatusCreated {
		t.Errorf("CreateApp via BFF cookie with valid CSRF header status = %d, want %d, body=%s", withTokenRec.Code, http.StatusCreated, withTokenRec.Body.String())
	}
}

// AuthenticateDeveloper on the in-memory store must exercise a dummy
// password-hash comparison for a nonexistent email — not merely return an
// error immediately — so failed-login latency doesn't leak whether an
// account exists (this mirrors the Postgres-backed Store's behavior).
func TestMemoryDevStore_AuthenticateDeveloper_NoTimingOracleForUnknownEmail(t *testing.T) {
	store := developer.NewMemoryDevStore()
	ctx := context.Background()

	if _, err := store.RegisterDeveloper(ctx, "dave@example.com", "Dave", "correct-password"); err != nil {
		t.Fatalf("RegisterDeveloper() error = %v", err)
	}

	if _, err := store.AuthenticateDeveloper(ctx, "dave@example.com", "wrong-password"); err == nil {
		t.Error("wrong password for existing account should fail")
	}
	if _, err := store.AuthenticateDeveloper(ctx, "nobody@example.com", "anything"); err == nil {
		t.Error("nonexistent account should fail")
	}
}

func TestRegister_RejectsInvalidEmail(t *testing.T) {
	h := newTestHandlers()
	body, _ := json.Marshal(map[string]string{
		"email":    "not-an-email",
		"name":     "Grace",
		"password": "supersecret123",
	})
	rec := httptest.NewRecorder()
	h.Register(rec, httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegister_RejectsShortPassword(t *testing.T) {
	h := newTestHandlers()
	body, _ := json.Marshal(map[string]string{
		"email":    "grace@example.com",
		"name":     "Grace",
		"password": "short",
	})
	rec := httptest.NewRecorder()
	h.Register(rec, httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

// Logout must actually invalidate the session server-side: since sessions
// are stateless 24h JWTs, clearing the cookie alone does nothing if the
// client (or an attacker who captured the cookie) replays it directly.
func TestLogout_RevokesSession(t *testing.T) {
	h := newTestHandlers()
	router := h.Router()

	if _, err := h.Store.RegisterDeveloper(context.Background(), "carol@example.com", "Carol", "supersecret123"); err != nil {
		t.Fatalf("RegisterDeveloper() error = %v", err)
	}

	loginBody, _ := json.Marshal(map[string]string{"email": "carol@example.com", "password": "supersecret123"})
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody)))
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body=%s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login response set no cookies")
	}

	// The session must work before logout.
	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	for _, c := range cookies {
		meReq.AddCookie(c)
	}
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("/me before logout status = %d, want %d, body=%s", meRec.Code, http.StatusOK, meRec.Body.String())
	}

	// Logout.
	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", nil)
	for _, c := range cookies {
		logoutReq.AddCookie(c)
	}
	logoutRec := httptest.NewRecorder()
	router.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", logoutRec.Code, http.StatusNoContent)
	}

	// Replaying the same (pre-logout) cookie must now be rejected.
	replayReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	for _, c := range cookies {
		replayReq.AddCookie(c)
	}
	replayRec := httptest.NewRecorder()
	router.ServeHTTP(replayRec, replayReq)
	if replayRec.Code != http.StatusUnauthorized {
		t.Errorf("/me after logout status = %d, want %d (revoked session must be rejected)", replayRec.Code, http.StatusUnauthorized)
	}
}
