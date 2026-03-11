package github

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Client handles GitHub App authentication — JWT generation and installation token caching.
type Client struct {
	appID      string
	clientID   string
	privateKey *rsa.PrivateKey
	httpClient *http.Client

	// In-memory installation token cache.
	mu     sync.RWMutex
	tokens map[int64]cachedToken
}

type cachedToken struct {
	Token     string
	ExpiresAt time.Time
}

// InstallationToken is the response from GitHub's installation token endpoint.
type InstallationToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewClient creates a GitHub App client. Returns nil if appID is empty (not configured).
func NewClient(appID, clientID, privateKeyPEM string) (*Client, error) {
	if appID == "" {
		return nil, nil
	}

	keyData := privateKeyPEM
	// Support file path instead of raw PEM.
	if !strings.HasPrefix(keyData, "-----BEGIN") {
		data, err := os.ReadFile(keyData)
		if err != nil {
			return nil, fmt.Errorf("github: read private key file: %w", err)
		}
		keyData = string(data)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyData))
	if err != nil {
		return nil, fmt.Errorf("github: parse private key: %w", err)
	}

	return &Client{
		appID:      appID,
		clientID:   clientID,
		privateKey: key,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		tokens:     make(map[int64]cachedToken),
	}, nil
}

// AppID returns the configured GitHub App ID.
func (c *Client) AppID() string {
	return c.appID
}

// ClientID returns the configured GitHub App client ID (used for install URL).
func (c *Client) ClientID() string {
	return c.clientID
}

// PrivateKey returns the parsed RSA private key (used for signing state JWTs).
func (c *Client) PrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

// GenerateJWT creates a GitHub App JWT for authenticating as the app.
// Valid for 10 minutes per GitHub's requirements.
func (c *Client) GenerateJWT() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    c.appID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(c.privateKey)
}

// InstallationTokenFor returns a valid installation token, using cache when possible.
// Refreshes the token 5 minutes before expiry.
func (c *Client) InstallationTokenFor(ctx context.Context, installationID int64) (*InstallationToken, error) {
	c.mu.RLock()
	cached, ok := c.tokens[installationID]
	c.mu.RUnlock()

	if ok && time.Until(cached.ExpiresAt) > 5*time.Minute {
		return &InstallationToken{Token: cached.Token, ExpiresAt: cached.ExpiresAt}, nil
	}

	return c.fetchInstallationToken(ctx, installationID)
}

func (c *Client) fetchInstallationToken(ctx context.Context, installationID int64) (*InstallationToken, error) {
	appJWT, err := c.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("github: generate jwt: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: fetch installation token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("github: read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("github: installation token request failed (%d): %s", resp.StatusCode, string(body))
	}

	var token InstallationToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("github: decode token response: %w", err)
	}

	// Cache the token.
	c.mu.Lock()
	c.tokens[installationID] = cachedToken{Token: token.Token, ExpiresAt: token.ExpiresAt}
	c.mu.Unlock()

	return &token, nil
}

// GetInstallation fetches installation details from GitHub.
func (c *Client) GetInstallation(ctx context.Context, installationID int64) (map[string]any, error) {
	appJWT, err := c.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("github: generate jwt: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d", installationID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: get installation: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("github: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: get installation failed (%d): %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("github: decode installation: %w", err)
	}

	return result, nil
}

// APIRequest makes an authenticated request to the GitHub API using an installation token.
func (c *Client) APIRequest(ctx context.Context, method, url, installToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+installToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return c.httpClient.Do(req)
}
