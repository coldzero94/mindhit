package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/ent/user"
)

// Session soft delete status value (uses "inactive" from SoftDeleteMixin)
const sessionStatusInactive = "inactive"

// Session service errors
var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionNotOwned     = errors.New("session not owned by user")
	ErrInvalidSessionState = errors.New("invalid session state transition")
)

// SessionService handles session-related business logic.
type SessionService struct {
	client *ent.Client
}

// NewSessionService creates a new SessionService instance.
func NewSessionService(client *ent.Client) *SessionService {
	return &SessionService{client: client}
}

// activeSessions returns a query filtered to active (non-deleted) sessions only.
func (s *SessionService) activeSessions() *ent.SessionQuery {
	return s.client.Session.Query().Where(session.StatusNEQ(session.Status(sessionStatusInactive)))
}

// Start creates a new recording session
func (s *SessionService) Start(ctx context.Context, userID uuid.UUID) (*ent.Session, error) {
	return s.client.Session.
		Create().
		SetUserID(userID).
		SetSessionStatus(session.SessionStatusRecording).
		SetStartedAt(time.Now()).
		Save(ctx)
}

// Pause pauses a recording session
func (s *SessionService) Pause(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
	sess, err := s.getOwnedSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	if sess.SessionStatus != session.SessionStatusRecording {
		return nil, ErrInvalidSessionState
	}

	return s.client.Session.
		UpdateOneID(sessionID).
		SetSessionStatus(session.SessionStatusPaused).
		Save(ctx)
}

// Resume resumes a paused session
func (s *SessionService) Resume(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
	sess, err := s.getOwnedSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	if sess.SessionStatus != session.SessionStatusPaused {
		return nil, ErrInvalidSessionState
	}

	return s.client.Session.
		UpdateOneID(sessionID).
		SetSessionStatus(session.SessionStatusRecording).
		Save(ctx)
}

// Stop stops a session and marks it for processing
func (s *SessionService) Stop(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
	sess, err := s.getOwnedSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	if sess.SessionStatus != session.SessionStatusRecording && sess.SessionStatus != session.SessionStatusPaused {
		return nil, ErrInvalidSessionState
	}

	now := time.Now()
	return s.client.Session.
		UpdateOneID(sessionID).
		SetSessionStatus(session.SessionStatusProcessing).
		SetEndedAt(now).
		Save(ctx)
}

// Get retrieves a session by ID with ownership check
func (s *SessionService) Get(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
	return s.getOwnedSession(ctx, sessionID, userID)
}

// GetWithDetails retrieves a session with all related data
func (s *SessionService) GetWithDetails(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
	sess, err := s.activeSessions().
		Where(session.IDEQ(sessionID)).
		WithUser().
		WithPageVisits(func(q *ent.PageVisitQuery) {
			q.WithURL()
		}).
		WithHighlights().
		WithMindmap().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if sess.Edges.User.ID != userID {
		return nil, ErrSessionNotOwned
	}

	return sess, nil
}

// ListByUser retrieves all active sessions for a user
func (s *SessionService) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ent.Session, error) {
	return s.activeSessions().
		Where(session.HasUserWith(user.IDEQ(userID))).
		Order(ent.Desc(session.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
}

// Update updates session metadata (title, description)
func (s *SessionService) Update(ctx context.Context, sessionID, userID uuid.UUID, title, description *string) (*ent.Session, error) {
	sess, err := s.getOwnedSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	update := s.client.Session.UpdateOneID(sess.ID)

	if title != nil {
		update.SetTitle(*title)
	}
	if description != nil {
		update.SetDescription(*description)
	}

	return update.Save(ctx)
}

// Delete soft-deletes a session by setting status to "deleted" and deleted_at timestamp.
func (s *SessionService) Delete(ctx context.Context, sessionID, userID uuid.UUID) error {
	sess, err := s.getOwnedSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = s.client.Session.
		UpdateOneID(sess.ID).
		SetStatus(session.Status(sessionStatusInactive)).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

// getOwnedSession retrieves an active session and verifies ownership.
func (s *SessionService) getOwnedSession(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
	sess, err := s.activeSessions().
		Where(session.IDEQ(sessionID)).
		WithUser().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if sess.Edges.User.ID != userID {
		return nil, ErrSessionNotOwned
	}

	return sess, nil
}
