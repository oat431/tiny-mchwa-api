package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"oat431/tiny-mchawa-api/internal/config"
	"oat431/tiny-mchawa-api/internal/database"
	"oat431/tiny-mchawa-api/internal/handler"
	mw "oat431/tiny-mchawa-api/internal/middleware"
	"oat431/tiny-mchawa-api/internal/repository"
	"oat431/tiny-mchawa-api/internal/service"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg := config.Load()

	// Logger
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	// DB
	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Wire layers
	repo := repository.NewTodolistRepository(db)
	svc := service.NewTodolistService(repo)
	h := handler.NewTodolistHandler(svc)

	// Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit:             10 * 1024 * 1024,

		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Middleware
	mw.Setup(app)

	// Health
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})

	// Routes
	v1 := app.Group("/api/v1")
	v1.Post("/todolists", h.Create)
	v1.Get("/todolists", h.List)
	v1.Get("/todolists/:id", h.GetByID)
	v1.Put("/todolists/:id", h.Update)
	v1.Delete("/todolists/:id", h.Delete)

	// 404
	app.Use(func(c fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	})

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
	}()

	slog.Info("starting server", "port", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
