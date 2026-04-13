# Missing FK Index on notification_log.agent_log_id

**Date:** 2026-04-12
**Assessment reference:** `docs/plans/code-assessment-2026-04-12.md`, issue #8

## Problem

Migration 019 (`019_notification_log.up.sql`) created the `notification_log` table with a foreign key on `agent_log_id`:

```sql
agent_log_id UUID REFERENCES agent_log(id) ON DELETE SET NULL
```

But no supporting index was created for this column. Migration 021 (`021_rls_missing_tables.up.sql`) later added missing FK indexes for `notification_log(channel_id)` and `agent_log(conversation_id)`, but overlooked `agent_log_id`.

**Impact:** Without an index, every `DELETE` on the `agent_log` table forces PostgreSQL to sequentially scan `notification_log` to verify the FK constraint (`ON DELETE SET NULL`). This gets progressively slower as the table grows — a particular concern during log retention cleanup jobs that delete `agent_log` rows in bulk.

## Fix

**Migration:** `024_notification_log_agent_log_idx`

```sql
-- up
CREATE INDEX IF NOT EXISTS idx_notification_log_agent_log_id ON notification_log(agent_log_id);

-- down
DROP INDEX IF EXISTS idx_notification_log_agent_log_id;
```

Uses `IF NOT EXISTS` / `IF EXISTS` guards for idempotency (in case of manual hotfixes on production).

## Verification

Migration was applied against the live Supabase database and verified in both directions:

```
24/u notification_log_agent_log_idx (136ms)   -- up
24/d notification_log_agent_log_idx (114ms)   -- down
24/u notification_log_agent_log_idx (105ms)   -- re-up
```

- `sqlc generate` — no drift, no generated code changes (index-only migration)
- `go build ./...` — passes
- Migration version now at 24, consistent with `schema_migrations` table
