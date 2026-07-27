package oauth

import (
	"testing"
)

func TestGetSocialAuthURL(t *testing.T) {
	cfg := SocialConfig{
		ClientID:    "test_client_id",
		RedirectURI: "https://myapp.com/callback",
	}

	googleURL, err := GetSocialAuthURL(ProviderGoogle, cfg, "state123")
	if err != nil {
		t.Fatalf("GetSocialAuthURL Google error: %v", err)
	}
	if googleURL == "" || len(googleURL) < 20 {
		t.Errorf("Google auth URL is invalid: %s", googleURL)
	}

	githubURL, err := GetSocialAuthURL(ProviderGitHub, cfg, "state123")
	if err != nil {
		t.Fatalf("GetSocialAuthURL GitHub error: %v", err)
	}
	if githubURL == "" || len(githubURL) < 20 {
		t.Errorf("GitHub auth URL is invalid: %s", githubURL)
	}
}
