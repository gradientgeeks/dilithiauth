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
	"os"
	"strconv"
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
	// GetByEmail looks up a user by email, used by the password-reset flow.
	GetByEmail(ctx context.Context, tenantID, email string) (*User, error)
	// SetPassword updates a user's password hash, used by the password-reset flow.
	SetPassword(ctx context.Context, tenantID, username, newPasswordHash string) error
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
	configureDBPool(db)
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

// RegisterForSubject is a convenience adapter that satisfies the
// oauth.UserRegistrar interface, returning only the subject string.
func (s *PostgresUserStore) RegisterForSubject(ctx context.Context, tenantID, username, email, password string) (string, error) {
	u, err := s.Register(ctx, tenantID, username, email, password)
	if err != nil {
		return "", err
	}
	return u.Subject, nil
}

// GetByEmail looks up a user by email within a tenant.
func (s *PostgresUserStore) GetByEmail(ctx context.Context, tenantID, email string) (*User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, username, email, password_hash, subject, created_at
		 FROM users WHERE tenant_id = $1 AND email = $2`,
		tenantID, email,
	).Scan(&u.ID, &u.TenantID, &u.Username, &u.Email, &u.PasswordHash, &u.Subject, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("userstore: get by email: %w", err)
	}
	return &u, nil
}

// SetPassword updates a user's password hash (Argon2id) within a tenant.
func (s *PostgresUserStore) SetPassword(ctx context.Context, tenantID, username, newPasswordHash string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE tenant_id = $2 AND username = $3`,
		newPasswordHash, tenantID, username,
	)
	if err != nil {
		return fmt.Errorf("userstore: set password: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("user not found")
	}
	return nil
}

// HashPassword exposes the package's Argon2id hashing for callers (e.g. the
// password-reset flow) that need to hash a new password before calling
// SetPassword.
func HashPassword(password string) (string, error) {
	return hashPassword(password)
}

// EmailExists reports whether an account with the given email exists,
// satisfying oauth.PasswordResetter.
func (s *PostgresUserStore) EmailExists(ctx context.Context, tenantID, email string) (bool, error) {
	_, err := s.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// SetPasswordForEmail hashes and sets a new password for the account with
// the given email, satisfying oauth.PasswordResetter.
func (s *PostgresUserStore) SetPasswordForEmail(ctx context.Context, tenantID, email, newPassword string) error {
	u, err := s.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return err
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.SetPassword(ctx, tenantID, u.Username, hash)
}

// Close releases the underlying database connection.
func (s *PostgresUserStore) Close() error { return s.db.Close() }

// Ping verifies the database connection is reachable, for use by readiness
// checks (e.g. an HTTP /readyz handler). It respects ctx's deadline/cancellation.
func (s *PostgresUserStore) Ping(ctx context.Context) error {
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
