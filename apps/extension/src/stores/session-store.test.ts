import { describe, it, expect, beforeEach } from "vitest";
import { useSessionStore } from "./session-store";

describe("SessionStore", () => {
  beforeEach(() => {
    useSessionStore.getState().reset();
  });

  describe("initial state", () => {
    it("should have idle status initially", () => {
      const state = useSessionStore.getState();
      expect(state.status).toBe("idle");
      expect(state.sessionId).toBeNull();
      expect(state.pageCount).toBe(0);
      expect(state.highlightCount).toBe(0);
    });
  });

  describe("startSession", () => {
    it("should set session id and status to recording", () => {
      useSessionStore.getState().startSession("test-session-id");
      const state = useSessionStore.getState();

      expect(state.sessionId).toBe("test-session-id");
      expect(state.status).toBe("recording");
      expect(state.startedAt).not.toBeNull();
      expect(state.pageCount).toBe(0);
      expect(state.highlightCount).toBe(0);
    });
  });

  describe("pauseSession", () => {
    it("should set status to paused", () => {
      useSessionStore.getState().startSession("test-session-id");
      useSessionStore.getState().pauseSession();

      expect(useSessionStore.getState().status).toBe("paused");
      expect(useSessionStore.getState().sessionId).toBe("test-session-id");
    });
  });

  describe("resumeSession", () => {
    it("should set status back to recording", () => {
      useSessionStore.getState().startSession("test-session-id");
      useSessionStore.getState().pauseSession();
      useSessionStore.getState().resumeSession();

      expect(useSessionStore.getState().status).toBe("recording");
    });
  });

  describe("stopSession", () => {
    it("should reset session state", () => {
      useSessionStore.getState().startSession("test-session-id");
      useSessionStore.getState().incrementPageCount();
      useSessionStore.getState().stopSession();

      const state = useSessionStore.getState();
      expect(state.sessionId).toBeNull();
      expect(state.status).toBe("idle");
      expect(state.startedAt).toBeNull();
    });
  });

  describe("incrementPageCount", () => {
    it("should increment page count", () => {
      useSessionStore.getState().incrementPageCount();
      useSessionStore.getState().incrementPageCount();

      expect(useSessionStore.getState().pageCount).toBe(2);
    });
  });

  describe("incrementHighlightCount", () => {
    it("should increment highlight count", () => {
      useSessionStore.getState().incrementHighlightCount();
      useSessionStore.getState().incrementHighlightCount();
      useSessionStore.getState().incrementHighlightCount();

      expect(useSessionStore.getState().highlightCount).toBe(3);
    });
  });

  describe("updateElapsedTime", () => {
    it("should update elapsed time when recording", () => {
      // Start session
      useSessionStore.getState().startSession("test-session-id");

      // Manually set startedAt to be 5 seconds ago
      const fiveSecondsAgo = Date.now() - 5000;
      useSessionStore.setState({ startedAt: fiveSecondsAgo });

      // Update elapsed time
      useSessionStore.getState().updateElapsedTime();

      expect(useSessionStore.getState().elapsedSeconds).toBeGreaterThanOrEqual(
        4
      );
      expect(useSessionStore.getState().elapsedSeconds).toBeLessThanOrEqual(6);
    });

    it("should not update elapsed time when paused", () => {
      useSessionStore.getState().startSession("test-session-id");
      useSessionStore.getState().pauseSession();

      const elapsedBefore = useSessionStore.getState().elapsedSeconds;
      useSessionStore.getState().updateElapsedTime();
      const elapsedAfter = useSessionStore.getState().elapsedSeconds;

      expect(elapsedAfter).toBe(elapsedBefore);
    });
  });

  describe("reset", () => {
    it("should reset all state", () => {
      useSessionStore.getState().startSession("test-session-id");
      useSessionStore.getState().incrementPageCount();
      useSessionStore.getState().incrementHighlightCount();
      useSessionStore.getState().reset();

      const state = useSessionStore.getState();
      expect(state.sessionId).toBeNull();
      expect(state.status).toBe("idle");
      expect(state.startedAt).toBeNull();
      expect(state.pageCount).toBe(0);
      expect(state.highlightCount).toBe(0);
      expect(state.elapsedSeconds).toBe(0);
    });
  });
});
