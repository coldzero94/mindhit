import { test, expect } from "@playwright/test";

test.describe("Dashboard", () => {
  test.describe("Sessions Page (Unauthenticated)", () => {
    test("should redirect to login when not authenticated", async ({
      page,
    }) => {
      await page.goto("/sessions");

      // Should redirect to login
      await expect(page).toHaveURL(/\/login/);
    });

    test("should redirect from session detail to login when not authenticated", async ({
      page,
    }) => {
      await page.goto("/sessions/some-session-id");

      // Should redirect to login
      await expect(page).toHaveURL(/\/login/);
    });
  });

  test.describe("Account Page (Unauthenticated)", () => {
    test("should redirect to login when not authenticated", async ({
      page,
    }) => {
      await page.goto("/account");

      // Should redirect to login
      await expect(page).toHaveURL(/\/login/);
    });
  });
});

test.describe("Dashboard (Authenticated)", () => {
  // NOTE: These tests require authentication setup
  // In a real scenario, you would use:
  // 1. A test user with known credentials
  // 2. beforeEach hook to login
  // 3. Or use storageState to persist auth

  test.skip("should display sessions list when authenticated", async ({
    page,
  }) => {
    // This test would require auth setup
    await page.goto("/sessions");

    await expect(page.getByRole("heading", { name: "내 세션" })).toBeVisible();
  });

  test.skip("should display session detail page", async ({ page }) => {
    // This test would require:
    // 1. Auth setup
    // 2. A known session ID in the test database
    await page.goto("/sessions/test-session-id");

    // Should show session detail elements
    await expect(page.getByText("방문한 페이지")).toBeVisible();
    await expect(page.getByText("하이라이트")).toBeVisible();
    await expect(page.getByRole("tab", { name: "이벤트" })).toBeVisible();
    await expect(page.getByRole("tab", { name: "마인드맵" })).toBeVisible();
  });

  test.skip("should navigate between tabs in session detail", async ({
    page,
  }) => {
    await page.goto("/sessions/test-session-id");

    // Click on mindmap tab
    await page.getByRole("tab", { name: "마인드맵" }).click();

    // Verify mindmap tab is active
    await expect(page.getByRole("tab", { name: "마인드맵" })).toHaveAttribute(
      "aria-selected",
      "true"
    );
  });

  test.skip("should show delete confirmation dialog", async ({ page }) => {
    await page.goto("/sessions/test-session-id");

    // Click delete button
    await page.getByRole("button", { name: /delete|삭제/i }).click();

    // Should show confirmation dialog
    await expect(page.getByText("세션을 삭제하시겠습니까?")).toBeVisible();
    await expect(page.getByRole("button", { name: "취소" })).toBeVisible();
    await expect(page.getByRole("button", { name: "삭제" })).toBeVisible();
  });

  test.skip("should cancel delete when clicking cancel", async ({ page }) => {
    await page.goto("/sessions/test-session-id");

    // Click delete button
    await page.getByRole("button", { name: /delete|삭제/i }).click();

    // Click cancel
    await page.getByRole("button", { name: "취소" }).click();

    // Dialog should close
    await expect(page.getByText("세션을 삭제하시겠습니까?")).not.toBeVisible();
  });
});

test.describe("Navigation", () => {
  test("should display home page", async ({ page }) => {
    await page.goto("/");

    // Home page should be accessible
    await expect(page).toHaveURL("/");
  });
});
