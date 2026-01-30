"use client";

import { ReactNode } from "react";
import { motion, useTransform } from "framer-motion";
import { useScrollProgress } from "@/lib/hooks/use-scroll-progress";
import { cn } from "@/lib/utils";

interface FullPageSectionProps {
  children: ReactNode;
  className?: string;
  /**
   * Enable fade-out effect as section scrolls away
   * @default true
   */
  enableFadeOut?: boolean;
  /**
   * Enable scale-down effect as section scrolls away
   * @default false
   */
  enableScale?: boolean;
  /**
   * Z-index for stacking order (higher values appear on top)
   * Each subsequent section should have a higher z-index
   * @default 10
   */
  zIndex?: number;
}

/**
 * FullPageSection - iPhone-style full-page sticky section wrapper
 *
 * Creates a full-viewport-height section that sticks to the top as you scroll,
 * then fades out and gets covered by the next section (iPhone product page style).
 *
 * Features:
 * - Exact 100vh height
 * - position: sticky with top: 0
 * - Scroll-linked fade-out animation
 * - Optional scale-down animation for depth effect
 * - Next section slides up from bottom to cover this section
 *
 * @example
 * <FullPageSection className="section-light">
 *   <div className="container mx-auto">
 *     <h1>Your Content</h1>
 *   </div>
 * </FullPageSection>
 */
export function FullPageSection({
  children,
  className,
  enableFadeOut = true,
  enableScale = false,
  zIndex = 10,
}: FullPageSectionProps) {
  const { ref, scrollYProgress } = useScrollProgress(["start start", "end start"]);

  // Map scroll progress to opacity (fade out as section exits)
  // When scrollYProgress = 0 (section at top), opacity = 1
  // When scrollYProgress = 0.8+ (section leaving), opacity starts to fade
  const opacity = useTransform(
    scrollYProgress,
    [0, 0.7, 1],
    [1, 1, enableFadeOut ? 0.3 : 1]
  );

  // Optional: Map scroll progress to scale (slight shrink for depth)
  const scale = useTransform(
    scrollYProgress,
    [0, 0.7, 1],
    [1, 1, enableScale ? 0.95 : 1]
  );

  return (
    <motion.section
      ref={ref}
      style={{
        opacity: enableFadeOut ? opacity : 1,
        scale: enableScale ? scale : 1,
        zIndex,
      }}
      className={cn(
        // Full-page sticky positioning - EXACT height
        "sticky top-0 h-screen",
        // Flexbox for centering content
        "flex items-center justify-center",
        // Ensure section doesn't overflow
        "overflow-hidden",
        // Default background (override with className)
        "bg-background",
        // Custom classes
        className
      )}
    >
      {children}
    </motion.section>
  );
}
