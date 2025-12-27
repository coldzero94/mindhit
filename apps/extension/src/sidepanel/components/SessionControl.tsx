import { useState } from "react";
import { useSessionStore } from "@/stores/session-store";
import { useAuthStore } from "@/stores/auth-store";
import { api } from "@/lib/api";

export function SessionControl() {
  const {
    status,
    sessionId,
    startSession,
    pauseSession,
    resumeSession,
    stopSession,
  } = useSessionStore();
  const { token } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);

  const handleStart = async () => {
    if (!token) return;
    setIsLoading(true);
    try {
      const response = await api.startSession(token);
      startSession(response.session.id);

      // Notify Background about session start
      chrome.runtime.sendMessage({
        type: "SESSION_STARTED",
        sessionId: response.session.id,
      });
    } catch (error) {
      console.error("Failed to start session:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handlePause = async () => {
    if (!token || !sessionId) return;
    setIsLoading(true);
    try {
      await api.pauseSession(token, sessionId);
      pauseSession();
      chrome.runtime.sendMessage({ type: "SESSION_PAUSED" });
    } catch (error) {
      console.error("Failed to pause session:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleResume = async () => {
    if (!token || !sessionId) return;
    setIsLoading(true);
    try {
      await api.resumeSession(token, sessionId);
      resumeSession();
      chrome.runtime.sendMessage({ type: "SESSION_RESUMED" });
    } catch (error) {
      console.error("Failed to resume session:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleStop = async () => {
    if (!token || !sessionId) return;
    setIsLoading(true);
    try {
      await api.stopSession(token, sessionId);
      stopSession();
      chrome.runtime.sendMessage({ type: "SESSION_STOPPED" });
    } catch (error) {
      console.error("Failed to stop session:", error);
    } finally {
      setIsLoading(false);
    }
  };

  if (status === "idle") {
    return (
      <div className="bg-white rounded-xl p-6 shadow-sm">
        <div className="text-center space-y-4">
          <div className="w-16 h-16 mx-auto bg-blue-100 rounded-full flex items-center justify-center">
            <svg
              className="w-8 h-8 text-blue-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
              />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
          <div>
            <h2 className="font-semibold text-gray-900">Start New Session</h2>
            <p className="text-sm text-gray-500 mt-1">
              Record your browsing activity and generate mindmaps
            </p>
          </div>
          <button
            onClick={handleStart}
            disabled={isLoading}
            className="btn btn-primary w-full"
          >
            {isLoading ? "Starting..." : "Start Recording"}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl p-4 shadow-sm">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <div
            className={`w-3 h-3 rounded-full ${
              status === "recording"
                ? "bg-red-500 animate-pulse"
                : "bg-yellow-500"
            }`}
          />
          <span className="font-medium text-gray-900">
            {status === "recording" ? "Recording" : "Paused"}
          </span>
        </div>
      </div>

      <div className="flex gap-2">
        {status === "recording" ? (
          <button
            onClick={handlePause}
            disabled={isLoading}
            className="btn btn-secondary flex-1"
          >
            Pause
          </button>
        ) : (
          <button
            onClick={handleResume}
            disabled={isLoading}
            className="btn btn-primary flex-1"
          >
            Resume
          </button>
        )}
        <button
          onClick={handleStop}
          disabled={isLoading}
          className="btn btn-danger flex-1"
        >
          Stop
        </button>
      </div>
    </div>
  );
}
