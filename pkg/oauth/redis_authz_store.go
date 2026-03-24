package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAuthzStore implements AuthzStore using Redis as the backend.
// It shares users and clients in memory for demonstration purposes,
// but distributes authorization codes through Redis.
type RedisAuthzStore struct {
	client *redis.Client
	ctx    context.Context
	mem    *MemoryAuthzStore
}

// NewRedisAuthzStore creates a new Redis-backed AuthzStore.
func NewRedisAuthzStore(redisAddr string) (*RedisAuthzStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisAuthzStore{
		client: rdb,
		ctx:    ctx,
		mem:    NewAuthzStore(), // Fallback to memory store for demo users and clients
	}, nil
}

// AddUser registers a demo user (in memory for demo).
func (s *RedisAuthzStore) AddUser(u *DemoUser) {
	s.mem.AddUser(u)
}

// AddClient registers a demo client (in memory for demo).
func (s *RedisAuthzStore) AddClient(c *DemoClient) {
	s.mem.AddClient(c)
}

// AuthenticateUser checks username/password.
func (s *RedisAuthzStore) AuthenticateUser(username, password string) (*DemoUser, error) {
	return s.mem.AuthenticateUser(username, password)
}

// ValidateClient checks that the client_id exists and the redirect_uri matches.
func (s *RedisAuthzStore) ValidateClient(clientID, redirectURI string) (*DemoClient, error) {
	return s.mem.ValidateClient(clientID, redirectURI)
}

// CreateCode generates and stores an authorization code in Redis.
func (s *RedisAuthzStore) CreateCode(subject, clientID, redirectURI, scope string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(buf)
	expiresAt := time.Now().Add(10 * time.Minute)

	ac := &AuthCode{
		Code:        code,
		Subject:     subject,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		ExpiresAt:   expiresAt,
		Used:        false,
	}

	data, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	// Store in Redis with TTL
	key := "authcode:" + code
	if err := s.client.Set(s.ctx, key, data, 10*time.Minute).Err(); err != nil {
		return "", err
	}

	return code, nil
}

// RedeemCode validates and consumes an authorization code from Redis.
func (s *RedisAuthzStore) RedeemCode(code, clientID, redirectURI string) (*AuthCode, error) {
	key := "authcode:" + code

	// Start a transaction to ensure single-use
	txf := func(tx *redis.Tx) error {
		// Get the current code
		res, err := tx.Get(s.ctx, key).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		if err == redis.Nil {
			return errors.New("invalid code")
		}

		var ac AuthCode
		if err := json.Unmarshal([]byte(res), &ac); err != nil {
			return err
		}

		if ac.Used {
			return errors.New("code already used")
		}
		if time.Now().After(ac.ExpiresAt) {
			return errors.New("code expired")
		}
		if ac.ClientID != clientID {
			return errors.New("client_id mismatch")
		}
		if ac.RedirectURI != redirectURI {
			return errors.New("redirect_uri mismatch")
		}

		// Mark as used
		ac.Used = true
		newData, err := json.Marshal(ac)
		if err != nil {
			return err
		}

		// Save back within transaction
		_, err = tx.TxPipelined(s.ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(s.ctx, key, newData, time.Until(ac.ExpiresAt))
			return nil
		})
		return err
	}

	// Watch the key for changes
	err := s.client.Watch(s.ctx, txf, key)
	if err != nil {
		if err.Error() == "invalid code" || err.Error() == "code already used" ||
			err.Error() == "code expired" || err.Error() == "client_id mismatch" ||
			err.Error() == "redirect_uri mismatch" {
			return nil, err
		}
		return nil, errors.New("failed to redeem code or code already redeemed")
	}

	// Fetch again to return (it was just marked as used)
	res, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var ac AuthCode
	if err := json.Unmarshal([]byte(res), &ac); err != nil {
		return nil, err
	}

	return &ac, nil
}
