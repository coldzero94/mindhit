// Package main is the entry point for the MindHit API server.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/controller"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/infrastructure/config"
	"github.com/mindhit/api/internal/infrastructure/logger"
	"github.com/mindhit/api/internal/infrastructure/middleware"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Initialize logger based on environment
	logger.Init(cfg.Environment)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database connection with connection pool settings
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	defer func() {
		if err := client.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	// Auto-migrate schema in development
	if cfg.Environment != "production" {
		if err := client.Schema.Create(context.Background()); err != nil {
			slog.Error("failed to create schema", "error", err)
		}
	}

	// Queue client for async job processing
	queueClient := queue.NewClient(cfg.RedisAddr)
	defer func() {
		if err := queueClient.Close(); err != nil {
			slog.Error("failed to close queue client", "error", err)
		}
	}()

	// Services
	jwtService := service.NewJWTService(cfg.JWTSecret)
	authService := service.NewAuthService(client)
	sessionService := service.NewSessionService(client, queueClient)
	urlService := service.NewURLService(client)
	eventService := service.NewEventService(client, urlService)

	// Controllers
	authController := controller.NewAuthController(authService, jwtService)
	sessionController := controller.NewSessionController(sessionService, jwtService)
	eventController := controller.NewEventController(eventService, sessionService, jwtService)

	// Combined handler implementing StrictServerInterface
	handler := controller.NewHandler(authController, sessionController, eventController)

	// Router
	r := gin.New()
	r.Use(middleware.RequestID())                           // 1. Generate Request ID
	r.Use(middleware.NewLoggingMiddleware(cfg.Environment)) // 2. HTTP logging
	r.Use(gin.Recovery())                                   // 3. Panic recovery
	r.Use(middleware.CORS())                                // 4. CORS
	r.Use(middleware.Metrics())                             // 5. Prometheus metrics

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Rate limiting for auth endpoints (10 requests per minute per IP)
	authRateLimiter := middleware.AuthRateLimit()
	r.POST("/v1/auth/signup", authRateLimiter)
	r.POST("/v1/auth/login", authRateLimiter)
	r.POST("/v1/auth/forgot-password", authRateLimiter)

	// Register API handlers using generated code
	strictHandler := generated.NewStrictHandler(handler, nil)
	generated.RegisterHandlers(r, strictHandler)

	slog.Info("starting server", "port", cfg.Port, "env", cfg.Environment)
	if err := r.Run(":" + cfg.Port); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
