package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

func setupEventServiceTest(t *testing.T) (*ent.Client, *service.EventService, *service.URLService, *service.SessionService, *service.AuthService) {
	t.Helper()
	client := testutil.SetupTestDB(t)
	urlService := service.NewURLService(client)
	eventService := service.NewEventService(client, urlService)
	sessionService := service.NewSessionService(client, nil) // nil queue client for tests
	authService := service.NewAuthService(client)
	return client, eventService, urlService, sessionService, authService
}

// ==================== ProcessBatchEvents Tests ====================

func TestEventService_ProcessBatchEvents_Success(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("event-test"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	events := []service.BatchEvent{
		{
			Type:      "page_visit",
			Timestamp: time.Now().UnixMilli(),
			URL:       "https://example.com/page1",
			Title:     "Example Page 1",
		},
		{
			Type:      "page_visit",
			Timestamp: time.Now().UnixMilli(),
			URL:       "https://example.com/page2",
			Title:     "Example Page 2",
		},
	}

	processed, err := eventService.ProcessBatchEvents(ctx, sess.ID, events)

	require.NoError(t, err)
	assert.Equal(t, 2, processed)
}

func TestEventService_ProcessBatchEvents_WithHighlight(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("highlight-test"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	events := []service.BatchEvent{
		{
			Type:      "highlight",
			Timestamp: time.Now().UnixMilli(),
			Payload: map[string]interface{}{
				"text":     "Important text",
				"selector": "#main p",
				"color":    "#FF0000",
			},
		},
	}

	processed, err := eventService.ProcessBatchEvents(ctx, sess.ID, events)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify highlight was created for this session
	highlights, err := sess.QueryHighlights().All(ctx)
	require.NoError(t, err)
	assert.Len(t, highlights, 1)
	assert.Equal(t, "Important text", highlights[0].Text)
	assert.Equal(t, "#FF0000", highlights[0].Color)
}

func TestEventService_ProcessBatchEvents_SessionNotFound(t *testing.T) {
	client, eventService, _, _, _ := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	fakeSessionID := uuid.New()

	events := []service.BatchEvent{
		{
			Type:      "page_visit",
			Timestamp: time.Now().UnixMilli(),
			URL:       "https://example.com",
		},
	}

	_, err := eventService.ProcessBatchEvents(ctx, fakeSessionID, events)

	assert.ErrorIs(t, err, service.ErrSessionNotFound)
}

func TestEventService_ProcessBatchEvents_SessionNotAcceptingEvents(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("stopped-session"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Stop the session
	_, err = sessionService.Stop(ctx, sess.ID, user.ID)
	require.NoError(t, err)

	events := []service.BatchEvent{
		{
			Type:      "page_visit",
			Timestamp: time.Now().UnixMilli(),
			URL:       "https://example.com",
		},
	}

	_, err = eventService.ProcessBatchEvents(ctx, sess.ID, events)

	assert.ErrorIs(t, err, service.ErrSessionNotAcceptingEvents)
}

func TestEventService_ProcessBatchEvents_PausedSessionAcceptsEvents(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("paused-session"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Pause the session
	_, err = sessionService.Pause(ctx, sess.ID, user.ID)
	require.NoError(t, err)

	events := []service.BatchEvent{
		{
			Type:      "page_visit",
			Timestamp: time.Now().UnixMilli(),
			URL:       "https://example.com",
		},
	}

	processed, err := eventService.ProcessBatchEvents(ctx, sess.ID, events)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

// ==================== GetEventsBySession Tests ====================

func TestEventService_GetEventsBySession_Success(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("list-events"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Create some events
	events := []service.BatchEvent{
		{Type: "page_visit", Timestamp: time.Now().UnixMilli(), URL: "https://example.com/1"},
		{Type: "page_visit", Timestamp: time.Now().UnixMilli(), URL: "https://example.com/2"},
		{Type: "highlight", Timestamp: time.Now().UnixMilli(), Payload: map[string]interface{}{"text": "test"}},
	}
	_, err = eventService.ProcessBatchEvents(ctx, sess.ID, events)
	require.NoError(t, err)

	// Get events
	result, total, err := eventService.GetEventsBySession(ctx, sess.ID, "", 50, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, result, 3)
}

func TestEventService_GetEventsBySession_FilterByType(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("filter-events"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Create mixed events
	events := []service.BatchEvent{
		{Type: "page_visit", Timestamp: time.Now().UnixMilli(), URL: "https://example.com/1"},
		{Type: "page_visit", Timestamp: time.Now().UnixMilli(), URL: "https://example.com/2"},
		{Type: "highlight", Timestamp: time.Now().UnixMilli(), Payload: map[string]interface{}{"text": "test"}},
	}
	_, err = eventService.ProcessBatchEvents(ctx, sess.ID, events)
	require.NoError(t, err)

	// Get only page_visit events
	result, total, err := eventService.GetEventsBySession(ctx, sess.ID, "page_visit", 50, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
}

func TestEventService_GetEventsBySession_Pagination(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("pagination"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Create 5 events
	for i := 0; i < 5; i++ {
		events := []service.BatchEvent{
			{Type: "page_visit", Timestamp: time.Now().UnixMilli() + int64(i), URL: "https://example.com"},
		}
		_, err = eventService.ProcessBatchEvents(ctx, sess.ID, events)
		require.NoError(t, err)
	}

	// Get first page
	result, total, err := eventService.GetEventsBySession(ctx, sess.ID, "", 2, 0)

	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, result, 2)

	// Get second page
	result2, _, err := eventService.GetEventsBySession(ctx, sess.ID, "", 2, 2)

	require.NoError(t, err)
	assert.Len(t, result2, 2)
}

// ==================== GetEventStats Tests ====================

func TestEventService_GetEventStats_Success(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("stats"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Create mixed events
	events := []service.BatchEvent{
		{Type: "page_visit", Timestamp: time.Now().UnixMilli(), URL: "https://example.com/1"},
		{Type: "page_visit", Timestamp: time.Now().UnixMilli(), URL: "https://example.com/2"},
		{Type: "page_visit", Timestamp: time.Now().UnixMilli(), URL: "https://example.com/1"}, // duplicate URL
		{Type: "highlight", Timestamp: time.Now().UnixMilli(), Payload: map[string]interface{}{"text": "test1"}},
		{Type: "highlight", Timestamp: time.Now().UnixMilli(), Payload: map[string]interface{}{"text": "test2"}},
	}
	_, err = eventService.ProcessBatchEvents(ctx, sess.ID, events)
	require.NoError(t, err)

	stats, err := eventService.GetEventStats(ctx, sess.ID)

	require.NoError(t, err)
	assert.Equal(t, 5, stats["total_events"])
	assert.Equal(t, 3, stats["page_visits"])
	assert.Equal(t, 2, stats["highlights"])
	assert.Equal(t, 2, stats["unique_urls"]) // 2 unique URLs
}

func TestEventService_GetEventStats_EmptySession(t *testing.T) {
	client, eventService, _, sessionService, authService := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("empty-stats"))
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	stats, err := eventService.GetEventStats(ctx, sess.ID)

	require.NoError(t, err)
	assert.Equal(t, 0, stats["total_events"])
	assert.Equal(t, 0, stats["page_visits"])
	assert.Equal(t, 0, stats["highlights"])
	assert.Equal(t, 0, stats["unique_urls"])
}

// ==================== URL Service Tests ====================

// uniqueURL generates a unique URL for test isolation
func uniqueURL(prefix string) string {
	return fmt.Sprintf("https://example.com/%s-%s", prefix, uuid.New().String()[:8])
}

func TestURLService_GetOrCreate_NewURL(t *testing.T) {
	client, _, urlService, _, _ := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	testURL := uniqueURL("new")

	url, err := urlService.GetOrCreate(ctx, testURL, "Test Page", "Page content")

	require.NoError(t, err)
	assert.NotNil(t, url)
	assert.Equal(t, testURL, url.URL)
	assert.Equal(t, "Test Page", url.Title)
	assert.Equal(t, "Page content", url.Content)
}

func TestURLService_GetOrCreate_ExistingURL(t *testing.T) {
	client, _, urlService, _, _ := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	testURL := uniqueURL("existing")

	// Create first
	url1, err := urlService.GetOrCreate(ctx, testURL, "Title 1", "")
	require.NoError(t, err)

	// Get existing
	url2, err := urlService.GetOrCreate(ctx, testURL, "Title 2", "")
	require.NoError(t, err)

	assert.Equal(t, url1.ID, url2.ID)
	assert.Equal(t, "Title 1", url2.Title) // Original title preserved
}

func TestURLService_GetOrCreate_UpdatesContent(t *testing.T) {
	client, _, urlService, _, _ := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	testURL := uniqueURL("content")

	// Create without content
	_, err := urlService.GetOrCreate(ctx, testURL, "Title", "")
	require.NoError(t, err)

	// Get with content - should update
	url, err := urlService.GetOrCreate(ctx, testURL, "Title", "New content")
	require.NoError(t, err)

	assert.Equal(t, "New content", url.Content)
}

func TestURLService_GetOrCreate_NormalizesURL(t *testing.T) {
	client, _, urlService, _, _ := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	basePath := uuid.New().String()[:8]

	// Create with trailing slash
	url1, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/%s/", basePath), "Title", "")
	require.NoError(t, err)

	// Get without trailing slash - should match
	url2, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/%s", basePath), "Title", "")
	require.NoError(t, err)

	assert.Equal(t, url1.ID, url2.ID)
}

func TestURLService_GetOrCreate_CaseInsensitive(t *testing.T) {
	client, _, urlService, _, _ := setupEventServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	basePath := uuid.New().String()[:8]

	// Create with uppercase
	url1, err := urlService.GetOrCreate(ctx, fmt.Sprintf("HTTPS://EXAMPLE.COM/%s", basePath), "Title", "")
	require.NoError(t, err)

	// Get with lowercase - should match
	url2, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/%s", basePath), "Title", "")
	require.NoError(t, err)

	assert.Equal(t, url1.ID, url2.ID)
}
