# Phase 9, Step 1 — Migration 017: notification_channels

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md)

---

## What was done

Created migration 017 (`notification_channels`) — the per-application table that stores notification channel configuration for email, Slack, and Discord.

## Files created

| File | Purpose |
|------|---------|
| `backend/migrations/017_notification_channels.up.sql` | Creates `notification_channels` table + app_id index |
| `backend/migrations/017_notification_channels.down.sql` | Drops the table |

## Schema

```sql
notification_channels (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id     UUID NOT NULL → applications(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,       -- 'email', 'slack', 'discord'
    name       TEXT NOT NULL,       -- user-friendly label
    config     JSONB NOT NULL,      -- type-specific config
    enabled    BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
)
```

## Design decisions

- **JSONB for config** — Each channel type has a different config shape (email: `recipients[]`, Slack/Discord: `webhook_url`). JSONB avoids schema proliferation; validation happens at the handler level.
- **CASCADE on app delete** — Matches the existing FK pattern throughout the data model (connections, app_agent_config, monitoring_state all cascade from applications).
- **`enabled` column** — Allows users to temporarily disable a channel without deleting it and losing the config.
- **`type` as TEXT, not ENUM** — Consistent with how `severity` is handled elsewhere in the codebase. Validation at the application layer keeps migrations simple and avoids ALTER TYPE when adding new channel types.

## Next step

Step 2 — Migration 018: `notification_preferences` (per-app master switch, severity threshold, cooldown).
