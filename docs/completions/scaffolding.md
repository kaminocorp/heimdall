# Scaffolding — Completion Notes

**Date:** 2026-02-19
**Reference:** [Scaffolding Spec](../executing/scaffolding.md) | [Blueprint](../plans/blueprint.md) | [DB Models](../executing/db-models.md)

---

## What Was Done

Full monorepo scaffolding implemented per the scaffolding specification. Both frontend and backend are initialised with all directories, stub files, dependencies, and configuration in place.

---

## Root Level

| File | Notes |
|------|-------|
| `Makefile` | All targets from spec: `dev`, `dev-frontend`, `dev-backend`, `build`, `test`, `lint`, `sqlc-generate`, `migrate-up/down/create`, `docker-up/down` |
| `docker-compose.yml` | Two services (backend, frontend). No local Postgres — connects to remote DB via env var. |
| `.gitignore` | Go build artifacts, node_modules, .env files, IDE files, OS files |
| `README.md` | Project overview, quick start commands, structure reference |

---

## Frontend

**Stack:** Vue 3 + Vite + TypeScript + Pinia + Tailwind CSS v4 + Axios

### Initialisation Approach

`npm create vue@latest` was not usable non-interactively, so the project was scaffolded manually with an explicit `package.json` and config files. This produces the same result — a standard Vue 3 + Vite + TypeScript project — with full control over the structure.

### Dependencies Installed

**Runtime:** `vue`, `vue-router`, `pinia`, `axios`
**Dev:** `vite`, `@vitejs/plugin-vue`, `typescript`, `vue-tsc`, `tailwindcss`, `@tailwindcss/vite`, `eslint`, `eslint-plugin-vue`, `@vue/tsconfig`, `@tsconfig/node22`

### Directory Structure

```
frontend/src/
├── api/                    # HTTP client layer (client.ts, connections.ts, agent.ts, logs.ts, reports.ts)
├── assets/styles/          # main.css (Tailwind v4 entry via @import "tailwindcss")
├── components/
│   ├── common/             # AppHeader, AppSidebar, StatusBadge, LoadingSpinner
│   ├── connections/        # ConnectionCard, ConnectionForm, ConnectionList
│   ├── agent/              # ChatWindow, ChatMessage, ChatInput
│   ├── log/                # LogFeed, LogEntry, LogFilters
│   └── reports/            # ReportCard, ReportDetail, ReportList
├── composables/            # useWebSocket, useAgent, useAuth
├── layouts/                # DefaultLayout (sidebar + header + main)
├── pages/                  # DashboardPage, ConnectionsPage, AgentConfigPage, AgentChatPage, AgentLogPage, ReportsPage, LoginPage
├── router/                 # Vue Router with all routes (lazy-loaded)
├── stores/                 # Pinia stores (auth, connections, agent, logs)
├── types/                  # TypeScript types (connection, agent, log, report, api)
├── utils/                  # format.ts (date helpers), constants.ts
├── App.vue                 # Root component with DefaultLayout wrapper
└── main.ts                 # Entry point (creates app, registers Pinia + Router)
```

### Configuration

- `vite.config.ts` — Vue plugin, Tailwind plugin, `@` path alias, dev proxy for `/api` → `:8080` and `/ws` → `ws://localhost:8080`
- `tsconfig.json` — project references setup (tsconfig.app.json + tsconfig.node.json)
- `eslint.config.js` — flat config with vue/essential rules

### Key Design Decisions

- **API layer** (`src/api/`) isolates all HTTP calls. Each file maps to a backend resource.
- **Composables** handle WebSocket lifecycle (`useWebSocket`) and agent chat state (`useAgent`).
- **All components are functional stubs** with correct props/emits typing — ready to be wired up once the backend returns real data.
- **Tailwind v4** used with the new `@import "tailwindcss"` syntax and `@tailwindcss/vite` plugin (no `tailwind.config.js` needed — Tailwind v4 uses CSS-based config).

### Verification

- `npx vue-tsc -b --noEmit` — clean, zero type errors.

---

## Backend

**Stack:** Go 1.x + Chi v5 + pgx v5 + coder/websocket + anthropic-sdk-go + golang-migrate

### Initialisation

- Module: `github.com/crimson-sun/heimdall/backend`
- Dependencies installed via `go get` with `GOPROXY=https://proxy.golang.org,direct` (direct DNS to `gopkg.in` was timing out)

### Directory Structure

```
backend/
├── cmd/heimdall/main.go           # Entry point: HTTP server with graceful shutdown
├── internal/
│   ├── config/config.go           # Env-based config (PORT, DATABASE_URL, ANTHROPIC_API_KEY)
│   ├── api/
│   │   ├── router.go              # Chi router with all routes
│   │   ├── middleware/             # auth.go, cors.go, logging.go
│   │   └── handlers/              # connections.go, agent.go, chat.go, logs.go, reports.go, auth.go
│   ├── agent/
│   │   ├── agent.go               # Agent struct + Start/Stop lifecycle
│   │   ├── loop.go                # Tool-use loop (placeholder)
│   │   ├── monitor.go             # 24/7 monitoring goroutine (placeholder)
│   │   ├── tools.go               # Tool registry + dispatch
│   │   ├── tools_db.go            # query_database tool
│   │   ├── tools_logs.go          # search_logs tool
│   │   ├── tools_codebase.go      # search_codebase tool
│   │   ├── tools_memory.go        # recall_similar_incidents, recall_lessons tools
│   │   └── prompt.go              # System prompt + builder
│   ├── connectors/
│   │   ├── connector.go           # Connector, StreamConnector, QueryConnector interfaces
│   │   ├── registry.go            # Runtime connector registry
│   │   ├── database/postgres.go   # PostgreSQL connector stub (+test)
│   │   ├── logs/webhook.go        # Webhook log ingestion stub (+test)
│   │   ├── logs/syslog.go         # Syslog connector stub
│   │   └── codebase/github.go     # GitHub connector stub (+test)
│   ├── memory/
│   │   ├── client.go              # Elephantasm HTTP client
│   │   ├── types.go               # Event, Memory, Lesson structs
│   │   └── memory.go              # High-level memory service
│   ├── ws/
│   │   ├── hub.go                 # WebSocket hub (manages clients, broadcast)
│   │   ├── client.go              # Individual WS client
│   │   └── message.go             # Message types + envelope
│   ├── reports/
│   │   ├── generator.go           # Report builder
│   │   └── templates.go           # Report structure/sections
│   └── db/queries/                # sqlc input SQL files
│       ├── connections.sql
│       ├── agent_config.sql
│       ├── investigations.sql
│       ├── conversations.sql
│       └── log_buffer.sql
├── migrations/                     # 5 migration pairs (up/down) per DB models spec
│   ├── 001_create_connections
│   ├── 002_create_agent_config
│   ├── 003_create_investigations
│   ├── 004_create_conversations
│   └── 005_create_log_buffer
├── sqlc.yaml                       # sqlc config (pgx/v5, uuid→uuid.UUID, jsonb→json.RawMessage, timestamptz→time.Time)
├── go.mod
└── go.sum
```

### Key Design Decisions

- **Handlers return stub JSON** — all endpoints are wired and respond (200 with empty arrays, or 501 not implemented). This means the frontend can make requests immediately; real implementations plug in later.
- **WebSocket chat handler** (`handlers/chat.go`) accepts connections, reads JSON messages, and responds with a placeholder. Uses `coder/websocket` with `wsjson` for typed JSON read/write.
- **Connector interfaces** define `Connect()`, `Health()`, `Close()` as the base, with `Stream()` for one-way feeds and `Query()` for on-demand access.
- **Agent tool dispatch** is a simple switch statement in `tools.go`. Each tool gets its own `tools_*.go` file.
- **sqlc queries** follow the exact patterns from the DB models spec. Generated code goes to `internal/db/` (not yet generated — requires a running Postgres for sqlc to validate against).

### Verification

- `go build ./...` — clean, zero errors.
- `go test ./...` — 3 test files pass (connector stubs).

---

## What's Not Done (By Design)

These are intentionally deferred per the scaffolding spec:

| Item | Why |
|------|-----|
| `sqlc generate` | Requires a running Postgres instance. Queries are written; run `make sqlc-generate` once DB is available. |
| `internal/db/*.go` (generated) | Will be auto-generated by sqlc. |
| Real handler implementations | Handlers are stubs returning placeholder responses. Implementations come when DB + connectors are wired up. |
| Auth middleware wiring | `auth.go` middleware exists but is a passthrough. Real auth is deferred per MVP scope. |
| Tailwind config file | Not needed — Tailwind v4 uses CSS-based configuration via `@import "tailwindcss"`. |
| `frontend/public/favicon.ico` | Empty placeholder file created. Replace with actual icon. |

---

## How to Run

```bash
# Frontend dev server
make dev-frontend    # → localhost:5173

# Backend dev server
make dev-backend     # → localhost:8080

# Both concurrently
make dev

# Run backend tests
cd backend && go test ./...

# Type-check frontend
cd frontend && npx vue-tsc -b --noEmit

# Generate sqlc code (once DB is available)
make sqlc-generate

# Run migrations (once DATABASE_URL is set)
make migrate-up
```
