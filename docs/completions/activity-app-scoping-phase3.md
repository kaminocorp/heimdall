# Activity Feed App-Scoping — Phase 3: Populate `app_id` at Write Time

## Status: Complete

## Summary

Every code path that inserts into `log_buffer` or `agent_log` now sets
`app_id` so new rows are always scoped to the correct application. This
spans the agent emit layer, all 6 log connectors, the webhook/OTLP
ingestion handlers, and the connector factory/resume infrastructure.

**Plan reference:** `docs/executing/activity-app-scoping.md`, Part 3

---

## SQL query changes

| Query | Change |
|-------|--------|
| `InsertLogEntry` | Added `app_id` as 6th parameter |
| `InsertAgentLog` | Added `app_id` as 7th parameter |

Both parameters are `pgtype.UUID` (nullable), matching the column.
`make sqlc-generate` regenerated all Go code.

---

## Agent emit layer (`emit.go`)

### Signature change

```go
// Before
func (a *Agent) EmitLog(ctx, userID, conversationID, entryType, summary, detail)
func (a *Agent) EmitLogWithSeverity(ctx, userID, conversationID, entryType, summary, detail, severity)

// After
func (a *Agent) EmitLog(ctx, userID, appID *uuid.UUID, conversationID, entryType, summary, detail)
func (a *Agent) EmitLogWithSeverity(ctx, userID, appID *uuid.UUID, conversationID, entryType, summary, detail, severity)
```

`appID` is a `*uuid.UUID` — `nil` means "no app context" (the column
stays NULL). The internal `emitLog` converts to `pgtype.UUID` for the
query param.

### Caller updates

| Caller | File | Where `appID` comes from |
|--------|------|--------------------------|
| Interactive chat loop | `loop.go` (`runConversationCore`) | `appID uuid.UUID` param → `&appID` (nil when `uuid.Nil`) |
| Monitoring tool calls | `loop.go` (`RunMonitoring`) | `&appConfig.AppID` (local copy `monAppID`) |
| Monitor assessments | `monitor.go` (`monitorApp`) | `&app.ID` |
| Scheduled investigations | `scheduler.go` (`RunScheduledInvestigation`) | `&s.AppID` |
| App created/deleted audit | `applications.go` | `nil` (org-level events survive app deletion) |

---

## `log_buffer` INSERT callers

### Webhook handler (`webhooks.go`)

`conn.AppID` (from `GetConnectionByWebhookToken` result) →
`pgtype.UUID{Bytes: conn.AppID, Valid: true}`.

### OTLP handler (`otlp.go`)

Same pattern — `conn.AppID` from the token lookup.

### Connectors (5 pollers + syslog)

Each connector struct gained an `appID uuid.UUID` field, threaded through
the constructor:

| Connector | Constructor | Struct field |
|-----------|------------|--------------|
| `Supabase` | `NewSupabase(config, connID, userID, appID)` | `s.appID` |
| `Flyio` | `NewFlyio(config, connID, userID, appID)` | `f.appID` |
| `Vercel` | `NewVercel(config, connID, userID, appID)` | `v.appID` |
| `Railway` | `NewRailway(config, connID, userID, appID)` | `r.appID` |
| `MongoDB` | `NewMongoDB(config, connID, userID, appID)` | `m.appID` |
| `Syslog` | `NewSyslog(config, connID, userID, appID, queries)` | `s.appID` |

Each `InsertLogEntry` call now includes
`AppID: pgtype.UUID{Bytes: x.appID, Valid: true}`.

### Factory (`factory.go`)

`StartPoller` signature gained `appID uuid.UUID`, passed through to each
`New*` constructor.

### Resume paths (`main.go`)

`resumePollers` and `resumeSyslogListeners` now pass `conn.AppID`
(available from `ListActiveConnectionsByType` which returns full
`Connection` rows with `AppID`).

### Connection create/update (`connections.go`)

Both the create and update handlers pass `conn.AppID` to `StartPoller`
and `NewSyslog`.

### Validation (`connections_validate.go`)

Validation-only calls pass `uuid.Nil` — no rows are inserted.

### Connection test handler (`connections_test_handler.go`)

Test-connection calls pass `conn.AppID` for constructor compatibility.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `internal/db/queries/log_buffer.sql` | Edit | `app_id` in INSERT |
| `internal/db/queries/agent_log.sql` | Edit | `app_id` in INSERT |
| `internal/db/log_buffer.sql.go` | Generated | Updated param struct |
| `internal/db/agent_log.sql.go` | Generated | Updated param struct |
| `internal/db/models.go` | Generated | `AppID` on both models |
| `internal/agent/emit.go` | Edit | `appID *uuid.UUID` param |
| `internal/agent/loop.go` | Edit | Thread `appIDPtr` to EmitLog calls |
| `internal/agent/monitor.go` | Edit | Pass `&app.ID` |
| `internal/agent/scheduler.go` | Edit | Pass `&s.AppID` |
| `internal/api/handlers/applications.go` | Edit | Pass `nil, nil` (no app, no conv) |
| `internal/api/handlers/webhooks.go` | Edit | `conn.AppID` in InsertLogEntry |
| `internal/api/handlers/otlp.go` | Edit | `conn.AppID` in InsertLogEntry |
| `internal/api/handlers/connections.go` | Edit | `conn.AppID` to StartPoller/NewSyslog |
| `internal/api/handlers/connections_validate.go` | Edit | `uuid.Nil` for validation |
| `internal/api/handlers/connections_test_handler.go` | Edit | `conn.AppID` for test calls |
| `internal/connectors/factory.go` | Edit | `appID` param on StartPoller |
| `internal/connectors/logs/supabase.go` | Edit | `appID` field + constructor + insert |
| `internal/connectors/logs/flyio.go` | Edit | Same pattern |
| `internal/connectors/logs/vercel.go` | Edit | Same pattern |
| `internal/connectors/logs/railway.go` | Edit | Same pattern + added pgtype import |
| `internal/connectors/logs/mongodb.go` | Edit | Same pattern |
| `internal/connectors/logs/syslog.go` | Edit | Same pattern |
| `internal/connectors/logs/supabase_test.go` | Edit | Extra `uuid.New()` arg |
| `internal/connectors/logs/syslog_test.go` | Edit | Extra `uuid.New()` arg |
| `cmd/heimdall/main.go` | Edit | `conn.AppID` in resume paths |

## Verification

- `go build ./...` — clean compilation
- `go test ./...` — all tests pass

---

## What's next

| Phase | Scope | Status |
|-------|-------|--------|
| 1 — Migration | Add columns, backfill, index | Done |
| 2 — Backend queries & handler | New sqlc queries, `app_id` query param | Done |
| **3 — Backend write paths** | Populate `app_id` on every INSERT | **Done** |
| 4 — Frontend | Pass `currentAppId` to API, re-fetch on app switch | Not started |
| 5 — Verification | Backend + frontend tests, manual QA | Not started |
