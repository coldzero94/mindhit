'use client';

import { ReactNode, useRef, useState } from 'react';
import { useFrame } from '@react-three/fiber';
import { useSpring, animated } from '@react-spring/three';
import * as THREE from 'three';

interface BigBangAnimationProps {
  children: ReactNode;
  isReady: boolean;
  onComplete?: () => void;
  duration?: number;
}

export function BigBangAnimation({
  children,
  isReady,
  onComplete,
  duration = 1500,
}: BigBangAnimationProps) {
  const groupRef = useRef<THREE.Group>(null);
  const [hasCompleted, setHasCompleted] = useState(false);

  // Spring animation configuration
  const { scale } = useSpring({
    scale: isReady ? 1 : 0,
    config: {
      mass: 1,
      tension: 120,
      friction: 14,
      duration: duration,
    },
    onRest: () => {
      if (isReady && !hasCompleted) {
        setHasCompleted(true);
        onComplete?.();
      }
    },
  });

  // Initial rotation animation
  useFrame((_, delta) => {
    if (groupRef.current && !hasCompleted) {
      groupRef.current.rotation.y += delta * 0.5;
    }
  });

  return (
    <animated.group ref={groupRef} scale={scale}>
      {children}
    </animated.group>
  );
}
