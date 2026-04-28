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
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

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
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
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
			created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

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
	dev := &Developer{ID: id, Email: email, Name: name}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO developer_accounts (id, email, password_hash, name) VALUES ($1, $2, $3, $4)`,
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
		_ = verifyPassword(password, "0000000000000000000000000000000000000000:0000000000000000000000000000000000000000000000000000000000000000")
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
		ID:          id,
		DeveloperID: developerID,
		Name:        name,
		Description: description,
		HomepageURL: homepageURL,
		ClientID:    clientID,
		RedirectURIs: redirectURIs,
		IsPublic:    isPublic,
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO oauth_apps
		 (id, developer_id, name, description, homepage_url, client_id, client_secret_hash, redirect_uris, is_public)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, developerID, name, description, homepageURL,
		clientID, secretHash, redirectURIs, isPublic,
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
			&app.ClientID, &app.RedirectURIs, &app.IsPublic, &app.CreatedAt, &app.SecretRotatedAt); err != nil {
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
		&app.RedirectURIs, &app.IsPublic, &app.CreatedAt, &app.SecretRotatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("app not found")
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
		return errors.New("app not found")
	}
	return nil
}

func (s *Store) Close() error { return s.db.Close() }

// --- helpers ---

func mustRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("developer: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

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
