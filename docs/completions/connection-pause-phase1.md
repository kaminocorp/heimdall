# Connection Pause — Phase 1 Completion

**Scope:** Backend accepts and enforces `paused` as a valid connection status.
**Plan:** `docs/executing/connection-pause.md`

---

## What Changed

### 1. Validation — accept `paused` status

**File:** `backend/internal/api/handlers/connections_validate.go:29`

`isValidConnectionStatus()` now accepts `"paused"` as a fourth valid value alongside `active`, `inactive`, and `error`.

**Why:** This is the single gate that determines which status values the API accepts. Without this change, any `PUT /api/connections/{id}` request with `"status": "paused"` would be rejected with 400.

### 2. Update handler — skip poller/listener restart when paused

**File:** `backend/internal/api/handlers/connections.go:350-375`

The `UpdateConnection` handler always stops the existing poller and listener for a connection before restarting them. The restart is now wrapped in a `if status != "paused"` guard.

**Before:** Stop poller → stop listener → start poller → start listener (unconditionally).
**After:** Stop poller → stop listener → start poller → start listener (only if not paused).

**Why:** Without this, setting a connection to `paused` via the API would still restart its poller/listener in the same request — the pause would only take effect on the next server restart (when `ListActiveConnectionsByType` would skip it). The guard makes pause immediate.

**What happens on resume:** When a user later sets the status back to `active`, the same handler runs again. This time `status != "paused"` is true, so the poller/listener starts up — resume is automatic.

### 3. Agent tool — block `query_database` on paused connections

**File:** `backend/internal/agent/tools_db.go:45-48`

Added a status check after fetching the connection and before creating the database connector. If the connection is paused, returns a descriptive error.

**Why:** The `query_database` tool fetches connections via `GetConnectionByUser` which has no status filter — it returns connections in any state. Without this check, the agent could still open database connections to a paused Postgres source, defeating the purpose of pausing it.

**How errors surface:** Agent tool errors are returned as `isError: true` tool results to Claude (per the project's existing pattern in `agent/loop.go`). The agent will see the message and can inform the user that the connection is paused rather than silently failing.

---

## What Did NOT Change (and why)

### No migration

The `status` column is `TEXT`, not a Postgres `ENUM`. Adding a new string value requires no schema change — only application-level validation, which is handled by `isValidConnectionStatus()`.

### No SQL query changes

The two queries that gate on active status — `ListActiveConnectionsByType` (`WHERE status = 'active'`) and `GetConnectionByWebhookToken` (`WHERE ... AND status = 'active'`) — already exclude any non-active status. A paused connection is automatically:
- Skipped during poller/listener resume on server startup
- Rejected for webhook and OTLP ingestion (returns 401, same as inactive)

### No webhook/OTLP handler changes

Paused webhook connections get the same 401 as any non-active connection. A distinct 403 was considered but deferred — the 401 is functional and doesn't leak that the token is valid.

### No `search_logs` tool changes

The `search_logs` agent tool queries `log_buffer` directly, not connections. Historical logs from a paused connection remain fully searchable. Pausing stops new data from arriving, but doesn't hide existing data — this is the intended behavior.

### No monitor loop changes

The monitoring loop processes logs from `log_buffer` for all active applications, regardless of which connection produced them. Pausing a connection means it stops producing new logs, but existing logs continue to be classified and assessed.

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — all packages pass (agent, handlers, connectors, connectors/logs, connectors/database, connectors/codebase)
