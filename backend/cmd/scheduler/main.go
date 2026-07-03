// Command scheduler is a cron-driven process that sweeps for active
// subscriptions whose billing period has elapsed and publishes a
// subscription.renew event to RabbitMQ for each one.
//
// The subscription_state worker picks these events up and charges the
// customer's tokenized card via Nomba.
//
// The scheduler is stateless — it only reads Postgres and publishes to
// RabbitMQ. All idempotency is enforced downstream by the worker.
//
// # Schedule
//
// Default: every 1 minute. Override with SCHEDULER_INTERVAL env var
// (e.g. "30s", "5m").
package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/config"
	"github.com/shamigam1/subba/internal/observability"
	"github.com/shamigam1/subba/internal/store"
	"github.com/shamigam1/subba/internal/store/db"
	"github.com/shamigam1/subba/internal/webhook"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := observability.NewLogger(cfg.LogLevel, cfg.AppEnv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Postgres (admin pool — reads are not tenant-scoped) ───────────────────
	adminPool, err := store.NewPool(ctx, cfg.AdminDatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("scheduler: postgres connect error")
	}
	defer adminPool.Close()
	log.Info().Msg("scheduler: postgres connected")

	// ── RabbitMQ ──────────────────────────────────────────────────────────────
	bc, err := broker.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("scheduler: rabbitmq connect error")
	}
	defer bc.Close()
	log.Info().Msg("scheduler: rabbitmq connected")

	// ── Tick interval ─────────────────────────────────────────────────────────
	interval := time.Minute
	if raw := os.Getenv("SCHEDULER_INTERVAL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			interval = d
		} else {
			log.Warn().Str("value", raw).Msg("scheduler: invalid SCHEDULER_INTERVAL, using 1m")
		}
	}
	log.Info().Dur("interval", interval).Msg("scheduler: starting sweep loop")

	q := db.New(adminPool)

	// Run once immediately on startup, then on each tick.
	sweep(ctx, q, bc.Ch, log)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("scheduler: shutdown signal received — exiting")
			return
		case <-ticker.C:
			sweep(ctx, q, bc.Ch, log)
		}
	}
}

// sweep queries for all active subscriptions whose billing period has elapsed
// and publishes a subscription.renew event for each one. Individual failures
// are logged and skipped; the next tick will catch anything missed.
func sweep(ctx context.Context, q *db.Queries, ch *amqp.Channel, log zerolog.Logger) {
	subs, err := q.ListDueSubscriptions(ctx)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: list due subscriptions failed")
		return
	}

	if len(subs) == 0 {
		log.Debug().Msg("scheduler: sweep complete — no due subscriptions")
		return
	}

	log.Info().Int("count", len(subs)).Msg("scheduler: publishing renewal events")

	for _, sub := range subs {
		// requestID is scoped to the minute — retrying within the same minute
		// hits the idempotency barrier in the worker and is a no-op.
		requestID := "sched-" + sub.ID.String() + "-" + time.Now().UTC().Format("2006-01-02T15:04")

		event := webhook.ToRenewalBrokerEvent(
			sub.TenantID.String(),
			sub.CustomerID.String(),
			sub.ID.String(),
			sub.PlanID.String(),
			requestID,
		)

		body, err := json.Marshal(event)
		if err != nil {
			log.Error().Err(err).
				Str("subscription_id", sub.ID.String()).
				Msg("scheduler: marshal renewal event failed — skipping")
			continue
		}

		err = ch.PublishWithContext(ctx,
			broker.ExchangeName,
			broker.RoutingKeySubscriptionRenew,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         body,
				MessageId:    requestID,
			},
		)
		if err != nil {
			log.Error().Err(err).
				Str("subscription_id", sub.ID.String()).
				Msg("scheduler: publish failed — will retry next tick")
			continue
		}

		log.Info().
			Str("subscription_id", sub.ID.String()).
			Str("tenant_id", sub.TenantID.String()).
			Str("customer_id", sub.CustomerID.String()).
			Str("request_id", requestID).
			Msg("scheduler: subscription.renew published")
	}
}
