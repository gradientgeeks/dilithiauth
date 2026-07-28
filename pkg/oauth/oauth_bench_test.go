package oauth_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

const (
	testTenantID = "tenant-bench-1"
	testSubject  = "user_sub_12345"
	testTTL      = time.Hour
)

func TestTokenSizesAndHeaderOverhead(t *testing.T) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair failed: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)

	// Issue Thin JWT
	thinRes, err := provider.IssueJWTThin(testTenantID, testSubject, testTTL)
	if err != nil {
		t.Fatalf("IssueJWTThin failed: %v", err)
	}

	// Issue Jumbo JWT
	jumboToken, err := provider.IssueJWTFull(testTenantID, testSubject, testTTL)
	if err != nil {
		t.Fatalf("IssueJWTFull failed: %v", err)
	}

	// Issue CWT
	cwtRes, err := provider.IssueCWT(testTenantID, testSubject, testTTL)
	if err != nil {
		t.Fatalf("IssueCWT failed: %v", err)
	}

	cwtB64 := base64.RawURLEncoding.EncodeToString(cwtRes.Token)

	thinHeader := "Authorization: Bearer " + thinRes.Token
	jumboHeader := "Authorization: Bearer " + jumboToken
	cwtHeader := "Authorization: Bearer " + cwtB64

	t.Logf("=== EMPIRICAL TOKEN SIZES ===")
	t.Logf("Thin JWT size:          %d bytes", len(thinRes.Token))
	t.Logf("Jumbo JWT size:         %d bytes", len(jumboToken))
	t.Logf("CWT (binary) size:      %d bytes", len(cwtRes.Token))
	t.Logf("CWT (Base64url) size:   %d bytes", len(cwtB64))
	t.Logf("=== HTTP HEADER OVERHEAD ===")
	t.Logf("Thin JWT Header size:   %d bytes", len(thinHeader))
	t.Logf("Jumbo JWT Header size:  %d bytes", len(jumboHeader))
	t.Logf("CWT Header size:        %d bytes", len(cwtHeader))
	t.Logf("=== SIDECAR STORAGE OVERHEAD ===")
	t.Logf("ML-DSA-44 Raw Signature:%d bytes", len(thinRes.PQCSig))
	t.Logf("Thin PQCRef Key size:   %d bytes (%q)", len(thinRes.PQCRef), thinRes.PQCRef)
	t.Logf("CWT JWTID Key size:     %d bytes (%q)", len(cwtRes.JWTID), cwtRes.JWTID)
	t.Logf("Total Thin Entry Size:  %d bytes (Key + Sig)", len(thinRes.PQCRef)+len(thinRes.PQCSig))
	t.Logf("Total CWT Entry Size:   %d bytes (Key + Sig)", len(cwtRes.JWTID)+len(cwtRes.PQCSig))

	if len(thinRes.Token) >= len(jumboToken) {
		t.Errorf("expected Thin JWT (%d) to be smaller than Jumbo JWT (%d)", len(thinRes.Token), len(jumboToken))
	}
	if len(cwtRes.Token) >= len(thinRes.Token) {
		t.Errorf("expected binary CWT (%d) to be smaller than Thin JWT (%d)", len(cwtRes.Token), len(thinRes.Token))
	}
}

func BenchmarkIssue_ThinJWT(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.IssueJWTThin(testTenantID, testSubject, testTTL)
		if err != nil {
			b.Fatalf("IssueJWTThin error: %v", err)
		}
	}
}

func BenchmarkIssue_JumboJWT(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.IssueJWTFull(testTenantID, testSubject, testTTL)
		if err != nil {
			b.Fatalf("IssueJWTFull error: %v", err)
		}
	}
}

func BenchmarkIssue_CWT(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.IssueCWT(testTenantID, testSubject, testTTL)
		if err != nil {
			b.Fatalf("IssueCWT error: %v", err)
		}
	}
}

func BenchmarkVerify_ThinJWT(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)
	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	thinRes, err := provider.IssueJWTThin(testTenantID, testSubject, testTTL)
	if err != nil {
		b.Fatalf("IssueJWTThin error: %v", err)
	}
	sigStore.Put(thinRes.PQCRef, thinRes.PQCSig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.VerifyJWTThin(thinRes.Token)
		if err != nil {
			b.Fatalf("VerifyJWTThin error: %v", err)
		}
	}
}

func BenchmarkVerify_FastPath_ThinJWT(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)
	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	thinRes, err := provider.IssueJWTThin(testTenantID, testSubject, testTTL)
	if err != nil {
		b.Fatalf("IssueJWTThin error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.VerifyFastPath(thinRes.Token)
		if err != nil {
			b.Fatalf("VerifyFastPath error: %v", err)
		}
	}
}

func BenchmarkVerify_JumboJWT(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)
	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	jumboToken, err := provider.IssueJWTFull(testTenantID, testSubject, testTTL)
	if err != nil {
		b.Fatalf("IssueJWTFull error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.VerifyJWTFull(jumboToken)
		if err != nil {
			b.Fatalf("VerifyJWTFull error: %v", err)
		}
	}
}

func BenchmarkVerify_CWT(b *testing.B) {
	kp, err := signer.GenerateHybridKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}
	provider := oauth.NewProvider("https://auth.example.com", kp)
	sigStore := transport.NewSigStore()
	v := verifier.NewVerifierWithKeyPair(kp, sigStore)

	cwtRes, err := provider.IssueCWT(testTenantID, testSubject, testTTL)
	if err != nil {
		b.Fatalf("IssueCWT error: %v", err)
	}
	sigStore.Put(cwtRes.JWTID, cwtRes.PQCSig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.VerifyCWT(cwtRes.Token)
		if err != nil {
			b.Fatalf("VerifyCWT error: %v", err)
		}
	}
}
