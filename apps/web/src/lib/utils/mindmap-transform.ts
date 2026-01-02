import type { MindmapMindmap } from "@/api/generated/types.gen";
import type { MindmapData, MindmapNode, MindmapNodeType } from "@/types/mindmap";

export function transformApiMindmap(apiMindmap: MindmapMindmap): MindmapData | null {
  if (!apiMindmap.data) {
    return null;
  }

  const { nodes, edges, layout } = apiMindmap.data;

  return {
    nodes: nodes.map(
      (node): MindmapNode => ({
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
        data: node.data ?? {},
      })
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
