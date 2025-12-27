import { create } from "zustand";
import { persist, createJSONStorage, StateStorage } from "zustand/middleware";
import type { User } from "@/types";

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
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

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      setAuth: (user, token) => set({ user, token, isAuthenticated: true }),
      logout: () => set({ user: null, token: null, isAuthenticated: false }),
    }),
    {
      name: "mindhit-auth",
      storage: createJSONStorage(() => chromeStorage),
    }
  )
);
