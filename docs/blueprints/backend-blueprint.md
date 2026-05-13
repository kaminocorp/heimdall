# Backend Blueprint

A top-down walkthrough of the Heimdall Go backend — how it boots, what lives where, and how the layers connect.

This is the "map of the territory" doc. For a narrower deep-dive on any single subsystem, see:

- **`agent-architecture.md`** — the three agent modes, provider abstraction, tool registry, `agent_log` emission.
- **`lumber-integration.md`** — the ONNX classifier pipeline that gates monitoring-mode LLM calls.
- **`connections.md`** — connection schema, connector types, ingestion paths.
- **`database-connection-blueprint.md`** — pgxpool + RLS via `SET LOCAL`.
- **`sdks-blueprint.md`** — JS/Python/Go client SDKs that post to `/api/webhooks/logs`.

---

## 1. Boot Sequence

`backend/cmd/heimdall/main.go` — the entrypoint. Everything starts here, in roughly this order:

```
config.Load()                    ← read env vars into Config struct
cfg.Validate()                   ← fail fast if DATABASE_URL / ANTHROPIC_API_KEY / SUPABASE_URL missing

slog JSON handler (optional)     ← enabled by LOG_FORMAT=json

jwks = NewJWKSClient(supabaseURL)
jwks.Fetch(ctx)                  ← pre-load ECDSA signing keys (fail if unreachable)

pool = pgxpool.New(ctx, url)     ← establish Postgres connection pool

classifier = ...                 ← branch on CLASSIFIER_MODE:
                                     on       → LumberClassifier, fail boot if it can't load
                                     off      → PassthroughClassifier
                                     fallback → try Lumber, fall back to passthrough on error

ghClient = github.NewClient(...) ← only if GITHUB_APP_ID is set; else nil

queries  = db.New(pool)
notifier = notifications.NewDispatcher(queries, cfg)
ag       = agent.New(queries, cfg, classifier, notifier, ghClient)
ag.Start(ctx)                    ← spawns Monitor, Prune, InvestigationScheduler goroutines

poller   = connectors.NewPoller(queries)
resumePollers(queries, poller)   ← rehydrate all active flyio/vercel/railway/mongodb/supabase pollers

listener = connectors.NewListenerManager()
resumeSyslogListeners(...)       ← rehydrate active syslog listeners (TLS cert/key injected from env)

router   = api.NewRouter(cfg, pool, ag, jwks, ghClient, poller, listener)

http.Server.ListenAndServe()     ← read 30s, write 5m, idle 2m

signal.Notify(SIGINT|SIGTERM)    ← on signal: srv.Shutdown(10s),
                                     then poller.StopAll / listener.StopAll / ag.Stop (10s timeout),
                                     then classifier.Close() to release ONNX memory
```

The dependency chain flows in one direction:

```
Config → JWKSClient → Pool → Classifier → GitHub → Queries → Notifier → Agent → Poller/Listener → Router → Server
```

Boot is deliberately fail-fast on things we can't run without (DB, Supabase JWKS, Anthropic key) and best-effort on things we can (GitHub App, OpenRouter key, Lumber model in `fallback` mode).

---

## 2. Project Layout

```
backend/
├── cmd/
│   └── heimdall/
│       └── main.go                     # Entrypoint, boot wiring, resume* helpers
│
├── internal/
│   ├── config/
│   │   └── config.go                   # Config struct, env loading, Validate()
│   │
│   ├── api/
│   │   ├── router.go                   # Chi route table + middleware wiring
│   │   ├── middleware/
│   │   │   ├── auth.go                 # JWKSClient, ValidateJWT, Auth() middleware
│   │   │   ├── cors.go                 # CORS headers
│   │   │   └── logging.go              # Request logging via slog
│   │   └── handlers/
│   │       ├── server.go               # Server struct — central dependency holder
│   │       ├── userqueries.go          # UserQueries() — RLS-scoped transaction helper
│   │       ├── helpers.go              # jsonError + authorizeApp helpers
│   │       ├── auth.go                 # GET /api/auth/me
│   │       ├── organizations.go        # GET /api/org, POST /api/onboard
│   │       ├── applications.go         # CRUD for /api/apps (+ per-app sub-resources)
│   │       ├── connections.go          # CRUD + test for /api/connections
│   │       ├── webhooks.go             # POST /api/webhooks/logs (bearer-token ingestion)
│   │       ├── webhook_parsers.go      # Content-type-aware payload parsing
│   │       ├── otlp.go                 # POST /api/v1/logs (OTLP HTTP receiver)
│   │       ├── chat.go                 # WebSocket /ws/chat (streaming tool-use loop)
│   │       ├── conversations.go        # GET /api/conversations[/id]
│   │       ├── investigation_schedules.go  # CRUD + run-now for /api/apps/{id}/schedules
│   │       ├── notifications.go        # Channels + preferences + history per app
│   │       ├── github.go               # GitHub App install + callback
│   │       ├── models.go               # GET /api/models (curated dropdown)
│   │       ├── logs.go                 # GET /api/logs (unified feed: raw + agent)
│   │       ├── reports.go              # Stubs for /api/reports
│   │       └── health.go               # GET /health
│   │
│   ├── agent/                          # See agent-architecture.md for full walkthrough
│   │   ├── agent.go                    # Agent struct, New(), Start(), Stop(), providerFor()
│   │   ├── loop.go                     # runConversationCore, RunConversation, RunMonitoring
│   │   ├── loop_stream.go              # RunConversationStream + AgentEvent (streaming wrapper)
│   │   ├── monitor.go                  # 15s monitoring ticker goroutine
│   │   ├── scheduler.go                # 1m investigation scheduler goroutine + cron parser
│   │   ├── pruner.go                   # 1h log_buffer retention pruner goroutine
│   │   ├── provider.go                 # Neutral Provider interface + ChatParams types
│   │   ├── provider_anthropic.go       # AnthropicProvider (wraps anthropic-sdk-go)
│   │   ├── provider_openrouter.go      # OpenRouterProvider (raw net/http, opt-in)
│   │   ├── models.go                   # AnthropicModels + OpenRouterModels catalogues
│   │   ├── prompt.go                   # systemPrompt + monitoringSystemPrompt
│   │   ├── classifier.go               # Classifier interface, PassthroughClassifier
│   │   ├── classifier_lumber.go        # LumberClassifier (ONNX via kaminocorp/lumber)
│   │   ├── severity_gate.go            # ShouldEscalate hardcoded policy table
│   │   ├── extract.go                  # JSON payload → plain text for classification
│   │   ├── tools.go                    # ToolRegistry + Dispatch
│   │   ├── tools_logs.go               # search_logs implementation
│   │   ├── tools_db.go                 # query_database implementation
│   │   ├── tools_codebase.go           # search_codebase (GitHub App backed)
│   │   ├── tools_memory.go             # Placeholder for post-MVP Elephantasm memory tools
│   │   ├── emit.go                     # EmitLog / EmitLogWithSeverity (fire-and-forget)
│   │   └── message.go                  # Message domain type for conversation storage
│   │
│   ├── connectors/                     # See connections.md for full walkthrough
│   │   ├── connector.go                # Connector / StreamConnector / QueryConnector interfaces
│   │   ├── registry.go                 # Thread-safe runtime registry
│   │   ├── factory.go                  # Type-based connector construction
│   │   ├── poller.go                   # Goroutine pool for pull-based connectors
│   │   ├── listener.go                 # ListenerManager for syslog TCP/TLS
│   │   ├── database/
│   │   │   └── postgres.go             # Read-only Postgres connector (query_database)
│   │   ├── logs/
│   │   │   ├── syslog.go               # Syslog TCP/TLS listener
│   │   │   ├── flyio.go                # Fly.io log poller
│   │   │   ├── vercel.go               # Vercel log poller
│   │   │   ├── railway.go              # Railway log poller
│   │   │   ├── mongodb.go              # MongoDB Atlas log poller
│   │   │   └── supabase.go             # Supabase log poller
│   │   └── codebase/
│   │       └── github.go               # GitHub App code-search connector (search_codebase)
│   │
│   ├── github/                         # GitHub App client (JWT minting, installation tokens)
│   ├── notifications/                  # Dispatcher for email/Slack/Discord fan-out
│   ├── metrics/                        # Prometheus counters/histograms
│   │
│   └── db/
│       ├── db.go                       # DBTX interface, Queries struct, WithTx()
│       ├── models.go                   # Generated Go types for all tables
│       ├── queries/                    # Raw SQL query definitions (input to sqlc)
│       └── *.sql.go                    # sqlc-generated query methods
│
├── migrations/                         # PostgreSQL migrations (001–023)
├── sqlc.yaml                           # sqlc codegen config
├── Dockerfile                          # Three-stage build: compile → fetch Lumber model → runtime
├── go.mod
└── go.sum
```

---

## 3. Server Struct

All handlers hang off a single `Server` struct (`handlers/server.go`). This is the central dependency holder:

```go
type Server struct {
    Config   *config.Config         // env vars (API keys, URLs, feature flags)
    Pool     *pgxpool.Pool          // Postgres connection pool (for UserQueries transactions)
    Queries  *db.Queries            // sqlc-generated queries (default, no RLS context)
    Agent    *agent.Agent           // shared agent instance — three modes share one struct
    JWKS     *middleware.JWKSClient // JWT signing key cache
    GitHub   *github.Client         // GitHub App client — nil if GITHUB_APP_ID unset
    Poller   *connectors.Poller     // goroutine pool for pull-based log connectors
    Listener *connectors.ListenerManager // syslog TCP/TLS listeners
}
```

`Queries` is used for non-user-scoped operations (webhook ingestion looking up tokens, listing org-level resources during boot). User-scoped handlers call `s.UserQueries(ctx, userID)` instead, which starts a transaction with RLS context.

---

## 4. Router & Middleware

### Middleware Chain

Applied globally to every request via `r.Use()`:

```
Request → Logging → CORS → Handler
```

| Middleware | File | What it does |
|-----------|------|-------------|
| **Logging** | `middleware/logging.go` | Wraps `ResponseWriter` to capture status code. Logs method, path, status, duration via `slog`. Implements `Unwrap()` so the WebSocket library can access the underlying `net.Conn`. |
| **CORS** | `middleware/cors.go` | Permissive CORS (`*`) — suitable because all writes are bearer-token authenticated and the frontend is served cross-origin during local dev. |

### Auth Middleware

Applied to protected route groups via `r.Group(func(r chi.Router) { r.Use(middleware.Auth(jwks)) })`.

| Component | What it does |
|-----------|-------------|
| **`Auth(jwks)`** | Extracts `Bearer <token>` from the `Authorization` header → calls `ValidateJWT(token, jwks)` → injects `userID` (UUID) into request context → passes to handler. Returns 401 on failure. |
| **`ValidateJWT`** | Parses JWT, validates ES256 signature against cached Supabase JWKS keys, checks expiry, extracts `sub` claim as user UUID. Shared by both HTTP middleware and WebSocket auth in `chat.go`. |
| **`JWKSClient`** | Fetches and caches Supabase ECDSA public keys. Background refresh when stale but a key still exists. Force refresh on unknown `kid` (handles key rotation). |

### Route Table

```
router.go
├── Global middleware: Logging, CORS
│
├── /api
│   │
│   ├── [Public — no JWT]
│   │   ├── POST /webhooks/logs                 ← IngestWebhookLogs (bearer token → webhook_token lookup)
│   │   ├── POST /v1/logs                       ← IngestOTLPLogs (OTLP HTTP receiver)
│   │   └── GET  /github/callback               ← GitHubCallback (state-JWT, not session)
│   │
│   └── [Auth middleware group — Supabase JWT]
│       ├── GET  /org                            ← GetOrganization
│       ├── POST /onboard                        ← Onboard (create org + first app)
│       │
│       ├── GET  /apps                           ← ListApplications
│       ├── POST /apps                           ← CreateApplication
│       │
│       ├── /apps/{appId}
│       │   ├── GET    /                         ← GetApplication
│       │   ├── DELETE /                         ← DeleteApplication
│       │   ├── GET    /connections              ← ListConnectionsByApp
│       │   ├── GET    /agent/config             ← GetAppAgentConfig
│       │   ├── PUT    /agent/config             ← UpdateAppAgentConfig
│       │   ├── GET    /monitoring/status        ← GetMonitoringStatus
│       │   ├── GET    /stats                    ← GetAppDashboardStats
│       │   │
│       │   ├── /notifications
│       │   │   ├── GET    /preferences          ← GetNotificationPreferences
│       │   │   ├── PUT    /preferences          ← UpdateNotificationPreferences
│       │   │   ├── GET    /channels             ← ListNotificationChannels
│       │   │   ├── POST   /channels             ← CreateNotificationChannel
│       │   │   ├── PUT    /channels/{channelId} ← UpdateNotificationChannel
│       │   │   ├── DELETE /channels/{channelId} ← DeleteNotificationChannel
│       │   │   ├── POST   /channels/{channelId}/test ← TestNotificationChannel
│       │   │   └── GET    /history              ← ListNotificationHistory
│       │   │
│       │   └── /schedules                       ← Scheduled investigations
│       │       ├── GET    /                     ← ListSchedules
│       │       ├── POST   /                     ← CreateSchedule
│       │       ├── PATCH  /{id}                 ← UpdateSchedule
│       │       ├── DELETE /{id}                 ← DeleteSchedule
│       │       └── POST   /{id}/run             ← RunScheduleNow (synchronous)
│       │
│       ├── GET /github/install                  ← InstallGitHub (redirect to GitHub App install)
│       │
│       ├── /connections
│       │   ├── GET    /                         ← ListConnections
│       │   ├── POST   /                         ← CreateConnection
│       │   ├── GET    /{id}                     ← GetConnection
│       │   ├── PUT    /{id}                     ← UpdateConnection
│       │   ├── DELETE /{id}                     ← DeleteConnection
│       │   ├── POST   /{id}/test                ← TestConnection
│       │   ├── GET    /{id}/github/repos        ← ListGitHubRepos
│       │   └── PUT    /{id}/github/repos        ← UpdateGitHubRepos
│       │
│       ├── GET /logs                            ← ListLogs (unified: raw + agent)
│       │
│       ├── /conversations
│       │   ├── GET /                            ← ListConversations
│       │   └── GET /{id}                        ← GetConversation
│       │
│       ├── GET /auth/me                         ← Me
│       ├── GET /models                          ← GetAvailableModels (Anthropic + OpenRouter)
│       │
│       └── /reports
│           ├── GET /                            ← ListReports (still a stub)
│           └── GET /{id}                        ← GetReport (still a stub)
│
├── GET /health                                   ← Health (public)
├── GET /metrics                                  ← Prometheus scrape endpoint (public)
└── GET /ws/chat                                  ← HandleChat — WebSocket, JWT via ?token= query param
```

Two things to notice:

1. **`/ws/chat` sits outside the `/api` group and outside the Auth middleware.** It handles JWT validation manually from the `?token=` query parameter because the browser WebSocket API cannot set custom headers. `/health` and `/metrics` are also outside so they can be hit without auth for k8s probes and Prometheus scraping.
2. **Per-app nesting is the dominant route shape.** Most user-facing operations sit under `/apps/{appId}/...` — connections, agent config, monitoring status, schedules, notifications. The older global `/api/agent/config` and `/api/agent/run` routes that used to live in this blueprint no longer exist; they were replaced by per-app equivalents in v0.12.0 (Multi-App UI & API).

---

## 5. Handler Layer

### How Handlers Access the Database

Two patterns, depending on whether the data is user-scoped:

**User-scoped** (most handlers):
```go
queries, done, err := s.UserQueries(ctx, userID)
defer done()
// queries is now scoped to a transaction with SET LOCAL app.current_user_id = userID
```

**Global** (webhook ingestion, monitoring goroutines, org-level reads):
```go
s.Queries.GetConnectionByWebhookToken(ctx, token)
```

The `authorizeApp` helper in `helpers.go` is used by per-app handlers to verify that the `{appId}` URL param belongs to the authenticated user's org before proceeding. It calls `GetApplicationByOrgUser` and 403s on mismatch.

### Key handlers

| File | Purpose |
|---|---|
| `applications.go` | CRUD for apps + per-app resources (agent config, monitoring status, stats). The `CreateApplication` handler writes a default `app_agent_config` row pointing at `agent.DefaultModelID`. |
| `connections.go` | Connection CRUD + `TestConnection`. For `postgres`/`database` type: runs a live ping via `database.New`. For webhook/syslog/poller/github: starts/stops the appropriate background worker (poller goroutine or syslog listener). |
| `webhooks.go` | `POST /api/webhooks/logs`. Bearer-token auth via `connections.config->>'webhook_token'` (indexed by migration 008). Delegates payload parsing to `webhook_parsers.go`, which dispatches on `Content-Type` for different source shapes. |
| `otlp.go` | `POST /api/v1/logs`. OTLP HTTP receiver — accepts OTel log payloads and writes them to `log_buffer` without a dedicated `connections` row. |
| `chat.go` | The WebSocket chat handler. Validates the JWT from `?token=`, hydrates conversation history, and calls `Agent.RunConversationStream` — which emits `tool_start`/`tool_result`/`message` events as the agent loop iterates. Full details in `agent-architecture.md`. |
| `investigation_schedules.go` | CRUD for scheduled investigations. Validates either `interval_secs` OR `cron_expr` (exactly-one-of, 400 otherwise). `POST /{id}/run` calls `Agent.RunScheduledInvestigation` synchronously. |
| `notifications.go` | Notification channels (email/Slack/Discord) and per-app preferences. Wired to `notifications.Dispatcher` which is invoked fire-and-forget by the monitoring loop after each assessment. |
| `github.go` | GitHub App install/callback. The callback endpoint is the one hit by GitHub's redirect and uses a **state JWT** rather than a session cookie so it can complete without the user being "logged in" in the usual sense. |
| `models.go` | `GET /api/models` — merges `agent.AnthropicModels` + `agent.OpenRouterModels` depending on whether `OPENROUTER_API_KEY` is set. Populates the frontend's model dropdown honestly. |
| `logs.go` | Unified feed merging raw `log_buffer` rows and `agent_log` rows. Filters: `severity`, `connection_id`, `source=raw|agent|all`. Pagination. |
| `reports.go` | Still stubs. `ListReports` returns `[]`, `GetReport` returns 501. Reports are in the product vision but not yet wired up. |

---

## 6. Agent System

See **`agent-architecture.md`** for the full walkthrough — this is just the handler-layer entry points.

The `Agent` struct is constructed once in `main.go` and injected everywhere it's needed (the HTTP server holds it for interactive chat; `ag.Start()` spawns three background goroutines for monitoring, pruning, and scheduling). It has:

- A **provider map** (`anthropic` always, `openrouter` if keyed) accessed via `providerFor(name)`.
- A shared **classifier** (Lumber ONNX or passthrough, depending on `CLASSIFIER_MODE`).
- A shared **rate limiter** (30 rpm / burst 5) used by both the monitoring loop and the scheduler.
- A reference to the **notifications dispatcher** for fire-and-forget alert fan-out.
- A reference to the **GitHub App client** for the `search_codebase` tool.

**Three agent modes** (see `agent-architecture.md` for the full story):

1. **Interactive** — `ws/chat` → `RunConversationStream` → `runConversationCore` (`loop.go`). Streams tool progress to the browser, persists history to `conversations.messages`.
2. **Monitoring** — 15-second `Monitor` goroutine. Classifies new logs via Lumber, escalates flagged ones (capped at 50/batch) to `RunMonitoring` (`loop.go`). Strict `Assessment/Severity/Action` output format parsed for severity routing.
3. **Scheduled** — 1-minute `InvestigationScheduler` goroutine. Fires cron/interval schedules, reusing `RunMonitoring` as its LLM call.

**Tool registry** (`agent/tools.go`): `search_logs`, `query_database`, `search_codebase`. All three modes share the same registry. A fourth placeholder (`tools_memory.go`) is reserved for post-MVP Elephantasm memory tools.

**Default model:** `claude-sonnet-4-6` (`agent.DefaultModelID`). Model/provider selection flows through per-app config (`app_agent_config`) with fallback to the global `agent_config` singleton.

---

## 7. Connectors

### Interface Hierarchy (`connectors/connector.go`)

```
Connector                      Connect(), Health(), Close()
├── StreamConnector            + Stream(ctx, chan<- []byte)
└── QueryConnector             + Query(ctx, query string) (any, error)
```

### Runtime Infrastructure

- **`connectors.Poller`** — goroutine pool for pull-based log connectors (Fly.io, Vercel, Railway, MongoDB, Supabase). Each active connection has a goroutine that polls its provider on an interval and writes normalised entries to `log_buffer`. `resumePollers` rehydrates them at boot.
- **`connectors.ListenerManager`** — manages long-lived syslog TCP/TLS listeners. `resumeSyslogListeners` reattaches listeners for all `active` syslog connections at boot, injecting server-level TLS cert/key from env if the connection row doesn't carry its own.
- **`connectors.Registry`** — thread-safe `map[string]Connector` used by the HTTP handlers and the agent tools for runtime lookup.

### Per-type implementations

| Connector | Package | Status |
|---|---|---|
| Postgres (read-only) | `connectors/database/postgres.go` | Fully implemented. Used by `query_database`. Forces `default_transaction_read_only=on`. |
| Webhook logs | handled directly by `webhooks.go` handler | Fully implemented. No runtime goroutine — HTTP-driven. |
| Syslog TCP/TLS | `connectors/logs/syslog.go` | Fully implemented. Managed by `ListenerManager`. |
| OTLP HTTP | handled directly by `otlp.go` handler | Fully implemented. No runtime goroutine — HTTP-driven. |
| Fly.io | `connectors/logs/flyio.go` | Fully implemented poller. |
| Vercel | `connectors/logs/vercel.go` | Fully implemented poller. |
| Railway | `connectors/logs/railway.go` | Fully implemented poller. |
| MongoDB Atlas | `connectors/logs/mongodb.go` | Fully implemented poller. |
| Supabase | `connectors/logs/supabase.go` | Fully implemented poller. |
| GitHub codebase | `connectors/codebase/github.go` | Fully implemented. GitHub App backed — used by `search_codebase`. |

---

## 8. Database Layer

### sqlc Configuration (`sqlc.yaml`)

- Engine: PostgreSQL
- Queries: `internal/db/queries/*.sql` → generated Go in `internal/db/*.sql.go`
- Type overrides: `uuid` → `google/uuid.UUID`, `jsonb` → `json.RawMessage`, `timestamptz` → `time.Time`

### Generated Code Pattern

For each `.sql` file in `queries/`, sqlc generates a corresponding `.sql.go` file with:
- A `const` for the raw SQL string
- A `Params` struct for each query's parameters
- A method on `*Queries` that executes the query and returns typed results

The `Queries` struct wraps a `DBTX` interface (satisfied by both `*pgxpool.Pool` and `pgx.Tx`), so the same queries can run directly on the pool or within a transaction. `db.New(pool)` creates Queries on the pool; `queries.WithTx(tx)` creates Queries on a transaction (used by `UserQueries`).

### Data Model Hierarchy

```
User (Supabase auth.users mirror)
  └─ Organization (1:1 via users.org_id)
       └─ Application (1:N)
            ├─ Connection (1:N) — log sources, databases, codebases
            ├─ AppAgentConfig (1:1) — model, mode, schedule, prompt override, provider
            ├─ MonitoringState (1:1) — cursor tracking (last_monitored_at)
            ├─ InvestigationSchedule (1:N) — cron/interval scheduled agent runs
            ├─ NotificationPreference (1:1)
            └─ NotificationChannel (1:N)
```

### Key Tables (non-exhaustive)

| Table | Purpose |
|---|---|
| `users` | Mirror of Supabase `auth.users`. Has `org_id` for the user's organization. |
| `organizations` | Tenancy boundary. Apps hang off orgs, not users directly. |
| `applications` | A monitored app. Has `org_id`, `name`, default mode, `schedule_interval_secs`. |
| `connections` | Integrations between the app and external systems. See `connections.md`. |
| `log_buffer` | Raw ingested logs. JSONB payload, severity, cursor by `ingested_at`. Pruned every hour by `agent.Prune`. |
| `agent_log` | Agent observations, tool calls, tool results, monitoring assessments, scheduled investigation results. The "Activity feed" in the UI. |
| `conversations` | Interactive chat state. `messages` is JSONB; user-scoped. |
| `agent_config` | Global singleton row. Legacy — mostly superseded by per-app config. |
| `app_agent_config` | Per-app model/mode/provider/prompt. Loaded by monitoring + chat + scheduler. |
| `monitoring_state` | Per-app cursor (`last_monitored_at`) driving the `Monitor` loop. |
| `investigation_schedules` | Scheduled agent runs — `cron_expr` or `interval_secs`, plus stored prompt and last-run metadata. |
| `notification_channels` | Email/Slack/Discord targets per app. |
| `notification_preferences` | Per-app severity threshold + quiet hours. |
| `notification_log` | History of dispatched notifications. |
| `github_repos` | Per-repo enablement for GitHub connections — which repos the agent may search. |
| `investigations` | Investigation lifecycle tracking. Queries defined; no full handler wiring yet. |

---

## 9. Schema Evolution (Migrations)

Current migrations: **001 → 023**.

```
001  Create connections table
002  Create agent_config table (singleton)
003  Create investigations table
004  Create conversations table
005  Create log_buffer table
006  Create users table + auto-sync trigger from Supabase auth.users
007  Add user_id to connections (backfill + NOT NULL)
008  Add index on connections.config->>'webhook_token' for fast webhook lookup
009  Add user_id to conversations
010  Create agent_log table
011  Add user_id to investigations
012  Add user_id to log_buffer
013  Enable RLS on all user-scoped tables + create app_current_user_id() helper
014  Create organizations + applications tables (adds org_id to users, app_id to connections)
015  Create app_agent_config (per-app model/provider/prompt override)
016  Create monitoring_state (per-app cursor tracking)
017  Create notification_channels
018  Create notification_preferences
019  Create notification_log
020  Create github_repos (per-repo enablement for GitHub connections)
021  Enable RLS on missing tables (retroactive pass)
022  Add provider column to agent_config / app_agent_config (for OpenRouter support)
023  Create investigation_schedules (stored prompts + cron_expr/interval_secs)
```

The pattern across the history: tables were created first (001–005), then user-scoping was added retroactively (006–012, 014 for app-scoping), then RLS was enforced (013 + 021), then per-app config and operational tables were layered on (015–020, 023). Migration 022 was a targeted addition to support a new provider without touching any other schema.

---

## 10. Security Model

### Authentication

**HTTP routes**: Supabase-issued JWT in `Authorization: Bearer <token>` header. Validated via ES256 against JWKS-cached public keys. See `middleware/auth.go`.

**WebSocket (`/ws/chat`)**: JWT passed as `?token=` query parameter (browser WebSocket API limitation — can't set custom headers). Same `ValidateJWT()` function used for both paths.

**Webhook ingestion (`/api/webhooks/logs`)**: Bearer token looked up in `connections.config->>'webhook_token'` (indexed) — maps to a connection and its owning `user_id` + `app_id`.

**GitHub App callback (`/api/github/callback`)**: State JWT rather than a session — the callback is hit by GitHub's redirect and the user's browser may not be authenticated to the SPA at that moment.

### Row-Level Security (RLS)

Every user-scoped table has an RLS policy:

```sql
CREATE POLICY [table]_owner ON [table]
  FOR ALL
  USING (user_id = app_current_user_id());

-- Helper function:
CREATE FUNCTION app_current_user_id() RETURNS uuid AS $$
  SELECT nullif(current_setting('app.current_user_id', true), '')::uuid;
$$ LANGUAGE sql STABLE;
```

**How it's activated from Go**:

```go
// handlers/userqueries.go
tx, _ := s.Pool.Begin(ctx)
tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", userID.String())
return s.Queries.WithTx(tx)  // all subsequent queries run inside this transaction
```

`SET LOCAL` is transaction-scoped, so concurrent requests from different users never share session state. This is the canonical way to combine RLS with connection pooling. See `database-connection-blueprint.md` for the full pattern.

**Current enforcement model** (post-rollout, completed via the eight-phase RLS role split — see `docs/completions/rls-enforcement-phase-{1..8}.md` and the archived [`rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md)):

- **Two runtime roles**, neither superuser:
  - `app_user` (`DATABASE_URL`) — authenticates the JWT-scoped HTTP/WebSocket path. **No `BYPASSRLS`.** Every protected table is `FORCE ROW LEVEL SECURITY`'d, so RLS is the actual access boundary, not a defence-in-depth layer.
  - `cron_user` (`CRON_DATABASE_URL`) — authenticates the no-JWT background path (monitor loop, scheduler, connector pollers, log-buffer pruner). Has `BYPASSRLS` for cross-tenant enumeration plus a deliberately narrow set of writes (`INSERT/UPDATE` on `monitoring_state`, `DELETE` on `log_buffer`) — every write is a documented exception.
- **One migration role**, kept superuser-only: `postgres` (`DIRECT_URL`), used by `make migrate-*` and nothing else. Runtime traffic never touches it.
- **Bypass-then-scope handoff** for non-JWT paths: `cron_user` does the smallest possible cross-tenant lookup ("which user owns this app?"), then closes that session and reopens via `Pools.WithUserQueries(ctx, ownerUserID)` on the `app_user` pool with `app.current_user_id` set — so the audit trail and policy evaluation match a normal JWT request.
- **One SECURITY DEFINER escape hatch**: `lookup_user_for_invite(email)` (migration 041) lets the org-member invite handler resolve a not-yet-org-mate user by email. The function is owned by `postgres`, has `search_path` pinned, and is `EXECUTE`-granted only to `app_user`.
- **Startup invariant**: when `HEIMDALL_ENV=production`, `cmd/heimdall/main.go:79–84` refuses to launch if `DATABASE_URL` and `CRON_DATABASE_URL` authenticate as the same role — caught in deployment health checks before traffic is accepted.

The `SET LOCAL app.current_user_id` snippet above is what makes RLS load-bearing under FORCE: every request runs inside a transaction whose first statement pins the GUC the policies read.

### Read-Only Database Queries

The Postgres connector (`connectors/database/postgres.go`) appends `default_transaction_read_only=on` to the connection string. All agent queries against user databases are forced read-only at the Postgres level — writes will error.

---

## 11. Data Flows

### Log Ingestion (webhook path)

```
External App / SDK
  │
  │  POST /api/webhooks/logs
  │  Authorization: Bearer <webhook_token>
  │  Content-Type: application/json
  │  Body: {source_type, severity?, payload} or JSON array
  ▼
webhooks.go:IngestWebhookLogs
  │
  ├─ Extract Bearer token → GetConnectionByWebhookToken → connection_id + user_id + app_id
  ├─ Read body (10 MB cap)
  ├─ parseWebhookPayload(body, contentType)  ← webhook_parsers.go dispatches on Content-Type
  ├─ Insert each entry into log_buffer with connection_id + user_id
  └─ Return 201 Created
         │
         ▼
    log_buffer table
         │
    ┌────┼──────────────────────────────┐
    ▼    ▼                              ▼
GET    agent.Monitor                   search_logs tool
/logs  (15s tick, classifier pipeline)  (on-demand during investigation)
```

Log ingestion has four other entry points that all land in `log_buffer` the same way: OTLP (`/api/v1/logs`), syslog TCP/TLS listeners, the provider pollers (Fly.io/Vercel/Railway/MongoDB/Supabase), and the agent's own `agent_log` writes (a separate table, same semantic feed).

### WebSocket Chat (interactive agent mode)

```
Browser
  │
  │  GET /ws/chat?token=JWT&app_id=UUID&conversation_id=UUID
  ▼
chat.go:HandleChat
  │
  ├─ Validate JWT from ?token= → userID
  ├─ Accept WebSocket upgrade
  ├─ Load or create conversation (user-scoped via UserQueries)
  ├─ Send {type: "system", conversation_id}
  │
  └─ Message loop (runs until disconnect):
       │
       ├─ Read {content} from client
       ├─ Append to storedMessages, persist (JSONB) to conversations.messages
       ├─ Send {type: "status", content: "thinking"}
       │
       ├─ events := Agent.RunConversationStream(userID, appID, &convID, history, input)
       │    │
       │    └─ runConversationCore (goroutine)
       │         ├─ loop (up to 10 iterations)
       │         │   ├─ provider.ChatCompletion(ctx, params)
       │         │   ├─ on tool_use: Dispatch(search_logs / query_database / search_codebase)
       │         │   │     ├─ emit AgentEvent{tool_start}
       │         │   │     ├─ execute tool (errors returned as IsError tool results)
       │         │   │     ├─ emit AgentEvent{tool_result}
       │         │   │     └─ EmitLog(tool_call / tool_result) → agent_log
       │         │   └─ on end_turn: EmitLog(observation), emit AgentEvent{message}
       │         └─ close(events)
       │
       ├─ For each event in channel:
       │     tool_start / tool_result → forward to WebSocket
       │     message                  → capture as fullResponse
       │     error                    → capture as streamErr
       │
       ├─ Append agent message to storedMessages, persist
       └─ Send {role: "agent", content: fullResponse, ...} to client
```

See `agent-architecture.md` §2.1 and §3 for the full tool-use loop diagram and the shared `runConversationCore` internals.

### Agent Monitoring (background mode)

```
agent.Start() → Monitor goroutine (15s tick)
   │
   └─ monitorTick(ctx, sem):
        │
        ├─ ListActiveApplications
        └─ For each app (semaphore cap 10, per-app 2min timeout):
              │
              ├─ shouldMonitor? (continuous → always; periodic → check interval)
              ├─ Load monitoring_state cursor (first run: init to now, skip)
              ├─ ListLogsSinceForApp (cap 200)
              ├─ classifier.Classify(logs) → (flagged, safeCount)
              │
              ├─ if len(flagged) == 0: advance cursor, done
              │
              ├─ Load app_agent_config (model, provider, prompt override)
              ├─ limiter.Wait(ctx)  ← 30 rpm / burst 5 shared with scheduler
              ├─ RunMonitoring(userID, appConfig, formattedFlaggedLogs)
              │    └─ Claude tool-use loop → (assessment, severity)
              │
              ├─ EmitLogWithSeverity(entry_type="monitoring", ...)
              ├─ notifier.Notify (fire-and-forget) — dispatches to email/Slack/Discord
              └─ Advance cursor to last processed log's ingested_at
```

See `agent-architecture.md` §2.2 and `lumber-integration.md` for the classifier pipeline that does most of the heavy lifting before any LLM is involved.

---

## 12. Configuration

### Environment Variables

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `PORT` | No | `8080` | HTTP server port |
| `LOG_FORMAT` | No | `text` | `text` or `json` — controls slog handler |
| `DATABASE_URL` | **Yes** | — | PostgreSQL connection string (Supabase) |
| `ANTHROPIC_API_KEY` | **Yes** | — | Claude API key (default provider) |
| `SUPABASE_URL` | **Yes** | — | Supabase project URL (for JWKS endpoint) |
| `OPENROUTER_API_KEY` | No | — | Optional — registers the OpenRouter provider and unlocks additional models in the dropdown |
| `CLASSIFIER_MODE` | No | `fallback` | `on` / `off` / `fallback` — see `lumber-integration.md` |
| `LUMBER_MODEL_DIR` | No | `/opt/lumber/models` | Directory containing the ONNX model files |
| `RESEND_API_KEY` | No | — | Email notification dispatch via Resend |
| `NOTIFICATION_FROM_EMAIL` | No | — | `From:` address for notification emails |
| `GITHUB_APP_ID` | No | — | Enables the GitHub App client (and `search_codebase` tool) |
| `GITHUB_PRIVATE_KEY` | No | — | PEM-encoded GitHub App private key |
| `GITHUB_CLIENT_ID` | No | — | GitHub App OAuth client ID |
| `GITHUB_APP_SLUG` | No | `heimdall-agent` | Slug used in install URLs |
| `GITHUB_WEBHOOK_SECRET` | No | — | Webhook signature validation |
| `FRONTEND_URL` | No | `http://localhost:5173` | Used for GitHub callback redirect + CORS origins |
| `SYSLOG_TLS_CERT` | No | — | PEM cert injected into syslog listeners that don't carry their own |
| `SYSLOG_TLS_KEY` | No | — | PEM key, paired with `SYSLOG_TLS_CERT` |
| `ELEPHANTASM_URL` | No | — | Placeholder — long-term memory integration (not yet wired) |
| `ELEPHANTASM_API_KEY` | No | — | Placeholder |

### Config Struct (`config/config.go`)

```go
type Config struct {
    Port                  string
    LogFormat             string
    DatabaseURL           string
    AnthropicKey          string
    OpenRouterKey         string
    ElephantasmURL        string
    ElephantasmKey        string
    SupabaseURL           string
    ClassifierMode        string
    LumberModelDir        string
    ResendAPIKey          string
    NotificationFromEmail string
    GitHubAppID           string
    GitHubPrivateKey      string
    GitHubClientID        string
    GitHubAppSlug         string
    GitHubWebhookSecret   string
    FrontendURL           string
    SyslogTLSCert         string
    SyslogTLSKey          string
}
```

`Load()` reads from env vars with `os.Getenv`. `Validate()` only requires the three load-bearing fields: `DATABASE_URL`, `ANTHROPIC_API_KEY`, `SUPABASE_URL`. Everything else is optional — the backend degrades gracefully (no OpenRouter in the dropdown, no GitHub tools, no email notifications, etc.).

---

## 13. Still Stubbed / Not Yet Wired

A much shorter list than it used to be. As of this writing:

| Component | Location | Status |
|-----------|----------|--------|
| **Reports** | `handlers/reports.go` | Still stubs. `ListReports` returns `[]`, `GetReport` returns 501. The product vision has incident reports as a first-class concept but the generation pipeline isn't built yet. |
| **Investigations** | `db/queries/investigations.sql` | sqlc queries exist (user-scoped, with severity/status/findings) but no HTTP handlers, no agent wiring. The `investigation_schedules` table (migration 023) is a separate concept — it's the scheduler's stored prompts, not the investigation lifecycle. |
| **Long-term memory** | `agent/tools_memory.go` | Empty file. `recall_similar_incidents` / `recall_lessons` are deferred until the Elephantasm integration lands. `ELEPHANTASM_URL` / `ELEPHANTASM_API_KEY` env vars exist as placeholders. |
| **Token-by-token streaming** | `agent/loop_stream.go` | Today's streaming is coarse-grained (tool transitions + final message). Per-token streaming inside a `ChatCompletion` call is flagged as a follow-up. |

Items that **used** to be in this list but have shipped since the last version of this blueprint: the monitoring loop, the investigation scheduler, the log retention pruner, the GitHub connector + `search_codebase` tool, all five pull-based log pollers, the syslog TCP/TLS listener, the OTLP HTTP receiver, the OpenRouter provider, and the notifications subsystem.
