# Protocol - Claude Code Context

TypeSpec API definitions - the Single Source of Truth for all API contracts.

## Directory Structure

```
packages/protocol/
├── src/
│   ├── common/
│   │   └── errors.tsp      # Common error types
│   └── auth/
│       └── auth.tsp        # Auth API definitions
├── tsp-output/
│   └── openapi/
│       └── openapi.yaml    # Generated OpenAPI spec (committed)
├── main.tsp                # Entry point
├── tspconfig.yaml          # TypeSpec compiler config
└── package.json
```

## Tech Stack

- **Language**: TypeSpec
- **Output**: OpenAPI 3.0 YAML

## Key Concepts

### TypeSpec is the Source of Truth

```
TypeSpec (.tsp) → OpenAPI (yaml) → Go Server (oapi-codegen)
                                 → TypeScript Client (@hey-api/openapi-ts)
```

All API changes start here. Never edit the generated OpenAPI directly.

### TypeSpec Features Used

| Feature | Example | Effect |
|---------|---------|--------|
| `@minLength(n)` | `@minLength(8) password: string` | Zod: `.min(8)` |
| `@doc("...")` | `@doc("User email")` | OpenAPI description |
| `@route("/path")` | `@route("/v1/auth")` | API path prefix |
| `@encodedName` | `@encodedName("application/json", "created_at")` | JSON field name |

### Adding a New API

1. Create or edit `.tsp` file in `src/`
2. Import in `main.tsp` if new file
3. Run `pnpm run build` to generate OpenAPI
4. Run `pnpm run generate` from root to update Go/TS code

## Commands

```bash
pnpm run build   # Compile TypeSpec → OpenAPI
pnpm run watch   # Watch mode for development
pnpm run format  # Format .tsp files
```

## Important Files

| File | Purpose |
|------|---------|
| `main.tsp` | Entry point, service metadata |
| `src/**/*.tsp` | API definitions by domain |
| `tsp-output/openapi/openapi.yaml` | Generated OpenAPI (committed) |
| `tspconfig.yaml` | Compiler settings |

## Example: Adding Auth Endpoint

```typespec
// src/auth/auth.tsp
@route("/v1/auth")
namespace AuthRoutes {
  @post
  @route("/login")
  @doc("User login")
  op login(@body body: LoginRequest): AuthResponse | Common.ErrorResponse;
}
```

## Notes

- Generated OpenAPI in `tsp-output/` is committed (not gitignored)
- Use `@encodedName` for snake_case JSON fields (TypeSpec uses camelCase)
- All routes should be under `/v1/` prefix
- Error responses use `Common.ErrorResponse` or `Common.ValidationError`
