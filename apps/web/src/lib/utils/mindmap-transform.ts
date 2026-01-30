import type { MindmapMindmap } from "@/api/generated/types.gen";
import type { MindmapData, MindmapNode, MindmapNodeType, MindmapNodeData } from "@/types/mindmap";

export function transformApiMindmap(apiMindmap: MindmapMindmap): MindmapData | null {
  if (!apiMindmap.data) {
    return null;
  }

  const { nodes, edges, layout } = apiMindmap.data;

  return {
    nodes: nodes.map(
      (node): MindmapNode => {
        const rawData = node.data ?? {};
        const data: MindmapNodeData = {
          description: typeof rawData.description === 'string' ? rawData.description : undefined,
          urls: Array.isArray(rawData.urls) ? rawData.urls as string[] : undefined,
          keywords: Array.isArray(rawData.keywords) ? rawData.keywords as string[] : undefined,
          visitCount: typeof rawData.visitCount === 'number' ? rawData.visitCount : undefined,
          totalDuration: typeof rawData.totalDuration === 'number' ? rawData.totalDuration : undefined,
          url_id: typeof rawData.url_id === 'string' ? rawData.url_id : undefined,
          url: typeof rawData.url === 'string' ? rawData.url : undefined,
          summary: typeof rawData.summary === 'string' ? rawData.summary : undefined,
          relevance: typeof rawData.relevance === 'number' ? rawData.relevance : undefined,
        };

        return {
          id: node.id,
          label: node.label,
          type: node.type as MindmapNodeType,
          size: node.size,
          color: node.color,
          position: node.position
            ? {
                x: node.position.x,
                y: node.position.y,
                z: node.position.z,
              }
            : undefined,
          data,
        };
      }
    ),
    edges: edges.map((edge) => ({
      source: edge.source,
      target: edge.target,
      weight: edge.weight,
    })),
    layout: {
      type: (layout.type as MindmapData["layout"]["type"]) || "galaxy",
      params: layout.params ?? {},
    },
  };
}
