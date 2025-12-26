// Package fixture provides test data creation helpers.
package fixture

import (
	"context"
	"testing"

	"github.com/mindhit/api/ent"
)

// CreateTestUser creates a user for testing
func CreateTestUser(t *testing.T, client *ent.Client, email string) *ent.User {
	t.Helper()
	user, err := client.User.Create().
		SetEmail(email).
		SetPasswordHash("$2a$10$testhashedpassword").
		Save(context.Background())
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

// CreateTestSession creates a session for testing
func CreateTestSession(t *testing.T, client *ent.Client, user *ent.User) *ent.Session {
	t.Helper()
	session, err := client.Session.Create().
		SetUser(user).
		Save(context.Background())
	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}
	return session
}
