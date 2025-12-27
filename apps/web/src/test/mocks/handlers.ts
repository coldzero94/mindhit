import { http, HttpResponse } from "msw";

// Handle both possible API URLs (env may not be loaded in tests)
const API_URLS = [
  "http://localhost:9000/v1",
  "http://localhost:8080/v1",
];

// Mock data
export const mockUser = {
  id: "user-1",
  email: "test@example.com",
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

export const mockSession = {
  id: "session-1",
  title: "Test Session",
  description: undefined,
  session_status: "completed" as const,
  started_at: "2025-01-01T10:00:00Z",
  ended_at: "2025-01-01T12:00:00Z",
  created_at: "2025-01-01T10:00:00Z",
  updated_at: "2025-01-01T12:00:00Z",
};

export const mockSessionWithDetails = {
  ...mockSession,
  event_count: 10,
  page_visit_count: 5,
  highlight_count: 3,
};

// Helper to create handlers for both URLs
function createHandlers(apiUrl: string) {
  return [
    // Auth endpoints
    http.post(`${apiUrl}/auth/signup`, async ({ request }) => {
      const body = (await request.json()) as { email: string; password: string };

      if (body.email === "existing@example.com") {
        return HttpResponse.json(
          { error: { message: "Email already exists", code: "EMAIL_EXISTS" } },
          { status: 409 }
        );
      }

      return HttpResponse.json({
        user: { ...mockUser, id: "user-new", email: body.email },
        token: "mock-access-token",
        refresh_token: "mock-refresh-token",
      });
    }),

    http.post(`${apiUrl}/auth/login`, async ({ request }) => {
      const body = (await request.json()) as { email: string; password: string };

      if (body.password === "wrongpassword") {
        return HttpResponse.json(
          { error: { message: "Invalid credentials", code: "INVALID_CREDENTIALS" } },
          { status: 401 }
        );
      }

      return HttpResponse.json({
        user: { ...mockUser, email: body.email },
        token: "mock-access-token",
        refresh_token: "mock-refresh-token",
      });
    }),

    http.get(`${apiUrl}/auth/me`, () => {
      return HttpResponse.json({ user: mockUser });
    }),

    http.post(`${apiUrl}/auth/logout`, () => {
      return HttpResponse.json({ message: "Logged out successfully" });
    }),

    // Session endpoints - uses limit/offset
    http.get(`${apiUrl}/sessions`, () => {
      return HttpResponse.json({
        sessions: [mockSession],
      });
    }),

    http.get(`${apiUrl}/sessions/:id`, ({ params }) => {
      const { id } = params;

      if (id === "not-found") {
        return HttpResponse.json(
          { error: { message: "Session not found", code: "NOT_FOUND" } },
          { status: 404 }
        );
      }

      return HttpResponse.json({
        session: { ...mockSessionWithDetails, id },
      });
    }),

    http.delete(`${apiUrl}/sessions/:id`, ({ params }) => {
      const { id } = params;

      if (id === "not-found") {
        return HttpResponse.json(
          { error: { message: "Session not found", code: "NOT_FOUND" } },
          { status: 404 }
        );
      }

      return HttpResponse.json({ message: "Session deleted successfully" });
    }),
  ];
}

// Create handlers for all possible API URLs
export const handlers = API_URLS.flatMap(createHandlers);
