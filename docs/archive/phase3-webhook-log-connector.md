# Phase 3 — Webhook Log Connector

Reference: [MVP Roadmap](./mvp-roadmap.md) · [Blueprint](../plans/blueprint.md)

**Goal:** Get real data flowing into Heimdall. After this phase, a user can send logs to a webhook endpoint, and browse them (paginated, filterable) in the Agent Log page.

---

## Design Decisions

### Webhook authentication

The webhook endpoint (`POST /api/webhooks/logs`) lives **outside** JWT auth middleware — external services (app servers, log shippers) won't hold a Supabase JWT.

Instead, each connection of type `webhook_logs` gets a **webhook token** (a random 32-byte hex string) stored in its `config` JSONB. The webhook endpoint authenticates via:

```
POST /api/webhooks/logs
Authorization: Bearer <webhook_token>
```

The handler looks up the connection by token, confirms it exists and is active, then inserts the log entry. This keeps the auth model simple and scoped per-connection.

### User-scoped log reads

`GET /api/logs` sits behind JWT auth. Logs belong to connections, and connections belong to users. The query joins `log_buffer → connections` to filter by `user_id`, ensuring a user only sees logs from their own connections.

### Pagination

All list queries use cursor-style pagination isn't needed at MVP — simple `LIMIT`/`OFFSET` with a default page size of 50 is sufficient. The frontend sends `?limit=50&offset=0`.

---

## Task 1 — sqlc Queries (user-scoped, paginated, filterable)

Rewrite `log_buffer.sql` to support the real query patterns needed by the API.

### Changes

**`backend/internal/db/queries/log_buffer.sql`** — replace all queries:

```sql
-- name: InsertLogEntry :one
INSERT INTO log_buffer (connection_id, source_type, severity, payload)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListLogsByUser :many
SELECT lb.* FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1
ORDER BY lb.ingested_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLogsByUserAndSeverity :many
SELECT lb.* FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1 AND lb.severity = $2
ORDER BY lb.ingested_at DESC
LIMIT $3 OFFSET $4;

-- name: ListLogsByUserAndConnection :many
SELECT lb.* FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1 AND lb.connection_id = $2
ORDER BY lb.ingested_at DESC
LIMIT $3 OFFSET $4;

-- name: CountLogsByUser :one
SELECT count(*) FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1;

-- name: GetConnectionByWebhookToken :one
SELECT * FROM connections
WHERE config->>'webhook_token' = $1 AND type = 'webhook_logs' AND status = 'active';

-- name: PruneExpiredLogs :execrows
DELETE FROM log_buffer WHERE ingested_at < now() - interval '24 hours';
```

**Key changes from scaffolding queries:**
- All read queries join through `connections` to scope by `user_id`
- `LIMIT`/`OFFSET` for pagination
- `InsertLogEntry` now returns the inserted row (`:one` instead of `:exec`)
- New `GetConnectionByWebhookToken` for webhook auth
- New `CountLogsByUser` to support total count in paginated responses
- Removed the hardcoded `1 hour` time window — let the client decide what to fetch

Then run:

```bash
cd backend && sqlc generate
```

### Files changed

```
backend/internal/db/queries/log_buffer.sql    (rewritten)
backend/internal/db/log_buffer.sql.go         (regenerated)
backend/internal/db/models.go                 (regenerated — should be unchanged)
```

---

## Task 2 — Backend: Webhook Ingestion Endpoint

Create the `POST /api/webhooks/logs` handler and register it outside the JWT middleware.

### Webhook token generation

When a connection of type `webhook_logs` is created via the existing `POST /api/connections`, the handler should auto-generate a `webhook_token` in the config if one isn't provided. This is a small addition to `CreateConnection`:

```go
// In CreateConnection handler, after parsing request:
if req.Type == "webhook_logs" {
    cfg := make(map[string]interface{})
    json.Unmarshal(config, &cfg)
    if _, ok := cfg["webhook_token"]; !ok {
        cfg["webhook_token"] = generateToken() // crypto/rand, 32 bytes hex
        config, _ = json.Marshal(cfg)
    }
}
```

### Handler: `IngestWebhookLogs`

**`backend/internal/api/handlers/webhooks.go`** (new file):

```go
func (s *Server) IngestWebhookLogs(w http.ResponseWriter, r *http.Request)
```

Flow:
1. Extract `Authorization: Bearer <token>` from header.
2. Call `GetConnectionByWebhookToken(ctx, token)` — validates connection exists, is type `webhook_logs`, and is active.
3. Parse request body — expects JSON:
   ```json
   {
     "source_type": "application",
     "severity": "warning",
     "payload": { "message": "disk usage 92%", "host": "web-01" }
   }
   ```
   Also accept a batch variant (array of entries).
4. Call `InsertLogEntry` for each entry with the resolved `connection_id`.
5. Return `201 Created` with the inserted entry (or entries).

**Validation:**
- `source_type` is required (400 if missing).
- `payload` is required and must be valid JSON (400 if missing).
- `severity` is optional — defaults to `null`.

### Router

Add the webhook route **outside** the `/api` group (no JWT middleware):

```go
// In router.go, after the /api group:
r.Post("/api/webhooks/logs", s.IngestWebhookLogs)
```

Because this is registered on the root router (before the `/api` group's `r.Use(middleware.Auth(...))`), it won't require a JWT. The handler does its own token-based auth.

### Files changed

```
backend/internal/api/handlers/webhooks.go     (new)
backend/internal/api/handlers/connections.go   (token auto-generation in CreateConnection)
backend/internal/api/router.go                 (new route)
```

---

## Task 3 — Backend: ListLogs Handler

Replace the stub `ListLogs` handler with a real implementation.

### Handler: `ListLogs`

**`backend/internal/api/handlers/logs.go`** — rewrite:

```go
func (s *Server) ListLogs(w http.ResponseWriter, r *http.Request)
```

Flow:
1. Extract `userID` from JWT context (401 if absent).
2. Parse query parameters:
   - `severity` (optional) — filter by severity level
   - `connection_id` (optional) — filter by specific connection
   - `limit` (optional, default 50, max 200)
   - `offset` (optional, default 0)
3. Call the appropriate sqlc query based on which filters are present:
   - Both empty → `ListLogsByUser(userID, limit, offset)`
   - Severity set → `ListLogsByUserAndSeverity(userID, severity, limit, offset)`
   - Connection set → `ListLogsByUserAndConnection(userID, connID, limit, offset)`
4. Call `CountLogsByUser(userID)` for the total count.
5. Return JSON response:
   ```json
   {
     "data": [ ...log entries... ],
     "total": 142,
     "limit": 50,
     "offset": 0
   }
   ```

### Files changed

```
backend/internal/api/handlers/logs.go          (rewritten)
```

---

## Task 4 — Frontend: Wire Up Log Display

The frontend scaffolding (types, store, API client, components, page) already exists. Update it to work with the real API response shape and add pagination + connection filtering.

### API layer

**`frontend/src/api/logs.ts`** — update response type:

```ts
export interface PaginatedLogs {
  data: LogEntry[]
  total: number
  limit: number
  offset: number
}

export function listRecentLogs(params?: {
  severity?: string
  connection_id?: string
  limit?: number
  offset?: number
}) {
  return client.get<PaginatedLogs>('/logs', { params })
}
```

### Store

**`frontend/src/stores/logs.ts`** — add pagination state:

```ts
const entries = ref<LogEntry[]>([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)
```

Update `fetchLogs` to unpack the paginated response and expose `nextPage` / `prevPage` actions.

### Components

**`LogFilters.vue`** — add a connection dropdown:
- Accept a `connections` prop (list of user's connections).
- Add a `<select>` for filtering by connection.
- Emit both `severity` and `connection_id` in the filter event.

**`LogFeed.vue`** — add pagination controls:
- Show "Showing X–Y of Z" text.
- Previous / Next buttons.
- Emit `page` event when navigating.

**`LogEntry.vue`** — display the payload:
- Show key fields from the `payload` JSONB (e.g. `message`, or a formatted JSON preview).
- Currently only shows timestamp, severity, and source_type — the actual log content is missing.

**`AgentLogPage.vue`** — wire up:
- Fetch connections on mount (for the filter dropdown).
- Pass connections to `LogFilters`.
- Handle pagination events.
- Show error state.

### Files changed

```
frontend/src/api/logs.ts                          (updated)
frontend/src/stores/logs.ts                       (updated)
frontend/src/components/log/LogFilters.vue        (updated)
frontend/src/components/log/LogFeed.vue           (updated)
frontend/src/components/log/LogEntry.vue          (updated)
frontend/src/pages/AgentLogPage.vue               (updated)
```

---

## Task 5 — Verification & Changelog

### Manual verification

1. **Create a webhook connection:**
   ```bash
   curl -X POST http://localhost:8080/api/connections \
     -H "Authorization: Bearer <jwt>" \
     -H "Content-Type: application/json" \
     -d '{"name": "My App Logs", "type": "webhook_logs"}'
   ```
   Confirm the response includes a `webhook_token` in `config`.

2. **Send a log entry via webhook:**
   ```bash
   curl -X POST http://localhost:8080/api/webhooks/logs \
     -H "Authorization: Bearer <webhook_token>" \
     -H "Content-Type: application/json" \
     -d '{"source_type": "application", "severity": "warning", "payload": {"message": "disk usage 92%"}}'
   ```
   Confirm 201 response.

3. **List logs via API:**
   ```bash
   curl http://localhost:8080/api/logs?limit=10 \
     -H "Authorization: Bearer <jwt>"
   ```
   Confirm the log entry appears in the response, scoped to the user.

4. **Frontend:** Open the Agent Log page, confirm logs appear with filtering and pagination.

### Changelog entry

Add `0.4.0 — Webhook Log Connector` to `docs/changelog.md`.

### Completion doc

Write `docs/completions/phase3-webhook-log-connector.md` summarising what was done.

### Files changed

```
docs/changelog.md                                        (updated)
docs/completions/phase3-webhook-log-connector.md         (new)
```

---

## Dependencies

```
Task 1: sqlc queries
    └──▶ Task 2: Webhook ingestion endpoint  (needs GetConnectionByWebhookToken)
    └──▶ Task 3: ListLogs handler            (needs user-scoped queries)
              └──▶ Task 4: Frontend wiring   (needs real API responses)
                        └──▶ Task 5: Verification & changelog
```

Tasks 2 and 3 can run in parallel after Task 1.

---

## Files Summary

```
backend/internal/db/queries/log_buffer.sql              (rewritten)
backend/internal/db/log_buffer.sql.go                   (regenerated)
backend/internal/api/handlers/webhooks.go               (new)
backend/internal/api/handlers/logs.go                   (rewritten)
backend/internal/api/handlers/connections.go            (token auto-gen)
backend/internal/api/router.go                          (new webhook route)
frontend/src/api/logs.ts                                (updated)
frontend/src/stores/logs.ts                             (updated)
frontend/src/components/log/LogFilters.vue              (updated)
frontend/src/components/log/LogFeed.vue                 (updated)
frontend/src/components/log/LogEntry.vue                (updated)
frontend/src/pages/AgentLogPage.vue                     (updated)
docs/changelog.md                                       (updated)
docs/completions/phase3-webhook-log-connector.md        (new)
```
