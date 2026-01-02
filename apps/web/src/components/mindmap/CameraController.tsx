'use client';

import { useRef, useEffect } from 'react';
import { useThree, useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import type { MindmapNode } from '@/types/mindmap';

interface CameraControllerProps {
  selectedNode: MindmapNode | null;
  defaultPosition?: [number, number, number];
  focusDistance?: number;
  lerpFactor?: number;
}

export function CameraController({
  selectedNode,
  defaultPosition = [0, 0, 500],
  focusDistance = 150,
  lerpFactor = 0.05,
}: CameraControllerProps) {
  const { camera } = useThree();
  const targetPosition = useRef(new THREE.Vector3(...defaultPosition));
  const targetLookAt = useRef(new THREE.Vector3(0, 0, 0));

  useEffect(() => {
    if (selectedNode && selectedNode.position) {
      // Move camera towards selected node
      const nodePos = new THREE.Vector3(
        selectedNode.position.x,
        selectedNode.position.y,
        selectedNode.position.z
      );

      // Position camera at focusDistance from node
      const direction = new THREE.Vector3()
        .subVectors(camera.position, nodePos)
        .normalize();

      targetPosition.current.copy(nodePos).add(direction.multiplyScalar(focusDistance));
      targetLookAt.current.copy(nodePos);
    } else {
      // Return to default position
      targetPosition.current.set(...defaultPosition);
      targetLookAt.current.set(0, 0, 0);
    }
  }, [selectedNode, camera.position, defaultPosition, focusDistance]);

  useFrame(() => {
    // Smooth camera movement (lerp)
    camera.position.lerp(targetPosition.current, lerpFactor);

    // Update camera look-at point
    const currentLookAt = new THREE.Vector3();
    camera.getWorldDirection(currentLookAt);

    const targetDir = targetLookAt.current.clone().sub(camera.position).normalize();
    currentLookAt.lerp(targetDir, lerpFactor);

    camera.lookAt(
      camera.position.x + currentLookAt.x,
      camera.position.y + currentLookAt.y,
      camera.position.z + currentLookAt.z
    );
  });

  return null;
}
