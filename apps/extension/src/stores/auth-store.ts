import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import type { User } from "@/types";
import { chromeSessionStorage } from "@/lib/chrome-storage";
import { STORAGE_KEYS } from "@/lib/constants";

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
}

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
      name: STORAGE_KEYS.AUTH,
      // Use session storage for auth data (cleared when browser closes)
      storage: createJSONStorage(() => chromeSessionStorage),
    }
  )
);
