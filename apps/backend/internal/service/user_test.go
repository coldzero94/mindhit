package service_test

import (
	"context"
	"testing"

	"github.com/mindhit/api/ent/user"
	"github.com/mindhit/api/internal/testutil"
)

func TestUserCreate(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Create user
	u, err := client.User.Create().
		SetEmail("test@example.com").
		SetPasswordHash("hashedpassword").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Verify
	if u.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", u.Email)
	}
	if u.Status != user.StatusActive {
		t.Errorf("expected status active, got %s", u.Status)
	}
}

func TestUserSoftDelete(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Create user
	u, err := client.User.Create().
		SetEmail("delete@example.com").
		SetPasswordHash("hashedpassword").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Soft delete
	_, err = client.User.UpdateOneID(u.ID).
		SetStatus(user.StatusInactive).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to soft delete user: %v", err)
	}

	// Query active users - should not find deleted user
	activeUsers, err := client.User.Query().
		Where(user.StatusEQ(user.StatusActive)).
		All(ctx)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if len(activeUsers) != 0 {
		t.Errorf("expected 0 active users, got %d", len(activeUsers))
	}
}
