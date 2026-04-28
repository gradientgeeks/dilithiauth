package developer

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Handlers provides REST API handlers for the developer portal.
// All endpoints return JSON. Authentication uses a JWT session cookie
// set at login/register and validated by the requireSession middleware.
type Handlers struct {
	Store         *Store
	SessionSecret []byte
}

// Router mounts all developer API routes under a sub-router.
// Mount this at /api/v1/developer in the main server.
func (h *Handlers) Router() http.Handler {
	r := chi.NewRouter()

	// Public endpoints
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)

	// Authenticated endpoints
	r.Group(func(r chi.Router) {
		r.Use(h.requireSession)
		r.Get("/me", h.Me)
		r.Get("/apps", h.ListApps)
		r.Post("/apps", h.CreateApp)
		r.Get("/apps/{id}", h.GetApp)
		r.Post("/apps/{id}/rotate-secret", h.RotateSecret)
		r.Delete("/apps/{id}", h.DeleteApp)
	})

	return r
}

// requireSession validates the JWT cookie and injects SessionClaims into context.
func (h *Handlers) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := SessionFromRequest(r, h.SessionSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, WithSessionClaims(r, claims))
	})
}

// Register creates a new developer account and issues a session cookie.
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
	dev, err := h.Store.RegisterDeveloper(r.Context(), body.Email, body.Name, body.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := IssueSessionCookie(w, h.SessionSecret, dev); err != nil {
		writeError(w, http.StatusInternalServerError, "session error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":    dev.ID,
		"email": dev.Email,
		"name":  dev.Name,
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
	if err := IssueSessionCookie(w, h.SessionSecret, dev); err != nil {
		writeError(w, http.StatusInternalServerError, "session error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":    dev.ID,
		"email": dev.Email,
		"name":  dev.Name,
	})
}

// Logout clears the session cookie.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me returns the authenticated developer's profile.
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"id":    claims.DeveloperID,
		"email": claims.Email,
		"name":  claims.Name,
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
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, app)
}

// RotateSecret generates a new client_secret for a confidential app.
// The new secret is returned in this response ONLY.
func (h *Handlers) RotateSecret(w http.ResponseWriter, r *http.Request) {
	claims := SessionClaimsFromContext(r.Context())
	appID := chi.URLParam(r, "id")

	app, secret, err := h.Store.RotateSecret(r.Context(), claims.DeveloperID, appID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
	appID := chi.URLParam(r, "id")

	if err := h.Store.DeleteApp(r.Context(), claims.DeveloperID, appID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
