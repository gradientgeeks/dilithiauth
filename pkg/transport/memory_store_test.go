package transport_test

import (
	"testing"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
)

func TestMemorySigStore_NoTTL_NeverExpires(t *testing.T) {
	s := transport.NewMemorySigStore()
	s.Put("jti1", []byte("sig1"))
	time.Sleep(20 * time.Millisecond)
	if _, ok := s.Get("jti1"); !ok {
		t.Error("entry without a TTL should never expire")
	}
}

func TestMemorySigStore_WithTTL_ExpiresAndIsHiddenFromGet(t *testing.T) {
	s := transport.NewMemorySigStoreWithTTL(10 * time.Millisecond)
	s.Put("jti1", []byte("sig1"))

	if _, ok := s.Get("jti1"); !ok {
		t.Fatal("entry should be readable immediately after Put")
	}

	time.Sleep(20 * time.Millisecond)
	if _, ok := s.Get("jti1"); ok {
		t.Error("expired entry should no longer be returned by Get, even before Purge runs")
	}
}

func TestMemorySigStore_Purge_RemovesExpiredEntries(t *testing.T) {
	s := transport.NewMemorySigStoreWithTTL(10 * time.Millisecond)
	s.Put("jti1", []byte("sig1"))
	time.Sleep(20 * time.Millisecond)

	if got := s.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1 before Purge (Purge must be called to actually reclaim memory)", got)
	}
	s.Purge()
	if got := s.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0 after Purge", got)
	}
}
