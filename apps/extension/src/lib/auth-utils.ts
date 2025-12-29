import { STORAGE_KEYS } from "./constants";

/**
 * Get the auth token from chrome storage.
 * This utility extracts the token from the persisted Zustand auth store.
 * Checks chrome.storage.session first (where auth store saves), then falls back to local.
 */
export async function getAuthToken(): Promise<string | null> {
  try {
    // Try session storage first (auth store's primary storage)
    if (chrome.storage.session) {
      const sessionData = await chrome.storage.session.get(STORAGE_KEYS.AUTH);
      const sessionValue = sessionData[STORAGE_KEYS.AUTH];
      if (sessionValue) {
        const parsed = typeof sessionValue === "string" ? JSON.parse(sessionValue) : sessionValue;
        const token = parsed?.state?.token ?? null;
        if (token) return token;
      }
    }

    // Fallback to local storage
    const authData = await chrome.storage.local.get(STORAGE_KEYS.AUTH);
    const rawValue = authData[STORAGE_KEYS.AUTH];
    const parsed = typeof rawValue === "string" ? JSON.parse(rawValue) : null;
    return parsed?.state?.token ?? null;
  } catch {
    return null;
  }
}
