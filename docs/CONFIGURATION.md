# Configuration Guide

All configuration is done through environment variables in the root `.env` file.

## Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `API_SECRET_KEY` | Secret key for API authentication | `openssl rand -hex 32` |
| `GROQ_API_KEY` | Groq API key for AI features | `gsk_xxx...` |

## Backend Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `API_PORT` | `9000` | API server port |
| `ENVIRONMENT` | `development` | Environment name |
| `DATABASE_URL` | - | PostgreSQL connection string |
| `REDIS_URL` | - | Redis connection string |
| `WORKER_CONCURRENCY` | `10` | Number of worker goroutines |

### Database URL Format

```
postgres://user:password@host:port/database?sslmode=disable
```

Example for local Docker:
```
DATABASE_URL=postgres://postgres:password@localhost:5433/mindhit?sslmode=disable
```

## AI Provider Configuration

MindHit uses AI for tag extraction and mindmap generation. At least one API key is required.

| Variable | Provider | Notes |
|----------|----------|-------|
| `GROQ_API_KEY` | Groq | Recommended - Free tier available |
| `OPENAI_API_KEY` | OpenAI | GPT models |
| `GEMINI_API_KEY` | Google | Gemini models |
| `CLAUDE_API_KEY` | Anthropic | Claude models |

### Getting a Groq API Key (Recommended)

1. Go to https://console.groq.com
2. Create an account
3. Generate an API key
4. Set `GROQ_API_KEY` in your `.env`

Groq offers a generous free tier with fast inference using Llama models.

## Frontend Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:9000` | Backend API URL |
| `NEXT_PUBLIC_API_KEY` | - | API key (same as `API_SECRET_KEY`) |

## Extension Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_URL` | `http://localhost:9000/v1` | Backend API URL |
| `VITE_WEB_APP_URL` | `http://localhost:3000` | Web app URL |
| `VITE_API_KEY` | - | API key (same as `API_SECRET_KEY`) |
| `VITE_EVENT_BATCH_SIZE` | `1` | Events per batch |
| `VITE_EVENT_FLUSH_INTERVAL` | `5000` | Flush interval (ms) |

### Event Batching

For production, consider increasing batch settings:

```bash
VITE_EVENT_BATCH_SIZE=10
VITE_EVENT_FLUSH_INTERVAL=30000
```

## Docker Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | `postgres` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `password` | PostgreSQL password |
| `POSTGRES_DB` | `mindhit` | Database name |

## Monitoring (Optional)

| Variable | Default | Description |
|----------|---------|-------------|
| `GF_SECURITY_ADMIN_USER` | `admin` | Grafana admin username |
| `GF_SECURITY_ADMIN_PASSWORD` | `admin` | Grafana admin password |
| `SLACK_WEBHOOK_URL` | - | Slack webhook for alerts |

## Example .env File

```bash
# Required
API_SECRET_KEY=your-32-character-random-string
GROQ_API_KEY=gsk_your_key

# Backend
API_PORT=9000
DATABASE_URL=postgres://postgres:password@localhost:5433/mindhit?sslmode=disable
REDIS_URL=redis://localhost:6380

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:9000
NEXT_PUBLIC_API_KEY=your-32-character-random-string

# Extension
VITE_API_URL=http://localhost:9000/v1
VITE_WEB_APP_URL=http://localhost:3000
VITE_API_KEY=your-32-character-random-string
```
