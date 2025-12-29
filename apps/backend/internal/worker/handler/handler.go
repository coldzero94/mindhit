// Package handler provides worker job handlers.
package handler

import (
	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/infrastructure/ai"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/service"
)

// RegisterHandlers registers all job handlers with the queue server.
func RegisterHandlers(
	server *queue.Server,
	client *ent.Client,
	aiManager *ai.ProviderManager,
	usageService *service.UsageService,
) {
	h := &handlers{
		client:       client,
		aiManager:    aiManager,
		usageService: usageService,
	}

	server.HandleFunc(queue.TypeSessionProcess, h.HandleSessionProcess)
	server.HandleFunc(queue.TypeSessionCleanup, h.HandleSessionCleanup)
	server.HandleFunc(queue.TypeURLTagExtraction, h.HandleURLTagExtraction)
	server.HandleFunc(queue.TypeMindmapGenerate, h.HandleMindmapGenerate)
}

type handlers struct {
	client       *ent.Client
	aiManager    *ai.ProviderManager
	usageService *service.UsageService
}
