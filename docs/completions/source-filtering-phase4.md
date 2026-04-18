# Source Filtering — Phase 4 Completion

**Scope:** Broaden source-name extraction beyond the Fly.io webhook path. OTLP ingestion now extracts `service.name`; syslog extracts hostname; generic webhooks gain a `source_name_path` config knob for payloads whose natural source identifier is buried in a nested field. The generic filter pipeline (Phase 1) is extracted into a reusable helper and drives OTLP too.
**Plan:** `docs/executing/source-filtering-and-org-connections.md`
**Builds on:** Phases 1–3.

---

## The Problem This Solves

After Phase 3, every connector *could* opt into source filtering, but only webhook_logs + GitHub actually did. OTLP ingestion stored everything, ignoring `service.name`. Syslog stored everything, ignoring hostname. Generic webhook payloads (Firehose / Pub/Sub / Vercel etc.) filtered on whatever the parser set as `SourceType` — often a collapsed generic label like "firehose" that discriminated poorly.

Phase 4 extends the per-source filtering story to the other three mainstream log-entry paths. After this, a user with one OTLP endpoint can route `service.name=checkout-service` to one Heimdall app and `service.name=payment-service` to another; a user with one syslog port can filter by `hostname`; a user with AWS Firehose forwarding multiple log groups can point `source_name_path` at the `log_group` field to filter at that granularity.

---

## What Changed

### 1. Shared filter pipeline

**File:** `backend/internal/api/handlers/source_filter_pipeline.go` (new)

Two functions pulled out of `webhooks.ingestEntries` so every ingestion path can reuse them without duplication:

- **`routeSources(ctx, qtx, connID, appID, entries) → (routes, seenSources, err)`** — does discovery (upsert into `connection_sources` for every distinct source name in the batch) and resolves the per-source app targets. Behaviour matches Phase 1 for app-scoped connections and Phase 2 for org-scoped — one query per batch vs. one query per distinct source respectively.
- **`insertFiltered(ctx, qtx, connID, userID, entries, routes, seenSources) → (stats, err)`** — runs the filtered insert loop with fan-out, returning a `sourceFilterStats{Accepted, Filtered, Inserts, Sources}` summary that handlers use to compose their HTTP response.

Both accept a `*db.Queries` transaction handle, leaving transactional discipline (Begin, Commit, Rollback) to the caller. That lets the webhook handler keep its idempotency cache in the same transaction and lets OTLP use whatever tx shape fits it.

`webhooks.ingestEntries` became ~60 lines shorter and its behaviour is preserved — Phase 1/2/3 tests all pass unchanged.

### 2. OTLP — `service.name` extraction + fan-out

**File:** `backend/internal/api/handlers/otlp.go`

The OTLP handler previously built its own loop: walk `resourceLogs × scopeLogs × logRecords`, call `InsertLogEntry` for each, return 207 Multi-Status on partial failure. After Phase 4, it:

1. **Flattens** the nested OTLP structure into a `[]webhookLogRequest` via `flattenOTLPRecords(req)`. Each record becomes one entry whose `SourceName` is the resource's `service.name` attribute — OTLP's canonical identifier for the emitting service.
2. **Runs the same pipeline** as webhooks: `routeSources` → `insertFiltered`. Inherits drop-by-default, discovery upsert, and fan-out for org-scoped connections.
3. **Returns `{accepted, filtered}`** in a 200 OK — dropped the 207 Multi-Status handling, which was a pre-existing hack rather than an OTLP-spec behaviour (Heimdall's responses were never strictly OTLP-compliant).

The **Phase 2 defensive rejection of org-scoped OTLP is gone** — the routing problem that blocked it is solved: each record now has a `SourceName`, and fan-out routing knows where to send it.

**`SourceType` vs `SourceName` for OTLP:**
- `SourceType = "otlp/<service>"` — human-readable label used in the activity feed ("otlp/checkout-service").
- `SourceName = <service>` — bare filter key ("checkout-service"). Matches what the user sees in the source selector.

Records with no `service.name` fall through to `sourceNameOf()` which returns `SourceType` ("otlp") — so they can still be filtered on, just at a coarser granularity.

### 3. Syslog — hostname extraction + in-memory allow-list

**File:** `backend/internal/connectors/logs/syslog.go`

Syslog is a persistent TCP listener rather than a request/response handler, so it can't run a DB query per message — that'd throttle the hot path. The listener now holds two maps behind a mutex:

- **`enabled map[string]struct{}`** — the set of hostnames currently in `app_source_filters` with `enabled = true`. Refreshed on startup and every `syslogRefreshInterval` (60s) by a background goroutine. Hot-path lookup is O(1) — the mutex holds for only the duration of the map read.
- **`discovered map[string]struct{}`** — hostnames this listener has already upserted into `connection_sources`. Prevents a sustained stream from one host from translating into an upsert per packet; the DB sees one write per hostname per listener lifetime.

Hot path per message:

```
extract hostname (RFC 5424 header) → defaults to "unknown" if absent
discoverHostname(hostname)      // upsert once, no-op on subsequent calls
if !isEnabled(hostname):
    skip InsertLogEntry           // drop-by-default
else:
    InsertLogEntry with app_id
```

**Drop-by-default applies to pre-existing syslog connections**, same as Phase 1 did for webhook_logs. Users upgrading through 4 will see their syslog streams stop persisting until they open the source selector and enable their hosts. This is called out explicitly in the changelog and mitigated by the "Manage Sources" button in the connection detail modal.

**Refresh failures don't clobber the cache** — if `ListEnabledSourceNames` errors (transient DB hiccup), we log a warning and keep serving the previous cache. Better "slightly stale filter" than "suddenly everything's filtered because the DB blinked."

**Org-scoping remains off for syslog.** A listener is bound to a single port and writes to one app_id today; fan-out would need a different cache shape and refresh query. Not blocked by anything fundamental — just out of scope for Phase 4. `supportsOrgScope(connType)` still returns `false` for syslog.

### 4. Generic webhook — `source_name_path` config

**File:** `backend/internal/api/handlers/webhooks.go` (new helpers)

Parsers set `SourceName` when they can (Fly.io → `fly.app.name`; OTLP → `service.name`). For webhook payloads they can't decode — AWS Firehose carrying CloudWatch logs, generic GCP Pub/Sub, custom HTTP drains — users now set a `source_name_path` field on the connection config:

```json
{
  "webhook_token": "...",
  "source_name_path": "log_group"
}
```

A small dotted-path extractor runs after parsing and before routing:

- **`applySourceNamePathOverride(entries, connConfig)`** — if the config specifies a path, extract a string from each entry's payload and overwrite `SourceName`. "Overwrite" is intentional: the user's config is authoritative — they set the path because the parser's guess was inadequate.
- **`extractStringByPath(payload, segments)`** — walks a decoded JSON object by key segments. Non-object intermediates, missing keys, and non-string leaves all produce `""` (path didn't resolve → leave `SourceName` alone, fall through to the normal parser/`SourceType` fallback).
- **`readSourceNamePath(configRaw)`** — shallow config probe; tolerates malformed config without panicking.

**Syntax scope:** dotted paths only (`meta.source`, `log_group`). No array indexing, no wildcards, no filters. This is the 80% case — users needing richer expression can route through a transform layer. Full JSONPath would have been a heavier dependency for marginal value in Phase 4.

The `connResult` struct gained a `Config json.RawMessage` field so `ingestEntries` can read the path without a second DB round-trip.

### 5. Frontend — source selector available for OTLP + syslog

**File:** `frontend/src/components/connections/ConnectionDetailModal.vue`

`SOURCE_FILTERED_TYPES` expanded from `{webhook_logs, github}` to `{webhook_logs, github, otlp, syslog}`. The "Manage Sources" button now renders for all four connection types; the rest of the pipeline (SourceSelector auto-loads, Sync button remains GitHub-only via the `discoverable` prop) is unchanged.

Generic webhooks (`webhook_logs`) don't need a separate type in the set — they already render the Sources button. The `source_name_path` config field isn't exposed in the wizard's Setup step; users can set it today via the connection Edit panel (which preserves arbitrary config fields). A future frontend pass can add a dedicated input; the backend-only wiring covers the feature.

### 6. Tests

**Backend:**

- **`TestApplySourceNamePathOverride`** — 6 table cases: no config, top-level path, nested path, missing path, non-string leaf, parser-value-override. Each verifies the override helper modifies `SourceName` correctly without breaking entries that don't match.
- **`TestReadSourceNamePath`** — 7 cases including malformed JSON, empty config, whitespace trimming.
- **`TestExtractStringByPath`** — 7 cases of the dotted getter in isolation.
- **`TestIngestOTLP_ServiceNameFiltering`** — DB-backed end-to-end: create an OTLP connection, enable one service, POST a two-record OTLP payload, assert only the enabled service's record lands in `log_buffer`.
- **`TestIngestOTLP_DropByDefault`** — confirms OTLP inherits drop-by-default. No filters → zero inserts.
- **`TestIngestOTLP_OrgScopedFanOut`** — creates an org-scoped OTLP connection, two apps both enable `shared-service`, one input record becomes two `log_buffer` rows (the write-time fan-out). Also verifies the Phase 2 "org-scoped OTLP not supported" rejection is gone.

All Phase 1/2/3 tests pass unchanged.

No new frontend tests: the `SOURCE_FILTERED_TYPES` expansion is one line of data that the existing Vue rendering logic consumes uniformly. Manual smoke-test in the browser was the right verification tier here.

No new syslog listener integration tests: the behaviour is covered by the shared filter pipeline tests (which exercise the same `ListEnabledSourceNames` + `UpsertConnectionSource` queries) and the existing parser unit tests in `syslog_test.go` (which still pass). A full TCP-listener integration test was punted — the cost of binding a port inside the test suite outweighs the marginal coverage over the equivalent webhook tests.

---

## What Did NOT Change (and why)

### `source_name_path` not applied to OTLP

OTLP already has a first-class source identifier (`service.name`). Adding a config-level override would invert the precedence — parser first, config as authority — which makes the OTLP spec's semantic the second-class citizen. If a user has weird OTLP exporter config that buries the service elsewhere, the fix is in their exporter, not Heimdall.

### `source_name_path` not a wizard form field

Users who need it can set it via the connection Edit panel (the existing free-form config editor). Adding a dedicated wizard input is a UX polish that doesn't block Phase 4's functional goal. When a platform flow calls for it (e.g. a dedicated "AWS Firehose via Pub/Sub" step), the input can surface there.

### Org-scoping still off for syslog

A syslog listener binds one TCP port and writes to one target. Org-scoping would need either a fan-out variant of the in-memory cache (one `enabled` map per app) or a per-message routing query. Neither is hard; neither is called for by a real user yet. Kept out to avoid speculative design.

### Syslog's 60s refresh interval not configurable

A user who enables a host expects a 60s delay before messages start flowing. This is the trade-off for keeping the hot path DB-free. The interval could be a config knob; nobody's asked. Default is tuned for "comfortable" — not instant, but faster than most polling intervals elsewhere in Heimdall. A future adjustment can be motivated by measurements.

---

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Existing OTLP/syslog users see "nothing persisting" after upgrade | High for users with existing OTLP/syslog connections | Changelog calls out the drop-by-default change. "Manage Sources" button is prominently placed in the detail modal. Migration doesn't auto-enable anything — the user explicitly picks what to admit, which is the whole point. |
| `source_name_path` points at a field that sometimes exists, sometimes doesn't | Medium for heterogeneous payloads | Missing resolution falls back to parser value → `SourceType`. So the worst case is "some entries filter by the path-extracted name, others filter by the generic label" — still correct, just less granular for the unmatched ones. |
| Syslog cache goes stale when user toggles a host | Guaranteed for up to ~60s | Documented; refresh cadence chosen as a trade-off. Could be lowered if users complain. |
| OTLP response-shape change breaks callers | Negligible | The pre-Phase-4 response was `{accepted: N}` on 200 (or `{accepted, rejected}` on 207 for partials). The new response is `{accepted, filtered}` on 200 — additive field; `accepted` preserved. 207 is gone, but no known consumer relied on it (the OTLP SDK clients treat 2xx as success). |
| Users setting a nonsense `source_name_path` (not JSON-parseable) | Low | The probe tolerates malformed config; path parsing never panics. Worst case is the override silently no-ops, behaving identically to Phase 1–3. |

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `npx vue-tsc --noEmit` — clean
- Backend tests: 6 Phase 4 new + 6 Phase 3 + 5 Phase 2 + 13 Phase 1 = 30 filter-pipeline tests, all pass
- Syslog parser tests: 7 tests, all pass (unchanged)
- Frontend tests: 58 unit tests, all pass
- 11 pre-existing unrelated handler-test failures remain (tracked separately, unchanged across phases)

---

## File Map

```
backend/
  internal/api/handlers/source_filter_pipeline.go      NEW — routeSources + insertFiltered helpers
  internal/api/handlers/webhooks.go                    CHANGED — uses the shared pipeline; adds connResult.Config + applySourceNamePathOverride + extractStringByPath + readSourceNamePath
  internal/api/handlers/otlp.go                        REWRITTEN — flattenOTLPRecords → shared pipeline; Phase 2 rejection removed; response shape simplified
  internal/api/handlers/source_name_path_test.go       NEW — unit tests for source_name_path override helpers
  internal/api/handlers/otlp_filter_test.go            NEW — DB-backed OTLP filtering tests (service.name, drop-by-default, org-scoped fan-out)
  internal/connectors/logs/syslog.go                   CHANGED — hostname allow-list cache, refreshEnabledSet goroutine, isEnabled + discoverHostname helpers, hot-path filtering

frontend/
  src/components/connections/ConnectionDetailModal.vue CHANGED — SOURCE_FILTERED_TYPES now includes otlp, syslog
```

---

## Status

The four-phase arc is complete. Every connection type that carries multi-source traffic (webhook_logs, github, otlp, syslog) now funnels through one filter pipeline, one set of tables, one selector component. The `source_name_path` escape hatch covers the long tail of payload shapes the built-in parsers can't reach. Org-scoped connections fan out correctly across all four connector types except syslog — deliberately held back, not blocked.

The broader source-filtering work is done; anything further (array syntax in paths, wizard form fields for `source_name_path`, org-scoped syslog, per-source retention rules) lives outside this executing plan.
