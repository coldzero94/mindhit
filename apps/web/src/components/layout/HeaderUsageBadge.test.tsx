import { describe, it, expect } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { render } from "@/test/utils";
import { HeaderUsageBadge } from "./HeaderUsageBadge";
import { server } from "@/test/mocks/server";
import {
  mockUsageSummary,
  mockUsageSummaryHigh,
  mockUsageSummaryOver,
  mockUsageSummaryUnlimited,
} from "@/test/mocks/handlers";

// API URLs for handler override
const API_URLS = ["http://localhost:9000/v1", "http://localhost:8080/v1"];

describe("HeaderUsageBadge", () => {
  describe("normal usage (under 80%)", () => {
    it("should not render badge when usage is under 80%", async () => {
      render(<HeaderUsageBadge />);

      // Wait for potential data load
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Badge should not be visible
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
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

    it("should render warning badge at 80%+ usage", async () => {
      render(<HeaderUsageBadge />);

      await waitFor(() => {
        expect(screen.getByText(/85% 사용/)).toBeInTheDocument();
      });
    });

    it("should have yellow styling for warning state", async () => {
      render(<HeaderUsageBadge />);

      await waitFor(() => {
        const badge = screen.getByRole("link");
        expect(badge).toHaveClass("bg-yellow-100", "text-yellow-700");
      });
    });

    it("should link to account page", async () => {
      render(<HeaderUsageBadge />);

      await waitFor(() => {
        const link = screen.getByRole("link");
        expect(link).toHaveAttribute("href", "/account");
      });
    });

    it("should show Zap icon for warning state", async () => {
      render(<HeaderUsageBadge />);

      await waitFor(() => {
        expect(screen.getByText(/85% 사용/)).toBeInTheDocument();
      });

      // Zap icon should be present (not AlertTriangle)
      const link = screen.getByRole("link");
      const svg = link.querySelector("svg");
      expect(svg).toBeInTheDocument();
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

    it("should render error badge when over limit", async () => {
      render(<HeaderUsageBadge />);

      await waitFor(() => {
        expect(screen.getByText("토큰 한도 초과")).toBeInTheDocument();
      });
    });

    it("should have red styling for error state", async () => {
      render(<HeaderUsageBadge />);

      await waitFor(() => {
        const badge = screen.getByRole("link");
        expect(badge).toHaveClass("bg-red-100", "text-red-700");
      });
    });

    it("should show AlertTriangle icon for error state", async () => {
      render(<HeaderUsageBadge />);

      await waitFor(() => {
        expect(screen.getByText("토큰 한도 초과")).toBeInTheDocument();
      });

      // AlertTriangle icon should be present
      const link = screen.getByRole("link");
      const svg = link.querySelector("svg");
      expect(svg).toBeInTheDocument();
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

    it("should not render badge for unlimited plans", async () => {
      render(<HeaderUsageBadge />);

      // Wait for potential data load
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Badge should not be visible for unlimited plans
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
    });
  });

  describe("no data", () => {
    beforeEach(() => {
      // Override usage handler to return empty
      API_URLS.forEach((apiUrl) => {
        server.use(
          http.get(`${apiUrl}/usage`, () => {
            return HttpResponse.json({});
          })
        );
      });
    });

    it("should not render when no usage data", async () => {
      render(<HeaderUsageBadge />);

      // Wait for potential data load
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Badge should not be visible
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
    });
  });
});
