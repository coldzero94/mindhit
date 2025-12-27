import { vi } from "vitest";

/**
 * Create mock router functions for testing
 */
export function createMockRouter() {
  return {
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
  };
}

/**
 * Mock pathname for testing
 */
export function mockPathname(pathname: string) {
  vi.doMock("next/navigation", () => ({
    useRouter: () => createMockRouter(),
    usePathname: () => pathname,
    useParams: () => ({}),
    useSearchParams: () => new URLSearchParams(),
  }));
}

/**
 * Mock params for dynamic routes
 */
export function mockParams(params: Record<string, string>) {
  vi.doMock("next/navigation", () => ({
    useRouter: () => createMockRouter(),
    usePathname: () => "/",
    useParams: () => params,
    useSearchParams: () => new URLSearchParams(),
  }));
}
