// Package service provides business logic for the application.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/passwordresettoken"
	"github.com/mindhit/api/ent/user"
)

// Auth service errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenUsed          = errors.New("token already used")
	ErrTokenInvalid       = errors.New("invalid token")
)

// AuthService handles user authentication operations.
type AuthService struct {
	client *ent.Client
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(client *ent.Client) *AuthService {
	return &AuthService{client: client}
}

// activeUsers returns a query filtered to active users only
func (s *AuthService) activeUsers() *ent.UserQuery {
	return s.client.User.Query().Where(user.StatusEQ(user.StatusActive))
}

// Signup creates a new user account with the given email and password.
func (s *AuthService) Signup(ctx context.Context, email, password string) (*ent.User, error) {
	// Check email duplication (active users only)
	exists, err := s.activeUsers().
		Where(user.EmailEQ(email)).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user (status defaults to "active" via SoftDeleteMixin)
	return s.client.User.
		Create().
		SetEmail(email).
		SetPasswordHash(string(hashedPassword)).
		Save(ctx)
}

// Login authenticates a user with email and password.
func (s *AuthService) Login(ctx context.Context, email, password string) (*ent.User, error) {
	// Query active users only
	u, err := s.activeUsers().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user has a password (Google OAuth users don't)
	if u.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u, nil
}

// GetUserByID retrieves a user by their ID.
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	// Query active users only
	u, err := s.activeUsers().
		Where(user.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// GetUserByEmail retrieves a user by their email address.
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	// Query active users only
	u, err := s.activeUsers().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// generateSecureToken creates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RequestPasswordReset creates a password reset token and returns it
// The caller is responsible for sending the email
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	// Find user (active only)
	u, err := s.activeUsers().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Security: return empty string for non-existent email (prevent enumeration)
			return "", nil
		}
		return "", err
	}

	// Invalidate existing unused tokens
	_, err = s.client.PasswordResetToken.
		Update().
		Where(
			passwordresettoken.UserIDEQ(u.ID),
			passwordresettoken.UsedEQ(false),
		).
		SetUsed(true).
		Save(ctx)
	if err != nil {
		return "", err
	}

	// Generate new token
	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	// Save token (1 hour expiry)
	_, err = s.client.PasswordResetToken.
		Create().
		SetToken(token).
		SetUserID(u.ID).
		SetExpiresAt(time.Now().Add(1 * time.Hour)).
		Save(ctx)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword validates the token and updates the password
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Find token
	resetToken, err := s.client.PasswordResetToken.
		Query().
		Where(
			passwordresettoken.TokenEQ(token),
			passwordresettoken.UsedEQ(false),
		).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrTokenInvalid
		}
		return err
	}

	// Check expiration
	if time.Now().After(resetToken.ExpiresAt) {
		return ErrTokenExpired
	}

	// Check user is active
	if resetToken.Edges.User.Status != user.StatusActive {
		return ErrUserInactive
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Use transaction to update password and mark token as used
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return err
	}

	// Update password
	_, err = tx.User.
		UpdateOneID(resetToken.Edges.User.ID).
		SetPasswordHash(string(hashedPassword)).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Mark token as used
	_, err = tx.PasswordResetToken.
		UpdateOneID(resetToken.ID).
		SetUsed(true).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
