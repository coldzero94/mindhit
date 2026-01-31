# Self-Hosting Guide

Get MindHit running on your own infrastructure in 5 minutes.

## Prerequisites

- Docker and Docker Compose v2+
- Go 1.24+
- Node.js 22+
- pnpm (install with `npm install -g pnpm`)
- Moon task runner (install with `npm install -g @moonrepo/cli`)
- Groq API key (free at [console.groq.com](https://console.groq.com))

## Quick Start

### 1. Clone and Configure

```bash
git clone https://github.com/coldzero94/mindhit.git
cd mindhit
cp .env.example .env
```

Edit `.env` and set the required values:

```bash
# Generate a random API key
API_SECRET_KEY=$(openssl rand -hex 32)

# Set your Groq API key
GROQ_API_KEY=gsk_your_key_here

# Set the same API key for frontend and extension
NEXT_PUBLIC_API_KEY=$API_SECRET_KEY
VITE_API_KEY=$API_SECRET_KEY
```

### 2. Start Infrastructure

```bash
docker-compose up -d
```

This starts PostgreSQL and Redis.

### 3. Run Database Migrations

```bash
cd apps/backend
go run ./cmd/api migrate
```

### 4. Install Moon Task Runner

```bash
npm install -g @moonrepo/cli
```

### 5. Start the Backend

```bash
# Terminal 1: API Server
moonx backend:dev-api

# Terminal 2: Worker
moonx backend:dev-worker
```

### 6. Start the Frontend

```bash
moonx web:dev
```

### 7. Access the Application

- Web UI: <http://localhost:3000>
- API: <http://localhost:9000>

## Chrome Extension Setup

See [EXTENSION_INSTALL.md](./EXTENSION_INSTALL.md) for building and installing the Chrome extension.

## Production Deployment

For production, we recommend:

1. Use a reverse proxy (nginx, Caddy) with HTTPS
2. Set strong `API_SECRET_KEY`
3. Use managed PostgreSQL (e.g., AWS RDS, Supabase)
4. Use managed Redis (e.g., Upstash, Redis Cloud)

### Docker Production Setup

Build the Docker images:

```bash
cd apps/backend
docker build -f Dockerfile.api -t mindhit-api .
docker build -f Dockerfile.worker -t mindhit-worker .
```

Run with your production configuration:

```bash
docker run -d \
  --name mindhit-api \
  -e API_SECRET_KEY=$API_SECRET_KEY \
  -e DATABASE_URL=$DATABASE_URL \
  -e REDIS_URL=$REDIS_URL \
  -e GROQ_API_KEY=$GROQ_API_KEY \
  -p 9000:9000 \
  mindhit-api
```

## Monitoring (Optional)

For monitoring with Prometheus, Grafana, and Loki:

```bash
cd infra/docker
docker-compose up -d
```

Access Grafana at <http://localhost:3010> (admin/admin).

## Troubleshooting

### API returns 401 Unauthorized

Check that your API key is configured correctly:

- `API_SECRET_KEY` in backend `.env`
- `NEXT_PUBLIC_API_KEY` for frontend
- `VITE_API_KEY` for extension

### AI features not working

1. Check that `GROQ_API_KEY` is set correctly
2. Verify the key at [console.groq.com](https://console.groq.com)
3. Check worker logs for errors

### Database connection errors

1. Ensure PostgreSQL is running: `docker-compose ps`
2. Check `DATABASE_URL` format in `.env`
3. Run migrations: `cd apps/backend && go run ./cmd/api migrate`
