// Package service provides business logic for the application.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/subscription"
	"github.com/mindhit/api/ent/tokenusage"
	"github.com/mindhit/api/ent/user"
)

// Default token limit for free plan users without subscription
const defaultFreeTokenLimit = 50000

// UsageService handles token usage tracking and limit checking.
type UsageService struct {
	client *ent.Client
}

// NewUsageService creates a new UsageService instance.
func NewUsageService(client *ent.Client) *UsageService {
	return &UsageService{client: client}
}

// UsageRecord represents a token usage record to be created.
type UsageRecord struct {
	UserID    uuid.UUID
	SessionID uuid.UUID // Optional, can be uuid.Nil
	Operation string    // 'summarize', 'mindmap', 'keywords'
	Tokens    int
	AIModel   string // Optional
}

// LimitStatus represents the current usage status for a user.
type LimitStatus struct {
	TokensUsed  int     `json:"tokens_used"`
	TokenLimit  int     `json:"token_limit"`
	IsUnlimited bool    `json:"is_unlimited"`
	PercentUsed float64 `json:"percent_used"`
	CanUseAI    bool    `json:"can_use_ai"`
}

// UsageSummary represents usage statistics for a billing period.
type UsageSummary struct {
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	TokensUsed  int            `json:"tokens_used"`
	TokenLimit  int            `json:"token_limit"`
	PercentUsed float64        `json:"percent_used"`
	IsUnlimited bool           `json:"is_unlimited"`
	CanUseAI    bool           `json:"can_use_ai"`
	ByOperation map[string]int `json:"by_operation"`
}

// RecordUsage creates a new token usage record.
func (s *UsageService) RecordUsage(ctx context.Context, record UsageRecord) error {
	periodStart := s.getCurrentPeriodStart(ctx, record.UserID)

	builder := s.client.TokenUsage.Create().
		SetUserID(record.UserID).
		SetOperation(record.Operation).
		SetTokensUsed(record.Tokens).
		SetPeriodStart(periodStart)

	if record.SessionID != uuid.Nil {
		builder.SetSessionID(record.SessionID)
	}
	if record.AIModel != "" {
		builder.SetAiModel(record.AIModel)
	}

	_, err := builder.Save(ctx)
	return err
}

// CheckLimit checks if the user can use AI based on their token limits.
func (s *UsageService) CheckLimit(ctx context.Context, userID uuid.UUID) (*LimitStatus, error) {
	// Get active subscription with plan
	sub, _ := s.client.Subscription.
		Query().
		Where(
			subscription.StatusEQ(subscription.StatusActive),
			subscription.HasUserWith(user.IDEQ(userID)),
		).
		WithPlan().
		Only(ctx)

	periodStart := s.getCurrentPeriodStart(ctx, userID)

	// Get total tokens used in current period
	var result []struct {
		Sum int `json:"sum"`
	}
	err := s.client.TokenUsage.
		Query().
		Where(
			tokenusage.HasUserWith(user.IDEQ(userID)),
			tokenusage.PeriodStartGTE(periodStart),
		).
		Aggregate(ent.Sum(tokenusage.FieldTokensUsed)).
		Scan(ctx, &result)

	usage := 0
	if err == nil && len(result) > 0 {
		usage = result[0].Sum
	}

	// Determine limits
	limit := defaultFreeTokenLimit
	isUnlimited := false

	if sub != nil && sub.Edges.Plan != nil {
		if sub.Edges.Plan.TokenLimit != nil {
			limit = *sub.Edges.Plan.TokenLimit
		} else {
			isUnlimited = true
		}
	}

	percentUsed := 0.0
	if !isUnlimited && limit > 0 {
		percentUsed = float64(usage) / float64(limit) * 100
	}

	return &LimitStatus{
		TokensUsed:  usage,
		TokenLimit:  limit,
		IsUnlimited: isUnlimited,
		PercentUsed: percentUsed,
		CanUseAI:    isUnlimited || usage < limit,
	}, nil
}

// GetCurrentUsage returns detailed usage statistics for the current billing period.
func (s *UsageService) GetCurrentUsage(ctx context.Context, userID uuid.UUID) (*UsageSummary, error) {
	periodStart := s.getCurrentPeriodStart(ctx, userID)
	periodEnd := periodStart.AddDate(0, 0, 30) // 30-day period

	// Get usage by operation
	usages, err := s.client.TokenUsage.
		Query().
		Where(
			tokenusage.HasUserWith(user.IDEQ(userID)),
			tokenusage.PeriodStartGTE(periodStart),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	byOperation := make(map[string]int)
	totalTokens := 0
	for _, u := range usages {
		byOperation[u.Operation] += u.TokensUsed
		totalTokens += u.TokensUsed
	}

	// Get limit info
	limitStatus, err := s.CheckLimit(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UsageSummary{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		TokensUsed:  totalTokens,
		TokenLimit:  limitStatus.TokenLimit,
		PercentUsed: limitStatus.PercentUsed,
		IsUnlimited: limitStatus.IsUnlimited,
		CanUseAI:    limitStatus.CanUseAI,
		ByOperation: byOperation,
	}, nil
}

// GetUsageHistory returns usage history for past billing periods.
func (s *UsageService) GetUsageHistory(ctx context.Context, userID uuid.UUID, months int) ([]UsageSummary, error) {
	if months <= 0 {
		months = 6 // Default to 6 months
	}

	history := make([]UsageSummary, 0, months)
	now := time.Now().UTC()

	for i := 0; i < months; i++ {
		// Calculate period start for each month
		periodStart := s.calculatePeriodStartForDate(ctx, userID, now.AddDate(0, 0, -30*i))
		periodEnd := periodStart.AddDate(0, 0, 30)

		// Get usage for this period
		usages, err := s.client.TokenUsage.
			Query().
			Where(
				tokenusage.HasUserWith(user.IDEQ(userID)),
				tokenusage.PeriodStartGTE(periodStart),
				tokenusage.PeriodStartLT(periodEnd),
			).
			All(ctx)
		if err != nil {
			continue
		}

		byOperation := make(map[string]int)
		totalTokens := 0
		for _, u := range usages {
			byOperation[u.Operation] += u.TokensUsed
			totalTokens += u.TokensUsed
		}

		// Get limit info (use current plan for simplicity)
		limitStatus, _ := s.CheckLimit(ctx, userID)
		limit := defaultFreeTokenLimit
		isUnlimited := false
		if limitStatus != nil {
			limit = limitStatus.TokenLimit
			isUnlimited = limitStatus.IsUnlimited
		}

		percentUsed := 0.0
		if !isUnlimited && limit > 0 {
			percentUsed = float64(totalTokens) / float64(limit) * 100
		}

		history = append(history, UsageSummary{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			TokensUsed:  totalTokens,
			TokenLimit:  limit,
			PercentUsed: percentUsed,
			IsUnlimited: isUnlimited,
			CanUseAI:    true, // Historical data
			ByOperation: byOperation,
		})
	}

	return history, nil
}

// getCurrentPeriodStart returns the start of the current billing period.
func (s *UsageService) getCurrentPeriodStart(ctx context.Context, userID uuid.UUID) time.Time {
	return s.calculatePeriodStartForDate(ctx, userID, time.Now().UTC())
}

// calculatePeriodStartForDate calculates the billing period start for a given date.
func (s *UsageService) calculatePeriodStartForDate(ctx context.Context, userID uuid.UUID, date time.Time) time.Time {
	// Check for active subscription first
	sub, err := s.client.Subscription.
		Query().
		Where(
			subscription.StatusEQ(subscription.StatusActive),
			subscription.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx)

	if err == nil && sub != nil {
		// Pro/Enterprise: use subscription period
		return sub.CurrentPeriodStart
	}

	// Free plan: calculate based on signup date (30-day rolling periods)
	return s.calculateFreePlanPeriodStart(ctx, userID, date)
}

// calculateFreePlanPeriodStart calculates billing period for free plan users.
func (s *UsageService) calculateFreePlanPeriodStart(ctx context.Context, userID uuid.UUID, date time.Time) time.Time {
	u, err := s.client.User.Get(ctx, userID)
	if err != nil {
		return date.Truncate(24 * time.Hour)
	}

	signupDate := u.CreatedAt.UTC()
	targetDate := date.UTC()

	daysSinceSignup := int(targetDate.Sub(signupDate).Hours() / 24)
	if daysSinceSignup < 0 {
		return signupDate
	}

	periodNumber := daysSinceSignup / 30
	periodStart := signupDate.AddDate(0, 0, periodNumber*30)

	return periodStart
}
