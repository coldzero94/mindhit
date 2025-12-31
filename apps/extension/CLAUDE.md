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
│   ├── popup/              # Popup React UI (toolbar click)
│   │   ├── index.html
│   │   ├── index.tsx
│   │   ├── App.tsx
│   │   ├── styles.css
│   │   └── components/
│   │       ├── SessionControl.tsx
│   │       ├── SessionStats.tsx
│   │       ├── LoginPrompt.tsx
│   │       └── GoogleSignInButton.tsx
│   ├── lib/                # Shared utilities
│   │   ├── api.ts          # API client
│   │   ├── chrome-storage.ts # Chrome storage adapters
│   │   ├── constants.ts    # Constants (storage keys, client ID)
│   │   ├── events.ts       # Event queue management
│   │   └── storage.ts      # Chrome storage helpers
│   ├── stores/             # Zustand stores
│   │   ├── auth-store.ts   # Auth state (uses session storage)
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

### Google OAuth (Authorization Code Flow)

The extension uses `chrome.identity.launchWebAuthFlow()` for Google OAuth:

```typescript
// 1. Build OAuth URL
const redirectUri = chrome.identity.getRedirectURL();
const authUrl = new URL("https://accounts.google.com/o/oauth2/v2/auth");
authUrl.searchParams.set("client_id", GOOGLE_CLIENT_ID);
authUrl.searchParams.set("redirect_uri", redirectUri);
authUrl.searchParams.set("response_type", "code");
authUrl.searchParams.set("scope", "openid email profile");

// 2. Launch OAuth flow
const responseUrl = await chrome.identity.launchWebAuthFlow({
  url: authUrl.toString(),
  interactive: true,
});

// 3. Extract authorization code
const url = new URL(responseUrl);
const code = url.searchParams.get("code");

// 4. Exchange code for tokens via backend
const result = await api.googleAuthCode(code, redirectUri);
```

**Important**: The redirect URI format is `https://<extension-id>.chromiumapp.org/`

### Chrome Storage Adapters

Two storage adapters are provided in `lib/chrome-storage.ts`:

| Adapter | Storage Type | Use Case |
|---------|-------------|----------|
| `chromeStorage` | `chrome.storage.local` | Persistent data (session state) |
| `chromeSessionStorage` | `chrome.storage.session` | Sensitive data (auth tokens) - cleared on browser close |

### Zustand Hydration Pattern

When using async storage (chrome.storage), wait for hydration before rendering:

```typescript
const [isHydrated, setIsHydrated] = useState(false);

useEffect(() => {
  const unsub = useAuthStore.persist.onFinishHydration(() => {
    setIsHydrated(true);
  });

  if (useAuthStore.persist.hasHydrated()) {
    setIsHydrated(true);
  }

  return () => unsub();
}, []);

if (!isHydrated) {
  return <Loading />;
}
```

### Message Passing

- **Popup → Background**: Session control (start, pause, resume, stop)
- **Background → Content Scripts**: Recording state changes
- **Content Scripts → Background**: Event collection

### Event Batching

- Events are batched (configurable via `VITE_EVENT_BATCH_SIZE`)
- Flush interval configurable via `VITE_EVENT_FLUSH_INTERVAL`
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

# Linting
pnpm lint             # Run ESLint
```

## Loading in Chrome

1. Run `pnpm build`
2. Open `chrome://extensions`
3. Enable "Developer mode"
4. Click "Load unpacked"
5. Select the `dist` folder

## API Integration

The extension communicates with the backend at `VITE_API_URL` (default: `http://localhost:9000/v1`):

- `POST /auth/google/code` - Google OAuth login (Authorization Code)
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

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API URL | `http://localhost:9000/v1` |
| `VITE_EVENT_BATCH_SIZE` | Events per batch | `1` (dev), `10` (prod) |
| `VITE_EVENT_FLUSH_INTERVAL` | Flush interval (ms) | `5000` (dev), `30000` (prod) |

## Known Issues

- Icons are placeholder (empty files) - need actual PNG icons
- @crxjs/vite-plugin shows deprecation warning (beta version)

## Testing Notes

- Chrome API is mocked in `src/test/setup.ts`
- Store tests verify state transitions
- Event tests verify type safety and batching logic
