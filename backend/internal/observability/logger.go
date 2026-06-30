// Package observability holds logging, metrics, and health wiring. Phase 0 ships
// structured JSON logging with a requestId correlation field; metrics and tracing
// arrive in Phase 5.
package observability

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// CorrelationKey is the log field carrying the Nomba requestId across the webhook
// receiver and all three fanout consumers, so one payment event is traceable end
// to end. Without it, the consumers produce disconnected log streams.
const CorrelationKey = "request_id"

// NewLogger builds a JSON logger. In development it falls back to a console writer
// for readability; everywhere else it emits structured JSON for querying.
func NewLogger(level, env string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var logger zerolog.Logger
	if env == "development" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Kitchen})
	} else {
		logger = zerolog.New(os.Stderr)
	}
	return logger.Level(lvl).With().Timestamp().Str("service", "subba").Logger()
}
