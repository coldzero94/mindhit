package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/infrastructure/ai"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/service"
)

const relationshipGraphPrompt = `Analyze the pages and tags from this browsing session to create a relationship graph.

## Session Data

### Visited Pages (URL + keywords + summary)

%s

### Highlights (user-selected text)

%s

## Requirements

1. **Core theme (core)**: One central theme spanning the entire session
2. **Main topics (topics)**: 3-5 groups based on common keywords
3. **Page connections**: Map pages to their relevant topics
4. **Topic connections**: Relationships between topics with overlapping keywords

## Respond in JSON format

{
  "core": {
    "label": "Core theme (Korean)",
    "description": "Session summary (1-2 sentences)"
  },
  "topics": [
    {
      "id": "topic-1",
      "label": "Topic name (Korean)",
      "keywords": ["related", "keywords"],
      "description": "Topic description",
      "pages": [
        {
          "url_id": "uuid",
          "title": "Page title",
          "relevance": 0.9
        }
      ]
    }
  ],
  "connections": [
    {
      "from": "topic-1",
      "to": "topic-2",
      "shared_keywords": ["shared keyword"],
      "reason": "Connection reason"
    }
  ]
}`

// RelationshipGraphResponse represents the AI response structure.
type RelationshipGraphResponse struct {
	Core struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	} `json:"core"`
	Topics []struct {
		ID          string   `json:"id"`
		Label       string   `json:"label"`
		Keywords    []string `json:"keywords"`
		Description string   `json:"description"`
		Pages       []struct {
			URLID     string  `json:"url_id"`
			Title     string  `json:"title"`
			Relevance float64 `json:"relevance"`
		} `json:"pages"`
	} `json:"topics"`
	Connections []struct {
		From           string   `json:"from"`
		To             string   `json:"to"`
		SharedKeywords []string `json:"shared_keywords"`
		Reason         string   `json:"reason"`
	} `json:"connections"`
}

// HandleMindmapGenerate processes mindmap generation for a session.
func (h *handlers) HandleMindmapGenerate(ctx context.Context, t *asynq.Task) error {
	var payload queue.MindmapGeneratePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	sessionID, err := uuid.Parse(payload.SessionID)
	if err != nil {
		return fmt.Errorf("parse session id: %w", err)
	}

	slog.Info("generating mindmap", "session_id", payload.SessionID)

	// Check if AI manager is available
	if h.aiManager == nil {
		slog.Warn("ai manager not configured, skipping mindmap generation")
		return nil
	}

	// Get session with all related data
	sess, err := h.client.Session.
		Query().
		Where(session.IDEQ(sessionID)).
		WithPageVisits(func(q *ent.PageVisitQuery) {
			q.WithURL()
		}).
		WithHighlights().
		WithUser().
		Only(ctx)

	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	// Build page data with keywords
	var pageData strings.Builder
	durationMsMap := make(map[string]int)

	for _, pv := range sess.Edges.PageVisits {
		if pv.Edges.URL == nil {
			continue
		}
		u := pv.Edges.URL

		durationMs := 0
		if pv.DurationMs != nil {
			durationMs = *pv.DurationMs
		}
		durationMsMap[u.ID.String()] = durationMs

		pageData.WriteString(fmt.Sprintf(`
- ID: %s
  Title: %s
  URL: %s
  Keywords: [%s]
  Summary: %s
  Duration: %dms
`,
			u.ID.String(),
			u.Title,
			u.URL,
			strings.Join(u.Keywords, ", "),
			u.Summary,
			durationMs,
		))
	}

	// Build highlights text
	var highlights strings.Builder
	if len(sess.Edges.Highlights) > 0 {
		for _, hl := range sess.Edges.Highlights {
			highlights.WriteString(fmt.Sprintf("- \"%s\"\n", hl.Text))
		}
	} else {
		highlights.WriteString("(No highlights)")
	}

	// Get user ID for metadata and usage tracking
	var userID uuid.UUID
	if sess.Edges.User != nil {
		userID = sess.Edges.User.ID
	}

	// Check usage limit before AI call
	if h.usageService != nil && userID != uuid.Nil {
		status, err := h.usageService.CheckLimit(ctx, userID)
		if err != nil {
			slog.Warn("failed to check usage limit", "error", err)
		} else if !status.CanUseAI {
			slog.Warn("user token limit exceeded, skipping mindmap generation",
				"user_id", userID,
				"tokens_used", status.TokensUsed,
				"token_limit", status.TokenLimit,
			)
			return fmt.Errorf("token limit exceeded: used %d/%d", status.TokensUsed, status.TokenLimit)
		}
	}

	// Generate relationship graph using AI
	req := ai.ChatRequest{
		UserPrompt: fmt.Sprintf(relationshipGraphPrompt, pageData.String(), highlights.String()),
		Options: ai.ChatOptions{
			MaxTokens: 4096,
			JSONMode:  true,
		},
		Metadata: map[string]string{
			"session_id": sessionID.String(),
			"user_id":    userID.String(),
		},
	}

	response, err := h.aiManager.Chat(ctx, ai.TaskMindmap, req)
	if err != nil {
		return fmt.Errorf("ai generate mindmap: %w", err)
	}

	// Record token usage
	if h.usageService != nil && userID != uuid.Nil {
		if err := h.usageService.RecordUsage(ctx, service.UsageRecord{
			UserID:    userID,
			SessionID: sessionID,
			Operation: "mindmap",
			Tokens:    response.TotalTokens,
			AIModel:   response.Model,
		}); err != nil {
			slog.Error("failed to record usage", "error", err)
		}
	}

	var aiResp RelationshipGraphResponse
	if err := json.Unmarshal([]byte(response.Content), &aiResp); err != nil {
		return fmt.Errorf("parse ai response: %w", err)
	}

	// Convert AI response to mindmap data
	mindmapData := buildMindmapFromRelationship(aiResp, durationMsMap)

	// Convert to storage format ([]map[string]interface{})
	nodesData := service.ConvertNodesToMaps(mindmapData.Nodes)
	edgesData := service.ConvertEdgesToMaps(mindmapData.Edges)
	layoutData := service.ConvertLayoutToMap(mindmapData.Layout)

	// Save mindmap to database
	_, err = h.client.MindmapGraph.
		Create().
		SetSessionID(sessionID).
		SetNodes(nodesData).
		SetGraphEdges(edgesData).
		SetLayout(layoutData).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("save mindmap: %w", err)
	}

	// Update session status to completed
	_, err = h.client.Session.
		UpdateOneID(sessionID).
		SetSessionStatus(session.SessionStatusCompleted).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}

	slog.Info("mindmap generated",
		"session_id", payload.SessionID,
		"topics", len(aiResp.Topics),
		"connections", len(aiResp.Connections),
		"provider", response.Provider,
		"tokens", response.TotalTokens,
	)
	return nil
}

func buildMindmapFromRelationship(
	resp RelationshipGraphResponse,
	durationMsMap map[string]int,
) service.MindmapData {
	var nodes []service.MindmapNode
	var edges []service.MindmapEdge

	// Core node
	coreID := "core"
	nodes = append(nodes, service.MindmapNode{
		ID:       coreID,
		Label:    resp.Core.Label,
		Type:     "core",
		Size:     100,
		Color:    "#FFD700",
		Position: &service.Position{X: 0, Y: 0, Z: 0},
		Data: map[string]interface{}{
			"description": resp.Core.Description,
		},
	})

	// Topic nodes
	topicCount := len(resp.Topics)
	for i, topic := range resp.Topics {
		topicID := topic.ID
		if topicID == "" {
			topicID = fmt.Sprintf("topic-%d", i)
		}

		angle := (float64(i) / float64(topicCount)) * 2 * math.Pi
		radius := 200.0
		topicSize := 40.0 + float64(len(topic.Pages))*10
		if topicSize > 80 {
			topicSize = 80
		}

		nodes = append(nodes, service.MindmapNode{
			ID:    topicID,
			Label: topic.Label,
			Type:  "topic",
			Size:  topicSize,
			Color: getTopicColor(i),
			Position: &service.Position{
				X: radius * math.Cos(angle),
				Y: radius * math.Sin(angle),
				Z: 0,
			},
			Data: map[string]interface{}{
				"description": topic.Description,
				"keywords":    topic.Keywords,
			},
		})

		edges = append(edges, service.MindmapEdge{
			Source: coreID,
			Target: topicID,
			Weight: 1.0,
		})

		// Page nodes
		for j, page := range topic.Pages {
			pageID := page.URLID
			subAngle := angle + (float64(j)-float64(len(topic.Pages))/2)*0.4
			subRadius := 60.0 + float64(j)*15

			size := 15.0
			if durationMs, ok := durationMsMap[page.URLID]; ok {
				size = math.Min(40, 15+float64(durationMs)/20000)
			}
			size *= (0.5 + page.Relevance*0.5)

			nodes = append(nodes, service.MindmapNode{
				ID:    pageID,
				Label: page.Title,
				Type:  "page",
				Size:  size,
				Color: getTopicColor(i),
				Position: &service.Position{
					X: radius*math.Cos(angle) + subRadius*math.Cos(subAngle),
					Y: radius*math.Sin(angle) + subRadius*math.Sin(subAngle),
					Z: 0,
				},
				Data: map[string]interface{}{
					"url_id":    page.URLID,
					"relevance": page.Relevance,
				},
			})

			edges = append(edges, service.MindmapEdge{
				Source: topicID,
				Target: pageID,
				Weight: page.Relevance,
			})
		}
	}

	// Cross-topic connections
	for _, conn := range resp.Connections {
		edges = append(edges, service.MindmapEdge{
			Source: conn.From,
			Target: conn.To,
			Weight: float64(len(conn.SharedKeywords)) * 0.2,
			Label:  conn.Reason,
		})
	}

	return service.MindmapData{
		Nodes: nodes,
		Edges: edges,
		Layout: service.MindmapLayout{
			Type: "galaxy",
			Params: map[string]interface{}{
				"center": []float64{0, 0, 0},
				"scale":  1.0,
			},
		},
	}
}

func getTopicColor(index int) string {
	colors := []string{
		"#3B82F6", "#10B981", "#F59E0B", "#EF4444",
		"#8B5CF6", "#EC4899", "#14B8A6", "#F97316",
	}
	return colors[index%len(colors)]
}
