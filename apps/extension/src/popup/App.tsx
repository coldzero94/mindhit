import { useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useSessionStore } from "@/stores/session-store";
import { SessionControl } from "./components/SessionControl";
import { SessionStats } from "./components/SessionStats";
import { LoginPrompt } from "./components/LoginPrompt";

export function App() {
  const { isAuthenticated, user, logout } = useAuthStore();
  const { status, updateElapsedTime, incrementPageCount, incrementHighlightCount } = useSessionStore();

  // Update elapsed time
  useEffect(() => {
    if (status === "recording") {
      const interval = setInterval(updateElapsedTime, 1000);
      return () => clearInterval(interval);
    }
  }, [status, updateElapsedTime]);

  // Listen for page count and highlight count updates from content scripts
  useEffect(() => {
    const handleMessage = (message: { type: string }) => {
      if (message.type === "INCREMENT_PAGE_COUNT") {
        incrementPageCount();
      } else if (message.type === "INCREMENT_HIGHLIGHT_COUNT") {
        incrementHighlightCount();
      }
    };

    chrome.runtime.onMessage.addListener(handleMessage);
    return () => chrome.runtime.onMessage.removeListener(handleMessage);
  }, [incrementPageCount, incrementHighlightCount]);

  if (!isAuthenticated) {
    return <LoginPrompt />;
  }

  return (
    <div className="p-3 space-y-3">
      {/* Header */}
      <header className="flex items-center justify-between">
        <h1 className="text-base font-bold text-gray-900">MindHit</h1>
        <div className="flex items-center gap-2">
          <span className="text-xs text-gray-500 truncate max-w-[120px]">{user?.email}</span>
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
