import { apiClient } from "./client";
import type {
  SubscriptionSubscriptionResponse,
  SubscriptionPlanListResponse,
  SubscriptionChangePlanResponse,
  SubscriptionCancelSubscriptionResponse,
} from "@/api/generated/types.gen";

export async function getSubscription(): Promise<SubscriptionSubscriptionResponse> {
  const response =
    await apiClient.get<SubscriptionSubscriptionResponse>("/subscription");
  return response.data;
}

export async function getPlans(): Promise<SubscriptionPlanListResponse> {
  const response =
    await apiClient.get<SubscriptionPlanListResponse>("/subscription/plans");
  return response.data;
}

export async function changePlan(
  planId: string
): Promise<SubscriptionChangePlanResponse> {
  const response = await apiClient.post<SubscriptionChangePlanResponse>(
    "/subscription/change",
    { plan_id: planId }
  );
  return response.data;
}

export async function cancelSubscription(): Promise<SubscriptionCancelSubscriptionResponse> {
  const response =
    await apiClient.post<SubscriptionCancelSubscriptionResponse>(
      "/subscription/cancel"
    );
  return response.data;
}

export async function reactivateSubscription(): Promise<SubscriptionSubscriptionResponse> {
  const response = await apiClient.post<SubscriptionSubscriptionResponse>(
    "/subscription/reactivate"
  );
  return response.data;
}
