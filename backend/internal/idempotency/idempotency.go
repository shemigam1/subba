// Package idempotency provides an at-most-once execution wrapper for message
// consumers. It uses a two-layer deduplication strategy:
//
//  1. Redis fast-check (O(1), sub-millisecond): a SET NX with a 24-hour TTL
//     eliminates redeliveries before they touch Postgres.  The key is
//     "{requestID}:{consumer}".
//
//  2. Postgres transaction anchor: when Redis has no record, we open a
//     transaction and INSERT INTO processed_events … ON CONFLICT DO NOTHING.
//     Only the goroutine whose INSERT returns a row continues to invoke the
//     handler; all others see pgx.ErrNoRows and stop (duplicate delivery).
//
// The handler signature is:
//
//	func(ctx context.Context, tx pgx.Tx, event broker.Event) error
//
// The handler runs inside the same transaction as the INSERT, so handler work
// and idempotency record are committed atomically.  After a successful commit
// the Redis key is set so future redeliveries skip Postgres entirely.
package idempotency

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/store"
	"github.com/shamigam1/subba/internal/store/db"
)

// HandlerFunc is the signature every consumer handler must satisfy.
// tx is the live Postgres transaction opened by the wrapper — handlers MAY
// issue additional SQL on it; the wrapper commits or rolls back after the call.
type HandlerFunc func(ctx context.Context, tx pgx.Tx, event broker.Event) error

// Wrapper holds the shared dependencies of the idempotency layer.
type Wrapper struct {
	pool     *pgxpool.Pool
	rdb      *redis.Client
	consumer string        // logical name, e.g. "invoicing", "payouts"
	ttl      time.Duration // Redis key TTL; defaults to 24 h
	log      zerolog.Logger
}

// New creates a Wrapper for the named consumer.
func New(pool *pgxpool.Pool, rdb *redis.Client, consumer string, log zerolog.Logger) *Wrapper {
	return &Wrapper{
		pool:     pool,
		rdb:      rdb,
		consumer: consumer,
		ttl:      24 * time.Hour,
		log:      log.With().Str("idempotency_consumer", consumer).Logger(),
	}
}

// redisKey returns the Redis key for a given request ID.
func (w *Wrapper) redisKey(requestID string) string {
	return fmt.Sprintf("idem:%s:%s", requestID, w.consumer)
}

// Run is the main entry point. It executes the handler if and only if the
// (requestID, consumer) pair has not been processed before.
//
// Return contract:
//   - nil → event processed (or was already processed — both are success from
//     the caller's perspective; the message should be acked).
//   - non-nil → transient error; the caller should nack and requeue.
func (w *Wrapper) Run(ctx context.Context, event broker.Event, h HandlerFunc) error {
	requestID := event.RequestID
	if requestID == "" {
		// Events without a request ID cannot be deduplicated safely; let them
		// through and log loudly so the webhook layer is fixed.
		w.log.Warn().Str("event_type", event.EventType).Msg("idempotency: empty request_id, bypassing dedup")
		return w.runHandler(ctx, event, h)
	}

	// ── 1. Redis fast-check ────────────────────────────────────────────────
	key := w.redisKey(requestID)
	exists, err := w.rdb.Exists(ctx, key).Result()
	if err != nil {
		// Redis unavailable — fall through to Postgres so we don't drop events.
		w.log.Warn().Err(err).Str("request_id", requestID).Msg("idempotency: redis unavailable, falling back to postgres")
	} else if exists > 0 {
		w.log.Debug().Str("request_id", requestID).Msg("idempotency: redis hit — already processed")
		return nil
	}

	// ── 2. Postgres transaction + INSERT ON CONFLICT DO NOTHING ───────────
	var handlerErr error
	txErr := store.WithWorker(ctx, w.pool, func(tx pgx.Tx) error {
		q := db.New(tx)

		_, insertErr := q.InsertProcessedEvent(ctx, db.InsertProcessedEventParams{
			RequestID: requestID,
			Consumer:  w.consumer,
		})
		if insertErr != nil {
			if errors.Is(insertErr, pgx.ErrNoRows) {
				// Another worker already claimed this event — skip.
				w.log.Debug().Str("request_id", requestID).Msg("idempotency: postgres conflict — already claimed")
				return nil
			}
			return fmt.Errorf("insert processed_event: %w", insertErr)
		}

		// INSERT won the race — invoke the handler inside the same tx.
		if herr := h(ctx, tx, event); herr != nil {
			// Mark as failed before rolling back so we can observe it in
			// Postgres (the MarkFailed runs on a new connection after rollback).
			handlerErr = herr
			return herr // triggers rollback via WithWorker's defer
		}

		// Mark done inside the tx — committed atomically with the handler work.
		if merr := q.MarkProcessedEventDone(ctx, db.MarkProcessedEventDoneParams{
			RequestID: requestID,
			Consumer:  w.consumer,
		}); merr != nil {
			return fmt.Errorf("mark done: %w", merr)
		}
		return nil
	})

	if txErr != nil {
		if handlerErr != nil {
			// Best-effort: record failure on a fresh connection so the slot is
			// not left stuck in 'processing'.
			go w.markFailed(requestID)
		}
		return fmt.Errorf("idempotency tx: %w", txErr)
	}

	// ── 3. Backfill Redis so future redeliveries skip Postgres ────────────
	if setErr := w.rdb.Set(ctx, key, "1", w.ttl).Err(); setErr != nil {
		// Non-fatal — next delivery will fall through to Postgres and see 'done'.
		w.log.Warn().Err(setErr).Str("request_id", requestID).Msg("idempotency: failed to backfill redis")
	}

	return nil
}

// runHandler executes the handler without any idempotency guard (used when
// request_id is absent).
func (w *Wrapper) runHandler(ctx context.Context, event broker.Event, h HandlerFunc) error {
	return store.WithWorker(ctx, w.pool, func(tx pgx.Tx) error {
		return h(ctx, tx, event)
	})
}

// markFailed writes a 'failed' record on a fresh connection after a rollback.
func (w *Wrapper) markFailed(requestID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := w.pool.Acquire(ctx)
	if err != nil {
		w.log.Error().Err(err).Str("request_id", requestID).Msg("idempotency: could not acquire conn to mark failed")
		return
	}
	defer conn.Release()

	q := db.New(conn)
	if err := q.MarkProcessedEventFailed(ctx, db.MarkProcessedEventFailedParams{
		RequestID: requestID,
		Consumer:  w.consumer,
	}); err != nil {
		w.log.Error().Err(err).Str("request_id", requestID).Msg("idempotency: failed to mark event as failed")
	}
}
