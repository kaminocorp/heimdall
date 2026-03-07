# Phase 8, Step 11 — Tests: Completion Notes

Reference: [Phase 8 Monitoring Mode Plan](../executing/phase-8-monitoring-mode.md)

---

## Overview

Updated the handler test infrastructure for the multi-app data model, fixed existing tests broken by the `app_id` requirement on connections, and added new integration tests for organization, application, per-app agent config, monitoring status, and app-scoped endpoints.

---

## Task 1 — Test Infrastructure Update

| Action | File |
|--------|------|
| Rewritten | `backend/internal/api/handlers/testhelpers_test.go` |

### Changes to `testSetup`

`testSetup` now creates a complete org → app → agent config hierarchy alongside the test user:

| Step | What | Why |
|------|------|-----|
| Create org | `INSERT INTO organizations` | Required for app creation |
| Link user | `UPDATE users SET org_id` | Required for org-scoped queries |
| Create app | `INSERT INTO applications` | Required for connection creation (`app_id NOT NULL`) |
| Create config | `INSERT INTO app_agent_config` | Required for per-app config tests |

### New `testEnv` fields

| Field | Type | Description |
|-------|------|-------------|
| `OrgID` | `string` | Test organization UUID |
| `AppID` | `string` | Test application UUID |

### New helper: `createTestConnection`

Creates a `webhook_logs` connection in the test app and returns its ID. Replaces the inline connection creation pattern used in many tests — reduces duplication and ensures `app_id` is always included.

### New routes in test router

All routes from `router.go` are now mirrored:
- `/api/org`, `/api/onboard`
- `/api/apps`, `/api/apps/{appId}/*`
- All existing routes preserved

### Cleanup

Org deletion (CASCADE) added before auth.users deletion to ensure clean teardown.

---

## Task 2 — Fix Existing Connection Tests

| Action | File |
|--------|------|
| Rewritten | `backend/internal/api/handlers/connections_test.go` |

All connection creation calls updated to include `app_id: env.AppID`. Uses `createTestConnection` helper where possible. Added `TestCreateConnection_MissingAppID` to verify the new validation.

| Test | Change |
|------|--------|
| `TestListConnections` | Uses `createTestConnection` helper |
| `TestCreateConnection` | Adds `app_id`, asserts `app_id` in response |
| `TestCreateConnection_MissingAppID` | **New** — verifies 400 without `app_id` |
| `TestCreateConnection_InvalidType` | Adds `app_id` to both sub-cases |
| `TestUpdateConnection` | Uses `createTestConnection` helper |
| `TestDeleteConnection` | Uses `createTestConnection` helper |

---

## Task 3 — Fix Logs & Webhook Tests

| Action | File |
|--------|------|
| Rewritten | `backend/internal/api/handlers/logs_test.go` |
| Rewritten | `backend/internal/api/handlers/webhooks_test.go` |

Both files updated to use `createTestConnection` or include `app_id` when creating connections. No test logic changes — only the connection creation calls updated.

---

## Task 4 — New Organization Tests

| Action | File |
|--------|------|
| Created | `backend/internal/api/handlers/organizations_test.go` |

| Test | What It Covers |
|------|----------------|
| `TestGetOrganization` | Returns test user's org with correct ID, name, slug |
| `TestOnboard` | Creates org + app, verifies response structure and data |
| `TestOnboard_DuplicateSlug` | Returns 409 Conflict on slug collision |
| `TestOnboard_MissingFields` | Returns 400 for missing `org_name` or `org_slug` |

---

## Task 5 — New Application & Per-App Tests

| Action | File |
|--------|------|
| Created | `backend/internal/api/handlers/applications_test.go` |

| Test | What It Covers |
|------|----------------|
| `TestListApplications` | Returns the default test app |
| `TestCreateApplication` | Creates app, verifies org_id, checks list count |
| `TestCreateApplication_MissingName` | Returns 400 for empty name |
| `TestGetApplication` | Fetches app by ID |
| `TestGetApplication_NotFound` | Returns 404 for nonexistent UUID |
| `TestGetAppAgentConfig` | Returns per-app config with defaults |
| `TestUpdateAppAgentConfig` | Updates all fields, verifies persistence via GET |
| `TestUpdateAppAgentConfig_Defaults` | Empty body uses sensible defaults |
| `TestGetMonitoringStatus` | Returns mode, interval, null last_monitored_at |
| `TestGetAppDashboardStats` | Returns app-scoped stats with connection count |
| `TestListConnectionsByApp` | Lists connections scoped to app, verifies app_id |

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — all packages pass (handler tests skip without DATABASE_URL)
- `go test ./internal/agent/` — 63 tests passing
- `go test ./internal/api/handlers/` — 28 tests compile and skip cleanly

---

## Test Count Summary

| File | Tests |
|------|-------|
| `connections_test.go` | 6 (was 5, +1 new) |
| `logs_test.go` | 3 (unchanged) |
| `webhooks_test.go` | 2 (unchanged) |
| `auth_test.go` | 1 (unchanged) |
| `agent_test.go` | 2 (unchanged, legacy config) |
| `organizations_test.go` | 4 (new) |
| `applications_test.go` | 11 (new) |
| **Total** | **29** |

---

## Files Summary

| Action | File | Task |
|--------|------|------|
| Rewritten | `backend/internal/api/handlers/testhelpers_test.go` | 1 |
| Rewritten | `backend/internal/api/handlers/connections_test.go` | 2 |
| Rewritten | `backend/internal/api/handlers/logs_test.go` | 3 |
| Rewritten | `backend/internal/api/handlers/webhooks_test.go` | 3 |
| Created | `backend/internal/api/handlers/organizations_test.go` | 4 |
| Created | `backend/internal/api/handlers/applications_test.go` | 5 |
