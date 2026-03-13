package transport

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSigStore is a Redis-backed implementation of SigStoreBackend.
// It provides fast, shared signature storage ideal for multi-instance
// deployments where low-latency lookups are critical.
type RedisSigStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// RedisSigStoreOption configures optional RedisSigStore parameters.
type RedisSigStoreOption func(*RedisSigStore)

// WithKeyPrefix sets a custom key prefix (default: "sigstore:").
func WithKeyPrefix(prefix string) RedisSigStoreOption {
	return func(r *RedisSigStore) { r.prefix = prefix }
}

// WithTTL sets an expiration time on stored signatures.
// Default is 0 (no expiration).
func WithTTL(ttl time.Duration) RedisSigStoreOption {
	return func(r *RedisSigStore) { r.ttl = ttl }
}

// NewRedisSigStore creates a RedisSigStore connected to the given Redis address.
// The addr should be in "host:port" format (e.g. "localhost:6379").
// It verifies connectivity before returning.
func NewRedisSigStore(addr string, opts ...RedisSigStoreOption) (*RedisSigStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis_store: ping failed: %w", err)
	}

	rs := &RedisSigStore{
		client: client,
		prefix: "sigstore:",
	}
	for _, opt := range opts {
		opt(rs)
	}
	return rs, nil
}

func (r *RedisSigStore) key(jti string) string {
	return r.prefix + jti
}

func (r *RedisSigStore) Put(jti string, sig []byte) {
	ctx := context.Background()
	r.client.Set(ctx, r.key(jti), sig, r.ttl)
}

func (r *RedisSigStore) Get(jti string) ([]byte, bool) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, r.key(jti)).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		return nil, false
	}
	return val, true
}

func (r *RedisSigStore) Delete(jti string) {
	ctx := context.Background()
	r.client.Del(ctx, r.key(jti))
}

func (r *RedisSigStore) Len() int {
	ctx := context.Background()
	var count int
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, r.prefix+"*", 100).Result()
		if err != nil {
			return count
		}
		count += len(keys)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return count
}

// Close closes the underlying Redis connection.
func (r *RedisSigStore) Close() error {
	return r.client.Close()
}
