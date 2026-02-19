# Heimdall

Autonomous AI monitoring agent for production applications.

Heimdall replaces passive, noisy alerting with an intelligent agent that watches your systems 24/7, investigates anomalies autonomously, and surfaces only what matters.

## Prerequisites

- **Go** 1.24+
- **Node.js** 22+
- **Docker** with Compose v2 (for Postgres)
- **golang-migrate** CLI (`migrate`)
- **sqlc** (for code generation from SQL)

## Setup

```bash
# 1. Copy env file and fill in your Anthropic API key
cp .env.example .env

# 2. Start Postgres
docker compose up -d postgres

# 3. Run migrations
DATABASE_URL=postgres://heimdall:heimdall@localhost:5432/heimdall?sslmode=disable make migrate-up

# 4. Install frontend deps
cd frontend && npm install && cd ..

# 5. Start dev servers
make dev
```

## Quick Start

```bash
# Frontend only
make dev-frontend

# Backend only
make dev-backend

# Both concurrently
make dev
```

## Project Structure

```
heimdall/
├── frontend/    # Vue 3 + Vite + TypeScript
├── backend/     # Go backend
├── docs/        # Documentation
└── Makefile     # Root-level task runner
```

See [docs/vision.md](docs/vision.md) for the full product vision.
