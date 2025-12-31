export type EventType =
  | "page_visit"
  | "page_leave"
  | "scroll"
  | "highlight"
  | "click";

export interface BaseEvent {
  type: EventType;
  timestamp: number;
  url: string;
}

export interface PageVisitEvent extends BaseEvent {
  type: "page_visit";
  title: string;
  referrer: string;
}

export interface PageLeaveEvent extends BaseEvent {
  type: "page_leave";
  duration_ms: number;
  max_scroll_depth: number;
}

export interface ScrollEvent extends BaseEvent {
  type: "scroll";
  scroll_depth: number;
}

export interface HighlightEvent extends BaseEvent {
  type: "highlight";
  text: string;
  selector: string;
}

export interface ClickEvent extends BaseEvent {
  type: "click";
  selector: string;
  text: string;
}

export type BrowsingEvent =
  | PageVisitEvent
  | PageLeaveEvent
  | ScrollEvent
  | HighlightEvent
  | ClickEvent;

export interface User {
  id: string;
  email: string;
}

export interface Session {
  id: string;
  title: string | null;
  session_status: "recording" | "paused" | "processing" | "completed" | "failed";
  started_at: string;
  ended_at?: string;
}
