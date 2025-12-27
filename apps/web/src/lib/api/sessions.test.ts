import { describe, it, expect, beforeEach } from "vitest";
import { sessionsApi } from "./sessions";
import { useAuthStore } from "@/stores/auth-store";

describe("sessionsApi", () => {
  beforeEach(() => {
    // Set up auth for API calls
    useAuthStore.getState().setAuth(
      { id: "user-1", email: "test@example.com" },
      "mock-access-token",
      "mock-refresh-token"
    );
  });

  describe("list", () => {
    it("returns sessions list", async () => {
      const result = await sessionsApi.list();

      expect(result.sessions).toBeDefined();
      expect(Array.isArray(result.sessions)).toBe(true);
    });

    it("accepts pagination params", async () => {
      const result = await sessionsApi.list(10, 5);

      expect(result.sessions).toBeDefined();
    });
  });

  describe("get", () => {
    it("returns single session", async () => {
      const result = await sessionsApi.get("session-1");

      expect(result).toBeDefined();
      expect(result.id).toBe("session-1");
    });

    it("throws error for not found", async () => {
      await expect(sessionsApi.get("not-found")).rejects.toThrow();
    });
  });

  describe("delete", () => {
    it("completes without error", async () => {
      await expect(sessionsApi.delete("session-1")).resolves.not.toThrow();
    });

    it("throws error for not found", async () => {
      await expect(sessionsApi.delete("not-found")).rejects.toThrow();
    });
  });
});
