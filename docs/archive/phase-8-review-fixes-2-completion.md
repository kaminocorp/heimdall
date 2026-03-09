# Phase 8 — Post-Implementation Review Fixes (Round 2)

Reference: [Phase 8 Monitoring Mode Plan](../executing/phase-8-monitoring-mode.md)

---

## Summary

Full production readiness audit of Phase 8 (11 steps + prior fixes). Four parallel code audits covered: monitor loop & agent lifecycle, API handlers & router, classifier pipeline, and frontend. ~30 raw findings triaged down to 3 real issues. All fixed.

---

## Issue 1 — CreateConnection Missing App Authorization (FIXED)

**Severity: HIGH (security)**

The `CreateConnection` handler accepted `app_id` from the request body without verifying the app belongs to the authenticated user's organization. A user who knew another org's app UUID could create a connection under it, injecting logs into a foreign org's monitoring pipeline.

**Root cause:** The connections handler predated the multi-app model. When `app_id` was added in Phase 8 Step 10, the authorization pattern from `applications.go` (`authorizeApp`) was not backported.

**Fix:** Added a `GetApplicationByOrgUser` check immediately after parsing `app_id`, before the connection is created. Uses the same query that `authorizeApp` uses — JOINs `applications` with `users` on `org_id` to verify ownership.

**Changes:**
- `backend/internal/api/handlers/connections.go` — Added authorization check in `CreateConnection` (7 lines)

---

## Issue 2 — TestConnection Error Information Leakage (FIXED)

**Severity: MEDIUM (information leakage)**

The `TestConnection` handler returned raw `err.Error()` from the Postgres connector directly to the client. Database connector errors can contain hostnames, IP addresses, credentials, or internal paths.

**Fix:** Replaced raw error messages with generic client-facing messages ("Failed to initialize database connector" / "Failed to connect to database"). Full error details are now logged server-side with the connection ID for debugging.

**Changes:**
- `backend/internal/api/handlers/connections.go` — Replaced `err.Error()` with generic messages and `log.Printf` calls in `TestConnection`; added `log` import

---

## Issue 3 — App.vue Init Has No Error Handling (FIXED)

**Severity: MEDIUM (UX)**

The `onMounted` handler in `App.vue` awaited `auth.init()` and `app.init()` without a try-catch. If either threw (e.g., network failure, Supabase outage), the app would show "Initializing..." forever with no recovery path.

**Fix:** Wrapped the initialization sequence in try-catch. On failure, logs the error and redirects to the login page so the user can retry. `ready.value = true` is always set (moved outside try block) so the loading overlay clears regardless.

**Changes:**
- `frontend/src/App.vue` — Added try-catch around init sequence with fallback redirect to login

---

## Verified Non-Issues (Triaged as False Positives)

| Flagged Finding | Why It's Fine |
|---|---|
| UTF-8 panic in `monitor.go` string truncation | Go string slicing by byte index is valid, never panics |
| Race condition on `Agent.cancel`/`wg` fields | `Start()` and `Stop()` are called sequentially at boot/shutdown |
| `nil` vs empty slice from `LumberClassifier.Classify` | `len(nil) == 0` in Go; all callers guard with `len()` |
| Model name `claude-sonnet-4-6` flagged as "typo" | Correct model ID for Claude Sonnet 4.6 |
| Default escalation for unknown types in severity gate | Intentional fail-safe per design doc |
| Router guard race condition | `App.vue` blocks `RouterView` render behind `ready` flag; `router.replace()` re-evaluates guard post-init |
| `UpdateConnection`/`TestConnection` cross-user bypass | Both scope queries by `user_id` — cross-user attacks impossible |

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./internal/agent/` — all tests passing
- `npx vue-tsc --noEmit` — clean
- `npx vitest run` — 17/17 passing

---

## Files Changed

| # | File | Issue |
|---|------|-------|
| 1 | `backend/internal/api/handlers/connections.go` | 1, 2 |
| 2 | `frontend/src/App.vue` | 3 |
