# Notifications — Architecture Proposal

## Context

0.15.0 shipped a working notifications subsystem. What's outstanding is less "build it" and more "turn it into a first-class escalation product." This doc takes stock of what's there, identifies the gaps, and proposes where it should live architecturally relative to the new **outbound connection** concept the user is introducing (alongside Drainage — see `drainage.md`).

### What already exists

Tables (`backend/migrations/017…019`):
- `notification_channels` — per-app channel rows (type + JSONB config + enabled flag)
- `notification_preferences` — per-app severity threshold + cooldown minutes
- `notification_log` — delivery audit (status, optional `agent_log_id` back-ref)

Code (`backend/internal/notifications/`):
- `notifier.go:48-127` — `Dispatcher` with severity gate, cooldown gate, fan-out to enabled channels
- `email.go` / `slack.go` / `discord.go` — three `Channel` implementations (Resend for email, webhooks for the other two)
- `format.go:1-57` — severity emoji/colour, email HTML template, subject builder

Integration:
- The monitoring loop calls `Dispatcher.Notify()` after `EmitLog` writes an assessment — **fire-and-forget**, `context.WithoutCancel()` so a closing monitor ctx doesn't kill delivery
- HTTP layer: `backend/internal/api/handlers/notifications.go:1-482` — full CRUD + `/test` send + paginated history
- Frontend: `NotificationsPage.vue` with preferences, channels, and history sub-components

### What's missing or weak

1. **No retry or queue.** A transient Slack 429 or SMTP hiccup results in a single logged failure and the notification is dropped. `notification_log` records it but nothing re-attempts.
2. **Cooldown is app-wide, not per-channel.** One noisy hour suppresses every channel identically — fine for email but wrong if one channel is "pager-duty-critical" and another is "background discord feed."
3. **No escalation rules.** The only gate is `severity >= threshold`. Users can't say "route warnings to Slack, errors to email + Slack, criticals to PagerDuty."
4. **No digesting / batching.** Every flagged assessment dispatches immediately. A storm fires N individual messages.
5. **No templates.** Channel message format is hardcoded in `format.go`.
6. **Disjoint from the `connectors/` abstraction.** All inbound data sources go through `backend/internal/connectors/` with a `Connector` interface and a factory. Notifications are a parallel hierarchy.

---

## Proposal

Two viable shapes. I lean toward **Option B** but flag the cost up front so you can decide.

### Option A — Keep notifications standalone, polish what's there

Treat notifications as its own bounded subsystem and close the gaps listed above without merging into a connector abstraction.

**Changes:**
- Add a `notification_dispatch_queue` table; Dispatcher enqueues, a worker consumes with exponential backoff and a `max_attempts` cap. Dead-lettered rows stay visible in the history UI.
- Extend `notification_channels` with `cooldown_minutes` (override app default) so cooldown can be per-channel.
- Introduce a simple rule model: `notification_rules (id, app_id, severity_min, severity_max, channel_ids[])`. Replace the Dispatcher's "all enabled channels ≥ threshold" fan-out with "evaluate rules, union matched channels."
- Optional: a digest mode — `notification_preferences.digest_interval_minutes`; when non-zero, the worker batches pending payloads into one message per channel per interval.

**Cost:** small, contained. Existing code stays. No mental-model shift.
**Downside:** the user's ask is "2 outbound connections" — treating notifications as one of two outbound categories. Option A preserves the parallel hierarchy that directly contradicts that framing.

### Option B — Introduce `OutboundConnector` and make notifications one flavour of it

Create the new abstraction for drainage (which has no existing code to migrate), then bring notifications in via an adapter so the two concepts live under one interface.

**New interface** (`backend/internal/connectors/outbound.go`):
```go
type OutboundConnector interface {
    Connector            // existing base: ID, Type, Close
    Kind() OutboundKind  // "notify" | "drain"
    Send(ctx context.Context, payload OutboundPayload) error
}

type OutboundPayload struct {
    Kind    string          // "assessment" | "log_batch" | "investigation"
    Subject string          // rendered summary (for notify channels)
    Body    json.RawMessage // typed body per kind
}
```

**Migration path for existing notifications:**
- Keep the `notification_channels` / `notification_preferences` / `notification_log` tables exactly as they are.
- Add a new column or a new sibling table `outbound_connections(id, app_id, kind, type, config, enabled)`; backfill it from `notification_channels` with `kind = 'notify'`.
- Wrap each existing `Channel` in a thin `OutboundConnector` adapter — `notify:email`, `notify:slack`, `notify:discord`.
- Dispatcher becomes a thin filter: reads `outbound_connections` where `kind = 'notify'`, applies severity/cooldown/rule gates, calls `Send()`.
- The old handler surface stays; internally it writes to the unified table.

**Benefit:** one mental model for "things Heimdall sends outward." Drainage slots in as `kind = 'drain'` with the same list/create/delete/test surface on the frontend (tabbed: Notifications | Drainage). Future kinds (e.g. PagerDuty, OpsGenie, S3 mirror) are additive.

**Cost:** one migration to unify tables (can be done as dual-write first, flip second release), one adapter layer, one handler consolidation. Existing UI barely changes — just gains a "Drainage" tab.

### Recommendation

Option B, **staged**:
1. **Phase 1 (this release):** Close the polish gaps from Option A — retry queue, per-channel cooldown, rules, optional digest. Doesn't need the unification to land; users want the reliability now.
2. **Phase 2 (next release):** Land `OutboundConnector` interface + `outbound_connections` table for **drainage only** (see `drainage.md`). Notifications stay on their own tables.
3. **Phase 3 (opt-in, maybe never):** Migrate notifications into the unified table via dual-write → cutover → drop old tables. Only worth it if a third outbound category appears.

This sequencing means Phase 2 gets the clean abstraction without paying the cost of migrating working code, and you can still tab the frontend as "Notifications | Drainage" from Phase 2 (the frontend doesn't care that they're backed by two tables).

---

## Questions

1. **Are PagerDuty / OpsGenie / generic webhook channels in scope?** These are the obvious next channel types and would validate the rule model before we ship it. If yes, they should be Phase-1 scoped or left explicitly for Phase 4.
2. **Per-rule cooldown vs. per-channel cooldown?** The rule model proposed above gates on severity; cooldown is per-channel. A more expressive model is per-rule cooldown ("burst to Slack every time, but cap email at one per hour"). More power, more UI complexity.
3. **What's the fate of `notification_log.agent_log_id`?** Today it's nullable and sometimes empty. If we tighten it to required, we gain end-to-end traceability ("which assessment caused this notification"), but we need to audit every `Notify()` call site first.
4. **Digest mode behaviour during incidents.** If the user has digest-every-15-min and a critical fires, do we break the digest window and send immediately? Severity-based "digest unless severity ≥ X" is probably the right default, but worth confirming.
5. **Frontend pattern if Option B stages as proposed.** Do you want a single "Outbound" section in the sidebar with two tabs from day one of Phase 2, or keep "Notifications" as its own sidebar item and add "Drainage" as a separate one?
