# Phase 9, Step 4 — sqlc Queries & Code Generation

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md)

---

## What was done

Created sqlc query files for all three notification tables and ran `sqlc generate` to produce type-safe Go code. Verified clean compilation.

## Files created

| File | Purpose |
|------|---------|
| `backend/internal/db/queries/notification_channels.sql` | 6 queries: List, Get, Create, Update, Delete, ListEnabled |
| `backend/internal/db/queries/notification_preferences.sql` | 2 queries: Get, Upsert |
| `backend/internal/db/queries/notification_log.sql` | 4 queries: Insert, UpdateStatus, ListByApp (JOIN), GetLastForApp |

## Generated files (auto, via `sqlc generate`)

| File | Contents |
|------|----------|
| `backend/internal/db/notification_channels.sql.go` | `NotificationChannel` CRUD + `ListEnabledChannelsByApp` |
| `backend/internal/db/notification_preferences.sql.go` | `GetNotificationPreferences` + `UpsertNotificationPreferences` |
| `backend/internal/db/notification_log.sql.go` | Insert, status update, paginated list with JOIN, cooldown lookup |
| `backend/internal/db/models.go` | Three new structs: `NotificationChannel`, `NotificationLog`, `NotificationPreference` |

## Query inventory

### notification_channels
| Query | Type | Purpose |
|-------|------|---------|
| `ListNotificationChannelsByApp` | `:many` | All channels for an app (API listing) |
| `GetNotificationChannel` | `:one` | Single channel by ID (edit/delete) |
| `CreateNotificationChannel` | `:one` | Create with RETURNING * |
| `UpdateNotificationChannel` | `:one` | Update name/config/enabled, auto-sets updated_at |
| `DeleteNotificationChannel` | `:exec` | Hard delete |
| `ListEnabledChannelsByApp` | `:many` | Only enabled channels (dispatcher fan-out) |

### notification_preferences
| Query | Type | Purpose |
|-------|------|---------|
| `GetNotificationPreferences` | `:one` | Load preferences for dispatcher check |
| `UpsertNotificationPreferences` | `:one` | Create or update (ON CONFLICT on app_id PK) |

### notification_log
| Query | Type | Purpose |
|-------|------|---------|
| `InsertNotificationLog` | `:one` | Create pending entry before send attempt |
| `UpdateNotificationLogStatus` | `:exec` | Mark sent/failed, auto-sets sent_at on success |
| `ListNotificationLogByApp` | `:many` | Paginated history with JOIN to get channel_type/name |
| `GetLastNotificationForApp` | `:one` | Most recent sent notification (cooldown check) |

## Key types generated

- `NotificationChannel` — `Config` is `json.RawMessage` (JSONB override in sqlc.yaml)
- `NotificationLog` — `AgentLogID` is `pgtype.UUID` (nullable FK), `SentAt` is `*time.Time` (nullable timestamptz)
- `NotificationPreference` — `CooldownMinutes` is `int32`
- `ListNotificationLogByAppRow` — Extended struct with `ChannelType` and `ChannelName` from the JOIN

## Next step

Step 5 — Notification channel interface + factory + channel implementations (Slack, Discord, Email).
