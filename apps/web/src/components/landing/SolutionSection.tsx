"use client";

import { FullPageSection } from "./animations/FullPageSection";
import { ScrollReveal } from "./animations/ScrollReveal";
import { Zap, Brain, Box } from "lucide-react";

/**
 * SolutionSection - Present MindHit's 3-step solution
 *
 * Features:
 * - Full-page sticky section with gray background
 * - Three icons with staggered scroll reveals
 * - Center-stage layout showcasing core features
 */
export function SolutionSection() {
  const features = [
    {
      icon: Zap,
      title: "자동 수집",
      description: "Chrome Extension이 자동으로 브라우징 추적",
      color: "text-yellow-500",
      bgColor: "bg-yellow-500/10",
      delay: 0,
    },
    {
      icon: Brain,
      title: "AI 분석",
      description: "키워드 추출 및 관계 그래프 생성",
      color: "text-purple-500",
      bgColor: "bg-purple-500/10",
      delay: 0.2,
    },
    {
      icon: Box,
      title: "3D 시각화",
      description: "아름다운 인터랙티브 마인드맵으로 표현",
      color: "text-blue-500",
      bgColor: "bg-blue-500/10",
      delay: 0.4,
    },
  ];

  return (
    <FullPageSection className="section-gray" zIndex={30}>
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-center min-h-screen py-20">
          {/* Headline */}
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold text-center text-foreground mb-16 sm:mb-20">
            MindHit가 해결합니다
          </h2>

          {/* Three Features with Icons */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 sm:gap-12 w-full max-w-5xl">
            {features.map((feature, index) => {
              const Icon = feature.icon;
              return (
                <ScrollReveal
                  key={index}
                  scaleRange={[0.8, 1, 1]}
                  opacityRange={[0, 1, 1]}
                  className="flex flex-col items-center text-center"
                >
                  {/* Icon */}
                  <div
                    className={`w-24 h-24 sm:w-28 sm:h-28 rounded-full ${feature.bgColor} flex items-center justify-center mb-6 shadow-lg`}
                  >
                    <Icon className={`w-12 h-12 sm:w-14 sm:h-14 ${feature.color}`} strokeWidth={1.5} />
                  </div>

                  {/* Title */}
                  <h3 className="text-xl sm:text-2xl font-bold text-foreground mb-3">
                    {feature.title}
                  </h3>

                  {/* Description */}
                  <p className="text-base sm:text-lg text-muted-foreground max-w-xs">
                    {feature.description}
                  </p>
                </ScrollReveal>
              );
            })}
          </div>

          {/* Connection Lines (Visual Enhancement) */}
          <div className="hidden md:block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-4xl h-0.5 bg-gradient-to-r from-transparent via-border to-transparent opacity-30" />
        </div>
      </div>
    </FullPageSection>
  );
}
