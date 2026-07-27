package dilithiauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config represents DilithiAuth Client SDK configuration.
type Config struct {
	Domain       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	HTTPClient   *http.Client
}

// Client is the DilithiAuth SDK client instance.
type Client struct {
	domain       string
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

// TokenResponse represents OAuth 2.0 token response payload.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token,omitempty"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	PQCRef      string `json:"pqc_ref,omitempty"`
	PQCAlg      string `json:"pqc_alg,omitempty"`
}

// Claims represents JWT user claims payload.
type Claims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	PQCRef    string `json:"pqc_ref,omitempty"`
	PQCAlg    string `json:"pqc_alg,omitempty"`
}

// DefaultDomain is the official live DilithiAuth IdP issuer domain.
const DefaultDomain = "https://dilithiauthapis.gradientgeeks.tech"

// NewClient initializes a new DilithiAuth SDK Client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.ClientID == "" {
		return nil, errors.New("dilithiauth: ClientID is required")
	}

	domain := cfg.Domain
	if domain == "" {
		domain = DefaultDomain
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	domain = strings.TrimRight(domain, "/")
	return &Client{
		domain:       domain,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURI:  cfg.RedirectURI,
		httpClient:   httpClient,
	}, nil
}

// GeneratePKCEPair generates a high-entropy PKCE code_verifier and S256 code_challenge.
func GeneratePKCEPair() (verifier string, challenge string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(bytes)

	h := sha256.New()
	h.Write([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return verifier, challenge, nil
}

// GetAuthorizationURL constructs PKCE authorization redirect URL.
func (c *Client) GetAuthorizationURL(redirectURI, state, codeChallenge string) string {
	if redirectURI == "" {
		redirectURI = c.redirectURI
	}
	u, _ := url.Parse(c.domain + "/authorize")
	q := u.Query()
	q.Set("client_id", c.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()
	return u.String()
}

// ExchangeCode exchanges authorization code for Post-Quantum access token.
func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI, codeVerifier string) (*TokenResponse, error) {
	if redirectURI == "" {
		redirectURI = c.redirectURI
	}

	endpoint := c.domain + "/api/v1/oauth/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.clientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if c.clientSecret != "" {
		req.SetBasicAuth(c.clientID, c.clientSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokResp TokenResponse
	if err := json.Unmarshal(body, &tokResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokResp, nil
}

// ParseClaims decodes and returns claims from JWT token string without verification.
func ParseClaims(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT format")
	}

	payloadSegment := parts[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSegment)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}

// VerifyToken decodes claims and verifies token expiration.
func (c *Client) VerifyToken(ctx context.Context, tokenStr string) (*Claims, error) {
	claims, err := ParseClaims(tokenStr)
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}
