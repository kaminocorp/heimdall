# Phase 9, Step 3 — Migration 019: notification_log

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md)

---

## What was done

Created migration 019 (`notification_log`) — delivery tracking table that records every notification attempt, its status, and any error details.

## Files created

| File | Purpose |
|------|---------|
| `backend/migrations/019_notification_log.up.sql` | Creates `notification_log` table + two indexes |
| `backend/migrations/019_notification_log.down.sql` | Drops the table |

## Schema

```sql
notification_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id          UUID NOT NULL → applications(id) ON DELETE CASCADE,
    channel_id      UUID NOT NULL → notification_channels(id) ON DELETE CASCADE,
    agent_log_id    UUID → agent_log(id) ON DELETE SET NULL,
    severity        TEXT NOT NULL,
    summary         TEXT NOT NULL,
    status          TEXT DEFAULT 'pending',   -- pending, sent, failed
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT now()
)

Indexes:
  idx_notification_log_app_created  — (app_id, created_at DESC)  → paginated history queries
  idx_notification_log_status       — (status)                   → failed notification lookup
```

## Design decisions

- **Separate table from `agent_log`** — Delivery tracking (status, retries, error messages) is a different concern from agent observations. Keeps `agent_log` clean and makes notification history independently queryable.
- **`agent_log_id` with `ON DELETE SET NULL`** — Links back to the monitoring entry that triggered the notification. SET NULL (not CASCADE) because we want to preserve the notification record even if the agent log entry is cleaned up.
- **`channel_id` with `ON DELETE CASCADE`** — If a channel is deleted, its notification history goes with it. Keeps the log table free of orphaned references.
- **Composite index `(app_id, created_at DESC)`** — Optimises the primary query pattern: "show me recent notifications for this app" with pagination.
- **No `updated_at`** — Notification log entries are append-mostly. The only mutation is the status update from `pending` → `sent`/`failed`, which sets `sent_at` instead.

## Next step

Step 4 — sqlc queries for all three notification tables, then `sqlc generate`.
