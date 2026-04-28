package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAuthzStore implements AuthzStore using Redis for authorization codes
// and the embedded MemoryAuthzStore for users and clients.
// Redis keys are scoped by tenantID to provide logical isolation between tenants.
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
		mem:    NewAuthzStore(),
	}, nil
}

func (s *RedisAuthzStore) AddUser(tenantID string, u *User) {
	s.mem.AddUser(tenantID, u)
}

func (s *RedisAuthzStore) AddClient(tenantID string, c *OAuthClient) {
	s.mem.AddClient(tenantID, c)
}

func (s *RedisAuthzStore) RegisterUser(tenantID string, u *User) error {
	return s.mem.RegisterUser(tenantID, u)
}

func (s *RedisAuthzStore) AuthenticateUser(tenantID, username, password string) (*User, error) {
	return s.mem.AuthenticateUser(tenantID, username, password)
}

func (s *RedisAuthzStore) ValidateClient(tenantID, clientID, redirectURI string) (*OAuthClient, error) {
	return s.mem.ValidateClient(tenantID, clientID, redirectURI)
}

func (s *RedisAuthzStore) ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI string) (*OAuthClient, error) {
	return s.mem.ValidateClientWithSecret(tenantID, clientID, clientSecret, redirectURI)
}

// CreateCode generates and stores an authorization code in Redis, scoped by tenantID.
func (s *RedisAuthzStore) CreateCode(tenantID, subject, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod string) (string, error) {
	if codeChallenge != "" && codeChallengeMethod == "" {
		codeChallengeMethod = "S256"
	}
	if codeChallenge != "" && codeChallengeMethod != "S256" {
		return "", errors.New("pkce: only S256 code_challenge_method is supported")
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(buf)
	expiresAt := time.Now().Add(10 * time.Minute)

	ac := &AuthCode{
		Code:                code,
		Subject:             subject,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		ExpiresAt:           expiresAt,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}

	data, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	key := "tenant:" + tenantID + ":authcode:" + code
	if err := s.client.Set(s.ctx, key, data, 10*time.Minute).Err(); err != nil {
		return "", err
	}
	return code, nil
}

// RedeemCode validates and atomically consumes an authorization code from Redis.
func (s *RedisAuthzStore) RedeemCode(tenantID, code, clientID, redirectURI, codeVerifier string) (*AuthCode, error) {
	key := "tenant:" + tenantID + ":authcode:" + code

	txf := func(tx *redis.Tx) error {
		res, err := tx.Get(s.ctx, key).Result()
		if err == redis.Nil {
			return errors.New("invalid code")
		}
		if err != nil {
			return err
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
		if ac.CodeChallenge != "" {
			if codeVerifier == "" {
				return errors.New("pkce: code_verifier required")
			}
			h := sha256.Sum256([]byte(codeVerifier))
			computed := base64.RawURLEncoding.EncodeToString(h[:])
			if subtle.ConstantTimeCompare([]byte(computed), []byte(ac.CodeChallenge)) != 1 {
				return errors.New("pkce: code_verifier mismatch")
			}
		}

		ac.Used = true
		newData, err := json.Marshal(ac)
		if err != nil {
			return err
		}
		_, err = tx.TxPipelined(s.ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(s.ctx, key, newData, time.Until(ac.ExpiresAt))
			return nil
		})
		return err
	}

	if err := s.client.Watch(s.ctx, txf, key); err != nil {
		return nil, err
	}

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
