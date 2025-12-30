import type { BrowsingEvent } from "@/types";
import { getAuthToken } from "@/lib/auth-utils";
import {
  API_BASE_URL,
  EVENT_BATCH_SIZE,
  EVENT_FLUSH_INTERVAL,
} from "@/lib/constants";

const REQUEST_TIMEOUT = 30000; // 30 seconds

interface SessionState {
  isRecording: boolean;
  sessionId: string | null;
  events: BrowsingEvent[];
  pageCount: number;
}

const state: SessionState = {
  isRecording: false,
  sessionId: null,
  events: [],
  pageCount: 0,
};

let flushIntervalId: ReturnType<typeof setInterval> | null = null;

// Badge helper functions
function updateBadge(): void {
  if (state.isRecording) {
    // Show page count when recording
    const text = state.pageCount > 0 ? String(state.pageCount) : "";
    chrome.action.setBadgeText({ text });
    chrome.action.setBadgeBackgroundColor({ color: "#EF4444" }); // Red
  } else if (state.sessionId) {
    // Paused state
    chrome.action.setBadgeText({ text: "||" });
    chrome.action.setBadgeBackgroundColor({ color: "#F59E0B" }); // Yellow
  } else {
    // Idle state - clear badge
    chrome.action.setBadgeText({ text: "" });
  }
}

function resetBadge(): void {
  state.pageCount = 0;
  chrome.action.setBadgeText({ text: "" });
}

// Valid message types from extension
const VALID_MESSAGE_TYPES = [
  "SESSION_STARTED",
  "SESSION_PAUSED",
  "SESSION_RESUMED",
  "SESSION_STOPPED",
  "EVENT",
  "GET_STATE",
  "INCREMENT_PAGE_COUNT",
  "INCREMENT_HIGHLIGHT_COUNT",
] as const;

// Message listener with sender verification
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  // Verify sender is from our extension
  if (sender.id !== chrome.runtime.id) {
    return false;
  }

  // Validate message type
  if (!VALID_MESSAGE_TYPES.includes(message.type)) {
    return false;
  }

  switch (message.type) {
    case "SESSION_STARTED":
      state.isRecording = true;
      state.sessionId = message.sessionId;
      state.events = [];
      state.pageCount = 0;
      startFlushInterval();
      notifyContentScripts("START_RECORDING");
      updateBadge();
      break;

    case "SESSION_PAUSED":
      state.isRecording = false;
      stopFlushInterval();
      notifyContentScripts("PAUSE_RECORDING");
      updateBadge();
      break;

    case "SESSION_RESUMED":
      state.isRecording = true;
      startFlushInterval();
      notifyContentScripts("RESUME_RECORDING");
      updateBadge();
      break;

    case "SESSION_STOPPED":
      state.isRecording = false;
      state.sessionId = null;
      stopFlushInterval();
      flushEvents();
      notifyContentScripts("STOP_RECORDING");
      resetBadge();
      break;

    case "EVENT":
      if (state.isRecording && state.sessionId) {
        state.events.push(message.event);

        // Send events every EVENT_BATCH_SIZE events
        if (state.events.length >= EVENT_BATCH_SIZE) {
          flushEvents();
        }
      }
      break;

    case "GET_STATE":
      sendResponse({
        isRecording: state.isRecording,
        sessionId: state.sessionId,
      });
      break;

    case "INCREMENT_PAGE_COUNT":
      // Update badge with page count
      state.pageCount++;
      updateBadge();
      break;

    case "INCREMENT_HIGHLIGHT_COUNT":
      // Pass through for popup to handle via store
      break;
  }

  return true;
});

// Tab update detection (page navigation)
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (state.isRecording && changeInfo.status === "complete" && tab.url) {
    // Notify content script about new page
    chrome.tabs.sendMessage(tabId, { type: "PAGE_LOADED" }).catch(() => {
      // Tab might not have content script loaded yet
    });
  }
});

// Start flush interval only when recording
function startFlushInterval(): void {
  if (flushIntervalId) return;
  flushIntervalId = setInterval(() => {
    if (state.isRecording && state.events.length > 0) {
      flushEvents();
    }
  }, EVENT_FLUSH_INTERVAL);
}

// Stop flush interval when not recording
function stopFlushInterval(): void {
  if (flushIntervalId) {
    clearInterval(flushIntervalId);
    flushIntervalId = null;
  }
}

// Network recovery: retry pending events
self.addEventListener("online", () => {
  retryPendingEvents();
});

// Fetch with timeout
async function fetchWithTimeout(
  url: string,
  options: RequestInit,
  timeout: number = REQUEST_TIMEOUT
): Promise<Response> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
    });
    return response;
  } finally {
    clearTimeout(timeoutId);
  }
}

async function flushEvents(): Promise<void> {
  if (state.events.length === 0 || !state.sessionId) return;

  const eventsToSend = [...state.events];
  state.events = [];

  try {
    const token = await getAuthToken();
    if (!token) {
      console.error("[MindHit] No auth token, cannot send events");
      return;
    }

    const response = await fetchWithTimeout(
      `${API_BASE_URL}/sessions/${state.sessionId}/events`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          events: eventsToSend,
        }),
      }
    );

    if (!response.ok) {
      console.error("[MindHit] Failed to send events:", response.status, response.statusText);
      // Re-add to queue on failure
      state.events = [...eventsToSend, ...state.events];
    }
  } catch (error) {
    console.error("[MindHit] Failed to send events:", error);
    state.events = [...eventsToSend, ...state.events];
  }
}

async function notifyContentScripts(action: string): Promise<void> {
  const tabs = await chrome.tabs.query({});
  for (const tab of tabs) {
    if (tab.id) {
      chrome.tabs.sendMessage(tab.id, { type: action }).catch(() => {
        // Tab might not have content script loaded
      });
    }
  }
}

async function retryPendingEvents(): Promise<void> {
  const storage = await chrome.storage.local.get(null);
  const pendingKeys = Object.keys(storage).filter((k) =>
    k.startsWith("mindhit-pending-events-")
  );

  for (const key of pendingKeys) {
    const events = storage[key] as BrowsingEvent[];
    if (events && state.sessionId) {
      try {
        const token = await getAuthToken();
        if (!token) continue;

        const response = await fetchWithTimeout(
          `${API_BASE_URL}/sessions/${state.sessionId}/events`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
              events,
            }),
          }
        );

        if (response.ok) {
          await chrome.storage.local.remove(key);
        }
      } catch (error) {
        console.error("Failed to retry pending events:", error);
      }
    }
  }
}
