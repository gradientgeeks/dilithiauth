package developer

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/ratelimit"
)

// Handlers provides REST API handlers for the developer portal.
// All endpoints return JSON. Authentication uses a JWT session cookie
// set at login/register and validated by the requireSession middleware.
type Handlers struct {
	Store DevStore

	OTPStore      *OTPVerificationStore
	SessionSecret []byte

	// SessionRevocation enables server-side session invalidation on logout.
	// Sessions are otherwise stateless JWTs valid for their full 24h MaxAge
	// regardless of the cookie being cleared client-side. Optional; if nil,
	// Logout only clears the cookie and revocation checks are skipped.
	SessionRevocation SessionRevocationStore
}

// Router mounts all developer API routes under a sub-router.
// Mount this at /api/v1/developer in the main server.
//
// Public endpoints are rate-limited per client IP to blunt brute-force and
// enumeration attacks: /resend-code and /login guard against OTP/password
// guessing, /register and /verify-email are limited more loosely. Limiters
// are created fresh per Router() call (not shared package-level state) so
// each Handlers instance — including one built per test — has its own budget.
func (h *Handlers) Router() http.Handler {
	r := chi.NewRouter()

	rl := func(l *ratelimit.Limiter) func(http.Handler) http.Handler {
		return ratelimit.Middleware(l, ratelimit.ClientIPKey())
	}

	// Public endpoints
	r.With(rl(ratelimit.NewLimiter(1.0/10, 5))).Post("/register", h.Register)       // ~1 req/10s sustained, burst 5
	r.With(rl(ratelimit.NewLimiter(1.0/5, 5))).Post("/login", h.Login)              // ~1 req/5s sustained, burst 5
	r.With(rl(ratelimit.NewLimiter(1.0/5, 8))).Post("/verify-email", h.VerifyEmail) // ~1 req/5s sustained, burst 8
	r.With(rl(ratelimit.NewLimiter(1.0/20, 3))).Post("/resend-code", h.ResendCode)  // ~1 req/20s sustained, burst 3

	// Authenticated endpoints
	r.Group(func(r chi.Router) {
		r.Use(h.requireSession)
		r.Post("/logout", h.Logout)
		r.Get("/me", h.Me)
		r.Get("/apps", h.ListApps)
		r.Post("/apps", h.CreateApp)
		r.Get("/apps/{id}", h.GetApp)
		r.Post("/apps/{id}/rotate-secret", h.RotateSecret)
		r.Delete("/apps/{id}", h.DeleteApp)
	})

	return r
}

// requireSession validates either the developer JWT cookie (da_sess) or BFF session cookie (dilithiauth_bff_session).
func (h *Handlers) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := SessionFromRequest(r, h.SessionSecret)
		if err == nil {
			if h.sessionRevoked(claims.ID) {
				writeError(w, http.StatusUnauthorized, "session has been revoked")
				return
			}
			next.ServeHTTP(w, WithSessionClaims(r, claims))
			return
		}

		// Fallback to checking BFF session cookie if da_sess is not set.
		// Uses the same ParseSessionJWT helper (and thus the same
		// jwt.WithValidMethods([]string{"HS256"}) restriction) as the
		// primary da_sess path above, so the two can't drift out of sync.
		if bffCookie, bffErr := r.Cookie("dilithiauth_bff_session"); bffErr == nil && bffCookie.Value != "" {
			if c, pErr := ParseSessionJWT(bffCookie.Value, h.SessionSecret); pErr == nil {
				if c.DeveloperID == "" {
					// The BFF authenticates against the end-user account
					// system (pkg/oauth), not developer_accounts — its
					// session carries an end-user Subject (e.g. "usr_..."),
					// which is never a valid developer_accounts.id. Using it
					// directly as DeveloperID would violate the
					// oauth_apps_developer_id_fkey constraint the first time
					// this identity tries to create an app. Instead, resolve
					// (or just-in-time provision) the matching developer
					// account by the verified email the BFF now supplies.
					if c.Email != "" {
						dev, devErr := h.Store.GetDeveloperByEmail(r.Context(), c.Email)
						if devErr != nil || dev == nil {
							name := c.Name
							if name == "" {
								name = c.Email
							}
							dev, devErr = h.Store.RegisterDeveloper(r.Context(), c.Email, name, mustRandomHex(16))
							if devErr != nil || dev == nil {
								// Lost a race with a concurrent request
								// provisioning the same account.
								dev, devErr = h.Store.GetDeveloperByEmail(r.Context(), c.Email)
							} else {
								_ = h.Store.MarkEmailVerified(r.Context(), c.Email)
							}
						}
						if dev != nil {
							c.DeveloperID = dev.ID
						}
					}
					if c.DeveloperID == "" {
						writeError(w, http.StatusUnauthorized, "unable to resolve developer account for this session")
						return
					}
				}
				if h.sessionRevoked(c.ID) {
					writeError(w, http.StatusUnauthorized, "session has been revoked")
					return
				}
				next.ServeHTTP(w, WithSessionClaims(r, c))
				return
			}
		}

		writeError(w, http.StatusUnauthorized, "authentication required")
	})
}

// sessionRevoked reports whether the given session jti has been revoked.
// Sessions issued before SessionRevocation tracked jtis (or when it is nil)
// have an empty ID and are never treated as revoked.
func (h *Handlers) sessionRevoked(jti string) bool {
	return h.SessionRevocation != nil && jti != "" && h.SessionRevocation.IsRevoked(jti)
}

// Register creates a new developer account and triggers email verification code generation.
// Does NOT issue a session cookie — the developer MUST verify their email first.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if _, err := mail.ParseAddress(body.Email); err != nil {
		writeError(w, http.StatusBadRequest, "invalid email address")
		return
	}
	if len(body.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	dev, err := h.Store.RegisterDeveloper(r.Context(), body.Email, body.Name, body.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.OTPStore != nil {
		_, _, _ = h.OTPStore.GenerateCode(dev.Email)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":                     dev.ID,
		"email":                  dev.Email,
		"name":                   dev.Name,
		"verification_code_sent": true,
	})
}

// Login authenticates a developer and issues a session cookie.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	dev, err := h.Store.AuthenticateDeveloper(r.Context(), body.Email, body.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	csrfToken, err := IssueSessionCookie(w, h.SessionSecret, dev)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         dev.ID,
		"email":      dev.Email,
		"name":       dev.Name,
		"csrf_token": csrfToken,
	})
}

// Logout clears the session cookie and, when a SessionRevocation store is
// configured, revokes the session's jti server-side so it can't be replayed
// for the remainder of its 24h validity even if the cookie is retained.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if h.SessionRevocation != nil {
		claims := SessionClaimsFromContext(r.Context())
		if claims != nil && claims.ID != "" {
			expiry := time.Now().Add(sessionTTL)
			if claims.ExpiresAt != nil {
				expiry = claims.ExpiresAt.Time
			}
			_ = h.SessionRevocation.Revoke(claims.ID, expiry)
		}
	}
	ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me returns the authenticated developer's profile, including the session's
// CSRF token — the frontend must echo this back via the X-CSRF-Token header
// on state-changing requests (CreateApp, RotateSecret, DeleteApp).
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         claims.DeveloperID,
		"email":      claims.Email,
		"name":       claims.Name,
		"csrf_token": claims.CSRFToken,
	})
}

// ListApps returns all OAuth apps owned by the authenticated developer.
func (h *Handlers) ListApps(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	apps, err := h.Store.ListApps(r.Context(), claims.DeveloperID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list apps")
		return
	}
	writeJSON(w, http.StatusOK, apps)
}

// CreateApp registers a new OAuth app and returns credentials.
// The client_secret is included in this response ONLY — it is never returned again.
func (h *Handlers) CreateApp(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	if !ValidateCSRF(r, claims) {
		writeError(w, http.StatusForbidden, "missing or invalid CSRF token")
		return
	}

	var body struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		HomepageURL  string   `json:"homepage_url"`
		RedirectURIs []string `json:"redirect_uris"`
		IsPublic     bool     `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Accept redirect URIs as either a pre-split slice or a newline-delimited string.
	uris := filterEmpty(body.RedirectURIs)

	app, secret, err := h.Store.CreateApp(r.Context(),
		claims.DeveloperID, body.Name, body.Description, body.HomepageURL, uris, body.IsPublic)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := map[string]any{
		"app": app,
	}
	if secret != "" {
		resp["client_secret"] = secret // shown ONCE
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GetApp returns details of a single app owned by the authenticated developer.
func (h *Handlers) GetApp(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	appID := chi.URLParam(r, "id")

	app, err := h.Store.GetApp(r.Context(), claims.DeveloperID, appID)
	if err != nil {
		if errors.Is(err, ErrAppNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load app")
		}
		return
	}
	writeJSON(w, http.StatusOK, app)
}

// RotateSecret generates a new client_secret for a confidential app.
// The new secret is returned in this response ONLY.
func (h *Handlers) RotateSecret(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	if !ValidateCSRF(r, claims) {
		writeError(w, http.StatusForbidden, "missing or invalid CSRF token")
		return
	}
	appID := chi.URLParam(r, "id")

	app, secret, err := h.Store.RotateSecret(r.Context(), claims.DeveloperID, appID)
	if err != nil {
		if errors.Is(err, ErrAppNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			// Other errors here are either the "public apps don't have
			// secrets" business rule (client error) or a scan/DB failure
			// surfaced via GetApp — both are safe to show as 400, but note
			// this differs from GetApp/ListApps, which return 500 for
			// genuine internal errors since they have no client-error case.
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"app":           app,
		"client_secret": secret, // shown ONCE
	})
}

// DeleteApp removes an app owned by the authenticated developer.
func (h *Handlers) DeleteApp(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	if !ValidateCSRF(r, claims) {
		writeError(w, http.StatusForbidden, "missing or invalid CSRF token")
		return
	}
	appID := chi.URLParam(r, "id")

	if err := h.Store.DeleteApp(r.Context(), claims.DeveloperID, appID); err != nil {
		if errors.Is(err, ErrAppNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "failed to delete app")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// VerifyEmail validates the 6-digit OTP code submitted by the user and issues a session cookie on success.
func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Code == "" {
		writeError(w, http.StatusBadRequest, "email and verification code are required")
		return
	}

	if h.OTPStore != nil {
		valid, err := h.OTPStore.VerifyCode(body.Email, body.Code)
		if err != nil || !valid {
			msg := "invalid verification code"
			if err != nil {
				msg = err.Error()
			}
			writeError(w, http.StatusBadRequest, msg)
			return
		}
	}

	// Mark email as verified in developer store
	_ = h.Store.MarkEmailVerified(r.Context(), body.Email)

	// Issue session cookie now that email is verified! Load the real
	// developer record rather than fabricating one from the email — the
	// session's DeveloperID is the authorization key used by ListApps/
	// CreateApp/GetApp/RotateSecret/DeleteApp, and it must match the ID
	// Login uses for the same account, or apps created under one identity
	// become invisible under the other.
	dev, err := h.Store.GetDeveloperByEmail(r.Context(), body.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "verified but could not load account")
		return
	}
	csrfToken, err := IssueSessionCookie(w, h.SessionSecret, dev)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":     "verified",
		"message":    "Email verified successfully!",
		"csrf_token": csrfToken,
	})
}

// ResendCode generates a fresh 6-digit OTP code.
func (h *Handlers) ResendCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		writeError(w, http.StatusBadRequest, "valid email is required")
		return
	}

	if h.OTPStore != nil {
		if _, _, err := h.OTPStore.GenerateCode(body.Email); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func filterEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
