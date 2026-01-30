"use client";

import { ReactNode } from "react";
import { motion } from "framer-motion";
import { useReducedMotion } from "@/lib/hooks/use-reduced-motion";
import { cn } from "@/lib/utils";

interface FadeInUpProps {
  children: ReactNode;
  className?: string;
  /**
   * Delay before animation starts (in seconds)
   * @default 0
   */
  delay?: number;
  /**
   * Animation duration (in seconds)
   * @default 0.8
   */
  duration?: number;
  /**
   * Distance to travel (in pixels)
   * @default 50
   */
  distance?: number;
}

/**
 * FadeInUp - Simple fade-in from bottom animation
 *
 * Triggers when element enters viewport (whileInView).
 * Useful for non-sticky content like footer, pricing cards, etc.
 *
 * Features:
 * - Fades in + slides up
 * - Triggers once when entering viewport
 * - Respects prefers-reduced-motion
 *
 * @example
 * <FadeInUp delay={0.2}>
 *   <h2>Your Content</h2>
 * </FadeInUp>
 */
export function FadeInUp({
  children,
  className,
  delay = 0,
  duration = 0.8,
  distance = 50,
}: FadeInUpProps) {
  const prefersReducedMotion = useReducedMotion();

  // If user prefers reduced motion, skip animation
  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>;
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: distance }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.3 }}
      transition={{
        duration,
        delay,
        ease: [0.25, 0.1, 0.25, 1], // Smooth cubic-bezier easing
      }}
      className={cn("will-animate", className)}
    >
      {children}
    </motion.div>
  );
}
