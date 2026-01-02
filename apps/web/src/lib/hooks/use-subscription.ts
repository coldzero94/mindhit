import { useQuery } from "@tanstack/react-query";
import { getSubscription, getPlans } from "@/lib/api/subscription";

export const subscriptionKeys = {
  all: ["subscription"] as const,
  plans: ["plans"] as const,
};

export function useSubscription() {
  return useQuery({
    queryKey: subscriptionKeys.all,
    queryFn: getSubscription,
    staleTime: 5 * 60 * 1000, // 5분
  });
}

export function usePlans() {
  return useQuery({
    queryKey: subscriptionKeys.plans,
    queryFn: getPlans,
    staleTime: 30 * 60 * 1000, // 30분 (플랜은 자주 변하지 않음)
  });
}
