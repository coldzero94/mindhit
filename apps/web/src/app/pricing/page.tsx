"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, Loader2 } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { useSubscription, usePlans } from "@/lib/hooks/use-subscription";
import { useChangePlan } from "@/lib/hooks/use-change-plan";
import { PlanCard, PlanComparisonTable } from "@/components/pricing";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

export default function PricingPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { data: subscription } = useSubscription();
  const { data: plansData, isLoading } = usePlans();
  const changePlan = useChangePlan();

  const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);

  const currentPlanId = subscription?.subscription?.plan?.id;
  const plans = plansData?.plans || [];

  const handleSelectPlan = (planId: string) => {
    if (planId === currentPlanId) return;

    setSelectedPlanId(planId);
    setShowConfirmDialog(true);
  };

  const handleConfirmChange = async () => {
    if (!selectedPlanId) return;

    try {
      await changePlan.mutateAsync(selectedPlanId);
      setShowConfirmDialog(false);
      setSelectedPlanId(null);
      router.push("/account");
    } catch {
      // Error handled in hook
    }
  };

  const getSelectedPlan = () => {
    return plans.find((p) => p.id === selectedPlanId);
  };

  const isUpgrade = () => {
    if (!currentPlanId || !selectedPlanId) return false;
    const order = ["free", "pro", "enterprise"];
    return order.indexOf(selectedPlanId) > order.indexOf(currentPlanId);
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-50 to-white">
      <div className="max-w-6xl mx-auto px-4 py-8 sm:py-16">
        {/* Header */}
        <div className="mb-8">
          <Link
            href={isAuthenticated ? "/sessions" : "/"}
            className="inline-flex items-center text-gray-600 hover:text-gray-900 mb-6"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            {isAuthenticated ? "대시보드로 돌아가기" : "홈으로 돌아가기"}
          </Link>
        </div>

        <div className="text-center mb-12">
          <h1 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            당신의 브라우징을 마인드맵으로
          </h1>
          <p className="text-lg sm:text-xl text-gray-600">
            필요에 맞는 플랜을 선택하세요
          </p>
          {isAuthenticated && currentPlanId && (
            <p className="mt-4 text-sm text-blue-600">
              현재{" "}
              <span className="font-semibold">
                {plans.find((p) => p.id === currentPlanId)?.name || currentPlanId}
              </span>{" "}
              플랜을 사용 중입니다
            </p>
          )}
        </div>

        {/* Loading */}
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-gray-400" />
          </div>
        ) : (
          <>
            {/* Plan Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 sm:gap-8 mb-16">
              {plans
                .sort((a, b) => {
                  const order = ["free", "pro", "enterprise"];
                  return order.indexOf(a.id) - order.indexOf(b.id);
                })
                .map((plan) => (
                  <PlanCard
                    key={plan.id}
                    plan={plan}
                    currentPlanId={currentPlanId}
                    isLoggedIn={isAuthenticated}
                    highlighted={plan.id === "pro"}
                    onSelect={handleSelectPlan}
                  />
                ))}
            </div>

            {/* Comparison Table */}
            <div className="bg-white rounded-2xl border border-gray-200 p-6 sm:p-8">
              <h2 className="text-xl font-bold text-gray-900 mb-6 text-center">
                기능 비교
              </h2>
              <PlanComparisonTable
                plans={plans}
                currentPlanId={currentPlanId}
              />
            </div>
          </>
        )}

        {/* FAQ or Additional Info */}
        <div className="mt-16 text-center text-gray-600">
          <p>
            질문이 있으신가요?{" "}
            <a
              href="mailto:support@mindhit.dev"
              className="text-blue-600 hover:underline"
            >
              support@mindhit.dev
            </a>
            로 문의해 주세요.
          </p>
        </div>
      </div>

      {/* Confirm Dialog */}
      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {isUpgrade() ? "플랜 업그레이드" : "플랜 변경"}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {getSelectedPlan()?.name} 플랜으로{" "}
              {isUpgrade() ? "업그레이드" : "변경"}하시겠습니까?
              {!isUpgrade() && selectedPlanId === "free" && (
                <span className="block mt-2 text-amber-600">
                  다운그레이드 시 일부 기능이 제한될 수 있습니다.
                </span>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={changePlan.isPending}>
              취소
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmChange}
              disabled={changePlan.isPending}
            >
              {changePlan.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  처리 중...
                </>
              ) : (
                "확인"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
