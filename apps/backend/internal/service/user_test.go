package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent/user"
	"github.com/mindhit/api/internal/testutil"
)

// uniqueTestEmail generates a unique email for test isolation
func uniqueTestEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@example.com", prefix, uuid.New().String()[:8])
}

func TestUserCreate(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueTestEmail("create")

	// Create user
	u, err := client.User.Create().
		SetEmail(email).
		SetPasswordHash("hashedpassword").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Verify
	if u.Email != email {
		t.Errorf("expected email %s, got %s", email, u.Email)
	}
	if u.Status != user.StatusActive {
		t.Errorf("expected status active, got %s", u.Status)
	}
}

func TestUserSoftDelete(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueTestEmail("delete")

	// Create user
	u, err := client.User.Create().
		SetEmail(email).
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

	// Query by ID to verify status changed
	updated, err := client.User.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if updated.Status != user.StatusInactive {
		t.Errorf("expected status inactive, got %s", updated.Status)
	}
}
