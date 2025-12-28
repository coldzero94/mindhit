// Package main is the entry point for the MindHit Worker server.
package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/infrastructure/config"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/worker/handler"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Load config
	cfg := config.Load()

	// Setup logger
	var logHandler slog.Handler
	if cfg.Environment == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(logHandler))

	slog.Info("starting worker", "env", cfg.Environment, "redis", cfg.RedisAddr)

	// Connect to database with connection pool settings
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
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

	// Run migrations in development
	if cfg.Environment != "production" {
		if err := client.Schema.Create(context.Background()); err != nil {
			slog.Error("failed to create schema", "error", err)
		}
	}

	// Create worker server
	server := queue.NewServer(queue.ServerConfig{
		RedisAddr:   cfg.RedisAddr,
		Concurrency: cfg.WorkerConcurrency,
	})

	// Register handlers
	handler.RegisterHandlers(server, client)

	// Create scheduler for periodic tasks
	scheduler, err := queue.NewScheduler(cfg.RedisAddr)
	if err != nil {
		return err
	}

	if err := scheduler.RegisterPeriodicTasks(); err != nil {
		return err
	}

	// Run scheduler in background
	go func() {
		if err := scheduler.Run(); err != nil {
			slog.Error("scheduler error", "error", err)
		}
	}()
	defer scheduler.Shutdown()

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down worker")
		server.Shutdown()
	}()

	// Start server (blocking)
	return server.Run()
}
