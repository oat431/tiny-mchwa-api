package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	DBUrl    string
	LogLevel string
}

func Load() Config {
	// ponytail: ignore error — .env is optional, env vars take precedence
	_ = godotenv.Load()

	return Config{
		Port:     getEnv("PORT", "8005"),
		DBUrl:    getEnv("DATABASE_URL", "postgres://postgres:***@localhost:5432/tiny_mchwa?sslmode=disable"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func SetupLogger(level string) {
	lvl := slog.LevelInfo
	if level == "debug" {
		lvl = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
