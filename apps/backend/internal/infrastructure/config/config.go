// Package config provides application configuration loading from environment variables.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration values.
type Config struct {
	Port        string
	Environment string
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
}

// Load reads configuration from environment variables and returns a Config struct.
func Load() *Config {
	// Load .env file from project root
	_ = godotenv.Load("../../.env")

	return &Config{
		Port:        getEnv("PORT", "9000"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5433/mindhit?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6380"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
