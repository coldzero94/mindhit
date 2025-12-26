package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/infrastructure/queue"
)

// HandleSessionProcess processes a completed session.
func (h *handlers) HandleSessionProcess(ctx context.Context, t *asynq.Task) error {
	var payload queue.SessionProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	slog.Info("processing session", "session_id", payload.SessionID)

	sessionID, err := uuid.Parse(payload.SessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	// Get session
	sess, err := h.client.Session.Query().
		Where(session.IDEQ(sessionID)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Verify session is in processing state
	if sess.SessionStatus != session.SessionStatusProcessing {
		slog.Warn("session not in processing state",
			"session_id", payload.SessionID,
			"status", sess.SessionStatus,
		)
		return nil // Not an error, just skip
	}

	// TODO: Phase 9에서 AI 처리 로직 추가
	// 1. URL 요약
	// 2. 마인드맵 생성

	// Mark as completed for now
	_, err = h.client.Session.UpdateOneID(sessionID).
		SetSessionStatus(session.SessionStatusCompleted).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	slog.Info("session processing completed", "session_id", payload.SessionID)
	return nil
}
