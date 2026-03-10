# Phase 8 — Pre-Push Issues

Post-implementation review of Phase 8 (Monitoring Mode). All builds and tests pass (`go vet`, `go test`, `vue-tsc`). Issues below are ordered by priority.

---

## P0 — Must Fix Before Push

### 1. `GetDashboardStats` ignores auth check result

**File:** `backend/internal/api/handlers/stats.go:11`

```go
userID, _ := middleware.UserIDFromContext(r.Context()) // ignores ok!
```

The legacy `/api/stats` handler discards the `ok` return value. If the Auth middleware fails to inject context (e.g., route restructuring), this passes a zero UUID to the database. Defence-in-depth requires checking `ok` and returning 401 on failure, consistent with every other handler.

**Fix:** Add `ok` check with early return, matching the pattern in all other handlers.

---

### 2. ConnectionsPage fetches all user connections, not per-app

**File:** `frontend/src/pages/ConnectionsPage.vue:17`

`store.fetchConnections()` calls `GET /api/connections` which returns **all** connections for the user, regardless of the selected application. The Dashboard page correctly uses `listConnectionsByApp(appId)`, but the Connections page still uses the pre-Phase-8 user-scoped store.

When a user switches apps in the sidebar, they still see connections from all apps — breaking the multi-app model.

**Fix:** Update `ConnectionsPage` to fetch connections via the app-scoped endpoint (`GET /api/apps/{appId}/connections`), either by updating the connections store to accept an `appId` parameter or by calling `listConnectionsByApp` directly.

---

## P1 — Should Fix Before Push

### 3. Legacy routes still active

**File:** `backend/internal/api/router.go:57-64`

The pre-multi-app endpoints are still registered:
- `GET /api/stats` — returns user-scoped stats (superseded by `GET /api/apps/{appId}/stats`)
- `GET /api/agent/config` — returns global singleton config (superseded by `GET /api/apps/{appId}/agent/config`)
- `PUT /api/agent/config` — same
- `POST /api/agent/run` — triggers agent run without app context

The frontend has moved to app-scoped endpoints. These legacy routes increase API surface area and could cause confusion. Should be removed or deprecated.

---

### 4. `UpdateConnectionStatus` error silently discarded

**File:** `backend/internal/api/handlers/connections.go:312`

```go
_ = queries.UpdateConnectionStatus(r.Context(), ...)
```

After a connection test, the status update error is silently ignored. The user sees "success" but the `status` column may remain stale. Should at minimum log the error.

---

### 5. Onboard endpoint lacks idempotency guard

**File:** `backend/internal/api/handlers/organizations.go`

`POST /onboard` can be called multiple times by the same user, creating duplicate organizations each time. Should check if `user.org_id` is already set and return the existing org (or 409) instead of creating a new one.

---

## P2 — Nice to Fix (Not Blocking)

### 6. No enum validation on mode/model fields

**File:** `backend/internal/api/handlers/applications.go:172-179`

`UpdateAppAgentConfig` accepts any string for `mode` and `model`. Invalid values (e.g., `mode: "banana"`) are silently stored. Should validate `mode` against `continuous|periodic|off` and optionally validate `model` against a known list.

---

### 7. `formatFlaggedLogs` uses byte slice instead of `strings.Builder`

**File:** `backend/internal/agent/monitor.go:202`

```go
var b []byte
b = append(b, fmt.Sprintf(...)...)
```

Functional but non-idiomatic Go. `strings.Builder` is the standard pattern for building strings in a loop. With the 50-log cap this won't cause performance issues, but should be cleaned up.

---

### 8. Dashboard error overwrites

**File:** `frontend/src/pages/DashboardPage.vue:28-32`

Multiple sequential try-catch blocks each set `fetchError.value`. If two API calls fail, only the last error message is shown. Should accumulate errors.

---

### 9. Log the `UpdateConnectionStatus` failure in TestConnection

**File:** `backend/internal/api/handlers/connections.go:312`

Related to issue #4. Replace `_ =` with a `log.Printf` call to capture the error server-side, matching the pattern used for test failures on lines 290 and 297.
