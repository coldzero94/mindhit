"use client";

import { BarChart3 } from "lucide-react";
import { useUsageHistory } from "@/lib/hooks/use-usage";
import { Skeleton } from "@/components/ui/skeleton";

export function UsageHistory() {
  const { data, isLoading } = useUsageHistory(6);

  if (isLoading) {
    return (
      <div className="p-6 bg-white rounded-xl border border-gray-200">
        <div className="flex items-center gap-2 mb-6">
          <Skeleton className="w-5 h-5 rounded" />
          <Skeleton className="h-5 w-32" />
        </div>
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="space-y-2">
              <div className="flex justify-between">
                <Skeleton className="h-4 w-20" />
                <Skeleton className="h-4 w-24" />
              </div>
              <Skeleton className="h-6 w-full rounded-lg" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  const history = data?.history;

  if (!history || history.length === 0) {
    return (
      <div className="p-6 bg-white rounded-xl border border-gray-200">
        <div className="flex items-center gap-2 mb-4">
          <BarChart3 className="w-5 h-5 text-gray-400" />
          <h3 className="text-lg font-semibold text-gray-900">
            사용량 히스토리
          </h3>
        </div>
        <p className="text-sm text-gray-500">아직 사용 기록이 없습니다.</p>
      </div>
    );
  }

  const maxUsage = Math.max(...history.map((h) => h.tokens_used));

  const formatMonth = (monthStr: string) => {
    const date = new Date(monthStr);
    return date.toLocaleDateString("ko-KR", { year: "numeric", month: "short" });
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  return (
    <div className="p-6 bg-white rounded-xl border border-gray-200">
      <div className="flex items-center gap-2 mb-6">
        <BarChart3 className="w-5 h-5 text-gray-400" />
        <h3 className="text-lg font-semibold text-gray-900">
          사용량 히스토리
        </h3>
      </div>

      <div className="space-y-4">
        {history.map((item) => {
          const barWidth =
            maxUsage > 0 ? (item.tokens_used / maxUsage) * 100 : 0;
          const usagePercentage = item.is_unlimited
            ? 0
            : (item.tokens_used / item.token_limit) * 100;

          return (
            <div key={item.period_start} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">
                  {formatMonth(item.period_start)}
                </span>
                <span className="text-gray-900 font-medium">
                  {formatNumber(item.tokens_used)}
                  {!item.is_unlimited && ` / ${formatNumber(item.token_limit)}`}
                </span>
              </div>
              <div className="relative h-6 bg-gray-100 rounded-lg overflow-hidden">
                <div
                  className={`absolute left-0 top-0 h-full rounded-lg transition-all duration-500 ${
                    usagePercentage >= 90
                      ? "bg-red-500"
                      : usagePercentage >= 80
                        ? "bg-yellow-500"
                        : "bg-blue-500"
                  }`}
                  style={{ width: `${barWidth}%` }}
                />
                {!item.is_unlimited && (
                  <span className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-500">
                    {usagePercentage.toFixed(0)}%
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
