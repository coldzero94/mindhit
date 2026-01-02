import { apiClient } from "./client";
import type {
  SubscriptionSubscriptionResponse,
  SubscriptionPlanListResponse,
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
