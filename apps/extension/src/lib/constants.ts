/**
 * Storage keys used by Zustand persist stores.
 */
export const STORAGE_KEYS = {
  AUTH: "mindhit-auth",
  SESSION: "mindhit-session",
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
