import { http, HttpResponse } from "msw";

const API_URL = "http://localhost:9000/v1";

// Helper to validate API key
function hasValidApiKey(request: Request): boolean {
  const authHeader = request.headers.get("Authorization");
  return authHeader?.startsWith("Bearer ") ?? false;
}

// Mock data
export const mockSession = {
  id: "session-1",
  title: "Test Session",
  session_status: "recording" as const,
  started_at: "2025-01-01T10:00:00Z",
  created_at: "2025-01-01T10:00:00Z",
  updated_at: "2025-01-01T10:00:00Z",
};

export const mockSessions = [
  {
    id: "session-1",
    title: "First Session",
    session_status: "completed" as const,
    started_at: "2025-01-01T10:00:00Z",
    ended_at: "2025-01-01T11:00:00Z",
  },
  {
    id: "session-2",
    title: "Second Session",
    session_status: "completed" as const,
    started_at: "2025-01-02T10:00:00Z",
    ended_at: "2025-01-02T11:30:00Z",
  },
  {
    id: "session-3",
    title: null,
    session_status: "paused" as const,
    started_at: "2025-01-03T10:00:00Z",
    ended_at: null,
  },
];

export const handlers = [
  // Sessions - List
  http.get(`${API_URL}/sessions`, ({ request }) => {
    if (!hasValidApiKey(request)) {
      return HttpResponse.json(
        { error: { message: "Unauthorized", code: "UNAUTHORIZED" } },
        { status: 401 }
      );
    }

    const url = new URL(request.url);
    const limit = parseInt(url.searchParams.get("limit") || "5", 10);

    return HttpResponse.json({
      sessions: mockSessions.slice(0, limit),
      total: mockSessions.length,
    });
  }),

  // Sessions - Update (title)
  http.patch(`${API_URL}/sessions/:id`, async ({ params, request }) => {
    if (!hasValidApiKey(request)) {
      return HttpResponse.json(
        { error: { message: "Unauthorized", code: "UNAUTHORIZED" } },
        { status: 401 }
      );
    }

    const { id } = params;
    if (id === "not-found") {
      return HttpResponse.json(
        { error: { message: "Session not found", code: "NOT_FOUND" } },
        { status: 404 }
      );
    }

    const body = (await request.json()) as { title?: string };

    return HttpResponse.json({
      session: {
        ...mockSession,
        id,
        title: body.title ?? mockSession.title,
        updated_at: new Date().toISOString(),
      },
    });
  }),

  // Sessions - Start
  http.post(`${API_URL}/sessions/start`, ({ request }) => {
    if (!hasValidApiKey(request)) {
      return HttpResponse.json(
        { error: { message: "Unauthorized", code: "UNAUTHORIZED" } },
        { status: 401 }
      );
    }

    return HttpResponse.json({
      session: {
        ...mockSession,
        id: "session-new",
        started_at: new Date().toISOString(),
      },
    });
  }),

  // Sessions - Pause
  http.patch(`${API_URL}/sessions/:id/pause`, ({ params, request }) => {
    if (!hasValidApiKey(request)) {
      return HttpResponse.json(
        { error: { message: "Unauthorized", code: "UNAUTHORIZED" } },
        { status: 401 }
      );
    }

    const { id } = params;
    if (id === "not-found") {
      return HttpResponse.json(
        { error: { message: "Session not found", code: "NOT_FOUND" } },
        { status: 404 }
      );
    }

    return HttpResponse.json({
      session: { ...mockSession, id, session_status: "paused" },
    });
  }),

  // Sessions - Resume
  http.patch(`${API_URL}/sessions/:id/resume`, ({ params, request }) => {
    if (!hasValidApiKey(request)) {
      return HttpResponse.json(
        { error: { message: "Unauthorized", code: "UNAUTHORIZED" } },
        { status: 401 }
      );
    }

    const { id } = params;
    if (id === "not-found") {
      return HttpResponse.json(
        { error: { message: "Session not found", code: "NOT_FOUND" } },
        { status: 404 }
      );
    }

    return HttpResponse.json({
      session: { ...mockSession, id, session_status: "recording" },
    });
  }),

  // Sessions - Stop
  http.post(`${API_URL}/sessions/:id/stop`, ({ params, request }) => {
    if (!hasValidApiKey(request)) {
      return HttpResponse.json(
        { error: { message: "Unauthorized", code: "UNAUTHORIZED" } },
        { status: 401 }
      );
    }

    const { id } = params;
    if (id === "not-found") {
      return HttpResponse.json(
        { error: { message: "Session not found", code: "NOT_FOUND" } },
        { status: 404 }
      );
    }

    return HttpResponse.json({
      session: {
        ...mockSession,
        id,
        session_status: "completed",
        ended_at: new Date().toISOString(),
      },
    });
  }),

  // Events - Send
  http.post(`${API_URL}/sessions/:id/events`, async ({ params, request }) => {
    if (!hasValidApiKey(request)) {
      return HttpResponse.json(
        { error: { message: "Unauthorized", code: "UNAUTHORIZED" } },
        { status: 401 }
      );
    }

    const { id } = params;
    if (id === "not-found") {
      return HttpResponse.json(
        { error: { message: "Session not found", code: "NOT_FOUND" } },
        { status: 404 }
      );
    }

    const body = (await request.json()) as { events: unknown[] };

    if (!body.events) {
      return HttpResponse.json(
        { error: { message: "Invalid request body", code: "BAD_REQUEST" } },
        { status: 400 }
      );
    }

    return HttpResponse.json({
      message: "Events processed successfully",
      count: body.events.length,
    });
  }),
];
