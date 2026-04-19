# Pipeline Page — Phase 2 Completion

**Scope:** User-visible Pipeline page at `/pipeline` — live SVG Sankey funnel + Canvas2D particle layer + per-stage detail panels + live log ticker + per-log journey modal + a stubbed-but-visually-complete Time Machine block. Hydrates via `/pipeline/bootstrap` and streams via `/pipeline/stream` SSE with `?since=` replay and ID-based dedupe. Backend is unchanged from Phase 1; this release is pure frontend.
**Plan:** `docs/executing/pipeline-page-plan.md` §Phase 2

---

## The Problem This Solves

Phase 1 wired the backend so every log's pipeline journey is being recorded and live-published — but nothing in the UI consumed it. An operator opening Heimdall could not see ingestion velocity, the gate's flagged share, or which Lumber rules were firing in real time. Debugging a misconfiguration still meant `tail`-ing server logs.

Phase 2 lights up the page. From the moment the user lands on `/pipeline`, they see:

- A funnel diagram with five stages (Ingestion → Lumber → Gate → Agent → Activity) whose stream widths breathe with the actual flagged/safe ratio over the last hour.
- Particles riding the funnel paths — green pre-Gate, amber post-Gate-flagged, muted-grey on the safe drain.
- Node cards above each stage showing the rolling count and per-second rate; clicking a card expands a per-stage auditing panel (top sources, classification histogram, rules-fired list, recent assessments).
- A live log ticker at the bottom — every event shown with its stage badge and a click-to-expand journey modal.
- A Time Machine block at the top that's visually complete but disabled, with a "Coming soon — every log's pipeline journey is already being recorded" tooltip.

Phase 2's bar was *user-visible page that hydrates synchronously and stays in sync indefinitely*. Both met.

---

## What Changed

### 1. Types and API client

**Files:** `frontend/src/types/pipeline.ts`, `frontend/src/api/pipeline.ts`

`PipelineEvent`, `PipelineStats`, `PipelineBootstrap`, `PipelineJourney`, and a `ConnectionState` string-literal union mirror the backend's `pipelineEventJSON` / `bootstrapResponse` shapes from `backend/internal/api/handlers/pipeline.go`. The wire shape uses `omitempty` on stage-specific fields, so the TS types use plain optional `string`/`number`/`boolean` — no null guards needed at every read site.

`fetchPipelineBootstrap(appId, { window?, tickerLimit? })` and `fetchLogJourney(appId, logId)` are thin axios wrappers that share the project's standard `client` instance (auth header, org header, 401 → logout interceptor all inherited automatically).

### 2. Pinia store — `usePipelineStore`

**File:** `frontend/src/stores/pipeline.ts`

Composition store, matching the codebase convention. Holds:

- `stats` — replaced wholesale on bootstrap. Tiny object, deep ref is fine.
- `events` — held in a **`shallowRef`** so Vue doesn't deep-watch hundreds of objects. `triggerRef` after every mutation. Cap at 500; ring buffer evicts oldest-first; mirrored `Set<string>` for **O(1) dedupe by event id**.
- `stageTimestamps` + 1s recompute tick → `rates` (events/sec, 60s window). The recompute is **a 1s `setInterval`** that's started/stopped by the composable's lifecycle hooks; consuming components never see the timer.
- `flaggedRatio` / `safeRatio` / `assessmentRatio` — `computed` derivations off `stats` for the funnel widths. **Defaults to 0.5/0.5** when no gate decisions exist, so the funnel still has visible geometry on first paint instead of collapsing to a hair-thin line.
- `journeyCache` — Map-backed LRU (cap 100). Re-clicking the same row in the ticker doesn't re-fetch `/journey`. Touches re-insert into the Map for natural insertion-order LRU.

`reset()` clears everything in one call — used when the user switches apps so we don't leak the previous app's events into the new app's view.

### 3. Composable — `usePipelineStream`

**File:** `frontend/src/composables/usePipelineStream.ts`

Hydration-to-stream glue. Exposes `start()`, `destroy()`, `reconnect()`, plus reactive `connectionState` and `lastError`. Sequence:

1. **`bootstrap()`** — fetch `/pipeline/bootstrap`, push stats + events into the store synchronously, capture `cursor`.
2. **`openStream()`** — open `EventSource` against `/pipeline/stream?since=<cursor>&token=<jwt>`.
3. Three event types handled separately:
   - `event: replay` — frames missed between bootstrap-read and SSE-subscribe. Inserted via the same `store.insertEvent` path — the dedupe set rejects the overlap with bootstrap, which is exactly what makes the architecture safe.
   - `event: live` — steady-state stream. First live frame flips `connectionState` to `'streaming'` (not `onopen`, because data flowing is the real signal).
   - `event: resync` — server hit its replay cap, the buffer is too stale to trust. We `store.reset()`, drop cursor, re-bootstrap, reconnect.
4. **Token-aware reconnect.** EventSource auto-reconnects, but its built-in retry doesn't refresh the URL — meaning a token rotation mid-stream would loop on 401. We tear down on `onerror` and reconnect manually with exponential backoff (1s → 30s cap) so a fresh token is picked up from the auth store on the next attempt.
5. **Cursor advances on every accepted event** — reconnects don't replay anything we've already seen. Server emits replay frames in ASC order and live frames at occurrence time, so a max-of-seen `occurred_at` is monotonic and sufficient.

`onUnmounted` calls `destroy()` automatically; the page also calls it explicitly on app switch before instantiating a fresh handle.

### 4. Page shell — `PipelinePage.vue`

**File:** `frontend/src/pages/PipelinePage.vue`

Route-level component. Mounts the composable, watches `app.currentAppId`, and tears down + restarts on app switch. Renders:

- Header with a state pill (`streaming` / `reconnecting` / `resyncing` / `error` / `idle` / `bootstrapping`) — colour-coded against the design tokens, breathing-glow on `streaming`.
- `<TimeMachineBlock />` (disabled stub).
- Funnel + particles + node cards in a single bordered card.
- Per-stage detail panel (`<PipelineNodeDetail />`) toggles open beneath the funnel when a card is clicked.
- `<PipelineLogTicker />` underneath.
- `<PipelineJourneyModal />` mounted lazily when `inspectingLogId` is set.

Two empty states: **no apps** ("Create an application to start watching its pipeline") and **streaming-but-zero-ingestion** ("No ingestion events in the last hour. Enable a source on a connection to start the flow."). The second matches Phase 2.6 of the plan — drop-by-default source filtering means a freshly onboarded user will see zero events forever until they enable sources, and we surface the right next step rather than spinning indefinitely.

### 5. Funnel — `PipelineFunnel.vue`

**File:** `frontend/src/components/pipeline/PipelineFunnel.vue`

Pure-SVG Sankey. **Five stages laid out as fractional x-positions**, projected into pixel space at resize time via a single `ResizeObserver`. The funnel itself is two SVG paths:

- **Pre-Gate trunk** — rectangle from Ingestion to Gate, gradient-filled phosphor-green.
- **Post-Gate split** — two cubic-bezier tapers. Top stream (flagged) rides amber and narrows into Activity; bottom stream (safe) drains away into a thin grey wisp.

Stream widths come from `store.flaggedRatio` and `store.safeRatio`, with a `Math.max(3px, …)` floor so a zero-flagged window still has visible body. Geometry (width, height, cy, trunkHalf, flaggedHalf, safeHalf, stagePoints) is **`defineExpose`-d** so the particle layer and node card row above can plant on the same anchor x's without re-implementing the funnel's pad math.

Skipped d3-sankey: five stages is too few to justify a layout-pass dependency, and the bespoke geometry gives us pixel-precise control over the post-Gate split shape that d3's standard sankey can't easily produce.

### 6. Particles — `PipelineParticles.vue`

**File:** `frontend/src/components/pipeline/PipelineParticles.vue`

Canvas2D layer absolutely positioned over the SVG. **Allocation-free per-frame** via a fixed-size pool of 600 `Particle` slots that recycle as old particles fade. New events arriving in the store trigger spawn calls; each particle gets:

- A travel segment (between two stage anchor x's) keyed by its event's stage.
- A lane y — pre-Gate ride the centre with slight jitter; post-Gate flagged ride above centre, safe ride below — matches the SVG split visually.
- An ease-out-cubic interpolation across `TRAVEL_MS = 1800`.
- Fade in over the first 15% of life, fade out over the last 15%.
- A 2.4px solid disc + a 5.5px low-alpha glow halo. Two `arc` calls per particle is the entire per-frame cost.

**Sampling under load.** `recentSpawns` is a rolling counter that decays per second; when it exceeds `SAMPLE_THRESHOLD_PER_S = 50`, only 1 in 4 events spawn a particle. The store still records every event; only the visual drops. Auto-relaxes when traffic calms.

**High-DPI.** Canvas backed at `devicePixelRatio` so particles stay crisp on Retina; CSS size stays = layout px so the canvas and SVG align.

### 7. Node card + detail panel

**Files:** `frontend/src/components/pipeline/PipelineNodeCard.vue`, `frontend/src/components/pipeline/PipelineNodeDetail.vue`

Cards are tone-aware (`default` / `warn` / `safe`) — Gate uses `warn` so it visually pops as the decision point. A live "active" pulse appears on the right when the per-stage rate > 0.

The detail panel renders different content per stage, all sourced from the same in-memory ring buffer (no extra round-trip):

- **Ingestion** — top sources by count + last 20 ingest rows.
- **Lumber** — Type.Category histogram + last 20 classifications with confidence and summary.
- **Gate** — escalation rules fired with hit counts + flagged-share % + last 20 gate decisions.
- **Agent** — recent assessments with severity + summary, deep-linkable to the assessment.
- **Activity** — severity breakdown + a `Open Activity →` link to the existing Activity page.

Every row in every panel is click-to-inspect → fires the journey modal.

### 8. Live ticker — `PipelineLogTicker.vue`

**File:** `frontend/src/components/pipeline/PipelineLogTicker.vue`

Bottom card, max-height 280px, scrolls. Renders the most recent 80 events from the 500-cap ring buffer — **DOM size bounded** even when the buffer is full. Each row is a one-line journey hint (timestamp · stage badge · stage-appropriate content), click-to-inspect.

Stage badges use the design-token semantics: ingestion = muted, classified = accent-bright, gate = warn, assessment = info.

### 9. Journey modal — `PipelineJourneyModal.vue`

**File:** `frontend/src/components/pipeline/PipelineJourneyModal.vue`

Phase 1's `/pipeline/logs/{logId}/journey` endpoint already returns the data; this is the consumer. Teleported to body so it doesn't inherit the page's overflow constraints. Lists each stage event with timestamp + stage label + structured detail (source, severity, classification, gate decision, assessment id). Trailing note: "Replay view coming soon — Phase 4 will animate this log through the funnel."

Per-log journeys are LRU-cached in the store (cap 100), so re-clicking the same row in the ticker is a free open.

### 10. Time Machine stub — `TimeMachineBlock.vue`

**File:** `frontend/src/components/pipeline/TimeMachineBlock.vue`

Visually complete: title, "soon" pill, two `<input type="datetime-local">` controls, a "Replay" button. All disabled with `cursor-not-allowed` styling and a tooltip pointing at "every log's journey is already being recorded." Phase 4 will swap the disabled attribute for behaviour and the picker will already look right.

### 11. Routing + nav

**Files:** `frontend/src/router/index.ts`, `frontend/src/components/common/AppSidebar.vue`

New route `/pipeline` → lazy-imported `PipelinePage`. Nav entry added to the **Infrastructure** section in `AppSidebar.vue`, immediately after Connections. The plan's wording mentioned the "top header" but the codebase keeps primary nav in the sidebar — the entry placement (between Connections and Agent Configuration in the rendered nav order) honours the plan's intent.

---

## Tests

All tests pass. 13 new tests across three files; pre-existing 60 still green.

### Unit — `stores/__tests__/pipeline.test.ts`

7 subtests covering the reactivity contract:

- **insertEvent dedupes by id** — two inserts of the same id produces one row.
- **insertEvent prepends newest** — verifies the ticker render order.
- **mergeEvents preserves newest-first batch order** — bootstrap delivers newest-first; the iterate-in-reverse + prepend trick keeps the order right.
- **flaggedRatio defaults to 0.5/0.5** — funnel stays visible on a fresh load.
- **flaggedRatio reflects stats counts** — 30/70 split → 0.3 / 0.7.
- **reset clears events, stats, and dedupe set** — re-inserting the same id after reset must succeed (catches a hidden Set leak).
- **ring buffer evicts oldest beyond cap** — 510 inserts → 500 stored, oldest 10 dropped, newest at index 0.

### Composable — `composables/__tests__/usePipelineStream.test.ts`

2 subtests with a stub `EventSource`:

- **bootstraps stats and events into the store** — mocks `client.get`, asserts the store hydrated, asserts the SSE URL carries the cursor as `?since=` (URL-encoded), asserts the EventSource is `.close()`d on unmount.
- **flips to streaming on first live frame and dedupes against bootstrap** — replays the bootstrap event as a `replay` frame, asserts the store rejects the duplicate, then sends a fresh `live` frame, asserts the buffer length is 2 with newest-first order and `connectionState === 'streaming'`.

### Component — `components/pipeline/__tests__/PipelineFunnel.test.ts`

3 subtests:

- **renders the five stage labels** — Ingestion, Lumber, Gate, Agent, Activity. (CSS `text-transform: uppercase` is a render-time effect; happy-dom returns source casing.)
- **renders three stream paths plus the inflow beam** — at least 4 `<path>` elements.
- **post-gate widths are derived from store ratios** — heavy-flagged stats produce flaggedHalf > safeHalf via the exposed geometry refs.

---

## Decisions Worth Remembering

- **Funnel stream widths come from a Pinia `computed`, not from the funnel's local state.** The page's reactive flow is: events arrive → store ratios recompute → funnel re-renders. One source of truth.
- **`shallowRef` for the events buffer.** A 500-element array of plain objects deep-watched would melt Vue under high-throughput streams. The shallow ref + manual `triggerRef` trades ergonomics for predictable cost.
- **First live frame, not `onopen`, flips to `streaming`.** Handshake is not data flow. Users assume "streaming" means data is arriving.
- **Token-aware manual reconnect, not EventSource auto-retry.** Built-in retry reuses the original URL — meaning a rotated JWT loops on 401 forever. Manual reconnect closes the socket and rebuilds the URL with whatever's in the auth store right now.
- **Cursor advances per accepted event, not per frame batch.** Reconnects never re-replay anything we've already inserted. A monotonic max-of-seen `occurred_at` is enough because the server emits in order.
- **Particle pool, not allocation per spawn.** 600 fixed slots, recycled as old particles fade. Sampling kicks in past 50 events/sec. Two `arc` calls per particle per frame is the whole cost — we have headroom for many multiples of expected throughput before WebGL becomes necessary.
- **Per-log journey cache LRU is in the store, not the modal.** Re-opening the same row in the ticker → cached. Re-opening across modal mount/unmount → still cached. The cache outlives the component lifecycle, which is the point.
- **App switch tears down the stream and resets the store.** No accidental leakage of one app's events into another's view; no orphaned EventSource connections accumulating across navigations.

---

## Follow-ups

1. **Phase 1b — wire poller-based connectors** (still open from Phase 1). Until then, apps using fly/syslog/vercel/railway/supabase/mongodb connectors will see a silent Ingestion node. Classified/Gate/Assessment still fire correctly.
2. **Phase 3 hardening.** Drop-counter metric (3.1), SSE per-user connection cap (3.2), `/bootstrap` 5s in-memory cache (3.3), abandoned-subscriber leak canary (3.4), RLS verification pass (3.5). All are isolated backend changes that don't require frontend rework.
3. **Phase 4 — Time Machine replay.** Backend is already populated; frontend stub is visually complete. Wiring the date/time picker → `/pipeline/logs?since=&until=` (note: that endpoint isn't yet implemented; the plan calls for it in Phase 1.6 but Phase 1 deferred it as the page didn't yet need it) + a per-log replay modal that animates a single log through the funnel.
4. **Bootstrap window picker.** Hardcoded to `1h` today. The backend supports `?window=` up to 24h with a 1h default. Adding a window selector (1h / 6h / 24h) in the page header is a one-day add — it would just need to invalidate `events` on change since the ring buffer is window-agnostic.
5. **Particle cap is hard-coded at 600 / sample threshold at 50/sec.** Both should become tunable via env-driven config once we have real telemetry on per-app throughput. Document tuning advice once Phase 3.1's `pipeline_bus_dropped_events_total` produces real data.

---

## Files Touched

**New frontend files:**

- `frontend/src/types/pipeline.ts`
- `frontend/src/api/pipeline.ts`
- `frontend/src/stores/pipeline.ts` + `__tests__/pipeline.test.ts`
- `frontend/src/composables/usePipelineStream.ts` + `__tests__/usePipelineStream.test.ts`
- `frontend/src/pages/PipelinePage.vue`
- `frontend/src/components/pipeline/PipelineFunnel.vue` + `__tests__/PipelineFunnel.test.ts`
- `frontend/src/components/pipeline/PipelineParticles.vue`
- `frontend/src/components/pipeline/PipelineNodeCard.vue`
- `frontend/src/components/pipeline/PipelineNodeDetail.vue`
- `frontend/src/components/pipeline/PipelineLogTicker.vue`
- `frontend/src/components/pipeline/PipelineJourneyModal.vue`
- `frontend/src/components/pipeline/TimeMachineBlock.vue`

**Modified frontend files:**

- `frontend/src/router/index.ts` — register `/pipeline`
- `frontend/src/components/common/AppSidebar.vue` — add Pipeline nav entry under Infrastructure

**Backend:** unchanged.

Frontend totals: ~1,180 LOC across new files + 5 LOC across modified files. Tests: ~190 LOC across three files. Bundle impact: PipelinePage chunk is 29.4 kB (9.2 kB gzip) — well within the budget for a major new surface. `vue-tsc` clean, vite build clean, all 73 frontend tests pass (13 new + 60 pre-existing).
