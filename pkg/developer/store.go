// Package developer provides the self-service developer portal for DilithiAuth.
// Developers register accounts, create OAuth apps, and receive client credentials.
package developer

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// ErrAppNotFound is returned by GetApp when no matching row exists, distinct
// from any other error (e.g. a scan/connection failure) so callers can tell
// "genuinely doesn't exist" (404) apart from "something went wrong" (500).
var ErrAppNotFound = errors.New("app not found")

// Argon2id parameters — match pkg/userstore for consistency.
const (
	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024 // 64 MiB
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

// Developer represents a registered developer account.
type Developer struct {
	ID            string
	Email         string
	Name          string
	EmailVerified bool
	CreatedAt     time.Time
}

// OAuthApp represents a developer's registered OAuth application.
type OAuthApp struct {
	ID              string
	DeveloperID     string
	Name            string
	Description     string
	HomepageURL     string
	ClientID        string
	RedirectURIs    []string
	IsPublic        bool
	CreatedAt       time.Time
	SecretRotatedAt *time.Time
}

// DevStore defines the storage interface for developer accounts and OAuth apps.
type DevStore interface {
	RegisterDeveloper(ctx context.Context, email, name, password string) (*Developer, error)
	MarkEmailVerified(ctx context.Context, email string) error
	AuthenticateDeveloper(ctx context.Context, email, password string) (*Developer, error)
	GetDeveloperByEmail(ctx context.Context, email string) (*Developer, error)
	CreateApp(ctx context.Context, developerID, name, description, homepageURL string, redirectURIs []string, isPublic bool) (*OAuthApp, string, error)
	RotateSecret(ctx context.Context, developerID, appID string) (*OAuthApp, string, error)
	ListApps(ctx context.Context, developerID string) ([]*OAuthApp, error)
	GetApp(ctx context.Context, developerID, appID string) (*OAuthApp, error)
	DeleteApp(ctx context.Context, developerID, appID string) error
}

// MemoryDevStore provides an in-memory implementation of DevStore.
type MemoryDevStore struct {
	mu   sync.RWMutex
	devs map[string]*Developer // email -> Developer
	pwd  map[string]string     // email -> argon2 hash
	apps map[string]*OAuthApp  // appID -> OAuthApp
}

// NewMemoryDevStore creates an in-memory developer store.
func NewMemoryDevStore() *MemoryDevStore {
	return &MemoryDevStore{
		devs: make(map[string]*Developer),
		pwd:  make(map[string]string),
		apps: make(map[string]*OAuthApp),
	}
}

func (m *MemoryDevStore) RegisterDeveloper(ctx context.Context, email, name, password string) (*Developer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || name == "" || password == "" {
		return nil, errors.New("email, name, and password are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.devs[email]; exists {
		return nil, errors.New("an account with this email already exists")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	dev := &Developer{
		ID:        mustRandomHex(16),
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}
	m.devs[email] = dev
	m.pwd[email] = hash
	return dev, nil
}

func (m *MemoryDevStore) MarkEmailVerified(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	m.mu.Lock()
	defer m.mu.Unlock()
	if dev, ok := m.devs[email]; ok {
		dev.EmailVerified = true
	}
	return nil
}

func (m *MemoryDevStore) AuthenticateDeveloper(ctx context.Context, email, password string) (*Developer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	m.mu.RLock()
	dev, ok := m.devs[email]
	hash := m.pwd[email]
	m.mu.RUnlock()

	if !ok {
		// Dummy verify keeps timing consistent regardless of whether the
		// account exists, matching the Postgres-backed Store's behavior —
		// otherwise a missing-account response returns measurably faster
		// than a wrong-password response, letting callers enumerate emails.
		_ = verifyPassword(password, dummyPasswordHash)
		return nil, errors.New("invalid email or password")
	}
	if !verifyPassword(password, hash) {
		return nil, errors.New("invalid email or password")
	}
	return dev, nil
}

func (m *MemoryDevStore) GetDeveloperByEmail(ctx context.Context, email string) (*Developer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	m.mu.RLock()
	defer m.mu.RUnlock()
	dev, ok := m.devs[email]
	if !ok {
		return nil, errors.New("developer not found")
	}
	return dev, nil
}

func (m *MemoryDevStore) CreateApp(ctx context.Context, developerID, name, description, homepageURL string, redirectURIs []string, isPublic bool) (*OAuthApp, string, error) {
	if name == "" || len(redirectURIs) == 0 {
		return nil, "", errors.New("app name and redirect URIs are required")
	}
	id := mustRandomHex(16)
	clientID := "da_" + mustRandomHex(16)
	var secret, secretHash string
	if !isPublic {
		secret = "dacs_" + mustRandomHex(32)
		h, _ := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
		secretHash = string(h)
	}

	app := &OAuthApp{
		ID:           id,
		DeveloperID:  developerID,
		Name:         name,
		Description:  description,
		HomepageURL:  homepageURL,
		ClientID:     clientID,
		RedirectURIs: redirectURIs,
		IsPublic:     isPublic,
		CreatedAt:    time.Now(),
	}
	m.mu.Lock()
	m.apps[id] = app
	m.mu.Unlock()
	_ = secretHash
	return app, secret, nil
}

func (m *MemoryDevStore) RotateSecret(ctx context.Context, developerID, appID string) (*OAuthApp, string, error) {
	app, err := m.GetApp(ctx, developerID, appID)
	if err != nil {
		return nil, "", err
	}
	secret := "dacs_" + mustRandomHex(32)
	now := time.Now()
	app.SecretRotatedAt = &now
	return app, secret, nil
}

func (m *MemoryDevStore) ListApps(ctx context.Context, developerID string) ([]*OAuthApp, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []*OAuthApp
	for _, app := range m.apps {
		if app.DeveloperID == developerID {
			res = append(res, app)
		}
	}
	return res, nil
}

func (m *MemoryDevStore) GetApp(ctx context.Context, developerID, appID string) (*OAuthApp, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	app, ok := m.apps[appID]
	if !ok || app.DeveloperID != developerID {
		return nil, ErrAppNotFound
	}
	return app, nil
}

func (m *MemoryDevStore) DeleteApp(ctx context.Context, developerID, appID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	app, ok := m.apps[appID]
	if !ok || app.DeveloperID != developerID {
		return ErrAppNotFound
	}
	delete(m.apps, appID)
	return nil
}

// Store manages developer accounts and OAuth app registrations in PostgreSQL.
type Store struct {
	db *sql.DB
}

// NewStore opens the database and runs schema migrations.
func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("developer: open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("developer: ping db: %w", err)
	}
	configureDBPool(db)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("developer: migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS developer_accounts (
			id             TEXT PRIMARY KEY,
			email          TEXT UNIQUE NOT NULL,
			password_hash  TEXT NOT NULL,
			name           TEXT NOT NULL,
			email_verified BOOLEAN NOT NULL DEFAULT FALSE,
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		ALTER TABLE developer_accounts ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

		CREATE TABLE IF NOT EXISTS oauth_apps (
			id                 TEXT PRIMARY KEY,
			developer_id       TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
			name               TEXT NOT NULL,
			description        TEXT NOT NULL DEFAULT '',
			homepage_url       TEXT NOT NULL DEFAULT '',
			client_id          TEXT UNIQUE NOT NULL,
			client_secret_hash TEXT NOT NULL DEFAULT '',
			redirect_uris      TEXT[] NOT NULL,
			is_public          BOOLEAN NOT NULL DEFAULT FALSE,
			created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			secret_rotated_at  TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_oauth_apps_developer_id ON oauth_apps(developer_id);
		CREATE INDEX IF NOT EXISTS idx_oauth_apps_client_id    ON oauth_apps(client_id);
	`)
	return err
}

// RegisterDeveloper creates a developer account with an Argon2id-hashed password.
func (s *Store) RegisterDeveloper(ctx context.Context, email, name, password string) (*Developer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || name == "" || password == "" {
		return nil, errors.New("email, name, and password are required")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("developer: hash password: %w", err)
	}
	id := mustRandomHex(16)
	dev := &Developer{ID: id, Email: email, Name: name, EmailVerified: false}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO developer_accounts (id, email, password_hash, name, email_verified) VALUES ($1, $2, $3, $4, FALSE)`,
		id, email, hash, name,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("an account with this email already exists")
		}
		return nil, fmt.Errorf("developer: register: %w", err)
	}
	return dev, nil
}

// MarkEmailVerified updates email_verified = TRUE for a developer account.
func (s *Store) MarkEmailVerified(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	_, err := s.db.ExecContext(ctx, `UPDATE developer_accounts SET email_verified = TRUE WHERE email = $1`, email)
	return err
}

// AuthenticateDeveloper verifies email and password, returning the Developer on success.
func (s *Store) AuthenticateDeveloper(ctx context.Context, email, password string) (*Developer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var dev Developer
	var hash string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, created_at FROM developer_accounts WHERE email = $1`,
		email,
	).Scan(&dev.ID, &dev.Email, &dev.Name, &hash, &dev.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Dummy verify keeps timing consistent regardless of whether account exists.
		_ = verifyPassword(password, dummyPasswordHash)
		return nil, errors.New("invalid email or password")
	}
	if err != nil {
		return nil, fmt.Errorf("developer: authenticate: %w", err)
	}
	if !verifyPassword(password, hash) {
		return nil, errors.New("invalid email or password")
	}
	return &dev, nil
}

// GetDeveloperByEmail looks up a developer account by email without
// authenticating a password. Used when a flow (e.g. email verification)
// needs the developer's real persisted ID rather than fabricating one.
func (s *Store) GetDeveloperByEmail(ctx context.Context, email string) (*Developer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var dev Developer
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, email_verified, created_at FROM developer_accounts WHERE email = $1`,
		email,
	).Scan(&dev.ID, &dev.Email, &dev.Name, &dev.EmailVerified, &dev.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("developer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("developer: get by email: %w", err)
	}
	return &dev, nil
}

// CreateApp registers a new OAuth app for the developer.
// Returns the app and the plaintext client_secret (shown ONCE — never stored).
// Pass isPublic=true for PKCE-only apps that don't need a secret.
func (s *Store) CreateApp(ctx context.Context, developerID, name, description, homepageURL string, redirectURIs []string, isPublic bool) (*OAuthApp, string, error) {
	if name == "" {
		return nil, "", errors.New("app name is required")
	}
	if len(redirectURIs) == 0 {
		return nil, "", errors.New("at least one redirect URI is required")
	}
	for _, uri := range redirectURIs {
		if err := validateRedirectURI(uri); err != nil {
			return nil, "", err
		}
	}

	id := mustRandomHex(16)
	clientID := "da_" + mustRandomHex(16) // "da_" prefix = DilithiAuth

	var plaintextSecret, secretHash string
	if !isPublic {
		plaintextSecret = "dacs_" + mustRandomHex(32) // "dacs_" = DilithiAuth Client Secret
		hash, err := bcrypt.GenerateFromPassword([]byte(plaintextSecret), bcrypt.DefaultCost)
		if err != nil {
			return nil, "", fmt.Errorf("developer: hash client secret: %w", err)
		}
		secretHash = string(hash)
	}

	app := &OAuthApp{
		ID:           id,
		DeveloperID:  developerID,
		Name:         name,
		Description:  description,
		HomepageURL:  homepageURL,
		ClientID:     clientID,
		RedirectURIs: redirectURIs,
		IsPublic:     isPublic,
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO oauth_apps
		 (id, developer_id, name, description, homepage_url, client_id, client_secret_hash, redirect_uris, is_public)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, developerID, name, description, homepageURL,
		clientID, secretHash, pq.Array(redirectURIs), isPublic,
	)
	if err != nil {
		return nil, "", fmt.Errorf("developer: create app: %w", err)
	}
	return app, plaintextSecret, nil
}

// RotateSecret generates a new client_secret for a confidential app.
// Returns the app and the new plaintext secret (shown ONCE).
func (s *Store) RotateSecret(ctx context.Context, developerID, appID string) (*OAuthApp, string, error) {
	app, err := s.GetApp(ctx, developerID, appID)
	if err != nil {
		return nil, "", err
	}
	if app.IsPublic {
		return nil, "", errors.New("public (PKCE-only) apps do not have client secrets")
	}

	plaintext := "dacs_" + mustRandomHex(32)
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("developer: hash rotated secret: %w", err)
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`UPDATE oauth_apps SET client_secret_hash = $1, secret_rotated_at = $2
		 WHERE id = $3 AND developer_id = $4`,
		string(hash), now, appID, developerID,
	)
	if err != nil {
		return nil, "", fmt.Errorf("developer: rotate secret: %w", err)
	}
	app.SecretRotatedAt = &now
	return app, plaintext, nil
}

// ListApps returns all apps owned by the given developer, ordered by creation date.
func (s *Store) ListApps(ctx context.Context, developerID string) ([]*OAuthApp, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, homepage_url, client_id, redirect_uris, is_public, created_at, secret_rotated_at
		 FROM oauth_apps WHERE developer_id = $1 ORDER BY created_at DESC`,
		developerID,
	)
	if err != nil {
		return nil, fmt.Errorf("developer: list apps: %w", err)
	}
	defer rows.Close()

	var apps []*OAuthApp
	for rows.Next() {
		app := &OAuthApp{DeveloperID: developerID}
		if err := rows.Scan(&app.ID, &app.Name, &app.Description, &app.HomepageURL,
			&app.ClientID, pq.Array(&app.RedirectURIs), &app.IsPublic, &app.CreatedAt, &app.SecretRotatedAt); err != nil {
			return nil, fmt.Errorf("developer: scan app: %w", err)
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

// GetApp returns a single app, verifying it belongs to developerID.
func (s *Store) GetApp(ctx context.Context, developerID, appID string) (*OAuthApp, error) {
	app := &OAuthApp{ID: appID, DeveloperID: developerID}
	err := s.db.QueryRowContext(ctx,
		`SELECT name, description, homepage_url, client_id, redirect_uris, is_public, created_at, secret_rotated_at
		 FROM oauth_apps WHERE id = $1 AND developer_id = $2`,
		appID, developerID,
	).Scan(&app.Name, &app.Description, &app.HomepageURL, &app.ClientID,
		pq.Array(&app.RedirectURIs), &app.IsPublic, &app.CreatedAt, &app.SecretRotatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAppNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("developer: get app: %w", err)
	}
	return app, nil
}

// DeleteApp removes an app owned by the given developer.
func (s *Store) DeleteApp(ctx context.Context, developerID, appID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM oauth_apps WHERE id = $1 AND developer_id = $2`, appID, developerID,
	)
	if err != nil {
		return fmt.Errorf("developer: delete app: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrAppNotFound
	}
	return nil
}

func (s *Store) Close() error { return s.db.Close() }

// Ping verifies the database connection is reachable, for use by readiness
// checks (e.g. an HTTP /readyz handler). It respects ctx's deadline/cancellation.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// configureDBPool applies production-sane connection pool limits to db,
// overriding Go's defaults (unbounded max-open, no idle timeout) which can
// otherwise let a single process exhaust the database's connection limit.
// DILITHIAUTH_DB_MAX_CONNS overrides the max-open-connections default; all
// other limits are fixed, sensible values for a small-VM deployment.
func configureDBPool(db *sql.DB) {
	maxOpen := 25
	if v := os.Getenv("DILITHIAUTH_DB_MAX_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpen = n
		}
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)
}

// --- helpers ---

func mustRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("developer: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// dummyPasswordHash is a syntactically valid (but not password-derived)
// salt:hash pair run through verifyPassword when an account lookup misses,
// so a nonexistent-account response takes the same time as a wrong-password
// response for an existing account.
const dummyPasswordHash = "0000000000000000000000000000000000000000:0000000000000000000000000000000000000000000000000000000000000000"

func hashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash), nil
}

func verifyPassword(password, encoded string) bool {
	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	// crypto/subtle is not imported here; use argon2's own output length for comparison.
	if len(hash) != len(expected) {
		return false
	}
	var diff byte
	for i := range hash {
		diff |= hash[i] ^ expected[i]
	}
	return diff == 0
}

func validateRedirectURI(uri string) error {
	if uri == "" {
		return errors.New("redirect URI cannot be empty")
	}
	if strings.HasPrefix(uri, "http://localhost") || strings.HasPrefix(uri, "http://127.0.0.1") {
		return nil // localhost HTTP is allowed for development
	}
	if !strings.HasPrefix(uri, "https://") {
		return fmt.Errorf("redirect URI %q must use HTTPS (http://localhost is allowed for development)", uri)
	}
	return nil
}
