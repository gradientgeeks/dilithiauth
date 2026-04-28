package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
)

// SigStorePutter is the interface for storing PQC signatures, satisfied by
// *transport.SigStore without importing the transport package.
type SigStorePutter interface {
	Put(jti string, sig []byte)
}

// UserAuthenticator is an optional production-grade authenticator backed by
// a hashed-password store (e.g. *userstore.PostgresUserStore).
// When set on AuthzHandlers, it takes precedence over AuthzStore.AuthenticateUser.
// It is satisfied by *userstore.PostgresUserStore via its AuthenticateForSubject method.
type UserAuthenticator interface {
	AuthenticateForSubject(ctx context.Context, tenantID, username, password string) (string, error)
}

// ClientValidator is the interface for authenticating OAuth clients.
// It is separate from AuthzStore so production deployments can use
// PostgresAuthzStore (backed by the developer portal's oauth_apps table)
// for client lookups while using a Redis/memory store for auth codes.
type ClientValidator interface {
	ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error)
	ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI string) (*OAuthClient, error)
}

// AuthzHandlers provides HTTP handlers for the OAuth 2.0 authorization code flow.
type AuthzHandlers struct {
	Provider        *Provider
	Store           AuthzStore
	UserStore       UserAuthenticator // optional; overrides Store.AuthenticateUser when set
	SigStore        SigStorePutter
	Revocation      RevocationStore // optional; enables POST /revoke when set
	ClientValidator ClientValidator // optional; overrides Store for all client lookups
}

// clientValidator returns the effective ClientValidator — the explicit one if
// set, otherwise falls back to Store (which also satisfies the interface).
func (h *AuthzHandlers) clientValidator() ClientValidator {
	if h.ClientValidator != nil {
		return h.ClientValidator
	}
	return h.Store
}

// extractClientCredentials reads client_id + client_secret from the request.
// Priority: HTTP Basic Auth (RFC 6749 §2.3.1), then POST body params.
func extractClientCredentials(r *http.Request) (clientID, clientSecret string) {
	if id, secret, ok := r.BasicAuth(); ok {
		return id, secret
	}
	return r.FormValue("client_id"), r.FormValue("client_secret")
}

const authPageStyle = `
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#0f0f1a;color:#e2e8f0;min-height:100vh;display:flex;align-items:center;justify-content:center;padding:1rem}
.card{background:#1a1a2e;border:1px solid #2d2d4e;border-radius:12px;padding:2rem;width:100%;max-width:400px;box-shadow:0 20px 60px rgba(0,0,0,.5)}
.logo{text-align:center;margin-bottom:1.5rem}
.logo svg{width:48px;height:48px;color:#6366f1}
h1{font-size:1.5rem;font-weight:700;text-align:center;margin-bottom:.25rem;color:#fff}
.subtitle{text-align:center;color:#94a3b8;font-size:.875rem;margin-bottom:1.5rem}
.app-name{color:#6366f1;font-weight:600}
label{display:block;font-size:.8rem;font-weight:500;color:#94a3b8;margin-bottom:.3rem;margin-top:.9rem}
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
    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>
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
    <label for="username">Username</label>
    <input id="username" type="text" name="username" placeholder="your username" autocomplete="username" required>
    <label for="password">Password</label>
    <input id="password" type="password" name="password" placeholder="••••••••" autocomplete="current-password" required>
    <button class="btn" type="submit">Sign in</button>
  </form>
  <div class="divider">or</div>
  <div class="link-row">
    No account? <a href="/register?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Create one</a>
  </div>
  <div class="shield">
    <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
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
    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>
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
    <label for="username">Username</label>
    <input id="username" type="text" name="username" placeholder="choose a username" autocomplete="username" required minlength="3">
    <label for="email">Email</label>
    <input id="email" type="email" name="email" placeholder="you@example.com" autocomplete="email" required>
    <label for="password">Password</label>
    <input id="password" type="password" name="password" placeholder="at least 8 characters" autocomplete="new-password" required minlength="8">
    <button class="btn" type="submit">Create account</button>
  </form>
  <div class="divider">or</div>
  <div class="link-row">
    Already have an account? <a href="/authorize?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}&scope={{.Scope}}&code_challenge={{.CodeChallenge}}&code_challenge_method={{.CodeChallengeMethod}}">Sign in</a>
  </div>
  <div class="shield">
    <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
    Secured by DilithiAuth · Post-Quantum
  </div>
</div>
</body></html>`))

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

	if err := h.Store.RegisterUser(tenantID, &User{
		Username: username,
		Email:    email,
		Password: password,
	}); err != nil {
		redirectToRegister(err.Error())
		return
	}

	// Registration succeeded — immediately issue an auth code so the user is
	// logged in to the calling app without a second sign-in step.
	code, err := h.Store.CreateCode(tenantID, username, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod)
	if err != nil {
		redirectToRegister("account created but could not start session: " + err.Error())
		return
	}

	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += "&state=" + state
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
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

	code, err := h.Store.CreateCode(tenantID, subject, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod)
	if err != nil {
		http.Error(w, "code generation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += "&state=" + state
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// TokenExchange handles the authorization_code grant type at POST /token.
func (h *AuthzHandlers) TokenExchange(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	if grantType != "authorization_code" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "unsupported_grant_type",
			"error_description": "only authorization_code is supported",
		})
		return
	}

	tenantID := tenant.FromContext(r.Context())
	clientID, clientSecret := extractClientCredentials(r)
	redirectURI := r.FormValue("redirect_uri")
	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")

	// Authenticate the client (confidential clients must supply client_secret;
	// public/PKCE-only clients must supply an empty secret).
	if _, err := h.clientValidator().ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_client",
			"error_description": err.Error(),
		})
		return
	}

	ac, err := h.Store.RedeemCode(tenantID, code, clientID, redirectURI, codeVerifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	result, err := h.Provider.IssueJWTThin(ac.Subject, time.Hour)
	if err != nil {
		http.Error(w, "token issuance failed", http.StatusInternalServerError)
		return
	}

	h.SigStore.Put(result.PQCRef, result.PQCSig)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": result.Token,
		"token_type":   "Bearer",
		"pqc_ref":      result.PQCRef,
	})
}

// RevokeToken implements RFC 7009 token revocation at POST /revoke.
// Always returns 200 to avoid leaking information about token validity.
func (h *AuthzHandlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	tokenStr := r.FormValue("token")
	if tokenStr == "" || h.Revocation == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse without verification to extract jti and exp (token may already be
	// expired, but the caller may still want to prevent replay attempts).
	parser := jwt.NewParser()
	tok, _, err := parser.ParseUnverified(tokenStr, &HybridClaims{})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	claims, ok := tok.Claims.(*HybridClaims)
	if !ok || claims.ID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	expiry := time.Now().Add(time.Hour) // safe fallback
	if claims.ExpiresAt != nil {
		expiry = claims.ExpiresAt.Time
	}

	_ = h.Revocation.Revoke(claims.ID, expiry)
	w.WriteHeader(http.StatusOK)
}
