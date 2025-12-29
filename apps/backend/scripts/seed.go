// Package main provides seed scripts for development database.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/aiconfig"
	"github.com/mindhit/api/ent/plan"
	"github.com/mindhit/api/ent/subscription"
	"github.com/mindhit/api/ent/user"

	_ "github.com/lib/pq"
)

const (
	// TestUserEmail is the email for the test user
	TestUserEmail = "test@mindhit.dev"
	// TestUserPassword is the password for the test user
	TestUserPassword = "test1234!"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./scripts/seed.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  plans         Create or update plans")
		fmt.Println("  test-user     Create or update test user")
		fmt.Println("  subscriptions Create or update test user subscription")
		fmt.Println("  ai-configs    Create or update AI configs")
		fmt.Println("  all           Run all seeds")
		return fmt.Errorf("no command specified")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/mindhit?sslmode=disable"
	}

	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	switch os.Args[1] {
	case "plans":
		if err := seedPlans(ctx, client); err != nil {
			return fmt.Errorf("failed to seed plans: %w", err)
		}
	case "test-user":
		if err := seedTestUser(ctx, client); err != nil {
			return fmt.Errorf("failed to seed test user: %w", err)
		}
	case "subscriptions":
		if err := seedSubscriptions(ctx, client); err != nil {
			return fmt.Errorf("failed to seed subscriptions: %w", err)
		}
	case "ai-configs":
		if err := seedAIConfigs(ctx, client); err != nil {
			return fmt.Errorf("failed to seed ai configs: %w", err)
		}
	case "all":
		if err := seedAll(ctx, client); err != nil {
			return fmt.Errorf("failed to seed: %w", err)
		}
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}

	return nil
}

func seedAll(ctx context.Context, client *ent.Client) error {
	if err := seedPlans(ctx, client); err != nil {
		return err
	}
	if err := seedTestUser(ctx, client); err != nil {
		return err
	}
	if err := seedSubscriptions(ctx, client); err != nil {
		return err
	}
	if err := seedAIConfigs(ctx, client); err != nil {
		return err
	}
	return nil
}

func seedPlans(ctx context.Context, client *ent.Client) error {
	plans := []struct {
		ID                    string
		Name                  string
		PriceCents            int
		BillingPeriod         string
		TokenLimit            *int
		SessionRetentionDays  *int
		MaxConcurrentSessions *int
		Features              map[string]bool
	}{
		{
			ID:                    "free",
			Name:                  "Free",
			PriceCents:            0,
			BillingPeriod:         "monthly",
			TokenLimit:            intPtr(50000),
			SessionRetentionDays:  intPtr(30),
			MaxConcurrentSessions: intPtr(1),
			Features: map[string]bool{
				"export_png": true,
			},
		},
		{
			ID:                    "pro",
			Name:                  "Pro",
			PriceCents:            1200,
			BillingPeriod:         "monthly",
			TokenLimit:            intPtr(500000),
			SessionRetentionDays:  nil, // unlimited
			MaxConcurrentSessions: intPtr(5),
			Features: map[string]bool{
				"export_png":       true,
				"export_svg":       true,
				"export_md":        true,
				"export_json":      true,
				"priority_support": true,
			},
		},
		{
			ID:                    "enterprise",
			Name:                  "Enterprise",
			PriceCents:            0, // custom pricing
			BillingPeriod:         "monthly",
			TokenLimit:            nil, // unlimited
			SessionRetentionDays:  nil, // unlimited
			MaxConcurrentSessions: nil, // unlimited
			Features: map[string]bool{
				"export_png":   true,
				"export_svg":   true,
				"export_md":    true,
				"export_json":  true,
				"api_access":   true,
				"team_sharing": true,
				"sso":          true,
				"custom_ai":    true,
			},
		},
	}

	for _, p := range plans {
		// Check if plan already exists
		exists, err := client.Plan.Query().
			Where(plan.IDEQ(p.ID)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check plan %s: %w", p.ID, err)
		}

		if exists {
			// Update existing plan
			update := client.Plan.UpdateOneID(p.ID).
				SetName(p.Name).
				SetPriceCents(p.PriceCents).
				SetBillingPeriod(p.BillingPeriod).
				SetFeatures(p.Features)

			if p.TokenLimit != nil {
				update.SetTokenLimit(*p.TokenLimit)
			} else {
				update.ClearTokenLimit()
			}
			if p.SessionRetentionDays != nil {
				update.SetSessionRetentionDays(*p.SessionRetentionDays)
			} else {
				update.ClearSessionRetentionDays()
			}
			if p.MaxConcurrentSessions != nil {
				update.SetMaxConcurrentSessions(*p.MaxConcurrentSessions)
			} else {
				update.ClearMaxConcurrentSessions()
			}

			if _, err := update.Save(ctx); err != nil {
				return fmt.Errorf("failed to update plan %s: %w", p.ID, err)
			}
			fmt.Printf("✓ Plan updated: %s\n", p.ID)
		} else {
			// Create new plan
			create := client.Plan.Create().
				SetID(p.ID).
				SetName(p.Name).
				SetPriceCents(p.PriceCents).
				SetBillingPeriod(p.BillingPeriod).
				SetFeatures(p.Features).
				SetCreatedAt(time.Now())

			if p.TokenLimit != nil {
				create.SetTokenLimit(*p.TokenLimit)
			}
			if p.SessionRetentionDays != nil {
				create.SetSessionRetentionDays(*p.SessionRetentionDays)
			}
			if p.MaxConcurrentSessions != nil {
				create.SetMaxConcurrentSessions(*p.MaxConcurrentSessions)
			}

			if _, err := create.Save(ctx); err != nil {
				return fmt.Errorf("failed to create plan %s: %w", p.ID, err)
			}
			fmt.Printf("✓ Plan created: %s\n", p.ID)
		}
	}

	return nil
}

func intPtr(i int) *int {
	return &i
}

func seedTestUser(ctx context.Context, client *ent.Client) error {
	// Check if user already exists
	existing, err := client.User.Query().
		Where(user.EmailEQ(TestUserEmail)).
		Only(ctx)

	if err == nil {
		// User exists, update password
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(TestUserPassword),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		_, err = client.User.UpdateOne(existing).
			SetPasswordHash(string(hashedPassword)).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update test user: %w", err)
		}

		fmt.Printf("✓ Test user updated: %s (ID: %s)\n", TestUserEmail, existing.ID)
		return nil
	}

	if !ent.IsNotFound(err) {
		return fmt.Errorf("failed to query user: %w", err)
	}

	// Create new user
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(TestUserPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newUser, err := client.User.Create().
		SetEmail(TestUserEmail).
		SetPasswordHash(string(hashedPassword)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	fmt.Printf("✓ Test user created: %s (ID: %s)\n", TestUserEmail, newUser.ID)
	return nil
}

func seedSubscriptions(ctx context.Context, client *ent.Client) error {
	// Get test user
	testUser, err := client.User.Query().
		Where(user.EmailEQ(TestUserEmail)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to find test user: %w", err)
	}

	// Check if subscription already exists
	exists, err := client.Subscription.Query().
		Where(subscription.HasUserWith(user.IDEQ(testUser.ID))).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check subscription: %w", err)
	}

	if exists {
		fmt.Printf("✓ Subscription already exists for %s\n", TestUserEmail)
		return nil
	}

	// Create subscription with free plan
	now := time.Now()
	_, err = client.Subscription.Create().
		SetUserID(testUser.ID).
		SetPlanID("free").
		SetStatus(subscription.StatusActive).
		SetCurrentPeriodStart(now).
		SetCurrentPeriodEnd(now.AddDate(0, 1, 0)). // 1 month from now
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	fmt.Printf("✓ Subscription created for %s (Plan: free)\n", TestUserEmail)
	return nil
}

func seedAIConfigs(ctx context.Context, client *ent.Client) error {
	configs := []struct {
		TaskType          string
		Provider          string
		Model             string
		FallbackProviders []string
		Temperature       float64
		MaxTokens         int
		ThinkingBudget    int
		JSONMode          bool
	}{
		{
			TaskType:          "default",
			Provider:          "openai",
			Model:             "gpt-4o",
			FallbackProviders: []string{"gemini", "claude"},
			Temperature:       0.7,
			MaxTokens:         4096,
			ThinkingBudget:    0,
			JSONMode:          false,
		},
		{
			TaskType:          "tag_extraction",
			Provider:          "gemini",
			Model:             "gemini-2.0-flash-exp",
			FallbackProviders: []string{"openai"},
			Temperature:       0.3,
			MaxTokens:         1024,
			ThinkingBudget:    0,
			JSONMode:          true,
		},
		{
			TaskType:          "mindmap",
			Provider:          "claude",
			Model:             "claude-sonnet-4-20250514",
			FallbackProviders: []string{"openai"},
			Temperature:       0.5,
			MaxTokens:         8192,
			ThinkingBudget:    10000,
			JSONMode:          true,
		},
	}

	for _, cfg := range configs {
		// Check if config already exists
		exists, err := client.AIConfig.Query().
			Where(aiconfig.TaskTypeEQ(cfg.TaskType)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check ai config %s: %w", cfg.TaskType, err)
		}

		if exists {
			// Update existing config
			_, err = client.AIConfig.Update().
				Where(aiconfig.TaskTypeEQ(cfg.TaskType)).
				SetProvider(cfg.Provider).
				SetModel(cfg.Model).
				SetFallbackProviders(cfg.FallbackProviders).
				SetTemperature(cfg.Temperature).
				SetMaxTokens(cfg.MaxTokens).
				SetThinkingBudget(cfg.ThinkingBudget).
				SetJSONMode(cfg.JSONMode).
				SetEnabled(true).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to update ai config %s: %w", cfg.TaskType, err)
			}
			fmt.Printf("✓ AI config updated: %s\n", cfg.TaskType)
		} else {
			// Create new config
			_, err = client.AIConfig.Create().
				SetTaskType(cfg.TaskType).
				SetProvider(cfg.Provider).
				SetModel(cfg.Model).
				SetFallbackProviders(cfg.FallbackProviders).
				SetTemperature(cfg.Temperature).
				SetMaxTokens(cfg.MaxTokens).
				SetThinkingBudget(cfg.ThinkingBudget).
				SetJSONMode(cfg.JSONMode).
				SetEnabled(true).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create ai config %s: %w", cfg.TaskType, err)
			}
			fmt.Printf("✓ AI config created: %s\n", cfg.TaskType)
		}
	}

	return nil
}
