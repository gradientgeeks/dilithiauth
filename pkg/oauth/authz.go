package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// User represents an identity stored in the demo in-memory AuthzStore.
// Production deployments should use pkg/userstore.PostgresUserStore instead.
type User struct {
	Username string
	Password string // plaintext — demo only; see pkg/userstore for Argon2id
	Subject  string
}

// OAuthClient represents a registered OAuth 2.0 client application.
type OAuthClient struct {
	ClientID    string
	RedirectURI string
	Name        string
}

// AuthCode represents an issued, single-use authorization code.
type AuthCode struct {
	Code        string    `json:"code"`
	Subject     string    `json:"subject"`
	ClientID    string    `json:"client_id"`
	RedirectURI string    `json:"redirect_uri"`
	Scope       string    `json:"scope"`
	ExpiresAt   time.Time `json:"expires_at"`
	Used        bool      `json:"used"`
}

// AuthzStore manages OAuth clients and authorization codes, scoped per tenant.
// All methods receive tenantID as their first argument so a single store
// instance can serve multiple isolated tenants.
type AuthzStore interface {
	AddUser(tenantID string, u *User)
	AddClient(tenantID string, c *OAuthClient)
	AuthenticateUser(tenantID, username, password string) (*User, error)
	ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error)
	CreateCode(tenantID, subject, clientID, redirectURI, scope string) (string, error)
	RedeemCode(tenantID, code, clientID, redirectURI string) (*AuthCode, error)
}

// MemoryAuthzStore is a fully in-memory, multi-tenant AuthzStore.
// It is intended for development and testing only — passwords are plaintext
// and state is lost on restart.
type MemoryAuthzStore struct {
	mu      sync.RWMutex
	users   map[string]map[string]*User        // [tenantID][username]
	clients map[string]map[string]*OAuthClient // [tenantID][clientID]
	codes   map[string]*AuthCode               // keyed by code (globally unique)
}

// NewAuthzStore creates an empty MemoryAuthzStore.
func NewAuthzStore() *MemoryAuthzStore {
	return &MemoryAuthzStore{
		users:   make(map[string]map[string]*User),
		clients: make(map[string]map[string]*OAuthClient),
		codes:   make(map[string]*AuthCode),
	}
}

func (s *MemoryAuthzStore) AddUser(tenantID string, u *User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.users[tenantID] == nil {
		s.users[tenantID] = make(map[string]*User)
	}
	s.users[tenantID][u.Username] = u
}

func (s *MemoryAuthzStore) AddClient(tenantID string, c *OAuthClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clients[tenantID] == nil {
		s.clients[tenantID] = make(map[string]*OAuthClient)
	}
	s.clients[tenantID][c.ClientID] = c
}

func (s *MemoryAuthzStore) AuthenticateUser(tenantID, username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tenant := s.users[tenantID]
	if tenant == nil {
		return nil, errors.New("invalid credentials")
	}
	u, ok := tenant[username]
	if !ok || u.Password != password {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

func (s *MemoryAuthzStore) ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tenant := s.clients[tenantID]
	if tenant == nil {
		return nil, errors.New("unknown client")
	}
	c, ok := tenant[clientID]
	if !ok {
		return nil, errors.New("unknown client")
	}
	if c.RedirectURI != redirectURI {
		return nil, errors.New("redirect_uri mismatch")
	}
	return c, nil
}

func (s *MemoryAuthzStore) CreateCode(tenantID, subject, clientID, redirectURI, scope string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(buf)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code] = &AuthCode{
		Code:        code,
		Subject:     subject,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		Used:        false,
	}
	return code, nil
}

func (s *MemoryAuthzStore) RedeemCode(tenantID, code, clientID, redirectURI string) (*AuthCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ac, ok := s.codes[code]
	if !ok {
		return nil, errors.New("invalid code")
	}
	if ac.Used {
		return nil, errors.New("code already used")
	}
	if time.Now().After(ac.ExpiresAt) {
		return nil, errors.New("code expired")
	}
	if ac.ClientID != clientID {
		return nil, errors.New("client_id mismatch")
	}
	if ac.RedirectURI != redirectURI {
		return nil, errors.New("redirect_uri mismatch")
	}
	ac.Used = true
	return ac, nil
}
