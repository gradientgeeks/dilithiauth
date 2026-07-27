package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// bcryptDummyHash is used for constant-time comparisons when a user is not
// found, preventing timing-based user enumeration.
var bcryptDummyHash []byte

func init() {
	var err error
	bcryptDummyHash, err = bcrypt.GenerateFromPassword([]byte("_dummy_dilithiauth_"), bcrypt.DefaultCost)
	if err != nil {
		panic("oauth: initializing bcrypt dummy hash: " + err.Error())
	}
}

// User represents an identity stored in the demo in-memory AuthzStore.
// Password is stored as a bcrypt hash; callers always provide plaintext.
// Production deployments should use pkg/userstore.PostgresUserStore instead.
type User struct {
	Username string
	Email    string
	Password string // bcrypt hash — AddUser hashes the plaintext on entry
	Subject  string
}

// OAuthClient represents a registered OAuth 2.0 client application.
type OAuthClient struct {
	ClientID         string
	ClientSecretHash string // bcrypt hash; "" = public PKCE-only client
	RedirectURI      string
	Name             string
	DeveloperID      string // owner developer account ID (empty for demo/seed clients)
}

// AuthCode represents an issued, single-use authorization code.
type AuthCode struct {
	Code                string    `json:"code"`
	Subject             string    `json:"subject"`
	ClientID            string    `json:"client_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	ExpiresAt           time.Time `json:"expires_at"`
	Used                bool      `json:"used"`
	CodeChallenge       string    `json:"code_challenge,omitempty"`        // PKCE S256 challenge
	CodeChallengeMethod string    `json:"code_challenge_method,omitempty"` // always "S256" when set
}

// AuthzStore manages OAuth clients and authorization codes, scoped per tenant.
// All methods receive tenantID as their first argument so a single store
// instance can serve multiple isolated tenants.
type AuthzStore interface {
	AddUser(tenantID string, u *User)
	AddClient(tenantID string, c *OAuthClient)
	AuthenticateUser(tenantID, username, password string) (*User, error)
	ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error)
	// CreateCode creates an authorization code. codeChallenge and codeChallengeMethod
	// are optional PKCE parameters; pass empty strings to skip PKCE binding.
	CreateCode(tenantID, subject, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod string) (string, error)
	// RedeemCode validates and consumes an authorization code. codeVerifier is the
	// PKCE verifier; pass an empty string when PKCE was not used.
	RedeemCode(tenantID, code, clientID, redirectURI, codeVerifier string) (*AuthCode, error)
	// ValidateClientWithSecret authenticates a confidential client. For public
	// clients (ClientSecretHash == ""), clientSecret must be empty.
	ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI string) (*OAuthClient, error)
	// ValidateClientSecretOnly authenticates a client by client_id+client_secret
	// alone, with no redirect_uri check — used by the refresh_token grant.
	ValidateClientSecretOnly(tenantID, clientID, clientSecret string) (*OAuthClient, error)
	// RegisterUser creates a new user account. Returns an error if the username
	// or email is already taken.
	RegisterUser(tenantID string, u *User) error
	// GetUserBySubject looks up a user by their stable JWT sub claim.
	GetUserBySubject(tenantID, subject string) (*User, error)
	// GetUserByEmail looks up a user by email, used by the password-reset flow.
	GetUserByEmail(tenantID, email string) (*User, error)
	// SetPassword updates a user's password (bcrypt hash), used by the
	// password-reset flow.
	SetPassword(tenantID, username, newPasswordHash string) error
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
	// Hash the password before acquiring the lock (bcrypt is expensive).
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		panic("oauth: hashing user password: " + err.Error())
	}
	sub := u.Subject
	if sub == "" {
		sub = u.Username
	}
	stored := &User{Username: u.Username, Email: u.Email, Password: string(hash), Subject: sub}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.users[tenantID] == nil {
		s.users[tenantID] = make(map[string]*User)
	}
	s.users[tenantID][u.Username] = stored
}

// RegisterUser creates a new user. Returns an error if the username is already taken.
func (s *MemoryAuthzStore) RegisterUser(tenantID string, u *User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	sub := u.Subject
	if sub == "" {
		sub = u.Username
	}
	stored := &User{Username: u.Username, Email: u.Email, Password: string(hash), Subject: sub}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.users[tenantID] == nil {
		s.users[tenantID] = make(map[string]*User)
	}
	for _, existing := range s.users[tenantID] {
		if existing.Username == u.Username {
			return errors.New("username already taken")
		}
		if u.Email != "" && existing.Email == u.Email {
			return errors.New("email already registered")
		}
	}
	s.users[tenantID][u.Username] = stored
	return nil
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
		// Dummy compare prevents timing oracle on tenant/user existence.
		_ = bcrypt.CompareHashAndPassword(bcryptDummyHash, []byte(password))
		return nil, errors.New("invalid credentials")
	}
	u, ok := tenant[username]
	if !ok {
		_ = bcrypt.CompareHashAndPassword(bcryptDummyHash, []byte(password))
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// GetUserBySubject looks up a user by their stable JWT sub claim within a tenant.
func (s *MemoryAuthzStore) GetUserBySubject(tenantID, subject string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users[tenantID] {
		if u.Subject == subject {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

// GetUserByEmail looks up a user by email within a tenant.
func (s *MemoryAuthzStore) GetUserByEmail(tenantID, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users[tenantID] {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

// SetPassword updates a user's password (bcrypt hash) within a tenant.
func (s *MemoryAuthzStore) SetPassword(tenantID, username, newPasswordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tenant := s.users[tenantID]
	if tenant == nil {
		return errors.New("user not found")
	}
	u, ok := tenant[username]
	if !ok {
		return errors.New("user not found")
	}
	u.Password = newPasswordHash
	return nil
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

func (s *MemoryAuthzStore) ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI string) (*OAuthClient, error) {
	client, err := s.ValidateClient(tenantID, clientID, redirectURI)
	if err != nil {
		// Dummy compare prevents timing oracle on client existence.
		_ = bcrypt.CompareHashAndPassword(bcryptDummyHash, []byte(clientSecret))
		return nil, err
	}
	if client.ClientSecretHash == "" {
		// Public client (PKCE-only): no secret expected.
		if clientSecret != "" {
			return nil, errors.New("public client does not accept client_secret")
		}
		return client, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret)); err != nil {
		return nil, errors.New("invalid client_secret")
	}
	return client, nil
}

// ValidateClientSecretOnly authenticates a client by client_id+client_secret
// alone, with no redirect_uri check (used by the refresh_token grant).
func (s *MemoryAuthzStore) ValidateClientSecretOnly(tenantID, clientID, clientSecret string) (*OAuthClient, error) {
	s.mu.RLock()
	tenantClients := s.clients[tenantID]
	var client *OAuthClient
	if tenantClients != nil {
		client = tenantClients[clientID]
	}
	s.mu.RUnlock()

	if client == nil {
		_ = bcrypt.CompareHashAndPassword(bcryptDummyHash, []byte(clientSecret))
		return nil, errors.New("unknown client")
	}
	if client.ClientSecretHash == "" {
		if clientSecret != "" {
			return nil, errors.New("public client does not accept client_secret")
		}
		return client, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret)); err != nil {
		return nil, errors.New("invalid client_secret")
	}
	return client, nil
}

func (s *MemoryAuthzStore) CreateCode(tenantID, subject, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod string) (string, error) {
	if codeChallenge != "" && codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}
	if codeChallenge != "" && codeChallengeMethod != "S256" {
		return "", errors.New("pkce: only S256 code_challenge_method is supported")
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(buf)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code] = &AuthCode{
		Code:                code,
		Subject:             subject,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}
	return code, nil
}

func (s *MemoryAuthzStore) RedeemCode(tenantID, code, clientID, redirectURI, codeVerifier string) (*AuthCode, error) {
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
	if ac.CodeChallenge != "" {
		if codeVerifier == "" {
			return nil, errors.New("pkce: code_verifier required")
		}
		h := sha256.Sum256([]byte(codeVerifier))
		computed := base64.RawURLEncoding.EncodeToString(h[:])
		if subtle.ConstantTimeCompare([]byte(computed), []byte(ac.CodeChallenge)) != 1 {
			return nil, errors.New("pkce: code_verifier mismatch")
		}
	}
	ac.Used = true
	return ac, nil
}

// CompositeClientValidator checks Primary ClientValidator first, falling back to Secondary.
type CompositeClientValidator struct {
	Primary   ClientValidator
	Secondary ClientValidator
}

func (c *CompositeClientValidator) ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error) {
	if c.Primary != nil {
		client, err := c.Primary.ValidateClient(tenantID, clientID, redirectURI)
		if err == nil {
			return client, nil
		}
	}
	if c.Secondary != nil {
		return c.Secondary.ValidateClient(tenantID, clientID, redirectURI)
	}
	return nil, errors.New("unknown client")
}

func (c *CompositeClientValidator) ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI string) (*OAuthClient, error) {
	if c.Primary != nil {
		client, err := c.Primary.ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI)
		if err == nil {
			return client, nil
		}
	}
	if c.Secondary != nil {
		return c.Secondary.ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI)
	}
	return nil, errors.New("unknown client")
}

func (c *CompositeClientValidator) ValidateClientSecretOnly(tenantID, clientID, clientSecret string) (*OAuthClient, error) {
	if c.Primary != nil {
		client, err := c.Primary.ValidateClientSecretOnly(tenantID, clientID, clientSecret)
		if err == nil {
			return client, nil
		}
	}
	if c.Secondary != nil {
		return c.Secondary.ValidateClientSecretOnly(tenantID, clientID, clientSecret)
	}
	return nil, errors.New("unknown client")
}
