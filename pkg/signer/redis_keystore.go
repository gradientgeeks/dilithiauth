package signer

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// RedisKeyStore implements KeyStore using Redis.
type RedisKeyStore struct {
	client *redis.Client
	ctx    context.Context
	prefix string
}

// NewRedisKeyStore creates a new RedisKeyStore.
func NewRedisKeyStore(redisAddr string) (*RedisKeyStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisKeyStore{
		client: rdb,
		ctx:    ctx,
		prefix: "dilithiauth:keys:",
	}, nil
}

func (s *RedisKeyStore) Save(kp *HybridKeyPair) error {
	data, err := MarshalKeyPair(kp)
	if err != nil {
		return fmt.Errorf("signer: marshalling key %s: %w", kp.Kid, err)
	}

	key := s.prefix + kp.Kid
	if err := s.client.Set(s.ctx, key, string(data), 0).Err(); err != nil {
		return fmt.Errorf("signer: writing key to redis: %w", err)
	}

	// Add to set of known kids for LoadAll
	if err := s.client.SAdd(s.ctx, s.prefix+"all", kp.Kid).Err(); err != nil {
		return fmt.Errorf("signer: tracking key in redis: %w", err)
	}

	return nil
}

func (s *RedisKeyStore) Load(kid string) (*HybridKeyPair, error) {
	key := s.prefix + kid
	data, err := s.client.Get(s.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("signer: key %s not found in redis", kid)
		}
		return nil, fmt.Errorf("signer: reading key from redis: %w", err)
	}

	return UnmarshalKeyPair(data)
}

func (s *RedisKeyStore) LoadAll() ([]*HybridKeyPair, error) {
	kids, err := s.client.SMembers(s.ctx, s.prefix+"all").Result()
	if err != nil {
		return nil, fmt.Errorf("signer: reading key list from redis: %w", err)
	}

	var keys []*HybridKeyPair
	for _, kid := range kids {
		kp, err := s.Load(kid)
		if err != nil {
			// Skip keys that failed to load
			continue
		}
		keys = append(keys, kp)
	}

	return keys, nil
}

func (s *RedisKeyStore) SetActiveKid(kid string) error {
	if err := s.client.Set(s.ctx, s.prefix+"active", kid, 0).Err(); err != nil {
		return fmt.Errorf("signer: setting active key in redis: %w", err)
	}
	return nil
}

func (s *RedisKeyStore) GetActiveKid() (string, error) {
	kid, err := s.client.Get(s.ctx, s.prefix+"active").Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("signer: no active key set in redis")
		}
		return "", fmt.Errorf("signer: getting active key from redis: %w", err)
	}
	return strings.TrimSpace(kid), nil
}
