package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
)

// SigStorePutter is the interface for storing PQC signatures, satisfied by
// *transport.SigStore without importing the transport package.
type SigStorePutter interface {
	Put(jti string, sig []byte)
}

// VerifiedTokenClaims is the minimal set of claims a TokenVerifier returns
// after successfully verifying a token's signature(s).
type VerifiedTokenClaims struct {
	Subject   string
	Issuer    string
	TokenID   string
	ExpiresAt time.Time
}

// TokenVerifier verifies the full hybrid (Ed25519 + ML-DSA-44) signature of a
// token. It is satisfied by *verifier.Verifier (via a small adapter) without
// pkg/oauth importing pkg/verifier, which itself imports pkg/oauth.
type TokenVerifier interface {
	VerifyJWT(tokenStr string) (*VerifiedTokenClaims, error)
}

// UserAuthenticator is an optional production-grade authenticator backed by
// a hashed-password store (e.g. *userstore.PostgresUserStore).
// When set on AuthzHandlers, it takes precedence over AuthzStore.AuthenticateUser.
// It is satisfied by *userstore.PostgresUserStore via its AuthenticateForSubject method.
type UserAuthenticator interface {
	AuthenticateForSubject(ctx context.Context, tenantID, username, password string) (string, error)
}

// UserRegistrar is an optional production-grade registrar backed by a
// durable store (e.g. *userstore.PostgresUserStore). When set on
// AuthzHandlers, it takes precedence over AuthzStore.RegisterUser, so that
// registration and authentication always target the same backend — without
// this, RegisterPOST would silently register into the ephemeral in-memory/
// Redis AuthzStore even when UserAuthenticator points at Postgres.
// It is satisfied by *userstore.PostgresUserStore via its RegisterForSubject method.
type UserRegistrar interface {
	RegisterForSubject(ctx context.Context, tenantID, username, email, password string) (string, error)
}

// PasswordResetter is an optional production-grade password-reset backend
// (e.g. *userstore.PostgresUserStore), used instead of AuthzStore for
// forgot-password/reset-password when a durable UserStore is configured —
// mirrors the UserAuthenticator/UserRegistrar override pattern so all three
// stay pointed at the same backend.
type PasswordResetter interface {
	// EmailExists reports whether an account with the given email exists.
	// Used only to decide whether to send a reset code — the HTTP response
	// is identical either way, so this never leaks an enumeration signal.
	EmailExists(ctx context.Context, tenantID, email string) (bool, error)
	// SetPasswordForEmail hashes and sets a new password for the account
	// with the given email.
	SetPasswordForEmail(ctx context.Context, tenantID, email, newPassword string) error
}

// OTPSender is satisfied by *developer.OTPVerificationStore, reused here for
// end-user password-reset codes without pkg/oauth depending on developer's
// account/session types (no import cycle: pkg/developer never imports
// pkg/oauth).
type OTPSender interface {
	GenerateCode(email string) (code string, expiresAt time.Time, err error)
	VerifyCode(email, code string) (bool, error)
}

// ClientValidator is the interface for authenticating OAuth clients.
// It is separate from AuthzStore so production deployments can use
// PostgresAuthzStore (backed by the developer portal's oauth_apps table)
// for client lookups while using a Redis/memory store for auth codes.
type ClientValidator interface {
	ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error)
	ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI string) (*OAuthClient, error)
	// ValidateClientSecretOnly authenticates a client by client_id+client_secret
	// alone, with no redirect_uri check — used by the refresh_token grant,
	// which (per RFC 6749) carries no redirect_uri parameter.
	ValidateClientSecretOnly(tenantID, clientID, clientSecret string) (*OAuthClient, error)
}

// AuthzHandlers provides HTTP handlers for the OAuth 2.0 authorization code flow.
type AuthzHandlers struct {
	Provider         *Provider
	Store            AuthzStore
	UserStore        UserAuthenticator // optional; overrides Store.AuthenticateUser when set
	UserRegistrar    UserRegistrar     // optional; overrides Store.RegisterUser when set (should mirror UserStore)
	PasswordResetter PasswordResetter  // optional; overrides Store for password reset when set (should mirror UserStore)
	OTPStore         OTPSender         // optional; enables /forgot-password and /reset-password when set
	SigStore         SigStorePutter
	Revocation       RevocationStore   // optional; enables POST /revoke when set
	ClientValidator  ClientValidator   // optional; overrides Store for all client lookups
	Verifier         TokenVerifier     // required for /revoke and /verify/jwt to perform real signature verification
	RefreshTokens    RefreshTokenStore // optional; enables refresh_token issuance/grant when set
	RefreshTTL       time.Duration     // refresh token lifetime; defaults to 30 days if zero
	GrantStore       GrantStore        // optional; enables the first-time-only consent screen when set
	ConsentSecret    []byte            // required when GrantStore is set; signs short-lived consent tickets
}

// refreshTTL returns the configured refresh token lifetime, or a 30-day default.
func (h *AuthzHandlers) refreshTTL() time.Duration {
	if h.RefreshTTL > 0 {
		return h.RefreshTTL
	}
	return 30 * 24 * time.Hour
}

// clientValidator returns the effective ClientValidator — the explicit one if
// set, otherwise falls back to Store (which also satisfies the interface).
func (h *AuthzHandlers) clientValidator() ClientValidator {
	if h.ClientValidator != nil {
		return h.ClientValidator
	}
	return h.Store
}

// isPublicClient reports whether clientID is a public (PKCE-only, no
// client_secret) client. Errors resolving the client are treated as "not
// public" here — the subsequent CreateCode/RedeemCode call will surface the
// real validation error for an unknown client.
func (h *AuthzHandlers) isPublicClient(tenantID, clientID, redirectURI string) bool {
	client, err := h.clientValidator().ValidateClient(tenantID, clientID, redirectURI)
	if err != nil || client == nil {
		return false
	}
	return client.ClientSecretHash == ""
}

// emailExists reports whether an account with the given email exists, using
// whichever backend owns end-user accounts (PasswordResetter/UserStore when
// configured, else the fallback AuthzStore). Used only to decide whether to
// send a reset code — never surfaced to the caller, so it can't be used to
// enumerate registered emails.
func (h *AuthzHandlers) emailExists(ctx context.Context, tenantID, email string) bool {
	if h.PasswordResetter != nil {
		ok, _ := h.PasswordResetter.EmailExists(ctx, tenantID, email)
		return ok
	}
	_, err := h.Store.GetUserByEmail(tenantID, email)
	return err == nil
}

// resetPasswordForEmail sets a new password for the account with the given
// email, using whichever backend owns it (mirrors emailExists).
func (h *AuthzHandlers) resetPasswordForEmail(ctx context.Context, tenantID, email, newPassword string) error {
	if h.PasswordResetter != nil {
		return h.PasswordResetter.SetPasswordForEmail(ctx, tenantID, email, newPassword)
	}
	u, err := h.Store.GetUserByEmail(tenantID, email)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return h.Store.SetPassword(tenantID, u.Username, string(hash))
}

// consentTicketClaims is a short-lived (5 min), HS256-signed JWT that
// carries the fully-authenticated authorization request from AuthorizePOST/
// RegisterPOST to ConsentPOST. Only ConsentPOST's "decision" form field is
// trusted from the resubmitted form; every other value (subject, client,
// redirect_uri, scope, PKCE challenge) is read back out of this signed
// ticket so a forged consent POST can't grant access to an arbitrary
// subject or redirect to an arbitrary URI.
type consentTicketClaims struct {
	jwt.RegisteredClaims
	Subject             string `json:"subject"`
	TenantID            string `json:"tenant_id"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	State               string `json:"state"`
	Scope               string `json:"scope"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

func (h *AuthzHandlers) signConsentTicket(c consentTicketClaims) (string, error) {
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return tok.SignedString(h.ConsentSecret)
}

func (h *AuthzHandlers) parseConsentTicket(raw string) (*consentTicketClaims, error) {
	tok, err := jwt.ParseWithClaims(raw, &consentTicketClaims{}, func(t *jwt.Token) (any, error) {
		return h.ConsentSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid {
		return nil, fmt.Errorf("invalid or expired consent ticket")
	}
	claims, ok := tok.Claims.(*consentTicketClaims)
	if !ok {
		return nil, fmt.Errorf("invalid consent ticket claims")
	}
	return claims, nil
}

// grantStore returns the effective GrantStore. When nil (consent screens
// not configured), completeAuthorization always proceeds straight to code
// issuance without ever rendering a consent page.
func (h *AuthzHandlers) grantStore() GrantStore {
	return h.GrantStore
}

// completeAuthorization is the shared tail of AuthorizePOST/RegisterPOST
// once a subject has been authenticated (by password or fresh registration):
// it validates the requested scope, and either issues the authorization
// code immediately (no consent needed, or consent screens disabled) or
// renders the consent screen.
func (h *AuthzHandlers) completeAuthorization(
	w http.ResponseWriter, r *http.Request,
	tenantID, subject, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod string,
	onError func(msg string),
) {
	if err := validateScope(scope); err != nil {
		onError("invalid_scope")
		return
	}

	if gs := h.grantStore(); gs != nil {
		grant, _ := gs.Get(tenantID, subject, clientID)
		if !grantCoversScope(grant, scope) {
			h.renderConsent(w, r, tenantID, subject, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod, onError)
			return
		}
	}

	h.finishAuthorization(w, r, tenantID, subject, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod, onError)
}

func (h *AuthzHandlers) renderConsent(
	w http.ResponseWriter, r *http.Request,
	tenantID, subject, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod string,
	onError func(msg string),
) {
	ticket, err := h.signConsentTicket(consentTicketClaims{
		Subject: subject, TenantID: tenantID, ClientID: clientID, RedirectURI: redirectURI,
		State: state, Scope: scope, CodeChallenge: codeChallenge, CodeChallengeMethod: codeChallengeMethod,
	})
	if err != nil {
		onError("could not start authorization")
		return
	}

	appName := clientID
	if client, cerr := h.clientValidator().ValidateClient(tenantID, clientID, redirectURI); cerr == nil && client != nil && client.Name != "" {
		appName = client.Name
	}

	w.Header().Set("Content-Type", "text/html")
	_ = consentTmpl.Execute(w, consentPageData{
		AppName: appName,
		Scopes:  humanScopes(scope),
		Ticket:  ticket,
	})
}

// finishAuthorization creates the authorization code and redirects to the
// client's redirect_uri — the final step of a successful login or an
// approved consent decision.
func (h *AuthzHandlers) finishAuthorization(
	w http.ResponseWriter, r *http.Request,
	tenantID, subject, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod string,
	onError func(msg string),
) {
	code, err := h.Store.CreateCode(tenantID, subject, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod)
	if err != nil {
		onError("code generation failed: " + err.Error())
		return
	}
	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += "&state=" + state
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// ConsentPOST handles the user's Allow/Deny decision on the consent screen.
// Every security-sensitive field (subject, client, redirect_uri, scope,
// PKCE challenge) is read from the signed ticket, never from the resubmitted
// form — only "decision" is trusted from the form itself.
func (h *AuthzHandlers) ConsentPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	claims, err := h.parseConsentTicket(r.FormValue("ticket"))
	if err != nil {
		http.Error(w, "invalid or expired consent request — please sign in again", http.StatusBadRequest)
		return
	}

	if r.FormValue("decision") != "allow" {
		u := fmt.Sprintf("%s?error=access_denied", claims.RedirectURI)
		if claims.State != "" {
			u += "&state=" + claims.State
		}
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	if gs := h.grantStore(); gs != nil {
		existing, _ := gs.Get(claims.TenantID, claims.Subject, claims.ClientID)
		newScope := claims.Scope
		if existing != nil {
			newScope = mergeScopes(existing.Scope, claims.Scope)
		}
		_ = gs.Save(&Grant{
			Subject: claims.Subject, TenantID: claims.TenantID, ClientID: claims.ClientID,
			Scope: newScope, GrantedAt: time.Now(),
		})
	}

	onError := func(msg string) {
		http.Error(w, msg, http.StatusBadRequest)
	}
	h.finishAuthorization(w, r, claims.TenantID, claims.Subject, claims.ClientID, claims.RedirectURI, claims.State, claims.Scope, claims.CodeChallenge, claims.CodeChallengeMethod, onError)
}

// extractClientCredentials reads client_id + client_secret from the request.
// Priority: HTTP Basic Auth (RFC 6749 §2.3.1), then POST body params.
func extractClientCredentials(r *http.Request) (clientID, clientSecret string) {
	if id, secret, ok := r.BasicAuth(); ok {
		return id, secret
	}
	return r.FormValue("client_id"), r.FormValue("client_secret")
}

// brandLogoSVG is the DilithiAuth brand mark — a hexagonal lattice shield
// (3x3 basis-vector grid, representing lattice-based ML-DSA cryptography)
// with a verified checkmark — adapted from frontend/public/favicon.svg for
// inline use in server-rendered pages. size sets both width and height in
// pixels; the artwork itself stays on a fixed 0 0 48 48 viewBox.
func brandLogoSVG(size int) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="%d" height="%d" role="img" aria-label="DilithiAuth logo">
  <defs>
    <linearGradient id="da-shield" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0" stop-color="#818cf8"/>
      <stop offset="1" stop-color="#4f46e5"/>
    </linearGradient>
    <linearGradient id="da-check" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#22d3ee"/>
      <stop offset="1" stop-color="#0ea5e9"/>
    </linearGradient>
  </defs>
  <path d="M24 2.5 L40.5 9 V22.5 C40.5 33.5 33.5 42 24 45.5 C14.5 42 7.5 33.5 7.5 22.5 V9 Z" fill="url(#da-shield)"/>
  <path d="M24 2.5 L40.5 9 V22.5 C40.5 33.5 33.5 42 24 45.5 C14.5 42 7.5 33.5 7.5 22.5 V9 Z" fill="none" stroke="#312e81" stroke-opacity="0.25" stroke-width="1"/>
  <g stroke="#e0e7ff" stroke-width="1.3" stroke-linecap="round" opacity="0.9">
    <line x1="16" y1="15" x2="24" y2="15"/>
    <line x1="24" y1="15" x2="32" y2="15"/>
    <line x1="16" y1="23" x2="24" y2="23"/>
    <line x1="24" y1="23" x2="32" y2="23"/>
    <line x1="16" y1="15" x2="16" y2="23"/>
    <line x1="24" y1="15" x2="24" y2="23"/>
    <line x1="32" y1="15" x2="32" y2="23"/>
    <line x1="16" y1="15" x2="24" y2="23"/>
    <line x1="24" y1="15" x2="32" y2="23"/>
  </g>
  <g fill="#eef2ff">
    <circle cx="16" cy="15" r="1.9"/>
    <circle cx="24" cy="15" r="1.9"/>
    <circle cx="32" cy="15" r="1.9"/>
    <circle cx="16" cy="23" r="1.9"/>
    <circle cx="24" cy="23" r="1.9"/>
    <circle cx="32" cy="23" r="1.9"/>
  </g>
  <path d="M17 31.5 L22 36.5 L32 25.5" fill="none" stroke="url(#da-check)" stroke-width="3.4" stroke-linecap="round" stroke-linejoin="round"/>
</svg>`, size, size)
}

// iconUser, iconEmail, iconLock, iconHash are small (16x16) currentColor
// inline glyphs shown inside form field labels, matching the muted label
// styling already in authPageStyle.
const iconUser = `<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14c-4.418 0-8 2.239-8 5v1h16v-1c0-2.761-3.582-5-8-5z"/></svg>`

const iconEmail = `<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/></svg>`

const iconLock = `<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>`

const iconHash = `<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v18M15 3v18M4 8h16M4 16h16"/></svg>`

// iconCheckBadge is a small check-in-a-badge glyph used per consent scope
// line item, replacing the plain CSS '✓' bullet with something that reads
// as an approval/verification mark consistent with the brand shield.
const iconCheckBadge = `<svg width="16" height="16" fill="none" viewBox="0 0 24 24" style="flex-shrink:0"><circle cx="12" cy="12" r="10" fill="#312e81"/><path stroke="#a5b4fc" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12.5l2.5 2.5L16 9.5"/></svg>`

const authPageStyle = `
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#0f0f1a;color:#e2e8f0;min-height:100vh;display:flex;align-items:center;justify-content:center;padding:1rem}
.card{background:#1a1a2e;border:1px solid #2d2d4e;border-radius:12px;padding:2rem;width:100%;max-width:400px;box-shadow:0 20px 60px rgba(0,0,0,.5)}
.logo{text-align:center;margin-bottom:1.5rem}
.logo svg{width:40px;height:40px}
h1{font-size:1.5rem;font-weight:700;text-align:center;margin-bottom:.25rem;color:#fff}
.subtitle{text-align:center;color:#94a3b8;font-size:.875rem;margin-bottom:1.5rem}
.app-name{color:#6366f1;font-weight:600}
label{display:flex;align-items:center;gap:.35rem;font-size:.8rem;font-weight:500;color:#94a3b8;margin-bottom:.3rem;margin-top:.9rem}
label svg{color:#64748b;flex-shrink:0}
input[type=text],input[type=email],input[type=password]{width:100%;padding:.65rem .85rem;background:#0f0f1a;border:1px solid #2d2d4e;border-radius:8px;color:#e2e8f0;font-size:.9rem;outline:none;transition:border-color .2s}
input:focus{border-color:#6366f1}
.error{background:#3b1515;border:1px solid #7f1d1d;color:#fca5a5;padding:.6rem .85rem;border-radius:8px;font-size:.85rem;margin-top:.75rem}
.btn{display:block;width:100%;margin-top:1.25rem;padding:.75rem;background:#6366f1;color:#fff;border:none;border-radius:8px;font-size:.9rem;font-weight:600;cursor:pointer;transition:background .2s}
.btn:hover{background:#4f46e5}
.divider{text-align:center;margin:.9rem 0;color:#475569;font-size:.8rem;position:relative}
.divider::before,.divider::after{content:'';position:absolute;top:50%;width:42%;height:1px;background:#2d2d4e}
.divider::before{left:0}.divider::after{right:0}
.link-row{text-align:center;margin-top:1rem;font-size:.85rem;color:#64748b}
.link-row a{color:#6366f1;text-decoration:none;font-weight:500}
.link-row a:hover{color:#818cf8}
.shield{display:flex;align-items:center;justify-content:center;gap:.4rem;font-size:.75rem;color:#475569;margin-top:1.25rem}
</style>`

var loginTmpl = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Sign in — DilithiAuth</title>` + authPageStyle + `</head>
<body>
<div class="card">
  <div class="logo">
    ` + brandLogoSVG(40) + `
  </div>
  <h1>Sign in</h1>
  <p class="subtitle">to continue to <span class="app-name">{{.AppName}}</span></p>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
  <form method="POST" action="/authorize">
    <input type="hidden" name="client_id" value="{{.ClientID}}">
    <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
    <input type="hidden" name="state" value="{{.State}}">
    <input type="hidden" name="scope" value="{{.Scope}}">
    <input type="hidden" name="code_challenge" value="{{.CodeChallenge}}">
    <input type="hidden" name="code_challenge_method" value="{{.CodeChallengeMethod}}">
    <label for="username">` + iconUser + `Username</label>
    <input id="username" type="text" name="username" placeholder="your username" autocomplete="username" required>
    <label for="password">` + iconLock + `Password</label>
    <input id="password" type="password" name="password" placeholder="••••••••" autocomplete="current-password" required>
    <button class="btn" type="submit">Sign in</button>
  </form>
  <div class="divider">or continue with</div>
  <div style="display: flex; gap: 0.5rem; margin-bottom: 1rem;">
    <a href="/auth/social/google?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}" style="flex: 1; display: inline-flex; align-items: center; justify-content: center; gap: 0.4rem; padding: 0.5rem; background: #1f2937; border: 1px solid #374151; border-radius: 6px; color: #f9fafb; font-size: 0.85rem; font-weight: 500; text-decoration: none;">
      Google
    </a>
    <a href="/auth/social/github?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}" style="flex: 1; display: inline-flex; align-items: center; justify-content: center; gap: 0.4rem; padding: 0.5rem; background: #1f2937; border: 1px solid #374151; border-radius: 6px; color: #f9fafb; font-size: 0.85rem; font-weight: 500; text-decoration: none;">
      GitHub
    </a>
  </div>
  <div class="link-row">
    No account? <a href="/register?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Create one</a>
    · <a href="/forgot-password?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Forgot password?</a>
  </div>
  <div class="shield">
    ` + iconLock + `
    Secured by DilithiAuth · Post-Quantum
  </div>
</div>
</body></html>`))

var registerTmpl = template.Must(template.New("register").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Create account — DilithiAuth</title>` + authPageStyle + `</head>
<body>
<div class="card">
  <div class="logo">
    ` + brandLogoSVG(40) + `
  </div>
  <h1>Create account</h1>
  <p class="subtitle">to continue to <span class="app-name">{{.AppName}}</span></p>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
  <form method="POST" action="/register">
    <input type="hidden" name="client_id" value="{{.ClientID}}">
    <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
    <input type="hidden" name="state" value="{{.State}}">
    <input type="hidden" name="scope" value="{{.Scope}}">
    <input type="hidden" name="code_challenge" value="{{.CodeChallenge}}">
    <input type="hidden" name="code_challenge_method" value="{{.CodeChallengeMethod}}">
    <label for="username">` + iconUser + `Username</label>
    <input id="username" type="text" name="username" placeholder="choose a username" autocomplete="username" required minlength="3">
    <label for="email">` + iconEmail + `Email</label>
    <input id="email" type="email" name="email" placeholder="you@example.com" autocomplete="email" required>
    <label for="password">` + iconLock + `Password</label>
    <input id="password" type="password" name="password" placeholder="at least 8 characters" autocomplete="new-password" required minlength="8">
    <button class="btn" type="submit">Create account</button>
  </form>
  <div class="divider">or</div>
  <div class="link-row">
    Already have an account? <a href="/authorize?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Sign in</a>
  </div>
  <div class="shield">
    ` + iconLock + `
    Secured by DilithiAuth · Post-Quantum
  </div>
</div>
</body></html>`))

var forgotPasswordTmpl = template.Must(template.New("forgot-password").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Reset password — DilithiAuth</title>` + authPageStyle + `</head>
<body>
<div class="card">
  <div class="logo">
    ` + brandLogoSVG(40) + `
  </div>
  <h1>Forgot password</h1>
  <p class="subtitle">We'll email a 6-digit code to reset it</p>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
  <form method="POST" action="/forgot-password">
    <input type="hidden" name="client_id" value="{{.ClientID}}">
    <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
    <input type="hidden" name="state" value="{{.State}}">
    <input type="hidden" name="scope" value="{{.Scope}}">
    <input type="hidden" name="code_challenge" value="{{.CodeChallenge}}">
    <input type="hidden" name="code_challenge_method" value="{{.CodeChallengeMethod}}">
    <label for="email">` + iconEmail + `Email</label>
    <input id="email" type="email" name="email" placeholder="you@example.com" autocomplete="email" required>
    <button class="btn" type="submit">Send reset code</button>
  </form>
  <div class="link-row">
    <a href="/authorize?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Back to sign in</a>
  </div>
</div>
</body></html>`))

var resetPasswordTmpl = template.Must(template.New("reset-password").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Reset password — DilithiAuth</title>` + authPageStyle + `</head>
<body>
<div class="card">
  <div class="logo">
    ` + brandLogoSVG(40) + `
  </div>
  <h1>Enter reset code</h1>
  <p class="subtitle">Sent to <span class="app-name">{{.Email}}</span> if an account exists</p>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
  {{if .Message}}<div class="subtitle" style="color:#86efac">{{.Message}}</div>{{end}}
  <form method="POST" action="/reset-password">
    <input type="hidden" name="client_id" value="{{.ClientID}}">
    <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
    <input type="hidden" name="state" value="{{.State}}">
    <input type="hidden" name="scope" value="{{.Scope}}">
    <input type="hidden" name="code_challenge" value="{{.CodeChallenge}}">
    <input type="hidden" name="code_challenge_method" value="{{.CodeChallengeMethod}}">
    <label for="email">` + iconEmail + `Email</label>
    <input id="email" type="email" name="email" value="{{.Email}}" autocomplete="email" required>
    <label for="code">` + iconHash + `6-digit code</label>
    <input id="code" type="text" name="code" placeholder="123456" inputmode="numeric" pattern="[0-9]{6}" maxlength="6" required>
    <label for="password">` + iconLock + `New password</label>
    <input id="password" type="password" name="password" placeholder="at least 8 characters" autocomplete="new-password" required minlength="8">
    <button class="btn" type="submit">Reset password</button>
  </form>
  <div class="link-row">
    <a href="/forgot-password?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Request a new code</a>
    · <a href="/authorize?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Back to sign in</a>
  </div>
</div>
</body></html>`))

var consentTmpl = template.Must(template.New("consent").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Authorize — DilithiAuth</title>` + authPageStyle + `<style>
.scope-list{list-style:none;margin:1rem 0}
.scope-list li{padding:.5rem 0;color:#cbd5e1;font-size:.88rem;display:flex;align-items:center;gap:.6rem}
.btn-row{display:flex;gap:.75rem;margin-top:1.25rem}
.btn-deny{background:transparent;border:1px solid #2d2d4e;color:#94a3b8}
.btn-deny:hover{background:#1a1a2e}
</style></head>
<body>
<div class="card">
  <div class="logo">
    ` + brandLogoSVG(40) + `
  </div>
  <h1>Authorize access</h1>
  <p class="subtitle"><span class="app-name">{{.AppName}}</span> wants to:</p>
  <ul class="scope-list">
    {{range .Scopes}}<li>` + iconCheckBadge + `{{.}}</li>{{end}}
  </ul>
  <form method="POST" action="/consent">
    <input type="hidden" name="ticket" value="{{.Ticket}}">
    <div class="btn-row">
      <button class="btn btn-deny" type="submit" name="decision" value="deny">Deny</button>
      <button class="btn" type="submit" name="decision" value="allow">Allow</button>
    </div>
  </form>
  <div class="shield">
    ` + iconLock + `
    Secured by DilithiAuth · Post-Quantum
  </div>
</div>
</body></html>`))

// consentPageData holds the template data for the consent page.
type consentPageData struct {
	AppName string
	Scopes  []string
	Ticket  string
}

// scopeDescriptions maps recognized scopes to a human-readable description
// shown on the consent screen.
var scopeDescriptions = map[string]string{
	"openid":         "Verify your identity",
	"profile":        "View your basic profile information",
	"email":          "View your email address",
	"offline_access": "Access your account while you're not present (stay signed in)",
}

// humanScopes converts a space-delimited scope string into display strings,
// always including at least a base "verify your identity" line even if no
// scopes were requested.
func humanScopes(scope string) []string {
	fields := strings.Fields(scope)
	if len(fields) == 0 {
		return []string{"Verify your identity"}
	}
	out := make([]string, 0, len(fields))
	for _, s := range fields {
		if desc, ok := scopeDescriptions[s]; ok {
			out = append(out, desc)
		} else {
			out = append(out, s)
		}
	}
	return out
}

// authPageData holds the common template data for login/register pages.
type authPageData struct {
	ClientID            string
	RedirectURI         string
	State               string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	AppName             string
	Error               string
}

// resetPasswordPageData holds the template data for the reset-password page.
type resetPasswordPageData struct {
	ClientID            string
	RedirectURI         string
	State               string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	Email               string
	Error               string
	Message             string
}

// AuthorizeGET renders the login form.
func (h *AuthzHandlers) AuthorizeGET(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")
	loginError := r.URL.Query().Get("error")

	if clientID == "" || redirectURI == "" {
		http.Error(w, "missing client_id or redirect_uri", http.StatusBadRequest)
		return
	}

	client, err := h.clientValidator().ValidateClient(tenantID, clientID, redirectURI)
	if err != nil {
		http.Error(w, "invalid client: "+err.Error(), http.StatusBadRequest)
		return
	}

	if codeChallenge != "" && codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}

	appName := clientID
	if client != nil && client.Name != "" {
		appName = client.Name
	}

	w.Header().Set("Content-Type", "text/html")
	_ = loginTmpl.Execute(w, authPageData{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		State:               state,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		AppName:             appName,
		Error:               loginError,
	})
}

// RegisterGET renders the user registration form.
func (h *AuthzHandlers) RegisterGET(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")
	regError := r.URL.Query().Get("error")

	if clientID == "" || redirectURI == "" {
		http.Error(w, "missing client_id or redirect_uri", http.StatusBadRequest)
		return
	}

	client, err := h.clientValidator().ValidateClient(tenantID, clientID, redirectURI)
	if err != nil {
		http.Error(w, "invalid client: "+err.Error(), http.StatusBadRequest)
		return
	}

	appName := clientID
	if client != nil && client.Name != "" {
		appName = client.Name
	}

	w.Header().Set("Content-Type", "text/html")
	_ = registerTmpl.Execute(w, authPageData{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		State:               state,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		AppName:             appName,
		Error:               regError,
	})
}

// RegisterPOST creates a new user account and immediately completes the OAuth flow.
func (h *AuthzHandlers) RegisterPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	tenantID := tenant.FromContext(r.Context())
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	scope := r.FormValue("scope")
	codeChallenge := r.FormValue("code_challenge")
	codeChallengeMethod := r.FormValue("code_challenge_method")

	redirectToRegister := func(errMsg string) {
		u := fmt.Sprintf(
			"/register?client_id=%s&redirect_uri=%s&state=%s&scope=%s&code_challenge=%s&code_challenge_method=%s&error=%s",
			clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod, errMsg,
		)
		http.Redirect(w, r, u, http.StatusSeeOther)
	}

	if len(username) < 3 {
		redirectToRegister("Username must be at least 3 characters")
		return
	}
	if len(password) < 8 {
		redirectToRegister("Password must be at least 8 characters")
		return
	}
	if codeChallenge == "" && h.isPublicClient(tenantID, clientID, redirectURI) {
		redirectToRegister("PKCE (code_challenge) is required for this application")
		return
	}

	// Registration always targets the same backend AuthorizePOST authenticates
	// against below: UserRegistrar/UserStore (durable, e.g. Postgres) when
	// configured, else the fallback AuthzStore (in-memory/Redis). Using one
	// for registration and the other for login would silently split accounts
	// across two stores.
	var subject string
	if h.UserRegistrar != nil {
		sub, err := h.UserRegistrar.RegisterForSubject(r.Context(), tenantID, username, email, password)
		if err != nil {
			redirectToRegister(err.Error())
			return
		}
		subject = sub
	} else {
		if err := h.Store.RegisterUser(tenantID, &User{
			Username: username,
			Email:    email,
			Password: password,
		}); err != nil {
			redirectToRegister(err.Error())
			return
		}
		subject = username
	}

	// Registration succeeded — hand off to the shared consent/code-issuance
	// path so a freshly registered user goes through the same first-time
	// consent screen (if configured) as any other login.
	h.completeAuthorization(w, r, tenantID, subject, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod, func(msg string) {
		redirectToRegister("account created but could not start session: " + msg)
	})
}

// AuthorizePOST authenticates the user and redirects with an authorization code.
func (h *AuthzHandlers) AuthorizePOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	tenantID := tenant.FromContext(r.Context())
	username := r.FormValue("username")
	password := r.FormValue("password")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	scope := r.FormValue("scope")
	codeChallenge := r.FormValue("code_challenge")
	codeChallengeMethod := r.FormValue("code_challenge_method")

	redirectToLogin := func(errMsg string) {
		u := fmt.Sprintf(
			"/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=%s&code_challenge=%s&code_challenge_method=%s&error=%s",
			clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod, errMsg,
		)
		http.Redirect(w, r, u, http.StatusSeeOther)
	}

	// Authenticate via UserStore (Argon2id) if available, else fall back to demo store.
	var subject string
	if h.UserStore != nil {
		sub, err := h.UserStore.AuthenticateForSubject(r.Context(), tenantID, username, password)
		if err != nil {
			redirectToLogin("Invalid username or password")
			return
		}
		subject = sub
	} else {
		user, err := h.Store.AuthenticateUser(tenantID, username, password)
		if err != nil {
			redirectToLogin("Invalid username or password")
			return
		}
		subject = user.Subject
	}

	if codeChallenge == "" && h.isPublicClient(tenantID, clientID, redirectURI) {
		redirectToLogin("PKCE (code_challenge) is required for this application")
		return
	}

	h.completeAuthorization(w, r, tenantID, subject, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod, redirectToLogin)
}

// oauthParamsFromQuery/oauthParamsFromForm read the six hidden OAuth params
// carried through every auth-adjacent page (login, register, forgot/reset
// password) so they survive the multi-step redirect chain.
type oauthParams struct {
	ClientID, RedirectURI, State, Scope, CodeChallenge, CodeChallengeMethod string
}

func oauthParamsFromQuery(r *http.Request) oauthParams {
	q := r.URL.Query()
	return oauthParams{
		ClientID: q.Get("client_id"), RedirectURI: q.Get("redirect_uri"), State: q.Get("state"),
		Scope: q.Get("scope"), CodeChallenge: q.Get("code_challenge"), CodeChallengeMethod: q.Get("code_challenge_method"),
	}
}

func oauthParamsFromForm(r *http.Request) oauthParams {
	return oauthParams{
		ClientID: r.FormValue("client_id"), RedirectURI: r.FormValue("redirect_uri"), State: r.FormValue("state"),
		Scope: r.FormValue("scope"), CodeChallenge: r.FormValue("code_challenge"), CodeChallengeMethod: r.FormValue("code_challenge_method"),
	}
}

func (p oauthParams) queryString() string {
	return fmt.Sprintf(
		"client_id=%s&redirect_uri=%s&state=%s&scope=%s&code_challenge=%s&code_challenge_method=%s",
		p.ClientID, p.RedirectURI, p.State, p.Scope, p.CodeChallenge, p.CodeChallengeMethod,
	)
}

// ForgotPasswordGET renders the "enter your email" form.
func (h *AuthzHandlers) ForgotPasswordGET(w http.ResponseWriter, r *http.Request) {
	p := oauthParamsFromQuery(r)
	w.Header().Set("Content-Type", "text/html")
	_ = forgotPasswordTmpl.Execute(w, authPageData{
		ClientID: p.ClientID, RedirectURI: p.RedirectURI, State: p.State,
		Scope: p.Scope, CodeChallenge: p.CodeChallenge, CodeChallengeMethod: p.CodeChallengeMethod,
		Error: r.URL.Query().Get("error"),
	})
}

// ForgotPasswordPOST fires a reset OTP if the email is registered, then
// always redirects to the reset-password page with an identical message
// regardless of whether the account exists — this must never leak an
// enumeration signal (mirrors developer.Handlers.ResendCode's behavior).
func (h *AuthzHandlers) ForgotPasswordPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	tenantID := tenant.FromContext(r.Context())
	email := r.FormValue("email")
	p := oauthParamsFromForm(r)

	if h.OTPStore != nil && email != "" && h.emailExists(r.Context(), tenantID, email) {
		_, _, _ = h.OTPStore.GenerateCode(email)
	}

	u := fmt.Sprintf("/reset-password?%s&email=%s", p.queryString(), email)
	http.Redirect(w, r, u, http.StatusSeeOther)
}

// ResetPasswordGET renders the "enter code + new password" form.
func (h *AuthzHandlers) ResetPasswordGET(w http.ResponseWriter, r *http.Request) {
	p := oauthParamsFromQuery(r)
	w.Header().Set("Content-Type", "text/html")
	_ = resetPasswordTmpl.Execute(w, resetPasswordPageData{
		ClientID: p.ClientID, RedirectURI: p.RedirectURI, State: p.State,
		Scope: p.Scope, CodeChallenge: p.CodeChallenge, CodeChallengeMethod: p.CodeChallengeMethod,
		Email: r.URL.Query().Get("email"), Error: r.URL.Query().Get("error"),
	})
}

// ResetPasswordPOST verifies the OTP and sets the new password.
func (h *AuthzHandlers) ResetPasswordPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	tenantID := tenant.FromContext(r.Context())
	email := r.FormValue("email")
	code := r.FormValue("code")
	newPassword := r.FormValue("password")
	p := oauthParamsFromForm(r)

	redirectToReset := func(errMsg string) {
		u := fmt.Sprintf("/reset-password?%s&email=%s&error=%s", p.queryString(), email, errMsg)
		http.Redirect(w, r, u, http.StatusSeeOther)
	}

	if len(newPassword) < 8 {
		redirectToReset("Password must be at least 8 characters")
		return
	}
	if h.OTPStore == nil {
		redirectToReset("Password reset is not available")
		return
	}
	valid, err := h.OTPStore.VerifyCode(email, code)
	if err != nil || !valid {
		msg := "Invalid or expired code"
		if err != nil {
			msg = err.Error()
		}
		redirectToReset(msg)
		return
	}

	if err := h.resetPasswordForEmail(r.Context(), tenantID, email, newPassword); err != nil {
		redirectToReset("Could not reset password: " + err.Error())
		return
	}

	u := fmt.Sprintf("/authorize?%s&error=%s", p.queryString(), "Password reset - please sign in")
	http.Redirect(w, r, u, http.StatusSeeOther)
}

// TokenExchange handles the authorization_code grant type at POST /token.
func (h *AuthzHandlers) TokenExchange(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	switch r.FormValue("grant_type") {
	case "authorization_code":
		h.tokenExchangeAuthCode(w, r)
	case "refresh_token":
		h.tokenExchangeRefresh(w, r)
	default:
		writeTokenError(w, http.StatusBadRequest, "unsupported_grant_type", "only authorization_code and refresh_token are supported")
	}
}

func writeTokenError(w http.ResponseWriter, status int, errCode, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             errCode,
		"error_description": description,
	})
}

func (h *AuthzHandlers) tokenExchangeAuthCode(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.FromContext(r.Context())
	clientID, clientSecret := extractClientCredentials(r)
	redirectURI := r.FormValue("redirect_uri")
	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")

	// Authenticate the client (confidential clients must supply client_secret;
	// public/PKCE-only clients must supply an empty secret).
	client, err := h.clientValidator().ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI)
	if err != nil {
		writeTokenError(w, http.StatusUnauthorized, "invalid_client", err.Error())
		return
	}

	ac, err := h.Store.RedeemCode(tenantID, code, clientID, redirectURI, codeVerifier)
	if err != nil {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", err.Error())
		return
	}

	// Defense in depth: a public client's authorization code must have been
	// created with a PKCE challenge. This should already be enforced at
	// AuthorizePOST/RegisterPOST time, but is re-checked here in case a
	// future AuthzStore implementation doesn't enforce it at CreateCode.
	if client != nil && client.ClientSecretHash == "" && ac.CodeChallenge == "" {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "PKCE is required for public clients")
		return
	}

	// New refresh-token lineage: a fresh family, since this is a brand new
	// login, not a rotation of an existing refresh token.
	familyID, err := generateFamilyID()
	if err != nil {
		http.Error(w, "token issuance failed", http.StatusInternalServerError)
		return
	}
	h.issueTokens(w, tenantID, ac.Subject, clientID, ac.Scope, familyID)
}

func (h *AuthzHandlers) tokenExchangeRefresh(w http.ResponseWriter, r *http.Request) {
	if h.RefreshTokens == nil {
		writeTokenError(w, http.StatusBadRequest, "unsupported_grant_type", "refresh_token grant is not enabled")
		return
	}

	tenantID := tenant.FromContext(r.Context())
	clientID, clientSecret := extractClientCredentials(r)
	refreshToken := r.FormValue("refresh_token")

	if _, err := h.clientValidator().ValidateClientSecretOnly(tenantID, clientID, clientSecret); err != nil {
		writeTokenError(w, http.StatusUnauthorized, "invalid_client", err.Error())
		return
	}

	rec, err := h.RefreshTokens.Get(refreshToken)
	if err != nil {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "invalid or revoked refresh token")
		return
	}
	if rec.TenantID != tenantID || rec.ClientID != clientID {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "refresh token was not issued to this client")
		return
	}
	if time.Now().After(rec.ExpiresAt) {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "refresh token expired")
		return
	}
	if rec.Used {
		// A used refresh token being presented again means it was either
		// replayed by an attacker who stole it, or the legitimate client
		// retried after losing the response to the rotation that consumed
		// it. Either way, the whole family can no longer be trusted.
		_ = h.RefreshTokens.RevokeFamily(rec.FamilyID)
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "refresh token reuse detected; session revoked")
		return
	}

	if err := h.RefreshTokens.MarkUsed(refreshToken); err != nil {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "invalid refresh token")
		return
	}
	h.issueTokens(w, tenantID, rec.Subject, clientID, rec.Scope, rec.FamilyID)
}

// issueTokens mints a new access token (and, if RefreshTokens is configured,
// a rotated refresh token in familyID) and writes the /token JSON response.
func (h *AuthzHandlers) issueTokens(w http.ResponseWriter, tenantID, subject, clientID, scope, familyID string) {
	result, err := h.Provider.IssueJWTThinWithScope(tenantID, subject, scope, time.Hour)
	if err != nil {
		http.Error(w, "token issuance failed", http.StatusInternalServerError)
		return
	}
	h.SigStore.Put(result.PQCRef, result.PQCSig)

	resp := map[string]string{
		"access_token": result.Token,
		"token_type":   "Bearer",
		"pqc_ref":      result.PQCRef,
	}

	// Only issue a refresh token if the client asked for offline access —
	// otherwise every login would silently grant long-lived (30-day)
	// refresh capability regardless of what was requested/consented to.
	if h.RefreshTokens != nil && hasScope(scope, "offline_access") {
		newToken, err := generateRefreshToken()
		if err != nil {
			http.Error(w, "token issuance failed", http.StatusInternalServerError)
			return
		}
		if err := h.RefreshTokens.Create(&RefreshTokenRecord{
			Token:     newToken,
			Subject:   subject,
			TenantID:  tenantID,
			ClientID:  clientID,
			Scope:     scope,
			FamilyID:  familyID,
			ExpiresAt: time.Now().Add(h.refreshTTL()),
		}); err != nil {
			http.Error(w, "token issuance failed", http.StatusInternalServerError)
			return
		}
		resp["refresh_token"] = newToken
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// RevokeToken implements RFC 7009 token revocation at POST /revoke.
// Always returns 200 to avoid leaking information about token validity.
// The token's full hybrid signature (Ed25519 + ML-DSA-44) must verify before
// its jti is revoked — otherwise anyone who observed/guessed a jti could
// revoke someone else's token without ever possessing it.
func (h *AuthzHandlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	tokenStr := r.FormValue("token")
	if tokenStr == "" || h.Revocation == nil || h.Verifier == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	claims, err := h.Verifier.VerifyJWT(tokenStr)
	if err != nil || claims.TokenID == "" {
		// Signature invalid/unverifiable — do not act on the claimed jti.
		w.WriteHeader(http.StatusOK)
		return
	}

	expiry := claims.ExpiresAt
	if expiry.IsZero() {
		expiry = time.Now().Add(time.Hour) // safe fallback
	}

	_ = h.Revocation.Revoke(claims.TokenID, expiry)
	w.WriteHeader(http.StatusOK)
}

// VerifyJWTPOST provides a REST endpoint (POST /verify/jwt) for lightweight clients
// to outsource PQC token validation to DilithiAuth without compiling native PQC libraries.
func (h *AuthzHandlers) VerifyJWTPOST(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "invalid request body (expected JSON with 'token')", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if h.Verifier == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"error": "verification is not configured",
		})
		return
	}

	claims, err := h.Verifier.VerifyJWT(req.Token)
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"subject": claims.Subject,
		"issuer":  claims.Issuer,
		"jti":     claims.TokenID,
		"exp":     claims.ExpiresAt.Unix(),
	})
}
