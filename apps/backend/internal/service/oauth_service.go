// Package service provides business logic implementations.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"google.golang.org/api/idtoken"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/user"
)

var (
	// ErrInvalidIDToken is returned when the Google ID token is invalid.
	ErrInvalidIDToken = errors.New("invalid Google ID token")
	// ErrEmailNotVerified is returned when the Google account email is not verified.
	ErrEmailNotVerified = errors.New("email not verified by Google")
	// ErrCodeExchangeFailed is returned when the authorization code exchange fails.
	ErrCodeExchangeFailed = errors.New("failed to exchange authorization code")
	// ErrMissingClientSecret is returned when GOOGLE_CLIENT_SECRET is not configured.
	ErrMissingClientSecret = errors.New("GOOGLE_CLIENT_SECRET is not configured")
)

// GoogleUserInfo contains user information from Google ID Token.
type GoogleUserInfo struct {
	GoogleID string
	Email    string
	Name     string
	Picture  string
}

// googleTokenResponse represents the response from Google's token endpoint.
type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// OAuthService handles OAuth authentication.
type OAuthService struct {
	client       *ent.Client
	clientID     string
	clientSecret string
}

// NewOAuthService creates a new OAuthService instance.
func NewOAuthService(client *ent.Client) *OAuthService {
	return &OAuthService{
		client:       client,
		clientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		clientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}
}

// ValidateGoogleIDToken validates a Google ID token and extracts user info.
func (s *OAuthService) ValidateGoogleIDToken(ctx context.Context, idTokenStr string) (*GoogleUserInfo, error) {
	// Validate ID token with Google's public keys
	payload, err := idtoken.Validate(ctx, idTokenStr, s.clientID)
	if err != nil {
		return nil, ErrInvalidIDToken
	}

	// Extract claims
	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	sub, _ := payload.Claims["sub"].(string) // Google ID

	if !emailVerified {
		return nil, ErrEmailNotVerified
	}

	return &GoogleUserInfo{
		GoogleID: sub,
		Email:    email,
		Name:     name,
		Picture:  picture,
	}, nil
}

// FindOrCreateGoogleUser finds or creates a user from Google OAuth.
// Returns the user, whether it's a new user, and any error.
func (s *OAuthService) FindOrCreateGoogleUser(ctx context.Context, info *GoogleUserInfo) (*ent.User, bool, error) {
	// 1. Find by Google ID first
	u, err := s.client.User.Query().
		Where(user.GoogleIDEQ(info.GoogleID)).
		Only(ctx)
	if err == nil {
		// Existing user - update profile
		updated, updateErr := s.client.User.
			UpdateOneID(u.ID).
			SetNillableAvatarURL(&info.Picture).
			Save(ctx)
		return updated, false, updateErr
	}
	if !ent.IsNotFound(err) {
		return nil, false, err
	}

	// 2. Find by email (user signed up with email/password)
	u, err = s.client.User.Query().
		Where(
			user.EmailEQ(info.Email),
			user.StatusEQ("active"),
		).
		Only(ctx)
	if err == nil {
		// Link Google ID to existing email account
		updated, updateErr := s.client.User.
			UpdateOneID(u.ID).
			SetGoogleID(info.GoogleID).
			SetNillableAvatarURL(&info.Picture).
			Save(ctx)
		return updated, false, updateErr
	}
	if !ent.IsNotFound(err) {
		return nil, false, err
	}

	// 3. Create new user
	newUser, err := s.client.User.
		Create().
		SetEmail(info.Email).
		SetGoogleID(info.GoogleID).
		SetNillableAvatarURL(&info.Picture).
		SetAuthProvider("google").
		SetStatus("active"). // Google already verified email
		Save(ctx)
	if err != nil {
		return nil, false, err
	}

	return newUser, true, nil // isNewUser = true
}

// ExchangeAuthorizationCode exchanges an authorization code for tokens and returns user info.
// This is used for the Authorization Code flow (Chrome Extension).
func (s *OAuthService) ExchangeAuthorizationCode(ctx context.Context, code, redirectURI string) (*GoogleUserInfo, error) {
	if s.clientSecret == "" {
		return nil, ErrMissingClientSecret
	}

	// Exchange authorization code for tokens
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCodeExchangeFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCodeExchangeFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var tokenResp googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response", ErrCodeExchangeFailed)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%w: %s - %s", ErrCodeExchangeFailed, tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.IDToken == "" {
		return nil, fmt.Errorf("%w: no ID token in response", ErrCodeExchangeFailed)
	}

	// Validate the ID token and extract user info
	return s.ValidateGoogleIDToken(ctx, tokenResp.IDToken)
}
