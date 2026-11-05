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
	"github.com/FahadAkash/factorypulse/backend/internal/incident"
	"github.com/FahadAkash/factorypulse/backend/internal/telemetry"
	"github.com/FahadAkash/factorypulse/backend/internal/ws"
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

	// WebSocket Hub
	hub := ws.NewHub(logger)
	g.Go(func() error {
		hub.Run(gCtx)
		return nil
	})

	// Incident service
	incidentSvc := incident.NewService(nil)

	// HTTP Server (REST + WebSocket)
	httpServer := api.NewServer(cfg.Port, hub, incidentSvc, logger)
	g.Go(func() error {
		return httpServer.Start(gCtx)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("fatal server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("FactoryPulse backend cleanly stopped")
}
