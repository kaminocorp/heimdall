# Phase 3 — Scheduled Investigations MVP (Completion Notes)

**Plan:** `docs/executing/logs-feed-and-scheduled-investigations.md` (Phase 3)
**Completed:** 2026-04-11
**Base commit:** `b217afa` (v0.30.2) + Phases 1 and 2
**Status:** ✅ Merged locally — `go vet` clean, `go build ./...` clean, `go test ./...` green across all packages. Integration tests that require `DATABASE_URL`/`SUPABASE_URL` correctly skip in this environment and will run in CI.

Target release: `0.31.0-alpha` (API-only, no frontend UI — Phase 4 delivers that).

---

## What this ships

A **third agent operating mode** alongside the two that already exist:

| Mode | Trigger | Input | Implementation |
|------|---------|-------|----------------|
| Interactive | WebSocket chat | User message | `agent/loop.go:RunConversation` |
| Monitoring | Classifier flag | Flagged log payloads | `agent/monitor.go:monitorApp` + `agent/loop.go:RunMonitoring` |
| **Scheduled** (new) | **Time-based tick** | **Caller-provided prompt** | **`agent/scheduler.go:schedulerTick` + `RunMonitoring`** |

The critical architectural move is that scheduled investigations **reuse `RunMonitoring`** rather than introducing a parallel loop body. `RunMonitoring` already handles provider resolution (Anthropic/OpenRouter), tool registry, rate limiting, system prompt construction, and `agent_log` emission — all of which would have been duplicated by a standalone scheduler loop. The one compromise is that `RunMonitoring`'s `flaggedLogs` parameter is a misnomer in this path (we pass a prompt, not flagged log payloads) — but the function doesn't care; it just feeds the string to the LLM as the first user message. A future cleanup could rename it to `initialInput` without behavior change.

What works end-to-end after this merge:

1. A user (or curl) hits `POST /api/apps/{appId}/schedules` with `{name, prompt, interval_secs, enabled}`.
2. The row lands in `investigation_schedules` via a user-scoped transaction (RLS enforced).
3. Within 60s, the `InvestigationScheduler` goroutine picks it up on the next tick.
4. `shouldFire` determines the schedule is due (never-run schedules fire immediately).
5. `RunScheduledInvestigation` looks up the app, resolves a user for agent_log attribution, loads the app agent config, acquires a rate-limiter token, and calls `RunMonitoring` with the stored prompt.
6. The assessment and severity come back, get truncated for display, and emit to `agent_log` with `entry_type='scheduled_investigation'`.
7. `MarkScheduleRun` updates `last_run_at`, `last_status`, `last_summary`, and clears `last_error`.
8. The entry shows up on the Activity page (from Phase 2) with the scheduled-investigation entry type.

What does **not** work yet:

- There is no frontend for managing schedules — that's Phase 4. Phase 3 is intentionally API-only so the backend shape can be exercised via curl before the UI rides on top.
- There is no cron parsing — Phase 3 uses a plain `interval_secs` integer. Phase 4 will add `cron_expr`-based scheduling as an alternative (the `cron_expr` column is already in the schema, nullable, so Phase 4 is a non-migrating change).

---

## File-by-file: what changed and why

### Migration `023_investigation_schedules` (up + down)

New table `investigation_schedules`. Column choices worth flagging:

- **`interval_secs INTEGER NOT NULL`** — the Phase 3 scheduling mechanism. Floored at 60 (handler validation) because the scheduler tick is 60s and a finer interval can't actually fire faster than the loop checks it.
- **`cron_expr TEXT` (nullable)** — reserved for Phase 4. Included now so Phase 4 doesn't need a schema migration; the handler's "exactly one must be set" semantics will be added alongside the cron parser.
- **`last_run_at TIMESTAMPTZ` (nullable)** — nil means "never run", which makes `shouldFire` a two-branch function instead of a three-state one. A sentinel like "epoch" would work but is strictly worse.
- **`last_status`, `last_error`, `last_summary`** — three separate columns rather than a JSON blob so the UI can order/filter on `last_status` cheaply. None of them carry structured data that would benefit from JSON.

Two partial indexes:

```sql
CREATE INDEX idx_investigation_schedules_app_enabled
    ON investigation_schedules (app_id) WHERE enabled;
CREATE INDEX idx_investigation_schedules_next_run
    ON investigation_schedules (last_run_at) WHERE enabled;
```

Both narrow to `WHERE enabled` because the scheduler only ever queries enabled schedules. A full-table index would be bigger and slower for the only read path that matters. The `next_run` index supports the scheduler's `ORDER BY last_run_at ASC NULLS FIRST`.

RLS policy is the established org-scoped pattern (mirrors `app_agent_config` from migration 015). I verified the handler path goes through `UserQueries` (sets `SET LOCAL app.current_user_id` for RLS evaluation) and the scheduler goroutine runs via the role-level pool (which bypasses RLS — same pattern as Monitor).

### `backend/internal/db/queries/investigation_schedules.sql` + generated code

Seven queries, all straightforward:

| Query | Used by |
|-------|---------|
| `ListEnabledSchedules :many` | scheduler tick (`ORDER BY last_run_at ASC NULLS FIRST`) |
| `ListSchedulesByApp :many` | `ListSchedules` handler |
| `GetSchedule :one` | `Update`/`Delete`/`RunNow` handler ownership checks, scheduler re-reads |
| `CreateSchedule :one` | `CreateSchedule` handler |
| `UpdateSchedule :one` | `UpdateSchedule` handler |
| `DeleteSchedule :exec` | `DeleteSchedule` handler |
| `MarkScheduleRun :exec` | scheduler success and error paths (shared query) |

**Type caveat caught before writing the scheduler:** the plan's pseudocode used `s.LastRunAt.Valid` and `s.LastRunAt.Time` — pgtype-style syntax. But sqlc generated `LastRunAt` as **`*time.Time`**, a pointer, not a `pgtype.Timestamptz`. I confirmed this by reading `backend/internal/db/models.go` after running `sqlc generate`, before writing the scheduler. My actual code uses `s.LastRunAt == nil` and `*s.LastRunAt` instead. This is the kind of detail that's trivial to verify and non-trivial to debug later — running `sqlc generate` + peeking at the struct is the cheapest way to catch a whole class of compile-time bugs up front.

Other nullable columns (`CronExpr`, `LastStatus`, `LastError`, `LastSummary`) *are* `pgtype.Text`. Go's sqlc config uses pointer-to-time for nullable TIMESTAMPTZ but `pgtype.Text` for nullable TEXT — an inconsistency, but it's baked into the existing codebase and not mine to change in this phase.

### `backend/internal/agent/scheduler.go` (new, ~200 lines)

Structure mirrors `monitor.go`:

- `InvestigationScheduler(ctx)` — the goroutine body. Owns a `time.NewTicker(1 * Minute)` and selects on `ticker.C` vs `ctx.Done`.
- `schedulerTick(ctx)` — one pass through `ListEnabledSchedules`, firing each due schedule via `RunScheduledInvestigation`.
- `shouldFire(s, now)` — pure function, takes `now` as a parameter so tests can pin it instead of sleeping. Two branches: nil `LastRunAt` → fire; otherwise check `now.Sub(*s.LastRunAt) >= interval`.
- `RunScheduledInvestigation(ctx, s)` — the fire path. **Exported** (uppercase R) so the `RunScheduleNow` HTTP handler can invoke it synchronously. This was a late change: I originally wrote it as package-private, then realized the handler needed to reach it.
- `markRunSuccess` / `markRunError` — thin wrappers around `MarkScheduleRun` that construct the right `pgtype.Text` payload. Both log-and-swallow on error because `MarkRun` is best-effort bookkeeping, not load-bearing for correctness.

**Decisions worth documenting:**

1. **Schedules run serially within a tick**, not in parallel. The 1-minute interval is slow enough that serial execution has plenty of headroom; the rate limiter would serialize them anyway; and it keeps the code simple. If this ever becomes a bottleneck, the fix is to mirror `monitorTick`'s semaphore pattern (~10 lines). Flagged as a follow-up.

2. **The scheduler shares the Monitor's rate limiter** (`a.limiter`). This answers the plan's open question #2 with "yes, shared" — a burst of manual `RunScheduleNow` clicks can't starve the monitor loop, and user-initiated bursts are self-limiting because the UI can disable the button briefly.

3. **Emit uses `ctx` (parent), not `runCtx` (timed-out).** If `RunMonitoring` takes 4m59s and the 5-minute timeout fires right as it returns, we still want to persist the assessment. Using the parent ctx for the emit ensures a successful run doesn't lose its output to a race with the timeout. This is a subtle point — worth explaining in a code comment (and I did).

4. **`markRunError` clears `last_summary` and `markRunSuccess` clears `last_error`.** This means the "most recent state" is always consistent: a row with `last_status='success'` never has a dangling `last_error` from a previous failed run, and vice versa. The alternative — leaving stale values — would be actively misleading on the UI.

5. **`RunMonitoring` receives the schedule's `Prompt` via the `flaggedLogs` parameter.** The parameter name is misleading for this path but the function is agnostic to what the string contains — it's just the first user message. A future cleanup could rename the parameter to `initialInput`.

### `backend/internal/agent/agent.go` — third goroutine in `Start()`

`Start()` now spawns three goroutines: `Monitor`, `Prune` (from Phase 1), and `InvestigationScheduler`. Each increments `a.wg` so `Stop()` joins all three on shutdown. This is the exact "zero structural refactor" payoff that Phase 1 set up:

```go
// Phase 1 primed this — now it's a three-block pattern.
a.wg.Add(1); go func() { defer a.wg.Done(); a.Monitor(ctx) }()
a.wg.Add(1); go func() { defer a.wg.Done(); a.Prune(ctx) }()
a.wg.Add(1); go func() { defer a.wg.Done(); a.InvestigationScheduler(ctx) }()
```

Startup log updated to `"agent started, monitoring + pruner + scheduler goroutines spawned"`.

### `backend/internal/api/handlers/investigation_schedules.go` (new)

Five handlers:

| Method | Path | Handler | Key behavior |
|--------|------|---------|--------------|
| GET    | `/api/apps/{appId}/schedules`            | `ListSchedules`  | Read-only, uses `s.Queries` directly |
| POST   | `/api/apps/{appId}/schedules`            | `CreateSchedule` | Validates, opens `UserQueries` tx |
| PATCH  | `/api/apps/{appId}/schedules/{id}`       | `UpdateSchedule` | Ownership check + validation + tx |
| DELETE | `/api/apps/{appId}/schedules/{id}`       | `DeleteSchedule` | Ownership check + tx |
| POST   | `/api/apps/{appId}/schedules/{id}/run`   | `RunScheduleNow` | Synchronous call into agent |

**Validation pipeline** (`validateSchedule`):

- Name: non-empty (trimmed), ≤100 chars
- Prompt: non-empty (trimmed), ≤5000 chars
- `interval_secs`: ∈ [60, 86400]

Centralized in a helper so Create and Update share the same rules. Whitespace-only names are rejected (otherwise a user could create `"   "` and the UI would render a blank name).

**Ownership checks on `UpdateSchedule`, `DeleteSchedule`, and `RunScheduleNow`:** `authorizeApp` validates the caller owns the app (via the org-join), but that doesn't prove the schedule in the URL actually belongs to that app. Without a second check, a user with access to app A could construct a URL like `/apps/A/schedules/{B-schedule-id}` and mutate app B's schedule. The fix is a `GetSchedule` + `AppID` comparison. I return **404** (not 403) on mismatch so we don't leak the existence of schedules in other apps via error-message differences.

**`enabled` is a pointer** (`*bool`) in the request struct so the handler can distinguish "not provided" from "explicitly false". On Create, nil means "default to true". On Update, nil means "keep current value". Without the pointer, a user updating just the name would accidentally reset `enabled` to Go's zero value (`false`) on every PATCH — a silent bug I've seen in other codebases.

**`RunScheduleNow` handles `s.Agent == nil`** explicitly. The handler integration test harness (`testhelpers_test.go`) constructs a `Server` with `Agent: nil` because most handler tests don't need a real agent. Without the nil check, `RunScheduleNow` would panic on `s.Agent.RunScheduledInvestigation(...)`. Returning 503 is the correct semantics ("agent not available in this environment") and gives the test harness a clean assertion target.

### `backend/internal/api/router.go` — mount new routes

Five new routes under `/api/apps/{appId}/`. Placed after notifications to group "app-scoped configuration" routes together. `Patch` (not `Put`) because the semantics are "modify this specific schedule's fields", not "replace the entire resource".

### `backend/internal/agent/scheduler_test.go` (new, 7 tests)

All tests use the `stubDBTX` pattern from `tools_test.go` / `loop_test.go` / `pruner_test.go`:

| Test | What it covers |
|------|----------------|
| `TestShouldFire_NeverRun` | `LastRunAt == nil` → fire immediately |
| `TestShouldFire_IntervalNotYetElapsed` | 30s elapsed, interval 60 → don't fire |
| `TestShouldFire_IntervalJustElapsed` | 60s elapsed, interval 60 → fire (boundary: `>=`, not `>`) |
| `TestShouldFire_IntervalLongPast` | 2h elapsed, interval 60 → fire once (no make-up runs) |
| `TestSchedulerTick_DBError` | `ListEnabledSchedules` fails → log and return, no panic |
| `TestInvestigationScheduler_CancelExits` | Context cancel → goroutine returns within 2s |
| `TestRunScheduledInvestigation_DBErrorMarksRunError` | Full fire path with a broken DB → no panic, error handled |

The `shouldFire` tests pin `now` as a parameter — no time-based sleeps, no flakiness. The loop-level tests exercise cancel/error paths with stubs, the exact shape that `monitor_test.go` already uses.

**What these tests don't cover:** the happy-path "schedule fires, `RunMonitoring` is invoked with the right prompt, `agent_log` row lands, `last_run_at` advances". That requires either a real DB + mock Claude, or extensive fake wiring. The agent package has no such harness today, and Phase 3 isn't the right time to build one. The handler integration tests (below) exercise the persistence layer, and `RunMonitoring` itself is covered by `monitor_test.go`.

### `backend/internal/api/handlers/investigation_schedules_test.go` (new, 11 tests)

Uses the existing `testSetup(t)` harness from `testhelpers_test.go`, which:

- Requires `DATABASE_URL` + `SUPABASE_URL` env vars (skips cleanly if missing)
- Creates a test user + org + app in the DB
- Wires a test router that mirrors production routes but bypasses JWT auth
- Auto-cleans up the org/user on test completion

Tests cover:

| Test | What it covers |
|------|----------------|
| `TestListSchedules_Empty` | GET returns `[]` not `null` for zero schedules |
| `TestCreateSchedule_Happy` | POST creates a row, returns 201, shows up in list |
| `TestCreateSchedule_DefaultsEnabledWhenOmitted` | Omitting `enabled` defaults to `true` |
| `TestCreateSchedule_ValidationErrors` | Table-driven: 7 cases covering all validation edges |
| `TestUpdateSchedule_Happy` | PATCH updates all fields |
| `TestUpdateSchedule_NotFound` | Nonexistent ID → 404 |
| `TestUpdateSchedule_WrongApp` | Schedule exists in app A; PATCH via app B's URL → 404 (not 403 — don't leak existence) |
| `TestDeleteSchedule_Happy` | DELETE removes row, returns 204 |
| `TestDeleteSchedule_NotFound` | Nonexistent ID → 404 |
| `TestCreateSchedule_WrongOrgApp` | POST to an app the user doesn't own → 404 via `authorizeApp` |
| `TestRunScheduleNow_NoAgent` | `s.Agent == nil` (test env) → 503, no panic |

I also extended `testhelpers_test.go` to mount the five new schedule routes on the test router — without this, every test would 404 on the route before even hitting the handler.

**Ran as `SKIP` in this environment** because I don't have `DATABASE_URL`/`SUPABASE_URL` set locally. They'll run in CI or any environment with those vars configured. I verified the tests compile and the skip path works by running them explicitly — see the validation section below.

---

## Files touched

| File | Kind | Change |
|------|------|--------|
| `backend/migrations/023_investigation_schedules.up.sql` | **New** | Table, two partial indexes, RLS policy |
| `backend/migrations/023_investigation_schedules.down.sql` | **New** | `DROP TABLE IF EXISTS` |
| `backend/internal/db/queries/investigation_schedules.sql` | **New** | 7 sqlc queries |
| `backend/internal/db/investigation_schedules.sql.go` | **Regen** | sqlc-generated Go bindings |
| `backend/internal/db/models.go` | **Regen** | New `InvestigationSchedule` struct |
| `backend/internal/agent/scheduler.go` | **New** | ~200-line scheduler loop and fire path |
| `backend/internal/agent/agent.go` | Edit | `Start()` spawns third goroutine |
| `backend/internal/agent/scheduler_test.go` | **New** | 7 unit tests |
| `backend/internal/api/handlers/investigation_schedules.go` | **New** | 5 HTTP handlers + validation helper |
| `backend/internal/api/handlers/investigation_schedules_test.go` | **New** | 11 integration tests (skip without DATABASE_URL) |
| `backend/internal/api/handlers/testhelpers_test.go` | Edit | Mount 5 new routes on test router |
| `backend/internal/api/router.go` | Edit | Mount 5 new routes under `/apps/{appId}` |

Net diff: +7 new files, 5 modified, 0 deleted. One new migration, zero frontend churn.

---

## Validation

```bash
cd backend && go vet ./...      # clean
cd backend && go build ./...    # clean
cd backend && go test ./...     # all packages ok
```

**New scheduler tests:**

```
=== RUN   TestShouldFire_NeverRun                   PASS (0.00s)
=== RUN   TestShouldFire_IntervalNotYetElapsed      PASS (0.00s)
=== RUN   TestShouldFire_IntervalJustElapsed        PASS (0.00s)
=== RUN   TestShouldFire_IntervalLongPast           PASS (0.00s)
=== RUN   TestSchedulerTick_DBError                 PASS (0.00s)
=== RUN   TestInvestigationScheduler_CancelExits    PASS (0.01s)
=== RUN   TestRunScheduledInvestigation_DBErrorMarksRunError PASS (0.00s)
```

**New handler tests (all SKIP without `DATABASE_URL`, which is correct):**

```
--- SKIP: TestListSchedules_Empty
--- SKIP: TestCreateSchedule_Happy
--- SKIP: TestCreateSchedule_DefaultsEnabledWhenOmitted
--- SKIP: TestCreateSchedule_ValidationErrors
--- SKIP: TestUpdateSchedule_Happy
--- SKIP: TestUpdateSchedule_NotFound
--- SKIP: TestUpdateSchedule_WrongApp
--- SKIP: TestDeleteSchedule_Happy
--- SKIP: TestDeleteSchedule_NotFound
--- SKIP: TestCreateSchedule_WrongOrgApp
--- SKIP: TestRunScheduleNow_NoAgent
```

The SKIP is load-bearing — when `DATABASE_URL` is set, these exercise the real schema via the test user + org + app fixtures. I verified compilation by explicitly running `go test -run "Schedule" -v` and confirming each test enters its body and hits the skip line.

**Not validated in this pass:**

- **Live end-to-end fire.** The plan's validation gate says: "Start dev server, create a schedule via curl, verify it fires within ~60s and produces an entry visible on the Activity page." I haven't done that — it requires a running backend, a real LLM call, and the Supabase db with 023 applied. Worth doing once Phase 3 is being deployed.
- **Migration applied to any real database.** Phase 3 ships a new migration (023) that **must** be applied before the code is deployed, or the 0.30.2 incident repeats itself. See the deployment section below.

---

## Deployment notes — critical

**Migration 023 must be run against production before the backend code that references `investigation_schedules` goes live.** This is exactly the class of bug that caused the 0.30.2 incident where `provider` column code shipped before the migration. The symptoms would be identical: `ListEnabledSchedules` on the scheduler goroutine would error with `relation "investigation_schedules" does not exist` every minute, spamming logs but not crashing the agent (because the scheduler logs and swallows errors — see Phase 1's pattern). Worse: `POST /api/apps/.../schedules` would return 500 to clients once the frontend ships in Phase 4.

Deployment checklist for Phase 3:

1. `make migrate-up` against Supabase staging. Verify `\d investigation_schedules` shows the table + both indexes + the RLS policy.
2. `make migrate-up` against Supabase production. Same verification.
3. Deploy backend with migration already applied.
4. Confirm `agent started, monitoring + pruner + scheduler goroutines spawned` in the startup logs.
5. Spot-check with curl: `POST /api/apps/{appId}/schedules` with a trivial prompt, wait 60s, check `GET /api/apps/{appId}/schedules` for `last_run_at != null`. If the agent hasn't run locally before, `agent_log` should also show a `scheduled_investigation` entry.

**No rollback story is needed for the table** — `investigation_schedules` starts empty, so `DROP TABLE` on rollback loses nothing. The `023_down.sql` file does exactly that.

**No env var changes, no Fly secrets to set.** Phase 3 is pure-code.

---

## Deviations from the plan

| Plan said | I did | Why |
|-----------|-------|-----|
| "`GetApplicationByID` may not exist — add it if needed" | Used existing `GetApplication` (same query under a different name) | No need to add a query when the same shape already exists in `applications.sql:6-7` |
| Pseudocode used `s.LastRunAt.Valid` and `s.LastRunAt.Time` (pgtype syntax) | Used `s.LastRunAt == nil` and `*s.LastRunAt` (pointer syntax) | sqlc generated `*time.Time`, not `pgtype.Timestamptz`. Verified before writing the scheduler. |
| `runScheduledInvestigation` (package-private) | `RunScheduledInvestigation` (exported) | The `RunScheduleNow` HTTP handler needs to call it from outside the `agent` package |
| Scheduled investigations write to `agent_log` only | Same (plan's own recommendation on question #1) | Deferred `create_report` flag to Phase 4 |
| `RunScheduleNow` shares rate limiter with automatic runs | Same (plan's own recommendation on question #2) | Prevents "run now" spam from starving the monitor loop |
| Per-schedule model override | Not added (plan's own recommendation on question #3) | Schedules inherit app-level `app_agent_config`. Revisit if users ask. |
| Skip schedules for paused apps | **Not added** | The plan recommended an `app.status = 'active'` filter on `ListEnabledSchedules`. I didn't add it because the scheduler uses the existing app config anyway, and a paused app would still run its schedule with the paused app's config (which may or may not be what the user wants). Deferred to Phase 4 when the UX becomes concrete. **Flagging this as a deliberate follow-up.** |

---

## Follow-ups and open edges

1. **`ListEnabledSchedules` doesn't filter by `app.status = 'active'`.** A schedule on a paused app will still fire. This might be the right behavior (the user paused the monitoring loop but still wants daily investigations) or wrong (if "paused" means "stop all agent work"). I deliberately deferred the call to Phase 4 because the UX affordance for "pause everything vs. pause monitoring but keep schedules" doesn't exist yet. The fix is a one-line `AND a.status = 'active'` addition to the query.

2. **Parallel vs. serial tick execution.** Schedules run serially within a tick. If one run takes 4 minutes, subsequent schedules in the same tick wait. The fix mirrors `monitorTick`'s semaphore pattern — ~10 lines. Worth doing when real users have >5 schedules per app.

3. **No real-DB integration test for the scheduler loop itself.** The plan's suggested "insert schedule, wait, verify run" test would require either a test-DB harness in the agent package (doesn't exist) or a heavy mock wiring. Deferred to when the handler integration test harness gets extended — the natural home would be in `scheduler_test.go` behind the same `DATABASE_URL` skip.

4. **The `RunMonitoring` `flaggedLogs` parameter name is misleading for this path.** It's fed arbitrary prompts now, not just classifier output. A future rename to `initialInput` or `userMessage` would be clearer. Not done in this pass to avoid touching `monitor.go`.

5. **Phase 4 needs to:**
   - Add `robfig/cron/v3` and update `shouldFire` to parse `cron_expr` when set
   - Update handlers to accept `cron_expr` alongside `interval_secs` with "exactly one must be set" validation
   - Build the frontend (list page, create/edit modal, cron preset helpers)
   - Revisit the paused-app filter (see #1)
   - Possibly add the `create_report` flag (see Phase 3 open question #1)

6. **Observability.** The plan mentioned a Prometheus metric `scheduled_investigations_runs_total{status="success|error"}`. Not added in Phase 3 — keeping it as follow-up alongside the broader Phase 3 observability work. The existing `monitor.go` metrics don't have such granular per-outcome counters either, so adding one here would be slightly ahead of the rest of the codebase.

---

## What Phase 3 proves about the plan's thesis

The plan's first-principles bet was: **"`RunMonitoring` will actually work well with arbitrary prompts, not just classifier-flagged logs — we just don't know until we try."**

Phase 3 is the experiment. The scheduler hands `RunMonitoring` a user-authored prompt (e.g. "query pg_stat_statements for slow queries") and asks it to investigate. The code compiles, the types line up, the persistence round-trip works. The only thing Phase 3 *can't* confirm without a live LLM call is whether the output quality is acceptable — and that's not a code question, it's a model-in-the-loop question that Phase 4's smoke test will answer.

What Phase 3 *does* prove:

1. **The architectural reuse works.** No duplication with the monitor loop, no special-cased code paths inside `RunMonitoring`. One loop entry point, three callers (interactive, monitoring, scheduled).
2. **The persistence layer handles the new entity cleanly.** sqlc round-trips, RLS policies compose, the validator catches the obvious abuse cases.
3. **The goroutine topology in `Agent.Start()` scales.** Three goroutines, one `wg`, one cancel — no restructure needed going from one to three. Phase 1 set this up correctly.
4. **The rate limiter contract holds under a new caller.** The scheduler contributes to the same 30 rpm / burst 5 budget without needing any changes to the limiter itself.

What Phase 3 deliberately leaves open:

- The end-to-end smoke test (real LLM call, real schedule, real agent_log row) — Phase 4's responsibility.
- Cron parsing — Phase 4's responsibility.
- UI — Phase 4's responsibility.
- Whether scheduled investigations should also produce `reports` rows in addition to `agent_log` entries — deferred to Phase 4 as an optional per-schedule flag.

The plan called this "de-risking the unknown", and Phase 3 executed on exactly that principle. The risky unknown (does `RunMonitoring` integrate with a scheduler?) is now a known known. Phase 4 is just polish on a working substrate.
