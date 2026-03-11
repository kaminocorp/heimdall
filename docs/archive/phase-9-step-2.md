# Phase 9, Step 2 — Migration 018: notification_preferences

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md)

---

## What was done

Created migration 018 (`notification_preferences`) — per-application 1:1 table that controls the notification master switch, severity threshold, and cooldown window.

## Files created

| File | Purpose |
|------|---------|
| `backend/migrations/018_notification_preferences.up.sql` | Creates `notification_preferences` table (PK = app_id) |
| `backend/migrations/018_notification_preferences.down.sql` | Drops the table |

## Schema

```sql
notification_preferences (
    app_id              UUID PRIMARY KEY → applications(id) ON DELETE CASCADE,
    enabled             BOOLEAN DEFAULT false,    -- master switch
    severity_threshold  TEXT DEFAULT 'warning',   -- minimum severity to notify
    cooldown_minutes    INTEGER DEFAULT 15,       -- dedup window in minutes
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now()
)
```

## Design decisions

- **`app_id` as PK** — Enforces 1:1 with applications at the database level. Same pattern as `monitoring_state` (migration 016) and `app_agent_config` (migration 015).
- **`enabled` defaults to `false`** — Notifications are opt-in. New apps don't send anything until the user explicitly enables notifications and configures at least one channel.
- **`severity_threshold` as TEXT** — Matches the existing severity convention used in `agent_log`, `parseSeverityFromResponse`, and the monitoring prompt. Ranking logic lives in Go (`severityRank()` helper).
- **`cooldown_minutes` defaults to 15** — Matches the monitor tick interval (15s polling). Prevents a noisy app from flooding channels every cycle while still being responsive. Range validated at handler level (1–1440).

## Next step

Step 3 — Migration 019: `notification_log` (delivery tracking table).
