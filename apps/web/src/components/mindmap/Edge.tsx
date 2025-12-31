'use client';

import { useMemo } from 'react';
import { Line } from '@react-three/drei';
import * as THREE from 'three';
import type { MindmapEdge, MindmapNode } from '@/types/mindmap';

interface EdgeProps {
  edge: MindmapEdge;
  sourceNode: MindmapNode;
  targetNode: MindmapNode;
  isHighlighted: boolean;
}

export function Edge({ edge, sourceNode, targetNode, isHighlighted }: EdgeProps) {
  // Create curved line points using quadratic bezier
  const points = useMemo(() => {
    const source = sourceNode.position || { x: 0, y: 0, z: 0 };
    const target = targetNode.position || { x: 0, y: 0, z: 0 };

    // Calculate control point for bezier curve
    const midX = (source.x + target.x) / 2;
    const midY = (source.y + target.y) / 2;
    const midZ = (source.z + target.z) / 2;

    // Add some curvature based on distance
    const distance = Math.sqrt(
      Math.pow(target.x - source.x, 2) +
      Math.pow(target.y - source.y, 2) +
      Math.pow(target.z - source.z, 2)
    );
    const curvature = Math.min(distance * 0.1, 30);

    const curve = new THREE.QuadraticBezierCurve3(
      new THREE.Vector3(source.x, source.y, source.z),
      new THREE.Vector3(midX, midY + curvature, midZ),
      new THREE.Vector3(target.x, target.y, target.z)
    );

    return curve.getPoints(20);
  }, [sourceNode, targetNode]);

  // Line width based on edge weight
  const lineWidth = isHighlighted ? 2 : Math.max(0.5, Math.min(edge.weight * 0.5, 1.5));

  return (
    <Line
      points={points}
      color={isHighlighted ? '#ffffff' : '#4B5563'}
      lineWidth={lineWidth}
      opacity={isHighlighted ? 0.8 : 0.3}
      transparent
      dashed={isHighlighted}
      dashScale={3}
      dashSize={3}
      dashOffset={0}
    />
  );
}
