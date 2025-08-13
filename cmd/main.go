package main

import (
	"context"
	"github.com/getsentry/sentry-go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"screenshorter/config"
	"screenshorter/internal/handlers"
	"screenshorter/internal/service"
	"screenshorter/pkg/httpserver"
	"screenshorter/pkg/logger"
	"syscall"
	"time"
)

func main() {

	// Config
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("configuration read error: %v", err)
	}

	// Logger
	lgr := logger.NewLogger(cfg)

	lgr.Info().
		Str("version", config.Version).
		Msg("Starting Screenshorter")

	// Sentry
	if cfg.LogTarget == "sentry" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SentryDsn,
			Environment: "production",
			//Debug:       true,
			Release: "Screenshoter@" + config.Version,
		})
		if err != nil {
			lgr.Fatal().Err(err).Msgf("sentry.Init %s", err.Error())
		}
		lgr.Info().Msg("Sentry initialized")
		defer sentry.Flush(2 * time.Second)
	}

	screenshoter, err := service.NewPlaywrite()
	if err != nil {
		lgr.Fatal().Err(err).Msgf("Failed to initialize Playwright")
	}
	s := service.NewService(screenshoter)
	h := handlers.NewHandler(s, cfg)
	srv := httpserver.NewServer()

	serverErr := make(chan error, 1)
	go func() {
		lgr.Info().Str("port", cfg.Port).Msg("Starting HTTP server")
		if err := srv.Run(cfg.Port, h.InitRoutes()); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case err := <-serverErr:
		lgr.Error().Err(err).Msg("Server runtime error")
	case sig := <-quit:
		lgr.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	lgr.Info().Msg("Shutting down server...")
	if err := srv.Stop(ctx); err != nil {
		lgr.Error().Err(err).Msg("Server shutdown error")
	}

	lgr.Info().Msg("Server stopped gracefully")
}
