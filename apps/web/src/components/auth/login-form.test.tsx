import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@/test/utils";
import userEvent from "@testing-library/user-event";
import { LoginForm } from "./login-form";
import { useAuthStore } from "@/stores/auth-store";

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe("LoginForm", () => {
  beforeEach(() => {
    // Reset auth store
    useAuthStore.getState().logout();
  });

  it("renders login form with all fields", () => {
    render(<LoginForm />);

    expect(screen.getByText("MindHit 계정으로 로그인하세요")).toBeInTheDocument();
    expect(screen.getByLabelText("이메일")).toBeInTheDocument();
    expect(screen.getByLabelText("비밀번호")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "로그인" })).toBeInTheDocument();
  });

  it("shows validation errors for empty fields", async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    const submitButton = screen.getByRole("button", { name: "로그인" });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("유효한 이메일을 입력하세요")).toBeInTheDocument();
    });
  });

  it("shows validation error for short password", async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    const emailInput = screen.getByLabelText("이메일");
    const passwordInput = screen.getByLabelText("비밀번호");
    const submitButton = screen.getByRole("button", { name: "로그인" });

    await user.type(emailInput, "test@example.com");
    await user.type(passwordInput, "short");
    await user.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("비밀번호는 8자 이상이어야 합니다")
      ).toBeInTheDocument();
    });
  });

  it("submits form with valid data and updates auth store", async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    const emailInput = screen.getByLabelText("이메일");
    const passwordInput = screen.getByLabelText("비밀번호");
    const submitButton = screen.getByRole("button", { name: "로그인" });

    await user.type(emailInput, "test@example.com");
    await user.type(passwordInput, "password123");
    await user.click(submitButton);

    // After successful login, auth store should be updated
    await waitFor(
      () => {
        const state = useAuthStore.getState();
        expect(state.isAuthenticated).toBe(true);
        expect(state.user?.email).toBe("test@example.com");
      },
      { timeout: 3000 }
    );
  });

  it("clears field error when user types", async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    const emailInput = screen.getByLabelText("이메일");
    const submitButton = screen.getByRole("button", { name: "로그인" });

    // Submit empty form to trigger validation
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("유효한 이메일을 입력하세요")).toBeInTheDocument();
    });

    // Type in email field
    await user.type(emailInput, "test@example.com");

    // Error should be cleared
    await waitFor(() => {
      expect(
        screen.queryByText("유효한 이메일을 입력하세요")
      ).not.toBeInTheDocument();
    });
  });
});
