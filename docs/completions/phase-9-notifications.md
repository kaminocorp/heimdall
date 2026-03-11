# Phase 9 — Notifications & Escalation: Implementation Plan

Reference: [Post-MVP Roadmap](./post-mvp-roadmap.md) | [Vision](../vision.md) | [Changelog](../changelog.md)

---

## Goal

Make Heimdall's monitoring findings actionable without the user opening the dashboard. When the agent assesses flagged logs as `warning`, `error`, or `critical`, Heimdall should notify the right people through their preferred channels — email, Slack, or Discord.

A monitoring agent that can't alert anyone is a smoke detector with no alarm. Phase 8 gave Heimdall the ability to see; Phase 9 gives it the ability to speak.

---

## Architecture Overview

```
Monitor loop emits agent_log entry (severity ≥ threshold)
   ↓
┌─────────────────────────────────────┐
│  Notification Dispatcher            │  ← In-process, async
│  1. Check app notification config   │
│  2. Apply severity threshold filter │
│  3. Apply cooldown (dedup window)   │
│  4. Dispatch to enabled channels    │
└─────────────┬───────────────────────┘
              │
     ┌────────┼────────────┐
     │        │            │
   Email    Slack       Discord
  (SMTP/    (Webhook)   (Webhook)
   API)
     │        │            │
     └────────┼────────────┘
              │
       Write to notification_log
       (delivery tracking)
```

**Key principle:** Notifications are fire-and-forget, like `EmitLog`. A failed notification must never block the monitoring loop. Delivery failures are logged and retried once — if both attempts fail, the notification is marked `failed` in the log and moves on.

---

## Current State

| Component | Status | Notes |
|-----------|--------|-------|
| `agent_log` table | Exists | Has `severity` column, `entry_type` for filtering |
| `EmitLogWithSeverity` | Working | Monitoring entries include severity + detail JSONB |
| `monitor.go` | Working | Emits `monitoring` entries with severity after Claude assessment |
| `app_agent_config` | Exists | Per-app config (model, mode, interval) — no notification fields |
| User `email` | Available | Stored in `users` table from Supabase auth |
| Notification system | **Missing** | No channels, config, dispatch, or delivery tracking |

---

## Implementation Steps

### Step 1 — Database: Notification Configuration

Per-application notification settings. Each app can have multiple notification channels, each with its own config.

**Migration 017 — `notification_channels`:**

```sql
-- Notification channels (per-application, multiple per app)
CREATE TABLE notification_channels (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id     UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,          -- 'email', 'slack', 'discord'
    name       TEXT NOT NULL,          -- user-friendly label, e.g. "Ops Slack"
    config     JSONB NOT NULL,         -- channel-specific config (see below)
    enabled    BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_channels_app_id ON notification_channels(app_id);
```

**Channel config shapes (JSONB):**

```jsonc
// Email
{ "recipients": ["ops@company.com", "dev@company.com"] }

// Slack (incoming webhook)
{ "webhook_url": "https://hooks.slack.com/services/T.../B.../xxx" }

// Discord (webhook)
{ "webhook_url": "https://discord.com/api/webhooks/123/abc" }
```

**Migration 018 — `notification_preferences`:**

```sql
-- Per-app notification preferences (1:1 with applications)
CREATE TABLE notification_preferences (
    app_id              UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    enabled             BOOLEAN NOT NULL DEFAULT false,   -- master switch
    severity_threshold  TEXT NOT NULL DEFAULT 'warning',  -- minimum severity: info, warning, error, critical
    cooldown_minutes    INTEGER NOT NULL DEFAULT 15,      -- suppress duplicate notifications for N minutes
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Why `severity_threshold` as text:** Matches the existing severity convention (`info`, `warning`, `error`, `critical`) used throughout `agent_log`, `parseSeverityFromResponse`, and the monitoring prompt. No enum needed — validation happens in the handler.

**Why `cooldown_minutes`:** Without dedup, a noisy app could generate a notification every 15 seconds (monitor tick interval). The cooldown window suppresses repeat notifications for the same app within N minutes, regardless of channel. Default 15 minutes balances responsiveness with noise reduction.

---

### Step 2 — Database: Notification Log

Track every notification attempt for observability and dedup.

**Migration 019 — `notification_log`:**

```sql
CREATE TABLE notification_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id          UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    channel_id      UUID NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    agent_log_id    UUID REFERENCES agent_log(id) ON DELETE SET NULL,
    severity        TEXT NOT NULL,
    summary         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, sent, failed
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_log_app_created ON notification_log(app_id, created_at DESC);
CREATE INDEX idx_notification_log_status ON notification_log(status);
```

**Why a separate table vs. reusing `agent_log`:** Notification delivery tracking (status, retries, error messages) is a different concern from agent observations. Keeping them separate avoids polluting the agent log with delivery metadata and makes the notification history queryable without filtering.

---

### Step 3 — Database: Supporting Queries

New sqlc queries for the notification pipeline:

```sql
-- notification_channels.sql

-- name: ListNotificationChannelsByApp :many
SELECT * FROM notification_channels
WHERE app_id = $1
ORDER BY created_at ASC;

-- name: GetNotificationChannel :one
SELECT * FROM notification_channels
WHERE id = $1;

-- name: CreateNotificationChannel :one
INSERT INTO notification_channels (app_id, type, name, config, enabled)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateNotificationChannel :one
UPDATE notification_channels
SET name = $2, config = $3, enabled = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteNotificationChannel :exec
DELETE FROM notification_channels WHERE id = $1;

-- name: ListEnabledChannelsByApp :many
SELECT * FROM notification_channels
WHERE app_id = $1 AND enabled = true;
```

```sql
-- notification_preferences.sql

-- name: GetNotificationPreferences :one
SELECT * FROM notification_preferences WHERE app_id = $1;

-- name: UpsertNotificationPreferences :one
INSERT INTO notification_preferences (app_id, enabled, severity_threshold, cooldown_minutes)
VALUES ($1, $2, $3, $4)
ON CONFLICT (app_id) DO UPDATE
SET enabled = EXCLUDED.enabled,
    severity_threshold = EXCLUDED.severity_threshold,
    cooldown_minutes = EXCLUDED.cooldown_minutes,
    updated_at = now()
RETURNING *;
```

```sql
-- notification_log.sql

-- name: InsertNotificationLog :one
INSERT INTO notification_log (app_id, channel_id, agent_log_id, severity, summary, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateNotificationLogStatus :exec
UPDATE notification_log
SET status = $2, error_message = $3, sent_at = CASE WHEN $2 = 'sent' THEN now() ELSE sent_at END
WHERE id = $1;

-- name: ListNotificationLogByApp :many
SELECT nl.*, nc.type AS channel_type, nc.name AS channel_name
FROM notification_log nl
JOIN notification_channels nc ON nc.id = nl.channel_id
WHERE nl.app_id = $1
ORDER BY nl.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLastNotificationForApp :one
SELECT * FROM notification_log
WHERE app_id = $1 AND status = 'sent'
ORDER BY created_at DESC
LIMIT 1;
```

---

### Step 4 — Notification Channels (Backend)

Create the notification dispatch layer. Each channel type implements a common interface.

**New package: `backend/internal/notifications/`**

```
notifications/
  notifier.go       — interface + dispatcher
  email.go          — email channel (Resend API)
  slack.go          — Slack incoming webhook
  discord.go        — Discord webhook
  format.go         — message formatting (shared)
```

**`notifier.go` — Interface:**

```go
package notifications

import "context"

// Payload represents a notification to be sent.
type Payload struct {
    AppName    string
    Severity   string
    Summary    string
    Assessment string // full assessment from agent_log detail
    Timestamp  string
}

// Channel sends notifications to a specific destination.
type Channel interface {
    // Type returns the channel type identifier ("email", "slack", "discord").
    Type() string
    // Send delivers the notification. Returns nil on success.
    Send(ctx context.Context, payload Payload) error
}

// NewChannel creates a Channel from its type and config JSONB.
func NewChannel(channelType string, config json.RawMessage) (Channel, error) {
    switch channelType {
    case "email":
        return newEmailChannel(config)
    case "slack":
        return newSlackChannel(config)
    case "discord":
        return newDiscordChannel(config)
    default:
        return nil, fmt.Errorf("unknown channel type: %s", channelType)
    }
}
```

**Why an interface with a factory:** Same pattern as the `Classifier` interface in `agent/classifier.go`. Keeps each channel implementation isolated and makes testing straightforward (mock channel that records calls).

---

### Step 5 — Email Channel (Resend)

Use [Resend](https://resend.com) for transactional email. Free tier covers 100 emails/day — more than sufficient for monitoring alerts. Simple REST API, no SMTP configuration needed.

**New env vars:**

```
RESEND_API_KEY=re_xxx          # Resend API key
NOTIFICATION_FROM_EMAIL=heimdall@yourdomain.com  # Verified sender
```

**`email.go`:**

```go
type EmailChannel struct {
    recipients []string
    apiKey     string
    fromEmail  string
}

type emailConfig struct {
    Recipients []string `json:"recipients"`
}

func newEmailChannel(configBytes json.RawMessage) (*EmailChannel, error) {
    var cfg emailConfig
    if err := json.Unmarshal(configBytes, &cfg); err != nil {
        return nil, err
    }
    return &EmailChannel{
        recipients: cfg.Recipients,
        apiKey:     os.Getenv("RESEND_API_KEY"),
        fromEmail:  os.Getenv("NOTIFICATION_FROM_EMAIL"),
    }, nil
}

func (e *EmailChannel) Type() string { return "email" }

func (e *EmailChannel) Send(ctx context.Context, p Payload) error {
    // POST https://api.resend.com/emails
    // Body: { from, to, subject, html }
    // Subject: "[Heimdall] {severity}: {app_name}"
    // HTML: formatted assessment using format.go template
}
```

**Why Resend over SMTP:** Zero configuration (no SMTP host/port/TLS), single HTTP call, free tier sufficient for monitoring volume. Swap to SendGrid/Mailgun later if needed — the interface isolates the provider.

---

### Step 6 — Slack Channel

Slack incoming webhooks are a single POST with a JSON body. No OAuth, no bot token, no Slack SDK needed.

**`slack.go`:**

```go
type SlackChannel struct {
    webhookURL string
}

type slackConfig struct {
    WebhookURL string `json:"webhook_url"`
}

func (s *SlackChannel) Type() string { return "slack" }

func (s *SlackChannel) Send(ctx context.Context, p Payload) error {
    // POST to s.webhookURL with Slack Block Kit payload:
    // - Header block: severity emoji + app name
    // - Section block: assessment summary
    // - Context block: timestamp
    // Severity emoji mapping: critical=🔴, error=🟠, warning=🟡, info=🔵
}
```

---

### Step 7 — Discord Channel

Discord webhooks accept a similar JSON format to Slack but use embeds.

**`discord.go`:**

```go
type DiscordChannel struct {
    webhookURL string
}

type discordConfig struct {
    WebhookURL string `json:"webhook_url"`
}

func (d *DiscordChannel) Type() string { return "discord" }

func (d *DiscordChannel) Send(ctx context.Context, p Payload) error {
    // POST to d.webhookURL with Discord embed:
    // - Color: severity-mapped (critical=red, error=orange, warning=yellow, info=blue)
    // - Title: "[severity] App Name"
    // - Description: assessment summary
    // - Footer: timestamp
}
```

---

### Step 8 — Notification Dispatcher

The dispatcher is the orchestrator. It checks preferences, applies cooldown, fans out to channels, and logs delivery status.

**`notifier.go` — Dispatcher:**

```go
type Dispatcher struct {
    queries *db.Queries
    config  *config.Config
}

func NewDispatcher(queries *db.Queries, cfg *config.Config) *Dispatcher {
    return &Dispatcher{queries: queries, config: cfg}
}

// Notify is called by the monitoring loop after emitting an agent_log entry.
// It checks notification config, applies filters, and dispatches to all enabled channels.
// Fire-and-forget: errors are logged, never returned.
func (d *Dispatcher) Notify(ctx context.Context, appID uuid.UUID, agentLogID uuid.UUID, severity, summary, assessment string) {
    // 1. Load notification preferences for this app.
    //    If not found or disabled → return early.
    prefs, err := d.queries.GetNotificationPreferences(ctx, appID)
    if err != nil || !prefs.Enabled {
        return
    }

    // 2. Check severity threshold.
    //    If severity < threshold → return early.
    if severityRank(severity) < severityRank(prefs.SeverityThreshold) {
        return
    }

    // 3. Check cooldown.
    //    If last notification for this app was < cooldown_minutes ago → return early.
    lastNotif, err := d.queries.GetLastNotificationForApp(ctx, appID)
    if err == nil && time.Since(lastNotif.CreatedAt) < time.Duration(prefs.CooldownMinutes)*time.Minute {
        slog.Debug("notification suppressed by cooldown", "app_id", appID, "last_sent", lastNotif.CreatedAt)
        return
    }

    // 4. Load enabled channels.
    channels, err := d.queries.ListEnabledChannelsByApp(ctx, appID)
    if err != nil || len(channels) == 0 {
        return
    }

    // 5. Resolve app name for notification content.
    //    (Available from agent_log detail or looked up from applications table.)

    // 6. Dispatch to each channel.
    payload := Payload{
        AppName:    appName,
        Severity:   severity,
        Summary:    summary,
        Assessment: assessment,
        Timestamp:  time.Now().Format(time.RFC3339),
    }

    for _, ch := range channels {
        d.dispatchToChannel(ctx, ch, appID, agentLogID, payload)
    }
}

func (d *Dispatcher) dispatchToChannel(ctx context.Context, ch db.NotificationChannel, appID, agentLogID uuid.UUID, p Payload) {
    // 1. Create notification_log entry (status: pending)
    logEntry, _ := d.queries.InsertNotificationLog(ctx, ...)

    // 2. Build channel from type + config
    sender, err := NewChannel(ch.Type, ch.Config)
    if err != nil {
        d.queries.UpdateNotificationLogStatus(ctx, logEntry.ID, "failed", err.Error())
        return
    }

    // 3. Attempt send
    if err := sender.Send(ctx, p); err != nil {
        slog.Warn("notification send failed, retrying once", "channel", ch.Name, "err", err)
        // 4. Single retry
        if retryErr := sender.Send(ctx, p); retryErr != nil {
            d.queries.UpdateNotificationLogStatus(ctx, logEntry.ID, "failed", retryErr.Error())
            return
        }
    }

    // 5. Mark as sent
    d.queries.UpdateNotificationLogStatus(ctx, logEntry.ID, "sent", "")
}
```

**Severity ranking helper:**

```go
func severityRank(s string) int {
    switch s {
    case "critical": return 4
    case "error":    return 3
    case "warning":  return 2
    case "info":     return 1
    default:         return 0
    }
}
```

---

### Step 9 — Wire into Monitor Loop

The notification dispatcher is called from `monitor.go` after emitting a `monitoring` agent_log entry. This is the only integration point — minimal surgery.

**`monitor.go` changes:**

```go
// In monitorApp(), after EmitLogWithSeverity for "monitoring" entry:
if len(flagged) > 0 {
    // ... existing: RunMonitoring, EmitLogWithSeverity ...

    // NEW: Dispatch notification (fire-and-forget).
    if a.notifier != nil {
        go a.notifier.Notify(ctx, app.ID, logEntryID, severity, summary, assessment)
    }
}
```

**`agent.go` changes — add notifier field:**

```go
type Agent struct {
    queries    *db.Queries
    client     *anthropic.Client
    config     *config.Config
    classifier Classifier
    notifier   *notifications.Dispatcher  // NEW
    cancel     context.CancelFunc
    wg         sync.WaitGroup
}

func New(queries *db.Queries, cfg *config.Config, classifier Classifier, notifier *notifications.Dispatcher) *Agent {
    // ...
    return &Agent{
        // ...,
        notifier: notifier,
    }
}
```

**`cmd/heimdall/main.go` changes:**

```go
notifier := notifications.NewDispatcher(queries, cfg)
ag := agent.New(queries, cfg, classifier, notifier)
```

**Why fire-and-forget goroutine:** Notification dispatch involves HTTP calls to external services (Resend, Slack, Discord). These can be slow (100ms–2s). Running in a goroutine prevents blocking the monitor's cursor advance and next-app processing. The dispatcher handles its own error logging.

**EmitLog return value:** Currently `EmitLog` doesn't return the inserted `agent_log.id`. We need the ID for the `notification_log.agent_log_id` FK. Two options:

- **(A) Return the ID from `emitLog`** — Change `emitLog` to return `(uuid.UUID, error)`. The existing callers that ignore the return value continue working (Go allows unused returns). New caller in monitor.go captures it.
- **(B) Query by content** — Look up the agent_log entry by `app_id + entry_type + created_at`. Fragile and unnecessary.

**Recommendation:** Option A. Minimal change — `InsertAgentLog` already returns the full row (sqlc `RETURNING *`).

---

### Step 10 — API Endpoints

New endpoints for notification configuration, following the existing per-app pattern.

**Routes (added to `router.go` under `/apps/{appId}`):**

```go
r.Route("/apps/{appId}", func(r chi.Router) {
    // ... existing routes ...

    // Notification management
    r.Get("/notifications/preferences", s.GetNotificationPreferences)
    r.Put("/notifications/preferences", s.UpdateNotificationPreferences)
    r.Get("/notifications/channels", s.ListNotificationChannels)
    r.Post("/notifications/channels", s.CreateNotificationChannel)
    r.Put("/notifications/channels/{channelId}", s.UpdateNotificationChannel)
    r.Delete("/notifications/channels/{channelId}", s.DeleteNotificationChannel)
    r.Post("/notifications/channels/{channelId}/test", s.TestNotificationChannel)
    r.Get("/notifications/history", s.ListNotificationHistory)
})
```

**New handler file: `handlers/notifications.go`**

All handlers use `authorizeApp()` for per-app authorization, matching existing pattern in `applications.go`.

| Handler | Method | Notes |
|---------|--------|-------|
| `GetNotificationPreferences` | GET | Returns preferences (or defaults if none set) |
| `UpdateNotificationPreferences` | PUT | Validates severity threshold enum, cooldown bounds |
| `ListNotificationChannels` | GET | Lists all channels for app |
| `CreateNotificationChannel` | POST | Validates type enum, config shape per type |
| `UpdateNotificationChannel` | PUT | Validates channel belongs to app |
| `DeleteNotificationChannel` | DELETE | Validates channel belongs to app |
| `TestNotificationChannel` | POST | Sends a test notification to verify config |
| `ListNotificationHistory` | GET | Paginated notification log |

**Test notification:** `POST /notifications/channels/{channelId}/test` sends a synthetic notification with `severity: info`, `summary: "Test notification from Heimdall"`. This lets users verify their webhook URL or email config before relying on it.

**Input validation:**

```go
// severity_threshold must be one of:
validThresholds := map[string]bool{"info": true, "warning": true, "error": true, "critical": true}

// cooldown_minutes: min 1, max 1440 (24 hours)

// channel type must be one of:
validTypes := map[string]bool{"email": true, "slack": true, "discord": true}

// channel config validated per type:
//   email: recipients must be non-empty array of valid email strings
//   slack: webhook_url must start with "https://hooks.slack.com/"
//   discord: webhook_url must start with "https://discord.com/api/webhooks/"
```

---

### Step 11 — Config Updates

**`config.go` — new optional env vars:**

```go
type Config struct {
    // ... existing ...
    ResendAPIKey          string  // RESEND_API_KEY
    NotificationFromEmail string  // NOTIFICATION_FROM_EMAIL
}
```

These are optional — email notifications only work if configured. Slack and Discord require no server-side config (webhook URLs are stored per-channel in the DB).

---

### Step 12 — Frontend: Notification Settings Page

New page under the Agent section for managing notification preferences and channels.

**New file: `frontend/src/pages/NotificationsPage.vue`**

**Layout (two sections):**

```
┌──────────────────────────────────────────────────┐
│  NOTIFICATION PREFERENCES                         │
│  ┌──────────────────────────────────────────────┐ │
│  │  Enabled: [toggle]                           │ │
│  │  Severity threshold: [dropdown: info|warn|…] │ │
│  │  Cooldown: [input] minutes                   │ │
│  │                              [Save]          │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
│  NOTIFICATION CHANNELS                             │
│  ┌─────────────────────────────────┐               │
│  │  📧 Ops Email         enabled  │ [Test] [Edit] │
│  │  recipients: ops@co, dev@co    │        [Del]  │
│  └─────────────────────────────────┘               │
│  ┌─────────────────────────────────┐               │
│  │  💬 Ops Slack          enabled │ [Test] [Edit] │
│  │  hooks.slack.com/…             │        [Del]  │
│  └─────────────────────────────────┘               │
│                                                    │
│  [+ Add Channel]                                   │
│                                                    │
│  RECENT NOTIFICATIONS                              │
│  ┌──────────────────────────────────────────────┐ │
│  │  🟠 error  │ Ops Slack │ sent    │ 2m ago   │ │
│  │  🟡 warning│ Ops Email │ sent    │ 18m ago  │ │
│  │  🔴 critical│ Ops Slack│ failed  │ 1h ago   │ │
│  └──────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────┘
```

**Component breakdown:**

| Component | File | Purpose |
|-----------|------|---------|
| `NotificationsPage.vue` | `pages/` | Page shell, fetches preferences + channels + history |
| `NotificationChannelForm.vue` | `components/notifications/` | Create/edit channel (type selector, config fields) |
| `NotificationChannelCard.vue` | `components/notifications/` | Display channel with test/edit/delete actions |
| `NotificationHistoryTable.vue` | `components/notifications/` | Paginated table of recent notifications |

**Reactivity:** Watch `appStore.currentAppId` to re-fetch preferences and channels when app changes (matching existing page pattern).

---

### Step 13 — Frontend: API & Types

**New file: `frontend/src/api/notifications.ts`**

```typescript
import client from './client'

export function getNotificationPreferences(appId: string) {
  return client.get(`/apps/${appId}/notifications/preferences`).then(r => r.data)
}

export function updateNotificationPreferences(appId: string, prefs: UpdatePreferencesPayload) {
  return client.put(`/apps/${appId}/notifications/preferences`, prefs).then(r => r.data)
}

export function listNotificationChannels(appId: string) {
  return client.get(`/apps/${appId}/notifications/channels`).then(r => r.data)
}

export function createNotificationChannel(appId: string, channel: CreateChannelPayload) {
  return client.post(`/apps/${appId}/notifications/channels`, channel).then(r => r.data)
}

export function updateNotificationChannel(appId: string, channelId: string, channel: UpdateChannelPayload) {
  return client.put(`/apps/${appId}/notifications/channels/${channelId}`, channel).then(r => r.data)
}

export function deleteNotificationChannel(appId: string, channelId: string) {
  return client.delete(`/apps/${appId}/notifications/channels/${channelId}`)
}

export function testNotificationChannel(appId: string, channelId: string) {
  return client.post(`/apps/${appId}/notifications/channels/${channelId}/test`).then(r => r.data)
}

export function listNotificationHistory(appId: string, page = 1, pageSize = 20) {
  return client.get(`/apps/${appId}/notifications/history`, {
    params: { limit: pageSize, offset: (page - 1) * pageSize }
  }).then(r => r.data)
}
```

**New file: `frontend/src/types/notifications.ts`**

```typescript
export interface NotificationPreferences {
  app_id: string
  enabled: boolean
  severity_threshold: 'info' | 'warning' | 'error' | 'critical'
  cooldown_minutes: number
  created_at: string
  updated_at: string
}

export interface UpdatePreferencesPayload {
  enabled: boolean
  severity_threshold: string
  cooldown_minutes: number
}

export type NotificationChannelType = 'email' | 'slack' | 'discord'

export interface NotificationChannel {
  id: string
  app_id: string
  type: NotificationChannelType
  name: string
  config: EmailConfig | SlackConfig | DiscordConfig
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface EmailConfig {
  recipients: string[]
}

export interface SlackConfig {
  webhook_url: string
}

export interface DiscordConfig {
  webhook_url: string
}

export interface CreateChannelPayload {
  type: NotificationChannelType
  name: string
  config: Record<string, unknown>
  enabled?: boolean
}

export interface UpdateChannelPayload {
  name: string
  config: Record<string, unknown>
  enabled: boolean
}

export interface NotificationLogEntry {
  id: string
  app_id: string
  channel_id: string
  channel_type: string
  channel_name: string
  agent_log_id: string | null
  severity: string
  summary: string
  status: 'pending' | 'sent' | 'failed'
  error_message: string | null
  sent_at: string | null
  created_at: string
}
```

---

### Step 14 — Frontend: Router & Sidebar

**`router/index.ts` — add route:**

```typescript
{
  path: '/notifications',
  name: 'notifications',
  component: () => import('@/pages/NotificationsPage.vue'),
}
```

**`AppSidebar.vue` — add nav item:**

```typescript
{
  label: 'Agent',
  items: [
    { name: 'Configuration', to: '/agent/config', routeName: 'agent-config' },
    { name: 'Chat', to: '/agent/chat', routeName: 'agent-chat' },
    { name: 'Log', to: '/agent/log', routeName: 'agent-log' },
    { name: 'Notifications', to: '/notifications', routeName: 'notifications' },  // NEW
  ],
}
```

**Why under Agent:** Notifications are a direct extension of the agent's monitoring output. Users configure what the agent tells them about and through which channels. Keeping it in the Agent section maintains logical grouping.

---

### Step 15 — Dashboard Integration

Add a notification status card to `DashboardPage.vue` alongside the existing monitoring card.

**New data in dashboard fetch:**

```typescript
// Fetch alongside existing stats
const notifPrefs = await getNotificationPreferences(appId)
const recentNotifs = await listNotificationHistory(appId, 1, 5)
```

**Card content:**

```
┌─ Notifications ──────────────────────┐
│  Status: Enabled                      │
│  Threshold: warning+                  │
│  Channels: 2 active                   │
│  Last sent: 12m ago                   │
│  Recent: 3 sent, 1 failed (24h)      │
└───────────────────────────────────────┘
```

---

### Step 16 — Tests

| Test | What it covers |
|------|----------------|
| `notifications_test.go` (handler) | CRUD for channels, preferences get/update, history pagination, test notification, authorization checks |
| `notifier_test.go` (unit) | Dispatcher: severity threshold filtering, cooldown dedup, channel dispatch fan-out |
| `slack_test.go` (unit) | Slack message formatting, webhook payload structure |
| `discord_test.go` (unit) | Discord embed formatting |
| `email_test.go` (unit) | Email HTML template, recipient list handling |
| Frontend store test | Notification preferences and channels CRUD with mocked API |

**Test approach:**

- **Handler tests** follow existing `testhelpers_test.go` pattern with full DB setup
- **Dispatcher tests** use a mock `Channel` implementation that records `Send()` calls
- **Channel tests** use `httptest.NewServer` to verify HTTP request format without hitting real APIs
- **No integration tests with real Slack/Discord/Resend** — manual verification during development, test endpoint for production verification

---

## Implementation Order

```
Step 1:  Migration 017 — notification_channels table
Step 2:  Migration 018 — notification_preferences table
Step 3:  Migration 019 — notification_log table
         ↓
Step 4:  Queries — sqlc queries for all three tables
Step 5:  sqlc generate
         ↓
Step 6:  Channel interface + factory (notifier.go)
Step 7:  Slack channel (slack.go)
Step 8:  Discord channel (discord.go)
Step 9:  Email channel (email.go)
         ↓
Step 10: Dispatcher — severity filter, cooldown, fan-out (notifier.go)
         ↓
Step 11: Wire into monitor loop (monitor.go, agent.go, main.go)
Step 12: EmitLog return value change (emit.go)
         ↓
Step 13: Config update (config.go)
Step 14: API handlers (handlers/notifications.go)
Step 15: Router update (router.go)
         ↓
Step 16: Frontend types + API client
Step 17: NotificationsPage + components
Step 18: Router + sidebar update
Step 19: Dashboard notification card
         ↓
Step 20: Tests (handler + unit + frontend)
```

**Suggested PR breakdown:**

| PR | Steps | Description |
|----|-------|-------------|
| PR 1 | 1–5 | Database: migrations, queries, sqlc generate |
| PR 2 | 6–10 | Notification engine: channels, dispatcher, severity/cooldown logic |
| PR 3 | 11–13 | Integration: wire dispatcher into monitor loop + config |
| PR 4 | 14–15 | API: notification CRUD endpoints + test endpoint |
| PR 5 | 16–19 | Frontend: settings page, components, dashboard card |
| PR 6 | 20 | Tests |

---

## Estimated File Changes

| Action | File |
|--------|------|
| Create | `backend/migrations/017_notification_channels.up.sql` |
| Create | `backend/migrations/017_notification_channels.down.sql` |
| Create | `backend/migrations/018_notification_preferences.up.sql` |
| Create | `backend/migrations/018_notification_preferences.down.sql` |
| Create | `backend/migrations/019_notification_log.up.sql` |
| Create | `backend/migrations/019_notification_log.down.sql` |
| Create | `backend/internal/db/queries/notification_channels.sql` |
| Create | `backend/internal/db/queries/notification_preferences.sql` |
| Create | `backend/internal/db/queries/notification_log.sql` |
| Create | `backend/internal/notifications/notifier.go` |
| Create | `backend/internal/notifications/email.go` |
| Create | `backend/internal/notifications/slack.go` |
| Create | `backend/internal/notifications/discord.go` |
| Create | `backend/internal/notifications/format.go` |
| Create | `backend/internal/notifications/notifier_test.go` |
| Create | `backend/internal/notifications/slack_test.go` |
| Create | `backend/internal/notifications/discord_test.go` |
| Create | `backend/internal/notifications/email_test.go` |
| Create | `backend/internal/api/handlers/notifications.go` |
| Create | `backend/internal/api/handlers/notifications_test.go` |
| Create | `frontend/src/pages/NotificationsPage.vue` |
| Create | `frontend/src/components/notifications/NotificationChannelForm.vue` |
| Create | `frontend/src/components/notifications/NotificationChannelCard.vue` |
| Create | `frontend/src/components/notifications/NotificationHistoryTable.vue` |
| Create | `frontend/src/api/notifications.ts` |
| Create | `frontend/src/types/notifications.ts` |
| Edit | `backend/internal/agent/agent.go` (add notifier field) |
| Edit | `backend/internal/agent/monitor.go` (call notifier after emit) |
| Edit | `backend/internal/agent/emit.go` (return agent_log ID) |
| Edit | `backend/internal/config/config.go` (Resend env vars) |
| Edit | `backend/internal/api/router.go` (notification routes) |
| Edit | `backend/internal/api/handlers/server.go` (if handler struct needs update) |
| Edit | `backend/cmd/heimdall/main.go` (create dispatcher, pass to agent) |
| Edit | `frontend/src/router/index.ts` (notifications route) |
| Edit | `frontend/src/components/common/AppSidebar.vue` (nav item) |
| Edit | `frontend/src/pages/DashboardPage.vue` (notification card) |
| Regenerate | `backend/internal/db/*.sql.go` (sqlc generate) |

---

## Decisions

1. **Email provider.** Resend — simple REST API, free tier (100/day), no SMTP config. Provider-agnostic via the `Channel` interface; swappable without touching dispatch logic.

2. **No webhook signature verification (inbound).** This phase is outbound-only — Heimdall sends notifications, it doesn't receive them. No webhook signature verification needed.

3. **Cooldown scope.** Per-app, not per-channel. If an app triggers a notification, *all* channels are suppressed for the cooldown period. This is simpler and prevents alert fatigue. Per-channel cooldown can be added later if users want different cadences per channel.

4. **Single retry.** One retry on failure, then mark as failed. No exponential backoff, no retry queue. Monitoring is continuous — if a notification fails, the next monitoring cycle will generate a new one if the issue persists. Over-engineering retries adds complexity without meaningful benefit at this stage.

5. **No acknowledgement/escalation (deferred).** The roadmap mentions "re-notify if unacknowledged for N minutes." This requires read-receipt tracking and a separate escalation timer — significant complexity. Deferred to a future phase. The cooldown mechanism provides basic noise reduction for now.

6. **Notification channel config in JSONB.** Each channel type has different config (email needs recipients, Slack/Discord need webhook URLs). JSONB avoids schema proliferation while keeping config queryable. Validated at the handler level.

7. **Dashboard integration is lightweight.** A small card showing notification status and recent history — not a full notification center. Keeps the dashboard focused.

---

## Deferred to Future Phases

| Priority | Item | Notes |
|----------|------|-------|
| P1 | Acknowledgement + escalation rules | Re-notify if critical goes unacknowledged for N minutes |
| P1 | PagerDuty integration | Dedicated incident management channel |
| P2 | Per-channel cooldown | Different suppression windows per channel |
| P2 | Notification templates | User-customizable message format |
| P2 | Notification batching | Group multiple findings into a single digest |
| P3 | Webhook channel (generic) | User-defined HTTP POST to arbitrary URLs |
| P3 | SMS notifications | Via Twilio or similar |
