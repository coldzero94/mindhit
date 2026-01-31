# Phase 11.1: React Three Fiber 환경 설정

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | React Three Fiber 및 3D 렌더링 환경 구성 |
| **선행 조건** | Phase 7 완료 |
| **예상 소요** | 1 Step |
| **결과물** | 3D Canvas 렌더링이 가능한 웹앱 환경 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 11.1.1 | 의존성 및 기본 설정 | ✅ |

---

## Step 11.1.1: 의존성 및 기본 설정

### 체크리스트

- [x] **의존성 설치**

  ```bash
  cd apps/web
  pnpm add three @react-three/fiber @react-three/drei @react-three/postprocessing
  pnpm add -D @types/three
  pnpm add framer-motion @react-spring/three
  ```

- [x] **Three.js 타입 설정**
  - [x] `src/types/three.d.ts`

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

- [x] **마인드맵 타입 정의**
  - [x] `src/types/mindmap.ts`

    ```typescript
    export interface MindmapNodePosition {
      x: number;
      y: number;
      z: number;
    }

    export type MindmapNodeType = 'core' | 'topic' | 'subtopic' | 'page';

    export interface MindmapNode {
      id: string;
      label: string;
      type: MindmapNodeType;
      size: number;
      color: string;
      position?: MindmapNodePosition;
      data: {
        description?: string;
        urls?: string[];
        visitCount?: number;
        totalDuration?: number;
        [key: string]: unknown;
      };
    }

    export interface MindmapEdge {
      source: string;
      target: string;
      weight: number;
    }

    export type MindmapLayoutType = 'galaxy' | 'tree' | 'radial';

    export interface MindmapLayout {
      type: MindmapLayoutType;
      params: Record<string, unknown>;
    }

    export interface MindmapData {
      nodes: MindmapNode[];
      edges: MindmapEdge[];
      layout: MindmapLayout;
    }

    // API Response 타입
    export interface MindmapResponse {
      mindmap: MindmapData | null;
      session_id: string;
      generated_at: string | null;
    }
    ```

- [x] **Canvas Provider 설정**
  - [x] `src/components/mindmap/MindmapCanvas.tsx`

    ```tsx
    'use client';

    import { Suspense, ReactNode } from 'react';
    import { Canvas } from '@react-three/fiber';
    import { OrbitControls, Stars, PerspectiveCamera } from '@react-three/drei';
    import { EffectComposer, Bloom } from '@react-three/postprocessing';

    interface MindmapCanvasProps {
      children: ReactNode;
      className?: string;
    }

    function LoadingFallback() {
      return (
        <mesh>
          <sphereGeometry args={[1, 16, 16]} />
          <meshBasicMaterial color="#4B5563" wireframe />
        </mesh>
      );
    }

    export function MindmapCanvas({ children, className }: MindmapCanvasProps) {
      return (
        <div className={`w-full h-full min-h-[600px] bg-gray-900 rounded-xl overflow-hidden ${className || ''}`}>
          <Canvas>
            {/* Camera */}
            <PerspectiveCamera makeDefault position={[0, 0, 500]} fov={60} />

            {/* Lighting */}
            <ambientLight intensity={0.3} />
            <pointLight position={[100, 100, 100]} intensity={1} />
            <pointLight position={[-100, -100, -100]} intensity={0.5} />

            {/* Background Stars */}
            <Stars
              radius={300}
              depth={50}
              count={5000}
              factor={4}
              saturation={0}
              fade
              speed={0.5}
            />

            {/* Camera Controls */}
            <OrbitControls
              enablePan={true}
              enableZoom={true}
              enableRotate={true}
              minDistance={100}
              maxDistance={1000}
              dampingFactor={0.05}
              enableDamping
            />

            {/* Post-processing Effects */}
            <EffectComposer>
              <Bloom
                luminanceThreshold={0.2}
                luminanceSmoothing={0.9}
                intensity={0.5}
              />
            </EffectComposer>

            {/* Content */}
            <Suspense fallback={<LoadingFallback />}>
              {children}
            </Suspense>
          </Canvas>
        </div>
      );
    }
    ```

- [x] **테스트 컴포넌트**
  - [x] `src/components/mindmap/TestSphere.tsx`

    ```tsx
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
    ```

- [x] **테스트 페이지**
  - [x] `src/app/(dashboard)/test-3d/page.tsx`

    ```tsx
    'use client';

    import { MindmapCanvas } from '@/components/mindmap/MindmapCanvas';
    import { TestSphere } from '@/components/mindmap/TestSphere';

    export default function Test3DPage() {
      return (
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">3D 렌더링 테스트</h1>
            <p className="text-gray-500 mt-1">
              React Three Fiber 환경 설정 확인용 페이지입니다.
            </p>
          </div>

          <div className="bg-white rounded-xl shadow-sm p-4">
            <h2 className="text-lg font-medium mb-4">테스트 Canvas</h2>
            <MindmapCanvas className="h-[500px]">
              <TestSphere />
            </MindmapCanvas>
          </div>

          <div className="bg-white rounded-xl shadow-sm p-4">
            <h2 className="text-lg font-medium mb-2">조작 방법</h2>
            <ul className="text-sm text-gray-600 space-y-1">
              <li>• 마우스 드래그: 회전</li>
              <li>• 스크롤: 줌</li>
              <li>• 우클릭 드래그: 이동 (Pan)</li>
              <li>• 구체 호버: 색상 변경</li>
            </ul>
          </div>
        </div>
      );
    }
    ```

- [x] **Next.js 설정 확인**
  - [x] `next.config.ts` 에서 Three.js 최적화 확인 (Turbopack 설정 추가)

    ```typescript
    // next.config.ts에 추가 (필요한 경우)
    const nextConfig = {
      // ... 기존 설정
      transpilePackages: ['three'],
      webpack: (config) => {
        config.externals = [...(config.externals || []), { canvas: 'canvas' }];
        return config;
      },
    };
    ```

### 검증

```bash
cd apps/web
pnpm dev

# http://localhost:3000/test-3d 접속 후:
# 1. 별 배경이 있는 3D Canvas 렌더링 확인
# 2. 파란 구체가 회전하는지 확인
# 3. 마우스로 회전/줌 컨트롤 확인
# 4. 구체 호버 시 색상 변경 확인
# 5. "MindHit 3D" 텍스트 표시 확인
```

```bash
# 빌드 테스트
pnpm build
# 빌드 에러 없이 완료 확인
```

---

## Phase 11.1 완료 확인

### 전체 검증 체크리스트

- [x] Three.js 의존성 설치 완료
- [x] 타입 정의 파일 생성 (`three.d.ts`, `mindmap.ts`)
- [x] MindmapCanvas 컴포넌트 동작
- [x] OrbitControls 동작 (회전, 줌, 이동)
- [x] Stars 배경 렌더링
- [x] Bloom post-processing 효과
- [x] 테스트 페이지 정상 동작
- [x] 프로덕션 빌드 성공

### 테스트

```bash
moonx web:typecheck
moonx web:lint
moonx web:build
```

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| Three.js 타입 | `src/types/three.d.ts` |
| 마인드맵 타입 | `src/types/mindmap.ts` |
| Canvas Provider | `src/components/mindmap/MindmapCanvas.tsx` |
| 테스트 컴포넌트 | `src/components/mindmap/TestSphere.tsx` |
| 테스트 페이지 | `src/app/(dashboard)/test-3d/page.tsx` |

---

## 다음 Phase

Phase 11.1 완료 후 [Phase 11.2: 3D 마인드맵 컴포넌트](./phase-11.2-mindmap-components.md)로 진행하세요.
