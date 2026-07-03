package nomba

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Client is the shared Nomba API client. One instance per process, injected
// into handlers and workers that need to talk to Nomba.
type Client struct {
	httpClient   *http.Client
	redis        *redis.Client
	baseURL      string
	clientID     string
	clientSecret string
	accountID    string
	sf           singleflight.Group // dedupes concurrent token refreshes
}

// Config holds the values needed to construct a Client.
type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	AccountID    string
	Redis        *redis.Client
}

// NewClient builds a Nomba API client from the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient:   &http.Client{},
		redis:        cfg.Redis,
		baseURL:      cfg.BaseURL,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		accountID:    cfg.AccountID,
	}
}
