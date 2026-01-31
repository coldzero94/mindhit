import { describe, it, expect } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { render } from "@/test/utils";
import { SubscriptionCard } from "./SubscriptionCard";

describe("SubscriptionCard", () => {
  describe("loading state", () => {
    it("should show skeleton while loading", () => {
      render(<SubscriptionCard />);

      // Should show skeleton elements
      const skeletons = document.querySelectorAll('[class*="animate-pulse"], [data-slot="skeleton"]');
      expect(skeletons.length).toBeGreaterThan(0);
    });
  });

  describe("with data", () => {
    it("should display plan name", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        expect(screen.getByText(/Free 플랜/)).toBeInTheDocument();
      });
    });

    it("should display active status badge", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        expect(screen.getByText("활성")).toBeInTheDocument();
      });
    });

    it("should display subscription period dates", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        // Check for date strings (Korean format)
        expect(screen.getByText(/시작일:/)).toBeInTheDocument();
        expect(screen.getByText(/종료일:/)).toBeInTheDocument();
      });
    });

    it("should display token limit from plan", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        // mockPlan.token_limit is 10000, displayed as "월 10,000 토큰 제공"
        expect(screen.getByText(/월.*10,000.*토큰/)).toBeInTheDocument();
      });
    });

    it("should show plan change button", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        expect(screen.getByText("플랜 변경")).toBeInTheDocument();
      });
    });
  });

  describe("status badges", () => {
    it("should display correct badge for active status", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        const badge = screen.getByText("활성");
        expect(badge).toHaveClass("bg-status-success-bg", "text-status-success-text");
      });
    });
  });

  describe("plan icon", () => {
    it("should show muted icon for free plan", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        expect(screen.getByText(/Free 플랜/)).toBeInTheDocument();
      });

      // Free plan should have muted background
      const iconContainer = document.querySelector(".bg-muted");
      expect(iconContainer).toBeInTheDocument();
    });
  });
});
