package mfa

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultStepSize = 30 // 30-second time step window
	DefaultDigits   = 6  // 6-digit passcode
)

// TOTP handles Time-based One-Time Password generation and verification (RFC 6238).
type TOTP struct {
	StepSize time.Duration
	Digits   int
}

func NewTOTP() *TOTP {
	return &TOTP{
		StepSize: DefaultStepSize * time.Second,
		Digits:   DefaultDigits,
	}
}

// GenerateSecret generates a random 20-byte (160-bit) Base32-encoded TOTP secret key.
func GenerateSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// GenerateKeyURI constructs the standard otpauth:// URL for authenticator apps (Google Authenticator / Authy).
func GenerateKeyURI(secret, accountName, issuer string) string {
	encodedIssuer := url.QueryEscape(issuer)
	encodedAccount := url.QueryEscape(accountName)
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		encodedIssuer, encodedAccount, secret, encodedIssuer, DefaultDigits, DefaultStepSize)
}

// GenerateCode calculates the current 6-digit TOTP passcode for a given timestamp and secret key.
func (t *TOTP) GenerateCode(secret string, timestamp time.Time) (string, error) {
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("invalid base32 secret: %w", err)
	}

	counter := uint64(timestamp.Unix()) / uint64(t.StepSize.Seconds())
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, secretBytes)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff
	codeNum := truncated % 1000000

	return fmt.Sprintf("%0*d", t.Digits, codeNum), nil
}

// ValidateCode verifies a TOTP passcode against a secret key with a skew window of ±1 step (90s window).
func (t *TOTP) ValidateCode(secret, code string, timestamp time.Time) bool {
	if len(code) != t.Digits {
		return false
	}

	for skew := -1; skew <= 1; skew++ {
		skewedTime := timestamp.Add(time.Duration(skew) * t.StepSize)
		generatedCode, err := t.GenerateCode(secret, skewedTime)
		if err != nil {
			continue
		}
		if hmac.Equal([]byte(generatedCode), []byte(code)) {
			return true
		}
	}
	return false
}
