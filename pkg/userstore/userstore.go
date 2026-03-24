// Package userstore provides a production-grade user store for DilithiAuth.
//
// Passwords are hashed with Argon2id (OWASP recommended parameters).
// The PostgresUserStore is the primary implementation; it auto-creates the
// required schema on first use.
package userstore

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/argon2"
)

// Argon2id parameters — OWASP minimum recommendations (2023).
const (
	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024 // 64 MiB
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

// User represents an end-user identity within a tenant.
type User struct {
	ID           string
	TenantID     string
	Username     string
	Email        string
	PasswordHash string
	Subject      string // stable identifier used in JWT sub claim
	CreatedAt    time.Time
}

// UserStore is the interface for managing user identities.
type UserStore interface {
	// Register creates a new user and returns the persisted record.
	Register(ctx context.Context, tenantID, username, email, password string) (*User, error)
	// Authenticate verifies credentials and returns the user on success.
	Authenticate(ctx context.Context, tenantID, username, password string) (*User, error)
	// GetBySubject looks up a user by their JWT subject claim.
	GetBySubject(ctx context.Context, tenantID, subject string) (*User, error)
}

// PostgresUserStore implements UserStore backed by PostgreSQL.
type PostgresUserStore struct {
	db *sql.DB
}

// NewPostgresUserStore opens a connection, pings the database, and
// auto-migrates the users table.
func NewPostgresUserStore(dsn string) (*PostgresUserStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("userstore: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("userstore: ping: %w", err)
	}
	s := &PostgresUserStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("userstore: migrate: %w", err)
	}
	return s, nil
}

func (s *PostgresUserStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id           TEXT PRIMARY KEY,
			tenant_id    TEXT NOT NULL,
			username     TEXT NOT NULL,
			email        TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			subject      TEXT UNIQUE NOT NULL,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (tenant_id, username),
			UNIQUE (tenant_id, email)
		)
	`)
	return err
}

// Register creates a new user with an Argon2id-hashed password.
// Returns an error if the username or email already exists in the tenant.
func (s *PostgresUserStore) Register(ctx context.Context, tenantID, username, email, password string) (*User, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("userstore: hashing password: %w", err)
	}

	id, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	subjectSuffix, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	subject := "usr_" + subjectSuffix
	now := time.Now().UTC()

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users (id, tenant_id, username, email, password_hash, subject, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, tenantID, username, email, hash, subject, now,
	)
	if err != nil {
		return nil, fmt.Errorf("userstore: register: %w", err)
	}
	return &User{
		ID: id, TenantID: tenantID, Username: username,
		Email: email, PasswordHash: hash, Subject: subject, CreatedAt: now,
	}, nil
}

// Authenticate verifies username and password for the given tenant.
// Returns a generic "invalid credentials" error for both unknown users and
// wrong passwords to prevent username enumeration.
func (s *PostgresUserStore) Authenticate(ctx context.Context, tenantID, username, password string) (*User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, username, email, password_hash, subject, created_at
		 FROM users WHERE tenant_id = $1 AND username = $2`,
		tenantID, username,
	).Scan(&u.ID, &u.TenantID, &u.Username, &u.Email, &u.PasswordHash, &u.Subject, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Run a dummy verify to keep timing consistent.
		_ = verifyPassword(password, "0000000000000000000000000000000000000000:0000000000000000000000000000000000000000000000000000000000000000")
		return nil, errors.New("invalid credentials")
	}
	if err != nil {
		return nil, fmt.Errorf("userstore: authenticate query: %w", err)
	}
	if !verifyPassword(password, u.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}
	return &u, nil
}

// GetBySubject looks up a user by their stable JWT sub claim.
func (s *PostgresUserStore) GetBySubject(ctx context.Context, tenantID, subject string) (*User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, username, email, password_hash, subject, created_at
		 FROM users WHERE tenant_id = $1 AND subject = $2`,
		tenantID, subject,
	).Scan(&u.ID, &u.TenantID, &u.Username, &u.Email, &u.PasswordHash, &u.Subject, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("userstore: get by subject: %w", err)
	}
	return &u, nil
}

// AuthenticateForSubject is a convenience adapter that satisfies the
// oauth.UserAuthenticator interface, returning only the subject string.
func (s *PostgresUserStore) AuthenticateForSubject(ctx context.Context, tenantID, username, password string) (string, error) {
	u, err := s.Authenticate(ctx, tenantID, username, password)
	if err != nil {
		return "", err
	}
	return u.Subject, nil
}

// Close releases the underlying database connection.
func (s *PostgresUserStore) Close() error { return s.db.Close() }

// ---------------------------------------------------------------------------
// Argon2id helpers
// ---------------------------------------------------------------------------

// hashPassword returns a string in the form "<salt_hex>:<hash_hex>".
func hashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("userstore: generating salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash), nil
}

// verifyPassword returns true if password matches the encoded hash.
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
	return subtle.ConstantTimeCompare(hash, expected) == 1
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("userstore: random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
