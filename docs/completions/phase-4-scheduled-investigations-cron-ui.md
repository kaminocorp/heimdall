# Phase 4 — Scheduled Investigations: Cron Parsing + Frontend UI (Completion Notes)

**Plan:** `docs/executing/logs-feed-and-scheduled-investigations.md` (Phase 4)
**Completed:** 2026-04-11
**Base commit:** `b217afa` (v0.30.2) + Phases 1, 2, and 3
**Status:** ✅ Merged locally. Backend: `go vet ./...` clean, `go build ./...` clean, `go test ./...` green across all packages. Frontend: `vue-tsc -b --noEmit` clean, `npm run build` clean (1.06s), `vitest run` 29/29 pass, `eslint` clean on all Phase 4 files. Integration tests requiring `DATABASE_URL`/`SUPABASE_URL` correctly skip in this environment.

Target release: `0.31.0` (full release — Phase 3 shipped as `0.31.0-alpha` API-only on the same day).

---

## What this ships

Phase 4 turns the Phase 3 MVP into a real feature:

1. **Real cron parsing in the scheduler** — the Phase 3 integer-interval path is still there as a fallback, but `cron_expr` is the primary scheduling mechanism going forward.
2. **A frontend page for managing schedules** — `/schedules`, full CRUD, lives in the Agent section of the sidebar.
3. **An exactly-one-of validation contract** at the API boundary — clients send either `interval_secs` or `cron_expr`, not both, with a 400 on either violation.
4. **Backward compatibility for Phase 3 schedules** — the modal has an Interval mode specifically so Phase 3 rows can be edited without being force-migrated to cron.

The architectural bet from Phase 3 — "reuse `RunMonitoring` rather than introducing a parallel loop body" — still holds. Phase 4 adds nothing new to that loop; all the work is around it.

---

## Backend changes

### `github.com/robfig/cron/v3` dependency

Added via `go get github.com/robfig/cron/v3@latest` — resolved to `v3.0.1`. This is the de-facto Go cron library:

- Two years of stable releases
- Zero external dependencies
- Correct DST handling (not something we need to test ourselves)
- 12k+ GitHub stars, widely used in Go ecosystem

Alternative: `gorhill/cronexpr` was considered but hasn't seen a release in years. Not worth the risk for a library that's definitionally load-bearing to this feature.

### `scheduler.go` — cron path

The key addition is a package-level `cronParser`:

```go
var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
```

Five-field (no seconds) matches what users expect from `crontab` and what the frontend's `humanizeCron` renders. Importantly, this parser rejects six-field expressions with a clear error — so `"* * * * * *"` gets a 400 instead of silently being accepted with unexpected semantics.

The parser is constructed once at package init because `cron.Parser` is stateless and reusable. The plan suggested "construct per-call for simplicity" but I went with the shared instance because:

1. `shouldFire` runs every minute against every enabled schedule — allocating a parser on every call is unnecessary garbage
2. Package-level `var` is already the convention for similar stateless helpers elsewhere in the codebase
3. Testability isn't affected — the parser has no mutable state to pollute between tests

`ParseCronExpression` is an exported wrapper so the HTTP handler can validate at the API boundary without importing `robfig/cron/v3` directly. This keeps the cron library as a single point of entry from the `agent` package — if we ever swap parsers (e.g. to support 6-field expressions), there's exactly one import site to change.

`shouldFire` now has three branches:

```go
if s.LastRunAt == nil { return true }  // never run
if s.CronExpr.Valid && s.CronExpr.String != "" {
    // cron path: parse, compute next, compare to now
}
// legacy interval path: unchanged from Phase 3
```

**Decision worth flagging:** on a malformed `cron_expr`, `shouldFire` logs the error and returns `false`. The alternative — returning `true` or panicking — is worse:

- `true` → infinite tight loop firing the same broken expression every minute
- `panic` → takes down the whole agent because the scheduler goroutine shares fate with `Start()`'s `wg`

Returning `false` creates a "stuck" schedule that a user can fix by editing the row via the UI. The error log is visible in server logs so operators can notice it. This is the same defensive pattern used in `monitor.go` for classifier errors.

### `investigation_schedules.sql` — parameter added, regenerated

Both `CreateSchedule` and `UpdateSchedule` now take `cron_expr` alongside `interval_secs`. sqlc regenerated cleanly — the generated param struct is:

```go
type CreateScheduleParams struct {
    AppID        uuid.UUID
    Name         string
    Prompt       string
    IntervalSecs int32
    CronExpr     pgtype.Text  // ← new in Phase 4
    Enabled      bool
}
```

`pgtype.Text` (not `*string`) because the `cron_expr` column is `TEXT NULL` and sqlc's existing config for nullable TEXT columns is `pgtype.Text` — same as `LastStatus`, `LastError`, `LastSummary`. Consistency with the codebase beats ergonomics.

**No migration.** The `cron_expr` column has been in migration 023 since Phase 3. This is the payoff for the Phase 3 decision to ship the column nullable even though it wasn't used yet — Phase 4 is a pure-code change, so the class of bugs that caused v0.30.2 (code references a column that isn't in prod) can't happen here.

### `investigation_schedules.go` (handler) — exactly-one-of validation

The `scheduleRequest` struct gained `CronExpr string \`json:"cron_expr,omitempty"\`` and the validation pipeline now enforces:

1. Name: non-empty (trimmed), ≤100 chars
2. Prompt: non-empty (trimmed), ≤5000 chars
3. **Exactly one of** `IntervalSecs > 0` OR `CronExpr != ""` — both or neither is a 400
4. If interval: `60 ≤ IntervalSecs ≤ 86400`
5. If cron: `len(CronExpr) ≤ 200` AND `agent.ParseCronExpression(CronExpr)` succeeds

The cron is parsed at the API boundary so invalid expressions never reach the database. The scheduler also handles parse failures defensively (see above) but catching them here gives users a crisp 400 with the parser's own error text (`"expected 5 fields, found 3"`) instead of a silent stuck schedule.

A new helper `toDBScheduleParams(req *scheduleRequest) (int32, pgtype.Text)` packs a validated request into the sqlc params shape. When `CronExpr` is set we return `(0, Valid Text)`; otherwise `(IntervalSecs, Invalid Text)`. This makes the stored row unambiguous — even though `shouldFire` is designed to tolerate both being set (cron wins), we don't want ambiguous rows in the database.

### Handler tests

Four new validation cases added to `TestCreateSchedule_ValidationErrors`:

| Case | Expected error |
|------|----------------|
| Both interval_secs and cron_expr | "provide either interval_secs or cron_expr, not both" |
| Neither interval_secs nor cron_expr | "provide either interval_secs or cron_expr" |
| Invalid cron expression | "invalid cron_expr: expected exactly 5 fields, ..." |
| Cron expression too long (>200 chars) | "cron_expr exceeds maximum length" |

Two new happy-path tests:

- **`TestCreateSchedule_CronMode`** — POST with `cron_expr`, verify `interval_secs: 0` stored as sentinel
- **`TestUpdateSchedule_IntervalToCron`** — PATCH a Phase-3 interval schedule into cron mode, verify the mode flip cleanly zeros `interval_secs`

The mode-flip test is the important one — it guards against a class of bug where `toDBScheduleParams` might accidentally retain a stale interval when the user switches to cron mode.

### Scheduler tests

Six new unit tests in `scheduler_test.go`, all using the `cronText()` helper I added to keep the fixture tables readable:

| Test | What it locks in |
|------|------------------|
| `TestShouldFire_CronDue` | Past-next-slot fires |
| `TestShouldFire_CronNotYetDue` | Future-next-slot doesn't |
| `TestShouldFire_CronPrefersCronOverInterval` | **Precedence tripwire** — cron wins when both fields are set |
| `TestShouldFire_CronInvalidExpression` | Malformed expressions return false without panic |
| `TestShouldFire_CronNeverRun` | Never-run short-circuits before touching the parser |
| `TestParseCronExpression_Valid` | Exported wrapper sanity check |

The precedence test is the most load-bearing: it locks in the documented "cron wins when set" rule in `scheduler.go:shouldFire`. Without an explicit test, a future refactor could silently invert the precedence and neither the interval path nor the cron path would visibly fail in isolation — both branches work correctly on their own inputs. The test constructs a row where interval says "don't fire" and cron says "do fire", and asserts the cron verdict wins.

---

## Frontend changes

### New files

| File | Purpose |
|------|---------|
| `frontend/src/types/schedule.ts` | `InvestigationSchedule` + `ScheduleInput` types |
| `frontend/src/api/schedules.ts` | Axios client wrapping 5 REST endpoints |
| `frontend/src/stores/schedules.ts` | Pinia setup store with `runningId` tracking |
| `frontend/src/utils/cron-presets.ts` | Preset catalogue + `humanizeCron` + `formatInterval` + `formatSchedule` |
| `frontend/src/pages/SchedulesPage.vue` | Page-level list + modal wiring |
| `frontend/src/components/schedules/ScheduleCard.vue` | Per-schedule row card |
| `frontend/src/components/schedules/ScheduleModal.vue` | Create/edit modal with 3 scheduling modes |
| `frontend/src/stores/__tests__/schedules.test.ts` | 6 vitest store tests |

### Modified files

| File | Change |
|------|--------|
| `frontend/src/router/index.ts` | Added `/schedules` → `SchedulesPage` (lazy-loaded) |
| `frontend/src/components/common/AppSidebar.vue` | Added "Schedules" nav entry in Agent section |
| `frontend/src/test/setup.ts` | Added `patch: vi.fn()` to global axios mock |

### Design decisions

#### 1. App-scoped via `useAppStore().currentAppId`, not URL param

The plan said `/schedules`. I considered `/apps/{appId}/schedules` to match the REST URL shape, but the codebase convention for app-scoped pages (set by `AgentConfigPage.vue`, `ActivityPage.vue`, etc.) is to use `useAppStore().currentAppId` instead of threading the app ID through the URL. Following the convention means the sidebar app-switcher dropdown "just works" — you don't need to navigate to a new URL, the page reacts via `watch`. I added the same `watch(() => appStore.currentAppId, ...)` pattern used by `AgentConfigPage.vue` so switching apps mid-session clears the modal and refetches.

#### 2. Three scheduling modes in the modal (Preset + Custom cron + Interval)

The plan specified two modes (Preset cron buttons + Custom cron textbox). I added a third **Interval** mode because:

1. Phase 3 shipped schedules with `interval_secs` only (no `cron_expr`). These rows exist in every environment where Phase 3 ran.
2. When a user edits one of those rows, the modal needs to render the current schedule somehow. Forcing it into a cron mode would require converting `interval_secs` → a cron expression at load time, which is lossy and surprising.
3. Keeping an explicit Interval mode means Phase 3 schedules round-trip cleanly: load as Interval, save as Interval, nothing changes unless the user explicitly flips the mode toggle.

The `seedForm()` function in `ScheduleModal.vue` inspects the incoming schedule and picks the right initial mode: cron → try to match a preset (fall back to Custom if no match), no cron → Interval.

#### 3. `runningId` for per-row button state, not a global spinner

The store exposes `runningId: string | null` — the ID of whatever schedule is mid-`runNow` call. The card uses this to disable its own Run-now button without freezing the rest of the list. A global `loading` ref would have been simpler but would block the user from clicking Edit on a different row while one is running — poor UX for what is actually an "ad-hoc fire" operation.

This pattern is lifted from `stores/connections.ts`'s `testingId` which does the same thing for connection tests.

#### 4. No `cronstrue` dependency for humanizing cron expressions

The plan suggested either writing a local humanizer or pulling in `cronstrue`. I went with local:

```ts
export function humanizeCron(expr: string): string {
  const preset = cronPresets.find(p => p.value === expr)
  if (preset) return preset.label
  return expr
}
```

Reasoning: `cronstrue` is ~20 KB for the privilege of rendering five presets. That's a lot of bundle weight for a feature whose "Custom cron" mode already presents the user with raw cron text — the user typing `*/7 * * * *` has already opted into raw cron syntax, so rendering it verbatim on the card is the correct affordance, not a regression. If users later ask for readable rendering of arbitrary expressions, bringing in `cronstrue` is a one-line change.

#### 5. `MockAxiosResponse<T>` type alias in the store tests

The existing store tests (`connections.test.ts`, `logs.test.ts`, `auth.test.ts`) all use `as any` casts on the mock return values. I broke from that pattern here because lint flags `@typescript-eslint/no-explicit-any` on new code:

```ts
type MockAxiosResponse<T> = { data: T }
vi.mocked(client.get).mockResolvedValueOnce({ data: [fixture()] } as MockAxiosResponse<InvestigationSchedule[]>)
```

Minimal overhead (one type alias at the top of the test file), keeps the tests type-safe, and gives the rest of the codebase a pattern to copy if its lint debt is ever addressed. It's not worth porting the existing stores to use this type — that would be scope creep.

#### 6. `extractApiError` helper duplicated between `SchedulesPage.vue` and `stores/schedules.ts`

Both files need to walk axios error shapes to pull out `e.response?.data?.error`. I considered factoring it into a shared `utils/api-errors.ts` but kept two copies because:

- Two callers is exactly at the "rule of three" threshold
- The two callers have slightly different fallback semantics (store returns a default message, page returns the axios error's own `message` as an intermediate fallback)
- Building a premature abstraction adds file churn for no gain

If a third duplicate ever appears, that's when lifting it is worth doing. Flagged as a conscious decision, not an oversight.

#### 7. Added `patch: vi.fn()` to global axios mock in `src/test/setup.ts`

The schedules API is the first in the codebase to use PATCH (`updateSchedule`). The shared mock in `setup.ts` had `get/post/put/delete` but not `patch`. I added `patch` to the global mock rather than mocking it per-test because:

- Future PATCH users inherit it for free
- No existing tests rely on `patch` being undefined
- One-line change, very low blast radius

---

## Files touched

| File | Kind | Change |
|------|------|--------|
| `backend/go.mod`, `backend/go.sum` | Edit | Added `github.com/robfig/cron/v3@v3.0.1` |
| `backend/internal/agent/scheduler.go` | Edit | Cron parser + `shouldFire` cron branch + `ParseCronExpression` wrapper |
| `backend/internal/agent/scheduler_test.go` | Edit | +6 cron tests, `cronText` helper |
| `backend/internal/db/queries/investigation_schedules.sql` | Edit | Create/Update take `cron_expr` |
| `backend/internal/db/investigation_schedules.sql.go` | **Regen** | sqlc regenerated (new `CronExpr` field in params) |
| `backend/internal/api/handlers/investigation_schedules.go` | Edit | `cron_expr` validation + `toDBScheduleParams` |
| `backend/internal/api/handlers/investigation_schedules_test.go` | Edit | +6 test cases (4 validation + 2 happy paths) |
| `frontend/src/types/schedule.ts` | **New** | `InvestigationSchedule` + `ScheduleInput` |
| `frontend/src/api/schedules.ts` | **New** | 5 axios endpoints |
| `frontend/src/stores/schedules.ts` | **New** | Pinia store with `runningId` |
| `frontend/src/stores/__tests__/schedules.test.ts` | **New** | 6 vitest tests |
| `frontend/src/utils/cron-presets.ts` | **New** | Presets + `humanizeCron` + `formatInterval` + `formatSchedule` |
| `frontend/src/pages/SchedulesPage.vue` | **New** | List page |
| `frontend/src/components/schedules/ScheduleCard.vue` | **New** | Per-schedule card |
| `frontend/src/components/schedules/ScheduleModal.vue` | **New** | Create/edit modal |
| `frontend/src/router/index.ts` | Edit | `/schedules` route |
| `frontend/src/components/common/AppSidebar.vue` | Edit | Nav entry |
| `frontend/src/test/setup.ts` | Edit | Added `patch: vi.fn()` |
| `docs/vision.md` | Edit | +Scheduled agent mode, +Platform Sections entry |
| `docs/changelog.md` | Edit | 0.31.0 entry |

Net diff: 11 new files, 9 modified, 0 deleted. **Zero new migrations** — the `cron_expr` column has been in the schema since Phase 3.

---

## Validation

### Backend

```bash
cd backend && go vet ./...      # clean
cd backend && go build ./...    # clean
cd backend && go test ./...     # all packages pass
```

Verbose scheduler test run:

```
=== RUN   TestShouldFire_NeverRun                        PASS
=== RUN   TestShouldFire_IntervalNotYetElapsed           PASS
=== RUN   TestShouldFire_IntervalJustElapsed             PASS
=== RUN   TestShouldFire_IntervalLongPast                PASS
=== RUN   TestShouldFire_CronDue                         PASS  (new)
=== RUN   TestShouldFire_CronNotYetDue                   PASS  (new)
=== RUN   TestShouldFire_CronPrefersCronOverInterval     PASS  (new)
=== RUN   TestShouldFire_CronInvalidExpression           PASS  (new)
=== RUN   TestShouldFire_CronNeverRun                    PASS  (new)
=== RUN   TestParseCronExpression_Valid                  PASS  (new)
=== RUN   TestSchedulerTick_DBError                      PASS
=== RUN   TestInvestigationScheduler_CancelExits         PASS
=== RUN   TestRunScheduledInvestigation_DBErrorMarksRunError PASS
```

The `TestShouldFire_CronInvalidExpression` and `TestRunScheduledInvestigation_DBErrorMarksRunError` emit `ERROR`-level log lines during the run — that's the *expected* behavior being asserted. The tests check that the error is logged and the function returns safely without panic, not that execution is silent.

Handler tests continue to skip without `DATABASE_URL`/`SUPABASE_URL`, same as Phase 3.

### Frontend

```bash
cd frontend && npm run test -- --run   # 29/29 pass across 5 files
cd frontend && npx vue-tsc -b --noEmit # clean
cd frontend && npm run build           # 1.06s, SchedulesPage chunk 16.89 kB (4.99 kB gzip)
cd frontend && npx eslint <Phase 4 files>  # zero errors
```

Full vitest run output:

```
 ✓ src/stores/__tests__/schedules.test.ts (6 tests)
 ✓ src/stores/__tests__/auth.test.ts (5 tests)
 ✓ src/stores/__tests__/connections.test.ts (5 tests)
 ✓ src/stores/__tests__/logs.test.ts (5 tests)
 ✓ src/components/connections/__tests__/ConnectionForm.test.ts (8 tests)

 Test Files  5 passed (5)
      Tests  29 passed (29)
```

### Lint note

Running `npm run lint` against the whole frontend reports ~1,150 pre-existing errors across the codebase (mostly `no-explicit-any` in existing stores, unused imports, and a few other baseline issues). **My Phase 4 files are lint-clean** when scoped specifically — I used `unknown` for catch clauses, `MockAxiosResponse<T>` for test mocks, and avoided introducing any new baseline debt. The pre-existing lint errors are out of scope for this phase.

### Not validated in this pass

- **Live end-to-end cron fire against a real LLM.** The plan's validation gate says: *"Create a schedule via the UI with `*/5 * * * *` + a prompt like 'Query the production DB for any deadlocked transactions in the last 10 minutes'. Observe it fire within 5 minutes. `last_run_at`, `last_status='success'`, `last_summary` all populated. Corresponding row visible on the Activity page with `entry_type='scheduled_investigation'`. Disabling the schedule from the UI prevents further runs. 'Run now' fires the schedule synchronously and returns a summary. Invalid cron expression in the modal shows an error before the request is sent."*

  This requires a running backend with `DATABASE_URL`, an Anthropic API key, a real app configured in the DB, and at least one connected data source the agent can actually investigate. Worth doing once Phase 4 is deployed to staging. The scheduler test coverage and handler test coverage exercise everything short of the real LLM round-trip.

- **Migration applied to any real database.** Not applicable — Phase 4 ships zero migrations. The `cron_expr` column was part of migration 023 (Phase 3), which must already be applied for the handlers to work at all.

---

## Deployment notes

**No new env vars. No new Fly secrets. No new migrations.** Phase 4 is pure-code.

Deployment checklist:

1. Merge branch, CI runs `make lint && make test`.
2. Deploy backend. Confirm startup logs still show `agent started, monitoring + pruner + scheduler goroutines spawned`.
3. Deploy frontend. Confirm the new "Schedules" entry appears in the sidebar under Agent.
4. Smoke test via the UI:
   - Navigate to `/schedules`. Observe empty state for any app with no schedules.
   - Click "New Schedule". Confirm the modal opens with Preset mode selected by default.
   - Click through the three mode toggles (Preset / Custom cron / Interval). Confirm the form updates.
   - Create a schedule with a Preset cron. Confirm it appears in the list with "Pending" status.
   - Click "Run now". Confirm the button shows "Running…" and the card updates with a last-summary after the backend responds.
   - Click "Edit", change the name, save. Confirm the card updates.
   - Click "Delete", confirm the native dialog, confirm the card vanishes.
   - Try to create a schedule with an invalid cron expression like `"garbage"`. Confirm the backend returns a 400 with the parser error text visible in the toast.

---

## Deviations from the plan

| Plan said | I did | Why |
|-----------|-------|-----|
| Modal has two modes: Preset + Custom cron | **Three modes**: Preset + Custom cron + **Interval** | Phase 3 schedules have `interval_secs` only; forcing them into cron mode on load would be lossy. Interval mode keeps them editable without conversion. |
| Per-call `cronParser := cron.NewParser(...)` inside `shouldFire` | **Package-level `var cronParser`** | The parser is stateless and reusable; per-call construction is needless allocation in a loop that runs every minute. |
| `cronstrue` for humanizing cron | **Local `humanizeCron`** that matches against presets and falls back to raw | 20 KB dep for rendering five presets is overkill. Raw cron is the correct affordance for users who opted into Custom mode. |
| Plan didn't specify `runningId` | Added per-row `runningId` state in the store | Mirrors `connections.ts testingId`. A global loading flag would block editing other rows during a Run-now call. |
| Handler `extractApiError` helper (plan didn't mention it) | Added to both store and page | Needed to pull backend validation text out of axios rejections; two copies is below the "rule of three" threshold. |
| Plan didn't discuss the global test mock | Added `patch: vi.fn()` to `src/test/setup.ts` | The schedules API is the first in the codebase to use PATCH. Fixing the global mock is a one-line change that benefits future PATCH users. |

---

## Follow-ups and open edges

1. **Paused-app filter on `ListEnabledSchedules`** — still not added. Phase 3 deferred this to Phase 4; Phase 4 is deferring it again because the UX question (should "pause app" stop scheduled investigations too, or keep them running as a separate kind of monitoring?) isn't concrete yet. A paused app will still fire its schedules. One-line fix when the UX is decided: `AND a.status = 'active'` on the query.

2. **Parallel vs serial tick execution** — still serial. If a user has >5 schedules per app and the per-run LLM call takes more than ~15 seconds, the tick can pile up. The fix is the same as the Phase 3 follow-up: mirror `monitorTick`'s semaphore pattern. Not urgent.

3. **`RunMonitoring` `flaggedLogs` parameter name** — still misleading for the scheduled-investigation path (where it's arbitrary prompt text). Rename to `initialInput` is a clean future change; not done here to avoid touching `monitor.go`.

4. **Prometheus metric `scheduled_investigations_runs_total{status="..."}`** — still not added. The plan flagged this in Phase 3 and Phase 4; neither phase shipped it. Worth doing as a small observability pass alongside a broader metrics review of `scheduler.go` + `monitor.go`.

5. **`create_report: boolean` flag on schedules** — the Phase 3 open question #1 ("should scheduled investigations produce `reports` rows in addition to `agent_log` entries?") is still open. I deliberately didn't add it in Phase 4 because the UX value is unclear — most schedules will be "every 5 minutes check pg_stats" which don't warrant a full incident report, and users who want reports already have the Reports page. Revisit if users actually ask for it.

6. **Humanized cron for custom expressions** — if users end up typing raw cron expressions a lot, the 20 KB `cronstrue` dep becomes more defensible. The drop-in point is `humanizeCron` in `utils/cron-presets.ts` — swap the fallback branch.

7. **Schedule run history table** — right now we only remember the most recent run's fields (`last_run_at`, `last_status`, `last_summary`, `last_error`). A full `schedule_runs` history table would enable "show me this schedule's last 10 runs" trend views. Don't build it speculatively — wait for a user to ask.

8. **The `extractApiError` duplication** — one in `stores/schedules.ts`, one in `pages/SchedulesPage.vue`. Lift to `utils/api-errors.ts` if a third copy appears.

---

## What Phase 4 proves about the broader plan

The plan's phased approach was:

1. Phase 1 & 2: small feed-polish and pruner changes to unblock and de-risk the agent startup topology
2. Phase 3: de-risk the `RunMonitoring` reuse with an API-only MVP
3. Phase 4: layer cron parsing + UI on the proven substrate

**Phase 4 was almost entirely additive.** No core scheduler logic changed, no migrations, no new goroutines. The Phase 1-3 substrate already handled the goroutine topology, the persistence layer, the rate limiter, and the `agent_log` emission. Phase 4 only had to:

- Swap `shouldFire` to prefer cron when set (20 lines)
- Update handlers to validate `cron_expr` (30 lines)
- Build the frontend (the bulk of the work, entirely in `frontend/`)

**The `cron_expr` column being nullable in migration 023** (Phase 3) was the single most load-bearing decision for Phase 4's non-migrating shape. If Phase 3 had shipped `interval_secs` only, Phase 4 would have needed migration 024 to add `cron_expr`, and the v0.30.2 deployment-drift risk would be back. Phase 3's completion doc explicitly flagged this as "reserved for Phase 4, schema-safe now" — and that deliberate over-provisioning paid off directly.

**The `RunMonitoring` reuse continues to hold.** Phase 4 does not touch `monitor.go`, `loop.go`, or any of the tool-registry or provider-resolution code. The scheduler still calls `a.RunMonitoring(runCtx, userID, appConfig, s.Prompt)` exactly like Phase 3 did; the only difference is that *when* it calls it is now cron-determined instead of interval-determined.

**The architecture passes the three-callers test.** `RunMonitoring` now has three distinct callers (interactive chat, monitor loop, scheduler), which is the point at which an abstraction is stable. Three callers is the threshold that separates "generalized code" from "premature abstraction" — and we got here organically, without refactoring the core loop for the scheduler use case.

Phase 4 is the last planned phase for the Logs Feed & Scheduled Investigations initiative. The feature is feature-complete for the "v1" plan scope. Everything beyond this is follow-up work driven by real-user signal.
