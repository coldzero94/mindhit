import { describe, it, expect, beforeEach } from "vitest";
import { useAuthStore } from "./auth-store";

describe("useAuthStore", () => {
  beforeEach(() => {
    // Reset store before each test
    useAuthStore.getState().logout();
  });

  describe("initial state", () => {
    it("has null user initially", () => {
      expect(useAuthStore.getState().user).toBeNull();
    });

    it("has null tokens initially", () => {
      expect(useAuthStore.getState().accessToken).toBeNull();
      expect(useAuthStore.getState().refreshToken).toBeNull();
    });

    it("is not authenticated initially", () => {
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });

  describe("setAuth", () => {
    it("sets user and tokens", () => {
      const user = { id: "user-1", email: "test@example.com" };
      useAuthStore.getState().setAuth(user, "access-token", "refresh-token");

      const state = useAuthStore.getState();
      expect(state.user).toEqual(user);
      expect(state.accessToken).toBe("access-token");
      expect(state.refreshToken).toBe("refresh-token");
    });

    it("sets isAuthenticated to true", () => {
      const user = { id: "user-1", email: "test@example.com" };
      useAuthStore.getState().setAuth(user, "access-token", "refresh-token");

      expect(useAuthStore.getState().isAuthenticated).toBe(true);
    });
  });

  describe("setTokens", () => {
    it("updates tokens without changing user", () => {
      const user = { id: "user-1", email: "test@example.com" };
      useAuthStore.getState().setAuth(user, "old-access", "old-refresh");
      useAuthStore.getState().setTokens("new-access", "new-refresh");

      const state = useAuthStore.getState();
      expect(state.user).toEqual(user);
      expect(state.accessToken).toBe("new-access");
      expect(state.refreshToken).toBe("new-refresh");
    });
  });

  describe("logout", () => {
    it("clears all auth state", () => {
      const user = { id: "user-1", email: "test@example.com" };
      useAuthStore.getState().setAuth(user, "access-token", "refresh-token");
      useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.refreshToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });
  });
});
