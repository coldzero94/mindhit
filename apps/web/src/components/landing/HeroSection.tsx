"use client";

import Link from "next/link";
import { FullPageSection } from "./animations/FullPageSection";
import { ScrollReveal } from "./animations/ScrollReveal";
import { TextReveal } from "./animations/TextReveal";
import { Button } from "@/components/ui/button";
import { ArrowDown } from "lucide-react";

/**
 * HeroSection - iPhone-style hero with center-stage visual
 *
 * Features:
 * - Full-page sticky section
 * - Center visual with scroll-linked scale/rotate
 * - Text reveal animation for headline
 * - Dual CTAs (primary + outline)
 */
export function HeroSection() {
  const handleScrollDown = () => {
    window.scrollTo({
      top: window.innerHeight,
      behavior: "smooth",
    });
  };

  return (
    <FullPageSection className="section-light" zIndex={10}>
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-center min-h-screen py-20">
          {/* Headline with text reveal */}
          <div className="text-center mb-8 sm:mb-12">
            <TextReveal
              text="브라우징을 인사이트로 변환하세요"
              as="h1"
              className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold text-foreground mb-4 sm:mb-6"
              delay={0.1}
            />
            <p className="text-lg sm:text-xl md:text-2xl text-muted-foreground max-w-3xl mx-auto">
              AI가 당신의 브라우징 히스토리를 분석하여 <br className="hidden sm:block" />
              아름다운 3D 마인드맵으로 시각화합니다
            </p>
          </div>

          {/* Center-stage visual with scroll-linked animation */}
          <div className="my-12 sm:my-16 md:my-20">
            <ScrollReveal
              scaleRange={[0.9, 1, 0.95]}
              rotateRange={[0, 0, 3]}
              className="w-full max-w-2xl mx-auto"
            >
              {/* Placeholder for hero visual */}
              {/* TODO: Replace with actual 3D preview or mockup image */}
              <div className="relative aspect-video rounded-2xl bg-gradient-to-br from-primary/5 via-accent/10 to-primary/5 border-2 border-border shadow-2xl overflow-hidden">
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="text-center p-8">
                    {/* Abstract mindmap visualization */}
                    <div className="relative w-48 h-48 mx-auto">
                      {/* Center node */}
                      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-16 h-16 rounded-full bg-gradient-to-br from-yellow-400 to-yellow-600 shadow-lg animate-pulse" />

                      {/* Surrounding nodes */}
                      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-12 h-12 rounded-full bg-gradient-to-br from-blue-400 to-blue-600 shadow-lg opacity-80" />
                      <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-12 h-12 rounded-full bg-gradient-to-br from-green-400 to-green-600 shadow-lg opacity-80" />
                      <div className="absolute top-1/2 left-0 -translate-y-1/2 w-12 h-12 rounded-full bg-gradient-to-br from-purple-400 to-purple-600 shadow-lg opacity-80" />
                      <div className="absolute top-1/2 right-0 -translate-y-1/2 w-12 h-12 rounded-full bg-gradient-to-br from-pink-400 to-pink-600 shadow-lg opacity-80" />

                      {/* Connection lines */}
                      <svg className="absolute inset-0" viewBox="0 0 192 192">
                        <line x1="96" y1="96" x2="96" y2="24" stroke="oklch(0.8 0 0)" strokeWidth="2" strokeOpacity="0.3" />
                        <line x1="96" y1="96" x2="96" y2="168" stroke="oklch(0.8 0 0)" strokeWidth="2" strokeOpacity="0.3" />
                        <line x1="96" y1="96" x2="24" y2="96" stroke="oklch(0.8 0 0)" strokeWidth="2" strokeOpacity="0.3" />
                        <line x1="96" y1="96" x2="168" y2="96" stroke="oklch(0.8 0 0)" strokeWidth="2" strokeOpacity="0.3" />
                      </svg>
                    </div>

                    <p className="mt-6 text-sm text-muted-foreground">
                      3D 마인드맵 프리뷰
                    </p>
                  </div>
                </div>
              </div>
            </ScrollReveal>
          </div>

          {/* CTAs */}
          <div className="flex flex-col sm:flex-row items-center gap-4 mb-12">
            <Link href="/signup">
              <Button size="lg" className="text-lg px-8 py-6">
                무료로 시작하기
              </Button>
            </Link>
            <Button
              size="lg"
              variant="outline"
              className="text-lg px-8 py-6"
              onClick={handleScrollDown}
            >
              자세히 알아보기
            </Button>
          </div>

          {/* Scroll indicator */}
          <button
            onClick={handleScrollDown}
            className="absolute bottom-8 left-1/2 -translate-x-1/2 text-muted-foreground hover:text-foreground transition-colors animate-bounce"
            aria-label="Scroll down"
          >
            <ArrowDown className="w-6 h-6" />
          </button>
        </div>
      </div>
    </FullPageSection>
  );
}
