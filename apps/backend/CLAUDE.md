# Backend - Claude Code Context

Go backend for MindHit API server and worker.

## Directory Structure

```
apps/backend/
├── cmd/
│   ├── api/           # API server entrypoint
│   └── worker/        # Worker entrypoint
├── ent/               # Ent ORM (schema + generated)
│   ├── schema/        # Entity definitions (source)
│   └── *.go           # Generated ORM code
├── internal/
│   ├── controller/    # HTTP handlers
│   ├── service/       # Business logic
│   ├── infrastructure/# External services (DB, Redis, etc.)
│   └── generated/     # oapi-codegen generated (from OpenAPI)
├── pkg/               # Shared code between API and Worker
├── test/              # Test utilities
├── oapi-codegen.yaml  # OpenAPI code generation config
├── atlas.hcl          # Database migration config
└── Makefile           # Build and generate commands
```

## Tech Stack

- **Framework**: Gin
- **ORM**: Ent
- **Database**: PostgreSQL
- **Queue**: Redis + Asynq (worker)
- **API Codegen**: oapi-codegen (from TypeSpec-generated OpenAPI)

## Key Patterns

### API Code Generation

Server code is generated from OpenAPI spec:

```bash
make generate-api  # Generates internal/generated/api.gen.go
```

Source: `packages/protocol/tsp-output/openapi/openapi.yaml`

### Ent Schema

Entity schemas are defined in `ent/schema/`. After changes:

```bash
go generate ./ent  # Regenerate ORM code
```

### Database Migrations

Using Atlas for migrations:

```bash
make migrate-diff name=<migration_name>  # Generate migration
make migrate-apply                        # Apply migrations
```

## Commands

```bash
# Development
make dev-api      # Run API server with hot reload
make dev-worker   # Run worker with hot reload

# Code Generation
make generate-api # Generate from OpenAPI
go generate ./ent # Generate Ent ORM

# Testing
make test         # Run all tests
make lint         # Run linter

# Database
make migrate-diff name=<name>  # Create migration
make migrate-apply             # Apply migrations
```

## Important Files

| File | Purpose |
|------|---------|
| `ent/schema/*.go` | Entity definitions |
| `internal/generated/api.gen.go` | Generated server interface |
| `oapi-codegen.yaml` | Code generation config |
| `atlas.hcl` | Migration config |

## Environment Variables

See `.env.local` for required variables:
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string (for worker)

## Notes

- Generated files in `internal/generated/` are committed (not gitignored)
- Controllers implement the generated `StrictServerInterface`
- All API routes are prefixed with `/v1/`
