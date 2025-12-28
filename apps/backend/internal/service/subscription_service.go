// Package service provides business logic for the application.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/subscription"
	"github.com/mindhit/api/ent/user"
)

// Free plan ID
const freePlanID = "free"

// SubscriptionService errors
var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrPlanNotFound         = errors.New("plan not found")
)

// SubscriptionService handles subscription-related business logic.
type SubscriptionService struct {
	client *ent.Client
}

// NewSubscriptionService creates a new SubscriptionService instance.
func NewSubscriptionService(client *ent.Client) *SubscriptionService {
	return &SubscriptionService{client: client}
}

// GetSubscription returns the active subscription for a user.
func (s *SubscriptionService) GetSubscription(ctx context.Context, userID uuid.UUID) (*ent.Subscription, error) {
	sub, err := s.client.Subscription.
		Query().
		Where(
			subscription.StatusEQ(subscription.StatusActive),
			subscription.HasUserWith(user.IDEQ(userID)),
		).
		WithPlan().
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil, ErrSubscriptionNotFound
	}
	return sub, err
}

// GetAvailablePlans returns all available plans.
func (s *SubscriptionService) GetAvailablePlans(ctx context.Context) ([]*ent.Plan, error) {
	return s.client.Plan.Query().All(ctx)
}

// CreateFreeSubscription creates a free subscription for a new user.
// This should be called during user registration.
func (s *SubscriptionService) CreateFreeSubscription(ctx context.Context, userID uuid.UUID) (*ent.Subscription, error) {
	freePlan, err := s.client.Plan.Get(ctx, freePlanID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}

	now := time.Now().UTC()
	periodEnd := now.AddDate(0, 0, 30) // 30-day period

	return s.client.Subscription.
		Create().
		SetUserID(userID).
		SetPlanID(freePlan.ID).
		SetStatus(subscription.StatusActive).
		SetCurrentPeriodStart(now).
		SetCurrentPeriodEnd(periodEnd).
		Save(ctx)
}

// GetUserPlan returns the current plan for a user.
// Returns free plan if no subscription exists.
func (s *SubscriptionService) GetUserPlan(ctx context.Context, userID uuid.UUID) (*ent.Plan, error) {
	sub, err := s.GetSubscription(ctx, userID)
	if err == nil && sub.Edges.Plan != nil {
		return sub.Edges.Plan, nil
	}

	// No active subscription, return free plan
	return s.client.Plan.Get(ctx, freePlanID)
}

// HasFeature checks if a user has access to a specific feature.
func (s *SubscriptionService) HasFeature(ctx context.Context, userID uuid.UUID, feature string) (bool, error) {
	plan, err := s.GetUserPlan(ctx, userID)
	if err != nil {
		return false, err
	}

	if plan.Features == nil {
		return false, nil
	}

	return plan.Features[feature], nil
}

// SubscriptionInfo represents subscription details for API response.
type SubscriptionInfo struct {
	ID                 uuid.UUID `json:"id"`
	Status             string    `json:"status"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CancelAtPeriodEnd  bool      `json:"cancel_at_period_end"`
	Plan               *PlanInfo `json:"plan"`
}

// PlanInfo represents plan details for API response.
type PlanInfo struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	PriceCents            int             `json:"price_cents"`
	BillingPeriod         string          `json:"billing_period"`
	TokenLimit            *int            `json:"token_limit"`
	SessionRetentionDays  *int            `json:"session_retention_days,omitempty"`
	MaxConcurrentSessions *int            `json:"max_concurrent_sessions,omitempty"`
	Features              map[string]bool `json:"features"`
}

// GetSubscriptionInfo returns formatted subscription info for API response.
func (s *SubscriptionService) GetSubscriptionInfo(ctx context.Context, userID uuid.UUID) (*SubscriptionInfo, error) {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			// Return virtual free subscription
			plan, planErr := s.client.Plan.Get(ctx, freePlanID)
			if planErr != nil {
				return nil, planErr
			}

			// Get user's signup date for period calculation
			u, userErr := s.client.User.Get(ctx, userID)
			if userErr != nil {
				return nil, userErr
			}

			// Calculate current period based on signup date
			signupDate := u.CreatedAt.UTC()
			now := time.Now().UTC()
			daysSinceSignup := int(now.Sub(signupDate).Hours() / 24)
			periodNumber := daysSinceSignup / 30
			periodStart := signupDate.AddDate(0, 0, periodNumber*30)
			periodEnd := periodStart.AddDate(0, 0, 30)

			return &SubscriptionInfo{
				ID:                 uuid.Nil,
				Status:             "active",
				CurrentPeriodStart: periodStart,
				CurrentPeriodEnd:   periodEnd,
				CancelAtPeriodEnd:  false,
				Plan:               planToInfo(plan),
			}, nil
		}
		return nil, err
	}

	return &SubscriptionInfo{
		ID:                 sub.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		Plan:               planToInfo(sub.Edges.Plan),
	}, nil
}

// planToInfo converts an Ent Plan to PlanInfo.
func planToInfo(p *ent.Plan) *PlanInfo {
	if p == nil {
		return nil
	}
	return &PlanInfo{
		ID:                    p.ID,
		Name:                  p.Name,
		PriceCents:            p.PriceCents,
		BillingPeriod:         p.BillingPeriod,
		TokenLimit:            p.TokenLimit,
		SessionRetentionDays:  p.SessionRetentionDays,
		MaxConcurrentSessions: p.MaxConcurrentSessions,
		Features:              p.Features,
	}
}
