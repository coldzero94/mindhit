/**
 * Sessions API Integration Tests
 * 실제 백엔드 서버와 통신하여 세션 플로우 검증
 *
 * 실행 전 필수: 백엔드 서버가 실행 중이어야 함
 * moonx backend:dev-api-test (rate limiting 비활성화)
 */
import { describe, it, expect, beforeAll, afterAll, beforeEach } from "vitest";
import { AxiosInstance } from "axios";
import {
  createAuthenticatedClient,
  checkServerHealth,
  TestUserManager,
} from "./setup";

describe("Sessions API Integration", () => {
  const userManager = new TestUserManager();
  let authClient: AxiosInstance;

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

  beforeEach(async () => {
    // 각 테스트마다 새 유저 생성
    const user = await userManager.createUser("session_test");
    authClient = createAuthenticatedClient(user.token);
  });

  describe("Session Lifecycle", () => {
    it("should start a new session", async () => {
      const response = await authClient.post("/sessions/start");

      expect(response.status).toBe(201);
      expect(response.data.session).toBeDefined();
      expect(response.data.session.id).toBeDefined();
      expect(response.data.session.session_status).toBe("recording");
    });

    it("should pause a recording session", async () => {
      // Start session
      const startRes = await authClient.post("/sessions/start");
      const sessionId = startRes.data.session.id;

      // Pause session
      const pauseRes = await authClient.patch(`/sessions/${sessionId}/pause`);

      expect(pauseRes.status).toBe(200);
      expect(pauseRes.data.session.session_status).toBe("paused");
    });

    it("should resume a paused session", async () => {
      // Start session
      const startRes = await authClient.post("/sessions/start");
      const sessionId = startRes.data.session.id;

      // Pause session
      await authClient.patch(`/sessions/${sessionId}/pause`);

      // Resume session
      const resumeRes = await authClient.patch(`/sessions/${sessionId}/resume`);

      expect(resumeRes.status).toBe(200);
      expect(resumeRes.data.session.session_status).toBe("recording");
    });

    it("should stop a session", async () => {
      // Start session
      const startRes = await authClient.post("/sessions/start");
      const sessionId = startRes.data.session.id;

      // Stop session
      const stopRes = await authClient.post(`/sessions/${sessionId}/stop`);

      expect(stopRes.status).toBe(200);
      // Status could be "processing" or "completed" depending on queue
      expect(["processing", "completed"]).toContain(
        stopRes.data.session.session_status
      );
    });
  });

  describe("Session CRUD", () => {
    it("should list sessions", async () => {
      // Create a session first
      await authClient.post("/sessions/start");

      // List sessions
      const response = await authClient.get("/sessions");

      expect(response.status).toBe(200);
      expect(response.data.sessions).toBeDefined();
      expect(Array.isArray(response.data.sessions)).toBe(true);
      expect(response.data.sessions.length).toBeGreaterThanOrEqual(1);
    });

    it("should get a single session", async () => {
      // Create a session
      const startRes = await authClient.post("/sessions/start");
      const sessionId = startRes.data.session.id;

      // Get session
      const getRes = await authClient.get(`/sessions/${sessionId}`);

      expect(getRes.status).toBe(200);
      expect(getRes.data.session.id).toBe(sessionId);
    });

    it("should update session title and description", async () => {
      // Create a session
      const startRes = await authClient.post("/sessions/start");
      const sessionId = startRes.data.session.id;

      // Update session
      const updateRes = await authClient.put(`/sessions/${sessionId}`, {
        title: "Updated Title",
        description: "Updated Description",
      });

      expect(updateRes.status).toBe(200);
      expect(updateRes.data.session.title).toBe("Updated Title");
      expect(updateRes.data.session.description).toBe("Updated Description");
    });

    it("should delete a session", async () => {
      // Create a session
      const startRes = await authClient.post("/sessions/start");
      const sessionId = startRes.data.session.id;

      // Delete session
      const deleteRes = await authClient.delete(`/sessions/${sessionId}`);
      expect(deleteRes.status).toBe(204);

      // Verify deletion
      try {
        await authClient.get(`/sessions/${sessionId}`);
        expect.fail("Should have thrown 404");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          expect(error.response.status).toBe(404);
        }
      }
    });
  });

  describe("Session Pagination", () => {
    it("should paginate sessions with limit and offset", async () => {
      // Create multiple sessions
      await authClient.post("/sessions/start");
      await authClient.post("/sessions/start");
      await authClient.post("/sessions/start");

      // Get first page
      const page1 = await authClient.get("/sessions", {
        params: { limit: 2, offset: 0 },
      });

      expect(page1.status).toBe(200);
      expect(page1.data.sessions.length).toBeLessThanOrEqual(2);

      // Get second page
      const page2 = await authClient.get("/sessions", {
        params: { limit: 2, offset: 2 },
      });

      expect(page2.status).toBe(200);
    });
  });

  describe("Session Authorization", () => {
    it("should not allow accessing other user's session", async () => {
      // Create session with first user
      const startRes = await authClient.post("/sessions/start");
      const sessionId = startRes.data.session.id;

      // Create second user
      const user2 = await userManager.createUser("other_user");
      const authClient2 = createAuthenticatedClient(user2.token);

      // Try to access first user's session
      try {
        await authClient2.get(`/sessions/${sessionId}`);
        expect.fail("Should have thrown 404 or 403");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          // 다른 유저의 세션은 404 (보안상 403 대신 404 반환)
          expect([403, 404]).toContain(error.response.status);
        }
      }
    });

    it("should reject requests without authentication", async () => {
      const { default: axios } = await import("axios");
      const unauthClient = axios.create({
        baseURL: "http://localhost:9000/v1",
      });

      try {
        await unauthClient.get("/sessions");
        expect.fail("Should have thrown 401");
      } catch (error: unknown) {
        if (
          error &&
          typeof error === "object" &&
          "response" in error &&
          error.response &&
          typeof error.response === "object" &&
          "status" in error.response
        ) {
          // 400 (missing auth header) or 401 (invalid/no token)
          expect([400, 401]).toContain(error.response.status);
        }
      }
    });
  });

  describe("Full Session Flow", () => {
    it("should complete create -> update -> pause -> resume -> stop -> delete flow", async () => {
      // Step 1: Start session
      const startRes = await authClient.post("/sessions/start");
      expect(startRes.status).toBe(201);
      const sessionId = startRes.data.session.id;

      // Step 2: Update title
      const updateRes = await authClient.put(`/sessions/${sessionId}`, {
        title: "Integration Test Session",
      });
      expect(updateRes.status).toBe(200);
      expect(updateRes.data.session.title).toBe("Integration Test Session");

      // Step 3: Pause
      const pauseRes = await authClient.patch(`/sessions/${sessionId}/pause`);
      expect(pauseRes.status).toBe(200);
      expect(pauseRes.data.session.session_status).toBe("paused");

      // Step 4: Resume
      const resumeRes = await authClient.patch(`/sessions/${sessionId}/resume`);
      expect(resumeRes.status).toBe(200);
      expect(resumeRes.data.session.session_status).toBe("recording");

      // Step 5: Stop
      const stopRes = await authClient.post(`/sessions/${sessionId}/stop`);
      expect(stopRes.status).toBe(200);

      // Step 6: Delete
      const deleteRes = await authClient.delete(`/sessions/${sessionId}`);
      expect(deleteRes.status).toBe(204);
    });
  });
});
