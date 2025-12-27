import { describe, it, expect, beforeEach } from "vitest";
import { authApi } from "./auth";
import { useAuthStore } from "@/stores/auth-store";

describe("authApi", () => {
  beforeEach(() => {
    useAuthStore.getState().logout();
  });

  describe("login", () => {
    it("returns user and token on success", async () => {
      const result = await authApi.login({
        email: "test@example.com",
        password: "password123",
      });

      expect(result.user).toBeDefined();
      expect(result.user.email).toBe("test@example.com");
      expect(result.token).toBe("mock-access-token");
    });

    it("throws error for wrong password", async () => {
      await expect(
        authApi.login({
          email: "test@example.com",
          password: "wrongpassword",
        })
      ).rejects.toThrow();
    });
  });

  describe("signup", () => {
    it("returns user and token on success", async () => {
      const result = await authApi.signup({
        email: "newuser@example.com",
        password: "password123",
      });

      expect(result.user).toBeDefined();
      expect(result.user.email).toBe("newuser@example.com");
      expect(result.token).toBe("mock-access-token");
    });

    it("throws error for existing email", async () => {
      await expect(
        authApi.signup({
          email: "existing@example.com",
          password: "password123",
        })
      ).rejects.toThrow();
    });
  });

  describe("me", () => {
    it("returns current user", async () => {
      // Set up auth first
      useAuthStore.getState().setAuth(
        { id: "user-1", email: "test@example.com" },
        "mock-access-token",
        "mock-refresh-token"
      );

      const user = await authApi.me();

      expect(user).toBeDefined();
      expect(user.id).toBe("user-1");
    });
  });

  describe("logout", () => {
    it("completes without error", async () => {
      await expect(authApi.logout()).resolves.not.toThrow();
    });
  });
});
