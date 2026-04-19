# Drainage — Architecture Proposal

## Context

Heimdall prunes `log_buffer` after 48 hours (`backend/internal/agent/pruner.go:46-55`, `log_buffer.sql:85`). That's fine as a default — it keeps the SaaS DB bounded and respects the "monitoring, not archive" framing from `docs/vision.md`. But users whose compliance or retrospective-analysis needs exceed 48h need somewhere to push the data.

**Drainage** is the outbound path that mirrors logs, assessments, and investigations to a destination the user controls — their warehouse, their S3 bucket, their SIEM, their own Postgres. Once there, Heimdall's 48h window is a cache, not a ceiling. This is a greenfield subsystem — no existing code to migrate, no tables to preserve.

This doc assumes Option B in `notifications.md` (introducing an `OutboundConnector` abstraction). Drainage is the clean-slate motivator for that abstraction.

---

## Shape of the problem

### What gets drained

Three payload categories, each with a natural cursor:

| Kind | Source | Cursor | Volume estimate |
|---|---|---|---|
| **Logs** | `log_buffer` | `ingested_at` | High — every ingested line |
| **Pipeline events** | `log_pipeline_events` (new in 038) | `created_at` | Up to 4× logs (ingest/classify/gate/assess) |
| **Investigations** | `investigations` + attached `conversations` + `agent_log` entries | `investigations.updated_at` | Low — hundreds per day at most |

Pipeline events are optional to drain (they're large and per-log derivatives), but they're where the **story of each log** lives. A user building their own dashboard will want them. We should drain them by default and offer an opt-out.

### Delivery destinations (initial catalogue)

Ranked by likely demand + implementation cost:

1. **Generic HTTPS webhook** — simplest, plugs into the user's own ingestion. Batched POSTs of NDJSON payloads. Low effort, covers 80% of asks.
2. **S3 / S3-compatible object storage** (R2, B2, MinIO) — hourly or size-triggered rollups as newline-delimited JSON files. Standard for compliance archives.
3. **Postgres (user-owned)** — write directly into a customer DB. Natural synergy with the BYODB work in `enterprise-archi.md` but requires DSN management and schema versioning on a DB we don't own.
4. **Kafka / Redpanda / NATS** (phase 2+) — streaming pipe for users who have a real-time warehouse.

Start with **webhook + S3** in phase 1. Postgres sink in phase 2. Kafka/message buses are explicitly out of scope until someone asks.

### Delivery semantics

- **At-least-once with idempotency keys.** Every payload carries a stable key (`log_buffer.id`, `log_pipeline_events.id`, `investigations.id` + `updated_at`). Consumers dedupe.
- **Cursor persisted per `outbound_connection`.** If the worker crashes mid-batch, it resumes from the last acknowledged cursor. No silent gaps.
- **Bounded backpressure.** A destination that's been failing for 24h stops retrying, raises a `drainage_paused` event in the Activity feed, and requires user acknowledgement. Prevents a dead Slack from tying up the worker.
- **Ordered within a batch, not across batches.** Logs within one POST are ordered by `ingested_at`; two batches racing is acceptable because consumers key on the id.

---

## Proposal

### Tables

```sql
-- 039_outbound_connections.up.sql
CREATE TABLE outbound_connections (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  app_id          uuid NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
  kind            text NOT NULL CHECK (kind IN ('notify','drain')),
  type            text NOT NULL,         -- 'webhook' | 's3' | 'postgres' | ...
  name            text NOT NULL,
  config          jsonb NOT NULL,        -- type-specific (URL, bucket, DSN, etc.)
  enabled         boolean NOT NULL DEFAULT true,
  drain_scopes    text[] NOT NULL DEFAULT ARRAY['logs','pipeline_events','investigations'],
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

-- 040_drainage_state.up.sql
CREATE TABLE drainage_cursors (
  connection_id   uuid NOT NULL REFERENCES outbound_connections(id) ON DELETE CASCADE,
  scope           text NOT NULL,         -- 'logs' | 'pipeline_events' | 'investigations'
  last_cursor_ts  timestamptz NOT NULL,
  last_cursor_id  uuid,                  -- tiebreaker for same-timestamp rows
  updated_at      timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (connection_id, scope)
);

CREATE TABLE drainage_delivery_log (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id   uuid NOT NULL REFERENCES outbound_connections(id) ON DELETE CASCADE,
  scope           text NOT NULL,
  started_at      timestamptz NOT NULL,
  finished_at     timestamptz,
  row_count       int NOT NULL,
  status          text NOT NULL,         -- 'ok' | 'retry' | 'failed' | 'paused'
  error           text,
  cursor_advanced boolean NOT NULL DEFAULT false
);
```

Why three tables and not one jumbo audit log: cursors are read-heavy hot-path (every worker tick), delivery log is write-heavy and occasionally queried. Separating them keeps the cursor lookup a pk hit.

### Interface

```go
// backend/internal/connectors/outbound.go
type DrainConnector interface {
    OutboundConnector  // see notifications.md Option B
    Drain(ctx context.Context, batch DrainBatch) error
}

type DrainBatch struct {
    Scope    string          // 'logs' | 'pipeline_events' | 'investigations'
    AppID    uuid.UUID
    Rows     []json.RawMessage
    MinTS    time.Time
    MaxTS    time.Time
}
```

Each destination type (`webhook`, `s3`, `postgres`) implements `Drain` according to its transport rules. Webhook POSTs NDJSON; S3 writes an object keyed `app_id/scope/YYYY/MM/DD/HH/maxts.ndjson`; Postgres does an `INSERT … ON CONFLICT DO NOTHING`.

### Worker

New goroutine started from `cmd/heimdall/main.go` alongside the existing monitor and pruner:

```
drainage.Worker.Run(ctx):
  every 30s (configurable):
    for each enabled outbound_connection where kind='drain':
      for each scope in conn.drain_scopes:
        cursor := load drainage_cursors
        rows   := query rows since cursor (limit = batch_size)
        if empty: continue
        try Drain(batch):
          on ok:    advance cursor, log delivery ok
          on error: log delivery retry, exponential backoff, pause after N failures
```

Batch size defaults to 500 rows per scope per tick. Tick interval defaults to 30s. Both per-connection overridable.

**Interaction with pruner**: drainage must complete before pruning for a given window. Either:
- (a) Delay the pruner by one drainage tick (simple, safe, ~30s buffer), or
- (b) Query logs by `ingested_at < now() - 47h 55m` for the drainage cursor, giving drainage a 5m head start.

(b) is cleaner — the pruner stays on its hourly cadence and doesn't need to coordinate with anything. Proposed.

### API surface

Mirrors the existing notifications CRUD so the frontend doesn't learn two patterns:

```
GET    /api/apps/{appId}/outbound                  — list all (notify + drain)
POST   /api/apps/{appId}/outbound                  — create
GET    /api/apps/{appId}/outbound/{id}             — fetch
PUT    /api/apps/{appId}/outbound/{id}             — update
DELETE /api/apps/{appId}/outbound/{id}             — delete
POST   /api/apps/{appId}/outbound/{id}/test        — send a canned batch
GET    /api/apps/{appId}/outbound/{id}/deliveries  — paginated delivery_log
```

If you'd rather keep notifications on its existing endpoints for now (Phase 1 of the Option B staging in `notifications.md`), drainage gets its own `/api/apps/{appId}/drainage` subtree that's structurally identical.

### Frontend

A new "Drainage" sub-page under the app surface (or a tab inside "Notifications" renamed to "Outbound"). Wizard flow matches the inbound connection wizard — pick destination type, fill config, hit Test, save.

The delivery log viewer is the killer feature for operators: last 100 batches per connection, with row count, duration, cursor position, and error if any. This is the "did my drainage actually work" answer they'll want.

---

## Enterprise / BYODB implications

`docs/executing/enterprise-archi.md` describes splitting control vs data plane. Drainage is unambiguously data-plane: the cursors, delivery log, and source tables all live next to `log_buffer`. The `outbound_connections` table has a tougher question — config rows (including secrets like S3 keys) are operator-managed state, which feels control-plane, but they're scoped to an app, which is data-plane.

**Proposal:** put `outbound_connections` in data-plane alongside `notification_channels` (both store destination credentials today; they've already made the same call for notifications). Encrypt `config.secrets_*` fields at rest — see open question below.

When BYODB lands, a drainage cursor is *per-customer-DB*, not global — the worker becomes "for each customer DB, tick each connection." This is naturally expressed if `drainage.Worker` takes a pool-getter function (`func(appID) *pgxpool.Pool`) rather than a single pool. Build it that way from day one.

---

## Phased plan

**Phase 1 — Plumbing + webhook + S3 (this release)**
- Migrations 039/040
- `DrainConnector` interface
- Worker goroutine with cursor tracking + retry
- `webhook` and `s3` destinations
- HTTP handlers (list/CRUD/test/deliveries)
- Basic frontend page (list, create, delete, delivery log)

**Phase 2 — Postgres destination + reliability polish**
- `postgres` destination (direct INSERT into customer DB)
- Dead-letter UI ("X deliveries paused, review")
- Per-scope enable/disable
- Schema doc for webhook payload shape (so consumers can code against it)

**Phase 3 — Investigation-shaped exports**
- Full investigation export includes the conversation and agent_log entries as a denormalised bundle (JSON tree, not three separate rows). Operators want one investigation = one row.

**Phase 4 — Streaming destinations (only if asked)**
- Kafka / Redpanda / NATS
- Real-time mode (sub-second batches instead of 30s ticks)

---

## Questions

1. **Secret management for destination configs.** S3 access keys, webhook bearer tokens, Postgres DSNs all live in `outbound_connections.config`. Today `notification_channels.config` stores Slack/Discord webhook URLs as plaintext JSONB. Do we want to introduce app-level envelope encryption (e.g. per-org KMS key) before drainage multiplies the attack surface, or accept parity with notifications for now and fix both together later?
2. **What's the "investigation" export boundary?** Options: (a) investigations row only; (b) investigation + its conversation messages; (c) investigation + conversation + every `agent_log` entry tagged with that `conversation_id`. (c) is most useful, (a) is cheapest. Probably (c) with a scope flag, but confirm.
3. **Is retention beyond 48h a destination responsibility or a Heimdall guarantee?** Once data is drained, Heimdall's job is done — we don't track whether the user's S3 bucket still has it. But a user might ask "I want you to keep 30 days in Heimdall **and** mirror to S3." That's a retention-knob question, not a drainage question, and I'd keep them separate: extend `log_buffer` retention via a separate `log_retention_hours` setting; drainage is orthogonal.
4. **Does drainage replay on backfill?** If a user adds a new drainage connection today, do we backfill the last 48h immediately, or only drain from "now" forward? Backfill is trivially supported by seeding `drainage_cursors.last_cursor_ts = now() - 48h` on create — the question is default behaviour. Proposed default: **backfill on create** (users expect everything-so-far); opt-out via a `start_from_now` flag in the create request.
5. **Rate limiting on outbound HTTP.** A webhook destination that can only absorb 100 req/s with 500-row batches = 50k logs/s headroom, which is plenty. But should batch size be negotiated per-connection, or a fixed global default? Per-connection override is cheap and inevitable; let's plan for it from the schema (already modellable in `config`).
6. **Does drainage count against the user's API quota or billing tier?** Almost certainly yes eventually. Worth a placeholder in `drainage_delivery_log.row_count` so metering has something to aggregate on from day one, even if we don't meter it yet.
7. **Observability of the drainage worker itself.** The worker is a background goroutine like the monitor and pruner. We have structured logs for those; should drainage emit its own `agent_log` entries (so it shows up in the Activity feed), or just rely on `drainage_delivery_log` + app logs? Activity feed visibility is probably right for "drainage paused" events, silent for normal ticks.
