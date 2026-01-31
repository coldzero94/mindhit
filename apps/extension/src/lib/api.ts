import type { Session, BrowsingEvent } from "@/types";
import { API_BASE_URL, API_KEY } from "@/lib/constants";

interface SessionResponse {
  session: Session;
}

interface SessionListResponse {
  sessions: Session[];
  total: number;
}

function getAuthHeaders(): HeadersInit {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };
  if (API_KEY) {
    headers["Authorization"] = `Bearer ${API_KEY}`;
  }
  return headers;
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      ...getAuthHeaders(),
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(
      (error as { error?: { message?: string } }).error?.message ||
        "Request failed"
    );
  }

  return response.json() as Promise<T>;
}

export const api = {
  // Sessions
  startSession: async (): Promise<SessionResponse> => {
    return request<SessionResponse>("/sessions/start", {
      method: "POST",
    });
  },

  pauseSession: async (sessionId: string): Promise<SessionResponse> => {
    return request<SessionResponse>(`/sessions/${sessionId}/pause`, {
      method: "PATCH",
    });
  },

  resumeSession: async (sessionId: string): Promise<SessionResponse> => {
    return request<SessionResponse>(`/sessions/${sessionId}/resume`, {
      method: "PATCH",
    });
  },

  stopSession: async (sessionId: string): Promise<SessionResponse> => {
    return request<SessionResponse>(`/sessions/${sessionId}/stop`, {
      method: "POST",
    });
  },

  // Events
  sendEvents: async (
    sessionId: string,
    events: BrowsingEvent[]
  ): Promise<void> => {
    await request(`/sessions/${sessionId}/events`, {
      method: "POST",
      body: JSON.stringify({ events }),
    });
  },

  // Session List
  getSessions: async (limit: number = 5): Promise<SessionListResponse> => {
    return request<SessionListResponse>(
      `/sessions?limit=${limit}&sort=started_at:desc`,
      {
        method: "GET",
      }
    );
  },

  // Update Session (for title change)
  updateSession: async (
    sessionId: string,
    data: { title?: string; description?: string }
  ): Promise<SessionResponse> => {
    return request<SessionResponse>(`/sessions/${sessionId}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  },
};
