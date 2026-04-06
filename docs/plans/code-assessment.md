# Code Assessment — Heimdall Repository

**Date:** 2026-04-06 (supersedes 2026-03-22 assessment)
**Scope:** Full repository — backend (Go), frontend (Vue 3 + TypeScript), infrastructure (migrations, Docker, config)
**Assessment criteria:** Functionality, accuracy, maintainability, clean code. Files >500 lines flagged for refactoring.

---

## Executive Summary

The codebase is in strong shape. All 12 issues from the March 22 assessment have been addressed — SQL injection patched, XSS escaped, CORS locked down, RLS added, search_logs fixed, pagination corrected, and timer leaks cleaned up. The remaining issues are maintenance-level: dead code that should be removed, one instance of code duplication worth consolidating, and two files slightly over the 500-line threshold.

**Overall score: 8.5/10** — up from 7.5. No critical or high-severity issues remain.

---

## Previous Assessment Status

All 12 issues from the 2026-03-22 assessment are resolved:

| # | Issue | Status |
|---|-------|--------|
| 1 | SQL injection in Supabase table name | **Fixed** — allowlist validation in `supabase.go:29-38` |
| 2 | XSS in email HTML | **Fixed** — `html.EscapeString()` in `format.go:55` |
| 3 | `search_logs` ignores query param | **Fixed** — reads `input["query"]` in `tools_logs.go:38` |
| 4 | Chat uses wrong agent config | **Fixed** — `loop.go:36` uses `GetAppAgentConfig` |
| 5 | Wildcard CORS | **Fixed** — reads `CORS_ALLOWED_ORIGINS` env var in `cors.go:12-28` |
| 6 | UserQueries always commits | **Fixed** — documented as safe; PG auto-rolls-back aborted txns |
| 7 | Pagination broken for merged sources | **Fixed** — merges then applies offset/limit in `logs.go:192-199` |
| 8 | Missing RLS on 5 tables | **Fixed** — migration `021_rls_missing_tables.up.sql` |
| 9 | Poller not restarted on update | **Fixed** — `UpdateConnection` stops/restarts in `connections.go:339-342` |
| 10 | Poller error silently swallowed | **Fixed** |
| 11 | Rate-limit sleep ineffective | **Fixed** |
| 12 | Timer leak in ConnectionTestModal | **Fixed** — cleanup in `onBeforeUnmount` |

---

## Current Issues

### 1. Dead code: unreachable handlers and frontend modules

**Severity:** Medium
**Effort:** Small

Four backend handlers exist but are not registered in `router.go`:

| Handler | File | Lines |
|---------|------|-------|
| `GetAgentConfig` | `handlers/agent.go:18` | 12 |
| `UpdateAgentConfig` | `handlers/agent.go:38` | 34 |
| `RunAgent` | `handlers/agent.go:81` | 27 |
| `GetDashboardStats` | `handlers/stats.go:10` | 22 |

These were superseded by per-app equivalents (`GetAppAgentConfig`, `GetAppDashboardStats`) but never removed.

Three frontend modules reference these dead endpoints:

| Module | File | Issue |
|--------|------|-------|
| `api/agent.ts` | `frontend/src/api/agent.ts` | Calls `/agent/config` — endpoint doesn't exist |
| `stores/agent.ts` | `frontend/src/stores/agent.ts` | Uses stale `AgentConfig` type; never imported by any page |
| `api/stats.ts` | `frontend/src/api/stats.ts` | Calls `/stats` — endpoint doesn't exist |

The `AgentConfig` type in `types/agent.ts` defines mode as `'continuous' | 'scheduled' | 'off'`, while the active `AppAgentConfig` in `types/organization.ts` correctly uses `'continuous' | 'periodic' | 'off'`. This type is only used by the dead agent store, so it's not a runtime bug — but it will confuse anyone reading the types.

**Fix:** Delete `handlers/agent.go`, `handlers/stats.go`, `frontend/src/api/agent.ts`, `frontend/src/stores/agent.ts`, `frontend/src/api/stats.ts`, and the `AgentConfig` interface from `types/agent.ts`. Keep the WebSocket message types in `types/agent.ts` as they are actively used.

---

### 2. Duplicated poller/listener initialization logic

**Severity:** Medium
**Effort:** Small

Poller startup logic is duplicated between two locations:

- **`backend/cmd/heimdall/main.go:177-238`** — `resumePollers()` initializes connectors for all 5 poll-based types on server boot
- **`backend/internal/api/handlers/connections.go:484-523`** — `startPoller()` initializes connectors when a new connection is created

Both contain the same switch/table over `supabase`, `flyio`, `vercel`, `railway`, `mongodb` with identical `New*` → `poller.Start` patterns. Adding a new connector type requires changes in both places.

**Fix:** Extract a shared factory function into `internal/connectors/`:

```go
// connectors/factory.go
func StartPoller(poller *Poller, connType string, config json.RawMessage, connID, userID uuid.UUID) error {
    // single switch over connector types
}
```

Both `resumePollers` and `startPoller` call this function.

---

### 3. Stub implementations (dead code)

**Severity:** Low
**Effort:** Small

Three packages contain only TODO stubs with no callers:

| File | Lines | Content |
|------|-------|---------|
| `backend/internal/reports/generator.go` | 16 | `Generate()` returns `nil, nil` |
| `backend/internal/memory/client.go` | 36 | `RecordEvent`, `QueryMemories`, `GetLessons` are all no-ops |
| `backend/internal/connectors/logs/webhook.go` | 30 | `Stream()` does nothing |

These were placeholders for future features. They have no callers and provide no functionality.

**Fix:** Either remove entirely, or keep only if active development is planned (in which case, add a `// TODO(feature-name)` comment referencing the tracking issue).

---

### 4. Silenced `json.Marshal` error in syslog TLS config injection

**Severity:** Low
**Effort:** Small
**File:** `backend/internal/api/handlers/connections.go:189`

```go
config, _ = json.Marshal(cfgMap)
```

During syslog connection creation, TLS cert/key are injected into the config map and re-marshalled. The marshal error is silenced. While `json.Marshal` on a `map[string]interface{}` with string values won't realistically fail, this deviates from the pattern of checking errors elsewhere in the codebase.

**Fix:** Handle the error:

```go
config, err = json.Marshal(cfgMap)
if err != nil {
    jsonError(w, "failed to marshal config", http.StatusInternalServerError)
    return
}
```

---

## Files Over 500 Lines

### `backend/internal/api/handlers/connections.go` — 559 lines

Contains 6 HTTP handlers, 3 validators, and the `startPoller` helper. The file is cohesive (all connection-related) but slightly over threshold.

**Recommendation:** Extract `startPoller` into `internal/connectors/factory.go` (see issue #2 above). This brings the file under 500 lines and eliminates the duplication simultaneously.

### `frontend/src/pages/NotificationsPage.vue` — 541 lines

Handles notification preferences (edit/view), channel CRUD (add/edit/delete/test), and notification history in a single component.

**Recommendation:** Extract into sub-components:

- `NotificationPreferences.vue` — severity threshold toggle and edit mode
- `NotificationChannels.vue` — channel list, add/edit/delete/test
- `NotificationHistory.vue` — history list with filtering
- `NotificationsPage.vue` — orchestrator (imports the three above)

---

## What's Working Well

- **Security posture** is solid: RLS on all user-scoped tables, SQL injection mitigated, XSS escaped, CORS locked to known origins, auth middleware on all protected routes
- **Error handling** is consistent: `jsonError`/`jsonServerError` helpers, agent tool errors surfaced as `isError` results
- **Resource cleanup** is thorough: deferred closes, `onBeforeUnmount` cleanup in Vue components, proper context cancellation
- **Type safety** in the Go layer via sqlc-generated code
- **Test coverage** on critical paths: agent monitor, Supabase connector, connection handlers, frontend stores
- **Clean separation**: connectors implement interfaces, handlers delegate to queries, stores abstract API calls

---

## Priority Summary

| # | Issue | Severity | Effort | Action |
|---|-------|----------|--------|--------|
| 1 | Dead code: old handlers + frontend modules | Medium | Small | Delete 5 files |
| 2 | Duplicated poller initialization | Medium | Small | Extract factory function |
| 3 | Stub implementations | Low | Small | Delete or annotate |
| 4 | Silenced json.Marshal error | Low | Small | Add error check |
| — | `connections.go` over 500 lines | Low | Small | Resolved by #2 |
| — | `NotificationsPage.vue` over 500 lines | Low | Medium | Extract 3 sub-components |
