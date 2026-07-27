package oauth_test

import (
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
)

func TestMemoryGrantStore_SaveGet(t *testing.T) {
	s := oauth.NewMemoryGrantStore()
	if g, err := s.Get("t1", "alice", "client1"); err != nil || g != nil {
		t.Fatalf("Get() before Save = (%v, %v), want (nil, nil)", g, err)
	}

	g := &oauth.Grant{Subject: "alice", TenantID: "t1", ClientID: "client1", Scope: "openid email"}
	if err := s.Save(g); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := s.Get("t1", "alice", "client1")
	if err != nil || got == nil {
		t.Fatalf("Get() = (%v, %v), want a saved grant", got, err)
	}
	if got.Scope != "openid email" {
		t.Errorf("Scope = %q, want %q", got.Scope, "openid email")
	}

	// Different tenant/client must not see the grant.
	if g, _ := s.Get("t2", "alice", "client1"); g != nil {
		t.Error("grant must be tenant-scoped")
	}
	if g, _ := s.Get("t1", "alice", "client2"); g != nil {
		t.Error("grant must be client-scoped")
	}
}

func TestScopeContains(t *testing.T) {
	tests := []struct {
		scope, want string
		expect      bool
	}{
		{"openid email", "email", true},
		{"openid email", "profile", false},
		{"", "email", false},
		{"openid  email", "openid", true},
	}
	for _, tt := range tests {
		if got := oauth.ScopeContains(tt.scope, tt.want); got != tt.expect {
			t.Errorf("ScopeContains(%q, %q) = %v, want %v", tt.scope, tt.want, got, tt.expect)
		}
	}
}
