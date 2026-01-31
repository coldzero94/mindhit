export interface MindmapNodePosition {
  x: number;
  y: number;
  z: number;
}

export type MindmapNodeType = 'core' | 'topic' | 'subtopic' | 'page';

export interface MindmapNodeData {
  description?: string;
  urls?: string[];
  keywords?: string[];
  visitCount?: number;
  totalDuration?: number;
  url_id?: string;
  url?: string;
  summary?: string;
  relevance?: number;
}

export interface MindmapNode {
  id: string;
  label: string;
  type: MindmapNodeType;
  size: number;
  color: string;
  position?: MindmapNodePosition;
  data: MindmapNodeData;
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

// API Response type
export interface MindmapResponse {
  mindmap: MindmapData | null;
  session_id: string;
  generated_at: string | null;
}
