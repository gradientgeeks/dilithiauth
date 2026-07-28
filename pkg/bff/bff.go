// Package bff provides a Backend-for-Frontend (BFF) implementation that allows
// the DilithiAuth Developer Portal (and third-party single-page applications)
// to authenticate using DilithiAuth's own post-quantum OAuth 2.0 / OIDC provider.
//
// Tokens are stored securely in HttpOnly, Secure, SameSite=Lax cookies,
// keeping PQC access tokens and client credentials completely inaccessible
// to browser-side JavaScript (XSS safe).
package bff

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

const (
	sessionCookieName = "dilithiauth_bff_session"
	pkceCookieName    = "dilithiauth_bff_pkce"
)

// Config holds configuration parameters for the BFF server.
type Config struct {
	IssuerURL     string        // e.g. https://dilithiauthapis.gradientgeeks.tech
	ClientID      string        // e.g. da_developer_portal
	ClientSecret  string        // optional for confidential clients
	RedirectURI   string        // e.g. https://dilithiauth.gradientgeeks.tech/auth/callback
	FrontendURL   string        // e.g. https://dilithiauth.gradientgeeks.tech
	SessionSecret []byte        // 32-byte secret for session cookie encryption/signing
	TokenTTL      time.Duration // default session TTL
}

// Server handles BFF authentication endpoints.
type Server struct {
	config Config
	v      *verifier.Verifier
}

// NewServer creates a new BFF Server instance.
func NewServer(cfg Config, v *verifier.Verifier) (*Server, error) {
	if cfg.IssuerURL == "" || cfg.ClientID == "" || cfg.RedirectURI == "" {
		return nil, errors.New("bff: IssuerURL, ClientID, and RedirectURI are required")
	}
	if len(cfg.SessionSecret) < 32 {
		return nil, errors.New("bff: SessionSecret must be at least 32 bytes")
	}
	if cfg.TokenTTL == 0 {
		cfg.TokenTTL = 24 * time.Hour
	}
	if cfg.FrontendURL == "" {
		cfg.FrontendURL = cfg.RedirectURI
	}
	return &Server{config: cfg, v: v}, nil
}

// Router returns an http.Handler with mounted BFF endpoints.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/login", s.HandleLogin)
	r.Get("/callback", s.HandleCallback)
	r.Get("/me", s.HandleMe)
	r.Post("/logout", s.HandleLogout)

	return r
}

type pkceSession struct {
	State        string `json:"state"`
	CodeVerifier string `json:"code_verifier"`
}

type userSessionClaims struct {
	jwt.RegisteredClaims
	Subject   string `json:"sub"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	PQCRef    string `json:"pqc_ref,omitempty"`
	CSRFToken string `json:"csrf,omitempty"`
}

// socialProviders is the set of providers HandleLogin will forward to via
// DilithiAuth's own /auth/social/{provider} endpoint instead of /authorize.
var socialProviders = map[string]bool{"google": true, "github": true}

// HandleLogin initiates the PKCE OAuth 2.0 Authorization Code flow. If a
// recognized `provider` query param is present (e.g. "google", "github"),
// the browser is sent to DilithiAuth's own /auth/social/{provider} endpoint
// instead of /authorize, so the user authenticates with that provider first;
// DilithiAuth still issues the authorization code redeemed below, using the
// same state/PKCE values generated here. Without this, a frontend linking
// straight to /auth/social/{provider} bypasses the PKCE cookie and state
// this handler sets up, and HandleCallback always rejects it.
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate random PKCE verifier (32 random bytes => 43 chars base64url)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		http.Error(w, "bff: generating PKCE verifier failed", http.StatusInternalServerError)
		return
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Compute PKCE code_challenge (S256)
	h := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

	// Generate random state parameter to prevent CSRF
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		http.Error(w, "bff: generating state failed", http.StatusInternalServerError)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Store PKCE state & verifier in short-lived HttpOnly cookie
	pkceData, err := json.Marshal(pkceSession{State: state, CodeVerifier: codeVerifier})
	if err != nil {
		http.Error(w, "bff: marshalling PKCE session failed", http.StatusInternalServerError)
		return
	}

	encodedPKCE := base64.RawURLEncoding.EncodeToString(pkceData)
	http.SetCookie(w, &http.Cookie{
		Name:     pkceCookieName,
		Value:    encodedPKCE,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.config.IssuerURL, "https"),
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect browser to DilithiAuth IdP /authorize endpoint, or to a social
	// provider via DilithiAuth's own /auth/social/{provider} if requested.
	authPath := "/authorize"
	if provider := r.URL.Query().Get("provider"); socialProviders[provider] {
		authPath = "/auth/social/" + provider
	}
	authURL, err := url.Parse(s.config.IssuerURL + authPath)
	if err != nil {
		http.Error(w, "bff: invalid authorize URL", http.StatusInternalServerError)
		return
	}

	q := authURL.Query()
	q.Set("response_type", "code")
	q.Set("client_id", s.config.ClientID)
	q.Set("redirect_uri", s.config.RedirectURI)
	q.Set("scope", "openid profile email")
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	authURL.RawQuery = q.Encode()

	http.Redirect(w, r, authURL.String(), http.StatusFound)
}

// HandleCallback receives the authorization code, validates PKCE, exchanges the code for a PQC token,
// and sets an encrypted session cookie.
func (s *Server) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errParam := r.URL.Query().Get("error")

	if errParam != "" {
		http.Error(w, "oauth error: "+errParam, http.StatusBadRequest)
		return
	}
	if code == "" || state == "" {
		http.Error(w, "missing code or state parameter", http.StatusBadRequest)
		return
	}

	// Read and validate PKCE cookie
	cookie, err := r.Cookie(pkceCookieName)
	if err != nil || cookie.Value == "" {
		http.Error(w, "missing or expired PKCE session cookie", http.StatusBadRequest)
		return
	}

	// Clear PKCE cookie immediately after single use
	http.SetCookie(w, &http.Cookie{
		Name:     pkceCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	pkceBytes, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "invalid PKCE session cookie", http.StatusBadRequest)
		return
	}

	var sess pkceSession
	if err := json.Unmarshal(pkceBytes, &sess); err != nil {
		http.Error(w, "failed to decode PKCE session", http.StatusBadRequest)
		return
	}

	if sess.State != state {
		http.Error(w, "state mismatch (possible CSRF attack)", http.StatusBadRequest)
		return
	}

	// Perform Code Exchange at DilithiAuth IdP /token endpoint
	tokenForm := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {s.config.ClientID},
		"code":          {code},
		"redirect_uri":  {s.config.RedirectURI},
		"code_verifier": {sess.CodeVerifier},
	}
	if s.config.ClientSecret != "" {
		tokenForm.Set("client_secret", s.config.ClientSecret)
	}

	tokenResp, err := http.PostForm(s.config.IssuerURL+"/token", tokenForm)
	if err != nil {
		http.Error(w, "failed to reach token endpoint: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		var errRes map[string]string
		_ = json.NewDecoder(tokenResp.Body).Decode(&errRes)
		http.Error(w, "token exchange failed: "+errRes["error_description"], http.StatusBadRequest)
		return
	}

	var tokenResult map[string]interface{}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		http.Error(w, "failed to parse token response", http.StatusInternalServerError)
		return
	}

	accessToken, _ := tokenResult["access_token"].(string)
	pqcRef, _ := tokenResult["pqc_ref"].(string)
	if accessToken == "" {
		http.Error(w, "access token missing from response", http.StatusInternalServerError)
		return
	}

	// Verify the Post-Quantum token using Verifier
	var subject string
	if s.v != nil {
		claims, verifyErr := s.v.VerifyJWT(accessToken)
		if verifyErr == nil {
			subject = claims.Subject
		}
	}
	if subject == "" {
		// Parse unverified fallback for display
		parser := jwt.NewParser()
		tok, _, _ := parser.ParseUnverified(accessToken, &oauth.HybridClaims{})
		if claims, ok := tok.Claims.(*oauth.HybridClaims); ok {
			subject = claims.Subject
		}
	}

	// Fetch the real name/email from /userinfo (scope requested above is
	// "openid profile email", so both should be present). Best-effort: the
	// session still works with just the bare subject if this fails, it just
	// won't have a human-readable display name/email.
	var displayName, email string
	if req, reqErr := http.NewRequest(http.MethodGet, s.config.IssuerURL+"/userinfo", nil); reqErr == nil {
		req.Header.Set("Authorization", "Bearer "+accessToken)
		if userinfoResp, uiErr := http.DefaultClient.Do(req); uiErr == nil {
			defer userinfoResp.Body.Close()
			if userinfoResp.StatusCode == http.StatusOK {
				var userinfo struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}
				if json.NewDecoder(userinfoResp.Body).Decode(&userinfo) == nil {
					displayName = userinfo.Name
					email = userinfo.Email
				}
			}
		}
	}

	// Create signed session cookie for the BFF
	csrfBytes := make([]byte, 16)
	if _, err := rand.Read(csrfBytes); err != nil {
		http.Error(w, "failed to create session cookie", http.StatusInternalServerError)
		return
	}
	now := time.Now()
	sessionClaims := userSessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.ClientID,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.TokenTTL)),
		},
		Subject:   subject,
		Username:  displayName,
		Email:     email,
		PQCRef:    pqcRef,
		CSRFToken: base64.RawURLEncoding.EncodeToString(csrfBytes),
	}

	sessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, sessionClaims)
	sessTokenStr, err := sessToken.SignedString(s.config.SessionSecret)
	if err != nil {
		http.Error(w, "failed to create session cookie", http.StatusInternalServerError)
		return
	}

	// Set HttpOnly, Secure, SameSite session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessTokenStr,
		Path:     "/",
		Expires:  now.Add(s.config.TokenTTL),
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.config.IssuerURL, "https"),
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to frontend dashboard
	http.Redirect(w, r, s.config.FrontendURL, http.StatusFound)
}

// HandleMe returns the current authenticated user's session claims.
func (s *Server) HandleMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &userSessionClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.config.SessionSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "invalid session cookie", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*userSessionClaims)
	if !ok {
		http.Error(w, "invalid session claims", http.StatusUnauthorized)
		return
	}

	email := claims.Email
	if email == "" {
		email = claims.Subject
	}
	name := claims.Username
	if name == "" {
		name = email
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"id":            claims.Subject,
		"email":         email,
		"name":          name,
		"subject":       claims.Subject,
		"pqc_ref":       claims.PQCRef,
		"expires_at":    claims.ExpiresAt.Time,
		"csrf_token":    claims.CSRFToken,
	})
}

// HandleLogout destroys the session cookie.
func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.config.IssuerURL, "https"),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}
