package oauth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
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
	return &PostgresAuthzStore{db: db}, nil
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

func (s *PostgresAuthzStore) Close() error { return s.db.Close() }
