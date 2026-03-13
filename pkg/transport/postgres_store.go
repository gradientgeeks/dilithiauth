package transport

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgresSigStore is a PostgreSQL-backed implementation of SigStoreBackend.
// It stores ML-DSA-44 signatures in a sig_store table for durable, shared storage
// across multiple server instances.
type PostgresSigStore struct {
	db *sql.DB
}

// NewPostgresSigStore creates a PostgresSigStore connected to the given DSN.
// It verifies connectivity and auto-creates the sig_store table if it does not exist.
func NewPostgresSigStore(dsn string) (*PostgresSigStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres_store: opening connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres_store: ping failed: %w", err)
	}

	createTable := `CREATE TABLE IF NOT EXISTS sig_store (
		jti TEXT PRIMARY KEY,
		sig BYTEA NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	)`
	if _, err := db.Exec(createTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres_store: creating table: %w", err)
	}

	return &PostgresSigStore{db: db}, nil
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
