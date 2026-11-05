package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FahadAkash/factorypulse/backend/internal/api"
	"github.com/FahadAkash/factorypulse/backend/internal/config"
	"github.com/FahadAkash/factorypulse/backend/internal/telemetry"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := telemetry.InitLogger(cfg.LogLevel)
	logger.Info("FactoryPulse backend starting",
		"port", cfg.Port,
		"workers", cfg.WorkerPoolCap,
		"mqtt", cfg.MQTTURL,
		"redis", cfg.RedisAddr,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(ctx)

	// HTTP Server (Health & Monitoring API)
	httpServer := api.NewServer(cfg.Port, logger)
	g.Go(func() error {
		return httpServer.Start(gCtx)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("fatal server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("FactoryPulse backend cleanly stopped")
}

