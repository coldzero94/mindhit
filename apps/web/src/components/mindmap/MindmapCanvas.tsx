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
    <div className={`w-full bg-gray-900 rounded-xl overflow-hidden ${className || 'h-[600px]'}`}>
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
