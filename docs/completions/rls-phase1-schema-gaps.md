# RLS Phase 1 — Schema Gaps (Completed)

**Date:** 2026-02-26
**Scope:** Add `user_id` columns to `investigations` and `log_buffer` so all user-scoped tables can support RLS policies in later phases.

---

## Why

RLS policies enforce row ownership via `WHERE user_id = ...`. Two tables were missing `user_id`:

- **`investigations`** — had no user scoping at all (queries were global).
- **`log_buffer`** — scoped indirectly via `JOIN connections` (which has `user_id`), but RLS needs a direct column on the table itself.

---

## What Changed

### Migrations

| File | Purpose |
|------|---------|
| `backend/migrations/011_add_user_id_to_investigations.up.sql` | Adds `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE` + index. Uses a temporary default for existing rows, then drops the default. |
| `backend/migrations/011_add_user_id_to_investigations.down.sql` | Drops index + column. |
| `backend/migrations/012_add_user_id_to_log_buffer.up.sql` | Adds `user_id UUID` (nullable first), backfills from `connections.user_id` via JOIN, then sets `NOT NULL` + index. |
| `backend/migrations/012_add_user_id_to_log_buffer.down.sql` | Drops index + column. |

### sqlc Queries

**`backend/internal/db/queries/investigations.sql`** — every query now includes `user_id`:
- `CreateInvestigation` — inserts `user_id` as a 7th param.
- `GetInvestigation` — `WHERE id = $1 AND user_id = $2`.
- `ListOpenInvestigations` — `WHERE user_id = $1 AND status IN (...)`.
- `ListInvestigationsByDateRange` — `WHERE user_id = $1 AND started_at BETWEEN ...`.
- `UpdateInvestigationFindings` — `WHERE id = $1 AND user_id = $6`.
- `ResolveInvestigation` — `WHERE id = $1 AND user_id = $3`.
- `DismissInvestigation` — `WHERE id = $1 AND user_id = $3`.

**`backend/internal/db/queries/log_buffer.sql`** — replaced JOIN-based scoping with direct `user_id` filtering:
- `InsertLogEntry` — now inserts `user_id` as a 5th param.
- `ListLogsByUser` — `WHERE user_id = $1` (was `JOIN connections`).
- `ListLogsByUserAndSeverity` — `WHERE user_id = $1 AND severity = $2` (was JOIN).
- `ListLogsByUserAndConnection` — `WHERE user_id = $1 AND connection_id = $2` (was JOIN).
- `CountLogsByUser` — `WHERE user_id = $1` (was JOIN).
- `GetConnectionByWebhookToken`, `PruneExpiredLogs` — unchanged.

### Generated Code (sqlc)

Ran `sqlc generate` — updated:
- `backend/internal/db/models.go` — `Investigation` and `LogBuffer` structs now include `UserID uuid.UUID`.
- `backend/internal/db/investigations.sql.go` — regenerated with new param structs.
- `backend/internal/db/log_buffer.sql.go` — regenerated; queries no longer JOIN to connections.

### Handler Fix

**`backend/internal/api/handlers/webhooks.go`** — added `UserID: conn.UserID` to the `InsertLogEntryParams` struct in `IngestWebhookLogs`. The user ID comes from the connection record (already looked up via webhook token).

No changes needed in `logs.go` or `agent/tools_logs.go` — the param struct field names (`UserID`, `Limit`, `Offset`, etc.) remained identical, so existing call sites compiled without modification.

---

## What Did NOT Change

- **No handler changes for investigations** — the monitoring loop (`agent/monitor.go`) is still a stub; no code currently creates or queries investigations at runtime.
- **`agent_config`** — system-wide single-row table, no user scoping needed.
- **`users`** — identity table, will get its own RLS policy in Phase 3 but needs no schema change.

---

## Verification

- `sqlc generate` — clean, no errors.
- `go build ./...` — compiles successfully.
- Models confirmed: `Investigation.UserID` and `LogBuffer.UserID` present in generated `models.go`.
