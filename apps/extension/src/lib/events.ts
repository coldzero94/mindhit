import type { BrowsingEvent } from "@/types";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:9000/v1";
const BATCH_SIZE = 10;
const FLUSH_INTERVAL = 30000; // 30 seconds
const MAX_RETRY = 3;

interface EventQueue {
  events: BrowsingEvent[];
  sessionId: string | null;
  flushTimer: ReturnType<typeof setInterval> | null;
  retryCount: number;
}

const queue: EventQueue = {
  events: [],
  sessionId: null,
  flushTimer: null,
  retryCount: 0,
};

export function initEventQueue(sessionId: string): void {
  queue.sessionId = sessionId;
  queue.events = [];
  queue.retryCount = 0;
  startFlushTimer();
}

export function addEvent(event: BrowsingEvent): void {
  if (!queue.sessionId) return;

  queue.events.push(event);

  if (queue.events.length >= BATCH_SIZE) {
    flush();
  }
}

export function stopEventQueue(): void {
  if (queue.flushTimer) {
    clearInterval(queue.flushTimer);
    queue.flushTimer = null;
  }
  flush(); // Send remaining events
  queue.sessionId = null;
}

function startFlushTimer(): void {
  if (queue.flushTimer) {
    clearInterval(queue.flushTimer);
  }
  queue.flushTimer = setInterval(flush, FLUSH_INTERVAL);
}

async function flush(): Promise<void> {
  if (queue.events.length === 0 || !queue.sessionId) return;

  const eventsToSend = [...queue.events];
  queue.events = [];

  try {
    const success = await sendEvents(queue.sessionId, eventsToSend);

    if (success) {
      queue.retryCount = 0;
    } else {
      handleFailure(eventsToSend);
    }
  } catch (error) {
    console.error("Failed to send events:", error);
    handleFailure(eventsToSend);
  }
}

function handleFailure(events: BrowsingEvent[]): void {
  queue.retryCount++;

  if (queue.retryCount < MAX_RETRY) {
    // Add failed events back to queue
    queue.events = [...events, ...queue.events];
  } else {
    // Save to local storage on max retry exceeded
    saveToLocal(events);
    queue.retryCount = 0;
  }
}

async function sendEvents(
  sessionId: string,
  events: BrowsingEvent[]
): Promise<boolean> {
  const authData = await chrome.storage.local.get("mindhit-auth");
  const rawValue = authData["mindhit-auth"];
  const parsed =
    typeof rawValue === "string" ? JSON.parse(rawValue) : null;
  const token = parsed?.state?.token;

  if (!token) return false;

  const response = await fetch(`${API_BASE_URL}/events/batch`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      session_id: sessionId,
      events,
    }),
  });

  return response.ok;
}

async function saveToLocal(events: BrowsingEvent[]): Promise<void> {
  const key = `mindhit-pending-events-${Date.now()}`;
  await chrome.storage.local.set({ [key]: events });
}

// Retry pending events (on app start)
export async function retryPendingEvents(sessionId: string): Promise<void> {
  const storage = await chrome.storage.local.get(null);
  const pendingKeys = Object.keys(storage).filter((k) =>
    k.startsWith("mindhit-pending-events-")
  );

  for (const key of pendingKeys) {
    const events = storage[key] as BrowsingEvent[];
    if (events.length > 0) {
      const success = await sendEvents(sessionId, events);
      if (success) {
        await chrome.storage.local.remove(key);
      }
    }
  }
}
