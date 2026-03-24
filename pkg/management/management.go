// Package management provides the DilithiAuth Management API.
//
// The Management API is a separate REST service (default :9090) that lets
// operators programmatically provision tenants, register OAuth clients, and
// create users. It is intended to be deployed alongside the Identity Provider
// and secured behind an API key (MANAGEMENT_API_KEY environment variable).
//
// Endpoints:
//
//	POST   /api/v1/tenants                         Create a new tenant
//	GET    /api/v1/tenants                         List all tenants
//	GET    /api/v1/tenants/{tenantID}               Get a specific tenant
//	POST   /api/v1/tenants/{tenantID}/clients       Register an OAuth client
//	GET    /api/v1/tenants/{tenantID}/clients       List OAuth clients
//	DELETE /api/v1/tenants/{tenantID}/clients/{id}  Remove an OAuth client
//	POST   /api/v1/tenants/{tenantID}/users         Register a new user
package management

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/userstore"
)

// Handler holds the management API dependencies.
type Handler struct {
	Store     TenantStore
	UserStore *userstore.PostgresUserStore // optional; nil disables user registration
	APIKey    string                       // required for all requests
}

// Router returns a chi.Router mounting all management endpoints.
func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(h.requireAPIKey)

	r.Post("/api/v1/tenants", h.createTenant)
	r.Get("/api/v1/tenants", h.listTenants)
	r.Get("/api/v1/tenants/{tenantID}", h.getTenant)

	r.Post("/api/v1/tenants/{tenantID}/clients", h.addClient)
	r.Get("/api/v1/tenants/{tenantID}/clients", h.listClients)
	r.Delete("/api/v1/tenants/{tenantID}/clients/{clientID}", h.deleteClient)

	r.Post("/api/v1/tenants/{tenantID}/users", h.registerUser)

	return r
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func (h *Handler) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.APIKey == "" {
			// No key configured — allow all (development mode).
			next.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		key := strings.TrimPrefix(auth, "Bearer ")
		if key == "" {
			key = r.Header.Get("X-API-Key")
		}
		if key != h.APIKey {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------------
// Tenant handlers
// ---------------------------------------------------------------------------

func (h *Handler) createTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Slug == "" {
		jsonError(w, "name and slug are required", http.StatusBadRequest)
		return
	}

	t, err := h.Store.CreateTenant(r.Context(), req.Name, req.Slug)
	if err != nil {
		jsonError(w, "failed to create tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, t, http.StatusCreated)
}

func (h *Handler) listTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.Store.ListTenants(r.Context())
	if err != nil {
		jsonError(w, "failed to list tenants: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tenants == nil {
		tenants = []*Tenant{}
	}
	jsonResponse(w, tenants, http.StatusOK)
}

func (h *Handler) getTenant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tenantID")
	t, err := h.Store.GetTenant(r.Context(), id)
	if err != nil {
		if err.Error() == "tenant not found" {
			jsonError(w, "tenant not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to get tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, t, http.StatusOK)
}

// ---------------------------------------------------------------------------
// OAuth client handlers
// ---------------------------------------------------------------------------

func (h *Handler) addClient(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	var req struct {
		Name        string `json:"name"`
		ClientID    string `json:"client_id"`
		RedirectURI string `json:"redirect_uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.ClientID == "" || req.RedirectURI == "" {
		jsonError(w, "name, client_id, and redirect_uri are required", http.StatusBadRequest)
		return
	}

	client, err := h.Store.AddClient(r.Context(), tenantID, req.Name, req.ClientID, req.RedirectURI)
	if err != nil {
		jsonError(w, "failed to add client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, client, http.StatusCreated)
}

func (h *Handler) listClients(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	clients, err := h.Store.ListClients(r.Context(), tenantID)
	if err != nil {
		jsonError(w, "failed to list clients: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if clients == nil {
		clients = []*ManagedClient{}
	}
	jsonResponse(w, clients, http.StatusOK)
}

func (h *Handler) deleteClient(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	clientID := chi.URLParam(r, "clientID")
	if err := h.Store.DeleteClient(r.Context(), tenantID, clientID); err != nil {
		if err.Error() == "client not found" {
			jsonError(w, "client not found", http.StatusNotFound)
			return
		}
		jsonError(w, "failed to delete client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// User registration handler
// ---------------------------------------------------------------------------

func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {
	if h.UserStore == nil {
		jsonError(w, "user store not configured", http.StatusNotImplemented)
		return
	}
	tenantID := chi.URLParam(r, "tenantID")
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		jsonError(w, "username, email, and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.UserStore.Register(r.Context(), tenantID, req.Username, req.Email, req.Password)
	if err != nil {
		jsonError(w, "failed to register user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Never return the password hash.
	jsonResponse(w, map[string]any{
		"id":         user.ID,
		"tenant_id":  user.TenantID,
		"username":   user.Username,
		"email":      user.Email,
		"subject":    user.Subject,
		"created_at": user.CreatedAt,
	}, http.StatusCreated)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func jsonResponse(w http.ResponseWriter, body any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	jsonResponse(w, map[string]string{"error": msg}, status)
}
