# Code Assessment v0.42.5

**Date:** 2026-04-14
**Scope:** Full codebase — backend (Go), frontend (Vue 3 + TypeScript), SQL migrations, build config, tests
**Objective:** Assess functional correctness, security, maintainability, and clean code. Flag files over 500 lines.

---

## Overall Score: 8.4 / 10

The codebase is well-structured with consistent patterns, clear separation of concerns, and no files exceeding 500 lines. The multi-org implementation across 8 phases is functionally sound. The issues below — primarily inconsistent RLS usage, missing test coverage for new handlers, and a few frontend correctness bugs — need to be addressed to reach production-ready 8.5+.

---

## File Length Check

**All files are under 500 lines.** The recent v0.42.4 refactoring (AgentNebula, AppHeader, github.go) was effective. Largest files:

| File | Lines | Status |
|------|-------|--------|
| `useAgentNebula.ts` | 419 | OK |
| `applications.go` (handlers) | 416 | OK |
| `notifications.go` (handlers) | 403 | OK |
| `loop.go` (agent) | 373 | OK |
| `models.go` (agent) | 366 | OK |

---

## Critical Issues (0)

None found.

---

## High Severity (7)

### H1. Inconsistent RLS bypass across write handlers

**Files:** `handlers/applications.go`, `handlers/org_members.go`, `handlers/organizations.go`

Several handlers introduced in the multi-org phases bypass `UserQueries()` for write operations, meaning RLS policies are not evaluated at the DB layer. Handler-level authorization (via `resolveOrgAndRole`) provides functional protection, but this breaks the defense-in-depth pattern used everywhere else.

**Affected operations:**
- `CreateApplication` — uses `s.Queries.CreateApplication` directly (applications.go:72-93)
- `ListApplications` — uses `s.Queries` directly (applications.go:29-39)
- `CreateNewOrganization` — manual transaction without `UserQueries()` (organizations.go:74-108)
- `Onboard` — same pattern (organizations.go:248-303)
- `InviteMember`, `UpdateMemberRole`, `RemoveMember` — all use `s.Queries` directly (org_members.go)
- `ListOrgMembers` — uses `s.Queries` directly (org_members.go:92)
- `authorizeApp` helper — uses `s.Queries` for auth check (applications.go:194-217)

**Fix:** Route all writes through `UserQueries()` to set `app.current_user_id` for RLS evaluation.

### H2. WebSocket origin accepts all origins

**File:** `handlers/chat.go:49`

`OriginPatterns: []string{"*"}` allows WebSocket connections from any origin. Combined with the JWT token in the query string (visible in browser history, server logs, proxy logs), a malicious page could potentially open a WebSocket connection.

**Fix:** Use the same origin allowlist as the CORS middleware configuration.

### H3. Reports handlers are non-functional stubs

**File:** `handlers/reports.go:8-16`

Both `ListReports` and `GetReport` return empty arrays unconditionally. `GetReport` accepts a path parameter `{id}` but ignores it entirely, and returns an array — semantically wrong for a single-resource GET. The entire reports feature (store + page + 3 components) is wired end-to-end but serves no data.

**Fix:** Either implement the handlers or remove the entire reports stack (handlers, routes, frontend store/pages/components) to avoid dead code.

### H4. Unhandled rejection in 401 interceptor prevents login redirect

**File:** `frontend/src/api/client.ts:37-39`

The 401 handler does `await auth.logout()` without try/catch. If Supabase `signOut` fails (expired session, network error), the `window.location.href = '/login'` redirect never executes. User gets stuck on a broken page.

**Fix:** Wrap `auth.logout()` in try/catch so the redirect always fires.

### H5. WebSocket `send()` throws on CONNECTING state

**File:** `frontend/src/composables/useWebSocket.ts:51-53`

`send()` guards against `ws` being null but not against `readyState !== OPEN`. Calling send on a CONNECTING socket throws `DOMException`. The `AgentChatPage` disables the button when status is not open, but any other caller would crash.

**Fix:** Add `ws.readyState === WebSocket.OPEN` guard in `send()`.

### H6. Dual `<script>` block variable scoping bug in wizard steps

**Files:** `StepWebhookSetup.vue:33,42-49`, `StepOTLPSetup.vue:33,54-73`

`payloadExample` is defined in a regular `<script lang="ts">` block but referenced in a `<script setup>` template. Variables from the non-setup block are not reliably available in the setup template scope. This renders `payloadExample` as `undefined` in some Vue compiler versions.

**Fix:** Move the constant definitions inside the `<script setup>` block.

### H7. No test coverage for notification handlers

**File:** `handlers/notifications.go`

Eight notification handler methods have zero test coverage. These CRUD operations touch the database and could break silently. The test router in `testhelpers_test.go` doesn't even register notification routes.

**Fix:** Add notification routes to the test router and write tests covering CRUD + authorization.

---

## Medium Severity (17)

### M1. Test router is out of sync with production router

**File:** `handlers/testhelpers_test.go:140-206`

Missing routes: `/api/models`, `/api/v1/logs`, `/api/github/*`, all notification routes, GitHub repo routes. Handler tests for these endpoints would silently get 404/405 from the test router.

**Fix:** Keep the test router in sync with `router.go`, or generate it from the same source.

### M2. No tests for org_members handlers

**File:** `handlers/org_members.go`

Four security-sensitive handler methods (ListOrgMembers, InviteMember, UpdateMemberRole, RemoveMember) have no tests. Role management is a common attack surface.

### M3. No tests for GitHub integration handlers

**Files:** `handlers/github_install.go`, `handlers/github_repos.go`

Five handler methods including the OAuth callback flow (which performs state-based JWT auth) are untested.

### M4. Prometheus `/metrics` endpoint is unauthenticated

**File:** `router.go:119`

Mounted outside the protected route group. Exposes operational data (goroutine counts, request latencies, error rates) to unauthenticated callers.

**Fix:** Move behind auth middleware or serve on a separate internal port.

### M5. Monitor semaphore blocks shutdown

**File:** `agent/monitor.go:68`

`sem <- struct{}{}` on the main goroutine blocks without selecting on `ctx.Done()`. If all 10 slots are full, the Monitor goroutine cannot respond to shutdown signals.

**Fix:** Use a select with `ctx.Done()` alongside the semaphore send.

### M6. `parseSeverityFromResponse` false-positives on common words

**File:** `agent/loop.go:346-361`

Keyword scan checks if the entire response contains "error" or "warning". "No errors detected" would be classified as severity `error`.

**Fix:** Only apply heuristic keywords in proximity to severity-related sentence patterns, or remove the fallback heuristic entirely and rely on the structured `severity:` prefix.

### M7. Webhook batch insertion is not atomic

**File:** `handlers/webhooks.go:57-86`

Inserts log entries one at a time in a loop with no wrapping transaction. A mid-batch failure commits early entries but returns 400, giving the client no indication of which entries succeeded.

**Fix:** Wrap the batch in a transaction, or return a 207 Multi-Status response indicating per-entry results.

### M8. `SearchLogsByUser` ILIKE does not escape wildcards

**File:** `db/queries/log_buffer.sql:42-43`

User input containing `%` or `_` is interpreted as LIKE wildcards, producing incorrect search results. Not a SQL injection risk (parameterized), but a correctness bug.

**Fix:** Escape `%` and `_` in the Go code before passing to the query, or add a SQL escape function.

### M9. Poller `Start` can cause concurrent `Poll()` for same connection

**File:** `connectors/poller.go:44-46`

Cancels old context and immediately starts new goroutine without waiting for the old one to finish. Two goroutines could poll the same connection concurrently, producing duplicate log entries.

**Fix:** Wait for the old goroutine to return (e.g., via a done channel) before starting the new one.

### M10. `ListenerManager.Start` holds mutex during `Close()`

**File:** `connectors/listener.go:50-53`

`Close()` (up to 5-second drain timeout) is called while holding `m.mu`. Blocks all other listener operations.

**Fix:** Release the lock before calling `Close()`, matching the pattern in `Stop()`.

### M11. Unbounded `storedMessages` slice growth in chat

**File:** `handlers/chat.go:146-153,234-241`

Messages accumulate without limit per WebSocket session. Each message triggers full re-serialization and DB write of the growing slice.

**Fix:** Cap conversation length or implement a sliding window for the persisted messages.

### M12. Dead navigation links to missing routes

**Files:** `PublicNav.vue:35-39`, `PublicFooter.vue:12,14,19`

Links to `/security`, `/terms`, `/privacy` have no corresponding routes. Users clicking these land on the 404 page.

**Fix:** Either add the pages or remove the links.

### M13. `AgentNebula.vue` dormant prop is not reactive

**File:** `components/connections/AgentNebula.vue:17-20`

The `dormant` prop is captured once at mount time. If connections change from 0 to >0 while mounted, the nebula doesn't transition between dormant/active states.

**Fix:** Watch the prop and update the shader uniform reactively.

### M14. DashboardPage sequential API calls

**File:** `pages/DashboardPage.vue:21-38`

Five independent API calls made sequentially. Creates a visible waterfall delay on page load.

**Fix:** Use `Promise.allSettled()` to parallelize.

### M15. Orphaned `ReportDetail.vue` component

**File:** `components/reports/ReportDetail.vue`

Defined but never imported anywhere. Dead code.

### M16. `app.ts` store swallows all errors as "needs onboarding"

**File:** `frontend/src/stores/app.ts:54-56`

The catch block in `init()` sets `needsOnboarding = true` for any error — including network failures, 500s, or auth issues. A transient server error incorrectly routes users to onboarding.

**Fix:** Check the error type/status. Only treat 404 (no org) as needing onboarding.

### M17. Migration 014 deletes data unconditionally

**File:** `migrations/014_organizations_applications.up.sql:31-32`

`DELETE FROM log_buffer; DELETE FROM connections;` with no safety guard. Acceptable for early development but would destroy production data if re-run.

**Note:** Since this migration has already been applied in production, this is a historical observation rather than an actionable fix. Document this as a non-replayable migration.

---

## Low Severity (14)

| # | File | Issue |
|---|------|-------|
| L1 | `agent/extract.go:51-56` | Extracted message text not truncated to `maxClassifyChars` (Lumber handles it, but inconsistent) |
| L2 | `agent/loop.go:364-372` | `extractText` concatenates text blocks without separator |
| L3 | `agent/agent.go:77-88` | `Start()`/`Stop()` read/write `a.cancel` without sync (unlikely concurrent call) |
| L4 | `handlers/applications.go:39-46` | `ListApplications` returns `null` instead of `[]` when no apps exist |
| L5 | `handlers/connections.go:117-123` | TOCTOU window between auth check and insert (different transactions) |
| L6 | `handlers/logs.go:260-274` | `source=all` merge strategy is documented approximation, can miss rows on deep pages |
| L7 | `connectors/logs/supabase.go:214-218` | Cursor not advanced past last timestamp (works by coincidence due to `>` operator) |
| L8 | `pages/NotFoundPage.vue:9,12` | "Return to Dashboard" links to `/` (public landing) instead of `/dashboard` |
| L9 | `pages/org/OrgTeamPage.vue:75-78` | Role change mutates object before API confirms — no rollback on error |
| L10 | `components/org/AppCard.vue:4` | Unused import `formatRelativeTime` |
| L11 | `components/common/OrgDropdown.vue:4` | Unused import `OrganizationWithRole` |
| L12 | `components/common/AppSidebar.vue:58-59` | `mobile` prop applies identical class in both branches (no-op) |
| L13 | `components/notifications/NotificationHistory.vue:10-11` | `bg-status-error`/`bg-status-warning` may not match Tailwind theme definitions |
| L14 | `frontend/src/stores/reports.ts` | Reports store has no tests (other stores do — inconsistency) |

---

## Recommended Fix Priority

### Must-fix before production push (8.5+ target)

| # | Issue | Effort |
|---|-------|--------|
| H1 | Route all multi-org handlers through `UserQueries()` for RLS | Medium |
| H2 | Restrict WebSocket origin allowlist | Small |
| H4 | Wrap 401 interceptor logout in try/catch | Small |
| H5 | Guard `send()` on `readyState === OPEN` | Small |
| H6 | Move `payloadExample` into `<script setup>` block | Small |
| M4 | Move `/metrics` behind auth or internal port | Small |
| M8 | Escape ILIKE wildcards in search query | Small |
| M12 | Remove dead nav links or add placeholder pages | Small |
| M16 | Differentiate network errors from "needs onboarding" in app store | Small |

### Should-fix soon after

| # | Issue | Effort |
|---|-------|--------|
| H3 | Decide: implement reports or remove the entire stack | Medium |
| H7 | Add notification handler tests + sync test router | Medium |
| M1 | Sync test router with production router | Small |
| M2 | Add org_members handler tests | Medium |
| M5 | Select on `ctx.Done()` during semaphore acquire | Small |
| M6 | Fix severity heuristic false-positives | Small |
| M9 | Wait for old poller goroutine before starting new | Small |
| M10 | Release mutex before listener `Close()` | Small |
| M14 | Parallelize DashboardPage API calls | Small |

### Can defer

All Low severity items and remaining Medium items (M7, M11, M13, M15, M17).

---

## What's Working Well

- **File length discipline** — all files under 500 lines after v0.42.4 refactoring
- **RLS pattern** — `UserQueries()` with `SET LOCAL app.current_user_id` is well-designed; the gaps above are deviations from an otherwise solid pattern
- **Agent package** — clean separation (classifier, tools, providers, scheduler), no files over 373 lines
- **SQL migrations** — all 27 pairs present, ordered, with down migrations. RLS policies comprehensive
- **Frontend stores** — consistent composition API pattern across all Pinia stores
- **Connector architecture** — interface hierarchy (Connector → StreamConnector/QueryConnector) is clean
- **Auth flow** — Supabase JWT → middleware → RLS chain is well-integrated
- **Multi-org data model** — `org_members` junction table with role-based access is architecturally sound
