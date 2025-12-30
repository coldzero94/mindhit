import type { User, Session, BrowsingEvent } from "@/types";
import { API_BASE_URL } from "@/lib/constants";

interface AuthResponse {
  user: User;
  token: string;
}

interface SessionResponse {
  session: Session;
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
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
  // Auth
  login: async (email: string, password: string): Promise<AuthResponse> => {
    return request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },

  googleAuth: async (credential: string): Promise<AuthResponse> => {
    return request<AuthResponse>("/auth/google", {
      method: "POST",
      body: JSON.stringify({ credential }),
    });
  },

  googleAuthCode: async (
    code: string,
    redirectUri: string
  ): Promise<AuthResponse> => {
    return request<AuthResponse>("/auth/google/code", {
      method: "POST",
      body: JSON.stringify({ code, redirect_uri: redirectUri }),
    });
  },

  // Sessions
  startSession: async (token: string): Promise<SessionResponse> => {
    return request<SessionResponse>("/sessions/start", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
  },

  pauseSession: async (
    token: string,
    sessionId: string
  ): Promise<SessionResponse> => {
    return request<SessionResponse>(`/sessions/${sessionId}/pause`, {
      method: "PATCH",
      headers: { Authorization: `Bearer ${token}` },
    });
  },

  resumeSession: async (
    token: string,
    sessionId: string
  ): Promise<SessionResponse> => {
    return request<SessionResponse>(`/sessions/${sessionId}/resume`, {
      method: "PATCH",
      headers: { Authorization: `Bearer ${token}` },
    });
  },

  stopSession: async (
    token: string,
    sessionId: string
  ): Promise<SessionResponse> => {
    return request<SessionResponse>(`/sessions/${sessionId}/stop`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
  },

  // Events
  sendEvents: async (
    token: string,
    sessionId: string,
    events: BrowsingEvent[]
  ): Promise<void> => {
    await request("/events/batch", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({ session_id: sessionId, events }),
    });
  },
};
