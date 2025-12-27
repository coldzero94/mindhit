import { describe, it, expect, beforeEach } from "vitest";
import { useAuthStore } from "./auth-store";

describe("AuthStore", () => {
  beforeEach(() => {
    useAuthStore.getState().logout();
  });

  describe("initial state", () => {
    it("should not be authenticated initially", () => {
      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
    });
  });

  describe("setAuth", () => {
    it("should set user and token", () => {
      const user = { id: "user-1", email: "test@example.com" };
      const token = "test-token";

      useAuthStore.getState().setAuth(user, token);

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(true);
      expect(state.user).toEqual(user);
      expect(state.token).toBe(token);
    });
  });

  describe("logout", () => {
    it("should clear user and token", () => {
      const user = { id: "user-1", email: "test@example.com" };
      useAuthStore.getState().setAuth(user, "test-token");
      useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
    });
  });
});
