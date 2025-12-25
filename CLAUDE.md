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
│   └── extension/        # Chrome Extension (Plasmo)
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
| Chrome Extension | TypeScript + Plasmo | `apps/extension/` |
| Worker | Go + Asynq | `apps/backend/` (same codebase) |
| API Protocol | TypeSpec → OpenAPI | `packages/protocol/` |
| Shared | TypeScript types, utilities | `packages/shared/` |

### Development Environment

- **3-Stage Environment**: `go run` (local) → `kind` (local K8s) → `EKS` (production)
- **Task Runner**: moonrepo (`moonx <project>:<task>`)
- **Database**: PostgreSQL
- **Queue**: Redis + Asynq

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
| `apps/extension/CLAUDE.md` | ⬜ TODO | Chrome extension (Phase 8) |
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
| Phase 2.1 | ⬜ Pending | Google OAuth (after Phase 6, before Phase 7) |
| Phase 3 | ⬜ Pending | Session Management API |
| Phase 4 | ⬜ Pending | Event Collection API |
| Phase 5 | ⬜ Pending | Monitoring & Infra (Basic) |
| Phase 6 | ⬜ Pending | Worker & Job Queue |
| Phase 7 | ⬜ Pending | Next.js Web App |
| Phase 8 | ⬜ Pending | Chrome Extension |
| Phase 9 | ⬜ Pending | Plan & Usage System |
| Phase 10 | ⬜ Pending | AI Mindmap Generation |
| Phase 11 | ⬜ Pending | Web App Dashboard |
| Phase 12 | ⬜ Pending | Production Monitoring |
| Phase 13 | ⬜ Pending | Deployment & Operations |
| Phase 14 | ⬜ Pending | Stripe Billing Integration |

> Detailed phase docs: `docs/development/phases/`

### Recent Changes

<!--
Record phase completions here (newest first):
- [YYYY-MM-DD] Phase X.X completed: Brief description
-->

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
moonx backend:seed-test-user # Create test user
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
moonx extension:dev    # Run dev mode
moonx extension:build  # Production build
moonx extension:test   # Run tests
```

### All Projects

```bash
moonx :test            # Run all tests
moonx :lint            # Lint all projects
moonx :build           # Build all projects
```

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
