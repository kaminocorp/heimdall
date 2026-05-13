# Heimdall

Autonomous AI monitoring agent for production applications.

Heimdall replaces passive, noisy alerting with an intelligent agent that watches your systems 24/7, classifies incoming logs with an ONNX model (Lumber), escalates anomalies to an LLM for investigation, and surfaces only what matters — on a chronological agent log, as incident reports, or via interactive chat.

## Features

- **Monitoring mode** — background loop polls connected apps every 15s, classifies logs via Lumber, escalates flagged entries to the LLM for assessment.
- **Interactive mode** — WebSocket chat with a tool-using agent (`search_logs`, `query_database`).
- **Multi-provider LLM** — Anthropic (Claude) and OpenRouter (GPT-4o, Gemini, Llama, direct Claude, …) selectable per application.
- **Multi-app** — one organization, many applications, each with its own connections, agent config, and monitoring cursor.
- **Ingestion** — webhook, OTLP HTTP, Syslog TLS, API pollers, and Python / Go / JS SDKs.
- **Connectors** — read-only Postgres (incl. Supabase), GitHub App for code context.
- **Notifications** — email escalation via Resend.
- **Production hygiene** — health endpoint, Prometheus metrics, JSON-structured logs, Postgres RLS.

See [docs/vision.md](docs/vision.md) for the product vision and [docs/changelog.md](docs/changelog.md) for release history.

## Prerequisites

- **Go** 1.25+
- **Node.js** 22+
- **Docker** with Compose v2 (for local Postgres)
- **golang-migrate** CLI (`migrate`)
- **sqlc** (for regenerating Go from SQL)

## Setup

```bash
# 1. Copy env file and fill in your keys (Anthropic and/or OpenRouter, Supabase)
cp .env.example .env

# 2. Start Postgres
docker compose up -d postgres

# 3. (Production-shape only) Bootstrap the runtime roles. Skip in dev — the
#    local Postgres role created by docker compose is a superuser and the
#    migrations work against it directly. See "Database URLs" below.
make bootstrap-roles

# 4. Run migrations
DATABASE_URL=postgres://heimdall:heimdall@localhost:5432/heimdall?sslmode=disable make migrate-up

# 5. Install frontend deps
cd frontend && npm install && cd ..

# 6. Start dev servers (frontend on :5173, backend on :8080)
make dev
```

## Database URLs

Heimdall splits database access across three URLs in production. Dev runs with
just `DATABASE_URL` set; the others fall back transparently:

| Var                    | Purpose                                           | Prod role          | Dev fallback           |
| ---------------------- | ------------------------------------------------- | ------------------ | ---------------------- |
| `DATABASE_URL`         | Per-request app-pool traffic, RLS-enforced        | `app_user`         | required               |
| `CRON_DATABASE_URL`    | Cross-tenant background work (monitor / pruner)   | `cron_user` (BYPASSRLS, narrow grants) | falls back to `DATABASE_URL` with a startup WARN |
| `DIRECT_URL`           | DDL / `make migrate-*` only                       | `postgres` (superuser) | falls back to `DATABASE_URL` |

`HEIMDALL_ENV=production` enables a startup invariant that refuses to launch
when `DATABASE_URL` and `CRON_DATABASE_URL` resolve to the same role. Set
this in production; leave unset in dev. See the **Database URLs (RLS role
split)** section in [CLAUDE.md](CLAUDE.md) for the full role / grant table
and the rationale behind the split (8-phase rollout, archived runbook in
`docs/archive/rls-enforcement-roadmap.md`).

## Commands

```bash
# Dev
make dev              # Frontend + backend concurrently
make dev-frontend     # Vite dev server only
make dev-backend      # Go backend only (loads .env)

# Build
make build            # Both
make build-frontend   # vue-tsc + vite build
make build-backend    # backend/bin/heimdall

# Test
make test             # Backend + frontend
cd backend && go test ./...
cd frontend && npm run test

# Lint
make lint             # go vet + eslint

# Database (uses DIRECT_URL, falls back to DATABASE_URL)
make migrate-up
make migrate-down
make migrate-create
make sqlc-generate

# Docker
make docker-up
make docker-down
```

Backend integration tests that need Postgres skip automatically when `DATABASE_URL` is unset.

## Architecture

```
User (Supabase auth)
  └─ Organization
       └─ Application
            ├─ Connection (logs, databases, code)
            ├─ AppAgentConfig (provider, model, mode, schedule)
            └─ MonitoringState (cursor)
```

- **Backend** — Go (Chi, pgxpool, sqlc). Agent loop in `backend/internal/agent`, handlers in `backend/internal/api/handlers`, connectors in `backend/internal/connectors`.
- **Frontend** — Vue 3 + TypeScript + Vite + Tailwind v4, Pinia stores, Supabase JS SDK for auth.
- **Database** — PostgreSQL (Supabase in prod), migrations in `backend/migrations/`, RLS policies scoped by `app.current_user_id`.
- **Classifier** — Lumber ONNX model on the hot path; only flagged logs reach the LLM.

See [CLAUDE.md](CLAUDE.md) for deeper architectural notes.

## Project Structure

```
heimdall/
├── frontend/      # Vue 3 + Vite + TypeScript
├── backend/       # Go backend (cmd/, internal/, migrations/)
├── docs/          # Vision, changelog, plans, completions
├── docker-compose.yml
└── Makefile       # Root-level task runner
```
