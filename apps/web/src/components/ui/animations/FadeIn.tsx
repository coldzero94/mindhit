"use client";

import { ReactNode } from "react";
import { motion } from "framer-motion";
import { useReducedMotion } from "@/lib/hooks/use-reduced-motion";
import { cn } from "@/lib/utils";

interface FadeInProps {
  children: ReactNode;
  className?: string;
  /**
   * Delay before animation starts (in seconds)
   * @default 0
   */
  delay?: number;
  /**
   * Animation duration (in seconds)
   * @default 0.5
   */
  duration?: number;
}

/**
 * FadeIn - Simple fade-in animation
 *
 * Triggers immediately on mount (not scroll-triggered).
 * Useful for dashboard content, modals, page sections.
 *
 * Features:
 * - Simple opacity fade from 0 to 1
 * - Respects prefers-reduced-motion
 * - Lightweight alternative to FadeInUp when no slide needed
 *
 * @example
 * <FadeIn delay={0.1}>
 *   <Card>...</Card>
 * </FadeIn>
 */
export function FadeIn({
  children,
  className,
  delay = 0,
  duration = 0.5,
}: FadeInProps) {
  const prefersReducedMotion = useReducedMotion();

  // If user prefers reduced motion, skip animation
  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>;
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{
        duration,
        delay,
        ease: "easeOut",
      }}
      className={cn("will-animate", className)}
    >
      {children}
    </motion.div>
  );
}
