"use client";

import Link from "next/link";
import { FadeInUp } from "./animations/FadeInUp";
import { Button } from "@/components/ui/button";
import { Check, Loader2 } from "lucide-react";
import { usePlans } from "@/lib/hooks/use-subscription";

/**
 * PricingSection - Preview of pricing plans with footer
 *
 * Features:
 * - Uses real API data from usePlans hook
 * - Links to full pricing page
 * - Fade-in animations
 * - Includes footer content (no separate Footer component needed)
 * - Regular flow section (not sticky) to avoid overlapping issues
 */
export function PricingSection() {
  const { data: plansData, isLoading } = usePlans();
  const plans = plansData?.plans || [];

  const formatPrice = (priceCents: number, planId: string) => {
    if (planId === "enterprise") return "문의";
    if (priceCents === 0) return "무료";
    return `$${priceCents / 100}`;
  };

  const getFeatures = (plan: typeof plans[0]) => {
    const features = [];

    // Token limit
    if (plan.token_limit === -1) {
      features.push("무제한 AI 토큰");
    } else if (plan.token_limit !== undefined) {
      features.push(`월 ${plan.token_limit.toLocaleString()} AI 토큰`);
    }

    // Max concurrent sessions
    if (plan.max_concurrent_sessions === -1) {
      features.push("무제한 동시 세션");
    } else if (plan.max_concurrent_sessions !== undefined) {
      features.push(`최대 ${plan.max_concurrent_sessions}개 동시 세션`);
    }

    // Session retention
    if (plan.session_retention_days === -1) {
      features.push("무제한 세션 보관");
    } else if (plan.session_retention_days !== undefined) {
      features.push(`${plan.session_retention_days}일 세션 보관`);
    }

    // Additional features by plan
    if (plan.id === "free") {
      features.push("기본 마인드맵");
      features.push("Chrome Extension");
    } else if (plan.id === "pro") {
      features.push("고급 AI 분석");
      features.push("3D 인터랙티브 마인드맵");
      features.push("우선 지원");
    } else if (plan.id === "enterprise") {
      features.push("팀 협업 기능");
      features.push("맞춤형 AI 모델");
      features.push("전담 지원");
      features.push("온프레미스 옵션");
    }

    return features;
  };

  return (
    <section className="relative w-full section-light min-h-screen pt-20 pb-12" style={{ zIndex: 80 }}>
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-center">
          {/* Title - Always visible */}
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold text-center text-foreground mb-4">
            모든 규모의 팀을 위한 플랜
          </h2>
          <p className="text-lg sm:text-xl text-center text-muted-foreground mb-16">
            필요에 맞는 플랜을 선택하세요
          </p>

          {/* Loading State */}
          {isLoading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-gray-400" />
            </div>
          ) : (
            <>
              {/* Plan Cards */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-5xl mb-12">
                {plans
                  .sort((a, b) => {
                    const order = ["free", "pro", "enterprise"];
                    return order.indexOf(a.id) - order.indexOf(b.id);
                  })
                  .map((plan, index) => {
                    const features = getFeatures(plan);
                    const highlighted = plan.id === "pro";

                    return (
                      <FadeInUp key={plan.id} delay={index * 0.1}>
                        <div
                          className={`p-6 sm:p-8 rounded-2xl border-2 transition-all hover:shadow-xl ${
                            highlighted
                              ? "border-primary bg-primary/5 shadow-lg scale-105"
                              : "border-border bg-card"
                          }`}
                        >
                          {highlighted && (
                            <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                              <span className="bg-primary text-primary-foreground text-xs font-medium px-3 py-1 rounded-full">
                                추천
                              </span>
                            </div>
                          )}
                          <h3 className="text-xl font-bold text-foreground mb-2">
                            {plan.name}
                          </h3>
                          <div className="mb-6">
                            <span className="text-3xl font-bold text-foreground">
                              {formatPrice(plan.price_cents, plan.id)}
                            </span>
                            {plan.price_cents > 0 && plan.id !== "enterprise" && (
                              <span className="text-muted-foreground ml-2">
                                / 월
                              </span>
                            )}
                          </div>
                          <ul className="space-y-3 mb-6">
                            {features.map((feature, i) => (
                              <li key={i} className="flex items-start gap-2 text-sm">
                                <Check className="w-5 h-5 text-green-500 flex-shrink-0 mt-0.5" />
                                <span className="text-foreground">{feature}</span>
                              </li>
                            ))}
                          </ul>
                        </div>
                      </FadeInUp>
                    );
                  })}
              </div>

              {/* CTA to Full Pricing Page */}
              <FadeInUp delay={0.4}>
                <Link href="/pricing">
                  <Button variant="outline" size="lg">
                    전체 플랜 및 기능 비교 보기
                  </Button>
                </Link>
              </FadeInUp>
            </>
          )}

          {/* Simple Footer */}
          <div className="w-full mt-20 pt-8 border-t border-border">
            <div className="flex flex-col sm:flex-row justify-between items-center gap-4 pb-12">
              {/* Logo & Copyright */}
              <div className="flex flex-col sm:flex-row items-center gap-2">
                <span className="font-bold text-foreground">MindHit</span>
                <span className="text-muted-foreground text-sm">
                  © 2026 MindHit. All rights reserved.
                </span>
              </div>

              {/* GitHub Link */}
              <div className="flex items-center gap-4">
                <a
                  href="https://github.com/coldzero94/mindhit"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-muted-foreground hover:text-foreground transition-colors"
                  aria-label="GitHub"
                >
                  <svg
                    className="w-5 h-5"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
                  </svg>
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
