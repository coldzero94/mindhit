import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@/test/utils";
import userEvent from "@testing-library/user-event";
import { SignupForm } from "./signup-form";
import { useAuthStore } from "@/stores/auth-store";

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe("SignupForm", () => {
  beforeEach(() => {
    // Reset auth store
    useAuthStore.getState().logout();
  });

  it("renders signup form with all fields", () => {
    render(<SignupForm />);

    expect(screen.getByText("MindHit 계정을 만드세요")).toBeInTheDocument();
    expect(screen.getByLabelText("이메일")).toBeInTheDocument();
    expect(screen.getByLabelText("비밀번호")).toBeInTheDocument();
    expect(screen.getByLabelText("비밀번호 확인")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "회원가입" })).toBeInTheDocument();
  });

  it("shows validation error for short password", async () => {
    const user = userEvent.setup();
    render(<SignupForm />);

    const emailInput = screen.getByLabelText("이메일");
    const passwordInput = screen.getByLabelText("비밀번호");
    const confirmInput = screen.getByLabelText("비밀번호 확인");
    const submitButton = screen.getByRole("button", { name: "회원가입" });

    await user.type(emailInput, "test@example.com");
    await user.type(passwordInput, "short");
    await user.type(confirmInput, "short");
    await user.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("비밀번호는 8자 이상이어야 합니다")
      ).toBeInTheDocument();
    });
  });

  it("shows validation error when passwords do not match", async () => {
    const user = userEvent.setup();
    render(<SignupForm />);

    const emailInput = screen.getByLabelText("이메일");
    const passwordInput = screen.getByLabelText("비밀번호");
    const confirmInput = screen.getByLabelText("비밀번호 확인");
    const submitButton = screen.getByRole("button", { name: "회원가입" });

    await user.type(emailInput, "test@example.com");
    await user.type(passwordInput, "password123");
    await user.type(confirmInput, "differentpassword");
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("비밀번호가 일치하지 않습니다")).toBeInTheDocument();
    });
  });

  it("submits form with valid data and updates auth store", async () => {
    const user = userEvent.setup();
    render(<SignupForm />);

    const emailInput = screen.getByLabelText("이메일");
    const passwordInput = screen.getByLabelText("비밀번호");
    const confirmInput = screen.getByLabelText("비밀번호 확인");
    const submitButton = screen.getByRole("button", { name: "회원가입" });

    await user.type(emailInput, "newuser@example.com");
    await user.type(passwordInput, "password123");
    await user.type(confirmInput, "password123");
    await user.click(submitButton);

    // After successful signup, auth store should be updated
    await waitFor(
      () => {
        const state = useAuthStore.getState();
        expect(state.isAuthenticated).toBe(true);
        expect(state.user?.email).toBe("newuser@example.com");
      },
      { timeout: 3000 }
    );
  });

  it("clears field error when user types", async () => {
    const user = userEvent.setup();
    render(<SignupForm />);

    const emailInput = screen.getByLabelText("이메일");
    const passwordInput = screen.getByLabelText("비밀번호");
    const confirmInput = screen.getByLabelText("비밀번호 확인");
    const submitButton = screen.getByRole("button", { name: "회원가입" });

    // Submit with mismatched passwords
    await user.type(emailInput, "test@example.com");
    await user.type(passwordInput, "password123");
    await user.type(confirmInput, "different");
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("비밀번호가 일치하지 않습니다")).toBeInTheDocument();
    });

    // Clear and retype confirm password
    await user.clear(confirmInput);
    await user.type(confirmInput, "password123");

    // Error should be cleared
    await waitFor(() => {
      expect(
        screen.queryByText("비밀번호가 일치하지 않습니다")
      ).not.toBeInTheDocument();
    });
  });
});
