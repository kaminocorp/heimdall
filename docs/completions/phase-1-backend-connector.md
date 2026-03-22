# Phase 1 Completion — Supabase Backend Connector

Phase 1 of the Supabase connector implementation is complete. This phase introduces the backend infrastructure for polling the Supabase Management API and inserting logs into `log_buffer`, where the existing monitoring loop classifies and escalates them.

---

## What Was Built

### 1. `PollConnector` Interface

**File:** `backend/internal/connectors/connector.go`

Added a new interface alongside the existing `StreamConnector` (push-based) and `QueryConnector` (on-demand):

```go
type PollConnector interface {
    Connector
    Poll(ctx context.Context, queries *db.Queries) error
}
```

**Why:** Polling is fundamentally pull-based — the connector initiates HTTP requests on a timer. The existing `StreamConnector.Stream()` assumes push-based data flowing into a channel. `Poll()` accepts `*db.Queries` so it can call `InsertLogEntry` directly, keeping insertion logic inside the connector without an external orchestrator.

---

### 2. Supabase Connector

**File:** `backend/internal/connectors/logs/supabase.go` (new)

Struct and constructor:

| Field | Purpose |
|-------|---------|
| `config SupabaseConfig` | Parsed from connection's JSONB: `project_ref`, `access_token`, `poll_tables`, `poll_interval_secs` |
| `connectionID` | UUID of the connection row |
| `userID` | UUID of the owning user (for `InsertLogEntry`) |
| `httpClient` | Shared HTTP client with 30s timeout |
| `cursors map[string]time.Time` | Per-table in-memory cursor (microsecond precision) |

Methods:

| Method | Behaviour |
|--------|-----------|
| `NewSupabase(configJSON, connectionID, userID)` | Parses config, validates required fields, applies defaults (`poll_tables` defaults to `["postgres_logs"]`, `poll_interval_secs` defaults to 30, minimum 15) |
| `Connect(ctx)` | Validates PAT by running `SELECT 1` against the Management API analytics endpoint. Returns error on 401/403/404. |
| `Health(ctx)` | Alias for `Connect` — lightweight API ping. |
| `Poll(ctx, queries)` | For each table in `poll_tables`: queries logs since cursor, parses response, inserts into `log_buffer`, advances cursor. Continues polling other tables if one fails. |
| `Close()` | No-op (stateless HTTP client). |
| `ParsedConfig()` | Returns the parsed `SupabaseConfig` for reading settings like poll interval. |

**API endpoint called:**

```
GET https://api.supabase.com/v1/projects/{ref}/analytics/endpoints/logs.all
  ?sql=SELECT timestamp, event_message, metadata FROM {table} WHERE timestamp > '{cursor}' ORDER BY timestamp ASC LIMIT 500
```

**Insertion mapping:**

| `log_buffer` column | Value |
|---------------------|-------|
| `connection_id` | The Supabase connection's UUID |
| `source_type` | `"supabase/{table}"` e.g. `"supabase/postgres_logs"` |
| `severity` | Derived from metadata (`error_severity`, `severity`, or `level` fields) |
| `payload` | Full row as JSONB: `{ "timestamp": ..., "event_message": ..., "metadata": ... }` |
| `user_id` | The connection owner's UUID |

**Cursor management:**

- In-memory map of `table_name -> last_timestamp` (microsecond Unix).
- Initialises to `now() - 5 minutes` on first poll (avoids pulling the full 24h window).
- After each successful poll, advances to the max timestamp seen.
- Ephemeral (in-memory only). On process restart, re-polls from `now() - 5 minutes`.

**Rate limit handling:**

- Reads `X-RateLimit-Remaining` from response headers.
- If `0`, reads `X-RateLimit-Reset` (Unix timestamp) and sleeps until then (context-aware).
- On HTTP 429, returns error so the poller logs it and retries on next tick.

---

### 3. Polling Loop Manager

**File:** `backend/internal/connectors/poller.go` (new)

Manages one goroutine per active poll-based connection:

```go
type Poller struct {
    queries *db.Queries
    mu      sync.Mutex
    pollers map[uuid.UUID]context.CancelFunc
}
```

| Method | Behaviour |
|--------|-----------|
| `NewPoller(queries)` | Constructor. |
| `Start(conn, connectionID, interval)` | Launches a goroutine that calls `conn.Poll()` on a ticker. Fires once immediately, then on each tick. Stops any existing poller for the same ID first. |
| `Stop(connectionID)` | Cancels the goroutine for this connection. |
| `StopAll()` | Cancels all goroutines (called on server shutdown). |

---

### 4. Server & Startup Wiring

**File:** `backend/internal/api/handlers/server.go` (modified)

- Added `Poller *connectors.Poller` field to `Server` struct.
- Updated `NewServer()` signature to accept and store `*connectors.Poller`.

**File:** `backend/internal/api/router.go` (modified)

- Updated `NewRouter()` signature to accept and pass `*connectors.Poller`.

**File:** `backend/cmd/heimdall/main.go` (modified)

- Creates `Poller` instance after DB pool init.
- Calls `resumePollers()` to restart polling for all active Supabase connections on startup.
- Calls `poller.StopAll()` during shutdown (before `ag.Stop()` and pool close).

**Resume on startup:** `resumePollers()` queries `ListActiveConnectionsByType("supabase")`, creates a `Supabase` connector for each, and starts the polling goroutine.

---

### 5. Connection Handler Updates

**File:** `backend/internal/api/handlers/connections.go` (modified)

**TestConnection** — added `case "supabase"`:
- Creates a `Supabase` connector from the connection's config.
- Calls `Connect()` with a 10-second timeout.
- Returns success/failure with the actual error message.

**CreateConnection** — after successful creation:
- If type is `"supabase"`, creates a `Supabase` connector and starts the polling goroutine via `s.Poller.Start()`.

**DeleteConnection** — before deleting:
- Calls `s.Poller.Stop(connID)` to cancel any active polling goroutine.

---

### 6. New sqlc Query

**File:** `backend/internal/db/queries/connections.sql` (modified)

```sql
-- name: ListActiveConnectionsByType :many
SELECT * FROM connections WHERE type = $1 AND status = 'active';
```

Used by `resumePollers()` on server startup to find all Supabase connections that should be polling.

**Generated code:** `backend/internal/db/connections.sql.go` (auto-generated via `make sqlc-generate`).

---

## Files Summary

### New Files

| File | Purpose |
|------|---------|
| `backend/internal/connectors/logs/supabase.go` | Supabase Management API polling connector |
| `backend/internal/connectors/poller.go` | Goroutine-per-connection polling loop manager |

### Modified Files

| File | Change |
|------|--------|
| `backend/internal/connectors/connector.go` | Added `PollConnector` interface |
| `backend/internal/api/handlers/server.go` | Added `Poller` field, updated `NewServer()` |
| `backend/internal/api/router.go` | Updated `NewRouter()` to accept `*connectors.Poller` |
| `backend/internal/api/handlers/connections.go` | Supabase test/create/delete handling |
| `backend/cmd/heimdall/main.go` | Create poller, resume on startup, shutdown |
| `backend/internal/db/queries/connections.sql` | Added `ListActiveConnectionsByType` query |
| `backend/internal/db/connections.sql.go` | Auto-generated by sqlc |

### No Database Migrations Required

Phase 1 uses existing tables (`connections`, `log_buffer`) and JSONB config. The `type` field stores `"supabase"` as a string value — no schema changes needed.

---

## Verification

- `go build ./...` — compiles cleanly
- `go vet ./...` — no issues
- No new dependencies added (uses stdlib `net/http` for API calls)

---

## What's Next

- **Phase 2:** Frontend — add `supabase` type to `ConnectionForm`, poll table multi-select, `ConnectionCard` display
- **Phase 3:** End-to-end validation, error scenario testing, backend + frontend unit tests
- **Phase 4:** Connection creation wizard (independent of Phase 3)
