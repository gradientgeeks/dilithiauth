package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// DemoUser represents a user in the demo identity store.
type DemoUser struct {
	Username string
	Password string
	Subject  string
}

// DemoClient represents a registered OAuth client.
type DemoClient struct {
	ClientID    string
	RedirectURI string
	Name        string
}

// AuthCode represents an issued authorization code.
type AuthCode struct {
	Code        string
	Subject     string
	ClientID    string
	RedirectURI string
	Scope       string
	ExpiresAt   time.Time
	Used        bool
}

// AuthzStore holds demo users, clients, and authorization codes.
type AuthzStore struct {
	mu      sync.RWMutex
	users   map[string]*DemoUser  // keyed by username
	clients map[string]*DemoClient // keyed by client_id
	codes   map[string]*AuthCode   // keyed by code
}

// NewAuthzStore creates an empty AuthzStore.
func NewAuthzStore() *AuthzStore {
	return &AuthzStore{
		users:   make(map[string]*DemoUser),
		clients: make(map[string]*DemoClient),
		codes:   make(map[string]*AuthCode),
	}
}

// AddUser registers a demo user.
func (s *AuthzStore) AddUser(u *DemoUser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.Username] = u
}

// AddClient registers a demo client.
func (s *AuthzStore) AddClient(c *DemoClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c.ClientID] = c
}

// AuthenticateUser checks username/password and returns the user on success.
func (s *AuthzStore) AuthenticateUser(username, password string) (*DemoUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[username]
	if !ok || u.Password != password {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// ValidateClient checks that the client_id exists and the redirect_uri matches.
func (s *AuthzStore) ValidateClient(clientID, redirectURI string) (*DemoClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.clients[clientID]
	if !ok {
		return nil, errors.New("unknown client")
	}
	if c.RedirectURI != redirectURI {
		return nil, errors.New("redirect_uri mismatch")
	}
	return c, nil
}

// CreateCode generates and stores an authorization code.
func (s *AuthzStore) CreateCode(subject, clientID, redirectURI, scope string) (string, error) {
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
	}
	return code, nil
}

// RedeemCode validates and consumes an authorization code. It checks expiry,
// single-use, and client_id/redirect_uri match.
func (s *AuthzStore) RedeemCode(code, clientID, redirectURI string) (*AuthCode, error) {
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
