// Package testutil provides testing utilities for database operations.
package testutil

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/mindhit/api/ent"
)

var (
	sharedDB   *sql.DB
	sharedOnce sync.Once
	schemaOnce sync.Once
)

// getTestDatabaseURL returns the test database URL from environment or default
func getTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://postgres:password@localhost:5433/mindhit_test?sslmode=disable"
}

// getSharedDB returns a shared database connection pool
func getSharedDB(t *testing.T) *sql.DB {
	sharedOnce.Do(func() {
		var err error
		sharedDB, err = sql.Open("postgres", getTestDatabaseURL())
		if err != nil {
			t.Fatalf("failed to open postgres: %v", err)
		}
		sharedDB.SetMaxOpenConns(25)
		sharedDB.SetMaxIdleConns(10)
	})
	return sharedDB
}

// ensureSchema ensures the database schema is created (only once)
func ensureSchema(t *testing.T, client *ent.Client) {
	schemaOnce.Do(func() {
		ctx := context.Background()
		if err := client.Schema.Create(ctx); err != nil {
			t.Fatalf("failed to create schema: %v", err)
		}
	})
}

// SetupTestDB creates a test database client.
// Uses a shared connection pool for efficiency.
// Tests should use unique identifiers (emails, etc.) to avoid conflicts.
func SetupTestDB(t *testing.T) *ent.Client {
	t.Helper()

	db := getSharedDB(t)
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	// Ensure schema exists (only once across all tests)
	ensureSchema(t, client)

	return client
}

// CleanupTestDB is kept for backward compatibility.
// With shared connection pool, we don't close individual clients.
func CleanupTestDB(t *testing.T, _ *ent.Client) {
	t.Helper()
	// No-op: using shared connection pool
}

// EnsureFreePlan ensures the free plan exists in the database for testing.
// Uses upsert to avoid conflicts when called multiple times.
func EnsureFreePlan(t *testing.T, client *ent.Client) (*ent.Plan, error) {
	t.Helper()

	ctx := context.Background()
	const freePlanID = "free"

	// Try to get existing plan first
	plan, err := client.Plan.Get(ctx, freePlanID)
	if err == nil {
		return plan, nil
	}

	// Create if not exists
	if ent.IsNotFound(err) {
		tokenLimit := 10000
		retentionDays := 7
		maxSessions := 3

		return client.Plan.
			Create().
			SetID(freePlanID).
			SetName("Free").
			SetPriceCents(0).
			SetBillingPeriod("monthly").
			SetTokenLimit(tokenLimit).
			SetSessionRetentionDays(retentionDays).
			SetMaxConcurrentSessions(maxSessions).
			SetFeatures(map[string]bool{
				"ai_mindmap":   true,
				"export_pdf":   false,
				"team_sharing": false,
			}).
			Save(ctx)
	}

	return nil, err
}
