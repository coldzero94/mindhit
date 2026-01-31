import { describe, it, expect, beforeEach, vi } from "vitest";
import { api } from "./api";
import type { BrowsingEvent } from "@/types";

// Mock fetch globally
const mockFetch = vi.fn();
globalThis.fetch = mockFetch as typeof fetch;

describe("API Client", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("startSession", () => {
    it("should start session successfully", async () => {
      const mockResponse = {
        session: {
          id: "session-1",
          session_status: "recording",
          started_at: "2024-01-01T00:00:00Z",
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.startSession();

      expect(result.session).toBeDefined();
      expect(result.session.id).toBe("session-1");
    });
  });

  describe("pauseSession", () => {
    it("should pause session successfully", async () => {
      const mockResponse = {
        session: {
          id: "session-1",
          session_status: "paused",
          started_at: "2024-01-01T00:00:00Z",
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.pauseSession("session-1");

      expect(result.session.session_status).toBe("paused");
    });
  });

  describe("resumeSession", () => {
    it("should resume session successfully", async () => {
      const mockResponse = {
        session: {
          id: "session-1",
          session_status: "recording",
          started_at: "2024-01-01T00:00:00Z",
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.resumeSession("session-1");

      expect(result.session.session_status).toBe("recording");
    });
  });

  describe("stopSession", () => {
    it("should stop session successfully", async () => {
      const mockResponse = {
        session: {
          id: "session-1",
          session_status: "completed",
          started_at: "2024-01-01T00:00:00Z",
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.stopSession("session-1");

      expect(result.session.session_status).toBe("completed");
    });
  });

  describe("sendEvents", () => {
    it("should send events successfully", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      const events: BrowsingEvent[] = [
        {
          type: "page_visit",
          timestamp: 1704067200000, // 2024-01-01T00:00:00Z in milliseconds
          url: "https://example.com",
          title: "Example",
          referrer: "",
        },
      ];

      await expect(
        api.sendEvents("session-1", events)
      ).resolves.not.toThrow();
    });
  });

  describe("getSessions", () => {
    it("should get sessions list successfully", async () => {
      const mockResponse = {
        sessions: [
          {
            id: "session-1",
            session_status: "completed",
            started_at: "2024-01-01T00:00:00Z",
          },
        ],
        total: 1,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.getSessions();

      expect(result.sessions).toBeDefined();
      expect(result.sessions.length).toBe(1);
      expect(result.total).toBe(1);
    });

    it("should respect limit parameter", async () => {
      const mockResponse = {
        sessions: [
          {
            id: "session-1",
            session_status: "completed",
            started_at: "2024-01-01T00:00:00Z",
          },
          {
            id: "session-2",
            session_status: "completed",
            started_at: "2024-01-01T00:00:00Z",
          },
        ],
        total: 2,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.getSessions(2);

      expect(result.sessions.length).toBe(2);
    });
  });

  describe("updateSession", () => {
    it("should update session title successfully", async () => {
      const mockResponse = {
        session: {
          id: "session-1",
          title: "New Title",
          session_status: "completed",
          started_at: "2024-01-01T00:00:00Z",
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.updateSession("session-1", {
        title: "New Title",
      });

      expect(result.session).toBeDefined();
      expect(result.session.title).toBe("New Title");
    });
  });

  describe("error handling", () => {
    it("should throw error on failed request", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          error: { message: "Unauthorized" },
        }),
      });

      await expect(api.startSession()).rejects.toThrow("Unauthorized");
    });
  });
});
