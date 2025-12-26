import { apiClient } from "./client";
import type {
  SessionSession,
  SessionSessionListResponse,
  SessionSessionResponse,
  SessionUpdateSessionRequest,
  EventsEventListResponse,
  EventsEventStatsResponse,
} from "@/api/generated/types.gen";

export type { SessionSession, SessionSessionListResponse };

export const sessionsApi = {
  list: async (
    limit = 20,
    offset = 0
  ): Promise<SessionSessionListResponse> => {
    const response = await apiClient.get<SessionSessionListResponse>(
      "/sessions",
      {
        params: { limit, offset },
      }
    );
    return response.data;
  },

  get: async (id: string): Promise<SessionSession> => {
    const response = await apiClient.get<SessionSessionResponse>(
      `/sessions/${id}`
    );
    return response.data.session;
  },

  start: async (): Promise<SessionSession> => {
    const response =
      await apiClient.post<SessionSessionResponse>("/sessions/start");
    return response.data.session;
  },

  update: async (
    id: string,
    data: SessionUpdateSessionRequest
  ): Promise<SessionSession> => {
    const response = await apiClient.put<SessionSessionResponse>(
      `/sessions/${id}`,
      data
    );
    return response.data.session;
  },

  pause: async (id: string): Promise<SessionSession> => {
    const response = await apiClient.patch<SessionSessionResponse>(
      `/sessions/${id}/pause`
    );
    return response.data.session;
  },

  resume: async (id: string): Promise<SessionSession> => {
    const response = await apiClient.patch<SessionSessionResponse>(
      `/sessions/${id}/resume`
    );
    return response.data.session;
  },

  stop: async (id: string): Promise<SessionSession> => {
    const response = await apiClient.post<SessionSessionResponse>(
      `/sessions/${id}/stop`
    );
    return response.data.session;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/sessions/${id}`);
  },

  // Events
  getEvents: async (
    sessionId: string,
    options?: { type?: string; limit?: number; offset?: number }
  ): Promise<EventsEventListResponse> => {
    const response = await apiClient.get<EventsEventListResponse>(
      `/sessions/${sessionId}/events`,
      { params: options }
    );
    return response.data;
  },

  getEventStats: async (sessionId: string): Promise<EventsEventStatsResponse> => {
    const response = await apiClient.get<EventsEventStatsResponse>(
      `/sessions/${sessionId}/events/stats`
    );
    return response.data;
  },
};
