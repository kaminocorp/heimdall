# Connection Test on Create

## Summary

Added a `POST /api/connections/{id}/test` endpoint that verifies connectivity after a connection is created. The frontend calls it automatically after create and shows explicit visual feedback throughout: a pulsing "TESTING" badge on the card while the test runs, then either "ACTIVE" (green) or "ERROR" (red) with a clear error message explaining what went wrong.

## Why

Previously, every new connection was created with `status: "inactive"` and stayed that way. Users had no way to know if their credentials were valid or the target was reachable until the agent tried to use the connection — at which point the error was buried in agent logs. This left users guessing.

## Test Logic by Type

| Type | What the test does |
|------|-------------------|
| `postgres` | `database.New(config)` → `Connect(ctx)` → `Close(ctx)` with a 5-second timeout |
| `webhook_logs` | No remote target — auto-pass, set `active` |
| `syslog` | Stub connector — auto-pass, set `active` |
| `github` | Stub connector — auto-pass, set `active` |

When real connectors are implemented for syslog/github, the handler naturally gains real testing — it just needs to call `Connect()`.

## Frontend Flow

1. User fills out the form and clicks Create
2. `POST /api/connections` → connection created with `status: "inactive"`
3. `POST /api/connections/{id}/test` fires immediately
4. While testing: connection card shows a pulsing green "TESTING" badge (replaces status badge)
5. On success: badge updates to "ACTIVE"
6. On failure: badge updates to "ERROR" + error banner appears: *"Connection created but test failed: \<reason\>"*

## Files Changed

### Backend (2 files)

1. **`backend/internal/api/handlers/connections.go`** — Added `TestConnection` handler:
   - Extracts user ID from JWT context, parses connection ID from URL
   - Fetches connection via `GetConnectionByUser` (user-scoped, prevents IDOR)
   - Switches on `conn.Type`: postgres gets a real connect/close cycle with 5s timeout; all others auto-pass
   - Updates status to `active` or `error` via `UpdateConnectionStatus`
   - Returns `{ "success": true/false, "message": "..." }` — always HTTP 200 since the request itself succeeded
   - New imports: `context`, `time`, `connectors/database`

2. **`backend/internal/api/router.go`** — Registered `r.Post("/{id}/test", s.TestConnection)` inside the `/connections` route group

### Frontend (4 files)

3. **`frontend/src/api/connections.ts`** — Added `testConnection(id)` function calling `POST /connections/${id}/test`

4. **`frontend/src/stores/connections.ts`** — Added:
   - `testingId` ref — tracks which connection ID is currently being tested
   - `testConnection(id)` action — calls API, updates local connection status from response, clears `testingId` in `finally`
   - Exposed `testingId` and `testConnection` in the store return

5. **`frontend/src/pages/ConnectionsPage.vue`** — Updated `handleCreate`:
   - After `store.createConnection(payload)`, calls `store.testConnection(conn.id)`
   - On test failure, sets `actionError` to `"Connection created but test failed: <message>"`
   - Passes `store.testingId` to `ConnectionList` via `:testing-id` prop

6. **`frontend/src/components/connections/ConnectionCard.vue`** — Added `testing` prop:
   - When `true`, replaces `StatusBadge` with a pulsing green "TESTING" indicator (accent-colored pill with `animate-pulse` dot)
   - When `false`, renders normal `StatusBadge` as before

7. **`frontend/src/components/connections/ConnectionList.vue`** — Added `testingId` prop, passes `testing` boolean to each `ConnectionCard` by comparing `testingId === conn.id`
