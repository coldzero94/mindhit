// Package main provides seed scripts for development database.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/mindhit/api/ent"
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
		fmt.Println("  test-user    Create or update test user")
		fmt.Println("  all          Run all seeds")
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
	case "test-user":
		if err := seedTestUser(ctx, client); err != nil {
			return fmt.Errorf("failed to seed test user: %w", err)
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
	if err := seedTestUser(ctx, client); err != nil {
		return err
	}
	// Add more seeds here as needed (e.g., plans in Phase 9)
	return nil
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
