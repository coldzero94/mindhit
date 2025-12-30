import { useState } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { api } from "@/lib/api";

export function LoginPrompt() {
  const { setAuth } = useAuthStore();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError("");

    try {
      const response = await api.login(email, password);
      setAuth(response.user, response.token);
    } catch {
      setError("Login failed. Please check your email and password.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="p-4">
      <div className="text-center mb-4">
        <h1 className="text-lg font-bold text-gray-900">MindHit</h1>
        <p className="text-xs text-gray-500 mt-1">Sign in to get started</p>
      </div>

      <form onSubmit={handleLogin} className="space-y-3">
        <div>
          <label className="block text-xs font-medium text-gray-700 mb-1">
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="input text-sm py-1.5"
            placeholder="email@example.com"
            required
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-700 mb-1">
            Password
          </label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="input text-sm py-1.5"
            required
          />
        </div>

        {error && <p className="text-xs text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={isLoading}
          className="btn btn-primary w-full text-sm py-2"
        >
          {isLoading ? "Signing in..." : "Sign In"}
        </button>
      </form>

      <p className="text-center text-xs text-gray-500 mt-3">
        Don&apos;t have an account?{" "}
        <a
          href="http://localhost:3000/signup"
          target="_blank"
          rel="noopener noreferrer"
          className="text-blue-600 hover:underline"
        >
          Sign up on web
        </a>
      </p>
    </div>
  );
}
