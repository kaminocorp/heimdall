# Activity Feed App-Scoping — Phase 1: Database Migration

## Status: Complete

## Summary

Added `app_id UUID` columns to `log_buffer` and `agent_log` tables so the
Activity feed can filter by the currently selected application. This is the
database foundation for per-app Activity scoping — subsequent phases wire
the queries, handlers, write paths, and frontend.

**Plan reference:** `docs/executing/activity-app-scoping.md`, Part 1

---

## Migration: `025_add_app_id_to_logs`

**Files created:**
| File | Purpose |
|------|---------|
| `backend/migrations/025_add_app_id_to_logs.up.sql` | Add columns, backfill, create indexes |
| `backend/migrations/025_add_app_id_to_logs.down.sql` | Drop indexes and columns |

### What the up migration does

1. **`log_buffer.app_id`** — nullable `UUID` FK to `applications(id)` with
   `ON DELETE CASCADE`. Backfilled from `connections.app_id` via join on
   `connection_id`. Every existing row gets a value because `connection_id`
   is `NOT NULL` and all connections have `app_id`.

2. **`agent_log.app_id`** — same column definition. Backfilled from
   `detail->>'app_id'` JSONB extraction where present. Monitoring and
   scheduled investigation entries have this field; interactive chat entries
   do not (those remain `NULL` until Phase 3 populates at write time).

3. **Partial indexes** on both tables — `(app_id, timestamp DESC) WHERE
   app_id IS NOT NULL`. Partial because:
   - Historical NULLs will never match a filtered query.
   - Keeps the index lean — only rows that participate in per-app views.

### Backfill results (production, 2026-04-13)

| Table | Total rows | Backfilled | NULL remaining |
|-------|-----------|------------|----------------|
| `log_buffer` | 0 | 0 | 0 |
| `agent_log` | 32 | 32 (100%) | 0 |

All `agent_log` rows had `app_id` in their JSONB `detail`, so backfill
achieved full coverage. `log_buffer` was empty at migration time (logs are
pruned by the background job).

### Down migration

Drops both indexes, then drops both columns. No data loss beyond the
denormalised `app_id` values, which can be re-derived from
`connections.app_id` and `agent_log.detail->>'app_id'`.

---

## Design decisions

### Why nullable (not NOT NULL)?

- `agent_log` interactive chat entries historically have no app context in
  their JSONB — can't backfill what doesn't exist.
- Some future entry types (system heartbeats, org-level events) may
  genuinely have no app scope.
- The query contract is simple: when `app_id` is in the WHERE clause,
  NULLs are excluded by SQL semantics. No special handling needed.

### Why denormalise instead of joining?

`log_buffer` already has a join path (`connection_id → connections.app_id`),
but adding `app_id` directly means:
- One WHERE clause instead of a JOIN per query.
- Matches the existing pattern (`user_id` was denormalised onto
  `log_buffer` the same way in migration 012).
- `app_id` never changes for a given connection, so no sync risk.

### Why partial indexes?

Standard B-tree indexes include every row. The `WHERE app_id IS NOT NULL`
filter excludes historical NULLs that will never match a per-app query,
keeping the index smaller and writes to NULL-app rows cheaper.

---

## What's next

| Phase | Scope | Status |
|-------|-------|--------|
| **1 — Migration** | Add columns, backfill, index | **Done** |
| 2 — Backend queries & handler | New sqlc queries, `app_id` query param in logs handler | Not started |
| 3 — Backend write paths | Populate `app_id` on every INSERT | Not started |
| 4 — Frontend | Pass `currentAppId` to API, re-fetch on app switch | Not started |
| 5 — Verification | Backend + frontend tests, manual QA | Not started |
