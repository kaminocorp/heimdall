# Phase 8, Step 5 — Monitor Loop Implementation: Completion Notes

Reference: [Phase 8 Monitoring Mode Plan](../executing/phase-8-monitoring-mode.md)

---

## Overview

Implemented the full monitoring loop that processes active applications concurrently, classifies logs through the Lumber pipeline (Step 4), and escalates flagged logs to the LLM agent. Also added the monitoring-specific system prompt, the `RunMonitoring` agent method, and wired agent lifecycle (`Start`/`Stop`) into `main.go`.

**Key principle:** The monitor runs as a long-lived goroutine with a 15-second global tick. Each tick polls all active applications, checks scheduling constraints, and processes new logs. A well-behaved production app generates only heartbeat entries — zero LLM calls.

---

## Task 1 — New DB Query: `GetFirstUserInOrg`

| Action | File |
|--------|------|
| Edited | `backend/internal/db/queries/users.sql` |
| Regenerated | `backend/internal/db/users.sql.go` |

The `agent_log` table requires a `user_id`, but monitoring is app-scoped (via org → app → connection). Added a query to resolve the first user in an organization for log attribution:

```sql
-- name: GetFirstUserInOrg :one
SELECT id FROM users WHERE org_id = $1 ORDER BY created_at ASC LIMIT 1;
```

sqlc generates this as `GetFirstUserInOrg(ctx, pgtype.UUID) (uuid.UUID, error)` — the `pgtype.UUID` parameter matches the nullable `users.org_id` column.

---

## Task 2 — Monitoring System Prompt

| Action | File |
|--------|------|
| Edited | `backend/internal/agent/prompt.go` |

Added `monitoringSystemPrompt` constant and `BuildMonitoringPrompt(override)` function. The monitoring prompt differs from the interactive prompt in several key ways:

| Interactive Prompt | Monitoring Prompt |
|-------------------|-------------------|
| "You are chatting with a user" | "You are NOT chatting with a user" |
| General-purpose investigation | Classifier-flagged log assessment |
| Open-ended response | Structured: what happened, severity, recommendation |
| No false alarm guidance | "False alarms erode trust" |

The prompt tells the LLM that a deterministic pipeline has already filtered safe logs, so what it sees has been pre-identified as unusual. This reduces unnecessary hedging and keeps assessments focused.

---

## Task 3 — EmitLog Severity Support

| Action | File |
|--------|------|
| Rewritten | `backend/internal/agent/emit.go` |

The original `EmitLog` never set the `severity` field on `InsertAgentLogParams` — it always defaulted to null. Monitoring entries need severity (`info`, `warning`, `error`, `critical`).

Refactored to:
- `EmitLog(...)` — existing signature, unchanged behavior (severity = null). All existing callers unaffected.
- `EmitLogWithSeverity(...)` — adds explicit severity string parameter.
- Both delegate to private `emitLog(...)` which handles the `pgtype.Text` conversion.

**Why not change the existing signature?** `EmitLog` is called in ~10 places in `loop.go` for interactive chat. Those entries don't have meaningful severity — they're just observations and tool results. Adding a required severity parameter to all those call sites would add noise for no value.

---

## Task 4 — `RunMonitoring` Agent Method

| Action | File |
|--------|------|
| Edited | `backend/internal/agent/loop.go` |

Added `RunMonitoring(ctx, userID, appConfig, flaggedLogs) (string, string)` — returns `(assessment, severity)`.

**Differences from `RunConversation`:**

| Aspect | `RunConversation` | `RunMonitoring` |
|--------|-------------------|-----------------|
| System prompt | Interactive prompt | Monitoring prompt |
| Config source | Global `agent_config` table | Per-app `app_agent_config` |
| History | Multi-turn conversation | Sessionless (no history) |
| Return value | `(string, error)` | `(string, string)` — assessment + severity |
| Error handling | Returns error to caller | Returns degraded assessment with severity |
| Conversation ID | Optional, for log linking | Always nil |

**Severity parsing:** After the LLM responds, `parseSeverityFromResponse(text)` extracts severity:
1. Checks for explicit `Severity: <level>` patterns (prompted in the system prompt)
2. Falls back to heuristic keyword scan (`critical` > `error` > `warning`)
3. Defaults to `info` if nothing found

**Error resilience:** If the Claude API call fails, `RunMonitoring` returns a degraded assessment string with `"error"` severity rather than propagating the error. This prevents a single API failure from crashing the monitoring loop.

---

## Task 5 — Monitor Loop (`monitor.go` Rewrite)

| Action | File |
|--------|------|
| Rewritten | `backend/internal/agent/monitor.go` |

### Architecture

```
Monitor(ctx)
  └── ticker (15s) → monitorTick(ctx, sem)
       └── ListActiveApplications()
            └── for each app (concurrent, bounded by sem):
                 shouldMonitor(ctx, app) → skip if interval not elapsed
                 monitorApp(ctx, app)
                   ├── resolveOrgUser(ctx, orgID)
                   ├── GetMonitoringState(ctx, appID) → cursor
                   ├── ListLogsSinceForApp(ctx, appID, cursor, 200)
                   ├── classifier.Classify(logs)
                   ├── EmitLogWithSeverity → heartbeat
                   ├── if flagged > 0:
                   │    ├── GetAppAgentConfig → per-app config
                   │    ├── formatFlaggedLogs → text for LLM
                   │    ├── RunMonitoring → (assessment, severity)
                   │    └── EmitLogWithSeverity → monitoring entry
                   └── UpsertMonitoringState → advance cursor
```

### Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `monitorTickInterval` | 15s | Global polling rate |
| `maxConcurrentApps` | 10 | Semaphore bound for concurrent app monitors |
| `logBatchLimit` | 200 | Max logs fetched per app per cycle |

### `shouldMonitor` Logic

| Mode | Behavior |
|------|----------|
| `continuous` | Always returns true (monitors every tick) |
| `periodic` | Checks `time.Since(lastMonitoredAt) >= scheduleIntervalSecs` |
| No state yet | Returns true (first run) |

### `monitorApp` Flow

1. **Resolve org user** — `GetFirstUserInOrg(orgID)` for agent_log attribution
2. **Get cursor** — if no `monitoring_state` row exists, initialize to `now()` and return (skip first cycle to avoid processing historical logs)
3. **Fetch logs** — `ListLogsSinceForApp` returns up to 200 logs ordered ASC
4. **Classify** — `classifier.Classify(logs)` returns `(flagged, safeCount)`
5. **Emit heartbeat** — always, with metadata `{logs_processed, safe, flagged}`
6. **Escalate** — if flagged > 0: load per-app config, format flagged logs, call `RunMonitoring`, emit monitoring entry with assessment and severity
7. **Advance cursor** — `UpsertMonitoringState` to the last log's `IngestedAt`

### `formatFlaggedLogs`

Builds structured text for the LLM containing:
- Application name
- Count of flagged logs
- Per-log: classification type/category, confidence, severity, summary, timestamp, raw payload

### Error Handling

Every step in `monitorApp` handles errors independently via logging + early return. No single failure propagates to crash the monitoring loop or affect other applications.

---

## Task 6 — Agent Lifecycle (`Start`/`Stop`)

| Action | File |
|--------|------|
| Rewritten | `backend/internal/agent/agent.go` |

### Agent Struct Changes

Added `cancel context.CancelFunc` and `wg sync.WaitGroup` fields for goroutine lifecycle management.

### `Start(ctx)`

Creates a child context with cancel, launches `Monitor(ctx)` in a goroutine tracked by `wg`.

### `Stop()`

Calls `cancel()` to signal the monitoring goroutine, then `wg.Wait()` to block until it exits cleanly.

---

## Task 7 — Wire into `main.go`

| Action | File |
|--------|------|
| Edited | `backend/cmd/heimdall/main.go` |

### Startup Sequence

```go
ag := agent.New(db.New(pool), cfg, classifier)
ag.Start(context.Background())  // ← NEW: launches monitoring goroutine
router := api.NewRouter(cfg, pool, ag, jwks)
```

### Shutdown Sequence

```
1. HTTP server shutdown (10s timeout)
2. ag.Stop()          ← NEW: cancel monitoring, wait for exit
3. classifier.Close()
4. pool.Close() (deferred)
```

**Ordering matters:** The agent must stop before the classifier closes (agent may be mid-classification) and before the pool closes (agent may be mid-query).

---

## Task 8 — Tests

| Action | File | Test Count |
|--------|------|------------|
| Created | `backend/internal/agent/monitor_test.go` | 16 tests |

### Test Summary

| Test | What It Covers |
|------|----------------|
| `TestFormatFlaggedLogs` | Single log formatting: app name, classification metadata, timestamp, raw payload |
| `TestFormatFlaggedLogs_MultipleLogs` | Multi-log formatting: correct numbering, all logs present |
| `TestParseSeverityFromResponse` (9 subtests) | Explicit `Severity: X` parsing, heuristic keyword fallback, priority ordering, default to info |
| `TestShouldMonitor_ContinuousMode` | Continuous mode always returns true |
| `TestShouldMonitor_PeriodicMode_NoState` | No monitoring state → should monitor (first run) |
| `TestRunMonitoring_SimpleResponse` | Simple end_turn response with severity extraction |
| `TestRunMonitoring_WithToolUse` | Tool use flow: tool_use → tool_result → end_turn, 2 API calls |
| `TestRunMonitoring_MaxIterations` | Infinite tool loop capped at maxIterations, returns degraded assessment |
| `TestMonitorStartStop` | Monitor goroutine starts and exits cleanly on context cancellation |
| `TestMonitorTick_NoApps` | DB error on ListActiveApplications → no panic, logs error |
| `TestBuildMonitoringPrompt` (2 subtests) | Default prompt, prompt with override appended |

### Test Patterns

- **Mock Anthropic API** — reuses `httptest.Server` + `newTestAgent` from `loop_test.go`
- **DB stubs** — reuses `stubDBTX` from `loop_test.go` which makes all DB operations fail gracefully
- **No ONNX needed** — monitor tests use `PassthroughClassifier`, testing pipeline logic without model files

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./internal/agent/ -v` — **63 tests passing** (47 existing + 16 new)
- No regressions in existing `loop_test.go`, `tools_test.go`, `classifier_test.go`, `extract_test.go`, or `severity_gate_test.go`

---

## Files Summary

| Action | File | Task |
|--------|------|------|
| Edited | `backend/internal/db/queries/users.sql` | 1 |
| Regenerated | `backend/internal/db/users.sql.go` | 1 |
| Edited | `backend/internal/agent/prompt.go` | 2 |
| Rewritten | `backend/internal/agent/emit.go` | 3 |
| Edited | `backend/internal/agent/loop.go` | 4 |
| Rewritten | `backend/internal/agent/monitor.go` | 5 |
| Rewritten | `backend/internal/agent/agent.go` | 6 |
| Edited | `backend/cmd/heimdall/main.go` | 7 |
| Created | `backend/internal/agent/monitor_test.go` | 8 |

---

## Agent Log Entry Types (Updated)

| entry_type | Source | Meaning |
|------------|--------|---------|
| `observation` | Interactive chat | Agent's response in chat |
| `tool_call` | Both | Agent invoked a tool |
| `tool_result` | Both | Tool returned a result |
| `monitoring` | Monitor loop | Agent's assessment of flagged logs (has severity) |
| `heartbeat` | Monitor loop | Monitor processed N logs, M safe, K flagged (severity: info) |
