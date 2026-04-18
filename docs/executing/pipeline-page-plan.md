# Pipeline Page — Implementation Plan

Visualise the five-stage log pipeline (Ingestion → Lumber → Gate → Agent → Activity) as an app-scoped Sankey/funnel diagram with a live SSE feed and a persisted per-log event trail for future replay ("Time Machine").

This plan follows and replaces the proposal in `pipeline-page.md`. Decisions made:

- **Proposal B (Sankey/funnel)** chosen over A and C.
- **App-scoped**, matching Connections/Agent Config/Activity.
- **Mobile deferred** — platform-wide overhaul will address it later.
- **Time Machine stubbed on the frontend** from day one, backend fully populated so no backfill gap exists when the UI lands.
- **Naming**: "Pipeline", top-level nav item between Connections and Agent Config.
- **Rendering**: SVG + Canvas2D for v1. WebGL/WebGPU deferred — see "Rendering tech trade-offs" below.

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

## Phase 1 — Backend: Persistence + Pub/Sub

**Goal:** Every log flowing through the pipeline publishes live events AND persists its full stage history. Replay-capable from day one.

### 1.1 Migration `034_log_pipeline_events`

**New file:** `backend/migrations/034_log_pipeline_events.up.sql`

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
- `GetRecentPipelineEventsByApp` — paginated recent events for an app (used by Time Machine list view)
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

- `GET /api/apps/{appId}/pipeline/stats?since=1h` — snapshot aggregates for initial render. Calls `PipelineStatsByApp`.
- `GET /api/apps/{appId}/pipeline/stream` — SSE. Subscribes to `PipelineBus`, writes `data: {...}\n\n` frames. Closes on client disconnect or context cancel. Auth: JWT via existing middleware. Also emits a periodic `event: heartbeat` every 15s so proxies don't idle-close.
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

- `api/pipeline.ts` — `fetchPipelineStats(appId, since)`, `fetchLogJourney(appId, logId)`, `fetchPipelineLogs(appId, params)`.
- `composables/usePipelineStream.ts` — wraps `EventSource` for `/api/apps/{appId}/pipeline/stream`. Exposes: reactive `events` ring buffer (cap 500), `stats` reactive snapshot, `connectionState`, `reconnect()`. Auto-reconnects with backoff on disconnect. Includes the JWT as a query-string token (SSE can't set headers), matching the existing WebSocket pattern.

### 2.4 Store

**New file:** `frontend/src/stores/pipeline.ts`

Pinia composition store holding:
- Current app's stats snapshot.
- Recent events ring buffer (for the live ticker).
- Per-stage rolling rates (events-per-minute, recomputed on a 1s interval).
- Derived `flaggedRatio` / `safeRatio` for Sankey width calculation.

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
- `backend/migrations/034_log_pipeline_events.up.sql` + `.down.sql`
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

## Remaining open questions

1. **Escalation rule naming for `rule_hit`.** Phase 1 persists which Gate rule fired. The rules are currently hardcoded in `severity_gate.go` — do they have stable string identifiers today, or do we need to introduce a rule-ID constant table as part of this work? (Ninety-second grep through `severity_gate.go` will answer this; flagging so we don't discover it mid-implementation.)

2. **Ingestion event granularity under batch webhooks.** A single webhook request can contain hundreds of logs (e.g., Fly drain bursts). Do we emit one `ingestion` event per log (truest to the data model, but 100× the pipeline event writes) or one aggregated event per batch with a `count` field and spawn the per-log events only at the `classified` stage? Recommendation: one event per log — writes are cheap, and per-log ingestion events are the only way to draw an accurate Ingestion-stage particle stream. But worth a conscious decision.

3. **Activity stage completeness.** The doc lists five stages but `log_pipeline_events.stage` only has four values (ingestion/classified/gate/assessment). The "Activity" visual node is really just the downstream view of assessments — no new event is emitted when something lands in `agent_log`; it's the same event as `assessment`. Confirming that's the right model, or do we want a fifth `activity` stage event for symmetry (e.g., to capture notification dispatch)?

4. **Time Machine retention messaging.** The 48h window is currently hardcoded in SQL. Is that still the right default for replay, or does Heimdall want a longer retention tier for paid tiers specifically for pipeline audit history? If yes, it changes the schema slightly (separate retention policy for `log_pipeline_events`) — cheap to add now, expensive to retrofit.

5. **Live ticker storage.** Phase 2 uses an in-memory ring buffer (cap 500) for the live ticker. On page reload, the ticker starts empty until new events arrive. Acceptable, or should the initial page load hydrate the ticker with e.g. the last 50 events from `log_pipeline_events`? (Recommendation: hydrate from DB on load — it's a one-query addition and avoids an empty-state feel.)
