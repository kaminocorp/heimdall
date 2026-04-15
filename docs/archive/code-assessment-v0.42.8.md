# Code Assessment — v0.42.8 (2026-04-14)

**Scope:** Full repository review — backend (Go), frontend (Vue 3/TypeScript), SQL/migrations, tests.
**Goal:** Production readiness at 8.5/10. Assess functionality, accuracy, maintainability, clean code.
**Files over 500 lines:** Only `db/log_buffer.sql.go` (503 lines, sqlc-generated — exempt).

---

## Overall Score: 7.0 / 10

The codebase has solid architecture, clean separation of concerns, and good patterns (sqlc, Chi router, Pinia composition API). However, the assessment uncovered **critical security gaps in SQL query scoping**, a **broken RLS migration**, several **frontend functional bugs**, and significant **test coverage holes**. These must be resolved before production deployment.

---

## Summary by Severity

| Severity | Count | Category |
|----------|-------|----------|
| Critical | 12 | SQL scoping, RLS recursion, frontend logic bugs |
| High | 18 | Auth gaps, concurrency bugs, state management |
| Medium | 25 | Consistency, performance, missing validation |
| Low | 15 | Code clarity, minor UX, dead code |

---

## CRITICAL — Must Fix Before Production

### C1. Self-referential RLS policy on `org_members` causes infinite recursion

**File:** `backend/migrations/027_org_members_rls.up.sql:6-9`

```sql
CREATE POLICY org_members_user_policy ON org_members
    FOR ALL USING (
        org_id IN (SELECT om.org_id FROM org_members om WHERE om.user_id = app_current_user_id())
    );
```

The policy on `org_members` queries `org_members` in its own USING clause. PostgreSQL will throw `ERROR: infinite recursion detected in policy for relation "org_members"`. This crashes any query against `org_members` when RLS is active for non-owner roles.

**Fix:** Create a `SECURITY DEFINER` function that bypasses RLS:

```sql
CREATE OR REPLACE FUNCTION app_user_org_ids() RETURNS SETOF UUID
SECURITY DEFINER SET search_path = public
LANGUAGE sql STABLE AS $$
  SELECT org_id FROM org_members WHERE user_id = app_current_user_id();
$$;

CREATE POLICY org_members_user_policy ON org_members
    FOR ALL USING (org_id IN (SELECT app_user_org_ids()));
```

---

### C2. Unscoped SQL queries — any authenticated user can read/modify/delete other orgs' data

The backend connects as the postgres owner role, which **bypasses all RLS policies** (documented in migration 013). This means the application code is the sole enforcement layer. The following queries have **no user/org/app ownership check** at the SQL level:

| Query | File | Risk |
|-------|------|------|
| `GetApplication` | `applications.sql:6` | Read any app by UUID |
| `UpdateApplication` | `applications.sql:33` | Rename/modify any app |
| `DeleteApplication` | `applications.sql:45` | Delete any app |
| `GetSchedule` | `investigation_schedules.sql:13` | Read any schedule |
| `UpdateSchedule` | `investigation_schedules.sql:25` | Modify any schedule |
| `DeleteSchedule` | `investigation_schedules.sql:38` | Delete any schedule |
| `GetNotificationChannel` | `notification_channels.sql:6` | Read any channel config |
| `UpdateNotificationChannel` | `notification_channels.sql:15` | Modify any channel |
| `DeleteNotificationChannel` | `notification_channels.sql:21` | Delete any channel |
| `UpdateNotificationLogStatus` | `notification_log.sql:6` | Modify any log entry |
| `DeleteGitHubRepo` | `github_repos.sql:17` | Delete any repo link |
| `UpdateConnectionStatus` | `connections.sql:21` | Change any connection's status |
| `UpdateOrganization` | `organizations.sql:27` | Rename any org |
| `DeleteOrganization` | `organizations.sql:64` | Delete any org (CASCADE) |

**Mitigation:** Many of these queries are called from handlers that do perform app-level authorization via `authorizeApp` or `resolveOrgAndRole` before reaching the query. The handlers act as the security boundary. However, this is a defense-in-depth gap — any new handler that calls these queries without the authorization check is an immediate vulnerability.

**Fix:** Add `AND app_id = $N` (joined through org_members) to every query, or create scoped variants and deprecate the unscoped ones. Priority: `DeleteApplication`, `DeleteOrganization`, `UpdateApplication` (highest blast radius).

---

### C3. `query_database` tool passes LLM-generated SQL with no statement validation

**File:** `backend/internal/agent/tools_db.go:15-16`

```go
sql, _ := input["sql"].(string)
```

The LLM's `sql` parameter is passed directly to the user's connected Postgres database. The only protection is `default_transaction_read_only=on` in the connection string, which can be bypassed with `SET default_transaction_read_only = off` in the same session.

**Fix (two layers):**
1. Parse the SQL statement and reject anything that isn't `SELECT`, `EXPLAIN`, or `WITH ... SELECT`.
2. Use a database role with `GRANT SELECT` only (not a session parameter).

---

### C4. `query_database` has no row-count limit

**File:** `backend/internal/connectors/database/postgres.go`

A `SELECT * FROM large_table` from the LLM returns unlimited rows, potentially OOMing the agent process.

**Fix:** Enforce `LIMIT 1000` (or configurable) at the connector level. If the query already has a LIMIT, use the smaller of the two.

---

### C5. `UpdateMemberRole` allows sole owner to demote themselves

**File:** `backend/internal/api/handlers/org_members.go:202-262`

An owner can change their own role to `member` or `admin`, leaving the org with zero owners. The sole-owner guard exists in `RemoveMember` but not here.

**Fix:** Add the same `CountOrgOwners` check: if `targetUserID == callerID && currentRole == owner && newRole != owner`, verify owner count > 1.

---

### C6. `ActivityPage.vue` passes ref object instead of `.value` to pagination

**File:** `frontend/src/pages/ActivityPage.vue:65-66`

```html
@next="logsStore.nextPage(activeFilters)"
@prev="logsStore.prevPage(activeFilters)"
```

`activeFilters` is a `Ref<{...}>`. The store receives the wrapper, not the plain object — filters are silently ignored on page navigation.

**Fix:** `logsStore.nextPage(activeFilters.value)` / `logsStore.prevPage(activeFilters.value)`

---

### C7. `LoginPage.vue` does not call `app.init()` after login

**File:** `frontend/src/pages/LoginPage.vue:20`

After `auth.login()`, the code calls `router.push('/dashboard')` without initializing the app store. `appStore.currentAppId` remains `null`, `applications` remains `[]`, and all dashboard API calls pass `undefined` as the app ID.

**Fix:** Add `await appStore.init()` between `auth.login()` and `router.push('/dashboard')`.

---

### C8. WebSocket `token` and `appId` are stale closures — not reactive

**File:** `frontend/src/composables/useAgent.ts:22-26`

`auth.token` and `appStore.currentAppId` are captured once at setup time. If the token refreshes (Supabase auto-refresh) or the user switches apps, the WebSocket continues using the old values.

**Fix:** Watch for token/appId changes and reconnect the WebSocket when they change.

---

### C9. No WebSocket reconnection logic

**File:** `frontend/src/composables/useWebSocket.ts:40-48`

Any network interruption, server restart, or idle timeout permanently kills the chat. The user sees a dead input with no way to recover without reloading the page.

**Fix:** Implement reconnection with exponential backoff (1s, 2s, 4s, max 30s). Cap retries, then show a "Reconnect" button.

---

### C10. Monitor cursor advances even when LLM provider fails

**File:** `backend/internal/agent/monitor.go:199-205`

If `RunMonitoring` returns a provider error string (not a Go error), the cursor is still advanced past unprocessed logs. Those logs are permanently skipped from re-analysis.

**Fix:** Only advance the cursor when `RunMonitoring` returns a non-error assessment. Check the severity or add a success boolean return value.

---

### C11. CORS middleware missing `PATCH` method

**File:** `backend/internal/api/middleware/cors.go:38`

```go
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
```

`PATCH` is missing. `PATCH /api/apps/{appId}/schedules/{id}` will fail CORS preflight in browsers.

**Fix:** Add `PATCH` to the allowed methods string.

---

### C12. `NotFoundPage.vue` links to `/` instead of `/dashboard`

**File:** `frontend/src/pages/NotFoundPage.vue:8`

The "Return to Dashboard" link goes to `/` (public landing page), not `/dashboard`.

**Fix:** Change `to="/"` to `to="/dashboard"`.

---

## HIGH — Fix Before Production

### H1. Race condition on `Agent.Start`/`Stop` — no mutex on `a.cancel`

**File:** `backend/internal/agent/agent.go:83-116`

`a.cancel` is read/written without synchronization. Concurrent `Start`/`Stop` calls race.

**Fix:** Add a `sync.Mutex` protecting `a.cancel` and `a.wg` access.

---

### H2. `Poller.Stop()` doesn't drain goroutine — `Stop` then `Start` causes double-poll

**File:** `backend/internal/connectors/poller.go:70-79`

`Stop()` cancels and deletes the entry, but doesn't wait for the goroutine to exit. An immediate `Start()` bypasses the `done` channel drain (because the entry was deleted).

**Fix:** Wait on `entry.done` before deleting from the map, same pattern as `Start()`.

---

### H3. `ListenerManager.Start()` holds mutex during `Close()` (up to 5s)

**File:** `backend/internal/connectors/listener.go:50-56`

The unlock-relock pattern around `Close()` also has a TOCTOU race — another goroutine can insert a new entry for the same connection ID between unlock and relock.

**Fix:** Use the `done` channel pattern from `poller.go`. Extract the entry under lock, release lock, drain, then re-acquire to insert the new entry.

---

### H4. Listener/Poller context isolation — `context.Background()` bypasses server shutdown

**Files:** `connectors/listener.go:58`, `connectors/poller.go:55`

Both use `context.Background()` instead of the server's root context. Goroutines survive server shutdown unless `StopAll()` is explicitly called.

**Fix:** Accept a parent `context.Context` in `Start()` and derive from it, or document that `StopAll()` is mandatory in the shutdown sequence and verify it's called.

---

### H5. Syslog `Close()` leaks a goroutine on shutdown timeout

**File:** `backend/internal/connectors/logs/syslog.go:212-224`

When the 5-second shutdown timeout fires, the `go func() { s.wg.Wait(); close(done) }()` goroutine is leaked forever.

**Fix:** Use `context.WithTimeout` and select on context cancellation, or accept the goroutine leak as intentional (document it).

---

### H6. MongoDB connector uses Basic Auth — Atlas v2 API requires Digest/ApiKey auth

**File:** `backend/internal/connectors/logs/mongodb.go:250`

`req.SetBasicAuth(m.config.PublicKey, m.config.PrivateKey)` — the Atlas v2 API does not support Basic Auth. Requests will return 401.

**Fix:** Use Digest Authentication or the `ApiKey` header per Atlas v2 docs.

---

### H7. MongoDB `getClusterHostname` called on every poll cycle

**File:** `backend/internal/connectors/logs/mongodb.go:103`

Makes an extra API call every poll to retrieve a hostname that doesn't change at runtime.

**Fix:** Cache the hostname at construction/first-connect time.

---

### H8. Cursor advances past failed log inserts in Fly.io/Vercel/Railway/MongoDB connectors

**Files:** `logs/flyio.go:248`, `logs/vercel.go:162`, `logs/railway.go:173`, `logs/mongodb.go:186`

When `InsertLogEntry` fails mid-batch, `maxTS` may already be non-zero from earlier successes. The cursor advances past the failed entry, permanently losing that log.

**Fix:** On insert error, return `maxTS` from the *last successful* insert (track separately), or return zero timestamp to force re-poll from the old cursor.

---

### H9. `auth.ts` — `onAuthStateChange` subscription never unsubscribed

**File:** `frontend/src/stores/auth.ts:28`

Memory leak on HMR or if `init()` is called twice. Return value of `onAuthStateChange` is discarded.

**Fix:** Store the subscription and unsubscribe on store disposal.

---

### H10. `auth.ts` — `refreshSession` error silently discarded

**File:** `frontend/src/stores/auth.ts:20-22`

If the refresh token is revoked, `session.value` is set to `null` silently. User is logged out without feedback.

**Fix:** Check `refreshed.error` and either show a toast or redirect to login with a message.

---

### H11. `OrgSettingsPage.vue` — Direct store mutation doesn't update `organizations[]` list

**File:** `frontend/src/pages/org/OrgSettingsPage.vue:54-57`

`appStore.organization.name = updated.name` mutates the single ref but not the matching entry in `appStore.organizations[]`. The `OrgDropdown` shows the old name until full reload.

**Fix:** Add an `updateOrganization()` action to the app store that updates both `organization` and the matching entry in `organizations[]`.

---

### H12. `ConnectionsPage.vue` — `bubbleEls` never shrinks on connection deletion

**File:** `frontend/src/pages/ConnectionsPage.vue:31`

Stale DOM refs in `bubbleEls` cause `FlowLines` to call `getBoundingClientRect()` on detached elements.

**Fix:** Reset `bubbleEls` to `[]` after connection deletion, or rebuild on each render cycle.

---

### H13. `useAgent.ts` — `isThinking`/`activeTools` not reset on WebSocket close

**File:** `frontend/src/composables/useAgent.ts:28-96`

If the connection drops mid-thought, the UI shows a permanent spinner.

**Fix:** Watch `status` and reset `isThinking = false` and `activeTools = []` when status becomes `'closed'`.

---

### H14. `DashboardPage.vue` — No race-condition guard on rapid app switch

**File:** `frontend/src/pages/DashboardPage.vue:48-49`

Multiple in-flight `Promise.allSettled` calls race; the last to resolve wins regardless of which app it was for.

**Fix:** Use an `AbortController` or generation counter to discard stale responses.

---

### H15. `GitHubCallback` makes outbound GitHub API call before verifying app ownership

**File:** `backend/internal/api/handlers/github_install.go:134-148`

The `GetInstallation` call uses an attacker-controlled `installation_id` from the URL. Ownership is only verified afterward.

**Fix:** Verify `appID` belongs to `userID` before making any outbound GitHub API call.

---

### H16. `TestConnection` updates status via unscoped `s.Queries` instead of RLS-scoped `queries`

**File:** `backend/internal/api/handlers/connections_test_handler.go:177-186`

The `UpdateConnectionStatus` call bypasses the RLS transaction pattern.

**Fix:** Use the user-scoped `queries` (already in scope) for the status update.

---

### H17. `DeleteApplication` uses bare `s.Queries` for count and delete

**File:** `backend/internal/api/handlers/applications.go:164-176`

Both `CountApplicationsByOrg` and `DeleteApplication` bypass the RLS transaction pattern.

**Fix:** Use `s.UserQueries()` for both operations.

---

### H18. Scheduler runs schedules serially — hung LLM blocks all others

**File:** `backend/internal/agent/scheduler.go:89-95`

`RunScheduledInvestigation` blocks up to 5 minutes. If many schedules fire at once, subsequent ticks are dropped, causing permanent schedule slip.

**Fix:** Run each schedule in its own goroutine with a concurrency semaphore (similar to the monitor's approach).

---

## MEDIUM — Should Fix

### M1. `connections.go` — `UpdateConnection` silently resets status to `"inactive"` if client omits the field

**File:** `handlers/connections.go:270-273`

**Fix:** Preserve the existing status when the field is empty.

### M2. `webhooks.go` — Polymorphic response shape (single object vs array)

**File:** `handlers/webhooks.go:90-94`

**Fix:** Always return an array.

### M3. `conversations.go` — Hardcoded `Limit: 50, Offset: 0` with no pagination

**File:** `handlers/conversations.go:37-39`

**Fix:** Parse `?limit=` and `?offset=` query params, same pattern as `ListLogs`.

### M4. No role-based authorization on app-level write operations

**Files:** `applications.go`, `investigation_schedules.go`, `notifications.go`

`authorizeApp` checks org membership but not role. Any `member` can change agent config, create schedules, modify notifications.

**Fix:** Add `hasMinRole("admin")` check to write operations, or document that all members have full app access.

### M5. `ListLogsSinceForApp` and `GetAppDashboardStats` don't use the `lb.app_id` index

**Files:** `monitoring.sql:27`, `stats.sql:3`

Both JOIN through `connections` instead of filtering directly on `lb.app_id` (available since migration 025).

**Fix:** Rewrite to `WHERE lb.app_id = $1 AND lb.ingested_at > $2` to use `idx_log_buffer_app_id`.

### M6. No `SearchLogsByUserAndApp` query variant

**File:** `log_buffer.sql`

`SearchLogsByUser` only scopes by `user_id`, not `app_id`. Per-app Activity page search is broken.

**Fix:** Add an app-scoped search query.

### M7. `investigations` and `conversations` tables still user-scoped, not org-scoped

**Files:** `investigations.sql`, `conversations.sql`

Inconsistent with the org-centric data model. Investigations aren't visible to other org members.

**Fix:** Add `app_id` to both tables and migrate existing data; or document as intentionally user-private.

### M8. `app.ts:selectOrg` swallows all errors silently

**File:** `frontend/src/stores/app.ts:83-92`

Network errors and 500s are treated as "org has no apps".

**Fix:** Only catch 404; re-throw other errors.

### M9. `useWebSocket.ts` — Identical consecutive messages may not trigger `watch`

**File:** `composables/useWebSocket.ts:37`

Vue's `watch` uses shallow equality. Two identical messages won't re-trigger.

**Fix:** Wrap in `{ payload, ts: Date.now() }` or use an event emitter.

### M10. `OrgTeamPage.vue` — Single `error` ref for fetch and action errors

**File:** `pages/org/OrgTeamPage.vue:74-78`

**Fix:** Use separate `fetchError` and `actionError` refs.

### M11. Stores mutation actions inconsistently handle errors

**Files:** `stores/connections.ts`, `stores/schedules.ts`

`fetchX` actions set `loading`/`error` properly. `createX`, `updateX`, `deleteX` do neither.

**Fix:** Add consistent error/loading state management or document that callers must wrap in try/catch.

### M12. `connections_test_handler.go` — Error details may leak credentials to client

**File:** `handlers/connections_test_handler.go:61-62`

`fmt.Sprintf("Failed to initialize: %v", err)` — connector errors may include DSN/API keys.

**Fix:** Return a generic message; log the detailed error server-side.

### M13. No slug format validation in `CreateNewOrganization` / `Onboard`

**File:** `handlers/organizations.go:69, 235`

Arbitrary strings accepted as slugs (spaces, unicode, control chars).

**Fix:** Validate with `/^[a-z0-9-]+$/` regex, 3-50 chars.

### M14. Empty tool-results message appended to conversation

**Files:** `agent/loop.go:217`, `agent/loop.go:334`

If all blocks are non-`tool_use`, an empty `ToolResults` message is sent to the provider.

**Fix:** Guard with `if len(toolResults) > 0` before appending.

### M15. `Vary: Origin` not set unconditionally in CORS middleware

**File:** `middleware/cors.go:36-45`

Caching intermediaries may incorrectly cache responses without CORS headers.

**Fix:** Set `Vary: Origin` on all responses, not just allowlisted origins.

### M16. Authenticated users can reach `/onboarding` after setup is complete

**File:** `frontend/src/router/index.ts:32-36`

No guard prevents re-entry. Could cause duplicate org creation.

**Fix:** Add a guard in the router or in `OnboardingPage` that redirects if `!app.needsOnboarding`.

### M17. GitHub codebase connector — empty repo list searches all of public GitHub

**File:** `backend/internal/connectors/codebase/github.go:109-130`

If no repos are connected, the search API call has no `repo:` filter.

**Fix:** Return an empty result if the repo list is empty.

### M18. `github_repos.go` — Unbounded repo array in `UpdateGitHubRepos`

**File:** `handlers/github_repos.go:208-225`

1MB body limit still allows thousands of entries, each triggering a DB upsert in one transaction.

**Fix:** Cap at 100 repos (or configurable limit).

---

## LOW — Nice to Have

| # | File | Issue |
|---|------|-------|
| L1 | `middleware/auth.go:234` | Duplicated `jsonError` function (same as `helpers.go`) |
| L2 | `org_members.go:127` | No email format validation on invite |
| L3 | `notifications.go:351` | Email recipients not validated / no count limit |
| L4 | `chat.go:280` | `persistMessages` silently swallows DB errors |
| L5 | `connections_test_handler.go:187` | `TestConnection` always returns HTTP 200 even on failure |
| L6 | `syslog.go:304` | RFC 5424 regex misparses STRUCTURED-DATA field |
| L7 | `railway.go:31` | `ServiceID`/`EnvironmentID` parsed but never used |
| L8 | `types/agent.ts:35` | `WSMessage` union missing `tool_start`/`tool_result` types |
| L9 | `types/api.ts:7` | `PaginatedResponse<T>` is dead code — never used |
| L10 | `useToast.ts:18` | Dangling `setTimeout` on manually dismissed toasts |
| L11 | `014.up.sql:12` | Duplicate unique index on `organizations.slug` |
| L12 | `api/notifications.ts:32` | `deleteNotificationChannel` returns `AxiosResponse` not `void` |
| L13 | `config.go:65` | `SupabaseURL` required in validation but unused in reviewed code |
| L14 | `classifier_lumber.go:57` | Low-confidence logs always escalated (inflates noise metrics) |
| L15 | `026.down.sql:44` | Inconsistent `public.` schema prefix in rollback policies |

---

## Test Coverage Gaps

### Untested Handler Endpoints (Zero Coverage)

| Endpoint | Handler |
|----------|---------|
| `GET /api/orgs` | `ListUserOrganizations` |
| `POST /api/orgs` | `CreateNewOrganization` |
| `PUT /api/org` | `UpdateOrganization` |
| `DELETE /api/org` | `DeleteOrganization` |
| `GET /api/connections/{id}` | `GetConnection` |
| `POST /api/connections/{id}/test` | `TestConnection` (HTTP layer) |
| `GET/PUT /api/connections/{id}/github/repos` | `ListGitHubRepos`, `UpdateGitHubRepos` |
| `GET /api/github/install` | `InstallGitHub` |
| `GET /api/github/callback` | `GitHubCallback` |
| All 8 notification endpoints | Preferences, Channels CRUD, Test, History |
| `GET /api/conversations[/{id}]` | `ListConversations`, `GetConversation` |
| `GET /api/models` | `GetAvailableModels` |

### Broken/Ineffective Tests

| Test | Issue |
|------|-------|
| `auth.test.ts:79-86` | `isAuthenticated` test always asserts `false` — never verifies post-session change |
| `applications_test.go:403` | Dead assertion block — first provider mismatch silently untested |
| `logs.test.ts:49` | `nextPage()` not awaited — test passes accidentally |

### Missing Critical Test Scenarios

- No cross-org connection access test (read/update/delete another org's connection)
- No test for webhook with missing `Authorization` header (vs. invalid token)
- No test for inactive connection token rejection
- `auth.signup` function entirely untested
- `auth.init()` refresh-session branch never exercised
- Postgres/GitHub codebase connectors tested only at constructor level

---

## Recommended Fix Priority

### Phase 1 — Security (blocks production)

1. **C1** — Fix `org_members` RLS recursion (migration 028)
2. **C2** — Add scoped query variants for the highest-risk unscoped queries (`DeleteApplication`, `DeleteOrganization`, `UpdateApplication`) and update handlers to use them
3. **C3** — Add SQL statement validation to `query_database` tool (reject non-SELECT)
4. **C4** — Add row-count limit to database connector
5. **C5** — Add sole-owner guard to `UpdateMemberRole`
6. **C11** — Add `PATCH` to CORS allowed methods
7. **H15** — Reorder GitHub callback to verify ownership before outbound call
8. **H16/H17** — Switch `TestConnection` and `DeleteApplication` to use `UserQueries()`

### Phase 2 — Frontend Functional Bugs

1. **C6** — Fix `ActivityPage.vue` ref vs `.value` bug
2. **C7** — Add `app.init()` to login flow
3. **C8/C9** — Implement WebSocket reconnection with reactive token/appId
4. **C12** — Fix NotFound page link
5. **H11** — Fix `OrgSettingsPage` store mutation pattern
6. **H12** — Fix `bubbleEls` stale refs
7. **H13** — Reset thinking state on WS close
8. **H14** — Add race-condition guard to DashboardPage

### Phase 3 — Backend Concurrency & Correctness

1. **H1** — Add mutex to `Agent.Start`/`Stop`
2. **H2/H3/H4** — Fix poller/listener drain and context isolation
3. **H5** — Fix syslog shutdown goroutine leak
4. **C10** — Fix monitor cursor advancement on provider failure
5. **H8** — Fix cursor advancement on insert failure across all connectors
6. **H18** — Parallelize scheduler execution
7. **M14** — Guard empty tool-results

### Phase 4 — Query Optimization & Consistency

1. **M5** — Rewrite queries to use `lb.app_id` index
2. **M6** — Add app-scoped log search query
3. **M7** — Plan org-scoping for investigations/conversations
4. **M1-M4** — Handler consistency fixes

### Phase 5 — Test Coverage

1. Add cross-org access tests for connections, schedules, notifications
2. Fix broken `auth.test.ts` `isAuthenticated` test
3. Add integration tests for notification endpoints
4. Add integration tests for org CRUD endpoints
5. Test `auth.init()` refresh-session branch
