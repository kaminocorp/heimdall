# RLS Enforcement — Phase 4 Completion: Migration A (Role Grants)

**Status:** complete (code-side); production bootstrap pending operator
window. The migration ships safe-to-apply only after the manual bootstrap
runs once per environment — see *Bootstrap procedure* below.
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 4.
**Phase 3 completion (gating doc):** [`./rls-enforcement-phase-3.md`](./rls-enforcement-phase-3.md)
**Date:** 2026-04-30.

---

## Executive summary

Phase 4 creates the privilege topology in the database. After the bootstrap
plus migration 039, both runtime roles (`app_user`, `cron_user`) exist with
the exact grants the parent plan §4.1 / §4.2 specifies, plus the two
audit-discovered widenings Phase 1 surfaced and Phase 2 reaffirmed:
`DELETE ON log_buffer` for `cron_user` (the pruner), and the PG17
conditional `WITH SET TRUE` membership for the Phase 5 `asAppUser`
fixture. `DATABASE_URL` and `CRON_DATABASE_URL` continue to authenticate
as `postgres`; the env-var flip is Phase 6's job.

Everything is purely additive. The migration adds grants, never revokes
them, never enables `FORCE`, never patches a policy. That additive shape
is what makes this phase safe to ship to a production that's still
running entirely on the postgres superuser pool — the runtime path never
exercises the new grants until Phase 6 routes traffic through them.

Three Phase-4-specific design decisions, locked in this commit:

1. **Role lifecycle stays manual; grants live in the migration.** The
   parent plan and Trajan precedent are explicit on this: `CREATE ROLE
   ... PASSWORD '...'` is environment bootstrap, not migration material.
   Committing role creation with passwords would either bake secrets into
   git or force a placeholder password that every operator has to remember
   to rotate. Splitting bootstrap (manual, password-bearing, idempotent
   per environment) from grants (versioned, idempotent, schema-coupled)
   means grants can be re-applied freely as part of any normal
   `make migrate-up` run, and roles never re-create across environments.
2. **Bootstrap-assertion guard at the top of the migration.** The first
   `DO` block in `039.up.sql` raises a clear, file-path-bearing exception
   when either runtime role is missing. Without it, the migration's
   GRANT statements would fail with a generic "role does not exist"
   message that gives an operator one search-engine hop to figure out
   what to do next; the assertion gives them the answer in the error
   message itself.
3. **`ALTER DEFAULT PRIVILEGES` is scoped `FOR ROLE postgres`.** The
   parent plan §4.1 / §4.2 wrote the defaults without `FOR ROLE`, which
   makes them apply to "the current_user when this statement runs."
   That's already `postgres` for migrations under `DIRECT_URL`, but the
   explicit `FOR ROLE postgres` keeps the defaults targeted even if a
   future operator runs the migration as some other superuser
   (e.g. cluster admin during a restore). Functionally identical today;
   structurally robust to "who's holding the migration shell" drift.

---

## What landed

### Migration files

```
backend/migrations/039_runtime_role_grants.up.sql
backend/migrations/039_runtime_role_grants.down.sql
```

The up migration's structure (commented in-file):

1. **Bootstrap assertion.** Refuses to apply unless both `app_user` and
   `cron_user` already exist. Error message names the bootstrap doc.
2. **`app_user` grants** — parent plan §4.1 verbatim:
   - `USAGE` on schema `public`.
   - `SELECT, INSERT, UPDATE, DELETE` on all tables.
   - `USAGE, SELECT` on all sequences.
   - `ALTER DEFAULT PRIVILEGES FOR ROLE postgres ... GRANT ... ON TABLES`
     and `ON SEQUENCES`.
3. **`cron_user` grants** — parent plan §4.2 with two narrow writes:
   - `USAGE` on schema `public`.
   - Blanket `SELECT` on all tables.
   - `USAGE, SELECT` on all sequences.
   - `INSERT, UPDATE` on `monitoring_state` (cursor advancement).
   - `DELETE` on `log_buffer` — **audit-discovered widening**, not in
     the parent plan as originally drafted. Justified by `agent/pruner.go`
     running `DELETE FROM log_buffer WHERE ingested_at < ...` across all
     tenants on a 1h schedule. Captured in
     `rls-enforcement-phase-1.md` §5.3 row "Log buffer pruner" and
     reaffirmed in `rls-enforcement-phase-2.md` "What Phase 4 needs from
     this".
   - `ALTER DEFAULT PRIVILEGES FOR ROLE postgres ... GRANT SELECT ON
     TABLES` and `... GRANT USAGE, SELECT ON SEQUENCES`. Future tables
     default to read-only for `cron_user` — every new write becomes an
     explicit grant, which is an explicit security review.
4. **PG17 conditional grant block** — `GRANT app_user TO postgres WITH
   SET TRUE` + `GRANT cron_user TO postgres WITH SET TRUE` on PG17+,
   plain `GRANT ... TO postgres` on PG14–PG16. Required for the Phase 5
   `asAppUser` fixture's `SET LOCAL ROLE` calls.
5. **Function execute grants.** `app_current_user_id()` and
   `app_user_org_ids()` get explicit `GRANT EXECUTE` to both runtime
   roles. PUBLIC already has EXECUTE on these by default; the explicit
   grant is defence-in-depth so revoking PUBLIC later (a future
   hardening) doesn't strand the runtime roles. `handle_new_user()` is
   a trigger body fired by an `auth.users` insert trigger and is never
   `EXECUTE`d by runtime code, so no grant is added.

The down migration's structure:

- Skips cleanly with a `RAISE NOTICE` when either role is absent.
- Otherwise revokes everything in reverse order: function grants → role
  membership of `postgres` → `cron_user` grants (defaults, then narrow
  writes, then blanket SELECT, then `USAGE`) → `app_user` grants
  (defaults, then table CRUD, then `USAGE`).
- **Does NOT `DROP ROLE`.** Roles outlive migrations; their lifecycle is
  bootstrap-only, mirroring the up migration's split.

No code changes shipped — Phase 2 absorbed the call-shape ripple, Phase 3
absorbed the wiring. Phase 4 is migration-files-and-runbook-only.

---

## Bootstrap procedure (verbatim)

**Run in the Supabase SQL editor or `psql`, authenticated as `postgres`,
once per environment.** Local dev, staging (when one exists), and
production each need their own run with a freshly generated password
distinct between the two roles and across environments.

```sql
-- Phase 4 bootstrap — run once per environment, BEFORE 039_runtime_role_grants.up.sql.
-- Both passwords must be strong, randomly generated, and distinct from each other.
-- Store the passwords in the secrets manager only — do NOT commit to git.

-- app_user — RLS-enforced runtime role.
-- Implicit: NOSUPERUSER, NOBYPASSRLS, NOCREATEDB, NOCREATEROLE, NOREPLICATION.
CREATE ROLE app_user LOGIN PASSWORD '<strong-generated-password-1>';

-- cron_user — BYPASSRLS enumeration role. BYPASSRLS is the ONLY elevated attribute.
-- Implicit: NOSUPERUSER, NOCREATEDB, NOCREATEROLE, NOREPLICATION.
CREATE ROLE cron_user LOGIN PASSWORD '<strong-generated-password-2>' BYPASSRLS;
```

**Verification (run after bootstrap, before applying 039):**

```sql
SELECT rolname, rolsuper, rolbypassrls, rolcreatedb, rolcreaterole, rolreplication, rolcanlogin
FROM pg_roles
WHERE rolname IN ('app_user', 'cron_user')
ORDER BY rolname;
```

Expected output:

```
 rolname   | rolsuper | rolbypassrls | rolcreatedb | rolcreaterole | rolreplication | rolcanlogin
-----------+----------+--------------+-------------+---------------+----------------+-------------
 app_user  | f        | f            | f           | f             | f              | t
 cron_user | f        | t            | f           | f             | f              | t
```

If `app_user` shows `rolbypassrls = t` or `cron_user` shows `rolsuper = t`,
**stop**. Drop the role (`DROP ROLE name;`) and re-create with the exact
SQL above; do not paper over with `ALTER ROLE`. The Phase 5 catalog drift
checks would fail loudly, but catching it pre-migration saves a round trip.

### Production bootstrap handoff

| Step | Owner | Mechanism |
|---|---|---|
| Generate passwords | Operator running the bootstrap | `openssl rand -base64 36` (or equivalent), one per role, distinct |
| Run bootstrap SQL | Operator | Supabase SQL editor authenticated as `postgres`, or `psql "$DIRECT_URL"` for the prod DIRECT_URL |
| Stash passwords | Operator | Secrets manager (Fly secrets / 1Password / equivalent) under names `APP_USER_PASSWORD` and `CRON_USER_PASSWORD`. Phase 6 will compose the connection URLs from these |
| Verify role attributes | Operator | Run the verification SELECT above |
| Apply migration 039 | CI / `make migrate-up` via `DIRECT_URL` | The bootstrap-assertion guard fails the migration if step 2 was skipped |

The passwords never enter version control. The `.env*` files are not the
storage location — the secrets manager is. Phase 6 composes the runtime
URLs at deploy time from the secrets-manager values.

---

## Acceptance check status

- [x] **Bootstrap SQL recorded verbatim in this completion doc** — see
      *Bootstrap procedure* above. Production-environment run is the
      operator's gating step before applying 039 in prod; the SQL itself
      is committed (in this doc) so any operator can reproduce the
      bootstrap deterministically.
- [x] **`make migrate-up` (now using `DIRECT_URL`) applies 039 cleanly**
      *after* the bootstrap. Locally rehearsed: bootstrap → `make
      migrate-up` runs the assertion guard green and applies all grants
      without error.
- [x] **`make migrate-down` reverses 039 cleanly** — locally rehearsed
      against the bootstrapped local DB. The `RAISE NOTICE` skip-path
      was also rehearsed against a DB where the bootstrap had been
      undone (`DROP ROLE`); the down migration emitted the notice and
      exited 0.
- [x] **`pg_roles` shows the right attributes** — verification SELECT
      above run locally post-bootstrap. `app_user`: `rolbypassrls = f`,
      `rolsuper = f`. `cron_user`: `rolbypassrls = t`, `rolsuper = f`.
- [x] **`information_schema.role_table_grants` reflects the §4.2
      narrow-write contract** — verified locally post-039 with:

      ```sql
      SELECT table_name, privilege_type
      FROM information_schema.role_table_grants
      WHERE grantee = 'cron_user'
        AND privilege_type IN ('INSERT', 'UPDATE', 'DELETE')
      ORDER BY table_name, privilege_type;
      ```

      Expected and observed (post-Phase 4):

      ```
       table_name        | privilege_type
      -------------------+----------------
       log_buffer        | DELETE
       monitoring_state  | INSERT
       monitoring_state  | UPDATE
      ```

      Three rows, exactly. Anything else is a regression and Phase 5's
      `cronWriteAllowlist` catalog test will fail.
- [x] **`DATABASE_URL` and `CRON_DATABASE_URL` still point at
      `postgres`.** No env change in Phase 4 — the topology is in place
      but unused.

Phase 5 (test infrastructure) is unblocked.

---

## Surprises and design notes

### One audit-discovered grant widening: `DELETE ON log_buffer`

Phase 1's §5.3 audit surfaced `agent/pruner.go` running
`PruneExpiredLogs` (a cross-tenant `DELETE FROM log_buffer WHERE
ingested_at < ...`) on a 1h cron schedule. The parent plan §4.2 was
SELECT-only on tenant tables. Two resolutions were possible:

1. **Add `DELETE ON log_buffer` to `cron_user`.** Keeps the pruner
   inside the cron-pool philosophy (cross-tenant enumerate-and-act
   shape). Costs: one verb, one table, audit-trail-visible in the
   migration.
2. **Move the pruner to a per-tenant `app_user` handoff.** Removes the
   cron-side write, but fans out into one transaction per tenant per
   tick. At even modest tenant counts that's a 30x connection-cost
   amplification for a janitorial sweep, plus per-tenant transaction
   bookkeeping where today there's one `DELETE`.

Resolution 1 won. The parent plan §4.2 itself anticipates this:
"Add further table-specific write grants here as the scheduler's needs
grow; keep the list auditable." The widening is now in the migration
under a code comment that names the audit reference and the writer
file path. The Phase 5 `cronWriteAllowlist` test gets `log_buffer →
{DELETE}` as a third entry alongside `monitoring_state →
{INSERT, UPDATE}`.

### `ALTER DEFAULT PRIVILEGES FOR ROLE postgres` versus implicit current_user

The parent plan's `ALTER DEFAULT PRIVILEGES` statements omit `FOR ROLE`,
which makes them apply to whatever `current_user` is when the statement
runs. Migration 039 is run via `DIRECT_URL` (= `postgres` today, =
`postgres` post-Phase-6 too — `DIRECT_URL` is the migrations-superuser
URL by design), so the implicit and explicit forms are functionally
identical today.

The explicit `FOR ROLE postgres` form is committed because the implicit
form silently breaks if some future migration (or some operator-side
restore) is run as a different superuser. The down migration mirrors
the explicit form exactly so the revoke matches the grant.

### The `DO` block over plain SQL for the assertion + PG17 branch

Two `DO $$ ... END $$` blocks bookend the migration. The pattern is
ugly but load-bearing:

- The bootstrap-assertion block needs `RAISE EXCEPTION` to fail the
  whole transaction with a custom message; that's plpgsql, not plain
  SQL.
- The PG17 conditional block needs `current_setting('server_version_num')::int`
  branching, which pure SQL also can't do without `CASE` over a query
  whose result is the migration text — uglier than the `DO` block.

Both blocks pay for themselves in clarity. golang-migrate handles
`DO $$ ... $$` cleanly; no special configuration needed.

### Function-execute grants are mostly defence-in-depth

`PUBLIC` already has `EXECUTE` on `app_current_user_id()` and
`app_user_org_ids()` because that's the default for `CREATE FUNCTION`
in Supabase's setup. The explicit `GRANT EXECUTE ... TO app_user,
cron_user` in 039 doesn't change behaviour today. The reason it's in
the migration anyway: a future hardening pass might `REVOKE EXECUTE
... FROM PUBLIC` on every helper as part of a Supabase-internal-API
audit. If that revoke lands without the explicit runtime-role grants
already in place, the runtime roles silently lose access to two helpers
the RLS policies depend on. Cheaper to commit the grants now than to
debug the regression later.

### Doc-side prerequisite from Phase 1, still pending

The roadmap references `docs/executing/rls-enforcement-mental-model.md`
(does not exist) and `docs/refs/trajan-db-roles.md` (lives in
`docs/executing/` not `docs/refs/`). Carried forward across Phases 1, 2,
and 3 completion docs. Still non-blocking for Phase 5; the roadmap
header links remain broken. Owner: doc-side PR before Phase 8.

---

## What Phase 5 needs from this

Phase 5 (test infrastructure) is now fully unblocked:

1. The `asAppUser` fixture can `SET LOCAL ROLE app_user` from a
   postgres-authenticated test transaction because the PG17 conditional
   grant block in 039 granted `app_user TO postgres WITH SET TRUE`
   (or the plain pre-PG17 form on older clusters).
2. The §5.6a catalog drift checks — `rolbypassrls`, `rolsuper`, the
   `cronWriteAllowlist` test — have a deterministic post-039 catalog
   to assert against. Initial allowlist content for Phase 5:

   ```go
   var cronWriteAllowlist = map[string]map[string]bool{
       "monitoring_state": {"INSERT": true, "UPDATE": true},
       "log_buffer":       {"DELETE": true},
   }
   ```

3. The §5.6 RLS regression tests can rely on `app_user` having
   `rolbypassrls = false` so the FORCE-flip in Phase 7 actually
   enforces against `app_user`-shaped reads.
4. Phase 5 also needs the §5.6c source-tree pairing tripwire's
   `rawDBAccessAllowlist` populated. Phase 1 §5.3 + §5.3a enumerated
   the surfaces; Phase 2 refactored them all to use `Pools.UserQueries`
   / `Pools.WithUserQueries` / `Pools.CronQueries`. Initial allowlist
   shape (file-by-file justifications):

   ```go
   var rawDBAccessAllowlist = map[string]string{
       // Pools wrappers — by definition the only place raw Pool.Begin lives.
       "internal/db/pools.go":                "primary chokepoint; UserQueries / CronQueries originate here",
       // Bootstrap / health probe — no tenant data touched.
       "internal/api/handlers/health.go":    "Pool.Ping() liveness check, no tenant data",
       "cmd/heimdall/main.go":               "pgxpool.New construction; assertRoleSplit's `SELECT current_user`",
       // Cron-pool handle — the legitimate non-UserQueries shape.
       // Files using `Pools.CronQueries()` enumerate; per-tenant work hands off.
   }
   ```

   Phase 5 is responsible for refining this into the final test-shape;
   the seed list is what Phase 2's refactor leaves behind.

The Phase 5 PR is test files plus two allowlist files (one for verb
coverage, one for raw-DB-access pairing). No production code changes.
