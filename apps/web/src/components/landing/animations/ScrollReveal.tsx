"use client";

import { ReactNode } from "react";
import { motion, useTransform } from "framer-motion";
import { useScrollProgress } from "@/lib/hooks/use-scroll-progress";
import { cn } from "@/lib/utils";

interface ScrollRevealProps {
  children: ReactNode;
  className?: string;
  /**
   * Scale range for the animation [start, middle, end]
   * @default [0.8, 1, 0.9] - scales up on enter, slightly down on exit
   */
  scaleRange?: [number, number, number];
  /**
   * Rotation range in degrees [start, middle, end]
   * @default [0, 0, 0] - no rotation
   */
  rotateRange?: [number, number, number];
  /**
   * Y-axis translation range in pixels [start, middle, end]
   * @default [0, 0, 0] - no translation
   */
  yRange?: [number, number, number];
  /**
   * Opacity range [start, middle, end]
   * @default [1, 1, 1] - always visible
   */
  opacityRange?: [number, number, number];
}

/**
 * ScrollReveal - Center-stage element with scroll-linked transforms
 *
 * Animates a center element (product image, 3D visual, etc.) based on scroll position.
 * Perfect for iPhone-style product page reveals where the visual transforms as you scroll.
 *
 * Features:
 * - Scroll-linked scale, rotate, y-translate, opacity
 * - Maps scroll progress (0 = entering, 0.5 = middle, 1 = leaving) to transform values
 * - Smooth, continuous animation (not triggered, but linked to scroll)
 *
 * @example
 * // Scale up on enter, rotate slightly on exit
 * <ScrollReveal
 *   scaleRange={[0.8, 1, 0.9]}
 *   rotateRange={[0, 0, 5]}
 * >
 *   <img src="/product.png" alt="Product" />
 * </ScrollReveal>
 */
export function ScrollReveal({
  children,
  className,
  scaleRange = [0.8, 1, 0.9],
  rotateRange = [0, 0, 0],
  yRange = [0, 0, 0],
  opacityRange = [1, 1, 1],
}: ScrollRevealProps) {
  const { ref, scrollYProgress } = useScrollProgress(["start end", "end start"]);

  // Map scroll progress to transform values
  // scrollYProgress: 0 = entering viewport, 0.5 = center, 1 = leaving viewport
  const scale = useTransform(scrollYProgress, [0, 0.5, 1], scaleRange);
  const rotate = useTransform(scrollYProgress, [0, 0.5, 1], rotateRange);
  const y = useTransform(scrollYProgress, [0, 0.5, 1], yRange);
  const opacity = useTransform(scrollYProgress, [0, 0.5, 1], opacityRange);

  return (
    <motion.div
      ref={ref}
      style={{
        scale,
        rotate,
        y,
        opacity,
      }}
      className={cn("will-animate", className)}
    >
      {children}
    </motion.div>
  );
}
