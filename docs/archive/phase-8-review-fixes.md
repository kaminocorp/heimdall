# Phase 8 — Post-Implementation Review & Fixes

Reference: [Phase 8 Implementation Plan](./phase-8-monitoring-mode.md) | [Post-MVP Roadmap](./post-mvp-roadmap.md)

---

## Summary

Phase 8 (Monitoring Mode) is feature-complete across all 11 steps. Build, vet, type-check, and all 63 agent + 29 handler tests pass cleanly. The review identified **5 issues** — 1 security fix required before production, 4 lower-severity items to address alongside or shortly after.

---

## Issue 1 — Authorization Gap on Per-App Endpoints

**Severity: Medium-High (security)**

### Problem

All `/api/apps/{appId}/*` handlers parse the `appId` from the URL and query the database directly without verifying the application belongs to the authenticated user's organization. Any authenticated user can read or modify any application's data by supplying its UUID.

**Affected handlers (all in `backend/internal/api/handlers/applications.go`):**

| Handler | Line | Risk |
|---------|------|------|
| `GetApplication` | 93 | Read another org's app metadata |
| `GetAppAgentConfig` | 110 | Read another org's agent config |
| `UpdateAppAgentConfig` | 141 | **Modify another org's agent config** |
| `GetMonitoringStatus` | 183 | Read another org's monitoring state |
| `GetAppDashboardStats` | 218 | Read another org's stats |
| `ListConnectionsByApp` | 235 | Read another org's connections |

**Root cause:** The existing connection handlers use `UserQueries` + `user_id` scoping (correct pattern). The new app handlers were written without an equivalent org-scoping check — they trust the `appId` URL parameter directly.

**Mitigation in place:** UUIDs are 128-bit random (impractical to guess), and `app_agent_config` has an RLS policy. But `applications`, `connections`, `monitoring_state`, and `log_buffer` do not have app-scoped RLS, so the application layer is the only line of defense for those tables.

### Proposed Fix

**Add an `authorizeApp` helper** that every `{appId}` handler calls before proceeding. The helper verifies the app belongs to the user's org.

**Step 1 — New sqlc query (`backend/internal/db/queries/applications.sql`):**

```sql
-- name: GetApplicationByOrgUser :one
-- Returns the application only if it belongs to the same org as the given user.
SELECT a.* FROM applications a
JOIN users u ON u.org_id = a.org_id
WHERE a.id = $1 AND u.id = $2;
```

**Step 2 — Helper function (`backend/internal/api/handlers/applications.go`):**

```go
// authorizeApp verifies the app belongs to the authenticated user's org.
// Returns the application on success, writes an HTTP error and returns nil on failure.
func (s *Server) authorizeApp(w http.ResponseWriter, r *http.Request) *db.Application {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        jsonError(w, "missing user context", http.StatusUnauthorized)
        return nil
    }

    appID, err := uuid.Parse(chi.URLParam(r, "appId"))
    if err != nil {
        jsonError(w, "invalid application id", http.StatusBadRequest)
        return nil
    }

    app, err := s.Queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
        ID:     appID,
        UserID: userID,
    })
    if err != nil {
        jsonError(w, "application not found", http.StatusNotFound)
        return nil
    }

    return &app
}
```

**Step 3 — Update each handler** to call `authorizeApp` at the top:

```go
func (s *Server) GetAppAgentConfig(w http.ResponseWriter, r *http.Request) {
    app := s.authorizeApp(w, r)
    if app == nil {
        return // error already written
    }
    // ... use app.ID instead of parsing appId again
}
```

Apply the same pattern to all 6 affected handlers.

**Step 4 — Add test** (`backend/internal/api/handlers/applications_test.go`):

- `TestGetApplication_WrongOrg` — Create a second user + org + app, verify that user 1 gets 404 when requesting user 2's app.

**Files changed:**

| Action | File |
|--------|------|
| Edit | `backend/internal/db/queries/applications.sql` |
| Regenerate | `backend/internal/db/applications.sql.go` |
| Edit | `backend/internal/api/handlers/applications.go` |
| Edit | `backend/internal/api/handlers/applications_test.go` |

---

## Issue 2 — Onboarding Handler Is Not Transactional

**Severity: Low**

### Problem

The `Onboard` handler (`backend/internal/api/handlers/organizations.go:36`) performs 4 sequential queries without a transaction:

1. `CreateOrganization` — creates the org
2. `SetUserOrg` — links the user to the org
3. `CreateApplication` — creates the default app
4. `UpsertAppAgentConfig` — creates default agent config

If step 3 fails, the user is linked to an org with no app. If step 4 fails, the app has no agent config. The user would see a broken state on reload.

### Proposed Fix

Wrap the 4 queries in a database transaction using pgxpool's `BeginTx`.

**Step 1 — Add a `WithTx` helper** (if one doesn't exist) or use `pool.BeginTx` directly:

```go
func (s *Server) Onboard(w http.ResponseWriter, r *http.Request) {
    // ... parse request, validate ...

    tx, err := s.Pool.Begin(r.Context())
    if err != nil {
        jsonError(w, "database error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(r.Context())

    qtx := s.Queries.WithTx(tx)

    org, err := qtx.CreateOrganization(r.Context(), ...)
    // ... all 4 queries use qtx ...

    if err := tx.Commit(r.Context()); err != nil {
        jsonError(w, "failed to complete onboarding", http.StatusInternalServerError)
        return
    }

    // ... write response ...
}
```

sqlc already generates a `WithTx` method on the `Queries` struct, so this requires no new infrastructure.

**Step 2 — Add test**: `TestOnboard_PartialFailure` — verify that if app creation fails (e.g., via a constraint), the org is not persisted.

**Files changed:**

| Action | File |
|--------|------|
| Edit | `backend/internal/api/handlers/organizations.go` |
| Edit | `backend/internal/api/handlers/organizations_test.go` |

---

## Issue 3 — Monitoring Prompt Uses Verbose Variant

**Severity: Low**

### Problem

The Phase 8 plan offered two prompt variants. Variant B (token-optimized) was marked as preferred because it:
- Uses fewer tokens per monitoring call
- Enforces a strict output format (`Severity: <level>`) that `parseSeverityFromResponse` can reliably parse
- Includes a decision tree that short-circuits false positives

The implementation uses Variant A (verbose), which lacks the strict format. This makes `parseSeverityFromResponse` (`loop.go:282`) rely on its heuristic keyword fallback more often. The heuristic has false-positive risk — e.g., an LLM response containing "the error rate returned to normal" would be parsed as severity `error`.

### Proposed Fix

Replace the `monitoringSystemPrompt` constant in `backend/internal/agent/prompt.go` with Variant B from the plan:

```go
const monitoringSystemPrompt = `You are Heimdall, an autonomous monitoring agent. No user is present.

Incoming logs were pre-filtered by a deterministic classifier. Routine logs are
already excluded. Each entry includes classification metadata and raw payload.

PROCESS
1. Read classifications and payloads.
2. False positives -> say so, stop. Do not investigate.
3. Real issues -> use tools (search_logs, query_database) only if raw data is
   insufficient. Produce assessment.

OUTPUT (strict format)
Assessment: <1-3 sentences: what happened, why it matters>
Severity: <info | warning | error | critical>
Action: <none | monitor | investigate>

SEVERITY
- info: notable but benign -- deploys, config changes, expected errors
- warning: may escalate -- elevated error rates, slow queries, auth failures
- error: active user-facing failure -- 5xx spikes, broken integrations
- critical: system-wide or data integrity risk -- cascading failures, breaches

RULES
- Never fabricate or speculate beyond available data.
- When uncertain, lower severity. False alarms erode trust.
- Brevity -- output feeds automated pipeline, not humans.`
```

Then tighten `parseSeverityFromResponse` to prioritize the structured `Severity: <level>` match and only fall back to heuristics if no structured match is found (current behavior is already correct for this — no code change needed if the prompt produces the structured format).

**Update the monitoring prompt test** (`TestBuildMonitoringPrompt`) to assert the new prompt content.

**Files changed:**

| Action | File |
|--------|------|
| Edit | `backend/internal/agent/prompt.go` |
| Edit | `backend/internal/agent/monitor_test.go` |

---

## Issue 4 — No Input Size Bound on Flagged Logs Sent to LLM

**Severity: Low**

### Problem

`formatFlaggedLogs` (`monitor.go:190`) concatenates up to 200 flagged logs with full raw payloads into a single string sent to the LLM. When using `PassthroughClassifier` (classifier off or fallback), ALL 200 logs are flagged. Large payloads (e.g., full HTTP request/response bodies) could produce a prompt that exceeds token limits or causes unexpectedly expensive API calls.

### Proposed Fix

Add two bounds:

**1. Per-payload truncation** — Truncate each log's raw payload to a maximum character length (e.g., 2000 chars) in `formatFlaggedLogs`:

```go
const maxPayloadChars = 2000

payload := string(cl.Log.Payload)
if len(payload) > maxPayloadChars {
    payload = payload[:maxPayloadChars] + "... [truncated]"
}
b = append(b, fmt.Sprintf("Raw payload: %s\n\n", payload)...)
```

**2. Batch cap for LLM escalation** — If more than N logs (e.g., 50) are flagged in a single cycle, only send the first N to the LLM and note the remainder in the heartbeat:

```go
const maxFlaggedForLLM = 50

escalated := flagged
if len(escalated) > maxFlaggedForLLM {
    escalated = escalated[:maxFlaggedForLLM]
    // Log that some flagged entries were deferred
}
```

This prevents runaway cost while still processing the most recent flagged logs.

**Files changed:**

| Action | File |
|--------|------|
| Edit | `backend/internal/agent/monitor.go` |
| Edit | `backend/internal/agent/monitor_test.go` |

---

## Issue 5 — Semaphore Can Block the Tick Loop

**Severity: Informational**

### Problem

In `monitorTick` (`monitor.go:60`), `sem <- struct{}{}` blocks if all 10 semaphore slots are occupied. If one `monitorApp` goroutine hangs (e.g., Claude API call with no deadline), subsequent apps in the same tick won't start. Combined with `wg.Wait()` at line 67, one hung app blocks the entire monitoring cycle.

### Proposed Fix

Pass a **deadline-scoped context** to each `monitorApp` invocation so that no single app can block indefinitely:

```go
const monitorAppTimeout = 2 * time.Minute

go func(app db.ListActiveApplicationsRow) {
    defer wg.Done()
    defer func() { <-sem }()

    appCtx, cancel := context.WithTimeout(ctx, monitorAppTimeout)
    defer cancel()
    a.monitorApp(appCtx, app)
}(app)
```

The 2-minute timeout ensures that even if the Claude API hangs, the goroutine will be cancelled and the semaphore slot released. The Claude SDK respects context cancellation, so this propagates cleanly.

**Files changed:**

| Action | File |
|--------|------|
| Edit | `backend/internal/agent/monitor.go` |

---

## Implementation Order

```
Issue 1: Authorization gap         [REQUIRED before prod]
Issue 2: Transactional onboarding  [recommended, quick win]
Issue 3: Monitoring prompt variant  [recommended, quick win]
Issue 4: Input size bounds          [recommended]
Issue 5: Per-app timeout            [nice-to-have]
```

Issues 1-3 are independent and can be done in parallel. Issue 4 and 5 are also independent of each other. All five can be done in a single PR.

---

## Estimated File Changes

| Action | File | Issue |
|--------|------|-------|
| Edit | `backend/internal/db/queries/applications.sql` | 1 |
| Regenerate | `backend/internal/db/applications.sql.go` | 1 |
| Edit | `backend/internal/api/handlers/applications.go` | 1 |
| Edit | `backend/internal/api/handlers/applications_test.go` | 1 |
| Edit | `backend/internal/api/handlers/organizations.go` | 2 |
| Edit | `backend/internal/api/handlers/organizations_test.go` | 2 |
| Edit | `backend/internal/agent/prompt.go` | 3 |
| Edit | `backend/internal/agent/monitor_test.go` | 3, 4 |
| Edit | `backend/internal/agent/monitor.go` | 4, 5 |
