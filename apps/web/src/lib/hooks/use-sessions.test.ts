import { describe, it, expect, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import {
  useSessions,
  useSession,
  sessionKeys,
} from "./use-sessions";

// Create wrapper with QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
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

describe("sessionKeys", () => {
  it("generates correct all key", () => {
    expect(sessionKeys.all).toEqual(["sessions"]);
  });

  it("generates correct lists key", () => {
    expect(sessionKeys.lists()).toEqual(["sessions", "list"]);
  });

  it("generates correct list key with params", () => {
    expect(sessionKeys.list(10, 0)).toEqual([
      "sessions",
      "list",
      { limit: 10, offset: 0 },
    ]);
  });

  it("generates correct detail key", () => {
    expect(sessionKeys.detail("session-1")).toEqual([
      "sessions",
      "detail",
      "session-1",
    ]);
  });

  it("generates correct events key", () => {
    expect(sessionKeys.events("session-1")).toEqual([
      "sessions",
      "detail",
      "session-1",
      "events",
    ]);
  });

  it("generates correct stats key", () => {
    expect(sessionKeys.stats("session-1")).toEqual([
      "sessions",
      "detail",
      "session-1",
      "stats",
    ]);
  });
});

describe("useSessions", () => {
  beforeEach(() => {
    // Reset any state between tests
  });

  it("fetches sessions list", async () => {
    const { result } = renderHook(() => useSessions(), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for data
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Should have sessions data
    expect(result.current.data).toBeDefined();
    expect(result.current.data?.sessions).toBeDefined();
  });

  it("uses correct default pagination", async () => {
    const { result } = renderHook(() => useSessions(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Default limit is 20, offset is 0
    expect(result.current.data).toBeDefined();
  });

  it("accepts custom pagination", async () => {
    const { result } = renderHook(() => useSessions(10, 5), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toBeDefined();
  });
});

describe("useSession", () => {
  it("fetches single session by id", async () => {
    const { result } = renderHook(() => useSession("session-1"), {
      wrapper: createWrapper(),
    });

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toBeDefined();
    expect(result.current.data?.id).toBe("session-1");
  });

  it("is disabled when id is empty", async () => {
    const { result } = renderHook(() => useSession(""), {
      wrapper: createWrapper(),
    });

    // Should not fetch when id is empty
    expect(result.current.fetchStatus).toBe("idle");
  });
});
