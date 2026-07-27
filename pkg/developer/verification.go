package developer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"
)

// OTPVerificationStore manages 6-digit email verification codes.
type OTPVerificationStore struct {
	mu     sync.RWMutex
	codes  map[string]otpItem // email -> otpItem
	Mailer *Mailer
}

type otpItem struct {
	Code      string
	ExpiresAt time.Time
}

// NewOTPStore creates an OTP verification store, optionally attached to a Mailer.
func NewOTPStore(mailer *Mailer) *OTPVerificationStore {
	return &OTPVerificationStore{
		codes:  make(map[string]otpItem),
		Mailer: mailer,
	}
}

// GenerateCode creates a 6-digit cryptographic numeric OTP valid for 15 minutes.
func (s *OTPVerificationStore) GenerateCode(email string) (string, time.Time, error) {
	email = normalizeEmail(email)
	if email == "" {
		return "", time.Time{}, errors.New("invalid email address")
	}

	// Generate random number between 100000 and 999999
	nBig, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generating OTP code: %w", err)
	}
	code := fmt.Sprintf("%06d", nBig.Int64()+100000)

	expiresAt := time.Now().Add(15 * time.Minute)

	s.mu.Lock()
	s.codes[email] = otpItem{
		Code:      code,
		ExpiresAt: expiresAt,
	}
	s.mu.Unlock()

	// Send email asynchronously via OCI Email Delivery if configured
	if s.Mailer != nil {
		go func(e, c string) {
			_ = s.Mailer.SendVerificationOTP(e, c)
		}(email, code)
	} else {
		log.Printf("[DEV MODE - NO SMTP] Verification code for %s: %s (expires in 15m)", email, code)
	}

	return code, expiresAt, nil
}

// VerifyCode checks whether the provided OTP code matches and has not expired.
func (s *OTPVerificationStore) VerifyCode(email, code string) (bool, error) {
	email = normalizeEmail(email)
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.codes[email]
	if !ok {
		return false, errors.New("no verification code found for this email — please request a new code")
	}

	if time.Now().After(item.ExpiresAt) {
		delete(s.codes, email)
		return false, errors.New("verification code has expired — please request a new code")
	}

	if item.Code != code {
		return false, errors.New("invalid verification code")
	}

	// Code is valid — clear single-use code
	delete(s.codes, email)
	return true, nil
}

func normalizeEmail(email string) string {
	return fmt.Sprintf("%s", email)
}
