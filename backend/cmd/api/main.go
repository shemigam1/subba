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

	"github.com/shamigam1/subba/internal/config"
	httpapi "github.com/shamigam1/subba/internal/http"
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
	plat, err := platform.New(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect dependencies")
	}
	defer plat.Close()
	log.Info().Msg("dependencies connected")

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.NewRouter(cfg, log, plat)}

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
