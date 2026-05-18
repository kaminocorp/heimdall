# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Heimdall?

Autonomous AI monitoring agent for production applications. It watches systems 24/7 via a background monitoring loop, classifies logs with an ONNX model (Lumber), escalates flagged entries to an LLM for assessment, and provides interactive chat for on-demand investigation. Multi-provider (Anthropic + OpenRouter) selectable per application.

See `docs/vision.md` for the product vision and `docs/changelog.md` for release history (treat the changelog as the source of truth for current state — versions move fast).

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
make test                                              # All tests (backend + frontend)
cd backend && go test ./...                            # All backend tests
cd backend && go test ./internal/agent                 # Single backend package
cd backend && go test ./internal/api/handlers -run TestCreateConnection  # Single test
cd frontend && npm run test                            # All frontend tests (vitest, watch mode)
cd frontend && npm run test -- --run                   # Single non-watch run (used in CI)
cd frontend && npm run test:watch                      # Frontend watch mode
```

Backend integration tests require `DATABASE_URL`; they skip cleanly when missing. A handful of RLS regression tests additionally gate on `HEIMDALL_ROLE_SPLIT_LIVE` / `HEIMDALL_RLS_FORCE_LIVE` so they only run against a database that has the role split applied.

### Lint
```bash
make lint                    # Both backend + frontend
cd backend && go vet ./...
cd frontend && npm run lint  # ESLint (test files have no-explicit-any disabled)
```

### Database
```bash
# Migrate commands use DIRECT_URL (falls back to DATABASE_URL when unset).
# make dev-backend auto-sources .env, but migrate commands do not — source it first:
#   set -a && . ./.env && set +a && make migrate-up
make migrate-up        # Apply all pending migrations
make migrate-down      # Rollback one migration
make migrate-create    # Interactive: create new migration pair
make sqlc-generate     # Regenerate Go code from SQL queries

# Production-shape only — bootstrap runtime roles before first migrate-up.
# Skip in dev where the local Postgres user is a superuser.
APP_USER_PASSWORD=... CRON_USER_PASSWORD=... make bootstrap-roles
```

`make sqlc-generate` must run on the pinned sqlc version — v1.30 changed nullable-inference for FILTER aggregates and `SECURITY DEFINER` returns, which silently rewrites a few generated files. If a diff appears in `*.sql.go` you didn't intend, check your local sqlc version against the toolchain.

### Database URLs (RLS role split)

Three URLs, three purposes — the RLS role split (v0.48.x) is fully shipped. Production runtime authenticates as the two non-superuser roles below; `DATABASE_URL` is the only required one in dev, with the others falling back transparently:

- **`DATABASE_URL`** — runtime app-pool URL. Production: `app_user` (per-tenant CRUD; `FORCE ROW LEVEL SECURITY` is the access boundary, binding even the table owner).
- **`CRON_DATABASE_URL`** — runtime cron-pool URL for cross-tenant enumeration. Empty → falls back to `DATABASE_URL` with a startup WARN. Production: `cron_user` (BYPASSRLS, narrow grants — `INSERT/UPDATE` on `monitoring_state`, `DELETE` on `log_buffer`, blanket `SELECT`).
- **`DIRECT_URL`** — superuser URL used by `make migrate-*`. Empty → falls back to `DATABASE_URL`. Production: `postgres` (the only path that retains DDL privileges; never serves runtime traffic).

`HEIMDALL_ENV=production` enables a startup invariant that refuses to launch when `DATABASE_URL` and `CRON_DATABASE_URL` authenticate as the same role. Set it in production; leave unset in dev (where all three URLs typically point at the local `postgres` superuser).

**Supabase pooler mode is load-bearing.** `SET LOCAL` and `app.current_user_id` are void in transaction-pooling mode — RLS would silently return zero rows for everything. The pooler must be in *session pooling* mode.

### Infrastructure
```bash
docker compose up -d postgres   # Start local Postgres
make docker-up                  # Start all Docker services
make docker-down                # Stop all Docker services
```

## Architecture

### Backend (Go)

**Module**: `github.com/hejijunhao/heimdall/backend`

The backend is a Go HTTP server using Chi router. Key subsystem directories under `backend/internal/`: `agent/` (loop, monitor, scheduler, pipeline, classifier, providers, tools), `api/handlers/`, `connectors/`, `db/` (pools + sqlc-generated queries), `config/`, `notifications/`, `metrics/`, `reports/`, `memory/`, `github/`, `ws/`.

#### Agent — three operating modes

- **Interactive mode** (`agent/loop.go`, `agent/loop_stream.go`): User-initiated via WebSocket. Tool-use loop with `search_logs`, `query_database`, `inspect_codebase` (GitHub). Conversation history persisted as JSONB in the conversations table.
- **Monitoring mode** (`agent/monitor.go`): Background goroutine polling every 15 seconds. For each active app: fetches new logs since cursor → classifies via Lumber ONNX (`agent/classifier_lumber.go`) → escalates flagged logs to an LLM provider (`provider_anthropic.go` or `provider_openrouter.go`) → emits results to `agent_log`. Concurrency limited by semaphore. Cursor tracked in `monitoring_state`.
- **Scheduled mode** (`agent/scheduler.go`): Cron-driven stored prompts (`investigation_schedules` table), useful for pull-only connectors that the monitoring loop can't watch passively.

#### Classifier pipeline

`agent/classifier.go` is an interface with two implementations: `LumberClassifier` (ONNX, threshold + hardcoded escalation rules in `severity_gate.go`) and `PassthroughClassifier` (escalates everything). Controlled by `CLASSIFIER_MODE` env var: `on`, `off`, or `fallback`. The pipeline writes structured events through `pipeline_bus.go` / `pipeline_writer.go`, which feed the Pipeline page's Sankey funnel and the time-machine replay.

#### Connectors (`internal/connectors/`)

Interface hierarchy: `Connector` (base) → `StreamConnector` (logs in) / `QueryConnector` (DB queries out). Implementations include webhook, Syslog TLS, OTLP HTTP, Fly.io drain, API pollers, Postgres (read-only — enforces `default_transaction_read_only=on`), Supabase, and GitHub App for code context. SDKs that emit to the webhook ingestion endpoint live in `packages/sdk-{go,js,python}`.

#### Source filtering & org-scoped connections (v0.46.x)

Connections were promoted from per-app to **org-level** so several apps can share one webhook endpoint. Per-app visibility is governed by `connection_sources` (which sources a connection exposes) and `app_source_filters` (which sources each app subscribes to). Anything that resolves an app from a connection needs to pass through `resolveSourceFilterApp` / equivalent — it validates org membership *and* the explicit `app.OrgID != conn.OrgID` cross-org refusal.

#### Database access patterns

Two transaction shapes — picking the right one is load-bearing:

- **`UserQueries(ctx, userID)`** — short HTTP handlers. Begins a txn, runs `SET LOCAL ROLE app_user` + `SET LOCAL app.current_user_id = userID` for RLS, returns a `done` func. Inherits Supabase's default per-statement `statement_timeout` (~8s) and `idle_in_transaction_session_timeout`.
- **`UserQueriesForLoop(ctx, userID)`** — agent loop only. Same role/GUC setup *plus* `SET LOCAL statement_timeout = 0` and `idle_in_transaction_session_timeout = 0` so the txn survives Claude's between-tool-call thinking time and slow tool queries. **Don't share a single loop txn across goroutines** — `pgx.Tx` is not goroutine-safe; consumer goroutines must fully drain producer events before the rollback `defer` fires (see `chat.go` consumer-loop comment).

Other patterns worth knowing:

- All DB queries are user-scoped via `user_id` or joined through the org hierarchy.
- `authorizeApp` validates an app belongs to the user's org via `GetApplicationByOrgUser`.
- Webhook / OTLP ingest re-checks connection ownership *inside* the user-scoped txn (TOCTOU guard against tokens-still-valid-but-user-removed cases — returns 401 explicitly).
- Agent tool errors are returned as `isError: true` tool results to Claude (not thrown).
- `EmitLog` is fire-and-forget — errors logged, never propagated.
- Webhook post-commit idempotency-cache writes use `context.WithoutCancel(r.Context())` with a 5s timeout because webhook senders disconnect right after reading the response status.

#### Generated code

sqlc generates type-safe Go from SQL in `internal/db/queries/*.sql`. Config in `backend/sqlc.yaml` maps uuid→`google/uuid.UUID`, jsonb→`json.RawMessage`, timestamptz→`time.Time`. Migrations live in `backend/migrations/` (currently in the 040s — the RLS role split lives in 039–041).

### Frontend (Vue 3 + TypeScript)

Vite-based SPA with Tailwind CSS v4 (techno-brutalist design: near-black canvas, muted green accent, JetBrains Mono). Vite proxies `/api` and `/ws` to `localhost:8080` in dev.

- **Store pattern** — all Pinia stores use the composition API (`defineStore` with a setup function). Stores include `auth`, `app` (org/apps/currentAppId persisted to localStorage), `connections`, `logs`, `agent`, `reports`, `pipeline`, `activity`, `notifications`, `scheduled`, plus org/team and onboarding stores.
- **Auth** — Supabase JS SDK → JWT in session → Axios interceptor injects `Authorization: Bearer` → 401 triggers logout. WebSocket auth via `?token=` query param.
- **WebSocket chat** — `useWebSocket` (connection wrapper) → `useAgent` (message parsing, thinking state, conversation hydration). Messages typed as system/status/error/chat.
- **API client** — `api/client.ts` is an Axios instance with `/api` base URL.
- **Tests** — Vitest with happy-dom. `src/test/setup.ts` creates fresh Pinia per test and globally mocks the Axios client. Test files are scoped to `no-explicit-any: 'off'` via the root `eslint.config.js` override — tests reach into Vue VM internals where `any` is pragmatic.

### Data model hierarchy

```
User (Supabase auth.users)
  └─ Organization (org_members for multi-org membership)
       ├─ Connection (1:N, org-scoped) — log sources, databases, code
       │   └─ connection_sources (per-source declarations)
       └─ Application (1:N)
            ├─ app_source_filters — which (connection, source) pairs this app sees
            ├─ AppAgentConfig (1:1) — provider, model, mode, schedule, prompt override
            └─ MonitoringState (1:1) — last_monitored_at cursor
```

Activity feed (`agent_log`) is the chronological master feed combining raw events, agent observations, and rule-based entries. Reports are produced when the agent finishes investigating an escalation.

### Docs as a workflow

Convention used throughout the project — when planning or reviewing, reach for the right folder:

- `docs/vision.md` — product vision (rarely changes).
- `docs/changelog.md` — single growing changelog, newest first. Source of truth for current state.
- `docs/blueprints/` — design docs for upcoming or in-progress work.
- `docs/executing/` — operator-facing runbooks for in-flight rollouts (e.g. `rls-enforcement-next-steps.md`).
- `docs/completions/` — per-phase post-ship records ("what / where / why" depth beyond the changelog).
- `docs/archive/` — retired plans and runbooks.

## Repository layout

```
heimdall/
├── frontend/                 # Vue 3 + Vite + TypeScript
├── backend/
│   ├── cmd/heimdall/         # entrypoint (main.go)
│   ├── internal/             # subsystems (see Architecture)
│   ├── migrations/           # SQL migrations (numbered, .up.sql / .down.sql pairs)
│   └── sqlc.yaml             # sqlc config
├── packages/sdk-{go,js,python}  # client SDKs that POST to the webhook endpoint
├── docs/                     # vision, changelog, blueprints, executing, completions, archive
├── docker-compose.yml
└── Makefile                  # root task runner
```
