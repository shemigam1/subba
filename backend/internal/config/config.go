// Package config loads runtime configuration from the environment. A local .env
// is loaded if present; real environments inject these as process env vars.
package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	LogLevel string
	HTTPAddr string

	// DatabaseURL connects as the non-superuser `subba_app` role, so RLS is
	// enforced. All tenant-scoped data access uses this pool.
	DatabaseURL string
	// AdminDatabaseURL connects as a superuser (bypasses RLS) for migrations and
	// tenant-agnostic work: signup, auth, webhook→tenant routing.
	AdminDatabaseURL string

	RedisURL    string
	RabbitMQURL string

	// PortalBaseURL is the public origin of the hosted customer portal; magic-link
	// access URLs are built from it.
	PortalBaseURL string
	// DashboardBaseURL is the public origin of the tenant dashboard, allowed as a
	// CORS origin alongside the portal.
	DashboardBaseURL string

	// MasterEncryptionKey (32 bytes, decoded from base64) encrypts tenant secrets
	// such as the Nomba client secret before they are stored.
	MasterEncryptionKey []byte

	// Resend transactional email (magic links, receipts, dunning).
	ResendAPIKey   string
	EmailFromName  string // display name, e.g. "Subba"
	EmailFromEmail string // verified sender, e.g. "notify@mail.subba.app"

	NombaBaseURL       string
	NombaClientID      string
	NombaClientSecret  string
	NombaAccountID     string // parent account ID, sent in the accountId header
	NombaSubAccountID  string // sub-account that scopes calls (revenue split target)
	NombaWebhookSecret string
}

// Load reads configuration, applying defaults suited to local development.
func Load() (*Config, error) {
	// Best-effort: ignore a missing .env so prod (env-injected) still works.
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:             getenv("APP_ENV", "development"),
		LogLevel:           getenv("LOG_LEVEL", "info"),
		HTTPAddr:           getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		AdminDatabaseURL:   os.Getenv("ADMIN_DATABASE_URL"),
		RedisURL:           getenv("REDIS_URL", "redis://localhost:6379/0"),
		RabbitMQURL:        getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		PortalBaseURL:      getenv("PORTAL_BASE_URL", "http://localhost:3000/pay"),
		DashboardBaseURL:   os.Getenv("DASHBOARD_BASE_URL"),
		ResendAPIKey:       os.Getenv("RESEND_API_KEY"),
		EmailFromName:      getenv("EMAIL_FROM_NAME", "Subba"),
		EmailFromEmail:     os.Getenv("EMAIL_FROM_EMAIL"),
		NombaBaseURL:       getenv("NOMBA_BASE_URL", "https://api.nomba.com"),
		NombaClientID:      os.Getenv("NOMBA_CLIENT_ID"),
		NombaClientSecret:  os.Getenv("NOMBA_CLIENT_SECRET"),
		NombaAccountID:     getenv("NOMBA_ACCOUNT_ID", "f666ef9b-888e-4799-85ce-acb505b28023"),
		NombaSubAccountID:  os.Getenv("NOMBA_SUBACCOUNT_ID"),
		NombaWebhookSecret: os.Getenv("NOMBA_WEBHOOK_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.AdminDatabaseURL == "" {
		return nil, fmt.Errorf("ADMIN_DATABASE_URL is required")
	}

	rawKey := os.Getenv("MASTER_ENCRYPTION_KEY")
	if rawKey == "" {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(rawKey)
	if err != nil {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY must be base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY must decode to 32 bytes, got %d", len(key))
	}
	cfg.MasterEncryptionKey = key

	if cfg.NombaClientID != "" && strings.HasPrefix(cfg.NombaClientID, "706df6c") {
		// If using TEST credentials, rigidly enforce sandbox URL to avoid 401/403.
		if cfg.NombaBaseURL == "https://api.nomba.com" {
			cfg.NombaBaseURL = "https://sandbox.nomba.com"
		} else if cfg.NombaBaseURL != "https://sandbox.nomba.com" {
			return nil, fmt.Errorf("NOMBA_BASE_URL must be https://sandbox.nomba.com when using TEST credentials")
		}
	} else if cfg.NombaClientID != "" && !strings.HasPrefix(cfg.NombaClientID, "706df6c") {
		return nil, fmt.Errorf("NOMBA_CLIENT_ID must be a TEST credential (starting with 706df6c) to prevent accidental real money movement")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
