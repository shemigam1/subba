// Command api is the synchronous HTTP entrypoint: liveness/readiness probes plus the
// tenant dashboard and customer portal APIs. The Nomba webhook receiver (Track A) mounts
// into the same server later.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/config"
	httpapi "github.com/shamigam1/subba/internal/http"
	"github.com/shamigam1/subba/internal/nomba"
	"github.com/shamigam1/subba/internal/observability"
	"github.com/shamigam1/subba/internal/platform"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := observability.NewLogger(cfg.LogLevel, cfg.AppEnv)

	ctx := context.Background()

	// =================================Connect to RabbitMQ================================================
	bc, err := broker.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("rabbitmq connect error")
	}
	defer bc.Close()

	// Declare topology — api owns this, workers/scheduler assume it exists
	if err := broker.DeclareTopology(bc.Ch); err != nil {
		log.Fatal().Err(err).Msg("topology declare error")
	}
	log.Info().Msg("rabbitmq topology declared successfully")

	// =================================Connect to Redis====================================
	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid redis url")
	}
	rdb := redis.NewClient(redisOpt)
	defer rdb.Close()

	nombaClient := nomba.NewClient(nomba.Config{
		BaseURL:      cfg.NombaBaseURL,
		ClientID:     cfg.NombaClientID,
		ClientSecret: cfg.NombaClientSecret,
		Redis:        rdb,
	})
	_ = nombaClient // TODO: inject into webhook handler once built

	plat, err := platform.New(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect dependencies")
	}
	defer plat.Close()
	log.Info().Msg("dependencies connected")

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.NewRouter(cfg, log, plat, bc.Ch)}

	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("api listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info().Msg("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}
