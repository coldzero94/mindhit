import { useState } from "react";
import { useSessionStore } from "@/stores/session-store";
import { useAuthStore } from "@/stores/auth-store";
import { api } from "@/lib/api";
import { SessionTitleInput } from "./SessionTitleInput";

interface ErrorState {
  message: string;
  action: string;
}

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
  const [error, setError] = useState<ErrorState | null>(null);

  const clearError = () => setError(null);

  const handleStart = async () => {
    if (!token) return;
    setIsLoading(true);
    clearError();
    try {
      const response = await api.startSession(token);
      startSession(response.session.id);

      // Notify Background about session start
      chrome.runtime.sendMessage({
        type: "SESSION_STARTED",
        sessionId: response.session.id,
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to start session";
      setError({ message, action: "start" });
    } finally {
      setIsLoading(false);
    }
  };

  const handlePause = async () => {
    if (!token || !sessionId) return;
    setIsLoading(true);
    clearError();
    try {
      await api.pauseSession(token, sessionId);
      pauseSession();
      chrome.runtime.sendMessage({ type: "SESSION_PAUSED" });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to pause session";
      setError({ message, action: "pause" });
    } finally {
      setIsLoading(false);
    }
  };

  const handleResume = async () => {
    if (!token || !sessionId) return;
    setIsLoading(true);
    clearError();
    try {
      await api.resumeSession(token, sessionId);
      resumeSession();
      chrome.runtime.sendMessage({ type: "SESSION_RESUMED" });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to resume session";
      setError({ message, action: "resume" });
    } finally {
      setIsLoading(false);
    }
  };

  const handleStop = async () => {
    if (!token || !sessionId) return;
    setIsLoading(true);
    clearError();
    try {
      await api.stopSession(token, sessionId);
      stopSession();
      chrome.runtime.sendMessage({ type: "SESSION_STOPPED" });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to stop session";
      setError({ message, action: "stop" });
    } finally {
      setIsLoading(false);
    }
  };

  if (status === "idle") {
    return (
      <div className="bg-white rounded-lg p-4 shadow-sm">
        <div className="text-center space-y-3">
          <div className="w-12 h-12 mx-auto bg-blue-100 rounded-full flex items-center justify-center">
            <svg
              className="w-6 h-6 text-blue-600"
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
            <h2 className="font-semibold text-gray-900 text-sm">Start New Session</h2>
            <p className="text-xs text-gray-500 mt-1">
              Record your browsing activity
            </p>
          </div>
          <button
            onClick={handleStart}
            disabled={isLoading}
            className="btn btn-primary w-full text-sm py-2"
          >
            {isLoading ? "Starting..." : "Start Recording"}
          </button>
          {error && (
            <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded-lg">
              <div className="flex items-start gap-2">
                <svg className="w-4 h-4 text-red-500 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
                <div className="flex-1">
                  <p className="text-xs text-red-700">{error.message}</p>
                  <button
                    onClick={clearError}
                    className="text-xs text-red-600 underline mt-1"
                  >
                    Dismiss
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg p-3 shadow-sm">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <div
            className={`w-2.5 h-2.5 rounded-full ${
              status === "recording"
                ? "bg-red-500 animate-pulse"
                : "bg-yellow-500"
            }`}
          />
          <span className="font-medium text-gray-900 text-sm">
            {status === "recording" ? "Recording" : "Paused"}
          </span>
        </div>
      </div>

      {sessionId && (
        <SessionTitleInput sessionId={sessionId} className="mb-2" />
      )}

      {error && (
        <div className="mb-2 p-2 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-start gap-2">
            <svg className="w-4 h-4 text-red-500 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
            <div className="flex-1">
              <p className="text-xs text-red-700">{error.message}</p>
              <button
                onClick={clearError}
                className="text-xs text-red-600 underline mt-1"
              >
                Dismiss
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="flex gap-2">
        {status === "recording" ? (
          <button
            onClick={handlePause}
            disabled={isLoading}
            className="btn btn-secondary flex-1 text-sm py-1.5"
          >
            Pause
          </button>
        ) : (
          <button
            onClick={handleResume}
            disabled={isLoading}
            className="btn btn-primary flex-1 text-sm py-1.5"
          >
            Resume
          </button>
        )}
        <button
          onClick={handleStop}
          disabled={isLoading}
          className="btn btn-danger flex-1 text-sm py-1.5"
        >
          Stop
        </button>
      </div>
    </div>
  );
}
