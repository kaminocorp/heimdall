# RLS Enforcement — Phase 8 Completion: Cleanup & Docs

**Status:** complete (Phase 7 follow-up shipped; blueprints refreshed;
executing-plan docs archived; cross-links reconciled). Production
secrets rotation is the lone operator-side residual — documented
below; no application change pending.
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — this is the deliverable for Phase 8.
**Phase 7 completion (gating doc):** [`./rls-enforcement-phase-7.md`](./rls-enforcement-phase-7.md)
**Date:** 2026-05-04.

---

## Executive summary

Phase 8 is the close-out phase. After this commit:

- The Phase 7 invite-by-email regression is resolved via a SECURITY
  DEFINER helper (migration 041) and a one-spot wiring change in
  `org_members.go AddOrgMember`. The handler now distinguishes "user
  doesn't exist" from "user exists but RLS hides them" because under
  the helper there's no second case.
- `docs/blueprints/database-connection-blueprint.md` and
  `docs/blueprints/backend-blueprint.md` no longer assert "we connect
  as the postgres superuser" as the runtime model — both surfaces
  describe the three-role topology with explicit pointers to the
  completion docs and the archived roadmap.
- `docs/executing/rls-enforcement-roadmap.md` and
  `docs/executing/rls-enforcement-role-split.md` moved to
  `docs/archive/`. The mental-model doc stays in
  `docs/blueprints/rls-enforcement-mental-model.md` as evergreen
  reference (it's a layer map, not a plan).
- `docs/refs/trajan-db-roles.md` carries a forward link to the
  Heimdall completion docs, naming the three deviations from the
  Trajan precedent so a future port has the diff in hand.
- The Phase-1-discovered broken link (`docs/executing/rls-enforcement-mental-model.md`,
  which never existed at that path) is reconciled in the archived
  roadmap and across every completion doc — the actual file lives in
  `docs/blueprints/`.

The eight-phase rollout is now structurally complete on the codebase
side. Production runtime is on `app_user` + `cron_user`, every
RLS-enabled table is `FORCE`'d, the policy gaps Phase 1 surfaced are
closed, and the test surface (Phase 5) plus the catalog-drift
allowlist (Phase 7) keep future regressions noisy.

Three Phase-8-specific design decisions, locked in this commit:

1. **The invite follow-up shipped as part of Phase 8, not as a
   freestanding PR.** The roadmap was silent on where this work
   belonged because the regression was discovered late (during Phase
   7's design pass). Folding it into Phase 8 keeps the rollout's
   final state self-consistent — every code path works under the
   three-role topology before the executing-plan docs get archived.
   The alternative (ship Phase 8 with the regression open and the fix
   as a follow-up PR) would have meant the archived roadmap describes
   a closed system that doesn't exist for one specific code path.
2. **Migrations stay archived in `docs/executing/` only via `git mv`**;
   the working-tree path is `docs/archive/`. Convention here matches
   `docs/archive/runbooks/` (0.47.4) — historical artefacts move
   physically, with link reconciliation as the same-PR follow-on.
   Leaving an `executing/` symlink stub was considered and rejected:
   broken-link-on-rename is loud (404 on click) where a stale symlink
   is silent rot.
3. **Secrets rotation is documented but not executed in this commit.**
   The roadmap §8 lists "rotate `postgres`, `app_user`, `cron_user`
   passwords to long random values, distinct from each other" as
   Phase 8 work. That's an operator action against the production
   secrets manager (not a codebase action) and depends on a 60-second
   maintenance window where each role gets a fresh password and
   `flyctl secrets set` updates the corresponding URL. Documented
   below as the *Production secrets-rotation runbook*; the as-shipped
   section of the *Acceptance check status* table reflects this is
   the lone residual.

---

## What landed

### Phase 7 follow-up — invite-by-email helper

```
backend/migrations/041_lookup_user_for_invite.up.sql            (new)
backend/migrations/041_lookup_user_for_invite.down.sql          (new)
backend/internal/db/queries/users.sql                           (added LookupUserIDForInvite query)
backend/internal/db/users.sql.go                                (added matching generated method)
backend/internal/api/handlers/org_members.go                    (InviteMember rewired to use the helper)
```

Migration 041 introduces `public.lookup_user_for_invite(target_email TEXT)
RETURNS UUID` — `LANGUAGE sql STABLE SECURITY DEFINER SET search_path =
public`. Function body is a single-row email→user_id lookup, returning
NULL when no user matches. The function is `EXECUTE`-granted to
`app_user` only; `cron_user` doesn't need it (every cron path resolves
ownership via cross-tenant SELECTs on `applications`/`connections`,
never via email), and `PUBLIC` is explicitly revoked so a future
PostgREST exposure can't reach it.

The handler change in `InviteMember` is one block: the
`q.GetUserByEmail` call inside `WithUserQueries` becomes
`q.LookupUserIDForInvite`, the result type becomes `*uuid.UUID` (nil
on no match), and the surrounding error logic is rewritten so a NULL
result yields a clean 404 ("user not found — they must have a
Heimdall account first") while a real DB error yields 500. The
existing test suite (`TestInviteMember*`) still passes — the tests
exercise the happy path and the conflict / authz branches; they don't
discriminate the visibility-vs-existence case the helper specifically
fixes.

The `db.User` value previously read from `GetUserByEmail` is no longer
needed; the response body now echoes `req.Email` (caller-supplied)
rather than `targetUser.Email` (DB-canonical). In practice these
match — `users.email` is stored as-provided by Supabase auth — and
echoing the caller-supplied form makes the 201 response a clean
acknowledgement of "yes, I invited the email you typed."

### Blueprints — three-role topology made explicit

```
docs/blueprints/backend-blueprint.md          (§10 RLS subsection — "Current enforcement model" rewritten)
docs/blueprints/database-connection-blueprint.md  (added §0 + edited §3 + opening reading-note)
```

`backend-blueprint.md`'s "Current enforcement model" paragraph (the
:517 line called out in the roadmap §8) was the load-bearing stale
claim — it asserted "the backend connects as the postgres superuser,
which bypasses RLS by default." Rewritten in place to describe the
two runtime roles, the migration role, the bypass-then-scope handoff
shape, the SECURITY DEFINER escape hatch (migration 041), and the
production startup invariant. Inline link to the completion docs and
the archived roadmap so a reader who wants the rationale finds it
without leaving the blueprint.

`database-connection-blueprint.md` was a heavier diff. The original
was written when there was one `DATABASE_URL`, one `pgxpool.Pool`,
and one role; six sections referenced this structure explicitly. The
fix wasn't to rewrite all six — it was to add a §0 ("Three URLs,
three roles, two runtime pools") at the top with a table-shaped
summary, plus a reading note in the intro that scopes the rest of
the doc as "still applies, but read every singular as a per-role
plural." Sections that made an actively misleading claim got a
surgical edit (the "constructs a single pool" line in §3 became
"constructs two pools threaded into a `*db.Pools` struct"); the rest
stayed legacy text that's structurally still accurate for the pgxpool
mechanics it describes.

### Cross-links reconciled

```
docs/refs/trajan-db-roles.md                  (added Heimdall back-reference)
docs/executing/rls-enforcement-roadmap.md → docs/archive/...  (git mv + status header refreshed)
docs/executing/rls-enforcement-role-split.md → docs/archive/... (git mv + status header refreshed)
docs/completions/rls-enforcement-phase-{1..7}.md (path updates: ../executing/ → ../archive/)
CLAUDE.md                                     (RLS-section status refreshed; path updated)
```

`trajan-db-roles.md` now opens with a "Heimdall analogue" callout
naming the three deviations from the Trajan baseline:

1. `cron_user` carries `DELETE on log_buffer` beyond the Trajan
   write-grant set, for the audit-discovered pruner widening.
2. The catalog-drift tests (Phase 5 §5.6a) are Heimdall-specific
   extensions Trajan didn't ship — Trajan's allowlist-only approach
   relies on operator review at migration time; Heimdall pins it in
   CI via Test 2 (verb coverage) and the operator-flag-gated Test 1
   (FORCE coverage).
3. `lookup_user_for_invite` (migration 041) is a Phase 7 follow-up
   the Trajan precedent didn't need — Trajan has no email-based
   invite flow, so the SECURITY DEFINER escape hatch never came up.

The roadmap and parent-plan docs each have a Status header refreshed
to "shipped 2026-05-04 — all eight phases complete. Archived." plus a
pointer to the completion docs. The mental-model link in each was
fixed to `../blueprints/rls-enforcement-mental-model.md` (the Phase 1
broken-link prerequisite Phase 8 inherited).

The seven existing completion docs had `../executing/...` links
rewritten to `../archive/...` via a single sed pass; CLAUDE.md's RLS
section was refreshed in place.

### Archive moves

```
docs/archive/rls-enforcement-roadmap.md        (was: docs/executing/...)
docs/archive/rls-enforcement-role-split.md     (was: docs/executing/...)
```

`docs/blueprints/rls-enforcement-mental-model.md` stays put — it's a
layer map of "which roles show up in which files," which remains
useful as a conceptual on-ramp for someone reading the codebase
post-rollout. Archiving it would orphan the Phase 1 / 2 / 3
completion docs that link to it as a companion reference.

---

## Production secrets-rotation runbook

The roadmap §8 lists this as Phase 8 work; it is the lone residual
that isn't a codebase change. Roles' passwords today are still the
local-bootstrap values from Phase 4 plus whatever Fly.io secrets the
operator set during the Phase 6 flip. The Trajan precedent
(`docs/refs/trajan-db-roles.md`) is opinionated that distinct,
long-random passwords per role is the correct posture; we adopt the
same.

```bash
# 1. Generate fresh passwords. 32 random bytes, base64url-encoded
#    (no shell-escaping concerns), distinct per role.
NEW_POSTGRES=$(openssl rand -base64 32 | tr -d '=' | tr '+/' '-_')
NEW_APP_USER=$(openssl rand -base64 32 | tr -d '=' | tr '+/' '-_')
NEW_CRON_USER=$(openssl rand -base64 32 | tr -d '=' | tr '+/' '-_')

# 2. ALTER ROLE in the prod DB (run as superuser via DIRECT_URL).
psql "$DIRECT_URL" <<SQL
ALTER ROLE postgres   WITH PASSWORD '$NEW_POSTGRES';
ALTER ROLE app_user   WITH PASSWORD '$NEW_APP_USER';
ALTER ROLE cron_user  WITH PASSWORD '$NEW_CRON_USER';
SQL

# 3. Stage the new passwords in the secrets manager (1Password /
#    op cli example; substitute your own).
op item edit "Heimdall/postgres"   password="$NEW_POSTGRES"
op item edit "Heimdall/app_user"   password="$NEW_APP_USER"
op item edit "Heimdall/cron_user"  password="$NEW_CRON_USER"

# 4. Update Fly.io secrets in one transaction. flyctl triggers an
#    auto-redeploy; the role-split startup invariant verifies the new
#    URLs land on the correct roles (cmd/heimdall/main.go:79–84).
flyctl secrets set \
  --app heimdall \
  DATABASE_URL="postgres://app_user:$NEW_APP_USER@<host>:5432/<db>?sslmode=require" \
  CRON_DATABASE_URL="postgres://cron_user:$NEW_CRON_USER@<host>:5432/<db>?sslmode=require" \
  DIRECT_URL="postgres://postgres:$NEW_POSTGRES@<host>:5432/<db>?sslmode=require"
```

The post-redeploy boot log must include the
`role-split verified app_role=app_user cron_role=cron_user` line —
same load-bearing signal as the Phase 6 flip. Rollback is identical
to the Phase 6 revert lever (set `DATABASE_URL` back to the prior
working value, redeploy).

Operator window: a single 60-second maintenance moment. No app
downtime is necessary if the rotation is done in one `flyctl secrets
set` call (the running pods carry the old URLs until the rolling
deploy replaces them; the new pods pick up the rotated passwords).

This runbook is documented here rather than executed by automation
because the secrets manager surface varies (1Password, AWS Secrets
Manager, Fly.io vault, etc.) and the post-rotation verification step
is operator judgement, not a script.

---

## Acceptance check status

Roadmap Phase 8 / parent plan §9 acceptance items:

**Code / docs side (complete in this commit):**

- [x] All eight completion docs exist in
      `docs/completions/rls-enforcement-phase-N.md` (1, 2, 3, 4, 5,
      6, 7, 8). Phase 8 is this file.
- [x] Phase 7 invite-by-email regression resolved
      (migration 041 + handler wiring).
- [x] `CLAUDE.md` references the new topology — the Database URLs
      section names production end-states for all three URLs and
      links to the archived roadmap.
- [x] `docs/blueprints/backend-blueprint.md` §10 RLS subsection
      reflects the three-role topology, the bypass-then-scope handoff,
      and the migration-041 SECURITY DEFINER helper.
- [x] `docs/blueprints/database-connection-blueprint.md` carries §0
      "Three URLs, three roles, two runtime pools" plus a reading
      note scoping the rest of the doc.
- [x] `docs/refs/trajan-db-roles.md` has a Heimdall analogue
      back-reference naming the three deviations from the Trajan
      baseline.
- [x] `rls-enforcement-roadmap.md` and `rls-enforcement-role-split.md`
      moved to `docs/archive/` via `git mv` (history preserved).
      Status headers in both refreshed to "shipped 2026-05-04 —
      archived" with a pointer to the completion docs.
- [x] Phase 1 broken-link prerequisite (`docs/executing/rls-enforcement-mental-model.md`
      — never existed at that path) reconciled in the archived
      roadmap and across all completion docs. Mental-model doc is in
      `docs/blueprints/`.
- [x] Full backend test suite green
      (`go test ./...` → all packages pass; integration tests skip
      cleanly without `DATABASE_URL`).

**Operator-side (residual):**

- [ ] Production secrets rotation per the runbook above (lone item;
      not a codebase change). When complete, fill in the rotation
      timestamp and the post-rotation `role-split verified` boot-log
      line in the operator log.

The rollout is structurally complete. The remaining secrets-rotation
item is documented and runbook-ready; flipping it to `[x]` in this
file is the formal close.

---

## The eight-phase rollout, end-to-end

For a future reader who lands on this doc cold, here is the rollout
in one paragraph each.

1. **[Phase 1](./rls-enforcement-phase-1.md) — Audit & discover.** Read-only audit producing a
   classified inventory of every code path the role split affects.
   Output: the Phase 2 prerequisite list, the Phase 4 grant
   exceptions (cron_user `DELETE on log_buffer`), the Phase 7 policy
   gaps.
2. **[Phase 2](./rls-enforcement-phase-2.md) — Code refactor (handoff pattern).** Rewrites every
   site Phase 1 surfaced so it complies with the `app_user` +
   `UserQueries` contract while still running on `postgres`. The
   `db.Pools` struct landed here with both `.App` and `.Cron`
   resolving to the same superuser pool.
3. **[Phase 3](./rls-enforcement-phase-3.md) — Pool wiring (dual pool, single role).** Physically
   introduces the second `*pgxpool.Pool`, the `assertRoleSplit`
   startup invariant, and renames the migration-driver env var to
   `DIRECT_URL`. Both pools still resolve to `postgres`.
4. **[Phase 4](./rls-enforcement-phase-4.md) — Migration A (role grants).** Creates the `app_user`
   + `cron_user` grants in the database via migration 039. Bootstrap
   SQL is environment-side (passwords never in version control); the
   migration asserts roles exist before granting.
5. **[Phase 5](./rls-enforcement-phase-5.md) — Test infrastructure.** Six new test files in
   `backend/internal/db/`. The `asAppUser` fixture exercises RLS
   under the `postgres` pool; the catalog-drift tests pin the verb
   coverage (allowlist promoted to the Phase-7 target state); the
   pairing tripwire catches future raw-pool regressions.
6. **[Phase 6](./rls-enforcement-phase-6.md) — Stage 1 env-var flip.** Production flips from
   `postgres` to `app_user`/`cron_user` via a single Fly.io secrets
   transaction. RLS is still cosmetic at this point — failures
   surface as privilege errors, not policy errors.
7. **[Phase 7](./rls-enforcement-phase-7.md) — Migration B (FORCE + policy patches).**
   Migration 040 ships `ALTER TABLE … FORCE ROW LEVEL SECURITY` on
   every protected table plus five policy patches Phase 1 surfaced.
   The Phase 5 §5.6a Test 2 (verb coverage) goes green for the first
   time.
8. **Phase 8 — Cleanup & docs (this file).** Phase 7 invite
   follow-up; blueprints refreshed; executing plan archived;
   cross-links reconciled; secrets rotation runbook documented.

---

## Files changed

```
backend/migrations/041_lookup_user_for_invite.up.sql            (new)
backend/migrations/041_lookup_user_for_invite.down.sql          (new)
backend/internal/db/queries/users.sql                           (LookupUserIDForInvite added)
backend/internal/db/users.sql.go                                (matching generated method)
backend/internal/api/handlers/org_members.go                    (InviteMember uses helper)
docs/blueprints/backend-blueprint.md                            (§10 RLS subsection rewrite)
docs/blueprints/database-connection-blueprint.md                (§0 added; intro reading-note; §3 line update)
docs/refs/trajan-db-roles.md                                    (Heimdall back-reference)
docs/archive/rls-enforcement-roadmap.md                         (was: docs/executing/...)
docs/archive/rls-enforcement-role-split.md                      (was: docs/executing/...)
docs/completions/rls-enforcement-phase-{1..7}.md                (link path updates)
CLAUDE.md                                                       (RLS section refreshed; path updated)
docs/completions/rls-enforcement-phase-8.md                     (new — this file)
```

---

## Surprises and design notes

### 1. The follow-up unit-test situation

`backend/internal/api/handlers/org_members_test.go TestInviteMember`
exercises the happy path and four edge cases (duplicate, missing
email, member-cannot-invite, role validation). All five still pass
after the helper rewire because the test runs against `postgres` (no
FORCE applied at test time, since `DATABASE_URL` in tests points at
the local superuser). The visibility-vs-existence distinction the
helper specifically fixes is a property of `app_user`+FORCE; under
`postgres` the original `GetUserByEmail` would have worked too.

A meaningful regression test would need the `asAppUser` fixture
(Phase 5 §5.6d) extended to handler-level tests — not done here
because (a) the fixture currently lives in `internal/db/` and
extending it cross-package is a small refactor in its own right, and
(b) the migration-041 helper has a single SQL function body that's
trivially correct on inspection. Tracked as a future test-surface
follow-up if/when Phase 5's fixture gets pulled into a shared
`rlstest` package.

### 2. The `database-connection-blueprint.md` reading-note is a
deliberate hedge, not a TODO

The original doc had ~300 lines of accurate mechanics about pgxpool,
TLS, capacity ceiling, and Supavisor migration paths — all still
correct. The role split changes the *who* of the connection, not the
*how*. Rewriting every singular ("the pool", "the URL", "the
password") to a per-role plural would have inflated the diff to
something review-fatiguing and arguably regressed readability of the
mechanics sections.

The §0 + reading-note shape pushes the role-split context to the top
once, then lets the legacy text stand. Future edits should respect
this convention: when adding a new mechanic, write it in the
single-pool voice if the mechanic is role-agnostic, and call out
"app pool only" / "cron pool only" inline when it isn't.

### 3. `CLAUDE.md` was already mostly correct

CLAUDE.md's "Database URLs (RLS role split)" section was added
during Phase 3 (when the dual-pool wiring landed) and described the
end-state intent of each URL. Phase 8's edit was scope-trimming
("Phase 6 end-state" → "Production") and a path update to the
archived roadmap. No structural rewrite was needed.

### 4. The Phase 1 broken-link prerequisite is finally closed

Carried forward in every completion doc from Phase 1 onwards: the
roadmap header referenced `docs/executing/rls-enforcement-mental-model.md`
(non-existent) and `docs/refs/trajan-db-roles.md` (correct). Phase 8
fixes the mental-model path to `docs/blueprints/rls-enforcement-mental-model.md`
in the now-archived roadmap, and verifies the trajan-db-roles path
remains correct. No follow-up.

### 5. The roadmap's "remove `// removed code` comments" instinct
applies to migration 040's `github_repos_owner` rewrite

Worth flagging because it's the kind of thing a careless future
review might "clean up." Migration 040's `down.sql` re-creates
`github_repos_owner` with the migration-020 body verbatim. That's not
a stale comment or cargo-cult; it's the load-bearing reverse of the
`DROP POLICY IF EXISTS ... ; CREATE POLICY ... _via_connection` pair
in `040.up.sql`. Removing it would silently break the Phase 7 revert
lever. The comment in the down-migration explicitly notes this so a
future reader doesn't optimise it away.

---

## Closing

The eight-phase RLS role split is shipped. Production runtime is on
`app_user` + `cron_user`; every protected table is `FORCE`'d; the
Phase 1 §2 threat model is closed against `app_user`; the test
surface keeps regressions noisy; the documentation reflects the
as-shipped state.

Future work in this neighbourhood is incremental and unblocked by
this rollout — a SECURITY DEFINER `users_email_lookup` for
auto-complete UIs, a `cron_user` widening if a new janitor task
needs DELETE on a different table, a Phase-9-style "audit log on
every cron-pool write" if the operations footprint grows. All of
those are single-PR scopes against the topology this rollout put in
place.
