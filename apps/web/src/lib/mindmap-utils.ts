import type { MindmapNodeType } from '@/types/mindmap';

// Default colors for node types
export const NODE_TYPE_COLORS: Record<MindmapNodeType, string> = {
  core: '#F59E0B',     // Amber - Core node
  topic: '#3B82F6',    // Blue - Topic
  subtopic: '#10B981', // Emerald - Subtopic
  page: '#8B5CF6',     // Violet - Page
};

// Topic color palette
export const TOPIC_COLORS = [
  '#3B82F6', // Blue
  '#10B981', // Emerald
  '#F59E0B', // Amber
  '#EF4444', // Red
  '#8B5CF6', // Violet
  '#EC4899', // Pink
  '#06B6D4', // Cyan
  '#F97316', // Orange
];

export function getTopicColor(index: number): string {
  return TOPIC_COLORS[index % TOPIC_COLORS.length];
}

// Calculate node size
export function calculateNodeSize(
  type: MindmapNodeType,
  visitCount?: number,
  totalDuration?: number
): number {
  const baseSize: Record<MindmapNodeType, number> = {
    core: 50,
    topic: 30,
    subtopic: 20,
    page: 15,
  };

  let size = baseSize[type];

  // Adjust size based on visit count
  if (visitCount && visitCount > 1) {
    size *= Math.min(1 + visitCount * 0.1, 1.5);
  }

  // Adjust size based on duration (milliseconds → minutes)
  if (totalDuration && totalDuration > 60000) {
    const minutes = totalDuration / 60000;
    size *= Math.min(1 + minutes * 0.05, 1.3);
  }

  return size;
}
