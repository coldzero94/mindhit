// Package service provides business logic services.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/mindmapgraph"
	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/ent/user"
)

// Mindmap service errors.
var (
	ErrMindmapNotFound = errors.New("mindmap not found")
	ErrSessionNotReady = errors.New("session not ready for mindmap generation")
)

// MindmapService handles mindmap operations.
type MindmapService struct {
	client *ent.Client
}

// NewMindmapService creates a new MindmapService.
func NewMindmapService(client *ent.Client) *MindmapService {
	return &MindmapService{client: client}
}

// GetBySessionID retrieves a mindmap for a session.
func (s *MindmapService) GetBySessionID(ctx context.Context, sessionID, userID uuid.UUID) (*ent.MindmapGraph, error) {
	// First verify the session belongs to the user
	sess, err := s.client.Session.Query().
		Where(
			session.ID(sessionID),
			session.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	// Get the mindmap for this session
	mindmap, err := s.client.MindmapGraph.Query().
		Where(mindmapgraph.HasSessionWith(session.ID(sess.ID))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrMindmapNotFound
		}
		return nil, err
	}

	return mindmap, nil
}

// CreatePending creates a new mindmap in pending state.
func (s *MindmapService) CreatePending(ctx context.Context, sessionID uuid.UUID) (*ent.MindmapGraph, error) {
	mindmap, err := s.client.MindmapGraph.Create().
		SetSessionID(sessionID).
		SetStatus(mindmapgraph.StatusPending).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mindmap, nil
}

// UpdateStatus updates the mindmap status.
func (s *MindmapService) UpdateStatus(ctx context.Context, mindmapID uuid.UUID, status mindmapgraph.Status) (*ent.MindmapGraph, error) {
	return s.client.MindmapGraph.UpdateOneID(mindmapID).
		SetStatus(status).
		Save(ctx)
}

// SetFailed marks a mindmap as failed with an error message.
func (s *MindmapService) SetFailed(ctx context.Context, mindmapID uuid.UUID, errMsg string) (*ent.MindmapGraph, error) {
	return s.client.MindmapGraph.UpdateOneID(mindmapID).
		SetStatus(mindmapgraph.StatusFailed).
		SetErrorMessage(errMsg).
		Save(ctx)
}

// SetCompleted marks a mindmap as completed with data.
func (s *MindmapService) SetCompleted(ctx context.Context, mindmapID uuid.UUID, data MindmapData) (*ent.MindmapGraph, error) {
	return s.client.MindmapGraph.UpdateOneID(mindmapID).
		SetStatus(mindmapgraph.StatusCompleted).
		SetNodes(ConvertNodesToMaps(data.Nodes)).
		SetGraphEdges(ConvertEdgesToMaps(data.Edges)).
		SetLayout(ConvertLayoutToMap(data.Layout)).
		Save(ctx)
}

// GetOrCreateForSession gets existing mindmap or creates a new pending one.
func (s *MindmapService) GetOrCreateForSession(ctx context.Context, sessionID, userID uuid.UUID) (*ent.MindmapGraph, bool, error) {
	// Verify session ownership
	sess, err := s.client.Session.Query().
		Where(
			session.ID(sessionID),
			session.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false, ErrSessionNotFound
		}
		return nil, false, err
	}

	// Try to get existing mindmap
	mindmap, err := s.client.MindmapGraph.Query().
		Where(mindmapgraph.HasSessionWith(session.ID(sess.ID))).
		Only(ctx)
	if err == nil {
		return mindmap, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, err
	}

	// Create new mindmap
	mindmap, err = s.CreatePending(ctx, sessionID)
	if err != nil {
		return nil, false, err
	}

	return mindmap, true, nil
}
