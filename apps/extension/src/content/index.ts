import type { BrowsingEvent } from "@/types";

let isRecording = false;
let pageEnteredAt: number | null = null;
let maxScrollDepth = 0;
let lastScrollTime = 0;

/**
 * Check if extension context is still valid.
 * Returns false if extension was unloaded/reloaded.
 */
function isContextValid(): boolean {
  try {
    return !!chrome.runtime?.id;
  } catch {
    return false;
  }
}

/**
 * Safely send a message to the background script.
 * Handles "Extension context invalidated" errors gracefully.
 */
function safeSendMessage(
  message: unknown,
  callback?: (response: unknown) => void
): void {
  if (!isContextValid()) {
    console.warn("[MindHit] Extension context invalidated, stopping recording");
    cleanup();
    return;
  }

  try {
    chrome.runtime.sendMessage(message, (response) => {
      // Check for runtime errors
      if (chrome.runtime.lastError) {
        const errorMessage = chrome.runtime.lastError.message || "";
        if (errorMessage.includes("context invalidated")) {
          console.warn("[MindHit] Extension context invalidated");
          cleanup();
          return;
        }
        console.warn("[MindHit] Message send error:", errorMessage);
        return;
      }
      callback?.(response);
    });
  } catch (error) {
    console.warn("[MindHit] Failed to send message:", error);
    cleanup();
  }
}

/**
 * Cleanup when extension context is invalidated.
 */
function cleanup(): void {
  isRecording = false;
  pageEnteredAt = null;
  window.removeEventListener("scroll", handleScroll);
  window.removeEventListener("beforeunload", handlePageLeave);
}

// Check initial state
safeSendMessage(
  { type: "GET_STATE" },
  (response: unknown) => {
    const state = response as { isRecording: boolean } | undefined;
    if (state?.isRecording) {
      startRecording();
    }
  }
);

// Message listener
try {
  chrome.runtime.onMessage.addListener((message) => {
    if (!isContextValid()) {
      cleanup();
      return;
    }

    switch (message.type) {
      case "START_RECORDING":
      case "RESUME_RECORDING":
      case "PAGE_LOADED":
        startRecording();
        break;

      case "PAUSE_RECORDING":
      case "STOP_RECORDING":
        stopRecording();
        break;
    }
  });
} catch {
  // Extension context may be invalidated on page load
  console.warn("[MindHit] Failed to add message listener");
}

function startRecording(): void {
  if (isRecording) return;
  if (!isContextValid()) return;

  isRecording = true;
  pageEnteredAt = Date.now();
  maxScrollDepth = 0;

  // Page visit event
  sendEvent({
    type: "page_visit",
    timestamp: Date.now(),
    url: window.location.href,
    title: document.title,
    referrer: document.referrer,
  });

  // Add event listeners
  window.addEventListener("scroll", handleScroll, { passive: true });
  window.addEventListener("beforeunload", handlePageLeave);

  // Update page count in Side Panel
  safeSendMessage({ type: "INCREMENT_PAGE_COUNT" });
}

function stopRecording(): void {
  if (!isRecording) return;

  handlePageLeave();

  window.removeEventListener("scroll", handleScroll);
  window.removeEventListener("beforeunload", handlePageLeave);

  isRecording = false;
  pageEnteredAt = null;
}

function handleScroll(): void {
  const now = Date.now();
  // Throttle scroll events (500ms)
  if (now - lastScrollTime < 500) return;
  lastScrollTime = now;

  const scrollTop = window.scrollY;
  const docHeight = document.documentElement.scrollHeight - window.innerHeight;
  const scrollDepth = docHeight > 0 ? Math.min(scrollTop / docHeight, 1) : 0;

  if (scrollDepth > maxScrollDepth) {
    maxScrollDepth = scrollDepth;

    sendEvent({
      type: "scroll",
      timestamp: now,
      url: window.location.href,
      scroll_depth: scrollDepth,
    });
  }
}

function handlePageLeave(): void {
  if (!pageEnteredAt) return;

  sendEvent({
    type: "page_leave",
    timestamp: Date.now(),
    url: window.location.href,
    duration_ms: Date.now() - pageEnteredAt,
    max_scroll_depth: maxScrollDepth,
  });
}

function sendEvent(event: BrowsingEvent): void {
  safeSendMessage({ type: "EVENT", event });
}
