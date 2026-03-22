# Phase 8 — Production Hardening (Code Assessment)

Comprehensive production-readiness review identified 19 issues across backend and frontend. All fixed in this phase.

## Critical Fixes

### 1. UserQueries transaction commit uses cancellable context

**File:** `backend/internal/api/handlers/userqueries.go`

The `done()` closure captured the HTTP request `ctx`. If the client disconnected before `defer done()` ran, `tx.Commit(ctx)` failed with `context canceled`, silently rolling back successful writes. This affected every write endpoint using `UserQueries` — connection CRUD, conversations, etc.

**Fix:** Use `context.WithoutCancel(ctx)` (Go 1.21+) for the commit call. This preserves context values (user ID) but decouples from the request lifecycle so the commit always completes.

### 2. Partial-create on Supabase poller failure

**File:** `backend/internal/api/handlers/connections.go`

When creating a Supabase connection, the DB insert was committed by `defer done()`, but if `NewSupabase()` failed afterwards, the handler returned HTTP 400. The client saw an error, but the connection persisted in the database — retrying would create duplicates.

**Fix:** Validate the Supabase config eagerly (with `NewSupabase(config, uuid.Nil, uuid.Nil)`) *before* the DB insert. If config is invalid, the handler returns 400 before anything is written. The post-insert `NewSupabase` call is now best-effort (logs on failure, doesn't return error to client).

### 3. `os.Exit(1)` in server goroutine skips cleanup

**File:** `backend/cmd/heimdall/main.go`

The `ListenAndServe` goroutine called `os.Exit(1)` on error, which terminates the process immediately — skipping all deferred cleanup (pool.Close, classifier.Close, poller.StopAll).

**Fix:** Replaced with an error channel. The main goroutine now `select`s on both the signal channel and the error channel, ensuring the shutdown sequence always runs.

## High-Severity Fixes

### 4. Poller panics on zero/negative interval

**File:** `backend/internal/connectors/poller.go`

`time.NewTicker(interval)` panics if `interval <= 0`. No validation existed.

**Fix:** Added `minPollInterval` constant (5s). `Start()` clamps any interval below this minimum.

### 5. No graceful drain on shutdown

**File:** `backend/internal/connectors/poller.go`

`StopAll()` cancelled contexts but returned immediately — in-flight polls could be interrupted mid-transaction.

**Fix:** Added `sync.WaitGroup` to the Poller struct. Each goroutine increments on start, decrements on exit. `StopAll()` cancels all contexts then calls `wg.Wait()` (outside the lock to prevent deadlock).

### 6. `Poll()` always returns nil

**File:** `backend/internal/connectors/logs/supabase.go`

The `PollConnector` interface defines an error return, but `Poll()` always returned `nil` — making the error log in `Poller.run()` dead code.

**Fix:** Track `firstErr` across table polls. Return the first error encountered (all tables are still attempted).

### 7. Pagination total count ignores active filters

**File:** `backend/internal/api/handlers/logs.go`, `backend/internal/db/queries/log_buffer.sql`

`CountLogsByUser` counted all logs regardless of severity or connection_id filters, producing incorrect pagination totals.

**Fix:** Added `CountLogsByUserAndSeverity` and `CountLogsByUserAndConnection` SQL queries. The handler now uses the filtered count that matches the active query path. Regenerated sqlc.

### 8. `source=all` pagination is O(offset)

**File:** `backend/internal/api/handlers/logs.go`

`fetchLimit = offset + limit` fetched up to `offset + limit` rows from each source. With large offsets, this consumed unbounded memory — trivially exploitable for resource exhaustion.

**Fix:** Capped offset at 10,000 (`maxOffset`). This is a reasonable limit for log browsing UIs. Deep pagination should use cursor-based approaches.

### 9. HTTP idle connections accumulate

**File:** `backend/internal/connectors/logs/supabase.go`

`Close()` was a no-op. Each destroyed/recreated Supabase connector leaked idle TCP connections from the `http.Client` transport.

**Fix:** `Close()` now calls `s.httpClient.CloseIdleConnections()`.

### 10. No input validation on type/direction/status

**File:** `backend/internal/api/handlers/connections.go`

Connection `type`, `direction`, and `status` fields accepted arbitrary strings with no allow-list validation.

**Fix:** Added `isValidConnectionType`, `isValidDirection`, and `isValidConnectionStatus` helpers with explicit allow-lists. Both `CreateConnection` and `UpdateConnection` now validate before proceeding.

## Medium-Severity Fixes

### 11. Orphaned connection on wizard abandon

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue`

If the user reached the test step (connection created server-side), then closed the wizard, the connection persisted in the database with no cleanup.

**Fix:** `handleClose()` now deletes the created connection (best-effort) before emitting `close`. Also added Escape key handler via document event listener with proper cleanup on unmount.

### 12. One-way prop sync in wizard step components

**Files:** `StepName.vue`, `StepSupabaseAuth.vue`, `StepSupabaseTables.vue`, `StepPostgresConfig.vue`

All step components copied props into local refs on mount but never reacted to prop changes. Re-entering a flow after going back to platform selection showed stale data from the previous attempt.

**Fix:** Added `watch(() => props.modelValue.config, ...)` (or `.name` for StepName) watchers that sync inward when the parent resets state. Guards prevent infinite loops (only update if value actually changed).

### 13. Migration 021 is not idempotent

**File:** `backend/migrations/021_rls_missing_tables.up.sql`

`CREATE POLICY` fails if the policy already exists. A partial failure mid-migration left a dirty state that couldn't be cleanly retried.

**Fix:** Added `DROP POLICY IF EXISTS` before each `CREATE POLICY`. Changed `CREATE INDEX` to `CREATE INDEX IF NOT EXISTS`.

### 14. CopyableField setTimeout leak

**File:** `frontend/src/components/common/CopyableField.vue`

The 2-second `setTimeout` was never cleared on unmount. If the component unmounted within 2 seconds of copying, the callback fired on a detached component.

**Fix:** Store the timer handle, clear previous timer before setting a new one, and clear on `onBeforeUnmount`.

### 15. notification_log RLS policy uses unnecessary 3-table join

**File:** `backend/migrations/021_rls_missing_tables.up.sql`

The `notification_log` table has a direct `app_id` column, but the RLS policy routed through `channel_id -> notification_channels -> applications`. This was slower and fragile if channels were deleted (orphaned log rows would become invisible).

**Fix:** Changed the policy to use the direct `app_id` column with a 2-table join (matching the pattern used for `notification_channels` and `notification_preferences`).

## Low-Severity Fixes

### 16. Escape key handler on modals

**Files:** `ConnectionWizard.vue`, `ConnectionTestModal.vue`

Neither modal had `@keydown.escape` handling. Users couldn't dismiss modals with Escape.

**Fix:** Added document-level `keydown` listeners for Escape, with proper cleanup on unmount. Test modal only allows Escape when not in the `testing` phase (preventing accidental dismissal during an in-flight test).

### 17. Delete confirmation on ConnectionCard

**File:** `frontend/src/components/connections/ConnectionCard.vue`

The delete button immediately emitted `delete` with no confirmation. A single click permanently deleted a connection.

**Fix:** Two-click confirmation pattern. First click changes button text to "Confirm?" with red styling. Second click within 3 seconds performs the delete. Auto-resets after 3 seconds if user doesn't confirm.

### 18. Action button visibility on touch/keyboard devices

**File:** `frontend/src/components/connections/ConnectionCard.vue`

Action buttons (Edit, Delete, Ping, Repos) were hidden with `sm:opacity-0 sm:group-hover:opacity-100`, making them invisible on touch screens and to keyboard users.

**Fix:** Added `focus-within:opacity-100` so buttons become visible when any button in the group receives focus (tab navigation). Also added `focus:text-accent focus:outline-none` to individual buttons for keyboard focus styling.

### 19. Dead code in ConnectionTestModal

**File:** `frontend/src/components/connections/ConnectionTestModal.vue`

The `statusColor` computed property was defined but never referenced in the template.

**Fix:** Removed the computed and the now-unused `watch` import.

## Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/userqueries.go` | Use `context.WithoutCancel` for commit |
| 2 | `backend/internal/api/handlers/connections.go` | Eager Supabase validation, input allow-lists |
| 3 | `backend/cmd/heimdall/main.go` | Error channel replaces `os.Exit(1)` |
| 4 | `backend/internal/connectors/poller.go` | Interval clamp, WaitGroup drain |
| 5 | `backend/internal/connectors/logs/supabase.go` | Return first error, close idle connections |
| 6 | `backend/internal/api/handlers/logs.go` | Filtered count queries, offset cap |
| 7 | `backend/internal/db/queries/log_buffer.sql` | New filtered count queries |
| 8 | `backend/internal/db/log_buffer.sql.go` | Regenerated sqlc output |
| 9 | `backend/migrations/021_rls_missing_tables.up.sql` | Idempotent policies, optimized notification_log join |
| 10 | `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Orphan cleanup, Escape key |
| 11 | `frontend/src/components/connections/wizard/steps/StepName.vue` | Inward prop sync |
| 12 | `frontend/src/components/connections/wizard/steps/StepSupabaseAuth.vue` | Inward prop sync |
| 13 | `frontend/src/components/connections/wizard/steps/StepSupabaseTables.vue` | Inward prop sync |
| 14 | `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue` | Inward prop sync |
| 15 | `frontend/src/components/common/CopyableField.vue` | Timer cleanup on unmount |
| 16 | `frontend/src/components/connections/ConnectionTestModal.vue` | Escape key, remove dead code |
| 17 | `frontend/src/components/connections/ConnectionCard.vue` | Delete confirmation, focus-within visibility |

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./internal/connectors/...` — all pass
- `npx vue-tsc --noEmit` — no type errors
- `npx vitest run` — 25/25 tests pass
