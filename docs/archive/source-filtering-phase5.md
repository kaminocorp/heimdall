# Source Filtering — Phase 5 Completion (Production-Readiness Fixes)

**Scope:** Post-review fixes addressing seven issues surfaced by the four-agent production-readiness assessment of Phases 1–4. One critical runtime regression, two operational safety issues, one security test gap, three API-hygiene fixes. No new features, no behavior changes that users would notice in the happy path — this is hardening, not functionality.
**Preceded by:** Phases 1–4 (source filtering end-to-end).

---

## The Problem This Solves

Four focused review agents assessed the four-phase overhaul against a production-readiness bar. They converged on seven issues: one was a broken-on-first-render frontend regression in the primary GitHub onboarding flow, two were operational-safety problems where silent failure modes could turn into oncall headaches, and four were API/error-code hygiene or test-coverage gaps that would bite during ops but weren't exploitable.

Phase 5 fixes all seven with minimal diff: ~200 lines of code + a 4-line migration, no schema changes, no behavior changes on the happy path.

---

## What Changed

### 1. GitHub install-callback crash — the rename miss

**File:** `frontend/src/pages/ConnectionsPage.vue:89-93`

Phase 3's `manage-repos → manage-sources` rename collapsed two ref names into one (`sourceSelectorConnectionId`), but the onMounted GitHub-callback branch still referenced the old `repoSelectorConnectionId`. No test exercised the `?github=installed` redirect, so the symbol-not-found was latent until a user completed a GitHub App install and landed on `/connections`.

Fixed by renaming to the live symbol and also setting `sourceSelectorDiscoverable.value = true` so the post-install selector opens in Sync mode — matching what `openSourceSelector()` does for a user-initiated GitHub click. Now the two entry points to the selector (click in the UI vs. post-install redirect) have identical behaviour.

### 2. Syslog startup hardening — don't silently drop everything

**File:** `backend/internal/connectors/logs/syslog.go`

Previously, if the initial `refreshEnabledSet` call failed during `Listen()` (say a DB hiccup during app deployment), the `enabled` map stayed empty and every incoming syslog message was silently dropped by drop-by-default semantics until the 60-second background ticker eventually succeeded. The TCP listener looked healthy, the connection showed `active` in the UI — but zero logs were persisting.

Three defensive layers now:

1. **Initial retry with backoff.** Try up to 3 times at `1s, 2s, 3s` intervals before the accept loop starts. Transient DB blips during startup no longer register as silent drops.
2. **`initialRefreshDone atomic.Bool` flag.** Set to `true` on the first successful load. If it's still `false` after all retries, the listener proceeds anyway (better than crashing), but…
3. **`Health()` now fails** if `initialRefreshDone` is false. The connection shows as `error` in the UI and monitoring dashboards — operators see it immediately rather than discovering the problem from a silent data gap.

The background refresh ticker keeps trying indefinitely; once it succeeds, `initialRefreshDone` flips and `Health()` recovers on its own.

### 3. Syslog `discovered` map — bounded memory

**File:** `backend/internal/connectors/logs/syslog.go`

The per-process `discovered` map dedupes `connection_sources` upserts across a listener's lifetime. It had no cap — a long-running listener accepting forged hostnames (malicious or accidental — container sprawl, for instance) could grow the map indefinitely. At ~100 bytes/entry × 1M distinct hosts, that's 100 MB per listener.

Added a 10,000-entry cap (`syslogMaxDiscovered`). When full, the whole map clears. Worst case: one extra `UpsertConnectionSource` per active host after the clear — and since `UpsertConnectionSource` is `ON CONFLICT DO UPDATE`, the DB row already exists, so the "extra write" is just a `last_seen_at` bump. Trades perfect in-process dedupe for bounded memory; an attacker spraying forged hostnames now forces extra DB writes instead of exhausting listener RAM.

Clears emit a `slog.Warn` so a real memory-exhaustion pattern shows up in logs.

### 4. Partial-index coverage for `ListEnabledSourceNames`

**Files:** `backend/migrations/037_source_filter_lookup_index.up.sql` + `.down.sql`

Migration 034's `idx_app_source_filters_lookup` partial index was keyed `(connection_id, source_name) WHERE enabled = true`. This is perfect for `ListAppsEnabledForSource` (the org-fan-out hot path, filters by `connection_id + source_name`), but `ListEnabledSourceNames` filters by `(connection_id, app_id)` — it was doing an index scan on `connection_id` and a heap-filter on `app_id`. Fine at small scale; wasteful at scale as filter rows accumulate per connection.

Migration 037 adds a second partial index: `(connection_id, app_id) WHERE enabled = true`. Both query shapes now use index-only filtering. The two indexes are both small (partial, enabled-only), so the storage overhead is minimal.

Why a second index rather than replacing the first: the two queries have different access patterns. `ListAppsEnabledForSource` filters by source_name, so a composite `(connection_id, app_id, source_name)` would still work for it — but no better than the existing 2-column index, and worse for `ListEnabledSourceNames` if Postgres can't prune on the third column. Two focused partial indexes is cleaner.

### 5. Authz error codes on source-filter endpoints — 400 → 403

**File:** `backend/internal/api/handlers/source_filters.go`

`resolveSourceFilterApp` was returning every failure as HTTP 400 Bad Request, regardless of whether the underlying issue was a syntactic problem (malformed app_id) or a policy denial (app from a different org, app/connection org mismatch). This broke two client conventions:

- Client retry logic that distinguishes 4xx "caller problem" from 403 "denied".
- Monitoring alarms that page on elevated 4xx rates expect a 4xx distribution, not an authz problem disguised as validation noise.

Signature changed to return `(uuid.UUID, int, error)` where the int is the HTTP status. Now:

- **400** — missing or malformed `app_id` (syntactic).
- **403** — app doesn't match the connection's app (app-scoped case), user isn't a member of the app's org, or app's org ≠ connection's org (the cross-org attack boundary from Phase 2).

All four call sites (`ListSourceFilters`, `UpdateSourceFilters`, `AddSourceFilter`, `DeleteSourceFilter`) updated together; signature change is local to one file.

### 6. `DiscoverSources` error mapping — everything was 502

**Files:** `backend/internal/api/handlers/source_filters_discover.go`, `source_filters_github_test.go`

Phase 3's `DiscoverSources` handler collapsed every failure in `discoverGitHubRepos` into `502 Bad Gateway`. That hid three distinct causes from clients and oncall:

- **503 Service Unavailable**: our deployment has no GitHub App configured (`s.GitHub == nil`). The installation is fine, we're misconfigured — 503 tells the client "retry later once the operator fixes this".
- **500 Internal Server Error**: our stored connection config is corrupt (unparseable JSON, missing `installation_id`). This is our bug, not the user's.
- **429 Too Many Requests**: GitHub rate-limited us. Client should back off.
- **502 Bad Gateway**: genuine upstream failure (network, non-2xx from a valid request).

Introduced a typed `discoverError` with a `status` field. The helper now tags every error with the correct category; the handler dispatches via `errors.As`, calling `jsonError` with the tagged status and side-band-logging 5xx via `slog.Error` so oncall still sees the underlying cause.

Test `TestDiscoverSources_GitHubAppNotConfigured` updated to assert `503` instead of `502` — the new behavior is correct; the old test was asserting the wrong thing. Test now carries a comment explaining the 502/503 distinction for future readers.

**Subtle fix in the handler:** initial version routed 5xx through `jsonServerError`, but that helper hardcodes `500` regardless of the passed-in status. The fix: call `jsonError` directly with `de.status` and emit the slog side-band separately. This is why the assertion failed in the first test-run — the 503 tag was being squashed to 500.

### 7. Cross-org attack test

**File:** `backend/internal/api/handlers/org_connections_test.go`

The Phase 2 security boundary — user in orgs A and B cannot use app from B against connection in A — had no direct test coverage. `TestSourceFilters_OrgScopedRequireAppID` only exercised the "missing app_id" branch.

Added two tests:

- **`TestResolveSourceFilterApp_CrossOrgRejected`** — creates a second org with the test user as a member, then attempts GET/POST/PUT/DELETE on an orgA connection using orgB's app_id. All four must return 403. Final assertion queries `app_source_filters` to confirm no side-effect rows leaked into orgB's app (belt-and-braces: a handler that 403s but still writes would be catastrophic).
- **`TestResolveSourceFilterApp_AppScopedMismatchRejected`** — verifies that on an *app-scoped* connection, passing a different-but-valid app_id from the same org returns 403, while passing the connection's own app_id (or omitting the param) returns 200. Covers the other resolver branch.

The tests use the harness's single-user-in-one-org primitive and layer the second-org setup directly on `env.Pool` — same pattern `TestIngestWebhookLogs_OrgFanOut` uses for its second-app setup. No changes to `testSetup`.

---

## What Did NOT Change (and why)

### No refactor of `jsonServerError`

Helper still hardcodes 500. Phase 5 worked around it for `DiscoverSources` only. The helper is called in dozens of places across the handlers — a refactor would be larger scope and not block production for filter work specifically. If there's a follow-up "error-code hygiene pass" across all handlers, this is a natural scope for it.

### No frontend tests for `SourceSelector.vue`

Review flagged this as the biggest remaining frontend-coverage gap. Component tests with happy-dom + Vitest would add meaningful coverage but the current setup doesn't have a pattern for async-polling components (the selector polls every 10s while open). Deferring to a future frontend-test-infra investment; the critical runtime regression (item 1) was the only *broken* frontend behaviour surfaced. Manual smoke-test covered the rest.

### No fan-out batch query for org-scoped routing

`routeSources` still issues one `ListAppsEnabledForSource` per distinct source name. Review flagged as M1: O(distinct sources) DB round-trips per webhook batch. Fine at current scale (single-digit apps per org, handful of sources per batch). A single batched query with `= ANY($2) GROUP BY source_name` would collapse this to one round-trip — the obvious fix if/when fan-out becomes a hotspot. Not blocking production; flagged in the Phase 2 risk table already.

### No rate-limit / cooldown on `DiscoverSources`

Review flagged M1 in the handler pass: a user could hammer Sync to drain their GitHub App's 5000/hr quota. The 429 pass-through (item 6) gives the client an accurate signal to back off, which is the most important part. A server-side per-connection cooldown would add belt-and-braces but needs a store (ephemeral or persisted) for the last-sync timestamp; out of scope for a hardening pass.

### No bound on `UpdateSourceFilters` per-item length

Review flagged M4: the 256-char cap is in the POST but not the PUT. Still asymmetric; low-severity (the backend's 1MB body limit bounds total payload size even without per-item caps). Can land alongside any other polish on this handler.

---

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Migration 037 locks `app_source_filters` during index build | Low (small table at current scale) | `CREATE INDEX` without `CONCURRENTLY` is fine for the current row count (~thousands). If rows grow, switch to `CREATE INDEX CONCURRENTLY` in a future op. |
| Syslog `discovered` clear at cap causes DB write spike | Low (cap is 10k, clear-on-overflow is rare in practice) | Warn-level log on clear; upserts are cheap (ON CONFLICT bump of a single timestamp column). |
| `initialRefreshDone` race with Health checker | Negligible | atomic.Bool load/store; no shared-state concerns. |
| Cross-org attack test uses raw SQL to bypass API | Accepted | Other tests in `org_connections_test.go` already use this pattern for second-app/org setup. Keeps the test focused on the authz boundary. |

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `npx vue-tsc --noEmit` — clean
- `make migrate-up` — migration 037 applied cleanly
- Backend filter-pipeline tests: **all passing** (30+ tests across Phases 1–5 + 2 new cross-org tests = 32 tests)
- Frontend tests: **58/58 passing** (unchanged — no new component tests this phase)
- The 11 pre-existing unrelated failures continue to track separately

---

## File Map

```
backend/
  migrations/037_source_filter_lookup_index.up.sql       NEW (4 lines)
  migrations/037_source_filter_lookup_index.down.sql     NEW
  internal/connectors/logs/syslog.go                     CHANGED — initialRefreshDone atomic.Bool, startup retry loop, discovered map cap + clear, Health() failure signal
  internal/api/handlers/source_filters.go                CHANGED — resolveSourceFilterApp returns (uuid, int, error); four call sites updated
  internal/api/handlers/source_filters_discover.go       CHANGED — typed discoverError with per-kind status mapping; handler dispatches via errors.As
  internal/api/handlers/source_filters_github_test.go    CHANGED — TestDiscoverSources_GitHubAppNotConfigured now asserts 503 (was 502)
  internal/api/handlers/org_connections_test.go          CHANGED — TestResolveSourceFilterApp_CrossOrgRejected and TestResolveSourceFilterApp_AppScopedMismatchRejected added

frontend/
  src/pages/ConnectionsPage.vue                          CHANGED — GitHub callback uses sourceSelectorConnectionId (was broken ref); sets discoverable=true for parity
```

---

## Status

Source-filtering work is **production-ready**. The critical runtime regression is fixed, the operational safety gaps are closed, and the security boundary carries direct test coverage. Remaining review findings (frontend component tests, jsonServerError refactor, fan-out batching, PUT length cap, discover cooldown) are all in acceptable tech-debt territory — none block deployment, and each can land when the motivating signal arrives.

Ready to ship.
