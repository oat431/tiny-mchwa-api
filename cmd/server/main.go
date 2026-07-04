package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"oat431/tiny-mchawa-api/internal/app"
	"oat431/tiny-mchawa-api/internal/config"
)

func main() {
	cfg := config.Load()
	config.SetupLogger(cfg.LogLevel)

	a, err := app.New(cfg)
	if err != nil {
		slog.Error("app init failed", "error", err)
		os.Exit(1)
	}
	defer a.DB.Close()

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = a.Fiber.ShutdownWithContext(ctx)
	}()

	slog.Info("starting server", "port", cfg.Port)
	if err := a.Fiber.Listen(":" + cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
