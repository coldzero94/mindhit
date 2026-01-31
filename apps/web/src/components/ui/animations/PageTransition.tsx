"use client";

import { ReactNode } from "react";
import { motion } from "framer-motion";
import { useReducedMotion } from "@/lib/hooks/use-reduced-motion";
import { cn } from "@/lib/utils";

interface PageTransitionProps {
  children: ReactNode;
  className?: string;
  /**
   * Animation duration (in seconds)
   * @default 0.3
   */
  duration?: number;
}

/**
 * PageTransition - Wrapper for page-level entry animations
 *
 * Provides a subtle fade + slide animation for page content.
 * Wrap the entire page content to give smooth entry feel.
 *
 * Features:
 * - Subtle opacity + y-axis animation
 * - Quick duration for snappy feel
 * - Respects prefers-reduced-motion
 *
 * @example
 * // In page.tsx
 * export default function MyPage() {
 *   return (
 *     <PageTransition>
 *       <div className="container">
 *         Page content here...
 *       </div>
 *     </PageTransition>
 *   );
 * }
 */
export function PageTransition({
  children,
  className,
  duration = 0.3,
}: PageTransitionProps) {
  const prefersReducedMotion = useReducedMotion();

  // If user prefers reduced motion, skip animation
  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>;
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration,
        ease: "easeOut",
      }}
      className={cn("will-animate", className)}
    >
      {children}
    </motion.div>
  );
}
