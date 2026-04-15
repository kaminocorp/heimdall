# Code Assessment — v0.42.10

**Date:** 2026-04-15
**Scope:** Full codebase (backend + frontend + database layer)
**Focus:** Functionality, accuracy, maintainability, clean code. No nice-to-haves.

---

## Summary

| Severity | Backend | Frontend | Database | Total |
|----------|---------|----------|----------|-------|
| Critical | 2       | 1        | 1        | **4** |
| High     | 8       | 2        | 2        | **12**|
| Medium   | 15      | 7        | 3        | **25**|

**Critical issues:** SQL read-only validation bypass (CTE + multi-statement), provider error detection via fragile string match, ActivityPage showing wrong connections, nullable `app_id` columns violating multi-app scoping.

**Key new findings (v0.42.10 pass):** Leaf data tables still user-scoped (breaks multi-member orgs), no panic recovery in background goroutines, monitor cursor skips excess flagged logs, `organization` ref loses role type guarantee after onboarding, slug validation missing on update.

---

## Critical Issues

### C1. `isReadOnlySQL` allows write operations inside CTEs

**File:** `backend/internal/agent/tools_db.go:108-110`
**Impact:** An LLM-generated query could execute write operations via a CTE.

The validation only checks if the normalized SQL starts with `SELECT`, `EXPLAIN`, or `WITH`. A CTE like this passes the check:

```sql
WITH x AS (DELETE FROM users RETURNING *) SELECT * FROM x;
```

**Mitigation in place:** The Postgres connector sets `default_transaction_read_only=on` in the connection string (`database/postgres.go:48`), so PostgreSQL itself will reject the write. This makes the risk **medium in practice** but **critical in principle** — the `isReadOnlySQL` function claims to validate read-only queries but doesn't. If the connector changes or another database type is added, this becomes an actual exploit.

**Additional vector:** Multi-statement injection via `;` is also unguarded. A query like `SELECT 1; DROP TABLE users;` passes the check because it starts with `SELECT`. The Postgres `default_transaction_read_only=on` blocks writes, but the app-layer check is incomplete.

**Fix:** (1) Reject statements containing `;` after stripping string literals and comments. (2) For `WITH`-prefixed queries, scan the normalized string for write keywords (`DELETE`, `UPDATE`, `INSERT`, `DROP`, `ALTER`, `TRUNCATE`, `CREATE`) appearing outside of string literals.

---

### C2. Provider error detection via fragile string match

**Files:** `backend/internal/agent/monitor.go:179`, `backend/internal/agent/loop.go:274`

The monitoring loop returns a hardcoded string `"Monitoring assessment failed: provider error"` from `RunMonitoring` when the LLM provider fails. The caller in `monitorApp` detects this by checking `strings.Contains(assessment, "provider error")`.

```go
// loop.go:274 — the source
return "Monitoring assessment failed: provider error", "error"

// monitor.go:179 — the detector
if severity == "error" && strings.Contains(assessment, "provider error") {
    slog.Warn("monitor: LLM provider failed, cursor not advanced", "app_id", app.ID)
    return
}
```

If someone edits the return string without updating the detector (or vice versa), the cursor will advance past logs that were never assessed. Those logs are permanently skipped.

**Fix:** Use a sentinel error or a dedicated boolean return value from `RunMonitoring` indicating provider failure. Replace the string match entirely.

---

### C3. ActivityPage fetches connections from all apps

**File:** `frontend/src/pages/ActivityPage.vue:17`

```typescript
connectionsStore.fetchConnections()  // fetches ALL user connections
```

This calls `fetchConnections()` which hits `GET /connections` (all connections for the user across all apps). The ConnectionsPage correctly uses `fetchConnectionsByApp(appId)`. As a result, the Activity page's connection dropdown shows connections from other apps, and the `LogFeed` component receives cross-app connection data.

The `watch` on `appStore.currentAppId` (line 20) refetches logs but not connections, so switching apps leaves stale cross-app connections.

**Fix:** Change to `connectionsStore.fetchConnectionsByApp(appStore.currentAppId)` and add connection refetch to the app-switch watcher.

---

### C4. Nullable `app_id` on `log_buffer` and `agent_log` tables

**File:** `backend/migrations/025_add_app_id_to_logs.up.sql`

The `app_id` column was added as nullable, with a backfill that only covers rows with valid connections (log_buffer) or rows where `detail->>'app_id'` exists (agent_log). Rows that don't match either condition retain `NULL app_id` permanently.

Queries like `ListLogsSinceForApp` filter by `app_id = $1`, silently excluding these rows from the activity feed. This creates gaps in monitoring history.

**Fix:** Either add `NOT NULL` constraint after running a validation query (`SELECT COUNT(*) FROM log_buffer WHERE app_id IS NULL`), or backfill remaining NULLs to a default/orphan app.

---

## High Issues

### H1. Onboard handler missing `Provider` field and hardcodes model

**File:** `backend/internal/api/handlers/organizations.go:312-317`

The Onboard handler creates a default agent config without setting `Provider` and with a hardcoded model string:

```go
// Onboard (organizations.go:312) — MISSING Provider, hardcoded model
UpsertAppAgentConfigParams{
    AppID:                app.ID,
    Model:                "claude-sonnet-4-6",   // should be agent.DefaultModelID
    Mode:                 "off",
    ScheduleIntervalSecs: 60,
    // Provider: missing!
}
```

Compare with `CreateApplication` (applications.go:90-95) which correctly uses `agent.DefaultModelID` and sets `Provider: "anthropic"`. Apps created via onboarding will have an empty Provider, which may cause failures in scheduled investigations or model resolution.

**Fix:** Add `Provider: "anthropic"` and replace the hardcoded string with `agent.DefaultModelID`.

---

### H2. `json.Marshal` errors silently ignored in 4 connectors

**Files:**
- `backend/internal/connectors/logs/flyio.go:232`
- `backend/internal/connectors/logs/vercel.go:140`
- `backend/internal/connectors/logs/railway.go:155`
- `backend/internal/connectors/logs/mongodb.go:168`

All four connectors use `payload, _ := json.Marshal(...)` when building log entry payloads. If marshaling fails, a nil/empty `[]byte` is inserted into the database, corrupting the log entry.

The Supabase connector (`supabase.go:184-192`) correctly handles this error. The other four should follow the same pattern.

**Fix:** Check the error and `continue` (skip the entry) on failure, matching the Supabase pattern.

---

### H3. Inconsistent poll error handling across connectors

**Files:** `flyio.go`, `vercel.go`, `railway.go`, `mongodb.go` vs `supabase.go`

Fly.io, Vercel, Railway, and MongoDB connectors all `return` immediately on the first insert error, aborting processing of remaining rows in the batch. The Supabase connector uses `continue` to skip bad rows and keep processing.

This means a single malformed log entry blocks the entire polling cycle for these four connectors, while Supabase gracefully skips it.

**Fix:** Standardize on the `continue` pattern (skip and log bad entries) across all connectors, or document the intentional difference.

---

### H4. `ListenerManager.StopAll()` doesn't wait for done channels

**File:** `backend/internal/connectors/listener.go:110-117`

`Stop()` correctly waits on `<-entry.done` after cancelling. `StopAll()` calls `cancel()` and `Close()` but only waits on the WaitGroup, not the individual done channels. This is a subtle ordering inconsistency — the goroutine may still be running briefly after `StopAll()` returns.

**Fix:** Either wait for each done channel in `StopAll()`, or accept the `wg.Wait()` as sufficient (it is, since the goroutine calls `wg.Done()` before reading from `done`). If the latter, add a comment documenting why.

---

### H5. ConnectionWizard async race on rapid "Next" clicks

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue:118-122`

`goNext()` creates a connection before advancing to the next step. Rapid double-clicks can fire two creation requests, potentially creating duplicate connections. The `createdConnectionId` ref gets overwritten by whichever request finishes last.

**Fix:** Add a `creating` guard ref that disables the button during the async operation.

---

### H6. NotificationChannels missing error handling on list refetch

**File:** `frontend/src/components/notifications/NotificationChannels.vue:119`

After a successful save or delete, the component calls `listNotificationChannels()` to refresh the list. If that refetch fails, the UI shows stale data with no error indication.

**Fix:** Wrap the refetch in a try/catch and set an error state on failure.

---

### H7. Dead API exports never imported

**Files:**
- `frontend/src/api/applications.ts:30` — `getApplication()`
- `frontend/src/api/connections.ts:9` — `getConnection()`
- `frontend/src/api/conversations.ts:4` — `listConversations()`

These three functions are exported but never imported anywhere in the codebase.

**Fix:** Remove them. They can be re-added when needed.

---

### H8. Migration 026 rollback is destructive for multi-org users

**File:** `backend/migrations/026_org_members.down.sql`

The down migration discards all org memberships except the earliest one per user. This is documented in the file, but running it in production permanently deletes multi-org data.

**Fix:** Consider adding a guard that prevents this migration from running if multi-org memberships exist, or accept the documented risk.

---

### H9. Leaf data tables still user-scoped — breaks multi-member orgs

**Files:** `backend/internal/api/handlers/connections.go:33`, `backend/migrations/013_enable_rls.up.sql:24-56`, `log_buffer`/`conversations`/`agent_log`/`investigations` queries

The org_members migration (026) enables multiple users per org, but the leaf data tables were never migrated to org-scoped access:

- `ListConnectionsByUser` filters by `user_id` — Member B cannot see connections created by Member A
- `log_buffer` RLS policies use `user_id = app_current_user_id()` — Member B sees zero logs
- `conversations`, `agent_log`, `investigations` queries are all scoped to `user_id`

If multi-member orgs ship, new team members see empty dashboards. The entire leaf data layer needs org-scoping through `org_members` joins, matching the pattern already used for `applications` and `organizations`.

**Fix (if shipping multi-member now):** Rewrite connection/log queries and RLS to scope through `org_members` like the app/org tables already do. **Fix (if deferring multi-member):** Add a comment and ensure the invite flow warns that multi-member data visibility is limited.

---

### H10. No panic recovery in monitor/scheduler per-app goroutines

**Files:** `backend/internal/agent/monitor.go:77-83`, `backend/internal/agent/scheduler.go:107-111`

Per-app goroutines in `monitorTick` and `schedulerTick` have no `recover()`. A panic in any classifier, tool dispatch, or DB call crashes the entire server process. Since `wg.Add(3)` is called in `Start()`, a panic also means `wg.Wait()` in `Stop()` hangs indefinitely — the process cannot shut down cleanly.

```go
// monitor.go:77 — no recover() in this goroutine
go func(app db.ListActiveApplicationsRow) {
    defer wg.Done()
    defer func() { <-sem }()
    appCtx, cancel := context.WithTimeout(ctx, monitorAppTimeout)
    defer cancel()
    a.monitorApp(appCtx, app)
}(app)
```

**Fix:** Add `defer func() { if r := recover(); r != nil { slog.Error("monitor: panic in per-app goroutine", "app_id", app.ID, "panic", r) } }()` at the top of each per-app goroutine in both `monitorTick` and `schedulerTick`.

---

### H11. Monitor advances cursor past flagged logs exceeding `maxFlaggedForLLM`

**File:** `backend/internal/agent/monitor.go:152+, 207-213`

When `len(flagged) > maxFlaggedForLLM`, only the first N flagged logs are sent to the LLM for assessment. The cursor is then advanced to the timestamp of the last log in the batch (`logs[len(logs)-1]`), regardless of whether all flagged logs were assessed. The excess flagged logs are classified but their LLM assessment is permanently lost — they are never reprocessed.

**Fix:** Either (a) only advance the cursor to the timestamp of the last *assessed* log, or (b) emit an agent_log entry noting that N flagged logs were dropped due to the cap, so the team has visibility.

---

## Medium Issues

### M1. Magic "provider error" string duplicated without constant

**Files:** `backend/internal/agent/loop.go:274`, `backend/internal/agent/monitor.go:179`

The string `"provider error"` appears in two locations. Even as a short-term fix before C2 is addressed, extract it to a package-level constant.

---

### M2. Dead file: `tools_memory.go`

**File:** `backend/internal/agent/tools_memory.go`

Contains only a placeholder comment about deferred memory tools. Should be removed to avoid confusion.

---

### M3. Malformed cron expressions silently skip schedules

**File:** `backend/internal/agent/scheduler.go:138-145`

When `cronParser.Parse()` fails, the error is logged but no agent_log entry is emitted. Users have no UI visibility into broken schedules — they just silently stop running.

**Fix:** Emit an agent_log entry so the issue appears in the Activity feed.

---

### M4. `notifications.ts` uses `.then()` pattern instead of `async/await`

**File:** `frontend/src/api/notifications.ts`

All 8 functions in this file use `.then(r => r.data)` while every other API module uses `async/await`. Functionally equivalent but inconsistent with codebase conventions.

**Fix:** Refactor to `async/await` pattern for consistency.

---

### M5. GitHub connector doesn't drain response body in `Health()`

**File:** `backend/internal/connectors/codebase/github.go:67-77`

The `Health()` method closes the response body without draining it, preventing HTTP connection reuse.

**Fix:** Add `io.Copy(io.Discard, resp.Body)` before `resp.Body.Close()`.

---

### M6. GitHub `apiGet` timeout can shadow caller's deadline

**File:** `backend/internal/connectors/codebase/github.go:342`

`context.WithTimeout(ctx, apiTimeout)` creates a new deadline. If the caller's context has a shorter deadline, Go takes the sooner one (this is correct). But if the caller has a *longer* deadline, the `apiTimeout` correctly caps it. No bug here on re-inspection — Go's `WithTimeout` always takes the sooner deadline. **Resolved — no change needed.**

---

### M7. Syslog `Stream()` method ignores `out` channel parameter

**File:** `backend/internal/connectors/logs/syslog.go:192-194`

`Stream()` receives an `out chan<- []byte` parameter but delegates to `Listen()` which doesn't use it. This satisfies the `StreamConnector` interface but violates its semantic contract.

**Fix:** Either implement proper channel output or document why this connector uses `Listen()` directly instead.

---

### M8. ConnectionForm type coercion for number fields

**File:** `frontend/src/components/connections/ConnectionForm.vue:126`

```typescript
config.value[field.key] = value === '' ? '' as unknown as number : Number(value)
```

Unsafe type assertion that casts empty string through `unknown` to `number`. Works at runtime but bypasses TypeScript's safety guarantees.

**Fix:** Use a proper sentinel value (e.g., `undefined` or `null`) for empty numeric fields.

---

### M9. BaseSelect keyboard navigation doesn't validate bounds

**File:** `frontend/src/components/common/BaseSelect.vue:89`

`list.children[focusedIndex.value]` accesses DOM children without validating that `focusedIndex` is within bounds. If the list is empty or index is -1, this silently fails.

**Fix:** Add bounds check before accessing `children[focusedIndex.value]`.

---

### M10. NotificationsPage template ref method calls not type-safe

**File:** `frontend/src/pages/NotificationsPage.vue:28-29, 53-54`

Calls `prefsRef.value?.resetEditing()` and `channelsRef.value?.resetForm()` on child component refs. If child components are refactored and these methods are removed, the calls fail silently.

**Fix:** Use `defineExpose` on child components and type the template refs against the exposed interface.

---

### M11. Auth token refresh treats all errors as session invalid

**File:** `frontend/src/stores/auth.ts:29-40`

`refreshSession()` errors are treated uniformly — any failure clears the session and logs the user out. A transient 500 error from Supabase would incorrectly log the user out.

**Fix:** Distinguish between auth errors (401/403 → clear session) and transient errors (500/network → retry or keep session).

---

### M12. API client toast race condition

**File:** `frontend/src/api/client.ts:27-52`

The 401 response interceptor uses a dynamic import for the toast composable. Multiple simultaneous 401 errors trigger multiple async imports, and the redirect on line 44 may execute before the toast is displayed.

**Fix:** Use a synchronous toast reference or queue messages before redirecting.

---

### M13. ScheduleModal stale validation on mode switch

**File:** `frontend/src/components/schedules/ScheduleModal.vue:89-110`

If a user enters a custom cron expression and then switches back to preset mode, the stale custom cron value may still influence validation through `buildInput()`.

**Fix:** Reset custom cron when switching away from custom mode.

---

### M14. Inconsistent `extractApiError` usage across components

**Files:** `NotificationChannels.vue`, `StepTest.vue` vs `ConnectionForm.vue`, `ConnectionTestModal.vue`

Some components use `extractApiError()` for user-facing error messages, others pass raw error messages. This creates inconsistent error formatting.

**Fix:** Standardize on `extractApiError()` for all user-facing errors.

---

### M15. `UpdateOrganization` doesn't validate slug format

**File:** `backend/internal/api/handlers/organizations.go` (update handler)

`CreateNewOrganization` and `Onboard` validate slugs against `slugRe` (`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`), but `UpdateOrganization` passes the slug directly to the DB without validation. An admin could update to an invalid format like `INVALID--SLUG!!`.

**Fix:** Apply `slugRe.MatchString(req.Slug)` in `UpdateOrganization` when the slug field is non-empty.

---

### M16. `organization` ref typed as `Organization | null` — loses `role` guarantee

**File:** `frontend/src/stores/app.ts:10`, `frontend/src/types/organization.ts:7,19-21`

The `organization` ref is typed `Organization | null` but is always assigned `OrganizationWithRole` values (which includes `role`). However, `onboard()` at line 109 sets `organization.value = result.organization` where the API response type is `Organization` (no guaranteed `role`). After onboarding, `org.role` is `undefined`, causing permission checks like `org.role === 'owner'` in `OrgSettingsPage` and `OrgTeamPage` to silently evaluate to `false` — hiding admin/owner controls from the org creator.

**Fix:** Type the ref as `OrganizationWithRole | null`. Ensure line 109 in `onboard()` spreads `role: 'owner'` onto the organization value (it already does this for the `organizations` array on line 110, but not for `organization.value`).

---

### M17. Org delete always redirects to onboarding (ignores other orgs)

**File:** `frontend/src/pages/org/OrgSettingsPage.vue:83-85`

After deleting an org, the code calls `appStore.reset()` then navigates to `onboarding`. If the user belongs to multiple orgs, they should be redirected to select another org, not forced through onboarding. `reset()` clears the `organizations` array, so the router guard sends them to onboarding regardless of other memberships.

**Fix:** After delete, call `appStore.init()` instead of `reset()`. If orgs remain, `init()` selects the first one and routes to `/org`. If none remain, `init()` sets `needsOnboarding = true` and the guard handles it.

---

### M18. Scheduler creates new semaphore per tick — no global concurrency cap

**File:** `backend/internal/agent/scheduler.go:90`

Unlike `monitorTick` which receives `sem` as a parameter (created once in `Monitor()`), `schedulerTick` allocates a fresh `chan struct{}` every minute. If a tick's schedules all take >1 minute, the next tick launches another batch, for a total exceeding `maxConcurrentSchedules` concurrent LLM calls.

**Fix:** Create the semaphore once in `InvestigationScheduler()` and pass it to `schedulerTick`, matching the monitor pattern.

---

### M19. `toolQueryDatabase` opens new connection per query — no pooling

**File:** `backend/internal/agent/tools_db.go:50-58`

Every tool invocation opens a fresh `pgx.Conn`, runs a query, and closes it. In a busy interactive session (10 iterations, multiple queries each), this creates connection storms against the user's monitored database. No connection pooling, reuse, or concurrent-connection cap per user database.

**Fix:** Cache connectors per connection ID for the duration of the agent loop, closing them at session end. Alternatively, use `pgxpool` with `max_conns=2`.

---

### M20. GitHub installation token never refreshes (1hr expiry)

**File:** `backend/internal/connectors/codebase/github.go:57-64`

The GitHub installation token is obtained once in `Connect()` and cached in `g.token` with no TTL check. GitHub installation tokens expire after 1 hour. A long-running monitoring loop holding a GitHub connector open fails silently after 60 minutes — `readFile`/`searchCode` calls get 401s surfaced as tool errors to the LLM.

**Fix:** Track token creation time and refresh when approaching expiry (e.g., after 50 minutes), or refresh on 401 responses.

---

### M21. Auth state change listener doesn't reset app store

**File:** `frontend/src/stores/auth.ts:46-49`

The `onAuthStateChange` listener sets `session.value = null` on session end (token revocation, timeout) but never calls `appStore.reset()`. The app store retains the previous user's org/app data in memory. The Axios 401 interceptor does `window.location.href = '/login'` which clears state via page reload, but that's a side effect, not a guarantee. Manual logout (`ProfileDropdown.vue:26-29`) correctly calls `app.reset()`.

**Fix:** In the `onAuthStateChange` callback, when session becomes null, also call `useAppStore().reset()`.

---

### M22. No `http.MaxBytesReader` on JSON-decoding endpoints

**Files:** Multiple handlers using `json.NewDecoder(r.Body).Decode(&req)`

The webhook handler (`webhooks.go:38`) correctly uses `io.LimitReader` with 10MB, but all other JSON-decoding endpoints have no body size limit. An attacker could send a massive JSON body to exhaust memory.

**Fix:** Add `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` (1MB) at the start of JSON-decoding handlers, or apply it globally via middleware.

---

### M23. No `Access-Control-Max-Age` on CORS preflight

**File:** `backend/internal/api/middleware/cors.go:46-49`

The OPTIONS handler returns 204 but no `Access-Control-Max-Age` header. Browsers re-send preflight requests for every cross-origin request instead of caching the result. Performance impact in production.

**Fix:** Add `w.Header().Set("Access-Control-Max-Age", "3600")` to the preflight response.

---

### M24. `OrgTeamPage` lets admin attempt to remove owner

**File:** `frontend/src/pages/org/OrgTeamPage.vue:287-289`

The remove button shows for `isAdmin && member.user_id !== currentUserId`. Since `isAdmin` is true for both admins and owners, an admin can click "Remove" on an owner. The backend correctly rejects this, but the UI allows the attempt and shows a confusing error.

**Fix:** Add `&& member.role !== 'owner'` to the remove button condition, or `!(member.role === 'owner' && role !== 'owner')`.

---

### M25. Redundant `idx_org_members_user_id` index

**File:** `backend/migrations/026_org_members.up.sql:22`

The primary key is `(user_id, org_id)`, which already provides a B-tree index with `user_id` as the leading column. The separate `idx_org_members_user_id` index is redundant write overhead.

**Fix:** Remove the index in a future migration. Low priority.

---

### M26. `InviteMember` TOCTOU race — duplicate check outside transaction

**File:** `backend/internal/api/handlers/org_members.go:165-181`

The `GetOrgMembership` check (line 165) uses `s.Queries` (unscoped), then `CreateOrgMember` (line 181) runs inside a `UserQueries` transaction. Between the check and the insert, a concurrent request could create the same membership. The DB unique constraint catches it, but the error surfaces as 500 ("failed to add member") rather than the intended 409.

**Fix:** Move the duplicate check inside the transaction, or catch the unique constraint violation and return 409.

---

### M27. `LumberClassifier` doesn't validate `len(events) == len(logs)`

**File:** `backend/internal/agent/classifier_lumber.go:54`

After `ClassifyBatch`, the code iterates `events` using index `i` to access `logs[i]`. If the Lumber library returns fewer events than input texts (e.g., on a partial failure), this panics with index-out-of-bounds.

**Fix:** Add `if len(events) != len(logs) { return nil, fmt.Errorf(...) }` guard after `ClassifyBatch`.

---

### M28. MongoDB hostname cache has double-checked locking race

**File:** `backend/internal/connectors/logs/mongodb.go:216-233`

`cachedHostname` releases the mutex before calling `getClusterHostname`, then re-acquires to store the result. Two concurrent `Poll` calls could both see an empty hostname, both make the API call, and both write. Benign in outcome (same value) but wasteful. If `getClusterHostname` fails for one caller, no hostname is cached, causing repeated API calls every poll.

**Fix:** Hold the lock across the API call (acceptable since it's a one-time init), or use `sync.Once`.

---

## Files Over 500 Lines

| File | Lines | Status |
|------|-------|--------|
| `backend/internal/db/log_buffer.sql.go` | 605 | **sqlc-generated** — exempt from refactoring |

No hand-written files exceed 500 lines. The largest hand-written files:

| File | Lines | Notes |
|------|-------|-------|
| `backend/internal/api/handlers/applications_test.go` | 476 | Test file — acceptable |
| `backend/internal/api/handlers/applications.go` | 432 | Near limit, monitor growth |
| `frontend/src/components/connections/useAgentNebula.ts` | 419 | WebGL rendering — inherently dense |
| `backend/internal/api/handlers/notifications.go` | 411 | Near limit, monitor growth |
| `backend/internal/api/handlers/connections.go` | 407 | Near limit, monitor growth |
| `backend/internal/connectors/logs/syslog.go` | 387 | Protocol parsing — acceptable density |
| `backend/internal/agent/loop.go` | 374 | Near limit, monitor growth |

None require immediate refactoring, but `applications.go`, `notifications.go`, and `connections.go` are trending toward the threshold.

---

## Recommended Fix Order

1. **C1 + C2** — Security and data integrity in the agent subsystem (same files, same PR)
2. **H10** — Panic recovery in monitor/scheduler goroutines (server stability)
3. **C3** — ActivityPage connection scoping (quick frontend fix)
4. **H1 + M15** — Onboard handler missing Provider + slug validation on update (same files)
5. **H2 + H3** — Connector error handling standardization (same subsystem, same PR)
6. **M16 + M17** — Organization type fix + org delete flow (frontend correctness)
7. **C4** — app_id NOT NULL migration (requires production data validation first)
8. **H9** — Leaf table org-scoping (decision required: ship multi-member now or defer?)
9. **H11 + M27** — Monitor cursor/classifier length guard (agent accuracy)
10. **H5 + H6 + H7** — Frontend cleanup batch
11. **M18 + M19** — Scheduler semaphore + tool DB pooling (resource management)
12. **M20 + M21 + M22 + M23** — Token refresh, auth state, body limits, CORS (hardening)
13. **M1–M14, M24–M28** — Remaining medium issues in priority order
