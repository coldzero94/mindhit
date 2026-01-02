# Web - Claude Code Context

Next.js web application for MindHit.

## Directory Structure

```
apps/web/
├── src/
│   ├── api/
│   │   └── generated/          # @hey-api/openapi-ts generated (committed)
│   │       ├── types.gen.ts    # TypeScript types
│   │       ├── sdk.gen.ts      # API SDK functions
│   │       ├── zod.gen.ts      # Zod v4 validation schemas
│   │       └── client.gen.ts   # HTTP client config
│   ├── app/                    # Next.js App Router pages
│   │   ├── (auth)/             # Auth routes (login, signup)
│   │   │   ├── login/page.tsx
│   │   │   ├── signup/page.tsx
│   │   │   └── layout.tsx
│   │   ├── (dashboard)/        # Protected routes (requires auth)
│   │   │   ├── sessions/
│   │   │   │   ├── page.tsx       # Session list
│   │   │   │   └── [id]/page.tsx  # Session detail with Mindmap tab
│   │   │   ├── account/page.tsx   # Account & usage page
│   │   │   ├── test-3d/page.tsx   # 3D test page
│   │   │   └── layout.tsx         # Dashboard layout with auth guard
│   │   ├── layout.tsx          # Root layout with Providers
│   │   ├── page.tsx            # Home (redirects to /login or /sessions)
│   │   └── providers.tsx       # React Query + Toaster
│   ├── components/
│   │   ├── ui/                 # shadcn/ui components (tabs, card, etc.)
│   │   ├── auth/               # Auth components (login-form, signup-form)
│   │   ├── sessions/           # Session components
│   │   │   └── SessionTitleEdit.tsx  # Inline title editing
│   │   ├── account/            # Account page components
│   │   │   ├── SubscriptionCard.tsx  # Subscription info card
│   │   │   ├── UsageCard.tsx         # Token usage card with progress
│   │   │   └── UsageHistory.tsx      # Usage history chart
│   │   ├── layout/             # Layout components
│   │   │   └── HeaderUsageBadge.tsx  # Header usage warning badge
│   │   └── mindmap/            # 3D mindmap components
│   │       ├── MindmapCanvas.tsx     # R3F Canvas wrapper
│   │       ├── MindmapViewer.tsx     # Mindmap viewer with API integration
│   │       ├── Galaxy.tsx            # Main 3D scene (with animations)
│   │       ├── Node.tsx              # 3D node with spring animations
│   │       ├── Edge.tsx              # 3D edge component
│   │       ├── BigBangAnimation.tsx  # Initial expansion animation
│   │       ├── ParticleField.tsx     # Background particle stars
│   │       ├── NebulaEffect.tsx      # Spiral nebula effect
│   │       ├── PostProcessing.tsx    # Bloom + Vignette effects
│   │       ├── CameraController.tsx  # Focus camera on selected node
│   │       └── AutoRotateCamera.tsx  # Idle auto-rotation
│   ├── lib/
│   │   ├── api/                # API client wrappers
│   │   │   ├── client.ts       # Axios client with interceptors
│   │   │   ├── auth.ts         # Auth API functions
│   │   │   ├── sessions.ts     # Sessions API functions
│   │   │   ├── mindmap.ts      # Mindmap API functions
│   │   │   ├── subscription.ts # Subscription API functions
│   │   │   └── usage.ts        # Usage API functions
│   │   ├── hooks/              # React Query hooks
│   │   │   ├── use-sessions.ts # Sessions query/mutation hooks
│   │   │   ├── use-mindmap.ts  # Mindmap query/mutation hooks
│   │   │   ├── use-subscription.ts  # Subscription hooks
│   │   │   └── use-usage.ts    # Usage hooks (with auto-refresh)
│   │   ├── utils/              # Utility functions
│   │   │   └── mindmap-transform.ts  # API to frontend type transform
│   │   └── utils.ts            # Utility functions (cn)
│   ├── stores/
│   │   └── auth-store.ts       # Zustand auth store with persist
│   └── test/                   # Test infrastructure
│       ├── integration/        # Integration tests (API + cleanup)
│       ├── mocks/              # MSW handlers and mock data
│       ├── helpers/            # Test helper functions
│       ├── setup.ts            # Vitest setup
│       └── utils.tsx           # Test render utilities
├── openapi-ts.config.ts        # Hey API configuration
├── vitest.config.ts            # Vitest configuration
├── moon.yml                    # Moon task definitions
└── package.json
```

## Tech Stack

- **Framework**: Next.js 16.1 (App Router)
- **Language**: TypeScript
- **Styling**: TailwindCSS v4
- **UI Components**: shadcn/ui
- **State Management**: Zustand (auth), React Query (server state)
- **API Client**: Axios + @hey-api/openapi-ts (types)
- **Validation**: Zod v4
- **Toast**: Sonner
- **3D Visualization**: React Three Fiber, @react-three/drei, @react-three/postprocessing

## Key Patterns

### Authentication Flow

1. **Auth Store** (`stores/auth-store.ts`): Zustand store with localStorage persistence
2. **API Client** (`lib/api/client.ts`): Axios interceptor adds Bearer token
3. **Dashboard Layout** (`app/(dashboard)/layout.tsx`): Auth guard redirects to /login
4. **Home Page** (`app/page.tsx`): Redirects based on auth state

### API Code Generation

Client types are generated from OpenAPI spec:

```bash
moonx web:generate  # Generates src/api/generated/*
```

Source: `packages/protocol/tsp-output/openapi/openapi.yaml`

### React Query Hooks

```typescript
// Session hooks
const { data, isLoading } = useSessions(limit, offset);
const { data: session } = useSession(id);
const deleteSession = useDeleteSession();
await deleteSession.mutateAsync(sessionId);
const updateSession = useUpdateSession();
await updateSession.mutateAsync({ id, data: { title: newTitle } });

// Mindmap hooks
const { data: mindmap } = useMindmap(sessionId);
const generateMindmap = useGenerateMindmap();
await generateMindmap.mutateAsync({ sessionId, options: { force: true } });

// Subscription hooks
const { data: subscription } = useSubscription();
const { data: plans } = usePlans();

// Usage hooks (auto-refresh every 5 minutes)
const { data: usage } = useUsage();
const { data: history } = useUsageHistory(6); // Last 6 months
```

### Form Validation with Zod v4

```typescript
const result = loginSchema.safeParse(formData);
if (!result.success) {
  result.error.issues.forEach((issue) => {
    // Handle validation errors
  });
}
```

## Commands

```bash
# Development
moonx web:dev        # Start dev server (localhost:3000)
pnpm dev             # Alternative

# Code Quality
moonx web:lint       # Run ESLint
moonx web:typecheck  # Run TypeScript check

# Testing
moonx web:test                   # Run all tests
pnpm vitest run                  # Run tests directly
pnpm test:integration            # Run integration tests only

# Build
moonx web:build      # Production build

# Code Generation
moonx web:generate   # Generate API client from OpenAPI
```

## Important Files

| File | Purpose |
|------|---------|
| `src/stores/auth-store.ts` | Zustand auth state (user, tokens) |
| `src/lib/api/client.ts` | Axios client with auth interceptor |
| `src/lib/hooks/use-sessions.ts` | React Query hooks for sessions (CRUD) |
| `src/components/sessions/SessionTitleEdit.tsx` | Inline session title editing |
| `src/lib/hooks/use-mindmap.ts` | React Query hooks for mindmap API |
| `src/lib/hooks/use-subscription.ts` | Subscription query hooks |
| `src/lib/hooks/use-usage.ts` | Usage query hooks with auto-refresh |
| `src/components/mindmap/MindmapViewer.tsx` | Mindmap viewer with API integration |
| `src/components/account/SubscriptionCard.tsx` | Subscription info display |
| `src/components/account/UsageCard.tsx` | Token usage with progress bar |
| `src/components/layout/HeaderUsageBadge.tsx` | Header badge for 80%+ usage |
| `src/app/(dashboard)/account/page.tsx` | Account & usage page |
| `src/app/providers.tsx` | QueryClient + Toaster setup |
| `eslint.config.mjs` | ESLint config (ignores generated files) |
| `openapi-ts.config.ts` | Hey API generation config |
| `vitest.config.ts` | Vitest test configuration |
| `src/test/setup.ts` | Vitest setup (MSW, cleanup) |
| `src/test/mocks/handlers.ts` | MSW request handlers |
| `src/test/integration/*.test.ts` | Integration tests |

## Testing

### Unit Tests

Unit tests are colocated with components (`*.test.tsx` next to `*.tsx`):

- `src/components/**/*.test.tsx` - Component tests
- `src/lib/hooks/*.test.ts` - Hook tests

### Integration Tests

Integration tests are in `src/test/integration/`:

- `auth.test.ts` - Authentication flow tests
- `sessions.test.ts` - Session management tests

Integration tests run against a real backend (`NEXT_PUBLIC_API_URL`) and include automatic cleanup.

Run with backend in test mode:

```bash
# Terminal 1: Start backend in test mode
moonx backend:dev-api-test

# Terminal 2: Run integration tests
pnpm test:integration
```

## Notes

- Generated files in `src/api/generated/` are committed (not gitignored)
- Generated files are excluded from ESLint (`src/api/generated/**`)
- Zod v4 is used (supports `z.iso.datetime()`, `z.int()`)
- Use `error.issues` instead of `error.errors` for Zod v4 compatibility
- Sonner is used for toasts (not deprecated toast component)
