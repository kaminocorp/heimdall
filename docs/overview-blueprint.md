# Heimdall — Overview Blueprint

A living reference of the current project architecture and structure. Updated as the codebase evolves.

**Last updated:** 2026-02-19 (v0.1.0 — scaffolding)

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Frontend | Vue 3 (Composition API) + Vite + TypeScript |
| Styling | Tailwind CSS v4 |
| State | Pinia |
| HTTP client | Axios |
| Backend | Go + Chi v5 |
| Database driver | pgx v5 |
| Query layer | sqlc (code generation from SQL) |
| Migrations | golang-migrate |
| WebSocket | coder/websocket |
| LLM | anthropic-sdk-go (Claude) |
| Agent memory | Elephantasm (planned, post-MVP) |

---

## Monorepo Structure

```
heimdall/
├── frontend/                  Vue 3 + Vite + TypeScript
├── backend/                   Go backend
├── docs/                      All documentation
│   ├── plans/                 Design docs & proposals
│   ├── executing/             Implementation specs
│   └── completions/           Post-implementation notes
├── Makefile                   Root-level task runner
├── docker-compose.yml         Containerised local dev
├── .gitignore
└── README.md
```

---

## Frontend — `frontend/`

### Architecture

```
src/
├── api/               HTTP client layer — one file per backend resource
├── assets/styles/     Global CSS (Tailwind entry)
├── components/        Reusable UI components, grouped by domain
│   ├── common/        Shared: AppHeader, AppSidebar, StatusBadge, LoadingSpinner
│   ├── connections/   ConnectionCard, ConnectionForm, ConnectionList
│   ├── agent/         ChatWindow, ChatMessage, ChatInput
│   ├── log/           LogFeed, LogEntry, LogFilters
│   └── reports/       ReportCard, ReportDetail, ReportList
├── composables/       Vue composables (useWebSocket, useAgent, useAuth)
├── layouts/           Page layouts (DefaultLayout)
├── pages/             Route-level views (one per route, lazy-loaded)
├── router/            Vue Router configuration
├── stores/            Pinia stores (auth, connections, agent, logs)
├── types/             TypeScript type definitions
├── utils/             Helpers (format, constants)
├── App.vue            Root component
└── main.ts            Entry point
```

### Routing

| Path | Page | Description |
|------|------|-------------|
| `/` | DashboardPage | System overview |
| `/connections` | ConnectionsPage | Manage integrations |
| `/agent/config` | AgentConfigPage | Agent model & behaviour settings |
| `/agent/chat` | AgentChatPage | Real-time chat with the agent |
| `/agent/log` | AgentLogPage | Chronological master feed |
| `/reports` | ReportsPage | Incident reports |
| `/login` | LoginPage | Authentication |

### Data Flow

```
Pages → Stores (Pinia) → API layer (Axios) → Backend REST API
Pages → Composables (useAgent) → WebSocket → Backend WS endpoint
```

---

## Backend — `backend/`

### Architecture

```
cmd/heimdall/main.go          Entry point: HTTP server + graceful shutdown

internal/
├── config/                    Env-based configuration
├── api/                       HTTP layer
│   ├── router.go              Chi route definitions
│   ├── middleware/             auth, cors, logging
│   └── handlers/              One file per resource (connections, agent, chat, logs, reports, auth)
├── agent/                     Agent engine
│   ├── agent.go               Struct + lifecycle (Start/Stop)
│   ├── loop.go                Core Claude tool-use loop
│   ├── monitor.go             24/7 monitoring goroutine
│   ├── tools.go               Tool registry + dispatch
│   ├── tools_db.go            query_database tool
│   ├── tools_logs.go          search_logs tool
│   ├── tools_codebase.go      search_codebase tool
│   ├── tools_memory.go        recall_similar_incidents, recall_lessons tools
│   └── prompt.go              System prompt construction
├── connectors/                External integration layer
│   ├── connector.go           Interfaces: Connector, StreamConnector, QueryConnector
│   ├── registry.go            Runtime connector registry
│   ├── database/              PostgreSQL connector
│   ├── logs/                  Webhook + Syslog connectors
│   └── codebase/              GitHub connector
├── memory/                    Elephantasm integration
│   ├── client.go              HTTP client
│   ├── types.go               Event, Memory, Lesson structs
│   └── memory.go              High-level memory service
├── ws/                        WebSocket management
│   ├── hub.go                 Connection hub + broadcast
│   ├── client.go              Individual client
│   └── message.go             Message types + envelope
├── reports/                   Report generation
│   ├── generator.go           Report builder
│   └── templates.go           Report structure/sections
└── db/queries/                sqlc input SQL (5 query files)
```

### API Routes

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| GET | `/api/connections` | ListConnections | stub |
| POST | `/api/connections` | CreateConnection | stub |
| GET | `/api/connections/{id}` | GetConnection | stub |
| PUT | `/api/connections/{id}` | UpdateConnection | stub |
| DELETE | `/api/connections/{id}` | DeleteConnection | stub |
| GET | `/api/agent/config` | GetAgentConfig | stub |
| PUT | `/api/agent/config` | UpdateAgentConfig | stub |
| GET | `/api/logs` | ListLogs | stub |
| GET | `/api/reports` | ListReports | stub |
| GET | `/api/reports/{id}` | GetReport | stub |
| GET | `/api/auth/login` | Login | stub |
| WS | `/ws/chat` | HandleChat | stub (echo) |

### Connector Interfaces

```go
Connector           → Connect(), Health(), Close()
StreamConnector     → Connector + Stream(ctx, chan<- []byte)
QueryConnector      → Connector + Query(ctx, query) (any, error)
```

### Agent Tools

| Tool | Description | File |
|------|-------------|------|
| `query_database` | Execute read-only SQL against monitored DB | tools_db.go |
| `search_logs` | Search recent logs for patterns | tools_logs.go |
| `search_codebase` | Search code via GitHub API | tools_codebase.go |
| `recall_similar_incidents` | Query Elephantasm for past incidents | tools_memory.go |
| `recall_lessons` | Retrieve learned lessons from Elephantasm | tools_memory.go |

---

## Database

5 tables, managed via golang-migrate migrations in `backend/migrations/`.

| Table | Purpose | Key columns |
|-------|---------|-------------|
| `connections` | Integration config | name, type, direction, config (JSONB), status |
| `agent_config` | Agent behaviour | model, mode, schedule, system_prompt_override |
| `investigations` | Agent findings | trigger_type, summary, severity, status, context/findings/tool_trace (JSONB) |
| `conversations` | Chat sessions | investigation_id (FK), messages (JSONB array) |
| `log_buffer` | Rolling 24h debug buffer | connection_id (FK), source_type, severity, payload (JSONB) |

Query layer: sqlc. SQL queries in `backend/internal/db/queries/`. Generated Go code goes to `backend/internal/db/` (not yet generated — requires running Postgres).

---

## Key Commands

```bash
make dev                # Run frontend + backend concurrently
make dev-frontend       # Frontend only (Vite → :5173)
make dev-backend        # Backend only (Go → :8080)
make build              # Production build (both)
make test               # Run all tests
make lint               # Lint both projects
make sqlc-generate      # Regenerate Go code from SQL
make migrate-up         # Run DB migrations
make migrate-down       # Rollback last migration
```

---

## What's Implemented vs Planned

| Component | Status |
|-----------|--------|
| Project structure | Done |
| Frontend components (stubs) | Done |
| Frontend routing + stores | Done |
| Backend HTTP server + routes | Done |
| Connector interfaces | Done |
| Agent tool registry + dispatch | Done |
| WebSocket chat (echo) | Done |
| Migrations + sqlc queries | Done (SQL written, not generated) |
| Real handler implementations | Pending — needs DB |
| Agent tool-use loop (Claude) | Pending |
| Monitoring goroutine | Pending |
| Elephantasm integration | Pending (post-MVP) |
| Authentication | Pending (deferred per MVP) |
