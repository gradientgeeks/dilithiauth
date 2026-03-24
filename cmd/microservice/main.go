package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	addr := getEnv("MICROSERVICE_ADDR", ":8082")
	internalSecret := []byte(getEnv("INTERNAL_JWT_SECRET", "super-secret-key-for-internal-services"))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
		// 1. Read the standard Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Validate using standard, lightweight HMAC (No PQC overhead here!)
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return internalSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 3. Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}

		subject := claims["sub"].(string)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello %s, you have successfully accessed the internal microservice using a standard lightweight JWT!"}`, subject)
	})

	log.Printf("Internal Microservice starting on %s (Standard headers only)", addr)

	// Notice we use standard http.Server, NOT the PQC-tuned one,
	// because the Edge Proxy stripped the massive headers.
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
