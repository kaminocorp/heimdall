# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Heimdall?

Autonomous AI monitoring agent for production applications. It watches systems 24/7 via a background monitoring loop, classifies logs with an ONNX model (Lumber), escalates flagged entries to Claude for assessment, and provides interactive chat for on-demand investigation.

## Commands

### Development
```bash
make dev              # Run frontend + backend concurrently
make dev-frontend     # Frontend only (Vite dev server on :5173)
make dev-backend      # Backend only (loads .env, runs on :8080)
```

### Build
```bash
make build-frontend   # vue-tsc + vite build
make build-backend    # go build → backend/bin/heimdall
```

### Test
```bash
make test                              # All tests (backend + frontend)
cd backend && go test ./...            # All backend tests
cd backend && go test ./internal/agent # Single backend package
cd backend && go test ./internal/api/handlers -run TestCreateConnection  # Single test
cd frontend && npm run test            # All frontend tests (vitest)
cd frontend && npm run test:watch      # Frontend watch mode
```

Backend integration tests require `DATABASE_URL` env var; they skip automatically if missing.

### Lint
```bash
make lint                    # Both backend + frontend
cd backend && go vet ./...   # Go vet
cd frontend && npm run lint  # ESLint
```

### Database
```bash
# Requires DATABASE_URL env var (defined in root .env).
# make dev-backend auto-sources .env, but migrate commands do not.
# Source it first: set -a && . ./.env && set +a && make migrate-up
make migrate-up        # Apply all pending migrations
make migrate-down      # Rollback one migration
make migrate-create    # Interactive: create new migration pair
make sqlc-generate     # Regenerate Go code from SQL queries
```

### Infrastructure
```bash
docker compose up -d postgres   # Start local Postgres
make docker-up                  # Start all Docker services
make docker-down                # Stop all Docker services
```

## Architecture

### Backend (Go)

**Module**: `github.com/hejijunhao/heimdall/backend`

The backend is a Go HTTP server using Chi router, with two main subsystems:

**Agent — Two Operating Modes:**
- **Interactive mode** (`agent/loop.go`): User-initiated via WebSocket. Runs a Claude tool-use loop (max 10 iterations) with `search_logs` and `query_database` tools. Conversation history persisted as JSONB in the conversations table.
- **Monitoring mode** (`agent/monitor.go`): Background goroutine polling every 15 seconds. For each active app: fetches new logs since cursor → classifies via Lumber ONNX (`agent/classifier_lumber.go`) → escalates flagged logs to Claude for assessment → emits results to agent_log. Concurrency limited to 10 apps via semaphore.

**Classifier pipeline** (`agent/classifier.go`): Interface with two implementations — `LumberClassifier` (ONNX model, confidence threshold 0.5, hardcoded escalation rules in `severity_gate.go`) and `PassthroughClassifier` (escalates everything). Controlled by `CLASSIFIER_MODE` env var: `on`, `off`, or `fallback`.

**Key architectural patterns:**
- All DB queries are user-scoped via `user_id` or joined through the org hierarchy (User → Organization → Application → Connection)
- `authorizeApp` helper in handlers validates app belongs to user's org via `GetApplicationByOrgUser` query
- `UserQueries()` begins a transaction and sets `SET LOCAL app.current_user_id` for PostgreSQL RLS policies
- Agent tool errors are returned as `isError: true` tool results to Claude (not thrown)
- `EmitLog` is fire-and-forget — errors logged, never propagated

**Database layer**: sqlc generates type-safe Go from SQL in `internal/db/queries/*.sql`. Migrations in `backend/migrations/` (001–016). Config in `sqlc.yaml` maps uuid→`google/uuid.UUID`, jsonb→`json.RawMessage`, timestamptz→`time.Time`.

**Connectors** (`internal/connectors/`): Interface hierarchy — `Connector` (base), `StreamConnector` (logs), `QueryConnector` (database). Postgres connector enforces `default_transaction_read_only=on`.

### Frontend (Vue 3 + TypeScript)

Vite-based SPA with Tailwind CSS v4 (techno-brutalist design: near-black backgrounds, muted green accent `#5a9e6a`, JetBrains Mono).

**Store pattern**: All Pinia stores use composition API (`defineStore` with setup function). Key stores: `auth` (Supabase session), `app` (org/apps/currentAppId persisted to localStorage), `connections`, `logs`, `agent`, `reports`.

**Auth flow**: Supabase JS SDK → JWT stored in session → Axios interceptor injects `Authorization: Bearer` header → 401 responses trigger logout. WebSocket auth via `?token=` query param.

**WebSocket chat**: `useWebSocket` (low-level connection wrapper) → `useAgent` (message parsing, thinking state, conversation hydration). Messages typed as system/status/error/chat.

**API client** (`api/client.ts`): Axios instance with `/api` base URL. Vite proxies `/api` and `/ws` to `localhost:8080` in dev.

**Test setup**: Vitest with happy-dom environment. `src/test/setup.ts` creates fresh Pinia per test and globally mocks the Axios client.

### API Routes

All protected routes require Supabase JWT (except webhook ingestion which uses bearer token).

```
POST /api/webhooks/logs              — Log ingestion (bearer token auth)
GET  /api/org                        — User's organization
POST /api/onboard                    — Create org + first app (idempotent, 409 if exists)
GET|POST /api/apps                   — List/create applications
GET  /api/apps/{appId}/*             — Per-app: connections, agent/config, monitoring/status, stats
CRUD /api/connections                — Connection management
GET  /api/logs                       — Paginated log queries
GET  /api/conversations[/{id}]       — Conversation history
GET  /api/reports[/{id}]             — Investigation reports
GET  /ws/chat                        — WebSocket (token in query param)
```

### Data Model Hierarchy

```
User (Supabase auth.users)
  └─ Organization (1:1 via org_id on users)
       └─ Application (1:N)
            ├─ Connection (1:N) — log sources, databases
            ├─ AppAgentConfig (1:1) — model, mode, schedule, prompt override
            └─ MonitoringState (1:1) — cursor tracking (last_monitored_at)
```
