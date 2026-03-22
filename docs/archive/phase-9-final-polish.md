# Phase 9 — Final Polish

Seven issues found during post-implementation code review. All fixed — no feature changes.

## Issues Fixed

### 1. ConnectionCard delete confirmation timer leak (LOW)

**File:** `frontend/src/components/connections/ConnectionCard.vue`

**Problem:** The 3-second `setTimeout` that auto-resets the delete confirmation state was never cleared on unmount. If the component unmounted within 3 seconds of clicking delete (e.g., navigation, parent re-render), the callback fired on a detached component instance. This is the same class of bug that Phase 7 fixed in `StepTest.vue` and `ConnectionTestModal.vue`, but was missed on `ConnectionCard`.

**Fix:** Store the timer handle in a module-scoped `let deleteTimer`. Clear it in `onBeforeUnmount` and when the user confirms deletion (to avoid the auto-reset firing after the card is removed).

### 2. CORS headers leaked to non-matching origins (LOW)

**File:** `backend/internal/api/middleware/cors.go`

**Problem:** `Access-Control-Allow-Methods` and `Access-Control-Allow-Headers` were set unconditionally on every response, even when the request origin didn't match the allowlist. While not a security vulnerability (browsers block the request without `Access-Control-Allow-Origin`), it leaked API capability information to arbitrary origins.

**Fix:** Moved all three CORS response headers (`Allow-Origin`, `Allow-Methods`, `Allow-Headers`) inside the origin-match conditional. Non-matching origins now receive no CORS headers at all.

### 3. `log.Printf` → `slog.Error` in connections handler (LOW)

**File:** `backend/internal/api/handlers/connections.go`

**Problem:** Seven log calls in the connections handler used `log.Printf` (Go's basic logger) with printf-style format strings, while the entire rest of the codebase uses `slog.Error` with structured key-value attributes. These log lines were invisible to any structured log aggregation (JSON log pipelines, Fly.io log drains, etc.).

**Fix:** Replaced all 7 calls with `slog.Error(msg, "connection_id", connID, "err", err)` structured logging. Replaced the `"log"` import with `"log/slog"`.

### 4. Hand-rolled `splitCSV` replaced with `strings.Split` (COSMETIC)

**File:** `backend/internal/api/middleware/cors.go`

**Problem:** A 12-line manual CSV parser (`splitCSV`) was used to parse the `CORS_ALLOWED_ORIGINS` environment variable, when `strings.Split(s, ",")` does the same thing. The custom implementation also didn't trim whitespace around values, so `"origin1, origin2"` (with a space) would silently fail to match.

**Fix:** Replaced with `strings.Split(env, ",")` + `strings.TrimSpace(o)` per element. Deleted the `splitCSV` function. Added `"strings"` import.

### 5. Missing `sb.Close()` in Supabase TestConnection handler (LOW)

**File:** `backend/internal/api/handlers/connections.go`

**Problem:** The `TestConnection` handler's Supabase case created a `Supabase` connector and called `Connect()` but never called `Close()`. Each test created an `http.Client` whose idle transport connections were never released. The Postgres case correctly called `defer pg.Close(ctx)`, making this an inconsistency.

**Fix:** Added `defer sb.Close()` in the success branch, matching the Postgres pattern.

### 6. Duplicate `supabaseTables` array across wizard and form (LOW)

**Files:**
- `frontend/src/components/connections/wizard/flows.ts` (modified)
- `frontend/src/components/connections/wizard/steps/StepSupabaseTables.vue` (modified)
- `frontend/src/components/connections/ConnectionForm.vue` (modified)

**Problem:** The list of six Supabase log tables (`postgres_logs`, `auth_logs`, etc.) was defined independently in both `ConnectionForm.vue` (used for editing) and `StepSupabaseTables.vue` (used in the wizard). Adding a new table to one but not the other would create a silent inconsistency.

**Fix:** Extracted the canonical list to `flows.ts` as `supabaseLogTables` (exported `as const` for type safety). Both `StepSupabaseTables.vue` and `ConnectionForm.vue` now import and reference this single definition.

### 7. Port validation gap in StepPostgresConfig wizard step (LOW)

**File:** `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue`

**Problem:** The `emit('valid')` check only verified host, database, and user were non-empty. Port validation (1–65535 range, NaN handling) happened in `sync()` where it silently defaulted invalid values to 5432, but the step still reported itself as valid. A user could enter `abc` as the port, see the Continue button enabled, and proceed — the config would silently store 5432 instead.

**Fix:** Added `portValid` check (`!Number.isNaN(parsed) && parsed >= 1 && parsed <= 65535`) to the `emit('valid')` condition. The Continue button is now disabled until the port is a valid integer in the TCP range.

## Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/components/connections/ConnectionCard.vue` | Timer handle stored, cleared on unmount and on confirm |
| 2 | `backend/internal/api/middleware/cors.go` | CORS headers scoped to matching origins, `splitCSV` → `strings.Split` |
| 3 | `backend/internal/api/handlers/connections.go` | `log.Printf` → `slog.Error` with structured attributes (7 calls); `defer sb.Close()` in TestConnection |
| 4 | `frontend/src/components/connections/wizard/flows.ts` | Exported `supabaseLogTables` constant |
| 5 | `frontend/src/components/connections/wizard/steps/StepSupabaseTables.vue` | Import `supabaseLogTables` from flows instead of inline definition |
| 6 | `frontend/src/components/connections/ConnectionForm.vue` | Import `supabaseLogTables` from flows instead of inline definition |
| 7 | `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue` | Port range validation in `emit('valid')` |

## Verification

```
go build ./...       — clean
go vet ./...         — clean
go test ./...        — all packages pass
npx vue-tsc --noEmit — no type errors
npx vitest run       — 25/25 tests pass
```
