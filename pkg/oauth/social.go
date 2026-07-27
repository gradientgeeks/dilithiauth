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
	TokenURL     string // Optional custom token endpoint URL (for testing or proxying)
	UserInfoURL  string // Optional custom userinfo endpoint URL (for testing or proxying)
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
		tokenURL := cfg.TokenURL
		if tokenURL == "" {
			tokenURL = "https://oauth2.googleapis.com/token"
		}
		userinfoURL := cfg.UserInfoURL
		if userinfoURL == "" {
			userinfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
		}

		data := url.Values{}
		data.Set("code", code)
		data.Set("client_id", cfg.ClientID)
		data.Set("client_secret", cfg.ClientSecret)
		data.Set("redirect_uri", cfg.RedirectURI)
		data.Set("grant_type", "authorization_code")

		req, _ := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
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
		if err := json.Unmarshal(body, &tok); err != nil {
			return nil, fmt.Errorf("google token json parse error: %w", err)
		}
		accessToken, _ := tok["access_token"].(string)
		if accessToken == "" {
			return nil, fmt.Errorf("google token response missing access_token")
		}

		// Fetch userinfo
		userReq, _ := http.NewRequestWithContext(ctx, "GET", userinfoURL, nil)
		userReq.Header.Set("Authorization", "Bearer "+accessToken)
		userResp, err := httpClient.Do(userReq)
		if err != nil {
			return nil, fmt.Errorf("google userinfo request failed: %w", err)
		}
		defer userResp.Body.Close()

		userBody, _ := io.ReadAll(userResp.Body)
		if userResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("google userinfo error (%d): %s", userResp.StatusCode, string(userBody))
		}
		var gUser map[string]interface{}
		if err := json.Unmarshal(userBody, &gUser); err != nil {
			return nil, fmt.Errorf("google userinfo json parse error: %w", err)
		}

		idVal := ""
		if id, ok := gUser["id"]; ok && id != nil {
			idVal = fmt.Sprintf("%v", id)
		} else if sub, ok := gUser["sub"]; ok && sub != nil {
			idVal = fmt.Sprintf("%v", sub)
		}

		emailVal := ""
		if em, ok := gUser["email"]; ok && em != nil {
			emailVal = fmt.Sprintf("%v", em)
			if emailVal == "<nil>" {
				emailVal = ""
			}
		}

		nameVal := ""
		if n, ok := gUser["name"]; ok && n != nil {
			nameVal = fmt.Sprintf("%v", n)
			if nameVal == "<nil>" {
				nameVal = ""
			}
		}

		return &SocialUser{
			ID:       idVal,
			Email:    emailVal,
			Name:     nameVal,
			Provider: ProviderGoogle,
		}, nil

	case ProviderGitHub:
		tokenURL := cfg.TokenURL
		if tokenURL == "" {
			tokenURL = "https://github.com/login/oauth/access_token"
		}
		userinfoURL := cfg.UserInfoURL
		if userinfoURL == "" {
			userinfoURL = "https://api.github.com/user"
		}

		data := url.Values{}
		data.Set("code", code)
		data.Set("client_id", cfg.ClientID)
		data.Set("client_secret", cfg.ClientSecret)
		data.Set("redirect_uri", cfg.RedirectURI)

		req, _ := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
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
		if err := json.Unmarshal(body, &tok); err != nil {
			return nil, fmt.Errorf("github token json parse error: %w", err)
		}
		accessToken, _ := tok["access_token"].(string)
		if accessToken == "" {
			return nil, fmt.Errorf("github token response missing access_token")
		}

		// Fetch userinfo
		userReq, _ := http.NewRequestWithContext(ctx, "GET", userinfoURL, nil)
		userReq.Header.Set("Authorization", "Bearer "+accessToken)
		userReq.Header.Set("User-Agent", "DilithiAuth-Server")
		userResp, err := httpClient.Do(userReq)
		if err != nil {
			return nil, fmt.Errorf("github user request failed: %w", err)
		}
		defer userResp.Body.Close()

		userBody, _ := io.ReadAll(userResp.Body)
		if userResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("github user error (%d): %s", userResp.StatusCode, string(userBody))
		}
		var ghUser map[string]interface{}
		if err := json.Unmarshal(userBody, &ghUser); err != nil {
			return nil, fmt.Errorf("github user json parse error: %w", err)
		}

		idVal := ""
		if id, ok := ghUser["id"]; ok && id != nil {
			idVal = fmt.Sprintf("%v", id)
		}

		emailVal := ""
		if em, ok := ghUser["email"]; ok && em != nil {
			emailVal = fmt.Sprintf("%v", em)
			if emailVal == "<nil>" {
				emailVal = ""
			}
		}

		nameVal := ""
		if n, ok := ghUser["name"]; ok && n != nil {
			nameVal = fmt.Sprintf("%v", n)
			if nameVal == "<nil>" {
				nameVal = ""
			}
		}
		if nameVal == "" {
			if login, ok := ghUser["login"]; ok && login != nil {
				nameVal = fmt.Sprintf("%v", login)
				if nameVal == "<nil>" {
					nameVal = ""
				}
			}
		}

		// If email is empty (private email setting on GitHub), fetch from /user/emails
		if emailVal == "" {
			emailsURL := userinfoURL + "/emails"
			if !strings.HasSuffix(userinfoURL, "/user") && !strings.HasSuffix(userinfoURL, "/userinfo") {
				if strings.HasSuffix(userinfoURL, "/") {
					emailsURL = userinfoURL + "emails"
				} else {
					emailsURL = userinfoURL + "/emails"
				}
			}
			emailsReq, _ := http.NewRequestWithContext(ctx, "GET", emailsURL, nil)
			emailsReq.Header.Set("Authorization", "Bearer "+accessToken)
			emailsReq.Header.Set("User-Agent", "DilithiAuth-Server")

			emailsResp, err := httpClient.Do(emailsReq)
			if err == nil {
				defer emailsResp.Body.Close()
				if emailsResp.StatusCode == http.StatusOK {
					emailsBody, _ := io.ReadAll(emailsResp.Body)
					var emailsList []map[string]interface{}
					if err := json.Unmarshal(emailsBody, &emailsList); err == nil {
						for _, em := range emailsList {
							eStr := fmt.Sprintf("%v", em["email"])
							if eStr != "" && eStr != "<nil>" {
								if v, ok := em["verified"].(bool); ok && !v {
									continue
								}
								primary, _ := em["primary"].(bool)
								if primary {
									emailVal = eStr
									break
								}
								if emailVal == "" {
									emailVal = eStr
								}
							}
						}
					}
				}
			}
		}

		return &SocialUser{
			ID:       idVal,
			Email:    emailVal,
			Name:     nameVal,
			Provider: ProviderGitHub,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported social provider: %s", provider)
	}
}

