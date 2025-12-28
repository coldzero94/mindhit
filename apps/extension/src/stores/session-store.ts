import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { chromeStorage } from "@/lib/chrome-storage";
import { STORAGE_KEYS } from "@/lib/constants";

// Extension local state (idle: no session, recording: recording, paused: paused)
// Backend state: recording, paused, processing, completed, failed
export type SessionStatus =
  | "idle"
  | "recording"
  | "paused"
  | "processing"
  | "completed"
  | "failed";

interface SessionState {
  sessionId: string | null;
  status: SessionStatus;
  startedAt: number | null;
  pageCount: number;
  highlightCount: number;
  elapsedSeconds: number;

  startSession: (sessionId: string) => void;
  pauseSession: () => void;
  resumeSession: () => void;
  stopSession: () => void;
  incrementPageCount: () => void;
  incrementHighlightCount: () => void;
  updateElapsedTime: () => void;
  reset: () => void;
}

export const useSessionStore = create<SessionState>()(
  persist(
    (set, get) => ({
      sessionId: null,
      status: "idle",
      startedAt: null,
      pageCount: 0,
      highlightCount: 0,
      elapsedSeconds: 0,

      startSession: (sessionId) =>
        set({
          sessionId,
          status: "recording",
          startedAt: Date.now(),
          pageCount: 0,
          highlightCount: 0,
          elapsedSeconds: 0,
        }),

      pauseSession: () => set({ status: "paused" }),

      resumeSession: () => set({ status: "recording" }),

      stopSession: () =>
        set({
          sessionId: null,
          status: "idle",
          startedAt: null,
        }),

      incrementPageCount: () =>
        set((state) => ({ pageCount: state.pageCount + 1 })),

      incrementHighlightCount: () =>
        set((state) => ({ highlightCount: state.highlightCount + 1 })),

      updateElapsedTime: () => {
        const { startedAt, status } = get();
        if (startedAt && status === "recording") {
          set({ elapsedSeconds: Math.floor((Date.now() - startedAt) / 1000) });
        }
      },

      reset: () =>
        set({
          sessionId: null,
          status: "idle",
          startedAt: null,
          pageCount: 0,
          highlightCount: 0,
          elapsedSeconds: 0,
        }),
    }),
    {
      name: STORAGE_KEYS.SESSION,
      storage: createJSONStorage(() => chromeStorage),
    }
  )
);
