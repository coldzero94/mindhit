import { useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useSessionStore } from "@/stores/session-store";
import { SessionControl } from "./components/SessionControl";
import { SessionStats } from "./components/SessionStats";
import { LoginPrompt } from "./components/LoginPrompt";

export function App() {
  const { isAuthenticated, user, logout } = useAuthStore();
  const { status, updateElapsedTime } = useSessionStore();

  // Update elapsed time
  useEffect(() => {
    if (status === "recording") {
      const interval = setInterval(updateElapsedTime, 1000);
      return () => clearInterval(interval);
    }
  }, [status, updateElapsedTime]);

  if (!isAuthenticated) {
    return <LoginPrompt />;
  }

  return (
    <div className="min-h-screen p-4 space-y-4">
      {/* Header */}
      <header className="flex items-center justify-between">
        <h1 className="text-lg font-bold text-gray-900">MindHit</h1>
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500">{user?.email}</span>
          <button
            onClick={logout}
            className="text-xs text-gray-400 hover:text-gray-600"
          >
            Logout
          </button>
        </div>
      </header>

      {/* Session Control */}
      <SessionControl />

      {/* Session Stats */}
      {status !== "idle" && <SessionStats />}
    </div>
  );
}
