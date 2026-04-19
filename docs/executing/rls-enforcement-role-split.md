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

Two URLs, two roles, clear separation of responsibilities:

```
DATABASE_URL  → app_user:<pw>@<host>:5432/postgres
                  └─ Non-superuser, RLS-enforced.
                  └─ Used by the Go server (pgxpool, UserQueries, all handlers, agent, notifications).
                  └─ Cannot DROP, cannot ALTER SYSTEM, cannot read auth.*, cannot SET ROLE.

DIRECT_URL    → postgres:<pw>@<host>:5432/postgres
                  └─ Superuser, retains DDL rights.
                  └─ Used ONLY by golang-migrate (`make migrate-up`, `make migrate-down`).
                  └─ Never read by the long-running backend process.
```

This is the Supabase-recommended production pattern and matches what Prisma, Drizzle, and most Supabase-native stacks ship with. The `DIRECT_URL` vs. `DATABASE_URL` split is the same naming convention those ORMs use, so it's recognisable to anyone onboarding onto the project.

### Port 5432, not 6543

This plan keeps both URLs on **port 5432 (direct / session mode)**. Moving `DATABASE_URL` to Supavisor on 6543 is a separate project with its own blockers (session GUCs on user-owned connectors, prepared-statement cache behaviour, future `LISTEN/NOTIFY`) — see `docs/blueprints/database-connection-blueprint.md` §8. Don't bundle that migration into this one; it roughly triples the blast radius.

---

## 4. What `app_user` needs (and doesn't)

### Grants

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

### FORCE RLS on every user-scoped table

Without `FORCE`, the table *owner* still bypasses RLS. Supabase's table ownership model varies across migrations, so rely on `FORCE` rather than assuming ownership:

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

### What must NOT be granted

- `BYPASSRLS` (the whole point is that it can't)
- `SUPERUSER`
- `CREATE` on schema `public` (prevents rogue table creation under the app role)
- Any access to `auth.*`, `storage.*`, `extensions.*`, or other Supabase internal schemas
- `CREATE EXTENSION`, `ALTER SYSTEM`, or anything that reshapes the server

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

### 5.3 Audit the webhook ingestion path

`/api/webhooks/logs` uses bearer-token auth (not JWT) and its DB writes currently go through the superuser pool. Under `app_user` the handler must:

1. Resolve the bearer token to the owning user ID.
2. `SET LOCAL app.current_user_id = $ownerID` on the transaction.
3. Perform the `INSERT INTO log_buffer`.

Confirm steps 1–3 happen in that order in the existing handler. If the handler currently inserts without `SET LOCAL` (relying on superuser bypass), it will start returning RLS failures at flip time — this is the single most likely "forgot about this path" bug.

Same audit for:
- OTLP receiver (`/api/otlp/v1/logs`)
- GitHub webhook receiver
- Syslog TLS listener (not HTTP, but writes to the same DB)
- Any poller-driven connector that writes to `log_buffer` or `agent_log`

### 5.4 Write the migrations (two, not one)

Split into two migration pairs so the steps remain independently reversible. Rolling role creation and FORCE RLS into one file couples two concerns that need different gating: the role has to land before CI can exercise `app_user`, but FORCE RLS is only safe to land once §5.2's policy audit has confirmed every INSERT / UPDATE / DELETE path has a matching policy. Splitting keeps the "each step is independently reversible until the final flip" property §5 promises.

**Migration A — `039_app_user_role.up.sql` / `.down.sql`**

1. `CREATE ROLE app_user LOGIN PASSWORD '<strong-generated>'` — no `BYPASSRLS`, `SUPERUSER`, `CREATEDB`, `CREATEROLE`, or `REPLICATION`.
2. The grants in §4 (`USAGE` on `public`, CRUD on all tables, `USAGE, SELECT` on all sequences, plus the two `ALTER DEFAULT PRIVILEGES` statements so future migrations inherit the grants).

After this lands: `app_user` exists and can be pointed at by a staging `DATABASE_URL` for the §5.1 audit and the §5.6 regression tests. RLS is still not enforced (no `FORCE` yet), so flipping the env var at this point is a no-op for correctness but lets CI validate that every handler at least *runs* under the new role.

Down: revoke grants, `DROP ROLE app_user`. Clean because we grant rather than transfer ownership.

**Migration B — `040_rls_force_enforcement.up.sql` / `.down.sql`**

1. `ALTER TABLE … FORCE ROW LEVEL SECURITY` on every table in §4's list.
2. Any policy patches surfaced in §5.2 (the likely gap is INSERT / UPDATE / DELETE policies on the 0.46.x additions, which today often declare only SELECT).

This is the migration that actually flips "RLS is cosmetic" to "RLS is enforced." It should land *after* §5.1–§5.3 audits are complete and Migration A has baked in staging, not as part of the same PR.

Down: drop `FORCE` on each table, revert any policy patches. Does not touch Migration A — if B rolls back, `app_user` still exists and is still usable.

**Why the numbering matters:** `039` and `040` land on separate commits, separate PRs, and — importantly — can land in separate deploys. If Migration B surfaces a missed code path in prod, rolling it back does not invalidate the role or force a re-seed. If they were one migration, a partial failure during application would leave the DB in a state neither down-migration cleanly undoes.

### 5.5 Env-var split + Makefile update

- Add `DIRECT_URL` to `.env.example` and the `README` / `CLAUDE.md` setup instructions.
- Update `Makefile` targets `migrate-up`, `migrate-down`, `migrate-create`, `sqlc-generate` (where it touches the DB) to read `DIRECT_URL` instead of `DATABASE_URL`.
- Leave `DATABASE_URL` pointing at `postgres` superuser temporarily during the rollout — don't break dev setups in the same PR that lands the migration.

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

### 5.7 Staged rollout

The migration split in §5.4 unlocks a two-stage rollout that separates "does the app run as a non-superuser at all" from "does RLS enforcement break anything." These are different failure modes and diagnosing them together is harder than diagnosing them apart.

**Stage 1 — `DATABASE_URL` → `app_user`, FORCE RLS not yet landed.**

Order: local dev → staging → prod. After Migration A lands but before Migration B:

- Point the backend at `app_user`. RLS is still cosmetic, so any failure here is a pure privilege-set issue, not an RLS-policy issue.
- Watch for `42501 permission denied` on DDL, `auth.*` reads, `SET ROLE`, `CREATE EXTENSION` — these are §5.1 misses. Each is a single grep-and-fix to patch up.
- Confirm `/api/webhooks/logs`, OTLP, syslog, GitHub ingestion still accept writes at normal volumes.
- Bake for at least a few hours of realistic traffic before moving to Stage 2.

At the end of Stage 1 we have: non-superuser app, no RLS enforcement. Strictly better than today (no DDL, no `auth.*`, no role escalation) but still not the full posture.

**Stage 2 — land Migration B (FORCE RLS + policy patches).**

Order: staging first, then prod. With Stage 1 already clean, any new error surfaced here is an RLS-policy issue, not a privilege issue:

- Watch for `RLS policy violation` / `new row violates row-level security policy` errors. These are §5.2 or §5.3 misses: a handler writing without `SET LOCAL`, or a table whose INSERT/UPDATE/DELETE policy doesn't match its SELECT policy.
- The webhook / OTLP / syslog / GitHub paths are the most likely to break here — they don't go through `UserQueries` in the same way JWT-authenticated handlers do. Confirm each works end-to-end against staging before flipping prod.
- Bake for 48h in staging before prod.

### 5.8 Clean up

- Rotate the `postgres` password after the flip (app doesn't need it anymore, only migrations do).
- Update `CLAUDE.md` to note the new topology.
- Update `docs/blueprints/database-connection-blueprint.md` — §4 (`backend-blueprint.md:517` note becomes stale), §2 (add the role distinction to the endpoint table), §8 (remove the "we connect as superuser" premise).
- Update `docs/blueprints/backend-blueprint.md:517` specifically — the current note will become a lie once this lands.

---

## 6. Related audits worth batching

These are orthogonal to the role split but easy to do in the same PR cycle:

- **`service_role` key usage**: grep the codebase for `service_role`, `SUPABASE_SERVICE_ROLE_KEY`, or similar. Any path that uses Supabase's service-role key against PostgREST bypasses RLS via a completely different route and needs the same "why is this privileged access necessary" audit.
- **RLS coverage on 0.46.x tables**: `connection_sources`, `app_source_filters` were added mid-0.46.x. `log_pipeline_events` (migration 038) is in flight. Each needs a full SELECT / INSERT / UPDATE / DELETE policy set, not just RLS-enabled. The `source_filtering-phase*` completion docs (now in `docs/archive/`) should be cross-referenced.
- **`SET LOCAL` usage audit**: verify that every code path that writes to a user-scoped table goes through `UserQueries`, not the raw pool directly. Any call site that uses `s.Pool` to begin a transaction without the `SET LOCAL` is already a latent bug under superuser (relies on app-code filtering) and a hard failure under `app_user` (RLS denies the write).
- **Idempotency / rate-limit tables**: 0.45.1 added RLS to the idempotency table. Check whether there are other infra-shape tables (e.g. `schema_migrations` — that one we don't want to lock down, since golang-migrate uses `DIRECT_URL` anyway) that need the same treatment.

---

## 7. Rough effort estimate

- §5.1–5.3 audits: **half a day** (read migrations, grep handlers, inspect ingestion paths).
- §5.4 migration: **an hour or two** once the audit is done.
- §5.5 env-var split: **an hour**, mostly Makefile edits and dev-setup docs.
- §5.6 RLS regression tests: **a day** — writing one test per critical table, wiring them to run against the staging DB in CI.
- §5.7 staged rollout: **elapsed time over ~3 days**, mostly waiting for staging to marinate before prod.
- §5.8 cleanup: **an hour**.

Total active work: ~2–3 days. Total wall-clock with staging baking: ~1 week. This assumes no surprises in §5.1 — if any code path turns out to need superuser for reasons we don't yet know, the scope expands.

---

## 8. Out of scope (don't bundle)

- **Supavisor migration.** Separate project with separate blockers. See blueprint §8.
- **Moving to a dedicated schema** (`CREATE SCHEMA heimdall`) with app-level ownership. Cleaner long-term, but not needed for enforcement — `public` + explicit grants is sufficient.
- **Read replicas / query routing.** Orthogonal scaling work.
- **RLS policy rewrites for performance.** Some policies currently do multi-join subqueries that would be cheaper as cached `app.current_org_id` reads. Worth doing eventually, definitely not in this PR.

---

## 9. Acceptance checklist

Ship-ready when all of these are true:

- [ ] Migration A (`039_app_user_role`) has landed in prod: `app_user` exists with the grants in §4 and none of the denied privileges.
- [ ] Migration B (`040_rls_force_enforcement`) has landed in prod: every user-scoped table has `FORCE ROW LEVEL SECURITY` set and a policy for each of SELECT / INSERT / UPDATE / DELETE.
- [ ] `DATABASE_URL` in prod points at `app_user`. `DIRECT_URL` in prod points at `postgres`.
- [ ] `make migrate-up` and `make migrate-down` still work (via `DIRECT_URL`).
- [ ] All backend tests pass against `app_user` in CI.
- [ ] At least one RLS regression test per critical table exists in CI and fails if BYPASSRLS is re-granted.
- [ ] No `42501 permission denied` errors in the staging log after 48h of realistic traffic.
- [ ] Webhook / OTLP / syslog / GitHub ingestion paths all confirmed working under the new role.
- [ ] The `postgres` password has been rotated.
- [ ] `docs/blueprints/database-connection-blueprint.md` and `docs/blueprints/backend-blueprint.md` reflect the new topology.
- [ ] `CLAUDE.md` includes `DIRECT_URL` in the setup instructions.
