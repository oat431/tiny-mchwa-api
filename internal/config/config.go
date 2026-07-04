package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Port     string
	DBUrl    string
	LogLevel string
}

func Load() Config {
	return Config{
		Port:     getEnv("PORT", "3000"),
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
