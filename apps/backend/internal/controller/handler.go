package controller

import (
	"context"

	"github.com/mindhit/api/internal/generated"
)

// Handler combines all controllers to implement StrictServerInterface
type Handler struct {
	*AuthController
	*SessionController
	*EventController
	*SubscriptionController
	*UsageController
	*OAuthController
}

// NewHandler creates a new Handler with all controllers
func NewHandler(
	auth *AuthController,
	session *SessionController,
	event *EventController,
	subscription *SubscriptionController,
	usage *UsageController,
	oauth *OAuthController,
) *Handler {
	return &Handler{
		AuthController:         auth,
		SessionController:      session,
		EventController:        event,
		SubscriptionController: subscription,
		UsageController:        usage,
		OAuthController:        oauth,
	}
}

// Ensure Handler implements StrictServerInterface at compile time
var _ generated.StrictServerInterface = (*Handler)(nil)

// Auth handlers are embedded from AuthController

// Session handlers are embedded from SessionController

// RoutesForgotPassword delegates to AuthController
func (h *Handler) RoutesForgotPassword(ctx context.Context, request generated.RoutesForgotPasswordRequestObject) (generated.RoutesForgotPasswordResponseObject, error) {
	return h.AuthController.RoutesForgotPassword(ctx, request)
}

// RoutesLogin delegates to AuthController
func (h *Handler) RoutesLogin(ctx context.Context, request generated.RoutesLoginRequestObject) (generated.RoutesLoginResponseObject, error) {
	return h.AuthController.RoutesLogin(ctx, request)
}

// RoutesLogout delegates to AuthController
func (h *Handler) RoutesLogout(ctx context.Context, request generated.RoutesLogoutRequestObject) (generated.RoutesLogoutResponseObject, error) {
	return h.AuthController.RoutesLogout(ctx, request)
}

// RoutesMe delegates to AuthController
func (h *Handler) RoutesMe(ctx context.Context, request generated.RoutesMeRequestObject) (generated.RoutesMeResponseObject, error) {
	return h.AuthController.RoutesMe(ctx, request)
}

// RoutesRefresh delegates to AuthController
func (h *Handler) RoutesRefresh(ctx context.Context, request generated.RoutesRefreshRequestObject) (generated.RoutesRefreshResponseObject, error) {
	return h.AuthController.RoutesRefresh(ctx, request)
}

// RoutesResetPassword delegates to AuthController
func (h *Handler) RoutesResetPassword(ctx context.Context, request generated.RoutesResetPasswordRequestObject) (generated.RoutesResetPasswordResponseObject, error) {
	return h.AuthController.RoutesResetPassword(ctx, request)
}

// RoutesSignup delegates to AuthController
func (h *Handler) RoutesSignup(ctx context.Context, request generated.RoutesSignupRequestObject) (generated.RoutesSignupResponseObject, error) {
	return h.AuthController.RoutesSignup(ctx, request)
}

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

// SubscriptionRoutesGetSubscription delegates to SubscriptionController
func (h *Handler) SubscriptionRoutesGetSubscription(ctx context.Context, request generated.SubscriptionRoutesGetSubscriptionRequestObject) (generated.SubscriptionRoutesGetSubscriptionResponseObject, error) {
	return h.SubscriptionController.SubscriptionRoutesGetSubscription(ctx, request)
}

// SubscriptionRoutesListPlans delegates to SubscriptionController
func (h *Handler) SubscriptionRoutesListPlans(ctx context.Context, request generated.SubscriptionRoutesListPlansRequestObject) (generated.SubscriptionRoutesListPlansResponseObject, error) {
	return h.SubscriptionController.SubscriptionRoutesListPlans(ctx, request)
}

// UsageRoutesGetUsage delegates to UsageController
func (h *Handler) UsageRoutesGetUsage(ctx context.Context, request generated.UsageRoutesGetUsageRequestObject) (generated.UsageRoutesGetUsageResponseObject, error) {
	return h.UsageController.UsageRoutesGetUsage(ctx, request)
}

// UsageRoutesGetUsageHistory delegates to UsageController
func (h *Handler) UsageRoutesGetUsageHistory(ctx context.Context, request generated.UsageRoutesGetUsageHistoryRequestObject) (generated.UsageRoutesGetUsageHistoryResponseObject, error) {
	return h.UsageController.UsageRoutesGetUsageHistory(ctx, request)
}

// RoutesGoogleAuth delegates to OAuthController
func (h *Handler) RoutesGoogleAuth(ctx context.Context, request generated.RoutesGoogleAuthRequestObject) (generated.RoutesGoogleAuthResponseObject, error) {
	return h.OAuthController.RoutesGoogleAuth(ctx, request)
}
