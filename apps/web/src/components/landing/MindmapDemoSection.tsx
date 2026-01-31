"use client";

import { lazy, Suspense, useState, useEffect, useRef } from "react";
import { Loader2 } from "lucide-react";
import { demoMindmapData } from "@/lib/data/demo-mindmap";

// Lazy-load heavy 3D components
const MindmapCanvas = lazy(() =>
  import("@/components/mindmap/MindmapCanvas").then((mod) => ({
    default: mod.MindmapCanvas,
  }))
);

const Galaxy = lazy(() =>
  import("@/components/mindmap/Galaxy").then((mod) => ({
    default: mod.Galaxy,
  }))
);

/**
 * MindmapDemoSection - 3D mindmap demo
 *
 * Features:
 * - Dark background for 3D to pop
 * - Lazy-loaded 3D components with IntersectionObserver
 * - Full-screen canvas with minimal UI overlay
 * - Regular flow section (not sticky) to avoid overlapping with pricing
 */
export function MindmapDemoSection() {
  const [shouldLoad, setShouldLoad] = useState(false);
  const sectionRef = useRef<HTMLDivElement>(null);

  // Use IntersectionObserver to load 3D only when section is visible
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setShouldLoad(true);
        }
      },
      {
        threshold: 0.1, // Load when 10% of section is visible
      }
    );

    const currentRef = sectionRef.current;
    if (currentRef) {
      observer.observe(currentRef);
    }

    return () => {
      if (currentRef) {
        observer.unobserve(currentRef);
      }
    };
  }, []);

  return (
    <section ref={sectionRef} className="relative w-full min-h-screen section-dark" style={{ zIndex: 70 }}>
      <div className="relative w-full h-screen">
          {/* Title Overlay */}
          <div className="absolute top-8 left-8 z-10">
            <h3 className="text-2xl sm:text-3xl font-bold text-white mb-2">
              Interactive 3D Mindmap
            </h3>
            <p className="text-sm sm:text-base text-gray-400">
              Rotate with mouse, click nodes for details
            </p>
          </div>

          {/* 3D Canvas */}
          <div className="w-full h-full">
            {shouldLoad ? (
              <Suspense
                fallback={
                  <div className="flex items-center justify-center h-full">
                    <div className="text-center">
                      <Loader2 className="w-12 h-12 animate-spin text-white mx-auto mb-4" />
                      <p className="text-white text-sm">Loading 3D Mindmap...</p>
                    </div>
                  </div>
                }
              >
                <MindmapCanvas className="h-full">
                  <Galaxy
                    data={demoMindmapData}
                    onNodeSelect={() => {}} // No-op for demo
                    enableAutoRotate={true}
                    enableAnimation={true}
                  />
                </MindmapCanvas>
              </Suspense>
            ) : (
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <div className="w-32 h-32 mx-auto mb-4 bg-gradient-to-br from-purple-500/20 to-blue-500/20 rounded-full flex items-center justify-center">
                    <div className="w-24 h-24 bg-gradient-to-br from-purple-500/30 to-blue-500/30 rounded-full animate-pulse" />
                  </div>
                  <p className="text-white text-sm">Scroll to view 3D Mindmap</p>
                </div>
              </div>
            )}
          </div>

          {/* Bottom Hint */}
          <div className="absolute bottom-8 left-1/2 -translate-x-1/2 text-center text-gray-400 text-xs sm:text-sm">
            <p>Mindmap example generated from real browsing data</p>
          </div>
        </div>
    </section>
  );
}
