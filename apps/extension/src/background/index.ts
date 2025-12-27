import type { BrowsingEvent } from "@/types";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:8080/v1";

interface SessionState {
  isRecording: boolean;
  sessionId: string | null;
  events: BrowsingEvent[];
}

const state: SessionState = {
  isRecording: false,
  sessionId: null,
  events: [],
};

// Open Side Panel when Extension icon is clicked
chrome.sidePanel.setPanelBehavior({ openPanelOnActionClick: true });

// Message listener
chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  switch (message.type) {
    case "SESSION_STARTED":
      state.isRecording = true;
      state.sessionId = message.sessionId;
      state.events = [];
      notifyContentScripts("START_RECORDING");
      break;

    case "SESSION_PAUSED":
      state.isRecording = false;
      notifyContentScripts("PAUSE_RECORDING");
      break;

    case "SESSION_RESUMED":
      state.isRecording = true;
      notifyContentScripts("RESUME_RECORDING");
      break;

    case "SESSION_STOPPED":
      state.isRecording = false;
      state.sessionId = null;
      flushEvents();
      notifyContentScripts("STOP_RECORDING");
      break;

    case "EVENT":
      if (state.isRecording && state.sessionId) {
        state.events.push(message.event);

        // Send events every 10 events or every 30 seconds
        if (state.events.length >= 10) {
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
    case "INCREMENT_HIGHLIGHT_COUNT":
      // These messages are for Side Panel to handle via store
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

// Periodic event flush (30 seconds)
setInterval(() => {
  if (state.isRecording && state.events.length > 0) {
    flushEvents();
  }
}, 30000);

// Network recovery: retry pending events
self.addEventListener("online", () => {
  retryPendingEvents();
});

async function flushEvents(): Promise<void> {
  if (state.events.length === 0 || !state.sessionId) return;

  const eventsToSend = [...state.events];
  state.events = [];

  try {
    const authData = await chrome.storage.local.get("mindhit-auth");
    const rawValue = authData["mindhit-auth"];
    const parsed =
      typeof rawValue === "string" ? JSON.parse(rawValue) : null;
    const token = parsed?.state?.token;

    if (!token) return;

    const response = await fetch(`${API_BASE_URL}/events/batch`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        session_id: state.sessionId,
        events: eventsToSend,
      }),
    });

    if (!response.ok) {
      // Re-add to queue on failure
      state.events = [...eventsToSend, ...state.events];
    }
  } catch (error) {
    console.error("Failed to send events:", error);
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
        const authData = await chrome.storage.local.get("mindhit-auth");
        const rawValue = authData["mindhit-auth"];
        const parsed =
          typeof rawValue === "string" ? JSON.parse(rawValue) : null;
        const token = parsed?.state?.token;

        if (!token) continue;

        const response = await fetch(`${API_BASE_URL}/events/batch`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            session_id: state.sessionId,
            events,
          }),
        });

        if (response.ok) {
          await chrome.storage.local.remove(key);
          console.log(`Retried ${events.length} pending events`);
        }
      } catch (error) {
        console.error("Failed to retry pending events:", error);
      }
    }
  }
}
