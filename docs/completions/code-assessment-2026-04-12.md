# Code Assessment — 2026-04-12

Comprehensive codebase assessment covering backend (Go), frontend (Vue 3 + TypeScript), and database layer. Focused on functionality, accuracy, maintainability, and clean code. Nice-to-have improvements excluded.

**Status: 8/8 resolved.** All items addressed.

---

## Files Over 500 Lines (Refactor Required) — RESOLVED

| File | Before | After | Resolution |
|------|--------|-------|------------|
| `backend/internal/api/handlers/connections.go` | 693 | 395 | Split into 3 files + extracted `validateConnectorConfig` helper |
| `frontend/src/pages/NotificationsPage.vue` | 541 | 102 | Decomposed into 3 sub-components under `components/notifications/` |

---

## Backend Issues

### 1. ~~Duplicated Connection Validation Logic — HIGH~~ RESOLVED

Extracted `validateConnectorConfig()` into `connections_validate.go`, moved `TestConnection` to `connections_test_handler.go`. File reduced from 693 → 395 lines. See `docs/completions/connections-handler-refactor.md`.

### 2. ~~Unchecked json.Marshal Error — MEDIUM~~ RESOLVED

Fixed in `backend/cmd/heimdall/main.go:174`. Marshal error is now checked — logs the error with the connection ID and `continue`s to the next connection.

### 3. ~~Inconsistent Listener Start Error Handling — LOW~~ RESOLVED

The response body already included `status: "error"` (the struct is mutated before encoding), so the 201 was correct REST semantics. Fixed the real issue: `UpdateConnectionStatus` errors were silently discarded with `_ =`. Both Create and Update handlers now log these failures.

---

## Frontend Issues

### 4. ~~NotificationsPage Monolith — HIGH~~ RESOLVED

Decomposed into `NotificationPreferences.vue` (151 lines), `NotificationChannels.vue` (294 lines), and `NotificationHistory.vue` (58 lines). Page reduced from 541 → 102 lines. See `docs/completions/notifications-page-refactor.md`.

### 5. ~~Duplicated `extractApiError` Helper — MEDIUM~~ RESOLVED

Created `frontend/src/utils/apiError.ts` with a canonical `extractApiError(e: unknown, fallback)` function. Removed 2 duplicate definitions, replaced 13 inline variants across 10 files. See `docs/completions/extract-api-error-utility.md`.

### 6. ~~Direct Store State Mutation from Components — LOW~~ RESOLVED

Added `fetchConnectionsByApp(appId)` action to the connections store. `ConnectionsPage.vue` now calls `store.fetchConnectionsByApp(appId)` instead of directly mutating `store.loading/error/connections`.

### 7. ~~Untyped Error Catches — LOW~~ RESOLVED

All 13 `catch (e: any)` blocks converted to `catch (e: unknown)` with the shared `extractApiError` utility. Zero `catch (e: any)` remaining in the codebase. Fixed alongside #5.

---

## Database Issues

### 8. ~~Missing Foreign Key Index — MEDIUM~~ RESOLVED

Added migration `024_notification_log_agent_log_idx` creating the missing index on `notification_log(agent_log_id)`. Applied and verified (up/down/up) against production. See `docs/completions/notification-log-missing-index.md`.

---

## Summary

| # | Severity | Category | Item | Status |
|---|----------|----------|------|--------|
| 1 | HIGH | Backend | Duplicated validation in connections.go (693 lines) | RESOLVED — `docs/completions/connections-handler-refactor.md` |
| 4 | HIGH | Frontend | NotificationsPage.vue monolith (541 lines) | RESOLVED — `docs/completions/notifications-page-refactor.md` |
| 5 | MEDIUM | Frontend | Duplicated `extractApiError` helper | RESOLVED — `docs/completions/extract-api-error-utility.md` |
| 8 | MEDIUM | Database | Missing index on `notification_log.agent_log_id` | RESOLVED — `docs/completions/notification-log-missing-index.md` |
| 2 | MEDIUM | Backend | Unchecked `json.Marshal` in main.go:174 | RESOLVED — checked error, log + continue on failure |
| 3 | LOW | Backend | Misleading 201 response on listener start failure | RESOLVED — log `UpdateConnectionStatus` errors instead of `_ =` |
| 6 | LOW | Frontend | Direct store mutation from component | RESOLVED — added `fetchConnectionsByApp` store action |
| 7 | LOW | Frontend | Untyped `catch (e: any)` in stores | RESOLVED — fixed alongside #5 |

### What's Clean

- **Agent package** — well-structured with clear separation between interactive/monitoring/scheduled modes
- **Database layer** — sqlc queries are properly user-scoped, migrations use appropriate CASCADE behaviors, generated code is consistent
- **Auth flow** — JWT validation, RLS session variable, WebSocket token auth all properly implemented
- **Resource cleanup** — all `addEventListener`/`removeEventListener` pairs properly paired in Vue components, `done()` properly called on all paths in chat handler
- **Store architecture** — Pinia composition API used consistently, minimal prop drilling
- **Connector abstraction** — clean interface hierarchy (Connector/StreamConnector/QueryConnector)
