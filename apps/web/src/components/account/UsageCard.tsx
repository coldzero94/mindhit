"use client";

import { Zap, AlertTriangle } from "lucide-react";
import { useUsage } from "@/lib/hooks/use-usage";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";

export function UsageCard() {
  const { data, isLoading } = useUsage();

  if (isLoading) {
    return (
      <div className="p-6 bg-white rounded-xl border border-gray-200">
        <div className="flex items-center gap-3 mb-4">
          <Skeleton className="w-12 h-12 rounded-xl" />
          <div>
            <Skeleton className="h-5 w-24 mb-2" />
            <Skeleton className="h-4 w-16" />
          </div>
        </div>
        <Skeleton className="h-8 w-32 mb-3" />
        <Skeleton className="h-2 w-full" />
      </div>
    );
  }

  if (!data?.usage) return null;

  const usage = data.usage;
  const percentage = usage.percent_used;
  const isNearLimit = percentage >= 80;
  const isOverLimit = percentage >= 100;

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  return (
    <div className="p-6 bg-white rounded-xl border border-gray-200">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div
            className={`p-3 rounded-xl ${isNearLimit ? "bg-yellow-100" : "bg-blue-100"}`}
          >
            <Zap
              className={`w-6 h-6 ${isNearLimit ? "text-yellow-600" : "text-blue-600"}`}
            />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900">토큰 사용량</h3>
            <p className="text-sm text-gray-500">이번 달</p>
          </div>
        </div>

        {isNearLimit && !isOverLimit && (
          <div className="flex items-center gap-1 px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm">
            <AlertTriangle className="w-4 h-4" />
            <span>80% 도달</span>
          </div>
        )}
        {isOverLimit && (
          <div className="flex items-center gap-1 px-3 py-1 bg-red-100 text-red-700 rounded-full text-sm">
            <AlertTriangle className="w-4 h-4" />
            <span>한도 초과</span>
          </div>
        )}
      </div>

      <div className="space-y-3">
        <div className="flex items-end justify-between">
          <span className="text-3xl font-bold text-gray-900">
            {formatNumber(usage.tokens_used)}
          </span>
          <span className="text-sm text-gray-500">
            / {usage.is_unlimited ? "무제한" : formatNumber(usage.token_limit)}{" "}
            토큰
          </span>
        </div>

        {!usage.is_unlimited && <Progress value={percentage} />}

        <p className="text-xs text-gray-400">
          {usage.is_unlimited
            ? "무제한 사용 중"
            : `${percentage.toFixed(1)}% 사용 중`}
        </p>
      </div>
    </div>
  );
}
