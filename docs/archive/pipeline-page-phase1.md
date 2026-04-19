# Pipeline Page — Phase 1 Completion

**Scope:** Backend persistence + in-memory pub/sub + HTTP endpoints that record every log's journey through the five-stage pipeline (Ingestion → Lumber → Gate → Agent → Activity) and stream live events to SSE subscribers. Frontend is Phase 2; user-visible surface is zero in Phase 1.
**Plan:** `docs/executing/pipeline-page-plan.md`

---

## The Problem This Solves

Heimdall's pipeline is a black box today. A log comes in via a webhook, gets classified by Lumber, either survives the severity gate or doesn't, maybe gets sent to Claude for assessment, and eventually lands in the activity feed — but nothing in the product shows that journey. Operators have no way to answer "did my log get classified?", "why did this one get escalated to the LLM?", or "what happened between ingestion and the assessment I'm looking at?" Debugging a misconfiguration means reading server logs.

Phase 1 lays the foundation for a Pipeline page that renders the journey visually: every log writes a row per stage to a new `log_pipeline_events` table and publishes the same event to an in-memory bus that an SSE endpoint fans out to the browser. Phase 2 builds the Sankey/funnel UI on top; Phase 4 wires a per-log replay ("Time Machine") using the same data. Phase 1 is pure plumbing — no user-visible change, but from the moment it ships every log's full history is being recorded so when Phase 2 lands the page has real data on day one.

---

## What Changed

### 1. New table — `log_pipeline_events`

**File:** `backend/migrations/038_log_pipeline_events.up.sql` (+ `.down.sql`)

One row per log per stage. Schema:

- `log_id UUID NOT NULL REFERENCES log_buffer(id) ON DELETE CASCADE` — the spine. Every pipeline event is anchored to the `log_buffer` row it describes, and the cascade means the existing 48h pruner (`agent/pruner.go`) sheds pipeline events in lockstep. No new retention logic.
- `app_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE` — for the per-app queries that back `/bootstrap` and `/stream`.
- `stage TEXT NOT NULL` — one of `'ingestion' | 'classified' | 'gate' | 'assessment'`. Plain TEXT (not enum) so a future stage adds one constant in Go, no DDL.
- Stage-specific nullable columns: `source_type`, `severity`, `type`, `category`, `confidence`, `summary`, `escalated`, `rule_hit`.
- `assessment_id UUID REFERENCES agent_log(id) ON DELETE SET NULL` — deliberately looser than CASCADE. Deleting an assessment row shouldn't erase the earlier stage events for the log it evaluated; the log's journey up to the gate should still be reconstructable.
- `metadata JSONB NOT NULL DEFAULT '{}'::jsonb` — forward-compat bag so rare extra fields don't need a column migration.

Three indexes: `(log_id, occurred_at)` for the journey lookup, `(app_id, occurred_at DESC)` for the ticker and `/bootstrap`, and `(app_id, stage, occurred_at DESC)` for the stage-partitioned aggregates in `PipelineStatsByApp`.

**RLS follows the system-table pattern** from migrations 030 and 033: enabled, zero policies. Non-owner Postgres roles (anon, authenticated) see zero rows; the backend owner-role pool bypasses RLS. The table is not in any direct-user access path — only `s.Queries` touches it — so a no-policy posture is the tightest setting that still works.

**Migration numbering:** 037 (source filter lookup index, 0.46.2 hardening) was the last slot. Pipeline is **038**. The 0.46.4 schema-drift audit confirmed prod matches `backend/migrations/*.up.sql` exactly through 037, so 038 lands on a verified-clean parent.

### 2. sqlc queries

**File:** `backend/internal/db/queries/log_pipeline_events.sql`

Five queries:

- `InsertPipelineEvent` — the single insert used by every stage call site. Metadata is passed as `json.RawMessage` (the writer always sends `{}` in Phase 1).
- `GetPipelineEventsByLog` — ordered stage history for a single log. Backs `/pipeline/logs/{logId}/journey` and the Time Machine replay in Phase 4.
- `GetRecentPipelineEventsByApp` — newest-N for an app, `DESC`. Used by `/bootstrap`'s `recentEvents` payload.
- `GetPipelineEventsSince` — catch-up drain for the SSE `?since=` replay. Capped via `$3` so the SSE handler can detect replay-cap overflow and tell the client to resync.
- `PipelineStatsByApp` — single-pass aggregate with `COUNT(*) FILTER (WHERE stage = ...)` per stage. Uses the `idx_lpe_stage` partial-ordered index so the planner doesn't scan the whole table for each bucket.

Regenerated via `sqlc generate`. One cleanup moment: the first pass used `COALESCE($13, '{}'::jsonb)` in the insert which left sqlc unable to infer the parameter type, producing a `Column13 interface{}` field. Dropped the COALESCE and kept the `NOT NULL DEFAULT '{}'` at the column level — sqlc then emitted `Metadata json.RawMessage` correctly.

### 3. `PipelineBus` — in-memory pub/sub

**File:** `backend/internal/agent/pipeline_bus.go` + `_test.go`

Per-app fan-out hub with non-blocking publish semantics. `Subscribe(appID) → (<-chan PipelineEvent, unsub func())`; `Publish(evt)` routes to every channel registered for `evt.AppID`.

**Buffered channels (cap 128) with non-blocking publish.** If a subscriber's buffer is full, its copy of the event is dropped and a counter ticks up. Log lines are emitted on drops #1, #2, #4, #8, #16… — a power-of-two cadence keeps the signal loud without spamming when a subscriber stays stuck. The alternative (blocking publish) would let one stalled browser tab back-pressure the entire ingestion path; unacceptable. Dropped events are recoverable because the persistence path (`log_pipeline_events`) runs before publish, so reconnecting + replay via `/bootstrap` + `?since=` closes the gap.

**Lock discipline:** the race-clean version holds `RLock` across the `select { case ch <- evt: default: }` block. The obvious snapshot-then-send alternative (copy channel list under RLock, release, then send) has a subtle race — a concurrent `unsub` can close a channel between the release and the send, panicking with "send on closed channel". The concurrent subscribe/unsub unit test under `-race` catches this in one run. Holding RLock is fine in practice because each per-channel send is bounded by a non-blocking `select`, so the lock is released in microseconds even under fan-out. Writers (subscribe/unsub) wait briefly; Go's RWMutex writer-preference prevents starvation.

**Drop counter and `SubscriberCount`** are exposed for Phase 3.1's `pipeline_bus_dropped_events_total` metric and the Phase 3.4 leak-canary sweeper.

### 4. `PipelineWriter` — persist-then-publish helper

**File:** `backend/internal/agent/pipeline_writer.go`

Thin helper that owns both the DB insert and the bus publish for a single event. Four stage-specific entry points — `WriteIngestion`, `WriteClassified`, `WriteGate`, `WriteAssessment` — each with a typed input struct.

**Persist first, publish second.** If the DB insert fails, the publish is skipped and the error is logged. Rationale: losing a live event is recoverable (replay via `/bootstrap`); leaving a particle on the page for a log whose journey isn't in the database is not — it'd imply functionality the Time Machine view can't actually deliver.

**Fire-and-forget from the caller.** All errors are logged at `WARN`; nothing is propagated back to `monitorApp` or the webhook handler. A `log_pipeline_events` outage must not degrade the underlying work. Mirrors the pre-existing `EmitLog` contract on `agent/emit.go`.

**Nil-safety by construction.** All four `Write*` methods return early if the receiver is nil. Tests and early-boot paths (pre-migration) can leave the writer unconfigured without crashing.

### 5. Severity gate rule IDs (Option A)

**File:** `backend/internal/agent/severity_gate.go`

`ShouldEscalate` was `(event lumber.Event) bool`; now it returns `(bool, string)`. The string is a stable rule ID per escalating branch — `error_type`, `request_server_error`, `access_login_failure`, `unknown_type`, etc. Safe branches return `RuleNone = ""`.

The IDs are exported constants (`RuleErrorType`, `RuleRequestSlowRequest`, …) so call sites and tests reference them by name rather than by string literal. Chosen per Option A in the plan — hardcoded mapping, no DB table, no expression engine. Alternatives documented in `pipeline-rule-options.md`.

Rule IDs land on `log_pipeline_events.rule_hit` for the Gate stage, so the Pipeline-page Gate detail panel can show a "rules that fired in the last hour" histogram without hard-coding Lumber taxonomy strings into the frontend.

### 6. `ClassifiedLog` + `Classifier.Classify` — return every log, not just the flagged subset

**Files:** `backend/internal/agent/classifier.go`, `classifier_lumber.go`, `classifier_test.go`

`ClassifiedLog` grew two fields: `Escalated bool` and `RuleID string`. The `Classifier.Classify` interface method changed from returning `(flagged []ClassifiedLog, safeCount int)` to returning `[]ClassifiedLog` — every input log, with the gate decision baked in.

**Why the signature change:** the Pipeline-page Classified and Gate stages need to emit events for *every* log, not just the escalated subset. The old interface threw away the safe-log metadata (only the count survived). Emitting inside the classifier would have required injecting a writer and a ctx into every Classifier implementation — cleaner to return the full list and let `monitorApp` drive the emit.

**Test refactor cost:** six call sites changed from `flagged, safeCount := c.Classify(logs)` to `flagged, safeCount := FilterFlagged(c.Classify(logs))`. `FilterFlagged` is a new free function that partitions by `Escalated`. Tests pass unchanged otherwise.

### 7. Instrumentation at three of four stages

**Files:** `backend/internal/agent/monitor.go`, `backend/internal/api/handlers/webhooks.go`, `backend/internal/api/handlers/otlp.go`, `backend/internal/api/handlers/source_filter_pipeline.go`

- **Classified + Gate:** `monitorApp` iterates the full `[]ClassifiedLog` returned by `a.classifier.Classify(logs)` and calls `pw.WriteClassified(...)` + `pw.WriteGate(...)` for each entry (escalated *and* safe). The detail panel needs both halves.
- **Assessment:** after `a.RunMonitoring` returns and the resulting `agent_log` row is inserted via `EmitLogWithSeverity`, the handler emits one Assessment event per log in the *post-cap* `escalated` slice (`monitorApp` caps at `maxFlaggedForLLM = 50` before calling the LLM). Guarded on `logEntryID != uuid.Nil` — a dangling `assessment_id` referencing an insert that failed would be misleading.
- **Ingestion (partial):** Emitted *after* `tx.Commit` in both `webhooks.go ingestEntries` and `otlp.go`, one event per inserted `log_buffer` row. Pre-commit emit would race the FK — the `log_pipeline_events.log_id` reference only resolves once `log_buffer` has committed.

**Returning inserted rows from `insertFiltered`.** Needed log IDs post-commit, so `insertFiltered` now returns `(sourceFilterStats, []insertedRow, error)` instead of `(sourceFilterStats, error)`. Both callers (webhooks + OTLP) pass the slice to `pw.WriteIngestion` after commit.

**Poller-based connectors not wired.** `flyio.go`, `syslog.go`, `vercel.go`, `railway.go`, `supabase.go`, `mongodb.go` all write to `log_buffer` independently of the shared `insertFiltered` helper. Wiring those six adds significant surface that wasn't the spec's focus ("webhook path or equivalent"). Classified + Gate + Assessment still fire for their logs (because `monitorApp` reads from `log_buffer` regardless of how the log got there), so the Pipeline page funnel is *mostly* complete for those apps — only the Ingestion stage is silent. Explicitly deferred to Phase 1b.

### 8. HTTP endpoints

**File:** `backend/internal/api/handlers/pipeline.go`

Three routes:

- `GET /api/apps/{appId}/pipeline/bootstrap?window=1h&tickerLimit=50` — bundled first-paint response: `{ stats, recent_events, cursor }` in one roundtrip. Stats come from `PipelineStatsByApp`; `recent_events` from `GetRecentPipelineEventsByApp`; `cursor` is the newest row's `occurred_at` (or `now()` if empty). Replaces the plan's originally-two-endpoint design with one — the ticker paints synchronously with the rest of the page instead of after a follow-up fetch. Clamps: `window` to 24h max (default 1h), `tickerLimit` to 200 (default 50).
- `GET /api/apps/{appId}/pipeline/logs/{logId}/journey` — ordered stage history for a single log. **Defence-in-depth:** `authorizeApp` verifies the URL's `appId`, but the underlying `GetPipelineEventsByLog` doesn't filter by `app_id`. Without a per-row `row.AppID != app.ID` check, a user could guess another tenant's `log_id` and pull its journey. The handler cross-checks and 404s on mismatch. There's a dedicated test for this (`TestPipeline_Integration/journey_refuses_cross-app_log`) that constructs a foreign org + app and asserts the 404.
- `GET /api/apps/{appId}/pipeline/stream?since=<rfc3339>&token=<jwt>` — SSE endpoint. Sequence per connection: (1) validate `token` via `middleware.ValidateJWT` + check app ownership via `GetApplicationByOrgUser`; (2) if `since` is set, drain `GetPipelineEventsSince` up to a 500-row cap and write each as `event: replay`; (3) if the cap was hit, emit `event: resync` and close cleanly so the client refetches `/bootstrap` with a fresh cursor; (4) otherwise subscribe to `PipelineBus` and forward live events as `event: live`; heartbeat every 15s so proxies don't idle-close. Auth via `?token=` matches the existing WebSocket chat pattern — `EventSource` can't set custom headers. The stream route lives *outside* the `MaxBodySize`-wrapped protected group in `router.go` because streaming responses shouldn't be body-limited.

### 9. Wire-up in `main.go` and `Agent`

**Files:** `backend/cmd/heimdall/main.go`, `backend/internal/agent/agent.go`

Construct one `PipelineBus` and one `PipelineWriter` per process (they're both safe-for-concurrent-use singletons), inject both into `agent.New(...)`. The Agent holds them and exposes `a.Pipeline()` + `a.PipelineBus()` as nil-safe accessors so the handlers package can reach them via `s.Agent.Pipeline()` / `s.Agent.PipelineBus()` without introducing a circular dep between `agent` and `handlers`.

---

## Tests

### Unit — `agent/pipeline_bus_test.go`

Four subtests, all pass under `-race`:

- **Fanout** — two subscribers on app A both receive an event; a subscriber on app B receives nothing.
- **DropsWhenFull** — fill the subscriber buffer (128 events), then publish 10 more in a goroutine; assert the publish doesn't block and `Dropped()` reports 10.
- **UnsubscribeCleanup** — after `unsub()`, the channel is closed, `SubscriberCount` drops to 0, and a double-`unsub` is idempotent.
- **ConcurrentSubscribeUnsubscribe** — 16 goroutines × 50 iterations of subscribe/publish/drain/unsub. This test caught the race bug in the original snapshot-then-send implementation of `Publish` — without `-race` it would have reached prod and crashed on reconnect storms (every SSE reconnect = unsubscribe + resubscribe).

### Integration — `api/handlers/pipeline_test.go`

Five subtests against real Postgres (skips when `DATABASE_URL` unset), all pass:

- **four-stage walk persists and fans out** — create a log, walk it through all four `pw.Write*` methods, assert (a) the subscriber receives four events with matching `log_id`, (b) `GetPipelineEventsByLog` returns four rows in order, (c) the Gate row carries `rule_hit = RuleErrorType`.
- **cascade delete sheds pipeline events with log_buffer** — precondition four rows, `DELETE FROM log_buffer WHERE id = $1`, assert zero rows remain. Verifies the `ON DELETE CASCADE` actually matches the pruner's behaviour.
- **bootstrap returns stats and recent events** — hit `/pipeline/bootstrap?window=1h&tickerLimit=10`, assert stats counts ≥ 1 for ingestion/classified/flagged/assessment, `window_seconds = 3600`, non-empty `recent_events`, non-zero `cursor`.
- **journey endpoint returns ordered stages for a log** — walk a log, GET `/pipeline/logs/{logId}/journey`, assert 4 ordered events.
- **journey refuses cross-app log** — construct a foreign org + user + app + connection + log, walk that log, then request the journey URL with the test user's `appId` but the foreign log's id. Must 404.

---

## Decisions Worth Remembering

- **Non-blocking publish with a drop counter > blocking publish**. Back-pressure from SSE subscribers into the ingestion hot path is unacceptable; reconstructability via persistence + replay makes drops recoverable.
- **Hold RLock across non-blocking sends in `Publish`**. Unconventional — "don't hold locks across channel sends" is Go orthodoxy — but only applies to blocking sends. `select/default` is microsecond-bounded, and the snapshot-then-send alternative has a closed-channel race.
- **Return all logs from `Classifier.Classify`, not just flagged**. The "emit inside the classifier" alternative required threading a writer + ctx into every implementation. Returning `[]ClassifiedLog` with `Escalated` lets `monitorApp` drive the per-log emit cleanly.
- **One `/bootstrap` endpoint, not stats + ticker separately**. First-paint latency matters for the ticker — synchronous render beats a second fetch-and-hydrate cycle.
- **Ingestion events post-commit, not inside the transaction**. FK resolution against `log_buffer` requires the parent row to be visible to the writer's query, which means post-commit.
- **`ON DELETE CASCADE` on `log_id`, `ON DELETE SET NULL` on `assessment_id`**. Log disappearing = the journey is gone; assessment disappearing = the journey up to the gate should still reconstruct.
- **`TEXT` stage column, not enum**. Future `notification_dispatch` stage adds one constant in Go, no migration.

---

## Follow-ups

1. **Phase 1b — wire poller-based connectors.** Six files (`flyio.go`, `syslog.go`, `vercel.go`, `railway.go`, `supabase.go`, `mongodb.go`) each need a `pw.WriteIngestion` call after their direct `queries.InsertLogEntry`. Until then, apps using those connectors see a silent Ingestion stage on the Pipeline page (but Classified/Gate/Assessment still work for their logs via `monitorApp`).
2. **Phase 2 — frontend Sankey + particle layer + ticker + stubbed Time Machine block.** Unblocked — all backend contracts (types, endpoints, SSE frame format) are stable.
3. **Phase 3 hardening.** Drop-counter metric (3.1), SSE connection cap per user (3.2), `/bootstrap` 5s in-memory cache (3.3), abandoned-subscriber leak canary (3.4), RLS verification pass (3.5).
4. **Classifier emit during ingestion cap.** `monitorApp` caps at `maxFlaggedForLLM = 50`; the dropped flagged logs currently emit Classified + Gate events but no Assessment event (correct — they weren't assessed). Users seeing the funnel may want a visual indicator that some flagged logs were dropped rather than assessed. Small UX note, not a Phase 1 blocker.

---

## Files Touched

**New backend files:**

- `backend/migrations/038_log_pipeline_events.up.sql` + `.down.sql`
- `backend/internal/db/queries/log_pipeline_events.sql` (+ generated `internal/db/log_pipeline_events.sql.go`)
- `backend/internal/agent/pipeline_bus.go` + `pipeline_bus_test.go`
- `backend/internal/agent/pipeline_writer.go`
- `backend/internal/api/handlers/pipeline.go` + `pipeline_test.go`

**Modified backend files:**

- `backend/cmd/heimdall/main.go` — construct `PipelineBus` + `PipelineWriter`, inject into `agent.New`
- `backend/internal/agent/agent.go` — new fields, accessors, updated `New` signature
- `backend/internal/agent/monitor.go` — Classified/Gate/Assessment emit sites + new `FilterFlagged` consumer
- `backend/internal/agent/classifier.go` — `ClassifiedLog.Escalated` + `RuleID`, interface return-shape change, `FilterFlagged` helper, Passthrough update
- `backend/internal/agent/classifier_lumber.go` — signature change, per-log gate decision inline
- `backend/internal/agent/classifier_test.go` — call-site updates (`FilterFlagged(c.Classify(...))`)
- `backend/internal/agent/severity_gate.go` — returns `(bool, ruleID)`, rule-ID constants
- `backend/internal/agent/severity_gate_test.go` — asserts non-empty rule on escalating branches, `RuleNone` on safe
- `backend/internal/api/handlers/webhooks.go` — post-commit Ingestion emit, `agent` import
- `backend/internal/api/handlers/otlp.go` — same post-commit Ingestion emit
- `backend/internal/api/handlers/source_filter_pipeline.go` — `insertFiltered` returns `[]insertedRow`
- `backend/internal/api/handlers/testhelpers_test.go` — test router registers `/pipeline/bootstrap` + `/pipeline/logs/{logId}/journey`
- `backend/internal/api/router.go` — route registration (2 protected, 1 public-with-token-auth)

Backend totals (excluding generated code and tests): ~550 LOC across new files + ~80 LOC across modified files. Tests: ~290 LOC (bus + integration). Matches the plan's ~600 LOC Phase 1 budget.
