# Phase 11.2: 3D 마인드맵 컴포넌트

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 3D 마인드맵 Node, Edge, Galaxy 컴포넌트 구현 |
| **선행 조건** | Phase 11.1 완료 |
| **예상 소요** | 2 Steps |
| **결과물** | 인터랙티브 3D 마인드맵 렌더링 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 11.2.1 | Node 및 Edge 컴포넌트 | ⬜ |
| 11.2.2 | Galaxy 컴포넌트 | ⬜ |

---

## Step 11.2.1: Node 및 Edge 컴포넌트

### 체크리스트

- [ ] **Node 컴포넌트 (행성/위성)**
  - [ ] `src/components/mindmap/Node.tsx`

    ```tsx
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
    ```

- [ ] **Edge 컴포넌트 (연결선)**
  - [ ] `src/components/mindmap/Edge.tsx`

    ```tsx
    'use client';

    import { useMemo, useRef } from 'react';
    import { useFrame } from '@react-three/fiber';
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
      const lineRef = useRef<THREE.Line>(null);

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

      // Animate dash offset for highlighted edges
      useFrame((_, delta) => {
        if (lineRef.current && isHighlighted) {
          const material = lineRef.current.material as THREE.LineDashedMaterial;
          if (material.dashOffset !== undefined) {
            material.dashOffset -= delta * 5;
          }
        }
      });

      // Line width based on edge weight
      const lineWidth = isHighlighted ? 2 : Math.max(0.5, Math.min(edge.weight * 0.5, 1.5));

      return (
        <Line
          ref={lineRef}
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
    ```

- [ ] **노드 색상 유틸리티**
  - [ ] `src/lib/mindmap-utils.ts`

    ```typescript
    import type { MindmapNodeType } from '@/types/mindmap';

    // 노드 타입별 기본 색상
    export const NODE_TYPE_COLORS: Record<MindmapNodeType, string> = {
      core: '#F59E0B',     // Amber - 중심 노드
      topic: '#3B82F6',    // Blue - 주제
      subtopic: '#10B981', // Emerald - 하위 주제
      page: '#8B5CF6',     // Violet - 페이지
    };

    // 주제별 색상 팔레트
    export const TOPIC_COLORS = [
      '#3B82F6', // Blue
      '#10B981', // Emerald
      '#F59E0B', // Amber
      '#EF4444', // Red
      '#8B5CF6', // Violet
      '#EC4899', // Pink
      '#06B6D4', // Cyan
      '#F97316', // Orange
    ];

    export function getTopicColor(index: number): string {
      return TOPIC_COLORS[index % TOPIC_COLORS.length];
    }

    // 노드 크기 계산
    export function calculateNodeSize(
      type: MindmapNodeType,
      visitCount?: number,
      totalDuration?: number
    ): number {
      const baseSize: Record<MindmapNodeType, number> = {
        core: 50,
        topic: 30,
        subtopic: 20,
        page: 15,
      };

      let size = baseSize[type];

      // 방문 횟수에 따른 크기 조정
      if (visitCount && visitCount > 1) {
        size *= Math.min(1 + visitCount * 0.1, 1.5);
      }

      // 체류 시간에 따른 크기 조정 (밀리초 → 분)
      if (totalDuration && totalDuration > 60000) {
        const minutes = totalDuration / 60000;
        size *= Math.min(1 + minutes * 0.05, 1.3);
      }

      return size;
    }
    ```

### 검증

```bash
cd apps/web
pnpm dev

# 테스트 페이지에서 Node, Edge 개별 테스트
# 1. Node 컴포넌트 import 후 렌더링
# 2. 호버 시 스케일 애니메이션 확인
# 3. Edge 곡선 렌더링 확인
```

---

## Step 11.2.2: Galaxy 컴포넌트

### 체크리스트

- [ ] **Galaxy 컴포넌트 (전체 마인드맵)**
  - [ ] `src/components/mindmap/Galaxy.tsx`

    ```tsx
    'use client';

    import { useState, useCallback, useMemo } from 'react';
    import { useThree } from '@react-three/fiber';
    import * as THREE from 'three';
    import { Node } from './Node';
    import { Edge } from './Edge';
    import type { MindmapData, MindmapNode } from '@/types/mindmap';

    interface GalaxyProps {
      data: MindmapData;
      onNodeSelect: (node: MindmapNode | null) => void;
    }

    export function Galaxy({ data, onNodeSelect }: GalaxyProps) {
      const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
      const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);
      const { controls } = useThree();

      // Create node map for edge lookup
      const nodeMap = useMemo(() => {
        return new Map(data.nodes.map((node) => [node.id, node]));
      }, [data.nodes]);

      // Get connected node IDs for highlighting
      const connectedNodeIds = useMemo(() => {
        const targetId = selectedNodeId || hoveredNodeId;
        if (!targetId) return new Set<string>();

        const connected = new Set<string>();

        data.edges.forEach((edge) => {
          if (edge.source === targetId) {
            connected.add(edge.target);
          }
          if (edge.target === targetId) {
            connected.add(edge.source);
          }
        });

        return connected;
      }, [data.edges, selectedNodeId, hoveredNodeId]);

      // Handle node click
      const handleNodeClick = useCallback(
        (node: MindmapNode) => {
          if (selectedNodeId === node.id) {
            // Deselect
            setSelectedNodeId(null);
            onNodeSelect(null);
          } else {
            // Select
            setSelectedNodeId(node.id);
            onNodeSelect(node);

            // Animate camera to node
            if (node.position && controls) {
              const orbitControls = controls as any;
              if (orbitControls.target) {
                // Smoothly move camera target to node position
                orbitControls.target.set(
                  node.position.x,
                  node.position.y,
                  node.position.z
                );
              }
            }
          }
        },
        [selectedNodeId, onNodeSelect, controls]
      );

      // Handle node hover
      const handleNodeHover = useCallback((nodeId: string, hovered: boolean) => {
        setHoveredNodeId(hovered ? nodeId : null);
      }, []);

      // Handle background click (deselect)
      const handleBackgroundClick = useCallback(() => {
        if (selectedNodeId) {
          setSelectedNodeId(null);
          onNodeSelect(null);
        }
      }, [selectedNodeId, onNodeSelect]);

      return (
        <group onClick={handleBackgroundClick}>
          {/* Invisible background plane for click detection */}
          <mesh position={[0, 0, -500]} visible={false}>
            <planeGeometry args={[2000, 2000]} />
            <meshBasicMaterial transparent opacity={0} />
          </mesh>

          {/* Render edges first (behind nodes) */}
          {data.edges.map((edge) => {
            const sourceNode = nodeMap.get(edge.source);
            const targetNode = nodeMap.get(edge.target);

            if (!sourceNode || !targetNode) return null;

            const isHighlighted =
              edge.source === selectedNodeId ||
              edge.target === selectedNodeId ||
              edge.source === hoveredNodeId ||
              edge.target === hoveredNodeId;

            return (
              <Edge
                key={`${edge.source}-${edge.target}`}
                edge={edge}
                sourceNode={sourceNode}
                targetNode={targetNode}
                isHighlighted={isHighlighted}
              />
            );
          })}

          {/* Render nodes */}
          {data.nodes.map((node) => (
            <Node
              key={node.id}
              node={node}
              isSelected={node.id === selectedNodeId}
              isHovered={
                node.id === hoveredNodeId ||
                connectedNodeIds.has(node.id)
              }
              onClick={() => handleNodeClick(node)}
              onHover={(hovered) => handleNodeHover(node.id, hovered)}
            />
          ))}
        </group>
      );
    }
    ```

- [ ] **테스트용 Mock 데이터**
  - [ ] `src/lib/mock-mindmap-data.ts`

    ```typescript
    import type { MindmapData } from '@/types/mindmap';

    export const mockMindmapData: MindmapData = {
      nodes: [
        {
          id: 'core',
          label: '브라우징 세션',
          type: 'core',
          size: 50,
          color: '#F59E0B',
          position: { x: 0, y: 0, z: 0 },
          data: { description: '2024년 12월 31일 세션' },
        },
        {
          id: 'topic-1',
          label: '개발',
          type: 'topic',
          size: 35,
          color: '#3B82F6',
          position: { x: -150, y: 50, z: 30 },
          data: { visitCount: 5 },
        },
        {
          id: 'topic-2',
          label: '디자인',
          type: 'topic',
          size: 30,
          color: '#10B981',
          position: { x: 100, y: -30, z: -50 },
          data: { visitCount: 3 },
        },
        {
          id: 'topic-3',
          label: '뉴스',
          type: 'topic',
          size: 25,
          color: '#EF4444',
          position: { x: 50, y: 120, z: 80 },
          data: { visitCount: 2 },
        },
        {
          id: 'sub-1-1',
          label: 'React',
          type: 'subtopic',
          size: 20,
          color: '#60A5FA',
          position: { x: -220, y: 100, z: 60 },
          data: { visitCount: 3 },
        },
        {
          id: 'sub-1-2',
          label: 'TypeScript',
          type: 'subtopic',
          size: 18,
          color: '#60A5FA',
          position: { x: -200, y: -20, z: 0 },
          data: { visitCount: 2 },
        },
        {
          id: 'sub-2-1',
          label: 'Figma',
          type: 'subtopic',
          size: 22,
          color: '#34D399',
          position: { x: 180, y: -80, z: -30 },
          data: { visitCount: 2 },
        },
        {
          id: 'page-1',
          label: 'React 공식 문서',
          type: 'page',
          size: 12,
          color: '#93C5FD',
          position: { x: -280, y: 150, z: 90 },
          data: { urls: ['https://react.dev'] },
        },
        {
          id: 'page-2',
          label: 'MDN Web Docs',
          type: 'page',
          size: 10,
          color: '#93C5FD',
          position: { x: -250, y: 50, z: 100 },
          data: { urls: ['https://developer.mozilla.org'] },
        },
      ],
      edges: [
        { source: 'core', target: 'topic-1', weight: 1 },
        { source: 'core', target: 'topic-2', weight: 0.8 },
        { source: 'core', target: 'topic-3', weight: 0.5 },
        { source: 'topic-1', target: 'sub-1-1', weight: 0.9 },
        { source: 'topic-1', target: 'sub-1-2', weight: 0.7 },
        { source: 'topic-2', target: 'sub-2-1', weight: 0.8 },
        { source: 'sub-1-1', target: 'page-1', weight: 0.6 },
        { source: 'sub-1-1', target: 'page-2', weight: 0.4 },
      ],
      layout: {
        type: 'galaxy',
        params: {},
      },
    };
    ```

- [ ] **테스트 페이지 업데이트**
  - [ ] `src/app/(dashboard)/test-3d/page.tsx` 수정

    ```tsx
    'use client';

    import { useState } from 'react';
    import { MindmapCanvas } from '@/components/mindmap/MindmapCanvas';
    import { Galaxy } from '@/components/mindmap/Galaxy';
    import { mockMindmapData } from '@/lib/mock-mindmap-data';
    import type { MindmapNode } from '@/types/mindmap';

    export default function Test3DPage() {
      const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);

      return (
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">3D 마인드맵 테스트</h1>
            <p className="text-gray-500 mt-1">
              Galaxy 컴포넌트 렌더링 테스트
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* 3D Canvas */}
            <div className="lg:col-span-2 bg-white rounded-xl shadow-sm p-4">
              <h2 className="text-lg font-medium mb-4">마인드맵</h2>
              <MindmapCanvas className="h-[600px]">
                <Galaxy data={mockMindmapData} onNodeSelect={setSelectedNode} />
              </MindmapCanvas>
            </div>

            {/* Node Detail Panel */}
            <div className="bg-white rounded-xl shadow-sm p-4">
              <h2 className="text-lg font-medium mb-4">노드 정보</h2>

              {selectedNode ? (
                <div className="space-y-4">
                  <div className="flex items-center gap-3">
                    <div
                      className="w-4 h-4 rounded-full"
                      style={{ backgroundColor: selectedNode.color }}
                    />
                    <span className="font-medium">{selectedNode.label}</span>
                  </div>

                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-500">타입</span>
                      <span className="capitalize">{selectedNode.type}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">크기</span>
                      <span>{selectedNode.size}</span>
                    </div>
                    {selectedNode.position && (
                      <div className="flex justify-between">
                        <span className="text-gray-500">위치</span>
                        <span className="text-xs">
                          ({selectedNode.position.x.toFixed(0)},{' '}
                          {selectedNode.position.y.toFixed(0)},{' '}
                          {selectedNode.position.z.toFixed(0)})
                        </span>
                      </div>
                    )}
                    {selectedNode.data.visitCount && (
                      <div className="flex justify-between">
                        <span className="text-gray-500">방문 횟수</span>
                        <span>{selectedNode.data.visitCount}</span>
                      </div>
                    )}
                  </div>

                  {selectedNode.data.description && (
                    <div className="pt-2 border-t">
                      <p className="text-sm text-gray-600">
                        {selectedNode.data.description}
                      </p>
                    </div>
                  )}

                  {selectedNode.data.urls && selectedNode.data.urls.length > 0 && (
                    <div className="pt-2 border-t">
                      <p className="text-sm font-medium text-gray-500 mb-1">관련 URL</p>
                      <ul className="space-y-1">
                        {selectedNode.data.urls.map((url, i) => (
                          <li key={i} className="text-sm text-blue-600 truncate">
                            {url}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-gray-400 text-sm">
                  노드를 클릭하여 상세 정보를 확인하세요
                </p>
              )}
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-sm p-4">
            <h2 className="text-lg font-medium mb-2">조작 방법</h2>
            <ul className="text-sm text-gray-600 space-y-1">
              <li>• 마우스 드래그: 회전</li>
              <li>• 스크롤: 줌</li>
              <li>• 노드 클릭: 선택 및 상세 정보 표시</li>
              <li>• 노드 호버: 연결된 노드 하이라이트</li>
              <li>• 빈 공간 클릭: 선택 해제</li>
            </ul>
          </div>
        </div>
      );
    }
    ```

### 검증

```bash
cd apps/web
pnpm dev

# http://localhost:3000/test-3d 접속 후:
# 1. 전체 마인드맵 렌더링 확인 (9개 노드, 8개 엣지)
# 2. 노드 호버 시 스케일 업 + 연결 노드 하이라이트
# 3. 노드 클릭 시 선택 링 표시 + 우측 패널에 정보 표시
# 4. 엣지 곡선 렌더링 확인
# 5. 빈 공간 클릭 시 선택 해제
```

---

## Phase 11.2 완료 확인

### 전체 검증 체크리스트

- [ ] Node 컴포넌트 렌더링 (4가지 타입)
- [ ] Node 호버 애니메이션 (스케일, emissive)
- [ ] Node 선택 시 링 효과
- [ ] Node 라벨 표시 (호버/선택/core)
- [ ] Edge 컴포넌트 (곡선 bezier)
- [ ] Edge 하이라이트 (색상, 점선)
- [ ] Galaxy 전체 조합
- [ ] 노드 선택/해제 로직
- [ ] 연결 노드 하이라이트
- [ ] 배경 클릭 선택 해제

### 테스트

```bash
moonx web:typecheck
moonx web:lint
moonx web:build
```

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| Node 컴포넌트 | `src/components/mindmap/Node.tsx` |
| Edge 컴포넌트 | `src/components/mindmap/Edge.tsx` |
| Galaxy 컴포넌트 | `src/components/mindmap/Galaxy.tsx` |
| 유틸리티 | `src/lib/mindmap-utils.ts` |
| Mock 데이터 | `src/lib/mock-mindmap-data.ts` |

---

## 다음 Phase

Phase 11.2 완료 후 [Phase 11.3: 세션 상세 페이지 개선](./phase-11.3-session-detail.md)으로 진행하세요.
