# Phase 9, Step 9 — Production Readiness Review & Fixes

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md)

---

## What was done

Thorough code review of the entire Phase 9 implementation (migrations, SQL queries, notifications package, handlers, router, agent wiring, frontend types/API/page/router/sidebar). Identified 5 issues and fixed all of them.

## Issues found & fixed

### 1. Context cancellation kills notification goroutine (CRITICAL)

**File:** `backend/internal/agent/monitor.go:187`

**Problem:** The notification goroutine received `ctx` — an `appCtx` with a 2-minute timeout and a deferred `cancel()`. When `monitorApp` returned after emitting the log, the context was cancelled before the goroutine could make HTTP calls to Slack/Discord/Resend. Every notification in production would have failed silently.

**Fix:** Changed `ctx` to `context.WithoutCancel(ctx)`. This creates a child context that inherits values but is not cancelled when the parent is — exactly what fire-and-forget work needs.

### 2. UTF-8 truncation bug (MINOR)

**File:** `backend/internal/notifications/slack.go:80-84`

**Problem:** `truncate()` used `len(s)` (byte length) and byte-level slicing. Multi-byte UTF-8 characters (common in Claude assessments) could be cut mid-character, producing invalid UTF-8.

**Fix:** Changed to `[]rune(s)` for length check and slicing.

### 3. Silent DB error swallowing (MINOR)

**File:** `backend/internal/notifications/notifier.go:126-139`

**Problem:** `markSent()` and `markFailed()` used `_ =` to discard DB errors. If the database was temporarily unreachable, notification statuses would be stuck at "pending" forever with no logs to debug.

**Fix:** Added `slog.Warn()` calls on error.

### 4. JSON null instead of empty array (LOW)

**File:** `backend/internal/api/handlers/notifications.go`

**Problem:** `ListNotificationChannels` and `ListNotificationHistory` returned `null` instead of `[]` when results were empty (Go nil slice marshals to `null`). Frontend expects an array.

**Fix:** Added nil-to-empty-slice coercion before encoding for both handlers.

### 5. No HTTP client timeout (LOW)

**Files:** `backend/internal/notifications/slack.go`, `backend/internal/notifications/email.go`

**Problem:** Both used `http.DefaultClient` which has no timeout. A hanging Slack/Discord/Resend endpoint could block the goroutine indefinitely.

**Fix:** Created a shared `webhookClient` with a 10-second timeout, used by both `postWebhook()` and the email channel's `Send()`.

## Files edited

| File | Change |
|------|--------|
| `backend/internal/agent/monitor.go` | `context.WithoutCancel(ctx)` for notification goroutine |
| `backend/internal/notifications/slack.go` | Rune-safe `truncate()`, `webhookClient` with 10s timeout |
| `backend/internal/notifications/email.go` | Use shared `webhookClient` instead of `http.DefaultClient` |
| `backend/internal/notifications/notifier.go` | Log errors in `markSent`/`markFailed` |
| `backend/internal/api/handlers/notifications.go` | Nil-to-empty-slice for channels and history responses |

## What was verified clean (no action needed)

- **Migrations** — schemas match plan, FKs correct, indexes present, down migrations are proper inverses
- **sqlc queries** — 12 queries across 3 files, all generate cleanly
- **Authorization** — all 8 handlers use `authorizeApp()`, channel ownership verified on update/delete
- **Input validation** — severity threshold, cooldown bounds, channel type, webhook URL prefixes
- **Frontend** — types match backend, API functions map to all endpoints, route registered, sidebar link present
- **Wiring** — dispatcher created in `main.go`, passed to agent, nil-safe check in monitor loop
- **Build** — `go build`, `go vet`, `vue-tsc`, all existing tests pass

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — all pass (agent tests re-ran due to monitor.go change)
- `vue-tsc --noEmit` — clean

---

## Post-review hardening (2026-03-11)

Second review pass identified 3 additional issues. All fixed.

### 6. `truncate()` panic on small `max` values (HIGH)

**File:** `backend/internal/notifications/slack.go:81`

**Problem:** `truncate(s, max)` computed `runes[:max-3]`. If `max < 4`, this evaluates to a negative or zero slice index, causing a runtime panic. While current callers pass large values (2000, 2900), any future caller passing a small value would crash the notification goroutine silently.

**Fix:** Added early return `if max < 4 { return s }` before the rune slicing.

### 7. Webhook URL prefix validation allows host spoofing (HIGH)

**File:** `backend/internal/api/handlers/notifications.go:367-380`

**Problem:** Validation used `strings.HasPrefix(url, "https://hooks.slack.com/")`. This is bypassable — `https://hooks.slack.com-evil.com/services/...` passes the prefix check because the string starts with the expected prefix, but the actual host is `hooks.slack.com-evil.com`. Same issue for Discord URLs.

**Fix:** Replaced prefix matching with `net/url.Parse()` + explicit host and path-prefix validation via a new `validateWebhookURL(raw, expectedHost, pathPrefix)` helper. This parses the URL structurally, checks `u.Scheme == "https"`, `u.Host == expectedHost` (exact match), and that `u.Path` starts with the expected path prefix and has content after it.

### 8. `io.ReadAll` error silently discarded in error responses (MEDIUM)

**Files:** `backend/internal/notifications/slack.go:105`, `backend/internal/notifications/email.go:74`

**Problem:** When a webhook/API returned a non-2xx status, the code read the response body for the error message: `respBody, _ := io.ReadAll(...)`. The `_` discards read errors. If the body read fails (connection reset, timeout), `respBody` is empty and the error message loses diagnostic detail.

**Fix:** Capture the read error. On failure, return a message noting the body was unreadable: `"webhook returned %d (could not read body)"`.

## Additional files edited (post-review)

| File | Change |
|------|--------|
| `backend/internal/notifications/slack.go` | `truncate()` guard for `max < 4`; `io.ReadAll` error handling in `postWebhook` |
| `backend/internal/notifications/email.go` | `io.ReadAll` error handling in `Send` |
| `backend/internal/api/handlers/notifications.go` | `validateWebhookURL()` helper using `net/url.Parse()`; replaced `strings.HasPrefix` checks for Slack/Discord URLs; added `fmt` and `net/url` imports |

## Post-review verification

- `go build ./...` — clean
- `go vet ./...` — clean
