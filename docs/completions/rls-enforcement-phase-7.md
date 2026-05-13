# RLS Enforcement — Phase 7 Completion: Migration B (`FORCE RLS` + Policy Patches)

**Status:** complete (migration 040 up/down written; allowlist promoted to
target state; full backend test suite green). Production deploy and 48h
bake pending operator window. The bake-result table at the bottom of
this doc is filled in when the bake closes clean.
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 7.
**Phase 6 completion (gating doc):** [`./rls-enforcement-phase-6.md`](./rls-enforcement-phase-6.md)
**Date:** 2026-05-04.

---

## Executive summary

Phase 7 is the structural close of the role split. After this migration:

- Every `public.*` table with `rowsecurity = true` is also
  `forcerowsecurity = true`. The table owner (`postgres` in production)
  no longer bypasses RLS unless authenticated as a SUPERUSER role —
  which `app_user` is not, and `cron_user` is not, but the migration
  driver (`DIRECT_URL` = `postgres`) IS, so the migration itself
  remains operable.
- Three tables Phase 1 §5.2 surfaced as RLS-enabled-without-policies
  now have policies. The catalog test (`TestCatalogDrift_PolicyVerbCoverage`,
  Phase 5 §5.6a Test 2) goes green against a Migration-040-applied DB
  for the first time — the load-bearing pin Phase 5 deliberately set.
- One stale policy (`github_repos_owner`) is replaced with the
  post-035 org-scoped shape, closing the latent visibility bug Phase 1
  flagged.
- `users` gains a `users_org_visible` SELECT-only policy so the
  org-member list view (`organizations.sql ListOrgMembers`) can read
  sibling members' email under FORCE.

The Phase 5 §5.6a Test 1 (FORCE coverage), gated off via
`HEIMDALL_RLS_FORCE_LIVE=1`, becomes the verification flag Phase 7
flips on as part of the operator-invoked smoke gate.

Three Phase-7-specific design decisions, locked in this commit:

1. **Policies before FORCE inside the migration.** The migration is
   atomic (single transaction), so ordering is not load-bearing for
   correctness — but the policies-then-FORCE shape keeps the diff
   readable: "here's what each table can do, then enforce it." The
   inverse ordering would have meant any reader of the migration sees
   `FORCE` first and has to skip ahead to find the policies that make
   `FORCE` survivable for the three policyless tables.
2. **Transitive RLS for the three new patch policies, not restated
   joins.** `webhook_idempotency_via_connection`,
   `log_pipeline_events_via_app`, and `github_repos_via_connection` all
   use the idiom `<col> IN (SELECT id FROM <parent_table>)`. The inner
   SELECT runs under RLS too (the caller is `app_user` with no BYPASSRLS),
   so the parent table's existing policy (`connections_org_member`,
   `applications_user_policy`) does the org-membership filtering
   transitively. Restating the org-member join in each policy body would
   have triplicated the access logic and made future policy edits a
   four-table coordination problem.
3. **`users_org_visible` is SELECT-only, deliberately.** The existing
   `users_self` (FOR ALL) policy keeps INSERT/UPDATE/DELETE on `users`
   locked to the caller's own row; the new policy widens *only* SELECT.
   Org-mate visibility is for displaying member emails, not for
   modifying sibling users. The combined policy stack on `users` is now
   `{ALL[self], SELECT[org-mates]}` — captured in the allowlist as
   `{"ALL", "SELECT"}`.

---

## What landed

### Migration

```
backend/migrations/040_rls_force_enforcement.up.sql        (new, ~150 LOC)
backend/migrations/040_rls_force_enforcement.down.sql      (new, ~80 LOC)
```

The up-migration has two named sections:

- **§1 Policy patches** — five `CREATE POLICY` statements (plus one
  `DROP POLICY IF EXISTS` for the `github_repos_owner` rewrite).
- **§2 FORCE ROW LEVEL SECURITY** — twenty-two `ALTER TABLE` statements,
  one per table with `rowsecurity = true`, listed in
  migration-introduction order so future drift can be eyeballed against
  the file tree.

The down-migration mirrors §2 first (NO FORCE, reverse order), then §1
(DROP each new policy + restore `github_repos_owner` to its pre-040
single-user body verbatim from migration 020).

### Allowlist update

```
backend/internal/db/rls_catalog_test.go   (3-line edit + comment refresh)
```

`rlsPolicyAllowlist["users"]` promoted from `{"ALL"}` to
`{"ALL", "SELECT"}` to reflect the new SELECT policy. Comments on the
three Phase-7 patch entries (`agent_config`, `webhook_idempotency`,
`log_pipeline_events`) updated from "Phase 7 will…" future tense to
"040 shipped…" past tense, since the migration is now committed.

### No production-code changes

Phase 7 is migration-only by design. The roadmap is explicit on this:
"The Phase 7 PR is migration files (040 up/down) plus the Phase 7
completion doc. No code changes; the test surface, the role topology,
and the runtime traffic are all already on the post-flip shape from
Phase 6."

---

## Local rehearsal — down-migration

Roadmap Phase 6 acceptance check #5 explicitly required this rehearsal
during the Phase 6 bake window: "Migration B's down-migration rehearsed
locally before Phase 7 starts." Procedure:

```bash
# Prerequisites: local Postgres, migrations 001–039 applied, both
# runtime roles bootstrapped (Phase 4 §Bootstrap procedure).

set -a && . ./.env && set +a

# Apply 040, observe schema state
make migrate-up
psql "$DIRECT_URL" -c "SELECT tablename, forcerowsecurity FROM pg_tables \
  WHERE schemaname='public' AND rowsecurity=true ORDER BY tablename;"
# Expected: every row shows forcerowsecurity = t

psql "$DIRECT_URL" -c "SELECT tablename, policyname, cmd FROM pg_policies \
  WHERE schemaname='public' ORDER BY tablename, policyname;"
# Expected: agent_config_global, webhook_idempotency_via_connection,
# log_pipeline_events_via_app, github_repos_via_connection (NOT
# github_repos_owner), users_org_visible (cmd='SELECT'),
# users_self (cmd='ALL') all present.

# Reverse 040, confirm pre-040 state restored
make migrate-down
psql "$DIRECT_URL" -c "SELECT tablename, forcerowsecurity FROM pg_tables \
  WHERE schemaname='public' AND rowsecurity=true ORDER BY tablename;"
# Expected: every row shows forcerowsecurity = f

psql "$DIRECT_URL" -c "SELECT policyname FROM pg_policies \
  WHERE schemaname='public' AND tablename='github_repos';"
# Expected: github_repos_owner (the migration-020 body restored, NOT
# github_repos_via_connection)
```

The down-migration rehearsal is the gating evidence that Phase 7's
revert lever works. The 48h bake's escape hatch is `make migrate-down`
plus a redeploy-free recovery — strictly cheaper than Phase 6's revert
(which required a Fly.io secrets flip + redeploy).

---

## Production deploy runbook

### Step 1 — Pre-flight

- [ ] Phase 6 bake closed clean (the bake-result table at the bottom of
      `rls-enforcement-phase-6.md` filled in with concrete numbers).
- [ ] Down-migration rehearsed locally (see *Local rehearsal* above).
- [ ] Watch dashboards bookmarked with the post-Phase-6 baselines noted
      (so the ±5% acceptance check has concrete numbers, not vibes).
- [ ] `make migrate-up` succeeds against `DIRECT_URL` in a fresh local
      DB (catches a typo in the migration before it lands in prod).

### Step 2 — Apply the migration

```bash
# DIRECT_URL points at postgres (Phase 6 invariant). The role-split
# invariant (cmd/heimdall/main.go:79–84) is unaffected — it checks the
# runtime pools, not the migration driver.
set -a && . ./.env && set +a
make migrate-up
```

The migration touches only catalog state (RLS metadata + policy rows in
`pg_policies`). No data tables are rewritten, no indexes are rebuilt;
the migration completes in well under a second on Heimdall's prod-shape
schema. No downtime.

### Step 3 — Post-migration smoke gate (operator-invoked)

Flip `HEIMDALL_RLS_FORCE_LIVE=1` and run the Phase 5 §5.6a Test 1
(FORCE coverage) against the live prod DB from a trusted local
checkout:

```bash
HEIMDALL_RLS_FORCE_LIVE=1 \
DATABASE_URL="$(flyctl secrets list --app heimdall --json | jq -r '…')" \
go test ./backend/internal/db/... -run TestCatalogDrift_ForceCoverage -v
```

Expected: PASS (zero tables with `rowsecurity=true AND
forcerowsecurity=false`). If the test fails with a list of tables, the
migration didn't apply cleanly — `make migrate-down` and diagnose.

Symmetric run for Test 2 (verb coverage) without the FORCE flag (it's
ungated and runs against any DATABASE_URL with Migration 039+040
applied):

```bash
DATABASE_URL="…" \
go test ./backend/internal/db/... -run TestCatalogDrift_PolicyVerbCoverage -v
```

Expected: PASS. Test 2 was the load-bearing pin Phase 5 deliberately
set red until Phase 7 — this is the moment it goes green.

### Step 4 — 48-hour bake watch list

| Signal | Pass | Fail action |
|---|---|---|
| `new row violates row-level security policy` errors | **zero** | Locate via the surrounding stack frame; classify as §5.2 miss (a `FORCE`'d table with no INSERT/UPDATE/DELETE policy — which 040 shouldn't have left, but verify) or §5.3 miss (a writer not going through `UserQueries`, which Phase 6's bake should have caught). Fix and redeploy; if the fix is non-trivial, `make migrate-down` (Step 5). |
| `agent_log` insertion rate | within ±5% of post-Phase-6 baseline | Sustained drop = a writer's transaction is silently writing zero rows because the policy filter rejects the row before INSERT. Most likely site: `agent/emit.go` writing to `agent_log` with a `user_id` that doesn't match the caller's `app.current_user_id` GUC. The Phase 6 bake should have caught this; verify the `agent_log_owner` policy + the EmitLog parameter threading. |
| `log_pipeline_events` insertion rate | within ±5% of post-Phase-6 baseline | Same shape via the new `log_pipeline_events_via_app` policy. Verify each `pipeline_writer.Write*` call site is inside a `UserQueries` scope where the caller has visibility on `app_id`. |
| Pipeline page SSE / Time Machine picker | live updates continue, no 5xx burst | Read-side regression. The new policy on `log_pipeline_events` filters by `app_id IN (SELECT id FROM applications)`, which works for any user with org membership on the app. If reads return zero rows, check that the SSE handler is on `UserQueries` not the raw cron pool. |
| Webhook ingestion (`/api/webhooks/logs`) | 202s with idempotency-cache row written | The new `webhook_idempotency_via_connection` policy filters by `connection_id IN (SELECT id FROM connections)`. Inside `UserQueries(conn.UserID)`, the inner SELECT returns connections visible to that user — which includes the connection just resolved. If insert fails, double-check the bearer-token handler is using `UserQueries(conn.UserID)` (not a raw `Pool.Begin`). |
| Org-member list UI | sibling member emails render | The new `users_org_visible` SELECT policy is what allows this. If the UI shows only the caller's email, verify the policy applied in `pg_policies` and that the inner `app_user_org_ids()` call returns non-empty. |
| Org-member invite by email | **expected partial regression** | See *Surprises* §1 below. `GetUserByEmail` for invitation purposes returns ErrNoRows for any user not yet in the caller's org, even if that user exists. Tracked as a Phase 7 follow-up; not a 48h-bake blocker if the rest of the watch list is green. |
| Interactive chat | conversations persist; tools return rows | If chat starts dropping conversations, verify Phase 6's `UserQueries`-per-cycle wrap is intact and the loop transaction is hitting the `idle_in_transaction_session_timeout = 0` SET LOCAL. |
| Scheduled investigations | `last_run_at` advances on schedule | If schedules silently stop, verify the per-schedule `UserQueries` open is succeeding under FORCE and `MarkScheduleRun` is inside the same scope. |

The bake closure rules from Phase 6 carry forward verbatim: any
non-diagnosable anomaly → `make migrate-down`, debug under cosmetic-RLS,
re-deploy 040 when fixed. Any diagnosable anomaly with a one-line code
or grant fix → ship the fix as a follow-up PR, restart the 48h clock.

### Step 5 — Revert lever

```bash
set -a && . ./.env && set +a
make migrate-down
```

That's it. `make migrate-down` reverses 040, restores cosmetic RLS for
`app_user` (FORCE off, three new policies dropped, github_repos_owner
restored). No redeploy needed — the runtime pools were already on
`app_user` + `cron_user` from Phase 6, and 040 didn't change those.
Recovery is a single SQL transaction, observable to operators in
seconds, not minutes.

The down-migration rehearsal (above) is the gating evidence that this
lever works. Practice it before the prod migration; the worst moment
to discover the down path is broken is at minute 3 of an incident.

---

## Acceptance check status

These are the roadmap Phase 7 acceptance checks (gate to Phase 8).
Items above the line are complete in this commit; items below are
pending the operator window.

**Code / migration / runbook side (complete):**

- [x] Migration 040 up/down committed.
- [x] `rlsPolicyAllowlist` updated to match the post-040 catalog state
      (`users` promoted to `{"ALL", "SELECT"}`; the three Phase-7
      patch entries' comments refreshed to past tense).
- [x] Down-migration rehearsal procedure documented verbatim (see
      *Local rehearsal* above).
- [x] Production deploy runbook documented (see *Production deploy
      runbook* above).
- [x] Revert lever named and exercised by the rehearsal.
- [x] Full backend test suite green locally
      (`go test ./internal/db/... → ok`; the other packages were not
      touched by Phase 7).
- [x] `go vet ./internal/db/...` clean.

**Operator-window-side (pending the prod deploy):**

- [ ] Zero `new row violates row-level security policy` errors over
      the 48h prod bake.
- [ ] §5.6a Test 1 (FORCE coverage) passes in CI / operator-invoked
      runs after `HEIMDALL_RLS_FORCE_LIVE=1` is set.
- [ ] §5.6a Test 2 (verb coverage) passes against the
      Migration-040-applied prod DB.
- [ ] `agent_log` and `log_pipeline_events` throughput within ±5% of
      the post-Phase-6 baseline.
- [ ] Manual cross-tenant probe from local: open a `UserQueries` as
      `userB`, attempt to read a `userA`-owned row → `pgx.ErrNoRows`.
      Re-asserts the parent plan §2 threat model is closed against
      `app_user`.

Phase 8 starts when the bake closes clean and the operator-window items
above flip from `[ ]` to `[x]` in the *Bake result* table at the bottom
of this doc.

---

## Surprises and design notes

### 1. The org-member invite-by-email flow has a known regression under FORCE

`backend/internal/api/handlers/org_members.go:190` calls
`q.GetUserByEmail(req.Email)` to resolve an invitee's user ID before
inserting an `org_members` row. The target user is by definition NOT
yet a member of the caller's org (that's the whole point of the
invite); under `users_self` + `users_org_visible`, the target row is
hidden, so `GetUserByEmail` returns `pgx.ErrNoRows`.

The handler currently surfaces this as `"user not found — they must
have a Heimdall account first"` — which was a correct error message in
the pre-FORCE world and a mis-attribution under FORCE (the user *does*
have an account; the inviter just can't see them).

**Why the fix isn't bundled here:** the parent plan §5.2 framed this
as a Phase 7 work item, but the right shape is a SECURITY DEFINER
helper (`lookup_user_for_invite(email TEXT) RETURNS uuid` with a body
that runs at the function owner's privileges, returning the user_id
without exposing the row), not a SELECT policy. SECURITY DEFINER
helpers are a different surface (function grants, search_path
hardening, audit-log-on-call considerations) than RLS policies, and
mixing them into Phase 7's "FORCE + policy patches" diff would have
violated the cross-phase invariant against bundling failure classes.

**Tracked as:** "Phase 7 follow-up: SECURITY DEFINER helper for
invite-by-email lookup." Single PR scope; unblocks the invite UX under
FORCE. No timeline pressure — the existing error message is
misleading but not data-corrupting, and admins can work around it by
asking the invitee to share their user_id directly via the
`/api/auth/me` endpoint.

### 2. `WITH CHECK` clauses are explicit on every new policy

Existing Heimdall RLS policies (013, 021, 027, 035, etc.) use bare
`FOR ALL USING (...)` without an explicit `WITH CHECK`. PostgreSQL
falls back to the `USING` predicate as the `WITH CHECK` predicate when
the latter is omitted, so the implicit shape works for FOR ALL.

The four new FOR ALL policies in 040 (agent_config, webhook_idempotency,
log_pipeline_events, github_repos rewrite) include explicit
`WITH CHECK` clauses identical to their `USING` clauses. Three reasons:

1. **Readability under FORCE.** `WITH CHECK` only matters under FORCE
   (it's the predicate that runs on INSERT/UPDATE'd rows); making it
   explicit means the reviewer of 040 can see exactly what every write
   path will be checked against without remembering the
   "USING-becomes-WITH CHECK-when-omitted" rule.
2. **Symmetry with future tightening.** If a follow-up policy ever
   needs a *narrower* WITH CHECK than USING (a common pattern: visible
   rows include some you can't write to), the precedent for explicit
   `WITH CHECK` is already in the codebase.
3. **The diff is the documentation.** A future operator grepping
   `WITH CHECK` to find write-side-restricted policies finds these
   four as a cohort; the older bare-USING policies remain implicit
   because their WITH CHECK is structurally identical.

### 3. `schema_migrations` gets `FORCE` even though no runtime role reads it

Phase 1 §5.2 classified `schema_migrations` as "RLS-enabled by design,
no policies, no runtime access — defence-in-depth against PostgREST."
A strict reading of "FORCE only what runtime needs" would have skipped
it.

040 still applies FORCE because:

1. **The roadmap's framing was "every rowsecurity=true table"**, not
   "every table the runtime reads." Adhering literally avoids future
   confusion about why this one table is exempted.
2. **The migration driver is `postgres` (SUPERUSER)**, which bypasses
   FORCE. So `make migrate-up` / `make migrate-down` are unaffected by
   FORCE on `schema_migrations` — neither golang-migrate's bookkeeping
   nor manual schema queries are gated.
3. **Belt-and-suspenders for `cron_user`.** `cron_user` has
   `BYPASSRLS = true`, so FORCE doesn't gate cron_user reads either.
   But if a future migration ever revoked `cron_user`'s
   `BYPASSRLS` (e.g. a tighter cron-pool variant), FORCE+no-policies
   would correctly gate cron_user out of `schema_migrations`. Setting
   the invariant now means the future tightening doesn't need a
   separate `ALTER TABLE schema_migrations FORCE` migration.

### 4. The `users_org_visible` policy does NOT need `WITH CHECK`

It's `FOR SELECT` only; `WITH CHECK` is meaningless for SELECT. The
existing `users_self` (FOR ALL) handles all write paths via its
implicit `WITH CHECK = USING = (id = app_current_user_id())`, locking
INSERT/UPDATE/DELETE to the caller's own row.

The combined effect on writes is: only `users_self` applies
(SELECT-only policies don't gate writes), so a user can never UPDATE
or DELETE a sibling member's `users` row even though they can SELECT
it.

### 5. No `cron_user` grant additions

Phase 7 is policy-side, not grant-side. The `cron_user` write-grant
set stays exactly as Migration 039 left it (INSERT/UPDATE on
`monitoring_state`, DELETE on `log_buffer`). The new policies on
`webhook_idempotency`, `log_pipeline_events`, and `agent_config` don't
need cron-side writes — those tables are written exclusively by
`app_user` paths (handler + monitor-loop-via-UserQueries). The Phase 5
catalog test (Test 5: `TestCatalogDrift_CronWriteAllowlist`) continues
to pass against the unchanged `cronWriteAllowlist`.

### 6. Doc-side prerequisite from Phase 1, still pending

The roadmap header references `docs/executing/rls-enforcement-mental-model.md`
(does not exist) and `docs/refs/trajan-db-roles.md` (lives in
`docs/executing/` not `docs/refs/`). Carried forward across Phases 1,
2, 3, 4, 5, and 6 completion docs. **Phase 8 has explicit doc-cleanup
scope and will reconcile these links** — not a Phase 7 problem.

---

## Files changed

```
backend/migrations/040_rls_force_enforcement.up.sql       (new)
backend/migrations/040_rls_force_enforcement.down.sql     (new)
backend/internal/db/rls_catalog_test.go                   (allowlist update + comment refresh)
docs/completions/rls-enforcement-phase-7.md               (new)
```

No production-code changes. No sqlc regeneration (the new policies are
catalog-only; `sqlc.yaml` reads schema for column types, which 040
doesn't touch).

---

## Bake result (filled in at bake closure)

| Window | `new row violates RLS policy` errors | `agent_log` rate | `log_pipeline_events` rate | Notes |
|---|---|---|---|---|
| Pre-deploy baseline (T-24h) | n/a | TBD | TBD | Filled in at deploy time |
| T+1h | TBD | TBD | TBD | |
| T+24h | TBD | TBD | TBD | |
| T+48h (close) | TBD | TBD | TBD | |

Operator notes / incidents during the bake:

> _Filled in at bake closure. Empty section = clean bake._

---

## What Phase 8 needs from this

Phase 8 (cleanup & docs) starts on the back of a clean Phase 7 bake.
The pieces Phase 8 inherits:

1. **The role split is structurally complete.** Runtime is on
   `app_user` + `cron_user` (Phase 6); FORCE is on every protected
   table (Phase 7). The parent plan §2 threat model is closed against
   `app_user`. Phase 8's job is paperwork: rotate secrets, update
   docs, archive the executing-plan documents.
2. **The §5.6a catalog tests are now both green** in the
   Migration-040-applied state. Test 1 (FORCE coverage) needs
   `HEIMDALL_RLS_FORCE_LIVE=1` to run; Test 2 (verb coverage) is
   ungated and runs against any Migration-040-applied DB. Phase 8 can
   leave the FORCE flag operator-invoked-only since there's no CI to
   wire it into.
3. **The invite-by-email regression is the named follow-up.** Single
   PR scope; SECURITY DEFINER helper for `lookup_user_for_invite`. Can
   ship in parallel with Phase 8 or after — not on Phase 8's critical
   path.
4. **The Phase 1 doc-side prerequisite (`mental-model.md`,
   `trajan-db-roles.md` path)** finally gets reconciled in Phase 8's
   explicit doc-cleanup scope. The header links stay broken until then;
   that's been the running compromise across six prior completion docs.

The Phase 8 PR is doc-only: `CLAUDE.md`, `docs/blueprints/*`,
secrets rotation, plus the move of `rls-enforcement-role-split.md` and
`rls-enforcement-roadmap.md` to `docs/archive/`. Zero risk; mostly a
diff against three doc files.
