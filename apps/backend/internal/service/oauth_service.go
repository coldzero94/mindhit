// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"
	"os"

	"google.golang.org/api/idtoken"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/user"
)

var (
	// ErrInvalidIDToken is returned when the Google ID token is invalid.
	ErrInvalidIDToken = errors.New("invalid Google ID token")
	// ErrEmailNotVerified is returned when the Google account email is not verified.
	ErrEmailNotVerified = errors.New("email not verified by Google")
)

// GoogleUserInfo contains user information from Google ID Token.
type GoogleUserInfo struct {
	GoogleID string
	Email    string
	Name     string
	Picture  string
}

// OAuthService handles OAuth authentication.
type OAuthService struct {
	client   *ent.Client
	clientID string
}

// NewOAuthService creates a new OAuthService instance.
func NewOAuthService(client *ent.Client) *OAuthService {
	return &OAuthService{
		client:   client,
		clientID: os.Getenv("GOOGLE_CLIENT_ID"),
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
