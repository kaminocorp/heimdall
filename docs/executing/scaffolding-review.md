# Scaffolding Review

**Date:** 2026-02-19
**Scope:** Full review of v0.1.0 scaffolding — root, frontend, backend, database

---

## Overall Verdict

Strong scaffolding. Both frontend and backend compile cleanly (`vue-tsc --noEmit` zero errors, `go build ./...` clean). Project structure is well-organised and architectural decisions are sound. Issues found are mostly about completeness and consistency rather than correctness.

---

## Must Fix

These will cause bugs, break tooling, or produce incorrect behaviour.

### 1. `go.mod` — all deps marked `// indirect`

**File:** `backend/go.mod` lines 6–18

Every dependency — including chi, pgx, websocket, anthropic-sdk-go — is tagged `// indirect`. This is incorrect; packages like `internal/api/router.go` directly import `chi/v5`. Running `go mod tidy` will either reclassify or remove them.

### 2. Handlers — `http.Error` overwrites Content-Type

**Files:** `backend/internal/api/handlers/connections.go` lines 14–15, 19–20, 24–25, 29–30; also `agent.go` line 18, `reports.go` line 14

Pattern: `w.Header().Set("Content-Type", "application/json")` followed by `http.Error(...)`. `http.Error` internally sets `Content-Type: text/plain`, overwriting the JSON header. JSON error responses are served as plain text.

**Fix:** Use `w.WriteHeader(status)` + `json.NewEncoder(w).Encode(...)`, or just accept plain text for stub errors.

### 3. `agent/tools.go` — incomplete tool registry

**File:** `backend/internal/agent/tools.go` lines 18–36 vs 40–53

`ToolRegistry()` only declares `query_database` and `search_logs`. Three other tools — `search_codebase`, `recall_similar_incidents`, `recall_lessons` — exist in the `Dispatch` switch but are never declared. The LLM will never call undeclared tools.

### 4. `useAgent` composable — WebSocket data never consumed

**File:** `frontend/src/composables/useAgent.ts` line 7

`data` from `useWebSocket` is destructured but never watched or consumed. Incoming agent messages will never appear in the chat. Needs a `watch` on `data` that parses and pushes into `messages`.

### 5. Log severity mismatch — `'error'` vs `'critical'`

**Files:** `frontend/src/components/log/LogEntry.vue` line 13; `frontend/src/components/log/LogFilters.vue` line 20

Components check for severity `'error'`, but the domain model (`types/log.ts`) and constants (`utils/constants.ts` — `SEVERITY_LEVELS = ['info', 'warning', 'critical']`) use `'critical'`. Mismatched string literals.

---

## Should Fix

Design issues that will cause friction or bugs as implementation progresses.

### 6. Router — no dependency injection

**File:** `backend/internal/api/router.go`

`NewRouter` accepts `*config.Config` but never uses it. Handlers are standalone functions with no access to DB pool, agent instance, or config. Needs a `Server` struct with methods as handlers, or closure-based handler factories.

### 7. Login endpoint is GET instead of POST

**File:** `backend/internal/api/router.go` line 38

`r.Get("/api/auth/login", ...)` — login should be POST since it accepts credentials. GET with a hardcoded token response is misleading even for a stub.

### 8. Unknown tool dispatch silently succeeds

**File:** `backend/internal/agent/tools.go` line 52

Unknown tool names return `"", nil`. Should return `fmt.Errorf("unknown tool: %s", name)` to surface dispatch errors.

### 9. Migration foreign key / constraint gaps

**Files:**
- `backend/migrations/002_*` — `agent_config` has no singleton constraint. Multiple rows allowed but app expects exactly one. Use `id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1)` or similar.
- `backend/migrations/004_*` — `conversations.investigation_id` FK has no `ON DELETE` clause. Deleting an investigation will cause FK violations. Consider `ON DELETE SET NULL`.
- `backend/migrations/005_*` — `log_buffer.connection_id` FK has no `ON DELETE CASCADE`. Deleting a connection leaves undeletable log entries.

### 10. Connector registry `Remove` doesn't close

**File:** `backend/internal/connectors/registry.go`

`Remove()` deletes from the map but never calls `Close()` on the removed connector. Resource leak risk. Either call `Close()` automatically or return the removed connector for the caller to close.

### 11. Duplicate auth logic — store vs composable

**Files:** `frontend/src/stores/auth.ts` + `frontend/src/composables/useAuth.ts`

Both provide token management and `isAuthenticated`. The composable uses a module-scope `ref` (poor man's store) and is never imported anywhere — dead code. Remove `useAuth` composable; the Pinia store is the correct pattern.

### 12. ESLint — no TypeScript parser

**File:** `frontend/eslint.config.js`

Config applies to `*.ts` files but no TypeScript parser is configured. ESLint will attempt to parse TS syntax as JS and fail on type annotations. Add `typescript-eslint` to devDependencies and include its flat config.

### 13. docker-compose.yml references nonexistent Dockerfiles

**File:** `docker-compose.yml` lines 4–5, 14–15

Backend service references `./backend/Dockerfile`, frontend references `./frontend/Dockerfile`. Neither exists. `docker compose up` will fail immediately.

### 14. Makefile uses deprecated `docker-compose`

**File:** `Makefile` lines 51, 54

`docker-compose` (hyphenated) is the legacy Python CLI, deprecated in favour of `docker compose` (space-separated, Go-based). Will fail on systems with only Compose v2.

### 15. No Postgres service in docker-compose.yml

**File:** `docker-compose.yml`

Blueprint (`docs/plans/blueprint.md` line 215) describes docker-compose as providing "backend + frontend + postgres". The actual file has no Postgres service. Either add one or update the blueprint.

---

## Nice to Have

Improvements that would increase quality but aren't blocking.

### 16. Config missing Elephantasm fields and validation

**File:** `backend/internal/config/config.go`

No `ElephantasmURL` or `ElephantasmAPIKey` fields despite `internal/memory/client.go` needing them. No `Validate() error` method to catch missing required config at startup.

### 17. No HTTP server timeouts

**File:** `backend/cmd/heimdall/main.go`

`http.Server` has no `ReadTimeout` or `WriteTimeout`. Even placeholder values (e.g. 30s) signal production intent.

### 18. Logging middleware doesn't capture status code

**File:** `backend/internal/api/middleware/logging.go`

`http.ResponseWriter` is passed through without wrapping. Logged requests have no status code. Needs a `ResponseWriter` wrapper that captures `WriteHeader` calls.

### 19. Upsert uses positional params instead of `EXCLUDED`

**File:** `backend/internal/db/queries/agent_config.sql` lines 6–8

`ON CONFLICT (id) DO UPDATE SET model = $1` — idiomatic PostgreSQL uses `EXCLUDED.model`. Positional params work but are less readable and more error-prone.

### 20. No index on `connections.status`

**File:** `backend/migrations/001_*/up.sql`

`UpdateConnectionStatus` queries by ID (indexed by PK), but if connections are ever listed by status, an index would help. Low priority until query patterns are known.

### 21. Router — no auth guard or 404 route

**File:** `frontend/src/router/index.ts`

No `router.beforeEach` navigation guard for auth redirects. No catch-all `/:pathMatch(.*)*` route for 404s.

### 22. Axios client — no auth interceptor

**File:** `frontend/src/api/client.ts`

Auth store holds a token but the axios instance never attaches it to requests. Needs a request interceptor that sets the `Authorization` header.

### 23. `ChatWindow.vue` uses index as `v-for` key

**File:** `frontend/src/components/agent/ChatWindow.vue` line 18

`v-for="(msg, i) in messages" :key="i"` — array index keys cause rendering issues when the list changes. Use a stable identifier (message id or timestamp).

### 24. Inconsistent page state management

**File:** `frontend/src/pages/ReportsPage.vue` vs other pages

`ReportsPage` uses local `ref`s for state while `ConnectionsPage`, `AgentLogPage`, etc. use Pinia stores. Either create a reports store or standardise on local state.

### 25. No `.env.example` or setup prerequisites in README

**Files:** root directory (missing `.env.example`); `README.md`

`docker-compose.yml` references `env_file: .env` but there's no `.env.example` documenting required variables. README has no prerequisites section (Go, Node, Docker, golang-migrate, sqlc) and no database setup instructions.

---

## Things Done Well

- **Go project structure** — `cmd/`, `internal/`, clean package separation follows standard conventions.
- **Connector interfaces** — three-tier hierarchy (`Connector` / `StreamConnector` / `QueryConnector`) is excellent, composable Go interface design.
- **Thread safety** — `connectors/registry.go` and `ws/hub.go` use `sync.RWMutex` correctly with defensive copies.
- **SQL migrations** — good index coverage, especially compound and partial indexes on `log_buffer`.
- **Graceful shutdown** in `main.go` is textbook correct.
- **Vue 3 Composition API** used exclusively with `<script setup lang="ts">` — no Options API anywhere.
- **TypeScript rigour** — `Record<string, unknown>` over `any`, proper generics on API calls, union literal types.
- **Tailwind v4** correctly set up with `@tailwindcss/vite` plugin and `@import "tailwindcss"`.
- **Code splitting** — all page routes lazy-loaded.
- **Consistent patterns** — every list page follows the same fetch-on-mount + spinner + list pattern.
- **Both `vue-tsc` and `go build ./...` pass clean.**
