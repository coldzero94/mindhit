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
│   ├── api/              # Go Backend (Gin + Ent)
│   ├── web/              # Next.js 15 Web App
│   ├── extension/        # Chrome Extension (Plasmo)
│   └── worker/           # Background Worker (Temporal)
├── packages/
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
| Backend API | Go + Gin + Ent + PostgreSQL | `apps/api/` |
| Web App | Next.js 15 + TypeScript + TailwindCSS | `apps/web/` |
| Chrome Extension | TypeScript + Plasmo | `apps/extension/` |
| Worker | Go + Temporal | `apps/worker/` |
| Shared | TypeScript types, utilities | `packages/shared/` |

### Development Environment

- **3-Stage Environment**: `go run` (local) → `kind` (local K8s) → `EKS` (production)
- **Task Runner**: moonrepo (`moon run <project>:<task>`)
- **Database**: PostgreSQL
- **Queue**: Temporal

---

## Documentation Guidelines

### Monorepo Documentation Structure

This is a monorepo. Documentation should be **split by scope**:

| Scope | CLAUDE.md | README.md | Language |
|-------|-----------|-----------|----------|
| Project-wide | `/CLAUDE.md` (this file) | `/README.md` | English |
| Backend | `/apps/api/CLAUDE.md` | `/apps/api/README.md` | English |
| Frontend | `/apps/web/CLAUDE.md` | `/apps/web/README.md` | English |
| Extension | `/apps/extension/CLAUDE.md` | `/apps/extension/README.md` | English |
| Worker | `/apps/worker/CLAUDE.md` | `/apps/worker/README.md` | English |
| Shared | `/packages/shared/CLAUDE.md` | `/packages/shared/README.md` | English |
| Development phases | - | `/docs/development/phases/` | Korean |
| API specs | - | `/docs/api/` | English (OpenAPI) |

### CLAUDE.md Files (Per App)

Each app should have its own `CLAUDE.md` with app-specific context:

```
apps/api/CLAUDE.md       # Backend-specific: routes, services, DB schema
apps/web/CLAUDE.md       # Frontend-specific: components, pages, state
apps/extension/CLAUDE.md # Extension-specific: background, content scripts
apps/worker/CLAUDE.md    # Worker-specific: jobs, workflows
packages/shared/CLAUDE.md # Shared types, utilities
```

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
| Phase 0 | ⬜ Pending | 3-Stage Dev Environment |
| Phase 1 | ⬜ Pending | Project Initialization |
| Phase 1.5 | ⬜ Pending | API Spec Standardization |
| Phase 2 | ⬜ Pending | Authentication System |
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

_Development has not started yet._

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

### Backend (apps/api)

```bash
moon run api:dev          # Run dev server
moon run api:test         # Run tests
moon run api:generate     # Generate Ent + OpenAPI
moon run api:migrate      # Apply migrations
moon run api:migrate-diff # Generate migration
moon run api:seed-test-user # Create test user
```

### Frontend (apps/web)

```bash
moon run web:dev          # Run dev server
moon run web:test         # Run tests
moon run web:build        # Production build
moon run web:lint         # Lint code
```

### Extension (apps/extension)

```bash
moon run extension:dev    # Run dev mode
moon run extension:build  # Production build
moon run extension:test   # Run tests
```

### All Projects

```bash
moon run :test            # Run all tests
moon run :lint            # Lint all projects
moon run :build           # Build all projects
```

---

## Key Decisions

<!-- Record major technical decisions made during development -->

_No decisions recorded yet._

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
- OpenAPI spec: `docs/api/openapi.yaml`

### Authentication Flow

- JWT-based (Access Token 15min + Refresh Token 7days)
- Access Token: API authentication
- Refresh Token: Token renewal only
