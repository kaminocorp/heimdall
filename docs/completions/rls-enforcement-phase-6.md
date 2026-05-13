# RLS Enforcement — Phase 6 Completion: Stage 1 Env-Var Flip

**Status:** complete (rehearsal procedure + production runbook + revert
lever locked); production flip and 48h bake pending operator window. The
runbook below is the gating step before Phase 7; the bake-result table
at the bottom of this doc is filled in when the bake closes clean.
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 6.
**Phase 5 completion (gating doc):** [`./rls-enforcement-phase-5.md`](./rls-enforcement-phase-5.md)
**Date:** 2026-05-02.

---

## Executive summary

Phase 6 flips production runtime traffic off the `postgres` superuser pool
and onto the two non-superuser roles Phase 4 created. After the flip:

- `DATABASE_URL` authenticates as `app_user` (RLS-enforced; no DDL; no
  `auth.*` access).
- `CRON_DATABASE_URL` authenticates as `cron_user` (BYPASSRLS for
  cross-tenant enumeration; narrow writes — `INSERT/UPDATE` on
  `monitoring_state`, `DELETE` on `log_buffer`; no other writes; no DDL).
- `DIRECT_URL` continues to authenticate as `postgres` (migrations only).

RLS itself is **still cosmetic** for `app_user` until Phase 7 ships
`FORCE`. That separation is the entire diagnostic argument behind the
two-stage rollout: any failure surfaced during this 48h bake is by
construction a *privilege-set* miss (a write going to the wrong pool,
a DDL we missed in §5.1, a helper that needs an explicit `EXECUTE`
grant), not a *policy* miss. Phase 7's bake handles the latter.

The Phase 3 invariant (`backend/cmd/heimdall/main.go:79–84`) is the
load-bearing safety net: with `HEIMDALL_ENV=production` set in the
*same* secrets-update transaction as the URL flip, the first post-flip
boot calls `assertRoleSplit`, which `SELECT current_user`s on each pool
and refuses to launch if both still resolve to the same role. Operator
ergonomics: a misconfigured deploy fails before traffic is accepted,
not after symptoms appear.

Three Phase-6-specific design decisions, locked in this commit:

1. **Single combined deploy: URL flip + `HEIMDALL_ENV=production` in
   one secrets transaction.** The split-deploy alternative (flip URLs
   first, then turn on the invariant later) was rejected: it gives the
   misconfigured-deploy failure mode a window to ship symptoms before
   the safety net fires. The `assertRoleSplit` boot-time check was
   built specifically for this moment; flipping `HEIMDALL_ENV` and the
   URLs in one secrets-update means the *first* boot post-flip
   verifies the role split, and any prior boot continues to skip the
   check (matching the dormant-tripwire posture Phase 3 shipped).
2. **Connectivity test runs as a manual deploy-time smoke gate, not
   in CI.** There is no `.github/workflows/` in this repo. The
   `HEIMDALL_ROLE_SPLIT_LIVE=1` flag the Phase 5 connectivity tests
   key off was originally framed as a CI gate; with no CI to wire it
   into, the realistic shape is an operator-invoked test against the
   live prod URLs (`HEIMDALL_ROLE_SPLIT_LIVE=1 DATABASE_URL=…
   CRON_DATABASE_URL=… go test ./backend/internal/db/... -run
   TestAppPool TestCronPool`). Documented in the *Production flip
   runbook* below as step 5.
3. **Two-step verification: boot log first, connectivity test second.**
   `assertRoleSplit`'s "role-split verified" `slog.Info` line is the
   primary signal that the flip landed. The Phase 5 connectivity test
   is the secondary signal — same shape but from outside the running
   process, exercised against fresh `pgxpool.New` calls. Keeping them
   independent means a regression in either path catches what the
   other might miss (e.g. `assertRoleSplit` runs on the long-lived
   pools the app actually uses; the connectivity test pulls a fresh
   pool from the same URL, so a deployment that somehow uses two
   different connection strings at boot vs. test time would surface).

---

## What landed

### Documentation

- **`docs/completions/rls-enforcement-phase-6.md`** (this file) — the
  rehearsal procedure, production flip runbook, watch list, and revert
  lever. Lifted directly from the roadmap §6 acceptance language and
  expanded with Heimdall-specific Fly.io invocation shapes.
- **`.env.example`** — added a commented "Phase 6 example fill" stanza
  showing the post-Phase-6 URL shape (`postgres://app_user:…`,
  `postgres://cron_user:…`) so an operator filling the file has the
  target form visible inline. The active uncommented values still mirror
  the pre-flip state — copy-paste a fresh `.env.example` produces a
  working pre-flip dev env, exactly as before.

### No production-code changes

Phase 2 absorbed the call-shape ripple. Phase 3 absorbed the wiring.
Phase 4 absorbed the migration. Phase 5 absorbed the test surface.
Phase 6's job is to set three environment variables and one
`HEIMDALL_ENV` flag in production, then watch. Code-side, the only
thing left to do is *not regress* — and the Phase 5 §5.6c pairing
tripwire is what catches future regression toward the pre-Phase-2
shape.

---

## Local rehearsal procedure (verbatim)

The roadmap's first Phase 6 task: "Point local `DATABASE_URL` and
`CRON_DATABASE_URL` at `app_user` and `cron_user` respectively. Manually
trigger every path in Phase 1's §5.3 audit." This is the only
pre-production test of the actual privilege topology — Heimdall has no
staging environment, so the rehearsal here is the only thing that gets
to fail under controlled conditions.

### Prerequisites

1. Local Postgres running (`docker compose up -d postgres`).
2. Migrations 001–039 applied locally via `make migrate-up` (which now
   reads `DIRECT_URL` per Phase 3, falling back to `DATABASE_URL`).
3. Bootstrap from Phase 4 §*Bootstrap procedure* run locally (creates
   `app_user` + `cron_user` with local-only passwords).

### Step 1 — Flip local env vars

In a copy of `.env` for rehearsal (do **not** check this in):

```bash
DIRECT_URL=postgres://postgres:postgres@localhost:5432/heimdall?sslmode=disable
DATABASE_URL=postgres://app_user:<local-app-password>@localhost:5432/heimdall?sslmode=disable
CRON_DATABASE_URL=postgres://cron_user:<local-cron-password>@localhost:5432/heimdall?sslmode=disable
HEIMDALL_ENV=production
```

### Step 2 — Boot the backend

```bash
make dev-backend
```

Expected boot log (the load-bearing line is the third one):

```
… database pool ready label=app  max_conns=30
… database pool ready label=cron max_conns=5
… role-split verified app_role=app_user cron_role=cron_user
… starting server port=8080
```

If the third line is missing or the process exits 1 with
`role-split startup invariant failed`, the URL flip didn't land —
inspect `.env`, do not proceed.

### Step 3 — Exercise every Phase 1 §5.3 path

For each path, look for `42501 permission denied`, `pgx.ErrNoRows`
where rows are expected, or sudden silent-zero-row behaviour.

| Path | How to exercise locally | Pass condition |
|---|---|---|
| Webhook ingestion | `curl -X POST http://localhost:8080/api/webhooks/logs -H 'Authorization: Bearer <local-conn-token>' -d '{"entries":[{"message":"rehearsal","level":"info","timestamp":"<rfc3339>"}]}'` | 202 + `log_buffer` row visible to the owning user via `asAppUser` fixture; idempotency cache row written |
| OTLP ingestion | Send an OTLP-HTTP payload at `/api/otlp/v1/logs` with the same bearer token shape | 200 + `log_buffer` rows |
| Syslog TLS listener | Bring up a syslog connection in the UI, fire a TLS-TCP message at the listener port | `log_buffer` row written; `connection_sources` upsert succeeds |
| Monitor loop iteration | Insert a fresh log row owned by an active app, wait one 15s tick | `agent_log` and `log_pipeline_events` rows materialise; `monitoring_state.last_monitored_at` advances |
| Scheduled investigation | Configure a 1-minute schedule, wait | `agent_log` entry typed `investigation_*`; `investigation_schedules.last_run_at` advances |
| Pruner sweep | Insert a `log_buffer` row with `ingested_at = now() - interval '49 hours'`, wait one 1h tick (or invoke the prune handler manually) | The row is gone — no `42501` |
| GitHub install flow | Click through a GitHub App install (requires GH App configured locally) | 200 + `github_repos` rows insertable |
| Notifications dispatch | Trigger a flagged log via the monitor loop with notification channels configured | `notification_log` row written; downstream send succeeds |
| Interactive chat | Open a chat WebSocket; send a prompt that uses `search_logs` and `query_database` | Tools return rows scoped to the caller's tenant; conversation persists |

Anything in this list that fails is a Phase 6 prerequisite that escaped
Phase 1's audit and Phase 2's refactor. Track each as a follow-up PR
that re-runs the relevant audit slice; do not paper over by widening
grants in Migration 039 unless the failure is structurally cron-side
(like the pruner's `DELETE on log_buffer` widening Phase 4 already
covered).

### Step 4 — Local revert rehearsal

Revert `.env` to the pre-flip state (`DATABASE_URL` → `postgres`,
`CRON_DATABASE_URL` blank, `HEIMDALL_ENV` blank), restart the backend,
confirm boot is unchanged from pre-Phase-6 (the WARN about
`CRON_DATABASE_URL` falling back to `DATABASE_URL` reappears; no
`role-split verified` line). This rehearses the production revert
path in <60 seconds end-to-end.

---

## Production flip runbook (verbatim)

Heimdall's production deploy is on Fly.io (`backend/fly.toml`). Adapt
the secrets-update commands if the deploy surface changes.

### Step 1 — Pre-flight

- [ ] Phase 5 acceptance checks all green in CI / locally
      (`go test ./backend/internal/db/... -run TestPairing
      TestCatalogDrift_AppUserAttributes TestCatalogDrift_CronUserAttributes
      TestCatalogDrift_CronWriteAllowlist`).
- [ ] Production bootstrap (Phase 4 §*Bootstrap procedure*) run against
      prod, verified via the post-bootstrap `pg_roles` SELECT.
- [ ] Migration 039 applied in prod via `make migrate-up` against
      `DIRECT_URL = postgres://postgres:…`.
- [ ] Migration 040's down-migration draft (Phase 7's responsibility)
      rehearsed locally — the roadmap requires this before Phase 7
      starts; piggyback the rehearsal on the Phase 6 bake window.
- [ ] Watch dashboards (whatever the team uses for `agent_log` and
      `log_pipeline_events` throughput) bookmarked, with the pre-flip
      baseline rates noted (so the ±5% acceptance check is concrete).
- [ ] Local rehearsal (above) green end-to-end; any §5.3 path that
      surfaced `42501` has a fix shipped.

### Step 2 — The flip

One Fly.io secrets-update transaction (a single `flyctl secrets set`
sets multiple values atomically):

```bash
flyctl secrets set \
  --app heimdall \
  DATABASE_URL="postgres://app_user:$(op read 'op://Heimdall/app_user/password')@<prod-host>:5432/<db>?sslmode=require" \
  CRON_DATABASE_URL="postgres://cron_user:$(op read 'op://Heimdall/cron_user/password')@<prod-host>:5432/<db>?sslmode=require" \
  DIRECT_URL="postgres://postgres:$(op read 'op://Heimdall/postgres/password')@<prod-host>:5432/<db>?sslmode=require" \
  HEIMDALL_ENV=production
```

(Substitute `op read` for whatever secrets-manager wrapper the operator
uses — the only invariant is "no plaintext password in shell history.")

`flyctl secrets set` triggers an automatic redeploy by default. The
post-flip boot log on the new machine **must** include:

```
role-split verified app_role=app_user cron_role=cron_user
```

If this line is missing — or if the boot log shows
`role-split startup invariant failed` — the deploy machine refused
to start traffic. Fly's health checks fail, the previous deploy
remains live, no user impact. Diagnose, fix, redeploy. The dormant
tripwire from Phase 3 is doing its job.

### Step 3 — Post-flip smoke gate (operator-invoked)

Run the Phase 5 §5.6b connectivity tests against the live prod URLs
from a trusted local checkout:

```bash
HEIMDALL_ROLE_SPLIT_LIVE=1 \
DATABASE_URL="$(flyctl secrets list --app heimdall --json | jq -r '…')" \
CRON_DATABASE_URL="$(…)" \
go test ./backend/internal/db/... -run TestAppPool_IsActuallyAppUser
go test ./backend/internal/db/... -run TestCronPool_IsActuallyCronUser
```

Expected: both tests PASS. If either fails with
`current_user must authenticate as app_user; got "postgres"` or the
cron sibling, `flyctl secrets list` will show the problem (the URL
in question still references the old role).

`HEIMDALL_ROLE_SPLIT_LIVE=1` stays in operator-invoked-only territory
— the prod URLs are not in CI, so the flag has no CI consumer to wire
into.

### Step 4 — 48-hour bake watch list

Watch the production logs continuously for 48 hours. The acceptance
gates (parent plan §5.7 / roadmap Phase 6 acceptance):

| Signal | Pass | Fail action |
|---|---|---|
| `42501 permission denied` errors | **zero** | Locate the file via the surrounding stack frame; classify as §5.1 miss (DDL or `auth.*` read under `app_user`), §5.3 miss (writer not on `UserQueries`), or grant gap (function `EXECUTE`). Fix and redeploy; if the fix is non-trivial, execute the revert lever (Step 5). |
| `agent_log` insertion rate | within ±5% of pre-flip baseline | A sustained drop = a background writer is hitting `cron_user` (no INSERT grant on tenant tables) instead of handing off via `UserQueries`. Locate via the EmitLog call site; the §5.6c pairing tripwire would have caught this if it were a raw-pool regression — so it's likely a missed `UserQueries` wrap inside one of the writers. |
| `log_pipeline_events` insertion rate | within ±5% of pre-flip baseline | Same shape as above; pipeline writer's `Write*` callers are the suspect set. Verify each `Write*` site is inside a `UserQueriesForLoop` / `WithUserQueries` scope. |
| Pipeline page SSE / Time Machine picker | live updates continue, no 5xx burst | If the SSE goes dark, the pipeline writer is mis-routed *or* the read-side queries (`pipeline.go:201, 210, 276, 447, 512, 554`) are hitting a non-RLS-scoped handle. Check Phase 1 §5.3 Class D entries for pipeline.go. |
| Webhook / OTLP / syslog / GitHub ingestion | successful 202s; no idempotency-cache write failures | Each of these is a Phase 1 §5.3 bearer-token row; a regression here means the bearer-token handler's `UserQueries` scope didn't land or didn't include the write that's failing. |
| Interactive chat | conversations persist; tools return rows; no transaction-timeout errors | If chat starts dropping conversations or returning empty tool results, `UserQueriesForLoop`'s `idle_in_transaction_session_timeout = 0` either didn't apply or isn't enough — re-check the loop entry point in `chat.go`. |
| Scheduled investigations | `last_run_at` advances on schedule; `agent_log` rows appear | If schedules silently stop running, the scheduler's per-schedule `UserQueries` open is failing — likely a connectivity error to `app_user` rather than a privilege error. |

Within the 48h:

- **Any anomaly that's not immediately diagnosable → execute revert
  lever (Step 5), debug under the postgres pool, re-flip when
  fixed.** No partial fixes, no "let's see if it gets better."
- **Any anomaly that *is* diagnosable and has a one-line code or
  grant fix → ship the fix as a follow-up PR, redeploy, restart the
  bake clock from zero.** The 48h is for confidence in stability,
  not raw wall-clock; restarting is honest.

### Step 5 — Revert lever (named, practiced)

The fastest path back to the pre-flip topology is one
`flyctl secrets set` that reverts `DATABASE_URL` and
`CRON_DATABASE_URL` to the `postgres` URL, plus blanks `HEIMDALL_ENV`:

```bash
flyctl secrets set \
  --app heimdall \
  DATABASE_URL="postgres://postgres:$(op read 'op://Heimdall/postgres/password')@<prod-host>:5432/<db>?sslmode=require" \
  CRON_DATABASE_URL="" \
  HEIMDALL_ENV=""
```

Setting `CRON_DATABASE_URL=""` returns the system to the Phase 3
fall-through (the `slog.Warn` reappears; cron pool authenticates as
`postgres` via the `DATABASE_URL` fallback). Setting `HEIMDALL_ENV=""`
deactivates the `assertRoleSplit` invariant, so the boot won't trip on
the now-identical pools.

`flyctl secrets set` triggers an automatic redeploy; full
revert-to-prior-topology in <60 seconds (Fly's rolling deploy time).
**Do not** `DROP ROLE`, do not down-migrate 039 — the roles and
grants stay in place so the next flip attempt is bootstrap-free.

The revert was rehearsed locally during Step 4 of the rehearsal
procedure (above) — that rehearsal is the gating evidence that this
lever actually works. Practice it before the prod flip; the worst
moment to discover the revert path is broken is at minute 3 of an
incident.

### Step 6 — Bake closure

When 48h elapse with the watch list above all green, the bake closes
clean. Fill in the *Bake result* table at the bottom of this doc with
concrete numbers:

- Pre-flip baseline (per-minute rates for `agent_log` and
  `log_pipeline_events`).
- Post-flip rate at +1h, +24h, +48h.
- Any blip > ±5% with attribution.

Phase 7 starts on the back of a green close.

---

## Acceptance check status

These are the roadmap Phase 6 acceptance checks (gate to Phase 7).
Items below the line are pending the operator window; items above
the line are complete in this commit.

**Code / runbook / rehearsal-side (complete):**

- [x] **Local rehearsal procedure documented** with exhaustive coverage
      of every Phase 1 §5.3 path — see *Local rehearsal procedure*.
- [x] **Production flip runbook documented** with the Fly.io secrets
      shape and the load-bearing single-transaction sequencing — see
      *Production flip runbook*.
- [x] **Revert lever named and rehearsed** — see Step 4 of the local
      rehearsal procedure (which exercises the same shape).
- [x] **`HEIMDALL_ROLE_SPLIT_LIVE` flag wired into Phase 5
      connectivity tests** (already in
      `backend/internal/db/rls_connectivity_test.go` since Phase 5);
      runbook step 3 codifies the operator-invoked smoke gate.
- [x] **`.env.example`** contains a Phase 6 example fill stanza so
      the post-flip URL shape is visible to any operator preparing a
      `.env` file.

**Operator-window-side (pending the prod flip):**

- [ ] Zero `42501 permission denied` errors over the 48h prod bake.
- [ ] `agent_log` and `log_pipeline_events` throughput within ±5% of
      the pre-flip baseline.
- [ ] Webhook / OTLP / syslog / GitHub ingestion confirmed working
      under the new role via prod traffic.
- [ ] Phase 5 §5.6b connectivity tests passing against the live prod
      URLs (operator-invoked smoke gate).
- [ ] Migration B's down-migration rehearsed locally before Phase 7
      starts (carried as a Phase 6 bake-window task per the roadmap).

Phase 7 starts when both the bake closes clean and the operator-window
items above flip from `[ ]` to `[x]` in the *Bake result* table at the
bottom of this doc.

---

## Surprises and design notes

### `HEIMDALL_ENV` is the on/off switch for the safety net, not a feature gate

Tempting to read `HEIMDALL_ENV=production` as "we're in prod now,
turn on prod-flavoured behaviour everywhere." Resist. Phase 3
(`config.go` field comment) made it explicit: this var exists to
gate startup *invariants* — `assertRoleSplit` today, possibly
"refuse to launch when migrations are behind" later. It is not a
feature flag for log format, CORS rules, or anything else (those
have their own knobs already).

The Phase 6 flip turns it on in prod *because* the role split makes
the invariant meaningful. Anyone reading this doc later and tempted
to add unrelated checks behind the same flag should add a sibling
flag instead — overloading invites the failure mode where flipping
one capability accidentally activates an unrelated one.

### The connectivity test's lack of CI is a honest limit, not a regression

Phase 5's design framed `HEIMDALL_ROLE_SPLIT_LIVE` as "Phase 6 sets the
flag in CI as part of the env-var rollout." Heimdall has no CI in this
repo — no `.github/workflows/`, no `Jenkinsfile`, no equivalent. The
roadmap's CI framing was aspirational; the realistic shape is what the
runbook step 3 codifies — operator-invoked test, run once at flip time
plus during the 48h bake whenever a dashboard signal spikes.

If/when CI is added to the repo, wiring `HEIMDALL_ROLE_SPLIT_LIVE=1`
into the integration job is a one-line change, but it's not Phase 6's
job to ship a CI surface that doesn't exist yet.

### `assertRoleSplit` is the single load-bearing piece of safety

If I had to point at one line of code that prevents Phase 6 from
shipping in a broken state, it's `main.go:79–84`. The `if
cfg.Environment == "production"` gate, the `SELECT current_user`
calls, and the same-role refusal-to-launch combine to make
"misconfigured deploy lands traffic on the wrong role" a structural
impossibility. The Phase 5 `TestAppPool_IsActuallyAppUser` test does
the same shape externally; the test-and-runtime pair means a
silent-misconfig regression has to defeat both.

This is why the runbook insists `HEIMDALL_ENV=production` and the URL
flip ship in *one* secrets transaction. The split-deploy alternative
("flip URLs first, monitor, then enable the invariant") gives a
window where a misconfig produces symptoms before the safety net
activates. Single transaction, single boot, single decision point.

### The Phase 5 catalog test (Test 2) versus a Migration-039-applied DB

Phase 5's completion doc surfaced this: `TestCatalogDrift_PolicyVerbCoverage`
will fail against any DB with Migration 039 applied but Migration 040
not yet, because the allowlist already promises the Phase-7 target
state for `agent_config`, `webhook_idempotency`, `log_pipeline_events`.

During the Phase 6 bake the prod DB is in exactly this state. **Do not
run integration tests against the prod DB during the bake.** The Phase
5 recommendation was "gate the catalog tests away from the integration
DB in Phase 6"; in Heimdall's no-CI world this gating is enforced by
operator discipline — don't point `DATABASE_URL` at prod when running
`go test ./backend/internal/db/...`. The connectivity test's separate
flag (`HEIMDALL_ROLE_SPLIT_LIVE`) is what scopes the operator-invoked
smoke gate to just the role checks.

### The Migration B down-migration rehearsal piggybacks here

Roadmap Phase 6 acceptance check #5: "Migration B's down-migration
rehearsed locally before Phase 7 starts." Phase 7 is the
`FORCE RLS` + policy patches migration. Rehearsing its down-migration
during the Phase 6 bake window costs zero — there's nothing else
happening in that window — and means Phase 7's PR doesn't have to
include the rehearsal as a gating step. The rehearsal output goes
into the Phase 7 completion doc, not this one.

### Doc-side prerequisite from Phase 1, still pending

The roadmap header references `docs/executing/rls-enforcement-mental-model.md`
(does not exist) and `docs/refs/trajan-db-roles.md` (lives in
`docs/executing/` not `docs/refs/`). Carried forward across Phases 1,
2, 3, 4, and 5 completion docs. Still non-blocking for Phase 7; the
roadmap header links remain broken. Owner: doc-side PR before Phase 8,
which has explicit doc-cleanup scope.

---

## Files changed

```
.env.example                                              (Phase 6 example fill stanza)
docs/completions/rls-enforcement-phase-6.md               (new)
```

No production-code changes. No migration changes. No test changes.

---

## Bake result (filled in at bake closure)

| Window | `agent_log` rate | `log_pipeline_events` rate | `42501` errors | Notes |
|---|---|---|---|---|
| Pre-flip baseline (T-24h) | TBD | TBD | n/a | Filled in at flip-time |
| T+1h | TBD | TBD | TBD | |
| T+24h | TBD | TBD | TBD | |
| T+48h (close) | TBD | TBD | TBD | |

Operator notes / incidents during the bake:

> _Filled in at bake closure. Empty section = clean bake._

---

## What Phase 7 needs from this

Phase 7 (Migration B — `FORCE RLS` + policy patches) starts on the
back of a clean Phase 6 bake. The pieces Phase 7 inherits from this
phase:

1. **Production runtime is already on `app_user` + `cron_user`.** RLS
   is cosmetic (no `FORCE` yet); Phase 7's migration is the flip from
   cosmetic to enforced. Any `new row violates row-level security
   policy` error during Phase 7's bake is by construction a *policy*
   miss (a table with `FORCE` but no INSERT/UPDATE/DELETE policy) or a
   *write-path* miss that Phase 6 should have caught (a writer not
   going through `UserQueries`). Phase 6's clean bake means the latter
   class is empty.
2. **The Migration B down-migration is rehearsed.** Roadmap acceptance
   check #5; happens during the Phase 6 bake window per the design
   note above.
3. **The Phase 5 §5.6a Test 1 (FORCE coverage) is the gating signal
   for Phase 7's correctness.** Today it's gated off by
   `HEIMDALL_RLS_FORCE_LIVE`. Phase 7's job is to ship the migration
   *and* turn on the flag in the operator-invoked verification step,
   confirming every `rowsecurity = true` table is now `forcerowsecurity
   = true`.
4. **The Phase 5 §5.6a Test 2 (verb coverage) is the gating signal
   for the policy patches.** The allowlist already promises the
   Phase-7 target state — Phase 7 makes the catalog match the promise.

The Phase 7 PR is migration files (040 up/down) plus the Phase 7
completion doc. No code changes; the test surface, the role topology,
and the runtime traffic are all already on the post-flip shape from
Phase 6.
