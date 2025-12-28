import type { StateStorage } from "zustand/middleware";

/**
 * Chrome storage adapter for Zustand persist.
 * This provides a centralized implementation to avoid duplication.
 */
export const chromeStorage: StateStorage = {
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

/**
 * Session storage adapter for sensitive data (like tokens).
 * Data is cleared when the browser is closed.
 * NOTE: chrome.storage.session requires Manifest V3 and "storage" permission.
 */
export const chromeSessionStorage: StateStorage = {
  getItem: async (name: string): Promise<string | null> => {
    // chrome.storage.session may not be available in all contexts
    if (!chrome.storage.session) {
      return chromeStorage.getItem(name);
    }
    const result = await chrome.storage.session.get(name);
    const value = result[name];
    return typeof value === "string" ? value : null;
  },
  setItem: async (name: string, value: string): Promise<void> => {
    if (!chrome.storage.session) {
      await chromeStorage.setItem(name, value);
      return;
    }
    await chrome.storage.session.set({ [name]: value });
  },
  removeItem: async (name: string): Promise<void> => {
    if (!chrome.storage.session) {
      await chromeStorage.removeItem(name);
      return;
    }
    await chrome.storage.session.remove(name);
  },
};
