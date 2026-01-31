# MindHit

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

AI-powered mindmap generation from your browsing history.

MindHit is a self-hosted application that collects your browsing sessions via a Chrome Extension and generates visual mindmaps using AI, helping you understand and recall your research patterns.

## Features

- **Session Recording** - Capture browsing sessions with page visits, highlights, and scroll depth
- **AI-Powered Analysis** - Extract topics and relationships using Groq, OpenAI, Gemini, or Claude
- **3D Mindmap Visualization** - Interactive galaxy-style mindmap with React Three Fiber
- **Self-Hosted** - Run on your own infrastructure, keep your data private
- **Single-User Design** - Simple API key authentication, no account management needed

## Quick Start

### Prerequisites

- Docker and Docker Compose v2+
- Go 1.24+
- Node.js 22+
- pnpm (install with `npm install -g pnpm`)
- Moon task runner (install with `npm install -g @moonrepo/cli`)
- Groq API key (free at [console.groq.com](https://console.groq.com))

### Setup

```bash
# Clone and configure
git clone https://github.com/coldzero94/mindhit.git
cd mindhit
cp .env.example .env

# Edit .env and set the following:
# - API_SECRET_KEY (generate with: openssl rand -hex 32)
# - GROQ_API_KEY (get from console.groq.com)
# - NEXT_PUBLIC_API_KEY (same as API_SECRET_KEY)
# - VITE_API_KEY (same as API_SECRET_KEY)

# Install moon task runner
npm install -g @moonrepo/cli

# Start infrastructure
docker-compose up -d

# Run migrations
cd apps/backend && go run ./cmd/api migrate

# Start backend (using moon task runner)
moonx backend:dev-api       # Terminal 1: API server
moonx backend:dev-worker    # Terminal 2: Worker

# Start frontend
moonx web:dev
```

Access the app at `http://localhost:3000`

### Chrome Extension

**Quick Install (Recommended):**
```bash
# Run the build script
./scripts/build-extension.sh   # macOS/Linux
# or
scripts\build-extension.bat     # Windows
```

**Manual Install:**
```bash
moonx extension:build
# Load apps/extension/dist in Chrome (chrome://extensions)
```

For detailed instructions, see [docs/SELF_HOSTING.md](docs/SELF_HOSTING.md).

## Documentation

- [Self-Hosting Guide](docs/SELF_HOSTING.md) - Full setup instructions
- [Configuration](docs/CONFIGURATION.md) - Environment variables
- [Extension Installation](docs/EXTENSION_INSTALL.md) - Chrome extension setup

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

docs/            # Documentation
infra/           # Infrastructure configs (see note below)
```

**Note on `infra/` folder:**

- Self-hosting users should use the **root `docker-compose.yml`** (PostgreSQL + Redis only)
- `infra/docker/` contains advanced monitoring stack (Prometheus, Grafana, Loki) - optional
- `infra/kind/`, `infra/helm/` are for Kubernetes deployments - advanced users only

## Development

### Commands

```bash
# Backend
moonx backend:dev-api     # Run API server
moonx backend:dev-worker  # Run worker
moonx backend:test        # Run tests

# Frontend
moonx web:dev             # Run dev server
moonx web:test            # Run tests
moonx web:build           # Production build

# Extension
moonx extension:dev       # Dev mode with hot reload
moonx extension:build     # Production build
moonx extension:test      # Run tests

# All projects
moonx :test               # Run all tests
moonx :lint               # Lint all projects
```

### Code Generation

```bash
moonx protocol:generate    # TypeSpec → OpenAPI → Go/TypeScript
```

## Contributing

Contributions are welcome! Please read our contribution guidelines before submitting PRs.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
