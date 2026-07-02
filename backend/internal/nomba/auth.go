package nomba

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const tokenCacheKey = "nomba:access_token"

// getToken returns a valid access token, fetching a new one if the cache is
// empty or expired. Concurrent callers collapse into a single Nomba call via
// singleflight — only one goroutine hits the network, the rest wait on it.
func (c *Client) getToken(ctx context.Context) (string, error) {
	if cached, err := c.redis.Get(ctx, tokenCacheKey).Result(); err == nil && cached != "" {
		return cached, nil
	}

	result, err, _ := c.sf.Do("token", func() (interface{}, error) {
		return c.fetchToken(ctx)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// fetchToken calls Nomba's OAuth endpoint directly and caches the result.
func (c *Client) fetchToken(ctx context.Context) (string, error) {
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/auth/token/issue", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accountId", c.accountID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("nomba token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nomba token request failed: status %d", resp.StatusCode)
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if tr.Data.AccessToken == "" {
		return "", fmt.Errorf("nomba token response missing access token")
	}

	// Cache with a safety margin so we never serve a token that's about to expire.
	ttl := time.Duration(tr.Data.ExpiresIn-60) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if err := c.redis.Set(ctx, tokenCacheKey, tr.Data.AccessToken, ttl).Err(); err != nil {
		return "", fmt.Errorf("cache token: %w", err)
	}

	return tr.Data.AccessToken, nil
}
