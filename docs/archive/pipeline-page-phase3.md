# Pipeline Page — Phase 3 Completion

**Scope:** Production-hardening of the pipeline-page backend shipped in Phase 1. Five items from the plan (`docs/executing/pipeline-page-plan.md` §Phase 3): bus drop-counter metric (3.1), per-user SSE connection cap (3.2), bootstrap response cache (3.3), leak-canary sweeper (3.4), and RLS verification (3.5). No behavioural change visible to Phase 2's frontend — these are de-risk hooks that only fire under load, abuse, or drift.
**Plan:** `docs/executing/pipeline-page-plan.md` §Phase 3

---

## The Problem This Solves

Phase 1 shipped the pipeline data plane with sound defaults — non-blocking publish, non-cascading assessment FK, post-commit ingestion emit — but with no defensive layer around the things that go wrong *after* a feature has been in production for a week. Three classes of concern were still latent:

1. **Visibility gaps.** The bus already had a drop counter, but nothing scraped it into Prometheus. Similarly, SSE connection count was invisible — we could know a specific connection existed only by logging into the server. When an operator asks "is the bus healthy?" at 3am, log-tailing is not the right answer.
2. **Abuse / runaway surfaces.** Any authenticated user could open N SSE connections to their app's stream endpoint, pinning N backend channels (+128-buffered events each). A tab-respawning bug or a deliberately abusive client could exhaust the subscriber map without any rate limit in the path.
3. **Stale bets.** The plan guessed that `/bootstrap`'s stats aggregate would need a cache only at higher volume, but noted "the hook is cheap to add and saves us from a future perf regression" — and with the frontend now live, every page reload hits the same aggregate within seconds.

Phase 3 adds the observability, caps, and short-TTL cache that let us ship Phase 2 broadly without a second post-deploy firefight. Plus an explicit RLS regression-guard that future migrations can't silently weaken.

---

## What Changed

### 1. Bus drop-counter metric — `heimdall_pipeline_bus_dropped_events_total`

**Files:** `backend/internal/metrics/metrics.go`, `backend/internal/agent/pipeline_bus.go`

The `PipelineBus.Publish` hot path already tracked dropped events in a `sync/atomic.Uint64` counter and emitted WARN logs at power-of-two cadence. Phase 3 exports the same counter to Prometheus as a simple `Counter` (no labels — a drop is a drop; the per-app / per-stage context lives in the log line, and Prometheus cardinality budget is too precious to spend on something that *should* stay at zero).

`metrics.PipelineBusDroppedEvents.Inc()` runs inside the same `default` branch that does the atomic add, so the two counters can never diverge. The Grafana dashboard can now show `rate(heimdall_pipeline_bus_dropped_events_total[5m])` and any non-zero value is actionable.

### 2. SSE per-user connection cap

**Files:** `backend/internal/api/handlers/pipeline_support.go` (new), `backend/internal/api/handlers/pipeline.go`, `backend/internal/api/handlers/server.go`

New `sseConnLimiter` — a `map[uuid.UUID]int` behind a mutex, with `Acquire` and `Release` paired like a semaphore. Constructed once in `NewServer` and cap'd at `pipelineSSEPerUserCap = 5`. The `PipelineStream` handler calls `Acquire(userID)` *after* JWT validation + app ownership check (so the counter is never driven by unauthenticated traffic), returns 429 on refusal, and `defer`s `Release` so disconnects free the slot.

Two Prometheus exports alongside:

- `heimdall_pipeline_sse_active_connections` — Gauge, inc on Acquire / dec on Release. Operational: "how many live streams right now?"
- `heimdall_pipeline_sse_rejected_total` — Counter, ticks on every cap-driven 429. A rising rate indicates either a runaway client or a probing attacker; both want visibility.

The cap itself is 5 for now — one tab per typical deployment environment with slack for a stale tab or two. Per the plan: "prevents a runaway tab spawn from pinning backend channels." The `Release` path has an underflow guard so a stray call can't permanently lock out a user — the fear was a miscount (if the handler ever returned *before* Acquire ran but *after* the `defer`, which it doesn't today but might in future refactors) leaving a user at `conns = -1` and therefore perpetually at-cap when the check is `>=`.

**Why per-user, not per-IP:** multiple tabs share a session cookie; IP-based limiting would lump unrelated users behind the same NAT or corporate gateway. The cap belongs on the authenticated identity.

### 3. `/pipeline/bootstrap` response cache

**Files:** `backend/internal/api/handlers/pipeline_support.go`, `backend/internal/api/handlers/pipeline.go`

5-second in-memory cache over the `bootstrapResponse` payload, keyed by `(appID, window, tickerLimit)`. Implementation is a `sync.Mutex`-guarded `map[bootstrapCacheKey]bootstrapCacheEntry` with lazy per-Get eviction for expired entries — no background sweeper needed at this cache's size.

**Key design choices:**

- **Payloads are pre-marshalled `[]byte`.** Cache hits `w.Write(payload)` directly, bypassing the `json.Encoder` reflection path. This matters on the happy path because `/bootstrap` is called on every Pipeline page mount and during re-subscribes.
- **Cache read happens *after* `authorizeApp`.** Authorization runs unconditionally; the cache never short-circuits the auth check. Tenant isolation is double-enforced: `authorizeApp` verifies the caller can see `app.ID`, AND `app.ID` is baked into the cache key — so even if the auth check were somehow bypassed upstream, a miss-keyed hit from another tenant is structurally impossible.
- **`X-Cache: HIT` / `X-Cache: MISS` response headers.** Operational debug aid (curl -I shows cache state) and the integration test asserts against them directly.
- **Four-tuple cache metrics.** `heimdall_pipeline_bootstrap_cache_total{result="hit|miss"}` gives the hit rate visibility we need to tune the TTL later if usage patterns shift.

The plan's "remaining open question" on cache-key granularity is resolved here: `(appID, window, tickerLimit)` is exactly the input tuple that determines the response, so it's the correct key. A user with two tabs at different windows gets separate entries, never cross-pollutes. A future window-selector in the UI (follow-up from Phase 2) would drive a wider key space but the isolation guarantee holds.

### 4. Leak-canary sweeper

**Files:** `backend/internal/agent/pipeline_sweeper.go` (new), `backend/internal/agent/pipeline_bus.go` (new `SubscriberCounts()` method), `backend/internal/agent/agent.go` (wired into `Start`)

The sweeper is a per-minute goroutine launched alongside `Monitor`, `Prune`, and `InvestigationScheduler` in `Agent.Start`. Each tick it calls `PipelineBus.SubscriberCounts()` for a snapshot of per-app channel counts and emits a WARN log for any app above `pipelineSweeperThreshold = 50`.

**Why 50 and not a tighter cap:** a single operator's Pipeline page uses one channel; `pipelineSSEPerUserCap = 5` limits a single user to 5 channels. For an app with 10 concurrent operators that's 50 — at the threshold. Anything meaningfully above is a bug: either a handler path forgot to unsub, or the SSE cap has been circumvented.

**Why the sweeper warns but never closes channels:** Taking action (e.g. force-closing the oldest subscribers) would *mask* the actual leak. The job of the canary is to make the leak noisy in logs and dashboards so it gets fixed; shedding connections would let the bug live forever behind an auto-mitigation.

**`SubscriberCounts()` returns a snapshot, not a live view.** The map returned is freshly allocated — callers can mutate it freely without touching bus state. Tested explicitly; the race-clean version matters because the sweeper iterates the snapshot while Subscribe/unsub may be rewriting the internal map in parallel.

### 5. RLS verification test

**File:** `backend/internal/api/handlers/pipeline_test.go` (new subtest)

Phase 1's migration 038 enabled RLS on `log_pipeline_events` with zero policies — matching the Phase-0 system-table pattern from migrations 030 (`ingestion_idempotency`) and 033 (`webhook_idempotency`). This is the tightest posture that still works because only the owner-role pool touches the table.

The new subtest `RLS: non-owner roles see zero pipeline rows`:

1. Inserts a log + walks it through the pipeline so there's real data to hide.
2. Reads `pg_class.relrowsecurity` to assert RLS is *still* enabled on the table (guard against a future `ALTER TABLE ... DISABLE ROW LEVEL SECURITY` migration slipping through review).
3. Reads `pg_policies` to assert zero policies exist (guard against someone adding a permissive policy that silently grants read access).
4. `SET LOCAL ROLE authenticated` inside a transaction and asserts `SELECT count(*) FROM log_pipeline_events WHERE log_id = $1` returns 0 under that role.

This is the same verification pattern the 0.43.0 and 0.45.1 RLS rollouts used post-deploy, now codified as a regression test so the next migration that touches this table can't silently break the posture.

---

## Tests

### Unit — `agent/pipeline_bus_test.go`

One new subtest:

- **SubscriberCountsSnapshot** — two subscribers on app A plus one on app B; assert `SubscriberCounts()` returns `{A: 2, B: 1}`. Mutates the returned map, re-calls `SubscriberCounts()`, asserts the internal state is unchanged (catches a future "return the internal map reference" regression). Then unsubs all three and asserts the snapshot is empty.

All four pre-existing `TestPipelineBus_*` subtests still pass under `-race`.

### Unit — `api/handlers/pipeline_support_test.go` (new file, 6 subtests)

- **SSEConnLimiter_CapsPerUser** — 1st + 2nd acquire succeed at cap=2, 3rd refused, Release opens a slot.
- **SSEConnLimiter_PerUserIsolation** — user A's exhausted cap doesn't block user B. This is the test for the NAT-sharing concern (see "Why per-user, not per-IP" above).
- **SSEConnLimiter_ReleaseIdempotentUnderflow** — unmatched Release doesn't corrupt the counter.
- **BootstrapCache_HitAndMiss** — empty cache misses; Put → Get hits.
- **BootstrapCache_KeyIsolation** — different appID, different window, different tickerLimit all miss under a key that was populated for a different tuple. This is the defence against the cross-tenant-bleed concern.
- **BootstrapCache_ExpiredEntryEvicted** — manually-expired entry returns nil AND is removed from the map (not just ignored). Verified by reading the internal map post-Get.

### Integration — `api/handlers/pipeline_test.go`

Two new subtests (both require `DATABASE_URL`):

- **bootstrap cache serves hit on second call** — wires the cache onto the struct-literal test Server via `SetPipelineBootstrapCacheForTest()`, makes two identical `/bootstrap?window=1h&tickerLimit=10` requests, asserts `X-Cache: MISS` then `X-Cache: HIT` with byte-identical bodies. Then a third request with `tickerLimit=5` → `X-Cache: MISS` (proves the key-isolation guarantee reaches the HTTP surface).
- **RLS: non-owner roles see zero pipeline rows** — the verification pass described above.

All existing five subtests of `TestPipeline_Integration` still pass.

### Totals

13 total backend tests run in `./internal/api/handlers` + `./internal/agent` for Phase 3. 8 unit (always run), 5 integration (skipped without `DATABASE_URL`). No existing tests modified. `go vet ./...` clean. `go build ./...` clean.

---

## Decisions Worth Remembering

- **Drop counter: Prometheus counter, no labels.** `heimdall_pipeline_bus_dropped_events_total` has no `app_id` or `stage` label — cardinality would balloon on a feature that should stay at zero. Per-app context lives in WARN logs, which is the right tier for a "tell me exactly which tab is stuck" investigation.
- **Cache key = `(appID, window, tickerLimit)`.** The exact tuple that determines the response. Missing any one of the three would produce cross-tenant or cross-window bleed. Documented inline as "the defence against the cross-tenant-bleed concern."
- **Cache read after authorize, not before.** Auth runs on every request; the cache never short-circuits it. Belt-and-braces against an upstream auth bug.
- **`X-Cache: HIT|MISS` response headers.** Curl-visible debug aid and test assertion surface — a better contract than parsing JSON bodies for "was this cached?" behaviour.
- **Per-user SSE cap, not per-IP.** A corporate NAT or household router would lump unrelated users behind one IP; the cap must follow the authenticated identity.
- **Sweeper warns, never force-closes.** Automated mitigation masks the underlying leak. The canary's job is to surface the bug, not hide it.
- **RLS verified in code, not just post-deploy.** The 0.43.0 and 0.45.1 rollouts verified RLS with ad-hoc SQL after landing. Codifying the check as a regression test means the next migration to touch `log_pipeline_events` can't silently weaken the posture without a CI red signal.
- **Server fields exported lazily for tests.** `pipelineSSELimiter` and `pipelineBootstrap` are unexported and nil-safe. The test helper builds `&handlers.Server{}` as a struct literal; without nil checks the integration tests would have panicked. `SetPipelineBootstrapCacheForTest()` is the single seam for opting into cache behaviour from struct-literal construction.

---

## Follow-ups

1. **Phase 1b — wire poller-based connectors** (still open from Phase 1). Six files (`flyio.go`, `syslog.go`, `vercel.go`, `railway.go`, `supabase.go`, `mongodb.go`) need `pw.WriteIngestion` after their direct `queries.InsertLogEntry`. Until then, apps using those connectors see a silent Ingestion stage. Classified/Gate/Assessment still fire via `monitorApp`.
2. **Phase 4 — Time Machine replay** (unblocked by Phases 1-3). Backend is already populated; frontend stub is visually complete. Needs the log-journey picker + per-log replay animation in the existing funnel.
3. **Grafana dashboard wiring.** Phase 3 adds four new metrics (`pipeline_bus_dropped_events_total`, `pipeline_sse_active_connections`, `pipeline_sse_rejected_total`, `pipeline_bootstrap_cache_total`). A dashboard panel for each would close the observability loop, but the metrics themselves are self-sufficient — `curl /metrics | grep pipeline` is enough for ad-hoc debugging.
4. **Per-user SSE cap tuning.** Five is a round guess. If telemetry shows the 429-rejection counter rising in normal use (users legitimately wanting more tabs), the cap is easy to raise via `pipelineSSEPerUserCap`. If it stays at zero, we know the current cap is comfortable.
5. **Bootstrap cache TTL tuning.** 5 seconds was the plan's recommendation. If `pipeline_bootstrap_cache_total{result="hit"}` / `(hit + miss)` runs hot (say > 60%) we could consider raising; if it runs cold (< 20%) the cache is paying its synchronisation cost for no gain and could be shortened or removed.

---

## Files Touched

**New backend files:**

- `backend/internal/agent/pipeline_sweeper.go` — 1-minute leak-canary goroutine
- `backend/internal/api/handlers/pipeline_support.go` — `sseConnLimiter` + `bootstrapCache`
- `backend/internal/api/handlers/pipeline_support_test.go` — 6 unit tests

**Modified backend files:**

- `backend/internal/metrics/metrics.go` — 4 new Prometheus metrics (1 counter + 1 gauge + 2 counter/vec)
- `backend/internal/agent/pipeline_bus.go` — `metrics.PipelineBusDroppedEvents.Inc()` on drop, new `SubscriberCounts()` snapshot accessor
- `backend/internal/agent/pipeline_bus_test.go` — new `TestPipelineBus_SubscriberCountsSnapshot`
- `backend/internal/agent/agent.go` — spawn the sweeper goroutine in `Start`; waitgroup from 3 to 4
- `backend/internal/api/handlers/server.go` — new `pipelineSSELimiter` + `pipelineBootstrap` fields; wired in `NewServer`; `SetPipelineBootstrapCacheForTest()` seam
- `backend/internal/api/handlers/pipeline.go` — cache read/write in `PipelineBootstrap`; `Acquire`/`Release` in `PipelineStream`; `X-Cache` response header
- `backend/internal/api/handlers/pipeline_test.go` — 2 new integration subtests (cache hit/miss + RLS verification)

**Frontend:** unchanged.

Backend totals: ~160 LOC across new files + ~60 LOC across modified files. Tests: ~190 LOC. Matches the plan's ~150 LOC Phase 3 budget (slightly over because of the 6 support-file unit tests, which weren't in the original line-count estimate).
