import { describe, it, expect } from "vitest";
import { api } from "./api";
import type { BrowsingEvent } from "@/types";

describe("API Client", () => {
  describe("login", () => {
    it("should login successfully with valid credentials", async () => {
      const result = await api.login("test@example.com", "password123");

      expect(result.user).toBeDefined();
      expect(result.user.email).toBe("test@example.com");
      expect(result.token).toBe("mock-access-token");
    });

    it("should throw error with invalid credentials", async () => {
      await expect(
        api.login("test@example.com", "wrongpassword")
      ).rejects.toThrow("Invalid credentials");
    });
  });

  describe("getSessions", () => {
    it("should get sessions list successfully", async () => {
      const result = await api.getSessions("mock-token");

      expect(result.sessions).toBeDefined();
      expect(result.sessions.length).toBeGreaterThan(0);
      expect(result.total).toBe(3);
    });

    it("should respect limit parameter", async () => {
      const result = await api.getSessions("mock-token", 2);

      expect(result.sessions.length).toBe(2);
    });

    it("should throw error without auth token", async () => {
      await expect(api.getSessions("")).rejects.toThrow();
    });
  });

  describe("updateSession", () => {
    it("should update session title successfully", async () => {
      const result = await api.updateSession("mock-token", "session-1", {
        title: "New Title",
      });

      expect(result.session).toBeDefined();
      expect(result.session.title).toBe("New Title");
    });

    it("should throw error for non-existent session", async () => {
      await expect(
        api.updateSession("mock-token", "not-found", { title: "Test" })
      ).rejects.toThrow("Session not found");
    });

    it("should throw error without auth token", async () => {
      await expect(
        api.updateSession("", "session-1", { title: "Test" })
      ).rejects.toThrow();
    });
  });

  describe("startSession", () => {
    it("should start a session successfully", async () => {
      const result = await api.startSession("mock-token");

      expect(result.session).toBeDefined();
      expect(result.session.id).toBe("session-new");
      expect(result.session.session_status).toBe("recording");
    });

    it("should throw error without auth token", async () => {
      await expect(api.startSession("")).rejects.toThrow();
    });
  });

  describe("pauseSession", () => {
    it("should pause a session successfully", async () => {
      const result = await api.pauseSession("mock-token", "session-1");

      expect(result.session).toBeDefined();
      expect(result.session.session_status).toBe("paused");
    });

    it("should throw error for non-existent session", async () => {
      await expect(
        api.pauseSession("mock-token", "not-found")
      ).rejects.toThrow("Session not found");
    });
  });

  describe("resumeSession", () => {
    it("should resume a session successfully", async () => {
      const result = await api.resumeSession("mock-token", "session-1");

      expect(result.session).toBeDefined();
      expect(result.session.session_status).toBe("recording");
    });

    it("should throw error for non-existent session", async () => {
      await expect(
        api.resumeSession("mock-token", "not-found")
      ).rejects.toThrow("Session not found");
    });
  });

  describe("stopSession", () => {
    it("should stop a session successfully", async () => {
      const result = await api.stopSession("mock-token", "session-1");

      expect(result.session).toBeDefined();
      expect(result.session.session_status).toBe("completed");
      expect(result.session.ended_at).toBeDefined();
    });

    it("should throw error for non-existent session", async () => {
      await expect(
        api.stopSession("mock-token", "not-found")
      ).rejects.toThrow("Session not found");
    });
  });

  describe("sendEvents", () => {
    it("should send events successfully", async () => {
      const events: BrowsingEvent[] = [
        {
          type: "page_visit",
          timestamp: Date.now(),
          url: "https://example.com",
          title: "Example Page",
          referrer: "",
        },
      ];

      // Should not throw
      await expect(
        api.sendEvents("mock-token", "session-1", events)
      ).resolves.toBeUndefined();
    });

    it("should throw error without auth token", async () => {
      const events: BrowsingEvent[] = [
        {
          type: "page_visit",
          timestamp: Date.now(),
          url: "https://example.com",
          title: "Example Page",
          referrer: "",
        },
      ];

      await expect(api.sendEvents("", "session-1", events)).rejects.toThrow();
    });
  });
});
