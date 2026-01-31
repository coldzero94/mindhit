"use client";

import { ReactNode } from "react";
import { motion, Variants } from "framer-motion";
import { useReducedMotion } from "@/lib/hooks/use-reduced-motion";
import { cn } from "@/lib/utils";

interface StaggerChildrenProps {
  children: ReactNode;
  className?: string;
  /**
   * Delay between each child animation (in seconds)
   * @default 0.1
   */
  staggerDelay?: number;
  /**
   * Duration of each child animation (in seconds)
   * @default 0.5
   */
  duration?: number;
}

/**
 * Animation variants for staggered children
 * Use with motion.div and variants={staggerItem}
 */
export const staggerItem: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.5,
      ease: [0.25, 0.1, 0.25, 1],
    },
  },
};

/**
 * StaggerChildren - Container for sequential child animations
 *
 * Animates children one after another with a stagger delay.
 * Children must use motion.div with variants={staggerItem}.
 *
 * Features:
 * - Sequential reveal of children
 * - Customizable stagger delay
 * - Respects prefers-reduced-motion
 *
 * @example
 * <StaggerChildren staggerDelay={0.1}>
 *   {items.map(item => (
 *     <motion.div key={item.id} variants={staggerItem}>
 *       <Card>{item.content}</Card>
 *     </motion.div>
 *   ))}
 * </StaggerChildren>
 */
export function StaggerChildren({
  children,
  className,
  staggerDelay = 0.1,
}: StaggerChildrenProps) {
  const prefersReducedMotion = useReducedMotion();

  // If user prefers reduced motion, skip animation
  if (prefersReducedMotion) {
    return <div className={className}>{children}</div>;
  }

  const containerVariants: Variants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: staggerDelay,
        delayChildren: 0.1,
      },
    },
  };

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={containerVariants}
      className={cn("will-animate", className)}
    >
      {children}
    </motion.div>
  );
}
