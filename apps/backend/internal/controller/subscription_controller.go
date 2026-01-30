// Package controller provides HTTP handlers for the API.
package controller

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
)

// SubscriptionController implements subscription-related handlers from StrictServerInterface.
type SubscriptionController struct {
	subscriptionService *service.SubscriptionService
	jwtService          *service.JWTService
}

// NewSubscriptionController creates a new SubscriptionController.
func NewSubscriptionController(subscriptionService *service.SubscriptionService, jwtService *service.JWTService) *SubscriptionController {
	return &SubscriptionController{
		subscriptionService: subscriptionService,
		jwtService:          jwtService,
	}
}

// extractUserID extracts and validates user ID from authorization header.
func (c *SubscriptionController) extractUserID(authHeader string) (uuid.UUID, error) {
	if authHeader == "" {
		return uuid.Nil, errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return uuid.Nil, errors.New("invalid authorization header format")
	}

	claims, err := c.jwtService.ValidateAccessToken(parts[1])
	if err != nil {
		return uuid.Nil, errors.New("invalid or expired access token")
	}

	return claims.UserID, nil
}

// SubscriptionRoutesGetSubscription handles GET /v1/subscription.
func (c *SubscriptionController) SubscriptionRoutesGetSubscription(ctx context.Context, request generated.SubscriptionRoutesGetSubscriptionRequestObject) (generated.SubscriptionRoutesGetSubscriptionResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.SubscriptionRoutesGetSubscription401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	subInfo, err := c.subscriptionService.GetSubscriptionInfo(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get subscription info", "error", err, "user_id", userID)
		return nil, err
	}

	return generated.SubscriptionRoutesGetSubscription200JSONResponse{
		Subscription: mapSubscriptionInfo(subInfo),
	}, nil
}

// SubscriptionRoutesListPlans handles GET /v1/subscription/plans.
func (c *SubscriptionController) SubscriptionRoutesListPlans(ctx context.Context, request generated.SubscriptionRoutesListPlansRequestObject) (generated.SubscriptionRoutesListPlansResponseObject, error) {
	_, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.SubscriptionRoutesListPlans401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	plans, err := c.subscriptionService.GetAvailablePlans(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get available plans", "error", err)
		return nil, err
	}

	apiPlans := make([]generated.SubscriptionPlan, 0, len(plans))
	for _, p := range plans {
		apiPlans = append(apiPlans, mapEntPlan(p))
	}

	return generated.SubscriptionRoutesListPlans200JSONResponse{
		Plans: apiPlans,
	}, nil
}

// mapSubscriptionInfo maps service.SubscriptionInfo to generated.SubscriptionSubscriptionInfo.
func mapSubscriptionInfo(info *service.SubscriptionInfo) generated.SubscriptionSubscriptionInfo {
	result := generated.SubscriptionSubscriptionInfo{
		Id:                 info.ID.String(),
		Status:             info.Status,
		CurrentPeriodStart: info.CurrentPeriodStart,
		CurrentPeriodEnd:   info.CurrentPeriodEnd,
		CancelAtPeriodEnd:  info.CancelAtPeriodEnd,
	}

	if info.Plan != nil {
		result.Plan = generated.SubscriptionPlan{
			Id:            info.Plan.ID,
			Name:          info.Plan.Name,
			PriceCents:    int32(info.Plan.PriceCents),
			BillingPeriod: info.Plan.BillingPeriod,
			Features:      info.Plan.Features,
		}
		if info.Plan.TokenLimit != nil {
			limit := int32(*info.Plan.TokenLimit)
			result.Plan.TokenLimit = &limit
		}
		if info.Plan.SessionRetentionDays != nil {
			days := int32(*info.Plan.SessionRetentionDays)
			result.Plan.SessionRetentionDays = &days
		}
		if info.Plan.MaxConcurrentSessions != nil {
			maxSessions := int32(*info.Plan.MaxConcurrentSessions)
			result.Plan.MaxConcurrentSessions = &maxSessions
		}
	}

	return result
}

// mapEntPlan maps ent.Plan to generated.SubscriptionPlan.
func mapEntPlan(p *ent.Plan) generated.SubscriptionPlan {
	result := generated.SubscriptionPlan{
		Id:            p.ID,
		Name:          p.Name,
		PriceCents:    int32(p.PriceCents),
		BillingPeriod: p.BillingPeriod,
		Features:      p.Features,
	}

	if p.TokenLimit != nil {
		limit := int32(*p.TokenLimit)
		result.TokenLimit = &limit
	}
	if p.SessionRetentionDays != nil {
		days := int32(*p.SessionRetentionDays)
		result.SessionRetentionDays = &days
	}
	if p.MaxConcurrentSessions != nil {
		maxSessions := int32(*p.MaxConcurrentSessions)
		result.MaxConcurrentSessions = &maxSessions
	}

	return result
}

// SubscriptionRoutesChangePlan handles POST /v1/subscription/change.
func (c *SubscriptionController) SubscriptionRoutesChangePlan(ctx context.Context, request generated.SubscriptionRoutesChangePlanRequestObject) (generated.SubscriptionRoutesChangePlanResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.SubscriptionRoutesChangePlan401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	if request.Body == nil || request.Body.PlanId == "" {
		return generated.SubscriptionRoutesChangePlan400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "plan_id is required",
			},
		}, nil
	}

	sub, err := c.subscriptionService.ChangePlan(ctx, userID, request.Body.PlanId)
	if err != nil {
		if errors.Is(err, service.ErrPlanNotFound) {
			return generated.SubscriptionRoutesChangePlan404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "plan not found",
				},
			}, nil
		}
		slog.ErrorContext(ctx, "failed to change plan", "error", err, "user_id", userID)
		return nil, err
	}

	// Load plan edge for response
	plan, _ := sub.QueryPlan().Only(ctx)

	subInfo := &service.SubscriptionInfo{
		ID:                 sub.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
	}
	if plan != nil {
		subInfo.Plan = &service.PlanInfo{
			ID:                    plan.ID,
			Name:                  plan.Name,
			PriceCents:            plan.PriceCents,
			BillingPeriod:         plan.BillingPeriod,
			TokenLimit:            plan.TokenLimit,
			SessionRetentionDays:  plan.SessionRetentionDays,
			MaxConcurrentSessions: plan.MaxConcurrentSessions,
			Features:              plan.Features,
		}
	}

	return generated.SubscriptionRoutesChangePlan200JSONResponse{
		Subscription: mapSubscriptionInfo(subInfo),
		Message:      "플랜이 변경되었습니다",
	}, nil
}

// SubscriptionRoutesCancelSubscription handles POST /v1/subscription/cancel.
func (c *SubscriptionController) SubscriptionRoutesCancelSubscription(ctx context.Context, request generated.SubscriptionRoutesCancelSubscriptionRequestObject) (generated.SubscriptionRoutesCancelSubscriptionResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.SubscriptionRoutesCancelSubscription401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sub, err := c.subscriptionService.CancelSubscription(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to cancel subscription", "error", err, "user_id", userID)
		return nil, err
	}

	// Load plan edge for response
	plan, _ := sub.QueryPlan().Only(ctx)

	subInfo := &service.SubscriptionInfo{
		ID:                 sub.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
	}
	if plan != nil {
		subInfo.Plan = &service.PlanInfo{
			ID:                    plan.ID,
			Name:                  plan.Name,
			PriceCents:            plan.PriceCents,
			BillingPeriod:         plan.BillingPeriod,
			TokenLimit:            plan.TokenLimit,
			SessionRetentionDays:  plan.SessionRetentionDays,
			MaxConcurrentSessions: plan.MaxConcurrentSessions,
			Features:              plan.Features,
		}
	}

	return generated.SubscriptionRoutesCancelSubscription200JSONResponse{
		Subscription: mapSubscriptionInfo(subInfo),
		Message:      "구독이 취소 예정으로 설정되었습니다. 현재 기간 종료 후 Free 플랜으로 전환됩니다.",
	}, nil
}

// SubscriptionRoutesReactivateSubscription handles POST /v1/subscription/reactivate.
func (c *SubscriptionController) SubscriptionRoutesReactivateSubscription(ctx context.Context, request generated.SubscriptionRoutesReactivateSubscriptionRequestObject) (generated.SubscriptionRoutesReactivateSubscriptionResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.SubscriptionRoutesReactivateSubscription401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	sub, err := c.subscriptionService.ReactivateSubscription(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reactivate subscription", "error", err, "user_id", userID)
		return nil, err
	}

	// Load plan edge for response
	plan, _ := sub.QueryPlan().Only(ctx)

	subInfo := &service.SubscriptionInfo{
		ID:                 sub.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
	}
	if plan != nil {
		subInfo.Plan = &service.PlanInfo{
			ID:                    plan.ID,
			Name:                  plan.Name,
			PriceCents:            plan.PriceCents,
			BillingPeriod:         plan.BillingPeriod,
			TokenLimit:            plan.TokenLimit,
			SessionRetentionDays:  plan.SessionRetentionDays,
			MaxConcurrentSessions: plan.MaxConcurrentSessions,
			Features:              plan.Features,
		}
	}

	return generated.SubscriptionRoutesReactivateSubscription200JSONResponse{
		Subscription: mapSubscriptionInfo(subInfo),
	}, nil
}
