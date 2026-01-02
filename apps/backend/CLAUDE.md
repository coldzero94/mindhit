# Backend - Claude Code Context

Go backend for MindHit API server and worker.

## Directory Structure

```
apps/backend/
├── cmd/
│   ├── api/              # API server entrypoint
│   └── worker/           # Worker entrypoint
├── ent/                  # Ent ORM (schema + generated)
│   ├── schema/           # Entity definitions (source)
│   ├── migrate/migrations/  # Atlas migration files
│   └── *.go              # Generated ORM code
├── internal/
│   ├── controller/       # HTTP handlers (*_test.go included)
│   ├── service/          # Business logic (*_test.go included)
│   ├── infrastructure/   # External services (DB, Redis, AI, etc.)
│   │   ├── ai/           # AI provider implementations
│   │   ├── database/     # Database connection
│   │   ├── middleware/   # HTTP middleware
│   │   └── queue/        # Redis + Asynq queue
│   ├── worker/           # Worker handlers
│   │   └── handler/      # Job handlers (*_test.go included)
│   ├── testutil/         # Test utilities (fixtures, helpers)
│   └── generated/        # oapi-codegen generated (from OpenAPI)
├── tests/
│   └── integration/      # Integration tests (API flow tests)
├── scripts/              # Development scripts (seed, etc.)
├── oapi-codegen.yaml     # OpenAPI code generation config
├── atlas.hcl             # Database migration config
└── Makefile              # Build and generate commands
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
# Development (via moonrepo)
moonx backend:dev-api       # Run API server with hot reload (Air)
moonx backend:dev-worker    # Run worker with hot reload (Air)
moonx backend:dev-api-test  # Run API in test mode (ENVIRONMENT=test)

# Code Generation
moonx backend:generate      # Generate Ent + OpenAPI code
go generate ./ent           # Generate Ent ORM only

# Testing
moonx backend:test          # Run all tests
go test ./...               # Run tests directly
golangci-lint run ./...     # Run linter

# Database (via moonrepo)
moonx backend:migrate-diff  # Generate migration from schema diff
moonx backend:migrate       # Apply migrations

# Seed Data (Development)
go run ./scripts/seed.go all        # Run all seeds
go run ./scripts/seed.go test-user  # Create test user only
```

## Important Files

| File | Purpose |
|------|---------|
| `ent/schema/*.go` | Entity definitions |
| `ent/migrate/migrations/*.sql` | Database migrations |
| `internal/generated/api.gen.go` | Generated server interface |
| `internal/testutil/fixtures.go` | Test fixtures and helpers |
| `tests/integration/*_test.go` | Integration tests (API flows) |
| `scripts/seed.go` | Development seed script |
| `oapi-codegen.yaml` | Code generation config |
| `atlas.hcl` | Migration config |

## Environment Variables

See `/.env` (project root) for required variables:
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string (for worker)
- `ENVIRONMENT` - Set to `test` for test mode (disables rate limiting, enables hard delete)
- `TEST_ACCESS_TOKEN` - Test token for development (optional)

## Test Credentials

For development/testing:

| Field    | Value              |
|----------|--------------------|
| Email    | `test@mindhit.dev` |
| Password | `test1234!`        |

Create with: `go run ./scripts/seed.go test-user`

## Testing

### Unit Tests

Unit tests are colocated with source files (`*_test.go` next to `*.go`):

- `internal/controller/*_test.go` - Controller tests
- `internal/service/*_test.go` - Service tests
- `internal/worker/handler/*_test.go` - Worker handler tests
- `internal/infrastructure/**/*_test.go` - Infrastructure tests

### Integration Tests

Integration tests are in `tests/integration/`:

- `auth_flow_test.go` - Authentication flow tests
- `session_flow_test.go` - Session management tests
- `event_flow_test.go` - Event collection tests
- `helpers_test.go` - Test helpers

Run integration tests:

```bash
go test ./tests/integration/...
```

## Notes

- Generated files in `internal/generated/` are committed (not gitignored)
- Controllers implement the generated `StrictServerInterface`
- All API routes are prefixed with `/v1/`
- Test mode (`ENVIRONMENT=test`) disables rate limiting and enables hard delete
- All FK constraints have CASCADE DELETE for proper data cleanup
