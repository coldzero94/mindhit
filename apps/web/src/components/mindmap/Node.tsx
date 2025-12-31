'use client';

import { useRef, useState } from 'react';
import { useFrame } from '@react-three/fiber';
import { Text, Sphere } from '@react-three/drei';
import { useSpring, animated } from '@react-spring/three';
import * as THREE from 'three';
import type { MindmapNode } from '@/types/mindmap';

interface NodeProps {
  node: MindmapNode;
  isSelected: boolean;
  isHovered: boolean;
  onClick: () => void;
  onHover: (hovered: boolean) => void;
}

const AnimatedSphere = animated(Sphere);

export function Node({ node, isSelected, isHovered, onClick, onHover }: NodeProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const [localHover, setLocalHover] = useState(false);

  // Animation spring
  const { scale, emissiveIntensity } = useSpring({
    scale: isHovered || localHover ? 1.2 : 1,
    emissiveIntensity: isSelected ? 0.5 : isHovered || localHover ? 0.3 : 0.1,
    config: { tension: 300, friction: 20 },
  });

  // Slow rotation for core node
  useFrame((_, delta) => {
    if (meshRef.current && node.type === 'core') {
      meshRef.current.rotation.y += delta * 0.1;
    }
  });

  const position: [number, number, number] = node.position
    ? [node.position.x, node.position.y, node.position.z]
    : [0, 0, 0];

  // 노드 타입별 크기 계산
  const getNodeSize = (): number => {
    const baseSize = node.size || 10;
    switch (node.type) {
      case 'core':
        return baseSize * 0.5;
      case 'topic':
        return baseSize * 0.4;
      case 'subtopic':
        return baseSize * 0.3;
      case 'page':
      default:
        return baseSize * 0.2;
    }
  };

  const nodeSize = getNodeSize();

  return (
    <group position={position}>
      {/* Main sphere */}
      <AnimatedSphere
        ref={meshRef}
        args={[nodeSize, 32, 32]}
        scale={scale}
        onClick={(e) => {
          e.stopPropagation();
          onClick();
        }}
        onPointerOver={(e) => {
          e.stopPropagation();
          setLocalHover(true);
          onHover(true);
          document.body.style.cursor = 'pointer';
        }}
        onPointerOut={() => {
          setLocalHover(false);
          onHover(false);
          document.body.style.cursor = 'auto';
        }}
      >
        <animated.meshStandardMaterial
          color={node.color}
          emissive={node.color}
          emissiveIntensity={emissiveIntensity}
          roughness={0.3}
          metalness={0.7}
        />
      </AnimatedSphere>

      {/* Glow effect for core node */}
      {node.type === 'core' && (
        <Sphere args={[nodeSize * 1.3, 32, 32]}>
          <meshBasicMaterial
            color={node.color}
            transparent
            opacity={0.15}
            side={THREE.BackSide}
          />
        </Sphere>
      )}

      {/* Ring effect for selected node */}
      {isSelected && (
        <mesh rotation={[Math.PI / 2, 0, 0]}>
          <ringGeometry args={[nodeSize * 1.5, nodeSize * 1.7, 32]} />
          <meshBasicMaterial color="#ffffff" transparent opacity={0.5} side={THREE.DoubleSide} />
        </mesh>
      )}

      {/* Label - show on hover, select, or for core nodes */}
      {(isHovered || localHover || isSelected || node.type === 'core') && (
        <Text
          position={[0, nodeSize + 15, 0]}
          fontSize={node.type === 'core' ? 14 : 10}
          color="white"
          anchorX="center"
          anchorY="middle"
          outlineWidth={0.5}
          outlineColor="black"
          maxWidth={150}
        >
          {node.label}
        </Text>
      )}
    </group>
  );
}
