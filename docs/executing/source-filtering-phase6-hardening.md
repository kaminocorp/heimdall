# Source Filtering — Phase 6 Hardening

**Status:** Not started
**Prereqs:** Phases 1–5 (committed as 18153bd)
**Target:** Clear a 8.5/10 production-readiness bar before deploy. Phases 1–5 ship at ~7.5/10 today; this plan closes the gap.

---

## Context

Four parallel review agents (backend authz, migrations, syslog runtime, frontend) assessed phases 1–5. No ship-blockers were introduced *by* phases 1–5, but seven issues keep the composite score below 8.5. Items here are ordered **deploy risk → operational risk → correctness polish → tech-debt follow-ups** so work can pause at any boundary and still leave the tree shippable.

**Not in this plan:**
- `InsertIdempotencyResult` running on the pool outside the batch tx (`webhooks.go:418`) — pre-existing from v0.45.0 (commit `bf5ab77`); not introduced by phases 1–5. File as a separate ticket.
- `jsonServerError` hardcoding 500 across dozens of handlers — broader error-hygiene pass, out of scope here.

---

## Tier 1 — Deploy-blocking (must land before prod push)

### 1.1 Migration 035 concurrent-writer safety — deploy runbook

**Files:** `deploy/runbooks/v0.46.0-preflight.sql` (new), `deploy/runbooks/v0.46.0-migrate.md` (new)
**Severity:** HIGH — could abort mid-deploy with `column "org_id" contains null values`.

Migration 035 does: `ADD COLUMN org_id UUID` (nullable) → `UPDATE connections SET org_id = ...` → `ALTER COLUMN org_id SET NOT NULL`. Between the UPDATE and the SET NOT NULL, a concurrent `INSERT INTO connections` lands with `org_id = NULL`, breaking the constraint.

**Migrations are immutable once committed** — editing `035_org_connections.up.sql` in place would cause schema drift on any environment that has already applied it. The `LOCK TABLE` and pre-flight can only help *while 035 runs*, so for environments where 035 has already applied successfully, this item is a no-op (the migration already passed). For environments where 035 has not yet run (prod cutover), protection goes in the **deploy runbook**, executed manually before `make migrate-up`.

**Fix:** Write `deploy/runbooks/v0.46.0-preflight.sql`:

```sql
-- v0.46.0 pre-flight: run BEFORE migrate-up on environments where 035 has not yet been applied.
-- Aborts if any connection's parent app has no org_id (would cause 035 to fail at SET NOT NULL).
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM connections c
    LEFT JOIN applications a ON c.app_id = a.id
    WHERE a.org_id IS NULL
  ) THEN
    RAISE EXCEPTION 'v0.46.0 pre-flight failed: connections exist whose parent app has no org_id. Investigate and resolve before running migrate-up.';
  END IF;
END $$;

-- Block concurrent writers for the duration of migration 035.
-- Release happens automatically at COMMIT of the migration tx.
LOCK TABLE connections IN EXCLUSIVE MODE;
```

And `deploy/runbooks/v0.46.0-migrate.md` explaining when and how to run it:

> 1. Check `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;` — skip this runbook if ≥ 35.
> 2. In the SAME psql session (so the LOCK TABLE persists across statements): `psql $DATABASE_URL -f deploy/runbooks/v0.46.0-preflight.sql -f <path to 035.up.sql> -f <path to 036.up.sql> -f <path to 037.up.sql>`
> 3. Verify: `SELECT count(*) FROM connections WHERE org_id IS NULL;` expect 0.
> 4. Record migration versions 035–037 in `schema_migrations` (the Go migrator does this automatically if you prefer to run via `make migrate-up` — but then LOCK TABLE won't persist across the separate migrator transactions).

Note the trade-off: running via the migrator (`make migrate-up`) applies migrations in separate transactions, so a `LOCK TABLE` in one migration doesn't persist. Runbook approach uses a single psql session for atomic lock semantics. Use the runbook approach for prod cutover; use `make migrate-up` for dev/staging where the race window is trivially short.

**Verification:** On a seeded DB, run the preflight against a simulated orphan row (`INSERT INTO applications (id, org_id, name) VALUES (...); UPDATE applications SET org_id = NULL WHERE id = ...;`) — the DO block must raise.

---

### 1.2 `source_name` length + null-byte validation on PUT and DELETE

**File:** `backend/internal/api/handlers/source_filters.go`
**Severity:** HIGH — unbounded user input lands in a partial index and hot-path cache.

`AddSourceFilter` (POST) rejects `len(name) > 256`. `UpdateSourceFilters` (PUT, bulk) and `DeleteSourceFilter` do not. A client can PUT arbitrary-length source names; they persist into `app_source_filters.source_name` (unbounded TEXT), leak into `idx_app_source_filters_app_lookup`, and get loaded into the syslog in-memory cache. Null bytes (`\x00`) are also accepted and break downstream tooling that treats strings as C-strings.

**Fix:** Extract the existing POST validation into a helper and call it from all three mutating endpoints:

```go
func validateSourceName(name string) error {
    name = strings.TrimSpace(name)
    if name == "" {
        return fmt.Errorf("source_name is required")
    }
    if len(name) > 256 {
        return fmt.Errorf("source_name exceeds 256 characters")
    }
    if strings.ContainsRune(name, 0) {
        return fmt.Errorf("source_name contains null byte")
    }
    return nil
}
```

Call sites: `AddSourceFilter`, each iteration of `UpdateSourceFilters`'s `req.Sources` loop (return 400 with the offending index), `DeleteSourceFilter`.

Also: in `UpdateSourceFilters`, treat empty `source_name` after trim as 400 rather than silent skip — the current silent behaviour hides client bugs.

**Test:** Extend `TestSourceFilters_CRUD` with a table case asserting 400 for `len=257`, `"\x00"`, and `""` on each of POST/PUT/DELETE.

---

## Tier 2 — Operational safety (strongly recommended pre-deploy)

### 2.1 Syslog `Health()` degrades on sustained refresh failure

**File:** `backend/internal/connectors/logs/syslog.go`
**Severity:** HIGH (operational) — cache can go stale for 10+ minutes with no external signal.

Phase 5 added `initialRefreshDone atomic.Bool` so `Health()` fails before the first successful load. But once set, it never goes back — post-startup refresh errors (logged via `slog.Warn` inside `refreshEnabledSet`) leave the cache frozen while Health() reports OK. Operators discover the problem from silent data gaps rather than a paging signal.

**Fix:** Track `lastRefreshAt time.Time` under `filterMu`. Update on every successful refresh. Have `Health()` return an error if `time.Since(lastRefreshAt) > 3 * syslogRefreshInterval` (i.e. more than ~3 minutes without a successful load — two missed ticks).

```go
// Inside refreshEnabledSet, after successful map swap:
s.filterMu.Lock()
s.enabled = newMap
s.lastRefreshAt = time.Now()
s.filterMu.Unlock()
s.initialRefreshDone.Store(true)

// In Health():
if !s.initialRefreshDone.Load() {
    return fmt.Errorf("syslog filter cache not yet loaded")
}
s.filterMu.Lock()
stale := time.Since(s.lastRefreshAt) > 3*syslogRefreshInterval
s.filterMu.Unlock()
if stale {
    return fmt.Errorf("syslog filter cache stale: last refresh %s ago", time.Since(s.lastRefreshAt))
}
return nil
```

**Test:** Add a unit test that flips the queries handle to a failing mock, advances a fake clock past `3*syslogRefreshInterval`, and asserts `Health()` errors. Recovery: swap the mock back, run refresh, assert Health() succeeds.

---

### 2.2 `discoverHostname` rollback race in syslog

**File:** `backend/internal/connectors/logs/syslog.go` (around lines 546–558)
**Severity:** MEDIUM — a failed first-packet upsert can leave a hostname permanently missing from `connection_sources`.

Current flow: release `filterMu` → call `UpsertConnectionSource` → on error, re-acquire `filterMu` and delete the `discovered` entry. Between the release and upsert, a second concurrent message from the same host observes the entry in `discovered` and short-circuits. If the first upsert fails and the second message arrives *during* that failure, the second message never triggers a retry. Later messages from the same host also short-circuit forever — the host never surfaces in the UI.

Low probability in practice (requires a DB error *and* a concurrent second packet *and* no subsequent listener restart), but when it fires the symptom is silent.

**Fix:** Replace the short-lived `discovered` entry with an "in-flight" marker. Simplest form: add a second map `discoveredPending map[string]struct{}` held while the upsert is in flight. A concurrent message that sees only the pending marker (not the `discovered` set) takes the slow path and retries the upsert itself — which is idempotent at the DB level (`ON CONFLICT DO UPDATE`). On success, promote to `discovered`; on failure, drop from pending.

Alternative (simpler but slower): hold `filterMu` across the upsert. Acceptable only if DB latency is typically sub-ms; profile first.

**Test:** Exercise with a mock queries handle that returns an error on the first call and success on the second; assert the second call is made.

---

### 2.3 `SourceSelector.discoverSources` forwards `appId`

**File:** `frontend/src/components/connections/SourceSelector.vue` (around line 139)
**Severity:** MEDIUM — latent 403 once OTLP/syslog gain discovery in a future phase.

The three CRUD helpers on `sources.ts` thread `{ appId }` into query params; `discoverSources` does not. Today the backend's `DiscoverSources` handler ignores `?app_id=` for GitHub (discovery is per-connection, not per-app). When discovery extends to other connector types that need app-scope resolution (e.g. a future Vercel projects endpoint), this becomes a silent 403 at post-install callback time — the exact regression class Phase 5 just fixed.

**Fix:** Pass `requestOpts.value` through:

```ts
await discoverSources(props.connectionId, requestOpts.value)
```

Update `discoverSources` in `frontend/src/api/sources.ts` to accept the same `{ appId }` bag and forward as `?app_id=`.

**Test:** Extend the existing `source.test.ts` with a contract check, or add a component smoke test once `SourceSelector.vue` gets a test harness (see Tier 4).

---

## Tier 3 — Correctness polish (should land same sprint)

### 3.1 `source_name_path` validation at connection write time

**File:** `backend/internal/api/handlers/connections.go` (in `CreateConnection` / `UpdateConnection` config validation), `backend/internal/api/handlers/webhooks.go` (where `extractStringByPath` is called).
**Severity:** LOW — silent no-op on malformed path hides operator config mistakes.

`strings.Split(".", ".")` → `["", ""]` → `extractStringByPath` returns `""`. Fine for defensiveness but means a typo (`"meta..source"`, `".log_group"`, `""`) silently disables filtering; the operator sees drop-by-default everywhere with no error.

**Fix:** In the config-parsing side of `CreateConnection`/`UpdateConnection`, validate:
- non-empty after trim
- no empty path segments when split on `.`
- no leading/trailing dot

Reject with 400 at write time. Leave `extractStringByPath` defensive (returns `""` on miss) — belt and braces.

**Test:** Unit tests in `source_name_path_test.go` covering `""`, `"."`, `"a.."`, `"..b"`, `".x"`, `"x."`.

---

### 3.2 Wizard scope step for OTLP

**File:** `frontend/src/components/connections/wizard/flows.ts`
**Severity:** MEDIUM — feature gap, not a correctness issue.

Backend `supportsOrgScope` returns true for `webhook_logs` AND `otlp` (confirmed in `connections.go`). Phase 4 removed the OTLP defensive rejection of org-scoped, and fan-out is supported. But `flows.ts` only inserts the scope step for `webhook_logs` and Fly.io drain — a user cannot create an org-scoped OTLP connection from the wizard. They must call the API directly.

**Fix:** Add the scope step to the OTLP wizard flow in `flows.ts`, matching the webhook_logs wiring. Verify the generated payload omits `app_id` when scope is `'org'`.

**Test:** Manual smoke-test in dev browser: create an org-scoped OTLP connection via the wizard, assert it lands with `app_id: null` and appears in every app's connection list.

---

### 3.3 `classifyStaleness` handles undefined safely

**File:** `frontend/src/types/source.ts` (line ~19)
**Severity:** LOW — defensive; current signature excludes `undefined`, but TypeScript narrowing doesn't always enforce that at runtime call sites.

If `undefined` ever reaches the function (partial server response, test mock), `new Date(undefined).getTime()` is `NaN`, and `NaN < hour` is `false` → always returns `'stale'` (red). That's misleading.

**Fix:** Widen the signature to `string | Date | null | undefined` and treat undefined as `'never'`:

```ts
export function classifyStaleness(lastSeenAt: string | Date | null | undefined): Staleness {
    if (lastSeenAt == null) return 'never'
    // ... rest unchanged
}
```

**Test:** Add a table case for `undefined` to `source.test.ts`.

---

### 3.4 `ConnectionsPage` GitHub callback with connection_id disambiguation

**File:** `frontend/src/pages/ConnectionsPage.vue` (around line 89)
**Severity:** LOW — today users rarely have two GitHub installs; tomorrow they might.

The `?github=installed` redirect branch does `store.connections.find(c => c.type === 'github')` — first match wins. With multiple GH App installs this can't disambiguate which one just finished installing.

**Fix:** Extend the server-side install callback redirect to include `&connection_id=<uuid>`. In `ConnectionsPage.vue`, look up by id: `store.connections.find(c => c.id === route.query.connection_id)`. Falls back to first-match if the param is absent (backwards compatible).

**Test:** Manual smoke-test with a freshly installed GitHub App. Check that the URL carries the connection id.

---

### 3.5 Stale index comment in `source_filters.sql`

**File:** `backend/internal/db/queries/source_filters.sql`
**Severity:** NITPICK — docs drift.

Update the comment (around line 58) claiming `ListEnabledSourceNames` is backed by `idx_app_source_filters_lookup`. After migration 037 the correct index for that query is `idx_app_source_filters_app_lookup` (the `(connection_id, app_id) WHERE enabled = true` partial index). The old index still serves `ListAppsEnabledForSource`.

**Skipped in this plan:** a `RAISE NOTICE` in `035.down.sql` warning about destructive deletion of org-scoped rows. Migrations are immutable once committed; the warning belongs in the deploy runbook (Item 1.1) instead. The file-level comment block at the top of `035.down.sql` already documents the destruction for anyone reading the file.

---

### 3.6 Page-visibility gate on `SourceSelector` polling

**File:** `frontend/src/components/connections/SourceSelector.vue` (line ~201)
**Severity:** NITPICK — 10s polling fires even on backgrounded tabs, wastes requests.

**Fix:** Guard the polling tick on `document.visibilityState === 'visible'`. On `visibilitychange`, resume. Minor but aligns with standard SPA hygiene.

---

## Tier 4 — Deferred tech debt (file as follow-ups, not blockers)

Listed here so they're not lost — each has a motivating signal that would prompt action later.

| Item | Motivating signal | Where to fix |
|---|---|---|
| `InsertIdempotencyResult` inside batch tx | Pre-existing from v0.45.0. Root cause: pool-scoped write after `tx.Commit`. | `webhooks.go:418` — move inside `qtx` before commit. |
| `jsonServerError` hardcoded 500 | Affects error-code hygiene across all handlers. | Refactor helper to accept status. |
| Fan-out batch query for org-scoped routing | Webhook latency climbs with "distinct sources per batch". | `source_filter_pipeline.go` `routeSources` → single query with `= ANY($2) GROUP BY source_name`. |
| `DiscoverSources` rate limit / cooldown | User hammers Sync → drains GH App 5000/hr quota. | Add per-connection last-sync timestamp; reject calls within cooldown window with 429. |
| `SourceSelector.vue` component tests | Regression like Phase 5 item #1 slips through type-checking. | Build a happy-dom + Vitest harness for async-polling components. |
| Syslog `atomic.Pointer[map]` for `enabled` | Hot-path lock contention at high QPS. | Lock-free read path. |
| Syslog per-connection TCP tracking | Shutdown stalls for up to `syslogReadTimeout` on idle connections. | Track sockets; force-close in `Close()`. |
| Syslog octet-counting framing (RFC 6587 §3.4.1) | User complaint from a non-newline-framed client. | Parse length-prefixed frames. |
| Migration 037 `CREATE INDEX CONCURRENTLY` | `app_source_filters` row count > ~100k. | Switch to concurrent index build; run outside tx. |
| `app_source_filters.updated_at` column | Audit trail for enable/disable flips. | Migration + sqlc regen. |
| `inserts` response field consumers | Diagnostic need on activity feed or stats pages. | Wire into ingestion stats UI. |
| Source-filter 403 vs 400 copy | User confusion on error toasts. | Map status to message in `extractApiError`. |

---

## Execution order

Work Tier 1 → Tier 2 → Tier 3 in sequence. Any stopping point after Tier 1 is shippable; stopping after Tier 2 hits the 8.5 bar; stopping after Tier 3 hits 9.0. Tier 4 is backlog.

Each item is independently landable — no cross-item dependencies within a tier. Suggested batching for commits:

1. **Commit A** — Tier 1 (items 1.1 + 1.2). Backend only. One migration edit + one handler edit + validation helper + tests.
2. **Commit B** — Tier 2 (items 2.1 + 2.2 + 2.3). Mixed backend + frontend. Syslog hardening + frontend one-liner.
3. **Commit C** — Tier 3.1 through 3.6. Mixed. Polish batch.

Each commit independently passes `go build ./... && go vet ./... && npx vue-tsc --noEmit && npm run test`.

---

## Verification checklist

- [ ] `go build ./...` clean
- [ ] `go vet ./...` clean
- [ ] `npx vue-tsc --noEmit` clean
- [ ] Backend source-filter tests all pass (32 tests from phases 1–5)
- [ ] New unit tests for validation helper, syslog staleness, discoverHostname race
- [ ] Frontend `npm run test -- --run` passes (58+ tests)
- [ ] Manual smoke-test: migration 035 on a seeded DB with concurrent writer simulation
- [ ] Manual smoke-test: syslog Health() degraded → recovered cycle
- [ ] Manual smoke-test: GitHub install callback with 1 and N GitHub connections
- [ ] Manual smoke-test: wizard creates an org-scoped OTLP connection end-to-end

---

## Handoff

On completion, this plan moves to `docs/completions/source-filtering-phase6.md` and is referenced from the next changelog entry (0.46.2 — Phase 6 Hardening). The Tier 4 backlog items get filed as separate tickets or inline TODO comments at their target file locations.
