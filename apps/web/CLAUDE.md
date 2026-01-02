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
│   │   │   ├── test-3d/page.tsx   # 3D test page
│   │   │   └── layout.tsx         # Dashboard layout with auth guard
│   │   ├── layout.tsx          # Root layout with Providers
│   │   ├── page.tsx            # Home (redirects to /login or /sessions)
│   │   └── providers.tsx       # React Query + Toaster
│   ├── components/
│   │   ├── ui/                 # shadcn/ui components (tabs, card, etc.)
│   │   ├── auth/               # Auth components (login-form, signup-form)
│   │   ├── sessions/           # Session components
│   │   │   ├── SessionTitleEdit.tsx  # Inline title editing
│   │   └── mindmap/            # 3D mindmap components
│   │       ├── MindmapCanvas.tsx   # R3F Canvas wrapper
│   │       ├── MindmapViewer.tsx   # Mindmap viewer with API integration
│   │       ├── Galaxy.tsx          # Main 3D scene
│   │       ├── Node.tsx            # 3D node component
│   │       └── Edge.tsx            # 3D edge component
│   ├── lib/
│   │   ├── api/                # API client wrappers
│   │   │   ├── client.ts       # Axios client with interceptors
│   │   │   ├── auth.ts         # Auth API functions
│   │   │   ├── sessions.ts     # Sessions API functions
│   │   │   └── mindmap.ts      # Mindmap API functions
│   │   ├── hooks/              # React Query hooks
│   │   │   ├── use-sessions.ts # Sessions query/mutation hooks
│   │   │   └── use-mindmap.ts  # Mindmap query/mutation hooks
│   │   ├── utils/              # Utility functions
│   │   │   └── mindmap-transform.ts  # API to frontend type transform
│   │   └── utils.ts            # Utility functions (cn)
│   └── stores/
│       └── auth-store.ts       # Zustand auth store with persist
├── openapi-ts.config.ts        # Hey API configuration
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
| `src/components/mindmap/MindmapViewer.tsx` | Mindmap viewer with API integration |
| `src/app/providers.tsx` | QueryClient + Toaster setup |
| `eslint.config.mjs` | ESLint config (ignores generated files) |
| `openapi-ts.config.ts` | Hey API generation config |

## Notes

- Generated files in `src/api/generated/` are committed (not gitignored)
- Generated files are excluded from ESLint (`src/api/generated/**`)
- Zod v4 is used (supports `z.iso.datetime()`, `z.int()`)
- Use `error.issues` instead of `error.errors` for Zod v4 compatibility
- Sonner is used for toasts (not deprecated toast component)
