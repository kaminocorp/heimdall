# Phase 9, Step 6 — Wire Dispatcher into Monitor Loop

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md) (Steps 9, 11, 12)

---

## What was done

Wired the notification dispatcher into the monitoring pipeline. The agent now dispatches notifications after emitting monitoring entries to `agent_log`.

## Files edited

| File | Change |
|------|--------|
| `backend/internal/agent/agent.go` | Added `notifier *notifications.Dispatcher` field, updated `New()` to accept it |
| `backend/internal/agent/emit.go` | Changed `EmitLog`, `EmitLogWithSeverity`, and `emitLog` to return `uuid.UUID` (the inserted agent_log ID) |
| `backend/internal/agent/monitor.go` | Captures `logEntryID` from `EmitLogWithSeverity`, calls `a.notifier.Notify()` in a goroutine |
| `backend/cmd/heimdall/main.go` | Creates `notifications.Dispatcher`, passes it to `agent.New()` |

## Integration flow

```
monitor.go: monitorApp()
    │
    ├─ RunMonitoring() → assessment, severity
    ├─ EmitLogWithSeverity() → logEntryID (uuid.UUID)
    │
    └─ go a.notifier.Notify(ctx, appID, logEntryID, appName, severity, summary, assessment)
         │
         └─ (fire-and-forget goroutine — see Step 5 for dispatcher internals)
```

## Design decisions

- **`emitLog` returns `uuid.UUID`** — Changed from no return value. Uses `uuid.Nil` as the zero value on failure (Go convention for value types). Existing callers that ignore the return continue to work unchanged.
- **Nil-safe notifier check** — `if a.notifier != nil` guards the call so tests that construct an Agent without a notifier don't panic. Also skips dispatch if the log entry failed to insert (`logEntryID != uuid.Nil`).
- **Goroutine for dispatch** — The `go a.notifier.Notify(...)` call prevents notification HTTP calls (Resend, Slack, Discord) from blocking cursor advancement and next-app processing. The dispatcher handles its own error logging internally.
- **`queries` extracted to variable** — In `main.go`, `db.New(pool)` is now stored in a `queries` variable shared by both the dispatcher and the agent, avoiding duplicate pool wrappers.

## Tests

- `go test ./internal/agent/...` passes — existing tests construct Agent without a notifier (nil), and the nil-safe check prevents panics.

## Next step

Step 7 — API endpoints for notification CRUD (`handlers/notifications.go`, `router.go`).
