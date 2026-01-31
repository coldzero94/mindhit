"use client";

import { BarChart3 } from "lucide-react";
import { useUsageHistory } from "@/lib/hooks/use-usage";
import { Skeleton } from "@/components/ui/skeleton";

export function UsageHistory() {
  const { data, isLoading } = useUsageHistory(6);

  if (isLoading) {
    return (
      <div className="p-6 bg-card rounded-xl border border-border">
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
      <div className="p-6 bg-card rounded-xl border border-border">
        <div className="flex items-center gap-2 mb-4">
          <BarChart3 className="w-5 h-5 text-muted-foreground" />
          <h3 className="text-lg font-semibold text-foreground">
            사용량 히스토리
          </h3>
        </div>
        <p className="text-sm text-muted-foreground">아직 사용 기록이 없습니다.</p>
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
    <div className="p-6 bg-card rounded-xl border border-border">
      <div className="flex items-center gap-2 mb-6">
        <BarChart3 className="w-5 h-5 text-muted-foreground" />
        <h3 className="text-lg font-semibold text-foreground">
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
                <span className="text-muted-foreground">
                  {formatMonth(item.period_start)}
                </span>
                <span className="text-foreground font-medium">
                  {formatNumber(item.tokens_used)}
                  {!item.is_unlimited && ` / ${formatNumber(item.token_limit)}`}
                </span>
              </div>
              <div className="relative h-6 bg-muted rounded-lg overflow-hidden">
                <div
                  className={`absolute left-0 top-0 h-full rounded-lg transition-all duration-500 ${
                    usagePercentage >= 90
                      ? "bg-status-error"
                      : usagePercentage >= 80
                        ? "bg-status-warning"
                        : "bg-status-info"
                  }`}
                  style={{ width: `${barWidth}%` }}
                />
                {!item.is_unlimited && (
                  <span className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
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
