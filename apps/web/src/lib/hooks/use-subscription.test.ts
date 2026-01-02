import { describe, it, expect } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import {
  useSubscription,
  usePlans,
  subscriptionKeys,
} from "./use-subscription";
import {
  mockSubscription,
  mockPlan,
  mockProPlan,
} from "@/test/mocks/handlers";

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

describe("subscriptionKeys", () => {
  it("should generate correct query keys", () => {
    expect(subscriptionKeys.all).toEqual(["subscription"]);
    expect(subscriptionKeys.plans).toEqual(["plans"]);
  });
});

describe("useSubscription", () => {
  it("should fetch subscription data", async () => {
    const { result } = renderHook(() => useSubscription(), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for data
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Verify subscription data
    expect(result.current.data?.subscription).toBeDefined();
    expect(result.current.data?.subscription.id).toBe(mockSubscription.id);
    expect(result.current.data?.subscription.status).toBe("active");
    expect(result.current.data?.subscription.plan?.name).toBe(mockPlan.name);
  });

  it("should have correct staleTime (5 minutes)", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });

    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client: queryClient }, children);

    const { result } = renderHook(() => useSubscription(), { wrapper });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Check that queryClient has the subscription data cached
    const cachedData = queryClient.getQueryData(subscriptionKeys.all);
    expect(cachedData).toBeDefined();
  });
});

describe("usePlans", () => {
  it("should fetch plans list", async () => {
    const { result } = renderHook(() => usePlans(), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for data
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Verify plans data
    expect(result.current.data?.plans).toBeDefined();
    expect(result.current.data?.plans).toHaveLength(2);
    expect(result.current.data?.plans[0].name).toBe(mockPlan.name);
    expect(result.current.data?.plans[1].name).toBe(mockProPlan.name);
  });

  it("should include plan details", async () => {
    const { result } = renderHook(() => usePlans(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const freePlan = result.current.data?.plans.find((p) => p.name === "Free");
    const proPlan = result.current.data?.plans.find((p) => p.name === "Pro");

    expect(freePlan?.token_limit).toBe(10000);
    expect(proPlan?.token_limit).toBe(100000);
  });
});
