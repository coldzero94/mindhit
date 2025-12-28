# Phase 11: 웹앱 대시보드 & 3D 마인드맵

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | React Three Fiber 기반 3D 마인드맵 시각화, 대시보드, 계정/사용량 페이지 |
| **선행 조건** | Phase 7, 9, 10 완료 |
| **예상 소요** | 5 Steps |
| **결과물** | 인터랙티브 3D "Knowledge Galaxy" 마인드맵 + 계정/사용량 UI |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 11.1 | React Three Fiber 설정 | ⬜ |
| 11.2 | 3D 마인드맵 컴포넌트 | ⬜ |
| 11.3 | 세션 상세 페이지 개선 | ⬜ |
| 11.4 | 애니메이션 & 인터랙션 | ⬜ |
| 11.5 | 계정 & 사용량 페이지 | ⬜ |

---

## Step 11.1: React Three Fiber 설정

### 체크리스트

- [ ] **의존성 설치**

  ```bash
  cd apps/web
  pnpm add three @react-three/fiber @react-three/drei @react-three/postprocessing
  pnpm add -D @types/three
  pnpm add framer-motion @react-spring/three
  ```

- [ ] **Three.js 타입 설정**
  - [ ] `src/types/three.d.ts`

    ```typescript
    import { Object3DNode } from '@react-three/fiber';
    import { Line2 } from 'three/examples/jsm/lines/Line2';
    import { LineGeometry } from 'three/examples/jsm/lines/LineGeometry';
    import { LineMaterial } from 'three/examples/jsm/lines/LineMaterial';

    declare module '@react-three/fiber' {
      interface ThreeElements {
        line2: Object3DNode<Line2, typeof Line2>;
        lineGeometry: Object3DNode<LineGeometry, typeof LineGeometry>;
        lineMaterial: Object3DNode<LineMaterial, typeof LineMaterial>;
      }
    }
    ```

- [ ] **마인드맵 타입 정의**
  - [ ] `src/types/mindmap.ts`

    ```typescript
    export interface MindmapNode {
      id: string;
      label: string;
      type: 'core' | 'topic' | 'subtopic' | 'page';
      size: number;
      color: string;
      position?: {
        x: number;
        y: number;
        z: number;
      };
      data: Record<string, unknown>;
    }

    export interface MindmapEdge {
      source: string;
      target: string;
      weight: number;
    }

    export interface MindmapLayout {
      type: 'galaxy' | 'tree' | 'radial';
      params: Record<string, unknown>;
    }

    export interface MindmapData {
      nodes: MindmapNode[];
      edges: MindmapEdge[];
      layout: MindmapLayout;
    }
    ```

- [ ] **Canvas Provider 설정**
  - [ ] `src/components/mindmap/MindmapCanvas.tsx`

    ```tsx
    'use client';

    import { Suspense } from 'react';
    import { Canvas } from '@react-three/fiber';
    import { OrbitControls, Stars, PerspectiveCamera } from '@react-three/drei';
    import { EffectComposer, Bloom } from '@react-three/postprocessing';

    interface MindmapCanvasProps {
      children: React.ReactNode;
    }

    export function MindmapCanvas({ children }: MindmapCanvasProps) {
      return (
        <div className="w-full h-full min-h-[600px] bg-gray-900 rounded-xl overflow-hidden">
          <Canvas>
            <PerspectiveCamera makeDefault position={[0, 0, 500]} fov={60} />

            {/* Lighting */}
            <ambientLight intensity={0.3} />
            <pointLight position={[100, 100, 100]} intensity={1} />
            <pointLight position={[-100, -100, -100]} intensity={0.5} />

            {/* Background */}
            <Stars
              radius={300}
              depth={50}
              count={5000}
              factor={4}
              saturation={0}
              fade
              speed={0.5}
            />

            {/* Controls */}
            <OrbitControls
              enablePan={true}
              enableZoom={true}
              enableRotate={true}
              minDistance={100}
              maxDistance={1000}
              dampingFactor={0.05}
            />

            {/* Post-processing */}
            <EffectComposer>
              <Bloom
                luminanceThreshold={0.2}
                luminanceSmoothing={0.9}
                intensity={0.5}
              />
            </EffectComposer>

            {/* Content */}
            <Suspense fallback={null}>
              {children}
            </Suspense>
          </Canvas>
        </div>
      );
    }
    ```

### 검증

```bash
pnpm dev
# http://localhost:3000 접속
# 3D Canvas 렌더링 확인 (별 배경 표시)
```

---

## Step 11.2: 3D 마인드맵 컴포넌트

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
    import { MindmapNode } from '@/types/mindmap';

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

      const getNodeSize = () => {
        switch (node.type) {
          case 'core':
            return node.size * 0.5;
          case 'topic':
            return node.size * 0.4;
          case 'subtopic':
            return node.size * 0.3;
          default:
            return node.size * 0.2;
        }
      };

      return (
        <group position={position}>
          {/* Main sphere */}
          <AnimatedSphere
            ref={meshRef}
            args={[getNodeSize(), 32, 32]}
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

          {/* Glow effect for core */}
          {node.type === 'core' && (
            <Sphere args={[getNodeSize() * 1.3, 32, 32]}>
              <meshBasicMaterial
                color={node.color}
                transparent
                opacity={0.15}
                side={THREE.BackSide}
              />
            </Sphere>
          )}

          {/* Label */}
          {(isHovered || localHover || isSelected || node.type === 'core') && (
            <Text
              position={[0, getNodeSize() + 15, 0]}
              fontSize={node.type === 'core' ? 14 : 10}
              color="white"
              anchorX="center"
              anchorY="middle"
              outlineWidth={0.5}
              outlineColor="black"
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
    import { MindmapEdge, MindmapNode } from '@/types/mindmap';

    interface EdgeProps {
      edge: MindmapEdge;
      sourceNode: MindmapNode;
      targetNode: MindmapNode;
      isHighlighted: boolean;
    }

    export function Edge({ edge, sourceNode, targetNode, isHighlighted }: EdgeProps) {
      const lineRef = useRef<THREE.Line>(null);

      const points = useMemo(() => {
        const source = sourceNode.position || { x: 0, y: 0, z: 0 };
        const target = targetNode.position || { x: 0, y: 0, z: 0 };

        // Create curved line using quadratic bezier
        const curve = new THREE.QuadraticBezierCurve3(
          new THREE.Vector3(source.x, source.y, source.z),
          new THREE.Vector3(
            (source.x + target.x) / 2,
            (source.y + target.y) / 2 + 20,
            (source.z + target.z) / 2
          ),
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

      return (
        <Line
          ref={lineRef}
          points={points}
          color={isHighlighted ? '#ffffff' : '#4B5563'}
          lineWidth={isHighlighted ? 2 : 1}
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

- [ ] **Galaxy 컴포넌트 (전체 마인드맵)**
  - [ ] `src/components/mindmap/Galaxy.tsx`

    ```tsx
    'use client';

    import { useState, useCallback, useMemo } from 'react';
    import { useThree } from '@react-three/fiber';
    import * as THREE from 'three';
    import { Node } from './Node';
    import { Edge } from './Edge';
    import { MindmapData, MindmapNode } from '@/types/mindmap';

    interface GalaxyProps {
      data: MindmapData;
      onNodeSelect: (node: MindmapNode | null) => void;
    }

    export function Galaxy({ data, onNodeSelect }: GalaxyProps) {
      const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
      const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);
      const { camera, controls } = useThree();

      // Create node map for edge lookup
      const nodeMap = useMemo(() => {
        return new Map(data.nodes.map((node) => [node.id, node]));
      }, [data.nodes]);

      // Get connected node IDs for highlighting
      const connectedNodeIds = useMemo(() => {
        if (!selectedNodeId && !hoveredNodeId) return new Set<string>();

        const targetId = selectedNodeId || hoveredNodeId;
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

      const handleNodeClick = useCallback(
        (node: MindmapNode) => {
          if (selectedNodeId === node.id) {
            setSelectedNodeId(null);
            onNodeSelect(null);
          } else {
            setSelectedNodeId(node.id);
            onNodeSelect(node);

            // Animate camera to node
            if (node.position && controls) {
              const targetPosition = new THREE.Vector3(
                node.position.x,
                node.position.y,
                node.position.z + 200
              );

              // Simple camera animation (for more complex, use gsap or spring)
              (controls as any).target.set(
                node.position.x,
                node.position.y,
                node.position.z
              );
            }
          }
        },
        [selectedNodeId, onNodeSelect, controls]
      );

      const handleNodeHover = useCallback((nodeId: string, hovered: boolean) => {
        setHoveredNodeId(hovered ? nodeId : null);
      }, []);

      return (
        <group>
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

### 검증

```bash
pnpm dev
# 마인드맵 데이터와 함께 Galaxy 컴포넌트 렌더링 확인
```

---

## Step 11.3: 세션 상세 페이지 개선

### 체크리스트

- [ ] **마인드맵 Hook**
  - [ ] `src/lib/hooks/use-mindmap.ts`

    ```typescript
    import { useQuery } from '@tanstack/react-query';
    import { apiClient } from '@/lib/api/client';
    import { MindmapData } from '@/types/mindmap';

    export function useMindmap(sessionId: string) {
      return useQuery({
        queryKey: ['mindmap', sessionId],
        queryFn: async () => {
          const response = await apiClient.get<{ mindmap: MindmapData }>(
            `/sessions/${sessionId}/mindmap`
          );
          return response.data.mindmap;
        },
        enabled: !!sessionId,
      });
    }
    ```

- [ ] **Node 상세 패널**
  - [ ] `src/components/mindmap/NodeDetailPanel.tsx`

    ```tsx
    'use client';

    import { motion, AnimatePresence } from 'framer-motion';
    import { X, ExternalLink, Clock, FileText } from 'lucide-react';
    import { MindmapNode } from '@/types/mindmap';
    import { Button } from '@/components/ui/button';
    import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

    interface NodeDetailPanelProps {
      node: MindmapNode | null;
      onClose: () => void;
    }

    export function NodeDetailPanel({ node, onClose }: NodeDetailPanelProps) {
      return (
        <AnimatePresence>
          {node && (
            <motion.div
              initial={{ x: 300, opacity: 0 }}
              animate={{ x: 0, opacity: 1 }}
              exit={{ x: 300, opacity: 0 }}
              transition={{ type: 'spring', damping: 25, stiffness: 300 }}
              className="absolute right-4 top-4 bottom-4 w-80 z-10"
            >
              <Card className="h-full overflow-auto">
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-lg">{node.label}</CardTitle>
                  <Button variant="ghost" size="icon" onClick={onClose}>
                    <X className="h-4 w-4" />
                  </Button>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Node type badge */}
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: node.color }}
                    />
                    <span className="text-sm text-gray-500 capitalize">
                      {node.type}
                    </span>
                  </div>

                  {/* Description */}
                  {node.data.description && (
                    <div className="space-y-1">
                      <h4 className="text-sm font-medium flex items-center gap-1">
                        <FileText className="h-4 w-4" />
                        설명
                      </h4>
                      <p className="text-sm text-gray-600">
                        {node.data.description as string}
                      </p>
                    </div>
                  )}

                  {/* Related URLs */}
                  {node.data.urls && (node.data.urls as string[]).length > 0 && (
                    <div className="space-y-2">
                      <h4 className="text-sm font-medium flex items-center gap-1">
                        <ExternalLink className="h-4 w-4" />
                        관련 페이지
                      </h4>
                      <ul className="space-y-1">
                        {(node.data.urls as string[]).slice(0, 5).map((url) => (
                          <li key={url} className="text-sm text-blue-600 truncate">
                            {url}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}

                  {/* Stats for core/topic nodes */}
                  {(node.type === 'core' || node.type === 'topic') && (
                    <div className="grid grid-cols-2 gap-2 pt-2">
                      <div className="bg-gray-50 rounded p-2">
                        <p className="text-xs text-gray-500">크기</p>
                        <p className="font-medium">{node.size.toFixed(0)}</p>
                      </div>
                      <div className="bg-gray-50 rounded p-2">
                        <p className="text-xs text-gray-500">위치</p>
                        <p className="font-medium text-xs">
                          {node.position
                            ? `${node.position.x.toFixed(0)}, ${node.position.y.toFixed(0)}`
                            : 'N/A'}
                        </p>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </motion.div>
          )}
        </AnimatePresence>
      );
    }
    ```

- [ ] **세션 상세 페이지 업데이트**
  - [ ] `src/app/(dashboard)/sessions/[id]/page.tsx` 업데이트

    ```tsx
    'use client';

    import { useState } from 'react';
    import { useParams, useRouter } from 'next/navigation';
    import { ArrowLeft, Trash2, Map, List } from 'lucide-react';

    import { useSession, useDeleteSession } from '@/lib/hooks/use-sessions';
    import { useMindmap } from '@/lib/hooks/use-mindmap';
    import { Button } from '@/components/ui/button';
    import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
    import { Skeleton } from '@/components/ui/skeleton';
    import { useToast } from '@/components/ui/use-toast';

    import { MindmapCanvas } from '@/components/mindmap/MindmapCanvas';
    import { Galaxy } from '@/components/mindmap/Galaxy';
    import { NodeDetailPanel } from '@/components/mindmap/NodeDetailPanel';
    import { SessionTimeline } from '@/components/sessions/SessionTimeline';
    import { SessionStats } from '@/components/sessions/SessionStats';
    import { MindmapNode } from '@/types/mindmap';

    export default function SessionDetailPage() {
      const params = useParams();
      const router = useRouter();
      const { toast } = useToast();
      const sessionId = params.id as string;

      const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);
      const [activeTab, setActiveTab] = useState<'mindmap' | 'timeline'>('mindmap');

      const { data: session, isLoading: isSessionLoading } = useSession(sessionId);
      const { data: mindmap, isLoading: isMindmapLoading } = useMindmap(sessionId);
      const deleteSession = useDeleteSession();

      const handleDelete = async () => {
        try {
          await deleteSession.mutateAsync(sessionId);
          toast({ title: '세션이 삭제되었습니다.' });
          router.push('/sessions');
        } catch (error) {
          toast({
            title: '삭제 실패',
            description: '세션을 삭제하는데 실패했습니다.',
            variant: 'destructive',
          });
        }
      };

      if (isSessionLoading) {
        return (
          <div className="space-y-4">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-[600px]" />
          </div>
        );
      }

      if (!session) {
        return (
          <div className="text-center py-8">
            <p className="text-red-500">세션을 찾을 수 없습니다.</p>
            <Button
              variant="outline"
              className="mt-4"
              onClick={() => router.push('/sessions')}
            >
              목록으로 돌아가기
            </Button>
          </div>
        );
      }

      return (
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => router.push('/sessions')}
              >
                <ArrowLeft className="h-5 w-5" />
              </Button>
              <div>
                <h1 className="text-2xl font-bold">
                  {session.title || '제목 없음'}
                </h1>
                <p className="text-sm text-gray-500">
                  {new Date(session.started_at).toLocaleString('ko-KR')}
                </p>
              </div>
            </div>
            <Button variant="destructive" size="icon" onClick={handleDelete}>
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>

          {/* Stats */}
          <SessionStats session={session} />

          {/* Main Content */}
          <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)}>
            <TabsList>
              <TabsTrigger value="mindmap" className="gap-2">
                <Map className="h-4 w-4" />
                마인드맵
              </TabsTrigger>
              <TabsTrigger value="timeline" className="gap-2">
                <List className="h-4 w-4" />
                타임라인
              </TabsTrigger>
            </TabsList>

            <TabsContent value="mindmap" className="mt-4">
              {isMindmapLoading ? (
                <Skeleton className="h-[600px] rounded-xl" />
              ) : mindmap ? (
                <div className="relative">
                  <MindmapCanvas>
                    <Galaxy data={mindmap} onNodeSelect={setSelectedNode} />
                  </MindmapCanvas>
                  <NodeDetailPanel
                    node={selectedNode}
                    onClose={() => setSelectedNode(null)}
                  />
                </div>
              ) : (
                <div className="h-[600px] bg-gray-100 rounded-xl flex items-center justify-center">
                  <div className="text-center">
                    <p className="text-gray-500">마인드맵이 아직 생성되지 않았습니다.</p>
                    <p className="text-sm text-gray-400 mt-1">
                      세션 처리가 완료되면 자동으로 생성됩니다.
                    </p>
                  </div>
                </div>
              )}
            </TabsContent>

            <TabsContent value="timeline" className="mt-4">
              <SessionTimeline session={session} />
            </TabsContent>
          </Tabs>
        </div>
      );
    }
    ```

- [ ] **세션 통계 컴포넌트**
  - [ ] `src/components/sessions/SessionStats.tsx`

    ```tsx
    import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
    import { Badge } from '@/components/ui/badge';
    import { SessionWithDetails } from '@/lib/api/sessions';

    interface SessionStatsProps {
      session: SessionWithDetails;
    }

    export function SessionStats({ session }: SessionStatsProps) {
      const totalDurationMs = session.page_visits.reduce(
        (acc, pv) => acc + (pv.duration_ms || 0),
        0
      );

      const formatDuration = (ms: number) => {
        const totalSeconds = Math.floor(ms / 1000);
        const hours = Math.floor(totalSeconds / 3600);
        const minutes = Math.floor((totalSeconds % 3600) / 60);

        if (hours > 0) {
          return `${hours}시간 ${minutes}분`;
        }
        return `${minutes}분`;
      };

      return (
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-gray-500">
                상태
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Badge variant={session.status === 'completed' ? 'default' : 'secondary'}>
                {session.status}
              </Badge>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-gray-500">
                방문한 페이지
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">{session.page_visits.length}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-gray-500">
                총 체류 시간
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">{formatDuration(totalDurationMs)}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-gray-500">
                하이라이트
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">{session.highlights.length}</p>
            </CardContent>
          </Card>
        </div>
      );
    }
    ```

- [ ] **세션 타임라인 컴포넌트**
  - [ ] `src/components/sessions/SessionTimeline.tsx`

    ```tsx
    import { formatDistanceToNow } from 'date-fns';
    import { ko } from 'date-fns/locale';
    import { ExternalLink, Clock, MessageSquare } from 'lucide-react';
    import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
    import { SessionWithDetails } from '@/lib/api/sessions';

    interface SessionTimelineProps {
      session: SessionWithDetails;
    }

    export function SessionTimeline({ session }: SessionTimelineProps) {
      const sortedVisits = [...session.page_visits].sort(
        (a, b) => new Date(a.entered_at).getTime() - new Date(b.entered_at).getTime()
      );

      return (
        <div className="space-y-4">
          {/* Page visits timeline */}
          <Card>
            <CardHeader>
              <CardTitle>방문 기록</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="relative">
                {/* Timeline line */}
                <div className="absolute left-4 top-0 bottom-0 w-px bg-gray-200" />

                <ul className="space-y-4">
                  {sortedVisits.map((visit, index) => (
                    <li key={visit.id} className="relative pl-10">
                      {/* Timeline dot */}
                      <div className="absolute left-2.5 top-1.5 w-3 h-3 rounded-full bg-blue-500 ring-4 ring-white" />

                      <div className="bg-gray-50 rounded-lg p-3">
                        <div className="flex items-start justify-between">
                          <div className="flex-1 min-w-0">
                            <p className="font-medium truncate">
                              {visit.url.title || '제목 없음'}
                            </p>
                            <a
                              href={visit.url.url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-sm text-blue-600 hover:underline truncate flex items-center gap-1"
                            >
                              {visit.url.url}
                              <ExternalLink className="h-3 w-3 flex-shrink-0" />
                            </a>
                          </div>
                        </div>

                        <div className="flex items-center gap-4 mt-2 text-sm text-gray-500">
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {visit.duration_ms
                              ? `${Math.floor(visit.duration_ms / 60000)}분 ${Math.floor((visit.duration_ms % 60000) / 1000)}초`
                              : '-'}
                          </span>
                          <span>
                            스크롤 {Math.round(visit.max_scroll_depth * 100)}%
                          </span>
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* Highlights */}
          {session.highlights.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <MessageSquare className="h-5 w-5" />
                  하이라이트
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2">
                  {session.highlights.map((highlight) => (
                    <li
                      key={highlight.id}
                      className="p-3 bg-yellow-50 rounded-lg border-l-4"
                      style={{ borderColor: highlight.color }}
                    >
                      <p className="text-sm">&ldquo;{highlight.text}&rdquo;</p>
                      <p className="text-xs text-gray-400 mt-1">
                        {formatDistanceToNow(new Date(highlight.created_at), {
                          addSuffix: true,
                          locale: ko,
                        })}
                      </p>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          )}
        </div>
      );
    }
    ```

### 검증

```bash
pnpm dev
# 세션 상세 페이지에서 마인드맵 및 타임라인 탭 전환 확인
```

---

## Step 11.4: 애니메이션 & 인터랙션

### 체크리스트

- [ ] **빅뱅 초기 애니메이션**
  - [ ] `src/components/mindmap/BigBangAnimation.tsx`

    ```tsx
    'use client';

    import { useRef, useEffect } from 'react';
    import { useFrame } from '@react-three/fiber';
    import { useSpring, animated } from '@react-spring/three';
    import * as THREE from 'three';

    interface BigBangAnimationProps {
      isPlaying: boolean;
      onComplete: () => void;
      children: React.ReactNode;
    }

    export function BigBangAnimation({
      isPlaying,
      onComplete,
      children,
    }: BigBangAnimationProps) {
      const groupRef = useRef<THREE.Group>(null);

      const { scale, opacity } = useSpring({
        from: { scale: 0.01, opacity: 0 },
        to: async (next) => {
          if (isPlaying) {
            await next({ scale: 1.2, opacity: 1, config: { duration: 800 } });
            await next({ scale: 1, opacity: 1, config: { duration: 300 } });
            onComplete();
          }
        },
        config: { tension: 200, friction: 20 },
      });

      return (
        <animated.group ref={groupRef} scale={scale}>
          {children}
        </animated.group>
      );
    }
    ```

- [ ] **궤도 애니메이션**
  - [ ] `src/components/mindmap/OrbitAnimation.tsx`

    ```tsx
    'use client';

    import { useRef } from 'react';
    import { useFrame } from '@react-three/fiber';
    import * as THREE from 'three';

    interface OrbitAnimationProps {
      children: React.ReactNode;
      radius: number;
      speed: number;
      startAngle: number;
      centerPosition: [number, number, number];
    }

    export function OrbitAnimation({
      children,
      radius,
      speed,
      startAngle,
      centerPosition,
    }: OrbitAnimationProps) {
      const groupRef = useRef<THREE.Group>(null);
      const angleRef = useRef(startAngle);

      useFrame((_, delta) => {
        if (groupRef.current) {
          angleRef.current += delta * speed;

          groupRef.current.position.x =
            centerPosition[0] + radius * Math.cos(angleRef.current);
          groupRef.current.position.y =
            centerPosition[1] + radius * Math.sin(angleRef.current) * 0.3; // Elliptical
          groupRef.current.position.z =
            centerPosition[2] + radius * Math.sin(angleRef.current);
        }
      });

      return <group ref={groupRef}>{children}</group>;
    }
    ```

- [ ] **파티클 효과**
  - [ ] `src/components/mindmap/ParticleField.tsx`

    ```tsx
    'use client';

    import { useRef, useMemo } from 'react';
    import { useFrame } from '@react-three/fiber';
    import * as THREE from 'three';

    interface ParticleFieldProps {
      count: number;
      radius: number;
    }

    export function ParticleField({ count, radius }: ParticleFieldProps) {
      const pointsRef = useRef<THREE.Points>(null);

      const positions = useMemo(() => {
        const positions = new Float32Array(count * 3);

        for (let i = 0; i < count; i++) {
          const theta = Math.random() * Math.PI * 2;
          const phi = Math.acos(2 * Math.random() - 1);
          const r = radius * (0.5 + Math.random() * 0.5);

          positions[i * 3] = r * Math.sin(phi) * Math.cos(theta);
          positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
          positions[i * 3 + 2] = r * Math.cos(phi);
        }

        return positions;
      }, [count, radius]);

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
              count={count}
              array={positions}
              itemSize={3}
            />
          </bufferGeometry>
          <pointsMaterial
            size={1}
            color="#4B5563"
            transparent
            opacity={0.5}
            sizeAttenuation
          />
        </points>
      );
    }
    ```

- [ ] **카메라 제어 개선**
  - [ ] `src/components/mindmap/CameraController.tsx`

    ```tsx
    'use client';

    import { useRef, useEffect } from 'react';
    import { useThree, useFrame } from '@react-three/fiber';
    import * as THREE from 'three';
    import { MindmapNode } from '@/types/mindmap';

    interface CameraControllerProps {
      focusNode: MindmapNode | null;
    }

    export function CameraController({ focusNode }: CameraControllerProps) {
      const { camera } = useThree();
      const targetPosition = useRef(new THREE.Vector3(0, 0, 500));
      const targetLookAt = useRef(new THREE.Vector3(0, 0, 0));

      useEffect(() => {
        if (focusNode?.position) {
          targetPosition.current.set(
            focusNode.position.x,
            focusNode.position.y + 50,
            focusNode.position.z + 200
          );
          targetLookAt.current.set(
            focusNode.position.x,
            focusNode.position.y,
            focusNode.position.z
          );
        } else {
          targetPosition.current.set(0, 0, 500);
          targetLookAt.current.set(0, 0, 0);
        }
      }, [focusNode]);

      useFrame((_, delta) => {
        // Smooth camera movement
        camera.position.lerp(targetPosition.current, delta * 2);

        // Smooth look-at
        const currentLookAt = new THREE.Vector3();
        camera.getWorldDirection(currentLookAt);
        currentLookAt.lerp(
          targetLookAt.current.clone().sub(camera.position).normalize(),
          delta * 2
        );
        camera.lookAt(
          camera.position.clone().add(currentLookAt)
        );
      });

      return null;
    }
    ```

- [ ] **최종 Galaxy 컴포넌트 업데이트**

  ```tsx
  // Galaxy.tsx에 애니메이션 추가
  import { useState } from 'react';
  import { BigBangAnimation } from './BigBangAnimation';
  import { ParticleField } from './ParticleField';
  import { CameraController } from './CameraController';

  export function Galaxy({ data, onNodeSelect }: GalaxyProps) {
    const [isAnimationComplete, setIsAnimationComplete] = useState(false);
    // ... 기존 코드

    return (
      <group>
        {/* Background particles */}
        <ParticleField count={500} radius={400} />

        {/* Camera controller */}
        <CameraController focusNode={selectedNodeId ? nodeMap.get(selectedNodeId) || null : null} />

        {/* Main content with animation */}
        <BigBangAnimation
          isPlaying={true}
          onComplete={() => setIsAnimationComplete(true)}
        >
          {/* Edges */}
          {data.edges.map((edge) => {
            // ... edge 렌더링
          })}

          {/* Nodes */}
          {data.nodes.map((node) => (
            // ... node 렌더링
          ))}
        </BigBangAnimation>
      </group>
    );
  }
  ```

### 검증

```bash
pnpm dev
# 3D 마인드맵 페이지에서:
# 1. 빅뱅 초기 애니메이션 확인
# 2. 노드 클릭 시 카메라 이동 확인
# 3. 호버 시 연결된 노드 하이라이트 확인
# 4. 배경 파티클 애니메이션 확인
```

---

## Step 11.5: 계정 & 사용량 페이지

### 목표

Phase 9에서 구현한 Subscription/Usage API를 프론트엔드에서 사용할 수 있도록 UI를 구현합니다.

### 체크리스트

- [ ] **API 래퍼 함수 생성**
  - [ ] `src/lib/api/subscription.ts`

    ```typescript
    import { client, handleApiResponse } from './client';
    import {
      subscriptionRoutesGetSubscription,
      subscriptionRoutesListPlans,
    } from '@/api/generated';
    import type {
      SubscriptionSubscriptionInfo,
      SubscriptionPlan,
    } from '@/api/generated';

    export type { SubscriptionSubscriptionInfo, SubscriptionPlan };

    export async function getSubscription(): Promise<SubscriptionSubscriptionInfo> {
      const response = await subscriptionRoutesGetSubscription({ client });
      return handleApiResponse(response).subscription;
    }

    export async function getPlans(): Promise<SubscriptionPlan[]> {
      const response = await subscriptionRoutesListPlans({ client });
      return handleApiResponse(response).plans;
    }
    ```

  - [ ] `src/lib/api/usage.ts`

    ```typescript
    import { client, handleApiResponse } from './client';
    import {
      usageRoutesGetUsage,
      usageRoutesGetUsageHistory,
    } from '@/api/generated';
    import type { UsageUsageSummary } from '@/api/generated';

    export type { UsageUsageSummary };

    export async function getUsage(): Promise<UsageUsageSummary> {
      const response = await usageRoutesGetUsage({ client });
      return handleApiResponse(response).usage;
    }

    export async function getUsageHistory(months?: number): Promise<UsageUsageSummary[]> {
      const response = await usageRoutesGetUsageHistory({
        client,
        query: { months },
      });
      return handleApiResponse(response).history;
    }
    ```

- [ ] **React Query Hooks 생성**
  - [ ] `src/lib/hooks/use-subscription.ts`

    ```typescript
    import { useQuery } from '@tanstack/react-query';
    import { getSubscription, getPlans } from '@/lib/api/subscription';

    export function useSubscription() {
      return useQuery({
        queryKey: ['subscription'],
        queryFn: getSubscription,
      });
    }

    export function usePlans() {
      return useQuery({
        queryKey: ['plans'],
        queryFn: getPlans,
      });
    }
    ```

  - [ ] `src/lib/hooks/use-usage.ts`

    ```typescript
    import { useQuery } from '@tanstack/react-query';
    import { getUsage, getUsageHistory } from '@/lib/api/usage';

    export function useUsage() {
      return useQuery({
        queryKey: ['usage'],
        queryFn: getUsage,
      });
    }

    export function useUsageHistory(months: number = 6) {
      return useQuery({
        queryKey: ['usage-history', months],
        queryFn: () => getUsageHistory(months),
      });
    }
    ```

- [ ] **계정 페이지 생성**
  - [ ] `src/app/(dashboard)/account/page.tsx`

    ```tsx
    'use client';

    import { useSubscription, usePlans } from '@/lib/hooks/use-subscription';
    import { useUsage, useUsageHistory } from '@/lib/hooks/use-usage';
    import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
    import { Progress } from '@/components/ui/progress';
    import { Badge } from '@/components/ui/badge';
    import { Skeleton } from '@/components/ui/skeleton';
    import { Button } from '@/components/ui/button';
    import { formatDate } from '@/lib/utils';

    export default function AccountPage() {
      const { data: subscription, isLoading: isSubLoading } = useSubscription();
      const { data: usage, isLoading: isUsageLoading } = useUsage();
      const { data: history } = useUsageHistory(6);

      if (isSubLoading || isUsageLoading) {
        return (
          <div className="space-y-6">
            <Skeleton className="h-8 w-48" />
            <div className="grid gap-6 md:grid-cols-2">
              <Skeleton className="h-48" />
              <Skeleton className="h-48" />
            </div>
          </div>
        );
      }

      const usagePercent = usage?.limit
        ? Math.round((usage.tokensUsed / usage.limit) * 100)
        : 0;

      return (
        <div className="space-y-6">
          <h1 className="text-2xl font-bold">계정</h1>

          <div className="grid gap-6 md:grid-cols-2">
            {/* 구독 정보 */}
            <Card>
              <CardHeader>
                <CardTitle>구독 정보</CardTitle>
                <CardDescription>현재 플랜 및 구독 상태</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-500">플랜</span>
                  <Badge variant="default">{subscription?.plan?.name || 'Free'}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-500">상태</span>
                  <Badge variant={subscription?.status === 'active' ? 'default' : 'secondary'}>
                    {subscription?.status || 'active'}
                  </Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-500">현재 기간</span>
                  <span className="text-sm">
                    {subscription?.currentPeriodStart && subscription?.currentPeriodEnd
                      ? `${formatDate(subscription.currentPeriodStart)} ~ ${formatDate(subscription.currentPeriodEnd)}`
                      : '-'}
                  </span>
                </div>
                {subscription?.plan?.priceCents === 0 && (
                  <Button className="w-full mt-4">Pro로 업그레이드</Button>
                )}
              </CardContent>
            </Card>

            {/* 사용량 */}
            <Card>
              <CardHeader>
                <CardTitle>토큰 사용량</CardTitle>
                <CardDescription>이번 달 AI 토큰 사용량</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>{usage?.tokensUsed?.toLocaleString() || 0} 토큰 사용</span>
                    <span className="text-gray-500">
                      {usage?.limit ? `${usage.limit.toLocaleString()} 제한` : '무제한'}
                    </span>
                  </div>
                  {usage?.limit && (
                    <Progress value={usagePercent} className="h-2" />
                  )}
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-500">기간</span>
                  <span className="text-sm">
                    {usage?.periodStart && usage?.periodEnd
                      ? `${formatDate(usage.periodStart)} ~ ${formatDate(usage.periodEnd)}`
                      : '-'}
                  </span>
                </div>
                {usagePercent >= 80 && (
                  <p className="text-sm text-orange-600">
                    ⚠️ 토큰 사용량이 80%를 초과했습니다.
                  </p>
                )}
              </CardContent>
            </Card>
          </div>

          {/* 사용량 히스토리 */}
          {history && history.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>사용량 히스토리</CardTitle>
                <CardDescription>최근 6개월 토큰 사용량</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {history.map((item, index) => (
                    <div key={index} className="flex items-center justify-between py-2 border-b last:border-0">
                      <span className="text-sm">
                        {formatDate(item.periodStart)} ~ {formatDate(item.periodEnd)}
                      </span>
                      <span className="text-sm font-medium">
                        {item.tokensUsed.toLocaleString()} 토큰
                      </span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      );
    }
    ```

- [ ] **사이드바에 계정 링크 추가**
  - [ ] `src/components/layout/Sidebar.tsx` 업데이트

    ```tsx
    // 사이드바 메뉴에 추가
    { href: '/account', icon: User, label: '계정' }
    ```

- [ ] **헤더에 사용량 표시 (옵션)**
  - [ ] `src/components/layout/Header.tsx` 업데이트

    ```tsx
    // 헤더 우측에 사용량 요약 표시
    import { useUsage } from '@/lib/hooks/use-usage';

    const { data: usage } = useUsage();
    const usagePercent = usage?.limit
      ? Math.round((usage.tokensUsed / usage.limit) * 100)
      : 0;

    // 80% 이상일 경우 경고 배지 표시
    ```

### 검증

```bash
pnpm dev
# http://localhost:3000/account 접속
# - 구독 정보 카드 표시 확인
# - 토큰 사용량 Progress 바 표시 확인
# - 사용량 히스토리 목록 표시 확인
```

---

## Phase 11 완료 확인

### 전체 검증 체크리스트

- [ ] React Three Fiber 렌더링
- [ ] 3D 노드 표시 (core, topic, subtopic)
- [ ] 노드 간 연결선 표시
- [ ] 노드 호버/클릭 인터랙션
- [ ] 노드 상세 패널
- [ ] 카메라 컨트롤 (회전, 줌)
- [ ] 빅뱅 초기 애니메이션
- [ ] 배경 별/파티클 효과
- [ ] 세션 통계 표시
- [ ] 타임라인 뷰
- [ ] 계정 페이지 (구독/사용량 UI)

### 테스트 요구사항

| 테스트 유형 | 대상 | 도구 |
| ----------- | ---- | ---- |
| 컴포넌트 테스트 | 마인드맵 노드/엣지 | Vitest + React Testing Library |
| 스냅샷 테스트 | 3D 렌더링 결과 | Vitest |
| E2E 테스트 | 마인드맵 인터랙션 | Playwright |

```bash
# Phase 11 테스트 실행
moonx web:test
moonx web:e2e
```

> **Note**: 3D 컴포넌트는 WebGL 컨텍스트가 필요하므로 CI에서는 headless 브라우저로 E2E 테스트를 실행합니다.

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| 마인드맵 Canvas | `src/components/mindmap/MindmapCanvas.tsx` |
| Galaxy 컴포넌트 | `src/components/mindmap/Galaxy.tsx` |
| Node 컴포넌트 | `src/components/mindmap/Node.tsx` |
| Edge 컴포넌트 | `src/components/mindmap/Edge.tsx` |
| 애니메이션 | `src/components/mindmap/BigBangAnimation.tsx` |
| 세션 상세 페이지 | `src/app/(dashboard)/sessions/[id]/page.tsx` |
| 테스트 | `src/components/mindmap/**/*.test.tsx` |

---

## 다음 Phase

Phase 11 완료 후 [Phase 12: 프로덕션 모니터링](./phase-12-monitoring.md)으로 진행하세요.
