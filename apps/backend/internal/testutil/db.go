package testutil

import (
	"context"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/lib/pq"

	"github.com/mindhit/api/ent"
)

// getTestDatabaseURL returns the test database URL from environment or default
func getTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://postgres:password@localhost:5432/mindhit_test?sslmode=disable"
}

// SetupTestDB creates a test database client with PostgreSQL
// It also cleans up existing data for a fresh test environment
func SetupTestDB(t *testing.T) *ent.Client {
	t.Helper()
	client, err := ent.Open(dialect.Postgres, getTestDatabaseURL())
	if err != nil {
		t.Fatalf("failed to open postgres: %v", err)
	}

	ctx := context.Background()

	// Auto migrate schema
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Clean up all tables for fresh test
	cleanupTables(t, client)

	return client
}

// cleanupTables deletes all data from tables (order matters due to foreign keys)
func cleanupTables(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	// Delete in reverse dependency order
	if _, err := client.MindmapGraph.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean mindmap_graphs: %v", err)
	}
	if _, err := client.RawEvent.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean raw_events: %v", err)
	}
	if _, err := client.Highlight.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean highlights: %v", err)
	}
	if _, err := client.PageVisit.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean page_visits: %v", err)
	}
	if _, err := client.URL.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean urls: %v", err)
	}
	if _, err := client.Session.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean sessions: %v", err)
	}
	if _, err := client.UserSettings.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean user_settings: %v", err)
	}
	if _, err := client.User.Delete().Exec(ctx); err != nil {
		t.Logf("failed to clean users: %v", err)
	}
}

// CleanupTestDB closes the test database client
func CleanupTestDB(t *testing.T, client *ent.Client) {
	t.Helper()
	if err := client.Close(); err != nil {
		t.Errorf("failed to close client: %v", err)
	}
}
