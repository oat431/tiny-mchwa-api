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
	todoRepo := repository.NewTodolistRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	todoSvc := service.NewTodolistService(todoRepo)
	taskSvc := service.NewTaskService(taskRepo, todoRepo)
	todoH := handler.NewTodolistHandler(todoSvc)
	taskH := handler.NewTaskHandler(taskSvc)

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

	// Todolist routes
	v1 := f.Group("/api/v1")
	v1.Post("/todolists", todoH.Create)
	v1.Get("/todolists", todoH.List)
	v1.Get("/todolists/:id", todoH.GetByID)
	v1.Put("/todolists/:id", todoH.Update)
	v1.Delete("/todolists/:id", todoH.Delete)

	// Task routes (nested under todolists)
	v1.Post("/todolists/:todolistId/tasks", taskH.Create)
	v1.Get("/todolists/:todolistId/tasks", taskH.List)

	// Task routes (direct access)
	v1.Get("/tasks/:id", taskH.GetByID)
	v1.Put("/tasks/:id", taskH.Update)
	v1.Delete("/tasks/:id", taskH.Delete)

	// 404
	f.Use(func(c fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	})

	return &App{Fiber: f, DB: db}, nil
}
