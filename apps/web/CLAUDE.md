# Web - Claude Code Context

Next.js web application for MindHit.

## Directory Structure

```
apps/web/
├── src/
│   ├── api/
│   │   └── generated/       # @hey-api/openapi-ts generated (committed)
│   │       ├── types.gen.ts   # TypeScript types
│   │       ├── sdk.gen.ts     # API SDK functions
│   │       ├── zod.gen.ts     # Zod v4 validation schemas
│   │       └── client.gen.ts  # HTTP client config
│   ├── app/                 # Next.js App Router pages
│   ├── components/          # React components
│   ├── hooks/               # Custom React hooks
│   └── lib/                 # Utilities and API client wrapper
├── openapi-ts.config.ts     # Hey API configuration
└── package.json
```

## Tech Stack

- **Framework**: Next.js 16.1 (App Router)
- **Language**: TypeScript
- **Styling**: TailwindCSS
- **API Client**: @hey-api/openapi-ts (generated from OpenAPI)
- **Validation**: Zod v4

## Key Patterns

### API Code Generation

Client code is generated from OpenAPI spec:

```bash
pnpm run generate  # Generates src/api/generated/*
```

Source: `packages/protocol/tsp-output/openapi/openapi.yaml`

### Generated Files

| File | Purpose |
|------|---------|
| `types.gen.ts` | Request/Response TypeScript types |
| `sdk.gen.ts` | API call functions (routesLogin, routesSignup, etc.) |
| `zod.gen.ts` | Zod schemas for runtime validation |
| `client.gen.ts` | HTTP client configuration |

### API Usage Example

```typescript
import { routesLogin, zAuthLoginRequest } from '@/lib/api';

// Validate input with Zod
const validated = zAuthLoginRequest.parse({ email, password });

// Type-safe API call
const { data, error } = await routesLogin({ body: validated });
```

## Commands

```bash
# Development
pnpm run dev       # Start dev server

# Code Generation
pnpm run generate  # Generate API client from OpenAPI

# Build
pnpm run build     # Production build
pnpm run lint      # Lint code
```

## Important Files

| File | Purpose |
|------|---------|
| `openapi-ts.config.ts` | Hey API generation config |
| `src/api/generated/*` | Generated API client (committed) |
| `src/lib/api.ts` | API client wrapper and re-exports |

## Notes

- Generated files in `src/api/generated/` are committed (not gitignored)
- Zod schemas include validation rules from TypeSpec (e.g., `@minLength`)
- Use `zod.gen.ts` for form validation before API calls
