// Package main is the entry point for the MindHit API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"entgo.io/ent/dialect"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/controller"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/infrastructure/config"
	"github.com/mindhit/api/internal/infrastructure/middleware"
	"github.com/mindhit/api/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database connection
	client, err := ent.Open(dialect.Postgres, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
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

	// Services
	jwtService := service.NewJWTService(cfg.JWTSecret)
	authService := service.NewAuthService(client)
	sessionService := service.NewSessionService(client)
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
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register API handlers using generated code
	strictHandler := generated.NewStrictHandler(handler, nil)
	generated.RegisterHandlers(r, strictHandler)

	slog.Info("starting server", "port", cfg.Port, "env", cfg.Environment)
	if err := r.Run(":" + cfg.Port); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
