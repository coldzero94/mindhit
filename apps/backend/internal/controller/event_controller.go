package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
)

// EventController handles event-related HTTP requests.
type EventController struct {
	eventService   *service.EventService
	sessionService *service.SessionService
}

// NewEventController creates a new EventController.
func NewEventController(
	eventService *service.EventService,
	sessionService *service.SessionService,
) *EventController {
	return &EventController{
		eventService:   eventService,
		sessionService: sessionService,
	}
}

// RoutesBatchEvents implements generated.StrictServerInterface
func (c *EventController) RoutesBatchEvents(ctx context.Context, request generated.RoutesBatchEventsRequestObject) (generated.RoutesBatchEventsResponseObject, error) {
	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesBatchEvents400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: "invalid session id"},
		}, nil
	}

	// Verify session exists
	_, err = c.sessionService.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return generated.RoutesBatchEvents404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "session not found"},
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
		// Build payload map for additional fields (selector, color)
		payload := make(map[string]interface{})
		if e.Selector != nil {
			payload["selector"] = *e.Selector
		}
		if e.Color != nil {
			payload["color"] = *e.Color
		}

		batchEvents[i] = service.BatchEvent{
			Type:      e.Type,
			Timestamp: e.Timestamp,
			URL:       ptrToString(e.Url),
			Title:     ptrToString(e.Title),
			Content:   ptrToString(e.Text),
			Payload:   payload,
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
	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesListEvents404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: "invalid session id"},
		}, nil
	}

	// Verify session exists
	_, err = c.sessionService.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return generated.RoutesListEvents404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "session not found"},
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

	// Get aggregated page visits (grouped by URL with summaries)
	pageVisits := make([]generated.EventsPageVisit, 0)
	if eventType == "" || eventType == "page_visit" {
		aggregated, err := c.eventService.GetAggregatedPageVisits(ctx, sessionID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get aggregated page visits", "error", err)
			return nil, err
		}

		for _, pv := range aggregated {
			visit := generated.EventsPageVisit{
				Id:        pv.URLID.String(),
				Url:       pv.URL,
				VisitedAt: pv.FirstVisitedAt,
			}
			if pv.Title != "" {
				visit.Title = &pv.Title
			}
			if pv.Summary != "" {
				visit.Summary = &pv.Summary
			}
			if len(pv.Keywords) > 0 {
				visit.Keywords = &pv.Keywords
			}
			visitCount := int32(pv.VisitCount)
			visit.VisitCount = &visitCount
			totalDuration := int32(pv.TotalDurationMs)
			visit.TotalDurationMs = &totalDuration
			pageVisits = append(pageVisits, visit)
		}
	}

	// Get highlights from raw events
	highlights := make([]generated.EventsHighlight, 0)
	if eventType == "" || eventType == "highlight" {
		// Use limit/offset for highlights only
		limit := 50
		if request.Params.Limit != nil {
			limit = int(*request.Params.Limit)
		}
		offset := 0
		if request.Params.Offset != nil {
			offset = int(*request.Params.Offset)
		}

		events, _, err := c.eventService.GetEventsBySession(ctx, sessionID, "highlight", limit, offset)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get highlights", "error", err)
			return nil, err
		}

		for _, e := range events {
			var payload map[string]interface{}
			if e.Payload != "" {
				_ = json.Unmarshal([]byte(e.Payload), &payload)
			}

			text := getStringFromPayload(payload, "content")
			color := getStringFromPayload(payload, "color")
			if color == "" {
				color = "#FFFF00"
			}
			h := generated.EventsHighlight{
				Id:        e.ID.String(),
				Text:      text,
				Color:     color,
				CreatedAt: e.CreatedAt,
			}
			highlights = append(highlights, h)
		}
	}

	total := len(pageVisits) + len(highlights)
	return generated.RoutesListEvents200JSONResponse{
		PageVisits: pageVisits,
		Highlights: highlights,
		Total:      int32(total),
	}, nil
}

// RoutesGetEventStats implements generated.StrictServerInterface
func (c *EventController) RoutesGetEventStats(ctx context.Context, request generated.RoutesGetEventStatsRequestObject) (generated.RoutesGetEventStatsResponseObject, error) {
	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.RoutesGetEventStats404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{Message: "invalid session id"},
		}, nil
	}

	// Verify session exists
	_, err = c.sessionService.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return generated.RoutesGetEventStats404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{Message: "session not found"},
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
