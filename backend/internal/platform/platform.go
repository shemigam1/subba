// Package platform wires the external dependencies (Postgres, Redis, RabbitMQ) and
// exposes the readiness checks behind /readyz. Readiness flipping to false on a
// dependency loss is the graceful-degradation signal the engine demonstrates live.
package platform

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/shamigam1/subba/internal/config"
	"github.com/shamigam1/subba/internal/store"
)

// Platform holds the long-lived dependency handles shared across the process.
type Platform struct {
	DB      *pgxpool.Pool // tenant-scoped (subba_app role, RLS enforced)
	AdminDB *pgxpool.Pool // privileged (bypasses RLS) for signup/auth/routing
	Redis   *redis.Client
	AMQP    *amqp.Connection
}

// New connects every dependency, failing fast if any is unreachable.
func New(ctx context.Context, cfg *config.Config) (*Platform, error) {
	db, err := store.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("tenant db: %w", err)
	}
	adminDB, err := store.NewPool(ctx, cfg.AdminDatabaseURL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("admin db: %w", err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		db.Close()
		adminDB.Close()
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(redisOpts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		db.Close()
		adminDB.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	amqpConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		db.Close()
		adminDB.Close()
		_ = rdb.Close()
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	return &Platform{DB: db, AdminDB: adminDB, Redis: rdb, AMQP: amqpConn}, nil
}

// Ready checks live connectivity to every dependency. Returns the first failure so
// /readyz can report exactly which dependency is down.
func (p *Platform) Ready(ctx context.Context) error {
	if err := p.DB.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	if err := p.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	if p.AMQP == nil || p.AMQP.IsClosed() {
		return errors.New("rabbitmq: connection closed")
	}
	return nil
}

// Close releases all dependency handles.
func (p *Platform) Close() {
	if p.AMQP != nil {
		_ = p.AMQP.Close()
	}
	if p.Redis != nil {
		_ = p.Redis.Close()
	}
	if p.AdminDB != nil {
		p.AdminDB.Close()
	}
	if p.DB != nil {
		p.DB.Close()
	}
}
