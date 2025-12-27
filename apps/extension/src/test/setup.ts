import { vi, beforeAll, afterAll, afterEach } from "vitest";
import { server } from "./mocks/server";

// Setup MSW server
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Mock chrome API
const mockChrome = {
  storage: {
    local: {
      get: vi.fn().mockResolvedValue({}),
      set: vi.fn().mockResolvedValue(undefined),
      remove: vi.fn().mockResolvedValue(undefined),
    },
  },
  runtime: {
    sendMessage: vi.fn(),
    onMessage: {
      addListener: vi.fn(),
    },
  },
  tabs: {
    query: vi.fn().mockResolvedValue([]),
    sendMessage: vi.fn().mockResolvedValue(undefined),
    onUpdated: {
      addListener: vi.fn(),
    },
  },
  sidePanel: {
    setPanelBehavior: vi.fn(),
  },
};

// @ts-expect-error - Mocking chrome API for tests
globalThis.chrome = mockChrome;
