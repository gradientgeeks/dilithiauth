package oauth

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// SigStorePutter is the interface for storing PQC signatures, satisfied by
// *transport.SigStore without importing the transport package.
type SigStorePutter interface {
	Put(jti string, sig []byte)
}

// AuthzHandlers provides HTTP handlers for the OAuth 2.0 authorization code flow.
type AuthzHandlers struct {
	Provider *Provider
	Store    AuthzStore
	SigStore SigStorePutter
}

var loginTmpl = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html>
<head><title>DilithiAuth Login</title></head>
<body>
<h1>DilithiAuth Login</h1>
<form method="POST" action="/authorize">
  <input type="hidden" name="client_id" value="{{.ClientID}}">
  <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
  <input type="hidden" name="state" value="{{.State}}">
  <input type="hidden" name="scope" value="{{.Scope}}">
  <label>Username: <input type="text" name="username"></label><br>
  <label>Password: <input type="password" name="password"></label><br>
  <button type="submit">Login</button>
</form>
</body>
</html>`))

// AuthorizeGET renders the login form.
func (h *AuthzHandlers) AuthorizeGET(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")

	if clientID == "" || redirectURI == "" {
		http.Error(w, "missing client_id or redirect_uri", http.StatusBadRequest)
		return
	}

	if _, err := h.Store.ValidateClient(clientID, redirectURI); err != nil {
		http.Error(w, "invalid client: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_ = loginTmpl.Execute(w, map[string]string{
		"ClientID":    clientID,
		"RedirectURI": redirectURI,
		"State":       state,
		"Scope":       scope,
	})
}

// AuthorizePOST authenticates the user and redirects with an authorization code.
func (h *AuthzHandlers) AuthorizePOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	scope := r.FormValue("scope")

	user, err := h.Store.AuthenticateUser(username, password)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	code, err := h.Store.CreateCode(user.Subject, clientID, redirectURI, scope)
	if err != nil {
		http.Error(w, "code generation failed", http.StatusInternalServerError)
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

	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")

	ac, err := h.Store.RedeemCode(code, clientID, redirectURI)
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

	// Store the PQC signature in the sidecar.
	h.SigStore.Put(result.PQCRef, result.PQCSig)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": result.Token,
		"token_type":   "Bearer",
		"pqc_ref":      result.PQCRef,
	})
}
