# Phase 11.5: 애니메이션 및 인터랙션

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 3D 마인드맵 애니메이션 효과 및 카메라 제어 개선 |
| **선행 조건** | Phase 11.2 완료 |
| **예상 소요** | 2 Steps |
| **결과물** | 빅뱅 애니메이션, 파티클 효과, 부드러운 카메라 이동 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 11.5.1 | 시각 효과 | ⬜ |
| 11.5.2 | 카메라 및 인터랙션 | ⬜ |

---

## Step 11.5.1: 시각 효과

### 체크리스트

- [ ] **빅뱅 초기 애니메이션**
  - [ ] `src/components/mindmap/BigBangAnimation.tsx`

    ```tsx
    'use client';

    import { ReactNode, useRef, useEffect, useState } from 'react';
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

      // 스프링 애니메이션 설정
      const { scale, opacity } = useSpring({
        scale: isReady ? 1 : 0,
        opacity: isReady ? 1 : 0,
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

      // 초기 회전 애니메이션
      useFrame((_, delta) => {
        if (groupRef.current && !hasCompleted) {
          groupRef.current.rotation.y += delta * 0.5;
        }
      });

      return (
        <animated.group
          ref={groupRef}
          scale={scale}
          // @ts-expect-error - animated props type issue
          material-opacity={opacity}
        >
          {children}
        </animated.group>
      );
    }
    ```

- [ ] **노드 개별 애니메이션 래퍼**
  - [ ] `src/components/mindmap/AnimatedNode.tsx`

    ```tsx
    'use client';

    import { useRef, useMemo } from 'react';
    import { useFrame } from '@react-three/fiber';
    import { useSpring, animated } from '@react-spring/three';
    import * as THREE from 'three';
    import type { MindmapNode } from '@/types/mindmap';

    interface AnimatedNodeProps {
      node: MindmapNode;
      index: number;
      isReady: boolean;
      isSelected: boolean;
      isHovered: boolean;
      onClick: () => void;
      onPointerOver: () => void;
      onPointerOut: () => void;
    }

    export function AnimatedNode({
      node,
      index,
      isReady,
      isSelected,
      isHovered,
      onClick,
      onPointerOver,
      onPointerOut,
    }: AnimatedNodeProps) {
      const meshRef = useRef<THREE.Mesh>(null);

      // 노드별 딜레이를 적용한 등장 애니메이션
      const delay = index * 50; // 50ms씩 딜레이

      const { position, scale } = useSpring({
        position: isReady
          ? [node.position?.x || 0, node.position?.y || 0, node.position?.z || 0]
          : [0, 0, 0],
        scale: isReady ? (isHovered || isSelected ? 1.2 : 1) : 0,
        delay: isReady ? delay : 0,
        config: {
          mass: 1,
          tension: 170,
          friction: 26,
        },
      });

      // Hover/Select 시 글로우 효과
      const glowIntensity = useMemo(() => {
        if (isSelected) return 0.5;
        if (isHovered) return 0.3;
        return 0.1;
      }, [isSelected, isHovered]);

      // 노드 타입별 크기 계수
      const sizeMultiplier = useMemo(() => {
        switch (node.type) {
          case 'core':
            return 1.5;
          case 'topic':
            return 1.2;
          case 'subtopic':
            return 1;
          case 'page':
            return 0.8;
          default:
            return 1;
        }
      }, [node.type]);

      const actualSize = node.size * sizeMultiplier;

      // 선택된 노드 펄스 애니메이션
      useFrame((state) => {
        if (meshRef.current && isSelected) {
          const pulse = Math.sin(state.clock.elapsedTime * 3) * 0.05 + 1;
          meshRef.current.scale.setScalar(actualSize * pulse);
        }
      });

      return (
        <animated.group position={position as unknown as [number, number, number]}>
          <animated.mesh
            ref={meshRef}
            scale={scale.to((s) => s * actualSize)}
            onClick={(e) => {
              e.stopPropagation();
              onClick();
            }}
            onPointerOver={(e) => {
              e.stopPropagation();
              onPointerOver();
              document.body.style.cursor = 'pointer';
            }}
            onPointerOut={(e) => {
              e.stopPropagation();
              onPointerOut();
              document.body.style.cursor = 'default';
            }}
          >
            <sphereGeometry args={[1, 32, 32]} />
            <meshStandardMaterial
              color={node.color}
              emissive={node.color}
              emissiveIntensity={glowIntensity}
              roughness={0.3}
              metalness={0.7}
            />
          </animated.mesh>

          {/* Glow sphere */}
          <animated.mesh scale={scale.to((s) => s * actualSize * 1.3)}>
            <sphereGeometry args={[1, 16, 16]} />
            <meshBasicMaterial
              color={node.color}
              transparent
              opacity={glowIntensity * 0.3}
              side={THREE.BackSide}
            />
          </animated.mesh>
        </animated.group>
      );
    }
    ```

- [ ] **파티클 배경 효과**
  - [ ] `src/components/mindmap/ParticleField.tsx`

    ```tsx
    'use client';

    import { useRef, useMemo } from 'react';
    import { useFrame } from '@react-three/fiber';
    import * as THREE from 'three';

    interface ParticleFieldProps {
      count?: number;
      radius?: number;
      size?: number;
      color?: string;
    }

    export function ParticleField({
      count = 2000,
      radius = 400,
      size = 1.5,
      color = '#ffffff',
    }: ParticleFieldProps) {
      const pointsRef = useRef<THREE.Points>(null);

      // 파티클 위치 생성 (구형 분포)
      const positions = useMemo(() => {
        const pos = new Float32Array(count * 3);

        for (let i = 0; i < count; i++) {
          // 구형 좌표계로 균일 분포
          const theta = Math.random() * Math.PI * 2;
          const phi = Math.acos(2 * Math.random() - 1);
          const r = radius * Math.cbrt(Math.random()); // 균일한 볼륨 분포

          pos[i * 3] = r * Math.sin(phi) * Math.cos(theta);
          pos[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
          pos[i * 3 + 2] = r * Math.cos(phi);
        }

        return pos;
      }, [count, radius]);

      // 파티클 크기 랜덤화
      const sizes = useMemo(() => {
        const s = new Float32Array(count);
        for (let i = 0; i < count; i++) {
          s[i] = size * (0.5 + Math.random() * 0.5);
        }
        return s;
      }, [count, size]);

      // 느린 회전 애니메이션
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
            <bufferAttribute
              attach="attributes-size"
              args={[sizes, 1]}
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
    ```

- [ ] **성운 효과 (Nebula)**
  - [ ] `src/components/mindmap/NebulaEffect.tsx`

    ```tsx
    'use client';

    import { useRef, useMemo } from 'react';
    import { useFrame } from '@react-three/fiber';
    import * as THREE from 'three';

    interface NebulaEffectProps {
      count?: number;
      radius?: number;
      colors?: string[];
    }

    export function NebulaEffect({
      count = 500,
      radius = 300,
      colors = ['#4F46E5', '#7C3AED', '#2563EB'],
    }: NebulaEffectProps) {
      const pointsRef = useRef<THREE.Points>(null);

      const { positions, colorArray } = useMemo(() => {
        const pos = new Float32Array(count * 3);
        const col = new Float32Array(count * 3);
        const colorObjects = colors.map((c) => new THREE.Color(c));

        for (let i = 0; i < count; i++) {
          // 나선형 분포
          const angle = (i / count) * Math.PI * 8;
          const r = (i / count) * radius;
          const noise = (Math.random() - 0.5) * 50;

          pos[i * 3] = Math.cos(angle) * r + noise;
          pos[i * 3 + 1] = (Math.random() - 0.5) * 100;
          pos[i * 3 + 2] = Math.sin(angle) * r + noise;

          // 랜덤 색상 선택
          const color = colorObjects[Math.floor(Math.random() * colorObjects.length)];
          col[i * 3] = color.r;
          col[i * 3 + 1] = color.g;
          col[i * 3 + 2] = color.b;
        }

        return { positions: pos, colorArray: col };
      }, [count, radius, colors]);

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
    ```

- [ ] **Post-processing 설정 개선**
  - [ ] `src/components/mindmap/PostProcessing.tsx`

    ```tsx
    'use client';

    import { EffectComposer, Bloom, Vignette } from '@react-three/postprocessing';
    import { BlendFunction } from 'postprocessing';

    interface PostProcessingProps {
      bloomIntensity?: number;
      bloomThreshold?: number;
    }

    export function PostProcessing({
      bloomIntensity = 0.5,
      bloomThreshold = 0.2,
    }: PostProcessingProps) {
      return (
        <EffectComposer>
          <Bloom
            luminanceThreshold={bloomThreshold}
            luminanceSmoothing={0.9}
            intensity={bloomIntensity}
            mipmapBlur
          />
          <Vignette
            offset={0.3}
            darkness={0.5}
            blendFunction={BlendFunction.NORMAL}
          />
        </EffectComposer>
      );
    }
    ```

### 검증

```bash
pnpm dev
# 세션 상세 > 마인드맵 탭에서:
# 1. 빅뱅 애니메이션 확인 (처음 로드 시)
# 2. 노드별 순차 등장 애니메이션 확인
# 3. 배경 파티클 확인
# 4. 성운 효과 확인
# 5. Bloom 효과 확인 (노드 glow)
# 6. 비네팅 효과 확인
```

---

## Step 11.5.2: 카메라 및 인터랙션

### 체크리스트

- [ ] **카메라 컨트롤러**
  - [ ] `src/components/mindmap/CameraController.tsx`

    ```tsx
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
          // 선택된 노드를 향해 카메라 이동
          const nodePos = new THREE.Vector3(
            selectedNode.position.x,
            selectedNode.position.y,
            selectedNode.position.z
          );

          // 노드에서 카메라 방향으로 focusDistance만큼 떨어진 위치
          const direction = new THREE.Vector3()
            .subVectors(camera.position, nodePos)
            .normalize();

          targetPosition.current.copy(nodePos).add(direction.multiplyScalar(focusDistance));
          targetLookAt.current.copy(nodePos);
        } else {
          // 기본 위치로 복귀
          targetPosition.current.set(...defaultPosition);
          targetLookAt.current.set(0, 0, 0);
        }
      }, [selectedNode, camera.position, defaultPosition, focusDistance]);

      useFrame(() => {
        // 부드러운 카메라 이동 (lerp)
        camera.position.lerp(targetPosition.current, lerpFactor);

        // 카메라가 바라보는 지점 업데이트
        const currentLookAt = new THREE.Vector3();
        camera.getWorldDirection(currentLookAt);
        currentLookAt.lerp(
          targetLookAt.current.clone().sub(camera.position).normalize(),
          lerpFactor
        );
      });

      return null;
    }
    ```

- [ ] **자동 회전 카메라 (Idle 상태)**
  - [ ] `src/components/mindmap/AutoRotateCamera.tsx`

    ```tsx
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
      const targetY = useRef(0);

      useEffect(() => {
        if (enabled) {
          // 현재 카메라 위치에서 각도 계산
          angleRef.current = Math.atan2(camera.position.x, camera.position.z);
        }
      }, [enabled, camera.position]);

      useFrame((_, delta) => {
        if (!enabled) return;

        angleRef.current += speed * delta;

        const x = Math.sin(angleRef.current) * radius;
        const z = Math.cos(angleRef.current) * radius;
        const y = camera.position.y + (targetY.current - camera.position.y) * 0.02;

        camera.position.set(x, y, z);
        camera.lookAt(0, 0, 0);
      });

      return null;
    }
    ```

- [ ] **궤도 애니메이션 (자식 노드)**
  - [ ] `src/components/mindmap/OrbitAnimation.tsx`

    ```tsx
    'use client';

    import { useRef, useMemo } from 'react';
    import { useFrame } from '@react-three/fiber';
    import * as THREE from 'three';

    interface OrbitAnimationProps {
      children: React.ReactNode;
      centerPosition: [number, number, number];
      radius: number;
      speed?: number;
      offset?: number;
      tilt?: number;
    }

    export function OrbitAnimation({
      children,
      centerPosition,
      radius,
      speed = 0.5,
      offset = 0,
      tilt = 0,
    }: OrbitAnimationProps) {
      const groupRef = useRef<THREE.Group>(null);

      // 틸트 행렬
      const tiltMatrix = useMemo(() => {
        const matrix = new THREE.Matrix4();
        matrix.makeRotationX(tilt);
        return matrix;
      }, [tilt]);

      useFrame((state) => {
        if (groupRef.current) {
          const angle = state.clock.elapsedTime * speed + offset;

          // 궤도 위치 계산
          let position = new THREE.Vector3(
            Math.cos(angle) * radius,
            0,
            Math.sin(angle) * radius
          );

          // 틸트 적용
          position.applyMatrix4(tiltMatrix);

          // 중심 위치 더하기
          position.add(new THREE.Vector3(...centerPosition));

          groupRef.current.position.copy(position);
        }
      });

      return <group ref={groupRef}>{children}</group>;
    }
    ```

- [ ] **마인드맵 인터랙션 훅**
  - [ ] `src/lib/hooks/use-mindmap-interaction.ts`

    ```typescript
    import { useState, useCallback, useEffect } from 'react';
    import type { MindmapNode } from '@/types/mindmap';

    interface UseMindmapInteractionOptions {
      onNodeSelect?: (node: MindmapNode | null) => void;
      autoRotateDelay?: number;
    }

    export function useMindmapInteraction(options: UseMindmapInteractionOptions = {}) {
      const { onNodeSelect, autoRotateDelay = 5000 } = options;

      const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);
      const [hoveredNode, setHoveredNode] = useState<MindmapNode | null>(null);
      const [isIdle, setIsIdle] = useState(true);

      // 노드 선택 핸들러
      const handleNodeClick = useCallback(
        (node: MindmapNode) => {
          const newNode = selectedNode?.id === node.id ? null : node;
          setSelectedNode(newNode);
          setIsIdle(false);
          onNodeSelect?.(newNode);
        },
        [selectedNode, onNodeSelect]
      );

      // 호버 핸들러
      const handleNodeHover = useCallback((node: MindmapNode | null) => {
        setHoveredNode(node);
        if (node) {
          setIsIdle(false);
        }
      }, []);

      // 배경 클릭 시 선택 해제
      const handleBackgroundClick = useCallback(() => {
        setSelectedNode(null);
        onNodeSelect?.(null);
      }, [onNodeSelect]);

      // Idle 감지 (일정 시간 상호작용 없으면 auto-rotate 활성화)
      useEffect(() => {
        if (selectedNode || hoveredNode) {
          setIsIdle(false);
          return;
        }

        const timer = setTimeout(() => {
          setIsIdle(true);
        }, autoRotateDelay);

        return () => clearTimeout(timer);
      }, [selectedNode, hoveredNode, autoRotateDelay]);

      return {
        selectedNode,
        hoveredNode,
        isIdle,
        handleNodeClick,
        handleNodeHover,
        handleBackgroundClick,
      };
    }
    ```

- [ ] **Galaxy 컴포넌트 업데이트 (애니메이션 통합)**
  - [ ] `src/components/mindmap/Galaxy.tsx` 업데이트

    ```tsx
    'use client';

    import { useState, useCallback, useMemo } from 'react';
    import { ThreeEvent } from '@react-three/fiber';
    import { Text } from '@react-three/drei';
    import type { MindmapNode, MindmapEdge } from '@/types/mindmap';
    import { BigBangAnimation } from './BigBangAnimation';
    import { AnimatedNode } from './AnimatedNode';
    import { Edge } from './Edge';
    import { ParticleField } from './ParticleField';
    import { NebulaEffect } from './NebulaEffect';
    import { CameraController } from './CameraController';
    import { AutoRotateCamera } from './AutoRotateCamera';
    import { PostProcessing } from './PostProcessing';
    import { useMindmapInteraction } from '@/lib/hooks/use-mindmap-interaction';

    interface GalaxyProps {
      nodes: MindmapNode[];
      edges: MindmapEdge[];
      onNodeClick?: (node: MindmapNode) => void;
      selectedNodeId?: string;
      showLabels?: boolean;
      enableAnimation?: boolean;
      enableAutoRotate?: boolean;
    }

    export function Galaxy({
      nodes,
      edges,
      onNodeClick,
      selectedNodeId,
      showLabels = true,
      enableAnimation = true,
      enableAutoRotate = true,
    }: GalaxyProps) {
      const [isAnimationReady, setIsAnimationReady] = useState(!enableAnimation);
      const [animationComplete, setAnimationComplete] = useState(!enableAnimation);

      const {
        selectedNode,
        hoveredNode,
        isIdle,
        handleNodeClick,
        handleNodeHover,
        handleBackgroundClick,
      } = useMindmapInteraction({
        onNodeSelect: onNodeClick,
      });

      // 노드 맵 생성 (Edge 렌더링용)
      const nodeMap = useMemo(() => {
        const map = new Map<string, MindmapNode>();
        nodes.forEach((node) => map.set(node.id, node));
        return map;
      }, [nodes]);

      // 선택된 노드 찾기
      const currentSelectedNode = useMemo(() => {
        if (selectedNodeId) {
          return nodes.find((n) => n.id === selectedNodeId) || null;
        }
        return selectedNode;
      }, [selectedNodeId, selectedNode, nodes]);

      // 배경 클릭 핸들러
      const handleCanvasClick = useCallback(
        (e: ThreeEvent<MouseEvent>) => {
          // 노드가 아닌 곳 클릭 시에만 처리
          if (e.object.type === 'Mesh') return;
          handleBackgroundClick();
        },
        [handleBackgroundClick]
      );

      // 애니메이션 시작 (컴포넌트 마운트 후)
      useState(() => {
        if (enableAnimation) {
          const timer = setTimeout(() => setIsAnimationReady(true), 100);
          return () => clearTimeout(timer);
        }
      });

      return (
        <>
          {/* 카메라 컨트롤 */}
          <CameraController selectedNode={currentSelectedNode} />
          {enableAutoRotate && (
            <AutoRotateCamera enabled={isIdle && animationComplete} speed={0.05} />
          )}

          {/* 배경 효과 */}
          <ParticleField count={3000} radius={500} size={1} color="#ffffff" />
          <NebulaEffect count={300} radius={350} />

          {/* 클릭 감지용 투명 평면 */}
          <mesh
            position={[0, 0, -100]}
            onClick={handleCanvasClick}
            visible={false}
          >
            <planeGeometry args={[2000, 2000]} />
            <meshBasicMaterial transparent opacity={0} />
          </mesh>

          {/* 메인 콘텐츠 */}
          <BigBangAnimation
            isReady={isAnimationReady}
            onComplete={() => setAnimationComplete(true)}
            duration={1500}
          >
            {/* Edges */}
            {edges.map((edge) => {
              const sourceNode = nodeMap.get(edge.source);
              const targetNode = nodeMap.get(edge.target);

              if (!sourceNode?.position || !targetNode?.position) return null;

              const isConnectedToSelected =
                currentSelectedNode &&
                (edge.source === currentSelectedNode.id ||
                  edge.target === currentSelectedNode.id);

              return (
                <Edge
                  key={`${edge.source}-${edge.target}`}
                  start={[
                    sourceNode.position.x,
                    sourceNode.position.y,
                    sourceNode.position.z,
                  ]}
                  end={[
                    targetNode.position.x,
                    targetNode.position.y,
                    targetNode.position.z,
                  ]}
                  weight={edge.weight}
                  color={isConnectedToSelected ? '#60A5FA' : '#4B5563'}
                  opacity={isConnectedToSelected ? 0.8 : 0.3}
                />
              );
            })}

            {/* Nodes */}
            {nodes.map((node, index) => {
              const isSelected = currentSelectedNode?.id === node.id;
              const isHovered = hoveredNode?.id === node.id;

              return (
                <group key={node.id}>
                  <AnimatedNode
                    node={node}
                    index={index}
                    isReady={isAnimationReady}
                    isSelected={isSelected}
                    isHovered={isHovered}
                    onClick={() => handleNodeClick(node)}
                    onPointerOver={() => handleNodeHover(node)}
                    onPointerOut={() => handleNodeHover(null)}
                  />

                  {/* Label */}
                  {showLabels && node.position && (
                    <Text
                      position={[
                        node.position.x,
                        node.position.y + node.size * 1.5 + 10,
                        node.position.z,
                      ]}
                      fontSize={node.type === 'core' ? 14 : node.type === 'topic' ? 12 : 10}
                      color="white"
                      anchorX="center"
                      anchorY="middle"
                      outlineWidth={0.5}
                      outlineColor="#000000"
                      visible={isSelected || isHovered || node.type === 'core'}
                    >
                      {node.label}
                    </Text>
                  )}
                </group>
              );
            })}
          </BigBangAnimation>

          {/* Post-processing */}
          <PostProcessing bloomIntensity={0.6} bloomThreshold={0.15} />
        </>
      );
    }
    ```

- [ ] **성능 최적화 유틸**
  - [ ] `src/lib/utils/three-performance.ts`

    ```typescript
    import { useMemo } from 'react';
    import type { MindmapNode } from '@/types/mindmap';

    // LOD (Level of Detail) 설정
    export interface LODConfig {
      highDetail: number;    // 이 거리 이내: 고품질
      mediumDetail: number;  // 이 거리 이내: 중간 품질
      lowDetail: number;     // 이 거리 이후: 저품질
    }

    export const DEFAULT_LOD: LODConfig = {
      highDetail: 100,
      mediumDetail: 300,
      lowDetail: 500,
    };

    // 노드 수에 따른 세그먼트 수 결정
    export function getOptimalSegments(nodeCount: number): number {
      if (nodeCount > 100) return 8;
      if (nodeCount > 50) return 16;
      return 32;
    }

    // 파티클 수 최적화
    export function getOptimalParticleCount(nodeCount: number): number {
      if (nodeCount > 100) return 1000;
      if (nodeCount > 50) return 2000;
      return 3000;
    }

    // 노드 필터링 (뷰포트 밖 노드 제외)
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

    // 디바운스된 상태 업데이트
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
    ```

### 검증

```bash
pnpm dev
# 마인드맵에서:
# 1. 빅뱅 애니메이션 (노드가 중심에서 퍼짐) 확인
# 2. 노드별 순차 등장 애니메이션 확인
# 3. 노드 클릭 시 카메라 이동 확인
# 4. 부드러운 transition 확인
# 5. 선택 해제 시 기본 위치 복귀 확인
# 6. 5초 Idle 후 자동 회전 확인
# 7. 파티클/성운 배경 효과 확인
# 8. 많은 노드에서 성능 확인 (60fps 유지)
```

---

## Phase 11.5 완료 확인

### 전체 검증 체크리스트

- [ ] 빅뱅 초기 애니메이션
- [ ] 노드별 순차 등장 애니메이션
- [ ] 파티클 배경 효과
- [ ] 성운 효과
- [ ] Bloom/Vignette post-processing
- [ ] 노드 선택 시 카메라 이동
- [ ] 부드러운 lerp transition
- [ ] Idle 시 자동 회전
- [ ] 성능 최적화 (60fps)

### 테스트

```bash
moonx web:typecheck
moonx web:lint
moonx web:build
```

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| 빅뱅 애니메이션 | `src/components/mindmap/BigBangAnimation.tsx` |
| 애니메이션 노드 | `src/components/mindmap/AnimatedNode.tsx` |
| 파티클 효과 | `src/components/mindmap/ParticleField.tsx` |
| 성운 효과 | `src/components/mindmap/NebulaEffect.tsx` |
| Post-processing | `src/components/mindmap/PostProcessing.tsx` |
| 카메라 컨트롤러 | `src/components/mindmap/CameraController.tsx` |
| 자동 회전 | `src/components/mindmap/AutoRotateCamera.tsx` |
| 궤도 애니메이션 | `src/components/mindmap/OrbitAnimation.tsx` |
| 인터랙션 훅 | `src/lib/hooks/use-mindmap-interaction.ts` |
| 성능 유틸 | `src/lib/utils/three-performance.ts` |
| Galaxy (통합) | `src/components/mindmap/Galaxy.tsx` |

---

## Phase 11 시리즈 완료

Phase 11.5 완료로 Phase 11 시리즈가 완료됩니다.

### Phase 11 전체 산출물

| Phase | 내용 | 주요 산출물 |
|-------|------|-------------|
| 11.1 | Three.js 환경 설정 | MindmapCanvas, 타입 정의, 테스트 페이지 |
| 11.2 | Node, Edge, Galaxy 컴포넌트 | 3D 노드/엣지/갤럭시, 레이아웃 유틸 |
| 11.3 | 세션 상세 마인드맵 통합 | API 연동, 탭 UI, 노드 상세 패널 |
| 11.4 | 계정/사용량 페이지 | 구독/사용량 API, 계정 페이지 |
| 11.5 | 애니메이션 및 인터랙션 | 빅뱅 애니메이션, 카메라 제어, 파티클 효과 |

### 최종 검증

```bash
# 전체 빌드 테스트
moonx web:typecheck
moonx web:lint
moonx web:build

# 개발 서버 실행
moonx web:dev

# 확인 항목:
# 1. /test-3d - 3D 렌더링 기본 테스트
# 2. /sessions/[id] - 마인드맵 탭 (빅뱅 애니메이션, 노드 상호작용)
# 3. /account - 계정 및 사용량 페이지
```

---

## 다음 Phase

Phase 11 시리즈 완료 후 [Phase 12: 프로덕션 모니터링](./phase-12-monitoring.md)으로 진행하세요.
