package transport

import (
	"context"
	"os"
	"testing"
	"time"
)

func newTestRedisStore(t *testing.T, opts ...RedisSigStoreOption) *RedisSigStore {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set – skipping Redis tests")
	}
	// Use a unique prefix per test to avoid collisions.
	prefix := "sigstore_test_" + t.Name() + ":"
	allOpts := append([]RedisSigStoreOption{WithKeyPrefix(prefix)}, opts...)
	rs, err := NewRedisSigStore(addr, allOpts...)
	if err != nil {
		t.Fatalf("NewRedisSigStore: %v", err)
	}
	t.Cleanup(func() {
		// Clean up test keys.
		ctx := context.Background()
		keys, _ := rs.client.Keys(ctx, prefix+"*").Result()
		if len(keys) > 0 {
			rs.client.Del(ctx, keys...)
		}
		rs.Close()
	})
	return rs
}

func TestRedisSigStore_PutGetDeleteLen(t *testing.T) {
	rs := newTestRedisStore(t)

	if n := rs.Len(); n != 0 {
		t.Fatalf("expected Len()=0, got %d", n)
	}

	sig := []byte("test-signature-data")
	rs.Put("jti-1", sig)

	got, ok := rs.Get("jti-1")
	if !ok {
		t.Fatal("expected ok=true for existing key")
	}
	if string(got) != string(sig) {
		t.Fatalf("expected %q, got %q", sig, got)
	}
	if n := rs.Len(); n != 1 {
		t.Fatalf("expected Len()=1, got %d", n)
	}

	rs.Delete("jti-1")
	_, ok = rs.Get("jti-1")
	if ok {
		t.Fatal("expected ok=false after delete")
	}
	if n := rs.Len(); n != 0 {
		t.Fatalf("expected Len()=0 after delete, got %d", n)
	}
}

func TestRedisSigStore_GetMiss(t *testing.T) {
	rs := newTestRedisStore(t)

	got, ok := rs.Get("nonexistent-key")
	if ok {
		t.Fatal("expected ok=false for missing key")
	}
	if got != nil {
		t.Fatalf("expected nil sig for missing key, got %v", got)
	}
}

func TestRedisSigStore_Upsert(t *testing.T) {
	rs := newTestRedisStore(t)

	rs.Put("jti-upsert", []byte("first"))
	rs.Put("jti-upsert", []byte("second"))

	got, ok := rs.Get("jti-upsert")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if string(got) != "second" {
		t.Fatalf("expected %q after upsert, got %q", "second", got)
	}
	if n := rs.Len(); n != 1 {
		t.Fatalf("expected Len()=1 after upsert, got %d", n)
	}
}

func TestRedisSigStore_TTL(t *testing.T) {
	rs := newTestRedisStore(t, WithTTL(1*time.Second))

	rs.Put("jti-ttl", []byte("expires-soon"))
	got, ok := rs.Get("jti-ttl")
	if !ok || string(got) != "expires-soon" {
		t.Fatal("expected key to exist immediately after put")
	}

	time.Sleep(1500 * time.Millisecond)

	_, ok = rs.Get("jti-ttl")
	if ok {
		t.Fatal("expected key to have expired")
	}
}

func TestNewSigStoreWithRedisBackend(t *testing.T) {
	rs := newTestRedisStore(t)

	store := NewSigStoreWithBackend(rs)

	store.Put("jti-wrapper", []byte("wrapped-sig"))
	got, ok := store.Get("jti-wrapper")
	if !ok {
		t.Fatal("expected ok=true via SigStore wrapper")
	}
	if string(got) != "wrapped-sig" {
		t.Fatalf("expected %q, got %q", "wrapped-sig", got)
	}
}
