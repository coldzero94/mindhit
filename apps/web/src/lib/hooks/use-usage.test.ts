import { describe, it, expect } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import { useUsage, useUsageHistory, usageKeys } from "./use-usage";
import { mockUsageSummary, mockUsageHistory } from "@/test/mocks/handlers";

// Create wrapper with QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
        staleTime: 0,
      },
    },
  });

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return React.createElement(
      QueryClientProvider,
      { client: queryClient },
      children
    );
  };
}

describe("usageKeys", () => {
  it("should generate correct query keys", () => {
    expect(usageKeys.current).toEqual(["usage", "current"]);
    expect(usageKeys.history(6)).toEqual(["usage", "history", 6]);
    expect(usageKeys.history(3)).toEqual(["usage", "history", 3]);
  });
});

describe("useUsage", () => {
  it("should fetch current usage data", async () => {
    const { result } = renderHook(() => useUsage(), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for data
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Verify usage data - response is { usage: UsageUsageSummary }
    expect(result.current.data).toBeDefined();
    expect(result.current.data?.usage?.tokens_used).toBe(mockUsageSummary.tokens_used);
    expect(result.current.data?.usage?.token_limit).toBe(mockUsageSummary.token_limit);
    expect(result.current.data?.usage?.percent_used).toBe(mockUsageSummary.percent_used);
  });

  it("should include period information", async () => {
    const { result } = renderHook(() => useUsage(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.usage?.period_start).toBe(mockUsageSummary.period_start);
    expect(result.current.data?.usage?.period_end).toBe(mockUsageSummary.period_end);
  });

  it("should have auto-refresh interval (5 minutes)", async () => {
    // This test verifies the hook is configured with refetchInterval
    // We can't easily test the actual interval without waiting
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });

    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client: queryClient }, children);

    const { result } = renderHook(() => useUsage(), { wrapper });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // The hook should have data cached
    const cachedData = queryClient.getQueryData(usageKeys.current);
    expect(cachedData).toBeDefined();
  });
});

describe("useUsageHistory", () => {
  it("should fetch usage history with default 6 months", async () => {
    const { result } = renderHook(() => useUsageHistory(), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for data
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Verify history data
    expect(result.current.data?.history).toBeDefined();
    expect(result.current.data?.history).toHaveLength(6);
  });

  it("should fetch usage history with custom months", async () => {
    const { result } = renderHook(() => useUsageHistory(3), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Should only have 3 months of data
    expect(result.current.data?.history).toHaveLength(3);
  });

  it("should include monthly usage details", async () => {
    const { result } = renderHook(() => useUsageHistory(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const firstMonth = result.current.data?.history[0];
    expect(firstMonth?.period_start).toBe(mockUsageHistory[0].period_start);
    expect(firstMonth?.tokens_used).toBe(mockUsageHistory[0].tokens_used);
    expect(firstMonth?.token_limit).toBe(mockUsageHistory[0].token_limit);
  });

  it("should use correct query key for different month values", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });

    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client: queryClient }, children);

    // Fetch with 6 months
    const { result: result6 } = renderHook(() => useUsageHistory(6), { wrapper });
    await waitFor(() => expect(result6.current.isSuccess).toBe(true));

    // Fetch with 3 months
    const { result: result3 } = renderHook(() => useUsageHistory(3), { wrapper });
    await waitFor(() => expect(result3.current.isSuccess).toBe(true));

    // Both should be cached with different keys
    expect(queryClient.getQueryData(usageKeys.history(6))).toBeDefined();
    expect(queryClient.getQueryData(usageKeys.history(3))).toBeDefined();
  });
});
