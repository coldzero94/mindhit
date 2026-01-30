import { STORAGE_KEYS } from "./constants";

/**
 * Get the auth token from chrome storage.
 * This utility extracts the token from the persisted Zustand auth store.
 */
export async function getAuthToken(): Promise<string | null> {
  try {
    const authData = await chrome.storage.local.get(STORAGE_KEYS.AUTH);
    const rawValue = authData[STORAGE_KEYS.AUTH];
    const parsed = typeof rawValue === "string" ? JSON.parse(rawValue) : null;
    return parsed?.state?.token ?? null;
  } catch {
    return null;
  }
}
