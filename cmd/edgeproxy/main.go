package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/edgeproxy"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

func main() {
	// Configuration
	addr := getEnv("EDGE_PROXY_ADDR", ":8081")
	upstreamURL := getEnv("UPSTREAM_URL", "http://localhost:8082") // Dummy microservice
	internalSecret := []byte(getEnv("INTERNAL_JWT_SECRET", "super-secret-key-for-internal-services"))

	log.Printf("Starting Edge Proxy on %s (routing to %s)", addr, upstreamURL)

	// Setup Verifier to validate DilithiAuth PQC tokens.
	// In a real environment, this proxy would load the public keys from
	// the DilithiAuth JWKS endpoint or share the Redis KeyStore.
	// For this demo, we'll assume it shares the same Redis KeyStore if configured.
	var km *signer.KeyManager
	keyStoreType := getEnv("DILITHIAUTH_KEYSTORE", "file")
	keysDir := getEnv("DILITHIAUTH_KEYS_DIR", "./data")

	var err error
	if keyStoreType == "redis" {
		redisAddr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		rs, err := signer.NewRedisKeyStore(redisAddr)
		if err != nil {
			log.Fatalf("FATAL: creating Redis KeyStore: %v", err)
		}
		km, err = signer.NewKeyManagerWithStore(rs)
		if err != nil {
			log.Fatalf("FATAL: creating KeyManager: %v", err)
		}
		log.Printf("Loaded public keys from Redis (%s)", redisAddr)
	} else {
		// Fallback to local files
		km, err = signer.NewKeyManager(keysDir)
		if err != nil {
			log.Fatalf("FATAL: loading local KeyManager: %v", err)
		}
		log.Printf("Loaded public keys from %s", keysDir)
	}

	// For Jumbo tokens, we don't even need a SigStore because the signature
	// is fully embedded in the token! We pass an empty one.
	emptySigStore := transport.NewSigStore()
	v := verifier.NewVerifier(km, emptySigStore)

	// Initialize the Edge Proxy
	proxy, err := edgeproxy.NewEdgeProxy(v, internalSecret, upstreamURL)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	// We MUST use the PQC-tuned server here because the edge proxy
	// is the component actually receiving the massive 16KB header from the client!
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(transport.JumboHeaderMiddleware)

	// Mount the proxy
	r.Mount("/", proxy.Handler())

	srv := transport.NewPQCServer(addr, r)
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
