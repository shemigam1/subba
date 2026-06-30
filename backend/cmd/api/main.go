// Command api is the synchronous HTTP entrypoint. In Phase 0 it stands up the
// process: config, logging, dependency wiring, and the liveness/readiness probes.
// The webhook receiver, tenant API, and portal API are added in later phases.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/shamigam1/subba/internal/config"
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

	if cfg.AppEnv != "development" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())

	// Liveness: the process is up. Cheap and dependency-free.
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Readiness: every dependency is reachable. Flips to 503 on dependency loss.
	router.GET("/readyz", func(c *gin.Context) {
		checkCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := plat.Ready(checkCtx); err != nil {
			log.Warn().Err(err).Msg("readiness check failed")
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: router}

	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("api listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server failed")
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
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
