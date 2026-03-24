package management

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Tenant represents an isolated customer namespace in DilithiAuth.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"` // used as subdomain: slug.auth.dilithiauth.com
	CreatedAt time.Time `json:"created_at"`
}

// ManagedClient is an OAuth 2.0 client registered via the Management API.
type ManagedClient struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	ClientID    string    `json:"client_id"`
	RedirectURI string    `json:"redirect_uri"`
	CreatedAt   time.Time `json:"created_at"`
}

// TenantStore persists tenant and OAuth client records.
type TenantStore interface {
	CreateTenant(ctx context.Context, name, slug string) (*Tenant, error)
	ListTenants(ctx context.Context) ([]*Tenant, error)
	GetTenant(ctx context.Context, id string) (*Tenant, error)

	AddClient(ctx context.Context, tenantID, name, clientID, redirectURI string) (*ManagedClient, error)
	ListClients(ctx context.Context, tenantID string) ([]*ManagedClient, error)
	DeleteClient(ctx context.Context, tenantID, clientID string) error
}

// PostgresTenantStore implements TenantStore backed by PostgreSQL.
type PostgresTenantStore struct {
	db *sql.DB
}

// NewPostgresTenantStore opens a connection, pings the database, and
// auto-migrates the tenants and oauth_clients tables.
func NewPostgresTenantStore(dsn string) (*PostgresTenantStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("management: open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("management: ping db: %w", err)
	}
	s := &PostgresTenantStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("management: migrate: %w", err)
	}
	return s, nil
}

func (s *PostgresTenantStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tenants (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			slug       TEXT UNIQUE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS oauth_clients (
			id           TEXT PRIMARY KEY,
			tenant_id    TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			name         TEXT NOT NULL,
			client_id    TEXT UNIQUE NOT NULL,
			redirect_uri TEXT NOT NULL,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func (s *PostgresTenantStore) CreateTenant(ctx context.Context, name, slug string) (*Tenant, error) {
	id, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO tenants (id, name, slug, created_at) VALUES ($1, $2, $3, $4)`,
		id, name, slug, now,
	)
	if err != nil {
		return nil, fmt.Errorf("management: create tenant: %w", err)
	}
	return &Tenant{ID: id, Name: name, Slug: slug, CreatedAt: now}, nil
}

func (s *PostgresTenantStore) ListTenants(ctx context.Context) ([]*Tenant, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, slug, created_at FROM tenants ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("management: list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		t := &Tenant{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func (s *PostgresTenantStore) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	t := &Tenant{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, slug, created_at FROM tenants WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("management: get tenant: %w", err)
	}
	return t, nil
}

func (s *PostgresTenantStore) AddClient(ctx context.Context, tenantID, name, clientID, redirectURI string) (*ManagedClient, error) {
	id, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO oauth_clients (id, tenant_id, name, client_id, redirect_uri, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, tenantID, name, clientID, redirectURI, now,
	)
	if err != nil {
		return nil, fmt.Errorf("management: add client: %w", err)
	}
	return &ManagedClient{
		ID: id, TenantID: tenantID, Name: name,
		ClientID: clientID, RedirectURI: redirectURI, CreatedAt: now,
	}, nil
}

func (s *PostgresTenantStore) ListClients(ctx context.Context, tenantID string) ([]*ManagedClient, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, client_id, redirect_uri, created_at
		 FROM oauth_clients WHERE tenant_id = $1 ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("management: list clients: %w", err)
	}
	defer rows.Close()

	var clients []*ManagedClient
	for rows.Next() {
		c := &ManagedClient{}
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.ClientID, &c.RedirectURI, &c.CreatedAt); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

func (s *PostgresTenantStore) DeleteClient(ctx context.Context, tenantID, clientID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM oauth_clients WHERE tenant_id = $1 AND client_id = $2`,
		tenantID, clientID,
	)
	if err != nil {
		return fmt.Errorf("management: delete client: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("client not found")
	}
	return nil
}

// Close releases the underlying database connection.
func (s *PostgresTenantStore) Close() error { return s.db.Close() }

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("management: random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
