# Code Assessment — v0.42.4 (Post Multi-Org Hardening)

> Comprehensive codebase review covering backend (Go), frontend (Vue 3 + TypeScript), database layer, connectors, and migrations.
> Date: 2026-04-13 | Scope: Full repository

## Overall Rating: 7.5 / 10

The multi-org implementation (phases 1–8) and the four hardening patches (v0.42.1–0.42.4) addressed many structural issues. However, this review found **2 critical**, **6 high**, **15 medium**, and **12 low** severity issues remaining. The criticals and highs must be fixed before production push. Once those are resolved, the codebase is at a solid 8.5+.

---

## Severity Legend

| Level | Meaning |
|-------|---------|
| **CRITICAL** | Broken core functionality or security exposure in production |
| **HIGH** | Auth bypass, data integrity bug, or cross-tenant leak |
| **MEDIUM** | Correctness edge case, race condition, or reliability issue |
| **LOW** | Inconsistency, minor UX bug, or latent risk |

---

## Critical Issues

### C1. OTLP ingestion endpoint is completely broken

**File:** `backend/internal/db/queries/log_buffer.sql:38`
**Used by:** `backend/internal/api/handlers/otlp.go:75`

`IngestOTLPLogs` calls `GetConnectionByWebhookToken`, but the SQL query hard-filters on `type = 'webhook_logs'`:

```sql
WHERE config->>'webhook_token' = @webhook_token::text AND type = 'webhook_logs' AND status = 'active';
```

An OTLP connection has `type = 'otlp'`, so the query will **never** match. Every OTLP ingest request returns 401 "invalid token" regardless of the token's validity. The entire OTLP ingestion path is non-functional.

**Fix:** Change the SQL to `AND type IN ('webhook_logs', 'otlp')`, or create a separate `GetConnectionByToken` query that accepts a type parameter.

---

### C2. Multi-org operations all resolve to the user's primary org only

**Files:** `backend/internal/api/handlers/org_members.go:24`, `organizations.go:18`, `applications.go:25,67`

`resolveOrgAndRole()` and every org-write handler calls `GetOrganizationByUser`, which returns the **earliest** membership:

```sql
SELECT o.* FROM organizations o
JOIN org_members om ON om.org_id = o.id
WHERE om.user_id = $1
ORDER BY om.created_at ASC LIMIT 1;
```

The frontend sends org-context requests (update org, manage members, list apps) after the user switches to org B via the header dropdown — but the backend always resolves to org A (primary). This means:

- A user belonging to orgs A and B **cannot** manage org B's members, settings, or apps
- All `PUT /api/org`, `DELETE /api/org`, member management, and `GET /api/apps` always operate on the primary org
- The frontend multi-org switcher gives the illusion of switching, but the backend ignores it

**Fix:** Either:
1. Thread an `org_id` param/header from the frontend and use `GetOrgMembership(userID, orgID)` to validate, or
2. Add an `X-Org-ID` header that the backend reads and verifies membership for

This is the single most impactful bug — it renders the entire multi-org feature non-functional for users with 2+ orgs.

---

## High Issues

### H1. WebSocket `app_id` param not authorization-checked — cross-tenant data access

**File:** `backend/internal/api/handlers/chat.go:57-64`

The WebSocket handler accepts an `app_id` query param and passes it directly to `RunConversationStream`. There is no `GetApplicationByOrgUser` check — any authenticated user who knows (or guesses) another org's app UUID can scope the agent's `search_codebase` tool to that org's GitHub repos.

```go
if appIDStr := r.URL.Query().Get("app_id"); appIDStr != "" {
    parsed, err := uuid.Parse(appIDStr)
    if err == nil {
        appID = parsed  // No ownership check
    }
}
```

**Fix:** Add `authorizeApp(userID, appID)` before accepting the app_id. Return a WebSocket error if the app doesn't belong to the user's org.

---

### H2. `RemoveMember` sole-owner guard checks total members, not owners

**File:** `backend/internal/api/handlers/org_members.go:260-269`

```go
if targetUserID == userID && callerRole == db.OrgMemberRoleOwner {
    count, err := s.Queries.CountOrgMembers(r.Context(), org.ID)
    // ...
    if count <= 1 {
```

`CountOrgMembers` counts **all** members (owners + admins + members). An org with 5 members where the caller is the **sole owner** passes this check (`count=5 > 1`) and successfully removes themselves — leaving the org permanently ownerless with no one able to manage it.

**Fix:** Create a `CountOrgOwners` query: `SELECT COUNT(*) FROM org_members WHERE org_id = $1 AND role = 'owner'`. Block removal when the owner count would drop to 0.

---

### H3. `org_members` table has no RLS enabled

**File:** `backend/migrations/026_org_members.up.sql`

Migration 026 creates the `org_members` table but never calls `ALTER TABLE org_members ENABLE ROW LEVEL SECURITY`. The handler-level `resolveOrgAndRole` check provides application-level protection, but any direct query through `UserQueries()` (which sets `app.current_user_id` for RLS) bypasses this gate. Without RLS on `org_members`, a query like `ListOrgMembers` with a manually crafted `org_id` would return results for any org.

**Fix:** Add a new migration:
```sql
ALTER TABLE org_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY org_members_user_policy ON org_members
    FOR ALL USING (org_id IN (
        SELECT org_id FROM org_members WHERE user_id = app_current_user_id()
    ));
```

---

### H4. `InviteMember` response has no Content-Type header

**File:** `backend/internal/api/handlers/org_members.go:146-148`

```go
w.WriteHeader(http.StatusCreated)           // headers flushed here
w.Header().Set("Content-Type", "application/json")  // too late — ignored
json.NewEncoder(w).Encode(...)
```

In Go's `net/http`, calling `WriteHeader` flushes headers. The `Content-Type` set afterward is silently dropped. The response body is JSON but the Content-Type defaults to `text/plain`, which may cause client parsing issues.

**Fix:** Swap the two lines — set the header before writing the status code.

---

### H5. UTF-8 string truncation can panic on multi-byte characters

**File:** `backend/internal/api/handlers/chat.go:150-152`

```go
if len(title) > 50 {
    title = title[:50] + "..."
}
```

`len()` on a Go string returns byte count. Slicing at byte index 50 on CJK text or emoji can split a multi-byte rune, producing invalid UTF-8 or a panic.

**Fix:** `runes := []rune(title); if len(runes) > 50 { title = string(runes[:50]) + "..." }`

---

### H6. `DeleteOrganization` and `UpdateOrganization` bypass RLS

**Files:** `backend/internal/api/handlers/organizations.go:162,205`

Both handlers call `s.Queries.UpdateOrganization()` / `s.Queries.DeleteOrganization()` directly on the shared `Queries` instance (which does not set `app.current_user_id`), not through `UserQueries()`. The handler-level `resolveOrgAndRole` check is the only protection. If a code path ever calls these queries outside the handler context, RLS won't help.

**Fix:** Route these through `UserQueries()` for defense-in-depth, matching the pattern used by all other write operations.

---

## Medium Issues

### M1. Monitor semaphore blocks context cancellation — delayed shutdown

**File:** `backend/internal/agent/monitor.go:67-68`

```go
sem <- struct{}{} // blocks if all 10 slots full — no ctx.Done() select
```

If all 10 goroutine slots are busy when `ctx` is cancelled (shutdown), the main loop blocks until a slot frees. This can delay clean shutdown by up to `monitorAppTimeout` (2 minutes).

**Fix:** Use `select { case sem <- struct{}{}: ... case <-ctx.Done(): return }`.

---

### M2. Rate-limiter interruption skips cursor advance — duplicate LLM escalation

**File:** `backend/internal/agent/monitor.go:160-198`

When `a.limiter.Wait(ctx)` returns an error (context cancelled), `monitorApp` returns early without advancing the cursor via `UpsertMonitoringState`. On the next tick, the same flagged logs are fetched and re-escalated to the LLM — producing duplicate alerts and wasting API quota.

**Fix:** Advance the cursor before the rate-limiter wait, or track escalated log IDs separately.

---

### M3. Lumber classification failure escalates all logs uncapped

**File:** `backend/internal/agent/classifier_lumber.go:37-49`

If `ClassifyBatch` errors, the entire batch is escalated as `UNCLASSIFIED/warning`. For a 200-log batch this bypasses the `maxFlaggedForLLM = 50` cap applied post-classification, potentially sending all 200 logs to Claude.

**Fix:** Apply the `maxFlaggedForLLM` cap inside the error branch as well.

---

### M4. `GetMonitoringState` called twice per app per tick

**File:** `backend/internal/agent/monitor.go:89,110`

`shouldMonitor` queries `GetMonitoringState`, then `monitorApp` queries it again for the cursor. Doubles the DB round-trips per app per tick across all periodic-mode apps.

**Fix:** Thread the state from `shouldMonitor` into `monitorApp`.

---

### M5. `UpdateConnection` treats all DB errors as 404

**File:** `backend/internal/api/handlers/connections.go:336`

Any `UpdateConnection` failure returns 404 "connection not found" regardless of cause. A transient DB error (timeout, pool exhaustion) is mis-reported. Other handlers correctly distinguish "not found" from server errors.

**Fix:** Check for `pgx.ErrNoRows` specifically; return `jsonServerError` for other errors.

---

### M6. `UpdateGitHubRepos` — no bound on input array size

**File:** `backend/internal/api/handlers/github_repos.go:208-225`

The 1 MB body limit permits ~50,000 small repo objects, each triggering an individual `UpsertGitHubRepo` DB round-trip with no transaction and no count cap.

**Fix:** Cap the array length (e.g. 500) and wrap in a transaction.

---

### M7. `resolveOrgAndRole` costs 2 DB round-trips per call

**File:** `backend/internal/api/handlers/org_members.go:17-40`

Calls `GetOrganizationByUser` then `GetOrgMembership` sequentially. Every member-management endpoint pays this double cost. `ListOrganizationsByUser` already returns the role — a combined query would halve the latency.

**Fix:** Create a `GetOrgAndRoleByUser` query returning org + role in one join.

---

### M8. `revealedFields` Set mutation doesn't trigger Vue reactivity

**File:** `frontend/src/components/connections/ConnectionDetailModal.vue:103-109`

`revealedFields` is `ref<Set<string>>`. In-place `.add()` / `.delete()` mutations on a Set held in a `ref` are not tracked by Vue 3's reactivity system. The Show/Hide toggle button will not update.

**Fix:** Replace with a `ref<string[]>` using array ops, or reassign to a new Set on each mutation.

---

### M9. `ConnectionsPage.vue` — stale `bubbleEls` DOM refs after app switch

**File:** `frontend/src/pages/ConnectionsPage.vue:31,56-66,90`

`bubbleEls` array is never cleared when the app changes. If the new app has fewer connections, stale DOM refs at the tail are passed to `FlowLines`, which calls `getBoundingClientRect()` on unmounted elements.

**Fix:** Reset `bubbleEls.value = []` when refetching connections.

---

### M10. `OrgSettingsPage.vue` — form fields not re-seeded on org switch

**File:** `frontend/src/pages/org/OrgSettingsPage.vue:36-41`

`editName` and `editSlug` are populated from `org.value` in `onMounted` only. If the user switches org via the header dropdown while on this page, the form still shows the old org's values. Saving would overwrite the wrong org's data (though C2 prevents this from reaching a non-primary org — once C2 is fixed, this becomes a data corruption path).

**Fix:** Add `watch(org, ...)` to re-seed form fields.

---

### M11. `ReportsPage.vue` — does not react to app switching

**File:** `frontend/src/pages/ReportsPage.vue:9-11`

`fetchReports()` called once in `onMounted` with no `watch` on `currentAppId`. Every other data page watches for app changes. Switching apps leaves stale report data visible.

**Fix:** Add `watch(() => appStore.currentAppId, () => store.fetchReports())`.

---

### M12. `AgentChatPage.vue` — WebSocket locked to app at mount time

**File:** `frontend/src/pages/AgentChatPage.vue:12`

`useAgent` captures `appStore.currentAppId` once at construction. Switching apps while on the chat page continues the session scoped to the old app. No `watch` triggers a reconnect.

**Fix:** Watch `currentAppId` and reinitialize the WebSocket on change (or redirect to force remount).

---

### M13. `ConnectionForm.vue` — supabase direction not locked when editing

**File:** `frontend/src/components/connections/ConnectionForm.vue:192-203`

The `watch(type, ...)` correctly locks direction to `one_way` when type changes to `supabase`, but when editing an existing supabase connection, the direction field remains interactive and can be changed to `two_way`.

**Fix:** Add `:disabled="type === 'supabase'"` to the direction select.

---

### M14. `SearchLogsByUser` does JSONB→text cast — no index, sequential scan at scale

**File:** `backend/internal/db/queries/log_buffer.sql:41-43`

```sql
WHERE user_id = @user_id AND payload::text ILIKE '%' || @query::text || '%'
```

Casting JSONB to text at query time prevents any index usage. At scale this will sequential-scan the entire `log_buffer` table for each search request.

**Fix:** Add a GIN index on `payload` or use `jsonb_to_tsvector` for full-text search.

---

### M15. `reports` store has no error state

**File:** `frontend/src/stores/reports.ts:10-17`

No `catch` block and no `error` ref. API failures propagate as unhandled promise rejections. Unlike `connections` and `logs` stores which expose error state, `reports` gives callers no way to show errors.

**Fix:** Add `error` ref and `catch` block matching the pattern in other stores.

---

## Low Issues

### L1. `shouldMonitor` silently treats DB errors as "first run"
**File:** `backend/internal/agent/monitor.go:89-93` — Any `GetMonitoringState` error causes the app to be included in every tick. Should log and return `false`.

### L2. `parseSeverityFromResponse` keyword fallback produces false positives
**File:** `backend/internal/agent/loop.go:346-361` — Scans for substring `"error"` anywhere; "no errors detected" triggers `error` severity.

### L3. `search_codebase` uses first connection's token for all repos
**File:** `backend/internal/agent/tools_codebase.go:46-55` — Multiple GitHub installations share one token. Repos from other installations will fail silently.

### L4. `CLASSIFIER_MODE` not validated at startup
**File:** `backend/internal/config/config.go` — Invalid values like `"FALLBACK"` or typos are silently accepted.

### L5. WebSocket wildcard origin `"*"` in production
**File:** `backend/internal/api/handlers/chat.go:49` — Should be restricted to `FrontendURL`.

### L6. `Onboard` hardcodes model ID instead of using `agent.DefaultModelID`
**File:** `backend/internal/api/handlers/organizations.go:294` — Will lag behind if the default model changes.

### L7. `NotFoundPage.vue` links to `/` (landing page) instead of `/dashboard`
**File:** `frontend/src/pages/NotFoundPage.vue:8` — Authenticated users should be sent to the dashboard.

### L8. `OrgTeamPage.vue` — optimistic role mutation not rolled back on API failure
**File:** `frontend/src/pages/org/OrgTeamPage.vue:70-78` — UI shows new role even when the API call failed.

### L9. Auth store never unsubscribes `onAuthStateChange` listener
**File:** `frontend/src/stores/auth.ts:28-31` — Leaks on re-init (HMR, future re-auth paths).

### L10. `useWebSocket.send()` throws uncaught when socket is closed
**File:** `frontend/src/composables/useWebSocket.ts:51` — `ws.send()` on a non-OPEN socket throws `DOMException`. No guard on `readyState`.

### L11. `Listener.Stop` does not wait for goroutine exit
**File:** `backend/internal/connectors/listener.go:71-84` — Can cause port bind conflicts if a new listener starts immediately after stop.

### L12. Dual `<script>` blocks in `StepWebhookSetup.vue` and `StepOTLPSetup.vue`
**Files:** `frontend/src/components/connections/wizard/steps/Step{Webhook,OTLP}Setup.vue` — Module-level `<script>` alongside `<script setup>` is fragile; move constants into setup block.

---

## Files Over 500 Lines

**None.** The v0.42.4 refactoring pass successfully split all previously oversized files. The largest files are now:

| File | Lines | Status |
|------|-------|--------|
| `useAgentNebula.ts` | 419 | OK |
| `applications.go` | 427 | OK |
| `ModelPicker.vue` | ~400 | OK |
| `OrgTeamPage.vue` | 352 | OK |

---

## Summary by Area

| Area | Critical | High | Medium | Low |
|------|----------|------|--------|-----|
| Backend — Handlers | 1 (C2) | 3 (H1,H2,H4) | 3 (M5,M6,M7) | 2 |
| Backend — Agent/Core | — | — | 4 (M1,M2,M3,M4) | 3 |
| Backend — DB/Migrations | 1 (C1) | 2 (H3,H6) | 1 (M14) | 1 |
| Frontend — Stores/API | — | — | 2 (M11,M15) | 2 |
| Frontend — Pages | — | 1 (H5) | 3 (M9,M10,M12) | 2 |
| Frontend — Components | — | — | 2 (M8,M13) | 2 |

---

## Recommended Fix Order

**Block production push — fix these first:**

1. **C2** — Multi-org operations pinned to primary org (renders multi-org non-functional)
2. **C1** — OTLP ingestion broken (entire endpoint non-functional)
3. **H1** — WebSocket app_id auth bypass (cross-tenant data access)
4. **H2** — Sole-owner guard counts wrong column (ownerless orgs)
5. **H3** — `org_members` missing RLS (DB-level auth gap)
6. **H4** — `InviteMember` Content-Type header order
7. **H5** — UTF-8 truncation panic
8. **H6** — Org mutations bypass RLS

**Fix before next release (non-blocking but important):**

9. M1–M4 — Monitor reliability (shutdown, cursor, classifier, double-query)
10. M5–M7 — Handler correctness (error mapping, unbounded input, double-query)
11. M8–M13 — Frontend reactivity and state consistency
12. M14–M15 — Search performance and error handling

**Address when convenient:**

13. L1–L12 — Minor consistency and reliability items
