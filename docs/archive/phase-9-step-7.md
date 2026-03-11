# Phase 9, Step 7 — API Endpoints & Router

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md) (Steps 10, 13–15)

---

## What was done

Created the notification CRUD API handlers and registered all 8 endpoints on the router under `/apps/{appId}/notifications/`.

## Files created

| File | Purpose |
|------|---------|
| `backend/internal/api/handlers/notifications.go` | 8 handlers: preferences get/update, channels CRUD, test, history |

## Files edited

| File | Change |
|------|--------|
| `backend/internal/api/router.go` | Registered 8 notification routes under `/apps/{appId}` |

## API Endpoints

| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| GET | `/apps/{appId}/notifications/preferences` | `GetNotificationPreferences` | Get preferences (returns defaults if no row) |
| PUT | `/apps/{appId}/notifications/preferences` | `UpdateNotificationPreferences` | Upsert preferences |
| GET | `/apps/{appId}/notifications/channels` | `ListNotificationChannels` | List all channels for app |
| POST | `/apps/{appId}/notifications/channels` | `CreateNotificationChannel` | Create channel with type/config validation |
| PUT | `/apps/{appId}/notifications/channels/{channelId}` | `UpdateNotificationChannel` | Update (validates channel belongs to app) |
| DELETE | `/apps/{appId}/notifications/channels/{channelId}` | `DeleteNotificationChannel` | Delete (validates channel belongs to app) |
| POST | `/apps/{appId}/notifications/channels/{channelId}/test` | `TestNotificationChannel` | Send test notification to verify config |
| GET | `/apps/{appId}/notifications/history` | `ListNotificationHistory` | Paginated notification log (limit/offset) |

## Input validation

- **severity_threshold**: Must be one of `info`, `warning`, `error`, `critical`
- **cooldown_minutes**: Must be 1–1440 (1 minute to 24 hours)
- **channel type**: Must be `email`, `slack`, or `discord`
- **email config**: `recipients` must be a non-empty array
- **slack config**: `webhook_url` must start with `https://hooks.slack.com/`
- **discord config**: `webhook_url` must start with `https://discord.com/api/webhooks/`
- **channel ownership**: Update/delete/test verify `channel.app_id == app.ID`

## Design decisions

- **`authorizeApp` for every handler** — Matches the existing per-app authorization pattern. Every notification endpoint validates the app belongs to the user's org before proceeding.
- **Defaults on missing preferences** — `GetNotificationPreferences` returns a synthetic default object instead of 404 when no row exists. This matches the `GetAppAgentConfig` pattern and simplifies frontend logic.
- **Test endpoint uses `notifications.NewChannel`** — Creates a real channel instance and calls `Send()` with a test payload. This validates both config parsing and actual delivery, catching issues like invalid webhook URLs early.
- **`validationError` type** — Lightweight error type for config validation, keeping error messages user-friendly in the JSON response.

## Next step

Step 8 — Frontend: types, API client, notification settings page, router, and sidebar.
