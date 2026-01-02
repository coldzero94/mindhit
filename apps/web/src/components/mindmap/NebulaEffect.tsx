'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

interface NebulaEffectProps {
  count?: number;
  radius?: number;
  colors?: string[];
  seed?: number;
}

// Simple seeded random number generator
function createSeededRandom(seed: number): () => number {
  let s = seed;
  return () => {
    s = (s * 1103515245 + 12345) % 2147483648;
    return s / 2147483648;
  };
}

export function NebulaEffect({
  count = 500,
  radius = 300,
  colors = ['#4F46E5', '#7C3AED', '#2563EB'],
  seed = 54321,
}: NebulaEffectProps) {
  const pointsRef = useRef<THREE.Points>(null);

  // Generate nebula data using seeded random (deterministic, pure function)
  const { positions, colorArray } = useMemo(() => {
    const pos = new Float32Array(count * 3);
    const col = new Float32Array(count * 3);
    const colorObjects = colors.map((c) => new THREE.Color(c));
    const random = createSeededRandom(seed);

    for (let i = 0; i < count; i++) {
      // Spiral distribution
      const angle = (i / count) * Math.PI * 8;
      const r = (i / count) * radius;
      const noise = (random() - 0.5) * 50;

      pos[i * 3] = Math.cos(angle) * r + noise;
      pos[i * 3 + 1] = (random() - 0.5) * 100;
      pos[i * 3 + 2] = Math.sin(angle) * r + noise;

      // Deterministic color selection
      const colorIndex = Math.floor(random() * colorObjects.length);
      const color = colorObjects[colorIndex];
      col[i * 3] = color.r;
      col[i * 3 + 1] = color.g;
      col[i * 3 + 2] = color.b;
    }

    return { positions: pos, colorArray: col };
  }, [count, radius, colors, seed]);

  useFrame((_, delta) => {
    if (pointsRef.current) {
      pointsRef.current.rotation.y += delta * 0.05;
    }
  });

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" args={[positions, 3]} />
        <bufferAttribute attach="attributes-color" args={[colorArray, 3]} />
      </bufferGeometry>
      <pointsMaterial
        size={8}
        vertexColors
        transparent
        opacity={0.4}
        sizeAttenuation
        depthWrite={false}
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}
