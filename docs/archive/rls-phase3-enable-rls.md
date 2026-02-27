# RLS Phase 3 — Enable Row Level Security (Completed)

**Date:** 2026-02-26
**Scope:** Enable RLS and create policies on all user-scoped tables via migration 013.

---

## Why

Phases 1–2 added `user_id` columns and `set_config` plumbing. This phase writes the actual RLS policies that enforce row-level isolation.

The backend connects as the `postgres` superuser (table owner), which **bypasses RLS by default**. This is intentional — the backend keeps full, unrestricted access. The policies protect against non-owner access paths:

- Supabase's `anon` and `authenticated` roles (used by PostgREST / client-side SDKs)
- Supabase SQL Editor / Dashboard queries (run as `authenticated`)
- Direct `psql` sessions using non-owner roles

---

## What Changed

### Migration `013_enable_rls.up.sql`

**Helper function:**
```sql
CREATE OR REPLACE FUNCTION app_current_user_id() RETURNS UUID
```
Safely reads `current_setting('app.current_user_id', true)` — returns `NULL` if unset (no error). Used by all policies.

**Policies created:**

| Table | Policy | Rule |
|-------|--------|------|
| `users` | `users_self` | `id = app_current_user_id()` |
| `connections` | `connections_owner` | `user_id = app_current_user_id()` |
| `conversations` | `conversations_owner` | `user_id = app_current_user_id()` |
| `agent_log` | `agent_log_owner` | `user_id = app_current_user_id()` |
| `log_buffer` | `log_buffer_owner` | `user_id = app_current_user_id()` |
| `investigations` | `investigations_owner` | `user_id = app_current_user_id()` |

All policies use `FOR ALL USING (...)` — covering SELECT, INSERT, UPDATE, DELETE in one rule.

**Not covered:**
- `agent_config` — system-wide single-row table, no user scoping needed.

### Migration `013_enable_rls.down.sql`

Drops all 6 policies, disables RLS on all 6 tables, drops the helper function.

### No backend changes

The `postgres` owner role bypasses RLS automatically. No handler, agent, or webhook modifications were required.

---

## Verification

- `go build ./...` — still compiles (no Go changes).
- After applying to Supabase, confirm RLS is active:
  ```sql
  SELECT tablename, rowsecurity
  FROM pg_tables
  WHERE schemaname = 'public';
  ```
  Should show `rowsecurity = true` for: `users`, `connections`, `conversations`, `agent_log`, `log_buffer`, `investigations`.

---

## Design Decision: No FORCE ROW LEVEL SECURITY

`FORCE ROW LEVEL SECURITY` would make even the table owner subject to policies. We chose **not** to use it because the `postgres` superuser connection should always have full access. The `set_config` plumbing from Phase 2 remains in place as future-proofing — if the backend ever migrates to a dedicated non-owner app role, RLS enforcement will activate automatically.
