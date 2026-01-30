"use client";

import { useState } from "react";
import Link from "next/link";
import { Crown, Calendar, ArrowUpRight, AlertTriangle, Loader2 } from "lucide-react";
import { useSubscription } from "@/lib/hooks/use-subscription";
import {
  useCancelSubscription,
  useReactivateSubscription,
} from "@/lib/hooks/use-change-plan";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

export function SubscriptionCard() {
  const { data, isLoading } = useSubscription();
  const cancelSubscription = useCancelSubscription();
  const reactivateSubscription = useReactivateSubscription();
  const [showCancelDialog, setShowCancelDialog] = useState(false);

  if (isLoading) {
    return (
      <div className="p-6 bg-white rounded-xl border border-gray-200">
        <div className="flex items-center gap-3 mb-6">
          <Skeleton className="w-12 h-12 rounded-xl" />
          <div>
            <Skeleton className="h-5 w-24 mb-2" />
            <Skeleton className="h-4 w-16" />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-full" />
        </div>
      </div>
    );
  }

  if (!data?.subscription) return null;

  const subscription = data.subscription;
  const plan = subscription.plan;
  const isFree = plan?.id === "free";
  const isPro = plan?.id === "pro";
  const cancelAtPeriodEnd = subscription.cancel_at_period_end;

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("ko-KR", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const handleCancel = async () => {
    await cancelSubscription.mutateAsync();
    setShowCancelDialog(false);
  };

  const handleReactivate = async () => {
    await reactivateSubscription.mutateAsync();
  };

  const statusLabels: Record<string, { label: string; color: string }> = {
    active: { label: "활성", color: "bg-green-100 text-green-700" },
    canceled: { label: "취소됨", color: "bg-red-100 text-red-700" },
    past_due: { label: "연체", color: "bg-yellow-100 text-yellow-700" },
  };

  const status = cancelAtPeriodEnd
    ? { label: "취소 예정", color: "bg-amber-100 text-amber-700" }
    : statusLabels[subscription.status] || statusLabels.active;

  return (
    <div className="p-6 bg-white rounded-xl border border-gray-200">
      <div className="flex items-start justify-between mb-6">
        <div className="flex items-center gap-3">
          <div
            className={`p-3 rounded-xl ${isFree ? "bg-gray-100" : "bg-gradient-to-br from-yellow-400 to-orange-500"}`}
          >
            <Crown
              className={`w-6 h-6 ${isFree ? "text-gray-600" : "text-white"}`}
            />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900">
              {plan?.name || "Free"} 플랜
            </h3>
            <span
              className={`inline-block px-2 py-0.5 text-xs font-medium rounded-full ${status.color}`}
            >
              {status.label}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Link href="/pricing">
            <Button
              variant="outline"
              size="sm"
              className="flex items-center gap-1"
            >
              플랜 변경
              <ArrowUpRight className="w-4 h-4" />
            </Button>
          </Link>
        </div>
      </div>

      {/* Cancel at period end warning */}
      {cancelAtPeriodEnd && (
        <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-lg flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
          <div className="flex-1">
            <p className="text-sm text-amber-800">
              현재 기간 종료 후 Free 플랜으로 전환됩니다.
            </p>
            <Button
              variant="link"
              size="sm"
              className="p-0 h-auto text-amber-700 hover:text-amber-900"
              onClick={handleReactivate}
              disabled={reactivateSubscription.isPending}
            >
              {reactivateSubscription.isPending ? (
                <>
                  <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                  처리 중...
                </>
              ) : (
                "취소 철회"
              )}
            </Button>
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div className="flex items-center gap-2 text-sm text-gray-600">
          <Calendar className="w-4 h-4 text-gray-400" />
          <span>시작일: {formatDate(subscription.current_period_start)}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-gray-600">
          <Calendar className="w-4 h-4 text-gray-400" />
          <span>종료일: {formatDate(subscription.current_period_end)}</span>
        </div>
      </div>

      {plan && (
        <div className="mt-4 pt-4 border-t border-gray-100 flex items-center justify-between">
          <p className="text-sm text-gray-500">
            월 {plan.token_limit?.toLocaleString() || "무제한"} 토큰 제공
          </p>
          {isPro && !cancelAtPeriodEnd && (
            <AlertDialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
              <AlertDialogTrigger asChild>
                <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50">
                  구독 취소
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>구독을 취소하시겠습니까?</AlertDialogTitle>
                  <AlertDialogDescription>
                    현재 기간({formatDate(subscription.current_period_end)})까지는 Pro 기능을 계속 사용할 수 있습니다.
                    이후 Free 플랜으로 전환됩니다.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel disabled={cancelSubscription.isPending}>
                    유지하기
                  </AlertDialogCancel>
                  <AlertDialogAction
                    onClick={handleCancel}
                    disabled={cancelSubscription.isPending}
                    className="bg-red-600 hover:bg-red-700"
                  >
                    {cancelSubscription.isPending ? (
                      <>
                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                        처리 중...
                      </>
                    ) : (
                      "구독 취소"
                    )}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
        </div>
      )}
    </div>
  );
}
