"use client";

import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { sessionsApi } from "@/lib/api/sessions";
import type {
  SessionSession,
  SessionUpdateSessionRequest,
} from "@/api/generated/types.gen";

export const sessionKeys = {
  all: ["sessions"] as const,
  lists: () => [...sessionKeys.all, "list"] as const,
  list: (limit: number, offset: number) =>
    [...sessionKeys.lists(), { limit, offset }] as const,
  details: () => [...sessionKeys.all, "detail"] as const,
  detail: (id: string) => [...sessionKeys.details(), id] as const,
  events: (id: string) => [...sessionKeys.detail(id), "events"] as const,
  stats: (id: string) => [...sessionKeys.detail(id), "stats"] as const,
};

export function useSessions(limit = 20, offset = 0) {
  return useQuery({
    queryKey: sessionKeys.list(limit, offset),
    queryFn: () => sessionsApi.list(limit, offset),
  });
}

export function useSession(id: string) {
  return useQuery({
    queryKey: sessionKeys.detail(id),
    queryFn: () => sessionsApi.get(id),
    enabled: !!id,
  });
}

export function useSessionEvents(
  sessionId: string,
  options?: { type?: string; limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: [...sessionKeys.events(sessionId), options],
    queryFn: () => sessionsApi.getEvents(sessionId, options),
    enabled: !!sessionId,
  });
}

export function useSessionStats(sessionId: string) {
  return useQuery({
    queryKey: sessionKeys.stats(sessionId),
    queryFn: () => sessionsApi.getEventStats(sessionId),
    enabled: !!sessionId,
  });
}

export function useStartSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => sessionsApi.start(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    },
  });
}

export function useUpdateSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: SessionUpdateSessionRequest;
    }) => sessionsApi.update(id, data),
    onSuccess: (data: SessionSession) => {
      queryClient.setQueryData(sessionKeys.detail(data.id), data);
      queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    },
  });
}

export function usePauseSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => sessionsApi.pause(id),
    onSuccess: (data: SessionSession) => {
      queryClient.setQueryData(sessionKeys.detail(data.id), data);
      queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    },
  });
}

export function useResumeSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => sessionsApi.resume(id),
    onSuccess: (data: SessionSession) => {
      queryClient.setQueryData(sessionKeys.detail(data.id), data);
      queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    },
  });
}

export function useStopSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => sessionsApi.stop(id),
    onSuccess: (data: SessionSession) => {
      queryClient.setQueryData(sessionKeys.detail(data.id), data);
      queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    },
  });
}

export function useDeleteSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => sessionsApi.delete(id),
    onSuccess: (_data, id) => {
      queryClient.removeQueries({ queryKey: sessionKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
    },
  });
}
