package oauth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// PostgresAuthzStore implements ClientValidator against the developer portal's
// oauth_apps table. It handles client authentication at /authorize and /token.
// Authorization codes are NOT stored here — those use the configured AuthzStore
// (memory or Redis). This store is the production source of truth for registered apps.
type PostgresAuthzStore struct {
	db *sql.DB
}

// NewPostgresAuthzStore opens a connection to the oauth_apps table.
// The schema is created by pkg/developer.Store.migrate(); this store is read-only.
func NewPostgresAuthzStore(dsn string) (*PostgresAuthzStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres_authz_store: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres_authz_store: ping: %w", err)
	}
	configureDBPool(db)
	return &PostgresAuthzStore{db: db}, nil
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

// ValidateClient checks that client_id exists and redirectURI is in the app's
// registered redirect_uris array. Used by AuthorizeGET before showing the login form.
func (s *PostgresAuthzStore) ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error) {
	var c OAuthClient
	err := s.db.QueryRowContext(context.Background(),
		`SELECT client_id, client_secret_hash, name, developer_id, is_public
		 FROM oauth_apps
		 WHERE client_id = $1 AND $2 = ANY(redirect_uris)`,
		clientID, redirectURI,
	).Scan(&c.ClientID, &c.ClientSecretHash, &c.Name, &c.DeveloperID, new(bool))

	if errors.Is(err, sql.ErrNoRows) {
		// Avoid revealing whether the client_id exists at all.
		return nil, errors.New("unknown client or redirect_uri mismatch")
	}
	if err != nil {
		return nil, fmt.Errorf("postgres_authz_store: validate client: %w", err)
	}
	c.RedirectURI = redirectURI // the specific URI that matched
	return &c, nil
}

// ValidateClientWithSecret authenticates a client at the token endpoint.
// For public clients (ClientSecretHash == ""), clientSecret must be empty.
func (s *PostgresAuthzStore) ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI string) (*OAuthClient, error) {
	client, err := s.ValidateClient(tenantID, clientID, redirectURI)
	if err != nil {
		// Dummy compare to prevent timing oracle even when client not found.
		_ = bcrypt.CompareHashAndPassword(bcryptDummyHash, []byte(clientSecret))
		return nil, err
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

// ValidateClientSecretOnly authenticates a client by client_id+client_secret
// alone, with no redirect_uri check (used by the refresh_token grant).
func (s *PostgresAuthzStore) ValidateClientSecretOnly(tenantID, clientID, clientSecret string) (*OAuthClient, error) {
	var c OAuthClient
	err := s.db.QueryRowContext(context.Background(),
		`SELECT client_id, client_secret_hash, name, developer_id FROM oauth_apps WHERE client_id = $1`,
		clientID,
	).Scan(&c.ClientID, &c.ClientSecretHash, &c.Name, &c.DeveloperID)

	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(bcryptDummyHash, []byte(clientSecret))
		return nil, errors.New("unknown client")
	}
	if err != nil {
		return nil, fmt.Errorf("postgres_authz_store: validate client secret: %w", err)
	}
	if c.ClientSecretHash == "" {
		if clientSecret != "" {
			return nil, errors.New("public client does not accept client_secret")
		}
		return &c, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(c.ClientSecretHash), []byte(clientSecret)); err != nil {
		return nil, errors.New("invalid client_secret")
	}
	return &c, nil
}

func (s *PostgresAuthzStore) Close() error { return s.db.Close() }

// Ping verifies the database connection is reachable, for use by readiness
// checks (e.g. an HTTP /readyz handler). It respects ctx's deadline/cancellation.
func (s *PostgresAuthzStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
