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

// Phase 11 Mock data

export const mockPlan = {
  id: "plan-free",
  name: "Free",
  description: "Free plan for individual users",
  token_limit: 10000,
  features: ["basic_mindmap", "session_history"],
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

export const mockProPlan = {
  id: "plan-pro",
  name: "Pro",
  description: "Pro plan for power users",
  token_limit: 100000,
  features: ["basic_mindmap", "session_history", "advanced_analytics", "priority_support"],
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

export const mockSubscription = {
  id: "sub-1",
  user_id: "user-1",
  plan_id: "plan-free",
  plan: mockPlan,
  status: "active" as const,
  current_period_start: "2025-01-01T00:00:00Z",
  current_period_end: "2025-02-01T00:00:00Z",
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

// Usage summary matches UsageUsageSummary type
export const mockUsageSummary = {
  period_start: "2025-01-01T00:00:00Z",
  period_end: "2025-02-01T00:00:00Z",
  tokens_used: 3500,
  token_limit: 10000,
  percent_used: 35.0,
  is_unlimited: false,
  can_use_ai: true,
  by_operation: {
    tag_extraction: 2000,
    mindmap_generation: 1500,
  },
};

// High usage (80%+) for warning tests
export const mockUsageSummaryHigh = {
  ...mockUsageSummary,
  tokens_used: 8500,
  percent_used: 85.0,
};

// Over limit (100%+) for error tests
export const mockUsageSummaryOver = {
  ...mockUsageSummary,
  tokens_used: 10500,
  percent_used: 105.0,
  can_use_ai: false,
};

// Unlimited usage for pro plans
export const mockUsageSummaryUnlimited = {
  ...mockUsageSummary,
  tokens_used: 50000,
  token_limit: 0,
  percent_used: 0,
  is_unlimited: true,
  can_use_ai: true,
};

export const mockUsageHistory = [
  { month: "2025-01", used_tokens: 3500, token_limit: 10000 },
  { month: "2024-12", used_tokens: 8000, token_limit: 10000 },
  { month: "2024-11", used_tokens: 5500, token_limit: 10000 },
  { month: "2024-10", used_tokens: 2000, token_limit: 10000 },
  { month: "2024-09", used_tokens: 7500, token_limit: 10000 },
  { month: "2024-08", used_tokens: 4000, token_limit: 10000 },
];

export const mockMindmapNode = {
  id: "node-core",
  label: "Test Session",
  type: "core" as const,
  size: 20,
  color: "#4F46E5",
  position: { x: 0, y: 0, z: 0 },
  data: { description: "Core node" },
};

export const mockMindmapNodes = [
  mockMindmapNode,
  {
    id: "node-topic-1",
    label: "Technology",
    type: "topic" as const,
    size: 15,
    color: "#10B981",
    position: { x: 100, y: 50, z: 0 },
    data: { description: "Tech related pages", urls: ["https://example.com/tech"] },
  },
  {
    id: "node-topic-2",
    label: "News",
    type: "topic" as const,
    size: 15,
    color: "#F59E0B",
    position: { x: -100, y: 50, z: 0 },
    data: { description: "News articles" },
  },
];

export const mockMindmapEdges = [
  { source: "node-core", target: "node-topic-1", weight: 1.0 },
  { source: "node-core", target: "node-topic-2", weight: 0.8 },
];

export const mockMindmap = {
  id: "mindmap-1",
  session_id: "session-1",
  nodes: mockMindmapNodes,
  edges: mockMindmapEdges,
  layout: { type: "galaxy" as const, params: {} },
  generated_at: "2025-01-01T12:00:00Z",
  created_at: "2025-01-01T12:00:00Z",
  updated_at: "2025-01-01T12:00:00Z",
};

// Helper to create Phase 11 handlers
function createPhase11Handlers(apiUrl: string) {
  return [
    // Subscription endpoints
    http.get(`${apiUrl}/subscription`, () => {
      return HttpResponse.json({
        subscription: mockSubscription,
      });
    }),

    http.get(`${apiUrl}/subscription/plans`, () => {
      return HttpResponse.json({
        plans: [mockPlan, mockProPlan],
      });
    }),

    // Usage endpoints
    http.get(`${apiUrl}/usage`, () => {
      return HttpResponse.json({ usage: mockUsageSummary });
    }),

    http.get(`${apiUrl}/usage/history`, ({ request }) => {
      const url = new URL(request.url);
      const months = parseInt(url.searchParams.get("months") || "6", 10);
      return HttpResponse.json({
        history: mockUsageHistory.slice(0, months),
      });
    }),

    // Mindmap endpoints
    http.get(`${apiUrl}/sessions/:sessionId/mindmap`, ({ params }) => {
      const { sessionId } = params;

      if (sessionId === "not-found") {
        return HttpResponse.json(
          { error: { message: "Mindmap not found", code: "NOT_FOUND" } },
          { status: 404 }
        );
      }

      if (sessionId === "no-mindmap") {
        return HttpResponse.json(
          { error: { message: "Mindmap not generated yet", code: "NOT_FOUND" } },
          { status: 404 }
        );
      }

      return HttpResponse.json({
        mindmap: { ...mockMindmap, session_id: sessionId as string },
      });
    }),

    http.post(`${apiUrl}/sessions/:sessionId/mindmap/generate`, async ({ params, request }) => {
      const { sessionId } = params;
      const body = await request.json() as { force?: boolean } | undefined;

      if (sessionId === "not-found") {
        return HttpResponse.json(
          { error: { message: "Session not found", code: "NOT_FOUND" } },
          { status: 404 }
        );
      }

      return HttpResponse.json({
        mindmap: {
          ...mockMindmap,
          session_id: sessionId as string,
          generated_at: new Date().toISOString(),
          ...(body?.force && { regenerated: true }),
        },
      });
    }),
  ];
}

// Create handlers for all possible API URLs
export const handlers = [
  ...API_URLS.flatMap(createHandlers),
  ...API_URLS.flatMap(createPhase11Handlers),
];
