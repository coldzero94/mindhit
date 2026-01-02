import { describe, it, expect } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { render } from "@/test/utils";
import { UsageCard } from "./UsageCard";
import { server } from "@/test/mocks/server";
import {
  mockUsageSummary,
  mockUsageSummaryHigh,
  mockUsageSummaryOver,
  mockUsageSummaryUnlimited,
} from "@/test/mocks/handlers";

// API URLs for handler override
const API_URLS = ["http://localhost:9000/v1", "http://localhost:8080/v1"];

describe("UsageCard", () => {
  describe("loading state", () => {
    it("should show skeleton while loading", () => {
      render(<UsageCard />);

      // Should show skeleton elements
      const skeletons = document.querySelectorAll('[class*="animate-pulse"], [data-slot="skeleton"]');
      expect(skeletons.length).toBeGreaterThan(0);
    });
  });

  describe("normal usage (under 80%)", () => {
    it("should display tokens used", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        // mockUsageSummary.tokens_used is 3500, should display as 3.5K
        expect(screen.getByText("3.5K")).toBeInTheDocument();
      });
    });

    it("should display token limit", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        // mockUsageSummary.token_limit is 10000
        expect(screen.getByText(/10.0K/)).toBeInTheDocument();
      });
    });

    it("should display usage percentage", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        expect(screen.getByText(/35.0% 사용 중/)).toBeInTheDocument();
      });
    });

    it("should show blue icon for normal usage", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        const iconContainer = document.querySelector(".bg-blue-100");
        expect(iconContainer).toBeInTheDocument();
      });
    });

    it("should not show warning badge", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        expect(screen.getByText("3.5K")).toBeInTheDocument();
      });

      expect(screen.queryByText("80% 도달")).not.toBeInTheDocument();
      expect(screen.queryByText("한도 초과")).not.toBeInTheDocument();
    });
  });

  describe("high usage (80%+)", () => {
    beforeEach(() => {
      // Override usage handler to return high usage
      API_URLS.forEach((apiUrl) => {
        server.use(
          http.get(`${apiUrl}/usage`, () => {
            return HttpResponse.json({ usage: mockUsageSummaryHigh });
          })
        );
      });
    });

    it("should show warning badge at 80%", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        expect(screen.getByText("80% 도달")).toBeInTheDocument();
      });
    });

    it("should show yellow icon for high usage", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        const iconContainer = document.querySelector(".bg-yellow-100");
        expect(iconContainer).toBeInTheDocument();
      });
    });
  });

  describe("over limit (100%+)", () => {
    beforeEach(() => {
      // Override usage handler to return over limit
      API_URLS.forEach((apiUrl) => {
        server.use(
          http.get(`${apiUrl}/usage`, () => {
            return HttpResponse.json({ usage: mockUsageSummaryOver });
          })
        );
      });
    });

    it("should show over limit badge", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        expect(screen.getByText("한도 초과")).toBeInTheDocument();
      });
    });

    it("should show red badge styling", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        const badge = screen.getByText("한도 초과").closest("div");
        expect(badge).toHaveClass("bg-red-100", "text-red-700");
      });
    });
  });

  describe("unlimited usage", () => {
    beforeEach(() => {
      // Override usage handler to return unlimited
      API_URLS.forEach((apiUrl) => {
        server.use(
          http.get(`${apiUrl}/usage`, () => {
            return HttpResponse.json({ usage: mockUsageSummaryUnlimited });
          })
        );
      });
    });

    it("should show unlimited text instead of limit", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        // Check for "무제한" in the token limit area
        expect(screen.getByText(/\/ 무제한/)).toBeInTheDocument();
      });
    });

    it("should show unlimited usage message", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        expect(screen.getByText("무제한 사용 중")).toBeInTheDocument();
      });
    });

    it("should not show progress bar", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        expect(screen.getByText("무제한 사용 중")).toBeInTheDocument();
      });

      // Progress component should not be rendered
      const progress = document.querySelector('[role="progressbar"]');
      expect(progress).not.toBeInTheDocument();
    });
  });

  describe("number formatting", () => {
    it("should format thousands as K", async () => {
      render(<UsageCard />);

      await waitFor(() => {
        expect(screen.getByText("3.5K")).toBeInTheDocument();
      });
    });
  });
});
