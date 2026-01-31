// Package controller provides HTTP handlers for API endpoints.
package controller

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/mindmapgraph"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/service"
)

// MindmapController implements mindmap-related handlers from StrictServerInterface.
type MindmapController struct {
	mindmapService *service.MindmapService
	queueClient    *queue.Client
}

// NewMindmapController creates a new MindmapController.
func NewMindmapController(mindmapService *service.MindmapService, queueClient *queue.Client) *MindmapController {
	return &MindmapController{
		mindmapService: mindmapService,
		queueClient:    queueClient,
	}
}

// MindmapRoutesGetMindmap handles GET /v1/sessions/{sessionId}/mindmap.
func (c *MindmapController) MindmapRoutesGetMindmap(ctx context.Context, request generated.MindmapRoutesGetMindmapRequestObject) (generated.MindmapRoutesGetMindmapResponseObject, error) {
	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.MindmapRoutesGetMindmap404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	mindmap, err := c.mindmapService.GetBySessionID(ctx, sessionID)
	if err != nil {
		return c.handleGetError(err)
	}

	return generated.MindmapRoutesGetMindmap200JSONResponse{
		Mindmap: mapMindmap(mindmap, sessionID),
	}, nil
}

// MindmapRoutesGenerateMindmap handles POST /v1/sessions/{sessionId}/mindmap/generate.
func (c *MindmapController) MindmapRoutesGenerateMindmap(ctx context.Context, request generated.MindmapRoutesGenerateMindmapRequestObject) (generated.MindmapRoutesGenerateMindmapResponseObject, error) {
	sessionID, err := uuid.Parse(request.Id)
	if err != nil {
		return generated.MindmapRoutesGenerateMindmap404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid session id",
			},
		}, nil
	}

	// Get or create mindmap
	mindmap, created, err := c.mindmapService.GetOrCreateForSession(ctx, sessionID)
	if err != nil {
		return c.handleGenerateError(err)
	}

	// If force is true and mindmap already exists, update to pending status
	if !created && request.Body.Force != nil && *request.Body.Force {
		mindmap, err = c.mindmapService.UpdateStatus(ctx, mindmap.ID, mindmapgraph.StatusPending)
		if err != nil {
			slog.Error("failed to reset mindmap status", "error", err)
			return nil, err
		}
	}

	// Enqueue mindmap generation job if status is pending
	if mindmap.Status == mindmapgraph.StatusPending && c.queueClient != nil {
		task, err := queue.NewMindmapGenerateTask(sessionID.String())
		if err != nil {
			slog.Error("failed to create mindmap task", "error", err)
		} else {
			_, err = c.queueClient.Enqueue(task, asynq.MaxRetry(3))
			if err != nil {
				slog.Error("failed to enqueue mindmap task", "error", err)
			} else {
				slog.Info("enqueued mindmap generation", "session_id", sessionID.String())
			}
		}
	}

	return generated.MindmapRoutesGenerateMindmap202JSONResponse{
		Mindmap: mapMindmap(mindmap, sessionID),
	}, nil
}

func (c *MindmapController) handleGetError(err error) (generated.MindmapRoutesGetMindmapResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.MindmapRoutesGetMindmap404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrMindmapNotFound):
		return generated.MindmapRoutesGetMindmap404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "mindmap not found for this session",
			},
		}, nil
	default:
		slog.Error("mindmap get failed", "error", err)
		return nil, err
	}
}

func (c *MindmapController) handleGenerateError(err error) (generated.MindmapRoutesGenerateMindmapResponseObject, error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		return generated.MindmapRoutesGenerateMindmap404JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session not found",
			},
		}, nil
	case errors.Is(err, service.ErrSessionNotReady):
		return generated.MindmapRoutesGenerateMindmap400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "session is not ready for mindmap generation",
			},
		}, nil
	default:
		slog.Error("mindmap generate failed", "error", err)
		return nil, err
	}
}

// mapMindmap converts an ent.MindmapGraph to generated.MindmapMindmap.
func mapMindmap(m *ent.MindmapGraph, sessionID uuid.UUID) generated.MindmapMindmap {
	result := generated.MindmapMindmap{
		Id:        m.ID.String(),
		SessionId: sessionID.String(),
		Status:    generated.MindmapMindmapStatus(m.Status.String()),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.ErrorMessage != nil {
		result.ErrorMessage = m.ErrorMessage
	}

	// Only include data if mindmap is completed
	if m.Status == mindmapgraph.StatusCompleted && len(m.Nodes) > 0 {
		data := &generated.MindmapMindmapData{
			Nodes:  mapNodes(m.Nodes),
			Edges:  mapEdges(m.GraphEdges),
			Layout: mapLayout(m.Layout),
		}
		result.Data = data
	}

	return result
}

func mapNodes(nodes []map[string]interface{}) []generated.MindmapMindmapNode {
	result := make([]generated.MindmapMindmapNode, len(nodes))
	for i, node := range nodes {
		result[i] = generated.MindmapMindmapNode{
			Id:    getString(node, "id"),
			Label: getString(node, "label"),
			Type:  getString(node, "type"),
			Size:  getFloat64(node, "size"),
			Color: getString(node, "color"),
		}
		if pos, ok := node["position"].(map[string]interface{}); ok {
			result[i].Position = &generated.MindmapPosition{
				X: getFloat64(pos, "x"),
				Y: getFloat64(pos, "y"),
				Z: getFloat64(pos, "z"),
			}
		}
		if data, ok := node["data"].(map[string]interface{}); ok {
			result[i].Data = &data
		}
	}
	return result
}

func mapEdges(edges []map[string]interface{}) []generated.MindmapMindmapEdge {
	result := make([]generated.MindmapMindmapEdge, len(edges))
	for i, edge := range edges {
		result[i] = generated.MindmapMindmapEdge{
			Source: getString(edge, "source"),
			Target: getString(edge, "target"),
			Weight: getFloat64(edge, "weight"),
		}
		if label := getString(edge, "label"); label != "" {
			result[i].Label = &label
		}
	}
	return result
}

func mapLayout(layout map[string]interface{}) generated.MindmapMindmapLayout {
	result := generated.MindmapMindmapLayout{
		Type: getString(layout, "type"),
	}
	if params, ok := layout["params"].(map[string]interface{}); ok {
		result.Params = &params
	}
	return result
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}
