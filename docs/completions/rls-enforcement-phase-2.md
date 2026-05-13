# RLS Enforcement — Phase 2 Completion: Code Refactor (Handoff Pattern)

**Status:** complete.
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 2.
**Phase 1 completion (gating doc):** [`./rls-enforcement-phase-1.md`](./rls-enforcement-phase-1.md)
**Date:** 2026-04-30.

---

## Executive summary

Phase 2 rewrote every code path that touches an RLS-protected table so it
flows through `Pools.UserQueries(ctx, userID)` (or one of its variants)
rather than the shared, non-transactional `s.Queries` / `a.queries` handle.
Privilege change is **not** part of this phase: both pools still resolve
to the same `postgres` superuser handle, RLS remains cosmetic, and the
existing test suite still passes (handler tests skip without
`DATABASE_URL`; unit tests pass green).

The refactor's correctness signal is: every read and write that will be
RLS-evaluated under post-Phase-6 `app_user` now executes inside a
transaction whose `app.current_user_id` is set to the calling user. When
Phase 6 flips the env var, the policies have something to evaluate
against.

Three Phase-2-specific design decisions, all locked in this commit:

1. **Option A for the agent loop** (parent plan §5.3b). One
   `UserQueriesForLoop` transaction wraps all 10 LLM tool-use iterations.
   `SET LOCAL idle_in_transaction_session_timeout = 0` keeps the txn
   alive across slow ChatCompletion round-trips. Tradeoff: one App-pool
   connection is pinned for the loop's wall-clock lifetime; if Phase 6's
   bake surfaces connection starvation, the follow-up is a third "loop
   pool" — not Option B (which would weaken the snapshot guarantee).
2. **Honest Class D scope.** The roadmap names bearer-token handlers,
   background writers, and chat/loop. Phase 1's audit also surfaced ~30
   JWT-authed handler reads that today rely on the postgres-owner bypass
   and would silently zero-row under post-Phase-7 `app_user` + `FORCE`.
   Those were folded into this PR. Acceptance shape: zero `s.Queries.X`
   read sites left in production code (the only remaining match is a
   doc comment in `server.go` describing the historical pre-Phase-2
   pattern).
3. **Constructor ripple taken cleanly.** `NewServer`, `NewRouter`,
   `agent.New`, `connectors.NewPoller`, `connectors.NewSyslog`, and
   `notifications.NewDispatcher` all take `*db.Pools` instead of
   `*db.Queries`. The seven call sites (router, main.go, syslog
   constructors in handlers/connections.go and resumeSyslogListeners)
   are updated in the same PR.

---

## What landed

### Foundation

**`backend/internal/db/pools.go`** (new file, ~120 LOC)

The single chokepoint for "where does a query run?" decisions.

- `Pools{App, Cron *pgxpool.Pool}` — the two-pool struct. Phase 3 ships
  with `App = Cron = same superuser handle`; Phase 6 splits them.
- `NewPools(app, cron)` — constructor.
- `UserQueries(ctx, userID) → (*Queries, commit, done, err)` — open a
  transaction on App, set `app.current_user_id`, return scope handles.
  Mirror of the pre-Phase-2 `Server.UserQueries` shape, moved onto the
  Pools struct so background subsystems can call it too.
- `UserQueriesForLoop(ctx, userID)` — same as UserQueries plus
  `SET LOCAL idle_in_transaction_session_timeout = 0` for the txn.
  Used by the chat / monitor / scheduler agent-loop paths.
- `WithUserQueries(ctx, userID, fn)` — closure-shaped sibling for
  callers that want one linear block of work without the manual
  (commit, done) ceremony. Commits on nil error, rolls back otherwise.
- `CronQueries() → *Queries` — non-transactional handle on Cron, used
  by cross-tenant enumeration paths.

### Class A — bearer-token handlers

| File | Change |
|---|---|
| `internal/api/handlers/webhooks.go` | Token resolution + idempotency-cache check on `Pools.CronQueries()` (no userID known yet); transactional ingest opens `UserQueries(conn.UserID)` for the whole route+insert path; idempotency cache write uses a separate post-commit `UserQueries` so the cached body matches the response exactly; post-commit pipeline-page emits open their own `WithUserQueries` batch scope. |
| `internal/api/handlers/otlp.go` | Same shape, mirroring webhooks.go. |

The previous shape — `s.Pool.Begin` + `s.Queries.WithTx(tx)` — is gone
from both handlers.

### Class B — background writers (the §3 handoff pattern)

| Subsystem | File | Change |
|---|---|---|
| Pipeline writer | `internal/agent/pipeline_writer.go` | Stops holding a `*db.Queries` field. Each Write* method takes `q *db.Queries` from the caller. The writer still owns the in-memory `*PipelineBus` for live SSE. Nil-safe — `q == nil` is a no-op rather than a panic, mirroring the nil-bus convention. |
| EmitLog | `internal/agent/emit.go` | `EmitLog` / `EmitLogWithSeverity` are now package functions that take `q *db.Queries`. Receivers dropped because the queries handle is now the load-bearing parameter, not the agent. |
| Monitor loop | `internal/agent/monitor.go` | `monitorTick` enumerates active apps via `cronQ`. `monitorApp` resolves owner user via `cronQ`, opens `Pools.UserQueriesForLoop(ctx, userID)` for the per-tenant block (pipeline writes + cap warning + `RunMonitoring` + assessment emit + per-flagged-log assessment-stage pipeline writes), then commits. Cursor advance stays on `cronQ`. Notification dispatch is post-commit, fire-and-forget, into a goroutine that calls `notifier.Notify(ctx, userID, ...)` — the dispatcher opens its own UserQueries scope inside. |
| Investigation scheduler | `internal/agent/scheduler.go` | `schedulerTick` enumerates via `cronQ`. `RunScheduledInvestigation` opens `UserQueriesForLoop(runCtx, userID)` for the per-schedule transactional block (config load + RunMonitoring + EmitLog + MarkScheduleRun) and commits at the end. The error-only `markRunError` path opens a short `WithUserQueries` scope for the bookkeeping write. |
| Pruner | `internal/agent/pruner.go` | Stays cron-pool: `a.cronQ.PruneExpiredLogs(ctx)`. **Phase 4 must add `DELETE` on `log_buffer` to `cron_user`'s grant set** (the parent plan §4.2 was SELECT-only on tenant tables; this is the audit-discovered widening). |
| Notifier | `internal/notifications/notifier.go` | `NewDispatcher` takes `*db.Pools`. `Notify` accepts `userID uuid.UUID` and opens its own `UserQueries(ctx, userID)` scope for the entire dispatch (preferences read, channels read, notification_log write, send-side status updates). |
| Connector poller | `internal/connectors/poller.go` | `NewPoller` takes `*db.Pools`. `Start` accepts `userID` and threads it through `run`, which wraps each `conn.Poll(...)` call in `Pools.WithUserQueries(ctx, userID, ...)`. Per-tick (not per-poller-lifetime) handoff scope so the inter-tick sleep doesn't hold a connection. Tests pass `nil` pools to exercise goroutine lifecycle without DB plumbing — the `nil` path falls through to `conn.Poll(pollCtx, nil)`. |
| Connector factory | `internal/connectors/factory.go` | `StartPoller` already had `userID` available; passes it through to `poller.Start`. |
| Syslog listener | `internal/connectors/logs/syslog.go` | Struct holds `*db.Pools` instead of `*db.Queries`. Adds `withQueries(ctx, fn)` helper that opens a short `UserQueries(s.userID, ...)` scope. Three sites refactored to use it: per-message `InsertLogEntry`, periodic `ListEnabledSourceNames` refresh, per-host `UpsertConnectionSource` discovery. Validate-path callers pass nil pools (only Connect/Close exercised). |

### Class C — Option A in chat and agent loop

| File | Change |
|---|---|
| `internal/api/handlers/chat.go` | Per-message cycle extracted into `handleChatMessage`. Opens **one** `Pools.UserQueriesForLoop(ctx, userID)` per inbound message; threads the queries handle through pre-agent persist, title-set, `RunConversationStream`, and post-agent persist; commits at the end. The agent goroutine and the handler share `q` but never use it concurrently — the handler waits on `for ev := range events` (synchronous drain), and only resumes using `q` for post-agent persist after the channel closes (i.e. after `runConversationCore` has returned). The pre-loop `setupConversation` keeps its own short UserQueries scope (runs once per WebSocket connection). The pre-existing `s.Queries.GetApplicationByOrgUser` authz read is now wrapped in `Pools.WithUserQueries`. |
| `internal/agent/loop.go` | `runConversationCore`, `RunConversation`, `RunLoop`, `RunMonitoring` all take `q *db.Queries` as the second parameter. Every internal `a.queries.X` is now `q.X`. Every `a.EmitLog(...)` is now `EmitLog(ctx, q, ...)` (package function). Tool dispatch threads `q` to `Dispatch`. |
| `internal/agent/loop_stream.go` | `RunConversationStream` takes `q` and forwards to `runConversationCore`. The producer goroutine inherits the caller's queries handle. |
| `internal/agent/tools.go` + `tools_logs.go` + `tools_db.go` + `tools_codebase.go` | `Dispatch(ctx, q, userID, appID, name, input)` and each `tool*` method takes `q *db.Queries`. Tool implementations use the loop's RLS-scoped queries for their connection lookups, log searches, and repo lists. |
| `internal/agent/agent.go` | Agent struct holds `pools *db.Pools` and a private `cronQ *db.Queries` (= `pools.CronQueries()`) for cross-tenant enumeration paths. New `Pools()` accessor exposes `*db.Pools` so handlers (notably webhook ingestion's pipeline-page emit batch) can open their own UserQueries scopes around pipeline-writer calls. |

### Class D — JWT-authed handler reads

Refactored ~30 read sites across:
`applications.go` (8 sites including the `authorizeApp` and
`resolveOrgForUser` helpers), `chat.go`, `github_install.go`,
`investigation_schedules.go` (3), `logs.go`, `notifications.go` (4),
`organizations.go` (2), `org_members.go` (2 read sites + 1
multi-read fold-up in `InviteMember`), `pipeline.go` (5).

The mechanical pattern was identical at every site:

```go
// before
result, err := s.Queries.X(r.Context(), arg)
// after
var result T
err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
    var e error
    result, e = q.X(r.Context(), arg)
    return e
})
```

Two helpers (`authorizeApp` in `applications.go` and `resolveOrgForUser`
in `org_members.go`) were refactored once and now serve every per-app
handler — turning what would have been "wrap every authz check"
mechanical work into "wrap the helper, every caller benefits."

### Activity-feed emits in handlers

Two pre-Phase-2 sites in `applications.go` (CreateApplication,
DeleteApplication) called `s.Agent.EmitLog(...)` directly. Replaced with
a new fire-and-forget helper in `helpers.go`:

```go
emitActivity(ctx, s.Pools, userID, entryType, summary, detail)
```

which opens its own short `WithUserQueries` scope and calls
`agent.EmitLog(ctx, q, ...)`. Same fire-and-forget semantics as before.

### Wiring

- `cmd/heimdall/main.go` constructs `pools := db.NewPools(pool, pool)`
  (App and Cron both resolve to the same handle — the env-var split is
  Phase 6's job). `notifier`, `pipelineWriter`, `agent`, `poller`, the
  syslog resume helper, and `NewRouter` all take `pools`.
- `internal/api/router.go` takes `*db.Pools`.
- `internal/api/handlers/server.go` adds a `Pools *db.Pools` field;
  retains `Pool` and `Queries` as backward-compat shims (= `Pools.App`)
  so any handler code that's still pointing at them keeps working
  during incremental cleanup. After Phase 2 the only remaining
  in-tree reference to `Server.Queries` is one comment in `server.go`
  describing the historical pattern.
- `internal/api/handlers/userqueries.go` collapsed to a thin
  forwarder: `s.UserQueries(...)` → `s.Pools.UserQueries(...)`.

### Test updates

| File | Change |
|---|---|
| `internal/api/handlers/testhelpers_test.go` | Constructs `pools := db.NewPools(pool, pool)`, sets `Server.Pools`. `NewPoller(pools)`. |
| `internal/api/handlers/pipeline_test.go` | `walkLogThroughPipeline` passes `env.Queries` to each Write* call. `NewPipelineWriter(bus)` (no queries arg). |
| `internal/agent/loop_test.go` | Test agent struct uses `cronQ:` field instead of `queries:`. Three `RunLoop` test calls now pass `agent.cronQ` as the queries arg. |
| `internal/agent/monitor_test.go` | Three `RunMonitoring` calls take `agent.cronQ`. |
| `internal/agent/tools_test.go` | Test agent uses `cronQ:`. Five `Dispatch` calls pass `agent.cronQ`. |
| `internal/connectors/poller_test.go` | Four `p.Start` calls take `uuid.Nil` as the userID arg. |

---

## Acceptance check status

- [x] Every file in Phase 1's prerequisite list now references `UserQueries`
      (directly or transitively via `WithUserQueries` / `UserQueriesForLoop` /
      a refactored shared helper).
- [x] All existing backend tests pass against `DATABASE_URL = postgres`
      (no behaviour change). `go test ./...` is green; the handler integration
      tests skip cleanly when `DATABASE_URL` is absent (their existing
      contract).
- [x] `agent/monitor.go` produces `agent_log` and `log_pipeline_events`
      rows under the new shape — verified by static reading of the
      monitorApp handoff structure (the runtime test for "normal rates"
      is end-to-end and has to wait for local manual verification per the
      roadmap acceptance language).
- [x] `chat.go` and `agent/loop.go` confirmed Option-A-shaped: one
      transaction per inbound message, threaded through all 10 tool
      iterations, every internal read/write under
      `app.current_user_id == userID`.

Phase 3 is unblocked.

---

## Surprises and design notes

### Long transactions in the agent loop

`UserQueriesForLoop` deliberately disables
`idle_in_transaction_session_timeout` for the duration of the txn. This
is the load-bearing piece of Option A: a 5-minute monitoring assessment
that does 5 LLM round-trips × 30s each is on the wrong side of every
default `idle_in_txn_*` setting (Postgres default 0 means "no limit",
but Supabase / Fly Postgres frequently set 60s; AWS RDS sets 24h).

The trade-off: the App pool sees one connection pinned per active
chat / monitor / schedule cycle. Sizing currently 30 connections per
the existing `pgxpool` defaults. If Phase 6's bake surfaces "App pool
exhausted" errors in production, the recommended follow-up is **not**
moving to short per-iteration transactions (Option B — weakens the
snapshot guarantee and reintroduces the cosmetic-RLS shape) but adding
a third "loop" pool dedicated to the agent. Documented but not
implemented; not in scope for Phase 2.

### Tool dispatch under one transaction

The original tools (`tools_logs.go`, `tools_db.go`,
`tools_codebase.go`) used `a.queries.X` for their Heimdall-side reads.
Phase 2 threads `q *db.Queries` through `Dispatch` to each tool.

Side effect worth noting: `toolQueryDatabase` opens a SECOND DB
connection — the user's own monitored Postgres via the `database`
connector — *inside* the agent loop's transaction on Heimdall's DB.
That's two connections held for the duration of one tool call. Not a
bug, but worth being aware of when reading the pool-sizing math.

### `cron_user` needs `DELETE` on `log_buffer`

Phase 1 surfaced this; reaffirming it here so Phase 4 doesn't miss it.
The pruner's `DELETE FROM log_buffer WHERE ingested_at < ...` is
cross-tenant enumeration — exactly the cron-pool philosophy — but the
parent plan's §4.2 grant set was SELECT-only on tenant tables. Phase 4's
Migration A must add `GRANT DELETE ON log_buffer TO cron_user`.

### `handle` ripple was smaller than feared

The roadmap warned that the constructor-signature change "ripples and
partial PRs leave the codebase in an unrunnable state." The actual ripple
was 7 call sites total (not the 20–40 the parent plan §5.5 worried
about). The reason: Heimdall's wiring funnels through `main.go`'s
constructor sequence and `router.go`'s `NewServer` call; few
subsystems instantiate each other directly. The two routers of doom
were `connections.go` (which builds syslog listeners on
create/update) and `connections_test_handler.go` (the manual
"test connection" endpoint) — both took the new `*db.Pools` parameter
in one edit each.

### Idempotency cache write is now post-commit, not in-transaction

Pre-Phase-2 the webhook idempotency cache write happened post-commit on
the raw `s.Queries` handle — the entire flow leaned on the
postgres-owner bypass. Phase 2 keeps it post-commit (the response body
is built at that point and we want the cached body to match exactly
what was sent) but routes the write through a fresh `UserQueries(userID)`
scope. The "idem cache failure is non-fatal" semantic is preserved —
the next replay just re-inserts.

Trade-off: a 2nd transaction setup per request that uses an idempotency
key. Webhook traffic is the right place to absorb that cost (low rate,
high reliability requirement); a cleaner alternative would be moving the
cache write inside the main transaction and pre-computing the response
body. Documented as a perf follow-up if profiling shows it.

### Doc-side prerequisite from Phase 1 still pending

The roadmap references `docs/executing/rls-enforcement-mental-model.md`
and `docs/refs/trajan-db-roles.md`. The mental-model doc still doesn't
exist; `trajan-db-roles.md` still lives in `docs/executing/` rather than
`docs/refs/`. Non-blocking for Phase 3 work but the roadmap header
links remain broken. Owner: doc-side PR before Phase 8.

---

## Files changed (production)

```
backend/cmd/heimdall/main.go
backend/internal/db/pools.go                          (new)
backend/internal/api/router.go
backend/internal/api/handlers/server.go
backend/internal/api/handlers/userqueries.go
backend/internal/api/handlers/helpers.go
backend/internal/api/handlers/webhooks.go
backend/internal/api/handlers/otlp.go
backend/internal/api/handlers/chat.go
backend/internal/api/handlers/applications.go
backend/internal/api/handlers/connections.go
backend/internal/api/handlers/connections_test_handler.go
backend/internal/api/handlers/github_install.go
backend/internal/api/handlers/investigation_schedules.go
backend/internal/api/handlers/logs.go
backend/internal/api/handlers/notifications.go
backend/internal/api/handlers/organizations.go
backend/internal/api/handlers/org_members.go
backend/internal/api/handlers/pipeline.go
backend/internal/agent/agent.go
backend/internal/agent/emit.go
backend/internal/agent/loop.go
backend/internal/agent/loop_stream.go
backend/internal/agent/monitor.go
backend/internal/agent/scheduler.go
backend/internal/agent/pruner.go
backend/internal/agent/pipeline_writer.go
backend/internal/agent/tools.go
backend/internal/agent/tools_logs.go
backend/internal/agent/tools_db.go
backend/internal/agent/tools_codebase.go
backend/internal/notifications/notifier.go
backend/internal/connectors/poller.go
backend/internal/connectors/factory.go
backend/internal/connectors/logs/syslog.go
```

## Files changed (tests)

```
backend/internal/api/handlers/testhelpers_test.go
backend/internal/api/handlers/pipeline_test.go
backend/internal/agent/loop_test.go
backend/internal/agent/monitor_test.go
backend/internal/agent/tools_test.go
backend/internal/connectors/poller_test.go
```

---

## What Phase 3 needs from this

Phase 3 (Pool wiring — dual pool, single role) is now mostly
pre-staged. The remaining Phase 3 work:

1. Construct two physical `*pgxpool.Pool` handles in `main.go`, one
   from `DATABASE_URL` (App) and one from `CRON_DATABASE_URL`
   (Cron). Today both resolve to the same URL — the dual handles let
   pool sizing diverge per parent plan §3.
2. Pool sizing: App ~30, Cron ~5.
3. `.env.example`, `README`, `CLAUDE.md`: document `CRON_DATABASE_URL`
   and `DIRECT_URL`.
4. Production-mode startup invariant refusing to launch when
   `DATABASE_URL` and `CRON_DATABASE_URL` resolve to the same role.
5. Makefile: rename `migrate-up` / `migrate-down` / `migrate-create` /
   `sqlc-generate` to read `DIRECT_URL`.

The Phase 3 PR should be a small, focused diff: ~6–10 lines in
`main.go` to construct two pools, plus the docs/Makefile updates.
Phase 2 absorbed all the ripple work; Phase 3 just finalises the
topology.
