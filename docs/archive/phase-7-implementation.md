# Phase 7 — Implementation Plan

Reference: [Post-MVP Roadmap](./post-mvp-roadmap.md) · [Vision](../vision.md)

---

## Pre-Implementation Notes

After auditing the codebase, several items listed in the roadmap are already complete:

| Roadmap Item | Actual Status |
|---|---|
| Login page (hardcoded token → Supabase) | **Done** — already uses `@supabase/supabase-js` with email/password login, token refresh, and logout |
| Dashboard page (placeholder) | **Partially done** — shows agent status, connection health, and recent activity. Needs polish, not a rewrite |
| Responsive layout | **Partially done** — sidebar has mobile hamburger/overlay. Individual pages need responsive audit |

The remaining work focuses on: agent config editing, toast system, loading skeletons, dashboard enhancement, and full test coverage.

---

## Phase 7a — UI Polish

### 7a.1 — Toast / Notification System

**Why first:** Every subsequent task (config save, error boundaries) needs a way to show transient feedback. Build the primitive first.

**Current state:** Errors are shown inline per-page via local `actionError` refs and red-bordered `<div>` blocks. No global notification mechanism exists.

#### Tasks

1. **Create `useToast` composable** (`frontend/src/composables/useToast.ts`)
   - Reactive array of toast objects: `{ id, message, type, duration }`
   - Types: `success`, `error`, `info`
   - `show(message, type?, duration?)` — adds toast, auto-removes after duration (default 4s)
   - `dismiss(id)` — manual removal
   - Singleton pattern via module-level state (shared across components without a provider)

2. **Create `ToastContainer` component** (`frontend/src/components/common/ToastContainer.vue`)
   - Fixed position bottom-right (`fixed bottom-4 right-4 z-50`)
   - Renders active toasts with enter/leave transitions
   - Styling uses existing design tokens: `--bg-elevated`, `--border`, status colours for type
   - Each toast has a dismiss button (×)

3. **Mount `ToastContainer` in `App.vue`**
   - Add once at app root level, outside router-view

4. **Migrate existing inline errors** (optional, can be done incrementally)
   - ConnectionsPage, LoginPage — keep inline errors for form validation, use toasts for action confirmations

#### Files

| Action | File |
|---|---|
| Create | `frontend/src/composables/useToast.ts` |
| Create | `frontend/src/components/common/ToastContainer.vue` |
| Edit | `frontend/src/App.vue` (mount ToastContainer) |

---

### 7a.2 — Agent Config Editing

**Why next:** The only major feature gap in the UI. Backend `PUT /api/agent/config` already works — this is purely frontend.

**Current state:** `AgentConfigPage.vue` displays config as a read-only key-value list. The agent store (`stores/agent.ts`) has `fetchConfig()` but no `updateConfig()`. The API module (`api/agent.ts`) already exports `updateAgentConfig()`.

#### Tasks

1. **Add `updateConfig` action to agent store** (`frontend/src/stores/agent.ts`)
   - Calls `agentApi.updateAgentConfig(data)`
   - Updates local `config` state on success
   - Returns promise for caller to handle success/error

2. **Rewrite `AgentConfigPage.vue` with edit form**
   - Display mode (current): shows config values as read-only text
   - Edit mode: toggled via "Edit Configuration" button
   - Form fields:
     - **Model** — text `<input>` with `<datalist>` of known models (`claude-sonnet-4-5-20250929`, `claude-haiku-4-5-20251001`, `claude-opus-4-20250115`). Free-text allows new model IDs without code changes — the backend casts the string directly to `anthropic.Model()` in `loop.go:32`
     - **Mode** — `<select>` with options: `continuous`, `scheduled`, `off`
     - **Schedule** — text input, conditionally shown when mode is `scheduled` (cron expression)
     - **System Prompt Override** — `<textarea>` with monospace font, placeholder showing purpose
   - Cancel button reverts to display mode without saving
   - Save button calls `agentStore.updateConfig()`, shows toast on success/error, returns to display mode
   - Loading state on save button

3. **Form styling** — consistent with existing design tokens:
   - Inputs: `bg-bg-surface border border-border text-text-primary` with focus ring
   - Use same patterns as `ConnectionForm.vue` for consistency

#### Files

| Action | File |
|---|---|
| Edit | `frontend/src/stores/agent.ts` (add updateConfig) |
| Edit | `frontend/src/pages/AgentConfigPage.vue` (add edit form) |

---

### 7a.3 — Dashboard Enhancement

**Why:** The dashboard works but can be more useful. Small improvements, not a rewrite.

**Current state:** Three sections — system status card (hardcoded "Active" agent status), connections card (count + list), recent activity (last 8 log entries). No loading states, no error handling, no log ingestion rate.

#### Tasks

1. **Add loading skeletons to dashboard**
   - Show pulse-animated placeholder blocks while stores are loading
   - Use a shared `SkeletonBlock` component (see 7a.4)

2. **Fix hardcoded agent status**
   - Currently always shows "Active" — should derive from `agentConfig.mode`:
     - `continuous` → "Active"
     - `scheduled` → "Scheduled" (with schedule value)
     - `off` → "Inactive"

3. **Add log ingestion rate**
   - New stat card showing approximate logs/hour or logs/day
   - Can be derived from `logsStore.total` and the timestamp range of visible entries
   - Or: add a lightweight backend endpoint `GET /api/stats` that returns `{ log_count_24h, connection_count, active_connections }` — keeps the dashboard fast with a single call

4. **Add error state**
   - If any store fetch fails, show an inline error with retry button instead of empty content

#### Files

| Action | File |
|---|---|
| Edit | `frontend/src/pages/DashboardPage.vue` |
| Create (if backend stats route chosen) | `backend/internal/api/handlers/stats.go` |
| Create (if backend stats route chosen) | `backend/internal/db/queries/stats.sql` |

**Decision: backend `GET /api/stats` endpoint.** A single query is cleaner than fetching all logs client-side just to count them.

---

### 7a.4 — Loading Skeletons

**Why:** Multiple pages fetch data on mount with no visual feedback during loading. Users see empty content or layout shifts.

**Current state:** Some pages have a simple `v-if="loading"` with a spinner div. No skeleton loaders exist.

#### Tasks

1. **Create `SkeletonBlock` component** (`frontend/src/components/common/SkeletonBlock.vue`)
   - Props: `width`, `height`, `rounded` (boolean)
   - Renders a div with pulse animation (`animate-pulse` from Tailwind)
   - Uses `--bg-surface` background with subtle shimmer

2. **Create page-level skeleton layouts**
   - `ConnectionsSkeleton` — mimics 3 connection cards
   - `LogsSkeleton` — mimics a log table with 8 rows
   - `ConfigSkeleton` — mimics the key-value config display
   - These can be inline in each page rather than separate components

3. **Apply to pages**
   - `ConnectionsPage.vue` — skeleton while `connectionsStore.loading`
   - `AgentLogPage.vue` — skeleton while `logsStore.loading`
   - `AgentConfigPage.vue` — skeleton while `agentStore.loading`
   - `DashboardPage.vue` — skeleton per card while respective stores load
   - `ReportsPage.vue` — skeleton while `reportsStore.loading`

#### Files

| Action | File |
|---|---|
| Create | `frontend/src/components/common/SkeletonBlock.vue` |
| Edit | `ConnectionsPage.vue`, `AgentLogPage.vue`, `AgentConfigPage.vue`, `DashboardPage.vue`, `ReportsPage.vue` |

---

### 7a.5 — Error Boundaries

**Why:** Unhandled promise rejections and API errors currently fail silently or show browser console errors only.

#### Tasks

1. **Global error handler in Vue app** (`frontend/src/main.ts`)
   - `app.config.errorHandler` — catches unhandled component errors, logs to console, shows toast
   - `window.addEventListener('unhandledrejection')` — catches unhandled promise rejections, shows toast

2. **Axios response interceptor** (`frontend/src/api/client.ts`)
   - On 401: trigger `authStore.logout()` and redirect to `/login` (session expired)
   - On 5xx: show error toast with generic message
   - On network error: show "Connection lost" toast
   - Individual API calls can still `.catch()` for specific handling (interceptor is the safety net)

#### Files

| Action | File |
|---|---|
| Edit | `frontend/src/main.ts` (error handler) |
| Edit | `frontend/src/api/client.ts` (response interceptor) |

---

### 7a.6 — Responsive Audit

**Why:** The sidebar is mobile-ready, but page content may overflow or look cramped on smaller screens.

#### Tasks

1. **Audit each page at 375px (mobile) and 768px (tablet) widths**
   - Dashboard cards — stack vertically on mobile (likely already works with flex-wrap)
   - Connections cards — ensure full-width on mobile
   - Agent log table — horizontal scroll wrapper if needed
   - Chat — already likely full-height, verify input doesn't get hidden by keyboard
   - Config page — form fields full-width on mobile

2. **Fix issues found during audit**
   - Mostly Tailwind responsive prefixes (`sm:`, `md:`, `lg:`)
   - No major layout rework expected

#### Files

| Action | File |
|---|---|
| Edit | Various page components (as needed based on audit findings) |

---

## Phase 7b — Test Coverage

### 7b.1 — Go Test Infrastructure

**Why first:** Establish patterns and helpers before writing individual tests.

**Current state:** 3 trivial stub tests, no test framework, no test helpers, no mocks. Uses only stdlib `testing`.

#### Tasks

1. **Add testify dependency**
   ```
   go get github.com/stretchr/testify
   ```
   - `assert` for readable assertions
   - `require` for fatal assertions
   - `mock` for interface mocking

2. **Create test helpers** (`backend/internal/api/handlers/testhelpers_test.go`)
   - `newTestServer(t)` — creates a `Server` with:
     - In-memory or test database connection (see decision below)
     - Mock agent (or nil agent for non-agent tests)
     - Test config
   - `authHeader(userID)` — generates a valid JWT for test requests
   - `makeRequest(method, path, body, headers)` — shorthand for `httptest` request/response

3. **Create mock agent** (`backend/internal/agent/mock_test.go` or `backend/internal/testutil/`)
   - Implements the same interface the handlers expect
   - Returns configurable responses for `RunLoop` / `RunConversation`

**Decision: tests run against the real production Supabase database.** This gives the most realistic coverage — no mocks hiding SQL bugs. However, since there is only one database, **every test must clean up after itself**.

#### Cleanup strategy

- Each test creates its own test user (unique email) and all test data is scoped to that user via `user_id`
- Tests use a `testSetup(t)` helper that:
  1. Connects to the prod DB using env config
  2. Creates a dedicated test user row
  3. Returns the pool, queries, and user ID
  4. Registers a `t.Cleanup()` function that deletes the test user — cascading `ON DELETE CASCADE` foreign keys wipe all related data (connections, conversations, logs, etc.)
- Tests must **never** modify global tables (`agent_config`) without restoring original state
- Tests must **never** hardcode IDs — always generate fresh UUIDs
- If a test panics or fails, `t.Cleanup()` still runs (Go guarantees this), so orphaned data is prevented

#### Files

| Action | File |
|---|---|
| Edit | `backend/go.mod` (add testify) |
| Create | `backend/internal/api/handlers/testhelpers_test.go` |

---

### 7b.2 — Handler Tests

**Why:** Handlers are the API surface — bugs here affect every user.

#### Tasks

1. **Connections handler tests** (`backend/internal/api/handlers/connections_test.go`)
   - `TestListConnections` — returns user-scoped connections
   - `TestCreateConnection` — validates required fields, returns created connection
   - `TestCreateConnection_InvalidType` — returns 400
   - `TestUpdateConnection` — updates fields, returns updated
   - `TestDeleteConnection` — returns 204
   - `TestTestConnection_Postgres` — validates connectivity test flow

2. **Agent config handler tests** (`backend/internal/api/handlers/agent_test.go`)
   - `TestGetAgentConfig_Default` — returns defaults when no DB row
   - `TestGetAgentConfig_Existing` — returns stored config
   - `TestUpdateAgentConfig` — upserts and returns updated config

3. **Logs handler tests** (`backend/internal/api/handlers/logs_test.go`)
   - `TestListLogs_All` — returns combined log_buffer + agent_log
   - `TestListLogs_FilterSource` — source=raw returns only log_buffer
   - `TestListLogs_Pagination` — respects limit/offset

4. **Webhook handler tests** (`backend/internal/api/handlers/webhooks_test.go`)
   - `TestIngestWebhookLogs` — valid payload inserts to log_buffer
   - `TestIngestWebhookLogs_InvalidToken` — returns 401
   - `TestIngestWebhookLogs_MissingFields` — returns 400
   - `TestIngestWebhookLogs_BatchArray` — accepts array of entries

5. **Auth handler tests** (`backend/internal/api/handlers/auth_test.go`)
   - `TestMe` — returns authenticated user
   - `TestMe_Unauthenticated` — returns 401

#### Files

| Action | File |
|---|---|
| Create | `backend/internal/api/handlers/connections_test.go` |
| Create | `backend/internal/api/handlers/agent_test.go` |
| Create | `backend/internal/api/handlers/logs_test.go` |
| Create | `backend/internal/api/handlers/webhooks_test.go` |
| Create | `backend/internal/api/handlers/auth_test.go` |

---

### 7b.3 — Agent Loop Tests

**Why:** The agent loop is the core intelligence layer. Tool dispatch, iteration limits, and error handling must be verified.

#### Tasks

1. **Create mock Anthropic client**
   - Returns configurable responses (text-only, tool-use, multi-turn)
   - Tracks calls for assertion

2. **Agent loop tests** (`backend/internal/agent/loop_test.go`)
   - `TestRunLoop_SimpleResponse` — Claude returns text, loop returns it
   - `TestRunLoop_ToolUse` — Claude requests tool, agent dispatches, Claude gets result, returns final text
   - `TestRunLoop_MaxIterations` — loop stops after 10 tool-use rounds
   - `TestRunLoop_ToolError` — tool returns error, sent as `isError` tool result

3. **Tool dispatch tests** (`backend/internal/agent/tools_test.go`)
   - `TestDispatch_UnknownTool` — returns error
   - `TestDispatch_SearchLogs` — correct delegation
   - `TestDispatch_QueryDatabase` — correct delegation, user scoping

#### Files

| Action | File |
|---|---|
| Create | `backend/internal/agent/loop_test.go` |
| Create | `backend/internal/agent/tools_test.go` |

---

### 7b.4 — Frontend Test Infrastructure

**Why:** No test runner exists. Install and configure before writing tests.

**Current state:** `"test": "echo \"No tests yet\""` in `package.json`. No vitest or jest.

#### Tasks

1. **Install vitest + testing utilities**
   ```
   npm install -D vitest @vue/test-utils happy-dom
   ```
   - `vitest` — Vite-native test runner (zero extra config with existing Vite setup)
   - `@vue/test-utils` — official Vue component testing
   - `happy-dom` — lightweight DOM implementation (faster than jsdom)

2. **Configure vitest** (`frontend/vitest.config.ts` or inline in `vite.config.ts`)
   - Environment: `happy-dom`
   - Globals: `true` (no manual imports of describe/it/expect)
   - Setup file for Pinia test plugin

3. **Update package.json scripts**
   - `"test": "vitest run"`
   - `"test:watch": "vitest"`

4. **Create test setup file** (`frontend/src/test/setup.ts`)
   - Install Pinia testing plugin
   - Mock `axios` globally or provide a mock API client

#### Files

| Action | File |
|---|---|
| Edit | `frontend/package.json` (add deps + scripts) |
| Edit | `frontend/vite.config.ts` (add test config) |
| Create | `frontend/src/test/setup.ts` |

---

### 7b.5 — Frontend Store Tests

**Why:** Stores are the data layer — correctness here prevents cascading UI bugs.

#### Tasks

1. **Auth store tests** (`frontend/src/stores/__tests__/auth.test.ts`)
   - `login` calls Supabase `signInWithPassword`, sets session + user
   - `logout` clears state
   - `init` restores session from Supabase cache
   - `isAuthenticated` computed tracks session presence

2. **Connections store tests** (`frontend/src/stores/__tests__/connections.test.ts`)
   - `fetchConnections` populates array
   - `createConnection` adds to array
   - `deleteConnection` removes from array
   - `testConnection` sets testingId during test
   - Error states handled correctly

3. **Logs store tests** (`frontend/src/stores/__tests__/logs.test.ts`)
   - `fetchLogs` populates entries + total
   - `nextPage` / `prevPage` adjust offset
   - `setSource` resets pagination and re-fetches

4. **Agent store tests** (`frontend/src/stores/__tests__/agent.test.ts`)
   - `fetchConfig` stores config
   - `updateConfig` sends PUT and updates local state

#### Files

| Action | File |
|---|---|
| Create | `frontend/src/stores/__tests__/auth.test.ts` |
| Create | `frontend/src/stores/__tests__/connections.test.ts` |
| Create | `frontend/src/stores/__tests__/logs.test.ts` |
| Create | `frontend/src/stores/__tests__/agent.test.ts` |

---

## Implementation Order

Phase 7a and 7b are independent tracks and can be worked in parallel. Within each track, the ordering is sequential — each step builds on the previous.

```
Phase 7a (UI Polish)                    Phase 7b (Test Coverage)
─────────────────────                   ────────────────────────
7a.1  Toast system                      7b.1  Go test infrastructure
  ↓                                       ↓
7a.2  Agent config editing              7b.2  Handler tests
  ↓                                       ↓
7a.3  Dashboard enhancement             7b.3  Agent loop tests
  ↓                                       ↓
7a.4  Loading skeletons                 7b.4  Frontend test infrastructure
  ↓                                       ↓
7a.5  Error boundaries                  7b.5  Frontend store tests
  ↓
7a.6  Responsive audit
```

### Estimated Scope

| Track | Sections | Files Created | Files Edited |
|---|---|---|---|
| 7a | 6 | ~4–6 | ~10–12 |
| 7b | 5 | ~10–12 | ~4 |

---

## Resolved Decisions

1. **Dashboard stats** — backend `GET /api/stats` endpoint (single query, keeps dashboard fast)
2. **Test database strategy** — real production Supabase DB with per-test user cleanup via `ON DELETE CASCADE`
3. **Model selector** — free-text `<input>` with `<datalist>` of known models (backend accepts any string via `anthropic.Model()` cast)
