# Production Readiness Fixes — v0.42.19 Assessment

**Date:** 2026-04-15
**Source:** Comprehensive code assessment across all subsystems
**Current score:** 8.1/10
**Target score:** 8.5+/10

---

## HIGH — Must fix before production

### H1. OrgOverviewPage shows stale apps after org switch

**File:** `frontend/src/pages/org/OrgOverviewPage.vue`
**Problem:** `fetchApps()` runs in `onMounted` only. If the user switches orgs via the dropdown while already on `/org`, Vue reuses the component instance without re-mounting. The page shows the previous org's applications until a manual refresh.
**Fix:** Add `watch(() => appStore.organization?.id, () => fetchApps(), { flush: 'post' })` to refetch when the org changes.

---

### H2. OrgTeamPage shows stale members after org switch

**File:** `frontend/src/pages/org/OrgTeamPage.vue`
**Problem:** Same root cause as H1 — `fetchMembers()` runs in `onMounted` only, no watcher on org context.
**Fix:** Add `watch(() => appStore.organization?.id, () => fetchMembers(), { flush: 'post' })`.

---

### H3. 401 interceptor race window

**File:** `frontend/src/api/client.ts` (lines 41–52)
**Problem:** `isLoggingOut` is reset in the `finally` block *before* `window.location.href = '/login'` executes on the next line. Other in-flight 401 responses can re-enter the handler during the brief async gap, producing duplicate `auth.logout()` calls.
**Fix:** Move `window.location.href = '/login'` inside the `try` block (after `await auth.logout()`), or into the `finally` block before resetting the flag. Alternatively, never reset `isLoggingOut` — the page is about to reload anyway.

---

### H4. Postgres connector `cfg.Host` not URL-escaped

**File:** `backend/internal/connectors/database/postgres.go` (line 64)
**Problem:** Host is interpolated directly into the connection string via `fmt.Sprintf`. A crafted host value like `evil.com?default_transaction_read_only=off&host=` could inject query parameters overriding the read-only guard. While this is user-controlled config (users can only affect their own DB), it undermines the defence-in-depth posture established by the SSLMode allowlist and SQL validation.
**Fix:** Validate `cfg.Host` against a hostname/IP regex, or use `url.PathEscape(cfg.Host)`.

---

## MEDIUM — Fix soon, not blocking

### M1. `Onboard` and `CreateNewOrganization` use raw `Pool.Begin`

**File:** `backend/internal/api/handlers/organizations.go`
**Problem:** Both handlers manually replicate the `UserQueries` pattern (begin tx, set RLS variable, commit/rollback) instead of calling `UserQueries()`. They miss the `context.WithoutCancel` that `UserQueries` uses for finalization — if a client disconnects mid-commit, the deferred rollback could use a cancelled context.
**Fix:** Refactor both to use `s.UserQueries()`, matching every other write handler.

---

### M2. `CreateConnection` org-auth check outside transaction

**File:** `backend/internal/api/handlers/connections.go`
**Problem:** `GetApplicationByOrgUser` runs against `s.Queries` (un-transacted), then the actual insert runs inside a `UserQueries` transaction. TOCTOU gap: a user could be removed from the org between the auth check and the write. RLS mitigates this, but the pattern differs from `DeleteApplication` which correctly does auth inside the transaction.
**Fix:** Move the `GetApplicationByOrgUser` check inside the `UserQueries` transaction.

---

### M3. Chat handler manual `done()` calls

**File:** `backend/internal/api/handlers/chat.go` (lines 86–131)
**Problem:** The conversation-setup transaction uses manual `done()` calls in each early-return branch instead of `defer done()`. Fragile — a future modification could forget `done()` in a new branch, leaking a transaction. The reason is the handler needs to release the transaction before entering the long-lived WebSocket message loop.
**Fix:** Extract the conversation-setup logic into a helper function where `defer done()` works naturally, returning the setup result to the outer handler.

---

### M4. LoginPage misattributes init failure as auth failure

**File:** `frontend/src/pages/LoginPage.vue`
**Problem:** After successful `auth.login()`, `appStore.init()` is called. If `init()` throws (network error, 500), the catch block shows "Authentication failed" — misleading since authentication succeeded.
**Fix:** Separate the error handling: catch auth errors from `login()` with an auth-specific message, and catch init errors with something like "Logged in but failed to load your workspace".

---

### M5. ProfileDropdown logout error handling

**File:** `frontend/src/components/common/ProfileDropdown.vue`
**Problem:** `handleLogout` calls `app.reset()` then `auth.logout()`. If `logout()` throws, app state is already cleared but the user stays on the current page in a partially reset state. Also, `auth.logout()` triggers `onAuthStateChange` which calls `app.reset()` again (harmless but redundant).
**Fix:** Wrap in try/catch. On failure, force-navigate to `/login` anyway since state is already cleared.

---

### M6. WebSocket `updateOptions` status flash

**File:** `frontend/src/composables/useWebSocket.ts`
**Problem:** When `updateOptions` closes the old socket and opens a new one, the old socket's async `onclose` fires after the new socket is already connecting, briefly setting `status = 'closed'`. The `useAgent` watcher on status resets `isThinking` and `activeTools` on `'closed'`, causing a visual glitch during token refresh.
**Fix:** Guard the `onclose` handler — skip `status = 'closed'` when `intentionalClose` is true, or use a connection generation counter.

---

### M7. Fragile org context detection in router

**File:** `frontend/src/router/index.ts`
**Problem:** `DefaultLayout.vue` uses `route.path.startsWith('/org')` to switch between `OrgSidebar` and `AppSidebar`. Any future route starting with `/org` (e.g. `/organic`) would incorrectly trigger the org sidebar.
**Fix:** Use `route.meta.context === 'org'` instead, adding a `meta` field to org routes.

---

### M8. ILIKE wildcard characters not escaped

**File:** `backend/internal/db/queries/log_buffer.sql` — `SearchLogsByUser`
**Problem:** The `payload::text ILIKE '%' || @query::text || '%'` pattern doesn't escape `%` and `_` in the user's search query. A search for literal `%` returns wildcard matches. Not a security issue (parameterized), but a functional correctness issue.
**Fix:** Escape `%` and `_` in the query parameter before passing to ILIKE, or add `ESCAPE '\'` and pre-escape the input in the Go handler.
