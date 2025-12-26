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

	// Find stale sessions
	staleSessions, err := h.client.Session.Query().
		Where(
			session.SessionStatusIn(session.SessionStatusRecording, session.SessionStatusPaused),
			session.UpdatedAtLT(threshold),
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query stale sessions: %w", err)
	}

	// Update each to failed status
	for _, sess := range staleSessions {
		_, err := h.client.Session.UpdateOne(sess).
			SetSessionStatus(session.SessionStatusFailed).
			Save(ctx)
		if err != nil {
			slog.Error("failed to update stale session",
				"session_id", sess.ID,
				"error", err,
			)
		}
	}

	slog.Info("session cleanup completed", "cleaned_count", len(staleSessions))
	return nil
}
