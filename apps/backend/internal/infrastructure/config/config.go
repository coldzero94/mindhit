// Package config provides application configuration loading from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration values.
type Config struct {
	Port              string
	Environment       string
	DatabaseURL       string
	JWTSecret         string
	RedisURL          string
	RedisAddr         string
	WorkerConcurrency int
}

// Load reads configuration from environment variables and returns a Config struct.
func Load() *Config {
	// Load .env file from project root
	_ = godotenv.Load("../../.env")

	return &Config{
		Port:              getEnv("PORT", "9000"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5433/mindhit?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6380"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6380"),
		WorkerConcurrency: getEnvInt("WORKER_CONCURRENCY", 10),
	}
}

// Validate checks if critical configuration values are properly set for production.
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.JWTSecret == "your-secret-key" || len(c.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be set to a secure value (at least 32 characters) in production")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
