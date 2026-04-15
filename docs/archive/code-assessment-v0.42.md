# Code Assessment — v0.42.15

**Date:** 2026-04-15
**Scope:** Full codebase (backend + frontend), all 8 phases + hardening series
**Method:** Automated test/lint suites, manual review of every production file across 6 parallel review passes, git history audit for flip-flops

---

## Test & Lint Results

| Check | Result |
|---|---|
| `go vet ./...` | Pass |
| `go test ./...` | Pass (6 packages, 52+ tests) |
| `vue-tsc --noEmit` | Pass (0 errors) |
| `vitest run` | Pass (7 suites, 52 tests) |
| `eslint src/` | **33 errors** (see L1, L2 below) |

---

## File Length Audit

No production file exceeds 500 lines. The only file over that threshold is `db/log_buffer.sql.go` (605 lines), which is sqlc-generated and should not be manually edited. The codebase is well-factored.

---

## Flip-Flop Audit

**1 confirmed flip-flop out of ~25+ hardening changes.** The hardenings are directionally consistent and additive.

**Confirmed:** Classifier confidence threshold in `classifier_lumber.go` went through three states: (1) dual-gated with Lumber + application code, (2) consolidated into application code only, (3) removed entirely. The final removal is well-reasoned (low-confidence auto-escalation caused PassthroughClassifier-like false positives), but the intermediate commit's stated intent was contradicted within the same hardening series.

**Not flip-flops (confirmed refinements):** `parseSeverityFromResponse` heuristic removal (refinement, structured markers preserved), scheduler serial-to-concurrent (anticipated evolution), connector insert error handling `return` to `continue` (consistent policy change), `tools_db.go` connection type narrowing (dead branch removal).

---

## Findings by Severity

### CRITICAL — Must fix before production push

#### C1. MaxBodySize middleware blocks webhook/OTLP ingestion over 1MB

**Files:** `api/router.go:22`, `handlers/webhooks.go:38`, `handlers/otlp.go`
**Impact:** Production log ingestion silently fails for payloads > 1MB

The global `MaxBodySize(1 << 20)` middleware at `router.go:22` wraps **all** request bodies in `http.MaxBytesReader(w, r.Body, 1MB)` before any handler runs. The webhook handler's own `io.LimitReader(r.Body, 10MB)` at `webhooks.go:38` wraps the *already-limited* body. Reading > 1MB hits the outer `MaxBytesReader` error first.

The comment on `middleware/cors.go:58` says "Handlers like the webhook endpoint that need larger bodies should apply their own `io.LimitReader` before this middleware fires" — but middleware fires *before* handlers, making this impossible.

**Fix:** Exempt `/api/webhooks/logs` and `/api/v1/logs` from the global body-size middleware, or apply `MaxBodySize` only to the authenticated route group rather than globally.

---

### HIGH — Should fix before production push

#### H1. GitHub connector: race condition on token fields

**File:** `connectors/codebase/github.go:70-86`

`token` and `tokenExpiresAt` are read and written without synchronization. `refreshTokenIfNeeded` is called from `apiGet`, which can run concurrently from multiple agent tool-use calls. A concurrent read of `tokenExpiresAt` while `refreshToken` writes it is a data race.

**Fix:** Add a `sync.Mutex` around token read/write in `refreshTokenIfNeeded` and `refreshToken`.

#### H2. SQL read-only bypass via PostgreSQL dollar-quoting

**File:** `agent/tools_db.go:140-200`

`containsSemicolon` and `stripStringLiterals` only handle single-quoted strings. PostgreSQL `$$`-quoting (e.g., `SELECT $$;$$ || (DELETE FROM users)`) bypasses the semicolon check. The `containsWriteKeyword` scan can also be evaded with dollar-quoted write keywords.

**Mitigated** by `default_transaction_read_only=on` at the Postgres connector level — the DB itself rejects writes. But the application-layer guard has a known hole.

**Fix:** Add `$$`-quoted string stripping alongside single-quote stripping, or use a simple SQL parser. Even a regex-based `\$[^$]*\$` detection would close the gap.

#### H3. ESLint config missing `dist/` exclusion

**File:** `frontend/eslint.config.js`

The flat ESLint config does not include an `ignores` entry for `dist/**`. This causes `npm run lint` to scan compiled build artifacts, inflating the error count from 33 (source-only) to 2113 and failing CI.

**Fix:** Add `{ ignores: ['dist/**'] }` to the config array.

---

### MEDIUM — Should fix, not blocking

#### M1. UserQueries done() always commits

**File:** `handlers/userqueries.go:35-39`

The `done()` callback unconditionally calls `tx.Commit()`. When used via `defer done()`, any handler that encounters an application-level error after a successful first write will still commit that write. The comment claims single-write-per-handler, but `CreateConnection` performs insert + poller start + status update within a single `UserQueries` scope.

**Fix:** Change `done()` to accept an error parameter (e.g., `done(err error)`) that calls `Rollback` when non-nil. Or switch to the common Go pattern of `defer tx.Rollback()` with explicit `tx.Commit()` only on the success path.

#### M2. Several read paths bypass RLS

**Files:** `handlers/applications.go:29,39,424`, `handlers/notifications.go:25,92,144,206,242,317`, `handlers/investigation_schedules.go:143`, `handlers/organizations.go:43`

Multiple handler read paths use `s.Queries` directly (the unscoped connection pool) instead of `UserQueries`. These queries are handler-level authorized via `authorizeApp`/`resolveOrgAndRole`, so the data returned is correct. But RLS policies are not active, meaning the database is not providing defense-in-depth for reads.

Write paths through `notifications.go` (`Create/Update/DeleteNotificationChannel`) also bypass RLS entirely.

**Risk:** Low — handler-level auth is consistently applied. But if RLS is intended as a security boundary (not just an optimization), these paths should use `UserQueries`.

#### M3. Scheduler tight-retries on provider failure

**File:** `agent/scheduler.go:232-234`

When `providerFailed` is true, `RunScheduledInvestigation` returns without advancing `last_run_at`. The schedule fires again on the next tick (every 60s), creating a tight retry loop that burns rate-limiter tokens while the provider is down.

**Fix:** On provider failure, either mark `last_run_at` (accept skipping this run) or implement exponential backoff (e.g., double the wait time on consecutive failures).

#### M4. MongoDB `sync.Once` permanently caches hostname resolution failures

**File:** `connectors/logs/mongodb.go:221-226`

If the first hostname resolution fails due to a transient network error, `sync.Once` caches the failure permanently. Every subsequent `Poll` call fails forever.

**Fix:** Replace `sync.Once` with a pattern that retries on error (e.g., `sync.OnceValues` with retry, or a manual once-with-retry).

#### M5. OrgTeamPage doesn't rollback optimistic role change on API error

**File:** `pages/org/OrgTeamPage.vue:76`

`handleRoleChange` sets `member.role = newRole` optimistically before the API call. On failure, `actionError` is set but the role is never reverted to its previous value. The UI shows the new role even though the server rejected it.

**Fix:** Save `oldRole` before the optimistic update and revert in the `catch` block.

#### M6. NotificationChannels deletes without confirmation

**File:** `components/notifications/NotificationChannels.vue:131`

`remove(ch)` calls the delete API immediately on click with no confirmation dialog. This removes a configured alert channel (Slack, Discord, email). Other destructive actions in the app consistently use confirmation modals.

**Fix:** Add a confirmation step consistent with the rest of the app.

#### M7. Inconsistent error extraction across frontend pages

**Files:** `pages/org/OrgTeamPage.vue`, `pages/org/OrgSettingsPage.vue`, `pages/org/OrgOverviewPage.vue`

Multiple files use inline `as { response?: { data?: { error?: string } } }` type assertions for error handling instead of the centralized `extractApiError` utility. This pattern is fragile and duplicated.

**Fix:** Replace inline error casting with `extractApiError()` calls.

---

### LOW — Clean up when convenient

#### L1. 7 unused variables/imports in production frontend code

| File | Unused symbol |
|---|---|
| `components/agent/ModelPicker.vue:84` | `formatContext` |
| `components/common/OrgDropdown.vue:4` | `OrganizationWithRole` |
| `components/org/AppCard.vue:4` | `formatRelativeTime` |
| `pages/AgentChatPage.vue:11` | `conversationId` |
| `pages/ConnectionsPage.vue:62-64` | `conn` (3 unused loop vars) |
| `pages/DashboardPage.vue:9` | `SkeletonBlock` |
| `test/setup.ts:1` | `config` |

#### L2. Dead code: `INVESTIGATION_STATUSES` constant and `ApiError` type

- `utils/constants.ts:7` — `INVESTIGATION_STATUSES` is never imported (remnant from removed reports feature)
- `types/api.ts` — `ApiError` interface is defined but never imported

#### L3. `SEVERITY_LEVELS` constant missing `error` level

`utils/constants.ts:5` — defines `['info', 'warning', 'critical']` but the notification type system includes `'error'`. Not currently causing a functional issue (the notification preferences UI hardcodes options), but inconsistent.

#### L4. `json.NewEncoder(w).Encode()` return value unchecked

Every handler ignores the error from `json.NewEncoder(w).Encode(...)`. In practice almost always nil, but if it fails after `WriteHeader`, the client gets a partial response.

#### L5. `extractText` concatenates text blocks without separator

`agent/loop.go:370-378` — multiple text blocks are joined with `+=` (no space/newline). Models rarely return multiple text blocks, so low risk.

#### L6. Duplicated `typeLabels` maps

`ConnectionBubble.vue:17` and `ConnectionDetailModal.vue:21` define nearly identical maps. Extract to a shared constant.

#### L7. Poller connectors skip `Connect()` call

`connectors/factory.go:17-50` — `StartPoller` calls `Poll()` directly without first calling `Connect()`. Benign for HTTP-based pollers but bypasses the `Connector` interface contract.

#### L8. Registry `Remove` calls `Close()` under lock

`connectors/registry.go:30-39` — holds `mu.Lock` while calling `c.Close()`, which could block (e.g., syslog's 5s drain timeout), blocking all other registry operations.

---

## Overall Assessment

### Score: 8.2 / 10

**Breakdown:**

| Dimension | Score | Notes |
|---|---|---|
| Functionality | 8.0 | C1 (MaxBodySize) is a production-blocking bug for large payloads |
| Security | 8.5 | Solid auth, RLS, read-only DB. H2 (dollar-quoting) is mitigated by connector |
| Maintainability | 8.5 | Clean architecture, consistent patterns, no oversized files |
| Code Quality | 8.0 | 33 ESLint errors, 7 unused symbols, some dead code |
| Test Coverage | 7.5 | Good unit test coverage where tests exist; some packages have no tests (middleware, config, db, github, notifications, ws) |
| Consistency | 8.5 | Hardenings are additive (1 minor flip-flop). Patterns are consistent across the codebase |

### Production Readiness

**Not yet safe to push** with C1 unresolved — the MaxBodySize middleware will silently reject webhook and OTLP payloads over 1MB, breaking production log ingestion.

After fixing C1, H1, and H3, the codebase clears 8.5/10 and is production-ready. The remaining medium/low findings are improvements that can ship incrementally.

### Recommended Fix Priority

1. **C1** — MaxBodySize webhook/OTLP conflict (blocks production ingestion)
2. **H3** — ESLint dist/ exclusion (blocks CI)
3. **H1** — GitHub token race condition (data race under concurrent access)
4. **H2** — Dollar-quoting SQL bypass (defense-in-depth gap)
5. **M1–M7** — Medium fixes (ship incrementally)
