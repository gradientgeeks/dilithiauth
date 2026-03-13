package transport

import (
	"os"
	"testing"
)

func newTestPostgresStore(t *testing.T) *PostgresSigStore {
	t.Helper()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN not set – skipping PostgreSQL tests")
	}
	pg, err := NewPostgresSigStore(dsn)
	if err != nil {
		t.Fatalf("NewPostgresSigStore: %v", err)
	}
	t.Cleanup(func() {
		// Clean up test data.
		pg.db.Exec("DELETE FROM sig_store")
		pg.Close()
	})
	return pg
}

func TestPostgresSigStore_PutGetDeleteLen(t *testing.T) {
	pg := newTestPostgresStore(t)

	// Empty store.
	if n := pg.Len(); n != 0 {
		t.Fatalf("expected Len()=0, got %d", n)
	}

	// Put and Get.
	sig := []byte("test-signature-data")
	pg.Put("jti-1", sig)

	got, ok := pg.Get("jti-1")
	if !ok {
		t.Fatal("expected ok=true for existing key")
	}
	if string(got) != string(sig) {
		t.Fatalf("expected %q, got %q", sig, got)
	}
	if n := pg.Len(); n != 1 {
		t.Fatalf("expected Len()=1, got %d", n)
	}

	// Delete.
	pg.Delete("jti-1")
	_, ok = pg.Get("jti-1")
	if ok {
		t.Fatal("expected ok=false after delete")
	}
	if n := pg.Len(); n != 0 {
		t.Fatalf("expected Len()=0 after delete, got %d", n)
	}
}

func TestPostgresSigStore_GetMiss(t *testing.T) {
	pg := newTestPostgresStore(t)

	got, ok := pg.Get("nonexistent-key")
	if ok {
		t.Fatal("expected ok=false for missing key")
	}
	if got != nil {
		t.Fatalf("expected nil sig for missing key, got %v", got)
	}
}

func TestPostgresSigStore_Upsert(t *testing.T) {
	pg := newTestPostgresStore(t)

	pg.Put("jti-upsert", []byte("first"))
	pg.Put("jti-upsert", []byte("second"))

	got, ok := pg.Get("jti-upsert")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if string(got) != "second" {
		t.Fatalf("expected %q after upsert, got %q", "second", got)
	}
	if n := pg.Len(); n != 1 {
		t.Fatalf("expected Len()=1 after upsert, got %d", n)
	}
}

func TestNewSigStoreWithPostgresBackend(t *testing.T) {
	pg := newTestPostgresStore(t)

	store := NewSigStoreWithBackend(pg)

	store.Put("jti-wrapper", []byte("wrapped-sig"))
	got, ok := store.Get("jti-wrapper")
	if !ok {
		t.Fatal("expected ok=true via SigStore wrapper")
	}
	if string(got) != "wrapped-sig" {
		t.Fatalf("expected %q, got %q", "wrapped-sig", got)
	}
}
