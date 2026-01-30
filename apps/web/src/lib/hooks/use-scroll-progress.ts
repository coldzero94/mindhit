import { useRef } from "react";
import { useScroll, type MotionValue } from "framer-motion";

/**
 * Custom hook for tracking scroll progress of a section
 * Returns a ref to attach to the section and the scroll progress (0-1)
 *
 * @param offset - Scroll offset configuration [start, end]
 *   - "start end" = element enters at bottom of viewport
 *   - "end start" = element exits at top of viewport
 * @returns { ref, scrollYProgress } - Ref for the section and scroll progress value
 *
 * @example
 * const { ref, scrollYProgress } = useScrollProgress();
 * const opacity = useTransform(scrollYProgress, [0, 1], [0, 1]);
 */
export function useScrollProgress(
  offset: readonly [string, string] = ["start end", "end start"]
): { ref: React.RefObject<HTMLDivElement | null>; scrollYProgress: MotionValue<number> } {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: offset as ["start end", "end start"],
  });

  return { ref, scrollYProgress };
}
