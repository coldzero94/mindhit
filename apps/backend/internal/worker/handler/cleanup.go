package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/infrastructure/queue"
)

// HandleSessionCleanup cleans up stale sessions.
func (h *handlers) HandleSessionCleanup(ctx context.Context, t *asynq.Task) error {
	var payload queue.SessionCleanupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	slog.Info("starting session cleanup", "max_age_hours", payload.MaxAgeHours)

	threshold := time.Now().Add(-time.Duration(payload.MaxAgeHours) * time.Hour)

	// Batch update stale sessions to failed status
	count, err := h.client.Session.Update().
		Where(
			session.SessionStatusIn(session.SessionStatusRecording, session.SessionStatusPaused),
			session.UpdatedAtLT(threshold),
		).
		SetSessionStatus(session.SessionStatusFailed).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update stale sessions: %w", err)
	}

	slog.Info("session cleanup completed", "cleaned_count", count)
	return nil
}
