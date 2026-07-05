package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func Setup(app *fiber.App) {
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // ponytail: tighten to Flowero Gate origin in production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	// Logger — skip health check
	app.Use(func(c fiber.Ctx) error {
		if c.Path() == "/health" {
			return c.Next()
		}
		slog.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"requestID", c.Locals("requestid"),
		)
		return c.Next()
	})
}
