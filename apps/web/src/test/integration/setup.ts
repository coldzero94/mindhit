/**
 * Integration test setup - Real backend API calls
 * Communicates with actual server without using MSW
 */
import axios from "axios";

// API client for integration tests (no Zustand store dependency)
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000";

// Delay to prevent rate limiting
export const delay = (ms: number) =>
  new Promise((resolve) => setTimeout(resolve, ms));

export const testApiClient = axios.create({
  baseURL: `${API_BASE_URL}/v1`,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
});

// Generate unique email for tests
export function uniqueEmail(prefix: string): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2)}@test.local`;
}

// Helper for creating test users with authentication
export async function createTestUser(prefix: string = "integration") {
  // Prevent rate limiting
  await delay(200);

  const email = uniqueEmail(prefix);
  const password = "testPassword123!";

  try {
    const response = await testApiClient.post("/auth/signup", {
      email,
      password,
    });

    return {
      email,
      password,
      token: response.data.token,
      user: response.data.user,
    };
  } catch (error: unknown) {
    if (
      error &&
      typeof error === "object" &&
      "response" in error &&
      error.response &&
      typeof error.response === "object" &&
      "status" in error.response &&
      error.response.status === 429
    ) {
      throw new Error(
        "RATE_LIMITED: Too many requests. Wait 1 minute and restart the backend server."
      );
    }
    throw error;
  }
}

// Create authenticated API client
export function createAuthenticatedClient(token: string) {
  const client = axios.create({
    baseURL: `${API_BASE_URL}/v1`,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    timeout: 10000,
  });
  return client;
}

// Check backend server health
export async function checkServerHealth(): Promise<boolean> {
  try {
    // Check with a simple request if health endpoint is not available
    await testApiClient.get("/auth/me", {
      validateStatus: () => true, // Don't treat any status code as an error
    });
    return true;
  } catch (error) {
    if (axios.isAxiosError(error) && error.code === "ECONNREFUSED") {
      return false;
    }
    return true; // Other errors (like 401) mean the server is running
  }
}

// Delete test user (hard delete - only allowed in test environment)
export async function deleteTestUser(token: string): Promise<void> {
  try {
    const authClient = createAuthenticatedClient(token);
    await authClient.delete("/auth/me", {
      params: { hard: true },
    });
  } catch (error) {
    // Ignore if already deleted or token expired
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return;
    }
    // Ignore other errors during test cleanup (just log)
    console.warn("Failed to delete test user:", error);
  }
}

// Helper to track and clean up users created during tests
export class TestUserManager {
  private createdUsers: Array<{ token: string; email: string }> = [];

  async createUser(prefix: string = "integration") {
    const user = await createTestUser(prefix);
    this.createdUsers.push({ token: user.token, email: user.email });
    return user;
  }

  async cleanup() {
    for (const user of this.createdUsers) {
      await deleteTestUser(user.token);
    }
    this.createdUsers = [];
  }
}
