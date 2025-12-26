// Package handler provides worker job handlers.
package handler

import (
	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/infrastructure/queue"
)

// RegisterHandlers registers all job handlers with the queue server.
func RegisterHandlers(server *queue.Server, client *ent.Client) {
	h := &handlers{client: client}

	server.HandleFunc(queue.TypeSessionProcess, h.HandleSessionProcess)
	server.HandleFunc(queue.TypeSessionCleanup, h.HandleSessionCleanup)
	// AI handlers are registered in Phase 9
	// server.HandleFunc(queue.TypeURLSummarize, h.HandleURLSummarize)
	// server.HandleFunc(queue.TypeMindmapGenerate, h.HandleMindmapGenerate)
}

type handlers struct {
	client *ent.Client
	// AI services will be added in Phase 9
}
