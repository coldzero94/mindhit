import { useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useSessionStore } from "@/stores/session-store";
import { SessionControl } from "./components/SessionControl";
import { SessionStats } from "./components/SessionStats";
import { LoginPrompt } from "./components/LoginPrompt";
import { SessionList } from "./components/SessionList";
import { DashboardLink } from "./components/DashboardLink";
import { NetworkBanner } from "./components/NetworkBanner";
import { Settings } from "./components/Settings";

export function App() {
  const { isAuthenticated, user, logout } = useAuthStore();
  const { status, sessionId, updateElapsedTime, incrementPageCount, incrementHighlightCount } = useSessionStore();
  const [isHydrated, setIsHydrated] = useState(false);
  const [showSessionList, setShowSessionList] = useState(false);
  const [showSettings, setShowSettings] = useState(false);

  // Wait for Zustand to hydrate from chrome.storage
  useEffect(() => {
    const unsubAuth = useAuthStore.persist.onFinishHydration(() => {
      setIsHydrated(true);
    });

    // Check if already hydrated
    if (useAuthStore.persist.hasHydrated()) {
      setIsHydrated(true);
    }

    return () => {
      unsubAuth();
    };
  }, []);

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

  // Show loading while hydrating from storage
  if (!isHydrated) {
    return (
      <div className="p-4 flex items-center justify-center min-h-[200px]">
        <div className="text-sm text-gray-500">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <LoginPrompt />;
  }

  return (
    <div className="min-w-[320px]">
      {/* Network Banner */}
      <NetworkBanner />

      {/* Settings Modal */}
      {showSettings && <Settings onClose={() => setShowSettings(false)} />}

      <div className="p-3 space-y-3">
        {/* Header */}
        <header className="flex items-center justify-between">
          <h1 className="text-base font-bold text-gray-900">MindHit</h1>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowSettings(true)}
              className="p-1 hover:bg-gray-100 rounded-full"
              title="Settings"
            >
              <svg
                className="w-4 h-4 text-gray-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                />
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
            </button>
            <DashboardLink />
            <span className="text-xs text-gray-500 truncate max-w-[100px]">
              {user?.email}
            </span>
            <button
              onClick={logout}
              className="text-xs text-gray-400 hover:text-gray-600"
            >
              Logout
            </button>
          </div>
        </header>

        {/* Session List Toggle */}
        {showSessionList ? (
          <SessionList
            currentSessionId={sessionId ?? undefined}
            onClose={() => setShowSessionList(false)}
          />
        ) : (
          <>
            {/* Session Control */}
            <SessionControl />

            {/* Session Stats */}
            {status !== "idle" && <SessionStats />}

            {/* View Sessions Button */}
            <button
              onClick={() => setShowSessionList(true)}
              className="w-full py-2 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-50 rounded-lg transition-colors border border-gray-200"
            >
              View recent sessions
            </button>
          </>
        )}
      </div>
    </div>
  );
}
