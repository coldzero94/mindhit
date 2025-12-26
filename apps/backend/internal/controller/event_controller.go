package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
)

// EventController handles event-related HTTP requests.
type EventController struct {
	eventService   *service.EventService
	sessionService *service.SessionService
	jwtService     *service.JWTService
}

// NewEventController creates a new EventController.
func NewEventController(
	eventService *service.EventService,
	sessionService *service.SessionService,
	jwtService *service.JWTService,
) *EventController {
	return &EventController{
		eventService:   eventService,
		sessionService: sessionService,
		jwtService:     jwtService,
	}
}

// extractUserID extracts and validates user ID from authorization header.
func (c *EventController) extractUserID(authHeader string) (uuid.UUID, error) {
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

// RoutesBatchEvents implements generated.StrictServerInterface
func (c *EventController) RoutesBatchEvents(ctx context.Context, request generated.RoutesBatchEventsRequestObject) (generated.RoutesBatchEventsResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesBatchEvents401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: err.Error()},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesBatchEvents400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: "invalid session id"},
		}, nil
	}

	// Verify session ownership
	_, err = c.sessionService.Get(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return generated.RoutesBatchEvents404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "session not found"},
			}, nil
		}
		if errors.Is(err, service.ErrSessionNotOwned) {
			return generated.RoutesBatchEvents403JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "access denied"},
			}, nil
		}
		slog.ErrorContext(ctx, "failed to get session", "error", err)
		return nil, err
	}

	if request.Body == nil || len(request.Body.Events) == 0 {
		return generated.RoutesBatchEvents400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: "no events provided"},
		}, nil
	}

	// Convert generated events to service events
	batchEvents := make([]service.BatchEvent, len(request.Body.Events))
	for i, e := range request.Body.Events {
		batchEvents[i] = service.BatchEvent{
			Type:      e.Type,
			Timestamp: e.Timestamp,
			URL:       ptrToString(e.Url),
			Title:     ptrToString(e.Title),
			Content:   ptrToString(e.Text),
		}
	}

	processed, err := c.eventService.ProcessBatchEvents(ctx, sessionID, batchEvents)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotAcceptingEvents) {
			return generated.RoutesBatchEvents400JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "session is not accepting events"},
			}, nil
		}
		slog.ErrorContext(ctx, "failed to process events", "error", err)
		return nil, err
	}

	return generated.RoutesBatchEvents200JSONResponse{
		Processed: int32(processed),
		Total:     int32(len(request.Body.Events)),
	}, nil
}

// RoutesListEvents implements generated.StrictServerInterface
func (c *EventController) RoutesListEvents(ctx context.Context, request generated.RoutesListEventsRequestObject) (generated.RoutesListEventsResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesListEvents401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: err.Error()},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesListEvents404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: "invalid session id"},
		}, nil
	}

	// Verify session ownership
	_, err = c.sessionService.Get(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return generated.RoutesListEvents404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "session not found"},
			}, nil
		}
		if errors.Is(err, service.ErrSessionNotOwned) {
			return generated.RoutesListEvents403JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "access denied"},
			}, nil
		}
		slog.ErrorContext(ctx, "failed to get session", "error", err)
		return nil, err
	}

	// Parse query params
	eventType := ""
	if request.Params.Type != nil {
		eventType = *request.Params.Type
	}
	limit := 50
	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
	}
	offset := 0
	if request.Params.Offset != nil {
		offset = int(*request.Params.Offset)
	}

	events, total, err := c.eventService.GetEventsBySession(ctx, sessionID, eventType, limit, offset)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get events", "error", err)
		return nil, err
	}

	// Convert to response types
	pageVisits := make([]generated.EventsPageVisit, 0)
	highlights := make([]generated.EventsHighlight, 0)

	for _, e := range events {
		// Parse payload JSON to extract fields
		var payload map[string]interface{}
		if e.Payload != "" {
			_ = json.Unmarshal([]byte(e.Payload), &payload)
		}

		if e.EventType == "page_visit" {
			pv := generated.EventsPageVisit{
				Id:        e.ID.String(),
				Url:       getStringFromPayload(payload, "url"),
				VisitedAt: e.CreatedAt,
			}
			if title := getStringFromPayload(payload, "title"); title != "" {
				pv.Title = &title
			}
			pageVisits = append(pageVisits, pv)
		} else if e.EventType == "highlight" {
			h := generated.EventsHighlight{
				Id:        e.ID.String(),
				Text:      getStringFromPayload(payload, "text"),
				Color:     "#FFFF00",
				CreatedAt: e.CreatedAt,
			}
			highlights = append(highlights, h)
		}
	}

	return generated.RoutesListEvents200JSONResponse{
		PageVisits: pageVisits,
		Highlights: highlights,
		Total:      int32(total),
	}, nil
}

// RoutesGetEventStats implements generated.StrictServerInterface
func (c *EventController) RoutesGetEventStats(ctx context.Context, request generated.RoutesGetEventStatsRequestObject) (generated.RoutesGetEventStatsResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.RoutesGetEventStats401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: err.Error()},
		}, nil
	}

	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesGetEventStats404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: "invalid session id"},
		}, nil
	}

	// Verify session ownership
	_, err = c.sessionService.Get(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return generated.RoutesGetEventStats404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "session not found"},
			}, nil
		}
		if errors.Is(err, service.ErrSessionNotOwned) {
			return generated.RoutesGetEventStats403JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "access denied"},
			}, nil
		}
		slog.ErrorContext(ctx, "failed to get session", "error", err)
		return nil, err
	}

	stats, err := c.eventService.GetEventStats(ctx, sessionID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get event stats", "error", err)
		return nil, err
	}

	return generated.RoutesGetEventStats200JSONResponse{
		TotalEvents: int32(stats["total_events"].(int)),
		PageVisits:  int32(stats["page_visits"].(int)),
		Highlights:  int32(stats["highlights"].(int)),
		UniqueUrls:  int32(stats["unique_urls"].(int)),
	}, nil
}

// ptrToString safely converts a string pointer to string
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getStringFromPayload extracts a string value from a JSON payload map
func getStringFromPayload(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}
	if v, ok := payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
