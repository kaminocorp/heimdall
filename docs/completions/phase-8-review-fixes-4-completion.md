# Phase 8 — Review Fixes Round 4 (Pre-Push Hardening)

**Date:** 2026-03-07
**Scope:** 3 P0 fixes identified during final comprehensive review before push, plus 1 additional UTF-8 truncation fix found during Round 5 audit.

---

## Context

A full-codebase review was conducted across all Phase 8 changes (11 implementation steps + 3 prior fix rounds). The review covered backend agent code, API handlers, database migrations/queries, and frontend changes. Three critical issues were identified and fixed immediately; remaining lower-priority items documented for Phase 9 prep.

---

## Fixes Applied

### Fix 1: `emitLog` Silently Drops Log on JSON Marshal Failure (P0)

**File:** `backend/internal/agent/emit.go`

**Problem:** When `json.Marshal(detail)` failed, the function returned early — silently dropping the entire log entry including summary and severity. A monitoring observation could be permanently lost if its detail map contained a non-serializable value.

**Fix:** Continue writing the log with `nil` detail instead of returning. The summary, severity, and entry type are still recorded.

```go
// Before
if err != nil {
    slog.Warn("agent emit: failed to marshal detail", "err", err)
    return  // <-- entire log entry lost
}

// After
if err != nil {
    slog.Warn("agent emit: failed to marshal detail, writing log without detail", "err", err)
    detailBytes = nil  // <-- log still written, just without detail JSON
}
```

---

### Fix 2: UTF-8 Truncation Can Corrupt Strings (P0)

**Files:** `backend/internal/agent/monitor.go`, `backend/internal/agent/loop.go`

**Problem:** `summary[:200]` and `resultSummary[:200]` slice by byte position. If truncation lands in the middle of a multi-byte UTF-8 character (e.g. non-Latin scripts, emojis in error messages), the stored string becomes invalid UTF-8. This could cause database storage failures or garbled display.

**Fix:** Use rune-safe truncation across all 4 truncation sites (1 in monitor.go, 3 in loop.go):

```go
// Before
if len(summary) > 200 {
    summary = summary[:200] + "..."
}

// After
if utf8.RuneCountInString(summary) > 200 {
    summary = string([]rune(summary)[:200]) + "..."
}
```

Added `unicode/utf8` import to both files.

---

### Fix 3: `Agent.Start()` Has No Double-Start Guard (P0)

**File:** `backend/internal/agent/agent.go`

**Problem:** If `Start()` is called twice without `Stop()`, the first goroutine's `cancel` function is overwritten and becomes unreachable. The orphaned monitoring goroutine runs indefinitely — a goroutine leak.

**Fix:** Check for existing cancel function and stop the previous instance before starting a new one:

```go
// Before
func (a *Agent) Start(ctx context.Context) {
    ctx, a.cancel = context.WithCancel(ctx)
    // ...

// After
func (a *Agent) Start(ctx context.Context) {
    if a.cancel != nil {
        slog.Warn("agent already running, stopping previous instance before restart")
        a.Stop()
    }
    ctx, a.cancel = context.WithCancel(ctx)
    // ...
```

---

### Fix 4: Remaining UTF-8 Byte-Slicing in `formatFlaggedLogs` (P0)

**File:** `backend/internal/agent/monitor.go`

**Problem:** The payload truncation in `formatFlaggedLogs` (line 213) used `len(payload)` (byte count) and byte slicing — the same class of bug fixed in Fix 2 for summary truncation. This site was missed because it operates on raw log payloads rather than agent response summaries. International log messages or payloads containing emojis/non-Latin text could be corrupted at the truncation boundary, potentially breaking the JSON sent to Claude's API.

**Fix:** Use rune-safe truncation consistent with every other truncation site in the codebase:

```go
// Before
if len(payload) > maxPayloadChars {
    payload = payload[:maxPayloadChars] + "... [truncated]"
}

// After
if utf8.RuneCountInString(payload) > maxPayloadChars {
    payload = string([]rune(payload)[:maxPayloadChars]) + "... [truncated]"
}
```

**Found during:** Round 5 comprehensive audit (2026-03-07).

---

## Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/emit.go` | Continue with nil detail on marshal failure instead of early return |
| 2 | `backend/internal/agent/monitor.go` | Added `unicode/utf8` import; rune-safe summary truncation |
| 3 | `backend/internal/agent/loop.go` | Added `unicode/utf8` import; rune-safe truncation at 3 sites |
| 4 | `backend/internal/agent/agent.go` | Double-start guard in `Start()` |
| 5 | `backend/internal/agent/monitor.go` | Rune-safe payload truncation in `formatFlaggedLogs` |

## Verification

- `go vet ./...` — clean
- `go test ./internal/agent/...` — all passing
- `vue-tsc --noEmit` — clean

---

## Deferred Items (P1/P2 — Phase 9 Prep)

These were identified during review but are not blockers for push:

| Priority | Issue | Location |
|----------|-------|----------|
| P1 | Missing org slug format validation (regex + max length) | `handlers/organizations.go` |
| P1 | No max-length on `SystemPromptOverride` | `handlers/applications.go` |
| P1 | `UpdateConnectionStatus` SQL lacks `user_id` scope (defense-in-depth) | `queries/connections.sql` |
| P1 | Missing indexes for monitoring queries (`applications(status)`, `connections(app_id, status)`) | New migration needed |
| P1 | Logs store doesn't filter by `app_id` | `frontend/src/stores/logs.ts` |
| P2 | Classifier thread-safety unverified for concurrent ONNX inference | `classifier_lumber.go` |
| P2 | Dead code: `store.fetchConnections()` calls old endpoint | `stores/connections.ts` |
| P2 | `ConnectionForm` type coercion (`'' as unknown as number`) | `ConnectionForm.vue` |
| P2 | Hardcoded model names in frontend datalist | `AgentConfigPage.vue` |
| P2 | No `disabled` attr on app selector during loading | `AppSidebar.vue` |
