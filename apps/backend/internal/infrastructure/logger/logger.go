// Package logger provides structured logging initialization based on environment.
package logger

import (
	"log/slog"
	"os"
)

// Init initializes the default slog logger based on the environment.
func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		AddSource: true,
	}

	switch env {
	case "production":
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
