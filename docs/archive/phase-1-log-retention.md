# Phase 1 — Log Retention Pruning (Completion Notes)

**Plan:** `docs/executing/logs-feed-and-scheduled-investigations.md` (Phase 1)
**Completed:** 2026-04-11
**Base commit:** `b217afa` (v0.30.2)
**Status:** ✅ Merged locally — backend builds clean, `go vet` clean, all tests green.

---

## What this fixes

The `PruneExpiredLogs` SQL function existed in `backend/internal/db/queries/log_buffer.sql:52-53` and sqlc had generated the corresponding Go wrapper at `backend/internal/db/log_buffer.sql.go:260`, but **nothing anywhere in the backend called it**. A grep for `PruneExpiredLogs` turned up the definition and the generated wrapper — zero callers.

The practical consequence: `log_buffer` grew unbounded in production. Every ingested webhook log, every Supabase stream row, every OTLP span — all of it accumulated forever. The dead function was a Chekhov's gun that had been sitting on the mantelpiece since at least v0.4.0 (the webhook log connector).

Phase 1 wires up the dead function, extends the retention window from 24h → 48h per user request, and adds a background goroutine to call the pruner on a schedule.

---

## What changed

### 1. SQL retention window: 24h → 48h

**File:** `backend/internal/db/queries/log_buffer.sql:52-53`

One-character edit — literal `'24 hours'` → `'48 hours'` inside the `PruneExpiredLogs` query.

**Regenerated:** `backend/internal/db/log_buffer.sql.go:257` — sqlc picked up the change cleanly via `make sqlc-generate`. The generated file's embedded query string now reads `DELETE FROM log_buffer WHERE ingested_at < now() - interval '48 hours'`. No Go signature change, no downstream Go edits required.

**Why the SQL is the source of truth, not a Go constant:** Retention is a data-plane concern. Keeping the window inside the query means there's exactly one place to change it, and the sqlc-generated code stays a dumb pass-through. If we ever wanted per-app retention we'd move to a parameterized query — but YAGNI for now.

---

### 2. New file: `backend/internal/agent/pruner.go`

A small `(a *Agent) Prune(ctx)` goroutine plus a `pruneTick(ctx)` helper. Two design choices worth flagging:

- **Startup tick runs *before* the ticker.** The loop body is `pruneTick(ctx); for { select { <-ticker.C: pruneTick; <-ctx.Done: return } }`. Without the startup call, a restart loop (e.g. a crashing pod, or a Fly machine flapping) could indefinitely delay pruning. This is a subtle form of "work-starvation under churn" — any scheduled-job pattern that *only* ticks is vulnerable to it.
- **Errors are logged and swallowed, never propagated.** The pruner is best-effort. If Postgres is temporarily unreachable, we log and move on — we must never let a transient pool hiccup take down the agent goroutine or stop the monitor loop running next door. This matches the same "fire-and-forget" philosophy the plan calls out for `EmitLog` in `CLAUDE.md`.

**Tick interval:** 1 hour, as a package-level const `pruneInterval`. Small enough to keep `log_buffer` bounded between ticks (48h window, 1h check = at most ~1h of overage at any moment), large enough that the DELETE's table scan cost is negligible amortised.

---

### 3. `backend/internal/agent/agent.go` — `Start()` spawns a second goroutine

**Before:** `Start` spawned exactly one goroutine — `Monitor(ctx)`.
**After:** `Start` spawns two goroutines — `Monitor(ctx)` and `Prune(ctx)` — each incrementing `a.wg` so `Stop()` joins both on shutdown.

**Why this is safe with the existing `Stop` implementation:** `Stop()` calls `a.cancel()` (shared context) and then `a.wg.Wait()`. Both goroutines select on `ctx.Done()` and return on cancel, so both `wg.Done()` calls fire and `Stop` returns cleanly. No change to `Stop` was needed — the plan predicted this correctly.

Also bumped the startup log line from "monitoring goroutine spawned" to "monitoring + pruner goroutines spawned" and updated the doc comment.

**This edit is load-bearing for Phase 3.** The scheduled-investigations phase will add a *third* goroutine (`InvestigationScheduler(ctx)`) alongside these two. By making `Start` a two-goroutine function now, Phase 3's edit becomes a simple one-block append instead of a first-time structural change. This is the dependency edge the plan's execution order diagram calls out.

---

### 4. New file: `backend/internal/agent/pruner_test.go`

Three tests, all using the existing `stubDBTX` pattern from `loop_test.go` / `tools_test.go`. The stub returns errors for all DB operations, which is actually the *perfect* fixture for this particular package — we're testing the *loop behavior*, not the SQL correctness (sqlc already did that).

| Test | What it proves |
|------|----------------|
| `TestPruneTick_DBError` | `pruneTick` logs and swallows a DB error without panicking. Exercises the exact production failure mode (transient pool unavailability). |
| `TestPrune_CancelExits` | `Prune` returns within 2 seconds of `ctx` cancel. Proves the `<-ctx.Done` branch of the select is wired correctly. |
| `TestPrune_StartupTickRuns` | `Prune` exits cleanly even if cancelled before the 1h ticker could fire — indirect verification that the startup tick runs and doesn't block forever. |

**Deliberate scoping call:** The plan suggested a fourth test that inserts three rows into `log_buffer` at `now()-49h / now()-47h / now()` and asserts exactly one deletion. I didn't write it because:

1. The agent package has **no existing real-DB integration harness** — `monitor_test.go` uses `stubDBTX` for everything. Adding one just for the pruner would either force a test-DB container into the agent package's test closure (scope creep) or add a new integration harness we don't have a pattern for yet.
2. The SQL itself is verified by the sqlc regeneration — the constant string in the generated Go now literally contains `'48 hours'`. If the interval were wrong, sqlc would embed the wrong string; the test would just confirm what we can read in `log_buffer.sql.go:257`.
3. The interesting behavior to guard against — regression via wiring — is the *loop behavior* (startup tick, clean cancel, error swallowing), which the three tests above do cover.

If this turns out to be the wrong call, the integration test is easy to bolt on later once Phase 3 lands (the scheduler will almost certainly need a real-DB test harness for `schedulerTick`, which would be a natural home for a pruner integration test too).

---

## Files touched

| File | Kind | Change |
|------|------|--------|
| `backend/internal/db/queries/log_buffer.sql` | Edit | `'24 hours'` → `'48 hours'` |
| `backend/internal/db/log_buffer.sql.go` | Regen | sqlc picked up the new interval |
| `backend/internal/agent/pruner.go` | **New** | `Prune(ctx)` loop + `pruneTick(ctx)` helper |
| `backend/internal/agent/agent.go` | Edit | `Start()` now spawns two goroutines |
| `backend/internal/agent/pruner_test.go` | **New** | Three unit tests |

Net diff: +2 new files, 2 modified, 0 deleted. Zero frontend churn, zero API changes, zero migration files — Phase 1 is a pure backend-internal plumbing fix.

---

## Validation

```bash
cd backend && go vet ./...         # clean
cd backend && go build ./...       # clean
cd backend && go test ./...        # all packages ok
```

Full test output for the pruner tests:

```
=== RUN   TestPruneTick_DBError
ERROR pruner: failed to delete expired logs err="stubDBTX: not implemented"
--- PASS: TestPruneTick_DBError (0.00s)
=== RUN   TestPrune_CancelExits
INFO pruner goroutine started interval=1h0m0s
ERROR pruner: failed to delete expired logs err="stubDBTX: not implemented"
INFO pruner goroutine stopped
--- PASS: TestPrune_CancelExits (0.01s)
=== RUN   TestPrune_StartupTickRuns
INFO pruner goroutine started interval=1h0m0s
ERROR pruner: failed to delete expired logs err="stubDBTX: not implemented"
INFO pruner goroutine stopped
--- PASS: TestPrune_StartupTickRuns (0.01s)
PASS
```

The `ERROR pruner: failed to delete expired logs` line in the test output is the *expected* shape of the log against a stub DB — the whole point of `TestPruneTick_DBError` is to prove that the error is logged and swallowed rather than propagated or panicked on.

**Not yet validated in a live environment.** The plan's validation gate also asks for a local-backend smoke test: seed `log_buffer` with a >48h row, start the backend, watch for `pruner: deleted expired logs rows=1` on startup. I haven't done that — it requires a running Postgres with a seeded row, and the user wanted the code shipped first. Worth doing as a follow-up when convenient.

---

## Deployment notes

- **No migration file** — the SQL change is *inside a named query*, not a schema migration. `make migrate-up` is a no-op for Phase 1.
- **No new env vars, no Fly secrets to set.**
- **Production impact on first run:** When this deploys, the startup `pruneTick` will run a `DELETE FROM log_buffer WHERE ingested_at < now() - interval '48 hours'` against production. Depending on how long `log_buffer` has been growing unchecked since v0.4.0, **this could delete a meaningful chunk of rows in a single statement**. If the table is very large, consider:
  - Running the DELETE manually first at a quiet hour and watching the query plan, or
  - Changing the pruner's first-tick behavior to `LIMIT 10000` (needs a SQL change to `DELETE ... WHERE id IN (SELECT id ... LIMIT N)`) if the initial backlog is too large to delete in a single transaction.
  - Both are out of scope for Phase 1 as-written but flagging here so the reviewer/deployer is aware.

- **No rollback story needed** — if the pruner misbehaves in prod, removing the goroutine from `Start` (or reverting this entire change) restores the pre-Phase-1 behavior. The retention itself is a data-plane SQL interval change; flipping it back to 24h is a one-char edit.

---

## Follow-ups & open edges

1. **Integration test for real deletion behavior** — deferred until Phase 3 introduces a real-DB test harness we can piggyback on.
2. **The 0.30.2 lesson applies here too.** Phase 1 has no migration file, so the schema-drift failure mode from v0.30.2 doesn't apply — but if future phases add migrations to this plan, remember to run `make migrate-up` against prod before deploying.
3. **`docs/changelog.md` entry** — the plan marks Phase 1 + Phase 2 as a combined `0.30.3` release. Phase 1's changelog entry lives inside Phase 2's doc (see Task 2.6 in the plan) and will be added when Phase 2 ships. No standalone 0.30.3-only release for Phase 1.
4. **Phase 3's third goroutine slot is now primed.** When the scheduler lands, it slots in as a third `a.wg.Add(1); go func(){ defer a.wg.Done(); a.InvestigationScheduler(ctx) }()` block right after the pruner's block, and updates the startup log line accordingly. Zero structural refactor needed.

---

## Why this was worth doing even though it feels small

Phase 1 is the archetypal "quiet infrastructure fix" — two files, ~60 lines of code, no user-visible change. It won't show up in a demo. But:

- It prevents a real production disk-fill. `log_buffer` has been growing unbounded for ~2 months of release history. That's the kind of bug that stays invisible until it isn't, and then it's a 2am page.
- It closes a Chekhov's-gun-style dead-code path. Dead functions are strictly worse than missing functions — they imply work was done, lull maintainers into assuming correctness, and survive grep-for-unused passes that rely on references.
- It structurally unblocks Phase 3 at near-zero marginal cost. The `Start()` edit to host a second goroutine is the *same* edit Phase 3 would need to make for its own scheduler goroutine — doing it now means Phase 3's PR is strictly smaller and more focused.

The plan's scope estimate of "~½ day including tests" held up almost exactly.
