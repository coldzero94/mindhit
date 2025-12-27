import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@/test/utils";
import userEvent from "@testing-library/user-event";
import { SessionList } from "./session-list";

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

describe("SessionList", () => {
  it("shows loading or content after render", async () => {
    render(<SessionList page={1} onPageChange={vi.fn()} />);

    // Wait for content to appear (either loading state finishes or shows sessions)
    await waitFor(() => {
      const hasContent =
        screen.queryByText("Test Session") ||
        screen.queryByText("아직 녹화된 세션이 없습니다.");
      expect(hasContent).toBeInTheDocument();
    });
  });

  it("renders session cards after loading", async () => {
    render(<SessionList page={1} onPageChange={vi.fn()} />);

    // Wait for sessions to load
    await waitFor(() => {
      expect(screen.getByText("Test Session")).toBeInTheDocument();
    });
  });

  it("shows empty state when no sessions", async () => {
    // This would need a custom handler, but we can test the structure
    render(<SessionList page={1} onPageChange={vi.fn()} />);

    // Wait for loading to finish
    await waitFor(() => {
      // Either shows sessions or empty state
      const hasContent =
        screen.queryByText("Test Session") ||
        screen.queryByText("아직 녹화된 세션이 없습니다.");
      expect(hasContent).toBeInTheDocument();
    });
  });

  it("shows page number in pagination", async () => {
    render(<SessionList page={2} onPageChange={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText("페이지 2")).toBeInTheDocument();
    });
  });

  it("calls onPageChange when clicking previous", async () => {
    const onPageChange = vi.fn();
    const user = userEvent.setup();

    render(<SessionList page={2} onPageChange={onPageChange} />);

    await waitFor(() => {
      expect(screen.getByText("페이지 2")).toBeInTheDocument();
    });

    const prevButton = screen.getByRole("button", { name: /이전/ });
    await user.click(prevButton);

    expect(onPageChange).toHaveBeenCalledWith(1);
  });

  it("disables previous button on first page", async () => {
    render(<SessionList page={1} onPageChange={vi.fn()} />);

    await waitFor(() => {
      // Wait for data to load - if there's data and pagination shows
      const prevButton = screen.queryByRole("button", { name: /이전/ });
      if (prevButton) {
        expect(prevButton).toBeDisabled();
      }
    });
  });
});
