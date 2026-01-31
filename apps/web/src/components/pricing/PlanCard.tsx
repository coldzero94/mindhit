"use client";

import { Check } from "lucide-react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { FadeInUp } from "@/components/ui/animations";
import { cn } from "@/lib/utils";
import type { SubscriptionPlan } from "@/api/generated/types.gen";

interface PlanCardProps {
  plan: SubscriptionPlan;
  currentPlanId?: string;
  isLoggedIn: boolean;
  highlighted?: boolean;
  onSelect?: (planId: string) => void;
  /** Animation delay index (for staggered animations) */
  index?: number;
}

export function PlanCard({
  plan,
  currentPlanId,
  isLoggedIn,
  highlighted = false,
  onSelect,
  index = 0,
}: PlanCardProps) {
  const router = useRouter();
  const isCurrentPlan = currentPlanId === plan.id;
  const isUpgrade =
    currentPlanId === "free" && (plan.id === "pro" || plan.id === "enterprise");
  const isDowngrade = currentPlanId === "pro" && plan.id === "free";

  const handleSelect = () => {
    if (!isLoggedIn) {
      router.push(`/login?redirect=/pricing`);
      return;
    }

    if (plan.id === "enterprise") {
      window.location.href =
        "mailto:sales@mindhit.dev?subject=Enterprise Plan Inquiry";
      return;
    }

    if (!isCurrentPlan && onSelect) {
      onSelect(plan.id);
    }
  };

  const formatPrice = () => {
    if (plan.id === "enterprise") return "문의";
    if (plan.price_cents === 0) return "무료";
    return `$${plan.price_cents / 100}/월`;
  };

  const getButtonText = () => {
    if (!isLoggedIn) {
      if (plan.id === "free") return "무료 시작";
      if (plan.id === "enterprise") return "문의하기";
      return "시작하기";
    }
    if (isCurrentPlan) return "현재 플랜";
    if (plan.id === "enterprise") return "문의하기";
    if (isDowngrade) return "다운그레이드";
    if (isUpgrade) return "업그레이드";
    return "선택";
  };

  const getButtonVariant = (): "default" | "outline" | "secondary" => {
    if (isCurrentPlan) return "secondary";
    if (highlighted) return "default";
    return "outline";
  };

  return (
    <FadeInUp delay={index * 0.1}>
      <div
        className={cn(
          "relative rounded-2xl border-2 p-6 bg-card transition-all",
          highlighted
            ? "border-primary bg-primary/5 shadow-lg scale-[1.02]"
            : "border-border hover:border-primary/50"
        )}
      >
        {highlighted && (
          <div className="absolute -top-3 left-1/2 -translate-x-1/2">
            <span className="bg-primary text-primary-foreground text-xs font-medium px-3 py-1 rounded-full">
              추천
            </span>
          </div>
        )}

        <div className="text-center mb-6">
          <h3 className="text-xl font-bold text-foreground">{plan.name}</h3>
          <div className="mt-4">
            <span className="text-4xl font-bold text-foreground">
              {formatPrice()}
            </span>
          </div>
        </div>

        <div className="space-y-3 mb-6">
          <div className="flex items-center gap-2 text-sm text-foreground">
            <Check className="w-4 h-4 text-status-success flex-shrink-0" />
            <span>
              {plan.token_limit
                ? `${plan.token_limit.toLocaleString()} 토큰/월`
                : "무제한 토큰"}
            </span>
          </div>
          <div className="flex items-center gap-2 text-sm text-foreground">
            <Check className="w-4 h-4 text-status-success flex-shrink-0" />
            <span>
              {plan.session_retention_days
                ? `${plan.session_retention_days}일 세션 보관`
                : "무제한 보관"}
            </span>
          </div>
          <div className="flex items-center gap-2 text-sm text-foreground">
            <Check className="w-4 h-4 text-status-success flex-shrink-0" />
            <span>
              {plan.max_concurrent_sessions
                ? `동시 ${plan.max_concurrent_sessions}개 세션`
                : "무제한 세션"}
            </span>
          </div>
        </div>

        <div className="space-y-2 mb-6 text-sm text-muted-foreground">
          {plan.features?.export_png && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>PNG 내보내기</span>
            </div>
          )}
          {plan.features?.export_svg && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>SVG 내보내기</span>
            </div>
          )}
          {plan.features?.export_md && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>Markdown 내보내기</span>
            </div>
          )}
          {plan.features?.export_json && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>JSON 내보내기</span>
            </div>
          )}
          {plan.features?.priority_support && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>우선 지원</span>
            </div>
          )}
          {plan.features?.api_access && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>API 접근</span>
            </div>
          )}
          {plan.features?.team_sharing && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>팀 공유</span>
            </div>
          )}
          {plan.features?.sso && (
            <div className="flex items-center gap-2">
              <Check className="w-4 h-4 text-status-success" />
              <span>SSO 로그인</span>
            </div>
          )}
        </div>

        <Button
          onClick={handleSelect}
          disabled={isCurrentPlan}
          variant={getButtonVariant()}
          className="w-full"
        >
          {getButtonText()}
        </Button>
      </div>
    </FadeInUp>
  );
}
