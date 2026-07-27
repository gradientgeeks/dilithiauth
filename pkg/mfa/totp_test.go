package mfa

import (
	"testing"
	"time"
)

func TestTOTPGenerationAndValidation(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}
	if len(secret) == 0 {
		t.Fatal("GenerateSecret returned empty string")
	}

	totp := NewTOTP()
	now := time.Now()

	code, err := totp.GenerateCode(secret, now)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("Expected 6-digit code, got %s", code)
	}

	valid := totp.ValidateCode(secret, code, now)
	if !valid {
		t.Errorf("ValidateCode failed for valid code %s", code)
	}

	invalid := totp.ValidateCode(secret, "000000", now)
	if invalid && code != "000000" {
		t.Errorf("ValidateCode succeeded for invalid code 000000")
	}
}

func TestKeyURIGeneration(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	uri := GenerateKeyURI(secret, "alice@example.com", "DilithiAuth")
	expectedPrefix := "otpauth://totp/DilithiAuth:alice%40example.com?secret=JBSWY3DPEHPK3PXP"
	if len(uri) == 0 || uri[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("GenerateKeyURI returned unexpected URI: %s", uri)
	}
}
