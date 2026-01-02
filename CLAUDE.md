# MindHit - Claude Code Context

This file provides project context for Claude Code sessions.
Each conversation should reference this file to understand the current project state.

---

## Project Overview

**MindHit** is a service that collects user browsing history and generates AI-powered mindmaps.

### Monorepo Structure

This is a **monorepo** project. Code and documentation should be organized accordingly:

```
mindhit/
├── apps/
│   ├── backend/          # Go Backend (API + Worker)
│   │   ├── cmd/
│   │   │   ├── api/      # API server entrypoint
│   │   │   └── worker/   # Worker entrypoint
│   │   ├── internal/     # Private code (API/Worker specific)
│   │   │   └── generated/  # oapi-codegen generated (committed)
│   │   └── pkg/          # Shared code (services, infra)
│   ├── web/              # Next.js 16.1 Web App
│   │   └── src/api/generated/  # @hey-api/openapi-ts generated (committed)
│   └── extension/        # Chrome Extension (Vite + CRXJS)
├── packages/
│   ├── protocol/         # TypeSpec API definitions (Single Source of Truth)
│   │   ├── src/          # *.tsp files
│   │   └── tsp-output/   # Generated OpenAPI spec (committed)
│   └── shared/           # Shared types/utilities (cross-platform)
├── docs/
│   └── development/
│       └── phases/       # Phase-based development guides
├── infra/                # IaC (Terraform, Kubernetes)
└── CLAUDE.md             # This file
```

### Tech Stack

| Component | Stack | Location |
|-----------|-------|----------|
| Backend API | Go + Gin + Ent + PostgreSQL | `apps/backend/` |
| Web App | Next.js 16.1 + TypeScript + TailwindCSS | `apps/web/` |
| Chrome Extension | React + Vite + CRXJS + Zustand | `apps/extension/` |
| Worker | Go + Asynq | `apps/backend/` (same codebase) |
| API Protocol | TypeSpec → OpenAPI | `packages/protocol/` |
| Shared | TypeScript types, utilities | `packages/shared/` |

### Development Environment

- **3-Stage Environment**: `go run` (local) → `kind` (local K8s) → `EKS` (production)
- **Task Runner**: moonrepo (`moonx <project>:<task>`)
- **Database**: PostgreSQL
- **Queue**: Redis + Asynq

### Environment Variables

All environment variables are managed in the **project root**:

| File           | Purpose                  | Git         |
|----------------|--------------------------|-------------|
| `/.env`        | Actual values (local dev)| `.gitignore`|
| `/.env.example`| Sample template          | Committed   |

**IMPORTANT**:

- Backend loads env from `../../.env` (relative to `apps/backend/`)
- Frontend (Next.js) loads env from `../../.env` via `next.config.ts` (dotenv)
- Docker services (Grafana, PostgreSQL, etc.) settings are in `docker-compose.yml`
- Do NOT create `.env` files in subdirectories (e.g., `apps/backend/.env`, `apps/web/.env.local`)

**Environment Variable Sections**:

| Section | Prefix | Used By |
|---------|--------|---------|
| Backend | `API_PORT`, `DATABASE_URL`, etc. | Go API/Worker |
| Frontend | `NEXT_PUBLIC_*` | Next.js (client-side exposed) |
| Extension | `EXTENSION_*` | Chrome Extension (Phase 8+) |
| Docker | `POSTGRES_*`, `GF_*` | docker-compose services |

---

## Documentation Guidelines

### Monorepo Documentation Structure

This is a monorepo. Documentation should be **split by scope**:

| Scope | CLAUDE.md | README.md | Language |
|-------|-----------|-----------|----------|
| Project-wide | `/CLAUDE.md` (this file) | `/README.md` | English |
| Backend | `/apps/backend/CLAUDE.md` | `/apps/backend/README.md` | English |
| Frontend | `/apps/web/CLAUDE.md` | `/apps/web/README.md` | English |
| Extension | `/apps/extension/CLAUDE.md` | `/apps/extension/README.md` | English |
| Shared | `/packages/shared/CLAUDE.md` | `/packages/shared/README.md` | English |
| Development phases | - | `/docs/development/phases/` | Korean |
| API specs | - | `/docs/api/` | English (OpenAPI) |

### CLAUDE.md Files (Per App)

Each app **MUST** have its own `CLAUDE.md` with app-specific context:

| File | Status | Description |
|------|--------|-------------|
| `apps/backend/CLAUDE.md` | ✅ Exists | Go API server, Ent ORM, oapi-codegen |
| `apps/web/CLAUDE.md` | ✅ Exists | Next.js, @hey-api/openapi-ts, Zod |
| `apps/extension/CLAUDE.md` | ✅ Exists | Chrome Extension (Vite + CRXJS) |
| `packages/protocol/CLAUDE.md` | ✅ Exists | TypeSpec API definitions |
| `packages/shared/CLAUDE.md` | ⬜ TODO | Shared utilities (when created) |

**IMPORTANT: Keep app-specific CLAUDE.md files updated!**

When making changes to an app, update its CLAUDE.md if:
- Directory structure changes
- New patterns or conventions are introduced
- Important files are added/removed
- Commands change
- New dependencies affect the workflow

**What to include in app-specific CLAUDE.md:**
- App architecture and folder structure
- Key patterns and conventions used
- Important files and their purposes
- App-specific commands
- Known issues or gotchas for that app

### Writing Guidelines

- **CLAUDE.md files**: Always in English (all of them)
- **Phase documents**: Korean (for the development team)
- **Code comments**: English
- **Commit messages**: English

---

## Current Development Status

### Phase Progress

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 0 | ✅ Done | 3-Stage Dev Environment |
| Phase 1 | ✅ Done | Project Initialization |
| Phase 1.5 | ✅ Done | API Spec Standardization |
| Phase 2 | ✅ Done | Authentication System |
| Phase 2.1 | ✅ Done | Google OAuth (GIS-based) |
| Phase 3 | ✅ Done | Session Management API |
| Phase 4 | ✅ Done | Event Collection API |
| Phase 5 | ✅ Done | Monitoring & Infra (Basic) |
| Phase 6 | ✅ Done | Worker & Job Queue |
| Phase 7 | ✅ Done | Next.js Web App |
| Phase 8 | ✅ Done | Chrome Extension |
| Phase 8.1 | ✅ Done | Extension UX Enhancement |
| Phase 9 | ✅ Done | Plan & Usage System |
| Phase 10 | ✅ Done | AI Provider Infrastructure |
| Phase 10.1 | ✅ Done | AI Settings & Logging |
| Phase 10.2 | ✅ Done | Mindmap Generation |
| Phase 11.1 | ✅ Done | React Three Fiber Setup |
| Phase 11.2 | ✅ Done | 3D Mindmap Components |
| Phase 11.3 | ✅ Done | Session Detail Page (Mindmap API) |
| Phase 11.4 | ✅ Done | Account & Usage Page |
| Phase 11.5 | ✅ Done | Animation & Interaction |
| Phase 12 | ✅ Done | Production Monitoring |
| Phase 13 | ⬜ Pending | Deployment & Operations |
| Phase 14 | ⬜ Pending | Stripe Billing Integration |

> Detailed phase docs: `docs/development/phases/`

### Recent Changes

<!--
Record phase completions here (newest first):
- [YYYY-MM-DD] Phase X.X completed: Brief description
-->

- [2026-01-02] Phase 12 completed: Production Monitoring - Business metrics (sessions, events, auth, worker, AI), Grafana dashboards (api-overview, business-metrics, ai-worker, infrastructure), Loki log aggregation, Alertmanager with alerts.yml
- [2026-01-02] Phase 11.5 completed: Animation & Interaction - BigBangAnimation, ParticleField, NebulaEffect, PostProcessing, CameraController, AutoRotateCamera, useMindmapInteraction hook, Galaxy integration with seeded random for React 19 compatibility
- [2026-01-02] Phase 11.4 completed: Account & Usage Page - subscription/usage API wrappers, useSubscription/useUsage hooks, SubscriptionCard, UsageCard, UsageHistory components, account page, HeaderUsageBadge
- [2026-01-02] Phase 11.3 updated: Session title editing - SessionTitleEdit component, useUpdateSession hook integration, inline edit with Enter/Escape support
- [2025-12-31] Phase 11.3 completed: Session Detail Page Enhancement - Mindmap API (TypeSpec + Go backend), MindmapViewer component, useMindmap hook, mindmap-transform util, Tabs UI with Events/Mindmap views, seed data for testing
- [2025-12-31] Phase 11.2 completed: 3D Mindmap Components - Node, Edge, Galaxy components with hover/select animations, mindmap-utils, mock data
- [2025-12-31] Phase 11.1 completed: React Three Fiber Setup - Three.js dependencies, MindmapCanvas component, TestSphere component, test-3d page, Turbopack configuration
- [2025-12-31] Phase 8.1 completed: Extension UX Enhancement - Session list view, web dashboard links, session title editing, network status banner, settings page with auto-start and URL configuration
- [2025-12-31] Extension Google OAuth: Chrome Extension Google OAuth login with Authorization Code flow (chrome.identity API), sidepanel→popup migration, Zustand hydration pattern for chrome.storage
- [2025-12-29] Phase 10 refactored: Code quality improvements - Centralized JSON validation, Default model constants, Cache race condition fix (double-check locking), Mindmap conversion functions moved to service layer, Optimized GetCurrentUsage query
- [2025-12-29] Phase 10.2 completed: Mindmap Generation - Tag extraction worker handler, Mindmap generation handler with relationship graph, UsageService integration for token tracking
- [2025-12-29] Phase 10.1 completed: AI Settings & Logging - ai_logs/ai_configs tables, AILogService, AIConfigService (5-min caching), ProviderManager with DB config + fallback support
- [2025-12-29] Phase 10 completed: AI Provider Infrastructure - Unified ChatRequest/ChatResponse types, OpenAI/Gemini/Claude provider implementations, Streaming support
- [2025-12-28] Phase 2.1 completed: Google OAuth - Google Identity Services (GIS) integration, OAuthService, OAuthController, Google Sign-In button on login/signup pages
- [2025-12-28] Phase 9 completed: Plan & Usage System - Plan/Subscription/TokenUsage schemas, UsageService, SubscriptionService, Subscription and Usage API endpoints
- [2025-12-28] Phase 8 completed: Chrome Extension - Side Panel UI, Session control, Event collection (page_visit, scroll, highlight), Batch sending with offline support
- [2025-12-27] Phase 7 completed: Next.js Web App - Auth UI (login/signup), Sessions list/detail pages, Zustand + React Query, zod v4
- [2025-12-26] Phase 6 completed: Worker & Job Queue - Asynq queue, session processing handler, cleanup scheduler, cmd/api and cmd/worker separation
- [2025-12-26] Phase 5 completed: Monitoring & Infra - Prometheus metrics, slog logger, Request ID, HTTP logging middleware
- [2025-12-26] Phase 4 completed: Event Collection API (batch events, list events, event stats) with TypeSpec integration
- [2025-12-26] Phase 3 completed: Session Management API (start, pause, resume, stop, list, get, update, delete)
- [2025-12-26] Phase 2 completed: JWT authentication (signup, login, refresh, password reset, logout, me)
- [2025-12-26] Google OAuth split to Phase 2.1 (to be implemented after Phase 6, before Phase 7)
- [2025-12-25] Phase 1.5 completed: TypeSpec → OpenAPI → Go/TypeScript code generation pipeline
- [2025-12-25] Phase 1 completed: Go backend project initialized with Ent ORM, PostgreSQL, and test infrastructure
- [2025-12-25] Phase 0 completed: Moon + Docker development environment setup

---

## Phase Completion Rules

### Before Starting a Phase

1. Read the phase document (`docs/development/phases/phase-X-*.md`)
2. Check prerequisites are completed
3. Update phase status to 🟡 in `docs/development/phases/README.md`

### During Phase Work

1. Follow step-by-step checklists
2. Update checkboxes as you complete items
3. Commit frequently with clear messages

### After Completing a Phase

**IMPORTANT**: Update the following files:

1. **Phase document**: Mark all checkboxes complete
2. **README.md** (`docs/development/phases/`): Update status to ✅
3. **CLAUDE.md** (this file):
   - Update "Phase Progress" table
   - Add entry to "Recent Changes" section
   - Update any relevant sections (Known Issues, Decisions, etc.)

This ensures other Claude Code sessions have accurate project context.

---

## Test Credentials

For development/testing environments only:

| Field | Value |
|-------|-------|
| Email | `test@mindhit.dev` |
| Password | `test1234!` |
| Access Token | `$TEST_ACCESS_TOKEN` env variable |

> Note: Test user is created via seed script. The User ID is assigned from the database.
> See: `docs/development/phases/phase-2-auth.md`

---

## Common Commands

### Backend (apps/backend)

```bash
moonx backend:dev-api     # Run API dev server
moonx backend:dev-worker  # Run Worker dev server
moonx backend:test        # Run tests
moonx backend:generate    # Generate Ent + OpenAPI
moonx backend:migrate     # Apply migrations
moonx backend:migrate-diff # Generate migration
moonx backend:seed        # Run all seeds (test user, etc.)

# Direct seed commands (from apps/backend/)
go run ./scripts/seed.go test-user  # Create/update test user only
go run ./scripts/seed.go all        # Run all seeds
```

### Frontend (apps/web)

```bash
moonx web:dev          # Run dev server
moonx web:test         # Run tests
moonx web:build        # Production build
moonx web:lint         # Lint code
```

### Extension (apps/extension)

```bash
moonx extension:dev        # Run dev mode
moonx extension:build      # Production build
moonx extension:test       # Run tests
moonx extension:typecheck  # TypeScript check
moonx extension:watch      # Build with watch mode
```

### All Projects

```bash
moonx :test            # Run all tests
moonx :lint            # Lint all projects
moonx :build           # Build all projects
```

---

## CI/CD

### GitHub Actions Workflows

| Workflow             | Trigger Paths                        | Jobs                                       |
| -------------------- | ------------------------------------ | ------------------------------------------ |
| `backend-ci.yml`     | `apps/backend/**`                    | lint, test (w/ PostgreSQL & Redis), build  |
| `web-ci.yml`         | `apps/web/**`, `packages/**`         | lint, typecheck, build, test               |
| `extension-ci.yml`   | `apps/extension/**`, `packages/**`   | lint, typecheck, build, test               |
| `protocol-ci.yml`    | `packages/protocol/**`               | validate (OpenAPI sync check)              |

All workflows trigger on:

- Push to `main` or `develop` branches
- Pull requests to `main`

### Docker Images

| Image            | Dockerfile                       | Description          |
| ---------------- | -------------------------------- | -------------------- |
| `mindhit-api`    | `apps/backend/Dockerfile.api`    | API server (Go 1.24) |
| `mindhit-worker` | `apps/backend/Dockerfile.worker` | Worker (Go 1.24)     |

Build locally:

```bash
cd apps/backend
docker build -f Dockerfile.api -t mindhit-api .
docker build -f Dockerfile.worker -t mindhit-worker .
```

---

## Code Quality

### Linting Rules

**Always run lint before committing changes:**

```bash
# Backend (Go)
cd apps/backend && golangci-lint run ./...

# All projects via moon
moonx :lint
```

### Backend Lint Guidelines

The backend uses `golangci-lint` with strict rules. Key requirements:

| Rule               | Description                            | How to Fix                                       |
| ------------------ | -------------------------------------- | ------------------------------------------------ |
| `exported`         | Exported types/functions need comments | Add `// TypeName does...` comment                |
| `package-comments` | Packages need comments                 | Add `// Package name provides...` at top         |
| `unused-parameter` | Unused function parameters             | **Use the parameter** (don't just rename to `_`) |
| `errcheck`         | Error return values must be checked    | Handle or explicitly ignore with `_ =`           |

#### Unused Parameters

When a parameter is currently unused but may be used later:

- ❌ Don't rename to `_` (e.g., `func(ctx context.Context)` → `func(_ context.Context)`)
- ✅ Do use the parameter meaningfully (e.g., `slog.InfoContext(ctx, "...")`)
- ✅ For test helpers, use `t.Helper()` to mark the function

```go
// ❌ Bad - loses future usability
setupState: func(_ *testing.T, _ *service.SessionService, ...) {
    // no-op
}

// ✅ Good - parameter is used
setupState: func(t *testing.T, _ *service.SessionService, ...) {
    t.Helper() // Already in correct state - no setup needed
}
```

### Excluded from Lint

These paths are excluded from `revive` rules in `.golangci.yml`:

- `ent/schema/` - Ent ORM schema files (follow Ent conventions)
- `internal/generated/` - Auto-generated code from oapi-codegen

---

### Code Generation

```bash
pnpm run generate              # Generate all (TypeSpec → OpenAPI → Go/TS)
pnpm run generate:protocol     # TypeSpec → OpenAPI only
pnpm run generate:api:go       # OpenAPI → Go server code
pnpm run generate:api:ts       # OpenAPI → TypeScript (types + SDK + Zod)
```

---

## Key Decisions

<!-- Record major technical decisions made during development -->

| Date | Decision | Rationale |
|------|----------|-----------|
| 2025-12-25 | Use `@hey-api/openapi-ts` for TypeScript code generation | Supports Zod v4 natively, single tool for types + SDK + validation, no Java dependency (unlike openapi-generator-cli) |
| 2025-12-25 | Commit generated code (not gitignore) | Enables code review for API changes, no CI generation step needed |

---

## Known Issues

<!-- Record known issues or technical debt -->

_No known issues._

---

## Architecture Notes

### Database Schema (Core Tables)

| Table | Description | Phase |
|-------|-------------|-------|
| `users` | User accounts | Phase 2 |
| `sessions` | Browsing sessions | Phase 3 |
| `events` | Page visit events | Phase 4 |
| `plans` | Subscription plans | Phase 9 |
| `subscriptions` | User subscriptions | Phase 9 |
| `token_usages` | AI token tracking | Phase 9 |
| `mindmaps` | Generated mindmaps | Phase 10 |

### API Versioning

- All APIs are prefixed with `/v1/`
- OpenAPI spec: `packages/protocol/tsp-output/openapi/openapi.yaml`

### Authentication Flow

- JWT-based (Access Token 15min + Refresh Token 7days)
- Access Token: API authentication
- Refresh Token: Token renewal only

### Error Handling

Unified error handling patterns are documented in `docs/development/09-error-handling.md`:

| Section | Scope              | Key Patterns                                        |
| ------- | ------------------ | --------------------------------------------------- |
| 1-9     | Backend (Go)       | Error types, HTTP responses, logging, worker errors |
| 10      | Frontend (Next.js) | Axios interceptor, Toast messages, Error Boundary   |
| 11      | Extension (Chrome) | chrome.runtime.lastError, offline handling          |
| 12      | AI                 | Provider errors, token limits, retry strategies     |

All phase documents reference this central guide for consistency.
