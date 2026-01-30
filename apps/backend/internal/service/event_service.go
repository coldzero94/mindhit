// Package service provides business logic implementations.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/pagevisit"
	"github.com/mindhit/api/ent/rawevent"
	"github.com/mindhit/api/ent/session"
	enturl "github.com/mindhit/api/ent/url"
	"github.com/mindhit/api/internal/infrastructure/metrics"
)

// Event service errors.
var (
	ErrSessionNotAcceptingEvents = errors.New("session is not accepting events")
)

// EventService handles event-related business logic.
type EventService struct {
	client     *ent.Client
	urlService *URLService
}

// NewEventService creates a new EventService instance.
func NewEventService(client *ent.Client, urlService *URLService) *EventService {
	return &EventService{
		client:     client,
		urlService: urlService,
	}
}

// BatchEvent represents a single event from the extension.
type BatchEvent struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	URL       string                 `json:"url,omitempty"`
	Title     string                 `json:"title,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// ProcessBatchEvents processes multiple events at once.
func (s *EventService) ProcessBatchEvents(
	ctx context.Context,
	sessionID uuid.UUID,
	events []BatchEvent,
) (int, error) {
	// Verify session exists and is in recording/paused state
	sess, err := s.client.Session.
		Query().
		Where(session.IDEQ(sessionID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, ErrSessionNotFound
		}
		return 0, fmt.Errorf("query session: %w", err)
	}

	if sess.SessionStatus != session.SessionStatusRecording && sess.SessionStatus != session.SessionStatusPaused {
		return 0, ErrSessionNotAcceptingEvents
	}

	// Record batch size metric
	metrics.EventBatchSize.Observe(float64(len(events)))

	processed := 0

	for _, event := range events {
		if err := s.processEvent(ctx, sessionID, event); err != nil {
			// Log error but continue processing
			continue
		}
		// Record event by type
		metrics.EventsReceived.WithLabelValues(event.Type).Inc()
		processed++
	}

	return processed, nil
}

func (s *EventService) processEvent(
	ctx context.Context,
	sessionID uuid.UUID,
	event BatchEvent,
) error {
	// Store raw event
	payload, err := toJSON(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = s.client.RawEvent.
		Create().
		SetSessionID(sessionID).
		SetEventType(event.Type).
		SetTimestamp(time.UnixMilli(event.Timestamp)).
		SetPayload(payload).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("save raw event: %w", err)
	}

	// Process specific event types
	switch event.Type {
	case "page_visit":
		return s.processPageVisit(ctx, sessionID, event)
	case "page_leave":
		return s.processPageLeave(ctx, sessionID, event)
	case "highlight":
		return s.processHighlight(ctx, sessionID, event)
	}

	return nil
}

func (s *EventService) processPageVisit(
	ctx context.Context,
	sessionID uuid.UUID,
	event BatchEvent,
) error {
	if event.URL == "" {
		return nil
	}

	// Get or create URL
	url, err := s.urlService.GetOrCreate(ctx, event.URL, event.Title, event.Content)
	if err != nil {
		return fmt.Errorf("get or create url: %w", err)
	}

	// Create page visit
	_, err = s.client.PageVisit.
		Create().
		SetSessionID(sessionID).
		SetURLID(url.ID).
		SetEnteredAt(time.UnixMilli(event.Timestamp)).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("create page visit: %w", err)
	}

	return nil
}

func (s *EventService) processHighlight(
	ctx context.Context,
	sessionID uuid.UUID,
	event BatchEvent,
) error {
	// Text is stored in Content field by the controller
	text := event.Content
	if text == "" {
		return nil
	}

	// Selector and color come from Payload map
	selector, _ := event.Payload["selector"].(string)
	color, _ := event.Payload["color"].(string)
	if color == "" {
		color = "#FFFF00"
	}

	_, err := s.client.Highlight.
		Create().
		SetSessionID(sessionID).
		SetText(text).
		SetSelector(selector).
		SetColor(color).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("create highlight: %w", err)
	}

	return nil
}

func (s *EventService) processPageLeave(
	ctx context.Context,
	sessionID uuid.UUID,
	event BatchEvent,
) error {
	if event.URL == "" {
		return nil
	}

	// Get duration_ms and max_scroll_depth from payload
	durationMs, _ := event.Payload["duration_ms"].(float64)
	maxScrollDepth, _ := event.Payload["max_scroll_depth"].(float64)

	// Find the most recent PageVisit for this URL in the session
	pv, err := s.client.PageVisit.
		Query().
		Where(
			pagevisit.HasSessionWith(session.IDEQ(sessionID)),
			pagevisit.HasURLWith(enturl.URLEQ(normalizeURL(event.URL))),
		).
		Order(ent.Desc(pagevisit.FieldEnteredAt)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// No matching page visit found, skip
			return nil
		}
		return fmt.Errorf("query page visit: %w", err)
	}

	// Update the page visit with leave time and duration
	_, err = pv.Update().
		SetLeftAt(time.UnixMilli(event.Timestamp)).
		SetNillableDurationMs(intPtr(int(durationMs))).
		SetMaxScrollDepth(maxScrollDepth).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("update page visit: %w", err)
	}

	return nil
}

// intPtr returns a pointer to an int value.
func intPtr(v int) *int {
	return &v
}

// ProcessBatchEventsFromJSON processes events from raw JSON.
func (s *EventService) ProcessBatchEventsFromJSON(
	ctx context.Context,
	sessionID uuid.UUID,
	jsonData string,
) (int, error) {
	eventsJSON := gjson.Get(jsonData, "events")
	if !eventsJSON.IsArray() {
		return 0, fmt.Errorf("events must be an array")
	}

	var events []BatchEvent
	eventsJSON.ForEach(func(_, value gjson.Result) bool {
		events = append(events, BatchEvent{
			Type:      value.Get("type").String(),
			Timestamp: value.Get("timestamp").Int(),
			URL:       value.Get("url").String(),
			Title:     value.Get("title").String(),
			Content:   value.Get("content").String(),
		})
		return true
	})

	return s.ProcessBatchEvents(ctx, sessionID, events)
}

// GetEventsBySession retrieves all events for a session.
func (s *EventService) GetEventsBySession(
	ctx context.Context,
	sessionID uuid.UUID,
	eventType string,
	limit int,
	offset int,
) ([]*ent.RawEvent, int, error) {
	query := s.client.RawEvent.
		Query().
		Where(rawevent.HasSessionWith(session.IDEQ(sessionID)))

	// Apply event type filter if provided
	if eventType != "" {
		query = query.Where(rawevent.EventTypeEQ(eventType))
	}

	// Get total count
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	// Apply pagination
	if limit <= 0 {
		limit = 50 // default limit
	}
	if limit > 200 {
		limit = 200 // max limit
	}

	events, err := query.
		Order(ent.Desc(rawevent.FieldTimestamp)).
		Limit(limit).
		Offset(offset).
		All(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("get events: %w", err)
	}

	return events, total, nil
}

// GetEventStats retrieves statistics for a session's events.
func (s *EventService) GetEventStats(
	ctx context.Context,
	sessionID uuid.UUID,
) (map[string]interface{}, error) {
	// Count by event type
	pageVisits, err := s.client.RawEvent.
		Query().
		Where(
			rawevent.HasSessionWith(session.IDEQ(sessionID)),
			rawevent.EventTypeEQ("page_visit"),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count page visits: %w", err)
	}

	highlights, err := s.client.RawEvent.
		Query().
		Where(
			rawevent.HasSessionWith(session.IDEQ(sessionID)),
			rawevent.EventTypeEQ("highlight"),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count highlights: %w", err)
	}

	totalEvents, err := s.client.RawEvent.
		Query().
		Where(rawevent.HasSessionWith(session.IDEQ(sessionID))).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count total events: %w", err)
	}

	// Get unique URLs count
	uniqueURLs, err := s.client.PageVisit.
		Query().
		Where(pagevisit.HasSessionWith(session.IDEQ(sessionID))).
		QueryURL().
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count unique urls: %w", err)
	}

	return map[string]interface{}{
		"total_events": totalEvents,
		"page_visits":  pageVisits,
		"highlights":   highlights,
		"unique_urls":  uniqueURLs,
	}, nil
}

// AggregatedPageVisit represents a page visit grouped by URL with aggregated stats.
type AggregatedPageVisit struct {
	URLID           uuid.UUID
	URL             string
	Title           string
	Summary         string
	Keywords        []string
	VisitCount      int
	TotalDurationMs int
	FirstVisitedAt  time.Time
	LastVisitedAt   time.Time
}

// GetAggregatedPageVisits returns page visits grouped by URL with summaries and stats.
func (s *EventService) GetAggregatedPageVisits(
	ctx context.Context,
	sessionID uuid.UUID,
) ([]AggregatedPageVisit, error) {
	// Get all page visits for this session with URL eager loading
	pageVisits, err := s.client.PageVisit.
		Query().
		Where(pagevisit.HasSessionWith(session.IDEQ(sessionID))).
		WithURL().
		Order(ent.Asc(pagevisit.FieldEnteredAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query page visits: %w", err)
	}

	// Group by URL and aggregate stats
	urlMap := make(map[uuid.UUID]*AggregatedPageVisit)
	for _, pv := range pageVisits {
		if pv.Edges.URL == nil {
			continue
		}
		u := pv.Edges.URL
		existing, ok := urlMap[u.ID]
		if !ok {
			existing = &AggregatedPageVisit{
				URLID:          u.ID,
				URL:            u.URL,
				Title:          u.Title,
				Summary:        u.Summary,
				Keywords:       u.Keywords,
				VisitCount:     0,
				FirstVisitedAt: pv.EnteredAt,
				LastVisitedAt:  pv.EnteredAt,
			}
			urlMap[u.ID] = existing
		}

		existing.VisitCount++
		if pv.DurationMs != nil {
			existing.TotalDurationMs += *pv.DurationMs
		}
		if pv.EnteredAt.After(existing.LastVisitedAt) {
			existing.LastVisitedAt = pv.EnteredAt
		}
	}

	// Convert to slice
	result := make([]AggregatedPageVisit, 0, len(urlMap))
	for _, v := range urlMap {
		result = append(result, *v)
	}

	return result, nil
}

func toJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
