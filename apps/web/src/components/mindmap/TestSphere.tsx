'use client';

import { useRef, useState } from 'react';
import { useFrame } from '@react-three/fiber';
import { Sphere, Text } from '@react-three/drei';
import * as THREE from 'three';

export function TestSphere() {
  const meshRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);

  useFrame((_, delta) => {
    if (meshRef.current) {
      meshRef.current.rotation.y += delta * 0.5;
    }
  });

  return (
    <group>
      <Sphere
        ref={meshRef}
        args={[30, 32, 32]}
        onPointerOver={() => setHovered(true)}
        onPointerOut={() => setHovered(false)}
      >
        <meshStandardMaterial
          color={hovered ? '#60A5FA' : '#3B82F6'}
          emissive={hovered ? '#60A5FA' : '#3B82F6'}
          emissiveIntensity={hovered ? 0.3 : 0.1}
          roughness={0.3}
          metalness={0.7}
        />
      </Sphere>

      <Text
        position={[0, 50, 0]}
        fontSize={12}
        color="white"
        anchorX="center"
        anchorY="middle"
      >
        MindHit 3D
      </Text>

      {/* Glow effect */}
      <Sphere args={[35, 32, 32]}>
        <meshBasicMaterial
          color="#3B82F6"
          transparent
          opacity={0.1}
          side={THREE.BackSide}
        />
      </Sphere>
    </group>
  );
}
