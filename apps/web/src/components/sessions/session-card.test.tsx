import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@/test/utils";
import { SessionCard } from "./session-card";
import type { SessionSession } from "@/api/generated/types.gen";

// Mock next/link
vi.mock("next/link", () => ({
  default: ({
    children,
    href,
  }: {
    children: React.ReactNode;
    href: string;
  }) => <a href={href}>{children}</a>,
}));

const createMockSession = (
  overrides: Partial<SessionSession> = {}
): SessionSession => ({
  id: "session-1",
  title: "Test Session",
  description: undefined,
  session_status: "completed",
  started_at: "2025-01-01T10:00:00Z",
  ended_at: "2025-01-01T12:00:00Z",
  created_at: "2025-01-01T10:00:00Z",
  updated_at: "2025-01-01T12:00:00Z",
  ...overrides,
});

describe("SessionCard", () => {
  it("renders session with title", () => {
    const session = createMockSession({ title: "My Test Session" });
    render(<SessionCard session={session} />);

    expect(screen.getByText("My Test Session")).toBeInTheDocument();
  });

  it("renders 'Untitled' when title is undefined", () => {
    const session = createMockSession({ title: undefined });
    render(<SessionCard session={session} />);

    expect(screen.getByText("Untitled")).toBeInTheDocument();
  });

  it("renders session description when provided", () => {
    const session = createMockSession({ description: "This is a test description" });
    render(<SessionCard session={session} />);

    expect(screen.getByText("This is a test description")).toBeInTheDocument();
  });

  it("does not render description when undefined", () => {
    const session = createMockSession({ description: undefined });
    render(<SessionCard session={session} />);

    expect(screen.queryByText("This is a test description")).not.toBeInTheDocument();
  });

  it("links to session detail page", () => {
    const session = createMockSession({ id: "session-123" });
    render(<SessionCard session={session} />);

    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/sessions/session-123");
  });

  it.each([
    ["recording", "Recording"],
    ["paused", "Paused"],
    ["processing", "Processing"],
    ["completed", "Completed"],
    ["failed", "Failed"],
  ] as const)("renders correct badge for %s status", (status, label) => {
    const session = createMockSession({ session_status: status });
    render(<SessionCard session={session} />);

    expect(screen.getByText(label)).toBeInTheDocument();
  });

  it("renders relative time", () => {
    // Mock current date to make test deterministic
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-01-01T11:00:00Z"));

    const session = createMockSession({
      started_at: "2025-01-01T10:00:00Z",
    });
    render(<SessionCard session={session} />);

    // Should show "about 1 hour ago" or similar
    expect(screen.getByText(/ago$/)).toBeInTheDocument();

    vi.useRealTimers();
  });
});
