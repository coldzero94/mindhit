# MindHit

AI-powered mindmap generation from your browsing history.

MindHit collects your browsing sessions via a Chrome Extension and generates visual mindmaps using AI, helping you understand and recall your research patterns.

## Features

- **Session Recording** - Capture browsing sessions with page visits, highlights, and scroll depth
- **AI-Powered Analysis** - Extract topics and relationships using OpenAI, Gemini, or Claude
- **3D Mindmap Visualization** - Interactive galaxy-style mindmap with React Three Fiber
- **Usage Tracking** - Monitor token usage with subscription plans

## Tech Stack

**Backend:** Go · Gin · Ent ORM · Asynq · PostgreSQL · Redis

**Frontend:** Next.js 16 · React · TailwindCSS · React Three Fiber

**Extension:** React · Vite · CRXJS · Manifest V3

## Project Structure

```text
apps/
├── backend/     # Go API server + Asynq worker
├── web/         # Next.js dashboard
└── extension/   # Chrome extension

packages/
├── protocol/    # TypeSpec API definitions → OpenAPI
└── shared/      # Shared TypeScript utilities
```

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- pnpm 9+
- Docker & Docker Compose
- [moonrepo](https://moonrepo.dev/) (`npm install -g @moonrepo/cli`)

### Setup

```bash
# Clone repository
git clone https://github.com/your-org/mindhit.git
cd mindhit

# Install dependencies
pnpm install

# Start infrastructure (PostgreSQL, Redis)
docker-compose up -d

# Copy environment file
cp .env.example .env

# Run database migrations
moonx backend:migrate

# Seed test data
cd apps/backend && go run ./scripts/seed.go all && cd ../..
```

### Development

```bash
# Backend API (Terminal 1)
moonx backend:dev-api

# Backend Worker (Terminal 2)
moonx backend:dev-worker

# Web App (Terminal 3)
moonx web:dev

# Extension (Terminal 4)
moonx extension:dev
```

### Testing

```bash
# All tests
moonx :test

# Backend only
moonx backend:test

# Frontend only
moonx web:test

# Extension only
moonx extension:test
```

### Linting

```bash
# All projects
moonx :lint

# Backend (Go)
cd apps/backend && golangci-lint run ./...

# Frontend (TypeScript)
moonx web:lint
```

## Documentation

- [Architecture](docs/development/01-architecture.md)
- [Data Structure](docs/development/02-data-structure.md)
- [API Spec Workflow](docs/development/07-api-spec-workflow.md)
- [Development Phases](docs/development/phases/README.md)
- [Test Coverage](docs/development/10-test-coverage.md)

## Development Phases

| Phase | Status | Description                      |
| ----- | ------ | -------------------------------- |
| 0-6   | ✅     | Core infrastructure              |
| 7-8   | ✅     | Web App + Extension              |
| 9-10  | ✅     | Plan/Usage + AI                  |
| 11    | ✅     | Dashboard + 3D Mindmap           |
| 12-14 | ⬜     | Monitoring, Deployment, Billing  |

## Environment Variables

All environment variables are managed in the project root `.env` file. See [.env.example](.env.example) for required variables.

## License

Private - All rights reserved.
