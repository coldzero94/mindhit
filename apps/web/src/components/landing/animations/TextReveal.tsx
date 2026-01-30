"use client";

import React from "react";
import { motion } from "framer-motion";
import { useReducedMotion } from "@/lib/hooks/use-reduced-motion";

interface TextRevealProps {
  text: string;
  className?: string;
  /**
   * Delay between each word (in seconds)
   * @default 0.08
   */
  delay?: number;
  /**
   * Animation duration per word (in seconds)
   * @default 0.5
   */
  duration?: number;
  /**
   * Tag to render (h1, h2, p, etc.)
   * @default "span"
   */
  as?: React.ElementType;
}

/**
 * TextReveal - Word-by-word stagger animation
 *
 * Splits text by spaces and animates each word with a stagger effect.
 * Perfect for hero headlines with dramatic reveals.
 *
 * Features:
 * - Word-by-word fade-in
 * - Stagger delay between words
 * - Respects prefers-reduced-motion
 *
 * @example
 * <TextReveal
 *   text="브라우징을 인사이트로 변환하세요"
 *   as="h1"
 *   className="text-5xl font-bold"
 * />
 */
export function TextReveal({
  text,
  className,
  delay = 0.08,
  duration = 0.5,
  as: Component = "span",
}: TextRevealProps) {
  const prefersReducedMotion = useReducedMotion();
  const words = text.split(" ");

  // If user prefers reduced motion, skip animation
  if (prefersReducedMotion) {
    return React.createElement(Component, { className }, text);
  }

  return React.createElement(
    Component,
    { className },
    words.map((word, index) => (
      <motion.span
        key={`${word}-${index}`}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{
          duration,
          delay: index * delay,
          ease: [0.25, 0.1, 0.25, 1],
        }}
        className="inline-block mr-[0.25em]"
      >
        {word}
      </motion.span>
    ))
  );
}
