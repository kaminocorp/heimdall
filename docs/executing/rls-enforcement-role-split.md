# RLS Enforcement via DB Role Split

**Status:** not started — planning doc.
**Owner:** TBD.
**Scheduling:** before any enterprise-shape customer onboarding. Not a code-red item today (no known active leak) but the defensive posture is one application bug away from a multi-tenant data leak, which is not where we want to be when selling "enterprise-grade."

---

## 1. The gap in one sentence

The backend connects to Supabase Postgres as the `postgres` **superuser**, which is `BYPASSRLS` by default, so every Row-Level Security policy attached to `applications`, `connections`, `log_buffer`, `conversations`, `connection_sources`, `app_source_filters`, `log_pipeline_events`, and every other user-scoped table is **compiled but not enforced** against our own backend. The `UserQueries` helper dutifully runs `SET LOCAL app.current_user_id = $1` inside every transaction, but the policies that read `current_setting('app.current_user_id', true)` never actually execute.

Confirmed in `docs/blueprints/backend-blueprint.md:517`:

> The backend connects as the `postgres` superuser, which bypasses RLS by default. The policies protect against non-owner access paths (Supabase dashboard roles, PostgREST, direct `psql` with other roles). The `SET LOCAL` plumbing is in place so that a migration to a non-owner app role activates RLS enforcement automatically.

The `SET LOCAL` plumbing was written with this migration in mind. We just haven't done the migration.

---

## 2. Why it matters — threat model

This is **not** currently a data leak. It becomes one under three scenarios, all of which grow more likely as the codebase grows:

| Scenario | Current (superuser) | With non-superuser `app_user` |
|---|---|---|
| **App code is perfect** — every query filters by user/org correctly | Safe. RLS is cosmetic. | Safe. RLS is belt-and-suspenders-plus-enforcement. |
| **App code has a bug** — a future handler, sqlc query, or AI-generated patch forgets `WHERE user_id = $1` or uses a join that crosses orgs | **Ships as a cross-tenant data leak.** Nothing catches it in CI, staging, or prod. | RLS catches it. The bad query returns zero rows, surfacing as a test failure or visible bug before data is exposed. |
| **SQL injection** anywhere in the stack — dynamic ORDER BY, string concat in a future connector, a bug in a raw-SQL code path | **Instant game over.** `DROP TABLE`, `SELECT FROM auth.users`, `SET ROLE`, `ALTER SYSTEM`. | Contained to `app_user`'s privilege set. No DDL, no auth schema, no role escalation. |

Scenarios #2 and #3 are the real risks. The 0.46.x source-filtering work added `connection_sources`, `app_source_filters`, and (in flight) `log_pipeline_events` — three new user-scoped write surfaces in as many weeks. The pace at which we're adding cross-tenant write paths is faster than the pace at which any single reviewer can mentally verify every predicate. RLS-as-enforcement is the mechanical backstop that lets the codebase scale without every new query being a potential breach.

A related multiplier: this codebase is maintained partly with AI assistance. An AI-generated sqlc query that drops a `WHERE user_id = $1` clause looks structurally identical to a correct one on review. RLS enforcement is the one layer that doesn't depend on a reviewer (human or AI) noticing the missing predicate.

---

## 3. The target topology

Three URLs, three roles, clear separation of responsibilities. Two non-superuser runtime roles (one RLS-enforced, one BYPASSRLS) plus a superuser retained for migrations only:

```
DATABASE_URL       → app_user:<pw>@<host>:5432/postgres
                       └─ Non-superuser, RLS-ENFORCED.
                       └─ Used by the Go HTTP server: handlers, WebSocket chat, UserQueries.
                       └─ Cannot DROP, cannot ALTER SYSTEM, cannot read auth.*, cannot SET ROLE, cannot BYPASSRLS.
                       └─ Receives ~all user-initiated traffic.

CRON_DATABASE_URL  → cron_user:<pw>@<host>:5432/postgres
                       └─ Non-superuser, BYPASSRLS (deliberate, scope-limited).
                       └─ Used ONLY by background work that must see across tenants:
                            - agent/monitor.go (the 15s monitoring loop, enumeration step)
                            - scheduled investigations (cron-shape scheduler)
                            - connector pollers (Fly.io drain, Supabase poller, etc.)
                            - pg_try_advisory_lock calls for cross-replica dedupe
                       └─ Enumerates and locks. NEVER writes tenant data directly.
                         Once it has resolved (app_id, owner_user_id), it hands off to a
                         UserQueries session on the app_user pool for the actual mutation.
                       └─ Same denied grants as app_user (no DDL, no auth.*, no SUPERUSER).

DIRECT_URL         → postgres:<pw>@<host>:5432/postgres
                       └─ Superuser, retains DDL rights.
                       └─ Used ONLY by golang-migrate (`make migrate-up`, `make migrate-down`).
                       └─ Never read by the long-running backend process.
```

This is the Supabase-recommended production pattern. The two-role-plus-migrations split is what Prisma, Drizzle, and most Supabase-native stacks ship with; the `CRON_DATABASE_URL` addition is what multi-tenant SaaS with scheduler workloads (e.g. Elephantasm, any codebase with APScheduler-shaped jobs) converge on independently. The `DIRECT_URL` vs. `DATABASE_URL` naming matches the convention those ORMs use, so it's recognisable to anyone onboarding.

### Why a third role and not "just let the monitor loop use DIRECT_URL"

Tempting, and wrong. The monitor loop is a long-running goroutine inside the app process; handing it a superuser connection re-creates the exact blast radius the role split is trying to eliminate. A SQL injection in a scheduler query, or a future AI-authored scheduled investigation that builds a dynamic `ORDER BY`, would have `DROP TABLE` / `ALTER SYSTEM` / `auth.*` access *and* the ability to run for hours without a request boundary to time it out. `cron_user` gives the scheduler exactly what it needs (cross-tenant `SELECT`, advisory locks) and nothing else.

### Why not `app_user` for the scheduler

Also tempting, also wrong. `app_user` is RLS-enforced; under FORCE RLS it returns zero rows to any query without `app.current_user_id` set. The monitor loop's enumeration step (`SELECT app_id, owner_user_id FROM applications WHERE monitoring_enabled`) is cross-tenant *by construction* — there is no single user whose identity would scope it correctly. `cron_user` with BYPASSRLS is the narrowest role that can answer that query.

### The handoff pattern (load-bearing)

`cron_user` is for **enumeration and coordination**, not tenant writes. Every write to a user-scoped table must go through `app_user` with `SET LOCAL app.current_user_id` set. In practice:

```
cron_engine.Query("SELECT app_id, owner_user_id FROM applications WHERE monitoring_enabled")
  → for each (app_id, owner_user_id):
      UserQueries(ctx, owner_user_id)            // app_user pool
        → SELECT logs, classify, escalate, INSERT INTO agent_log + log_pipeline_events
        → commit
```

The `cron_user` connection should almost never `INSERT` / `UPDATE` / `DELETE` against `public.*`. If a future contributor reaches for the cron pool to do a write, that's a design smell — the right question is "which user_id should this write be scoped to?" Enforced by convention today, enforceable by schema-level `REVOKE INSERT/UPDATE/DELETE ON public.* FROM cron_user` if we ever want to make it mechanical.

### Port 5432, not 6543

This plan keeps all three URLs on **port 5432 (direct / session mode)**. Moving `DATABASE_URL` and `CRON_DATABASE_URL` to Supavisor on 6543 is a separate project with its own blockers (prepared-statement cache behaviour, future `LISTEN/NOTIFY`, session-scoped GUCs on user-owned Postgres connectors) — see `docs/blueprints/database-connection-blueprint.md` §8 and the companion plan at `docs/executing/db-connection-type-improvement.md`. Don't bundle that migration into this one; it roughly triples the blast radius and mixes two unrelated failure modes (role privileges vs. pooler session semantics).

---

## 4. What each role needs (and doesn't)

### 4.1 `app_user` — the RLS-enforced runtime role

```sql
CREATE ROLE app_user LOGIN PASSWORD '<strong-generated>';
-- Deliberately NO: SUPERUSER, BYPASSRLS, CREATEDB, CREATEROLE, REPLICATION.

GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- Apply automatically to future tables/sequences created by migrations.
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO app_user;
```

### 4.2 `cron_user` — the BYPASSRLS enumeration role

```sql
CREATE ROLE cron_user LOGIN PASSWORD '<strong-generated-distinct-from-app_user>' BYPASSRLS;
-- Deliberately NO: SUPERUSER, CREATEDB, CREATEROLE, REPLICATION.
-- BYPASSRLS is the whole point, but it is the ONLY elevated attribute.

GRANT USAGE ON SCHEMA public TO cron_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cron_user;

-- Optional tightening (recommended): the cron role should almost never write.
-- Grant narrow write access only on the specific infra tables the scheduler
-- owns end-to-end (e.g. monitoring_state cursor updates, advisory-lock helpers).
-- Do NOT grant blanket INSERT/UPDATE/DELETE on public.*.
GRANT INSERT, UPDATE ON monitoring_state TO cron_user;
-- Add further table-specific write grants here as the scheduler's needs grow;
-- keep the list auditable. Every new grant is a potential cross-tenant write
-- path and should be reviewed at the same bar as a new RLS policy.

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO cron_user;

-- Future-table defaults: SELECT only. Force explicit opt-in for writes.
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO cron_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO cron_user;
```

The narrow-write grants matter. If `cron_user` can `INSERT` into any tenant table, a bug in the scheduler can still produce cross-tenant writes — BYPASSRLS silences the RLS policy that would otherwise catch it. By keeping writes scoped to infra-shape tables (`monitoring_state`, future scheduler-metadata tables) and forcing every write to tenant data through `app_user` under `UserQueries`, we preserve the "RLS is the backstop" property for the overwhelming majority of write volume.

### 4.3 FORCE RLS on every user-scoped table

Without `FORCE`, the table *owner* still bypasses RLS. Supabase's table ownership model varies across migrations, so rely on `FORCE` rather than assuming ownership. Note: `FORCE` does NOT apply to roles with the `BYPASSRLS` attribute — that's `cron_user`'s whole design, and the §4.2 narrow-write grants are what keep it safe. RLS enforcement remains real against `app_user` (which is where 99%+ of write volume flows) and cosmetic only for `cron_user` (which writes to exactly the tables you've explicitly allowed).

```sql
ALTER TABLE users                 FORCE ROW LEVEL SECURITY;
ALTER TABLE organizations         FORCE ROW LEVEL SECURITY;
ALTER TABLE applications          FORCE ROW LEVEL SECURITY;
ALTER TABLE connections           FORCE ROW LEVEL SECURITY;
ALTER TABLE connection_sources    FORCE ROW LEVEL SECURITY;
ALTER TABLE app_source_filters    FORCE ROW LEVEL SECURITY;
ALTER TABLE log_buffer            FORCE ROW LEVEL SECURITY;
ALTER TABLE conversations         FORCE ROW LEVEL SECURITY;
ALTER TABLE reports               FORCE ROW LEVEL SECURITY;
ALTER TABLE agent_log             FORCE ROW LEVEL SECURITY;
ALTER TABLE app_agent_config      FORCE ROW LEVEL SECURITY;
ALTER TABLE monitoring_state      FORCE ROW LEVEL SECURITY;
ALTER TABLE log_pipeline_events   FORCE ROW LEVEL SECURITY;
-- ...plus any user-scoped table added in the meantime.
```

This list must be exhaustive at the time of the migration. Treat `pg_tables WHERE schemaname='public'` as the source of truth and cross-check every entry.

### 4.4 What must NOT be granted (to either runtime role)

- `BYPASSRLS` on `app_user` (the whole point is that it can't)
- `SUPERUSER` on either role
- `CREATE` on schema `public` on either role (prevents rogue table creation under a runtime role)
- Any access to `auth.*`, `storage.*`, `extensions.*`, or other Supabase internal schemas
- `CREATE EXTENSION`, `ALTER SYSTEM`, or anything that reshapes the server
- `CREATEDB`, `CREATEROLE`, `REPLICATION` on either role

`cron_user` has `BYPASSRLS` — that single attribute is the entire reason the role exists, and its scope is contained by the narrow-write grants in §4.2. Anything else elevated is out of bounds.

---

## 5. Execution plan (Phase 1)

Order matters — each step is independently reversible until the final flip.

### 5.1 Audit every query for superuser-only behaviour

Run the backend test suite and a representative prod-shape workload against `app_user` in a staging database **before** flipping the env var. Specifically look for:

- `CREATE EXTENSION`, `ALTER SYSTEM`, `SET` of restricted GUCs — these will fail hard under `app_user`.
- Dependency on `auth.users` reads — these will fail, and anywhere they appear we probably want to cache the relevant columns in `public.users` instead.
- Any `SET ROLE` or `RESET ROLE` — these will fail.
- Extension-owning or catalog writes from app code (should be zero).

Expected outcome: nothing in the app needs superuser today. Confirm, don't assume.

### 5.2 Audit every user-scoped table for RLS policy coverage

For every table in the FORCE list above, confirm there's a policy that:
- **SELECT:** scopes by `current_setting('app.current_user_id', true)` via the User → Org → App → row join.
- **INSERT:** scopes the same way, typically via `WITH CHECK`.
- **UPDATE / DELETE:** ditto.

Known gap risk: the 0.46.x additions (`connection_sources`, `app_source_filters`, `log_pipeline_events`). Read each migration; if it enabled RLS but only declared a SELECT policy, the INSERT/UPDATE path is open once we drop superuser. Fix before flip, not after.

### 5.3 Audit every non-JWT write path

This is the single most likely "forgot about this path" bug class. Any code path that writes to a user-scoped table *without* a JWT in scope today relies — whether it realises it or not — on superuser bypass. Under `app_user` every such path either fails loud (insert rejected) or fails quiet (select returns zero rows and the loop silently monitors nothing). The second class is the scary one.

Split the audit into two categories by how they resolve an owning `user_id`:

**Bearer-token-authenticated handlers.** These know the owning user from the token:

- `/api/webhooks/logs` (log ingestion)
- `/api/otlp/v1/logs` (OTLP receiver)
- GitHub webhook receiver
- Syslog TLS listener (not HTTP, but writes to the same DB)

Under `app_user`, each must:
1. Resolve the bearer token to the owning `user_id`.
2. Call `s.UserQueries(ctx, ownerUserID)` (which runs `SET LOCAL app.current_user_id`).
3. Perform the `INSERT` through the returned `*db.Queries`.

If any of these currently calls `s.Pool.Begin` directly and inserts without `SET LOCAL` (relying on superuser bypass), that's a Stage 2 breakage waiting to happen.

**Background goroutines with no user context.** These are the paths `cron_user` exists for:

- `agent/monitor.go` — the 15-second monitoring loop. Enumerates active apps across all tenants, then for each app fetches new logs → classifies → escalates to Claude → writes to `agent_log`, `log_pipeline_events`, and `reports`.
- Scheduled investigations (0.31.0) — cron-shape scheduler running stored prompts on user apps.
- Connector pollers — Fly.io drain, Supabase log poller, future pull-based connectors.
- Pipeline writer — `agent/pipeline_writer.go` (post-0.47.0) inserts one row per stage per log. Hot write path.
- Notifications dispatcher — reads per-user notification config, sends Slack/email on agent findings.

The required pattern for each (the "handoff" pattern from §3):

1. Use the `cron_engine` (backed by `CRON_DATABASE_URL`) for **enumeration only**: `SELECT app_id, owner_user_id FROM applications WHERE monitoring_enabled`.
2. For each returned `(app_id, owner_user_id)`, switch to the `app_user` pool via `s.UserQueries(ctx, owner_user_id)`.
3. Do all classify / escalate / insert work inside that RLS-scoped transaction. Commit.

This means every background subsystem needs two handles: a `*pgxpool.Pool` for the cron engine and the existing `*handlers.Server` (or equivalent) for the app engine. Wire these through `agent.Agent`, the scheduler, and the notifications dispatcher — they currently take a single `*db.Queries` façade, which needs to become a pair (or a small struct holding both).

**The `pipeline_writer` case deserves special attention.** It runs inside `agent/monitor.go`'s per-app goroutine, which already has `owner_user_id` resolved at the top of the iteration. Writes should go through `app_user` + `UserQueries(ctx, owner_user_id)`, not the cron pool. If the writer was designed to use a raw `*pgxpool.Pool` for throughput, that design needs to change. `log_pipeline_events` is RLS-enabled (migration 038) with zero policies today — under `app_user` without a proper INSERT policy or a `SET LOCAL`, the entire Pipeline Page goes dark at flip time and the SSE ticker stops. Tier-1 check.

**Advisory locks.** When Heimdall goes multi-replica, the monitor loop needs `pg_try_advisory_lock` to prevent two replicas from classifying the same logs. Advisory locks must hold on the *same connection* for the duration of the critical section, which means they need session semantics — perfectly compatible with 5432 direct, and exactly the kind of thing `cron_user` is for. Not a gating concern today (single replica) but the handoff pattern should anticipate it: the cron engine is the right place for locks, the app engine is the right place for tenant writes.

### 5.4 Write the migrations (two, not one)

Split into two migration pairs so the steps remain independently reversible. Rolling role creation and FORCE RLS into one file couples two concerns that need different gating: the role has to land before CI can exercise `app_user`, but FORCE RLS is only safe to land once §5.2's policy audit has confirmed every INSERT / UPDATE / DELETE path has a matching policy. Splitting keeps the "each step is independently reversible until the final flip" property §5 promises.

**Migration A — `039_runtime_roles.up.sql` / `.down.sql`**

Creates both non-superuser roles in one migration. They land together because the code changes to switch the app over (§5.3 handoff pattern) touch both pools as a unit — splitting the role creation across two migrations adds file count without decoupling the rollout. The env-var flips (§5.5) remain independently reversible, which is what actually gives the staged rollout property §5 promises.

1. `CREATE ROLE app_user LOGIN PASSWORD '<strong-generated>'` — no `BYPASSRLS`, `SUPERUSER`, `CREATEDB`, `CREATEROLE`, or `REPLICATION`.
2. The §4.1 grants (`USAGE` on `public`, CRUD on all tables, `USAGE, SELECT` on all sequences, plus the two `ALTER DEFAULT PRIVILEGES` statements so future migrations inherit the grants).
3. `CREATE ROLE cron_user LOGIN PASSWORD '<distinct-strong-generated>' BYPASSRLS` — deliberately `BYPASSRLS`, deliberately nothing else elevated.
4. The §4.2 grants (`USAGE` on `public`, blanket `SELECT` on all tables, narrow `INSERT, UPDATE` on `monitoring_state`, `USAGE, SELECT` on all sequences, plus `ALTER DEFAULT PRIVILEGES ... GRANT SELECT` so future tables default to read-only for the cron role).

After this lands: both non-superuser roles exist. Neither is in the hot path yet (env vars still point at `postgres`), so the risk of this migration is effectively zero — it only creates roles and grants, it doesn't revoke anything from `postgres` or flip any RLS posture. RLS is still not enforced (no `FORCE` yet), so pointing staging's `DATABASE_URL` / `CRON_DATABASE_URL` at the new roles at this point is a no-op for correctness but lets CI validate that every handler and every background goroutine actually *runs* under the new roles.

Down: revoke grants in reverse order, `DROP ROLE app_user`, `DROP ROLE cron_user`. Clean because we grant rather than transfer ownership.

**Migration B — `040_rls_force_enforcement.up.sql` / `.down.sql`**

1. `ALTER TABLE … FORCE ROW LEVEL SECURITY` on every table in §4.3's list.
2. Any policy patches surfaced in §5.2 (the likely gap is INSERT / UPDATE / DELETE policies on the 0.46.x additions, which today often declare only SELECT).

This is the migration that actually flips "RLS is cosmetic" to "RLS is enforced" for `app_user`. It remains cosmetic for `cron_user` by design (BYPASSRLS) — `cron_user` safety comes from the §4.2 narrow-write grants, not from RLS. Migration B should land *after* §5.1–§5.3 audits are complete, after Migration A has baked in staging, and after the code-side handoff pattern refactor has shipped.

Down: drop `FORCE` on each table, revert any policy patches. Does not touch Migration A — if B rolls back, both runtime roles still exist and are still usable.

**Why the numbering matters:** `039` and `040` land on separate commits, separate PRs, and — importantly — can land in separate deploys. If Migration B surfaces a missed code path in prod, rolling it back does not invalidate the roles or force a re-seed. If they were one migration, a partial failure during application would leave the DB in a state neither down-migration cleanly undoes.

### 5.5 Env-var split + Makefile update + code wiring

- Add `DIRECT_URL` **and** `CRON_DATABASE_URL` to `.env.example`, `README`, and `CLAUDE.md` setup instructions.
- Update `Makefile` targets `migrate-up`, `migrate-down`, `migrate-create`, `sqlc-generate` (where it touches the DB) to read `DIRECT_URL` instead of `DATABASE_URL`.
- Leave `DATABASE_URL` pointing at `postgres` superuser temporarily during the rollout — don't break dev setups in the same PR that lands the migration.
- Add a second `*pgxpool.Pool` in `backend/cmd/heimdall/main.go` for the cron engine, parallel to the existing one. Construct it from `CRON_DATABASE_URL`. Thread it through to `agent.Agent`, the scheduler, and anywhere else §5.3 identified as a background writer. If `CRON_DATABASE_URL` is empty at startup, **fall back to the primary pool and emit a loud `WARN`** — the scheduler will still run (superuser bypass works the same way), but the warning makes the degraded posture visible in logs. In prod, treat missing `CRON_DATABASE_URL` as a config error rather than a silent downgrade.
- Refactor every site identified in §5.3's "background goroutines" list to use the handoff pattern: cron pool for enumeration, app pool + `UserQueries` for per-user writes.
- Add a config-level invariant: in production builds, refuse to start if `DATABASE_URL` and `CRON_DATABASE_URL` resolve to the same role (would silently undo the entire plan).

### 5.6 The RLS regression test

Add a test under `backend/internal/api/handlers/` that does NOT get skipped when `DATABASE_URL` is set:

```go
// Pseudocode — structure matches existing testhelpers_test.go pattern.
func TestRLS_CrossTenantReadIsDenied(t *testing.T) {
    userA := createUser(t, ...)
    userB := createUser(t, ...)

    rowInA := createConnection(t, userA, ...)

    // Using userB's identity, attempt to read userA's row.
    q, done := server.UserQueries(ctx, userB.ID)
    defer done()

    got, err := q.GetConnectionByID(ctx, rowInA.ID)
    require.Error(t, err)
    require.True(t, errors.Is(err, pgx.ErrNoRows))
    require.Nil(t, got)
}
```

This test **fails** on the current superuser setup (it would return `rowInA`) and **passes** on `app_user`. That's the signal. It encodes the invariant "RLS is enforced" so future regressions (including accidentally re-granting BYPASSRLS) fail CI.

A variant for INSERT / UPDATE / DELETE should exist for each critical table — `log_buffer`, `conversations`, `connection_sources`, `app_source_filters` are the minimum.

**Additional tests for the cron role:**

- `TestCronRole_CanEnumerateAcrossTenants`: using a connection as `cron_user`, `SELECT COUNT(*) FROM applications` returns rows from all tenants (proves BYPASSRLS works as designed).
- `TestCronRole_CannotDropTables`: `DROP TABLE applications` returns `42501 permission denied`. Proves BYPASSRLS doesn't leak DDL.
- `TestCronRole_CannotReadAuthUsers`: `SELECT * FROM auth.users` returns `42501 permission denied`. Proves the cron role can't escalate through Supabase internals.
- `TestCronRole_CannotWriteBlanketTenantTable`: `INSERT INTO log_buffer (...)` returns `42501 permission denied` (unless that table is explicitly in the narrow-write allowlist from §4.2). Proves the "cron is for enumeration, not tenant writes" invariant is enforced by grants, not just convention.

Together these pin the posture: `cron_user` can see everything, can write almost nothing, and cannot reshape anything. If a future migration accidentally grants blanket INSERT to `cron_user`, the fourth test fails in CI.

### 5.6a Catalog drift checks (posture as SQL invariants)

The §5.6 tests above verify *behaviour* — what each role can and can't do when exercised. These verify *catalog state* directly: RLS flags and role attributes that, if drifted, silently weaken the posture even while the behaviour tests still pass. Each is a ~15 line Go test wrapping one SQL query, running against the same `DATABASE_URL`-backed integration DB as the rest of the suite. Assertion in every case: the query returns zero rows.

```sql
-- Test 1: every RLS-enabled table in public must also be FORCE'd.
-- A future migration adds a user-scoped table and forgets FORCE → without
-- this check, RLS silently reverts to cosmetic for that table, and the
-- only signal is a cross-tenant leak after the fact.
SELECT tablename FROM pg_tables
WHERE schemaname = 'public' AND rowsecurity = true AND forcerowsecurity = false;

-- Test 2: every RLS-enabled table has SELECT/INSERT/UPDATE/DELETE policies.
-- 0.46.x connection_sources / app_source_filters and 0.47.0 log_pipeline_events
-- each shipped with partial policy coverage at some point in their migration
-- history. Prevents recurrence as the schema grows.
SELECT t.tablename FROM pg_tables t
WHERE t.schemaname = 'public' AND t.rowsecurity = true
  AND (SELECT COUNT(DISTINCT cmd) FROM pg_policies p
       WHERE p.schemaname = t.schemaname AND p.tablename = t.tablename
         AND cmd IN ('SELECT','INSERT','UPDATE','DELETE')) < 4;

-- Test 3: app_user must never gain BYPASSRLS.
-- A stray ALTER ROLE app_user BYPASSRLS would pass every behaviour test
-- (no tenant write runs *as* app_user in the test suite) and silently
-- collapse enforcement everywhere. Only this check catches it.
SELECT rolname FROM pg_roles WHERE rolname = 'app_user' AND rolbypassrls = true;

-- Test 4: cron_user must never gain SUPERUSER.
SELECT rolname FROM pg_roles WHERE rolname = 'cron_user' AND rolsuper = true;
```

Rationale split: behaviour tests catch broken *policy logic*; catalog tests catch broken *flags and role attributes*. Both failure classes exist, neither class subsumes the other. Four tests, zero new infrastructure — they ride the same Go integration harness and the same test Postgres as everything else in `backend/internal/api/handlers/`.

### 5.7 Staged rollout

The migration split in §5.4 unlocks a two-stage rollout that separates "does the app run as a non-superuser at all" from "does RLS enforcement break anything." These are different failure modes and diagnosing them together is harder than diagnosing them apart.

**No staging environment.** Heimdall runs only a production deployment today — a deliberate choice during the <10-customer Beta phase. That raises the bar on the two-stage structure rather than removing it: both stages still land in prod, but separating them over time is what lets us attribute any error to the right cause. Collapsing Stage 1 and Stage 2 into a single deploy would fuse privilege errors and policy errors into one incident, which is exactly what the split is meant to prevent. The bake windows below refer to **production traffic**, not staging.

**Stage 1 — `DATABASE_URL` → `app_user` AND `CRON_DATABASE_URL` → `cron_user`, FORCE RLS not yet landed.**

Order: local dev → prod (no staging to vet through first). After Migration A lands but before Migration B:

- Point the HTTP backend at `app_user`. Point the background goroutines (monitor loop, scheduler, connector pollers, notifications dispatcher) at `cron_user` via the cron engine wired in §5.5. RLS is still cosmetic, so any failure here is a pure privilege-set issue, not an RLS-policy issue.
- Watch for `42501 permission denied` on DDL, `auth.*` reads, `SET ROLE`, `CREATE EXTENSION` — these are §5.1 misses. Each is a single grep-and-fix to patch up.
- Watch for errors on background writes that the refactor missed. The monitor loop should still be producing `agent_log` and `log_pipeline_events` rows at normal rates — a sudden drop to zero means a writer is trying to go through `cron_user` (which lacks the grants) instead of handing off to `app_user`. Alert on it.
- Confirm `/api/webhooks/logs`, OTLP, syslog, GitHub ingestion still accept writes at normal volumes.
- Confirm the pipeline writer is still producing rows in `log_pipeline_events` (SSE ticker should still be moving, bootstrap endpoint should still return non-empty summaries).
- **Bake for 48h of production traffic before moving to Stage 2** — long enough for every scheduled investigation to have fired at least once, every connector poller to have run a full cycle, and every rarely-exercised ingestion path to have seen at least one request. Without a staging environment, prod traffic *is* the exercise surface; compressing this window means carrying a latent privilege-set bug into Stage 2 where it will be misdiagnosed as (and incorrectly fixed as) a policy bug.

At the end of Stage 1 we have: no runtime code on superuser, no RLS enforcement. Strictly better than today (no DDL, no `auth.*`, no role escalation from either runtime role) but still not the full posture.

**Stage 2 — land Migration B (FORCE RLS + policy patches).**

Order: prod directly. With Stage 1 already clean, any new error surfaced here is an RLS-policy issue, not a privilege issue:

- Watch for `new row violates row-level security policy` errors on the `app_user` pool. These are §5.2 or §5.3 misses: a handler writing without `SET LOCAL`, or a table whose INSERT/UPDATE/DELETE policy doesn't match its SELECT policy. `cron_user` is unaffected (BYPASSRLS).
- The webhook / OTLP / syslog / GitHub paths remain the most likely to break here — they don't go through `UserQueries` in the same way JWT-authenticated handlers do. Without a staging environment to vet these end-to-end against real traffic, the §5.2 policy audit is the single most important gate — treat any table with missing INSERT/UPDATE/DELETE policy coverage as a deploy-blocker, not a fix-forward item. The §5.6a Test 2 catalog check must be green before Migration B ships.
- Background subsystems (monitor loop, scheduler) should be unaffected at Stage 2 because their tenant writes already go through `app_user` + `UserQueries` (wired in §5.5). If the monitor loop starts producing policy-violation errors, that's a refactor miss — a background writer is still using `cron_user` for a tenant write. Fix the writer, not the policy.
- **Rehearse the rollback locally before the prod deploy.** Migration B's down-migration (drop `FORCE`, revert policy patches) is small and clean, but in a single-environment rollout the ability to revert quickly is load-bearing. Run the down-migration against a local DB end-to-end so the muscle memory exists under pressure.
- Bake for 48h in prod before declaring the rollout complete.

**Compensating controls for the missing staging environment:**

- Exercise both stages end-to-end against a local Postgres with representative seed data before each prod deploy. Every path in §5.3's background-writer list needs at least one manual trigger — webhook ingestion, OTLP receive, syslog receive, GitHub webhook, a scheduled investigation cycle, a full monitor-loop iteration.
- Keep the deploys small by design: Migration A is purely additive (creates roles, grants privileges, revokes nothing), Migration B is a flag flip plus targeted policy patches. Neither is structurally hard to revert.
- The "strictly better than today" checkpoint at the end of Stage 1 is where the risk concentrates. If something looks off at hour 48 — elevated error rate, a scheduled job that didn't fire, missing `log_pipeline_events` rows — hold Stage 2 and investigate. Don't press forward on the theory that FORCE RLS will surface the root cause; it will, but it'll do so in a way that mixes two error classes.
- When customer count crosses ~10 or paid revenue enters the picture, the independent case for a staging environment strengthens. Worth revisiting the rollout playbook at that point — not as a blocker on this work, but as a follow-up that would have made this rollout materially safer.

### 5.8 Clean up

- Rotate the `postgres` password after the flip (app doesn't need it anymore, only migrations do).
- Rotate `app_user` and `cron_user` passwords to long random values, distinct from each other and from `postgres`.
- Update `CLAUDE.md` to note the new topology — three URLs, three roles, what each is for.
- Update `docs/blueprints/database-connection-blueprint.md` — §4 (`backend-blueprint.md:517` note becomes stale), §2 (add all three roles to the endpoint table, distinguishing `app_user` on port 5432 from `cron_user` on port 5432), §8 (remove the "we connect as superuser" premise, add the "two runtime pools plus a migrations-only superuser pool" description).
- Update `docs/blueprints/backend-blueprint.md:517` specifically — the current note will become a lie once this lands.
- Cross-link `docs/executing/db-connection-type-improvement.md` from the blueprint's §8 so the 5432→6543 follow-up is discoverable.

---

## 6. Related audits worth batching

These are orthogonal to the role split but easy to do in the same PR cycle:

- **`service_role` key usage**: grep the codebase for `service_role`, `SUPABASE_SERVICE_ROLE_KEY`, or similar. Any path that uses Supabase's service-role key against PostgREST bypasses RLS via a completely different route and needs the same "why is this privileged access necessary" audit.
- **RLS coverage on 0.46.x tables**: `connection_sources`, `app_source_filters` were added mid-0.46.x. `log_pipeline_events` (migration 038) is in flight. Each needs a full SELECT / INSERT / UPDATE / DELETE policy set, not just RLS-enabled. The `source_filtering-phase*` completion docs (now in `docs/archive/`) should be cross-referenced.
- **`SET LOCAL` usage audit**: verify that every code path that writes to a user-scoped table goes through `UserQueries`, not the raw pool directly. Any call site that uses `s.Pool` to begin a transaction without the `SET LOCAL` is already a latent bug under superuser (relies on app-code filtering) and a hard failure under `app_user` (RLS denies the write).
- **Idempotency / rate-limit tables**: 0.45.1 added RLS to the idempotency table. Check whether there are other infra-shape tables (e.g. `schema_migrations` — that one we don't want to lock down, since golang-migrate uses `DIRECT_URL` anyway) that need the same treatment.

---

## 7. Rough effort estimate

- §5.1–5.3 audits: **one day** (read migrations, grep handlers, inspect every ingestion path *and* every background writer). The background-writer audit is the half-day addition over the two-role version of this plan.
- §5.4 migration: **two to three hours** once the audit is done — both roles go into Migration A, so SQL is modestly longer but still trivial.
- §5.5 env-var split + cron pool wiring: **half a day** — Makefile edits, env-var docs, the new `*pgxpool.Pool` in `main.go`, and threading it into the agent / scheduler / connector pollers / notifications dispatcher. The handoff-pattern refactor for background writers is the bulk of this. Call it **one to two days** if any writer turns out to be entangled with `s.Pool` in a way that needs untangling.
- §5.6 RLS regression tests + §5.6a catalog drift checks: **one to two days** — one test per critical table for `app_user`, plus the four `cron_user` posture tests and the four catalog drift checks, wired to run against the integration test DB in CI.
- §5.7 staged rollout: **elapsed time over ~4–5 days**, dominated by two 48h prod bake windows (Stage 1, then Stage 2). No staging means prod traffic is the exercise surface — don't compress the bakes to save calendar time.
- §5.8 cleanup: **an hour**.

Total active work: **3–5 days**. Total wall-clock with prod bake windows: **~1.5–2 weeks**. The three-role version is roughly a day longer than the two-role version, almost entirely driven by the background-writer refactor in §5.5 — but that refactor is the whole point, because it's what makes RLS enforcement actually reach the monitor loop and the pipeline writer.

This assumes no surprises in §5.1. If any code path turns out to need superuser for reasons we don't yet know (or if any background writer is structurally entangled with the raw pool in ways that need redesign rather than rewiring), the scope expands.

---

## 8. Out of scope (don't bundle)

- **Supavisor migration.** Separate project with separate blockers. See blueprint §8.
- **Moving to a dedicated schema** (`CREATE SCHEMA heimdall`) with app-level ownership. Cleaner long-term, but not needed for enforcement — `public` + explicit grants is sufficient.
- **Read replicas / query routing.** Orthogonal scaling work.
- **RLS policy rewrites for performance.** Some policies currently do multi-join subqueries that would be cheaper as cached `app.current_org_id` reads. Worth doing eventually, definitely not in this PR.

---

## 9. Acceptance checklist

Ship-ready when all of these are true:

**Migrations and roles**
- [ ] Migration A (`039_runtime_roles`) has landed in prod: `app_user` exists with the §4.1 grants and none of the denied privileges; `cron_user` exists with the §4.2 narrow grants, `BYPASSRLS`, and none of the denied privileges.
- [ ] Migration B (`040_rls_force_enforcement`) has landed in prod: every user-scoped table has `FORCE ROW LEVEL SECURITY` set and a policy for each of SELECT / INSERT / UPDATE / DELETE.

**Environment wiring**
- [ ] `DATABASE_URL` in prod points at `app_user`.
- [ ] `CRON_DATABASE_URL` in prod points at `cron_user`.
- [ ] `DIRECT_URL` in prod points at `postgres`.
- [ ] Production startup refuses to launch if `DATABASE_URL` and `CRON_DATABASE_URL` resolve to the same role.
- [ ] `make migrate-up` and `make migrate-down` still work (via `DIRECT_URL`).

**Code & tests**
- [ ] Every background writer identified in §5.3 uses the cron-pool-for-enumeration, app-pool-for-writes handoff pattern. No background goroutine calls `s.Pool.Begin` directly on the superuser pool.
- [ ] All backend tests pass against `app_user` + `cron_user` in CI.
- [ ] At least one RLS regression test per critical table exists in CI and fails if BYPASSRLS is re-granted to `app_user`.
- [ ] The four `cron_user` behaviour tests from §5.6 exist in CI and fail if cron_user grows `SUPERUSER`, loses `BYPASSRLS`, gains blanket INSERT on `public.*`, or gains access to `auth.*`.
- [ ] The four §5.6a catalog drift checks exist in CI: (1) every RLS-enabled public table is `FORCE`'d, (2) every RLS-enabled table has SELECT/INSERT/UPDATE/DELETE policy coverage, (3) `app_user` does not have `BYPASSRLS`, (4) `cron_user` does not have `SUPERUSER`.

**Observability (production, no staging)**
- [ ] No `42501 permission denied` errors in the production log after the 48h Stage 1 bake window.
- [ ] No `new row violates row-level security policy` errors in the production log after the 48h Stage 2 bake window.
- [ ] Monitor loop throughput (rows/min to `agent_log` and `log_pipeline_events`) unchanged from pre-flip baseline across both bake windows.
- [ ] Webhook / OTLP / syslog / GitHub ingestion paths all confirmed working under the new role via production traffic (not staging) across both stages.
- [ ] Pipeline page SSE ticker and bootstrap endpoint both serve non-empty responses under the new role.
- [ ] Migration B rollback rehearsed against a local DB before the Stage 2 prod deploy.

**Cleanup**
- [ ] The `postgres` password has been rotated.
- [ ] `app_user` and `cron_user` passwords are distinct, random, and stored only in the secrets manager (not in any committed `.env*`).
- [ ] `docs/blueprints/database-connection-blueprint.md` and `docs/blueprints/backend-blueprint.md` reflect the three-role topology.
- [ ] `CLAUDE.md` includes `DIRECT_URL` and `CRON_DATABASE_URL` in the setup instructions.
- [ ] `docs/executing/db-connection-type-improvement.md` is cross-linked from the blueprint's §8 as the follow-up for the 5432→6543 question.
