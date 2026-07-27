package oauth_test

import (
	"testing"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
)

func TestMemoryRefreshTokenStore_CreateGet(t *testing.T) {
	s := oauth.NewMemoryRefreshTokenStore()
	rec := &oauth.RefreshTokenRecord{
		Token: "rt_abc", Subject: "alice", TenantID: "default", ClientID: "demo-client",
		FamilyID: "fam1", ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := s.Create(rec); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	got, err := s.Get("rt_abc")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Subject != "alice" {
		t.Errorf("Subject = %q, want %q", got.Subject, "alice")
	}
}

func TestMemoryRefreshTokenStore_GetMissing(t *testing.T) {
	s := oauth.NewMemoryRefreshTokenStore()
	if _, err := s.Get("nonexistent"); err == nil {
		t.Error("Get() for a missing token should fail")
	}
}

func TestMemoryRefreshTokenStore_RevokeFamily(t *testing.T) {
	s := oauth.NewMemoryRefreshTokenStore()
	rec := &oauth.RefreshTokenRecord{
		Token: "rt_abc", FamilyID: "fam1", ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := s.Create(rec); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := s.RevokeFamily("fam1"); err != nil {
		t.Fatalf("RevokeFamily() error = %v", err)
	}
	if _, err := s.Get("rt_abc"); err == nil {
		t.Error("Get() should fail for a token in a revoked family")
	}
}

func TestMemoryRefreshTokenStore_Purge(t *testing.T) {
	s := oauth.NewMemoryRefreshTokenStore()
	expired := &oauth.RefreshTokenRecord{Token: "rt_old", FamilyID: "fam1", ExpiresAt: time.Now().Add(-time.Hour)}
	fresh := &oauth.RefreshTokenRecord{Token: "rt_new", FamilyID: "fam2", ExpiresAt: time.Now().Add(time.Hour)}
	_ = s.Create(expired)
	_ = s.Create(fresh)

	s.Purge()

	if _, err := s.Get("rt_old"); err == nil {
		t.Error("Purge() should have removed the expired token")
	}
	if _, err := s.Get("rt_new"); err != nil {
		t.Error("Purge() should not remove a token that hasn't expired")
	}
}

func TestMemoryRefreshTokenStore_MarkUsed(t *testing.T) {
	s := oauth.NewMemoryRefreshTokenStore()
	rec := &oauth.RefreshTokenRecord{Token: "rt_abc", FamilyID: "fam1", ExpiresAt: time.Now().Add(time.Hour)}
	_ = s.Create(rec)

	if err := s.MarkUsed("rt_abc"); err != nil {
		t.Fatalf("MarkUsed() error = %v", err)
	}
	got, err := s.Get("rt_abc")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !got.Used {
		t.Error("record should be marked Used after MarkUsed()")
	}

	if err := s.MarkUsed("nonexistent"); err == nil {
		t.Error("MarkUsed() for a missing token should fail")
	}
}
