// Package config loads runtime configuration from the environment. A local .env
// is loaded if present; real environments inject these as process env vars.
package config

import (
	"fmt"
	"os"

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

	NombaBaseURL       string
	NombaClientID      string
	NombaClientSecret  string
	NombaAccountID     string
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
		NombaBaseURL:       getenv("NOMBA_BASE_URL", "https://api.nomba.com"),
		NombaClientID:      os.Getenv("NOMBA_CLIENT_ID"),
		NombaClientSecret:  os.Getenv("NOMBA_CLIENT_SECRET"),
		NombaAccountID:     os.Getenv("NOMBA_ACCOUNT_ID"),
		NombaWebhookSecret: os.Getenv("NOMBA_WEBHOOK_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.AdminDatabaseURL == "" {
		return nil, fmt.Errorf("ADMIN_DATABASE_URL is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
