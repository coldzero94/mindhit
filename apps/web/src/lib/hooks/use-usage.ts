import { useQuery } from "@tanstack/react-query";
import { getUsage, getUsageHistory } from "@/lib/api/usage";

export const usageKeys = {
  current: ["usage", "current"] as const,
  history: (months: number) => ["usage", "history", months] as const,
};

export function useUsage() {
  return useQuery({
    queryKey: usageKeys.current,
    queryFn: getUsage,
    staleTime: 60 * 1000, // 1분
    refetchInterval: 5 * 60 * 1000, // 5분마다 자동 갱신
  });
}

export function useUsageHistory(months: number = 6) {
  return useQuery({
    queryKey: usageKeys.history(months),
    queryFn: () => getUsageHistory(months),
    staleTime: 5 * 60 * 1000, // 5분
  });
}
