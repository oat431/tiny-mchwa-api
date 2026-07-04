package config

import (
	"os"
)

type Config struct {
	Port      string
	DBUrl     string
	LogLevel  string
}

func Load() Config {
	return Config{
		Port:      getEnv("PORT", "3000"),
		DBUrl:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tiny_mchwa?sslmode=disable"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
