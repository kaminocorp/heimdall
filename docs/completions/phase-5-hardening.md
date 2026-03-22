# Phase 5 Completion — Code Assessment Hardening

Phase 5 fixes the 8 critical and high-priority issues identified in the comprehensive code assessment (`docs/plans/code-assessment.md`). These are security fixes, correctness bugs, and data integrity improvements — no feature work.

---

## What Was Fixed

### 1. SQL Injection in Supabase Connector (Critical)

**File:** `backend/internal/connectors/logs/supabase.go` (modified)

**Problem:** The `pollTable` method interpolated user-controlled table names directly into a SQL string sent to the Supabase Management API. A crafted `poll_tables` config value like `postgres_logs; DROP TABLE x--` would inject arbitrary SQL.

**Fix:** Added a `validSupabaseTables` allowlist map (6 known Supabase log tables). Validation happens at two levels:
- **Parse time** — `NewSupabase()` rejects any config with table names not in the allowlist, preventing invalid connections from being created.
- **Poll time** — `pollTable()` validates again as a defense-in-depth guard, in case the struct is constructed without going through `NewSupabase()`.

**New test:** `TestNewSupabase_InvalidTableName` — verifies that a config with a malicious table name is rejected with "invalid table name".

---

### 2. XSS in Email Notification HTML (Critical)

**File:** `backend/internal/notifications/format.go` (modified)

**Problem:** `FormatEmailHTML()` interpolated LLM-generated text (`p.Summary`, `p.Assessment`) and user-controlled values (`p.AppName`, `p.Severity`) directly into HTML without escaping. An LLM response containing `<script>` tags or a user naming their app with HTML would render in the email.

**Fix:** All 6 interpolated values in the HTML template now use `html.EscapeString()`:
- `p.AppName`, `p.Severity`, `p.Summary`, `p.Assessment`, `p.Timestamp`, and `FormatSubject(...)`.

Added `"html"` import.

---

### 3. `search_logs` Tool Ignores Query Parameter (Critical)

**Files:**
- `backend/internal/db/queries/log_buffer.sql` (modified) — 2 new queries
- `backend/internal/db/log_buffer.sql.go` (auto-generated)
- `backend/internal/agent/tools_logs.go` (modified)

**Problem:** The `search_logs` tool schema declared a required `query` parameter, but `toolSearchLogs` never read `input["query"]`. Every "search" was actually a "list recent logs", making the agent's investigation capability fundamentally broken.

**Fix:**

**New SQL queries:**
```sql
-- name: SearchLogsByUser :many
SELECT * FROM log_buffer
WHERE user_id = @user_id AND payload::text ILIKE '%' || @query::text || '%'
ORDER BY ingested_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: SearchLogsByUserAndSeverity :many
SELECT * FROM log_buffer
WHERE user_id = @user_id AND payload::text ILIKE '%' || @query::text || '%' AND severity = @severity
ORDER BY ingested_at DESC
LIMIT @row_limit OFFSET @row_offset;
```

**Updated tool dispatch** — `toolSearchLogs` now reads `input["query"]` and routes to one of 4 query paths:
1. `query` + `severity` → `SearchLogsByUserAndSeverity`
2. `query` only → `SearchLogsByUser`
3. `severity` only → `ListLogsByUserAndSeverity` (existing)
4. Neither → `ListLogsByUser` (existing, fallback)

This uses `ILIKE` for case-insensitive substring matching against the full payload JSONB text, so the agent can search for error messages, service names, status codes, etc.

---

### 4. Interactive Chat Uses Wrong Config Source (High)

**File:** `backend/internal/agent/loop.go` (modified)

**Problem:** `RunConversation` loaded the global `agent_config` singleton instead of the per-app `app_agent_configs` row. Per-app model selection and system prompt overrides configured via the UI were ignored during interactive chat (the monitoring loop correctly used per-app config).

**Fix:** `RunConversation` now checks `appID != uuid.Nil` and calls `GetAppAgentConfig(ctx, appID)` for the per-app config. Falls back to the global `GetAgentConfig` only when no `appID` is provided (e.g., `RunLoop` which passes `uuid.Nil`).

The chat handler (`chat.go:59-64`) already parses `app_id` from the WebSocket query params and passes it to `RunConversation`, so this fix activates per-app config for all interactive chat sessions that include `?app_id=...`.

---

### 5. Wildcard CORS Policy (High)

**File:** `backend/internal/api/middleware/cors.go` (modified)

**Problem:** `Access-Control-Allow-Origin: *` allowed any website to make authenticated API calls on behalf of a logged-in Heimdall user. Combined with the `Authorization` header being allowed, this is a real cross-origin attack surface.

**Fix:** CORS is now origin-checked against an allowlist:
- **`CORS_ALLOWED_ORIGINS` env var** — comma-separated list of allowed origins (e.g., `https://heimdall.app,https://app.heimdall.dev`). Set this in production.
- **Dev defaults** — when the env var is unset, allows `http://localhost:5173` and `http://localhost:4173` (Vite dev and preview servers).
- **`Vary: Origin`** header is set when an origin matches, as required by the CORS spec for non-wildcard responses.

Non-matching origins receive no `Access-Control-Allow-Origin` header, causing the browser to block the request.

---

### 6. UserQueries Transaction Safety (High)

**File:** `backend/internal/api/handlers/userqueries.go` (modified)

**Problem:** The assessment flagged that `done()` always calls `Commit`, with no rollback path if a handler encounters an error.

**Resolution:** After analysis, the always-commit pattern is safe for this codebase due to two PostgreSQL guarantees:
1. If any statement fails within a transaction, the transaction enters an "aborted" state and `COMMIT` automatically behaves as `ROLLBACK`.
2. Each handler performs at most one write operation, so there is no window for partial commits.

Changing `done()` to rollback would silently break all write handlers (creates, updates, deletes would be discarded). Instead, added detailed documentation explaining why the pattern is safe, so future maintainers don't repeat this analysis.

---

### 7. Log Pagination Broken for Combined Sources (High)

**File:** `backend/internal/api/handlers/logs.go` (modified)

**Problem:** When `source == "all"`, the handler fetched `limit` rows from raw logs and `limit` rows from agent logs, each with the same `offset`. This produced inconsistent results across pages — page 2 (offset=50) fetched raw logs 50-100 and agent logs 50-100, but the user expected the next 50 entries from the merged, timestamp-sorted set.

**Fix:** When `source == "all"`:
1. Fetch `offset + limit` rows from each source starting at offset 0 — this ensures we have enough rows to cover the merged window.
2. Merge and sort by timestamp descending (existing logic).
3. Apply offset to the merged result (skip the first `offset` entries).
4. Apply limit to the remaining entries.

For single-source queries (`source == "raw"` or `source == "agent"`), behavior is unchanged — offset and limit are passed directly to the database.

The `total` count remains the sum of both source counts, which is correct for the merged view — it tells the client how many total entries exist across both sources.

---

### 8. Missing RLS on 5 Tables + Missing FK Indexes (High)

**Files:**
- `backend/migrations/021_rls_missing_tables.up.sql` (new)
- `backend/migrations/021_rls_missing_tables.down.sql` (new)

**Problem:** Six tables had no Row Level Security policies, inconsistent with the pattern established for all other user-scoped tables. If the Supabase `authenticated` role accessed these tables directly via PostgREST, any user could read/modify another user's data.

**Fix:** Migration 021 enables RLS and creates `FOR ALL USING (...)` policies on:

| Table | Policy logic |
|-------|-------------|
| `organizations` | User's `org_id` matches the organization `id` |
| `applications` | Application's `org_id` matches the user's `org_id` |
| `monitoring_state` | App → org → user chain via join |
| `notification_channels` | App → org → user chain via join |
| `notification_preferences` | App → org → user chain via join |
| `notification_log` | Channel → app → org → user chain via join |

All policies use `current_setting('app.current_user_id', true)::uuid` to read the session variable set by `UserQueries()`, consistent with existing RLS policies.

**Missing FK indexes** also added in the same migration:
- `idx_notification_log_channel_id` — for `ON DELETE CASCADE` performance
- `idx_agent_log_conversation_id` — for `ON DELETE SET NULL` performance

The down migration cleanly reverses all changes.

---

## Files Summary

### New Files

| File | Purpose |
|------|---------|
| `backend/migrations/021_rls_missing_tables.up.sql` | RLS policies for 6 tables + 2 FK indexes |
| `backend/migrations/021_rls_missing_tables.down.sql` | Reverse migration |

### Modified Files

| File | Fix # | Change |
|------|-------|--------|
| `backend/internal/connectors/logs/supabase.go` | 1 | Table name allowlist validation at parse + poll time |
| `backend/internal/connectors/logs/supabase_test.go` | 1 | New test: `TestNewSupabase_InvalidTableName` |
| `backend/internal/notifications/format.go` | 2 | `html.EscapeString()` on all interpolated values |
| `backend/internal/db/queries/log_buffer.sql` | 3 | 2 new search queries with `ILIKE` keyword matching |
| `backend/internal/db/log_buffer.sql.go` | 3 | Auto-generated by sqlc |
| `backend/internal/agent/tools_logs.go` | 3 | Read `query` param, dispatch to search queries |
| `backend/internal/agent/loop.go` | 4 | Use per-app config in `RunConversation` when `appID` is set |
| `backend/internal/api/middleware/cors.go` | 5 | Origin allowlist from env var, dev defaults |
| `backend/internal/api/handlers/userqueries.go` | 6 | Updated documentation explaining tx safety |
| `backend/internal/api/handlers/logs.go` | 7 | Fixed merged-source pagination logic |

### Auto-Generated Files

| File | Change |
|------|--------|
| `backend/internal/db/log_buffer.sql.go` | Regenerated by `make sqlc-generate` |

---

## Verification

```
go build ./...              — clean
go vet ./...                — clean
go test ./...               — 6 packages, all pass (including new TestNewSupabase_InvalidTableName)
npm run test                — 5 files, 25 tests, all pass
```

---

## Deployment Notes

### Environment Variable Required

**`CORS_ALLOWED_ORIGINS`** must be set in production to the frontend origin(s), comma-separated:

```
CORS_ALLOWED_ORIGINS=https://heimdall.app,https://app.heimdall.dev
```

If unset, only `localhost:5173` and `localhost:4173` are allowed (dev mode).

### Migration Required

Run `make migrate-up` to apply migration 021 (RLS policies + indexes). The down migration is safe to run if rollback is needed.
