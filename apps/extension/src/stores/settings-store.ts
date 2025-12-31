import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { chromeStorage } from "@/lib/chrome-storage";
import { STORAGE_KEYS } from "@/lib/constants";

interface Settings {
  // Developer settings
  apiUrl: string;
  webUrl: string;
  // User settings
  autoStartSession: boolean;
  collectAllTabs: boolean; // true: all tabs, false: current tab only
}

interface SettingsState extends Settings {
  updateSettings: (settings: Partial<Settings>) => void;
  resetSettings: () => void;
}

const defaultSettings: Settings = {
  apiUrl: import.meta.env.VITE_API_URL || "http://localhost:9000/v1",
  webUrl: import.meta.env.VITE_WEB_URL || "http://localhost:3000",
  autoStartSession: false,
  collectAllTabs: true,
};

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      ...defaultSettings,

      updateSettings: (newSettings) =>
        set((state) => ({ ...state, ...newSettings })),

      resetSettings: () => set(defaultSettings),
    }),
    {
      name: STORAGE_KEYS.SETTINGS,
      storage: createJSONStorage(() => chromeStorage),
    }
  )
);
