"use client";

import Link from "next/link";
import { motion, useScroll, useTransform } from "framer-motion";
import { Button } from "@/components/ui/button";

/**
 * LandingNavbar - Fixed navigation bar with scroll-triggered background
 *
 * Features:
 * - Fixed positioning at top
 * - Transparent initially, becomes opaque on scroll
 * - Smooth background transition
 */
export function LandingNavbar() {
  const { scrollY } = useScroll();

  // Background opacity increases as user scrolls down
  // 0-50px scroll: opacity 0 → 1
  const backgroundOpacity = useTransform(scrollY, [0, 50], [0, 1]);

  return (
    <motion.nav
      style={{
        backgroundColor: useTransform(
          backgroundOpacity,
          (value) => `oklch(1 0 0 / ${value * 0.95})` // white with opacity
        ),
        backdropFilter: useTransform(
          backgroundOpacity,
          (value) => `blur(${value * 10}px)` // blur increases with scroll
        ),
      }}
      className="fixed top-0 left-0 right-0 z-50 border-b border-border/0 transition-colors"
    >
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link
            href="/"
            className="text-xl font-bold text-foreground hover:text-primary transition-colors"
          >
            MindHit
          </Link>

          {/* Navigation Links */}
          <div className="flex items-center gap-4">
            <Link href="/sessions">
              <Button>Dashboard</Button>
            </Link>
          </div>
        </div>
      </div>
    </motion.nav>
  );
}
