import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  changePlan,
  cancelSubscription,
  reactivateSubscription,
} from "@/lib/api/subscription";
import { toast } from "sonner";

export function useChangePlan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: changePlan,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
      toast.success(data.message || "플랜이 변경되었습니다");
    },
    onError: (error: Error & { response?: { data?: { error?: { message?: string } } } }) => {
      toast.error(
        error.response?.data?.error?.message || "플랜 변경에 실패했습니다"
      );
    },
  });
}

export function useCancelSubscription() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: cancelSubscription,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
      toast.success(data.message || "구독이 취소 예정으로 설정되었습니다");
    },
    onError: (error: Error & { response?: { data?: { error?: { message?: string } } } }) => {
      toast.error(
        error.response?.data?.error?.message || "구독 취소에 실패했습니다"
      );
    },
  });
}

export function useReactivateSubscription() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: reactivateSubscription,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["subscription"] });
      toast.success("구독이 재활성화되었습니다");
    },
    onError: (error: Error & { response?: { data?: { error?: { message?: string } } } }) => {
      toast.error(
        error.response?.data?.error?.message || "재활성화에 실패했습니다"
      );
    },
  });
}
