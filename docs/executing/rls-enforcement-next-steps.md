# RLS Enforcement — Manual Next Steps

**Status:** v0.48.0 (Phases 1–9) is structurally complete and committed. The
remaining work is operator-side: bootstrap two database roles in production,
apply migrations 039–041, flip three environment variables, run a 48h bake,
and clean up the gate-blockers from the pre-deploy review.

This doc is the single ordered checklist. Work top-to-bottom; do not skip steps.

**Key reference docs:**
- Bootstrap SQL canonical source: [`docs/completions/rls-enforcement-phase-4.md`](../completions/rls-enforcement-phase-4.md)
- Production flip runbook: [`docs/completions/rls-enforcement-phase-6.md`](../completions/rls-enforcement-phase-6.md)
- FORCE + policy patches: [`docs/completions/rls-enforcement-phase-7.md`](../completions/rls-enforcement-phase-7.md)
- Pre-deploy polish: [`docs/completions/rls-enforcement-phase-9.md`](../completions/rls-enforcement-phase-9.md)

---

## 0. Pre-flight gate-blockers (from the v0.48.0 review)

**Status: all closed in v0.48.1 + v0.48.2.** Section retained for historical
traceability — every item below shipped in code before the operator window
opens. See `docs/completions/rls-enforcement-phase-10.md` and the v0.48.2
changelog entry for the what/where/why of each fix.

- [x] **Regenerate sqlc.** Closed in v0.48.1 (Phase 10 C3). `users.sql`
      rewritten with `COALESCE(..., uuid.Nil)::uuid` sentinel so future
      regens are stable; `org_members.go::AddOrgMember` updated to compare
      against `uuid.Nil`. `log_pipeline_events.sql.go` deliberately *not*
      regenerated — the committed shape works; the sqlc v1.30
      nullable-inference regression for FILTER aggregates is documented as
      a known upstream wart, to revisit at next sqlc upgrade.
- [x] **Add `SET LOCAL statement_timeout = 0` to the loop variant of `userTx`**
      in `backend/internal/db/pools.go`. Closed in v0.48.1 (Phase 10 C2).
      Lives alongside the existing `idle_in_transaction_session_timeout = 0`
      in the `loop=true` branch (`pools.go:142, :146`).
- [x] **Always-on role logging at startup.** Closed in v0.48.1 (Phase 10 H1).
      `logPoolRoles` (`main.go:235`) runs unconditionally and emits a WARN
      ("set HEIMDALL_ENV=production to make this fatal") if both pools
      resolve to the same role. `assertRoleSplit`'s hard-fail behaviour
      under `HEIMDALL_ENV=production` is preserved.
- [x] **Webhook/OTLP TOCTOU guard.** Closed in v0.48.1 (Phase 10 H2) — but
      *not* via the zero-rows-affected guard originally proposed. The
      shipped fix is stronger: a `GetConnectionByUser` visibility re-check
      inside the user-scoped txn (`webhooks.go:362`, `otlp.go:139`). If
      the connection is no longer visible to the bearer-token's owner
      (deleted user, removed from org), the handler returns an explicit
      `401 Unauthorized` instead of a 5xx from a WITH CHECK violation.
      Behavioural coverage in `ingest_toctou_test.go`.
- [x] **Frontend ESLint clean.** Closed in v0.48.1 (Phase 10 C4). Nine
      production unused vars/imports fixed directly; 24 `any`-in-tests
      addressed by scoping `@typescript-eslint/no-explicit-any: 'off'` to
      `**/__tests__/**`, `**/*.test.{ts,tsx}`, and `src/test/**` via an
      override block in `eslint.config.js`. Production rules unchanged.
      `npm run lint` exits zero.
- [x] **Migration 040 vs 036 ordering.** Closed in v0.48.1 (Phase 10 C1).
      The `github_repos` policy patch + `ALTER TABLE … FORCE` block was
      removed from both `040.up.sql` and `040.down.sql`; only commentary
      remains explaining that 036 retired the table. `rls_catalog_test.go`'s
      `rlsPolicyAllowlist` no longer carries the `"github_repos"` entry.

---

## 1. Confirm Supabase pooler mode (load-bearing — do not skip)

The entire role-split depends on `SET LOCAL` and `set_config('app.current_user_id', …, true)`
surviving across queries within a transaction. Both only work in **session
pooling** mode. In transaction pooling mode every query lands on a fresh
server-side connection, the `app.current_user_id` GUC is empty, and `app_user`
+ `FORCE RLS` returns **zero rows for everything** — silently.

- [ ] In the Supabase dashboard, open **Project Settings → Database → Connection Pooling**.
- [ ] Confirm the connection string used for `DATABASE_URL` (the `app_user`
      pool) is **session mode**, not transaction mode. The session-mode URL
      uses port `5432`; transaction-mode uses port `6543`.
- [ ] Same check for `CRON_DATABASE_URL` (the `cron_user` pool). Session mode.
- [ ] `DIRECT_URL` should bypass the pooler entirely (direct connection to
      the database host on port `5432`).

If the pooler is currently in transaction mode for production traffic, switch
to session mode *before* the env-var flip in Step 4. Validate dev/staging
under session mode first.

---

## 2. Bootstrap the two non-superuser roles (Supabase Dashboard SQL Editor)

Roles are bootstrapped manually so passwords never enter version control.
Migration 039 will refuse to run until both roles exist.

**Generate two strong passwords first** (e.g. `openssl rand -base64 32` —
twice). Store them in your secrets manager (`APP_USER_PASSWORD`,
`CRON_USER_PASSWORD`) before running the SQL below; you will not be able to
recover them after creation.

- [ ] Open the Supabase dashboard for the **production** project.
- [ ] Navigate to **SQL Editor → New query**.
- [ ] Paste and run the block below, replacing `__APP_PW__` and `__CRON_PW__`
      with the freshly generated values:

      ```sql
      -- Heimdall RLS role split — production bootstrap.
      -- Source of truth: docs/completions/rls-enforcement-phase-4.md
      -- Do NOT commit this file with passwords filled in.

      CREATE ROLE app_user  LOGIN PASSWORD '__APP_PW__';
      CREATE ROLE cron_user LOGIN PASSWORD '__CRON_PW__' BYPASSRLS;

      -- Verify: app_user must show super=f, bypass=f.
      --        cron_user must show super=f, bypass=t.
      SELECT rolname, rolsuper, rolbypassrls
      FROM pg_roles
      WHERE rolname IN ('app_user', 'cron_user')
      ORDER BY rolname;
      ```

- [ ] Confirm the result table shows exactly:
      - `app_user`  — `rolsuper=f`, `rolbypassrls=f`
      - `cron_user` — `rolsuper=f`, `rolbypassrls=t`
- [ ] Store both passwords in the secrets manager (1Password / Doppler / Fly
      secrets — whichever production uses) under names `APP_USER_PASSWORD`
      and `CRON_USER_PASSWORD`.
- [ ] **Do the same in staging** if a separate Supabase project exists for
      staging — staging must mirror production for the bake to be meaningful.

> Local dev environments that already work as `postgres` superuser can skip
> bootstrap entirely. `make bootstrap-roles` is the local equivalent if you
> want a dev mirror of the prod role topology.

---

## 3. Apply migrations 039, 040, 041

Migrations must be applied via `DIRECT_URL` (the `postgres` superuser) — this
is the only path that retains DDL privileges.

- [ ] On staging first, source the env file and run:

      ```bash
      set -a && . ./.env.staging && set +a
      make migrate-up
      ```

- [ ] Confirm migrations 039, 040, 041 all apply without error. Specifically:
      - 039 — additive grants only (no REVOKE, no FORCE).
      - 040 — `FORCE ROW LEVEL SECURITY` on every RLS-enabled `public.*`
        table + the four policy patches Phase 1 surfaced. Holds
        `lock_timeout = 5s`.
      - 041 — Phase-8 invitee-lookup function + grant cleanup.
- [ ] Spot-check a few tables after 040 lands:
      ```sql
      SELECT relname, relrowsecurity, relforcerowsecurity
      FROM pg_class
      WHERE relkind = 'r' AND relnamespace = 'public'::regnamespace
      ORDER BY relname;
      ```
      Every `relrowsecurity = true` row should also have `relforcerowsecurity = true`.
- [ ] Run the Phase 5 connectivity tests against the staging URLs:
      ```bash
      HEIMDALL_ROLE_SPLIT_LIVE=1 \
      DATABASE_URL='postgres://app_user:…@…/postgres' \
      CRON_DATABASE_URL='postgres://cron_user:…@…/postgres' \
        go test ./backend/internal/db/... -run 'TestAppPool|TestCronPool|TestRLS|TestCatalogDrift'
      ```
      All should pass. `TestCatalogDrift_ForceCoverage` requires
      `HEIMDALL_RLS_FORCE_LIVE=1` — set it once 040 has applied.
- [ ] Repeat against **production** during the maintenance window. Do not
      proceed to Step 4 until staging has soaked under load for at least 24h.

---

## 4. Production env-var flip + `HEIMDALL_ENV=production`

Single combined deploy: flip the URLs and turn on `HEIMDALL_ENV=production` in
the same secrets-update transaction. Splitting them gives the misconfigured-
deploy failure mode a window to ship symptoms before `assertRoleSplit` fires.

- [ ] Compose the three URLs (replace `__APP_PW__`, `__CRON_PW__`, `__SUPER_PW__`
      with the values from your secrets manager):

      ```
      DATABASE_URL='postgres://app_user:__APP_PW__@db.<project>.supabase.co:5432/postgres'
      CRON_DATABASE_URL='postgres://cron_user:__CRON_PW__@db.<project>.supabase.co:5432/postgres'
      DIRECT_URL='postgres://postgres:__SUPER_PW__@db.<project>.supabase.co:5432/postgres'
      ```

      Both runtime URLs point at the **session-mode** pooler endpoint
      confirmed in Step 1 (port `5432`, NOT `6543`). `DIRECT_URL` bypasses
      the pooler.
- [ ] Update Fly secrets in one batch:

      ```bash
      fly secrets set \
        DATABASE_URL='…' \
        CRON_DATABASE_URL='…' \
        DIRECT_URL='…' \
        HEIMDALL_ENV='production'
      ```

      `fly secrets set` triggers a rolling restart automatically.
- [ ] Watch the boot log of the first restarted machine. The success line is:

      ```
      level=INFO msg="role split verified" app_role=app_user cron_role=cron_user
      ```

      If `assertRoleSplit` finds both pools resolving to the same role, the
      process exits with a clear error and Fly will mark the deploy failed
      before any traffic is accepted. This is the safety net.
- [ ] After all machines roll, exercise the live URL connectivity test against
      production (Step 3 command, prod URLs).
- [ ] Smoke test from the UI: open the dashboard, send a chat message, ingest
      one webhook log, view the activity feed. All must work.

---

## 5. 48h bake + watch list

After the flip, watch for these specific failure shapes for **48 hours** before
declaring the rollout closed:

- [ ] **Permission errors in app logs** (`SQLSTATE 42501`). Any indicates a
      missing grant; check whether the operation belongs in `app_user`'s
      grant set (CRUD on `public.*`) or `cron_user`'s narrow write set
      (`monitoring_state` insert/update, `log_buffer` delete).
- [ ] **Zero-row reads where rows should exist.** Symptoms: dashboard pages
      empty for users who definitely have data; webhooks return 201 but
      activity feed stays empty. Either pooler is in transaction mode (Step 1
      regression) or `app.current_user_id` is not being set somewhere.
- [ ] **`idle_in_transaction_session_timeout` aborts.** Should be impossible
      given `SET LOCAL idle_in_transaction_session_timeout = 0` in
      `userTx(loop=true)`, but watch for chat loop transactions terminating
      mid-LLM-roundtrip.
- [ ] **`statement_timeout` aborts mid-tool-call.** If Step 0's
      `SET LOCAL statement_timeout = 0` patch did not land, the *next* tool
      query inside a long chat loop will fail at 8s.
- [ ] **Cron pool errors.** Background monitor / scheduler ticks failing —
      surface as missed monitoring runs in the activity feed.

Revert lever: a single `fly secrets set DATABASE_URL='postgres://postgres:…' CRON_DATABASE_URL='postgres://postgres:…' HEIMDALL_ENV=''`
restores pre-flip topology in <60s. Migrations 039–041 do not need to be
rolled back to revert; they are compatible with the postgres-superuser pool
shape too.

---

## 6. Post-bake cleanup

When the bake closes clean (no escalations, no permission errors for 48h):

- [ ] Move this file from `docs/executing/` to `docs/archive/` (or delete it;
      the completion docs in `docs/completions/rls-enforcement-phase-{1..9}.md`
      remain the runbook-grade record).
- [ ] Update the bake-result table at the bottom of
      `docs/completions/rls-enforcement-phase-6.md` with the actual flip date
      and any incidents observed.
- [ ] Add a changelog entry under v0.48.0 noting the production flip date.
- [ ] Rotate the role passwords once on a quarterly cadence going forward
      (operator decision; not part of this rollout).

---

## Quick reference — what each role can do

| Role        | Auth as           | RLS    | DDL | Tenant CRUD | Cross-tenant SELECT | Narrow writes |
|-------------|-------------------|--------|-----|-------------|---------------------|---------------|
| `postgres`  | `DIRECT_URL`      | bypass | yes | yes         | yes                 | n/a (super)   |
| `app_user`  | `DATABASE_URL`    | FORCE  | no  | yes         | no                  | n/a           |
| `cron_user` | `CRON_DATABASE_URL` | bypass | no | no (no INSERT) | yes              | `monitoring_state` insert/update, `log_buffer` delete |

If a future feature needs `cron_user` to write a new table, that's an
explicit grant in a new migration — which is an explicit security review.
This is the security model.
