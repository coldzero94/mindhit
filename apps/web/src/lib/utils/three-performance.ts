import { useMemo } from 'react';
import type { MindmapNode } from '@/types/mindmap';

// LOD (Level of Detail) configuration
export interface LODConfig {
  highDetail: number;   // Within this distance: high quality
  mediumDetail: number; // Within this distance: medium quality
  lowDetail: number;    // Beyond this distance: low quality
}

export const DEFAULT_LOD: LODConfig = {
  highDetail: 100,
  mediumDetail: 300,
  lowDetail: 500,
};

// Determine optimal segments based on node count
export function getOptimalSegments(nodeCount: number): number {
  if (nodeCount > 100) return 8;
  if (nodeCount > 50) return 16;
  return 32;
}

// Optimize particle count based on node count
export function getOptimalParticleCount(nodeCount: number): number {
  if (nodeCount > 100) return 1000;
  if (nodeCount > 50) return 2000;
  return 3000;
}

// Filter nodes outside viewport
export function useVisibleNodes(
  nodes: MindmapNode[],
  cameraPosition: [number, number, number],
  maxDistance: number = 800
) {
  return useMemo(() => {
    const [cx, cy, cz] = cameraPosition;

    return nodes.filter((node) => {
      if (!node.position) return true;
      const dx = node.position.x - cx;
      const dy = node.position.y - cy;
      const dz = node.position.z - cz;
      const distance = Math.sqrt(dx * dx + dy * dy + dz * dz);
      return distance <= maxDistance;
    });
  }, [nodes, cameraPosition, maxDistance]);
}

// Debounced state update
export function debounce<T extends (...args: Parameters<T>) => void>(
  fn: T,
  delay: number
): T {
  let timeoutId: ReturnType<typeof setTimeout>;

  return ((...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), delay);
  }) as T;
}
