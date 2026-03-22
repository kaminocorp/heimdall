# Code Assessment — Heimdall Repository

**Date:** 2026-03-22
**Scope:** Full repository — backend (Go), frontend (Vue 3 + TypeScript), infrastructure (migrations, Docker, config)
**Assessment criteria:** Functionality, accuracy, maintainability, clean code. Files >500 lines flagged for refactoring.

---

## Executive Summary

The codebase is well-structured with clean separation of concerns, consistent patterns, and good test coverage for the new Supabase connector work. No Go files exceed 500 lines. One frontend file (541 lines) needs decomposition. However, several **functional bugs** and **security gaps** need attention before production deployment. The most critical are SQL injection in the Supabase connector, unescaped HTML in email notifications, and the `search_logs` agent tool silently ignoring its `query` parameter.

**Overall score: 7.5/10** — solid architecture, but the issues below need fixing to reach 8.5+.

---

## Critical Issues (Must Fix)

### 1. SQL injection in Supabase connector

**File:** `backend/internal/connectors/logs/supabase.go:126-129`

```go
sql := fmt.Sprintf(
    "SELECT timestamp, event_message, metadata FROM %s WHERE ...",
    table, cursorMicro, supabasePollLimit,
)
```

The `table` variable comes from user-provided config JSONB (`poll_tables`). It's interpolated directly into the SQL string sent to the Supabase Management API with no validation. A crafted table name like `postgres_logs; DROP TABLE x--` injects arbitrary SQL.

**Fix:** Validate `table` against an allowlist of known Supabase log tables:

```go
var validTables = map[string]bool{
    "postgres_logs": true, "auth_logs": true, "edge_logs": true,
    "function_logs": true, "storage_logs": true, "realtime_logs": true,
}

func (s *Supabase) pollTable(ctx context.Context, queries *db.Queries, table string) error {
    if !validTables[table] {
        return fmt.Errorf("supabase: invalid table name: %q", table)
    }
    // ...
}
```

Also validate in `NewSupabase()` at parse time so invalid tables are rejected early.

---

### 2. XSS in email notification HTML

**File:** `backend/internal/notifications/format.go:43-52`

```go
func FormatEmailHTML(p Payload) string {
    return fmt.Sprintf(`...
<p><strong>Summary:</strong> %s</p>
<div ...>%s</div>
...`, ..., p.Summary, p.Assessment, ...)
}
```

`p.Summary` and `p.Assessment` contain LLM-generated text interpolated directly into HTML without escaping. `p.AppName` is user-controlled and also unescaped. If Claude's response includes angle brackets or if a user names their app `<script>alert(1)</script>`, the HTML email will render it.

**Fix:** Use `html.EscapeString()` on all interpolated values:

```go
import "html"

// In FormatEmailHTML:
html.EscapeString(p.Summary)
html.EscapeString(p.Assessment)
html.EscapeString(p.AppName)
```

---

### 3. `search_logs` tool ignores the `query` parameter

**File:** `backend/internal/agent/tools_logs.go:14-86`

The tool schema declares a required `query` parameter, but `toolSearchLogs` never reads `input["query"]`. It only filters by `severity` and `limit`. The agent believes it's searching by keyword — it is not. This means every "search" in interactive chat is actually just "list recent logs", making the agent's investigation capability fundamentally broken.

**Fix:** Read the `query` parameter and use it for full-text filtering. Either:
- Add a `SearchLogsByKeyword` sqlc query with `WHERE payload::text ILIKE '%' || $1 || '%'`
- Or filter in Go after fetching (less efficient but simpler)

---

### 4. Interactive chat uses wrong config source

**File:** `backend/internal/agent/loop.go:34`

```go
cfg, err := a.queries.GetAgentConfig(ctx)
```

`RunConversation` loads the global `agent_config` singleton. The monitoring loop correctly uses `GetAppAgentConfig`. This means per-app model selection and system prompt overrides configured via the UI are ignored during interactive chat.

**Fix:** Accept `appID` in `RunConversation` and use `GetAppAgentConfig(ctx, appID)` when `appID != uuid.Nil`. Fall back to the global config only when no app-specific config exists.

---

## High Priority Issues

### 5. CORS wildcard allows cross-origin authenticated requests

**File:** `backend/internal/api/middleware/cors.go:7`

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

A wildcard CORS policy combined with `Authorization` header allowance means any website can make authenticated API calls on behalf of a logged-in Heimdall user. For a production monitoring tool with database access, this is a real attack surface.

**Fix:** Restrict to known frontend origins (e.g., the Vercel deployment URL and `localhost:5173` for dev). Read the allowed origin from an environment variable.

---

### 6. UserQueries always commits, never rolls back on error

**File:** `backend/internal/api/handlers/userqueries.go:26`

```go
return s.Queries.WithTx(tx), func() { tx.Commit(ctx) }, nil
```

The `done()` function always commits. Handlers use `defer done()`, so even if they return early due to an error, the transaction commits. For read-only handlers this is harmless, but write handlers (`CreateConnection`, `UpdateConnection`, `DeleteConnection`) could commit partial state.

**Fix:** Return a `done` that rolls back, and require explicit commit:

```go
return s.Queries.WithTx(tx), func() { tx.Rollback(ctx) }, func() error { return tx.Commit(ctx) }, nil
```

Or simpler: change `done()` to rollback (rollback after commit is a no-op in pgx):

```go
return s.Queries.WithTx(tx), func() { tx.Rollback(ctx) }, nil
```

Then have handlers explicitly commit when they succeed. Alternatively, since the current handlers always return before the next write on error, document this as an accepted design constraint.

---

### 7. Log pagination broken for combined sources

**File:** `backend/internal/api/handlers/logs.go:82-166`

When `source == "all"`, the handler fetches `limit` rows from raw logs AND `limit` rows from agent logs, merges them, sorts, and truncates. But `offset` is applied independently to each source, so paginating produces inconsistent and duplicate results. Page 2 (offset=50) fetches raw logs 50-100 and agent logs 50-100, but the user expects the next 50 from the merged sorted set.

**Fix:** Either:
- Use a SQL `UNION ALL` query with shared `ORDER BY` / `LIMIT` / `OFFSET`
- Or fetch from both sources without offset, merge in memory, then apply offset/limit to the merged result (acceptable for small datasets)

---

### 8. Missing RLS on 5 tables

**Files:** Migrations `014`, `016`, `017`, `018`, `019`

The following tables lack `ENABLE ROW LEVEL SECURITY` and RLS policies, inconsistent with the pattern established for all other user-scoped tables:

| Table | Migration |
|-------|-----------|
| `organizations` | 014 |
| `applications` | 014 |
| `monitoring_state` | 016 |
| `notification_channels` | 017 |
| `notification_preferences` | 018 |
| `notification_log` | 019 |

If the Supabase `authenticated` role accesses these tables via PostgREST, any user can read/modify any other user's data.

**Fix:** Add a new migration enabling RLS and creating appropriate policies for each table.

---

## Medium Priority Issues

### 9. Supabase poller not restarted on connection update

**File:** `backend/internal/api/handlers/connections.go` — `UpdateConnection` handler

When a Supabase connection's config is updated (e.g., changing `poll_tables` or `poll_interval_secs`), the running poller continues with the old config. Only `DeleteConnection` calls `s.Poller.Stop()`.

**Fix:** In `UpdateConnection`, after a successful DB update, stop and restart the poller if the connection type is `supabase`:

```go
if req.Type == "supabase" {
    s.Poller.Stop(connID)
    sb, err := logs.NewSupabase(config, connID, userID)
    if err == nil {
        cfg := sb.ParsedConfig()
        s.Poller.Start(sb, connID, time.Duration(cfg.PollIntervalSecs)*time.Second)
    }
}
```

---

### 10. Supabase poller creation error silently swallowed

**File:** `backend/internal/api/handlers/connections.go:178-185`

```go
if req.Type == "supabase" {
    sb, err := logs.NewSupabase(config, conn.ID, userID)
    if err == nil {
        // starts poller
    }
    // err silently dropped
}
```

If `NewSupabase` fails, the connection is created with status "inactive" and the user gets a 201 response with no indication that polling failed to start.

**Fix:** Log the error at minimum. Optionally return a warning in the response body.

---

### 11. Rate-limit sleep is ineffective

**File:** `backend/internal/connectors/logs/supabase.go:213-233`

When `X-RateLimit-Remaining` is 0, the code sleeps until the reset time, then falls through to read the response body of the *current* request (which may be a successful 200). The sleep doesn't retry — it just delays the return. The poller will retry on the next tick regardless.

**Fix:** Remove the inline sleep. Instead, track the rate-limit reset time and skip polling until it passes:

```go
// In Supabase struct:
rateLimitUntil time.Time

// In pollTable:
if time.Now().Before(s.rateLimitUntil) {
    return nil // skip this tick
}
```

---

### 12. Missing indexes on FK columns

| Table | Column | Issue |
|-------|--------|-------|
| `notification_log` | `channel_id` | FK with `ON DELETE CASCADE` — cascade requires full table scan without index |
| `agent_log` | `conversation_id` | FK with `ON DELETE SET NULL` — same issue |

**Fix:** Add a migration with:

```sql
CREATE INDEX idx_notification_log_channel_id ON notification_log(channel_id);
CREATE INDEX idx_agent_log_conversation_id ON agent_log(conversation_id);
```

---

### 13. `ConnectionTestModal` timer leak on unmount

**File:** `frontend/src/components/connections/ConnectionTestModal.vue:19`

The `setInterval` timer for the elapsed counter is started on mount but has no `onUnmounted` cleanup. If the modal is closed while the test is in-flight, the interval continues updating state on a destroyed component.

**Fix:**

```typescript
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
```

---

### 14. Dashboard sequential API calls

**File:** `frontend/src/pages/DashboardPage.vue:29-33`

Five `await` calls run sequentially when they could run in parallel:

```typescript
const [agentResult, connsResult, ...] = await Promise.allSettled([
  getAppAgentConfig(appId),
  listConnectionsByApp(appId),
  logsStore.fetchLogs(),
  getAppStats(appId),
  getMonitoringStatus(appId),
])
```

This would cut dashboard load time significantly.

---

## Low Priority Issues

### 15. `CopyableField` silently fails on copy

**File:** `frontend/src/components/common/CopyableField.vue:12-14`

The `catch` block swallows the error with no user feedback. Should show a toast or fallback to `document.execCommand('copy')`.

---

### 16. `StepPostgresConfig` doesn't validate password

**File:** `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue:34`

The validation check includes host, database, and username but not password. The wizard allows proceeding without a password, deferring the error to the connection test step.

---

### 17. Dead code: Unused handlers and store

**Backend:** `GetAgentConfig`, `UpdateAgentConfig`, `RunAgent` (agent.go), `GetDashboardStats` (stats.go) — defined but not registered in router.go.

**Frontend:** `agent` store (`stores/agent.ts`) — uses stale `AgentConfig` type with `'scheduled'` mode while the app uses `'periodic'`. Pages bypass this store entirely. `api/stats.ts` (`getDashboardStats`) is also unused — pages call `getAppStats` instead.

**Fix:** Remove dead code to reduce confusion.

---

### 18. Duplicate unique index on `organizations.slug`

**File:** `backend/migrations/014_organizations_applications.up.sql:8,12`

The column is defined with `UNIQUE` (implicit index) AND has an explicit `CREATE UNIQUE INDEX`. Two indexes are maintained for the same constraint.

**Fix:** Remove the explicit `CREATE UNIQUE INDEX` in a new migration, or leave as-is (harmless but wasteful).

---

## Refactoring Required

### `NotificationsPage.vue` — 541 lines (exceeds 500-line threshold)

This page combines preferences editing, channel CRUD, channel form state, notification history display, and multiple utility functions in a single SFC.

**Recommendation:** Decompose into:
- `NotificationPreferences.vue` — severity threshold, toggle
- `NotificationChannelList.vue` — channel cards, add/edit/delete
- `NotificationChannelForm.vue` — channel form (type, webhook URL, etc.)
- `NotificationHistory.vue` — history list with filtering

The page component becomes an orchestrator that imports these sub-components.

---

## Verification Results

| Check | Result |
|-------|--------|
| `go build ./...` | Clean |
| `go vet ./...` | Clean |
| `go test ./internal/connectors/...` | 4 packages, all pass |
| `npm run build` | Clean |
| `npm run test` | 5 files, 25 tests, all pass |
| Backend files >500 lines | None |
| Frontend files >500 lines | 1 (`NotificationsPage.vue` — 541 lines) |

---

## Priority Summary

| # | Issue | Severity | Effort |
|---|-------|----------|--------|
| 1 | SQL injection in Supabase table name | Critical | Small |
| 2 | XSS in email HTML | Critical | Small |
| 3 | `search_logs` ignores query param | Critical | Medium |
| 4 | Chat uses wrong agent config | High | Small |
| 5 | Wildcard CORS | High | Small |
| 6 | UserQueries always commits | High | Small |
| 7 | Pagination broken for merged sources | High | Medium |
| 8 | Missing RLS on 5 tables | High | Medium |
| 9 | Poller not restarted on update | Medium | Small |
| 10 | Poller error silently swallowed | Medium | Small |
| 11 | Rate-limit sleep ineffective | Medium | Small |
| 12 | Missing FK indexes | Medium | Small |
| 13 | Timer leak in test modal | Medium | Small |
| 14 | Dashboard sequential fetches | Medium | Small |
| 15 | CopyableField silent failure | Low | Small |
| 16 | Postgres step missing password validation | Low | Small |
| 17 | Dead code cleanup | Low | Small |
| 18 | Duplicate slug index | Low | Small |
| — | Refactor NotificationsPage.vue | Low | Medium |
