# Phase 3 — Webhook Log Connector

First real data ingestion pipeline — logs flow into Heimdall via webhooks and are displayed in the Agent Log page with pagination and filtering.

---

## sqlc Queries

### `internal/db/queries/log_buffer.sql`

Rewrote all queries from the scaffolding originals:

| Old query | New query | Change |
|-----------|-----------|--------|
| `InsertLogEntry` (`:exec`) | `InsertLogEntry` (`:one`) | Now returns the inserted row via `RETURNING *` |
| `ListRecentLogs` | `ListLogsByUser` | Joins `connections` to scope by `user_id`, adds `LIMIT`/`OFFSET` |
| `ListLogsByConnection` | `ListLogsByUserAndConnection` | Adds `user_id` scope via join, adds pagination |
| `ListLogsBySeverity` | `ListLogsByUserAndSeverity` | Adds `user_id` scope via join, adds pagination |
| — | `CountLogsByUser` | New — total count for paginated responses |
| — | `GetConnectionByWebhookToken` | New — looks up active `webhook_logs` connection by token in `config` JSONB |
| `PruneExpiredLogs` | `PruneExpiredLogs` | Unchanged |

**Why the JOIN pattern:** The `log_buffer` table has no `user_id` column — it references `connections` via `connection_id`. To ensure a user only sees their own logs, every read query joins `log_buffer → connections` and filters on `connections.user_id`. This avoids adding a redundant `user_id` column to `log_buffer` while maintaining proper data isolation.

**Why `@webhook_token::text`:** sqlc inferred `json.RawMessage` for the `$1` parameter in `config->>'webhook_token' = $1` because the `config` column is JSONB. Using a named parameter with an explicit `::text` cast forces sqlc to generate a clean `string` parameter.

Ran `sqlc generate` — regenerated `log_buffer.sql.go`.

---

## Webhook Ingestion Endpoint

### `POST /api/webhooks/logs`

New endpoint for external services to push log data into Heimdall.

**Auth model:** Each connection of type `webhook_logs` has a `webhook_token` (32-byte hex string) stored in its `config` JSONB field. The webhook endpoint authenticates via `Authorization: Bearer <webhook_token>`. The handler calls `GetConnectionByWebhookToken` which validates the token, connection type, and active status in a single query.

**Why not JWT auth:** Log shippers (Fluentd, Vector, app HTTP loggers) don't authenticate via Supabase — they need a simple, static bearer token. The route is registered outside the `/api` group's JWT middleware and handles its own authentication.

**Payload format — single entry:**
```json
{
  "source_type": "application",
  "severity": "warning",
  "payload": { "message": "disk usage 92%", "host": "web-01" }
}
```

**Payload format — batch (array):**
```json
[
  { "source_type": "application", "severity": "info", "payload": { "message": "request completed" } },
  { "source_type": "application", "severity": "critical", "payload": { "message": "OOM killed" } }
]
```

**Validation:**
- `source_type` is required (400 if missing).
- `payload` is required and must be valid JSON (400 if missing or `null`).
- `severity` is optional — stored as `NULL` if absent.

**Response:** 201 Created with the inserted entry (single) or entries (batch).

### Webhook Token Auto-Generation

Modified `CreateConnection` in `connections.go` — when `type` is `webhook_logs`, the handler auto-generates a `webhook_token` in the config JSONB if one isn't already provided. Uses `crypto/rand` for 32 cryptographically random bytes, hex-encoded to a 64-character string.

### Router

Restructured the `/api` route group to support mixed auth. The webhook route is registered directly on the `/api` subrouter (no middleware), while all JWT-protected routes are wrapped in an inner `r.Group(...)` that applies the `Auth` middleware. This avoids the Chi routing pitfall where a parent-level route at `/api/webhooks/logs` would be swallowed by the `/api` subrouter.

### Migration `008_add_webhook_token_index`

Added a partial functional index on `connections.config->>'webhook_token'` filtered to `type = 'webhook_logs'`. This ensures the `GetConnectionByWebhookToken` query uses an index scan instead of a sequential scan on the connections table.

---

## ListLogs Handler

### `GET /api/logs`

Replaced the stub (which returned `[]`) with a real implementation.

**Query parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `severity` | string | — | Filter by severity (`info`, `warning`, `critical`) |
| `connection_id` | UUID | — | Filter by connection |
| `limit` | int | 50 | Page size (max 200) |
| `offset` | int | 0 | Pagination offset |

**Routing logic:** Uses a `switch` on which filters are present to call the appropriate sqlc query (`ListLogsByUser`, `ListLogsByUserAndSeverity`, or `ListLogsByUserAndConnection`). This avoids building dynamic SQL while keeping the query plan optimal for each case.

**Response shape:**
```json
{
  "data": [ ...log entries... ],
  "total": 142,
  "limit": 50,
  "offset": 0
}
```

---

## Frontend

### API — `api/logs.ts`

- Added `PaginatedLogs` interface matching the new backend response shape.
- Renamed `listRecentLogs` to `listLogs` — accepts `severity`, `connection_id`, `limit`, `offset` params.

### Store — `stores/logs.ts`

- Added `total`, `limit`, `offset`, `error` state refs.
- `fetchLogs` now unpacks the paginated response and handles errors.
- Added `nextPage(filters)` and `prevPage(filters)` actions for pagination.
- Added `resetPagination()` to reset offset when filters change.

### Components

**`LogFilters.vue`**
- Now accepts a `connections` prop (list of user's connections).
- Added a connection `<select>` dropdown for filtering by source.
- Emits typed filter object with optional `severity` and `connection_id`.

**`LogEntry.vue`**
- Now displays the log payload content — extracts `message` field from payload if present, otherwise shows formatted JSON.
- Added human-readable timestamp formatting via `toLocaleString()`.
- Two-line layout: metadata row (timestamp, severity, source_type) + content row (payload message).

**`LogFeed.vue`**
- Accepts new props: `connections`, `total`, `limit`, `offset`.
- Passes connections to `LogFilters`.
- Shows empty-state message when no entries found.
- Pagination controls: Previous/Next buttons with disabled states, "X–Y of Z" summary.
- Emits `next` and `prev` events for pagination.

**`AgentLogPage.vue`**
- Fetches both logs and connections on mount.
- Passes connections store data to `LogFeed` for the filter dropdown.
- Tracks active filters in a ref, resets pagination on filter change.
- Shows error banner when log fetching fails.
- Passes all pagination props and events to `LogFeed`.

---

## Files Changed

```
backend/internal/db/queries/log_buffer.sql              (rewritten)
backend/internal/db/log_buffer.sql.go                   (regenerated)
backend/internal/api/handlers/webhooks.go               (new)
backend/internal/api/handlers/logs.go                   (rewritten)
backend/internal/api/handlers/connections.go            (token auto-gen in CreateConnection)
backend/internal/api/router.go                          (restructured — public + protected groups)
backend/migrations/008_add_webhook_token_index.up.sql   (new)
backend/migrations/008_add_webhook_token_index.down.sql (new)
frontend/src/api/logs.ts                                (updated)
frontend/src/stores/logs.ts                             (updated)
frontend/src/components/log/LogFilters.vue              (updated)
frontend/src/components/log/LogFeed.vue                 (updated)
frontend/src/components/log/LogEntry.vue                (updated)
frontend/src/pages/AgentLogPage.vue                     (updated)
docs/changelog.md                                       (0.4.0 entry)
docs/completions/phase3-webhook-log-connector.md        (this file)
```
