package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SocialProvider represents supported Social Identity Providers.
type SocialProvider string

const (
	ProviderGoogle SocialProvider = "google"
	ProviderGitHub SocialProvider = "github"
)

// SocialUser represents unified social profile claims returned from Google/GitHub.
type SocialUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Provider SocialProvider `json:"provider"`
}

// SocialConfig holds client credentials for a Social Identity Provider.
type SocialConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// GetSocialAuthURL constructs authorization redirect URL for Google or GitHub.
func GetSocialAuthURL(provider SocialProvider, cfg SocialConfig, state string) (string, error) {
	switch provider {
	case ProviderGoogle:
		u, _ := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
		q := u.Query()
		q.Set("client_id", cfg.ClientID)
		q.Set("redirect_uri", cfg.RedirectURI)
		q.Set("response_type", "code")
		q.Set("scope", "openid profile email")
		q.Set("state", state)
		u.RawQuery = q.Encode()
		return u.String(), nil

	case ProviderGitHub:
		u, _ := url.Parse("https://github.com/login/oauth/authorize")
		q := u.Query()
		q.Set("client_id", cfg.ClientID)
		q.Set("redirect_uri", cfg.RedirectURI)
		q.Set("scope", "read:user user:email")
		q.Set("state", state)
		u.RawQuery = q.Encode()
		return u.String(), nil

	default:
		return "", fmt.Errorf("unsupported social provider: %s", provider)
	}
}

// ExchangeSocialCode exchanges authorization code for social profile claims.
func ExchangeSocialCode(ctx context.Context, provider SocialProvider, cfg SocialConfig, code string) (*SocialUser, error) {
	httpClient := &http.Client{}

	switch provider {
	case ProviderGoogle:
		data := url.Values{}
		data.Set("code", code)
		data.Set("client_id", cfg.ClientID)
		data.Set("client_secret", cfg.ClientSecret)
		data.Set("redirect_uri", cfg.RedirectURI)
		data.Set("grant_type", "authorization_code")

		req, _ := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("google token request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("google token error (%d): %s", resp.StatusCode, string(body))
		}

		var tok map[string]interface{}
		json.Unmarshal(body, &tok)
		accessToken, _ := tok["access_token"].(string)

		// Fetch userinfo
		userReq, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		userReq.Header.Set("Authorization", "Bearer "+accessToken)
		userResp, err := httpClient.Do(userReq)
		if err != nil {
			return nil, fmt.Errorf("google userinfo request failed: %w", err)
		}
		defer userResp.Body.Close()

		userBody, _ := io.ReadAll(userResp.Body)
		var gUser map[string]interface{}
		json.Unmarshal(userBody, &gUser)

		return &SocialUser{
			ID:       fmt.Sprintf("%v", gUser["id"]),
			Email:    fmt.Sprintf("%v", gUser["email"]),
			Name:     fmt.Sprintf("%v", gUser["name"]),
			Provider: ProviderGoogle,
		}, nil

	case ProviderGitHub:
		data := url.Values{}
		data.Set("code", code)
		data.Set("client_id", cfg.ClientID)
		data.Set("client_secret", cfg.ClientSecret)
		data.Set("redirect_uri", cfg.RedirectURI)

		req, _ := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("github token request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("github token error (%d): %s", resp.StatusCode, string(body))
		}

		var tok map[string]interface{}
		json.Unmarshal(body, &tok)
		accessToken, _ := tok["access_token"].(string)

		// Fetch userinfo
		userReq, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
		userReq.Header.Set("Authorization", "Bearer "+accessToken)
		userReq.Header.Set("User-Agent", "DilithiAuth-Server")
		userResp, err := httpClient.Do(userReq)
		if err != nil {
			return nil, fmt.Errorf("github user request failed: %w", err)
		}
		defer userResp.Body.Close()

		userBody, _ := io.ReadAll(userResp.Body)
		var ghUser map[string]interface{}
		json.Unmarshal(userBody, &ghUser)

		return &SocialUser{
			ID:       fmt.Sprintf("%v", ghUser["id"]),
			Email:    fmt.Sprintf("%v", ghUser["email"]),
			Name:     fmt.Sprintf("%v", ghUser["name"]),
			Provider: ProviderGitHub,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported social provider: %s", provider)
	}
}
