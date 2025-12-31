import { useState, useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { api } from "@/lib/api";
import { WEB_APP_URL } from "@/lib/constants";
import type { Session } from "@/types";

interface SessionListProps {
  currentSessionId?: string;
  onClose: () => void;
}

export function SessionList({ currentSessionId, onClose }: SessionListProps) {
  const { token } = useAuthStore();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSessions = async () => {
      if (!token) return;

      try {
        setIsLoading(true);
        setError(null);
        const response = await api.getSessions(token, 10);
        setSessions(response.sessions);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load sessions");
      } finally {
        setIsLoading(false);
      }
    };

    fetchSessions();
  }, [token]);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("ko-KR", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getStatusBadge = (status: Session["session_status"]) => {
    const styles: Record<string, string> = {
      recording: "bg-green-100 text-green-700",
      paused: "bg-yellow-100 text-yellow-700",
      processing: "bg-blue-100 text-blue-700",
      completed: "bg-gray-100 text-gray-700",
      failed: "bg-red-100 text-red-700",
    };

    const labels: Record<string, string> = {
      recording: "Recording",
      paused: "Paused",
      processing: "Processing",
      completed: "Completed",
      failed: "Failed",
    };

    return (
      <span
        className={`px-1.5 py-0.5 text-[10px] font-medium rounded ${styles[status] || styles.completed}`}
      >
        {labels[status] || status}
      </span>
    );
  };

  const openSessionInWeb = (sessionId: string) => {
    chrome.tabs.create({ url: `${WEB_APP_URL}/sessions/${sessionId}` });
  };

  return (
    <div className="space-y-2">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-gray-900">Recent Sessions</h2>
        <button
          onClick={onClose}
          className="text-xs text-gray-400 hover:text-gray-600"
        >
          Close
        </button>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="py-4 text-center text-xs text-gray-500">Loading...</div>
      )}

      {/* Error */}
      {error && (
        <div className="py-4 text-center text-xs text-red-500">{error}</div>
      )}

      {/* Session List */}
      {!isLoading && !error && sessions.length === 0 && (
        <div className="py-4 text-center text-xs text-gray-500">
          No sessions yet
        </div>
      )}

      {!isLoading && !error && sessions.length > 0 && (
        <div className="space-y-1.5 max-h-[200px] overflow-y-auto">
          {sessions.map((session) => (
            <button
              key={session.id}
              onClick={() => openSessionInWeb(session.id)}
              className={`w-full p-2 text-left rounded-lg border transition-colors ${
                session.id === currentSessionId
                  ? "border-blue-300 bg-blue-50"
                  : "border-gray-200 hover:border-gray-300 hover:bg-gray-50"
              }`}
            >
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-medium text-gray-900 truncate">
                    {session.title || `Session #${session.id.slice(0, 8)}`}
                  </p>
                  <p className="text-[10px] text-gray-500 mt-0.5">
                    {formatDate(session.started_at)}
                  </p>
                </div>
                {getStatusBadge(session.session_status)}
              </div>
            </button>
          ))}
        </div>
      )}

      {/* View All Button */}
      <button
        onClick={() => chrome.tabs.create({ url: `${WEB_APP_URL}/sessions` })}
        className="w-full py-2 text-xs text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded-lg transition-colors"
      >
        View all in dashboard →
      </button>
    </div>
  );
}
