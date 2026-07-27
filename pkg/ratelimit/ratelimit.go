// Package ratelimit provides a simple in-memory per-key rate limiter and an
// HTTP middleware built on top of it, used to blunt brute-force and
// enumeration attacks against authentication-adjacent endpoints (login,
// registration, OTP verification/resend, token exchange, revocation).
package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Limiter is a token-bucket rate limiter keyed by an arbitrary string (e.g.
// client IP or normalized email). Safe for concurrent use.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens added per second
	burst    float64 // max tokens (and starting tokens) per bucket
	lastSeen map[string]time.Time
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewLimiter creates a Limiter allowing burst immediate requests per key,
// refilling at rate requests/sec afterward (i.e. an average of rate req/s
// sustained, with short bursts up to burst).
func NewLimiter(rate float64, burst int) *Limiter {
	return &Limiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   float64(burst),
	}
}

// Allow reports whether a request for key is permitted right now, consuming
// one token if so.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, lastRefill: now}
		l.buckets[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Purge removes buckets that have been full (i.e. idle) for longer than
// olderThan, bounding memory growth for a long-running process.
func (l *Limiter) Purge(olderThan time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := time.Now().Add(-olderThan)
	for key, b := range l.buckets {
		if b.tokens >= l.burst && b.lastRefill.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
}

// KeyFunc extracts a rate-limit key from a request, e.g. client IP or a
// normalized identifier parsed from the request body.
type KeyFunc func(r *http.Request) string

// ClientIPKey returns a KeyFunc that keys by remote IP (ignoring port).
func ClientIPKey() KeyFunc {
	return func(r *http.Request) string {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr
		}
		return host
	}
}

// Middleware returns HTTP middleware that rejects requests exceeding the
// limiter's rate for the key produced by keyFn, responding 429 with a
// Retry-After header.
func Middleware(l *Limiter, keyFn KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)
			if key != "" && !l.Allow(key) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
