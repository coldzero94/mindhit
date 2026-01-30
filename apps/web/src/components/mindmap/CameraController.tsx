'use client';

import { useRef, useEffect, useCallback } from 'react';
import { useThree, useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import type { MindmapNode } from '@/types/mindmap';

interface CameraControllerProps {
  selectedNode: MindmapNode | null;
  focusDistance?: number;
  animationDuration?: number;
  onAnimationComplete?: () => void;
}

export function CameraController({
  selectedNode,
  focusDistance = 120,
  animationDuration = 800,
  onAnimationComplete,
}: CameraControllerProps) {
  const { camera, controls } = useThree();
  const isAnimating = useRef(false);
  const animationProgress = useRef(0);
  const startPosition = useRef(new THREE.Vector3());
  const targetPosition = useRef(new THREE.Vector3());
  const startTarget = useRef(new THREE.Vector3());
  const targetTarget = useRef(new THREE.Vector3());

  // Easing function for smooth animation
  const easeOutCubic = (t: number): number => {
    return 1 - Math.pow(1 - t, 3);
  };

  const animateToNode = useCallback((node: MindmapNode) => {
    if (!node.position) return;

    const nodePos = new THREE.Vector3(
      node.position.x,
      node.position.y,
      node.position.z
    );

    // Calculate camera position - move towards node but keep some distance
    const direction = new THREE.Vector3()
      .subVectors(camera.position, nodePos)
      .normalize();

    // If direction is near zero (camera at node), use default direction
    if (direction.lengthSq() < 0.001) {
      direction.set(0, 0, 1);
    }

    startPosition.current.copy(camera.position);
    targetPosition.current.copy(nodePos).add(direction.multiplyScalar(focusDistance));

    // Get current look-at target from controls
    if (controls && 'target' in controls) {
      startTarget.current.copy((controls as { target: THREE.Vector3 }).target);
    } else {
      startTarget.current.set(0, 0, 0);
    }
    targetTarget.current.copy(nodePos);

    isAnimating.current = true;
    animationProgress.current = 0;
  }, [camera, controls, focusDistance]);

  // Watch for selectedNode changes
  useEffect(() => {
    if (selectedNode) {
      animateToNode(selectedNode);
    }
    // Don't auto-return to default when deselected - let user control
  }, [selectedNode, animateToNode]);

  useFrame((_, delta) => {
    if (!isAnimating.current) return;

    // Progress animation (delta is in seconds, animationDuration is in ms)
    animationProgress.current += (delta * 1000) / animationDuration;

    if (animationProgress.current >= 1) {
      // Animation complete
      animationProgress.current = 1;
      isAnimating.current = false;
      onAnimationComplete?.();
    }

    const t = easeOutCubic(animationProgress.current);

    // Interpolate camera position
    camera.position.lerpVectors(startPosition.current, targetPosition.current, t);

    // Interpolate controls target
    if (controls && 'target' in controls) {
      const target = (controls as { target: THREE.Vector3 }).target;
      target.lerpVectors(startTarget.current, targetTarget.current, t);
    }
  });

  return null;
}
