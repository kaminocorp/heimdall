# Changelog

- [0.2.1 — Users Table & Auth Groundwork](#021--users-table--auth-groundwork-2026-02-22)
- [0.2.0 — Supabase Database](#020--supabase-database-2026-02-22)
- [0.1.3 — Infrastructure & DevOps](#013--infrastructure--devops-2026-02-20)
- [0.1.2 — Frontend Fixes](#012--frontend-fixes-2026-02-20)
- [0.1.1 — Backend Fixes & Hardening](#011--backend-fixes--hardening-2026-02-20)
- [0.1.0 — Scaffolding](#010--scaffolding-2026-02-19)

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
