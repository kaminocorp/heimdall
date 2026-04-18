# Pipeline Page — Implementation Plan

Visualise the five-stage log pipeline (Ingestion → Lumber → Gate → Agent → Activity) as an app-scoped Sankey/funnel diagram with a live SSE feed and a persisted per-log event trail for future replay ("Time Machine").

This plan follows and replaces the proposal in `pipeline-page.md`. Decisions made:

- **Proposal B (Sankey/funnel)** chosen over A and C.
- **App-scoped**, matching Connections/Agent Config/Activity.
- **Mobile deferred** — platform-wide overhaul will address it later.
- **Time Machine stubbed on the frontend** from day one, backend fully populated so no backfill gap exists when the UI lands.
- **Naming**: "Pipeline", top-level nav item between Connections and Agent Config.
- **Rendering**: SVG + Canvas2D for v1. WebGL/WebGPU deferred — see "Rendering tech trade-offs" below.
- **Rule-hit telemetry**: hardcoded escalation rules gain stable rule IDs (Option A). Alternatives — DB-backed rules (Option B), expression engine (Option C) — documented in `pipeline-rule-options.md` for future reference.
- **Hydration**: bundled `/bootstrap` endpoint (stats + last N events + cursor in one roundtrip) + SSE with `?since=<cursor>` replay + client ID-based dedupe. See Phase 1.6 and Phase 2.3.
- **Source-filtering interaction**: the funnel shows *what Heimdall processes*, not *what was sent*. Ingestion events emit only for logs that pass source filtering and land in `log_buffer`. See "Interaction with source filtering" below.

---

## Rendering tech trade-offs (why SVG/Canvas2D, not WebGL/WebGPU)

Downsides of jumping to WebGL/WebGPU for v1:

1. **Text rendering is painful.** WebGL has no native text — you either pre-render glyph atlases or overlay DOM/SVG text. Our node cards, stats, and live-log ticker are text-heavy; a mixed SVG-text / WebGL-particle stack introduces two coordinate systems that must stay aligned under resize and zoom.
2. **Accessibility regression.** SVG nodes are DOM elements — screen readers, keyboard focus, and right-click inspection work for free. WebGL is a canvas blob. For an audit/monitoring surface that users may need to inspect and copy from, that matters.
3. **Debug + devtool loss.** When a particle animates wrong in SVG, you inspect the DOM. In WebGL you reach for a shader debugger. On a small team that's a real velocity tax.
4. **WebGPU browser support** is still uneven (Safari shipped only recently; older iPads lag). Even Heimdall's desktop-first stance doesn't let us ignore Safari.
5. **It's premature.** Heimdall's per-app throughput is realistically tens-per-second for most users, hundreds-per-second at the very top end. SVG with ~500 concurrent animated elements is fine on any 2020+ laptop. We don't have a perf problem to solve yet.

**Upsides** (why we keep the door open): GPU-accelerated particle counts in the 10k+ range, smoother animation under pathological load, cooler demos.

**Decision**: Build v1 on SVG (Sankey paths, node cards) + Canvas2D (particle layer, overlaid). Sample to 1-particle-per-N at high volume. Revisit WebGL only if real user telemetry shows p95 frame time > 16ms under normal load. This keeps the upgrade path open — the Canvas2D particle layer is the exact surface we'd swap for WebGL later, without touching SVG/Sankey geometry.

---

## Architecture overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        PIPELINE PAGE                            │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Sankey Funnel (SVG) — stages + proportional stream widths│  │
│  │  + Canvas2D particle overlay                              │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Time Machine block (stubbed v1)  ← click → replay (v2)  │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Live Log Ticker (bottom) — per-log journey, click-to-   │  │
│  │  expand granular detail panel                            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
        ▲ SSE stream (live events)         ▲ REST (stats snapshot,
        │                                   │  replay query)
        │                                   │
┌───────┴───────────────────────────────────┴───────────────────┐
│  Backend                                                       │
│                                                                │
│  PipelineBus (in-mem pub/sub per app)                          │
│       ▲                                                        │
│       │ publish events                                         │
│       │                                                        │
│  IngestWebhookLogs  ──▶  monitorApp ──▶ (Lumber) ──▶ (Gate)    │
│       │                      │             │           │       │
│       ▼                      ▼             ▼           ▼       │
│  log_pipeline_events (persisted; ON DELETE CASCADE on log_id)  │
└────────────────────────────────────────────────────────────────┘
```

Two data paths, one source of truth:

- **Live path** — `PipelineBus` emits events to SSE subscribers. Ephemeral.
- **Replay path** — same events written to `log_pipeline_events`, keyed by `log_id`. Cascading delete tied to `log_buffer` retention (currently 48h).

---

## Interaction with source filtering

The source-filtering plan (`source-filtering-and-org-connections.md`, Phase 1 shipped in migration 034) drops entries at webhook ingestion when no `app_source_filters` row with `enabled = true` matches the source name. **Filtered entries never land in `log_buffer`** — they have no `log_id`, so they cannot appear in `log_pipeline_events`.

Decisions:

1. **The funnel shows *what Heimdall processes*, not *what was sent*.** The Ingestion stage emits one event per log that *passes* the source filter and gets inserted into `log_buffer`. Filtered-out entries are out of scope for the per-log journey — they have no identity to replay.

2. **Pre-ingestion drops surface as a separate counter, not a funnel stage.** The Ingestion node card shows a secondary muted stat: *"N filtered from disabled sources in the last hour"*. This mirrors the "Showing X filtered entries/hr from disabled sources" indicator already in the source-filtering UI, and is sourced from a lightweight in-memory counter incremented in the webhook handler at the same point where filtering happens. Adding a fifth funnel stage would (a) require recording filtered entries in a new table (no `log_id` to reuse), (b) blur the funnel's meaning. A sibling counter keeps the funnel crisp.

3. **Org-scoped connections fan-out cleanly.** When Phase 2 of source-filtering ships, a single webhook batch can fan-out into N `log_buffer` rows (one per app with the source enabled). Each insert produces its own `log_id` and its own Ingestion event per app. The Pipeline page is app-scoped, so each app's funnel naturally shows only its share. No changes needed to the pipeline event schema.

4. **New-user empty state.** With drop-by-default source filtering, a freshly onboarded user will see zero Ingestion events until they enable sources. The Pipeline page's empty state should explicitly say "No sources enabled yet — [Manage sources]" rather than a generic "waiting for logs" spinner, to avoid the mystery of "the drain is configured but nothing arrives."

5. **Migration numbering.** Source-filtering Phase 1 shipped as migration 034. Phase 2 has reserved 035. The Pipeline's `log_pipeline_events` migration is therefore numbered **036**. See Phase 1.1.

---

## Phase 1 — Backend: Persistence + Pub/Sub

**Goal:** Every log flowing through the pipeline publishes live events AND persists its full stage history. Replay-capable from day one.

### 1.1 Migration `036_log_pipeline_events`

**New file:** `backend/migrations/036_log_pipeline_events.up.sql`

> **Migration numbering:** 033 (`rls_webhook_idempotency`) shipped in v0.45.1. 034 (`source_filtering` Phase 1) shipped with `source-filtering-and-org-connections.md`. 035 is reserved by `source-filtering` Phase 2 (org-scoped connections). Pipeline therefore takes **036**. If the Pipeline lands before source-filtering Phase 2, renumber at merge time — migration numbers are cheap to bump.

```sql
CREATE TABLE log_pipeline_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_id          UUID NOT NULL REFERENCES log_buffer(id) ON DELETE CASCADE,
    app_id          UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    stage           TEXT NOT NULL,      -- 'ingestion' | 'classified' | 'gate' | 'assessment'
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Stage-specific fields (nullable; filled per stage):
    source_type     TEXT,               -- ingestion
    severity        TEXT,               -- ingestion
    type            TEXT,               -- classified (Lumber Type)
    category        TEXT,               -- classified (Lumber Category)
    confidence      DOUBLE PRECISION,   -- classified
    summary         TEXT,               -- classified (Lumber summary text)
    escalated       BOOLEAN,            -- gate
    rule_hit        TEXT,               -- gate (which escalation rule fired)
    assessment_id   UUID REFERENCES agent_log(id) ON DELETE SET NULL, -- assessment
    -- Raw event payload for forward-compat (rare extra fields)
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_lpe_log_id ON log_pipeline_events (log_id, occurred_at);
CREATE INDEX idx_lpe_app_occurred ON log_pipeline_events (app_id, occurred_at DESC);
CREATE INDEX idx_lpe_stage ON log_pipeline_events (app_id, stage, occurred_at DESC);

-- RLS: follow the system-table pattern (migration 030). Enabled, no policies.
-- Only the owner-role pool (via s.Queries) writes/reads this table.
ALTER TABLE log_pipeline_events ENABLE ROW LEVEL SECURITY;
```

**Down migration** drops the table and indexes.

**Cascading cleanup**: `ON DELETE CASCADE` on `log_id` means the existing `PruneExpiredLogs` in `agent/pruner.go` automatically sheds pipeline events in lockstep. No new pruner code. `assessment_id` uses `ON DELETE SET NULL` so an assessment being deleted doesn't orphan the earlier stage events — the log's journey still reconstructs up to the Gate.

### 1.2 sqlc queries

**New file:** `backend/internal/db/queries/log_pipeline_events.sql`

Queries:
- `InsertPipelineEvent` — single insert
- `InsertPipelineEventsBatch` — batch insert for assessment stage (one event per flagged log in the batch)
- `GetPipelineEventsByLog` — all stage events for a single `log_id`, ordered by `occurred_at`
- `GetRecentPipelineEventsByApp` — paginated recent events for an app (used by Time Machine list view and by `/bootstrap` for ticker hydration). Accepts `@limit` and optional `@before` cursor.
- `GetPipelineEventsSince` — events with `occurred_at > @since` for a given app, ordered ascending. Used by the SSE endpoint's catch-up replay to close the hydration-to-stream gap. Capped at e.g. 500 rows; if the result hits the cap the stream responds with a `resync` frame telling the client to refetch `/bootstrap`.
- `PipelineStatsByApp` — time-windowed aggregates: counts per stage, % flagged, avg confidence. Accepts `@since` param.

Run `make sqlc-generate`.

### 1.3 PipelineBus

**New file:** `backend/internal/agent/pipeline_bus.go`

```go
type PipelineEvent struct {
    LogID        uuid.UUID
    AppID        uuid.UUID
    Stage        string
    OccurredAt   time.Time
    // stage-specific fields — mirror DB columns
    SourceType   string
    Severity     string
    Type         string
    Category     string
    Confidence   float64
    Summary      string
    Escalated    bool
    RuleHit      string
    AssessmentID *uuid.UUID
}

type PipelineBus struct {
    mu   sync.RWMutex
    subs map[uuid.UUID]map[chan PipelineEvent]struct{} // appID → set of subscriber channels
}

func (b *PipelineBus) Subscribe(appID uuid.UUID) (<-chan PipelineEvent, func())
func (b *PipelineBus) Publish(evt PipelineEvent)
```

Buffered channels (cap 128) with non-blocking publish — if a slow subscriber fills its buffer, we drop for that subscriber, not for others. Log a warning counter so we can detect subscriber stalls.

**Pattern reference:** `agent/loop_stream.go` uses the same idiom for `AgentEvent` channels; follow that structure for consistency.

### 1.4 Event writer

**New file:** `backend/internal/agent/pipeline_writer.go`

A small helper that does two things for every event:
1. Write to `log_pipeline_events` via sqlc.
2. Publish to `PipelineBus`.

Order: **persist first, then publish**. If the DB write fails, we log and skip the publish — better to lose a live event than to show a user a particle for a log whose journey isn't recoverable later. Fire-and-forget at the call site (same pattern as `EmitLog`): errors logged, never propagated.

### 1.5 Instrumentation points

Four call sites add a `pipelineWriter.Write(evt)` call:

| Stage | File | When |
|-------|------|------|
| `ingestion` | `internal/api/handlers/webhooks.go` (or equivalent `IngestWebhookLogs` path) | After successful `InsertLog` DB write, one event per log |
| `classified` | `agent/monitor.go` / `agent/classifier_lumber.go` | After Lumber returns for each log |
| `gate` | `agent/severity_gate.go` | After gate decision for each classified log |
| `assessment` | `agent/monitor.go` | After assessment batch completes; one event per flagged log in the batch, carrying the resulting `agent_log.id` |

**Important:** The `assessment` event is written per-log, not per-batch. The UI needs to know *which* logs landed in which assessment — that's the spine of the replay view.

### 1.6 HTTP endpoints

**New handler file:** `backend/internal/api/handlers/pipeline.go`

- `GET /api/apps/{appId}/pipeline/bootstrap?window=1h&tickerLimit=50` — **bundled first-paint response**. Returns `{ stats, recentEvents: [...], cursor }` in a single roundtrip:
  - `stats` — aggregate snapshot from `PipelineStatsByApp(window)`.
  - `recentEvents` — last N events from `GetRecentPipelineEventsByApp(tickerLimit)` (N capped server-side at 200). Ordered newest-first for ticker render; the frontend can reverse in-memory if chronological order is needed elsewhere.
  - `cursor` — RFC3339 timestamp of the newest event in the payload, or "now" if `recentEvents` is empty.
  - Response is ETaggable per app+window; short 5s server-side cache (Phase 3.3). Gzip in transit.
  - This supersedes the originally planned `/stats`-only endpoint. One roundtrip, not two — the ticker can paint synchronously with the rest of the page instead of after a follow-up fetch.
- `GET /api/apps/{appId}/pipeline/stream?since=<cursor>` — SSE live stream with catch-up replay:
  1. On connect, if `since` is set, the server first drains `GetPipelineEventsSince(since)` and writes those frames (each tagged `event: replay`). This closes the window between `/bootstrap` read-time and SSE subscribe-time.
  2. After replay completes (or immediately if `since` is absent), subscribes to `PipelineBus` and writes live frames (`event: live`).
  3. Heartbeat frame every 15s so proxies don't idle-close.
  4. If the replay result hits its cap (500 events — implies the client was disconnected too long), emits one `event: resync` frame instructing the client to drop its buffer and refetch `/bootstrap`, then continues with live events.
  5. Auth: JWT via query-string token (SSE can't set headers), matching the existing WebSocket pattern.
  6. Closes on client disconnect or context cancel; bus subscription cleaned up promptly.
- `GET /api/apps/{appId}/pipeline/logs/{logId}/journey` — returns the ordered stage history for a single log (for Time Machine replay; frontend stub doesn't call it yet but the endpoint ships in Phase 1 so Phase 4 can light up instantly).
- `GET /api/apps/{appId}/pipeline/logs?since=&until=&limit=&cursor=` — paginated log-journey list for Time Machine picker.

All routes scoped via `authorizeApp` helper (follows existing pattern).

**Wire-up:** Register routes in `internal/api/server.go` under the `/api/apps/{appId}` subrouter.

### 1.7 Tests

- Unit test `PipelineBus` — fanout, slow-subscriber drop behaviour, unsubscribe cleanup.
- Integration test (requires `DATABASE_URL`) — insert a log, walk it through all four stages, assert `GetPipelineEventsByLog` returns 4 rows in order.
- Integration test — cascading delete: insert log + 4 events, delete log, assert 0 pipeline events remain.
- Handler test — `/pipeline/stats` returns correct aggregates on seeded data.

**Phase 1 deliverables**: migration + sqlc queries + bus + writer + 4 instrumentation points + 4 endpoints + tests. Live stream works end-to-end, persistence populated from the first deployed request, `/journey` endpoint returns real data.

---

## Phase 2 — Frontend: Live Sankey Funnel (v1 UI)

**Goal:** Ship the user-visible Pipeline page. Live flow works, stats populate, Time Machine block is present but stubbed.

### 2.1 Routing + nav

- New route `/pipeline` registered in `frontend/src/router/`.
- New top-level nav item between Connections and Agent Config in the top header.
- Active-app scoping: reads `currentAppId` from the `app` Pinia store; if unset, show the standard "select an app" empty state used on other per-app pages.

### 2.2 Types

**New file:** `frontend/src/types/pipeline.ts`

TypeScript mirrors of the backend event shape, stats shape, and journey shape. Stage is a string-literal union `'ingestion' | 'classified' | 'gate' | 'assessment'`.

### 2.3 API + composable

- `api/pipeline.ts` — `fetchPipelineBootstrap(appId, { window, tickerLimit })`, `fetchLogJourney(appId, logId)`, `fetchPipelineLogs(appId, params)`.
- `composables/usePipelineStream.ts` — the hydration-to-stream glue. Sequence:
  1. On mount, calls `fetchPipelineBootstrap` once. Seeds `stats` reactive snapshot and pushes `recentEvents` into the ring buffer (cap 500) **synchronously before first paint**. Captures `cursor`.
  2. Opens `EventSource` against `/pipeline/stream?since=<cursor>&token=<jwt>`. Handles three frame types:
     - `event: replay` — events missed between bootstrap and connect. Inserted via the same dedupe-by-id path as live events.
     - `event: live` — steady-state stream.
     - `event: resync` — drop the ring buffer, refetch `/bootstrap`, reconnect with the new cursor. Rare but load-bearing for long disconnects.
  3. Every insert into the ring buffer de-duplicates by `log_pipeline_events.id` via a small `Set<string>` of ids currently in the buffer. Eviction removes from both buffer and set. De-dupe is what makes the overlap between `recentEvents` and `replay` safe.
  4. `connectionState` reactive enum: `idle | bootstrapping | streaming | reconnecting | resyncing | error`.
  5. Auto-reconnect on disconnect with exponential backoff (1s → 30s cap). On reconnect, reuses the current `cursor` so no events are missed even across network flaps.
  6. Exposes: `events` (reactive ring buffer), `stats` (reactive snapshot), `connectionState`, `reconnect()`, `destroy()`.
  7. Token refresh: if auth expires mid-stream, `EventSource` errors → `reconnect()` grabs a fresh token from the auth store before re-opening.

### 2.4 Store

**New file:** `frontend/src/stores/pipeline.ts`

Pinia composition store holding:
- Current app's stats snapshot.
- Recent events ring buffer (for the live ticker). Backed by a fixed-capacity `Array<PipelineEvent>` + `Set<eventId>` for O(1) dedupe.
- Per-stage rolling rates (events-per-minute, recomputed on a 1s interval via `setInterval` that is cleared on store teardown — no orphaned timers on SPA navigation).
- Derived `flaggedRatio` / `safeRatio` for Sankey width calculation, memoised via Vue's `computed`.
- Per-log journey cache (LRU, cap 100) so clicking the same log twice in the ticker doesn't re-fetch `/journey`.
- Perf invariants documented inline: ring-buffer insert is O(1), dedupe check is O(1), stats update is snapshot-replace (no per-field diffing). No reactivity traps — events flow through a shallow ref to keep Vue from deep-watching hundreds of objects.

### 2.5 Components

| Component | Responsibility |
|-----------|----------------|
| `pages/PipelinePage.vue` | Route shell. Mounts composable, wires store, renders children. |
| `components/pipeline/PipelineFunnel.vue` | SVG Sankey. Computes per-segment widths from stats; renders 5 stage labels. Handles resize. |
| `components/pipeline/PipelineParticles.vue` | Canvas2D overlay, absolutely positioned over the SVG. Reads events from store, spawns particles along precomputed bezier centerlines, animates via `requestAnimationFrame`. Sampling (1 particle per N logs) kicks in when events-per-second > threshold (start at 50/sec). |
| `components/pipeline/PipelineNodeCard.vue` | Glass-morphism card with live stat + click-to-expand. Click opens the detail panel. |
| `components/pipeline/PipelineNodeDetail.vue` | Expandable panel per stage. Renders stage-specific content (see section 2.6). |
| `components/pipeline/PipelineLogTicker.vue` | Bottom live-log stream. Each row is a mini-journey: `12:04:32  fly/app  ERROR  connection_failure  → FLAGGED → Agent`. Click opens the per-log journey modal (Phase 4 stub in v1: opens modal with "Replay coming soon" placeholder + raw event list). |
| `components/pipeline/TimeMachineBlock.vue` | **Stubbed in v1.** Visible block near the top of the page with a date/time range picker control (disabled), a "Replay a past log's journey" label, and a tooltip: "Coming soon — every log's pipeline journey is already being recorded." The picker control is styled fully so Phase 4 just wires behaviour. |

### 2.6 Granular detail panels (auditing-grade)

Per-stage content when a node card is expanded:

- **Ingestion** — per-source breakdown (source_type → logs/min), format distribution ring, **last 20 raw payloads** (expandable to full JSON), filter by severity.
- **Lumber** — classification distribution bar chart (Type.Category histogram), confidence distribution histogram, **last 20 classified logs** with before/after (raw text → Type.Category + confidence + summary).
- **Gate** — escalation rules with hit counts (sourced from `rule_hit` aggregates), flagged % sparkline over the last hour, **last 20 gate decisions** with the rule that fired.
- **Agent** — recent assessments list (summary + auto-severity + linked assessment_id → deep link to Activity page), avg tool calls per assessment, avg latency. Each assessment expands to show **every flagged log** that fed into it (via `log_pipeline_events WHERE assessment_id = …`).
- **Activity** — severity distribution over the last 24h, links to the Activity page filtered by this app.

All "last 20" lists are click-to-expand to full raw data. Every log in every panel has a "View journey" button that opens the same per-log journey modal as the ticker.

### 2.7 Visual design

- Dark glass-morphism node cards, same treatment as `ConnectionBubble`.
- Sankey stream paths: gradient-filled SVG paths. Green tint pre-Gate, split at Gate into amber (flagged → Agent) and muted grey (safe → sink).
- Stream widths: calculated from 1-hour rolling stats, updated every 5s so the funnel breathes rather than jitters.
- Particles: 2–4px glowing dots, green pre-Gate, amber post-Gate-on-flagged path, faded grey on safe path.
- Dormant state: when events-per-second < 1 for 30s, particles drift slowly (mirror `AgentNebula` dormant mode).

### 2.8 Tests

- Vitest: `usePipelineStream` composable — event buffering, reconnect, rate calculation.
- Vitest: `pipeline` store — derived ratios, ring-buffer eviction.
- Component test: `PipelineFunnel` renders correct path widths for given stats.

**Phase 2 deliverables**: user-visible page at `/pipeline` with live flow, granular detail panels, stubbed Time Machine block, stubbed per-log journey modal.

---

## Phase 3 — Backfill guard + production hardening

**Goal:** De-risk the rollout.

### 3.1 Subscriber backpressure metrics

Counter logged per app: `pipeline_bus_dropped_events_total`. If we see drops in production, we know a subscriber is stuck.

### 3.2 SSE connection cap per user

Cap concurrent SSE subscriptions per user at e.g. 5 (one per app tab). Exceeding returns 429. Prevents a runaway tab spawn from pinning backend channels.

### 3.3 Stats endpoint caching

`/pipeline/stats` is called on every page load; `PipelineStatsByApp` scans `log_pipeline_events` with time-windowed aggregates. Cache the result per app for 5s in-memory. At 48h retention and our expected row counts this is fine without caching, but the hook is cheap to add and saves us from a future perf regression.

### 3.4 Abandoned-subscriber cleanup

If an SSE connection closes, the bus must drop the subscriber channel promptly. Unit test covers this; add a background sweeper that warns if the subscriber map has any app with > 50 live channels (leak canary).

### 3.5 RLS verification

Run the standard RLS test pass against `log_pipeline_events`. Follows the system-table pattern from migration 030 — owner-role pool only.

---

## Phase 4 — Time Machine (replay, future)

**Goal:** Light up the stubbed Time Machine block. Backend is already complete from Phase 1 — this is almost entirely a frontend phase.

### 4.1 Wire the date/time picker

The stubbed `TimeMachineBlock.vue` gets real behaviour: range picker → calls `/api/apps/{appId}/pipeline/logs?since=&until=` → renders a virtualised list of log journeys.

### 4.2 Per-log replay view

Clicking a log in the picker (or in the live ticker) opens a modal that:
1. Fetches `/pipeline/logs/{logId}/journey`.
2. Renders the **same Sankey funnel as the live view**, but with only this log's particle traced through it — each stage card lights up with this log's specific data (Lumber classification, gate decision, assessment link).
3. Bottom of modal: timeline view (one row per stage event with timestamp and full metadata), plus "View full payload" and "Open assessment" actions.

### 4.3 Journey URL

Make the journey view routable: `/pipeline/logs/:logId` for deep-linking from Activity page or Slack notifications.

### 4.4 Gap disclosure

Because `log_pipeline_events` inherits `log_buffer`'s 48h retention, the picker surfaces a clear "Only the last 48h of journeys are available — logs are pruned after this window" note so users aren't surprised by missing history. If long-term retention becomes a product need, we solve that by changing `log_buffer` retention, not by special-casing pipeline events.

---

## File summary

**New backend files:**
- `backend/migrations/036_log_pipeline_events.up.sql` + `.down.sql`
- `backend/internal/db/queries/log_pipeline_events.sql`
- `backend/internal/agent/pipeline_bus.go` + `_test.go`
- `backend/internal/agent/pipeline_writer.go`
- `backend/internal/api/handlers/pipeline.go` + `_test.go`

**Modified backend files:**
- `backend/internal/api/handlers/webhooks.go` — add ingestion event emit
- `backend/internal/agent/monitor.go` — add classified/gate/assessment emits
- `backend/internal/agent/classifier_lumber.go` (or wherever Lumber returns) — emit classified
- `backend/internal/agent/severity_gate.go` — emit gate
- `backend/internal/api/server.go` — register routes
- `backend/cmd/heimdall/main.go` — construct `PipelineBus`, inject into handlers + agent

**New frontend files:**
- `frontend/src/types/pipeline.ts`
- `frontend/src/api/pipeline.ts`
- `frontend/src/composables/usePipelineStream.ts`
- `frontend/src/stores/pipeline.ts`
- `frontend/src/pages/PipelinePage.vue`
- `frontend/src/components/pipeline/PipelineFunnel.vue`
- `frontend/src/components/pipeline/PipelineParticles.vue`
- `frontend/src/components/pipeline/PipelineNodeCard.vue`
- `frontend/src/components/pipeline/PipelineNodeDetail.vue`
- `frontend/src/components/pipeline/PipelineLogTicker.vue`
- `frontend/src/components/pipeline/TimeMachineBlock.vue`

**Modified frontend files:**
- `frontend/src/router/index.ts` — register `/pipeline`
- Top header component — add nav item

---

## Complexity estimate

- Phase 1 (backend persistence + pub/sub + endpoints): ~600 LOC
- Phase 2 (frontend live UI + granular panels + stubs): ~1400 LOC
- Phase 3 (hardening): ~150 LOC
- Phase 4 (replay, future): ~400 LOC

Phase 2 is larger than the original `pipeline-page.md` estimate (800 LOC) because of (a) granular auditing detail panels per stage, (b) click-to-expand behaviour everywhere, (c) stubbed Time Machine block that must be visually complete.

---

## Rollout order

1. **Phase 1** merges behind no flag — backend persistence and endpoints are harmless without the UI consuming them. Verifies `log_pipeline_events` is being populated in production before any user sees the page.
2. **Phase 2** merges once Phase 1 has been writing events for at least 24h in production (so the initial Pipeline page render has real data).
3. **Phase 3** folds into Phase 2's final PR or lands as a follow-up.
4. **Phase 4** is scheduled separately — gated on real user feedback on v1.

---

## Resolved decisions

1. **Escalation rule naming for `rule_hit`** — **resolved as Option A**. `ShouldEscalate` returns `(bool, ruleID string)` with stable IDs per branch (`error_type`, `request_server_error`, `system_resource_alert`, …). Full rule-ID table in `pipeline-rule-options.md`. Non-escalated events persist with `rule_hit = ""`. Alternatives (DB-backed rules, expression engine) are documented but not pursued.

2. **Ingestion event granularity** — **one event per log**. 48h retention bounds the row count (ballpark: 1000 logs/min × 60 × 48 × 4 stages ≈ 11.5M rows, within comfort for indexed Postgres). Aggregate events would lose the per-log particle stream that is the feature's point.

3. **Activity stage completeness** — **remain at 4 event stages**. Activity is a downstream *view* over assessments, not a new event. The Activity funnel node visually exists but is painted from `assessment`-stage events plus a link out to the existing Activity page. A fifth `notification_dispatch` event may be added later if notifications become first-class in the pipeline UI — the `stage` column is a TEXT and trivially extensible.

4. **Time Machine retention tiers** — **deferred**. 48h is the v1 default, inherited from `log_buffer`. Longer retention (for paid tiers or compliance) becomes a separate scoped plan; it is cheap to introduce later by decoupling `log_pipeline_events` retention from `log_buffer` via a soft reference + independent pruner.

5. **Ticker hydration** — **bundled `/bootstrap` endpoint + SSE `?since=` replay + client ID-based dedupe**. Architecture in Phase 1.6 and Phase 2.3. Guarantees: (a) first paint has populated stats + ticker in one roundtrip, (b) zero gap between hydration and live stream, (c) zero duplicates, (d) correct under reconnects and long disconnects (via `resync`).

## Remaining open questions

1. **`PipelineBus` channel capacity at burst.** The default buffered-channel cap (128) may be low for assessment-stage fan-out when Claude completes a large batch. If production telemetry (`pipeline_bus_dropped_events_total`, Phase 3.1) shows drops during normal bursts, raise the cap or introduce a per-stage channel. Track during Phase 3 rollout; not a Phase 1 blocker.

2. **Bootstrap cache key granularity.** Phase 3.3 adds a 5s server-side cache on `/bootstrap`. Cache key must include `(appId, window, tickerLimit)` — a user with two tabs open at different windows (e.g. 1h vs 24h) must not cross-pollute. Noting for Phase 3 implementation.
