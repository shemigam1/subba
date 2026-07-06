// Command worker is the async event-processing entrypoint. It connects to RabbitMQ,
// declares the full topic topology, and starts one consumer per queue with a
// dedicated worker pool. Idempotency is enforced per-consumer via the
// internal/idempotency package (Redis fast-check + Postgres INSERT ON CONFLICT).
//
// Queue → handler mapping:
//
//	subba.invoicing          → invoicing.NewHandler   (payment.succeeded, subscription.renew)
//	subba.subscription_state → substate.NewHandler    (payment.succeeded, subscription.renew)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/config"
	"github.com/shamigam1/subba/internal/idempotency"
	"github.com/shamigam1/subba/internal/invoicing"
	"github.com/shamigam1/subba/internal/nomba"
	"github.com/shamigam1/subba/internal/observability"
	"github.com/shamigam1/subba/internal/store"
	"github.com/shamigam1/subba/internal/substate"
)

// consumerCfg describes one queue consumer with its concurrency and retry settings.
type consumerCfg struct {
	queue    string
	prefetch int
	poolSize int
	retryQ   string
	handler  idempotency.HandlerFunc
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := observability.NewLogger(cfg.LogLevel, cfg.AppEnv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Postgres ──────────────────────────────────────────────────────────────
	// Workers use the admin pool (bypasses RLS) because processed_events is not
	// tenant-scoped. Handler DB work that IS tenant-scoped opens additional tx
	// via store.WithTenant inside the handler when needed.
	adminPool, err := store.NewPool(ctx, cfg.AdminDatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("worker: postgres connect error")
	}
	defer adminPool.Close()
	log.Info().Msg("worker: postgres connected")

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("worker: invalid redis url")
	}
	rdb := redis.NewClient(redisOpt)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("worker: redis ping failed")
	}
	log.Info().Msg("worker: redis connected")

	// ── Nomba client ──────────────────────────────────────────────────────────
	nombaClient := nomba.NewClient(nomba.Config{
		BaseURL:      cfg.NombaBaseURL,
		ClientID:     cfg.NombaClientID,
		ClientSecret: cfg.NombaClientSecret,
		AccountID:    cfg.NombaAccountID,
		Redis:        rdb,
	})

	// ── RabbitMQ ──────────────────────────────────────────────────────────────
	bc, err := broker.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("worker: rabbitmq connect error")
	}
	defer bc.Close()

	// Declare topology — safe to call on every startup (idempotent). The worker
	// owns this; the API process only publishes and does not declare topology.
	if err := broker.DeclareTopology(bc.Ch); err != nil {
		log.Fatal().Err(err).Msg("worker: topology declare error")
	}
	log.Info().Msg("worker: rabbitmq topology declared")

	// ── Idempotency wrappers (one per logical consumer) ───────────────────────
	invoicingIdem := idempotency.New(adminPool, rdb, "invoicing", log)
	subStateIdem  := idempotency.New(adminPool, rdb, "subscription_state", log)

	// ── Consumer registry ─────────────────────────────────────────────────────
	consumers := []consumerCfg{
		{
			queue:    broker.QueueInvoicing,
			prefetch: 10,
			poolSize: 5,
			retryQ:   broker.QueueInvoicingRetry,
			handler:  invoicing.NewHandler(log),
		},
		{
			queue:    broker.QueueSubscriptionState,
			prefetch: 10,
			poolSize: 5,
			retryQ:   broker.QueueSubscriptionStateRetry,
			handler:  substate.NewHandler(nombaClient, log),
		},
	}

	wrappers := map[string]*idempotency.Wrapper{
		broker.QueueInvoicing:         invoicingIdem,
		broker.QueueSubscriptionState: subStateIdem,
	}

	// ── Start one consumer loop per queue ─────────────────────────────────────
	for _, c := range consumers {
		idem := wrappers[c.queue]
		if err := startConsumer(ctx, bc, c, idem); err != nil {
			log.Fatal().Err(err).Str("queue", c.queue).Msg("worker: failed to start consumer")
		}
		log.Info().
			Str("queue", c.queue).
			Int("pool_size", c.poolSize).
			Int("prefetch", c.prefetch).
			Msg("worker: consumer started")
	}

	log.Info().Msg("worker: all consumers running — awaiting shutdown signal")
	<-ctx.Done()
	log.Info().Msg("worker: shutdown signal received — draining in-flight messages")

	// Give in-flight messages time to finish before deferred Closes run.
	time.Sleep(10 * time.Second)
	log.Info().Msg("worker: shutdown complete")
	os.Exit(0)
}

// startConsumer opens a dedicated AMQP channel per queue (isolated QoS),
// then spawns poolSize goroutines that drain the delivery channel.
func startConsumer(
	ctx context.Context,
	bc *broker.Conn,
	c consumerCfg,
	idem *idempotency.Wrapper,
) error {
	ch, err := bc.NewChannel()
	if err != nil {
		return fmt.Errorf("open channel for %s: %w", c.queue, err)
	}

	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return fmt.Errorf("set prefetch for %s: %w", c.queue, err)
	}

	deliveries, err := ch.Consume(
		c.queue,
		"",    // server-generated consumer tag
		false, // autoAck=false — we ack/nack manually
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume %s: %w", c.queue, err)
	}

	for i := 0; i < c.poolSize; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-deliveries:
					if !ok {
						return
					}
					processDelivery(ctx, ch, d, c, idem)
				}
			}
		}()
	}
	return nil
}

// processDelivery deserialises one AMQP message, runs the idempotency wrapper,
// and acks on success or publishes to the retry queue on transient failure.
func processDelivery(
	ctx context.Context,
	ch *amqp.Channel,
	d amqp.Delivery,
	c consumerCfg,
	idem *idempotency.Wrapper,
) {
	var event broker.Event
	if err := json.Unmarshal(d.Body, &event); err != nil {
		// Undecodable message can never succeed; nack without requeue → DLQ.
		_ = d.Nack(false, false)
		return
	}

	if err := idem.Run(ctx, event, c.handler); err != nil {
		if pubErr := publishToRetry(ctx, ch, c.retryQ, d.Body); pubErr != nil {
			_ = d.Nack(false, false) // retry publish failed → DLQ
		} else {
			_ = d.Ack(false) // acked; copy is in retry queue
		}
		return
	}

	_ = d.Ack(false)
}

// publishToRetry sends the raw body to the named retry queue via the default
// exchange. The retry queue's x-message-ttl re-routes it back after the TTL.
func publishToRetry(ctx context.Context, ch *amqp.Channel, retryQueue string, body []byte) error {
	return ch.PublishWithContext(ctx,
		"",         // default exchange
		retryQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
