# RLS Enforcement — Phase 9 Completion: Pre-Deploy Polish

**Status:** complete (all seven items shipped; tripwires updated; build/vet/tests green).
**Parent plan:** [`../archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md)
**Roadmap:** [`../archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) — the eight-phase plan declared structural completion in Phase 8; this is the post-implementation polish pass that came out of the pre-deploy audit.
**Phase 8 completion (gating doc):** [`./rls-enforcement-phase-8.md`](./rls-enforcement-phase-8.md)
**Date:** 2026-05-04.

---

## Executive summary

Phase 9 is not in the original roadmap. It's the polish pass that came out of the
**pre-deploy code assessment** — three independent reviews (spec compliance,
security/migration safety, code quality) of the eight-phase work, run before
push. All three converged on a "cleared to push" verdict with a small list of
edits worth doing first to reach a 9/10 quality bar. This phase is that list.

After this commit:

- Migration 040 has a `SET LOCAL lock_timeout = '5s';` guard so the 22
  `ALTER TABLE … FORCE` statements fail fast instead of jamming production
  writers behind a stalled lock-wait.
- `agent_config_global` is `FOR SELECT` only. The legacy global singleton
  is read-only at runtime (verified: only `loop.go:104`'s `GetAgentConfig`
  hits this table); the prior `FOR ALL USING (true) WITH CHECK (true)`
  shape would have let any tenant `app_user` session globally rewrite
  default model / mode / system_prompt_override post-FORCE.
- `Server.Pool` and `Server.Queries` are gone from `handlers/server.go`.
  Both were transitional shims left over from Phase 2's incremental refactor;
  Phase 8's invite-by-email helper was the last remaining handler that
  needed `s.Queries`. Removing them eliminates the footgun where a future
  handler could write `s.Queries.GetX(...)` and silently bypass RLS.
- `Health` pings both pools instead of just `App`. Post-Phase-6 the cron
  pool authenticates as a different role with potentially different
  connectivity (different DSN, different PgBouncer client tier); a
  cron-pool outage was previously invisible to the LB.
- `README.md` documents the three-URL topology in a "Database URLs" table
  pointing at CLAUDE.md for the full role/grant rationale. New contributors
  reading the README first now learn about `CRON_DATABASE_URL` /
  `DIRECT_URL` / `HEIMDALL_ENV` immediately.
- A new `make bootstrap-roles` target wraps the Phase 4 bootstrap SQL
  with env-var-supplied passwords, ON_ERROR_STOP, and a verification
  query. Closes the "fresh-DB onboarding cliff" where migration 039
  raised an EXCEPTION with no in-tree path to satisfy it.
- The `cmd/heimdall/main.go` cron-pool failure path keeps its explicit
  `appPool.Close()` (the pre-deploy review flagged it as a redundant
  double-close, but `os.Exit` skips deferred functions, so removing it
  would have leaked the appPool's connections at process exit). A comment
  now documents the asymmetry so a future cleanup pass doesn't repeat the
  mistake.

Phase-9-specific design decisions, locked in this commit:

1. **Migration 040 was edited in place rather than amended via a new migration 042.**
   040 is on the in-flight branch but **has not yet shipped to production**
   (Phase 7 operator deploy is the next gate). Editing in place keeps the
   policy story in one file — anyone reading 040 sees the final shape with
   its lock_timeout and SELECT-only `agent_config_global` policy, not a
   "see migration 042 for the patch" trail. If 040 had already deployed,
   the right answer would be a forward-only 042; the pre-deploy window
   makes the in-place edit safe.
2. **`Server.Pool` and `Server.Queries` were dropped, not deprecated.**
   The Phase 2 doc acknowledged them as compat shims for incremental
   refactor; Phase 8's `LookupUserIDForInvite` was the last in-tree caller
   of `s.Queries`. With no remaining production callers, leaving the
   fields exported for "future safety" would have been the precise
   anti-pattern the role-split was meant to close — an unused raw-pool
   handle that any future handler can reach for. The pairing tripwire
   (`rls_pairing_test.go`) now has `App.Ping` / `Cron.Ping` regex patterns
   so health.go's new shape is still caught and the allowlist entry stays
   honest.
3. **The `agent_config_global` tightening did not also strip the unused
   `UpsertAgentConfig` sqlc query.** That belongs in a follow-up if it
   happens at all — sqlc-generated code is mechanical, the query file
   `agent_config.sql` doesn't carry the runtime-write story, and the
   policy is the actual access control. Touching the query without
   touching the SQL source would just re-generate it on the next
   `sqlc generate`. Right scope: leave the generated code, gate writes
   at the policy layer.
4. **`make bootstrap-roles` requires env-var passwords rather than
   prompting.** Two reasons: scriptability (CI / IaC paths) and accident
   prevention (an interactive prompt with a copy-pasted password from
   a secrets manager bypasses any operator's own `set +o history`
   discipline). The target fails loudly with the `openssl rand -base64 36`
   command in the error message when the vars are missing.

---

## What landed

### 1. Migration 040 — `lock_timeout` guard

**File:** `backend/migrations/040_rls_force_enforcement.up.sql`

Added at the top of the transaction, before any DDL:

```sql
SET LOCAL lock_timeout = '5s';
```

The `ALTER TABLE ... FORCE ROW LEVEL SECURITY` statement is metadata-only
(no row rewrite), so the *duration* of each statement is microseconds.
The risk is the *wait* — if any of the 22 affected tables is currently
held by a long-running query (autovacuum, an investigation tool query,
a slow report) when `migrate` reaches that table, every writer queues
behind us until the lock is released. `SET LOCAL lock_timeout = '5s'`
makes the migration abort the transaction in 5s instead of jamming
prod traffic indefinitely. golang-migrate runs each migration in its
own transaction, so SET LOCAL is the right scope (expires with the
txn, doesn't leak into subsequent migrations).

If the deploy aborts on lock_timeout, the right operator response is
"investigate what's holding the lock, then re-run `make migrate-up`" —
not "re-run with a longer timeout." 5s is comfortably above any
acceptable application query latency; if a statement is still running
after that, it's a sign of a deeper problem.

### 2. Migration 040 — `agent_config_global` tightened to SELECT-only

**File:** `backend/migrations/040_rls_force_enforcement.up.sql`

Before:

```sql
CREATE POLICY agent_config_global ON agent_config
    FOR ALL
    USING (true)
    WITH CHECK (true);
```

After:

```sql
CREATE POLICY agent_config_global ON agent_config
    FOR SELECT
    USING (true);
```

The `agent_config` table is the legacy global singleton, read by
`agent/loop.go:104` as a fallback when an app's per-app config row is
absent. Verified pre-edit that `UpsertAgentConfig` (sqlc-generated) has
**no runtime caller** — only `GetAgentConfig` is invoked. Under the old
policy + FORCE, any `app_user` session could have rewritten the
singleton's row, globally affecting default model selection / agent
mode / prompt override for any app falling through to the legacy
fallback. The SELECT-only policy preserves the runtime read path while
closing the write footgun.

If the singleton genuinely needs runtime writes in a future feature,
the right shape is a SECURITY DEFINER helper (mirroring migration 041's
`lookup_user_for_invite`), not a widening of this policy. Documented in
the policy comment so the next reader has the option in hand.

**Catalog allowlist updated** — `backend/internal/db/rls_catalog_test.go`
changed the `agent_config` entry from `{"ALL"}` to `{"SELECT"}`.
`TestCatalogDrift_PolicyVerbCoverage` would have caught the mismatch on
the next CI run; the allowlist update keeps the drift check honest.

### 3. `Server.Pool` and `Server.Queries` removed

**File:** `backend/internal/api/handlers/server.go`

Both fields and the `Pool: pools.App, Queries: db.New(pools.App)`
struct-init lines are gone. The `pgxpool` import dropped with them.
The doc comment that described them as "convenience references" for
the incremental Class D refactor was rewritten to describe the final
shape: "Pools is the only DB entry point; handlers go through
Pools.UserQueries or Pools.CronQueries."

**File:** `backend/internal/api/handlers/health.go`

`Health` now calls `s.Pools.App.Ping(...)` and `s.Pools.Cron.Ping(...)`
in sequence, returning 503 with `db: "app_pool_unreachable"` or
`db: "cron_pool_unreachable"` to identify which pool is down. The
expanded comment documents why the cron probe is load-bearing
post-Phase-6 (different role / DSN / connectivity tier).

**File:** `backend/internal/db/rls_pairing_test.go`

The pairing tripwire's regex `\bPool\.Ping\(` doesn't match
`s.Pools.App.Ping(...)` because the word boundary between `Pools` and
`.` blocks it. Without action the allowlist entry for `health.go`
would have gone stale and `TestPairing_AllowlistJustifications` would
have failed on the next CI run. Two new patterns —
`\bApp\.Ping\(` and `\bCron\.Ping\(` — restore the tripwire's
coverage of health.go's new shape and pre-emptively catch any future
handler that tries to bypass the chokepoint via the same syntactic
path. The allowlist entry's justification was rewritten accordingly.

**File:** `backend/internal/api/handlers/testhelpers_test.go`

Two struct-literal field assignments dropped (`Pool: pool, Queries:
queries`); the local `queries := db.New(pool)` line stayed because
`testEnv.Queries` is still consumed by `pipeline_test.go`,
`source_filters_github_test.go`, and `organizations_test.go` for
fixture setup. A short comment explains the asymmetry: production
`Server` no longer has `Queries`; tests still need a non-RLS read
path for fixture work.

### 4. `cmd/heimdall/main.go` cron-pool failure path

**File:** `backend/cmd/heimdall/main.go`

The pre-deploy review flagged the explicit `appPool.Close()` on the
cron-pool init-failure path as a "redundant double-close" with the
deferred close earlier in `main`. **The reviewer was wrong** —
`os.Exit(1)` skips deferred functions, so removing the explicit
close would have leaked the appPool's connections at process exit
(the OS reclaims them, but pgx's graceful drain doesn't fire and
the DB sees a TCP-close instead of a clean shutdown). The explicit
close stayed; a new comment documents the asymmetry so a future
cleanup pass doesn't repeat the same mis-reading. This is the only
Phase-9 review item I pushed back on rather than implemented as
suggested.

### 5. `README.md` — Database URLs section

Added between **Setup** and **Commands**: a three-row table mapping each
URL to its purpose, prod role, and dev fallback, plus the
`HEIMDALL_ENV=production` invariant note and a pointer to CLAUDE.md
for the full grant table. Also added a new step 3 to the Setup block
calling out `make bootstrap-roles` for production-shape installs and
naming dev as a skip.

### 6. `make bootstrap-roles` target

**File:** `Makefile`

```make
bootstrap-roles:
    @if [ -z "$$APP_USER_PASSWORD" ] || [ -z "$$CRON_USER_PASSWORD" ]; then \
        echo "error: APP_USER_PASSWORD and CRON_USER_PASSWORD must both be set."; \
        echo "  generate with: openssl rand -base64 36"; \
        echo "  then store in your secrets manager and re-run: APP_USER_PASSWORD=... CRON_USER_PASSWORD=... make bootstrap-roles"; \
        exit 1; \
    fi
    @psql "$${DIRECT_URL:-$$DATABASE_URL}" -v ON_ERROR_STOP=1 \
        -v app_pw="$$APP_USER_PASSWORD" -v cron_pw="$$CRON_USER_PASSWORD" \
        -c "CREATE ROLE app_user LOGIN PASSWORD :'app_pw';" \
        -c "CREATE ROLE cron_user LOGIN PASSWORD :'cron_pw' BYPASSRLS;" \
        -c "SELECT rolname, rolsuper, rolbypassrls FROM pg_roles WHERE rolname IN ('app_user', 'cron_user') ORDER BY rolname;"
    @echo "bootstrap complete. Verify the table above shows app_user(super=f, bypass=f) and cron_user(super=f, bypass=t), then run: make migrate-up"
```

Reuses the same `$${DIRECT_URL:-$$DATABASE_URL}` selector as
`migrate-up` / `migrate-down`, so it consistently uses the superuser
URL that the migrations themselves expect. `ON_ERROR_STOP=1` makes
`psql` non-zero exit on the first error (so a re-run against a DB
where roles already exist exits cleanly with `ERROR: role "app_user"
already exists` and the operator can choose between drop-and-re-run
or `ALTER ROLE ... PASSWORD ...`). The verification SELECT at the
end mirrors the Phase 4 doc's expected-output table.

### 7. Build / vet / test pass

- `cd backend && go build ./...` — clean
- `cd backend && go vet ./...` — clean
- `cd backend && go test ./...` — green; integration tests skip
  cleanly without `DATABASE_URL` set, exactly as designed.

---

## What did NOT land

### Pre-deploy review follow-ups deferred to post-bake

These items were called out in the pre-deploy review as "non-blocking
polish." They're tracked here so a future pass has the list:

- **Extend `rls_regression_test.go` to cover more tables.** The current
  cross-tenant regression covers `connections`, `log_buffer`,
  `conversations`, `connection_sources`, `app_source_filters` (5
  tables). The catalog test confirms policy *existence* on all
  RLS-enabled tables but not policy *correctness* — so a leak in
  `agent_log`, `investigations`, `monitoring_state`, `notification_log`,
  `org_members`, `applications`, `app_agent_config`,
  `investigation_schedules`, etc. would not be caught by regression
  today. Estimated 1–2 hours of test plumbing.
- **Always log `current_user` summary at startup.** Currently the
  role-split summary is logged only when `HEIMDALL_ENV=production`.
  A dev that accidentally points at a non-superuser role without
  setting `HEIMDALL_ENV` would have all cron-pool queries silently
  return zero rows. Two-line change in `main.go` to lift the
  `slog.Info("role-split verified", ...)` line out of `assertRoleSplit`
  and into the unconditional startup path.
- **Strip the unused `UpsertAgentConfig` sqlc query.** Generated code
  is mechanical, but the query file (`backend/internal/db/queries/agent_config.sql`)
  has a documentation footprint — leaving an unused write query when
  the policy has just removed the write path is mildly inconsistent.
  Defer to a future sqlc cleanup PR.

### Pre-deploy review item rebutted, not implemented

- **"Remove redundant appPool.Close in main.go."** Rebutted: `os.Exit`
  skips deferred functions, so the close is not redundant. The code
  stays as-is with a comment explaining the asymmetry. Documented in
  §4 above.

### Operator-side residuals (unchanged from Phase 8)

These are still pending after Phase 9 because they're operator actions,
not codebase actions. Phase 9 doesn't change their status:

- **Production role bootstrap** (Phase 4 operator window — now
  `make bootstrap-roles` makes this one command).
- **Phase 6 secrets flip** in the production secrets manager.
- **Phase 7 deploy of migration 040** (now with `lock_timeout` safety).
- **Phase 8 password rotation.**

---

## Acceptance check status

| Item | Status |
|---|---|
| 040 migration has `lock_timeout` | ✅ |
| 040 `agent_config_global` is SELECT-only | ✅ |
| Catalog allowlist updated to match | ✅ |
| `Server.Pool` / `Server.Queries` removed | ✅ |
| Pairing tripwire updated (`App.Ping` / `Cron.Ping` patterns + allowlist justification) | ✅ |
| `Health` probes both pools | ✅ |
| `cmd/heimdall/main.go` cron-failure path documented (rebutted as non-redundant) | ✅ |
| README documents three-URL topology | ✅ |
| `make bootstrap-roles` target added | ✅ |
| `go build ./...` clean | ✅ |
| `go vet ./...` clean | ✅ |
| `go test ./...` green | ✅ |
| Production role bootstrap executed | [ ] (operator) |
| Phase 6 secrets flip | [ ] (operator) |
| Phase 7 migration 040 deployed to prod | [ ] (operator) |
| Phase 8 password rotation | [ ] (operator) |

---

## Files changed

```
backend/migrations/040_rls_force_enforcement.up.sql    | lock_timeout guard + agent_config_global → SELECT-only
backend/internal/db/rls_catalog_test.go                | agent_config allowlist {"ALL"} → {"SELECT"}
backend/internal/api/handlers/server.go                | Drop Pool / Queries fields; rewrite doc comment
backend/internal/api/handlers/health.go                | Probe both pools; per-pool error attribution
backend/internal/db/rls_pairing_test.go                | Add App.Ping / Cron.Ping patterns; refresh allowlist note
backend/internal/api/handlers/testhelpers_test.go      | Drop Pool / Queries from struct-literal Server
backend/cmd/heimdall/main.go                           | Comment explaining why explicit appPool.Close stays
README.md                                              | Database URLs section + bootstrap-roles step
Makefile                                               | bootstrap-roles target
docs/completions/rls-enforcement-phase-9.md            | This doc.
```

Lines changed: ~80 LOC across modified files, ~210 LOC of new docs.
Build clean, vet clean, tests green.
