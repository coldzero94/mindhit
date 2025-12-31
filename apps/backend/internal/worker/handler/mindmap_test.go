package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

func TestHandleMindmapGenerate_NoAIManager(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{
		client:    client,
		aiManager: nil, // No AI manager
	}

	// Create task payload
	payload, _ := json.Marshal(queue.MindmapGeneratePayload{SessionID: uuid.New().String()})
	task := asynq.NewTask(queue.TypeMindmapGenerate, payload)

	// Should return nil (skip) when AI manager is not configured
	err := h.HandleMindmapGenerate(ctx, task)
	assert.NoError(t, err)
}

func TestHandleMindmapGenerate_InvalidPayload(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with invalid payload
	task := asynq.NewTask(queue.TypeMindmapGenerate, []byte("invalid json"))

	err := h.HandleMindmapGenerate(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal payload")
}

func TestHandleMindmapGenerate_InvalidSessionID(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with invalid UUID
	payload, _ := json.Marshal(queue.MindmapGeneratePayload{SessionID: "not-a-uuid"})
	task := asynq.NewTask(queue.TypeMindmapGenerate, payload)

	err := h.HandleMindmapGenerate(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse session id")
}

func TestHandleMindmapGenerate_SessionNotFound(t *testing.T) {
	// Skip if no AI manager (which is true in this test)
	t.Skip("Requires AI manager mock")
}

func TestBuildMindmapFromRelationship(t *testing.T) {
	response := RelationshipGraphResponse{
		Core: struct {
			Label       string `json:"label"`
			Description string `json:"description"`
		}{
			Label:       "Core Theme",
			Description: "Session about programming",
		},
		Topics: []struct {
			ID          string   `json:"id"`
			Label       string   `json:"label"`
			Keywords    []string `json:"keywords"`
			Description string   `json:"description"`
			Pages       []struct {
				URLID     string  `json:"url_id"`
				Title     string  `json:"title"`
				Relevance float64 `json:"relevance"`
			} `json:"pages"`
		}{
			{
				ID:          "topic-1",
				Label:       "Go Programming",
				Keywords:    []string{"go", "golang"},
				Description: "Go language resources",
				Pages: []struct {
					URLID     string  `json:"url_id"`
					Title     string  `json:"title"`
					Relevance float64 `json:"relevance"`
				}{
					{
						URLID:     uuid.New().String(),
						Title:     "Go Tutorial",
						Relevance: 0.9,
					},
				},
			},
			{
				ID:          "topic-2",
				Label:       "Testing",
				Keywords:    []string{"test", "unit test"},
				Description: "Testing resources",
				Pages:       nil,
			},
		},
		Connections: []struct {
			From           string   `json:"from"`
			To             string   `json:"to"`
			SharedKeywords []string `json:"shared_keywords"`
			Reason         string   `json:"reason"`
		}{
			{
				From:           "topic-1",
				To:             "topic-2",
				SharedKeywords: []string{"programming"},
				Reason:         "Both about coding",
			},
		},
	}

	durationMsMap := make(map[string]int)
	for _, topic := range response.Topics {
		for _, page := range topic.Pages {
			durationMsMap[page.URLID] = 30000 // 30 seconds
		}
	}

	result := buildMindmapFromRelationship(response, durationMsMap)

	// Verify core node
	assert.NotEmpty(t, result.Nodes)
	coreNode := result.Nodes[0]
	assert.Equal(t, "core", coreNode.ID)
	assert.Equal(t, "Core Theme", coreNode.Label)
	assert.Equal(t, "core", coreNode.Type)
	assert.Equal(t, "#FFD700", coreNode.Color)

	// Verify topic nodes exist
	topicCount := 0
	for _, node := range result.Nodes {
		if node.Type == "topic" {
			topicCount++
		}
	}
	assert.Equal(t, 2, topicCount)

	// Verify edges
	assert.NotEmpty(t, result.Edges)

	// Verify layout
	assert.Equal(t, "galaxy", result.Layout.Type)
}

func TestBuildMindmapFromRelationship_EmptyTopics(t *testing.T) {
	response := RelationshipGraphResponse{
		Core: struct {
			Label       string `json:"label"`
			Description string `json:"description"`
		}{
			Label:       "Empty Session",
			Description: "No topics",
		},
		Topics:      nil,
		Connections: nil,
	}

	result := buildMindmapFromRelationship(response, make(map[string]int))

	// Should still have core node
	assert.Len(t, result.Nodes, 1)
	assert.Equal(t, "core", result.Nodes[0].ID)
	assert.Empty(t, result.Edges)
}

func TestGetTopicColor(t *testing.T) {
	// First color
	assert.Equal(t, "#3B82F6", getTopicColor(0))

	// Second color
	assert.Equal(t, "#10B981", getTopicColor(1))

	// Should wrap around
	assert.Equal(t, "#3B82F6", getTopicColor(8)) // Same as index 0
}

func TestHandleMindmapGenerate_WithSession(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Create test user
	user, err := client.User.Create().
		SetEmail("mindmap-test-" + uuid.New().String() + "@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	require.NoError(t, err)

	// Create test session in processing state
	sess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusProcessing).
		SetStartedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Create handler without AI manager (will skip processing)
	h := &handlers{
		client:    client,
		aiManager: nil,
	}

	payload, _ := json.Marshal(queue.MindmapGeneratePayload{SessionID: sess.ID.String()})
	task := asynq.NewTask(queue.TypeMindmapGenerate, payload)

	// Should skip without error since no AI manager
	err = h.HandleMindmapGenerate(ctx, task)
	assert.NoError(t, err)
}

func TestMindmapNodePositioning(t *testing.T) {
	// Test that nodes are positioned correctly in a radial layout
	response := RelationshipGraphResponse{
		Core: struct {
			Label       string `json:"label"`
			Description string `json:"description"`
		}{
			Label:       "Core",
			Description: "Test",
		},
		Topics: []struct {
			ID          string   `json:"id"`
			Label       string   `json:"label"`
			Keywords    []string `json:"keywords"`
			Description string   `json:"description"`
			Pages       []struct {
				URLID     string  `json:"url_id"`
				Title     string  `json:"title"`
				Relevance float64 `json:"relevance"`
			} `json:"pages"`
		}{
			{ID: "topic-0", Label: "Topic 0"},
			{ID: "topic-1", Label: "Topic 1"},
			{ID: "topic-2", Label: "Topic 2"},
			{ID: "topic-3", Label: "Topic 3"},
		},
	}

	result := buildMindmapFromRelationship(response, make(map[string]int))

	// Core should be at center
	assert.NotNil(t, result.Nodes[0].Position)
	assert.Equal(t, 0.0, result.Nodes[0].Position.X)
	assert.Equal(t, 0.0, result.Nodes[0].Position.Y)

	// Topics should be at radius 200 from center
	for i, node := range result.Nodes {
		if node.Type == "topic" && node.Position != nil {
			// Calculate distance from origin
			dist := node.Position.X*node.Position.X + node.Position.Y*node.Position.Y
			// Should be approximately 200^2 = 40000
			assert.InDelta(t, 40000, dist, 1, "Topic %d should be at radius 200", i)
		}
	}
}

func TestConversionFunctions(t *testing.T) {
	// Test that conversion functions exist and work
	nodes := []service.MindmapNode{
		{ID: "test", Label: "Test", Type: "core"},
	}
	edges := []service.MindmapEdge{
		{Source: "a", Target: "b", Weight: 1.0},
	}
	layout := service.MindmapLayout{
		Type:   "galaxy",
		Params: map[string]interface{}{"scale": 1.0},
	}

	nodesMap := service.ConvertNodesToMaps(nodes)
	edgesMap := service.ConvertEdgesToMaps(edges)
	layoutMap := service.ConvertLayoutToMap(layout)

	assert.Len(t, nodesMap, 1)
	assert.Len(t, edgesMap, 1)
	assert.Equal(t, "galaxy", layoutMap["type"])
}
