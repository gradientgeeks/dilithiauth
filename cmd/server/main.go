// Command server is the DilithiAuth Identity Provider entrypoint.
//
// It starts a quantum-safe OAuth 2.0 / OIDC server with:
//   - Hybrid signing: ML-DSA-44 (NIST FIPS 204) + Ed25519
//   - OIDC Discovery endpoint (/.well-known/openid-configuration)
//   - JWKS endpoint (/.well-known/jwks.json)
//   - JWT issuance in both "thin" (sidecar) and "full" (jumbo) modes
//   - CBOR Web Token (CWT, RFC 8392) issuance
//   - MaxHeaderBytes tuned to 16 KiB for PQC-sized tokens
//
// Configuration is read from environment variables:
//
//	DILITHIAUTH_ISSUER  – base URL of the issuer (default: https://auth.dilithiauth.io)
//	DILITHIAUTH_ADDR    – listen address (default: :8080)
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
)

func main() {
	issuer := getEnv("DILITHIAUTH_ISSUER", "https://auth.dilithiauth.io")
	addr := getEnv("DILITHIAUTH_ADDR", ":8080")

	log.Printf("DilithiAuth starting – issuer=%s addr=%s", issuer, addr)

	// Generate hybrid key pair (ML-DSA-44 + Ed25519).
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		log.Fatalf("FATAL: generating hybrid key pair: %v", err)
	}
	log.Printf("Hybrid key pair generated – kid=%s", kp.Kid)

	// Build the OAuth provider.
	provider := oauth.NewProvider(issuer, kp)

	// Signature sidecar store (in-memory; replace with Redis in production).
	sigStore := transport.NewSigStore()

	// Build the router.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(transport.JumboHeaderMiddleware)
	r.Use(transport.SigStoreMiddleware(sigStore))

	// OIDC Discovery & JWKS.
	r.Get("/.well-known/openid-configuration", provider.DiscoveryHandler())
	r.Get("/.well-known/jwks.json", provider.JWKSHandler())

	// Token issuance endpoints (demo – no auth guard for brevity).
	r.Post("/token/thin", makeThinTokenHandler(provider, sigStore))
	r.Post("/token/full", makeFullTokenHandler(provider))
	r.Post("/token/cwt", makeCWTTokenHandler(provider, sigStore))

	// Health check.
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Create a PQC-ready server with MaxHeaderBytes = 16 KiB.
	srv := transport.NewPQCServer(addr, r)
	srv.ReadTimeout = 15 * time.Second
	srv.WriteTimeout = 15 * time.Second
	srv.IdleTimeout = 60 * time.Second

	log.Printf("Listening on %s (MaxHeaderBytes=%d)", addr, srv.MaxHeaderBytes)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: %v", err)
	}
}

// makeThinTokenHandler issues a "thin" JWT with a pqc_ref sidecar entry.
func makeThinTokenHandler(p *oauth.Provider, store *transport.SigStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("sub")
		if subject == "" {
			subject = "anonymous"
		}

		result, err := p.IssueJWTThin(subject, time.Hour)
		if err != nil {
			http.Error(w, "token issuance failed", http.StatusInternalServerError)
			return
		}

		// Store the ML-DSA-44 signature in the sidecar for later verification.
		store.Put(result.PQCRef, result.PQCSig)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": result.Token,
			"token_type":   "Bearer",
			"pqc_ref":      result.PQCRef,
		})
	}
}

// makeFullTokenHandler issues a "full" / jumbo JWT with embedded ML-DSA-44 sig.
func makeFullTokenHandler(p *oauth.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("sub")
		if subject == "" {
			subject = "anonymous"
		}

		tokenStr, err := p.IssueJWTFull(subject, time.Hour)
		if err != nil {
			http.Error(w, "token issuance failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": tokenStr,
			"token_type":   "Bearer",
		})
	}
}

// makeCWTTokenHandler issues a CBOR Web Token with sidecar PQC signature.
func makeCWTTokenHandler(p *oauth.Provider, store *transport.SigStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("sub")
		if subject == "" {
			subject = "anonymous"
		}

		result, err := p.IssueCWT(subject, time.Hour)
		if err != nil {
			http.Error(w, "CWT issuance failed", http.StatusInternalServerError)
			return
		}

		// Store PQC signature in sidecar keyed by JWTID.
		store.Put(result.JWTID, result.PQCSig)

		// Return the raw CBOR bytes with the appropriate content type.
		w.Header().Set("Content-Type", "application/cbor")
		w.Header().Set("X-Token-Id", result.JWTID)
		w.Header().Set("X-PQC-Ref", result.PQCRef)
		_, _ = w.Write(result.Token)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
