# RLS: webhook_idempotency Table

**Status:** Complete
**Date:** 2026-04-16
**Migration:** `backend/migrations/033_rls_webhook_idempotency.up.sql`

## What

Enabled Row Level Security on the `webhook_idempotency` table with no policies — the same defence-in-depth pattern used for `agent_config` and `schema_migrations` in migration 030.

## Why

Migration 032 created the `webhook_idempotency` table but didn't enable RLS. Every other table in the schema has RLS enabled. Without it, non-owner Postgres roles (`anon`, `authenticated`) could query the table directly via PostgREST or a leaked connection string, exposing cached webhook response bodies and connection IDs.

## Why no policies

The table is only accessed by the webhook ingestion handler (`handlers/webhooks.go` → `ingestEntries`), which uses `s.Queries` — the application's main connection pool running as the Postgres owner role. The owner role bypasses RLS entirely, so no policies are needed for the handler to function.

With RLS enabled and no policies:
- **Owner role (backend):** Full access, unaffected.
- **Non-owner roles (anon, authenticated via PostgREST):** Zero rows visible, zero writes possible.

This matches the two-pattern RLS approach used across the codebase:

| Pattern | Tables | Policy |
|---------|--------|--------|
| User-scoped | `connections`, `log_buffer`, `conversations`, etc. | `user_id = app_current_user_id()` or FK join through org hierarchy |
| System/internal | `agent_config`, `schema_migrations`, `webhook_idempotency` | RLS enabled, no policies — non-owner sees nothing |

## Files

| File | Change |
|------|--------|
| `backend/migrations/033_rls_webhook_idempotency.up.sql` | `ALTER TABLE webhook_idempotency ENABLE ROW LEVEL SECURITY` |
| `backend/migrations/033_rls_webhook_idempotency.down.sql` | `ALTER TABLE webhook_idempotency DISABLE ROW LEVEL SECURITY` |

## Verification

```sql
SELECT relname, relrowsecurity FROM pg_class WHERE relname = 'webhook_idempotency';
--  relname             | relrowsecurity
-- ---------------------+----------------
--  webhook_idempotency | t
```
