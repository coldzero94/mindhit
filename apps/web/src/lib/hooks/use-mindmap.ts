"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { mindmapApi } from "@/lib/api/mindmap";
import type {
  MindmapMindmap,
  MindmapGenerateMindmapRequest,
} from "@/api/generated/types.gen";

export const mindmapKeys = {
  all: ["mindmaps"] as const,
  detail: (sessionId: string) => [...mindmapKeys.all, sessionId] as const,
};

export function useMindmap(sessionId: string) {
  return useQuery({
    queryKey: mindmapKeys.detail(sessionId),
    queryFn: () => mindmapApi.get(sessionId),
    enabled: !!sessionId,
    retry: (failureCount, error) => {
      // Don't retry on 404 (mindmap not found)
      if ((error as { response?: { status?: number } })?.response?.status === 404) {
        return false;
      }
      return failureCount < 3;
    },
  });
}

export function useGenerateMindmap() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      sessionId,
      options,
    }: {
      sessionId: string;
      options?: MindmapGenerateMindmapRequest;
    }) => mindmapApi.generate(sessionId, options),
    onSuccess: (data: MindmapMindmap) => {
      queryClient.setQueryData(mindmapKeys.detail(data.session_id), data);
    },
  });
}
