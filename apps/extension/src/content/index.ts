import type { BrowsingEvent } from "@/types";

let isRecording = false;
let pageEnteredAt: number | null = null;
let maxScrollDepth = 0;
let lastScrollTime = 0;

// Check initial state
chrome.runtime.sendMessage(
  { type: "GET_STATE" },
  (response: { isRecording: boolean } | undefined) => {
    if (response?.isRecording) {
      startRecording();
    }
  }
);

// Message listener
chrome.runtime.onMessage.addListener((message) => {
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

function startRecording(): void {
  if (isRecording) return;

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
  chrome.runtime.sendMessage({ type: "INCREMENT_PAGE_COUNT" });
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
  chrome.runtime.sendMessage({ type: "EVENT", event });
}
