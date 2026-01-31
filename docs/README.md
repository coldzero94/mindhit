# MindHit Documentation

Welcome to the MindHit documentation.

## Quick Start

- [Self-Hosting Guide](./SELF_HOSTING.md) - Get started in 5 minutes with Docker
- [Configuration](./CONFIGURATION.md) - Environment variables and settings
- [Chrome Extension](./EXTENSION_INSTALL.md) - Build and install the extension

## What is MindHit?

MindHit is a self-hosted browsing activity tracker that generates AI-powered mindmaps from your browsing sessions. It collects your browsing history and uses AI to extract key topics and relationships.

### Features

- **Session Recording**: Track browsing activity in real-time
- **AI-Powered Analysis**: Extract keywords and generate mindmaps from your browsing data
- **Visual Mindmaps**: Interactive 3D visualization of topics and relationships
- **Chrome Extension**: Easy-to-use browser extension for recording sessions
- **Self-Hosted**: Run on your own infrastructure, keep your data private

## Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ Chrome Extension│────▶│   Go API Server │────▶│   PostgreSQL    │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Background     │
                        │  Worker (Asynq) │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  AI Provider    │
                        │  (Groq, etc.)   │
                        └─────────────────┘
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.24+ with Gin, Ent ORM |
| Frontend | Next.js 16.1, React, Three.js |
| Extension | Vite, React, CRXJS |
| Database | PostgreSQL 16 |
| Queue | Redis + Asynq |
| AI | Groq (default), OpenAI, Gemini, Claude |

## Internal Documentation

For development phase documents and internal guides (Korean), see [docs/internal/](./internal/).
