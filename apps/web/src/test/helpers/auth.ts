import { useAuthStore } from "@/stores/auth-store";
import { mockUser } from "../mocks/handlers";

/**
 * Set up authenticated state for tests
 */
export function setupAuthenticatedUser(
  overrides: Partial<typeof mockUser> = {}
) {
  const user = { ...mockUser, ...overrides };
  useAuthStore.getState().setAuth(user, "mock-access-token", "mock-refresh-token");
  return user;
}

/**
 * Clear authentication state
 */
export function clearAuth() {
  useAuthStore.getState().logout();
}

/**
 * Get current auth state
 */
export function getAuthState() {
  return useAuthStore.getState();
}
