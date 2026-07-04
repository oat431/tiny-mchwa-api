package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func setupTestApp() *fiber.App {
	app := fiber.New()
	app.Get("/health", func(c fiber.Ctx) error { return c.SendStatus(200) })
	// ponytail: no DB needed — just testing 404 and health routes
	app.Use(func(c fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	})
	return app
}

func TestHealthEndpoint(t *testing.T) {
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestNotFoundRoute(t *testing.T) {
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(body, &result)
	if result["error"] != "not found" {
		t.Fatalf("expected 'not found' error, got %v", result["error"])
	}
}
