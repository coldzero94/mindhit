export interface MindmapNodePosition {
  x: number;
  y: number;
  z: number;
}

export type MindmapNodeType = 'core' | 'topic' | 'subtopic' | 'page';

export interface MindmapNode {
  id: string;
  label: string;
  type: MindmapNodeType;
  size: number;
  color: string;
  position?: MindmapNodePosition;
  data: {
    description?: string;
    urls?: string[];
    visitCount?: number;
    totalDuration?: number;
    [key: string]: unknown;
  };
}

export interface MindmapEdge {
  source: string;
  target: string;
  weight: number;
}

export type MindmapLayoutType = 'galaxy' | 'tree' | 'radial';

export interface MindmapLayout {
  type: MindmapLayoutType;
  params: Record<string, unknown>;
}

export interface MindmapData {
  nodes: MindmapNode[];
  edges: MindmapEdge[];
  layout: MindmapLayout;
}

// API Response 타입
export interface MindmapResponse {
  mindmap: MindmapData | null;
  session_id: string;
  generated_at: string | null;
}
