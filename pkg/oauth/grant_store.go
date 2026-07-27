package oauth

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// recognizedScopes is the fixed set of scopes DilithiAuth understands. A
// request for anything outside this set is rejected with invalid_scope
// (RFC 6749 §5.2) rather than silently accepted and ignored.
var recognizedScopes = map[string]bool{
	"openid":         true,
	"profile":        true,
	"email":          true,
	"offline_access": true,
}

// validateScope splits a space-delimited scope string and rejects it if any
// requested scope isn't recognized. An empty scope string is always valid
// (no scopes requested).
func validateScope(scope string) error {
	for _, s := range strings.Fields(scope) {
		if !recognizedScopes[s] {
			return errors.New("unknown scope: " + s)
		}
	}
	return nil
}

// scopeSet parses a space-delimited scope string into a lookup set.
func scopeSet(scope string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range strings.Fields(scope) {
		set[s] = true
	}
	return set
}

// hasScope reports whether scope contains the given individual scope value.
func hasScope(scope, want string) bool {
	return scopeSet(scope)[want]
}

// ScopeContains reports whether a space-delimited scope string contains the
// given individual scope value. Exported for callers outside pkg/oauth
// (e.g. the /userinfo handler) that need to gate claims by granted scope.
func ScopeContains(scope, want string) bool {
	return hasScope(scope, want)
}

// mergeScopes returns the union of two space-delimited scope strings as a
// single space-delimited string (used when a repeat consent grants
// additional scopes on top of ones already approved).
func mergeScopes(a, b string) string {
	set := scopeSet(a)
	for s := range scopeSet(b) {
		set[s] = true
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	return strings.Join(out, " ")
}

// Grant records that a subject has approved a client's access to a set of
// scopes, so the consent screen can be skipped on subsequent logins unless
// new scopes are requested (matching Google/GitHub's OAuth UX).
type Grant struct {
	Subject   string
	TenantID  string
	ClientID  string
	Scope     string // space-delimited granted scopes
	GrantedAt time.Time
}

// GrantStore manages consent grants.
type GrantStore interface {
	// Get returns the existing grant for (tenantID, subject, clientID), or
	// nil (with no error) if none exists yet.
	Get(tenantID, subject, clientID string) (*Grant, error)
	// Save creates or overwrites the grant for (tenantID, subject, clientID).
	Save(g *Grant) error
}

// MemoryGrantStore is an in-memory GrantStore.
type MemoryGrantStore struct {
	mu     sync.RWMutex
	grants map[string]*Grant // key: tenantID + "|" + subject + "|" + clientID
}

// NewMemoryGrantStore creates an empty MemoryGrantStore.
func NewMemoryGrantStore() *MemoryGrantStore {
	return &MemoryGrantStore{grants: make(map[string]*Grant)}
}

func grantKey(tenantID, subject, clientID string) string {
	return tenantID + "|" + subject + "|" + clientID
}

func (s *MemoryGrantStore) Get(tenantID, subject, clientID string) (*Grant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.grants[grantKey(tenantID, subject, clientID)], nil
}

func (s *MemoryGrantStore) Save(g *Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grants[grantKey(g.TenantID, g.Subject, g.ClientID)] = g
	return nil
}

// grantCoversScope reports whether an existing grant (which may be nil)
// already covers every scope in requestedScope.
func grantCoversScope(g *Grant, requestedScope string) bool {
	if g == nil {
		return requestedScope == ""
	}
	granted := scopeSet(g.Scope)
	for s := range scopeSet(requestedScope) {
		if !granted[s] {
			return false
		}
	}
	return true
}
