# RLS Enforcement — Implementation Roadmap

**Status:** not started — execution playbook.
**Owner:** TBD.
**Parent plan:** [`rls-enforcement-role-split.md`](./rls-enforcement-role-split.md) — the *what* and *why*. This doc is the *how* and *in what order*.
**Companion:** [`rls-enforcement-mental-model.md`](./rls-enforcement-mental-model.md) — conceptual map of which roles show up in which layers.
**Reference precedent:** [`../refs/trajan-db-roles.md`](../refs/trajan-db-roles.md) — Trajan's shipped three-role topology.

---

## How to use this doc

The parent plan is comprehensive but structured for *understanding* the change, not for executing it incrementally. This roadmap re-shapes the same scope into **eight gated phases**, each one:

- Independently reversible (until the final flips in Phase 6 and Phase 7)
- Bounded in scope so a single PR can ship it
- Gated by an explicit set of acceptance checks before the next phase starts
- Mapped back to the parent plan's section numbers (`§5.x`) so the rationale always traces to the source

Work the phases in order. **Do not bundle.** The parent plan's §5 dedicates several paragraphs to why mixing privilege errors with policy errors makes diagnosis materially harder; the same reasoning applies one level up — mixing audit findings with code refactor with role-creation makes attribution of any single failure ambiguous.

After each phase ships, write its completion doc into `docs/completions/rls-enforcement-phase-N.md` with the same shape (what was done, where, why, surprises, follow-ups). The completion docs are the source of truth for the next phase's gating checks.

---

## Phase summary

| # | Phase | Touches | Reversible? | Active effort | Wall-clock |
|---|---|---|---|---|---|
| 1 | **Audit & discover** | Read-only (one doc) | N/A | 1–2 days | 1–2 days |
| 2 | **Code refactor — handoff pattern** | `backend/internal/{api,agent,connectors,notifications}` | Yes (single revert) | 1–2 days | 2–3 days |
| 3 | **Pool wiring (dual pool, single role)** | `backend/cmd/heimdall/main.go`, `handlers.Server`, constructors | Yes | 0.5–1 day | 1 day |
| 4 | **Migration A — role grants** (`039`) | `backend/migrations/`, manual bootstrap | Yes (down-migration) | 0.5 day | 1 day |
| 5 | **Test infrastructure** | `backend/internal/db/`, `backend/internal/api/handlers/`, allowlist files | Yes | 2–3 days | 2–3 days |
| 6 | **Stage 1 — env-var flip to non-superuser** | `.env`, deploy config, prod 48h bake | Yes (env revert) | 0.5 day + 48h bake | 3–4 days |
| 7 | **Migration B — `FORCE RLS` + policy patches** (`040`) | `backend/migrations/`, prod 48h bake | Yes (down-migration) | 0.5 day + 48h bake | 3–4 days |
| 8 | **Cleanup & docs** | `CLAUDE.md`, blueprints, secrets rotation | N/A (doc-only) | 0.5 day | 0.5 day |

**Total active work: 6–9 days. Total wall-clock with bakes: ~2–3 weeks.**

This matches the parent plan's §7 estimate. The numbers shift slightly because some work the parent plan groups under §5.5 ("env-var split + Makefile + code wiring") is split here into Phase 2 (the refactor) and Phase 3 (the pool wiring) — separating the constructor-signature ripple from the per-writer rewrite is what lets each one ship as a tractable PR.

---

## Phase 1 — Audit & discover

**Goal:** produce a written, classified inventory of every code path the role split affects, before any code changes the calling shape. The audit's findings shape every subsequent phase — the migration grants, the refactor scope, the policy allowlist, and the tripwire allowlist all depend on what surfaces here.

**Maps to parent plan:** §5.0, §5.1, §5.2, §5.3, §5.3a, §5.3b (named-target audit only), §4.5.

**Inputs:**
- Parent plan read end-to-end.
- Read access to `backend/`, `frontend/`, all migrations.
- A local Postgres with current production-shape schema (for `pg_policies` / `pg_tables` queries).

**Tasks:**

1. **§5.0 — `service_role` audit (gating).** One grep:
   ```bash
   grep -rn 'service_role\|SUPABASE_SERVICE_ROLE_KEY\|SERVICE_ROLE_KEY' \
     backend/ frontend/ --include='*.go' --include='*.ts' --include='*.vue'
   ```
   For each hit, classify: rewrite-through-`UserQueries` (preferred), or document-as-allowlisted-exception. **If any hit cannot be resolved or allowlisted, Phase 2+ is blocked** — `service_role` bypasses RLS via PostgREST, an entirely different route than the `postgres` superuser connection, and the role split does nothing to protect it.
2. **§5.1 — superuser-only behaviour audit.** Grep `backend/` for: `CREATE EXTENSION`, `ALTER SYSTEM`, `SET ROLE`, `RESET ROLE`, any `auth.users` or `auth.*` reads, any restricted-GUC `SET`. Expected result: empty. Confirm, don't assume.
3. **§5.2 — RLS policy verb coverage audit.** For every public table, run:
   ```sql
   SELECT tablename, array_agg(DISTINCT cmd ORDER BY cmd) AS verbs
   FROM pg_policies
   WHERE schemaname = 'public'
   GROUP BY tablename
   ORDER BY tablename;
   ```
   Cross-reference against `pg_tables.rowsecurity = true`. Flag every table with `rowsecurity = true` and a verb-set that's not `{SELECT, INSERT, UPDATE, DELETE}` (or, for documented append-only tables, the appropriate subset). Special focus on the 0.46.x additions: `connection_sources`, `app_source_filters`, `log_pipeline_events`. Each gap becomes a Phase 7 policy patch.
4. **§5.3 — non-JWT write-path audit.** Two passes:
   - **Bearer-token-authenticated handlers** — every endpoint in `backend/internal/api/handlers/` that uses `Authorization: Bearer` rather than a Supabase JWT. Confirmed candidates: `webhooks.go`, `otlp.go`, the GitHub webhook receiver, the syslog TLS listener (not HTTP, but writes to the same DB). Classify each as "already goes through `UserQueries`" or "writes via raw `Pool.Begin` and relies on superuser bypass."
   - **Background goroutines** — every long-lived writer in `backend/internal/agent/`, `backend/internal/connectors/`, `backend/internal/notifications/`. Confirmed candidates: `agent/monitor.go`, `agent/scheduler.go`, `agent/pipeline_writer.go`, `agent/emit.go`, the Fly.io drain, the Supabase poller, the notifications dispatcher. Classify each as "enumeration-only (cron-pool fits)" or "tenant-write (must hand off to app pool via `UserQueries`)."
5. **§5.3a — known-user background-task audit.** One grep:
   ```bash
   grep -rn '^\s*go func\|go func()' backend/internal backend/cmd --include='*.go'
   ```
   For each `go func`, classify: does it touch an RLS-protected table? If yes, does it receive `userID` and call `UserQueries(ctx, userID)`? Record every site that doesn't — these are silent-zero-row bugs waiting to happen under `app_user`. Expected hit count from the parent plan: ~3–8 sites.
6. **§5.3b — post-commit context audit (named targets only).** Read `backend/internal/api/handlers/chat.go` and `backend/internal/agent/loop.go`. Confirm each one matches **Option A** from the parent plan: one `UserQueries` per prompt cycle, no commit-and-continue inside the cycle, no fresh `UserQueries` per tool iteration. If either violates, log it as a Phase 2 prerequisite. Other call sites are short-lived and structurally guaranteed; don't waste time auditing them.
7. **§4.5 — RLS helper-function audit.** Grep migrations:
   ```bash
   grep -n 'CREATE FUNCTION\|CREATE OR REPLACE FUNCTION' backend/migrations/*.sql
   ```
   For each, check whether the function body references an RLS-protected table and whether `SECURITY DEFINER` is set. Record any `SECURITY INVOKER` (default) helpers — they need a Phase 4 retrofit migration. Expected hit count for Heimdall (per parent plan): possibly zero.

**Output (deliverable):**
`docs/completions/rls-enforcement-phase-1.md` containing:
- §5.0 results: zero hits, OR a numbered list of hits with disposition per hit
- §5.1 results: zero hits expected; any hit is a Phase 2 prerequisite
- §5.2 results: per-table verb coverage matrix; flagged gaps become Phase 7 work items
- §5.3 results: classified call-site list with file:line for every bearer-token handler and every background writer
- §5.3a results: classified `go func` list with file:line; every "missing `userID`" site is a Phase 2 fix
- §5.3b results: chat.go + loop.go disposition (Option A compliant or not)
- §4.5 results: helper inventory with `SECURITY DEFINER` status
- A consolidated **Phase 2 prerequisite list** — every fix that must land before the env-var flip in Phase 6, derived from the audits above

**Acceptance checks (gate to Phase 2):**
- [ ] Phase 1 completion doc written and committed.
- [ ] §5.0 disposition is "zero hits" OR an allowlist exists with justifications.
- [ ] Phase 2 prerequisite list is enumerated, not aspirational — every item names a file path.

**Risk / blast radius:** zero (read-only).
**Estimated effort:** 1–2 days active work, same wall-clock.

---

## Phase 2 — Code refactor (handoff pattern)

**Goal:** rewrite every site Phase 1 surfaced so it complies with the parent plan's `app_user` + `UserQueries` contract, *while still running on the `postgres` superuser pool*. This phase changes calling shape, not privileges. RLS remains cosmetic; the only signal that the refactor is correct is that the existing test suite still passes.

**Maps to parent plan:** §5.3 (rewrite), §5.3a (rewrite), §5.3b (Option A enforcement), §4.5 (helper retrofit if any).

**Inputs:**
- Phase 1 completion doc, specifically the consolidated Phase 2 prerequisite list.
- The named-target file paths from §5.3 and §5.3a.

**Tasks:**

1. **Bearer-token handlers — `webhooks.go`, `otlp.go`, GitHub webhook receiver, syslog listener.** Each one currently (per the parent plan §5.6c) opens `Pool.Begin` directly. Refactor to:
   - Resolve the bearer token to the owning `userID` (already happens elsewhere in the call chain — wire it through).
   - Call `s.UserQueries(ctx, ownerUserID)`.
   - Perform the `INSERT` through the returned `*db.Queries`.
2. **Background writers — `agent/monitor.go`, `agent/scheduler.go`, `agent/pipeline_writer.go`, `agent/emit.go`, connector pollers, notifications dispatcher.** Each one needs the handoff pattern from parent plan §3:
   ```
   cron pool: enumerate (app_id, owner_user_id) for active tenants
   for each tenant:
       UserQueries(ctx, owner_user_id) on app pool
       → all classify / escalate / insert work
       → commit
   ```
   At this phase the cron pool doesn't physically exist yet — Phase 3 introduces it. To avoid coupling, refactor against a `db.Pools` struct that today is constructed with `App: superuser, Cron: superuser` (same pool, twice). Phase 3 splits them.
3. **§5.3a known-user goroutines.** Every site Phase 1 flagged as missing `userID` gets the parameter threaded through and wrapped in `UserQueries(ctx, userID)`. No structural changes — just plumbing.
4. **§5.3b Option A invariant.** Add the doc comment to `UserQueries` declaring the single-transaction lifetime contract. If `chat.go` or `agent/loop.go` violate it (per Phase 1), fix them here. Pattern: one `UserQueries` per prompt cycle, threaded through all 10 tool iterations, committed at the end.
5. **§4.5 helper retrofits (if any).** For each `SECURITY INVOKER` helper Phase 1 surfaced, drop and recreate as `SECURITY DEFINER` in a small companion migration (label it `038a_security_definer_helpers.up.sql` or piggyback onto Migration A in Phase 4 — both work, prefer the smaller scope here).

**Output (deliverable):**
- All listed files refactored.
- A `db.Pools` struct (or equivalent) introduced, today returning the same superuser pool for both `.App` and `.Cron`.
- Phase 2 completion doc: which files changed, which patterns each followed, any structural surprises (e.g. a writer that resisted the handoff shape).

**Acceptance checks (gate to Phase 3):**
- [ ] Every file in Phase 1's prerequisite list now references `UserQueries`.
- [ ] All existing backend tests pass against `DATABASE_URL` = `postgres` (no behaviour change).
- [ ] `agent/monitor.go` still produces `agent_log` and `log_pipeline_events` rows at normal rates in local end-to-end testing.
- [ ] `chat.go` and `agent/loop.go` confirmed Option-A-shaped per the named-target audit.

**Risk / blast radius:** moderate. The refactor is mechanical but wide — the parent plan's §5.5 calls out 20–40 affected files and notes review fatigue as the dominant failure mode. Land as one PR, not many; the constructor-signature change ripples and partial PRs leave the codebase in an unrunnable state.
**Estimated effort:** 1–2 days active work, 2–3 days wall-clock with review.

---

## Phase 3 — Pool wiring (dual pool, single role)

**Goal:** physically introduce the second `*pgxpool.Pool` in `main.go` and thread it through the subsystems Phase 2 prepared, *while both pools still resolve to the same `postgres` superuser*. This phase changes the pool topology without changing privileges. The env-var split is a Phase 6 concern.

**Maps to parent plan:** §5.5 (the wiring half — the env-var flip half is Phase 6).

**Inputs:**
- Phase 2 merged.
- `db.Pools` struct in place from Phase 2.

**Tasks:**

1. **`backend/cmd/heimdall/main.go`** — construct two `*pgxpool.Pool` handles side-by-side:
   - `appPool` from `DATABASE_URL` (today: `postgres`).
   - `cronPool` from `CRON_DATABASE_URL` (today: empty → fall back to `DATABASE_URL` and emit a `WARN` per parent plan §5.5).
   - Pool sizing per parent plan §3: app ~30, cron ~5.
2. **`handlers.Server`** — already holds the app pool; add the cron pool. Expose it only on the path that constructs background subsystems; do not pass it into HTTP request handlers (none of them need cross-tenant enumeration).
3. **Constructor ripple** — `agent.Agent`, the scheduler, connector pollers, notifications dispatcher each take `db.Pools` instead of a single facade. The interactive chat path uses `.App` only; the monitoring loop uses both.
4. **Production invariant** — at startup, refuse to launch if `DATABASE_URL` and `CRON_DATABASE_URL` resolve to the same role *in production builds*. Today both resolve to `postgres`, so the invariant must be guarded by a build flag or env (`HEIMDALL_ENV=production`) — otherwise dev breaks. Alternative: gate the invariant on "neither URL contains `postgres@`," which is brittle but env-flag-free. Pick the cleaner option at implementation time.
5. **`.env.example`, `README`, `CLAUDE.md` setup** — add `CRON_DATABASE_URL` and `DIRECT_URL` as documented variables, with default values that mirror `DATABASE_URL` for dev. The values stay identical until Phase 6.
6. **`Makefile`** — update `migrate-up`, `migrate-down`, `migrate-create`, and any `sqlc-generate` target that touches the DB to read `DIRECT_URL` instead of `DATABASE_URL`. Today `DIRECT_URL` is identical to `DATABASE_URL`, so no behaviour changes; the rename pre-positions the migrations to keep working when Phase 6 splits them.

**Output (deliverable):**
- Two pools in `main.go`, both resolving to `postgres`.
- `db.Pools` threaded through every constructor that Phase 1 identified as a background-writer subsystem.
- Updated docs and Makefile.
- Phase 3 completion doc.

**Acceptance checks (gate to Phase 4):**
- [ ] `make dev-backend` boots with both pools constructed.
- [ ] All backend tests still pass (no behaviour change expected).
- [ ] Pool sizes match parent plan §3 (app ~30, cron ~5).
- [ ] Production-mode startup refuses to launch when both URLs are identical (verified locally with a fake `HEIMDALL_ENV=production` boot).
- [ ] `make migrate-up` still works via the renamed `DIRECT_URL`.

**Risk / blast radius:** low. The pools resolve to the same role, so no privilege change. The risk is purely "did we wire the constructors correctly" — caught by the existing test suite plus a single local boot.
**Estimated effort:** 0.5–1 day active work.

---

## Phase 4 — Migration A (role grants)

**Goal:** create the privilege topology in the database. No traffic flips yet — `DATABASE_URL` and `CRON_DATABASE_URL` still point at `postgres`. After this phase, both runtime roles exist and have the correct grants; pointing the env vars at them in a *staging* DB would work, but Heimdall has no staging, so the env-var flip is deferred to Phase 6.

**Maps to parent plan:** §4.1, §4.2, §5.4 (Migration A specifically), §5.6d (the conditional PG17 grant block lives in Migration A).

**Inputs:**
- Phase 1 completion doc — confirms §4.5 helper-function status and §5.2 policy gaps (the latter is Migration B / Phase 7's problem, but Phase 1 must have surfaced them).
- DBA / operator access to run the manual bootstrap step in production.

**Tasks:**

1. **Manual bootstrap (per environment, before Migration A runs).** Run in Supabase SQL editor or `psql` as `postgres`:
   ```sql
   CREATE ROLE app_user LOGIN PASSWORD '<strong-generated>';
   CREATE ROLE cron_user LOGIN PASSWORD '<strong-generated-distinct>' BYPASSRLS;
   ```
   Per the parent plan §3 / Trajan precedent, role creation with passwords is environment bootstrap, not migration material. Bootstrap once per environment (local, prod). Passwords go into the secrets manager; no committed `.env` file gets them.
2. **Migration `039_runtime_role_grants.up.sql`** — assumes the bootstrap has run and **fails loudly if not**:
   ```sql
   DO $$
   BEGIN
     IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
       RAISE EXCEPTION 'app_user role missing; run the manual bootstrap from rls-enforcement-roadmap.md Phase 4 before applying this migration';
     END IF;
     IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cron_user') THEN
       RAISE EXCEPTION 'cron_user role missing; run the manual bootstrap from rls-enforcement-roadmap.md Phase 4 before applying this migration';
     END IF;
   END $$;
   ```
   Then the §4.1 grants for `app_user` (CRUD on all `public.*` tables, USAGE+SELECT on sequences, default-privileges for future tables), the §4.2 grants for `cron_user` (USAGE on `public`, blanket SELECT, narrow `INSERT, UPDATE` on `monitoring_state` only, default-privileges scoped to SELECT only for future tables).
3. **PG17 conditional grant block** (parent plan §5.6d):
   ```sql
   DO $$
   BEGIN
     IF current_setting('server_version_num')::int >= 170000 THEN
       EXECUTE 'GRANT app_user TO postgres WITH SET TRUE';
       EXECUTE 'GRANT cron_user TO postgres WITH SET TRUE';
     ELSE
       EXECUTE 'GRANT app_user TO postgres';
       EXECUTE 'GRANT cron_user TO postgres';
     END IF;
   END $$;
   ```
   Required for the Phase 5 `asAppUser` test fixture; harmless on production-mode connections (which connect as `app_user` directly and never run `SET ROLE`).
4. **`SECURITY DEFINER` retrofit** — if Phase 1's §4.5 audit surfaced any helpers, re-declare them `SECURITY DEFINER` here and `GRANT EXECUTE ... TO app_user, cron_user`.
5. **Migration `039_runtime_role_grants.down.sql`** — revoke grants in reverse order. **Do not** `DROP ROLE` (role lifecycle stays manual, mirrors bootstrap).

**Output (deliverable):**
- Migration 039 up/down committed.
- Bootstrap SQL recorded in the Phase 4 completion doc verbatim, including the prod password handoff procedure (who runs it, how the passwords get into the secrets manager).
- Phase 4 completion doc.

**Acceptance checks (gate to Phase 5):**
- [ ] Bootstrap SQL has run in production and local; both roles exist with the intended attributes.
- [ ] `make migrate-up` (now using `DIRECT_URL`) applies 039 cleanly.
- [ ] `make migrate-down` reverses 039 cleanly (rehearsed locally).
- [ ] `pg_roles` shows `app_user` with `rolbypassrls = false`, `rolsuper = false`; `cron_user` with `rolbypassrls = true`, `rolsuper = false`.
- [ ] `information_schema.role_table_grants` shows `cron_user` with `INSERT/UPDATE` only on `monitoring_state` (plus blanket `SELECT`).
- [ ] `DATABASE_URL` and `CRON_DATABASE_URL` still point at `postgres` (we are not flipping yet).

**Risk / blast radius:** very low. Migration is purely additive — no `REVOKE`, no `FORCE RLS`, no privilege removal. Worst-case failure: migration aborts because the bootstrap didn't run, which is the failure mode the assertion is designed for.
**Estimated effort:** 0.5 day active work, 1 day wall-clock (bootstrap coordination).

---

## Phase 5 — Test infrastructure

**Goal:** add every test the parent plan §5.6 / §5.6a / §5.6b / §5.6c / §5.6d enumerates, *before* the env-var flip in Phase 6. The tests pass green today (running as `postgres`, RLS cosmetic) for the behaviour tests, and pin the catalog state for the drift tests. After Phase 6's flip, the behaviour tests will start *meaningfully* asserting RLS enforcement.

**Maps to parent plan:** §5.6, §5.6a, §5.6b, §5.6c, §5.6d.

**Inputs:**
- Phase 4 merged (so `app_user` and `cron_user` exist for the catalog tests to query).
- Phase 1 completion doc (so the §5.6c `rawDBAccessAllowlist` and §5.6a `rlsPolicyAllowlist` start with informed initial values).

**Tasks:**

1. **§5.6 RLS regression tests** — one per critical table (`connections`, `log_buffer`, `conversations`, `connection_sources`, `app_source_filters` minimum), each one asserting cross-tenant read returns `pgx.ErrNoRows` *under `app_user`*. Tests run via the §5.6d `asAppUser` fixture so they exercise enforcement even though `DATABASE_URL` is still `postgres`.
2. **§5.6 cron-role behaviour tests** — four tests against a cron-pool connection: enumerate-across-tenants succeeds, `DROP TABLE` denied (42501), `auth.users` read denied (42501), blanket tenant-table INSERT denied (42501).
3. **§5.6a catalog drift checks** — five Go tests wrapping SQL queries:
   - Test 1: every `rowsecurity = true` table is `forcerowsecurity = true`. (Will fail until Phase 7; gate Phase 7's accept-check on this turning green.)
   - Test 2: per-table verb coverage matches the allowlist in `backend/internal/db/rls_policy_allowlist.go`. (Initial allowlist content from parent plan §5.6a, refined by Phase 1's §5.2 findings.)
   - Test 3: `app_user` does not have `rolbypassrls = true`.
   - Test 4: `cron_user` does not have `rolsuper = true`.
   - Test 5: `cron_user`'s writable-table set in `information_schema.role_table_grants` matches `cronWriteAllowlist` exactly (initial: `{"monitoring_state": {"INSERT", "UPDATE"}}`).
4. **§5.6b connectivity tests** — `TestCronPool_IsActuallyCronUser`, `TestAppPool_IsActuallyAppUser`. Both will *fail today* because Phase 6 hasn't flipped env vars; mark them with a build tag or `t.Skip` gated on an env var (`HEIMDALL_ROLE_SPLIT_LIVE=1`) so they only run post-flip. Better: write them now, let them fail in CI with a clear message, and unblock them in Phase 6.
5. **§5.6c source-tree pairing tripwire** — `backend/internal/db/rls_pairing_test.go`. Walks every `.go` file in `backend/internal/` and `backend/cmd/`, asserts every file using raw DB access (`Pool.Begin`, raw `*db.Queries` in long-lived agent/notifications/connectors paths) either references `UserQueries` or is in `rawDBAccessAllowlist`. Initial allowlist content lifted from parent plan §5.6c, refined by Phase 1's §5.3 + §5.3a findings.
   - Second sub-test: every allowlist entry still exists and still matches the raw-access pattern it claims to justify (catches stale-allowlist drift).
6. **§5.6d opt-in `asAppUser` fixture** — wrapper in `testhelpers_test.go` that does `SET LOCAL ROLE app_user` + `set_config('app.current_user_id', $1, true)` from a `postgres`-authenticated test transaction. Unblocks the §5.6 RLS regression tests above.

**Output (deliverable):**
- All test files written, committed, and green in CI (modulo the connectivity tests which are gated on Phase 6).
- `backend/internal/db/rls_policy_allowlist.go` exists and is the source of truth.
- `backend/internal/db/raw_db_access_allowlist.go` (or inline in the test file, matching parent plan §5.6c shape) exists.
- Phase 5 completion doc.

**Acceptance checks (gate to Phase 6):**
- [ ] All §5.6 RLS regression tests pass via the `asAppUser` fixture.
- [ ] All four §5.6 cron-role behaviour tests pass.
- [ ] §5.6a Test 1 (`FORCE` coverage) is **failing or skipped** — that's expected; Phase 7 makes it green.
- [ ] §5.6a Tests 2–5 pass.
- [ ] §5.6b connectivity tests are written and gated on the post-Phase-6 env.
- [ ] §5.6c pairing tripwire is green — every raw-DB-access file is in the allowlist with a justification.
- [ ] §5.6d `asAppUser` fixture is in `testhelpers_test.go` and used by at least one test per critical table.

**Risk / blast radius:** zero (test-only).
**Estimated effort:** 2–3 days active work. The pairing tripwire and the allowlist files are the long-tail items.

---

## Phase 6 — Stage 1: env-var flip to non-superuser

**Goal:** point `DATABASE_URL` at `app_user` and `CRON_DATABASE_URL` at `cron_user` in production. RLS is still cosmetic at this point (`FORCE` lands in Phase 7), so any failure surfaced here is a *privilege-set* issue, not a *policy* issue. Diagnostically clean.

**Maps to parent plan:** §5.5 (the env-var flip half), §5.7 (Stage 1).

**Inputs:**
- Phases 1–5 all merged.
- Phase 1's prerequisite list is exhaustively addressed (every file fixed, every `go func` audited, every helper retrofitted).
- 48-hour bake window committed on the calendar.
- The named revert lever (parent plan §5.7 last subsection) practiced locally.

**Tasks:**

1. **Local end-to-end rehearsal.** Point local `DATABASE_URL` and `CRON_DATABASE_URL` at `app_user` and `cron_user` respectively. Manually trigger every path in Phase 1's §5.3 audit — webhook ingestion, OTLP, syslog, GitHub webhook, a scheduled investigation cycle, a full monitor-loop iteration. Anything that fails locally with `42501 permission denied` is a Phase 6 prerequisite that escaped Phase 1.
2. **Production env-var flip.** Update prod env (Fly.io secrets, or whatever the deploy mechanism is) so `DATABASE_URL` → `app_user`, `CRON_DATABASE_URL` → `cron_user`, `DIRECT_URL` → `postgres`. Redeploy.
3. **Watch the production logs for 48 hours.** Specifically:
   - `42501 permission denied` errors → these are §5.1 misses (DDL or `auth.*` reads under `app_user`).
   - Sudden drop in `agent_log` row rate → a background writer is going through `cron_user` (no INSERT grant on tenant tables) instead of handing off.
   - Sudden drop in `log_pipeline_events` row rate → same shape, different writer.
   - SSE ticker / Pipeline page going dark → pipeline writer is mis-routed.
   - Webhook / OTLP / syslog / GitHub ingestion failing → bearer-token handler is still on raw pool.
4. **Unblock §5.6b connectivity tests.** Once prod is on the new pools, the `HEIMDALL_ROLE_SPLIT_LIVE=1` gate goes from off to on; the tests now meaningfully assert "the cron pool really is `cron_user`."
5. **Acceptance bake.** 48h with no anomalies → proceed to Phase 7. Anomaly within the 48h → execute the named revert lever (env vars back to `postgres`, redeploy, <60 seconds), debug, re-Phase-1 the issue, re-flip when fixed.

**Output (deliverable):**
- Production running on `app_user` + `cron_user`.
- 48h of clean logs documented in the Phase 6 completion doc, including throughput metrics for `agent_log` / `log_pipeline_events` showing no regression.
- Phase 6 completion doc.

**Acceptance checks (gate to Phase 7):**
- [ ] Zero `42501 permission denied` errors over the 48h prod bake.
- [ ] `agent_log` and `log_pipeline_events` throughput within ±5% of the pre-flip baseline.
- [ ] Webhook / OTLP / syslog / GitHub ingestion confirmed working under the new role via prod traffic.
- [ ] §5.6b connectivity tests passing in CI.
- [ ] Migration B's down-migration rehearsed locally before Phase 7 starts.

**Risk / blast radius:** moderate. RLS still cosmetic, so worst-case is privilege denial (not data leak). Strictly better posture than today (no DDL, no `auth.*`, no role escalation). The named revert lever is a <60-second flip back.
**Estimated effort:** 0.5 day active work (rehearsal + flip), 48h elapsed for the bake.

---

## Phase 7 — Migration B (`FORCE RLS` + policy patches)

**Goal:** flip RLS from cosmetic to enforced for `app_user` by `FORCE`-ing it on every user-scoped table, and ship any missing INSERT/UPDATE/DELETE policies surfaced in Phase 1's §5.2. After this phase, the role split is structurally complete and the parent plan's §2 threat model is closed against `app_user`.

**Maps to parent plan:** §4.3, §5.2 (the policy patches half), §5.4 (Migration B specifically), §5.7 (Stage 2).

**Inputs:**
- Phase 6 baked clean for 48h.
- §5.6a Test 2 (verb-set allowlist) has been updated with the verb-set Phase 7 will produce — i.e. the allowlist is the **target state**, not the pre-Phase-7 state. Phase 7 makes the test go green.
- Phase 1 completion doc lists every policy gap; the migration patches each one.

**Tasks:**

1. **Migration `040_rls_force_enforcement.up.sql`:**
   - `ALTER TABLE … FORCE ROW LEVEL SECURITY` on every table from parent plan §4.3, validated against `pg_tables.rowsecurity = true` at migration-write time (don't trust the parent plan's enumeration from memory; read the catalog).
   - Add the missing INSERT/UPDATE/DELETE policies Phase 1 surfaced. Policy bodies follow the existing pattern: scope by `current_setting('app.current_user_id', true)` via the User → Org → App → row join.
2. **Migration `040_rls_force_enforcement.down.sql`** — drop `FORCE` on each table, revert each policy patch. Rehearse locally before deploy (parent plan §5.7 explicitly calls this out).
3. **Production deploy.** `make migrate-up` (via `DIRECT_URL`). The migration touches only catalog state; no app downtime.
4. **Watch the production logs for 48 hours:**
   - `new row violates row-level security policy` errors → §5.2 miss (a table with `FORCE` but no INSERT/UPDATE/DELETE policy) or §5.3 miss (a writer not going through `UserQueries`). Both are immediate fix-or-revert decisions.
   - Background subsystem regressions → `cron_user` is unaffected (BYPASSRLS), so any monitor-loop or scheduler regression here is a refactor miss. Fix the writer, not the policy.
5. **§5.6a Test 1 (`FORCE` coverage)** — should now go green. Verify in CI.
6. **Acceptance bake.** 48h with no anomalies → declare the rollout complete and proceed to Phase 8.

**Output (deliverable):**
- Migration 040 up/down committed and applied in prod.
- §5.6a Test 1 green.
- 48h of clean logs in the Phase 7 completion doc.
- Phase 7 completion doc.

**Acceptance checks (gate to Phase 8):**
- [ ] Zero `new row violates row-level security policy` errors over the 48h prod bake.
- [ ] §5.6a Test 1 (every RLS-enabled table is `FORCE`'d) passes in CI.
- [ ] Throughput metrics for `agent_log` / `log_pipeline_events` still within ±5% of pre-flip baseline.
- [ ] Manual cross-tenant probe from local: open a `UserQueries` as `userB`, attempt to read a `userA`-owned row → `pgx.ErrNoRows`. Re-asserts the parent plan §2 threat model is closed.

**Risk / blast radius:** moderate. The migration is small but the failure mode (silent zero-row reads, write-policy violations) is exactly the class Phase 6 was designed to surface in advance. Down-migration is a single `make migrate-down` and brings the system back to "Stage 1 clean" without redeploy.
**Estimated effort:** 0.5 day active work, 48h elapsed for the bake.

---

## Phase 8 — Cleanup & docs

**Goal:** rotate secrets, update every doc that asserted "we connect as superuser," and cross-link this rollout from the Trajan reference.

**Maps to parent plan:** §5.8.

**Inputs:**
- Phase 7 baked clean for 48h.
- Secrets manager access for password rotation.

**Tasks:**

1. **Password rotation.** Rotate `postgres`, `app_user`, `cron_user` passwords to long random values, distinct from each other. Store in the secrets manager only — no committed `.env*`.
2. **`CLAUDE.md`** — update setup instructions to include `DIRECT_URL` and `CRON_DATABASE_URL`, with one-line description per role.
3. **`docs/blueprints/database-connection-blueprint.md`** — §2 (endpoint table: add all three roles), §4 (the `backend-blueprint.md:517` note becomes stale; rewrite both), §8 (remove "we connect as superuser" premise; describe the two-runtime-pools-plus-migrations-superuser topology).
4. **`docs/blueprints/backend-blueprint.md:517`** — the original "we connect as superuser" note. Rewrite to reflect the new posture; cross-link to this roadmap and to the Phase 7 completion doc.
5. **Cross-link `docs/refs/trajan-db-roles.md`** — add a back-reference to this roadmap as Heimdall's analogue. Add a forward-reference from this roadmap.
6. **Cross-link `docs/executing/db-connection-type-improvement.md`** from the blueprint's §8 as the 5432→6543 follow-up.
7. **Move `rls-enforcement-role-split.md` and `rls-enforcement-roadmap.md` to `docs/archive/`** once Phase 8 ships — they describe shipped work, not in-flight work. Keep `rls-enforcement-mental-model.md` in `docs/refs/` as evergreen.

**Output (deliverable):**
- Rotated secrets.
- Updated `CLAUDE.md` and both blueprints.
- Phase 8 completion doc — final rollout summary, links to all eight phase completion docs.

**Acceptance checks (rollout complete):**
- [ ] All eight completion docs exist in `docs/completions/rls-enforcement-phase-N.md`.
- [ ] Parent plan §9 (the acceptance checklist) is fully ticked.
- [ ] `CLAUDE.md` references the new topology.
- [ ] Both blueprints reflect the three-role topology.
- [ ] Original parent plan and this roadmap moved to `docs/archive/`.

**Risk / blast radius:** zero (doc + secrets only).
**Estimated effort:** 0.5 day.

---

## Cross-phase invariants

These hold throughout the rollout — they are not a phase, but a checklist to verify before any phase ships.

- **No phase bundles two failure classes.** Privilege errors (Phase 6) and policy errors (Phase 7) ship in distinct phases for diagnostic clarity. If a phase ever feels like it's doing two things, split it.
- **Every phase is reversible until Phase 6.** Phases 1–5 touch only docs, code, tests, and additive grants. Phase 6 is the first irreversible-by-redeploy step (it's still reversible by env-var flip). Phase 7 is reversible by down-migration. Phase 8 is doc-only.
- **Every phase produces a completion doc.** The completion doc is the gating evidence for the next phase, not the PR description.
- **The parent plan is the source of truth on rationale.** This roadmap's section numbers (§5.x) trace back to the parent plan. If a phase deviates, update the parent plan, don't fork the rationale.
- **Bake windows are non-negotiable.** No staging environment means production traffic is the exercise surface. Compressing a 48h bake to ship faster reintroduces the exact failure mode the bake exists to catch.

---

## Where to start

Begin with **Phase 1**. Output goes to `docs/completions/rls-enforcement-phase-1.md`. Every subsequent phase is gated on Phase 1's findings, so no other phase has correct inputs without it.
