# Phase 8 — Review Fixes Round 3 (Pre-Push Hardening)

Final review pass before pushing Phase 8 to production. Nine issues identified and fixed across backend and frontend.

## Verification

- `go vet ./...` — clean
- `go test ./...` — all passing
- `vue-tsc --noEmit` — no TypeScript errors

---

## Fixes

### Fix 1 — `GetDashboardStats` Auth Check (P0)

**File:** `backend/internal/api/handlers/stats.go`

The legacy `/api/stats` handler discarded the `ok` return from `UserIDFromContext`, meaning a zero UUID could be passed to the database if auth context was missing.

**Change:** Added `ok` check with 401 early return, consistent with all other handlers.

---

### Fix 2 — ConnectionsPage App-Scoped Fetching (P0)

**File:** `frontend/src/pages/ConnectionsPage.vue`

The Connections page called `store.fetchConnections()` which hit the user-scoped `GET /api/connections` endpoint — returning all connections regardless of selected app. When switching apps in the sidebar, the wrong connections were shown.

**Change:** Replaced with `listConnectionsByApp(appId)` from the applications API, matching the pattern used by DashboardPage. Added `fetchAppConnections()` function that reads `appStore.currentAppId` and sets store state directly. Also re-fetches after create/update/test operations to keep the list in sync.

---

### Fix 3 — Legacy Routes Removed (P1)

**Files:** `backend/internal/api/router.go`, `backend/internal/api/handlers/testhelpers_test.go`

Removed pre-multi-app routes that were superseded by per-app endpoints:
- `GET /api/stats` → replaced by `GET /api/apps/{appId}/stats`
- `GET /api/agent/config` → replaced by `GET /api/apps/{appId}/agent/config`
- `PUT /api/agent/config` → replaced by `PUT /api/apps/{appId}/agent/config`
- `POST /api/agent/run` → no replacement needed (monitoring loop handles this)

Removed corresponding route registrations from test helper. Deleted `agent_test.go` (legacy global config tests — per-app equivalents already exist in `applications_test.go`). Removed `TestGetDashboardStats` from `logs_test.go` (per-app equivalent in `applications_test.go`).

---

### Fix 4 — `UpdateConnectionStatus` Error Logged (P1)

**File:** `backend/internal/api/handlers/connections.go:312`

The `_ =` silent discard on `UpdateConnectionStatus` after a connection test meant status column staleness went unnoticed.

**Change:** Replaced with `log.Printf` to capture the error server-side, matching the pattern used for test failures on lines 290 and 297.

---

### Fix 5 — Onboard Idempotency Guard (P1)

**File:** `backend/internal/api/handlers/organizations.go`

`POST /onboard` could be called multiple times, creating duplicate orgs each time.

**Change:** Before starting the transaction, checks `GetUser` and returns 409 if `user.OrgID.Valid` is true. Updated three tests (`TestOnboard`, `TestOnboard_DuplicateSlug`, new `TestOnboard_AlreadyOnboarded`) to use a `createBareUser` helper that creates a user without an org, since the idempotency guard now blocks the testSetup user.

---

### Fix 6 — Mode Enum Validation (P2)

**File:** `backend/internal/api/handlers/applications.go:175`

`UpdateAppAgentConfig` accepted any string for `mode`. Invalid values like `"banana"` were stored silently.

**Change:** Added `switch` validation: only `continuous`, `periodic`, `off` are accepted. Returns 400 otherwise. Added `TestUpdateAppAgentConfig_InvalidMode` test.

---

### Fix 7 — `formatFlaggedLogs` Uses `strings.Builder` (P2)

**File:** `backend/internal/agent/monitor.go:202`

Replaced `var b []byte` + repeated `append` with idiomatic `strings.Builder` + `fmt.Fprintf`. Added `"strings"` import.

---

### Fix 8 — Dashboard Error Accumulation (P2)

**File:** `frontend/src/pages/DashboardPage.vue`

Multiple sequential try-catch blocks each overwrote `fetchError.value`. If two API calls failed, only the last error was shown.

**Change:** Errors are now accumulated into a `string[]` and joined: `"Failed to load: agent config, connections"`.

---

## Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/stats.go` | Auth check on `UserIDFromContext` |
| 2 | `backend/internal/api/handlers/connections.go` | Log `UpdateConnectionStatus` error |
| 3 | `backend/internal/api/handlers/organizations.go` | Idempotency guard on onboard |
| 4 | `backend/internal/api/handlers/applications.go` | Mode enum validation |
| 5 | `backend/internal/api/router.go` | Removed legacy routes |
| 6 | `backend/internal/agent/monitor.go` | `strings.Builder` for formatFlaggedLogs |
| 7 | `frontend/src/pages/ConnectionsPage.vue` | App-scoped connection fetching |
| 8 | `frontend/src/pages/DashboardPage.vue` | Error accumulation |

## Test Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/testhelpers_test.go` | Removed legacy route registrations |
| 2 | `backend/internal/api/handlers/organizations_test.go` | `createBareUser`/`envForUser` helpers, updated onboard tests, added `TestOnboard_AlreadyOnboarded` |
| 3 | `backend/internal/api/handlers/applications_test.go` | Added `TestUpdateAppAgentConfig_InvalidMode` |
| 4 | `backend/internal/api/handlers/logs_test.go` | Removed `TestGetDashboardStats` (superseded by app-scoped test) |

## Files Deleted

| # | File | Reason |
|---|------|--------|
| 1 | `backend/internal/api/handlers/agent_test.go` | Tests for removed legacy `/api/agent/config` routes; per-app equivalents exist in `applications_test.go` |
