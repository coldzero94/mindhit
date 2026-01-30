"use client";

import { FullPageSection } from "./animations/FullPageSection";
import { ScrollReveal } from "./animations/ScrollReveal";

/**
 * ProblemSection - Illustrate the information overload problem
 *
 * Features:
 * - Full-page sticky section with off-white background
 * - Center visual: scattered nodes representing chaos/information overload
 * - Nodes disperse and fade as user scrolls (ScrollReveal)
 */
export function ProblemSection() {
  return (
    <FullPageSection className="section-lighter" zIndex={20}>
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-center min-h-screen py-20">
          {/* Headline */}
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold text-center text-foreground mb-8 sm:mb-12">
            정보의 홍수 속에서 <br className="hidden sm:block" />
            길을 잃으셨나요?
          </h2>

          {/* Center Visual - Information Overload */}
          <div className="my-12 sm:my-16 relative">
            <ScrollReveal
              scaleRange={[1, 1, 1.2]}
              opacityRange={[1, 1, 0.5]}
              className="w-full max-w-xl mx-auto"
            >
              <div className="relative w-64 h-64 sm:w-80 sm:h-80 mx-auto">
                {/* Scattered nodes representing information chaos */}
                {[
                  { top: "10%", left: "20%", size: "w-12 h-12", color: "from-red-400 to-red-600", delay: 0 },
                  { top: "15%", left: "70%", size: "w-10 h-10", color: "from-blue-400 to-blue-600", delay: 0.1 },
                  { top: "30%", left: "10%", size: "w-14 h-14", color: "from-green-400 to-green-600", delay: 0.2 },
                  { top: "35%", left: "80%", size: "w-8 h-8", color: "from-yellow-400 to-yellow-600", delay: 0.3 },
                  { top: "50%", left: "45%", size: "w-16 h-16", color: "from-purple-400 to-purple-600", delay: 0.4 },
                  { top: "60%", left: "25%", size: "w-10 h-10", color: "from-pink-400 to-pink-600", delay: 0.5 },
                  { top: "65%", left: "65%", size: "w-12 h-12", color: "from-indigo-400 to-indigo-600", delay: 0.6 },
                  { top: "80%", left: "40%", size: "w-8 h-8", color: "from-orange-400 to-orange-600", delay: 0.7 },
                ].map((node, index) => (
                  <div
                    key={index}
                    className={`absolute ${node.size} rounded-full bg-gradient-to-br ${node.color} shadow-lg opacity-70`}
                    style={{
                      top: node.top,
                      left: node.left,
                      animation: `float ${3 + index * 0.5}s ease-in-out infinite`,
                      animationDelay: `${node.delay}s`,
                    }}
                  />
                ))}

                {/* Connection lines (subtle) */}
                <svg className="absolute inset-0 opacity-20" viewBox="0 0 320 320">
                  <line x1="64" y1="32" x2="160" y2="160" stroke="currentColor" strokeWidth="1" />
                  <line x1="224" y1="48" x2="160" y2="160" stroke="currentColor" strokeWidth="1" />
                  <line x1="32" y1="96" x2="160" y2="160" stroke="currentColor" strokeWidth="1" />
                  <line x1="256" y1="112" x2="160" y2="160" stroke="currentColor" strokeWidth="1" />
                  <line x1="80" y1="192" x2="160" y2="160" stroke="currentColor" strokeWidth="1" />
                  <line x1="208" y1="208" x2="160" y2="160" stroke="currentColor" strokeWidth="1" />
                  <line x1="128" y1="256" x2="160" y2="160" stroke="currentColor" strokeWidth="1" />
                </svg>
              </div>
            </ScrollReveal>
          </div>

          {/* Description */}
          <p className="text-lg sm:text-xl text-center text-muted-foreground max-w-2xl">
            매일 수많은 웹페이지를 방문하지만, <br className="hidden sm:block" />
            정작 중요한 인사이트는 놓치고 계십니다.
          </p>
        </div>
      </div>

      {/* Floating animation keyframes */}
      <style jsx>{`
        @keyframes float {
          0%, 100% {
            transform: translateY(0px) rotate(0deg);
          }
          50% {
            transform: translateY(-20px) rotate(5deg);
          }
        }
      `}</style>
    </FullPageSection>
  );
}
