import type { MindmapNodeType } from '@/types/mindmap';

// 노드 타입별 기본 색상
export const NODE_TYPE_COLORS: Record<MindmapNodeType, string> = {
  core: '#F59E0B',     // Amber - 중심 노드
  topic: '#3B82F6',    // Blue - 주제
  subtopic: '#10B981', // Emerald - 하위 주제
  page: '#8B5CF6',     // Violet - 페이지
};

// 주제별 색상 팔레트
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

// 노드 크기 계산
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

  // 방문 횟수에 따른 크기 조정
  if (visitCount && visitCount > 1) {
    size *= Math.min(1 + visitCount * 0.1, 1.5);
  }

  // 체류 시간에 따른 크기 조정 (밀리초 → 분)
  if (totalDuration && totalDuration > 60000) {
    const minutes = totalDuration / 60000;
    size *= Math.min(1 + minutes * 0.05, 1.3);
  }

  return size;
}
