# RLS Enforcement — Phase 3 Completion: Pool Wiring (Dual Pool, Single Role)

**Status:** complete.
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 3.
**Phase 2 completion (gating doc):** [`./rls-enforcement-phase-2.md`](./rls-enforcement-phase-2.md)
**Date:** 2026-04-30.

---

## Executive summary

Phase 3 promotes the App / Cron pool topology from "one `*pgxpool.Pool`
passed twice to `db.NewPools`" to "two physical pools with explicit
sizing." Both pools still authenticate as the same `postgres` superuser
today — the env-var split is Phase 6's job — but every other piece of
the eventual topology is now in place:

- Two `pgxpool.Pool` handles constructed side-by-side, sized App=30 /
  Cron=5 per parent plan §3.
- Three documented runtime URLs (`DATABASE_URL`, `CRON_DATABASE_URL`,
  `DIRECT_URL`) with sensible fall-throughs that keep current
  configurations working unchanged.
- A dormant production-mode startup invariant that refuses to launch
  when both pools authenticate as the same role. Gated on
  `HEIMDALL_ENV=production`; left unset everywhere today, so the
  tripwire fires only when an operator explicitly opts in (Phase 6
  flips it on after the URLs actually differ).
- `make migrate-up` / `make migrate-down` now read `DIRECT_URL` (with
  fall-through to `DATABASE_URL`), pre-positioning the migration path
  to keep working when Phase 6 downgrades `DATABASE_URL` to a
  non-superuser role.

Privilege change: zero. Behaviour change: zero. The signal that the
refactor is correct is that the existing test suite still passes
green and `make dev-backend` boots with two pools constructed.

---

## Three Phase-3-specific design decisions

1. **Two physical pools today, not a shared handle.** Phase 2 left
   `db.NewPools(pool, pool)` in place — one physical pool reused. Phase 3
   could have kept that shape (one pool when URLs match, two when they
   differ) and only physically split when Phase 6 sets `CRON_DATABASE_URL`.
   Rejected. Sharing a handle means the per-pool `MaxConns` (App=30,
   Cron=5) can't diverge until Phase 6, which means we'd be testing an
   untested topology at flip time. Always-two-physical-pools today costs
   one extra idle pgxpool struct (and lazy connections up to the cron
   ceiling) but exercises the eventual shape continuously. The dev cost
   is bounded by `MaxConns=5` on the cron pool — no observable resource
   increase.
2. **Production invariant queries `current_user`, not URL strings.**
   The parent plan called for "refuse to launch when DATABASE_URL and
   CRON_DATABASE_URL resolve to the same role." A naive implementation
   would string-compare the two URLs. Rejected. Two distinct URLs can
   still both authenticate as `postgres` (different `?application_name=`,
   different `host=` aliases for the same DB, etc.); two identical URLs
   could in principle resolve differently if a `pgbouncer` rewrite is
   in play. The invariant uses `SELECT current_user` on each pool and
   compares the actual authenticated role — the strongest possible
   signal that the env-var flip landed.
3. **Invariant is dormant by default.** Today every prod environment
   has both URLs resolving to `postgres`. If we made the invariant
   unconditional, Phase 3 would brick prod immediately. Gated on
   `HEIMDALL_ENV=production`, set nowhere yet. Phase 6 turns it on
   (after setting `CRON_DATABASE_URL`) so the first prod deploy that
   has the role split actually verified gets the tripwire. The
   acceptance check ("Production-mode startup refuses to launch when
   both URLs are identical") was rehearsed locally per the
   "verified locally with a fake `HEIMDALL_ENV=production` boot"
   roadmap clause — see *Local rehearsal* below.

---

## What landed

### Configuration

**`backend/internal/config/config.go`**

Three new fields on `Config`, each with documented Phase 6 end-states:

- `CronDatabaseURL` (`CRON_DATABASE_URL`) — cron-pool URL.
- `DirectURL` (`DIRECT_URL`) — superuser URL for migrations.
- `Environment` (`HEIMDALL_ENV`) — gates startup invariants.

`Validate()` was deliberately *not* extended — none of the new vars are
required today. The fall-through behaviour is documented in the field
comments and surfaced in startup logs (the `WARN` for missing
`CRON_DATABASE_URL`).

### Pool construction

**`backend/cmd/heimdall/main.go`**

Replaced the single `pgxpool.New(ctx, cfg.DatabaseURL)` call with two
calls through a new `buildPool` helper:

```go
appPool, err := buildPool(ctx, cfg.DatabaseURL, 30, "app")
…
cronURL := cfg.CronDatabaseURL
if cronURL == "" {
    slog.Warn("CRON_DATABASE_URL unset; falling back to DATABASE_URL until Phase 6 of the RLS role split")
    cronURL = cfg.DatabaseURL
}
cronPool, err := buildPool(ctx, cronURL, 5, "cron")
```

`buildPool` parses via `pgxpool.ParseConfig`, sets `MaxConns` explicitly,
opens via `pgxpool.NewWithConfig`, and emits one `slog.Info` per pool
with the label and ceiling. Both pools get `defer Close()`; on cron-pool
failure, `appPool.Close()` runs eagerly to keep the failure path tidy.

`pools := db.NewPools(appPool, cronPool)` replaces the
`db.NewPools(pool, pool)` shape. Every downstream call site stays
unchanged because Phase 2 already routed everything through the
`*db.Pools` type.

### Production invariant

`assertRoleSplit(ctx, appPool, cronPool)` runs `SELECT current_user`
against each pool and returns an error when the two roles match.
Called only when `cfg.Environment == "production"`. The return value
is logged via `slog.Error` and the process exits 1 — same shape as
the existing `cfg.Validate()` failure path. Success path emits a
`slog.Info` with both role names so the post-Phase-6 boot logs
demonstrably record "we are running on app_user / cron_user."

### Documented surfaces

| File | Change |
|---|---|
| `.env.example` | Three new entries (`CRON_DATABASE_URL`, `DIRECT_URL`, `HEIMDALL_ENV`) with Phase 6 end-state comments. |
| `Makefile` | `migrate-up` / `migrate-down` read `$${DIRECT_URL:-$$DATABASE_URL}` instead of `$$DATABASE_URL`. `migrate-create` and `sqlc-generate` are unchanged — the former doesn't open a DB connection (it just creates files), and `sqlc generate` reads no env (it only consumes `sqlc.yaml` + the SQL files on disk). |
| `CLAUDE.md` | New "Database URLs (RLS role split)" subsection with one-line role descriptions and the `HEIMDALL_ENV` posture. The existing "requires DATABASE_URL" line in the Database section now reads "requires DIRECT_URL (falls back to DATABASE_URL when unset)." |
| `README.md` | One-line tweak in the commands block to match the new fallback. |

---

## Acceptance check status

- [x] `make dev-backend` boots with both pools constructed (verified
      via local rebuild + boot — one `database pool ready` line per pool
      in the startup log, sizing 30/5 reported).
- [x] All backend tests still pass (`go test ./...` green; integration
      tests that need `DATABASE_URL` skip cleanly when absent — same
      contract as before).
- [x] `go vet ./...` clean.
- [x] Pool sizes match parent plan §3 (App=30, Cron=5; emitted in
      startup logs).
- [x] Production-mode startup refuses to launch when both URLs resolve
      to the same role (verified locally with `HEIMDALL_ENV=production`
      and `DATABASE_URL == CRON_DATABASE_URL` — `assertRoleSplit`
      returned `… both authenticate as "postgres"; production requires
      distinct app_user / cron_user roles` and exited 1).
- [x] `make migrate-up` still works via the `DIRECT_URL` fallback (no
      `.env` change required; `DIRECT_URL` is empty so the Makefile's
      `${DIRECT_URL:-$DATABASE_URL}` substitution lands on `DATABASE_URL`
      — bit-identical to the pre-Phase-3 invocation).

Phase 4 is unblocked.

---

## Surprises and design notes

### `defer pool.Close()` ordering on the failure path

The pre-Phase-3 main.go had one `defer pool.Close()` directly after the
single pool construction — guaranteed to run if any later step exited
the function. Phase 3 has two pools and an early-exit between them
(if `buildPool(cron)` fails, `appPool` is already constructed but
`defer cronPool.Close()` doesn't yet exist). The fix in main.go is an
eager `appPool.Close()` inside the cron-failure branch *before*
`os.Exit(1)`. Slightly ugly; the alternative (extract the whole
two-pool construction into a helper that returns both or neither)
would have been cleaner but introduces a new function in main.go for
a five-line gain. Kept inline; the comment in the failure branch
makes the intent obvious.

### `HEIMDALL_ENV` is environment classification, not feature flag

Tempting to add other feature flags onto this same env var ("set
`HEIMDALL_ENV=production` to enable Sentry, structured logs, …"). Did
not. `HEIMDALL_ENV` exists today specifically to gate startup
*invariants* — checks that should be tighter in production than in
dev. `LOG_FORMAT` and the existing CORS knobs already cover the
"different defaults in prod" axis without overloading this var.
Future invariants (e.g. "refuse to launch when migrations are
behind") can layer onto the same env without confusion.

### Cron-pool fallback is a `Warn`, not a hard error

Leaving `CRON_DATABASE_URL` unset was a hard error in an earlier
draft. Backed off. The fallback is *the entire point* of the
transitional Phase-3 shape — every operator booting between Phase 3
ship and Phase 6 flip will hit this code path. A `Warn` puts the
"this is the transitional shape" message in every boot log without
breaking anyone. Phase 6 deletes the `Warn` (and probably hardens
the empty-cron-url case to an error) when the env is universally
populated.

### `sqlc generate` was never going to need `DIRECT_URL`

Phase 3 §1.5 of the roadmap said "any `sqlc-generate` target that
touches the DB to read `DIRECT_URL` instead of `DATABASE_URL`."
Inspection of `sqlc.yaml` confirms `sqlc generate` reads only the
on-disk SQL queries and migrations to produce Go code — no live DB
connection at any point. The Makefile target needs no change.
Documented here so a future re-audit doesn't reopen the question.

### Test fixtures kept the single-pool shape

`testhelpers_test.go` builds `pools := db.NewPools(pool, pool)` —
one physical pool, passed twice. Phase 3 deliberately did not touch
this. The test environment is already sized for a single Postgres
connection and exercising the two-physical-pools path in tests would
double the DB connection pressure for zero diagnostic value (the
pool wiring is exercised by `make dev-backend` boot). When Phase 6
ships and the App/Cron pools authenticate as different roles in
production, test fixtures will continue running both as `postgres` —
that's the existing `asAppUser` story (Phase 5) and out of scope for
Phase 3.

### Doc-side prerequisite from Phase 1 still pending

The roadmap references `docs/executing/rls-enforcement-mental-model.md`
and `docs/refs/trajan-db-roles.md`. Mental-model doc still doesn't
exist; `trajan-db-roles.md` still lives in `docs/executing/` rather
than `docs/refs/`. Carried forward from the Phase 1 / 2 completion
docs. Non-blocking for Phase 4; the roadmap header links remain
broken. Owner: doc-side PR before Phase 8.

---

## Files changed

```
backend/cmd/heimdall/main.go
backend/internal/config/config.go
.env.example
Makefile
CLAUDE.md
README.md
```

## Local rehearsal log

```
$ # default boot — both URLs unset for cron / direct
$ make dev-backend
… database pool ready label=app max_conns=30
WARN  CRON_DATABASE_URL unset; falling back to DATABASE_URL until Phase 6 of the RLS role split
… database pool ready label=cron max_conns=5
… starting server port=8080

$ # production-mode boot with same URL on both — must abort
$ HEIMDALL_ENV=production CRON_DATABASE_URL="$DATABASE_URL" make dev-backend
… database pool ready label=app max_conns=30
… database pool ready label=cron max_conns=5
ERROR  role-split startup invariant failed err=DATABASE_URL and CRON_DATABASE_URL both authenticate as "postgres"; production requires distinct app_user / cron_user roles
exit status 1
```

---

## What Phase 4 needs from this

Phase 4 (Migration A — role grants) is now mostly pre-staged. The
remaining work is database-side, not code-side:

1. **Manual bootstrap** — create `app_user` and `cron_user` roles with
   passwords (per parent plan §3 / Trajan precedent: bootstrap, not
   migration).
2. **Migration `039_runtime_role_grants.up.sql`** — assertion guard +
   per-role grant set per parent plan §4.1 / §4.2.
3. **PG17 conditional grant block** — `GRANT app_user TO postgres
   WITH SET TRUE` (and likewise for cron_user) for the Phase 5 test
   fixture.
4. **Audit-discovered widening** — `cron_user` needs `DELETE ON
   log_buffer` (the pruner's grant), per Phase 1 / Phase 2's
   reaffirmed surprise. Goes into Migration A's grant set, not a
   separate migration.

The Phase 4 PR is migration files + a bootstrap-procedure runbook in
the completion doc. No code changes; Phase 3 absorbed the wiring,
Phase 2 absorbed the call-shape ripple. Phase 4 is the first phase
that actually changes database state beyond the runtime catalog.
