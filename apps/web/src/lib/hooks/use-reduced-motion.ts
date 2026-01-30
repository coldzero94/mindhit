import { useSyncExternalStore } from "react";

/**
 * Custom hook to check if user prefers reduced motion
 * Respects the prefers-reduced-motion CSS media query for accessibility
 *
 * @returns boolean - true if user prefers reduced motion
 *
 * @example
 * const prefersReducedMotion = useReducedMotion();
 *
 * <motion.div
 *   animate={prefersReducedMotion ? {} : { opacity: 1, y: 0 }}
 * />
 */
export function useReducedMotion() {
  const prefersReducedMotion = useSyncExternalStore(
    (callback) => {
      const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
      mediaQuery.addEventListener("change", callback);
      return () => mediaQuery.removeEventListener("change", callback);
    },
    () => window.matchMedia("(prefers-reduced-motion: reduce)").matches,
    () => false // Server-side snapshot
  );

  return prefersReducedMotion;
}
