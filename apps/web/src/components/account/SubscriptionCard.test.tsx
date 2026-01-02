import { describe, it, expect, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { render } from "@/test/utils";
import { SubscriptionCard } from "./SubscriptionCard";
import { mockSubscription, mockPlan } from "@/test/mocks/handlers";

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
        // mockPlan.token_limit is 10000
        expect(screen.getByText(/10,000 토큰/)).toBeInTheDocument();
      });
    });

    it("should show upgrade button for free plan", async () => {
      const onUpgrade = vi.fn();
      render(<SubscriptionCard onUpgrade={onUpgrade} />);

      await waitFor(() => {
        expect(screen.getByText("업그레이드")).toBeInTheDocument();
      });
    });

    it("should call onUpgrade when upgrade button is clicked", async () => {
      const user = userEvent.setup();
      const onUpgrade = vi.fn();
      render(<SubscriptionCard onUpgrade={onUpgrade} />);

      await waitFor(() => {
        expect(screen.getByText("업그레이드")).toBeInTheDocument();
      });

      await user.click(screen.getByText("업그레이드"));
      expect(onUpgrade).toHaveBeenCalledTimes(1);
    });
  });

  describe("status badges", () => {
    it("should display correct badge for active status", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        const badge = screen.getByText("활성");
        expect(badge).toHaveClass("bg-green-100", "text-green-700");
      });
    });
  });

  describe("plan icon", () => {
    it("should show gray icon for free plan", async () => {
      render(<SubscriptionCard />);

      await waitFor(() => {
        expect(screen.getByText(/Free 플랜/)).toBeInTheDocument();
      });

      // Free plan should have gray background
      const iconContainer = document.querySelector(".bg-gray-100");
      expect(iconContainer).toBeInTheDocument();
    });
  });
});
