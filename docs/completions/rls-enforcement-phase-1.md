# RLS Enforcement — Phase 1 Completion: Audit & Discover

**Status:** complete.
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 1.
**Date:** 2026-04-30.
**Audited against commit:** local working tree at start of session (migrations 001–038, no working-tree changes).

---

## Executive summary

The audit confirms the parent plan's structural assumptions but expands the
Phase 2 prerequisite list well beyond the named call sites. The key surprises:

1. **Almost every handler that hits an RLS-protected table is at risk under
   `app_user`.** The `s.Queries.X` direct-call pattern is the dominant shape
   in `backend/internal/api/handlers/`; it bypasses RLS today only because
   the connection is `postgres`. With `app_user` and `FORCE`, every such
   site that doesn't go through `UserQueries` collapses to zero rows.
   This is not just `webhooks.go` / `otlp.go` — it is order-of-30 sites
   across the handlers package.
2. **The agent loop and chat handler both fail Option A.** `agent/loop.go`
   and `agent/handlers/chat.go` operate on the shared, non-transactional
   `a.queries` / `s.Queries` handles for the duration of a prompt cycle.
   No `set_config('app.current_user_id', ...)` is ever applied to the
   reads inside `runConversationCore` / `RunMonitoring`. Both are Phase 2
   prerequisites.
3. **Four tables have RLS enabled with zero policies** — the
   "system-table" idiom from migrations 030 / 033 / 038 that relies on the
   `postgres` owner bypass. Three of them (`webhook_idempotency`,
   `log_pipeline_events`, `agent_config`) are touched by the runtime
   `app_user` path. Under `FORCE`, every read and write fails.
   Phase 7 (Migration B) must add policies, not just sprinkle `FORCE`.
4. **One stale policy: `github_repos_owner` (migration 020) is pre-035
   shape.** It still scopes by `connections.user_id`, but migration 035
   moved connections to org-scoped. Other org members can't see GitHub
   repos for their org's connections under FORCE. Pre-existing latent bug;
   surfaces hard at FORCE-flip time.
5. **`schedule_migrations` and `agent_config`** RLS-enabled-but-policyless
   are arguably benign because the runtime `app_user` should never read them
   — but the audit can't prove "never reads them" without a code-side check
   too. See Phase 2 prerequisite list for confirmation work.
6. **`§4.5` is genuinely empty** as the parent plan predicted. All three
   user-defined functions are either `SECURITY DEFINER` (`handle_new_user`,
   `app_user_org_ids`) or read no RLS-protected table (`app_current_user_id`
   reads only a GUC). No retrofit migration needed.
7. **`cron_user`'s parent-plan-§4.2 grant set is too narrow for the pruner.**
   `agent/pruner.go` issues `DELETE FROM log_buffer` on a 1h schedule across
   all tenants. The plan grants `cron_user` SELECT on tenant tables and
   INSERT/UPDATE only on `monitoring_state`. Phase 4's Migration A will need
   either a `DELETE` grant on `log_buffer` for `cron_user`, or the pruner
   moves to a per-tenant `app_user` handoff (much heavier). Recommendation:
   widen the grant by one verb on one table; the pruner is enumeration-only
   so this stays inside the cron-pool philosophy.

The above are itemised in the **Phase 2 prerequisite list** below. Total
estimated refactor surface is ~30–40 files — consistent with the parent
plan's §5.5 estimate, but the work is more concentrated in the handlers
package than the plan implied.

---

## §5.0 — `service_role` audit

**Result: zero hits in Heimdall code.**

```bash
grep -rn 'service_role\|SUPABASE_SERVICE_ROLE_KEY\|SERVICE_ROLE_KEY' \
  backend/ frontend/ --include='*.go' --include='*.ts' --include='*.vue'
```

All matches are inside `frontend/node_modules/@supabase/auth-js/...`
(vendored SDK source/typings). No call site in `backend/`, no application
code in `frontend/src/`, no environment variable named `SUPABASE_SERVICE_ROLE_KEY`
in the project's `.env` shape (verified by absence of any such reference in
non-`node_modules` paths).

**Disposition:** clean. Phase 2+ unblocked on this gate. PostgREST is not
exposed for tenant data in Heimdall today, so `service_role` is structurally
absent rather than allowlisted.

---

## §5.1 — Superuser-only behaviour audit

**Result: zero production hits; test-only references to `auth.users`.**

Searched `backend/` for `CREATE EXTENSION`, `ALTER SYSTEM`, `SET ROLE`,
`RESET ROLE`, `auth.users`, `auth.identities`, `auth.sessions`.

Hits, classified:

| Path | Pattern | Disposition |
|---|---|---|
| `backend/internal/api/handlers/testhelpers_test.go:71,116` | `INSERT INTO auth.users`, `DELETE FROM auth.users` | **Test-only.** Test fixtures bootstrap an auth user before exercising the public.users trigger. Tests run via `DATABASE_URL` (which stays `postgres` for the test profile via `DIRECT_URL`-shaped reuse). No runtime impact under `app_user`. |
| `backend/internal/api/handlers/organizations_test.go:27,36` | same | **Test-only.** Same disposition. |
| `backend/internal/api/handlers/pipeline_test.go:364,370` | same | **Test-only.** Same disposition. |

No runtime code reads `auth.*`, runs DDL, switches roles, or sets restricted
GUCs. The migrations themselves use `CREATE TABLE`, `CREATE POLICY`, etc.,
but those are Phase-4-DIRECT_URL territory, not runtime.

**Disposition:** clean for Phase 2. Test fixtures continue to use the
postgres-superuser pool via `DIRECT_URL` — that's the intended Phase 5
posture for `asAppUser`-shaped tests anyway, since `SET ROLE` requires
membership the test runner already has via Migration A's PG17-conditional
grant block.

---

## §5.2 — RLS policy verb coverage matrix

Built from `backend/migrations/*.up.sql` (no live DB query — schema state
is fully reproducible from the migration tree).

### Per-table verb coverage

| Table | RLS | Policies (final state) | Verbs | Owner predicate | Status |
|---|---|---|---|---|---|
| `users` | ✅ | `users_self FOR ALL` | S/I/U/D | `id = app_current_user_id()` | **OK + caveat** (see footnote) |
| `connections` | ✅ | `connections_org_member FOR ALL` (035) | S/I/U/D | `org_id IN (org_members of caller)` | OK |
| `conversations` | ✅ | `conversations_owner FOR ALL` | S/I/U/D | `user_id = app_current_user_id()` | OK |
| `agent_log` | ✅ | `agent_log_owner FOR ALL` | S/I/U/D | `user_id = app_current_user_id()` | OK |
| `log_buffer` | ✅ | `log_buffer_owner FOR ALL` | S/I/U/D | `user_id = app_current_user_id()` | OK |
| `investigations` | ✅ | `investigations_owner FOR ALL` | S/I/U/D | `user_id = app_current_user_id()` | OK |
| `organizations` | ✅ | `organizations_user_policy FOR ALL` | S/I/U/D | `id IN (org_members of caller)` | OK |
| `applications` | ✅ | `applications_user_policy FOR ALL` | S/I/U/D | `org_id IN (org_members of caller)` | OK |
| `monitoring_state` | ✅ | `monitoring_state_user_policy FOR ALL` | S/I/U/D | applications→org_members | OK (cron_user has direct grant + BYPASSRLS) |
| `notification_channels` | ✅ | `notification_channels_user_policy FOR ALL` | S/I/U/D | applications→org_members | OK |
| `notification_preferences` | ✅ | `notification_preferences_user_policy FOR ALL` | S/I/U/D | applications→org_members | OK |
| `notification_log` | ✅ | `notification_log_user_policy FOR ALL` | S/I/U/D | applications→org_members | OK |
| `app_agent_config` | ✅ | `app_agent_config_org FOR ALL` (rewritten 026) | S/I/U/D | applications→org_members | OK |
| `investigation_schedules` | ✅ | `investigation_schedules_org FOR ALL` (rewritten 026) | S/I/U/D | applications→org_members | OK |
| `connection_sources` | ✅ | `connection_sources_org_member FOR ALL` (035 rewrote 034) | S/I/U/D | connections→org_members | OK |
| `app_source_filters` | ✅ | `app_source_filters_owner FOR ALL` | S/I/U/D | applications→org_members | OK |
| `org_members` | ✅ | `org_members_user_policy FOR ALL` (027 + 028 fix) | S/I/U/D | own org via `app_user_org_ids()` (DEFINER) | OK |
| **`github_repos`** | ✅ | `github_repos_owner FOR ALL` | S/I/U/D | **`connections.user_id = app_current_user_id()`** (stale!) | **GAP** — see policy patch list |
| **`agent_config`** | ✅ | **none** | ∅ | — | **GAP** — see policy patch list |
| **`schema_migrations`** | ✅ | **none** | ∅ | — | benign (DIRECT_URL only) |
| **`webhook_idempotency`** | ✅ | **none** | ∅ | — | **GAP** — `app_user` writes |
| **`log_pipeline_events`** | ✅ | **none** | ∅ | — | **GAP** — `app_user` writes |

> **Footnote (`users_self`):** `users.id = app_current_user_id()` means
> a user can only see their own `users` row. Org-member listings that
> do `JOIN users u ON u.id = om.user_id` (e.g. for displaying member
> emails) will collapse to "only the caller" under FORCE. Confirmed
> presence of such queries via `org_members.go:165` (`GetUserByEmail`).
> Phase 7 needs either a SELECT policy that allows reading users in
> shared orgs, **or** the org-member listing queries route through
> `app_user_org_ids()`-style SECURITY DEFINER helpers. Recommended
> resolution: add a `users_org_visible` SELECT policy whose `USING`
> clause is `id IN (SELECT user_id FROM org_members WHERE org_id IN
> (SELECT app_user_org_ids()))`. Documented as a Phase 7 work item.

### Policy patches required in Phase 7 (Migration B)

1. **`agent_config`** — global singleton legacy table. Audit code use:
   `agent/loop.go:95` (`GetAgentConfig`) reads it as fallback when no
   `appID` is set. Since chat with `app_id=""` reads this under `app_user`,
   Phase 7 needs a policy. Cheapest: `FOR ALL USING (true)` since the
   table is a single-row global with no tenant data — RLS-enable was
   defence-in-depth against PostgREST, not isolation. Document the choice.
2. **`webhook_idempotency`** — written by `webhooks.go:438`
   (`InsertIdempotencyResult`) and read by `webhooks.go:296`
   (`GetIdempotencyResult`). Both will run under `app_user` post-Phase-2.
   Policy: scope by `connection_id IN (caller's connections)`, joining
   through the existing `connections_org_member` shape — i.e. the
   idempotency row is visible to anyone who can see the parent connection.
3. **`log_pipeline_events`** — written by `agent/pipeline_writer.go`
   (every Write* method) and read by `pipeline.go` handlers. Both run
   under `app_user`. Policy: scope by `app_id IN (applications visible to
   caller)`, mirroring the existing `applications_user_policy` shape.
4. **`github_repos` (stale policy patch).** Drop the existing
   `github_repos_owner` and recreate as
   `FOR ALL USING (connection_id IN (SELECT c.id FROM connections c JOIN
   org_members om ON om.org_id = c.org_id WHERE om.user_id =
   app_current_user_id()))`. This matches the post-035 connections shape
   and unblocks org-member access.
5. **`users` SELECT widening (footnote above).** Either a new
   `users_org_visible` policy or migrate org-member listing queries to
   a SECURITY DEFINER helper.

### `schema_migrations`

Touched only by `make migrate-up` / `make migrate-down`, which run via
`DATABASE_URL` today and will run via `DIRECT_URL` (postgres) post-Phase-3.
Neither `app_user` nor `cron_user` reads it. No policy needed; the RLS-enable
is defence-in-depth against PostgREST and stays valid.

---

## §5.3 — Non-JWT write-path audit

### Bearer-token handlers

| Handler | File:line | Current shape | Touches RLS tables? | Refactor |
|---|---|---|---|---|
| Auto-detecting webhook ingestion | `handlers/webhooks.go:244-271 (readWebhookRequest)`, `:288-449 (ingestEntries)` | Token resolved via `s.Queries.GetConnectionByWebhookToken`. Idempotency via `s.Queries.Get/InsertIdempotencyResult`. Body insert via `s.Pool.Begin` + `s.Queries.WithTx` (raw qtx). | Yes — `connections` (token resolve), `webhook_idempotency`, `connection_sources`, `app_source_filters`, `log_buffer`. | Resolve token → `conn.UserID` → entire transaction (idempotency check + route + insert + idempotency write) inside one `UserQueries(ctx, conn.UserID)` scope. |
| Format-specific webhook ingestion | same file `:200-222` | Reuses `readWebhookRequest` + `ingestEntries`. | Same. | Same. |
| OTLP ingestion | `handlers/otlp.go:74-178 (IngestOTLPLogs)` | Token resolved via `s.Queries.GetConnectionByWebhookToken`. Insert via `s.Pool.Begin` + `s.Queries.WithTx` (raw qtx). No idempotency layer here (OTLP path does not have idempotency keys today). | Yes — `connections`, `connection_sources`, `app_source_filters`, `log_buffer`. | Same shape: resolve token → `conn.UserID` → `UserQueries(ctx, conn.UserID)` for the whole insert. |
| GitHub webhook receiver | searched: no dedicated `github_webhooks.go` exists in this tree. The GitHub *App* installation flow is in `handlers/github_install.go` and is JWT-authed. | n/a | n/a | The parent plan §5.3 listed this as "confirmed candidate"; in the current tree it is **not** a bearer-token write path. Note for plan: this candidate doesn't apply. |
| Syslog TLS listener | `connectors/logs/syslog.go:200-269 (Listen)` plus `handleConnection`; constructed in `main.go:200` and `handlers/connections.go:251,428`. | Listener-package code holds a `*db.Queries` from the raw pool and inserts into `log_buffer` per accepted line. | Yes — `log_buffer`. | Cannot use a request-scoped `UserQueries` because the listener is long-lived and accepts traffic from many remote senders. Options: (a) per-message `UserQueries(ctx, s.userID)` on the connection's owning user — heavy but straight-line; (b) hand off to a small ingestion helper that the handler / agent expose, mirroring the webhook path. (a) is closer to the parent plan's pattern; pick (a). |

### Background writers

| Subsystem | File:line | Current shape | Touches RLS tables? | Refactor |
|---|---|---|---|---|
| Monitor goroutine | `agent/monitor.go:29 (Monitor)`, `:48 (monitorTick)`, `:113 (monitorApp)` | `a.queries` (shared) used for `ListActiveApplications` (cross-tenant enumerate), `GetMonitoringState`, `ListLogsSinceForApp`, `GetAppAgentConfig`, `UpsertMonitoringState`, `EmitLog*`, `pipeline.Write*`. | Yes — `agent_log`, `log_pipeline_events`, plus reads on `app_agent_config`, `monitoring_state`, `log_buffer`. | Split: enumerate via cron pool; per-tenant work hands off to `UserQueries(ctx, ownerUserID)` (the user resolved by `resolveOrgUser`). The handoff covers `monitorApp` from cursor-read through emit. `UpsertMonitoringState` can stay on cron pool (cron_user has INSERT/UPDATE per parent §4.2). |
| Investigation scheduler | `agent/scheduler.go:63 (InvestigationScheduler)`, `:88 (schedulerTick)`, `:187 (RunScheduledInvestigation)` | Same `a.queries` shape. `ListEnabledSchedules` is cross-tenant; `RunScheduledInvestigation` does per-tenant `GetApplication`, `GetAppAgentConfig`, `EmitLog*`, `MarkScheduleRun`. | Yes — `investigation_schedules` (write `last_run_at`), `agent_log`. | Enumerate via cron pool; per-schedule run hands off to `UserQueries(ctx, ownerUserID)`. `MarkScheduleRun` writes `investigation_schedules` (RLS-protected) — must run inside the `UserQueries` scope. |
| Pipeline writer | `agent/pipeline_writer.go` (whole file) | Inserts into `log_pipeline_events` via shared `w.queries` from every Write* method. | Yes — `log_pipeline_events`. | Two callers exist: (a) webhook/OTLP ingestion (handler-side, has `userID` available — pass through and use the request's `UserQueries`), (b) `agent/monitor.go` (background, owner_user_id resolved by monitorApp — use the per-tenant `UserQueries` from the monitor refactor). Make `Write*` accept a `*db.Queries` parameter rather than holding one internally. |
| Notification dispatcher | `notifications/notifier.go:60 (Notify)`, `:95 (dispatchToChannel)`, `:129 (markSent)`, `:138 (markFailed)` | `d.queries` (shared) reads `notification_preferences`, `notification_channels`, writes `notification_log`. | Yes — all three. | Pass `userID` into `Notify`; resolve to `UserQueries(ctx, userID)` and thread the queries handle through `dispatchToChannel`/`markSent`/`markFailed`. The user is already resolvable in the call site (`monitor.go:261` has `userID` in scope). |
| Log buffer pruner | `agent/pruner.go:21 (Prune)`, `:46 (pruneTick)` | `a.queries.PruneExpiredLogs` — DELETE on `log_buffer` across all tenants. | Yes — `log_buffer`. | Cron-pool work (cross-tenant enumerate-and-delete is exactly the cron-pool philosophy). **Phase 4 grant addition required:** `cron_user` needs `DELETE` on `log_buffer`, plus default-privileges include `DELETE` on future tenant tables only if a similar pruner appears later. |
| Connector pollers (poll-based) | `connectors/poller.go:102 (run)`, `:112` (`conn.Poll(pollCtx, p.queries)`) | Passes shared `p.queries` into each `PollConnector.Poll`. Each connector (`flyio.go`, `vercel.go`, `railway.go`, `supabase.go`, `mongodb.go`) writes to `log_buffer` via that handle. | Yes — `log_buffer`. | Same handoff shape as syslog: each `Poll` call resolves to one connection / one owning user, so wrap the per-poll body in `UserQueries(ctx, ownerUserID)`. The `Poller` struct must now hold `db.Pools` so each `run()` iteration can mint a fresh user-scoped queries handle. |
| Connector listeners (accept loops) | `connectors/listener.go:71`, syslog listener internal goroutines (`syslog.go:224, 265, 300`) | Same shape — long-lived accept loop with shared queries. | Yes — `log_buffer`. | Per-connection handoff (option (a) from the bearer-token table). |
| Pipeline sweeper (PipelineBus retention) | `agent/agent.go:144 (runPipelineSweeper)` | Need to inspect; `agent.go` shows it as a goroutine but I did not deep-read its body. | Probably no DB writes (memory bus), but **flag for Phase 2 confirmation**. |  Verify-only. If it touches `log_pipeline_events`, fold into the same handoff as the writer. |

---

## §5.3a — Known-user goroutine audit

`grep -nE '^\s*go func' backend/internal backend/cmd --include='*.go'` results,
classified (test-file hits omitted):

| File:line | Purpose | Touches RLS table? | Receives userID? | Disposition |
|---|---|---|---|---|
| `agent/agent.go:128` | `Start` → `Monitor` | Yes (transitively) | No (resolves per-app inside `monitorApp`) | **Phase 2 fix** — see Monitor row above. The fix lives inside `monitorApp`, not at the `go func` site. |
| `agent/agent.go:133` | `Start` → `Prune` | Yes (`log_buffer` DELETE) | n/a (cross-tenant) | Stays cron-pool. Phase 4 grant addition. |
| `agent/agent.go:138` | `Start` → `InvestigationScheduler` | Yes (transitively) | No (resolves per-schedule) | **Phase 2 fix** — fix lives in `RunScheduledInvestigation`. |
| `agent/agent.go:143` | `Start` → `runPipelineSweeper` | Likely no | n/a | Verify in Phase 2; mark "no fix needed" once confirmed memory-only. |
| `agent/monitor.go:77` | per-app monitor goroutine | Yes | No (resolves orgUser inside) | **Phase 2 fix** — fix lives in `monitorApp`. |
| `agent/monitor.go:258` | fire-and-forget notifier dispatch | Yes (notifier writes notification_log) | Yes (`userID` is in scope at call site, but currently NOT passed to `notifier.Notify`) | **Phase 2 fix** — extend `Notify(...)` signature to accept `userID`, thread through. |
| `agent/loop_stream.go:57` | `RunConversationStream` → `runConversationCore` | Yes (every EmitLog, every config read) | Yes (`userID` is the very first parameter) | **Phase 2 fix** — Option A handoff in `runConversationCore`. |
| `agent/scheduler.go:123` | per-schedule goroutine | Yes | No (resolves inside) | **Phase 2 fix** — covered by `RunScheduledInvestigation` handoff. |
| `connectors/poller.go:61` | per-poller goroutine | Yes | Yes (`userID` is captured by `StartPoller`) | **Phase 2 fix** — wrap each `conn.Poll` call in `UserQueries(ctx, ownerUserID)`. |
| `connectors/listener.go:71` | per-listener goroutine | Yes (transitively, via the listener's own `handleConnection`) | Yes for syslog; need to inspect codebase listener if added | **Phase 2 fix** — handoff inside listener body. |
| `connectors/logs/syslog.go:224` | enabled-set refresh ticker | No (reads `app_source_filters` enable list — RLS-protected) | Yes | **Phase 2 fix** — read inside `UserQueries(ctx, s.userID)`. Note: `app_source_filters` IS RLS-protected, so the current refresh path silently breaks under FORCE without the handoff. |
| `connectors/logs/syslog.go:265` | per-connection handler goroutine | Yes (`log_buffer` insert) | Yes (`s.userID` captured at construct time) | **Phase 2 fix** — handoff in `handleConnection`. |
| `connectors/logs/syslog.go:300` | shutdown drain wait | No | n/a | OK. |
| `cmd/heimdall/main.go:123` | HTTP server `ListenAndServe` | No directly (handlers do their own DB) | n/a | OK. |
| `cmd/heimdall/main.go:148` | shutdown coordinator | No | n/a | OK. |

**Summary:** 9 goroutine sites need refactoring, but the fix is concentrated
in 7 functions/methods (the goroutine sites just dispatch into them). The
parent plan's "expected hit count: ~3–8 sites" is at the high end of its
range; the sweep was thorough.

---

## §5.3b — Post-commit context audit (named targets only)

### `chat.go`

**Does NOT match Option A.** Current shape:

- `:67` — pre-flight `s.Queries.GetApplicationByOrgUser` (raw, no UserQueries) for app authz.
- `:128` — `UserQueries` opened, used to set conversation title, committed, closed. (One-shot.)
- `:229` — `setupConversation` opens its own `UserQueries`, commits, closes.
- `:301` — `persistMessages` opens its own `UserQueries`, commits, closes. Called twice per message cycle (before agent run, after agent run).
- `:159` — `s.Agent.RunConversationStream(ctx, userID, ...)` runs the agent loop *outside* any `UserQueries` scope. The loop's many DB reads/writes (`a.queries.GetAppAgentConfig`, `ListConnectionsByApp`, `EmitLog`, etc.) all hit the raw shared `a.queries` handle.

**Phase 2 work:** consolidate to one `UserQueries(ctx, userID)` per inbound
message cycle. Open at the top of the message loop body, pass the resulting
`*db.Queries` down through `setupConversation` (refactored to accept it),
`persistMessages` (refactored), the conversation-title write, AND the
`RunConversationStream` call (refactored to accept a per-call queries
handle rather than reaching for `a.queries`). Commit at the end of the
cycle. The parent plan's 10-tool-iteration constraint applies inside
`runConversationCore`; chat.go's outer loop is the per-message scope.

### `agent/loop.go`

**Does NOT match Option A.** Current shape:

- `runConversationCore` (`:68`) and `RunMonitoring` (`:258`) operate
  exclusively on the shared `a.queries` handle. Every iteration's
  `a.EmitLog` writes to `agent_log` outside any user-scoped transaction.
- Tool dispatch (`a.Dispatch` at `:199` and `:328`) is also called
  outside any user-scoped transaction; whatever the tool does inside its
  own DB calls is on the same shared handle.

**Phase 2 work:** introduce a `loopQueries *db.Queries` parameter on
`runConversationCore` and `RunMonitoring`. Caller (chat.go for
interactive, monitor.go for monitoring, scheduler.go for scheduled)
opens `UserQueries(ctx, userID)` and threads the resulting queries
through. Inside the loop, replace every `a.queries.X` with `loopQueries.X`.
The 10-iteration ceiling stays inside one transaction — that's the
load-bearing Option A invariant. `a.Dispatch` similarly takes a
`*db.Queries` parameter so tool implementations can read/write on the
same scope.

**Risk:** long-running tool calls (LLM round-trips) are *outside* the DB
transaction in wall-clock terms but *inside* it in PostgreSQL terms.
A 2-minute tool dispatch holds an open transaction. Need to verify
`statement_timeout` / `idle_in_transaction_session_timeout` won't
abort. Recommend setting `idle_in_transaction_session_timeout = 0`
inside the transaction (`SET LOCAL`) for the agent loop only, **OR**
keep the existing per-emit short transactions (Option B in the parent
plan) for the agent loop specifically. **Phase 2 design decision** —
flag for the implementing PR.

---

## §4.5 — RLS helper-function audit

```bash
grep -n 'CREATE FUNCTION\|CREATE OR REPLACE FUNCTION' backend/migrations/*.sql
```

| Function | File:line | Mode | Reads RLS table? | Disposition |
|---|---|---|---|---|
| `public.handle_new_user()` | `006_create_users.up.sql:9` | `SECURITY DEFINER` | writes `public.users` (RLS-enabled) | OK as-is. |
| `app_current_user_id()` | `013_enable_rls.up.sql:12` | `LANGUAGE sql STABLE` (default `SECURITY INVOKER`) | reads only `current_setting('app.current_user_id', true)` — no table access | **OK as-is.** Default INVOKER is safe under `app_user` because the function never touches an RLS-protected row. |
| `app_user_org_ids()` | `028_fix_org_members_rls.up.sql:9` | `SECURITY DEFINER` (with `SET search_path = public`) | reads `org_members` | OK as-is. The DEFINER + fixed search_path are the original RLS recursion fix. |

**Disposition:** zero retrofits required. Matches the parent plan's
predicted "possibly zero" estimate exactly.

---

## Phase 2 prerequisite list (consolidated)

Every item below must land before the env-var flip in Phase 6. Each names a
file path; none are aspirational. Items are grouped by structural class.

### Class A — Bearer-token handler refactors

1. **`backend/internal/api/handlers/webhooks.go`** — collapse
   `readWebhookRequest` + `ingestEntries` so the entire transaction
   (token-resolution-aware idempotency check → `routeSources` →
   `insertFiltered` → idempotency cache write) runs inside one
   `UserQueries(ctx, conn.UserID)` scope. Idempotency check before the
   `UserQueries` open is acceptable (read-only) but easier to keep
   inside.
2. **`backend/internal/api/handlers/otlp.go`** — same shape.

### Class B — Background-writer handoff (the §3 pattern)

3. **`backend/internal/agent/monitor.go`** —
   `monitorTick` enumerates active apps via cron pool; `monitorApp`
   resolves owner user (`resolveOrgUser`), opens
   `UserQueries(ctx, ownerUserID)` for the cursor read,
   classify+pipeline+escalate+emit work, and (separately) advances
   `monitoring_state` via cron pool (or via `UserQueries`, either is
   correct now that monitoring_state has both a policy AND a direct
   cron grant). `EmitLog` and `pipeline.Write*` accept a `*db.Queries`.
4. **`backend/internal/agent/scheduler.go`** —
   `schedulerTick` enumerates via cron pool; `RunScheduledInvestigation`
   resolves owner, opens `UserQueries`, threads through `RunMonitoring`,
   `EmitLog*`, `MarkScheduleRun`.
5. **`backend/internal/agent/pipeline_writer.go`** — `Write*` methods
   take `queries *db.Queries` (caller-supplied) instead of holding a
   field. Caller-supplied means each call site already has a
   user-scoped queries handle — webhooks/OTLP from `UserQueries` they
   already opened, monitor from the per-app `UserQueries`.
6. **`backend/internal/agent/emit.go`** — `EmitLog*` take `queries
   *db.Queries`. Or split into `(a *Agent) EmitLog(ctx, q, ...)` that
   takes the handle, leaving the receiver only for severity-mapping logic.
7. **`backend/internal/notifications/notifier.go`** — `Notify` takes
   `userID uuid.UUID` (already resolvable at the call site in
   `monitor.go:261`). Internally opens `UserQueries(ctx, userID)` for the
   prefs/channels reads and the notification_log writes.
8. **`backend/internal/agent/pruner.go`** — stays cron-pool, but the
   migration must add `DELETE` grant on `log_buffer` to `cron_user`.
   Roadmap Phase 4 needs this widening; document in the Phase 4
   completion doc.
9. **`backend/internal/connectors/poller.go`** — `Poll` callback
   signature changes to receive a per-call user-scoped queries handle
   (or the poller's `run` opens `UserQueries(ctx, ownerUserID)` per
   tick and passes the queries through). Touches every connector under
   `backend/internal/connectors/logs/` (flyio, vercel, railway, supabase,
   mongodb) — they currently take `*db.Queries` directly.
10. **`backend/internal/connectors/logs/syslog.go`** — refresh-ticker
    goroutine and per-connection `handleConnection` both need
    `UserQueries(ctx, s.userID)` wrapping their DB writes.
11. **`backend/internal/connectors/listener.go`** — listener manager
    holds connection-shape today; will need access to `db.Pools` so
    listeners can mint per-message `UserQueries`. Constructor signature
    changes ripple through `main.go:104` and `handlers/connections.go:251,428`.

### Class C — Option A in interactive paths (§5.3b)

12. **`backend/internal/api/handlers/chat.go`** — one `UserQueries` per
    inbound message cycle. `setupConversation`, `persistMessages`, and
    `RunConversationStream` all accept a queries handle. The pre-flight
    app-authz lookup at `:67` either moves inside the cycle's
    `UserQueries` or stays out as an explicit "app exists for org" check
    via a SECURITY DEFINER helper (preferred: keep inside `UserQueries`).
13. **`backend/internal/agent/loop.go`** —
    `runConversationCore` and `RunMonitoring` accept
    `loopQueries *db.Queries`. Every internal `a.queries.X` becomes
    `loopQueries.X`. Decision needed: long-running tool dispatches inside
    one transaction (Option A strict) versus per-iteration short
    transactions (Option B). Recommendation: Option A with
    `SET LOCAL idle_in_transaction_session_timeout = 0` for the
    duration of the loop transaction.
14. **`backend/internal/agent/loop_stream.go`** — surface the
    `loopQueries` parameter through `RunConversationStream`.
15. **`backend/internal/agent/tool_*.go` files / `Dispatch`** —
    `Dispatch` accepts `*db.Queries`; every tool implementation that
    does its own DB read uses the supplied handle.

### Class D — Other handler clean-ups

These are JWT-authed handlers using `s.Queries.X` directly (read paths).
None are bearer-token; none are background workers; but all of them hit
RLS-protected tables and will return zero rows under `app_user` + `FORCE`.

The pattern is mechanical: replace each `s.Queries.X(ctx, ...)` read with
`q, _, done, err := s.UserQueries(ctx, userID); defer done(); q.X(ctx, ...)`.

Sites identified in the audit (file:line):

- `applications.go:29, 39, 232, 260, 406, 418, 434, 450`
- `chat.go:67`
- `github_install.go:75`
- `connections_test_handler.go:161` (test file — skip)
- `health.go:13` — pure ping (`s.Pool.Ping`); harmless.
- `investigation_schedules.go:141, 358, 374`
- `notifications.go:26, 113, 332, 386`
- `logs.go:89` — pre-flight authz, same pattern as chat.go:67.
- `organizations.go:42, 263`
- `org_members.go:45, 52, 60, 64, 93, 165, 172`
- `pipeline.go:201, 210, 276, 447, 512, 554` — Pipeline page reads.
  These run on the SSE path which is long-lived; needs Option-A-style
  per-cycle `UserQueries`, or per-poll-tick handoff. **Phase 2
  decision flag.**

The mechanical part should ship as one PR, with an exhaustive grep-based
checklist for review. The Pipeline SSE path deserves its own design
note inside the PR.

### Class E — Schema patches Phase 7 must ship (so Phase 5 tests can target the right state)

These are *not* Phase 2 work but need to be enumerated now so the Phase 5
allowlist files target the right verb sets:

16. Add policy on `agent_config` (likely `FOR ALL USING (true)` —
    documented as defence-in-depth-only).
17. Add policies on `webhook_idempotency` (`connection_id IN (caller's
    connections)`).
18. Add policies on `log_pipeline_events` (`app_id IN (caller's apps)`).
19. Drop and recreate `github_repos_owner` to use the org-member shape
    (not single-user shape).
20. Add `users_org_visible` SELECT policy (or rewrite the org-member
    listing queries through a SECURITY DEFINER helper).

### Class F — Migration A grant additions (Phase 4)

21. `cron_user` needs `DELETE` on `log_buffer` for the pruner.

### Class G — Doc-side prerequisites

22. The roadmap references `docs/executing/rls-enforcement-mental-model.md`
    and `docs/refs/trajan-db-roles.md`. The mental-model doc does not
    exist; the trajan-db-roles doc lives in `docs/executing/`, not
    `docs/refs/`. Either create / move, or update the roadmap's path
    references. **Non-blocking** for code work, but the roadmap header
    will 404 until reconciled.

---

## Acceptance checks (Phase 1 → Phase 2 gate)

- [x] Phase 1 completion doc written and committed (this file).
- [x] §5.0 disposition is "zero hits" — no allowlist required.
- [x] Phase 2 prerequisite list is enumerated, not aspirational — every
      item names a file path or a migration target.

**Risk / blast radius:** zero (read-only audit).
**Effort spent:** in-session review of migrations, handlers, agent, connectors, notifications.

Phase 2 is unblocked.
