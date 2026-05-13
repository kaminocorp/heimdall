# RLS Enforcement — Phase 5 Completion: Test Infrastructure

**Status:** complete (pairing tripwire green in CI; DB-bound tests skip
without `DATABASE_URL`; flag-gated tests skip until Phases 6 / 7 turn
the flags on).
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 5.
**Phase 4 completion (gating doc):** [`./rls-enforcement-phase-4.md`](./rls-enforcement-phase-4.md)
**Date:** 2026-04-30.

---

## Executive summary

Phase 5 lands the test surface that the rest of the rollout is gated on.
Six new test files in `backend/internal/db/`, ~700 LOC total, structured
so each parent-plan subsection (§5.6, §5.6a, §5.6b, §5.6c, §5.6d) lives
in its own file and fails for one specific reason if it regresses.

The big wins:

- **The §5.6c source-tree pairing tripwire passes today with a
  three-entry allowlist** — `internal/db/pools.go`,
  `cmd/heimdall/main.go`, `internal/api/handlers/health.go`. Phase 2's
  Class D sweep was thorough enough that no other production file in
  `backend/internal/` or `backend/cmd/` does raw DB access without
  routing through `Pools.UserQueries` / `Pools.CronQueries` /
  `Pools.WithUserQueries`. The tripwire is now CI's catch for any
  future regression toward the pre-Phase-2 shape.
- **The §5.6a verb-coverage allowlist is the Phase-7 *target* state.**
  Tables Phase 1 §5.2 surfaced as RLS-enabled-without-policies
  (`agent_config`, `webhook_idempotency`, `log_pipeline_events`) are
  listed with the verb sets Migration B (Phase 7) will produce. The
  catalog-drift test fails until the policies land — Phase 7's
  acceptance check is "this test goes green."
- **The §5.6d `asAppUser` fixture is the load-bearing piece.** It opens
  a postgres-authenticated transaction, runs `SET LOCAL ROLE app_user`,
  and pins `app.current_user_id` to the caller. RLS evaluates as if
  the connection were `app_user` itself, so the regression suite fires
  meaningful enforcement assertions today — *before* Phase 6 flips
  `DATABASE_URL`. Migration 039's PG17 conditional `GRANT app_user TO
  postgres WITH SET TRUE` is what unblocks this fixture.

Three Phase-5-specific design decisions:

1. **Allowlists live in `_test.go` files alongside the tests that
   consume them.** The roadmap suggested standalone `.go` files
   (`rls_policy_allowlist.go`). Rejected. Allowlists are test data —
   shipping them in production binaries inflates the build artefact
   and creates a public API surface that nothing-not-tests should
   read. Co-locating allowlist + test means one PR diff captures
   "what we promised to enforce" + "the enforcement," and reviewers
   see drift and intent in the same place.
2. **Flag gates over `t.Skip`-with-magic-strings.** The §5.6a Test 1
   (FORCE coverage) and §5.6b connectivity tests need to be *written*
   today and *off* until Phase 7 / Phase 6 respectively. Two env-var
   flags carry the gate: `HEIMDALL_RLS_FORCE_LIVE=1` (Phase 7's
   responsibility to set in CI) and `HEIMDALL_ROLE_SPLIT_LIVE=1`
   (Phase 6's). Searchable, documented, and a pre-flip operator can
   manually flip the env var locally to run the tests against a
   prepared DB without code changes.
3. **`asAppUser` and `asCronUser` are package-scoped test helpers, not
   a separate `rlstest` package.** The roadmap's wording ("wrapper in
   `testhelpers_test.go`") suggests per-package adoption. Today only
   `internal/db` consumes the fixture, and the per-package shape avoids
   bringing test-only Postgres-membership semantics into the production
   import graph. If a future test surface (e.g. handler-level RLS
   regression) wants the fixture, the cheapest path is a small
   `internal/db/rlstest/` extraction — not done yet because there's
   no second consumer.

---

## What landed

### Test files

```
backend/internal/db/rls_helpers_test.go        — asAppUser, asCronUser, makeTenant, requireRLSEnv, requireRoleExists
backend/internal/db/rls_regression_test.go     — §5.6 cross-tenant RLS regression suite (5 critical tables)
backend/internal/db/rls_cron_test.go           — §5.6 cron-role behaviour (4 posture tests)
backend/internal/db/rls_catalog_test.go        — §5.6a catalog drift (5 tests + rlsPolicyAllowlist + cronWriteAllowlist)
backend/internal/db/rls_connectivity_test.go   — §5.6b connectivity (TestAppPool_IsActuallyAppUser, TestCronPool_IsActuallyCronUser)
backend/internal/db/rls_pairing_test.go        — §5.6c source-tree tripwire + rawDBAccessAllowlist
```

No production-code changes shipped. Phase 5 is test-only.

### §5.6d `asAppUser` / `asCronUser` fixtures

```go
asAppUser(t, ctx, pool, userID) → (pgx.Tx, cleanup)
asCronUser(t, ctx, pool)        → (pgx.Tx, cleanup)
```

Both open a postgres-authenticated transaction on the shared test pool,
run `SET LOCAL ROLE <role>`, and (for `asAppUser`) pin
`app.current_user_id`. The caller defers `cleanup()`, which rolls back
the transaction. Each fixture front-loads `requireRoleExists` so a
DB without Migration 039 applied skips with a clear message naming
`docs/completions/rls-enforcement-phase-4.md`.

### §5.6 RLS regression suite (5 critical tables)

`TestRLSRegression_CrossTenant` builds two synthetic tenants A and B
(via `makeTenant` — fresh user, org, app, connection) under postgres,
then per-table:

1. Insert a row owned by tenant B under postgres.
2. Open `asAppUser(tenantA.UserID)`.
3. `SELECT WHERE id = <tenantB row>` → expect `pgx.ErrNoRows`.

Coverage: `connections`, `log_buffer`, `conversations`,
`connection_sources`, `app_source_filters`. `log_pipeline_events` and
`webhook_idempotency` are deliberately excluded — they have no policies
yet (Phase 1 §5.2). Phase 7's Migration B adds policies AND the matching
regression subtests.

### §5.6 cron-role behaviour (4 tests)

`TestCronRole_Behaviour`:

- `EnumerateAcrossTenantsSucceeds` — cross-tenant SELECT returns rows.
- `DropTableDenied` — `DROP TABLE` returns `42501` (insufficient_privilege).
- `AuthUsersReadDenied` — `SELECT FROM auth.users` returns `42501`.
- `BlanketTenantTableInsertDenied` — `INSERT INTO log_buffer` returns
  `42501`. (Note: this is the BYPASSRLS / grants distinction the parent
  plan §4.2 spent paragraphs on. BYPASSRLS suppresses *RLS evaluation*;
  it doesn't add *grants*. Without `INSERT` in the cron grant set, the
  attempt fails at the privilege layer, not silently succeeds.)

Each subtest matches the SQLSTATE code (`42501`) rather than the human
message — the message text varies by locale; the code is contract.

### §5.6a catalog drift (5 tests)

| Test | What it pins |
|---|---|
| `TestCatalogDrift_ForceCoverage` | Every `rowsecurity = true` table has `forcerowsecurity = true`. **Skipped until `HEIMDALL_RLS_FORCE_LIVE=1` (Phase 7).** |
| `TestCatalogDrift_PolicyVerbCoverage` | Per-table verb-set matches `rlsPolicyAllowlist` exactly. Bidirectional — extra catalog rows AND missing catalog rows fail. |
| `TestCatalogDrift_AppUserAttributes` | `app_user.rolbypassrls = false`, `rolsuper = false`. |
| `TestCatalogDrift_CronUserAttributes` | `cron_user.rolsuper = false`, `rolbypassrls = true`. |
| `TestCatalogDrift_CronWriteAllowlist` | `information_schema.role_table_grants` for `cron_user` matches `cronWriteAllowlist` exactly. |

`rlsPolicyAllowlist` ships with 22 entries (the 18 tenant tables, 3
Phase-7 patch targets, and `schema_migrations` as the documented
"RLS-enabled with no policies, by design" exception). `cronWriteAllowlist`
is two entries: `monitoring_state → {INSERT, UPDATE}` (parent plan §4.2)
and `log_buffer → {DELETE}` (Phase 1 audit-discovered widening for the
pruner).

### §5.6b connectivity (2 tests)

`TestAppPool_IsActuallyAppUser` opens a fresh `pgxpool` from
`DATABASE_URL` and asserts `current_user = 'app_user'`. Sibling for
`CRON_DATABASE_URL` / `cron_user`. Both gated on
`HEIMDALL_ROLE_SPLIT_LIVE=1`. Phase 6 sets the flag in CI as part of
the env-var rollout — the test then becomes the catch for "did the
deploy environment actually pick up the new URL?"

### §5.6c source-tree pairing tripwire

`TestPairing_RawDBAccessIsAllowlisted` walks every `.go` file under
`backend/internal/` and `backend/cmd/` (excluding `*_test.go` and
`*.sql.go`), regex-matches against eight raw-DB-access patterns
(`pgxpool.New`, `Pool.Begin`, `Pool.Exec`, `Pool.Query`,
`Pool.QueryRow`, `Pool.Ping`, `App.Begin`, `Cron.Begin`), and asserts
each match either references the chokepoint API
(`UserQueries` / `WithUserQueries` / `CronQueries`) or is in
`rawDBAccessAllowlist`.

Today's allowlist:

```go
"internal/db/pools.go":                "primary chokepoint; UserQueries / CronQueries / WithUserQueries originate here"
"cmd/heimdall/main.go":                "pgxpool.NewWithConfig pool construction + assertRoleSplit's SELECT current_user; pre-Pools wiring"
"internal/api/handlers/health.go":     "Pool.Ping() liveness probe; no tenant data"
```

Plus a second sub-test (`TestPairing_AllowlistJustifications`) that
asserts every allowlist entry still names an existing file AND still
matches a raw-access pattern — catches stale allowlist entries left
behind after a refactor.

---

## Acceptance check status

- [x] **§5.6 RLS regression tests** — five subtests covering
      connections / log_buffer / conversations / connection_sources /
      app_source_filters, each using `asAppUser`. Will execute once a
      `DATABASE_URL` + Migration-039-applied DB is available; skip
      cleanly otherwise. Static-readable shape mirrors the parent plan.
- [x] **§5.6 cron-role behaviour tests** — four subtests, all asserting
      SQLSTATE 42501 for the negative cases.
- [x] **§5.6a Test 1 (FORCE coverage)** — written; gated off via
      `HEIMDALL_RLS_FORCE_LIVE`. Phase 7 turns it on. **Expected to be
      skipped today** — that's the whole point.
- [x] **§5.6a Tests 2–5 (verb coverage / role attrs / cron allowlist)** —
      written; runnable today against a Migration-039-applied DB.
- [x] **§5.6b connectivity tests** — written; gated off via
      `HEIMDALL_ROLE_SPLIT_LIVE`. Phase 6 turns them on.
- [x] **§5.6c pairing tripwire** — green in CI today against the
      three-entry `rawDBAccessAllowlist`. Verified via:

      ```
      $ go test ./internal/db/... -run TestPairing -v
      === RUN   TestPairing_RawDBAccessIsAllowlisted
      --- PASS: TestPairing_RawDBAccessIsAllowlisted (0.06s)
      === RUN   TestPairing_AllowlistJustifications
      --- PASS: TestPairing_AllowlistJustifications (0.00s)
      PASS
      ```
- [x] **§5.6d `asAppUser` fixture** — present in `rls_helpers_test.go`,
      consumed by every subtest in `TestRLSRegression_CrossTenant`.
      `asCronUser` sibling for the cron-role behaviour suite.
- [x] **Full backend suite still green.** `go test ./... ` returns OK
      across every package; integration tests skip cleanly when
      `DATABASE_URL` is absent (existing contract).

Phase 6 is unblocked.

---

## Surprises and design notes

### The pairing tripwire's allowlist is *small*

Pre-Phase-5 the back-of-envelope expectation (per the parent plan §5.6c
example) was an allowlist of 8–12 entries — every long-lived background
writer, every bearer-token handler, plus the chokepoint and the bootstrap
files. Phase 2's refactor was thorough enough that today's allowlist is
**three** entries: the chokepoint itself, `main.go` (pre-Pools wiring),
and `health.go` (`Pool.Ping`).

That's the strongest available evidence that Phase 2 didn't paper over
any raw-access drift. Every other production file under `backend/internal/`
and `backend/cmd/` either references `UserQueries` / `CronQueries` /
`WithUserQueries`, or doesn't open raw DB connections at all.

### `conversations` has no `app_id` column

The first draft of the regression suite tried to insert into
`conversations(id, user_id, app_id, title, messages)` — the `app_id`
column doesn't exist. Migrations 004 and 009 establish the table with
just `user_id`; no `app_id` was ever added. The regression test now
uses `(id, user_id, title, messages)`.

This is worth a future doc-side cross-check: the `conversations` policy
is `user_id = app_current_user_id()`, so a conversation belongs to a
user, not an app. That's structurally fine but slightly out of pattern
with the rest of the per-app data model. Not a Phase 5 problem; flag
for future review.

### Allowlist commits the Phase 7 *target state* deliberately

`rlsPolicyAllowlist` lists `agent_config`, `webhook_idempotency`, and
`log_pipeline_events` with `{"ALL"}` — verb sets Migration B will
produce. Today's catalog has zero policies on those tables, so
`TestCatalogDrift_PolicyVerbCoverage` will fail against any DB that
has Migration 039 applied but Migration 040 not yet.

This is the load-bearing pin on Phase 7. The roadmap acceptance check
is "Test 1 (FORCE coverage) goes green," but Test 2 (verb coverage)
is the same shape — the migration must add the policies the allowlist
already promises. If a Phase-7 PR ships FORCE without the policies,
Test 2 fails loudly with a per-table mismatch message.

If we ship Phase 5 to a CI environment that runs the catalog tests
against a real DB before Phase 7, Test 2 will go red. Two acceptable
mitigations:

1. **Don't run integration tests against a Migration-039-applied DB
   until Phase 7 is ready** — i.e. CI's integration job stays on the
   pre-039 schema until Phase 7 ships. Today's CI shape (skip when
   `DATABASE_URL` absent) achieves this passively.
2. **Add a `HEIMDALL_RLS_PHASE_5_TESTING=1` env-flag gate to Test 2**
   — symmetric with Tests 1 and the connectivity tests. Considered
   and rejected: Test 2 is *the* drift detector, and gating it would
   weaken Phase 5's value proposition. The right place to handle this
   is at the CI-job level, not in the test.

Documenting both options here so a future operator wiring up CI for
the role-split rollout has the choice in hand.

### `_realOSGetenv`-shaped indirection chain (now removed)

The first draft of `rls_catalog_test.go` defined `envFlagOn` through a
ten-deep chain of internal wrappers ending at `os.LookupEnv`, on the
theory that "isolating the os import to one line keeps the rest of the
file unentangled with the os package." That theory is wrong: it
inflates the file by ~50 LOC of nonsense, makes the test harder to
read, and provides zero diagnostic value when something fails. The
shipped version uses `import "os"` and `os.Getenv` directly. Mention
here as a reminder that "minimise imports" is not by itself a goal —
import isolation matters when the alternative is genuine coupling, not
when the alternative is calling `os.Getenv` with the same arity it
already has.

### `webhook_idempotency` has no per-row owner column

The Phase 7 work item Phase 1 §5.2 named for `webhook_idempotency` is
"scope by `connection_id IN (caller's connections)`." The table's
schema has `connection_id` as a column and a FK to `connections`, so
the policy body lands cleanly. Documented here so the Phase 7 author
doesn't have to re-derive it from `webhook_idempotency.up.sql` plus
`connections_org_member`.

### Doc-side prerequisite from Phase 1, still pending

`docs/executing/rls-enforcement-mental-model.md` (does not exist) and
`docs/refs/trajan-db-roles.md` (lives in `docs/executing/` not
`docs/refs/`). Carried forward across Phases 1, 2, 3, and 4 completion
docs. Still non-blocking for Phase 6; the roadmap header links remain
broken. Owner: doc-side PR before Phase 8.

---

## Files changed

```
backend/internal/db/rls_helpers_test.go         (new)
backend/internal/db/rls_regression_test.go      (new)
backend/internal/db/rls_cron_test.go            (new)
backend/internal/db/rls_catalog_test.go         (new)
backend/internal/db/rls_connectivity_test.go    (new)
backend/internal/db/rls_pairing_test.go         (new)
```

No production-code changes.

---

## What Phase 6 needs from this

Phase 6 (Stage 1: env-var flip to non-superuser) is now fully unblocked.
The test surface is ready to *meaningfully* assert correctness against
the post-flip topology:

1. **Local end-to-end rehearsal.** Phase 6's first task is to point
   local `DATABASE_URL` at `app_user` and `CRON_DATABASE_URL` at
   `cron_user`, then exercise every Phase 1 §5.3 path (webhook, OTLP,
   syslog, monitor loop, scheduled investigation). Every privilege
   miss surfaces as a `42501` error, attributable to a single file by
   the audit's own structure.
2. **Set `HEIMDALL_ROLE_SPLIT_LIVE=1` in CI** as part of the same
   commit that flips the env vars in production. The §5.6b
   connectivity tests then become the catch for "did the deploy
   environment actually pick up the new URL?"
3. **Test 1 (FORCE coverage) stays skipped.** Phase 6 doesn't ship
   `FORCE`; Phase 7 does. The skip message names the gating phase.
4. **Test 2 (verb coverage) stays the load-bearing pin for Phase 7.**
   It will fail today against a Migration-039-applied DB because the
   three Phase-7 policy gaps (`agent_config`, `webhook_idempotency`,
   `log_pipeline_events`) haven't been patched. Phase 6 either gates
   integration tests away from a Migration-039 DB, or accepts the red
   on those three rows and uses them as a daily reminder that Phase 7
   is the next stop. Recommendation: gate the catalog tests away from
   the integration DB in Phase 6 (CI-job-level decision) and let
   Phase 7 unblock them.

The Phase 6 PR is configuration only: env-var flip, CI flag set, plus
the manual rehearsal log in the completion doc. No code changes; Phase 5
absorbed the test surface, Phase 4 absorbed the migration, Phase 3
absorbed the wiring, Phase 2 absorbed the call-shape ripple.
