/**
 * Shared Animation Components
 *
 * Reusable animation wrappers for consistent UX across the app.
 * All animations respect prefers-reduced-motion for accessibility.
 */

// New global animations
export { FadeIn } from "./FadeIn";
export { StaggerChildren, staggerItem } from "./StaggerChildren";
export { PageTransition } from "./PageTransition";

// Re-export landing page animations for convenience
// These are more specialized but can be used elsewhere if needed
export { FadeInUp } from "@/components/landing/animations/FadeInUp";
