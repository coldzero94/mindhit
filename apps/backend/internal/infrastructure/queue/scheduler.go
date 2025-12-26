package queue

import (
	"log/slog"

	"github.com/hibiken/asynq"
)

// Scheduler manages periodic task scheduling.
type Scheduler struct {
	scheduler *asynq.Scheduler
}

// NewScheduler creates a new scheduler.
func NewScheduler(redisAddr string) (*Scheduler, error) {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		nil,
	)
	return &Scheduler{scheduler: scheduler}, nil
}

// RegisterPeriodicTasks registers all periodic tasks.
func (s *Scheduler) RegisterPeriodicTasks() error {
	// Cleanup stale sessions every hour
	cleanupTask, err := NewSessionCleanupTask(24) // 24 hours max age
	if err != nil {
		return err
	}

	_, err = s.scheduler.Register("@every 1h", cleanupTask)
	if err != nil {
		return err
	}

	slog.Info("registered periodic cleanup task", "interval", "1h")
	return nil
}

// Run starts the scheduler (blocking).
func (s *Scheduler) Run() error {
	return s.scheduler.Run()
}

// Shutdown stops the scheduler.
func (s *Scheduler) Shutdown() {
	s.scheduler.Shutdown()
}
