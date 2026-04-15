# Defense-in-Depth Hardening — v0.42.20 Assessment Findings

Post-assessment findings from a comprehensive review of the v0.42.0–v0.42.20 codebase. None of these are functionally blocking — the app works correctly for normal usage. These are defense-in-depth improvements for production resilience at scale.

**Current rating:** 7.5/10 — targeting 8.5/10 after these fixes.

---

## C1. Webhook token index misses OTLP type

**Severity:** Critical (performance)
**File:** `backend/migrations/008_add_webhook_token_index.up.sql`
**Query:** `backend/internal/db/queries/log_buffer.sql:38`

The partial index `idx_connections_webhook_token` has predicate `WHERE type = 'webhook_logs'`, but `GetConnectionByWebhookToken` filters on `type IN ('webhook_logs', 'otlp')`. OTLP connections fall through to a sequential scan on `connections`. Not a correctness bug — every OTLP ingest request just gets slower as the table grows.

### Fix

New migration:

```sql
DROP INDEX IF EXISTS idx_connections_webhook_token;
CREATE INDEX idx_connections_webhook_token
    ON connections ((config->>'webhook_token'))
    WHERE type IN ('webhook_logs', 'otlp');
```

---

## H1. `SELECT set_config(...)` bypasses read-only SQL validation

**Severity:** High (defense-in-depth)
**File:** `backend/internal/agent/tools_db.go:132`

The SQL validator allows any statement starting with `SELECT`. The PostgreSQL function `set_config('default_transaction_read_only', 'off', false)` is callable via `SELECT set_config(...)` and would pass validation. This could disable the write guard within the session.

**Mitigations already in place:**
- Each `toolQueryDatabase` call opens a fresh `pgx.Conn` (line 53–61), so `set_config` with `is_local=false` only affects the single connection, which is closed immediately after.
- The Postgres connector sets `default_transaction_read_only=on` at the connection URL level.
- The LLM would need to craft this specific escape intentionally.

### Fix

Add a function-call blocklist to `isReadOnlySQL`. After the existing write-keyword check, scan for dangerous function names:

```go
dangerousFuncs := []string{"SET_CONFIG", "PG_READ_FILE", "PG_WRITE_FILE", "LO_IMPORT", "LO_EXPORT"}
for _, fn := range dangerousFuncs {
    if strings.Contains(upper, fn) {
        return false
    }
}
```

Alternatively, wrap all user queries in `SET TRANSACTION READ ONLY` at the connector level for a more robust guarantee.

---

## H2. `monitorTick` / `schedulerTick` early return skips `wg.Wait()`

**Severity:** High (correctness)
**Files:**
- `backend/internal/agent/monitor.go:70-76`
- `backend/internal/agent/scheduler.go:114-120`

When the context is cancelled while waiting for the semaphore, the code does `wg.Done(); return`. This returns from the tick function immediately, but `wg.Wait()` at the end of the function is never reached. Goroutines launched for earlier apps in the same loop iteration are orphaned — they complete eventually (their context is cancelled) but have no join point. If a goroutine is mid-write (e.g. `UpsertMonitoringState`), the process could exit before that write completes.

### Fix

Use `defer wg.Wait()` at the top of both functions:

```go
func (a *Agent) monitorTick(ctx context.Context, sem chan struct{}) {
    // ...
    var wg sync.WaitGroup
    defer wg.Wait() // always join, even on early return
    // ...
}
```

Same change in `schedulerTick`.

---

## H3. TOCTOU gap in notification channel + schedule Update/Delete

**Severity:** High (defense-in-depth)
**Files:**
- `backend/internal/api/handlers/notifications.go:223-228` (Update), `:291-300` (Delete)
- `backend/internal/api/handlers/investigation_schedules.go:231-241` (Update), `:307-315` (Delete)

The ownership check (`GetNotificationChannel` / `GetSchedule`) runs via `s.Queries` (outside the transaction), and the SQL queries (`UpdateNotificationChannel`, `DeleteNotificationChannel`, `DeleteSchedule`) filter only by `id` with no `AND app_id` guard. Two compounding problems: the check and write aren't in the same transactional snapshot, and even if they were, the SQL itself doesn't enforce ownership.

### Fix

Two-part fix:

**Part A — Add `app_id` to the SQL queries:**

```sql
-- notification_channels.sql
UPDATE notification_channels SET ... WHERE id = $1 AND app_id = $7;
DELETE FROM notification_channels WHERE id = $1 AND app_id = $2;

-- investigation_schedules.sql
UPDATE investigation_schedules SET ... WHERE id = $1 AND app_id = $8;
DELETE FROM investigation_schedules WHERE id = $1 AND app_id = $2;
```

Then `make sqlc-generate`.

**Part B — Move ownership checks inside the transaction:**

Replace `s.Queries.GetNotificationChannel(...)` with `queries.GetNotificationChannel(...)` (using the transactional queries object returned by `UserQueries`). Same for `GetSchedule`.

---

## M1. Postgres connector uses `url.PathEscape` for connection string userinfo

**Severity:** Medium (correctness)
**File:** `backend/internal/connectors/database/postgres.go:71-75`

The connection string is a URL (`postgres://user:pass@host:port/db?params`). User and password are escaped with `url.PathEscape`, which does not escape `@` or `:` — characters that have structural meaning in the userinfo portion of a URL. A password containing `@` would be misinterpreted as the host delimiter.

### Fix

Use Go's `url.URL` struct builder, which handles encoding correctly:

```go
u := &url.URL{
    Scheme:   "postgres",
    User:     url.UserPassword(cfg.User, cfg.Password),
    Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
    Path:     cfg.Database,
    RawQuery: fmt.Sprintf("sslmode=%s&default_transaction_read_only=on", cfg.SSLMode),
}
connStr := u.String()
```

Also add port range validation:

```go
if cfg.Port < 1 || cfg.Port > 65535 {
    return nil, fmt.Errorf("postgres: port must be between 1 and 65535")
}
```

---

## M2. `EmitLog` called before `commit()` in `CreateApplication`

**Severity:** Medium (correctness)
**File:** `backend/internal/api/handlers/applications.go:108-123`

`EmitLog` fires before `commit()`. If the commit fails, a phantom audit log entry is written for an application that was rolled back. This contradicts the project's own pattern — `DeleteApplication` (line 199–214) correctly calls `EmitLog` *after* the commit, with a comment: "emitted after the delete succeeds so we never record a deletion that didn't happen."

### Fix

Move the `EmitLog` call to after the `commit()` call:

```go
if err := commit(); err != nil {
    jsonServerError(w, "failed to create application", err)
    return
}

s.Agent.EmitLog(ctx, created.ID, userID, "application_created",
    fmt.Sprintf("Application '%s' created", created.Name), nil)
```

---

## M3. `UpdateConnection` reads existing connection outside transaction

**Severity:** Medium (correctness)
**File:** `backend/internal/api/handlers/connections.go:269-276`

The handler fetches the existing connection via `s.Queries.GetConnectionByUser(...)` (outside the transaction) to read defaults for omitted fields (status, webhook_token). The actual `UpdateConnection` runs inside a `UserQueries` transaction. Between the read and write, another request could modify the connection. At worst, a status or webhook token would be stale by one request.

### Fix

Move `GetConnectionByUser` inside the `UserQueries` transaction:

```go
queries, commit, done, err := s.UserQueries(ctx, userID)
if err != nil { ... }
defer done()

existing, err := queries.GetConnectionByUser(ctx, db.GetConnectionByUserParams{...})
if err != nil { ... }

// Apply defaults from existing, then update
```

---

## M5. Notification dispatch goroutine has no timeout

**Severity:** Medium (resource leak)
**File:** `backend/internal/agent/monitor.go:211-213`

The notification dispatch uses `context.WithoutCancel(ctx)` to survive the `monitorApp` return, which is correct. However, this context has no timeout. If the notification target (webhook, email service) hangs indefinitely, this goroutine leaks permanently.

### Fix

```go
notifyCtx, notifyCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
defer notifyCancel()
go a.Dispatcher.Dispatch(notifyCtx, ...)
```

Note: since this is a fire-and-forget goroutine, the `defer notifyCancel()` should be inside the goroutine, not outside:

```go
go func() {
    notifyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
    defer cancel()
    a.Dispatcher.Dispatch(notifyCtx, ...)
}()
```

---

## M8. Nullable JSONB columns map to `[]byte` instead of `json.RawMessage`

**Severity:** Medium (data serialization)
**File:** `backend/sqlc.yaml`, `backend/internal/db/models.go`

Nullable JSONB columns (`agent_log.detail`, `investigations.findings`, `investigations.tool_trace`) map to `[]byte` in the generated Go code. When these are marshaled to JSON for API responses, they produce base64-encoded strings rather than inline JSON objects.

### Fix

Add a nullable JSONB override to `sqlc.yaml`:

```yaml
overrides:
  - db_type: "jsonb"
    nullable: true
    go_type: "encoding/json.RawMessage"
```

Then `make sqlc-generate`. Verify the generated code compiles and test that API responses return inline JSON for these fields.

---

## M9. Duplicate unique index on `organizations.slug`

**Severity:** Medium (performance)
**File:** `backend/migrations/014_organizations_applications.up.sql:8,12`

Line 8 defines `slug TEXT NOT NULL UNIQUE` (which implicitly creates a unique index). Line 12 explicitly creates `CREATE UNIQUE INDEX idx_organizations_slug ON organizations(slug)`. PostgreSQL maintains both — double the index storage, double the write overhead on every insert/update.

### Fix

New migration:

```sql
DROP INDEX IF EXISTS idx_organizations_slug;
```

The `UNIQUE` constraint on the column definition is sufficient.

---

## Implementation Notes

- Fixes C1, M8, and M9 require new migrations — batch them into a single migration file (e.g. `030_assessment_fixes`).
- Fix H3 Part A requires SQL query changes followed by `make sqlc-generate`.
- All other fixes are Go/Vue source changes with no schema impact.
- Run `go build ./...`, `go vet ./...`, `go test ./...`, `vue-tsc --noEmit`, and `vite build` after applying.
