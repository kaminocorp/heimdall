# Heimdall — Overview Blueprint

A living reference of the current project architecture and structure. Updated as the codebase evolves.

**Last updated:** 2026-04-19 (v0.47.4 — Pipeline page, org-scoped connections, source filtering, RLS everywhere)

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Frontend | Vue 3 (Composition API) + Vite + TypeScript |
| Styling | Tailwind CSS v4 (techno-brutalist, JetBrains Mono, muted-green `#5a9e6a` on near-black) |
| State | Pinia (composition-API `defineStore` with setup functions) |
| HTTP client | Axios (with JWT interceptor, 401 → logout) |
| Testing (FE) | Vitest + happy-dom |
| Backend | Go 1.25 + Chi v5 |
| Database driver | pgx v5 (`pgxpool`) |
| Query layer | sqlc (code generation from SQL) |
| Migrations | golang-migrate |
| WebSocket | coder/websocket |
| LLM SDK | anthropic-sdk-go + OpenRouter HTTP provider |
| Classifier | Lumber ONNX model (in-process, `onnxruntime_go`) |
| Auth | Supabase (JWT verified against Supabase JWKS) |
| Observability | slog + Prometheus (`/metrics` behind auth) |
| Notifications | Slack / Discord webhooks + SMTP email |
| GitHub integration | GitHub App (JWT-signed installation tokens) |

---

## Monorepo Structure

```
heimdall/
├── frontend/                  Vue 3 + Vite + TypeScript
├── backend/                   Go backend
├── docs/                      All documentation
│   ├── blueprints/            Component-level blueprints
│   ├── plans/                 Design docs & proposals
│   ├── executing/             In-flight implementation specs
│   ├── completions/           Post-implementation notes
│   ├── archive/               Historical runbooks & plans
│   ├── vision.md              Product vision
│   ├── changelog.md           Release-by-release history
│   └── overview-blueprint.md  (this file)
├── Makefile                   Root-level task runner
├── docker-compose.yml         Local Postgres + services
├── fly.toml                   Fly.io deployment
├── .env / .env.example
└── README.md
```

---

## Frontend — `frontend/`

### Architecture

```
src/
├── api/                 HTTP client layer — one file per backend resource
│                        (applications, connections, conversations, github, logs,
│                         models, notifications, organizations, pipeline, schedules,
│                         sources; shared client.ts)
├── assets/styles/       Global CSS (Tailwind v4 entry)
├── components/          Reusable UI components, grouped by domain
│   ├── common/          AppHeader, AppSidebar, OrgSidebar, StatusBadge,
│                        dropdowns, toasts, skeletons
│   ├── connections/     ConnectionBubble/Detail/Form/Test, AgentNebula (3D),
│                        SourceSelector, wizard/ (platform-first flow)
│   ├── agent/           ChatWindow, ChatMessage, ChatInput, ModelPicker
│   ├── log/             LogFeed, LogEntry, LogFilters, LogDetailModal
│   ├── pipeline/        PipelineFunnel (Sankey), LogTicker, JourneyModal,
│                        NodeCard/Detail, Particles, TimeMachineBlock
│   ├── schedules/       ScheduleCard, ScheduleModal
│   ├── notifications/   Channels, Preferences, History
│   ├── org/             AppCard, AppListRow, CreateOrgModal
│   ├── app-wizard/      New-app wizard + steps
│   ├── settings/        DeleteAppModal
│   ├── public/          HeroMesh, PublicNav, PublicFooter
│   └── icons/
├── composables/         useWebSocket, useAgent, usePipelineStream, useToast
├── layouts/             DefaultLayout, PublicLayout
├── pages/               Route-level views (lazy-loaded)
├── router/              Vue Router config (auth + onboarding guards)
├── stores/              Pinia stores (composition API)
├── test/                Vitest setup (fresh Pinia per test, Axios mocked)
├── types/               TypeScript type definitions
├── utils/               Formatters, constants
├── lib/                 Third-party client wrappers (supabase)
├── App.vue
└── main.ts
```

### Routing

| Path | Page | Description |
|------|------|-------------|
| `/` | LandingPage | Public marketing landing |
| `/features` | FeaturesPage | Public features page |
| `/pricing` | PricingPage | Public pricing page |
| `/login` | LoginPage | Supabase auth |
| `/onboarding` | OnboardingPage | First-run org + app creation |
| `/dashboard` | DashboardPage | Per-app overview and health |
| `/connections` | ConnectionsPage | Unified three-lane connection manager |
| `/pipeline` | PipelinePage | Live Sankey funnel + ticker + Time Machine |
| `/pipeline/logs/:logId` | PipelinePage | Deep-link to a log's journey replay |
| `/agent/config` | AgentConfigPage | Model, mode, schedule, prompt override |
| `/agent/chat` | AgentChatPage | Real-time chat with the agent |
| `/schedules` | SchedulesPage | Cron-scheduled investigation prompts |
| `/activity` | ActivityPage | Chronological master feed (was `/agent/log`) |
| `/notifications` | NotificationsPage | Channels, preferences, history |
| `/org` | OrgOverviewPage | Org-scoped landing |
| `/org/team` | OrgTeamPage | Member management |
| `/org/settings` | OrgSettingsPage | Org + apps management |
| `/org/billing` | OrgBillingPage | Billing |

### Stores

| Store | Responsibility |
|-------|----------------|
| `auth` | Supabase session, JWT, user profile |
| `app` | Org, apps list, currentAppId (persisted to localStorage) |
| `connections` | Connection CRUD + source filters |
| `logs` | Paginated log buffer + filters |
| `pipeline` | Rolling pipeline-event window (SSE-fed) |
| `schedules` | Scheduled investigations |

### Data Flow

```
Pages → Stores (Pinia) → API layer (Axios) → Backend REST API
Pages → Composables (useAgent)          → WebSocket /ws/chat?token=…
Pages → Composables (usePipelineStream) → SSE /api/apps/:id/pipeline/stream
```

Vite dev server proxies `/api` and `/ws` → `localhost:8080`. JWTs are injected
by an Axios interceptor; 401 responses trigger logout. WebSocket and SSE auth
use a `?token=` query param because EventSource/WS cannot set headers.

---

## Backend — `backend/`

Module: `github.com/hejijunhao/heimdall/backend`

### Architecture

```
cmd/
├── heimdall/main.go       Entry point: config → pool → classifier → GH client
│                          → agent → connectors (listener + poller) → HTTP server
├── dbping/                DB connectivity smoke-test binary
└── refresh-models/        Periodic OpenRouter model-catalogue refresh

internal/
├── config/                Env-based configuration + validation
├── api/                   HTTP layer
│   ├── router.go          Chi route definitions (see "API Routes" below)
│   ├── middleware/        Auth (Supabase JWKS), CORS, logging, MaxBodySize
│   └── handlers/          One file per resource — see "Handlers" below
├── agent/                 Agent engine
│   ├── agent.go           Struct + lifecycle (Start/Stop)
│   ├── loop.go            Core Claude tool-use loop (max 10 iterations)
│   ├── loop_stream.go     Streaming variant for chat
│   ├── monitor.go         Background polling loop (15s, max 10 concurrent apps)
│   ├── classifier.go      Classifier interface
│   ├── classifier_lumber.go  ONNX inference (confidence ≥ 0.5)
│   ├── severity_gate.go   Hardcoded escalation rules over classifier output
│   ├── scheduler.go       Cron-driven scheduled investigations
│   ├── pipeline_bus.go    In-process event bus (writer → SSE)
│   ├── pipeline_writer.go Per-stage event persistence (ingest/classify/gate/assess)
│   ├── pipeline_sweeper.go Drop-counter + stuck-stage sweeper
│   ├── pruner.go          Log-buffer retention pruner
│   ├── tools.go           Tool registry + dispatch
│   ├── tools_logs.go      search_logs implementation
│   ├── tools_db.go        query_database implementation
│   ├── tools_codebase.go  search_codebase (GitHub)
│   ├── provider.go        Provider interface
│   ├── provider_anthropic.go  Anthropic SDK adapter
│   ├── provider_openrouter.go OpenRouter HTTP adapter
│   ├── prompt.go          System prompt construction
│   ├── emit.go            agent_log emission (fire-and-forget)
│   ├── extract.go         Model → structured assessment parsing
│   ├── message.go         Provider-agnostic message types
│   └── models.go          Model catalogue
├── connectors/            External integration layer
│   ├── connector.go       Interfaces: Connector, StreamConnector,
│                          QueryConnector, PollConnector
│   ├── factory.go         Build a connector from a stored DB row
│   ├── registry.go        Runtime registry
│   ├── listener.go        Manager for long-lived push listeners (syslog, OTLP)
│   ├── poller.go          Timer-driven pull loop for PollConnectors
│   ├── database/          postgres.go (read-only SQL, enforces RO txn mode)
│   ├── logs/              flyio, syslog, vercel, railway, supabase, mongodb
│   └── codebase/          github (search, read, tree)
├── db/                    sqlc-generated Go
│   ├── db.go              Queries wrapper + RLS session var helper
│   ├── models.go          Struct per table
│   ├── *.sql.go           Generated query methods
│   └── queries/           Hand-written SQL (sqlc input)
├── github/                GitHub App client (JWT → installation token)
├── notifications/         Slack, Discord, Email (SMTP), formatter, Notifier
├── metrics/               Prometheus registry + counters/gauges
├── ws/                    WebSocket hub (client, hub, message envelope)
├── memory/                (empty — Elephantasm integration not implemented)
└── reports/               (empty — report generation folded into agent_log)

migrations/                golang-migrate SQL (001–038, up + down pairs)
```

### Agent — Two Operating Modes

| Mode | Trigger | Loop | File |
|------|---------|------|------|
| Interactive | `/ws/chat` connection | Claude tool-use loop, max 10 iterations; conversation persisted as JSONB | `agent/loop.go`, `loop_stream.go` |
| Monitoring  | 15s background tick per app | Fetch new logs → Lumber classify → gate → escalate flagged to Claude → emit agent_log | `agent/monitor.go` |
| Scheduled   | Cron-driven per schedule row | Run stored prompt, emit agent_log, optionally escalate to Activity | `agent/scheduler.go` |

### Agent Tools

| Tool | Description | File |
|------|-------------|------|
| `search_logs` | Search recent logs by query + severity (user-scoped) | `tools_logs.go` |
| `query_database` | Read-only SQL against a user's connected Postgres | `tools_db.go` |
| `search_codebase` | GitHub: `search_code` / `read_file` / `list_tree` | `tools_codebase.go` |

Tool errors are returned as `isError: true` results to Claude — never thrown.

### Classifier pipeline

`CLASSIFIER_MODE` env var controls behaviour: `on` (require Lumber, exit on
failure), `off` (passthrough — escalate everything), `fallback` (try Lumber,
degrade to passthrough on init error). Threshold and severity mapping live
in `severity_gate.go`.

### Pipeline page data plane

Four pipeline stages — `ingest`, `classify`, `gate`, `assess` — each emit at
most one `log_pipeline_events` row per `log_id`. Rows cascade-delete with
their parent `log_buffer` row (48h retention ceiling). An in-process bus
fan-outs events to SSE subscribers; Prometheus counters cover bus drops and
per-user SSE connection caps.

### Handlers (`internal/api/handlers/`)

One file per resource, each carrying its own test file:

```
applications, auth, chat, connections (+ test/validate helpers),
conversations, github_install, health, helpers, investigation_schedules,
logs, models, notifications, organizations, org_members, otlp,
pipeline (+ support), source_filters (+ discover, GitHub, pipeline),
source_name_path, userqueries, webhook_parsers, webhooks
```

`userqueries.go` is load-bearing: `UserQueries()` begins a transaction and
sets `SET LOCAL app.current_user_id` so PostgreSQL RLS policies scope every
read/write to the caller. `authorizeApp` validates an `appId` route param
belongs to the caller's org via `GetApplicationByOrgUser`.

### API Routes

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/api/webhooks/logs` | Bearer token | Default-format webhook ingestion |
| POST | `/api/webhooks/logs/{format}` | Bearer token | Format-aware webhook ingestion |
| POST | `/api/v1/logs` | Bearer token | OTLP HTTP log receiver |
| GET  | `/api/github/callback` | State JWT | GitHub App install redirect |
| GET  | `/api/apps/{appId}/pipeline/stream` | `?token=` | SSE pipeline event stream |
| GET  | `/health` | public | Liveness |
| GET  | `/ws/chat` | `?token=` | Agent chat WebSocket |
| GET  | `/metrics` | JWT | Prometheus exposition |
| GET / PUT / DELETE | `/api/org` | JWT | User's active organization |
| POST | `/api/onboard` | JWT | Idempotent org + first-app creation |
| GET / POST | `/api/orgs` | JWT | Multi-org list + create |
| CRUD | `/api/org/members` | JWT | Invite / role / remove |
| GET / POST | `/api/apps` | JWT | List / create applications |
| GET / DELETE | `/api/apps/{appId}` | JWT | Single-app get / delete |
| GET | `/api/apps/{appId}/connections` | JWT | Connections for an app |
| GET / PUT | `/api/apps/{appId}/agent/config` | JWT | Per-app agent config |
| GET | `/api/apps/{appId}/monitoring/status` | JWT | Monitoring cursor + health |
| GET | `/api/apps/{appId}/stats` | JWT | Dashboard stats |
| GET | `/api/apps/{appId}/pipeline/bootstrap` | JWT | Initial funnel snapshot |
| GET | `/api/apps/{appId}/pipeline/logs` | JWT | Time Machine picker |
| GET | `/api/apps/{appId}/pipeline/logs/{logId}/journey` | JWT | Single-log journey |
| CRUD | `/api/apps/{appId}/notifications/*` | JWT | Channels, prefs, history |
| CRUD | `/api/apps/{appId}/schedules` | JWT | Scheduled investigations |
| POST | `/api/apps/{appId}/schedules/{id}/run` | JWT | Run schedule now |
| GET | `/api/github/install` | JWT | Start GitHub App install flow |
| CRUD | `/api/connections` | JWT | User/org-scoped connection management |
| POST | `/api/connections/{id}/test` | JWT | Connection probe |
| CRUD | `/api/connections/{id}/sources` | JWT | Source filters |
| POST | `/api/connections/{id}/sources/discover` | JWT | Upstream source discovery |
| GET | `/api/logs` | JWT | Paginated log feed |
| GET | `/api/conversations[/{id}]` | JWT | Conversation history |
| GET | `/api/auth/me` | JWT | Current user profile |
| GET | `/api/models` | JWT | Model picker catalogue |

### Connector Implementations

| Domain | Connector | Interface(s) | Notes |
|--------|-----------|--------------|-------|
| Database | Postgres (`database/postgres.go`) | `QueryConnector` | Enforces `default_transaction_read_only=on` |
| Database | Supabase | `QueryConnector` | Postgres under the hood, Supabase-specific config |
| Logs | Webhook (HTTP endpoint, not connector) | — | `/api/webhooks/logs[/{format}]` with bearer token |
| Logs | OTLP HTTP (endpoint) | — | `/api/v1/logs` |
| Logs | Syslog TLS (`logs/syslog.go`) | `StreamConnector` via ListenerManager | Per-org TLS listener |
| Logs | Fly.io log drain (`logs/flyio.go`) | `PollConnector` | HTTP drain handler |
| Logs | Vercel (`logs/vercel.go`) | `PollConnector` | Vercel log drain API |
| Logs | Railway (`logs/railway.go`) | `PollConnector` | Railway API poller |
| Logs | Supabase logs (`logs/supabase.go`) | `PollConnector` | Supabase pg_logs poller |
| Logs | MongoDB (`logs/mongodb.go`) | `PollConnector` | MongoDB change-stream / oplog |
| Codebase | GitHub (`codebase/github.go`) | `QueryConnector` | App-auth install token, `search_code`/`read_file`/`list_tree` |

Listener vs. poller split: long-lived push sources (syslog, OTLP) live under
`ListenerManager`; pull-based sources run inside `Poller` on a shared timer
loop.

---

## Database

38 migrations to date. Query layer: sqlc — hand-written SQL in
`backend/internal/db/queries/`, generated Go under `backend/internal/db/`
(`sqlc.yaml` maps `uuid → google/uuid.UUID`, `jsonb → json.RawMessage`,
`timestamptz → time.Time`).

### Tables

| Table | Purpose | Notable columns |
|-------|---------|-----------------|
| `users` | Mirrors Supabase auth.users + org pointer | `org_id` |
| `organizations` | Top-level tenant | `name`, `owner_user_id` |
| `org_members` | Many-to-many user↔org with role | `role` (owner/admin/member) |
| `applications` | Per-org monitoring target | `org_id` (NOT NULL), `status` |
| `connections` | Integration config | `org_id`, `app_id` (nullable — org-scoped if null), `type`, `direction`, `config` JSONB |
| `connection_sources` | Observed source names per connection (auto-discovered) | `source_name`, `first_seen_at`, `last_seen_at` |
| `app_source_filters` | Per-app enable/disable over connection sources | `(app_id, connection_id, source_name)`, `enabled` |
| `app_agent_config` | Per-app agent tuning | `model`, `mode`, `provider`, `schedule_interval_secs`, `system_prompt_override` |
| `agent_config` | Legacy singleton config (retained for migration) | — |
| `monitoring_state` | Per-app cursor | `last_monitored_at` |
| `log_buffer` | Rolling log store (48h retention) | `app_id`, `severity`, `source_type`, `payload` JSONB |
| `log_pipeline_events` | One row per (log, stage) for Pipeline page | `log_id` FK, `stage`, `occurred_at`, `escalated`, summary fields |
| `investigations` | Agent findings | `trigger_type`, `severity`, `status`, `context/findings/tool_trace` JSONB |
| `investigation_schedules` | Cron-driven investigation prompts | `cron`, `prompt`, `last_run_at` |
| `conversations` | Chat sessions (interactive mode) | `messages` JSONB array |
| `agent_log` | Unified activity feed entries | `entry_type`, `summary`, `detail` JSONB, `severity` |
| `notification_channels` | Slack/Discord/email destinations | `kind`, `config` JSONB |
| `notification_preferences` | Per-app routing rules | trigger matrix |
| `notification_log` | Delivery audit trail | `status`, `agent_log_id` |
| `webhook_idempotency` | Dedupe webhook ingestion | `(connection_id, idempotency_key)` |

### Row-Level Security

RLS is enabled on every user-data and system table (migrations 013, 021, 027,
028, 030, 033, 045). Enforcement works by:

1. Handlers call `UserQueries(ctx, userID)` which begins a transaction and
   `SET LOCAL app.current_user_id = '<uuid>'`.
2. Each table's `FOR ALL USING / WITH CHECK` policy reads
   `current_setting('app.current_user_id')` and joins through
   `users → organizations → applications → connections` as needed.
3. Integration tests assert RLS holds — a regression test codifies the
   posture so future migrations can't silently weaken it.

---

## Key Commands

```bash
make dev                Run frontend + backend concurrently
make dev-frontend       Vite dev server on :5173
make dev-backend        Loads .env, runs backend on :8080
make build-frontend     vue-tsc + vite build
make build-backend      go build → backend/bin/heimdall
make test               Backend + frontend tests
make lint               go vet + ESLint
make migrate-up         Apply pending migrations (needs DATABASE_URL)
make migrate-down       Rollback one migration
make migrate-create     Interactive: create new migration pair
make sqlc-generate      Regenerate db/*.sql.go from queries/*.sql
docker compose up -d postgres   Local Postgres
make docker-up / down           All Docker services
```

Backend integration tests skip automatically if `DATABASE_URL` is unset.
Migration commands do not auto-source `.env` — prefix with
`set -a && . ./.env && set +a &&`.

---

## Data Model Hierarchy

```
User (Supabase auth.users)
  └─ Organization (via org_members, many-to-many w/ role)
       └─ Application (1:N, org-scoped)
            ├─ Connection (1:N) — can also be org-scoped (app_id NULL)
            │    ├─ ConnectionSource (1:N, auto-discovered)
            │    └─ AppSourceFilter (per app × connection × source)
            ├─ AppAgentConfig (1:1)
            ├─ MonitoringState (1:1) — cursor tracking
            ├─ InvestigationSchedule (1:N)
            └─ NotificationPreference / NotificationChannel
```

---

## What's Implemented vs Planned

| Component | Status |
|-----------|--------|
| Core agent loop (interactive + monitoring + scheduled) | Done |
| Lumber ONNX classifier + severity gate | Done |
| Multi-provider LLM (Anthropic + OpenRouter) | Done |
| Multi-org / multi-app / team roles | Done |
| Row-level security across all tables | Done |
| Connection wizard + platform-first flow | Done |
| Log connectors: webhook, OTLP, syslog TLS, Fly.io, Vercel, Railway, Supabase, MongoDB | Done |
| Database connectors: Postgres, Supabase | Done |
| Codebase connector: GitHub App | Done |
| Source filtering (discovery + per-app enable/disable) | Done |
| Pipeline page: funnel, ticker, journey replay, Time Machine | Done |
| Activity feed (unified `agent_log`) | Done |
| Scheduled investigations | Done |
| Notifications: Slack / Discord / email + preferences + history | Done |
| Observability: Prometheus `/metrics`, SSE drop counters | Done |
| Public marketing site (landing / features / pricing) | Done |
| Elephantasm long-term memory | Not started (`internal/memory/` empty) |
| Dedicated reports engine | Deferred — folded into `agent_log` |
| Pipeline coverage for poller-based connectors (Phase 1b) | Outstanding follow-up |
