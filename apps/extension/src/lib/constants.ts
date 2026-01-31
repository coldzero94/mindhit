/**
 * Storage keys used by Zustand persist stores.
 */
export const STORAGE_KEYS = {
  SESSION: "mindhit-session",
  SETTINGS: "mindhit-settings",
} as const;

/**
 * Web app URL (for opening dashboard, settings, etc.)
 */
export const WEB_APP_URL =
  import.meta.env.VITE_WEB_APP_URL || "http://localhost:3000";

/**
 * API base URL
 */
export const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:9000/v1";

/**
 * API Key for authentication (single-user self-hosted mode)
 */
export const API_KEY = import.meta.env.VITE_API_KEY || "";

/**
 * Event batching configuration
 */
export const EVENT_BATCH_SIZE = import.meta.env.VITE_EVENT_BATCH_SIZE
  ? Number(import.meta.env.VITE_EVENT_BATCH_SIZE)
  : 1; // Default: 1 for dev (send immediately)

export const EVENT_FLUSH_INTERVAL = import.meta.env.VITE_EVENT_FLUSH_INTERVAL
  ? Number(import.meta.env.VITE_EVENT_FLUSH_INTERVAL)
  : 5000; // Default: 5 seconds for dev (production: 30000)
