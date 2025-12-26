package queue

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Server wraps asynq.Server for processing jobs.
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// ServerConfig holds configuration for the queue server.
type ServerConfig struct {
	RedisAddr   string
	Concurrency int
	Queues      map[string]int // queue name -> priority
}

// NewServer creates a new queue server.
func NewServer(cfg ServerConfig) *Server {
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 10
	}
	if cfg.Queues == nil {
		cfg.Queues = map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		}
	}

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues:      cfg.Queues,
			ErrorHandler: asynq.ErrorHandlerFunc(func(_ context.Context, task *asynq.Task, err error) {
				slog.Error("task failed",
					"type", task.Type(),
					"error", err,
				)
			}),
		},
	)

	return &Server{
		server: server,
		mux:    asynq.NewServeMux(),
	}
}

// HandleFunc registers a handler function for a task type.
func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

// Run starts the server and blocks until shutdown.
func (s *Server) Run() error {
	slog.Info("starting worker server")
	return s.server.Run(s.mux)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	s.server.Shutdown()
}
