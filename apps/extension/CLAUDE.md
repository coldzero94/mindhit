# MindHit Chrome Extension - Claude Code Context

Extension-specific development context. For project-wide context, see `/CLAUDE.md`.

## Overview

Chrome Extension (Manifest V3) for recording browsing sessions and sending events to the backend.

## Directory Structure

```
apps/extension/
├── src/
│   ├── background/         # Service Worker (event handling, API calls)
│   │   └── index.ts
│   ├── content/            # Content Script (DOM event collection)
│   │   └── index.ts
│   ├── sidepanel/          # Side Panel React UI
│   │   ├── index.html
│   │   ├── index.tsx
│   │   ├── App.tsx
│   │   ├── styles.css
│   │   └── components/
│   │       ├── SessionControl.tsx
│   │       ├── SessionStats.tsx
│   │       └── LoginPrompt.tsx
│   ├── lib/                # Shared utilities
│   │   ├── api.ts          # API client
│   │   ├── events.ts       # Event queue management
│   │   └── storage.ts      # Chrome storage helpers
│   ├── stores/             # Zustand stores
│   │   ├── auth-store.ts
│   │   └── session-store.ts
│   ├── types/              # TypeScript types
│   │   └── index.ts
│   ├── test/               # Test setup
│   │   └── setup.ts
│   └── vite-env.d.ts
├── public/
│   └── icons/              # Extension icons
├── manifest.json           # Chrome Extension manifest
├── vite.config.ts
├── vitest.config.ts
├── tsconfig.json
└── package.json
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Build Tool | Vite + @crxjs/vite-plugin |
| UI | React 19 |
| State | Zustand (with chrome.storage adapter) |
| Styling | Tailwind CSS v4 |
| Testing | Vitest |

## Key Patterns

### Chrome Storage Adapter for Zustand

```typescript
const chromeStorage: StateStorage = {
  getItem: async (name: string): Promise<string | null> => {
    const result = await chrome.storage.local.get(name);
    const value = result[name];
    return typeof value === "string" ? value : null;
  },
  setItem: async (name: string, value: string): Promise<void> => {
    await chrome.storage.local.set({ [name]: value });
  },
  removeItem: async (name: string): Promise<void> => {
    await chrome.storage.local.remove(name);
  },
};
```

### Message Passing

- **Side Panel → Background**: Session control (start, pause, resume, stop)
- **Background → Content Scripts**: Recording state changes
- **Content Scripts → Background**: Event collection

### Event Batching

- Events are batched (max 10 events or 30-second interval)
- Failed events are saved to local storage and retried on reconnection

## Commands

```bash
# Development
pnpm dev              # Start dev server
pnpm watch            # Build with watch mode

# Build
pnpm build            # Production build

# Testing
pnpm test             # Run tests
pnpm test:watch       # Run tests in watch mode

# Type checking
pnpm typecheck        # TypeScript check
```

## Loading in Chrome

1. Run `pnpm build`
2. Open `chrome://extensions`
3. Enable "Developer mode"
4. Click "Load unpacked"
5. Select the `dist` folder

## API Integration

The extension communicates with the backend at `VITE_API_URL` (default: `http://localhost:8080/v1`):

- `POST /auth/login` - User authentication
- `POST /sessions/start` - Start recording session
- `PATCH /sessions/:id/pause` - Pause session
- `PATCH /sessions/:id/resume` - Resume session
- `POST /sessions/:id/stop` - Stop session
- `POST /events/batch` - Send collected events

## Event Types

| Type | Description | Data |
|------|-------------|------|
| `page_visit` | Page loaded | url, title, referrer |
| `page_leave` | Page closed/navigated | duration_ms, max_scroll_depth |
| `scroll` | User scrolled | scroll_depth |
| `highlight` | Text selected | text, selector |
| `click` | Element clicked | selector, text |

## Known Issues

- Icons are placeholder (empty files) - need actual PNG icons
- @crxjs/vite-plugin shows deprecation warning (beta version)

## Testing Notes

- Chrome API is mocked in `src/test/setup.ts`
- Store tests verify state transitions
- Event tests verify type safety and batching logic
