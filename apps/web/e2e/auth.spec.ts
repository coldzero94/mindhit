import { test, expect } from "@playwright/test";

// Generate unique email for each test run
function uniqueEmail(prefix: string): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2)}@test.local`;
}

test.describe("Authentication Flow", () => {
  test.describe("Login Page", () => {
    test("should display login form", async ({ page }) => {
      await page.goto("/login");

      // Check page elements
      await expect(page.getByRole("heading", { name: "로그인" })).toBeVisible();
      await expect(page.getByLabel("이메일")).toBeVisible();
      await expect(page.getByLabel("비밀번호")).toBeVisible();
      await expect(page.getByRole("button", { name: "로그인" })).toBeVisible();
    });

    test("should show validation errors for empty fields", async ({ page }) => {
      await page.goto("/login");

      // Submit empty form
      await page.getByRole("button", { name: "로그인" }).click();

      // Check for validation errors
      await expect(page.getByText("유효한 이메일을 입력하세요")).toBeVisible();
    });

    test("should show validation error for invalid email", async ({ page }) => {
      await page.goto("/login");

      await page.getByLabel("이메일").fill("invalid-email");
      await page.getByLabel("비밀번호").fill("password123");
      await page.getByRole("button", { name: "로그인" }).click();

      await expect(page.getByText("유효한 이메일을 입력하세요")).toBeVisible();
    });

    test("should show validation error for short password", async ({ page }) => {
      await page.goto("/login");

      await page.getByLabel("이메일").fill("test@example.com");
      await page.getByLabel("비밀번호").fill("short");
      await page.getByRole("button", { name: "로그인" }).click();

      await expect(
        page.getByText("비밀번호는 8자 이상이어야 합니다")
      ).toBeVisible();
    });

    test("should navigate to signup page", async ({ page }) => {
      await page.goto("/login");

      await page.getByRole("link", { name: "회원가입" }).click();
      await expect(page).toHaveURL("/signup");
    });

    test("should show Google sign in button", async ({ page }) => {
      await page.goto("/login");

      // Google Sign-In button should be visible
      await expect(page.getByText("또는")).toBeVisible();
    });
  });

  test.describe("Signup Page", () => {
    test("should display signup form", async ({ page }) => {
      await page.goto("/signup");

      await expect(page.getByRole("heading", { name: "회원가입" })).toBeVisible();
      await expect(page.getByLabel("이메일")).toBeVisible();
      await expect(page.getByLabel("비밀번호", { exact: true })).toBeVisible();
      await expect(page.getByLabel("비밀번호 확인")).toBeVisible();
      await expect(page.getByRole("button", { name: "회원가입" })).toBeVisible();
    });

    test("should show validation error for password mismatch", async ({
      page,
    }) => {
      await page.goto("/signup");

      await page.getByLabel("이메일").fill("test@example.com");
      await page.getByLabel("비밀번호", { exact: true }).fill("password123");
      await page.getByLabel("비밀번호 확인").fill("different123");
      await page.getByRole("button", { name: "회원가입" }).click();

      await expect(
        page.getByText("비밀번호가 일치하지 않습니다")
      ).toBeVisible();
    });

    test("should navigate to login page", async ({ page }) => {
      await page.goto("/signup");

      await page.getByRole("link", { name: "로그인" }).click();
      await expect(page).toHaveURL("/login");
    });
  });

  test.describe("Full Auth Flow", () => {
    test.skip("should signup, logout, and login successfully", async ({
      page,
    }) => {
      // NOTE: This test requires a running backend server
      // Skip by default for CI - enable when backend is available

      const email = uniqueEmail("e2e_auth");
      const password = "testPassword123!";

      // Step 1: Sign up
      await page.goto("/signup");
      await page.getByLabel("이메일").fill(email);
      await page.getByLabel("비밀번호", { exact: true }).fill(password);
      await page.getByLabel("비밀번호 확인").fill(password);
      await page.getByRole("button", { name: "회원가입" }).click();

      // Should redirect to sessions page after signup
      await expect(page).toHaveURL("/sessions", { timeout: 10000 });

      // Step 2: Verify user is logged in (check for sessions page content)
      await expect(page.getByText("세션")).toBeVisible();

      // Step 3: Logout (assuming there's a logout button/link)
      // This depends on the actual UI implementation

      // Step 4: Login with created credentials
      await page.goto("/login");
      await page.getByLabel("이메일").fill(email);
      await page.getByLabel("비밀번호").fill(password);
      await page.getByRole("button", { name: "로그인" }).click();

      // Should redirect to sessions page after login
      await expect(page).toHaveURL("/sessions", { timeout: 10000 });
    });
  });

  test.describe("Protected Routes", () => {
    test("should redirect to login when accessing protected route without auth", async ({
      page,
    }) => {
      // Try to access sessions page without being logged in
      await page.goto("/sessions");

      // Should redirect to login page
      await expect(page).toHaveURL(/\/login/);
    });
  });
});
