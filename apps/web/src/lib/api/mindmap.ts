import { apiClient } from "./client";
import type {
  MindmapMindmap,
  MindmapMindmapResponse,
  MindmapGenerateMindmapRequest,
} from "@/api/generated/types.gen";

export type { MindmapMindmap, MindmapMindmapResponse };

export const mindmapApi = {
  get: async (sessionId: string): Promise<MindmapMindmap> => {
    const response = await apiClient.get<MindmapMindmapResponse>(
      `/sessions/${sessionId}/mindmap`
    );
    return response.data.mindmap;
  },

  generate: async (
    sessionId: string,
    options?: MindmapGenerateMindmapRequest
  ): Promise<MindmapMindmap> => {
    const response = await apiClient.post<MindmapMindmapResponse>(
      `/sessions/${sessionId}/mindmap/generate`,
      options ?? {}
    );
    return response.data.mindmap;
  },
};
