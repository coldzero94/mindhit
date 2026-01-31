"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Sparkles } from "lucide-react";

/**
 * FinalCTASection - Large call-to-action before footer
 *
 * Features:
 * - Dark gradient background
 * - Large CTA button with glow effect
 * - Compelling final message
 * - Not sticky - normal flow element for clean footer transition
 */
export function FinalCTASection() {
  return (
    <section className="relative min-h-screen section-dark overflow-hidden flex items-center justify-center" style={{ zIndex: 90 }}>
      {/* Gradient Background Effect */}
      <div className="absolute inset-0 bg-gradient-to-br from-purple-900/20 via-transparent to-blue-900/20" />

      {/* Glowing Orbs */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-pulse" />
      <div
        className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl animate-pulse"
        style={{ animationDelay: "1s" }}
      />

      <div className="container mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="flex flex-col items-center justify-center min-h-screen py-20 text-center">
          {/* Icon */}
          <Sparkles className="w-16 h-16 sm:w-20 sm:h-20 text-white mb-8 animate-pulse" />

          {/* Headline */}
          <h2 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-6 max-w-4xl">
            Browse Smarter
            <br />
            Starting Today
          </h2>

          {/* Subheadline */}
          <p className="text-lg sm:text-xl text-gray-300 mb-12 max-w-2xl">
            Take full control of your data with self-hosting.
            <br className="hidden sm:block" />
            Get started now.
          </p>

          {/* CTA Button with Glow */}
          <Link href="/sessions">
            <Button
              size="lg"
              className="text-xl px-12 py-8 shadow-2xl hover:shadow-purple-500/50 transition-all duration-300 animate-pulse"
            >
              <Sparkles className="w-6 h-6 mr-3" />
              Open Dashboard
            </Button>
          </Link>

          {/* Small note */}
          <p className="text-sm text-gray-400 mt-8">
            Open Source • Full Data Ownership
          </p>
        </div>
      </div>
    </section>
  );
}
