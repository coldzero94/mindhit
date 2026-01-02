import { apiClient } from "./client";
import type {
  UsageUsageResponse,
  UsageUsageHistoryResponse,
} from "@/api/generated/types.gen";

export async function getUsage(): Promise<UsageUsageResponse> {
  const response = await apiClient.get<UsageUsageResponse>("/usage");
  return response.data;
}

export async function getUsageHistory(
  months: number = 6
): Promise<UsageUsageHistoryResponse> {
  const response = await apiClient.get<UsageUsageHistoryResponse>(
    `/usage/history?months=${months}`
  );
  return response.data;
}
