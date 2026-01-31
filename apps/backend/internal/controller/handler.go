package controller

import (
	"context"

	"github.com/mindhit/api/internal/generated"
)

// Handler combines all controllers to implement StrictServerInterface.
type Handler struct {
	*SessionController
	*EventController
	*MindmapController
}

// NewHandler creates a new Handler with all controllers.
func NewHandler(
	session *SessionController,
	event *EventController,
	mindmap *MindmapController,
) *Handler {
	return &Handler{
		SessionController: session,
		EventController:   event,
		MindmapController: mindmap,
	}
}

// Ensure Handler implements StrictServerInterface at compile time
var _ generated.StrictServerInterface = (*Handler)(nil)

// RoutesList delegates to SessionController
func (h *Handler) RoutesList(ctx context.Context, request generated.RoutesListRequestObject) (generated.RoutesListResponseObject, error) {
	return h.SessionController.RoutesList(ctx, request)
}

// RoutesStart delegates to SessionController
func (h *Handler) RoutesStart(ctx context.Context, request generated.RoutesStartRequestObject) (generated.RoutesStartResponseObject, error) {
	return h.SessionController.RoutesStart(ctx, request)
}

// RoutesDelete delegates to SessionController
func (h *Handler) RoutesDelete(ctx context.Context, request generated.RoutesDeleteRequestObject) (generated.RoutesDeleteResponseObject, error) {
	return h.SessionController.RoutesDelete(ctx, request)
}

// RoutesGet delegates to SessionController
func (h *Handler) RoutesGet(ctx context.Context, request generated.RoutesGetRequestObject) (generated.RoutesGetResponseObject, error) {
	return h.SessionController.RoutesGet(ctx, request)
}

// RoutesUpdate delegates to SessionController
func (h *Handler) RoutesUpdate(ctx context.Context, request generated.RoutesUpdateRequestObject) (generated.RoutesUpdateResponseObject, error) {
	return h.SessionController.RoutesUpdate(ctx, request)
}

// RoutesPause delegates to SessionController
func (h *Handler) RoutesPause(ctx context.Context, request generated.RoutesPauseRequestObject) (generated.RoutesPauseResponseObject, error) {
	return h.SessionController.RoutesPause(ctx, request)
}

// RoutesResume delegates to SessionController
func (h *Handler) RoutesResume(ctx context.Context, request generated.RoutesResumeRequestObject) (generated.RoutesResumeResponseObject, error) {
	return h.SessionController.RoutesResume(ctx, request)
}

// RoutesStop delegates to SessionController
func (h *Handler) RoutesStop(ctx context.Context, request generated.RoutesStopRequestObject) (generated.RoutesStopResponseObject, error) {
	return h.SessionController.RoutesStop(ctx, request)
}

// RoutesBatchEvents delegates to EventController
func (h *Handler) RoutesBatchEvents(ctx context.Context, request generated.RoutesBatchEventsRequestObject) (generated.RoutesBatchEventsResponseObject, error) {
	return h.EventController.RoutesBatchEvents(ctx, request)
}

// RoutesListEvents delegates to EventController
func (h *Handler) RoutesListEvents(ctx context.Context, request generated.RoutesListEventsRequestObject) (generated.RoutesListEventsResponseObject, error) {
	return h.EventController.RoutesListEvents(ctx, request)
}

// RoutesGetEventStats delegates to EventController
func (h *Handler) RoutesGetEventStats(ctx context.Context, request generated.RoutesGetEventStatsRequestObject) (generated.RoutesGetEventStatsResponseObject, error) {
	return h.EventController.RoutesGetEventStats(ctx, request)
}

// MindmapRoutesGetMindmap delegates to MindmapController
func (h *Handler) MindmapRoutesGetMindmap(ctx context.Context, request generated.MindmapRoutesGetMindmapRequestObject) (generated.MindmapRoutesGetMindmapResponseObject, error) {
	return h.MindmapController.MindmapRoutesGetMindmap(ctx, request)
}

// MindmapRoutesGenerateMindmap delegates to MindmapController
func (h *Handler) MindmapRoutesGenerateMindmap(ctx context.Context, request generated.MindmapRoutesGenerateMindmapRequestObject) (generated.MindmapRoutesGenerateMindmapResponseObject, error) {
	return h.MindmapController.MindmapRoutesGenerateMindmap(ctx, request)
}
