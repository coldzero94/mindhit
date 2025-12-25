package controller

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
)

// AuthController implements auth-related handlers from StrictServerInterface
type AuthController struct {
	authService *service.AuthService
	jwtService  *service.JWTService
}

// NewAuthController creates a new AuthController
func NewAuthController(authService *service.AuthService, jwtService *service.JWTService) *AuthController {
	return &AuthController{
		authService: authService,
		jwtService:  jwtService,
	}
}

// RoutesSignup handles POST /v1/auth/signup
func (c *AuthController) RoutesSignup(ctx context.Context, request generated.RoutesSignupRequestObject) (generated.RoutesSignupResponseObject, error) {
	if request.Body == nil {
		return generated.RoutesSignup400JSONResponse{
			Error: struct {
				Details *[]generated.CommonValidationDetail `json:"details,omitempty"`
				Message string                              `json:"message"`
			}{
				Message: "request body is required",
			},
		}, nil
	}

	user, err := c.authService.Signup(ctx, request.Body.Email, request.Body.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			return generated.RoutesSignup409JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "email already exists",
				},
			}, nil
		}
		slog.Error("signup failed", "error", err, "email", request.Body.Email)
		return nil, err
	}

	tokenPair, err := c.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		slog.Error("failed to generate token pair", "error", err, "user_id", user.ID)
		return nil, err
	}

	return generated.RoutesSignup201JSONResponse{
		Token: tokenPair.AccessToken,
		User: generated.AuthUser{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// RoutesLogin handles POST /v1/auth/login
func (c *AuthController) RoutesLogin(ctx context.Context, request generated.RoutesLoginRequestObject) (generated.RoutesLoginResponseObject, error) {
	if request.Body == nil {
		return generated.RoutesLogin401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "request body is required",
			},
		}, nil
	}

	user, err := c.authService.Login(ctx, request.Body.Email, request.Body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return generated.RoutesLogin401JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "invalid email or password",
				},
			}, nil
		}
		slog.Error("login failed", "error", err, "email", request.Body.Email)
		return nil, err
	}

	tokenPair, err := c.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		slog.Error("failed to generate token pair", "error", err, "user_id", user.ID)
		return nil, err
	}

	return generated.RoutesLogin200JSONResponse{
		Token: tokenPair.AccessToken,
		User: generated.AuthUser{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// RoutesRefresh handles POST /v1/auth/refresh
func (c *AuthController) RoutesRefresh(ctx context.Context, request generated.RoutesRefreshRequestObject) (generated.RoutesRefreshResponseObject, error) {
	// Extract token from Authorization header
	authHeader := request.Params.Authorization
	if authHeader == "" {
		return generated.RoutesRefresh401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "authorization header is required",
			},
		}, nil
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return generated.RoutesRefresh401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid authorization header format",
			},
		}, nil
	}

	tokenString := parts[1]

	// Validate refresh token
	claims, err := c.jwtService.ValidateRefreshToken(tokenString)
	if err != nil {
		return generated.RoutesRefresh401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid or expired refresh token",
			},
		}, nil
	}

	// Verify user still exists
	_, err = c.authService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return generated.RoutesRefresh401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "user not found",
			},
		}, nil
	}

	// Generate new access token
	accessToken, _, err := c.jwtService.GenerateAccessToken(claims.UserID)
	if err != nil {
		slog.Error("failed to generate access token", "error", err, "user_id", claims.UserID)
		return nil, err
	}

	return generated.RoutesRefresh200JSONResponse{
		Token: accessToken,
	}, nil
}

// RoutesMe handles GET /v1/auth/me
func (c *AuthController) RoutesMe(ctx context.Context, request generated.RoutesMeRequestObject) (generated.RoutesMeResponseObject, error) {
	// Extract token from Authorization header
	authHeader := request.Params.Authorization
	if authHeader == "" {
		return generated.RoutesMe401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "authorization header is required",
			},
		}, nil
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return generated.RoutesMe401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid authorization header format",
			},
		}, nil
	}

	tokenString := parts[1]

	// Validate access token
	claims, err := c.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		return generated.RoutesMe401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid or expired access token",
			},
		}, nil
	}

	// Get user info
	user, err := c.authService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return generated.RoutesMe401JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "user not found",
				},
			}, nil
		}
		slog.Error("failed to get user", "error", err, "user_id", claims.UserID)
		return nil, err
	}

	return generated.RoutesMe200JSONResponse{
		User: generated.AuthUser{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// RoutesLogout handles POST /v1/auth/logout
func (c *AuthController) RoutesLogout(ctx context.Context, request generated.RoutesLogoutRequestObject) (generated.RoutesLogoutResponseObject, error) {
	// Extract token from Authorization header
	authHeader := request.Params.Authorization
	if authHeader == "" {
		return generated.RoutesLogout401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "authorization header is required",
			},
		}, nil
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return generated.RoutesLogout401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid authorization header format",
			},
		}, nil
	}

	tokenString := parts[1]

	// Validate access token (just to ensure it's a valid token)
	claims, err := c.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		return generated.RoutesLogout401JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "invalid or expired access token",
			},
		}, nil
	}

	// For stateless JWT, we just log the logout event
	// In production, you might want to add the token to a blacklist
	slog.Info("user logged out", "user_id", claims.UserID)

	return generated.RoutesLogout200JSONResponse{
		Message: "successfully logged out",
	}, nil
}

// RoutesForgotPassword handles POST /v1/auth/forgot-password
func (c *AuthController) RoutesForgotPassword(ctx context.Context, request generated.RoutesForgotPasswordRequestObject) (generated.RoutesForgotPasswordResponseObject, error) {
	if request.Body == nil {
		return generated.RoutesForgotPassword400JSONResponse{
			Error: struct {
				Details *[]generated.CommonValidationDetail `json:"details,omitempty"`
				Message string                              `json:"message"`
			}{
				Message: "request body is required",
			},
		}, nil
	}

	token, err := c.authService.RequestPasswordReset(ctx, request.Body.Email)
	if err != nil {
		// Log internal errors but don't expose them to client
		slog.Error("failed to create reset token", "error", err)
	}

	// If token was generated, email should be sent here
	if token != "" {
		// TODO: Send email with reset link (implement in Phase 5 or later)
		slog.Info("password reset requested", "email", request.Body.Email)
	}

	// Security: Always return same response regardless of email existence
	return generated.RoutesForgotPassword200JSONResponse{
		Message: "If the email exists, a password reset link has been sent.",
	}, nil
}

// RoutesResetPassword handles POST /v1/auth/reset-password
func (c *AuthController) RoutesResetPassword(ctx context.Context, request generated.RoutesResetPasswordRequestObject) (generated.RoutesResetPasswordResponseObject, error) {
	if request.Body == nil {
		return generated.RoutesResetPassword400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: "request body is required",
			},
		}, nil
	}

	err := c.authService.ResetPassword(ctx, request.Body.Token, request.Body.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenInvalid):
			return generated.RoutesResetPassword400JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "invalid or expired token",
				},
			}, nil
		case errors.Is(err, service.ErrTokenExpired):
			return generated.RoutesResetPassword400JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "token has expired",
				},
			}, nil
		case errors.Is(err, service.ErrTokenUsed):
			return generated.RoutesResetPassword400JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "token has already been used",
				},
			}, nil
		case errors.Is(err, service.ErrUserInactive):
			return generated.RoutesResetPassword400JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "user account is inactive",
				},
			}, nil
		default:
			slog.Error("password reset failed", "error", err)
			return nil, err
		}
	}

	return generated.RoutesResetPassword200JSONResponse{
		Message: "Password has been reset successfully.",
	}, nil
}
