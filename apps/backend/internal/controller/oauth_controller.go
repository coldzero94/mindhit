// Package controller provides HTTP handlers for the API.
package controller

import (
	"context"
	"log/slog"

	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
)

// OAuthController implements OAuth-related handlers.
type OAuthController struct {
	oauthService        *service.OAuthService
	jwtService          *service.JWTService
	subscriptionService *service.SubscriptionService
}

// NewOAuthController creates a new OAuthController.
func NewOAuthController(
	oauthService *service.OAuthService,
	jwtService *service.JWTService,
	subscriptionService *service.SubscriptionService,
) *OAuthController {
	return &OAuthController{
		oauthService:        oauthService,
		jwtService:          jwtService,
		subscriptionService: subscriptionService,
	}
}

// RoutesGoogleAuth handles POST /v1/auth/google.
func (c *OAuthController) RoutesGoogleAuth(
	ctx context.Context,
	request generated.RoutesGoogleAuthRequestObject,
) (generated.RoutesGoogleAuthResponseObject, error) {
	// 1. Validate Google ID Token
	userInfo, err := c.oauthService.ValidateGoogleIDToken(ctx, request.Body.Credential)
	if err != nil {
		slog.WarnContext(ctx, "invalid Google ID token", "error", err)
		return generated.RoutesGoogleAuth401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "Invalid Google credentials",
			},
		}, nil
	}

	// 2. Find or create user
	user, isNewUser, err := c.oauthService.FindOrCreateGoogleUser(ctx, userInfo)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find or create Google user", "error", err)
		return nil, err
	}

	// 3. Create free subscription for new users
	if isNewUser {
		if _, subErr := c.subscriptionService.CreateFreeSubscription(ctx, user.ID); subErr != nil {
			slog.ErrorContext(ctx, "failed to create free subscription", "error", subErr, "user_id", user.ID)
			// Continue even if subscription creation fails
		}
	}

	// 4. Generate JWT token pair
	tokenPair, err := c.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate token pair", "error", err)
		return nil, err
	}

	return generated.RoutesGoogleAuth200JSONResponse{
		Token: tokenPair.AccessToken,
		User: generated.AuthUser{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// RoutesGoogleAuthCode handles POST /v1/auth/google/code.
// This endpoint exchanges an authorization code for tokens (used by Chrome Extension).
func (c *OAuthController) RoutesGoogleAuthCode(
	ctx context.Context,
	request generated.RoutesGoogleAuthCodeRequestObject,
) (generated.RoutesGoogleAuthCodeResponseObject, error) {
	// 1. Exchange authorization code for tokens and get user info
	userInfo, err := c.oauthService.ExchangeAuthorizationCode(ctx, request.Body.Code, request.Body.RedirectUri)
	if err != nil {
		slog.WarnContext(ctx, "failed to exchange authorization code", "error", err)
		return generated.RoutesGoogleAuthCode401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "Invalid authorization code",
			},
		}, nil
	}

	// 2. Find or create user
	user, isNewUser, err := c.oauthService.FindOrCreateGoogleUser(ctx, userInfo)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find or create Google user", "error", err)
		return nil, err
	}

	// 3. Create free subscription for new users
	if isNewUser {
		if _, subErr := c.subscriptionService.CreateFreeSubscription(ctx, user.ID); subErr != nil {
			slog.ErrorContext(ctx, "failed to create free subscription", "error", subErr, "user_id", user.ID)
			// Continue even if subscription creation fails
		}
	}

	// 4. Generate JWT token pair
	tokenPair, err := c.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate token pair", "error", err)
		return nil, err
	}

	return generated.RoutesGoogleAuthCode200JSONResponse{
		Token: tokenPair.AccessToken,
		User: generated.AuthUser{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}
