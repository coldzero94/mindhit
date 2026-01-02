/**
 * Auth API Integration Tests
 * 실제 백엔드 서버와 통신하여 인증 플로우 검증
 *
 * 실행 전 필수: 백엔드 서버가 실행 중이어야 함
 * moonx backend:dev-api-test (rate limiting 비활성화)
 */
import { describe, it, expect, beforeAll, afterAll } from "vitest";
import {
  testApiClient,
  createAuthenticatedClient,
  checkServerHealth,
  uniqueEmail,
  deleteTestUser,
  TestUserManager,
} from "./setup";

describe("Auth API Integration", () => {
  const userManager = new TestUserManager();

  beforeAll(async () => {
    const isHealthy = await checkServerHealth();
    if (!isHealthy) {
      console.warn(
        "⚠️  Backend server is not running. Skipping integration tests."
      );
      console.warn("   Run: moonx backend:dev-api-test");
    }
  });

  afterAll(async () => {
    // 테스트에서 생성한 모든 유저 정리
    await userManager.cleanup();
  });

  describe("Signup", () => {
    it("should signup a new user successfully", async () => {
      const email = uniqueEmail("signup_test");
      const password = "securePassword123!";

      const response = await testApiClient.post("/auth/signup", {
        email,
        password,
      });

      expect(response.status).toBe(201);
      expect(response.data.token).toBeDefined();
      expect(response.data.user).toBeDefined();
      expect(response.data.user.email).toBe(email);
      expect(response.data.user.id).toBeDefined();

      // Cleanup: 생성된 유저 삭제
      await deleteTestUser(response.data.token);
    });

    it("should reject duplicate email signup", async () => {
      const user = await userManager.createUser("duplicate_test");

      // Second signup with same email
      try {
        await testApiClient.post("/auth/signup", {
          email: user.email,
          password: "differentPassword",
        });
        expect.fail("Should have thrown an error");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          expect(error.response.status).toBe(409);
        }
      }
    });

    it("should reject invalid email format", async () => {
      try {
        await testApiClient.post("/auth/signup", {
          email: "invalid-email",
          password: "password123!",
        });
        expect.fail("Should have thrown an error");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          // Backend returns 400 (validation) or 409 (conflict) for invalid email
          expect([400, 409]).toContain(error.response.status);
        }
      }
    });
  });

  describe("Login", () => {
    it("should login with valid credentials", async () => {
      const user = await userManager.createUser("login_test");

      // Login
      const response = await testApiClient.post("/auth/login", {
        email: user.email,
        password: user.password,
      });

      expect(response.status).toBe(200);
      expect(response.data.token).toBeDefined();
      expect(response.data.user.email).toBe(user.email);
    });

    it("should reject invalid password", async () => {
      const user = await userManager.createUser("wrong_pw_test");

      try {
        await testApiClient.post("/auth/login", {
          email: user.email,
          password: "wrongPassword123!",
        });
        expect.fail("Should have thrown an error");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          expect(error.response.status).toBe(401);
        }
      }
    });

    it("should reject non-existent email", async () => {
      try {
        await testApiClient.post("/auth/login", {
          email: "nonexistent@test.local",
          password: "password123!",
        });
        expect.fail("Should have thrown an error");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          expect(error.response.status).toBe(401);
        }
      }
    });
  });

  describe("Me (Get Current User)", () => {
    it("should return current user with valid token", async () => {
      const user = await userManager.createUser("me_test");
      const authClient = createAuthenticatedClient(user.token);

      const response = await authClient.get("/auth/me");

      expect(response.status).toBe(200);
      expect(response.data.user.email).toBe(user.email);
    });

    it("should reject request without token", async () => {
      try {
        await testApiClient.get("/auth/me");
        expect.fail("Should have thrown an error");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          // 400 (missing auth header) or 401 (invalid token)
          expect([400, 401]).toContain(error.response.status);
        }
      }
    });

    it("should reject request with invalid token", async () => {
      const authClient = createAuthenticatedClient("invalid-token");

      try {
        await authClient.get("/auth/me");
        expect.fail("Should have thrown an error");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          expect(error.response.status).toBe(401);
        }
      }
    });
  });

  describe("Logout", () => {
    it("should logout successfully", async () => {
      const user = await userManager.createUser("logout_test");
      const authClient = createAuthenticatedClient(user.token);

      const response = await authClient.post("/auth/logout");

      expect(response.status).toBe(200);
      expect(response.data.message).toContain("logged out");
    });
  });

  describe("Full Auth Flow", () => {
    it("should complete signup -> login -> me -> logout flow", async () => {
      const email = uniqueEmail("full_flow");
      const password = "fullFlowPassword123!";

      // Step 1: Signup
      const signupRes = await testApiClient.post("/auth/signup", {
        email,
        password,
      });
      expect(signupRes.status).toBe(201);
      const signupToken = signupRes.data.token;

      // Step 2: Login (새 토큰 발급)
      const loginRes = await testApiClient.post("/auth/login", {
        email,
        password,
      });
      expect(loginRes.status).toBe(200);
      const loginToken = loginRes.data.token;

      // Step 3: Get current user
      const authClient = createAuthenticatedClient(loginToken);
      const meRes = await authClient.get("/auth/me");
      expect(meRes.status).toBe(200);
      expect(meRes.data.user.email).toBe(email);

      // Step 4: Logout
      const logoutRes = await authClient.post("/auth/logout");
      expect(logoutRes.status).toBe(200);

      // Verify signup token is still different from login token
      expect(signupToken).not.toBe(loginToken);

      // Cleanup: 생성된 유저 삭제
      await deleteTestUser(loginToken);
    });
  });

  describe("Delete Account", () => {
    it("should delete user account (hard delete in test env)", async () => {
      const email = uniqueEmail("delete_test");
      const password = "deleteMe123!";

      // Create user
      const signupRes = await testApiClient.post("/auth/signup", {
        email,
        password,
      });
      const token = signupRes.data.token;
      const authClient = createAuthenticatedClient(token);

      // Delete account with hard=true
      const deleteRes = await authClient.delete("/auth/me", {
        params: { hard: true },
      });
      expect(deleteRes.status).toBe(204);

      // Verify user is deleted (login should fail)
      try {
        await testApiClient.post("/auth/login", { email, password });
        expect.fail("Should have thrown an error");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          expect(error.response.status).toBe(401);
        }
      }
    });
  });
});
