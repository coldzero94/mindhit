"use client";

import { AlertTriangle, Zap } from "lucide-react";
import Link from "next/link";
import { useUsage } from "@/lib/hooks/use-usage";

export function HeaderUsageBadge() {
  const { data } = useUsage();

  if (!data?.usage) return null;

  const usage = data.usage;

  // 무제한이거나 80% 미만이면 표시하지 않음
  if (usage.is_unlimited || usage.percent_used < 80) return null;

  const isOver = usage.percent_used >= 100;

  return (
    <Link
      href="/account"
      className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
        isOver
          ? "bg-red-100 text-red-700 hover:bg-red-200"
          : "bg-yellow-100 text-yellow-700 hover:bg-yellow-200"
      }`}
    >
      {isOver ? (
        <AlertTriangle className="w-3.5 h-3.5" />
      ) : (
        <Zap className="w-3.5 h-3.5" />
      )}
      <span>
        {isOver ? "토큰 한도 초과" : `${usage.percent_used.toFixed(0)}% 사용`}
      </span>
    </Link>
  );
}
