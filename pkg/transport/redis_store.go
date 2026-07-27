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
type RedisSigStoreOption func(*redisSigStoreConfig)

type redisSigStoreConfig struct {
	prefix   string
	ttl      time.Duration
	username string
	password string
}

// WithKeyPrefix sets a custom key prefix (default: "sigstore:").
func WithKeyPrefix(prefix string) RedisSigStoreOption {
	return func(c *redisSigStoreConfig) { c.prefix = prefix }
}

// WithTTL sets an expiration time on stored signatures.
// Default is 0 (no expiration).
func WithTTL(ttl time.Duration) RedisSigStoreOption {
	return func(c *redisSigStoreConfig) { c.ttl = ttl }
}

// WithAuth sets the Redis ACL username/password used to authenticate the
// connection. Username may be left empty to authenticate against the
// legacy single-password `requirepass` scheme instead of a named ACL user.
func WithAuth(username, password string) RedisSigStoreOption {
	return func(c *redisSigStoreConfig) { c.username = username; c.password = password }
}

// NewRedisSigStore creates a RedisSigStore connected to the given Redis address.
// The addr should be in "host:port" format (e.g. "localhost:6379").
// It verifies connectivity before returning.
func NewRedisSigStore(addr string, opts ...RedisSigStoreOption) (*RedisSigStore, error) {
	cfg := &redisSigStoreConfig{prefix: "sigstore:"}
	for _, opt := range opts {
		opt(cfg)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: cfg.username,
		Password: cfg.password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis_store: ping failed: %w", err)
	}

	rs := &RedisSigStore{
		client: client,
		prefix: cfg.prefix,
		ttl:    cfg.ttl,
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

// Ping verifies the Redis connection is reachable, for use by readiness
// checks (e.g. an HTTP /readyz handler). It respects ctx's deadline/cancellation.
func (r *RedisSigStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
