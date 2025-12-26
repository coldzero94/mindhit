package controller

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
)

// SessionController implements session-related handlers from StrictServerInterface
type SessionController struct {
	sessionService *service.SessionService
	jwtService     *service.JWTService
}

// NewSessionController creates a new SessionController
func NewSessionController(sessionService *service.SessionService, jwtService *service.JWTService) *SessionController {
	return &SessionController{
		sessionService: sessionService,
		jwtService:     jwtService,
	}
}

// extractUserID extracts and validates user ID from authorization header
func (c *SessionController) extractUserID(authHeader string) (uuid.UUID, error) {
	if authHeader == "" {
		return uuid.Nil, errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return uuid.Nil, errors.New("invalid authorization header format")
	}

	claims, err := c.jwtService.ValidateAccessToken(parts[1])
	if err != nil {
		return uuid.Nil, errors.New("invalid or expired access token")
	}

	return claims.UserID, nil
}

// RoutesStart handles POST /v1/sessions/start
func (c *SessionController) RoutesStart(ctx context.Context, request generated.RoutesStartRequestObject) (generated.RoutesStartResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesStart401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sess, err := c.sessionService.Start(ctx, userID)
	if err != nil {
		slog.Error("failed to create session", "error", err, "user_id", userID)
		return nil, err
	}

	return generated.RoutesStart201JSONResponse{
		Session: mapSession(sess),
	}, nil
}

// RoutesList handles GET /v1/sessions
func (c *SessionController) RoutesList(ctx context.Context, request generated.RoutesListRequestObject) (generated.RoutesListResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesList401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	limit := 20
	offset := 0
	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
	}
	if request.Params.Offset != nil {
		offset = int(*request.Params.Offset)
	}

	sessions, err := c.sessionService.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		slog.Error("failed to list sessions", "error", err, "user_id", userID)
		return nil, err
	}

	result := make([]generated.SessionSession, len(sessions))
	for i, s := range sessions {
		result[i] = mapSession(s)
	}

	return generated.RoutesList200JSONResponse{
		Sessions: result,
	}, nil
}

// RoutesGet handles GET /v1/sessions/{id}
func (c *SessionController) RoutesGet(ctx context.Context, request generated.RoutesGetRequestObject) (generated.RoutesGetResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesGet401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesGet404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	sess, err := c.sessionService.GetWithDetails(ctx, sessionID, userID)
	if err != nil {
		return c.handleGetError(err)
	}

	return generated.RoutesGet200JSONResponse{
		Session: mapSession(sess),
	}, nil
}

// RoutesUpdate handles PUT /v1/sessions/{id}
func (c *SessionController) RoutesUpdate(ctx context.Context, request generated.RoutesUpdateRequestObject) (generated.RoutesUpdateResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesUpdate401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesUpdate404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	sess, err := c.sessionService.Update(ctx, sessionID, userID, request.Body.Title, request.Body.Description)
	if err != nil {
		return c.handleUpdateError(err)
	}

	return generated.RoutesUpdate200JSONResponse{
		Session: mapSession(sess),
	}, nil
}

// RoutesPause handles PATCH /v1/sessions/{id}/pause
func (c *SessionController) RoutesPause(ctx context.Context, request generated.RoutesPauseRequestObject) (generated.RoutesPauseResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesPause401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesPause404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	sess, err := c.sessionService.Pause(ctx, sessionID, userID)
	if err != nil {
		return c.handlePauseError(err)
	}

	return generated.RoutesPause200JSONResponse{
		Session: mapSession(sess),
	}, nil
}

// RoutesResume handles PATCH /v1/sessions/{id}/resume
func (c *SessionController) RoutesResume(ctx context.Context, request generated.RoutesResumeRequestObject) (generated.RoutesResumeResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesResume401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesResume404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	sess, err := c.sessionService.Resume(ctx, sessionID, userID)
	if err != nil {
		return c.handleResumeError(err)
	}

	return generated.RoutesResume200JSONResponse{
		Session: mapSession(sess),
	}, nil
}

// RoutesStop handles POST /v1/sessions/{id}/stop
func (c *SessionController) RoutesStop(ctx context.Context, request generated.RoutesStopRequestObject) (generated.RoutesStopResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesStop401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesStop404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	sess, err := c.sessionService.Stop(ctx, sessionID, userID)
	if err != nil {
		return c.handleStopError(err)
	}

	return generated.RoutesStop200JSONResponse{
		Session: mapSession(sess),
	}, nil
}

// RoutesDelete handles DELETE /v1/sessions/{id}
func (c *SessionController) RoutesDelete(ctx context.Context, request generated.RoutesDeleteRequestObject) (generated.RoutesDeleteResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesDelete401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesDelete404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	err = c.sessionService.Delete(ctx, sessionID, userID)
	if err != nil {
		return c.handleDeleteError(err)
	}

	return generated.RoutesDelete204Response{}, nil
}

// Error handlers for different operations

func (c *SessionController) handleGetError(err error) (generated.RoutesGetResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.RoutesGet404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrSessionNotOwned):
		return generated.RoutesGet403JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "access denied",
			},
		}, nil
	default:
		slog.Error("session get failed", "error", err)
		return nil, err
	}
}

func (c *SessionController) handleUpdateError(err error) (generated.RoutesUpdateResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.RoutesUpdate404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrSessionNotOwned):
		return generated.RoutesUpdate403JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "access denied",
			},
		}, nil
	default:
		slog.Error("session update failed", "error", err)
		return nil, err
	}
}

func (c *SessionController) handlePauseError(err error) (generated.RoutesPauseResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.RoutesPause404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrSessionNotOwned):
		return generated.RoutesPause403JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "access denied",
			},
		}, nil
	case errors.Is(err, service.ErrInvalidSessionState):
		return generated.RoutesPause400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session cannot be paused from current state",
			},
		}, nil
	default:
		slog.Error("session pause failed", "error", err)
		return nil, err
	}
}

func (c *SessionController) handleResumeError(err error) (generated.RoutesResumeResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.RoutesResume404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrSessionNotOwned):
		return generated.RoutesResume403JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "access denied",
			},
		}, nil
	case errors.Is(err, service.ErrInvalidSessionState):
		return generated.RoutesResume400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session cannot be resumed from current state",
			},
		}, nil
	default:
		slog.Error("session resume failed", "error", err)
		return nil, err
	}
}

func (c *SessionController) handleStopError(err error) (generated.RoutesStopResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.RoutesStop404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrSessionNotOwned):
		return generated.RoutesStop403JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "access denied",
			},
		}, nil
	case errors.Is(err, service.ErrInvalidSessionState):
		return generated.RoutesStop400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session cannot be stopped from current state",
			},
		}, nil
	default:
		slog.Error("session stop failed", "error", err)
		return nil, err
	}
}

func (c *SessionController) handleDeleteError(err error) (generated.RoutesDeleteResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.RoutesDelete404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrSessionNotOwned):
		return generated.RoutesDelete403JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "access denied",
			},
		}, nil
	default:
		slog.Error("session delete failed", "error", err)
		return nil, err
	}
}

// mapSession converts an ent.Session to generated.SessionSession
func mapSession(s *ent.Session) generated.SessionSession {
	result := generated.SessionSession{
		Id:            s.ID.String(),
		SessionStatus: generated.SessionSessionStatus(s.SessionStatus),
		StartedAt:     s.StartedAt,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}

	if s.Title != nil {
		result.Title = s.Title
	}
	if s.Description != nil {
		result.Description = s.Description
	}
	if s.EndedAt != nil {
		result.EndedAt = s.EndedAt
	}

	return result
}
