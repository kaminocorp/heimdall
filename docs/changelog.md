# Changelog

- [0.7.1 — Supabase Auth Wiring](#071--supabase-auth-wiring-2026-02-26)
- [0.7.0 — Agent Log](#070--agent-log-2026-02-24)
- [0.6.1 — Post-Implementation Fixes](#061--post-implementation-fixes-2026-02-24)
- [0.6.0 — Agent Chat](#060--agent-chat-2026-02-24)
- [0.5.0 — Agent Loop](#050--agent-loop-2026-02-23)
- [0.4.0 — Webhook Log Connector](#040--webhook-log-connector-2026-02-23)
- [0.3.0 — Connections CRUD](#030--connections-crud-2026-02-23)
- [0.2.4 — Auth Me Endpoint & DB Pool](#024--auth-me-endpoint--db-pool-2026-02-22)
- [0.2.3 — Auth-Protected Routes](#023--auth-protected-routes-2026-02-22)
- [0.2.2 — JWT Verification Middleware](#022--jwt-verification-middleware-2026-02-22)
- [0.2.1 — Users Table & Auth Groundwork](#021--users-table--auth-groundwork-2026-02-22)
- [0.2.0 — Supabase Database](#020--supabase-database-2026-02-22)
- [0.1.3 — Infrastructure & DevOps](#013--infrastructure--devops-2026-02-20)
- [0.1.2 — Frontend Fixes](#012--frontend-fixes-2026-02-20)
- [0.1.1 — Backend Fixes & Hardening](#011--backend-fixes--hardening-2026-02-20)
- [0.1.0 — Scaffolding](#010--scaffolding-2026-02-19)

---

## 0.7.1 — Supabase Auth Wiring (2026-02-26)

Replaced the placeholder login flow with real Supabase email/password authentication. The frontend now uses `@supabase/supabase-js` for sign-in, sign-up, session recovery, and automatic token refresh — no backend changes required.

### Auth Store (full rewrite)

- `init()` recovers session on page load via `supabase.auth.getSession()` and subscribes to `onAuthStateChange` for transparent token refresh. ([phase7])
- `login(email, password)` and `signup(email, password)` call Supabase auth methods directly. ([phase7])
- `token` is a computed from `session.access_token` — auto-updates on refresh, consumed by the existing axios interceptor and WebSocket `?token=` param with zero changes to either. ([phase7])

### Login Page (full rewrite)

- Real email/password form with loading state, error banner, and sign-in / sign-up toggle. ([phase7])
- Signup with email confirmation shows "Check your email" message. ([phase7])

### App & Router

- `App.vue` gates rendering behind `auth.initialized` — prevents flash of login page on reload while session recovery is in progress. ([phase7])
- Router guard skips redirect until auth is initialized; redirects authenticated users away from `/login` → `/`. ([phase7])

### Sidebar

- Displays authenticated user's email at the bottom of the nav. ([phase7])
- "Sign out" calls `supabase.auth.signOut()` and redirects to `/login`. ([phase7])

### New Files

- `frontend/src/lib/supabase.ts` — singleton Supabase client reading `VITE_SUPABASE_URL` and `VITE_SUPABASE_ANON_KEY` from env. ([phase7])
- `.env.example` updated with frontend Supabase env vars. ([phase7])

[phase7]: completions/phase7-supabase-auth-wiring.md

---

## 0.7.0 — Agent Log (2026-02-24)

Phase 6 — unified chronological feed combining raw ingested log entries with agent observations. The agent now emits structured log entries during its tool-use loop, and the log endpoint merges both data sources into a single timeline.

### Database

- Migration `010_create_agent_log`: new `agent_log` table for agent-emitted observations (`tool_call`, `tool_result`, `observation`). Indexes on `(user_id, created_at DESC)` and `(entry_type)`. ([phase6])
- `conversation_id` uses `ON DELETE SET NULL` — agent log entries survive conversation deletion, preserving the audit trail. ([phase6])
- sqlc queries: `InsertAgentLog`, `ListAgentLogByUser`, `ListAgentLogByUserAndType`, `CountAgentLogByUser`. ([phase6])

### Agent

- **`EmitLog` method** — fire-and-forget agent log emission. Prevents observability writes from degrading the agent's primary work. ([phase6])
- **Emit hooks in tool-use loop** — three emit points: `tool_call` on dispatch, `tool_result` on success/failure, `observation` on final response. ([phase6])
- `RunConversation` signature updated to accept `conversationID *uuid.UUID` — agent log entries emitted during chat link to the conversation. ([phase6])

### Backend

- **Unified `GET /api/logs`** — merges `log_buffer` and `agent_log` into a single feed. New `source` query param (`all`, `raw`, `agent`). When `connection_id` filter is set, `source` is forced to `raw`. ([phase6])
- Unified response shape with `source`, `source_type`, `summary`, and `detail` fields normalised across both entry types. ([phase6])

### Frontend

- **Source filter** — new dropdown in LogFilters (`All sources` / `Raw logs` / `Agent activity`). ([phase6])
- **Agent entry styling** — purple border and `AGENT` badge for agent-sourced entries. Human-readable labels: `tool_call` → "Tool Call", `tool_result` → "Tool Result", `observation` → "Observation". ([phase6])
- Updated `LogEntry` type, API client, Pinia store, and page wiring for `source` filtering. ([phase6])

[phase6]: completions/phase6-agent-log.md

---

## 0.6.1 — Post-Implementation Fixes (2026-02-24)

Review of Phases 1–5 identified error-handling gaps and repo hygiene issues. All fixes target resilience and cleanliness — no functional changes.

### Backend

- **`rand.Read` error handling** — `crypto/rand.Read` failure in webhook token generation now returns 500 instead of silently producing a zero-value token. ([phase5-fixes])
- **JSON marshal/unmarshal error handling** — malformed config JSON in `CreateConnection` now returns 400/500 instead of being silently swallowed. ([phase5-fixes])
- **WebSocket write error handling** — all five `wsjson.Write` calls in `HandleChat` now check return values; write failures terminate the handler cleanly instead of continuing to run the agent loop for a dead connection. ([phase5-fixes])

### Frontend

- **Missing `onerror` handler** — `useWebSocket` now handles WebSocket errors, transitioning status to `'closed'` instead of showing a stale "connecting" state. ([phase5-fixes])
- **Type safety bypass removed** — widened `ChatMessage.role` to include `'assistant'` and removed `as any` cast in `useAgent.ts`. ([phase5-fixes])

### Repo Hygiene

- **Removed `backend/heimdall` binary from git** — 19MB compiled binary was tracked since Phase 3, creating large deltas on every build. Added to `.gitignore` and removed from tracking. ([phase5-fixes])
- **Deleted duplicate `frontend/vite.config.js`** — identical to existing `vite.config.ts`. Vite prefers `.ts`, so the `.js` copy was dead weight. ([phase5-fixes])

[phase5-fixes]: completions/phase5-fixes.md

---

## 0.6.0 — Agent Chat (2026-02-24)

Phase 5 — WebSocket-based agent chat with multi-turn conversation context and persistence.

### Database

- Migration `009_add_user_id_to_conversations`: user-scopes the conversations table. ([phase5])
- Rewrote sqlc queries — all reads/writes now scoped by `user_id`. Added `UpdateConversationTitleByUser`. ([phase5])

### Auth

- Extracted `ValidateJWT(tokenStr, jwtSecret)` helper from HTTP middleware — shared by REST routes and WebSocket auth. ([phase5])
- WebSocket authentication via `?token=` query parameter — validates before upgrade, rejects with HTTP 401 if invalid. ([phase5])

### Agent

- Added `Message` domain type in `agent/message.go` — represents stored chat messages. ([phase5])
- Added `RunConversation(ctx, userID, history, input)` — converts stored messages to Claude params for multi-turn context. ([phase5])
- `RunLoop` now delegates to `RunConversation` with nil history — backward compatible. ([phase5])

### Backend

- Rewrote `HandleChat` WebSocket handler: JWT auth, conversation create/load, message loop with agent integration, persistence, status signaling. ([phase5])
- New `GET /api/conversations` — user-scoped list (summaries without messages). ([phase5])
- New `GET /api/conversations/:id` — full conversation with messages. ([phase5])

### Frontend

- `useWebSocket` accepts auth options — appends `token` and `conversation_id` as query params. ([phase5])
- `useAgent` handles four message types: `system`, `status`, `error`, and chat messages. Exposes `conversationId`, `isThinking`, `error`, `loadMessages`. ([phase5])
- New `api/conversations.ts` — `listConversations()` and `getConversation(id)`. ([phase5])
- `AgentChatPage` shows connection status indicator, loads conversation history on mount, displays error banner. ([phase5])
- `ChatWindow` shows animated thinking indicator, auto-scrolls on new messages. ([phase5])
- `ChatInput` disables during thinking and when disconnected. ([phase5])

[phase5]: completions/phase5-agent-chat.md

---

## 0.5.0 — Agent Loop (2026-02-23)

Phase 4 — Claude API tool-use integration, turning Heimdall from a log viewer into an AI monitoring agent.

### Dependencies

- Added `anthropic-sdk-go v1.26.0` — official Go SDK for Claude API. ([phase4])

### Agent

- Refactored `Agent` struct with real fields: `*db.Queries`, `*anthropic.Client`, `*config.Config`. ([phase4])
- Replaced custom `ToolDefinition`/`ToolParam` types with SDK-native `anthropic.ToolUnionParam`. ([phase4])
- Registered two tools: `search_logs` (severity filter, paginated results) and `query_database` (user-scoped, read-only). ([phase4])
- Implemented `RunLoop(ctx, userID, input)` — full Claude API tool-use loop with max 10 iterations. ([phase4])
- Tool errors returned as `isError` tool results — Claude handles failures gracefully. ([phase4])
- System prompt scoped to registered tools only — avoids wasted loop iterations on nonexistent tools. ([phase4])

### Postgres Connector

- Real implementation in `connectors/database/postgres.go`: parses config JSONB, connects with `default_transaction_read_only=on`, executes queries, returns `[]map[string]any`. ([phase4])
- Read-only enforcement at the PostgreSQL session level — prevents mutations regardless of SQL content. ([phase4])

### Backend

- Wired Agent into Server: `main.go` creates Agent → `NewRouter(cfg, pool, ag)` → `NewServer(cfg, pool, ag)`. ([phase4])
- `WriteTimeout` raised to 5 minutes — agent loop makes multiple Claude API calls that exceed the previous 30s limit. ([phase4])
- `GET /api/agent/config` now reads from DB with fallback defaults. ([phase4])
- `PUT /api/agent/config` now persists via `UpsertAgentConfig`. ([phase4])
- New `POST /api/agent/run` endpoint — JWT-protected, synchronous test harness for the agent loop. ([phase4])

[phase4]: completions/phase4-agent-loop.md

---

## 0.4.0 — Webhook Log Connector (2026-02-23)

Phase 3 — webhook-based log ingestion, user-scoped log queries, and frontend log display.

### Database

- Rewrote sqlc queries in `log_buffer.sql` — all reads now join through `connections` to scope by `user_id`. ([phase3])
- Added `LIMIT`/`OFFSET` pagination to all list queries. ([phase3])
- Added `CountLogsByUser` query for paginated response totals. ([phase3])
- Added `GetConnectionByWebhookToken` query for webhook auth. ([phase3])
- `InsertLogEntry` now returns the inserted row (`:one` instead of `:exec`). ([phase3])
- Ran `sqlc generate` — regenerated `log_buffer.sql.go`. ([phase3])

### Backend

- New `POST /api/webhooks/logs` endpoint — accepts log payloads (single or batch), authenticates via per-connection webhook token. ([phase3])
- Restructured `/api` router: public webhook route + `r.Group(...)` for JWT-protected routes — fixes Chi subrouter precedence. ([phase3])
- Migration `008_add_webhook_token_index`: partial functional index on `config->>'webhook_token'` for webhook auth lookups. ([phase3])
- Auto-generates `webhook_token` (32-byte hex) in connection `config` when creating `webhook_logs` type connections. ([phase3])
- Replaced stub `ListLogs` handler with real implementation — supports `severity`, `connection_id`, `limit`, `offset` query params. ([phase3])
- Paginated JSON response: `{ data, total, limit, offset }`. ([phase3])

### Frontend

- Updated API layer for paginated response shape (`PaginatedLogs` interface). ([phase3])
- Expanded logs store with pagination state (`total`, `limit`, `offset`), `error` handling, `nextPage`/`prevPage` actions. ([phase3])
- `LogFilters` now includes connection dropdown for filtering by source. ([phase3])
- `LogEntry` displays payload content (extracts `message` field or shows formatted JSON). ([phase3])
- `LogFeed` shows pagination controls and "X–Y of Z" summary. ([phase3])
- `AgentLogPage` wires connections store for filter dropdown, handles pagination and error display. ([phase3])

[phase3]: completions/phase3-webhook-log-connector.md

---

## 0.3.0 — Connections CRUD (2026-02-23)

Phase 2 — full CRUD for connections, scoped to the authenticated user.

### Database

- Migration `007_add_user_id_to_connections`: added `user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE` with index. ([phase2])
- Rewrote sqlc queries to scope all operations by `user_id` (`ListConnectionsByUser`, `GetConnectionByUser`, `DeleteConnectionByUser`). ([phase2])
- Ran `sqlc generate` — updated `connections.sql.go` and `models.go`. ([phase2])

### Backend

- Replaced 5 stub handlers with real implementations: List, Get, Create, Update, Delete. ([phase2])
- All handlers extract user UUID from JWT context; return 401 if absent. ([phase2])
- Create defaults: `direction` → `one_way`, `config` → `{}`, `status` → `inactive`. Returns 201. ([phase2])
- Delete returns 204 No Content. ([phase2])

### Frontend

- Expanded Pinia store with `createConnection`, `updateConnection`, `deleteConnection` actions and error state. ([phase2])
- `ConnectionForm` now includes direction field, emits typed `CreateConnectionPayload`, supports cancel. ([phase2])
- `ConnectionCard` shows direction, has Delete button. ([phase2])
- `ConnectionList` shows empty-state message, forwards delete events. ([phase2])
- `ConnectionsPage` has "New Connection" toggle, error banner, and full create/delete flow. ([phase2])

[phase2]: completions/phase2-connections-crud.md

---

## 0.2.4 — Auth Me Endpoint & DB Pool (2026-02-22)

Phase 1 (Auth) Task 4 — `/api/auth/me` endpoint and database connection pool.

### Server

- Created `pgxpool.Pool` in `main.go` from `DATABASE_URL` — first real database connection in the server. ([phase1-task4])
- `Server` struct now holds `*db.Queries`; `NewServer` accepts the pool and wraps it with `db.New(pool)`. ([phase1-task4])

### Endpoint

- Added `GET /api/auth/me` — returns the authenticated user's `id`, `email`, and `created_at` from `public.users`. ([phase1-task4])
- Uses `UserIDFromContext` (Task 2) to read the JWT subject and `GetUser` (Task 1) to query the database. ([phase1-task4])

### Dependencies

- Promoted `golang-jwt/jwt/v5`, `google/uuid`, `jackc/pgx/v5` from indirect to direct in `go.mod`. ([phase1-task4])
- Added `pgxpool` transitive deps (`jackc/puddle/v2`, `x/sync`). ([phase1-task4])

[phase1-task4]: completions/phase1-task4-auth-me.md

---

## 0.2.3 — Auth-Protected Routes (2026-02-22)

Phase 1 (Auth) Task 3 — apply JWT middleware to protected routes.

### Router

- Applied `middleware.Auth(cfg.SupabaseJWTSecret)` to the `/api` route group — all `/api/*` requests now require a valid Supabase JWT. ([phase1-task3])
- Removed `POST /api/auth/login` stub — Supabase Auth handles login directly. ([phase1-task3])
- Removed dead `Login` handler from `handlers/auth.go`. ([phase1-task3])
- `/ws/chat` remains unprotected — WebSocket auth deferred to Phase 5. ([phase1-task3])

[phase1-task3]: completions/phase1-task3-auth-routes.md

---

## 0.2.2 — JWT Verification Middleware (2026-02-22)

Phase 1 (Auth) Task 2 — backend JWT verification.

### Auth

- Replaced stub auth middleware with real Supabase JWT validation (`internal/api/middleware/auth.go`). ([phase1-task2])
- Validates HMAC-SHA256 signature, checks expiry, extracts user UUID from `sub` claim. ([phase1-task2])
- Added `Auth(jwtSecret) → middleware` constructor and `UserIDFromContext(ctx)` context helper. ([phase1-task2])

### Config

- Added `SupabaseJWTSecret` field to `Config`, loaded from `SUPABASE_JWT_SECRET` env var, required in `Validate()`. ([phase1-task2])

### Dependencies

- Added `github.com/golang-jwt/jwt/v5`. ([phase1-task2])

[phase1-task2]: completions/phase1-task2-jwt-middleware.md

---

## 0.2.1 — Users Table & Auth Groundwork (2026-02-22)

Phase 1 (Auth) Task 1 — database foundation for user identity.

### Database

- Added migration `006_create_users`: `public.users` table with FK to `auth.users(id) ON DELETE CASCADE`. ([phase1-task1])
- Added `handle_new_user()` trigger function (`SECURITY DEFINER`) + `on_auth_user_created` trigger — auto-syncs Supabase Auth sign-ups into `public.users`. ([phase1-task1])
- Applied migration against live Supabase instance.

### sqlc

- Added `GetUser` query (`internal/db/queries/users.sql`). ([phase1-task1])
- Ran `sqlc generate` — produced `User` struct in `models.go` and `GetUser` method in `users.sql.go`. ([phase1-task1])

[phase1-task1]: completions/phase1-task1-users-migration.md

---

## 0.2.0 — Supabase Database (2026-02-22)

Connected Heimdall to a live Supabase Postgres instance and ran all migrations.

### Database

- Switched from local Postgres to Supabase (direct connection, port 5432).
- Added `cmd/dbping` utility — standalone connection smoke test (`SELECT 1`).
- Ran all 5 migrations against Supabase: `connections`, `agent_config`, `investigations`, `conversations`, `log_buffer`.
- Installed `golang-migrate` CLI (`go install` with `postgres` tag).

### Config

- `DATABASE_URL` is the sole database env var — no separate host/port/user/password fields needed.

---

## 0.1.3 — Infrastructure & DevOps (2026-02-20)

Docker production readiness, database tooling, and developer onboarding.

### Docker

- Added multi-stage frontend Dockerfile: builds with Node, serves with nginx. ([fix-013])
- Added multi-stage backend Dockerfile: builds with Go, runs minimal Alpine binary. ([fix-013])
- Added nginx `try_files` SPA routing so all frontend routes resolve to `index.html`. ([fix-026])
- Fixed frontend port mapping in `docker-compose.yml` — was `5173:5173` (Vite dev), now `3000:80` (nginx). ([fix-026])
- Removed stale dev-mode volume mounts from frontend service. ([fix-026])
- Added Postgres 17 service to `docker-compose.yml` with persistent volume. ([fix-015])
- Added Postgres `pg_isready` healthcheck; backend depends on `service_healthy`. ([fix-026])
- Switched `Makefile` from deprecated `docker-compose` to `docker compose` (Compose v2). ([fix-014])

### Database

- Ran `sqlc generate` — produced type-safe Go code for all 5 query files (24 queries total). ([impl-sqlc])
- Generated: `db.go`, `models.go`, `connections.sql.go`, `agent_config.sql.go`, `investigations.sql.go`, `conversations.sql.go`, `log_buffer.sql.go`. ([impl-sqlc])
- Added index `idx_connections_status` on `connections.status` for filtered queries. ([fix-020])

### Docs & Config

- Created `.env.example` documenting all required and optional environment variables. ([fix-025])
- Added setup instructions, prerequisites, and migration guide to `README.md`. ([fix-025])

[fix-013]: completions/fix-013-add-dockerfiles.md
[fix-014]: completions/fix-014-docker-compose-v2.md
[fix-015]: completions/fix-015-add-postgres-service.md
[fix-020]: completions/fix-020-connections-status-index.md
[fix-025]: completions/fix-025-env-example-readme.md
[fix-026]: completions/fix-026-dockerfile-healthcheck.md
[impl-sqlc]: completions/impl-sqlc-generate.md

## 0.1.2 — Frontend Fixes (2026-02-20)

WebSocket data flow, routing guards, auth wiring, and component correctness.

### Composables & Stores

- Fixed `useAgent` — added `watch` on WebSocket data so agent replies are actually parsed and displayed. ([fix-004])
- Added axios request interceptor to attach `Authorization: Bearer` header from auth store. ([fix-022])
- Removed dead `useAuth` composable — logic already lived in `useAuthStore` Pinia store. ([fix-011])
- Extracted reports state into `useReportsStore` Pinia store for consistency with other pages. ([fix-024])

### Routing

- Added `beforeEach` auth guard — unauthenticated users redirect to `/login`. ([fix-021])
- Added catch-all `/:pathMatch(.*)*` route and `NotFoundPage.vue` for 404s. ([fix-021])

### Components

- Changed log severity class from `'error'` to `'critical'` to match the domain model type. ([fix-005])
- Updated `LogFilters.vue` dropdown value from `'error'` to `'critical'`. ([fix-005])
- Changed `ChatWindow.vue` `v-for` key from array index to `msg.id` (stable identity). ([fix-023])
- Added `id` field to `ChatMessage` type; assigned via `crypto.randomUUID()`. ([fix-023])

### Tooling

- Configured `typescript-eslint` parser so ESLint can lint `.ts` and `.vue` files. ([fix-012])

[fix-004]: completions/fix-004-useagent-websocket-data.md
[fix-005]: completions/fix-005-log-severity-mismatch.md
[fix-011]: completions/fix-011-remove-useauth-composable.md
[fix-012]: completions/fix-012-eslint-typescript-parser.md
[fix-021]: completions/fix-021-router-auth-guard-404.md
[fix-022]: completions/fix-022-axios-auth-interceptor.md
[fix-023]: completions/fix-023-chat-vfor-key.md
[fix-024]: completions/fix-024-reports-pinia-store.md

## 0.1.1 — Backend Fixes & Hardening (2026-02-20)

Dependency corrections, HTTP semantics, server hardening, and migration constraints.

### HTTP & Handlers

- Added `jsonError()` helper — replaces `http.Error` so JSON error responses keep `Content-Type: application/json`. ([fix-002])
- Introduced `Server` struct with `NewServer` constructor; converted all handlers from free functions to methods for dependency injection. ([fix-006])
- Changed `/api/auth/login` from `GET` to `POST`. ([fix-007])

### Agent & Connectors

- Registered 3 missing tools (`search_codebase`, `recall_similar_incidents`, `recall_lessons`) in `ToolRegistry()`. ([fix-003])
- Tool dispatch now returns an error for unknown tool names instead of silent empty success. ([fix-008])
- `Registry.Remove()` now calls `Close()` on the connector before deleting to prevent resource leaks. ([fix-010])

### Server

- Added `ReadTimeout`, `WriteTimeout`, `IdleTimeout` to `http.Server`. ([fix-017])
- Wrapped `ResponseWriter` in logging middleware to capture HTTP status codes. ([fix-018])

### Config

- Added `ElephantasmURL` and `ElephantasmKey` fields to `Config` struct. ([fix-016])
- Added `Config.Validate()` method; called on startup to fail fast on missing required env vars. ([fix-016])

### Migrations

- Enforced singleton constraint on `agent_config` (`id = 1` with `CHECK`). ([fix-009])
- Added `ON DELETE SET NULL` to `conversations.investigation_id` FK. ([fix-009])
- Added `ON DELETE CASCADE` to `log_buffer.connection_id` FK. ([fix-009])

### Dependencies

- Reclassified direct dependencies in `go.mod` (chi, websocket, etc. were incorrectly marked `// indirect`). ([fix-001])
- Removed 11 unused transitive entries from `go.mod` and `go.sum`. ([fix-001])

[fix-001]: completions/fix-001-go-mod-indirect.md
[fix-002]: completions/fix-002-http-error-content-type.md
[fix-003]: completions/fix-003-incomplete-tool-registry.md
[fix-006]: completions/fix-006-router-dependency-injection.md
[fix-007]: completions/fix-007-login-get-to-post.md
[fix-008]: completions/fix-008-unknown-tool-dispatch-error.md
[fix-009]: completions/fix-009-migration-fk-constraints.md
[fix-010]: completions/fix-010-registry-remove-close.md
[fix-016]: completions/fix-016-config-elephantasm-validate.md
[fix-017]: completions/fix-017-http-server-timeouts.md
[fix-018]: completions/fix-018-logging-status-code.md

---

## 0.1.0 — Scaffolding (2026-02-19)

Full monorepo scaffolding. Project structure, frontend, backend, database schema, and tooling — all initialised and compiling.

### Root

- Added `Makefile` with targets for dev, build, test, lint, sqlc, migrations, and Docker.
- Added `docker-compose.yml` (backend + frontend services, no local Postgres).
- Added `.gitignore` and `README.md`.

### Frontend (Vue 3 + Vite + TypeScript + Pinia + Tailwind v4)

- Initialised Vue 3 project with Vite, TypeScript, and Tailwind CSS v4.
- Installed runtime deps: `vue`, `vue-router`, `pinia`, `axios`.
- Created API client layer (`src/api/`) — one file per backend resource.
- Created 15 components across 5 domains: common, connections, agent chat, log feed, reports.
- Created 7 page views with Vue Router (lazy-loaded).
- Created 4 Pinia stores (auth, connections, agent, logs).
- Created 3 composables: `useWebSocket`, `useAgent`, `useAuth`.
- Created TypeScript types for all domain models + API envelope types.
- Created utility helpers (date formatting, app constants).
- Configured Vite dev proxy (`/api` → `:8080`, `/ws` → `ws://localhost:8080`).
- Verified: `vue-tsc` type-check passes with zero errors.

### Backend (Go + Chi v5 + pgx v5 + coder/websocket + anthropic-sdk-go)

- Initialised Go module (`github.com/crimson-sun/heimdall/backend`).
- Created HTTP server entry point with graceful shutdown.
- Created Chi router with all REST routes + WebSocket endpoint.
- Created 3 middleware layers: request logging, CORS, auth (passthrough placeholder).
- Created 6 handler files — all endpoints wired and responding (stubs).
- Created WebSocket chat handler with JSON read/write (placeholder echo).
- Created agent engine package: lifecycle, tool-use loop, monitoring goroutine, system prompt, 5 tool implementations (all stubs).
- Created connector layer: `Connector`, `StreamConnector`, `QueryConnector` interfaces + registry.
- Created 4 connector implementations: PostgreSQL, webhook logs, syslog, GitHub (all stubs with tests).
- Created Elephantasm memory client package (HTTP client, types, service layer).
- Created WebSocket hub package (hub, client, message types).
- Created report generation package (generator, templates).
- Verified: `go build ./...` and `go test ./...` both pass clean.

### Database

- Wrote 5 migration pairs (up/down) for: connections, agent_config, investigations, conversations, log_buffer.
- Wrote 5 sqlc query files covering all CRUD operations per the DB models spec.
- Created `sqlc.yaml` config (pgx/v5, uuid→uuid.UUID, jsonb→json.RawMessage, timestamptz→time.Time).
- Note: `sqlc generate` not yet run — requires a running Postgres instance.

### Docs

- Added `docs/overview-blueprint.md` — living architecture reference.
- Added `docs/completions/scaffolding.md` — detailed implementation notes.
