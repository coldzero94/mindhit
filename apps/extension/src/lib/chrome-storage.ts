import type { StateStorage } from "zustand/middleware";

/**
 * Check if extension context is valid.
 * Returns false if extension was unloaded/reloaded.
 */
function isContextValid(): boolean {
  try {
    // This will throw if context is invalidated
    return !!chrome.runtime?.id;
  } catch {
    return false;
  }
}

/**
 * Chrome storage adapter for Zustand persist.
 * This provides a centralized implementation to avoid duplication.
 */
export const chromeStorage: StateStorage = {
  getItem: async (name: string): Promise<string | null> => {
    if (!isContextValid()) {
      console.warn("[MindHit] Extension context invalidated, cannot read storage");
      return null;
    }
    try {
      const result = await chrome.storage.local.get(name);
      const value = result[name];
      return typeof value === "string" ? value : null;
    } catch (error) {
      console.error("[MindHit] Failed to read from storage:", error);
      return null;
    }
  },
  setItem: async (name: string, value: string): Promise<void> => {
    if (!isContextValid()) {
      console.warn("[MindHit] Extension context invalidated, cannot write storage");
      return;
    }
    try {
      await chrome.storage.local.set({ [name]: value });
    } catch (error) {
      console.error("[MindHit] Failed to write to storage:", error);
    }
  },
  removeItem: async (name: string): Promise<void> => {
    if (!isContextValid()) {
      console.warn("[MindHit] Extension context invalidated, cannot remove from storage");
      return;
    }
    try {
      await chrome.storage.local.remove(name);
    } catch (error) {
      console.error("[MindHit] Failed to remove from storage:", error);
    }
  },
};

/**
 * Session storage adapter for sensitive data (like tokens).
 * Data is cleared when the browser is closed.
 * NOTE: chrome.storage.session requires Manifest V3 and "storage" permission.
 */
export const chromeSessionStorage: StateStorage = {
  getItem: async (name: string): Promise<string | null> => {
    if (!isContextValid()) {
      return null;
    }
    // chrome.storage.session may not be available in all contexts
    if (!chrome.storage.session) {
      return chromeStorage.getItem(name);
    }
    try {
      const result = await chrome.storage.session.get(name);
      const value = result[name];
      return typeof value === "string" ? value : null;
    } catch (error) {
      console.error("[MindHit] Failed to read from session storage:", error);
      return null;
    }
  },
  setItem: async (name: string, value: string): Promise<void> => {
    if (!isContextValid()) {
      return;
    }
    if (!chrome.storage.session) {
      await chromeStorage.setItem(name, value);
      return;
    }
    try {
      await chrome.storage.session.set({ [name]: value });
    } catch (error) {
      console.error("[MindHit] Failed to write to session storage:", error);
    }
  },
  removeItem: async (name: string): Promise<void> => {
    if (!isContextValid()) {
      return;
    }
    if (!chrome.storage.session) {
      await chromeStorage.removeItem(name);
      return;
    }
    try {
      await chrome.storage.session.remove(name);
    } catch (error) {
      console.error("[MindHit] Failed to remove from session storage:", error);
    }
  },
};
