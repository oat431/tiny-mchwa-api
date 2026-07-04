package app

import (
	"oat431/tiny-mchawa-api/internal/config"
	"oat431/tiny-mchawa-api/internal/database"
	"oat431/tiny-mchawa-api/internal/handler"
	mw "oat431/tiny-mchawa-api/internal/middleware"
	"oat431/tiny-mchawa-api/internal/repository"
	"oat431/tiny-mchawa-api/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
)

type App struct {
	Fiber *fiber.App
	DB    *sqlx.DB
}

func New(cfg config.Config) (*App, error) {
	// DB
	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		return nil, err
	}

	// Wire layers
	repo := repository.NewTodolistRepository(db)
	svc := service.NewTodolistService(repo)
	h := handler.NewTodolistHandler(svc)

	// Fiber
	f := fiber.New(fiber.Config{
		BodyLimit:      10 * 1024 * 1024,
		StructValidator: mw.NewValidator(),
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Middleware
	mw.Setup(f)

	// Health
	f.Get("/health", func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})

	// Routes
	v1 := f.Group("/api/v1")
	v1.Post("/todolists", h.Create)
	v1.Get("/todolists", h.List)
	v1.Get("/todolists/:id", h.GetByID)
	v1.Put("/todolists/:id", h.Update)
	v1.Delete("/todolists/:id", h.Delete)

	// 404
	f.Use(func(c fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	})

	return &App{Fiber: f, DB: db}, nil
}
