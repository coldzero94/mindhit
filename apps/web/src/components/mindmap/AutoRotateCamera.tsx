'use client';

import { useRef, useEffect } from 'react';
import { useThree, useFrame } from '@react-three/fiber';

interface AutoRotateCameraProps {
  enabled: boolean;
  speed?: number;
  radius?: number;
}

export function AutoRotateCamera({
  enabled,
  speed = 0.1,
  radius = 500,
}: AutoRotateCameraProps) {
  const { camera } = useThree();
  const angleRef = useRef(0);

  useEffect(() => {
    if (enabled) {
      // Calculate current angle from camera position
      angleRef.current = Math.atan2(camera.position.x, camera.position.z);
    }
  }, [enabled, camera.position]);

  useFrame((_, delta) => {
    if (!enabled) return;

    angleRef.current += speed * delta;

    const x = Math.sin(angleRef.current) * radius;
    const z = Math.cos(angleRef.current) * radius;
    const y = camera.position.y;

    camera.position.set(x, y, z);
    camera.lookAt(0, 0, 0);
  });

  return null;
}
