// Command management runs the DilithiAuth Management API server.
//
// The Management API lets operators provision tenants, register OAuth clients,
// and create users without touching config files or restarting the IdP.
//
// All endpoints require the Authorization header (Bearer <key>) or the
// X-API-Key header matching MANAGEMENT_API_KEY. If the variable is unset,
// auth is disabled (development mode only).
//
// Configuration:
//
//	MANAGEMENT_ADDR       – listen address (default: :9090)
//	MANAGEMENT_API_KEY    – required API key (default: empty = no auth)
//	MANAGEMENT_POSTGRES_DSN – PostgreSQL DSN for tenants + clients
//	MANAGEMENT_USERSTORE_DSN – PostgreSQL DSN for users (can be same as above)
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/management"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/userstore"
)

func main() {
	addr := getEnv("MANAGEMENT_ADDR", ":9090")
	apiKey := getEnv("MANAGEMENT_API_KEY", "")

	dsn := getEnv("MANAGEMENT_POSTGRES_DSN", "postgres://localhost:5432/dilithiauth?sslmode=disable")
	store, err := management.NewPostgresTenantStore(dsn)
	if err != nil {
		log.Fatalf("FATAL: connecting to PostgreSQL: %v", err)
	}
	defer store.Close()

	// UserStore is optional — the /users endpoint is disabled if not configured.
	var us *userstore.PostgresUserStore
	if userDSN := getEnv("MANAGEMENT_USERSTORE_DSN", ""); userDSN != "" {
		us, err = userstore.NewPostgresUserStore(userDSN)
		if err != nil {
			log.Fatalf("FATAL: connecting to user store: %v", err)
		}
		defer us.Close()
		log.Printf("UserStore: PostgreSQL-backed")
	} else {
		log.Printf("UserStore: not configured (POST /users will return 501)")
	}

	h := &management.Handler{
		Store:     store,
		UserStore: us,
		APIKey:    apiKey,
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      h.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if apiKey == "" {
		log.Printf("WARNING: MANAGEMENT_API_KEY is not set — API is open to all callers")
	}
	log.Printf("Management API listening on %s", addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
