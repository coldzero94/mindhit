"use client";

import { Crown, Calendar, ArrowUpRight } from "lucide-react";
import { useSubscription } from "@/lib/hooks/use-subscription";
import { Skeleton } from "@/components/ui/skeleton";

interface SubscriptionCardProps {
  onUpgrade?: () => void;
}

export function SubscriptionCard({ onUpgrade }: SubscriptionCardProps) {
  const { data, isLoading } = useSubscription();

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
  const isFree = plan?.name?.toLowerCase() === "free";

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("ko-KR", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const statusLabels: Record<string, { label: string; color: string }> = {
    active: { label: "활성", color: "bg-green-100 text-green-700" },
    canceled: { label: "취소됨", color: "bg-red-100 text-red-700" },
    past_due: { label: "연체", color: "bg-yellow-100 text-yellow-700" },
  };

  const status = statusLabels[subscription.status] || statusLabels.active;

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

        {isFree && onUpgrade && (
          <button
            onClick={onUpgrade}
            className="flex items-center gap-1 px-4 py-2 text-sm font-medium
                       bg-gradient-to-r from-blue-600 to-purple-600 text-white
                       rounded-lg hover:from-blue-700 hover:to-purple-700
                       transition-all"
          >
            업그레이드
            <ArrowUpRight className="w-4 h-4" />
          </button>
        )}
      </div>

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
        <div className="mt-4 pt-4 border-t border-gray-100">
          <p className="text-sm text-gray-500">
            월 {plan.token_limit?.toLocaleString() || "무제한"} 토큰 제공
          </p>
        </div>
      )}
    </div>
  );
}
