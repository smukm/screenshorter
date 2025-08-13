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
	lgr.Info().Msgf("Screenshorter %s", config.Version)

	// Sentry
	if cfg.LogTarget == "sentry" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SentryDsn,
			Environment: "production",
			//Debug:       true,
			Release: "Screenshorter@" + config.Version,
		})
		if err != nil {
			lgr.Fatal().Err(err).Msgf("sentry.Init %s", err.Error())
		}
		lgr.Info().Msg("Sentry initialized")
		defer sentry.Flush(2 * time.Second)
	}

	screenshorter, err := service.NewPlaywrite()
	if err != nil {
		lgr.Fatal().Err(err).Msgf("NewPlaywrite")
	}
	s := service.NewService(screenshorter)
	h := handlers.NewHandler(s, cfg)
	srv := httpserver.NewServer()
	go func() {
		if err := srv.Run("8033", h.InitRoutes()); err != nil && err != http.ErrServerClosed {
			lgr.Fatal().Err(err).Msgf("http server run error %s", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	lgr.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		lgr.Error().Err(err).Msg("Server forced to shutdown")
	}
}
