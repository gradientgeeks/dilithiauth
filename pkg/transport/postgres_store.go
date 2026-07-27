package transport

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// PostgresSigStore is a PostgreSQL-backed implementation of SigStoreBackend.
// It stores ML-DSA-44 signatures in a sig_store table for durable, shared storage
// across multiple server instances.
type PostgresSigStore struct {
	db  *sql.DB
	ttl time.Duration // 0 = Purge() is a no-op; entries are never auto-expired
}

// NewPostgresSigStore creates a PostgresSigStore connected to the given DSN.
// It verifies connectivity and auto-creates the sig_store table if it does
// not exist. Purge() is a no-op until a TTL is set via
// NewPostgresSigStoreWithTTL.
func NewPostgresSigStore(dsn string) (*PostgresSigStore, error) {
	return NewPostgresSigStoreWithTTL(dsn, 0)
}

// NewPostgresSigStoreWithTTL is like NewPostgresSigStore but Purge() deletes
// rows older than ttl (based on their created_at timestamp).
func NewPostgresSigStoreWithTTL(dsn string, ttl time.Duration) (*PostgresSigStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres_store: opening connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres_store: ping failed: %w", err)
	}
	configureDBPool(db)

	createTable := `CREATE TABLE IF NOT EXISTS sig_store (
		jti TEXT PRIMARY KEY,
		sig BYTEA NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	)`
	if _, err := db.Exec(createTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres_store: creating table: %w", err)
	}

	return &PostgresSigStore{db: db, ttl: ttl}, nil
}

func (p *PostgresSigStore) Put(jti string, sig []byte) {
	_, _ = p.db.Exec(
		`INSERT INTO sig_store (jti, sig) VALUES ($1, $2)
		 ON CONFLICT (jti) DO UPDATE SET sig = $2`,
		jti, sig,
	)
}

func (p *PostgresSigStore) Get(jti string) ([]byte, bool) {
	var sig []byte
	err := p.db.QueryRow(`SELECT sig FROM sig_store WHERE jti = $1`, jti).Scan(&sig)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return nil, false
	}
	return sig, true
}

func (p *PostgresSigStore) Delete(jti string) {
	_, _ = p.db.Exec(`DELETE FROM sig_store WHERE jti = $1`, jti)
}

func (p *PostgresSigStore) Len() int {
	var count int
	if err := p.db.QueryRow(`SELECT COUNT(*) FROM sig_store`).Scan(&count); err != nil {
		return 0
	}
	return count
}

// Close closes the underlying database connection.
func (p *PostgresSigStore) Close() error {
	return p.db.Close()
}

// Ping verifies the database connection is reachable, for use by readiness
// checks (e.g. an HTTP /readyz handler). It respects ctx's deadline/cancellation.
func (p *PostgresSigStore) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
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

// Purge deletes rows older than the configured ttl, bounding table growth.
// A no-op if ttl is 0 (NewPostgresSigStore without NewPostgresSigStoreWithTTL).
// Call periodically (e.g. hourly) from a background goroutine.
func (p *PostgresSigStore) Purge() {
	if p.ttl <= 0 {
		return
	}
	_, _ = p.db.Exec(`DELETE FROM sig_store WHERE created_at < $1`, time.Now().Add(-p.ttl))
}
