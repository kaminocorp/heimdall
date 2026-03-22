# Phase 3 Completion — Supabase Connector Testing

Phase 3 adds comprehensive unit tests for the Supabase connector backend (connector + poller) and frontend (ConnectionForm component), plus fixes to the test infrastructure.

---

## What Was Built

### 1. Supabase Connector Tests

**File:** `backend/internal/connectors/logs/supabase_test.go` (new — 12 tests)

| Test | What it verifies |
|------|------------------|
| `TestNewSupabase_ValidConfig` | All config fields parsed correctly from JSON |
| `TestNewSupabase_Defaults` | Missing `poll_tables` defaults to `["postgres_logs"]`, missing interval defaults to 30 |
| `TestNewSupabase_MissingProjectRef` | Returns error when `project_ref` is empty |
| `TestNewSupabase_MissingAccessToken` | Returns error when `access_token` is empty |
| `TestNewSupabase_InvalidJSON` | Returns error for malformed JSON |
| `TestNewSupabase_PollIntervalMinimum` | Interval below 15s is clamped to default (30) |
| `TestSupabase_Connect_Success` | Validates PAT by calling mock API, checks Authorization header |
| `TestSupabase_Connect_Unauthorized` | Returns error containing "401" on unauthorized response |
| `TestSupabase_Connect_NotFound` | Returns error containing "404" on project not found |
| `TestSupabase_ParseResponse` | `deriveSeverity()` correctly extracts from `error_severity`, `severity`, `level`, or returns invalid for nil/missing |
| `TestSupabase_CursorAdvancement` | Mock API returns two rows; verifies max timestamp is correctly identified |
| `TestSupabase_Close` | No-op close returns nil |
| `TestSupabase_Health` | Health check delegates to Connect successfully |
| `TestSupabase_RateLimit429` | Returns error on 429 response |

**Testing approach:** Uses `httptest.NewServer` to mock the Supabase Management API. An `apiBase` field (defaulting to the production URL) is overridden in tests to point at the mock server. This avoids hitting the real API while exercising the full HTTP path.

**Production code change:** Added `apiBase string` field to `Supabase` struct, defaulting to `supabaseAPIBase` constant. Used in `queryAPI()` instead of the hardcoded constant. Zero-cost in production (same value), enables test injection.

---

### 2. Poller Tests

**File:** `backend/internal/connectors/poller_test.go` (new — 4 tests)

| Test | What it verifies |
|------|------------------|
| `TestPoller_StartStop` | Goroutine starts, polls at least once, stops cleanly, no more polls after stop |
| `TestPoller_StopAll` | Multiple goroutines start and stop together |
| `TestPoller_StartReplacesExisting` | Starting a new connector on the same ID stops the old one |
| `TestPoller_StopNonExistent` | Stopping a non-existent ID doesn't panic |

**Testing approach:** Uses a `mockPollConnector` with an `atomic.Int32` counter. Tests use short intervals (10ms) and brief sleeps (50ms) to keep them fast (<300ms total) while verifying goroutine lifecycle.

---

### 3. Test Infrastructure Fix

**File:** `backend/internal/api/handlers/testhelpers_test.go` (modified)

The `testSetup()` function creates a `Server` struct directly. After Phase 1 added the `Poller` field to `Server`, the test helpers needed updating:

- Added `connectors` import
- Set `Poller: connectors.NewPoller(queries)` on the `Server` struct

This ensures existing integration tests (connections, logs, etc.) continue to compile and pass.

---

### 4. ConnectionForm Component Tests

**File:** `frontend/src/components/connections/__tests__/ConnectionForm.test.ts` (new — 8 tests)

| Test | What it verifies |
|------|------------------|
| `shows Supabase in the type dropdown options` | Component renders without errors |
| `renders Supabase config fields when type is supabase` | Project Reference, Personal Access Token, Poll Interval, Log Tables labels appear |
| `shows all six Supabase log tables as checkboxes` | 6 checkboxes render with correct table names |
| `defaults poll_tables to postgres_logs and auth_logs` | 2 checkboxes are checked by default |
| `auto-sets direction to one_way for Supabase` | Selecting Supabase type forces direction to `one_way` |
| `emits correct config payload with poll_tables on submit` | Submit emits payload with `poll_tables` array and `poll_interval_secs` as number |
| `shows helper text about Supabase Management API` | Info text about PAT generation appears |
| `populates selectedTables from initialValues when editing` | Editing a Supabase connection restores saved table selections |

**Testing approach:** Uses `@vue/test-utils` `mount()` with `happy-dom`. Mocks `@/api/github` and `@/stores/app` since ConnectionForm imports them. Sets component state via `wrapper.vm` for programmatic type selection.

---

### 5. Bug Fix — Edit Mode Table Selection

**File:** `frontend/src/components/connections/ConnectionForm.vue` (modified)

**Bug found during testing:** When editing a Supabase connection, the `type` watcher fired after the `initialValues` watcher, resetting `selectedTables` back to defaults and discarding the saved table selections.

**Fix:** The `type` watcher now only resets `selectedTables` to defaults when `props.initialValues` is null (i.e., creating a new connection). When editing, the `initialValues` watcher's table assignment is preserved.

---

## Files Summary

### New Files

| File | Tests |
|------|-------|
| `backend/internal/connectors/logs/supabase_test.go` | 12 tests |
| `backend/internal/connectors/poller_test.go` | 4 tests |
| `frontend/src/components/connections/__tests__/ConnectionForm.test.ts` | 8 tests |

### Modified Files

| File | Change |
|------|--------|
| `backend/internal/connectors/logs/supabase.go` | Added `apiBase` field for test injection |
| `backend/internal/api/handlers/testhelpers_test.go` | Added `Poller` field to test Server |
| `frontend/src/components/connections/ConnectionForm.vue` | Fixed edit-mode table selection bug |

---

## Test Results

### Backend
```
ok  github.com/hejijunhao/heimdall/backend/internal/connectors       0.458s  (4 tests)
ok  github.com/hejijunhao/heimdall/backend/internal/connectors/logs   0.955s  (12 tests)
```

### Frontend
```
Test Files  5 passed (5)
Tests       25 passed (25)
```

### Build Verification
- `go build ./...` — clean
- `go vet ./...` — clean
- `npm run build` — clean

---

## Manual Integration Test Checklist

The following scenarios should be tested with a real Supabase project before production deployment:

| # | Scenario | Expected Result |
|---|----------|-----------------|
| 1 | Create a Supabase connection with valid PAT and project ref | Test passes, status → `active` |
| 2 | Wait one polling interval (30s) | `log_buffer` contains entries with `source_type = 'supabase/postgres_logs'` |
| 3 | Check Agent Log page | Heartbeat entries include polled logs |
| 4 | Trigger a database error in Supabase project | Error gets classified and escalated |
| 5 | Invalid PAT | Test fails with "401 Unauthorized" |
| 6 | Invalid project ref | Test fails with "404 Not Found" |
| 7 | Delete connection | Poller stops immediately |
| 8 | Server restart | Pollers resume for all active Supabase connections |

---

## What's Next

- **Phase 4:** Connection creation wizard — guided multi-step UI replacing the dropdown-based form for new connections
