# Activity Feed App-Scoping — Phase 2: Backend Queries & Handler

## Status: Complete

## Summary

Added app-scoped sqlc query variants for both `log_buffer` and `agent_log`,
and updated the `ListLogs` handler to accept an optional `app_id` query
parameter. When provided, the handler validates app ownership and routes to
the `*ByApp` queries; when omitted, behaviour is unchanged (user-scoped).

**Plan reference:** `docs/executing/activity-app-scoping.md`, Part 2

---

## 2a — New sqlc queries

### `log_buffer.sql` — 4 new queries

| Query | Purpose |
|-------|---------|
| `ListLogsByApp` | Paginated raw logs filtered by `user_id` + `app_id` |
| `ListLogsByAppAndSeverity` | Same, additionally filtered by `severity` |
| `CountLogsByApp` | Count for pagination totals |
| `CountLogsByAppAndSeverity` | Count with severity filter |

### `agent_log.sql` — 2 new queries

| Query | Purpose |
|-------|---------|
| `ListAgentLogByApp` | Paginated agent logs filtered by `user_id` + `app_id` |
| `CountAgentLogByApp` | Count for pagination totals |

All queries mirror the existing `*ByUser` variants with an additional
`AND app_id = $N` clause. The partial indexes from Phase 1 (`WHERE app_id
IS NOT NULL`) serve these queries directly.

### 2c — sqlc-generate

`make sqlc-generate` regenerated all Go code. The `db.LogBuffer` and
`db.AgentLog` model structs now include `AppID pgtype.UUID` (nullable).
Six new param structs and query functions were generated.

---

## 2b — Handler changes (`logs.go`)

### New query parameter: `app_id`

The `ListLogs` handler now accepts an optional `app_id` query parameter.

**Validation flow:**
1. Parse as UUID — return 400 if malformed.
2. Call `GetApplicationByOrgUser(appID, userID)` — the same query used by
   `authorizeApp`. Return 404 if the app doesn't belong to the user's org.
3. Store as `pgtype.UUID{Bytes: parsed, Valid: true}` for use in query
   param structs.

**Why not reuse `authorizeApp`?** That helper extracts `appId` from Chi
URL path params (`chi.URLParam(r, "appId")`). The logs endpoint receives
`app_id` as a query parameter, so inline validation was simpler than
refactoring the helper.

### Query routing

The fetch and count switch statements now have additional cases that check
`appID.Valid` before falling through to the user-scoped defaults:

| Condition | Raw log query | Agent log query |
|-----------|--------------|-----------------|
| `connection_id` set | `ListLogsByUserAndConnection` (unchanged) | — |
| `app_id` + `severity` | `ListLogsByAppAndSeverity` | `ListAgentLogByApp` |
| `app_id` only | `ListLogsByApp` | `ListAgentLogByApp` |
| `severity` only | `ListLogsByUserAndSeverity` | `ListAgentLogByUser` |
| neither | `ListLogsByUser` | `ListAgentLogByUser` |

The merge logic (fetch both sources → sort by timestamp → apply
offset/limit) is unchanged — only the underlying queries differ.

### Backwards compatibility

- `app_id` is optional. Omitting it produces identical behaviour to before.
- Existing API consumers (frontend, tests) are unaffected.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/db/queries/log_buffer.sql` | Edit | +4 app-scoped queries |
| `backend/internal/db/queries/agent_log.sql` | Edit | +2 app-scoped queries |
| `backend/internal/db/log_buffer.sql.go` | Generated | New query functions + param structs |
| `backend/internal/db/agent_log.sql.go` | Generated | New query functions + param structs |
| `backend/internal/db/models.go` | Generated | `AppID` field on `LogBuffer` and `AgentLog` |
| `backend/internal/api/handlers/logs.go` | Edit | `app_id` parsing, ownership check, query routing |

## Verification

- `go build ./...` — clean compilation
- `go test ./...` — all tests pass (handlers: 0.357s, agent: 0.932s)
