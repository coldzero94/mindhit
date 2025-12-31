import { useState, useRef, useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useSessionStore } from "@/stores/session-store";
import { api } from "@/lib/api";

interface SessionTitleInputProps {
  sessionId: string;
  className?: string;
}

export function SessionTitleInput({ sessionId, className }: SessionTitleInputProps) {
  const { token } = useAuthStore();
  const { sessionTitle, setSessionTitle } = useSessionStore();
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(sessionTitle || "");
  const [isSaving, setIsSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  // Update edit value when sessionTitle changes
  useEffect(() => {
    setEditValue(sessionTitle || "");
  }, [sessionTitle]);

  // Focus input when editing starts
  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleSave = async () => {
    if (!token || !sessionId) return;

    const trimmedValue = editValue.trim();
    if (trimmedValue === (sessionTitle || "")) {
      setIsEditing(false);
      return;
    }

    setIsSaving(true);
    try {
      await api.updateSession(token, sessionId, { title: trimmedValue || undefined });
      setSessionTitle(trimmedValue || null);
      setIsEditing(false);
    } catch (err) {
      console.error("Failed to update session title:", err);
      // Revert on error
      setEditValue(sessionTitle || "");
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    setEditValue(sessionTitle || "");
    setIsEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSave();
    } else if (e.key === "Escape") {
      handleCancel();
    }
  };

  const displayTitle = sessionTitle || `Session #${sessionId.slice(0, 8)}`;

  if (isEditing) {
    return (
      <div className={`flex items-center gap-1 ${className || ""}`}>
        <input
          ref={inputRef}
          type="text"
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Enter session title..."
          className="flex-1 px-2 py-1 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
          disabled={isSaving}
        />
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="p-1 text-green-600 hover:bg-green-50 rounded disabled:opacity-50"
          title="Save"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        </button>
        <button
          onClick={handleCancel}
          disabled={isSaving}
          className="p-1 text-gray-400 hover:bg-gray-100 rounded disabled:opacity-50"
          title="Cancel"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    );
  }

  return (
    <div className={className}>
      <button
        onClick={() => setIsEditing(true)}
        className="group flex items-center gap-1 text-left max-w-full"
        title="Click to edit title"
      >
        <span className="text-xs font-medium text-gray-700 truncate">
          {displayTitle}
        </span>
        <svg
          className="w-3 h-3 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
          />
        </svg>
      </button>
    </div>
  );
}
