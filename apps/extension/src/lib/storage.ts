const STORAGE_KEY_PREFIX = "mindhit";

export const storage = {
  async get<T>(key: string): Promise<T | null> {
    const fullKey = `${STORAGE_KEY_PREFIX}-${key}`;
    const result = await chrome.storage.local.get(fullKey);
    return (result[fullKey] as T) || null;
  },

  async set<T>(key: string, value: T): Promise<void> {
    const fullKey = `${STORAGE_KEY_PREFIX}-${key}`;
    await chrome.storage.local.set({ [fullKey]: value });
  },

  async remove(key: string): Promise<void> {
    const fullKey = `${STORAGE_KEY_PREFIX}-${key}`;
    await chrome.storage.local.remove(fullKey);
  },

  async getAll(): Promise<Record<string, unknown>> {
    const result = await chrome.storage.local.get(null);
    const filtered: Record<string, unknown> = {};

    for (const [key, value] of Object.entries(result)) {
      if (key.startsWith(STORAGE_KEY_PREFIX)) {
        filtered[key.replace(`${STORAGE_KEY_PREFIX}-`, "")] = value;
      }
    }

    return filtered;
  },
};

// Network status detection
export function onOnline(callback: () => void): void {
  self.addEventListener("online", callback);
}

export function onOffline(callback: () => void): void {
  self.addEventListener("offline", callback);
}

export function isOnline(): boolean {
  return navigator.onLine;
}
