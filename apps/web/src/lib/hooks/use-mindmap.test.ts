import { describe, it, expect, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import { useMindmap, useGenerateMindmap, mindmapKeys } from "./use-mindmap";
import { mockMindmapNodes, mockMindmapEdges } from "@/test/mocks/handlers";

// Create wrapper with QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
        staleTime: 0,
      },
      mutations: {
        retry: false,
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

// Create wrapper that exposes queryClient for cache testing
function createWrapperWithClient() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
        staleTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });

  const wrapper = ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);

  return { wrapper, queryClient };
}

describe("mindmapKeys", () => {
  it("should generate correct query keys", () => {
    expect(mindmapKeys.all).toEqual(["mindmaps"]);
    expect(mindmapKeys.detail("session-1")).toEqual(["mindmaps", "session-1"]);
    expect(mindmapKeys.detail("session-abc")).toEqual(["mindmaps", "session-abc"]);
  });
});

describe("useMindmap", () => {
  it("should fetch mindmap data for a session", async () => {
    const { result } = renderHook(() => useMindmap("session-1"), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for data
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Verify mindmap data
    expect(result.current.data).toBeDefined();
    expect(result.current.data?.session_id).toBe("session-1");
    expect(result.current.data?.data?.nodes).toHaveLength(mockMindmapNodes.length);
    expect(result.current.data?.data?.edges).toHaveLength(mockMindmapEdges.length);
  });

  it("should include node details", async () => {
    const { result } = renderHook(() => useMindmap("session-1"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Check core node
    const coreNode = result.current.data?.data?.nodes.find((n) => n.type === "core");
    expect(coreNode).toBeDefined();
    expect(coreNode?.label).toBe("Test Session");

    // Check topic nodes
    const topicNodes = result.current.data?.data?.nodes.filter((n) => n.type === "topic");
    expect(topicNodes).toHaveLength(2);
  });

  it("should not fetch when sessionId is empty", async () => {
    const { result } = renderHook(() => useMindmap(""), {
      wrapper: createWrapper(),
    });

    // Should not be loading when disabled
    expect(result.current.isLoading).toBe(false);
    expect(result.current.fetchStatus).toBe("idle");
  });

  it("should handle 404 error (mindmap not found)", async () => {
    const { result } = renderHook(() => useMindmap("not-found"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toBeDefined();
  });

  it("should handle no-mindmap case (not generated yet)", async () => {
    const { result } = renderHook(() => useMindmap("no-mindmap"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });

  it("should not retry on 404 errors", async () => {
    const retryFn = vi.fn();
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: (failureCount, error) => {
            retryFn(failureCount, error);
            // Simulate 404 response
            if ((error as { response?: { status?: number } })?.response?.status === 404) {
              return false;
            }
            return failureCount < 3;
          },
          gcTime: 0,
        },
      },
    });

    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client: queryClient }, children);

    const { result } = renderHook(() => useMindmap("not-found"), { wrapper });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});

describe("useGenerateMindmap", () => {
  it("should generate mindmap successfully", async () => {
    const { result } = renderHook(() => useGenerateMindmap(), {
      wrapper: createWrapper(),
    });

    let responseData: unknown;
    await act(async () => {
      responseData = await result.current.mutateAsync({
        sessionId: "session-1",
      });
    });

    // Check returned data directly
    expect(responseData).toBeDefined();
    expect((responseData as { session_id: string })?.session_id).toBe("session-1");

    // Wait for state to update
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });

  it("should generate mindmap with force option", async () => {
    const { result } = renderHook(() => useGenerateMindmap(), {
      wrapper: createWrapper(),
    });

    let responseData: unknown;
    await act(async () => {
      responseData = await result.current.mutateAsync({
        sessionId: "session-1",
        options: { force: true },
      });
    });

    expect(responseData).toBeDefined();

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });

  it("should handle session not found error", async () => {
    const { result } = renderHook(() => useGenerateMindmap(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      try {
        await result.current.mutateAsync({
          sessionId: "not-found",
        });
      } catch {
        // Expected to throw
      }
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });

  it("should update cache after successful generation", async () => {
    const { wrapper, queryClient } = createWrapperWithClient();

    const { result } = renderHook(() => useGenerateMindmap(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        sessionId: "session-2",
      });
    });

    // Wait for mutation to complete and cache to update
    await waitFor(() => {
      const cachedData = queryClient.getQueryData(mindmapKeys.detail("session-2"));
      expect(cachedData).toBeDefined();
    });
  });
});
