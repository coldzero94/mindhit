'use client';

import { Suspense, ReactNode, useRef, useEffect, useState, type ElementRef } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, Stars, PerspectiveCamera } from '@react-three/drei';
import { EffectComposer, Bloom } from '@react-three/postprocessing';
import { RotateCcw, Move3d } from 'lucide-react';

interface MindmapCanvasProps {
  children: ReactNode;
  className?: string;
  onResetView?: () => void;
}

function LoadingFallback() {
  return (
    <mesh>
      <sphereGeometry args={[1, 16, 16]} />
      <meshBasicMaterial color="#4B5563" wireframe />
    </mesh>
  );
}

interface ControlsProps {
  freeRotation: boolean;
  onReset: () => void;
}

function Controls({ freeRotation, onReset }: ControlsProps) {
  const controlsRef = useRef<ElementRef<typeof OrbitControls>>(null);

  // Expose reset function
  useEffect(() => {
    if (controlsRef.current) {
      // Store the reset function
      (window as unknown as { __mindmapControlsReset?: () => void }).__mindmapControlsReset = () => {
        controlsRef.current?.reset();
        onReset();
      };
    }
    return () => {
      delete (window as unknown as { __mindmapControlsReset?: () => void }).__mindmapControlsReset;
    };
  }, [onReset]);

  // Turntable-style: Limit vertical rotation to prevent flipping
  // When freeRotation is enabled, allow full rotation
  const minPolarAngle = freeRotation ? 0 : Math.PI * 0.2;  // ~36 degrees from top
  const maxPolarAngle = freeRotation ? Math.PI : Math.PI * 0.8;  // ~144 degrees

  return (
    <OrbitControls
      ref={controlsRef}
      makeDefault
      enablePan={true}
      enableZoom={true}
      enableRotate={true}
      minDistance={50}
      maxDistance={1500}
      // Smooth damping for natural feel
      enableDamping={true}
      dampingFactor={0.05}
      // Pan settings - primary navigation method
      screenSpacePanning={true}
      panSpeed={1.0}
      // Zoom settings
      zoomSpeed={1.0}
      // Rotation settings - slower for precision
      rotateSpeed={0.4}
      // Turntable-style rotation limits (unless freeRotation enabled)
      minPolarAngle={minPolarAngle}
      maxPolarAngle={maxPolarAngle}
      // Mouse buttons: Left=Pan, Middle=Zoom, Right=Rotate
      mouseButtons={{
        LEFT: 2,   // PAN - primary interaction
        MIDDLE: 1, // DOLLY (zoom)
        RIGHT: 0,  // ROTATE - secondary, with limits
      }}
      // Touch settings
      touches={{
        ONE: 2,  // PAN
        TWO: 1,  // DOLLY_PAN
      }}
    />
  );
}

export function MindmapCanvas({ children, className }: MindmapCanvasProps) {
  const [freeRotation, setFreeRotation] = useState(false);

  const handleReset = () => {
    const resetFn = (window as unknown as { __mindmapControlsReset?: () => void }).__mindmapControlsReset;
    if (resetFn) resetFn();
  };

  return (
    <div className={`relative w-full bg-gray-900 rounded-xl overflow-hidden ${className || 'h-[600px]'}`}>
      {/* Control buttons - top right */}
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        {/* Free rotation toggle */}
        <button
          onClick={() => setFreeRotation(!freeRotation)}
          className={`p-2 rounded-lg backdrop-blur transition-all ${
            freeRotation
              ? 'bg-blue-500/80 text-white'
              : 'bg-gray-900/80 text-gray-400 hover:text-gray-200'
          }`}
          title={freeRotation ? 'Free Rotation ON' : 'Free Rotation OFF'}
        >
          <Move3d className="w-4 h-4" />
        </button>
        {/* Reset view button */}
        <button
          onClick={handleReset}
          className="p-2 rounded-lg bg-gray-900/80 text-gray-400 hover:text-gray-200 backdrop-blur transition-colors"
          title="Reset View"
        >
          <RotateCcw className="w-4 h-4" />
        </button>
      </div>

      {/* Help tooltip - bottom left */}
      <div className="absolute bottom-4 left-4 z-10 text-xs text-gray-400 bg-gray-900/80 px-3 py-2 rounded-lg backdrop-blur">
        <div className="space-y-0.5">
          <div><span className="text-gray-300">Drag</span>: Pan</div>
          <div><span className="text-gray-300">Scroll</span>: Zoom</div>
          <div><span className="text-gray-300">Right-click</span>: Rotate{!freeRotation && <span className="text-gray-500"> (limited)</span>}</div>
          <div><span className="text-gray-300">Click node</span>: Details</div>
        </div>
      </div>

      {/* Free rotation indicator */}
      {freeRotation && (
        <div className="absolute bottom-4 right-4 z-10 text-xs text-blue-400 bg-gray-900/80 px-2 py-1 rounded backdrop-blur">
          Free Rotation Mode
        </div>
      )}

      <Canvas>
        {/* Camera */}
        <PerspectiveCamera makeDefault position={[0, 0, 400]} fov={60} near={1} far={5000} />

        {/* Lighting */}
        <ambientLight intensity={0.4} />
        <pointLight position={[200, 200, 200]} intensity={1.2} />
        <pointLight position={[-200, -200, -200]} intensity={0.6} />
        <pointLight position={[0, 300, 0]} intensity={0.4} color="#6366f1" />

        {/* Background Stars */}
        <Stars
          radius={400}
          depth={100}
          count={3000}
          factor={4}
          saturation={0}
          fade
          speed={0.3}
        />

        {/* Camera Controls */}
        <Controls freeRotation={freeRotation} onReset={() => {}} />

        {/* Post-processing Effects */}
        <EffectComposer>
          <Bloom
            luminanceThreshold={0.15}
            luminanceSmoothing={0.9}
            intensity={0.6}
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
