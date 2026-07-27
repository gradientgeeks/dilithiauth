package developer_test

import (
	"testing"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/developer"
)

func TestOTPVerificationStore(t *testing.T) {
	store := developer.NewOTPStore(nil)
	email := "alice@example.com"

	code, expiresAt, err := store.GenerateCode(email)
	if err != nil {
		t.Fatalf("failed to generate OTP code: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %s", code)
	}

	if expiresAt.IsZero() {
		t.Errorf("expected non-zero expiration time")
	}

	t.Run("Verify Correct Code", func(t *testing.T) {
		valid, err := store.VerifyCode(email, code)
		if err != nil {
			t.Fatalf("expected valid verification, got error: %v", err)
		}
		if !valid {
			t.Fatalf("expected code to be valid")
		}
	})

	t.Run("Verify Single Use Code", func(t *testing.T) {
		_, err := store.VerifyCode(email, code)
		if err == nil {
			t.Fatalf("expected error on reused code")
		}
	})
}
