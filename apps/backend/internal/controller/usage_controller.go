// Package controller provides HTTP handlers for the API.
package controller

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
)

// UsageController implements usage-related handlers from StrictServerInterface.
type UsageController struct {
	usageService *service.UsageService
	jwtService   *service.JWTService
}

// NewUsageController creates a new UsageController.
func NewUsageController(usageService *service.UsageService, jwtService *service.JWTService) *UsageController {
	return &UsageController{
		usageService: usageService,
		jwtService:   jwtService,
	}
}

// extractUserID extracts and validates user ID from authorization header.
func (c *UsageController) extractUserID(authHeader string) (uuid.UUID, error) {
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

// UsageRoutesGetUsage handles GET /v1/usage.
func (c *UsageController) UsageRoutesGetUsage(ctx context.Context, request generated.UsageRoutesGetUsageRequestObject) (generated.UsageRoutesGetUsageResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.UsageRoutesGetUsage401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	usage, err := c.usageService.GetCurrentUsage(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get current usage", "error", err, "user_id", userID)
		return nil, err
	}

	return generated.UsageRoutesGetUsage200JSONResponse{
		Usage: mapUsageSummary(usage),
	}, nil
}

// UsageRoutesGetUsageHistory handles GET /v1/usage/history.
func (c *UsageController) UsageRoutesGetUsageHistory(ctx context.Context, request generated.UsageRoutesGetUsageHistoryRequestObject) (generated.UsageRoutesGetUsageHistoryResponseObject, error) {
	userID, err := c.extractUserID(request.Params.Authorization)
	if err != nil {
		return generated.UsageRoutesGetUsageHistory401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
			},
		}, nil
	}

	months := 6 // default
	if request.Params.Months != nil {
		months = int(*request.Params.Months)
	}

	history, err := c.usageService.GetUsageHistory(ctx, userID, months)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get usage history", "error", err, "user_id", userID)
		return nil, err
	}

	apiHistory := make([]generated.UsageUsageSummary, 0, len(history))
	for _, h := range history {
		apiHistory = append(apiHistory, mapUsageSummary(&h))
	}

	return generated.UsageRoutesGetUsageHistory200JSONResponse{
		History: apiHistory,
	}, nil
}

// mapUsageSummary maps service.UsageSummary to generated.UsageUsageSummary.
func mapUsageSummary(u *service.UsageSummary) generated.UsageUsageSummary {
	byOperation := make(map[string]int32)
	for k, v := range u.ByOperation {
		byOperation[k] = int32(v)
	}

	return generated.UsageUsageSummary{
		PeriodStart: u.PeriodStart,
		PeriodEnd:   u.PeriodEnd,
		TokensUsed:  int32(u.TokensUsed),
		TokenLimit:  int32(u.TokenLimit),
		PercentUsed: u.PercentUsed,
		IsUnlimited: u.IsUnlimited,
		CanUseAi:    u.CanUseAI,
		ByOperation: byOperation,
	}
}
