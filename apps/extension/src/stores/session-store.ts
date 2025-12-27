import { create } from "zustand";
import { persist, createJSONStorage, StateStorage } from "zustand/middleware";

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

// Chrome storage adapter for Zustand persist
const chromeStorage: StateStorage = {
  getItem: async (name: string): Promise<string | null> => {
    const result = await chrome.storage.local.get(name);
    const value = result[name];
    return typeof value === "string" ? value : null;
  },
  setItem: async (name: string, value: string): Promise<void> => {
    await chrome.storage.local.set({ [name]: value });
  },
  removeItem: async (name: string): Promise<void> => {
    await chrome.storage.local.remove(name);
  },
};

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
      name: "mindhit-session",
      storage: createJSONStorage(() => chromeStorage),
    }
  )
);
