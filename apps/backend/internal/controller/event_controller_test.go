package controller

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

func setupEventControllerTest(t *testing.T) (*EventController, *service.SessionService, *service.AuthService, *service.JWTService, func()) {
	client := testutil.SetupTestDB(t)
	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	sessionService := service.NewSessionService(client, nil)
	urlService := service.NewURLService(client)
	eventService := service.NewEventService(client, urlService)
	controller := NewEventController(eventService, sessionService, jwtService)

	cleanup := func() {
		testutil.CleanupTestDB(t, client)
	}

	return controller, sessionService, authService, jwtService, cleanup
}

func createUserSessionAndToken(t *testing.T, authService *service.AuthService, sessionService *service.SessionService, jwtService *service.JWTService, email string) (string, uuid.UUID) {
	ctx := context.Background()
	user, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)

	return "Bearer " + tokenPair.AccessToken, sess.ID
}

// ==================== BatchEvents Tests ====================

func TestEventController_RoutesBatchEvents(t *testing.T) {
	controller, sessionService, authService, jwtService, cleanup := setupEventControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token, sessionID := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("batch"))

	t.Run("successful batch events", func(t *testing.T) {
		url := "https://example.com/page1"
		title := "Page 1"
		text := "Some highlighted text"
		now := time.Now().UnixMilli()

		req := generated.RoutesBatchEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesBatchEventsParams{
				Authorization: token,
			},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{
					{
						Type:      "page_visit",
						Timestamp: now,
						Url:       &url,
						Title:     &title,
					},
					{
						Type:      "highlight",
						Timestamp: now,
						Url:       &url,
						Text:      &text,
					},
				},
			},
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesBatchEvents200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, int32(2), successResp.Processed)
		assert.Equal(t, int32(2), successResp.Total)
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesBatchEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesBatchEventsParams{
				Authorization: "",
			},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{},
			},
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("invalid session id returns 400", func(t *testing.T) {
		req := generated.RoutesBatchEventsRequestObject{
			Id: "invalid-uuid",
			Params: generated.RoutesBatchEventsParams{
				Authorization: token,
			},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{},
			},
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents400JSONResponse)
		assert.True(t, ok, "expected 400 response")
	})

	t.Run("non-existent session returns 404", func(t *testing.T) {
		req := generated.RoutesBatchEventsRequestObject{
			Id: uuid.New().String(),
			Params: generated.RoutesBatchEventsParams{
				Authorization: token,
			},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{},
			},
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("other user's session returns 403", func(t *testing.T) {
		otherToken, _ := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("other"))

		req := generated.RoutesBatchEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesBatchEventsParams{
				Authorization: otherToken,
			},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{},
			},
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents403JSONResponse)
		assert.True(t, ok, "expected 403 response")
	})

	t.Run("empty events returns 400", func(t *testing.T) {
		req := generated.RoutesBatchEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesBatchEventsParams{
				Authorization: token,
			},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{},
			},
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents400JSONResponse)
		assert.True(t, ok, "expected 400 response for empty events")
	})

	t.Run("nil body returns 400", func(t *testing.T) {
		req := generated.RoutesBatchEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesBatchEventsParams{
				Authorization: token,
			},
			Body: nil,
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents400JSONResponse)
		assert.True(t, ok, "expected 400 response for nil body")
	})

	t.Run("stopped session returns 400", func(t *testing.T) {
		// Create a new session and stop it
		token2, sessionID2 := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("stopped"))

		// Get user ID from token
		claims, err := jwtService.ValidateAccessToken(token2[7:]) // Remove "Bearer "
		require.NoError(t, err)

		_, err = sessionService.Stop(ctx, sessionID2, claims.UserID)
		require.NoError(t, err)

		url := "https://example.com"
		req := generated.RoutesBatchEventsRequestObject{
			Id: sessionID2.String(),
			Params: generated.RoutesBatchEventsParams{
				Authorization: token2,
			},
			Body: &generated.RoutesBatchEventsJSONRequestBody{
				Events: []generated.EventsEventData{
					{Type: "page_visit", Timestamp: time.Now().UnixMilli(), Url: &url},
				},
			},
		}

		resp, err := controller.RoutesBatchEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesBatchEvents400JSONResponse)
		assert.True(t, ok, "expected 400 response for stopped session")
	})
}

// ==================== ListEvents Tests ====================

func TestEventController_RoutesListEvents(t *testing.T) {
	controller, sessionService, authService, jwtService, cleanup := setupEventControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token, sessionID := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("list"))

	// Add some events first
	url := "https://example.com/page1"
	title := "Page 1"
	text := "Highlighted text"
	now := time.Now().UnixMilli()

	batchReq := generated.RoutesBatchEventsRequestObject{
		Id: sessionID.String(),
		Params: generated.RoutesBatchEventsParams{
			Authorization: token,
		},
		Body: &generated.RoutesBatchEventsJSONRequestBody{
			Events: []generated.EventsEventData{
				{Type: "page_visit", Timestamp: now, Url: &url, Title: &title},
				{Type: "highlight", Timestamp: now, Url: &url, Text: &text},
			},
		},
	}
	_, err := controller.RoutesBatchEvents(ctx, batchReq)
	require.NoError(t, err)

	t.Run("successful list events", func(t *testing.T) {
		limit := int32(50)
		offset := int32(0)
		req := generated.RoutesListEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesListEventsParams{
				Authorization: token,
				Limit:         &limit,
				Offset:        &offset,
			},
		}

		resp, err := controller.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesListEvents200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.GreaterOrEqual(t, len(successResp.PageVisits), 1)
		assert.GreaterOrEqual(t, len(successResp.Highlights), 1)
	})

	t.Run("filter by type", func(t *testing.T) {
		limit := int32(50)
		offset := int32(0)
		eventType := "page_visit"
		req := generated.RoutesListEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesListEventsParams{
				Authorization: token,
				Limit:         &limit,
				Offset:        &offset,
				Type:          &eventType,
			},
		}

		resp, err := controller.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesListEvents200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.GreaterOrEqual(t, len(successResp.PageVisits), 1)
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesListEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesListEventsParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesListEvents401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("invalid session id returns 404", func(t *testing.T) {
		req := generated.RoutesListEventsRequestObject{
			Id: "invalid-uuid",
			Params: generated.RoutesListEventsParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesListEvents404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("non-existent session returns 404", func(t *testing.T) {
		req := generated.RoutesListEventsRequestObject{
			Id: uuid.New().String(),
			Params: generated.RoutesListEventsParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesListEvents404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("other user's session returns 403", func(t *testing.T) {
		otherToken, _ := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("other-list"))

		req := generated.RoutesListEventsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesListEventsParams{
				Authorization: otherToken,
			},
		}

		resp, err := controller.RoutesListEvents(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesListEvents403JSONResponse)
		assert.True(t, ok, "expected 403 response")
	})
}

// ==================== GetEventStats Tests ====================

func TestEventController_RoutesGetEventStats(t *testing.T) {
	controller, sessionService, authService, jwtService, cleanup := setupEventControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token, sessionID := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("stats"))

	// Add some events first
	url1 := "https://example.com/page1"
	url2 := "https://example.com/page2"
	title := "Page"
	text := "Highlighted text"
	now := time.Now().UnixMilli()

	batchReq := generated.RoutesBatchEventsRequestObject{
		Id: sessionID.String(),
		Params: generated.RoutesBatchEventsParams{
			Authorization: token,
		},
		Body: &generated.RoutesBatchEventsJSONRequestBody{
			Events: []generated.EventsEventData{
				{Type: "page_visit", Timestamp: now, Url: &url1, Title: &title},
				{Type: "page_visit", Timestamp: now, Url: &url2, Title: &title},
				{Type: "highlight", Timestamp: now, Url: &url1, Text: &text},
			},
		},
	}
	_, err := controller.RoutesBatchEvents(ctx, batchReq)
	require.NoError(t, err)

	t.Run("successful get stats", func(t *testing.T) {
		req := generated.RoutesGetEventStatsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesGetEventStatsParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesGetEventStats(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesGetEventStats200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, int32(3), successResp.TotalEvents)
		assert.Equal(t, int32(2), successResp.PageVisits)
		assert.Equal(t, int32(1), successResp.Highlights)
		assert.Equal(t, int32(2), successResp.UniqueUrls)
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesGetEventStatsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesGetEventStatsParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesGetEventStats(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGetEventStats401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("invalid session id returns 404", func(t *testing.T) {
		req := generated.RoutesGetEventStatsRequestObject{
			Id: "invalid-uuid",
			Params: generated.RoutesGetEventStatsParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesGetEventStats(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGetEventStats404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("non-existent session returns 404", func(t *testing.T) {
		req := generated.RoutesGetEventStatsRequestObject{
			Id: uuid.New().String(),
			Params: generated.RoutesGetEventStatsParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesGetEventStats(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGetEventStats404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("other user's session returns 403", func(t *testing.T) {
		otherToken, _ := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("other-stats"))

		req := generated.RoutesGetEventStatsRequestObject{
			Id: sessionID.String(),
			Params: generated.RoutesGetEventStatsParams{
				Authorization: otherToken,
			},
		}

		resp, err := controller.RoutesGetEventStats(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGetEventStats403JSONResponse)
		assert.True(t, ok, "expected 403 response")
	})
}

// ==================== Helper Function Tests ====================

func TestEventController_ExtractUserID(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	jwtService := service.NewJWTService("test-secret")
	controller := NewEventController(nil, nil, jwtService)

	ctx := context.Background()

	// Create a user and get token
	authService := service.NewAuthService(client)
	user, err := authService.Signup(ctx, uniqueEmail("extract"), "password123")
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		userID, err := controller.extractUserID("Bearer " + tokenPair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, user.ID, userID)
	})

	t.Run("empty header", func(t *testing.T) {
		_, err := controller.extractUserID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization header is required")
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := controller.extractUserID("InvalidFormat")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := controller.extractUserID("Bearer invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired access token")
	})
}

// ==================== Additional Coverage Tests ====================

func TestEventController_PtrToString(t *testing.T) {
	t.Run("nil pointer returns empty string", func(t *testing.T) {
		result := ptrToString(nil)
		assert.Equal(t, "", result)
	})

	t.Run("valid pointer returns string", func(t *testing.T) {
		s := "test"
		result := ptrToString(&s)
		assert.Equal(t, "test", result)
	})
}

func TestEventController_GetStringFromPayload(t *testing.T) {
	t.Run("nil payload returns empty string", func(t *testing.T) {
		result := getStringFromPayload(nil, "key")
		assert.Equal(t, "", result)
	})

	t.Run("missing key returns empty string", func(t *testing.T) {
		payload := map[string]interface{}{"other": "value"}
		result := getStringFromPayload(payload, "key")
		assert.Equal(t, "", result)
	})

	t.Run("non-string value returns empty string", func(t *testing.T) {
		payload := map[string]interface{}{"key": 123}
		result := getStringFromPayload(payload, "key")
		assert.Equal(t, "", result)
	})

	t.Run("valid string value returns string", func(t *testing.T) {
		payload := map[string]interface{}{"key": "value"}
		result := getStringFromPayload(payload, "key")
		assert.Equal(t, "value", result)
	})
}

// Test for session not accepting events (not recording/paused)
func TestEventController_SessionNotAcceptingEvents(t *testing.T) {
	controller, sessionService, authService, jwtService, cleanup := setupEventControllerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	email := uniqueEmail("not-accepting")
	user, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	// Create session and complete it
	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Stop the session (changes status to processing)
	_, err = sessionService.Stop(ctx, sess.ID, user.ID)
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)
	token := "Bearer " + tokenPair.AccessToken

	url := "https://example.com"
	req := generated.RoutesBatchEventsRequestObject{
		Id: sess.ID.String(),
		Params: generated.RoutesBatchEventsParams{
			Authorization: token,
		},
		Body: &generated.RoutesBatchEventsJSONRequestBody{
			Events: []generated.EventsEventData{
				{Type: "page_visit", Timestamp: time.Now().UnixMilli(), Url: &url},
			},
		},
	}

	resp, err := controller.RoutesBatchEvents(ctx, req)
	require.NoError(t, err)

	_, ok := resp.(generated.RoutesBatchEvents400JSONResponse)
	assert.True(t, ok, "expected 400 response for session not accepting events")
}

// Test GetEventStats empty session
func TestEventController_RoutesGetEventStats_EmptySession(t *testing.T) {
	controller, sessionService, authService, jwtService, cleanup := setupEventControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token, sessionID := createUserSessionAndToken(t, authService, sessionService, jwtService, uniqueEmail("empty-stats"))

	req := generated.RoutesGetEventStatsRequestObject{
		Id: sessionID.String(),
		Params: generated.RoutesGetEventStatsParams{
			Authorization: token,
		},
	}

	resp, err := controller.RoutesGetEventStats(ctx, req)
	require.NoError(t, err)

	successResp, ok := resp.(generated.RoutesGetEventStats200JSONResponse)
	require.True(t, ok, "expected 200 response")
	assert.Equal(t, int32(0), successResp.TotalEvents)
	assert.Equal(t, int32(0), successResp.PageVisits)
	assert.Equal(t, int32(0), successResp.Highlights)
	assert.Equal(t, int32(0), successResp.UniqueUrls)
}

// Ensure Session status checks work correctly
func TestEventController_SessionStatusValidation(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	sessionService := service.NewSessionService(client, nil)
	urlService := service.NewURLService(client)
	eventService := service.NewEventService(client, urlService)
	controller := NewEventController(eventService, sessionService, jwtService)

	ctx := context.Background()
	email := uniqueEmail("status-check")
	user, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)
	token := "Bearer " + tokenPair.AccessToken

	// Create a session directly with completed status
	sess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusCompleted).
		SetStartedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	url := "https://example.com"
	req := generated.RoutesBatchEventsRequestObject{
		Id: sess.ID.String(),
		Params: generated.RoutesBatchEventsParams{
			Authorization: token,
		},
		Body: &generated.RoutesBatchEventsJSONRequestBody{
			Events: []generated.EventsEventData{
				{Type: "page_visit", Timestamp: time.Now().UnixMilli(), Url: &url},
			},
		},
	}

	resp, err := controller.RoutesBatchEvents(ctx, req)
	require.NoError(t, err)

	_, ok := resp.(generated.RoutesBatchEvents400JSONResponse)
	assert.True(t, ok, "expected 400 response for completed session")
}
