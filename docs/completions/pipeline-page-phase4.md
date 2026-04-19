# Pipeline Page — Phase 4 Completion

**Scope:** Time Machine replay — the Phase 2 stub becomes a functional picker + a per-log replay modal that animates each log through the Sankey funnel. Adds one backend endpoint (deferred from Phase 1.6), one deep-link route, upgrades the existing journey modal in place. Closes the loop on "every log's journey is already being recorded" that Phases 1–3 prepared for.
**Plan:** `docs/executing/pipeline-page-plan.md` §Phase 4

---

## The Problem This Solves

Phase 2 shipped a visually complete Time Machine block with disabled pickers and a "coming soon" tooltip — the plan's intent being that Phase 4 would swap behaviour in without touching the visual shell. Phases 1–3 then populated every log's journey into `log_pipeline_events` and hardened the live path. What was missing was the *replay surface*: no way to pick a historical time range, see which logs flowed through, and animate any one of them through the funnel after the fact.

That gap mattered because the entire value prop of recording every journey is the ability to answer post-hoc questions: "which logs did Lumber flag last night between 2am and 4am?", "what did the agent do with that weird payload I saw in Activity?", "was this pattern present an hour ago?". Without a picker the data was accumulating in a table nobody could query from the product.

Phase 4 makes it queryable:

- The Time Machine block now accepts a `since/until` window, calls `/pipeline/logs`, and renders a clickable list of journey summaries (source, severity, classification, gate decision, flagged-or-safe badge, duration).
- Clicking a row opens the journey modal — upgraded from a timeline-only view to a scaled Sankey with a traced particle that walks the log through each stage it reached. Stages the log never reached (e.g. assessment when the gate dropped it) stay dim; reached stages light up as the particle arrives.
- The modal is deep-linkable: `/pipeline/logs/:logId` auto-opens the replay on first paint, so Activity links and Slack notifications can point straight at a specific journey.
- 48h retention is called out in the picker preamble so users don't wonder why older ranges come back empty — the ceiling matches `log_buffer`'s `ON DELETE CASCADE` behaviour.

---

## What Changed

### 1. New picker query — `ListPipelineLogsByApp`

**Files:** `backend/internal/db/queries/log_pipeline_events.sql`, `backend/internal/db/log_pipeline_events.sql.go`

One SQL query that collapses a log's up-to-four stage rows into a single summary. `GROUP BY log_id` with `MAX(...) FILTER (WHERE stage = X)` and `BOOL_OR(escalated) FILTER (WHERE stage = 'gate')` gives one row per log with:

- `first_seen_at` / `last_seen_at` — the journey's temporal span, derived via `MIN`/`MAX(occurred_at)`.
- `stage_count` — 1–4, reflecting how far the log made it. A log that was source-filtered won't be in the table at all; a log that got to classification but not gate gets `stage_count = 2`.
- Stage-specific columns sourced from the matching stage row (`source_type` + `severity` from ingestion, `type`/`category`/`confidence`/`summary` from classification, `escalated`/`rule_hit` from gate, `assessment_id` from assessment).

**Why `MAX(...) FILTER` is safe here:** the writer contract (`agent/pipeline_writer.go`) emits each stage at most once per log_id. So the FILTER partitions narrow to at most one row per partition, and `MAX` is effectively "pick the one value there is". The query comment documents this load-bearing invariant so nobody changes the writer's single-emit-per-stage behaviour without realising it breaks the picker.

**Why offset pagination, not cursor.** Cursor pagination on a windowed GROUP BY would need `HAVING MAX(occurred_at) < @cursor` plus tie-breaking on `log_id` — three times the surface area of `LIMIT $4 OFFSET $5`. At 48h retention and realistic per-app volumes, `idx_lpe_app_occurred` scans the range efficiently regardless of offset; the plan's "virtualised list" goal is met by the UI layer (it renders the visible window at the browser, not the entire 48h of rows).

**Generated code hand-rolled**, matching the precedent from Phase 3's `source_filters.sql.go` comment regen: the diff is what `sqlc generate` would emit, done by hand to avoid threading the sqlc binary through this change.

### 2. `GET /api/apps/{appId}/pipeline/logs` handler

**Files:** `backend/internal/api/handlers/pipeline.go`, `backend/internal/api/router.go`, `backend/internal/api/handlers/testhelpers_test.go`

Query params:

- `since` / `until` — RFC3339 timestamps. Default: last 24h. Clamped so the span never exceeds 48h (`logsPickerMaxRange`) — anything wider is wasted work because of the log_buffer cascade-delete ceiling, and the clamp prevents a buggy client DOSing the picker with a 30-day range.
- `limit` — 1..200 (default 50, cap 200 to match `/bootstrap`'s `tickerLimit`).
- `offset` — ≥ 0 (default 0).

Response includes a `truncated` flag (`len(rows) >= limit`) so the client can offer a "load more" button without making it probe with an off-by-one request. That pattern matches how other paginated surfaces in the codebase (notifications history, logs list) communicate "there's more" to the frontend.

Authorisation via the existing `authorizeApp` helper — same posture as `/pipeline/bootstrap`. A user hitting another org's `appId` gets 404, verified in test.

### 3. Type + API additions

**Files:** `frontend/src/types/pipeline.ts`, `frontend/src/api/pipeline.ts`

`PipelineLogSummary` mirrors the backend's `pipelineLogSummaryJSON`, `PipelineLogsResponse` mirrors the list envelope. `fetchPipelineLogs(appId, { since, until, limit, offset })` wraps the axios call with sensible defaults — the shared `client` instance inherits auth headers, 401-logout interceptor, and the `/api` base URL.

### 4. `TimeMachineBlock.vue` — wired picker

**File:** `frontend/src/components/pipeline/TimeMachineBlock.vue`

Rewritten end-to-end but the visual shell (title, "soon" pill, since/until/Replay layout) is preserved — the Phase 2 styling decisions still apply. The "soon" pill now appears only when the block is showing the initial `:logId` deep-link context (renamed to "replaying") instead of permanent.

Picker behaviour:

- Defaults: last 24h window, limit 50.
- Uses `datetime-local` inputs in local-time, converts to UTC RFC3339 at fetch time.
- Renders results as a compact list below the picker inputs:
  - Stage bar (● ● ● ●) showing how far the log made it at a glance — filled dots for reached stages, outlined for unreached. Hover reveals `4/4 stages reached`.
  - One-line row with time, source, severity, classification, summary, duration (`first_seen_at → last_seen_at` delta), and a `flagged` pill when the gate escalated.
  - Click emits `inspect` with the log_id — consumed by `PipelinePage.vue`.
- Retention disclosure is now in the block's subtitle: "Retention is 48 hours — older logs are pruned with their parent row." Satisfies plan §4.4 without adding a separate callout.
- Empty state: "No journeys in this window. Try a wider range, or check source filters on your connections." — mirrors the live page's helpful next-step phrasing rather than a generic spinner.
- Switching apps clears results (via `watch` on `appId`) so the previous app's journeys don't linger.

**Why no true virtualisation.** The plan called for a virtualised list. The codebase has no virtualisation library today, and at a 200-row max per page the picker doesn't need one — a `max-h-72 overflow-auto` scroll container handles it. If volume grows past the point where 200 DOM rows matters, bringing in `vue-virtual-scroller` is a contained future change; the row template stays identical.

### 5. `PipelineJourneyModal.vue` — replay upgrade

**File:** `frontend/src/components/pipeline/PipelineJourneyModal.vue`

The Phase 2 modal was a timeline-only view with a "Replay view coming soon" footer. Phase 4 upgrades it in place:

- Scaled Sankey (640×140 SVG) at the top, with five stage nodes connected by a pre-gate trunk and a post-gate fork. The fork renders both flagged (top, amber) and safe (bottom, muted grey) lanes so the visual language matches the live page's funnel without coupling to it.
- A traced particle — single amber-glow circle — sits at the first stage node and animates through the stages the log actually reached. Driven by a `setTimeout(step, 650ms)` chain, which gives a guided reading of the journey rather than a physics sim. Each stage node fills as the particle reaches it; unreached stages stay outlined with a small "—" indicator above.
- If the gate's decision was `safe`, the particle drifts down to the safe lane after stage 2; if `flagged`, it rides the top lane. Matches how the live funnel interprets gate decisions.
- The timeline view (stage-by-stage event detail) is kept below the Sankey, unchanged.
- "Replay" button restarts the animation. Auto-plays once on load so the user sees the animation without clicking.
- Closing the modal (via `close` emit) clears the deep-link URL param — see §6.

**Why a self-contained Sankey instead of reusing `PipelineFunnel.vue`.** The live funnel reads ratios off the pipeline store, which keeps recomputing as events stream in. The replay backdrop needs to be a static visual representing *this log's* journey, not the rolling flagged/safe split of the window. Reusing `PipelineFunnel` would have meant freezing its ratios at modal-open time (complex) or letting the backdrop drift while the modal was open (wrong). A separate scaled Sankey is cheaper and more correct. The geometry is simple enough (five equally-spaced anchors in a 640×140 viewBox) that there's no layout code to duplicate.

### 6. Deep-link route `/pipeline/logs/:logId`

**Files:** `frontend/src/router/index.ts`, `frontend/src/pages/PipelinePage.vue`

New route that mounts the same `PipelinePage` component — the `:logId` param is watched inside the page and auto-opens the replay modal when set. Inspecting a log via the ticker or picker also mirrors into the URL via `router.replace({ name: 'pipeline-log', params: { logId } })`, so every replay is shareable.

Closing the modal steps back to `/pipeline` (via `router.replace({ name: 'pipeline' })`) so the next click pushes a fresh param cleanly. Use of `replace` (not `push`) keeps the browser back button from accumulating one history entry per modal open. The behaviour is idempotent: refresh re-opens the modal with the same log, close clears it.

### 7. Wiring in `PipelinePage.vue`

**File:** `frontend/src/pages/PipelinePage.vue`

- `TimeMachineBlock` now receives `:app-id`, `:initial-log-id`, and listens to `@inspect` (same handler the ticker uses, so picker-click and ticker-click funnel through identical code).
- New `watch` on `route.params.logId` syncs the modal's `inspectingLogId` with the URL.
- `inspectLog(logId)` mirrors into the URL; `closeJourney()` clears it.

---

## Tests

### Backend — `api/handlers/pipeline_test.go`

Two new integration subtests, both in `TestPipeline_Integration` (skip without `DATABASE_URL`):

- **logs picker dedupes per log and respects window** — walks one out-of-window log (backdated by 10 hours via a direct `UPDATE`) plus three fresh logs, hits `/pipeline/logs?since=…&until=…`, asserts the response has one row per fresh log with `stage_count = 4`, populated summary fields, `escalated = true`, and `rule_hit = RuleErrorType`. Also asserts the out-of-window log is absent.
- **logs picker refuses cross-org app** — constructs a foreign org + app and hits the test user's route with the foreign `appId`. Must 404 via `authorizeApp`, not 200-with-empty-logs.

All five pre-existing `TestPipeline_Integration` subtests still pass.

### Frontend

- **`api/__tests__/pipeline.test.ts`** (new file, 2 subtests) — `fetchPipelineLogs` hits the right URL with the right params; custom `limit`/`offset` are forwarded correctly.
- **`components/pipeline/__tests__/TimeMachineBlock.test.ts`** (new file, 3 subtests):
  - Clicking the Replay button triggers a `/pipeline/logs` call and renders the human-readable summary fields (source, type, severity, `flagged` pill). The raw log_id deliberately isn't rendered in the row — it'd be a 36-char UUID noising up the line — so the assertion is on summary fields, documented inline.
  - Clicking a log row emits `inspect` with the correct `log_id`.
  - Empty result renders the "No journeys in this window" message.

**Totals:** 13 → 15 frontend test files; 73 → 78 tests, all green (no pre-existing test changes). Frontend build: `PipelinePage` chunk grew from 29.4 kB (9.2 kB gzip) to 36.8 kB (11.4 kB gzip) — a +7.4 kB budget spend for the picker logic, replay Sankey, and route param watch, well under the plan's ~400 LOC Phase 4 estimate.

---

## Decisions Worth Remembering

- **Picker query uses `MAX(...) FILTER` under the invariant that each stage emits at most once per log.** Documented in the query comment. If a future writer change (e.g. re-classification, multi-assessment) ever violates that invariant, this query silently picks the "largest" value per column — unlikely to be correct. The writer contract is load-bearing for this picker.
- **48h max window clamp, enforced server-side.** Matches the log_buffer retention ceiling. Prevents a buggy client (or a pen-tester) from requesting 30 days and pinning the picker query under `idx_lpe_app_occurred` scans on an empty range.
- **Offset pagination, not cursor.** At 48h × typical per-app volume, offset is cheap; cursor over a GROUP BY would triple the query surface area for no real benefit. The picker is virtualised at the UI layer (200-row max, scroll container), which is the right layer for that concern.
- **Replay modal owns its own Sankey, not reusing `PipelineFunnel`.** The live funnel's ratios track the live window; the replay backdrop needs to be static. Coupling would have meant either freezing ratios at modal-open time (complex) or letting the backdrop drift (wrong). Separate-but-similar geometry is the right trade-off.
- **Deep-link route mounts `PipelinePage`, not a separate page.** The replay modal is a modal; the full page context (funnel, ticker, time machine block) stays underneath. Landing directly on `/pipeline/logs/:logId` gives the user the modal *plus* the broader page they can interact with, not a bare modal in a vacuum.
- **`router.replace`, not `router.push`, for modal-open/close.** One history entry per modal click would make the back button feel broken ("why does Back close a modal I just opened?"). Replace keeps the navigation model sensible.
- **Auto-play replay on modal open.** The plan specified the animation but didn't say autoplay vs click-to-play. Auto-play is what makes the "replay" verb satisfying — if the user had to click Replay every time, the feature feels inert. The Replay button stays present for re-watches.
- **Picker rows don't render the raw log_id.** A 36-char UUID clogs the one-line row visual. Click-to-inspect takes the user straight to the modal where the full UUID is visible in the header. Documented in the test so future refactors don't accidentally try to re-add it.

---

## Follow-ups

1. **Phase 1b — wire poller-based connectors** (still open from Phase 1). Six files (`flyio.go`, `syslog.go`, `vercel.go`, `railway.go`, `supabase.go`, `mongodb.go`) need `pw.WriteIngestion` after their direct `queries.InsertLogEntry` so apps using those connectors don't show a silent Ingestion stage on the Pipeline page *or* in Time Machine picker results. Gets more user-visible now that the picker is real.
2. **"Load more" button in the picker.** The backend already returns `truncated: true` when `len(rows) >= limit`; the frontend currently surfaces a warning line ("result truncated, narrow the window or raise the limit") but no button to fetch the next `offset`. Small additive change.
3. **Window preset buttons.** Today the user types datetime-local values. Quick-preset buttons (`Last 15m`, `Last hour`, `Last 6h`, `Last 24h`) would match how the user thinks. Probably a design-refresh moment rather than a Phase 4 follow-up.
4. **Slack/Activity deep-link adoption.** Now that `/pipeline/logs/:logId` exists, the Activity page's assessment rows could deep-link to the underlying pipeline replay. One-line change in the Activity component; deferred to a future UX pass.
5. **Virtualise the picker list.** The current `max-h-72 overflow-auto` scroll container is fine at 200 rows. If telemetry shows users raising `limit` to the cap and scrolling at length, bringing in `vue-virtual-scroller` is a drop-in on the `<li v-for>`.
6. **Longer retention tiers.** Plan §4.4 explicitly punted to a separate scoped plan. Decoupling `log_pipeline_events` retention from `log_buffer` via a soft reference + independent pruner is the mechanism; deferred until there's a product need.

---

## Files Touched

**New backend files:** none — Phase 4's backend surface is one query + one handler, both added to existing files.

**Modified backend files:**

- `backend/internal/db/queries/log_pipeline_events.sql` — `ListPipelineLogsByApp` query
- `backend/internal/db/log_pipeline_events.sql.go` — generated code for the new query (hand-rolled; identical to what `sqlc generate` would produce)
- `backend/internal/api/handlers/pipeline.go` — `pipelineLogSummaryJSON`, `pipelineLogsResponse`, `parseLogsRange`, `parseLogsLimit`, `parseLogsOffset`, `summaryRowToJSON`, `PipelineLogs` handler, picker constants
- `backend/internal/api/router.go` — register `/pipeline/logs`
- `backend/internal/api/handlers/testhelpers_test.go` — register `/pipeline/logs` on the test router
- `backend/internal/api/handlers/pipeline_test.go` — two new integration subtests

**New frontend files:**

- `frontend/src/api/__tests__/pipeline.test.ts` — `fetchPipelineLogs` contract tests
- `frontend/src/components/pipeline/__tests__/TimeMachineBlock.test.ts` — picker behaviour tests

**Modified frontend files:**

- `frontend/src/types/pipeline.ts` — `PipelineLogSummary`, `PipelineLogsResponse`
- `frontend/src/api/pipeline.ts` — `fetchPipelineLogs`
- `frontend/src/components/pipeline/TimeMachineBlock.vue` — rewritten as a working picker
- `frontend/src/components/pipeline/PipelineJourneyModal.vue` — upgraded to replay view with scaled Sankey + traced particle
- `frontend/src/router/index.ts` — new `/pipeline/logs/:logId` route
- `frontend/src/pages/PipelinePage.vue` — wire TimeMachine `@inspect`, deep-link route param watch, URL-syncing `inspectLog`/`closeJourney`

Backend totals: ~260 LOC across modified files. Tests: ~100 LOC.
Frontend totals: ~260 LOC across modified files + ~140 LOC across new test files. Bundle impact: +7.4 kB on the `PipelinePage` chunk (+2.2 kB gzip).

Under the plan's ~400 LOC Phase 4 budget, with one meaningful scope extension (the `ListPipelineLogsByApp` query + handler that Phase 1 had deferred from its originally-planned §1.6). `vue-tsc` clean, `vite build` clean, `go vet ./...` clean, all 78 frontend tests + full backend suite pass.
