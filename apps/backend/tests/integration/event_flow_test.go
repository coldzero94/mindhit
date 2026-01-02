//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/internal/controller"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

// TestEventFlow_CollectAndQuery tests event collection and querying:
// 1. User creates a session
// 2. User sends batch events
// 3. User queries events
// 4. User gets event stats
func TestEventFlow_CollectAndQuery(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	// Setup services
	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	sessionService := service.NewSessionService(client, nil)
	urlService := service.NewURLService(client)
	eventService := service.NewEventService(client, urlService)

	authController := controller.NewAuthController(authService, jwtService)
	sessionController := controller.NewSessionController(sessionService, jwtService)
	eventController := controller.NewEventController(eventService, sessionService, jwtService)

	ctx := context.Background()

	// Setup user and session
	var accessToken string
	var sessionID string

	t.Run("Setup: Create user and session", func(t *testing.T) {
		email := uniqueEmail("event_flow")
		signupReq := generated.RoutesSignupRequestObject{
			Body: &generated.RoutesSignupJSONRequestBody{
				Email:    email,
				Password: "password123!",
			},
		}
		resp, err := authController.RoutesSignup(ctx, signupReq)
		require.NoError(t, err)
		signupResp := resp.(generated.RoutesSignup201JSONResponse)
		accessToken = "Bearer " + signupResp.Token

		startReq := generated.RoutesStartRequestObject{
			Params: generated.RoutesStartParams{Authorization: accessToken},
		}
		startResp, err := sessionController.RoutesStart(ctx, startReq)
		require.NoError(t, err)
		sessionID = startResp.(generated.RoutesStart201JSONResponse).Session.Id
	})

	now := time.Now().UnixMilli()

	// Step 1: Send batch events
	t.Run("Step 1: Send batch events", func(t *testing.T) {
		url1 := "https://example.com/page1"
		title1 := "Page 1"
		url2 := "https://example.com/page2"
		title2 := "Page 2"
		highlightText := "Important text to remember"

		events := []generated.EventsEventData{
			{
				Type:      "page_visit",
				Timestamp: now,
				Url:       &url1,
				Title:     &title1,
			},
			{
				Type:      "page_visit",
				Timestamp: now + 60000,
				Url:       &url2,
				Title:     &title2,
			},
			{
				Type:      "highlight",
				Timestamp: now + 90000,
				Url:       &url2,
				Text:      &highlightText,
			},
		}

		req := generated.RoutesBatchEventsRequestObject{
			Id:     sessionID,
			Params: generated.RoutesBatchEventsParams{Authorization: accessToken},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: events,
			},
		}

		resp, err := eventController.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		batchResp, ok := resp.(generated.RoutesBatchEvents200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, int32(3), batchResp.Processed)
		assert.Equal(t, int32(3), batchResp.Total)
	})

	// Step 2: Query events
	t.Run("Step 2: List events", func(t *testing.T) {
		req := generated.RoutesListEventsRequestObject{
			Id:     sessionID,
			Params: generated.RoutesListEventsParams{Authorization: accessToken},
		}

		resp, err := eventController.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		listResp, ok := resp.(generated.RoutesListEvents200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.GreaterOrEqual(t, len(listResp.Events), 3)

		// Check event types
		typeCount := make(map[string]int)
		for _, e := range listResp.Events {
			typeCount[e.Type]++
		}
		assert.Equal(t, 2, typeCount["page_visit"])
		assert.Equal(t, 1, typeCount["highlight"])
	})

	// Step 3: Get event stats
	t.Run("Step 3: Get event stats", func(t *testing.T) {
		req := generated.RoutesGetEventStatsRequestObject{
			Id:     sessionID,
			Params: generated.RoutesGetEventStatsParams{Authorization: accessToken},
		}

		resp, err := eventController.RoutesGetEventStats(ctx, req)
		require.NoError(t, err)

		statsResp, ok := resp.(generated.RoutesGetEventStats200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.GreaterOrEqual(t, int(statsResp.Stats.TotalEvents), 3)
		assert.GreaterOrEqual(t, int(statsResp.Stats.UniqueUrls), 2)
	})

	// Step 4: Filter events by type
	t.Run("Step 4: Filter by event type", func(t *testing.T) {
		eventType := "highlight"
		req := generated.RoutesListEventsRequestObject{
			Id: sessionID,
			Params: generated.RoutesListEventsParams{
				Authorization: accessToken,
				Type:          &eventType,
			},
		}

		resp, err := eventController.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		listResp, ok := resp.(generated.RoutesListEvents200JSONResponse)
		require.True(t, ok)

		for _, e := range listResp.Events {
			assert.Equal(t, "highlight", e.Type)
		}
	})
}

// TestEventFlow_EmptyBatch tests sending empty event batch.
func TestEventFlow_EmptyBatch(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	sessionService := service.NewSessionService(client, nil)
	urlService := service.NewURLService(client)
	eventService := service.NewEventService(client, urlService)

	authController := controller.NewAuthController(authService, jwtService)
	sessionController := controller.NewSessionController(sessionService, jwtService)
	eventController := controller.NewEventController(eventService, sessionService, jwtService)

	ctx := context.Background()

	// Setup
	email := uniqueEmail("empty_batch")
	signupReq := generated.RoutesSignupRequestObject{
		Body: &generated.RoutesSignupJSONRequestBody{
			Email:    email,
			Password: "password123!",
		},
	}
	resp, _ := authController.RoutesSignup(ctx, signupReq)
	signupResp := resp.(generated.RoutesSignup201JSONResponse)
	accessToken := "Bearer " + signupResp.Token

	startReq := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{Authorization: accessToken},
	}
	startResp, _ := sessionController.RoutesStart(ctx, startReq)
	sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

	// Send empty batch
	t.Run("Empty batch returns 200 with zero processed", func(t *testing.T) {
		req := generated.RoutesBatchEventsRequestObject{
			Id:     sessionID,
			Params: generated.RoutesBatchEventsParams{Authorization: accessToken},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{},
			},
		}

		resp, err := eventController.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		batchResp, ok := resp.(generated.RoutesBatchEvents200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, int32(0), batchResp.Processed)
	})
}

// TestEventFlow_InvalidSession tests events for non-existent session.
func TestEventFlow_InvalidSession(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	sessionService := service.NewSessionService(client, nil)
	urlService := service.NewURLService(client)
	eventService := service.NewEventService(client, urlService)

	authController := controller.NewAuthController(authService, jwtService)
	eventController := controller.NewEventController(eventService, sessionService, jwtService)

	ctx := context.Background()

	// Setup user
	email := uniqueEmail("invalid_session")
	signupReq := generated.RoutesSignupRequestObject{
		Body: &generated.RoutesSignupJSONRequestBody{
			Email:    email,
			Password: "password123!",
		},
	}
	resp, _ := authController.RoutesSignup(ctx, signupReq)
	signupResp := resp.(generated.RoutesSignup201JSONResponse)
	accessToken := "Bearer " + signupResp.Token

	// Try to send events to non-existent session
	t.Run("Events to invalid session returns 404", func(t *testing.T) {
		fakeSessionID := uuid.New().String()
		url := "https://example.com"
		now := time.Now().UnixMilli()

		req := generated.RoutesBatchEventsRequestObject{
			Id:     fakeSessionID,
			Params: generated.RoutesBatchEventsParams{Authorization: accessToken},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{
					{
						Type:      "page_visit",
						Timestamp: now,
						Url:       &url,
					},
				},
			},
		}

		resp, err := eventController.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents404JSONResponse)
		require.True(t, ok, "expected 404 for non-existent session")
	})
}
