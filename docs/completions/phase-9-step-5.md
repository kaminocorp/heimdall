# Phase 9, Step 5 — Notification Channels, Dispatcher & Config

**Status:** Complete
**Date:** 2026-03-11
**Plan reference:** [Phase 9 — Notifications & Escalation](../executing/phase-9-notifications.md) (Steps 4–9, 11)

---

## What was done

Created the `backend/internal/notifications/` package with the channel interface, three channel implementations (Slack, Discord, Email), message formatting, and the central dispatcher. Also added Resend env vars to the config.

## Files created

| File | Purpose |
|------|---------|
| `backend/internal/notifications/notifier.go` | `Channel` interface, `NewChannel` factory, `Dispatcher` (severity filter, cooldown, fan-out, retry) |
| `backend/internal/notifications/slack.go` | Slack incoming webhook — Block Kit payload (header, section, context) |
| `backend/internal/notifications/discord.go` | Discord webhook — embed with severity color, title, description, footer |
| `backend/internal/notifications/email.go` | Resend API — HTML email with monospace template |
| `backend/internal/notifications/format.go` | Shared helpers: `SeverityEmoji`, `SeverityColor`, `FormatSubject`, `FormatEmailHTML` |

## Files edited

| File | Change |
|------|--------|
| `backend/internal/config/config.go` | Added `ResendAPIKey` and `NotificationFromEmail` fields + env var loading |

## Architecture

```
Dispatcher.Notify(ctx, appID, agentLogID, appName, severity, summary, assessment)
    │
    ├─ GetNotificationPreferences → enabled? severity ≥ threshold?
    ├─ GetLastNotificationForApp  → within cooldown window?
    ├─ ListEnabledChannelsByApp   → fan-out to each channel
    │
    └─ dispatchToChannel(ch)
         ├─ InsertNotificationLog (status: pending)
         ├─ NewChannel(type, config) → Channel interface
         ├─ channel.Send(payload) → retry once on failure
         └─ UpdateNotificationLogStatus (sent | failed)
```

## Design decisions

- **`Channel` interface with factory** — Same pattern as `Classifier` in `agent/classifier.go`. Each implementation is isolated; testing uses mock channels.
- **`postWebhook` shared helper** — Slack and Discord both POST JSON to a webhook URL. The shared function in `slack.go` avoids duplication. Discord reuses it.
- **`truncate` for message limits** — Slack blocks have a 3000-char limit, Discord embeds 4096. Messages are truncated with `...` to avoid API rejections.
- **Config passed to factory** — Only the email channel needs server-side config (`ResendAPIKey`, `NotificationFromEmail`). Slack/Discord config is entirely per-channel (webhook URLs in JSONB). The `*config.Config` is threaded through `NewChannel` so email can access it without env var lookups at send time.
- **Fire-and-forget with single retry** — Matches the `EmitLog` pattern: notification failures never block the monitoring loop. One retry covers transient network issues; persistent failures are logged and visible in `notification_log`.
- **Resend env vars are optional** — Email channel returns an error if they're missing, but the app starts fine without them. Slack and Discord work with zero server-side config.

## Next step

Step 6 — Wire the dispatcher into the monitor loop (`agent.go`, `monitor.go`, `emit.go`, `main.go`).
