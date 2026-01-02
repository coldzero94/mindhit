'use client';

import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

interface ParticleFieldProps {
  count?: number;
  radius?: number;
  size?: number;
  color?: string;
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

export function ParticleField({
  count = 2000,
  radius = 400,
  size = 1.5,
  color = '#ffffff',
  seed = 12345,
}: ParticleFieldProps) {
  const pointsRef = useRef<THREE.Points>(null);

  // Generate positions using seeded random (deterministic, pure function)
  const positions = useMemo(() => {
    const pos = new Float32Array(count * 3);
    const random = createSeededRandom(seed);

    for (let i = 0; i < count; i++) {
      // Spherical coordinates for uniform distribution
      const theta = random() * Math.PI * 2;
      const phi = Math.acos(2 * random() - 1);
      const r = radius * Math.cbrt(random()); // Uniform volume distribution

      pos[i * 3] = r * Math.sin(phi) * Math.cos(theta);
      pos[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
      pos[i * 3 + 2] = r * Math.cos(phi);
    }

    return pos;
  }, [count, radius, seed]);

  // Slow rotation animation
  useFrame((_, delta) => {
    if (pointsRef.current) {
      pointsRef.current.rotation.y += delta * 0.02;
      pointsRef.current.rotation.x += delta * 0.01;
    }
  });

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          args={[positions, 3]}
        />
      </bufferGeometry>
      <pointsMaterial
        size={size}
        color={color}
        transparent
        opacity={0.6}
        sizeAttenuation
        depthWrite={false}
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}
