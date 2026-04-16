# Changelog

- [0.45.2 — Connection Pause / Resume](#0452--connection-pause--resume-2026-04-16)
- [0.45.1 — RLS on Idempotency Table](#0451--rls-on-idempotency-table-2026-04-16)
- [0.45.0 — Webhook Ingestion Overhaul](#0450--webhook-ingestion-overhaul-2026-04-16)
- [0.44.3 — Fly.io Drain Wizard: Guided Setup](#0443--flyio-drain-wizard-guided-setup-2026-04-16)
- [0.44.2 — Connection Detail: URLs & Copy Buttons](#0442--connection-detail-urls--copy-buttons-2026-04-16)
- [0.44.1 — Webhook Token Display Fix](#0441--webhook-token-display-fix-2026-04-16)
- [0.44.0 — Fly.io Log Integration](#0440--flyio-log-integration-2026-04-16)
- [0.43.1 — Defense-in-Depth Hardening](#0431--defense-in-depth-hardening-2026-04-15)
- [0.43.0 — RLS on System Tables](#0430--rls-on-system-tables-2026-04-15)
- [0.42.21 — TypeScript Build Fixes](#04221--typescript-build-fixes-2026-04-15)
- [0.42.20 — Post-Assessment Hardening](#04220--post-assessment-hardening-2026-04-15)
- [0.42.19 — Final Production Hardening](#04219--final-production-hardening-2026-04-15)
- [0.42.18 — Pre-Deploy Hardening](#04218--pre-deploy-hardening-2026-04-15)
- [0.42.17 — Production Readiness Fixes](#04217--production-readiness-fixes-2026-04-15)
- [0.42.16 — Post-Assessment Hardening](#04216--post-assessment-hardening-2026-04-15)
- [0.42.15 — Medium-Severity Fixes](#04215--medium-severity-fixes-2026-04-15)
- [0.42.14 — High-Severity Fixes](#04214--high-severity-fixes-2026-04-15)
- [0.42.13 — Critical Security & Correctness Fixes](#04213--critical-security--correctness-fixes-2026-04-15)
- [0.42.12 — Low-Severity Polish](#04212--low-severity-polish-2026-04-15)
- [0.42.11 — Medium-Severity Fixes](#04211--medium-severity-fixes-2026-04-15)
- [0.42.10 — High-Severity Hardening](#04210--high-severity-hardening-2026-04-14)
- [0.42.9 — Critical Security & Correctness Fixes](#0429--critical-security--correctness-fixes-2026-04-14)
- [0.42.8 — Concurrency, Correctness & Test Coverage](#0428--concurrency-correctness--test-coverage-2026-04-14)
- [0.42.7 — Medium-Severity Fixes](#0427--medium-severity-fixes-2026-04-14)
- [0.42.6 — Post-Assessment Hardening](#0426--post-assessment-hardening-2026-04-14)
- [0.42.5 — Critical & High-Severity Fixes](#0425--critical--high-severity-fixes-2026-04-14)
- [0.42.4 — File Length Refactoring](#0424--file-length-refactoring-2026-04-13)
- [0.42.3 — Medium-Severity Fixes](#0423--medium-severity-fixes-2026-04-13)
- [0.42.2 — High-Severity Fixes](#0422--high-severity-fixes-2026-04-13)
- [0.42.1 — Post-Assessment Hardening](#0421--post-assessment-hardening-2026-04-13)
- [0.42.0 — Multi-Org Support & Team Management](#0420--multi-org-support--team-management-2026-04-13)
- [0.41.1 — Unified Connections Page: Three-Lane Layout](#0411--unified-connections-page-three-lane-layout-2026-04-13)
- [0.41.0 — Neutral Canvas Colour Rebalance](#0410--neutral-canvas-colour-rebalance-2026-04-13)
- [0.40.0 — Connection Wizard: Platform-First Redesign](#0400--connection-wizard-platform-first-redesign-2026-04-13)
- [0.39.0 — Ingestion Page Redesign: 3D Agent Nebula](#0390--ingestion-page-redesign-3d-agent-nebula-2026-04-13)
- [0.38.0 — Navigation Restructure & Infrastructure Triptych](#0380--navigation-restructure--infrastructure-triptych-2026-04-13)
- [0.37.0 — Activity Feed App-Scoping](#0370--activity-feed-app-scoping-2026-04-13)
- [0.36.0 — Top Header Nav Bar & Sidebar Slim-Down](#0360--top-header-nav-bar--sidebar-slim-down-2026-04-13)
- [0.35.0 — Technical Retro-Futurism UI Refresh](#0350--technical-retro-futurism-ui-refresh-2026-04-13)
- [0.34.1 — Dev-Mode Background Job Kill-Switch](#0341--dev-mode-background-job-kill-switch-2026-04-13)
- [0.34.0 — Activity Detail Modal & Supabase Poller Tuning](#0340--activity-detail-modal--supabase-poller-tuning-2026-04-13)
- [0.33.1 — Code Quality & Structural Cleanup](#0331--code-quality--structural-cleanup-2026-04-12)
- [0.33.0 — Expanded Model Catalogue & Picker](#0330--expanded-model-catalogue--picker-2026-04-12)
- [0.32.0 — Multi-App Setup & Settings](#0320--multi-app-setup--settings-2026-04-12)
- [0.31.0 — Scheduled Investigations](#0310--scheduled-investigations-2026-04-11)
- [0.30.3 — Activity Feed Rename & Log Retention](#0303--activity-feed-rename--log-retention-2026-04-11)
- [0.30.2 — Production Schema Drift Fix](#0302--production-schema-drift-fix-2026-04-10)
- [0.30.1 — GitHub Connection Flow Fixes](#0301--github-connection-flow-fixes-2026-04-10)
- [0.30.0 — GitHub App Provisioning](#0300--github-app-provisioning-2026-04-10)
- [0.29.1 — OpenRouter Post-Assessment Polish](#0291--openrouter-post-assessment-polish-2026-04-08)
- [0.29.0 — OpenRouter Multi-Provider Support](#0290--openrouter-multi-provider-support-2026-04-08)
- [0.28.0 — Fly.io VM Memory Upgrade](#0280--flyio-vm-memory-upgrade-2026-04-06)
- [0.27.1 — Go 1.25 Build Image](#0271--go-125-build-image-2026-04-06)
- [0.27.0 — Lumber Classifier Hardening](#0270--lumber-classifier-hardening-2026-04-06)
- [0.26.1 — Postgres Connector Interface Fix](#0261--postgres-connector-interface-fix-2026-04-06)
- [0.26.0 — Observability & Deployment Hygiene](#0260--observability--deployment-hygiene-2026-04-06)
- [0.25.0 — Code Assessment Cleanup](#0250--code-assessment-cleanup-2026-04-06)
- [0.24.5 — Poller, Parser & Shutdown Hardening](#0245--poller-parser--shutdown-hardening-2026-04-06)
- [0.24.4 — SDK Shutdown Safety](#0244--sdk-shutdown-safety-2026-04-06)
- [0.24.3 — UpdateConnection Validation](#0243--updateconnection-validation-2026-04-06)
- [0.24.2 — Production Hardening Pass](#0242--production-hardening-pass-2026-04-06)
- [0.24.1 — Ingestion Hardening](#0241--ingestion-hardening-2026-04-06)
- [0.24.0 — Webhook Parsers, API Pollers & Python/Go SDKs](#0240--webhook-parsers-api-pollers--pythongo-sdks-2026-04-06)
- [0.23.0 — OTLP HTTP Receiver & JS SDK](#0230--otlp-http-receiver--js-sdk-2026-04-05)
- [0.22.0 — Syslog TLS Listener](#0220--syslog-tls-listener-2026-04-05)
- [0.21.0 — Kamino Design System Alignment](#0210--kamino-design-system-alignment-2026-04-05)
- [0.20.5 — Stale-Asset Reload on Deploy](#0205--stale-asset-reload-on-deploy-2026-04-02)
- [0.20.4 — Public Site Header Overlap Fix](#0204--public-site-header-overlap-fix-2026-04-02)
- [0.20.3 — Lumber v0.9.0 Upgrade](#0203--lumber-v090-upgrade-2026-04-02)
- [0.20.2 — Blueprint View & Wizard Guard](#0202--blueprint-view--wizard-guard-2026-03-25)
- [0.20.1 — Action Button Color](#0201--action-button-color-2026-03-22)
- [0.20.0 — Security & Production Hardening](#0200--security--production-hardening-2026-03-22)
- [0.19.0 — Connection Wizard](#0190--connection-wizard-2026-03-22)
- [0.18.0 — Supabase Connector](#0180--supabase-connector-2026-03-22)
- [0.17.3 — Custom Dropdown Component](#0173--custom-dropdown-component-2026-03-21)
- [0.17.2 — SPA Routing Fix](#0172--spa-routing-fix-2026-03-21)
- [0.17.1 — Connection Test Modal & Dashboard Fix](#0171--connection-test-modal--dashboard-fix-2026-03-21)
- [0.17.0 — RLS Session Variable Fix](#0170--rls-session-variable-fix-2026-03-15)
- [0.16.1 — Server-Side Error Logging](#0161--server-side-error-logging-2026-03-15)
- [0.16.0 — GitHub App Integration](#0160--github-app-integration-2026-03-11)
- [0.15.1 — Notifications Build Fix](#0151--notifications-build-fix-2026-03-11)
- [0.15.0 — Notifications & Escalation](#0150--notifications--escalation-2026-03-11)
- [0.14.5 — Logo & Favicon](#0145--logo--favicon-2026-03-11)
- [0.14.4 — Feldgrau Colour Theme](#0144--feldgrau-colour-theme-2026-03-11)
- [0.14.3 — Public Layout, Heading & Mesh Refinement](#0143--public-layout-heading--mesh-refinement-2026-03-10)
- [0.14.2 — Hero Mesh Animation](#0142--hero-mesh-animation-2026-03-10)
- [0.14.1 — Public Site Header & Footer Redesign](#0141--public-site-header--footer-redesign-2026-03-10)
- [0.14.0 — Dockerfile Model Fix](#0140--dockerfile-model-fix-2026-03-09)
- [0.13.0 — Phase 8 Hardening](#0130--phase-8-hardening-2026-03-07)
- [0.12.0 — Multi-App UI & API](#0120--multi-app-ui--api-2026-03-07)
- [0.11.0 — Monitoring Mode](#0110--monitoring-mode-2026-03-07)
- [0.10.0 — Multi-App Data Model](#0100--multi-app-data-model-2026-03-07)
- [0.9.1 — UI Polish & Test Coverage](#091--ui-polish--test-coverage-2026-03-06)
- [0.9.0 — Public Website](#090--public-website-2026-02-27)
- [0.8.8 — Row Level Security](#088--row-level-security-2026-02-26)
- [0.8.7 — Connection Edit & Ping](#087--connection-edit--ping-2026-02-26)
- [0.8.6 — Connection Test on Create](#086--connection-test-on-create-2026-02-26)
- [0.8.5 — Connection Config Fields](#085--connection-config-fields-2026-02-26)
- [0.8.4 — Disable Scale-to-Zero](#084--disable-scale-to-zero-2026-02-26)
- [0.8.3 — Auth Guard Race Condition Fix](#083--auth-guard-race-condition-fix-2026-02-26)
- [0.8.2 — Missing Migration Fix](#082--missing-migration-fix-2026-02-26)
- [0.8.1 — Darker Background Tuning](#081--darker-background-tuning-2026-02-26)
- [0.8.0 — Techno-Brutalist Redesign](#080--techno-brutalist-redesign-2026-02-26)
- [0.7.1 — Supabase Auth Wiring](#071--supabase-auth-wiring-2026-02-26)
- [0.7.0 — Agent Log](#070--agent-log-2026-02-24)
- [0.6.1 — Post-Implementation Fixes](#061--post-implementation-fixes-2026-02-24)
- [0.6.0 — Agent Chat](#060--agent-chat-2026-02-24)
- [0.5.0 — Agent Loop](#050--agent-loop-2026-02-23)
- [0.4.0 — Webhook Log Connector](#040--webhook-log-connector-2026-02-23)
- [0.3.0 — Connections CRUD](#030--connections-crud-2026-02-23)
- [0.2.4 — Auth Me Endpoint & DB Pool](#024--auth-me-endpoint--db-pool-2026-02-22)
- [0.2.3 — Auth-Protected Routes](#023--auth-protected-routes-2026-02-22)
- [0.2.2 — JWT Verification Middleware](#022--jwt-verification-middleware-2026-02-22)
- [0.2.1 — Users Table & Auth Groundwork](#021--users-table--auth-groundwork-2026-02-22)
- [0.2.0 — Supabase Database](#020--supabase-database-2026-02-22)
- [0.1.3 — Infrastructure & DevOps](#013--infrastructure--devops-2026-02-20)
- [0.1.2 — Frontend Fixes](#012--frontend-fixes-2026-02-20)
- [0.1.1 — Backend Fixes & Hardening](#011--backend-fixes--hardening-2026-02-20)
- [0.1.0 — Scaffolding](#010--scaffolding-2026-02-19)

---

## 0.45.2 — Connection Pause / Resume (2026-04-16)

Connections can now be paused and resumed without deleting them. Useful for maintenance windows, cost control, debugging noisy sources, or keeping seasonal connections configured but dormant. No migration required — the existing `TEXT` status column accepts `paused` as a fourth value alongside `active`, `inactive`, and `error`, and the existing `WHERE status = 'active'` filters across ingestion, pollers, and listeners automatically exclude paused connections.

**Plan & assessment:** `docs/completions/connection-pause.md`
**Completion notes:** `docs/completions/connection-pause-phase{1..3}.md`

### Phase 1 — Backend: Accept and Enforce `paused` Status

**Files:** `handlers/connections_validate.go`, `handlers/connections.go`, `agent/tools_db.go`

#### Validation

`isValidConnectionStatus()` now accepts `"paused"` as a fourth valid value. This is the single gate that determines which status values the API accepts — without it, `PUT /api/connections/{id}` with `"status": "paused"` would return 400.

#### Update handler: immediate poller/listener shutdown

The `UpdateConnection` handler always stops the existing poller and listener for a connection before restarting them. The restart is now wrapped in a `if status != "paused"` guard. Without this, setting a connection to `paused` via the API would still restart its poller/listener in the same request — the pause would only take effect on the next server restart. On resume (`status = "active"`), the guard evaluates to true and the poller/listener starts up automatically.

#### Agent tool gate: block `query_database` on paused connections

After fetching the connection and before creating the database connector, `toolQueryDatabase` checks `conn.Status == "paused"` and returns a descriptive error. The error surfaces as an `isError: true` tool result to Claude (per the existing agent pattern), so the agent can inform the user rather than silently failing.

### Phase 2 — Frontend: Visual Treatment & Pause/Resume Button

**Files:** `types/connection.ts`, `ConnectionBubble.vue`, `ConnectionDetailModal.vue`, `stores/connections.ts`, `ConnectionsPage.vue`

#### TypeScript types

Added `'paused'` to `Connection.status` and `UpdateConnectionPayload.status` unions for compile-time exhaustiveness checking.

#### ConnectionBubble visual treatment

Four visual cues for paused connections: amber status dot with matching glow (distinct from green/red/grey), amber logo tint, dashed amber border communicating a "suspended" state, and 60% opacity (85% on hover) for dormancy. No pulse animation — paused is deliberately still.

#### ConnectionDetailModal: pause/resume button

A toggle button between Edit and Delete in the action bar. Only renders for `active` or `paused` connections — pausing an `inactive` or `error` connection has no practical effect. Active connections show an amber "Pause" button; paused connections show a green "Resume" button.

#### Store actions

`pauseConnection(id)` and `resumeConnection(id)` construct the full `UpdateConnectionPayload` from the existing connection, flipping only `status`. The existing `testConnection` action was updated to skip the local status update when a connection is paused — a successful ping no longer auto-resumes a paused connection.

### Phase 3 — Polish & Edge Cases

**Files:** `handlers/connections_test_handler.go`, `agent/loop.go`

#### Backend test endpoint: preserve paused status on ping

The `TestConnection` handler previously set the persisted status to `active` on success unconditionally. Now wrapped in `if conn.Status != "paused"` — paused connections still run the reachability test, but the persisted status is not touched. Defence-in-depth: the frontend guard (Phase 2) is a UX optimisation, this is the security boundary.

#### Agent system prompt: connection context injection

When the agent loop has an app context, it queries `ListConnectionsByApp` and appends a "Connected data sources" section to the system prompt listing each connection's name, type, direction, status, and ID. The closing instruction tells Claude that paused connections cannot be queried and to inform users they must resume first. This prevents blind tool calls and enables proactive user guidance.

### Post-assessment fix: CSS animation fill override

The `bubble-enter` animation filled forward at `opacity: 1`, silently overriding `.bubble-paused { opacity: 0.6 }`. Replaced with a dedicated `bubble-enter-paused` keyframe targeting `opacity: 0.6`, and added the paused state to the `prefers-reduced-motion` media query.

---

## 0.45.1 — RLS on Idempotency Table (2026-04-16)

Enabled Row Level Security on the `webhook_idempotency` table created in 0.45.0. Migration 032 added the table but missed RLS — the only table in the schema without it.

**Migration:** `033_rls_webhook_idempotency.up.sql`
**Completion notes:** `docs/completions/rls-webhook-idempotency.md`

Follows the system-table pattern from migration 030: RLS enabled with no policies. The table is only accessed by the webhook handler via the owner-role pool (`s.Queries`), which bypasses RLS. Non-owner roles (`anon`, `authenticated`) now see zero rows, closing the PostgREST exposure gap.

---

## 0.45.0 — Webhook Ingestion Overhaul (2026-04-16)

Five-phase overhaul of the webhook ingestion pipeline addressing eight gaps identified in a production assessment: partial-failure batch corruption, opaque error responses, detection false positives, internal ID leakage, no idempotency mechanism, no explicit format selection, payload field naming confusion, and Fly.io wizard accuracy. Every phase is independently shippable with zero cross-phase regressions.

**Assessment:** `docs/executing/webhook-ingestion-assessment.md`
**Plan:** `docs/executing/webhook-ingestion-improvements-plan.md`
**Completion notes:** `docs/completions/webhook-ingestion-phase{1..5}-*.md`

### Phase 1 — Correctness & Safety

**Files:** `handlers/webhooks.go`, `handlers/webhook_parsers.go`, `handlers/webhook_parsers_test.go`, `handlers/webhooks_test.go`

#### Transactional batch inserts

The handler previously inserted entries one at a time. If entry N failed validation, entries 1–(N-1) were already committed — silent duplicates on retry.

**Fix:** Two-pass approach. All entries validated upfront for `source_type` and `payload` before any database work. If any entry fails, a 400 is returned immediately with nothing written. All inserts then run inside a single `pgxpool.Pool.Begin()` / `Commit()` transaction with `Queries.WithTx(tx)`. Deferred `tx.Rollback()` ensures atomicity on insert failure.

#### Minimal response (no internal IDs)

The 201 response previously returned full `LogBuffer` database rows including `user_id`, `app_id`, `connection_id` — internal UUIDs meaningless to external callers.

**After:**
```json
{ "accepted": 5, "format": "native_batch" }
```

`parseWebhookPayload` now returns a 3-tuple `([]webhookLogRequest, string, error)` with a format identifier from each parser: `native`, `native_batch`, `vercel_ndjson`, `aws_firehose`, `gcp_pubsub`, `flyio_vector`.

### Phase 2 — Structured Errors & Format Transparency

**File:** `handlers/webhooks.go`

#### Structured error responses

New `webhookError` and `webhookFieldError` types replace bare error strings. Machine-readable error codes with per-entry field errors:

| Code | Status | When |
|------|--------|------|
| `unauthorized` | 401 | Missing/invalid bearer token |
| `invalid_json` | 400 | Body not valid JSON |
| `parse_failed` | 400 | Format detected but parsing failed |
| `empty_payload` | 400 | Parsing succeeded, zero entries |
| `validation_failed` | 400 | One or more entries missing required fields |

`validation_failed` collects **all** errors (not just the first), so automated callers fixing a batch of 50 entries don't need 50 retry cycles:

```json
{
  "error": "validation_failed",
  "message": "one or more entries failed validation",
  "format_detected": "native_batch",
  "details": [
    { "index": 2, "field": "source_type", "error": "required" },
    { "index": 4, "field": "payload", "error": "required" }
  ]
}
```

#### Server-side structured logging

Three `slog` log points: `Info` on successful ingestion (with `connection_id`, `format`, `entries`, `duration_ms`), `Warn` on parse error, `Warn` on validation failure.

### Phase 3 — Explicit Format Routes & Detection Hardening

**Files:** `handlers/webhooks.go`, `handlers/webhook_parsers.go`, `router.go`, `handlers/testhelpers_test.go`

#### Format-specific ingestion routes

```
POST /api/webhooks/logs              ← auto-detect (unchanged)
POST /api/webhooks/logs/{format}     ← format-specific (new)
```

Where `{format}` is one of: `flyio`, `vercel`, `firehose`, `pubsub`, `native`. Single parameterised route with a validated set — adding new formats requires one line in `validFormats` plus one `case` in `parseByFormat`. Invalid slugs return 404 listing valid formats.

**Refactored helpers:** `readWebhookRequest` (auth + body reading) and `ingestEntries` (validate + transact + respond) are shared by both handlers with zero duplication.

#### X-Heimdall-Format header override

The auto-detect handler checks `X-Heimdall-Format` before calling `parseWebhookPayload`. If present, auto-detection is skipped entirely. Gives callers who prefer the base URL a way to get deterministic parsing without endpoint reconfiguration.

#### Detection hardening

| Parser | Before | After | Why |
|--------|--------|-------|-----|
| NDJSON | Structural check (`isNDJSON` — any multi-line `{...}` input) | `Content-Type: application/x-ndjson` only | False-positived on Fly.io batches and native arrays |
| Firehose | `requestId` + non-empty `records` array | Also requires `records[0].data != ""` | App logs with `requestId`/`records` fields misrouted |
| Fly.io | `strings.HasPrefix(source_type, "fly")` | Exact match set: `fly_app`, `fly_io`, `fly_log_shipper`, `fly_app_logs` | `"flywheel"`, `"flutter"` no longer false-positive |

Removed dead `isNDJSON` function after it was decoupled from auto-detection.

### Phase 4 — Native Format v2

**Files:** `handlers/webhook_parsers.go`, `wizard/steps/StepWebhookSetup.vue`, `wizard/steps/StepFlyioDrainSetup.vue`

#### Dual-format native parsing

New v2 format with clearer field names, parsed alongside v1 with transparent normalisation:

| v2 field | v1 equivalent | Required | Rationale |
|----------|--------------|----------|-----------|
| `source` | `source_type` | Yes | Shorter, more natural |
| `level` | `severity` | No (defaults to `"info"`) | Standard across logging frameworks |
| `message` | (buried in `payload`) | No | First-class — the thing developers search for |
| `attrs` | `payload` | No | No collision with "payload" as HTTP body concept |

Version probing: `source` present → v2; `source_type` present without `source` → v1. If both present, v2 wins. Mixed v1+v2 batches rejected with clear error. Format strings: `native_v2`, `native_v2_batch`.

#### Deprecation signals for v1

When the request parses as v1 native format, the response includes RFC 8594 headers (`Sunset: 2026-10-01`, `Deprecation: true`, `Link` to successor docs) and a `"deprecated": true` field in the body. `slog.Warn` logged to track migration progress.

#### UI example payload updates

`StepWebhookSetup.vue` example updated from near-v2 to exact v2 spec. `StepFlyioDrainSetup.vue` Step 4 rewritten: VRL transform now outputs v2 format, clarified that enriched `fly` metadata auto-detection only applies to HTTP log drains (not NATS-based Log Shipper). Reference section split into **Native v2 (recommended)** and **Auto-detected (HTTP log drain only)** with muted secondary styling.

### Phase 5 — Idempotency

**Files:** `migrations/032_webhook_idempotency.up.sql`, `migrations/032_webhook_idempotency.down.sql`, `db/queries/webhook_idempotency.sql`, `db/webhook_idempotency.sql.go`, `db/models.go`, `handlers/webhooks.go`

#### Idempotency cache table

```sql
CREATE TABLE webhook_idempotency (
    connection_id   UUID  NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    idempotency_key TEXT  NOT NULL,
    response_status INT   NOT NULL,
    response_body   JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (connection_id, idempotency_key)
);
```

Composite PK — two different connections can reuse the same key. `ON DELETE CASCADE` from `connections` for automatic cleanup. Expiry index on `created_at` for efficient pruning.

#### Handler logic

Opt-in via `X-Idempotency-Key` header. Check happens in `ingestEntries` so it works for both auto-detect and format-specific routes:

1. Read header → query `GetIdempotencyResult` (24-hour TTL in `WHERE`)
2. **Cache hit:** return stored status + body, set `X-Idempotency-Replay: true` header
3. **Cache miss:** normal flow (validate → transaction → insert → commit)
4. After commit, cache the 201 response via `InsertIdempotencyResult` (`ON CONFLICT DO NOTHING`)

Only success responses cached — errors return normally without caching so callers can fix and retry. Cache insert is fire-and-forget (error discarded) — log entries are already committed, worst case is a future retry inserts duplicates (same as without idempotency).

`PruneExpiredIdempotencyKeys` query generated for periodic cleanup; not yet wired into a background job. The 24-hour TTL in `GetIdempotencyResult` means stale rows are never returned even unpruned.

### Test coverage

44 tests across all phases, all passing. `go vet` clean, `vue-tsc` clean.

| Phase | New/modified tests | Notable coverage |
|-------|-------------------|-----------------|
| 1 | 17 parser tests updated for 3-return-value signature, integration test updated | Batch semantics, minimal response shape |
| 2 | 7 new | Error serialization, format echo, `omitempty` behaviour |
| 3 | 7 new | Explicit format dispatch, hardened detection negatives (`flywheel`, no-data Firehose) |
| 4 | 9 new | v2 single/batch, no-attrs, no-level defaults, mixed batch rejection, both-fields precedence |
| 5 | 2 new | Response caching round-trips, deprecated field persistence |

---

## 0.44.3 — Fly.io Drain Wizard: Guided Setup (2026-04-16)

Rewrote the Fly.io Log Drain wizard step into a guided, step-by-step setup flow with pre-filled values — no more placeholder URLs or guesswork about payload format. Triggered by a real-world integration issue where a client's Fly Log Shipper was sending logs without the `source_type` field Heimdall requires, resulting in 400 errors with no clear guidance on what to fix.

**Files:** `wizard/ConnectionWizard.vue`, `wizard/steps/StepFlyioDrainSetup.vue`, `wizard/flows.ts`

### Early Connection Creation

The wizard now creates the webhook connection *before* entering the drain setup step (same pattern already used for the test step). This means the setup step has access to the real webhook URL and bearer token — no more `<your-heimdall-webhook-url>` placeholders.

**Change:** Added `flyio_drain` to the pre-creation check in `ConnectionWizard.vue`'s `goNext()`. Cleanup on wizard abandonment already handled by `handleClose()`.

### Five Numbered Setup Steps

The drain setup step is now a clear five-step walkthrough:

| Step | Action | What it shows |
|------|--------|---------------|
| 1. Launch the Log Shipper | `fly launch` command | Image name, guidance on region and deploy prompt |
| 2. Set Fly Access Token | `fly secrets set ACCESS_TOKEN=...` | Explains NATS log stream access |
| 3. Point it at Heimdall | `fly secrets set HTTP_URL=... HTTP_TOKEN=...` | **Pre-filled** with the real webhook URL and token |
| 4. Configure Vector | Conditional transform snippet | Stock image vs. custom config guidance |
| 5. Deploy | `fly deploy` | Sets expectation for when logs appear |

### Payload Format Documentation

Step 4 now explicitly documents Heimdall's Fly.io auto-detection contract: each log event must include either a `source_type` field starting with `"fly"` or a nested `fly` metadata object. For custom Vector configs, a ready-to-paste `remap` transform is provided.

A collapsible **Reference — Expected Payload Format** section lists every field Heimdall extracts: `fly.app.name`, `fly.machine.id`, `fly.region`, `log.level`, `message`, `timestamp`.

---

## 0.44.2 — Connection Detail: URLs & Copy Buttons (2026-04-16)

The connection detail modal now shows everything a user needs to configure a log shipper — no more hunting for URLs or re-reading docs.

**File:** `connections/ConnectionDetailModal.vue`

### Webhook URL & OTLP Endpoint

Webhook (`webhook_logs`) and OpenTelemetry (`otlp`) connections now display their full ingestion URL as a copiable row in the detail modal. The URL is derived from `window.location.origin` so it resolves correctly across environments (localhost in dev, `heimdallwatch.com` in production).

| Connection type | New row |
|-----------------|---------|
| `webhook_logs` | **Webhook URL** — `{origin}/api/webhooks/logs` |
| `otlp` | **Endpoint** — `{origin}/api/v1/logs` |

OTLP connections also now surface their Bearer Token (was previously a no-op `// OTLP shows endpoint info` comment).

### Inline Copy Buttons

Added a `copiable` flag to detail items. Any field marked copiable gets a small **Copy** / **Copied** button inline, using the same clipboard-with-fallback pattern as `CopyableField.vue`. Applied to: Webhook URL, OTLP Endpoint, and both Bearer Tokens.

---

## 0.44.1 — Webhook Token Display Fix (2026-04-16)

Fixed the connection detail modal not displaying the Bearer Token for webhook connections.

**File:** `connections/ConnectionDetailModal.vue`

The backend stores the auto-generated webhook token under the config key `webhook_token`, but the detail modal was reading `cfg.token` — a field that doesn't exist. The token was generated and persisted correctly; it simply never rendered in the UI.

**Fix:** `cfg.token` → `cfg.webhook_token` in the `webhook_logs` case of the detail builder.

---

## 0.44.0 — Fly.io Log Integration (2026-04-16)

Full Fly.io log integration: a rewritten backend poller, a dual-mode connection wizard, and a webhook parser for the Fly Log Shipper. Users can connect Fly.io apps to Heimdall via two paths — **Log Drain** (near-real-time, recommended) or **API Polling** (zero-setup fallback) — from a single unified wizard flow.

**Proposal:** `docs/executing/flyio-log-integration.md`
**Completion notes:** `docs/completions/flyio-phase{1,2,3}-*.md`

### 1. App-Level Poller Rewrite
**File:** `connectors/logs/flyio.go`

Replaced the N+1 per-machine polling approach with a single call to the Fly.io app-level NDJSON logs endpoint (`GET /v1/apps/{app}/logs`).

**Before:** The poller listed all machines via `GET /v1/apps/{app}/machines`, filtered to `started`/`running` state, then fetched `GET /v1/apps/{app}/machines/{id}/logs?limit=200` for each one. A 10-machine app made 11 API calls every 30 seconds. Stopped and crashed machines — whose final log lines are arguably the most important — were silently skipped.

**After:** A single API call returns NDJSON across all machines regardless of state. Parsed with `bufio.Scanner` for streaming memory efficiency. The `Connect()` / `Health()` validation now also hits the logs endpoint (proving both token validity and app existence) instead of the machines list.

| Metric | Before | After |
|--------|--------|-------|
| API calls per poll (10 machines) | 11 | 1 |
| Crash logs captured | No | Yes |
| Endpoint stability | Undocumented per-machine | Semi-public app-level (`flyctl` depends on it) |
| Per-machine 200-entry cap | Yes | No (streams all since cursor) |

**Removed:** `pollMachineLogs()` — all per-machine logic was in this function; no longer needed.

**Added:** `logsEndpoint()` helper (URL builder with proper escaping), `flyLogEntry` struct capturing the full NDJSON schema (`timestamp`, `message`, `level`, `instance`, `region`, `meta`). Stored payloads now include `instance` and `region` fields that weren't available in the per-machine approach.

**Tests:** 15 new unit tests in `flyio_test.go` — constructor validation, Connect auth/not-found, Poll request formation, empty responses, rate limiting, server errors, malformed NDJSON, cursor filtering, URL escaping, `normalizeSev` mappings, lifecycle.

### 2. Dual-Mode Connection Wizard
**Files:** `wizard/flows.ts`, `wizard/ConnectionWizard.vue`, `wizard/steps/StepFlyioMode.vue`, `wizard/steps/StepFlyioAuth.vue`, `wizard/steps/StepFlyioDrainSetup.vue`

The Fly.io entry in the platform grid is now available (`available: true`) with a dual-mode wizard flow:

```
Step 1: Name           → Connection name
Step 2: Mode           → "Log Drain (recommended)" or "API Polling"

Drain path:
  Step 3: Drain Setup  → Webhook URL/token info + copy-paste Fly Log Shipper CLI commands
                          → Creates a webhook_logs connection

Polling path:
  Step 3: Auth         → Fly app name + API token
  Step 4: Test         → Validates credentials against Fly.io Logs API
                          → Creates a flyio connection
```

**Dynamic `connectorType`** — this is the first wizard flow where the backend connector type changes based on user input mid-wizard. Two new computeds in `ConnectionWizard.vue`:

- `effectiveSteps` — filters the step list based on `state.config.flyio_mode` (drain excludes Auth + Test; polling excludes Drain Setup). Falls through to the static `selectedFlow.steps` for all other flows.
- `effectiveConnectorType` — returns `'webhook_logs'` for drain or `'flyio'` for polling. Falls through to `selectedFlow.connectorType` for all other flows.

The `flyio_mode` key is stripped from `state.config` before the API call — it's wizard-internal routing state that the backend doesn't need.

**New step components:**

- `StepFlyioMode` — two large toggle buttons with descriptions and badges ("Recommended" / "Zero setup"). Drain pre-selected as default. Always valid.
- `StepFlyioAuth` — app name + API token form fields. Follows the `StepSupabaseAuth` pattern. Valid when both non-empty.
- `StepFlyioDrainSetup` — display-only step with 4 copyable CLI commands for deploying the Fly Log Shipper. Uses the existing `CopyableField` component. Auto-valid on mount.

**Flow metadata updated:** description changed from "Ship logs from Fly.io apps via syslog drain" to "Ship logs from Fly.io apps via log drain or API polling"; subtitle from "via Syslog drain" to "Log Drain or API Polling"; `connectorType` from `'syslog'` to `'flyio'` (default, overridden dynamically).

### 3. Vector/Fly Log Shipper Webhook Parser
**File:** `handlers/webhook_parsers.go`

Added auto-detection and parsing for the JSON format sent by the Fly Log Shipper (a Vector container that connects to Fly's internal NATS log stream and forwards to HTTP sinks).

**Detection heuristic** (`isVectorFlyPayload`): checks for a `fly` nested JSON object (Fly Log Shipper's enriched format) or a `source_type` starting with `"fly"` (simpler Vector configurations). Handles both single objects and arrays (Vector batch mode). Sits in the detection chain after GCP Pub/Sub and before the Heimdall native fallback.

**Input format:**
```json
{
  "message": "request completed in 12ms",
  "timestamp": "2026-04-15T12:00:00.123Z",
  "host": "e784079c",
  "source_type": "fly_io",
  "fly": {
    "app": { "name": "my-fly-app" },
    "machine": { "id": "e784079c" },
    "region": "lhr"
  },
  "log": { "level": "info" }
}
```

**Field extraction:**

| Output field | Source | Fallback |
|-------------|--------|----------|
| `SourceType` | `"flyio/" + fly.app.name` | `source_type` from Vector, then `"flyio"` |
| `Severity` | `log.level` → `normalizeSeverity()` | `extractSeverityFromMap()` (checks `severity`, `level`, `log_level`, `error_severity`) |
| `Payload` | Enriched JSON with `message`, `timestamp`, `host`, `app_name`, `machine_id`, `region` | — |

**Tests:** 6 new tests — single entry, batch array, severity mapping, detection by source_type, fallback without fly metadata, negative detection (verifies native/other formats don't false-positive).

### End-to-End Drain Path

With all three phases in place, the drain path works as follows:

```
User selects "Fly.io → Log Drain" in wizard
  → Wizard creates webhook_logs connection (server generates bearer token)
  → User deploys Fly Log Shipper in their Fly org with webhook URL + token
  → Log Shipper connects to Fly's internal NATS stream
  → Log Shipper POSTs JSON to /api/webhooks/logs
  → isVectorFlyPayload() detects Fly format
  → parseVectorFlyEntry() extracts message, severity, Fly metadata
  → InsertLogEntry() stores in log_buffer
  → Monitoring loop classifies via Lumber → escalates to Claude
```

---

## 0.43.1 — Defense-in-Depth Hardening (2026-04-15)

Ten fixes from a post-v0.42.20 assessment covering one critical performance issue, three high-severity defense-in-depth gaps, and six medium correctness/robustness improvements. None were functionally blocking — all are hardening for production resilience at scale.

**Migration:** `031_assessment_fixes`

### C1. Webhook token index now covers OTLP connections
**File:** `migrations/031_assessment_fixes.up.sql`
**Fix:** The partial index `idx_connections_webhook_token` had predicate `WHERE type = 'webhook_logs'`, but `GetConnectionByWebhookToken` filters on `type IN ('webhook_logs', 'otlp')`. OTLP connections fell through to a sequential scan on `connections`. Dropped and recreated the index with the wider predicate.

### H1. SQL validator blocks dangerous Postgres functions
**File:** `agent/tools_db.go`
**Fix:** `isReadOnlySQL` allowed any statement starting with `SELECT`. The PostgreSQL function `set_config('default_transaction_read_only', 'off', false)` is callable via `SELECT set_config(...)` and would pass validation, potentially disabling the write guard. Added a blocklist for `SET_CONFIG`, `PG_READ_FILE`, `PG_WRITE_FILE`, `LO_IMPORT`, and `LO_EXPORT`. Six new test cases in `tools_db_test.go`.

### H2. `monitorTick` / `schedulerTick` always join goroutines
**Files:** `agent/monitor.go`, `agent/scheduler.go`
**Fix:** When the context was cancelled while waiting for the semaphore, both functions returned immediately — `wg.Wait()` at the end was never reached, orphaning in-flight goroutines. Moved to `defer wg.Wait()` at the top of both functions so goroutines are always joined regardless of control flow.

### H3. TOCTOU gap closed in notification channel + schedule CRUD
**Files:** `handlers/notifications.go`, `handlers/investigation_schedules.go`, `queries/notification_channels.sql`, `queries/investigation_schedules.sql`
**Fix:** Two compounding problems: ownership checks ran outside the transaction (via `s.Queries`), and the Update/Delete SQL filtered only by `id` with no `app_id` guard. Added `AND app_id` to the `UPDATE` and `DELETE` queries for both `notification_channels` and `investigation_schedules`, and moved all ownership checks inside the `UserQueries` transaction.

### M1. Postgres connector uses `url.URL` struct builder
**File:** `connectors/database/postgres.go`
**Fix:** User and password were escaped with `url.PathEscape`, which does not escape `@` or `:` — characters with structural meaning in URL userinfo. A password containing `@` would be misinterpreted as the host delimiter. Replaced `fmt.Sprintf` + `PathEscape` with Go's `url.URL` struct builder (`url.UserPassword` handles encoding correctly). Also added port range validation (1–65535).

### M2. `EmitLog` moved after `commit()` in `CreateApplication`
**File:** `handlers/applications.go`
**Fix:** `EmitLog` fired before `commit()`. If the commit failed, a phantom audit log entry was written for an application that was rolled back. Moved the call to after commit, matching the existing pattern in `DeleteApplication`.

### M3. `UpdateConnection` reads existing row inside transaction
**File:** `handlers/connections.go`
**Fix:** The handler fetched the existing connection via `s.Queries` (outside the transaction) to read defaults for omitted fields, then wrote inside a `UserQueries` transaction — a TOCTOU gap. Moved `GetConnectionByUser` and the webhook token preservation logic inside the transaction.

### M5. Notification dispatch goroutine has 30s timeout
**File:** `agent/monitor.go`
**Fix:** The fire-and-forget notification goroutine used `context.WithoutCancel` with no timeout. If the notification target hung indefinitely, the goroutine would leak permanently. Wrapped in a 30-second `context.WithTimeout` inside the goroutine.

### M8. Nullable JSONB columns map to `json.RawMessage`
**File:** `sqlc.yaml`
**Fix:** Nullable JSONB columns (`agent_log.detail`, `investigations.findings`, `investigations.tool_trace`) mapped to `[]byte` in generated Go code. When marshaled to JSON for API responses, they produced base64-encoded strings instead of inline JSON objects. Added a nullable JSONB override to `sqlc.yaml` and regenerated.

### M9. Duplicate unique index on `organizations.slug` dropped
**File:** `migrations/031_assessment_fixes.up.sql`
**Fix:** The column definition had `slug TEXT NOT NULL UNIQUE` (implicit index) and an explicit `CREATE UNIQUE INDEX idx_organizations_slug`. PostgreSQL maintained both — double storage, double write overhead. The explicit index is now dropped.

---

## 0.43.0 — RLS on System Tables (2026-04-15)

Defence-in-depth: enabled Row Level Security on the two remaining tables that lacked it — `agent_config` and `schema_migrations`. Both are system-internal tables with no user-scoped data, so RLS is enabled with **no policies**, meaning non-owner roles (Supabase `anon`, `authenticated`, PostgREST) see zero rows while the backend's `postgres` owner role bypasses RLS as before.

### 1. `agent_config` — global singleton locked down
**Migration:** `030_rls_system_tables`
**Why:** This table holds the fallback agent model, mode, and optional system prompt override. Without RLS, a leaked Supabase key or PostgREST access could read the configuration. Now returns zero rows to all non-owner roles.

### 2. `schema_migrations` — migration metadata locked down
**Migration:** `030_rls_system_tables`
**Why:** The `golang-migrate` bookkeeping table exposes which migration version the database is at. Low-sensitivity data, but no reason to leave it accessible. Now invisible to non-owner roles.

**Every table in the database now has RLS enabled.**

---

## 0.42.21 — TypeScript Build Fixes (2026-04-15)

Three type errors that broke the Vercel production build (`vue-tsc` exit code 2).

### 1. ConnectionForm empty-number field typed as `undefined`
**File:** `components/connections/ConnectionForm.vue`
**Fix:** `setFieldValue` stored `undefined` when a number field was cleared, but `getFieldValue` declared its return type as `string | number`. Changed the empty-field fallback from `undefined` to `''` so the stored value stays within the declared type.

### 2. NotificationPreferences threshold typed as bare `string`
**File:** `components/notifications/NotificationPreferences.vue`
**Fix:** `formThreshold` was `ref<string>`, but the `updateNotificationPreferences` payload expects `'info' | 'warning' | 'error' | 'critical'`. Narrowed the ref to the four-member union so the assignment on line 45 type-checks.

### 3. ActivityPage uses `.value` on ref inside template
**File:** `pages/ActivityPage.vue`
**Fix:** `activeFilters` is a `ref`, and Vue auto-unwraps refs in templates — so `activeFilters` in the template is already the plain object. Accessing `.value` tried to read a non-existent property on `{ severity?: string; connection_id?: string }`. Removed `.value` from both `@next` and `@prev` handlers.

---

## 0.42.20 — Post-Assessment Hardening (2026-04-15)

Eleven fixes from a v0.42.19 production readiness assessment covering four high-severity bugs (stale page state, 401 interceptor race, host injection), and seven medium correctness and robustness improvements across backend handlers, frontend state lifecycle, and routing.

### H1. OrgOverviewPage refetches apps on org switch
**File:** `pages/org/OrgOverviewPage.vue`
**Fix:** `fetchApps()` ran in `onMounted` only. When the user switched orgs via the dropdown while already on `/org`, Vue reused the component instance without re-mounting — the page showed the previous org's applications. Replaced `onMounted(fetchApps)` with `watch(() => appStore.organization?.id, () => fetchApps(), { immediate: true, flush: 'post' })` so data refetches whenever the org context changes.

### H2. OrgTeamPage refetches members on org switch
**File:** `pages/org/OrgTeamPage.vue`
**Fix:** Same root cause as H1 — `fetchMembers()` ran in `onMounted` only. Applied the same reactive watch pattern.

### H3. 401 interceptor no longer resets deduplication flag
**File:** `api/client.ts`
**Fix:** `isLoggingOut` was reset in the `finally` block before `window.location.href = '/login'` executed. Other in-flight 401 responses could re-enter the handler during the brief async gap, producing duplicate `auth.logout()` calls. Removed the `finally` reset entirely — the hard redirect destroys the JS context, making flag reset unnecessary.

### H4. Postgres connector validates and escapes `Host`
**File:** `connectors/database/postgres.go`
**Fix:** `cfg.Host` was interpolated directly into the connection string. A crafted value like `evil.com?default_transaction_read_only=off&host=` could inject query parameters, disabling the read-only enforcement. Now validated against a hostname/IP regex and URL-escaped with `url.PathEscape` before interpolation.

### M1. `Onboard` and `CreateNewOrganization` use `UserQueries`
**File:** `handlers/organizations.go`
**Fix:** Both handlers manually replicated the `UserQueries` pattern (begin tx, set RLS variable, commit/rollback) instead of calling the centralised helper. They missed the `context.WithoutCancel` that `UserQueries` uses for finalization — if a client disconnected mid-commit, the deferred rollback used a cancelled context. Refactored both to use `s.UserQueries()`.

### M2. `CreateConnection` auth check inside transaction
**File:** `handlers/connections.go`
**Fix:** `GetApplicationByOrgUser` ran against `s.Queries` (un-transacted), then the insert ran inside a `UserQueries` transaction — a TOCTOU gap where a user could be removed from the org between the auth check and the write. Moved the auth check inside the `UserQueries` transaction.

### M3. Chat handler uses `defer done()` via helper
**File:** `handlers/chat.go`
**Fix:** The conversation-setup transaction used manual `done()` calls in each early-return branch. Fragile — a future modification could forget `done()` in a new branch, leaking a transaction. Extracted the setup logic into `setupConversation()` where `defer done()` works naturally, returning the conversation ID and messages to the outer handler.

### M4. LoginPage distinguishes auth vs init failure
**File:** `pages/LoginPage.vue`
**Fix:** After successful `auth.login()`, if `appStore.init()` threw (network error, 500), the catch block showed "Authentication failed" — misleading since authentication succeeded. Separated error handling: auth errors show "Authentication failed", init errors show "Logged in but failed to load your workspace".

### M5. ProfileDropdown catches logout errors
**File:** `components/common/ProfileDropdown.vue`
**Fix:** `handleLogout` called `app.reset()` then `auth.logout()`. If `logout()` threw, app state was already cleared but the user stayed on the current page in a partially reset state. Wrapped in try/catch so navigation to `/login` proceeds regardless of logout outcome.

### M6. WebSocket `onclose` skips status flash on intentional close
**File:** `composables/useWebSocket.ts`
**Fix:** When `updateOptions` closed the old socket and opened a new one, the old socket's async `onclose` fired after the new socket was already connecting, briefly setting `status = 'closed'`. The `useAgent` watcher on status resets `isThinking` and `activeTools` on `'closed'`, causing a visual glitch during token refresh. Now guards the `onclose` handler — skips `status = 'closed'` when `intentionalClose` is true.

### M7. Org context detection uses `route.meta` instead of `startsWith`
**Files:** `router/index.ts`, `layouts/DefaultLayout.vue`, `components/common/AppHeader.vue`
**Fix:** `DefaultLayout` and `AppHeader` used `route.path.startsWith('/org')` to switch between org and app sidebars. Any future route starting with `/org` (e.g. `/organic`) would incorrectly trigger the org sidebar. Added `meta: { context: 'org' }` to all org routes and switched both components to `route.meta.context === 'org'`.

---

## 0.42.19 — Final Production Hardening (2026-04-15)

Six fixes from a comprehensive cross-subsystem production readiness assessment covering one high-severity security gap, one high-severity frontend state bug, and four medium correctness and robustness improvements.

### H1. Postgres connector validates SSLMode against allowlist
**File:** `connectors/database/postgres.go`
**Fix:** The `ssl_mode` field from connection config was interpolated directly into the connection URL without validation. A crafted value like `require&default_transaction_read_only=off` could inject additional query parameters, disabling the read-only enforcement that protects against LLM-generated write queries. Now validated against `{disable, require, verify-ca, verify-full}` before URL construction. The application-layer SQL validation (`tools_db.go`) remains the primary guard; this closes the defence-in-depth gap.

### H2. OrgSettingsPage seeds form on late-arriving org data
**File:** `pages/org/OrgSettingsPage.vue`
**Fix:** The settings form was populated in `onMounted`, but if the app store hadn't finished its async init yet, `org` was null and the name/slug fields stayed blank. Replaced `onMounted` with a reactive `watch(org, ..., { immediate: true })` so the form is seeded as soon as org data becomes available — whether that's before or after mount.

### M1. OrgSettingsPage sanitizes slug input
**File:** `pages/org/OrgSettingsPage.vue`
**Fix:** The slug edit field accepted arbitrary input (uppercase, spaces, special characters), relying entirely on backend validation. Added a `watch` on `editSlug` that applies the same sanitization pipeline used by `CreateOrgModal`: lowercase, strip non-alphanumeric characters (except hyphens), collapse consecutive hyphens, cap at 48 characters.

### M2. StatusBadge renders `paused` and `archived` statuses
**File:** `components/common/StatusBadge.vue`
**Fix:** The data model defines `status: 'active' | 'paused' | 'archived'`, but StatusBadge only styled `active`, `inactive`, `error`, and `warning`. `paused` and `archived` apps rendered with no color classes at all — no border tint, no dot color, no glow. Now maps `paused` to the warn/amber style and `archived` to the inactive/muted style.

### M3. Sole-owner guards run inside the transaction
**File:** `handlers/org_members.go`
**Fix:** `UpdateMemberRole` and `RemoveMember` both checked `CountOrgOwners` via `s.Queries` (outside the RLS-scoped transaction), then performed the write inside a separate `UserQueries()` transaction. Two concurrent requests to demote the last two owners could both pass the guard, then both commit, leaving zero owners. Moved the `GetOrgMembership` and `CountOrgOwners` reads inside the `UserQueries()` transaction so the guard and the write share the same transactional snapshot.

### M4. `app.init()` deduplicates concurrent calls
**File:** `stores/app.ts`
**Fix:** `init()` had no re-entry guard — overlapping calls from `App.vue` mount and the login handler could race, issuing parallel `listUserOrganizations` + `listApplications` calls with unpredictable interleaving. Added a shared `initPromise` that deduplicates concurrent callers: the first call runs, subsequent calls await the same promise, and the promise is cleared on completion so future calls work normally. Also fixed `reset()` to clear `loading`, `selectOrgSeq`, and `initPromise` for complete state teardown.

---

## 0.42.18 — Pre-Deploy Hardening (2026-04-15)

Six fixes from a comprehensive cross-subsystem review covering two high-severity concurrency bugs, two high-severity frontend state lifecycle issues, one medium commit-error handling gap, and one medium race condition in org switching.

### H1. GitHub connector `Health()` no longer reads token without mutex
**File:** `connectors/codebase/github.go`
**Fix:** `Health()` read `g.token` directly without holding `g.mu`, racing with `refreshTokenIfNeeded()` which updates the token under the lock. Now calls `refreshTokenIfNeeded(ctx)` and uses the returned value, matching the pattern already used by `apiGet()`. Eliminates a data race under concurrent agent tool-use calls.

### H2. `app.reset()` clears `initialized` flag
**File:** `stores/app.ts`
**Fix:** `reset()` cleared all reactive state and localStorage keys but left `initialized = true`. After a 401-triggered logout, the router guard saw `initialized === true` and skipped the init/onboarding flow — the app stayed on a stale page with empty state. Now sets `initialized = false` so the next navigation triggers a full `init()`.

### H3. Agent `Start()` inlines stop sequence to close lifecycle race
**File:** `agent/agent.go`
**Fix:** `Start()` previously unlocked `a.mu`, called `Stop()` (which re-acquires the lock), then re-locked — leaving a window where concurrent `Stop()` or `Start()` calls could observe inconsistent state (cancel set but wg count stale). Now inlines the stop logic: nils `a.cancel` under the lock before releasing it for `cancel()` + `wg.Wait()`, so any concurrent `Stop()` becomes a no-op.

### M1. 401 interceptor resets `isLoggingOut` after redirect
**File:** `api/client.ts`
**Fix:** The `isLoggingOut` deduplication flag was set on the first 401 but never reset. If the user logged back in within the same page session (no full reload), subsequent 401s were silently swallowed — the user got stuck in a zombie state where logout couldn't fire. Now resets in a `finally` block after the logout call completes.

### M2. `selectOrg()` discards stale responses on rapid org switch
**File:** `stores/app.ts`
**Fix:** Rapidly switching orgs (A → B → C) caused the async `listApplications()` response from org B to overwrite org C's state when it resolved late. Added a monotonic sequence counter (`selectOrgSeq`); the response is only applied if the counter still matches the value captured before the fetch.

### M3. `TestConnection` returns 500 on commit failure
**File:** `handlers/connections_test_handler.go`
**Fix:** When `commit()` failed after `UpdateConnectionStatus`, the error was logged but the handler continued to write a 200/502 HTTP response as if the status had been persisted. The UI showed "test passed" but the status wasn't saved. Now returns 500 and exits early if the commit fails.

---

## 0.42.17 — Production Readiness Fixes (2026-04-15)

Eight fixes from a comprehensive production-readiness review covering one critical RLS bypass, two high-severity bugs, and five medium/low consistency and correctness improvements.

### C1. `UpdateAppAgentConfig` now uses RLS-scoped transaction
**File:** `handlers/applications.go`
**Fix:** `UpdateAppAgentConfig` was the sole write handler bypassing `UserQueries` — it wrote via the unscoped `s.Queries` pool, skipping the `SET LOCAL app.current_user_id` RLS session variable. Now follows the same `UserQueries` + `commit()` + `defer done()` pattern as every other write handler.

### H1. `TestConnection` status update no longer silently rolled back
**File:** `handlers/connections_test_handler.go`
**Fix:** The `commit` return from `UserQueries` was discarded (`_`), so `done()` always rolled back the transaction. `UpdateConnectionStatus` writes were lost — connections showed stale status after testing. Now captures and calls `commit()` after the status write.

### H2. Post-delete navigation uses correct route name
**File:** `pages/org/OrgSettingsPage.vue`
**Fix:** After deleting an organisation, `router.push({ name: 'org' })` referenced a non-existent route (`'org'` vs the actual `'org-overview'`). The user was stranded on a dead page. Changed to `{ name: 'org-overview' }`.

### M1. `app.reset()` clears localStorage on logout
**File:** `stores/app.ts`
**Fix:** `reset()` cleared reactive state but left `heimdall_current_org` and `heimdall_current_app` in localStorage. A different user logging into the same browser would briefly attempt to load the previous user's org/app data. Now calls `localStorage.removeItem` for both keys.

### M2. SQL literal parser handles escaped single quotes (`''`)
**File:** `agent/tools_db.go`, `agent/tools_db_test.go`
**Fix:** PostgreSQL represents a literal `'` inside strings as `''` (two adjacent quotes). `stripAllLiterals` treated the second `'` as a closing delimiter, prematurely exiting the literal and potentially causing false rejections of valid queries. Now consumes `''` pairs and continues blanking inside the literal. Three new test cases added.

### M3. `CreateOrgModal` uses centralised `extractApiError()`
**File:** `components/org/CreateOrgModal.vue`
**Fix:** Replaced inline `as { response?: { data?: ... } }` type assertion with the `extractApiError()` utility used by every other page. Picks up the `response.data.message` fallback path the inline version missed.

### M4. 401 interceptor deduplicates logout on session expiry
**File:** `api/client.ts`
**Fix:** When a session expires, multiple in-flight requests each independently called `auth.logout()` and set `window.location.href`. Added an `isLoggingOut` guard so only the first 401 triggers the logout/redirect cycle.

### L1. Removed unused `formatRelativeTime` import
**File:** `components/org/AppCard.vue`
**Fix:** `formatRelativeTime` was imported but never used (only `formatDate` is referenced in the template). Removed the dead import.

---

## 0.42.16 — Post-Assessment Hardening (2026-04-15)

Eleven fixes from the v0.42 comprehensive code assessment, spanning one critical production-blocking bug, three high-severity issues, and seven medium correctness/consistency improvements.

### C1. MaxBodySize middleware no longer blocks webhook/OTLP ingestion
**Files:** `api/router.go`, `middleware/cors.go`
**Fix:** The global 1MB `MaxBodySize` middleware was wrapping request bodies *before* the webhook/OTLP handlers could apply their own 10MB limit. Moved `MaxBodySize` into the authenticated route group so ingestion endpoints are exempt. This was a production-blocking bug — large webhook payloads would silently fail with a generic "request body too large" error.

### H1. GitHub connector token access is now thread-safe
**File:** `connectors/codebase/github.go`
**Fix:** Added `sync.Mutex` around `token` and `tokenExpiresAt` fields. `refreshTokenIfNeeded` now returns the token under the lock, and `apiGet` uses the returned value instead of reading the field directly. Prevents a data race when concurrent agent tool-use calls trigger simultaneous token refreshes.

### H2. SQL read-only validation handles PostgreSQL dollar-quoting
**File:** `agent/tools_db.go`
**Fix:** `containsSemicolon` and `containsWriteKeyword` now strip `$$..$$` and `$tag$..$tag$` dollar-quoted literals alongside single-quoted strings. Also extended write-keyword scanning to SELECT/EXPLAIN statements (previously only applied to CTE/WITH). Defence-in-depth — the Postgres connector's `default_transaction_read_only=on` remains the primary guard.

### H3. ESLint config excludes `dist/` build artifacts
**File:** `frontend/eslint.config.js`
**Fix:** Added `{ ignores: ['dist/**'] }` to the flat config array. `npm run lint` previously scanned compiled build output, inflating the error count from 33 (source-only) to 2113.

### M1. `UserQueries` uses rollback-by-default transaction pattern
**File:** `handlers/userqueries.go` + 27 call sites across 12 handler files
**Fix:** Changed return signature from `(queries, done, err)` to `(queries, commit, done, err)`. `done()` now rolls back by default (safe to defer). Write handlers must call `commit()` explicitly on the success path. Read-only handlers use `_` for the commit return. This prevents partial commits when a handler returns early after a successful first write.

### M2. Notification write handlers use RLS-scoped transactions
**File:** `handlers/notifications.go`
**Fix:** `CreateNotificationChannel`, `UpdateNotificationChannel`, `DeleteNotificationChannel`, and `UpdateNotificationPreferences` now use `UserQueries` instead of the unscoped `s.Queries`. This ensures RLS policies are active for notification writes, matching the defence-in-depth pattern used by all other write handlers.

### M3. Scheduler marks failure on LLM provider error
**File:** `agent/scheduler.go`
**Fix:** When `providerFailed` is true, the scheduler now calls `markRunError` to advance `last_run_at`. Previously it returned silently without advancing the cursor, causing a tight retry loop (every 60s) that burned rate-limiter tokens while the provider was down.

### M4. MongoDB hostname resolution retries on transient failure
**File:** `connectors/logs/mongodb.go`
**Fix:** Replaced `sync.Once` caching with a mutex-guarded cache that only stores successful lookups. A transient DNS or network failure no longer permanently caches the error — the next `Poll` retries the lookup.

### M5. OrgTeamPage reverts role on failed API call
**File:** `pages/org/OrgTeamPage.vue`
**Fix:** `handleRoleChange` now saves `oldRole` before the optimistic update and reverts `member.role` in the catch block. Previously the UI displayed the new role even when the server rejected the change.

### M6. NotificationChannels requires confirmation before delete
**File:** `components/notifications/NotificationChannels.vue`
**Fix:** Replaced immediate delete-on-click with a confirmation modal (backdrop + Cancel/Delete buttons), matching the confirmation pattern used throughout the rest of the app for destructive actions.

### M7. Consistent `extractApiError` usage across org pages
**Files:** `pages/org/OrgTeamPage.vue`, `pages/org/OrgSettingsPage.vue`, `pages/org/OrgOverviewPage.vue`
**Fix:** Replaced 6 instances of inline `as { response?: { data?: ... } }` error type casting with the centralised `extractApiError()` utility. Eliminates duplicated fragile error-extraction logic.

---

## 0.42.15 — Medium-Severity Fixes (2026-04-15)

Twenty-five of twenty-eight medium issues from the v0.42.10 code assessment. M1 was already resolved by C2, M6 was a false positive (no change needed), and M25 (redundant index) is deferred. Covers correctness, consistency, hardening, and cleanup across both backend and frontend.

### M2. Dead `tools_memory.go` removed
**Fix:** Deleted placeholder file containing only a deferred-to-Phase-5 comment.

### M3. Malformed cron expressions emit to Activity feed
**Files:** `agent/scheduler.go`, `agent/scheduler_test.go`
**Fix:** `shouldFire` now returns `(bool, string)` — the string carries the parse error message. `schedulerTick` emits an agent_log entry so broken schedules are visible in the Activity feed, not just server logs.

### M4. `notifications.ts` refactored to `async/await`
**File:** `api/notifications.ts`
**Fix:** All 8 functions converted from `.then(r => r.data)` to `async/await` with explicit return types, matching every other API module.

### M5. GitHub Health() drains response body
**File:** `connectors/codebase/github.go`
**Fix:** Added `io.Copy(io.Discard, resp.Body)` before close to allow HTTP connection reuse.

### M7. Syslog Stream() contract documented
**File:** `connectors/logs/syslog.go`
**Fix:** Documented why the `out` channel parameter is unused (listener writes directly to DB). Renamed param to `_` for clarity.

### M8. ConnectionForm number field uses `undefined` sentinel
**File:** `components/connections/ConnectionForm.vue`
**Fix:** Replaced `'' as unknown as number` type assertion with `undefined` for empty numeric fields.

### M9. BaseSelect keyboard navigation validates bounds
**File:** `components/common/BaseSelect.vue`
**Fix:** `scrollToFocused()` now checks `focusedIndex` is within `[0, children.length)` before accessing DOM.

### M10. NotificationsPage template refs — already correct
**Assessment note:** Both child components already use `defineExpose` and the parent types refs via `InstanceType<typeof Component>`. No change needed.

### M11. Auth refresh distinguishes auth vs transient errors
**File:** `stores/auth.ts`
**Fix:** `init()` now checks whether a refresh failure is an auth error (401/403, revoked token) vs transient (network, 500). Transient errors keep the cached session; auth errors clear it.

### M12. API client uses synchronous toast import
**File:** `api/client.ts`
**Fix:** Replaced `import('@/composables/useToast').then(...)` with a top-level `import` so toasts display before redirect.

### M13. ScheduleModal resets custom cron on mode switch
**File:** `components/schedules/ScheduleModal.vue`
**Fix:** Added a `watch(mode)` that clears `customCron` when switching away from custom mode.

### M14. StepTest uses `extractApiError`
**File:** `components/connections/wizard/steps/StepTest.vue`
**Fix:** Replaced manual `err.response?.data?.message ?? err.message` chain with `extractApiError()`.

### M15. `UpdateOrganization` validates slug format
**File:** `handlers/organizations.go`
**Fix:** Applied `slugRe.MatchString(req.Slug)` validation, matching the existing pattern in `CreateNewOrganization` and `Onboard`.

### M16. Organization ref typed as `OrganizationWithRole`
**File:** `stores/app.ts`
**Fix:** Changed `ref<Organization | null>` to `ref<OrganizationWithRole | null>`. Fixed `onboard()` to spread `role: 'owner'` onto `organization.value`.

### M17. Org delete re-initializes instead of forcing onboarding
**File:** `pages/org/OrgSettingsPage.vue`
**Fix:** After delete, calls `appStore.init()` instead of `reset()`. Routes to org selector if other orgs exist, onboarding only if none remain.

### M18. Scheduler semaphore created once, not per tick
**File:** `agent/scheduler.go`
**Fix:** Semaphore hoisted from `schedulerTick` to `InvestigationScheduler`, passed as parameter. Matches the monitor pattern and prevents concurrent ticks from exceeding `maxConcurrentSchedules`.

### M19. `toolQueryDatabase` connection-per-query documented
**File:** `agent/tools_db.go`
**Fix:** Added TODO comment documenting the no-pooling limitation and the recommended fix (cache connectors per connection ID for loop duration).

### M20. GitHub installation token refreshes before expiry
**File:** `connectors/codebase/github.go`
**Fix:** Tracks `tokenExpiresAt` from the GitHub API response. `refreshTokenIfNeeded()` proactively refreshes when within 10 minutes of expiry. Called before every `apiGet`.

### M21. Auth state change resets app store
**File:** `stores/auth.ts`
**Fix:** `onAuthStateChange` callback now calls `appStore.reset()` when session transitions from authenticated to null.

### M22. Global `MaxBytesReader` middleware (1MB)
**Files:** `middleware/cors.go`, `api/router.go`
**Fix:** New `MaxBodySize` middleware applied globally. Caps request bodies at 1MB. The webhook endpoint's own `io.LimitReader(10MB)` takes precedence for its route.

### M23. CORS preflight caches for 1 hour
**File:** `middleware/cors.go`
**Fix:** Added `Access-Control-Max-Age: 3600` to OPTIONS responses.

### M24. Admin cannot attempt to remove owner in UI
**File:** `pages/org/OrgTeamPage.vue`
**Fix:** Added `&& member.role !== 'owner'` to remove button's `v-if` condition.

### M25. Redundant index — deferred
**Status:** Low write overhead, requires a dedicated migration. Deferred per assessment recommendation.

### M26. `InviteMember` catches unique constraint race
**File:** `handlers/org_members.go`
**Fix:** Catches duplicate key / unique constraint error from `CreateOrgMember` and returns 409 instead of 500.

### M27. LumberClassifier validates output count
**File:** `agent/classifier_lumber.go`
**Fix:** Added `len(events) != len(logs)` guard after `ClassifyBatch`. On mismatch, logs an error and escalates all logs (same fallback as classification failure).

### M28. MongoDB hostname uses `sync.Once`
**File:** `connectors/logs/mongodb.go`
**Fix:** Replaced manual double-checked locking with `sync.Once`. Eliminates duplicate API calls from concurrent `Poll` invocations.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `agent/tools_memory.go` | Deleted | Dead placeholder file |
| `agent/scheduler.go` | Fix | `shouldFire` returns error msg; semaphore hoisted |
| `agent/scheduler_test.go` | Update | Tests for new `shouldFire` signature |
| `agent/classifier_lumber.go` | Fix | Output length guard |
| `agent/tools_db.go` | Doc | Connection-per-query TODO |
| `api/notifications.ts` | Refactor | `async/await` pattern |
| `connectors/codebase/github.go` | Fix | Token refresh + body drain |
| `connectors/logs/syslog.go` | Doc | Stream() contract |
| `connectors/logs/mongodb.go` | Fix | `sync.Once` hostname cache |
| `ConnectionForm.vue` | Fix | `undefined` sentinel for empty numbers |
| `BaseSelect.vue` | Fix | Bounds check on keyboard nav |
| `stores/auth.ts` | Fix | Transient vs auth error; app store reset |
| `api/client.ts` | Fix | Synchronous toast import |
| `ScheduleModal.vue` | Fix | Reset custom cron on mode switch |
| `StepTest.vue` | Fix | `extractApiError` usage |
| `handlers/organizations.go` | Fix | Slug validation on update |
| `stores/app.ts` | Fix | `OrganizationWithRole` ref type + onboard role |
| `OrgSettingsPage.vue` | Fix | `init()` after org delete |
| `OrgTeamPage.vue` | Fix | Hide remove for owners |
| `handlers/org_members.go` | Fix | TOCTOU race → 409; limitation documented |
| `middleware/cors.go` | Fix | `Max-Age` header + `MaxBodySize` middleware |
| `api/router.go` | Fix | Apply `MaxBodySize` globally |

---

## 0.42.14 — High-Severity Fixes (2026-04-15)

All eleven high-severity issues from the v0.42.10 code assessment. Fixes onboarding defaults, standardizes connector error handling, adds panic recovery to background goroutines, guards against UI double-clicks, removes dead code, and documents deferred multi-member org scoping.

### H1. Onboard handler uses `agent.DefaultModelID` and sets Provider

**File:** `handlers/organizations.go`
**Issue:** The Onboard handler hardcoded `"claude-sonnet-4-6"` and omitted the `Provider` field, while `CreateApplication` correctly used `agent.DefaultModelID` and set `Provider: "anthropic"`. Apps created via onboarding had an empty Provider.
**Fix:** Replaced hardcoded model with `agent.DefaultModelID` and added `Provider: "anthropic"`.

### H2. `json.Marshal` errors handled in 4 connectors

**Files:** `connectors/logs/flyio.go`, `vercel.go`, `railway.go`, `mongodb.go`
**Issue:** All four used `payload, _ := json.Marshal(...)`, silently ignoring marshal errors. A nil payload would corrupt the DB entry.
**Fix:** Check the error and `continue` (skip the entry) on failure, matching the Supabase connector pattern.

### H3. Standardized poll error handling across connectors

**Files:** `connectors/logs/flyio.go`, `vercel.go`, `railway.go`, `mongodb.go`
**Issue:** These four connectors `return`ed immediately on the first insert error, aborting the entire polling batch. The Supabase connector used `continue` to skip bad rows.
**Fix:** Changed all four from `return fmt.Errorf(...)` to `continue` on insert error, matching Supabase. A single malformed entry no longer blocks the entire poll cycle.

### H4. `StopAll()` WaitGroup sufficiency documented

**File:** `connectors/listener.go`
**Issue:** `Stop()` waits on individual done channels; `StopAll()` uses only `wg.Wait()`. Subtle ordering inconsistency.
**Fix:** Added comment explaining why `wg.Wait()` is sufficient — goroutines call `wg.Done()` as their final action.

### H5. ConnectionWizard double-click guard on `goNext()`

**File:** `components/connections/wizard/ConnectionWizard.vue`
**Issue:** The "Continue" button was `:disabled="creating"`, but Vue batches DOM updates. Two click events in the same frame could both enter `goNext()` before the button was visually disabled.
**Fix:** Added `creating.value` check as an early return in `goNext()` — a JS-level guard that doesn't depend on DOM update timing.

### H6. NotificationChannels error handling on list refetch

**File:** `components/notifications/NotificationChannels.vue`
**Issue:** After a successful save, `listNotificationChannels()` was called to refresh the list. If that refetch failed, the UI showed stale data with no indication.
**Fix:** Wrapped the refetch in a nested try/catch. On failure, a toast alerts the user to reload.

### H7. Dead API exports removed

**Files:** `api/applications.ts`, `api/connections.ts`, `api/conversations.ts`
**Issue:** `getApplication()`, `getConnection()`, and `listConversations()` were exported but never imported anywhere.
**Fix:** Removed all three. Cleaned up the now-unused `ConversationSummary` import.

### H8. Migration 026 rollback guard for multi-org data

**File:** `migrations/026_org_members.down.sql`
**Issue:** The down migration silently discarded all org memberships except the earliest per user — permanent data loss for multi-org users.
**Fix:** Added a `DO $$ ... RAISE EXCEPTION ... $$` guard that aborts the rollback if any user has memberships in more than one org.

### H9. Multi-member org scoping limitation documented (deferred)

**File:** `handlers/org_members.go`
**Issue:** Leaf data tables (`connections`, `log_buffer`, `conversations`, `agent_log`, `investigations`) are user-scoped, not org-scoped. New team members see empty dashboards.
**Fix:** Documented the limitation in the `InviteMember` handler. Full org-scoping rewrite deferred to a future release — requires RLS policy changes, leaf query rewrites, and a migration.

### H10. Panic recovery in monitor/scheduler per-app goroutines

**Files:** `agent/monitor.go`, `agent/scheduler.go`
**Issue:** Per-app goroutines had no `recover()`. A panic would crash the entire server and hang `Stop()` indefinitely (via stuck `wg.Wait()`).
**Fix:** Added `defer func() { if r := recover(); r != nil { slog.Error(...) } }()` to both `monitorTick` and `schedulerTick` goroutines.

### H11. Agent log emitted when flagged logs exceed per-cycle cap

**File:** `agent/monitor.go`
**Issue:** When `len(flagged) > maxFlaggedForLLM`, excess logs were silently dropped. The cursor advanced past them, so they were never reassessed.
**Fix:** Emit an agent_log entry noting the drop count, total flagged, and cap value. The dropped logs still advance past the cursor (changing that risks infinite reprocessing loops), but the team now has Activity feed visibility into the cap being hit.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `handlers/organizations.go` | Fix | `agent.DefaultModelID` + `Provider: "anthropic"` |
| `connectors/logs/flyio.go` | Fix | Marshal error check + `continue` on insert error |
| `connectors/logs/vercel.go` | Fix | Marshal error check + `continue` on insert error |
| `connectors/logs/railway.go` | Fix | Marshal error check + `continue` on insert error |
| `connectors/logs/mongodb.go` | Fix | Marshal error check + `continue` on insert error |
| `connectors/listener.go` | Doc | Comment on `StopAll` WaitGroup sufficiency |
| `wizard/ConnectionWizard.vue` | Fix | `creating.value` early return in `goNext()` |
| `NotificationChannels.vue` | Fix | Refetch error handling with toast fallback |
| `api/applications.ts` | Cleanup | Removed dead `getApplication()` |
| `api/connections.ts` | Cleanup | Removed dead `getConnection()` |
| `api/conversations.ts` | Cleanup | Removed dead `listConversations()` + unused import |
| `migrations/026_org_members.down.sql` | Guard | `RAISE EXCEPTION` if multi-org memberships exist |
| `handlers/org_members.go` | Doc | Documented leaf-table user-scoping limitation |
| `agent/monitor.go` | Fix | Panic recovery + flagged-cap agent_log emit |
| `agent/scheduler.go` | Fix | Panic recovery in per-schedule goroutines |

---

## 0.42.13 — Critical Security & Correctness Fixes (2026-04-15)

All four critical issues from the v0.42.10 code assessment. Hardens SQL validation against CTE-based write injection, replaces a fragile provider-error string match with a typed return value, scopes ActivityPage connections to the current app, and enforces NOT NULL on `app_id` across leaf data tables.

### C1. `isReadOnlySQL` hardened against CTE writes and multi-statement injection

**File:** `agent/tools_db.go`
**Issue:** The read-only SQL validator only checked whether the query started with `SELECT`, `EXPLAIN`, or `WITH`. A CTE like `WITH x AS (DELETE FROM users RETURNING *) SELECT * FROM x` passed the check. Multi-statement injection via `;` (`SELECT 1; DROP TABLE users`) was also unguarded. The Postgres connector's `default_transaction_read_only=on` mitigated this at the DB layer, but the application-layer check was incomplete — a future non-Postgres connector would have no protection.
**Fix:** (1) Reject queries containing `;` outside of single-quoted string literals. (2) For `WITH`-prefixed queries, scan the normalized statement for write keywords (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`, `CREATE`, `GRANT`, `REVOKE`) as whole words, stripping string literals to avoid false positives on data values like `'DELETE ME'`. Added 26 test cases covering allowed queries, CTE attacks, multi-statement injection, comments, and edge cases.

### C2. Provider error detection uses typed return value instead of string match

**Files:** `agent/loop.go`, `agent/monitor.go`, `agent/scheduler.go`
**Issue:** `RunMonitoring` returned a hardcoded string `"Monitoring assessment failed: provider error"` when the LLM failed. The caller in `monitorApp` detected this via `strings.Contains(assessment, "provider error")`. If either string drifted, the monitoring cursor would advance past logs that were never assessed — permanent data loss.
**Fix:** `RunMonitoring` now returns a third value `providerFailed bool`. The monitor and scheduler check this boolean directly instead of pattern-matching on the assessment string. The string-based contract is eliminated entirely.

### C3. ActivityPage scoped to current app's connections

**File:** `pages/ActivityPage.vue`
**Issue:** `onMounted` called `connectionsStore.fetchConnections()` which hits `GET /connections` — returning all connections across all apps for the user. The connection filter dropdown showed connections from other apps. The app-switch watcher refetched logs but not connections, leaving stale cross-app data.
**Fix:** Changed to `connectionsStore.fetchConnectionsByApp(appId)` matching the pattern already used by `ConnectionsPage`. Added connection refetch to the `currentAppId` watcher so switching apps updates both logs and connections.

### C4. `app_id` NOT NULL constraint on `log_buffer` and `agent_log`

**Files:** `migrations/029_app_id_not_null.up.sql`, `migrations/029_app_id_not_null.down.sql`, sqlc-regenerated files, 10 callers updated
**Issue:** Migration 025 added `app_id` as nullable with a partial backfill. Rows that didn't match the backfill conditions retained `NULL app_id` permanently. These rows were invisible to the Activity feed (`app_id = $1` filters excluded them) but represented gaps in monitoring history and a schema integrity violation.
**Fix:** New migration 029: (1) deletes orphaned rows where `app_id IS NULL` (already invisible to queries), (2) adds `NOT NULL` constraint on both tables, (3) replaces partial indexes (`WHERE app_id IS NOT NULL`) with full indexes. sqlc regenerated to produce `uuid.UUID` instead of `pgtype.UUID` for these fields. Ten callers across connectors, handlers, and agent code updated from `pgtype.UUID{Bytes: x, Valid: true}` to plain `uuid.UUID`.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `agent/tools_db.go` | Rewrite | `isReadOnlySQL` + `containsSemicolon`, `containsWriteKeyword`, `stripStringLiterals` helpers |
| `agent/tools_db_test.go` | New | 26 test cases for SQL validation |
| `agent/loop.go` | Signature | `RunMonitoring` returns `(string, string, bool)` |
| `agent/monitor.go` | Fix | Uses `providerFailed` bool; removed `pgtype` import |
| `agent/scheduler.go` | Fix | Handles `providerFailed` from `RunMonitoring` |
| `agent/monitor_test.go` | Update | Three tests updated for new return signature |
| `agent/emit.go` | Fix | `pgtype.UUID` → `uuid.UUID` for `app_id` |
| `pages/ActivityPage.vue` | Fix | `fetchConnectionsByApp(appId)` + watcher refetch |
| `migrations/029_app_id_not_null.up.sql` | New | NOT NULL constraint + index rebuild |
| `migrations/029_app_id_not_null.down.sql` | New | Rollback to nullable + partial indexes |
| `handlers/logs.go` | Fix | `pgtype.UUID` → `uuid.UUID` + `hasAppID` bool |
| `handlers/applications.go` | Fix | `pgtype.UUID` → `uuid.UUID` for dashboard stats |
| `handlers/webhooks.go` | Fix | `pgtype.UUID` → `uuid.UUID` for log insert |
| `handlers/otlp.go` | Fix | `pgtype.UUID` → `uuid.UUID` for log insert |
| `connectors/logs/flyio.go` | Fix | `pgtype.UUID` → `uuid.UUID` |
| `connectors/logs/mongodb.go` | Fix | `pgtype.UUID` → `uuid.UUID` |
| `connectors/logs/railway.go` | Fix | `pgtype.UUID` → `uuid.UUID`; removed `pgtype` import |
| `connectors/logs/supabase.go` | Fix | `pgtype.UUID` → `uuid.UUID` |
| `connectors/logs/syslog.go` | Fix | `pgtype.UUID` → `uuid.UUID` |
| `connectors/logs/vercel.go` | Fix | `pgtype.UUID` → `uuid.UUID` |
| `db/*.sql.go` | Regenerated | sqlc output reflects NOT NULL `app_id` |

---

## 0.42.12 — Low-Severity Polish (2026-04-15)

Eleven of fifteen low-severity items from the v0.42.8 code assessment. Adds input validation, fixes a syslog parsing accuracy issue, removes dead types, and tightens the classifier escalation policy. Four items accepted as-is or deferred.

### L2. Email format validation on invite

**File:** `handlers/org_members.go`, `handlers/helpers.go`
**Issue:** `InviteMember` accepted any non-empty string as an email address.
**Fix:** Added `isValidEmail()` using Go's `net/mail.ParseAddress`. Rejects display names, checks for a dot in the domain, and caps length at 254 characters.

### L3. Notification email recipients validated and capped

**File:** `handlers/notifications.go`
**Issue:** Email notification channel config accepted any strings as recipients with no count limit. Typos or injection of non-email strings would only fail at send time.
**Fix:** Each recipient is validated with `isValidEmail()`. Maximum 20 recipients per channel.

### L5. TestConnection returns 502 on failure

**File:** `handlers/connections_test_handler.go`
**Issue:** `POST /connections/{id}/test` always returned HTTP 200 with `{"success": false}` on test failure. Clients relying on status codes couldn't distinguish success from failure.
**Fix:** Returns `502 Bad Gateway` when `result.Success == false`.

### L6. RFC 5424 regex now handles STRUCTURED-DATA

**File:** `connectors/logs/syslog.go`
**Issue:** The regex lumped the RFC 5424 STRUCTURED-DATA field into the message capture group. Structured data blocks like `[exampleSDID@32473 iut="3"]` were incorrectly appended to the message text.
**Fix:** Added an optional capture group for STRUCTURED-DATA (`-` or `[...]` blocks). The message is now captured in group 8. The group is optional to handle non-compliant syslog implementations that omit it.

### L8. WSMessage union includes tool_start/tool_result types

**File:** `types/agent.ts`
**Issue:** The `WSMessage` discriminated union was missing `WSToolStartMessage` and `WSToolResultMessage` interfaces. The `useAgent` composable handled these message types but TypeScript had no exhaustiveness checking.
**Fix:** Added both interfaces and included them in the union.

### L9. Dead `PaginatedResponse<T>` type removed

**File:** `types/api.ts`
**Issue:** `PaginatedResponse<T>` used `page`/`per_page` pagination but was never imported anywhere. The logs API uses its own `PaginatedLogs` type with `limit`/`offset`.
**Fix:** Removed the dead type.

### L10. Toast timers cleared on manual dismiss

**File:** `composables/useToast.ts`
**Issue:** Manually dismissing a toast left a dangling `setTimeout` callback that would fire after `duration` ms, calling `dismiss()` again on an already-removed toast.
**Fix:** Track timers in a `Map<number, Timeout>`. `dismiss()` calls `clearTimeout` before removing the toast.

### L12. `deleteNotificationChannel` returns `void`

**File:** `api/notifications.ts`
**Issue:** Returned the raw `AxiosResponse` instead of `void`, inconsistent with all other API functions that chain `.then(r => r.data)`.
**Fix:** Added `.then(() => {})` and explicit `Promise<void>` return type.

### L13. `SupabaseURL` no longer required in config validation

**File:** `config/config.go`
**Issue:** `Validate()` required `SUPABASE_URL` but the value was never consumed by the agent, connectors, or handlers in the reviewed code. Missing it would prevent the server from starting even when Supabase connectors weren't configured.
**Fix:** Removed from `Validate()`. The field is still loaded and available for the Supabase connector when configured.

### L14. Classifier no longer escalates on low confidence alone

**File:** `agent/classifier_lumber.go`
**Issue:** `Confidence < 0.5` alone triggered escalation regardless of the severity gate. A poorly calibrated model would escalate nearly everything, inflating noise metrics and LLM API costs.
**Fix:** Removed the confidence check. Escalation is now solely determined by `ShouldEscalate(event)` which checks type, category, and severity.

### L15. Consistent `public.` schema prefix in 026 rollback

**File:** `migrations/026_org_members.down.sql`
**Issue:** One policy JOIN referenced `users` without the `public.` schema prefix while others used `public.users`. On databases where `public` is not in `search_path`, the unqualified reference could fail.
**Fix:** Changed to `public.users` for consistency.

### Accepted / Deferred

- **L1** (duplicated `jsonError` in middleware vs handlers): Intentional — the packages cannot share a private function without creating a circular dependency or a new shared package. The 4-line function is acceptable duplication.
- **L4** (`persistMessages` swallows errors): Intentional fire-and-forget design for the WebSocket chat flow. Blocking on persistence would freeze the UI. Errors are logged server-side.
- **L7** (Railway unused `ServiceID`/`EnvironmentID` fields): Removing them would break existing connection configs stored in the database. Deferred until the Railway connector adds per-service filtering.
- **L11** (duplicate unique index on `organizations.slug`): Requires a migration to drop the redundant index. Marginal write overhead. Deferred.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `handlers/helpers.go` | New fn | `isValidEmail()` using `net/mail` |
| `handlers/org_members.go` | Fix | Email validation on invite |
| `handlers/notifications.go` | Fix | Email validation + 20-recipient cap |
| `handlers/connections_test_handler.go` | Fix | 502 status on test failure |
| `connectors/logs/syslog.go` | Fix | RFC 5424 STRUCTURED-DATA capture |
| `agent/classifier_lumber.go` | Fix | Removed low-confidence auto-escalation |
| `config/config.go` | Fix | SupabaseURL no longer required |
| `migrations/026_org_members.down.sql` | Fix | Consistent `public.` prefix |
| `types/agent.ts` | Fix | WSToolStart/Result types added |
| `types/api.ts` | Cleanup | Dead `PaginatedResponse` removed |
| `composables/useToast.ts` | Fix | Timer cleanup on dismiss |
| `api/notifications.ts` | Fix | `deleteNotificationChannel` returns void |

---

## 0.42.11 — Medium-Severity Fixes (2026-04-15)

Fifteen medium-severity issues from the v0.42.8 code assessment. Fixes silent status resets, inconsistent response shapes, missing pagination, credential leaks in error messages, and query performance gaps. Three items (M4, M7, M11) are deferred to future releases as they require schema migrations or policy decisions.

### M1. UpdateConnection preserves existing status

**File:** `handlers/connections.go`
**Issue:** When a client sent a `PUT /connections/{id}` without a `status` field (e.g. a rename-only update), the handler defaulted to `"inactive"`, silently deactivating a running connection.
**Fix:** Fetch the existing connection before applying defaults. When `status` is omitted, preserve the existing value.

### M2. Webhook ingest always returns an array

**File:** `handlers/webhooks.go`
**Issue:** `POST /api/webhooks/logs` returned a single object when one entry was inserted but an array when multiple were inserted. This polymorphic response broke clients that expected a consistent shape.
**Fix:** Always return an array.

### M3. Conversations endpoint supports pagination

**File:** `handlers/conversations.go`
**Issue:** `GET /api/conversations` hardcoded `LIMIT 50, OFFSET 0` with no query parameter support. Users with more than 50 conversations silently lost the rest.
**Fix:** Parse `?limit=` (1–200, default 50) and `?offset=` (default 0) query parameters.

### M5. Monitoring and stats queries use `lb.app_id` index

**Files:** `db/queries/monitoring.sql`, `db/queries/stats.sql`
**Issue:** `ListLogsSinceForApp` and `GetAppDashboardStats` joined through the `connections` table to filter by app, ignoring the `idx_log_buffer_app_id` partial index added in migration 025.
**Fix:** Rewrote both queries to filter directly on `log_buffer.app_id`, eliminating the join and enabling index use. Updated callers to pass `pgtype.UUID` for the nullable `app_id` column.

### M6. App-scoped log search queries added

**File:** `db/queries/log_buffer.sql`
**Issue:** `SearchLogsByUser` only scoped by `user_id`, making per-app Activity page search impossible.
**Fix:** Added `SearchLogsByApp` and `SearchLogsByAppAndSeverity` queries that filter on both `user_id` and `app_id`.

### M8. `selectOrg` no longer swallows network errors

**File:** `stores/app.ts`
**Issue:** The `catch` in `selectOrg` treated all errors — including network failures, 500s, and auth errors — as "org has no apps", showing an empty app list with no feedback.
**Fix:** Only catch 404 (no apps found). All other errors re-throw so the UI can display them.

### M9. WebSocket data wrapped to avoid same-value watch skip

**Files:** `composables/useWebSocket.ts`, `composables/useAgent.ts`
**Issue:** Vue's `watch` uses shallow equality. Two identical consecutive WebSocket messages (e.g. repeated `{"type":"status","content":"thinking"}`) would not trigger the watcher because the string value hadn't changed.
**Fix:** `data` ref now holds `{ payload: string, ts: number }` instead of a raw string. Each message gets a unique `ts`, guaranteeing the watch fires. `useAgent` reads `msg.payload` instead of the raw ref.

### M10. OrgTeamPage uses separate fetch and action error refs

**File:** `pages/org/OrgTeamPage.vue`
**Issue:** A single `error` ref was shared between fetch errors (loading the member list) and action errors (role changes, removals). Dismissing an action error could hide a still-relevant fetch error.
**Fix:** Split into `fetchError` (shown when member list is empty) and `actionError` (shown inline with dismiss button).

### M12. Connection test errors no longer leak internal details

**File:** `handlers/connections_test_handler.go`
**Issue:** Error messages included raw connector errors via `fmt.Sprintf("Failed to ...: %v", err)`. These could expose DSN strings, API keys, or internal hostnames to the browser.
**Fix:** Replaced all `%v` error formatting with static, user-friendly messages. Detailed errors are logged server-side only.

### M13. Slug format validation added

**File:** `handlers/organizations.go`
**Issue:** `CreateNewOrganization` and `Onboard` accepted arbitrary strings as org slugs — including spaces, unicode, control characters, and `/`.
**Fix:** Added `slugRe` regex (`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`) validating 3–50 characters, lowercase alphanumeric and hyphens, must start and end with a letter or digit.

### M14. Empty tool-results no longer appended to conversation

**File:** `agent/loop.go`
**Issue:** If all blocks in a tool-use response were non-`tool_use` types, an empty `ToolResults` message was appended. The provider would receive a malformed `user` turn with neither text nor tool results.
**Fix:** Guarded with `if len(toolResults) > 0` before appending, in both `RunConversation` and `RunMonitoring`.

### M15. `Vary: Origin` set unconditionally

**File:** `middleware/cors.go`
**Issue:** `Vary: Origin` was only set when the request origin matched the allowlist. Caching intermediaries could incorrectly serve a non-CORS response to a cross-origin request.
**Fix:** Moved `Vary: Origin` before the origin check so it's always present.

### M16. Onboarding route re-entry guard

**File:** `router/index.ts`
**Issue:** Authenticated users who had already completed onboarding could navigate directly to `/onboarding`, potentially creating duplicate organizations.
**Fix:** Added a router guard that redirects to `/dashboard` when `auth.isAuthenticated && !app.needsOnboarding && to.name === 'onboarding'`.

### M17. GitHub search returns empty when no repos connected

**File:** `connectors/codebase/github.go`
**Issue:** When no repos were connected and no specific repo was requested, `searchCode` built a query with no `repo:` filter, searching all of public GitHub — burning rate limit quota and returning irrelevant results.
**Fix:** Return an empty result with a helpful message when `repo == ""` and `len(g.repos) == 0`.

### M18. `UpdateGitHubRepos` capped at 100 entries

**File:** `handlers/github_repos.go`
**Issue:** The 1MB body limit still allowed thousands of repo entries, each triggering a separate DB upsert in a single transaction.
**Fix:** Added a `len(repos) > 100` check returning 400 before processing.

### Deferred

- **M4** (role-based auth on app writes): Requires a policy decision on whether `member` role should have write access to agent config, schedules, and notifications. Deferred to a feature release.
- **M7** (investigations/conversations org-scoping): Requires schema migration + data backfill to add `app_id` columns. Deferred to a feature release.
- **M11** (store mutation error handling consistency): Low functional impact since callers already wrap mutations in try/catch. Deferred.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `handlers/connections.go` | Fix | Fetch existing connection for status default |
| `handlers/webhooks.go` | Fix | Always return array |
| `handlers/conversations.go` | Fix | Pagination via `?limit=`/`?offset=` query params |
| `handlers/connections_test_handler.go` | Fix | Sanitized error messages (no `%v` leak) |
| `handlers/organizations.go` | Fix | Slug regex validation |
| `handlers/github_repos.go` | Fix | 100-repo cap on `UpdateGitHubRepos` |
| `handlers/applications.go` | Fix | `pgtype.UUID` for stats query param |
| `agent/loop.go` | Fix | Guard empty tool-results in both loops |
| `agent/monitor.go` | Fix | `pgtype.UUID` for `ListLogsSinceForApp` param |
| `connectors/codebase/github.go` | Fix | Empty-repo guard in `searchCode` |
| `middleware/cors.go` | Fix | `Vary: Origin` unconditional |
| `db/queries/monitoring.sql` | Perf | Direct `app_id` filter, no join |
| `db/queries/stats.sql` | Perf | Direct `app_id` filter, no join |
| `db/queries/log_buffer.sql` | New | `SearchLogsByApp` + `SearchLogsByAppAndSeverity` |
| `db/*sql.go` (generated) | Regen | sqlc regenerated from updated queries |
| `stores/app.ts` | Fix | `selectOrg` only catches 404 |
| `composables/useWebSocket.ts` | Fix | Data wrapped in `{ payload, ts }` |
| `composables/useAgent.ts` | Fix | Reads `msg.payload` from wrapped data |
| `pages/org/OrgTeamPage.vue` | Fix | Separate `fetchError` / `actionError` refs |
| `router/index.ts` | Fix | Onboarding re-entry guard |

---

## 0.42.10 — High-Severity Hardening (2026-04-14)

Thirteen high-severity issues from the v0.42.8 code assessment. Fixes concurrency races in the agent and connector subsystems, hardens the auth store and dashboard against stale state, and adds defense-in-depth to the application delete path. Three assessment items (H6 MongoDB auth, H8 cursor advancement, H15 GitHub callback ordering) were confirmed correct on deeper review and are documented below as resolved-no-change.

### H1. Agent Start/Stop race condition fixed

**File:** `agent/agent.go`
**Issue:** `a.cancel` was read and written from `Start()` and `Stop()` without synchronization. Concurrent calls could race on the cancel function.
**Fix:** Added `sync.Mutex` (`a.mu`) protecting `a.cancel`. `Start()` acquires the lock, checks for an existing instance, releases before calling `Stop()` (which also acquires the lock), then re-acquires to set the new cancel. `Stop()` copies the cancel func under the lock, nils it, releases, then calls cancel + wg.Wait outside the lock.

### H2. Poller Stop() now drains the goroutine

**File:** `connectors/poller.go`
**Issue:** `Stop()` cancelled the context and deleted the map entry but did not wait for the goroutine to exit. A subsequent `Start()` could race with the still-running old goroutine.
**Fix:** `Stop()` now waits on `<-entry.done` after cancelling, matching the drain pattern already used in `Start()`.

### H3. ListenerManager no longer holds mutex during Close()

**File:** `connectors/listener.go`
**Issue:** `Start()` held the mutex while calling `Close()` on an existing listener (which may block for up to 5 seconds). The unlock-relock pattern also had a TOCTOU race where another goroutine could insert a new entry between unlock and relock.
**Fix:** Rewrote to use the `done` channel pattern from the poller. The old entry is extracted under the lock, the lock is released, then `cancel()` + `Close()` + `<-done` runs without blocking other operations. `Stop()` and `StopAll()` follow the same pattern.

### H5. Syslog shutdown timeout documented as intentional

**File:** `connectors/logs/syslog.go`
**Issue:** The `go func() { s.wg.Wait(); close(done) }()` goroutine in `Close()` was leaked when the shutdown timeout fired.
**Fix:** Added a comment documenting that the leaked goroutine will finish naturally once connections hit their read deadline (`syslogReadTimeout`). Switched from `time.After` to `time.NewTimer` with proper cleanup via `defer timer.Stop()`.

### H6. MongoDB auth — confirmed correct (no change)

**File:** `connectors/logs/mongodb.go`
**Issue:** Assessment flagged `SetBasicAuth` as incorrect for Atlas v2 API. On deeper review, Atlas v2 programmatic API keys DO support HTTP Basic auth (public key as username, private key as password) when the `Accept: application/vnd.atlas.2023-01-01+json` header opts into the v2 contract. Digest auth is a legacy v1.0 requirement only.
**Fix:** Corrected the misleading comment. No functional change.

### H7. MongoDB hostname cached after first lookup

**File:** `connectors/logs/mongodb.go`
**Issue:** `getClusterHostname()` made an extra API call on every poll cycle to retrieve a hostname that doesn't change at runtime, doubling the request rate against the Atlas API.
**Fix:** Added `cachedHostname()` which resolves the hostname on first call and caches it in `m.hostname` (protected by `m.mu`) for subsequent calls.

### H8. Cursor advancement on insert failure — confirmed safe (no change)

**Files:** `logs/flyio.go`, `logs/vercel.go`, `logs/railway.go`, `logs/mongodb.go`
**Issue:** Assessment flagged cursor advancement past failed inserts. On deeper review, all four connectors `return fmt.Errorf(...)` immediately on insert failure, which exits `Poll()` before the cursor-advance block runs. For Fly.io specifically, the `pollMachineLogs` error causes the outer loop to `continue` (skipping the failed machine's `maxTS`), and successfully-inserted rows from other machines are correctly reflected in the cursor. No data loss occurs in any path.
**Fix:** No change needed. Documented the control flow.

### H9. Auth subscription leak fixed

**File:** `stores/auth.ts`
**Issue:** `supabase.auth.onAuthStateChange()` returned a subscription object that was never unsubscribed. On HMR or if `init()` was called twice, duplicate listeners would fire.
**Fix:** Store the subscription in `authSubscription`. On re-init, `unsubscribe()` the previous listener before registering a new one.

### H10. Auth refreshSession error now handled

**File:** `stores/auth.ts`
**Issue:** `refreshSession()` errors were silently discarded. A revoked refresh token would set `session.value = null` without any feedback — the user was silently logged out.
**Fix:** Destructure and check the `error` property from `refreshSession()`. On failure, log a warning and clear the session explicitly so the user is redirected to login.

### H11. OrgSettingsPage no longer mutates store directly

**Files:** `pages/org/OrgSettingsPage.vue`, `stores/app.ts`
**Issue:** `appStore.organization.name = updated.name` directly mutated the store's ref, bypassing Pinia's action tracking. The matching entry in `appStore.organizations[]` was not updated, so the `OrgDropdown` showed the old name until a full reload.
**Fix:** Added `updateOrg()` action to the app store that replaces both `organization` and the matching entry in `organizations[]` via spread (immutable update). `OrgSettingsPage` now calls `appStore.updateOrg()`.

### H12. ConnectionsPage bubble refs reset on fetch

**File:** `pages/ConnectionsPage.vue`
**Issue:** `bubbleEls` array never shrank when connections were deleted or the app switched. Stale DOM refs caused `FlowLines` to call `getBoundingClientRect()` on detached elements.
**Fix:** Reset `bubbleEls.value = []` at the start of `fetchAppConnections()` before the store fetch repopulates the list.

### H14. DashboardPage race condition on rapid app switch

**File:** `pages/DashboardPage.vue`
**Issue:** Multiple in-flight `Promise.allSettled` calls raced when the user switched apps rapidly. Whichever resolved last won, potentially overwriting the current app's data with stale data from a previous app.
**Fix:** Added a generation counter (`loadGeneration`). Each `loadData` call increments it and captures the current value. After `await Promise.allSettled`, the results are discarded if a newer call has started (`gen !== loadGeneration`).

### H15. GitHub callback — confirmed correct (no change)

**File:** `handlers/github_install.go`
**Issue:** Assessment flagged that the GitHub API call (`GetInstallation`) happened before the ownership check. On re-reading the code, `GetApplicationByOrgUser` (line 141) runs before `GetInstallation` (line 151). The ownership check is already in the correct order.
**Fix:** No change needed.

### H17. DeleteApplication now uses RLS-scoped transaction

**File:** `handlers/applications.go`
**Issue:** `CountApplicationsByOrg` and `DeleteApplication` used the bare `s.Queries` (pool, no RLS session variable), bypassing the defense-in-depth pattern used by other handlers.
**Fix:** Switched to `s.UserQueries()` for the entire authorization + delete path, matching the pattern used by all other write handlers.

### H18. Scheduler now runs schedules concurrently

**File:** `agent/scheduler.go`
**Issue:** `schedulerTick` ran schedules serially. A hung LLM call (up to 5-minute timeout) blocked all subsequent schedules, causing permanent schedule slip when many schedules fired on the same tick.
**Fix:** Schedules now run in goroutines with a semaphore cap (`maxConcurrentSchedules = 5`) and shutdown-aware acquisition, matching the monitor's concurrency pattern. A `sync.WaitGroup` ensures the tick waits for all in-flight schedules before returning.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `agent/agent.go` | Fix | Mutex on Start/Stop protecting cancel + wg |
| `agent/scheduler.go` | Fix | Concurrent schedule execution with semaphore |
| `connectors/poller.go` | Fix | Stop() waits on done channel before returning |
| `connectors/listener.go` | Fix | Done channel pattern, no mutex during Close() |
| `connectors/logs/syslog.go` | Fix | Documented timeout goroutine, use NewTimer |
| `connectors/logs/mongodb.go` | Fix | Cached hostname, corrected auth comment |
| `handlers/applications.go` | Fix | DeleteApplication uses UserQueries() |
| `stores/auth.ts` | Fix | Subscription cleanup + refresh error handling |
| `stores/app.ts` | Fix | New updateOrg() action for immutable updates |
| `pages/org/OrgSettingsPage.vue` | Fix | Uses appStore.updateOrg() |
| `pages/ConnectionsPage.vue` | Fix | Reset bubbleEls on fetch |
| `pages/DashboardPage.vue` | Fix | Generation counter prevents stale data |

---

## 0.42.9 — Critical Security & Correctness Fixes (2026-04-14)

Twelve critical issues identified in the v0.42.8 comprehensive code assessment. Fixes a broken RLS migration, adds SQL injection prevention to the agent's database tool, hardens the WebSocket chat layer, and resolves several frontend functional bugs.

### C1. Self-referential RLS policy on `org_members` fixed

**File:** `migrations/028_fix_org_members_rls.up.sql` (new)
**Issue:** Migration 027's RLS policy on `org_members` queried `org_members` in its own USING clause, causing PostgreSQL to throw `ERROR: infinite recursion detected`. This would crash every query against the table under RLS.
**Fix:** Replaced with a `SECURITY DEFINER` function (`app_user_org_ids()`) that bypasses RLS when resolving the caller's org memberships, then the policy references that function.

### C3. SQL statement validation added to `query_database` tool

**File:** `agent/tools_db.go`
**Issue:** LLM-generated SQL was passed directly to the user's connected Postgres database with no statement-type validation. Despite `default_transaction_read_only=on`, a `SET` statement could disable the guard.
**Fix:** Added `isReadOnlySQL()` which strips leading comments, normalizes to uppercase, and only allows statements beginning with `SELECT`, `EXPLAIN`, or `WITH`. All other statement types (UPDATE, DELETE, DROP, SET, etc.) are rejected before reaching the database.

### C4. Row-count limit added to database connector

**File:** `connectors/database/postgres.go`
**Issue:** `Query()` fetched all rows with no limit. An LLM-generated `SELECT * FROM large_table` could OOM the agent process.
**Fix:** Results are now capped at 1,000 rows. When truncated, a `_truncated` sentinel row is appended explaining the limit and suggesting a `LIMIT` clause.

### C5. Sole-owner guard added to `UpdateMemberRole`

**File:** `handlers/org_members.go`
**Issue:** An owner could demote themselves to `member` or `admin` via `PUT /api/org/members/{userId}/role`, leaving the org with zero owners and no way to recover. The guard existed in `RemoveMember` but not here.
**Fix:** Added the same `CountOrgOwners` check: if the target is the current sole owner and the new role is not `owner`, the request returns `409 Conflict`.

### C6. Activity page pagination filters fixed

**File:** `pages/ActivityPage.vue:65-66`
**Issue:** `@next` and `@prev` event handlers passed `activeFilters` (a `Ref` wrapper) instead of `activeFilters.value`. The logs store received the raw ref object, causing filters to be silently ignored on page navigation.
**Fix:** Changed to `activeFilters.value`.

### C7. Login flow now initializes app store

**File:** `pages/LoginPage.vue`
**Issue:** After `auth.login()`, the code called `router.push('/dashboard')` without calling `appStore.init()`. The app store remained uninitialized — `currentAppId` was `null`, `applications` was empty, and all dashboard API calls passed `undefined` as the app ID.
**Fix:** Added `await appStore.init()` after login. If the user needs onboarding, redirects to `/onboarding` instead of `/dashboard`.

### C8. WebSocket token and appId are now reactive

**File:** `composables/useAgent.ts`
**Issue:** `auth.token` and `appStore.currentAppId` were captured once at setup time. If the Supabase session refreshed or the user switched apps, the WebSocket continued using stale values.
**Fix:** Added `watch()` on both `auth.token` and `appStore.currentAppId` that call `updateOptions()` on the WebSocket, triggering a reconnect with fresh credentials when they change.

### C9. WebSocket reconnection with exponential backoff

**File:** `composables/useWebSocket.ts`
**Issue:** Any network interruption, server restart, or idle timeout permanently killed the WebSocket. The chat page showed a dead input with no recovery path short of a full page reload.
**Fix:** Added automatic reconnection with exponential backoff (1s base, 30s cap). Intentional closes (component unmount, explicit `close()`) do not trigger reconnection. The `updateOptions()` method allows the parent composable to update connection parameters and trigger a reconnect.

### C9b. Thinking/tool state reset on WebSocket close

**File:** `composables/useAgent.ts`
**Issue:** If the WebSocket disconnected while the agent was mid-thought, `isThinking` and `activeTools` remained set indefinitely — the UI showed a permanent spinner.
**Fix:** Added a `watch(status)` that resets `isThinking = false` and `activeTools = []` when status becomes `'closed'`.

### C10. Monitor cursor no longer advances on LLM provider failure

**File:** `agent/monitor.go`
**Issue:** When `RunMonitoring` returned a provider error, the cursor was still advanced past the unprocessed logs. Those logs were permanently skipped from re-analysis on the next tick.
**Fix:** Added an early return (before cursor advancement) when the assessment indicates a provider failure. Logs will be reprocessed on the next monitoring tick.

### C11. CORS middleware now allows PATCH and X-Org-ID

**File:** `middleware/cors.go`
**Issue:** `Access-Control-Allow-Methods` was missing `PATCH`. The `PATCH /api/apps/{appId}/schedules/{id}` endpoint would fail CORS preflight in browsers. Additionally, the `X-Org-ID` header (used for multi-org context) was not in `Access-Control-Allow-Headers`.
**Fix:** Added `PATCH` to allowed methods and `X-Org-ID` to allowed headers.

### C12. 404 page "Return to Dashboard" link fixed

**File:** `pages/NotFoundPage.vue`
**Issue:** The link pointed to `/` (public landing page) instead of `/dashboard`.
**Fix:** Changed to `/dashboard`.

### Test fix: `parseSeverityFromResponse` test cases updated

**File:** `agent/monitor_test.go`
**Issue:** Three test cases still expected the old heuristic keyword fallback removed in v0.42.8 (M6). Tests for bare "error"/"critical"/"warning" keywords (without `severity:` markers) expected those keywords to be parsed as severities, but the parser now correctly defaults to `"info"` for unstructured text.
**Fix:** Updated expected values to `"info"` to match the v0.42.8 parser behavior.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `migrations/028_fix_org_members_rls.up.sql` | New | SECURITY DEFINER function + fixed RLS policy |
| `migrations/028_fix_org_members_rls.down.sql` | New | Reverts to original policy |
| `agent/tools_db.go` | Fix | SQL statement validation (`isReadOnlySQL`) |
| `connectors/database/postgres.go` | Fix | 1,000-row query limit |
| `handlers/org_members.go` | Fix | Sole-owner demotion guard |
| `middleware/cors.go` | Fix | PATCH method + X-Org-ID header |
| `agent/monitor.go` | Fix | Skip cursor advance on provider failure |
| `agent/monitor_test.go` | Fix | Updated severity parser test expectations |
| `pages/ActivityPage.vue` | Fix | Pass `.value` not ref to pagination |
| `pages/LoginPage.vue` | Fix | Call `appStore.init()` after login |
| `pages/NotFoundPage.vue` | Fix | Dashboard link points to `/dashboard` |
| `composables/useWebSocket.ts` | Fix | Reconnection with exponential backoff |
| `composables/useAgent.ts` | Fix | Reactive token/appId + close state reset |

---

## 0.42.8 — Concurrency, Correctness & Test Coverage (2026-04-14)

Seven "should-fix" medium-severity items from the v0.42.5 code assessment. Fixes three concurrency bugs in the backend, removes a severity-parsing false-positive source, parallelizes dashboard loading, and adds integration tests for the org member management endpoints.

### M1. Test router sync — already resolved

**Resolved in:** v0.42.6 (H7)
The test router in `testhelpers_test.go` was fully synced with `router.go` during the H7 fix. All `/api` routes now match. Non-API routes (`/health`, `/metrics`, `/ws/chat`) are intentionally omitted as they're not handler-level concerns.

### M2. Integration tests for org_members handlers

**File:** `handlers/org_members_test.go` (new)
**Issue:** Four security-sensitive handler methods (`ListOrgMembers`, `InviteMember`, `UpdateMemberRole`, `RemoveMember`) had zero test coverage.
**Fix:** Added 9 integration tests covering:
- List members (verifies owner from testSetup)
- Invite member (success, duplicate 409, missing email 400)
- Authorization: member-role users cannot invite (403)
- Update role (owner promotes member to admin, non-owner blocked)
- Remove member (success, sole-owner blocked with 409, member-role cannot remove)

The `envForUser` helper in `organizations_test.go` was expanded to mount the `/org/members` route group.

### M5. Monitor semaphore no longer blocks shutdown

**File:** `agent/monitor.go:68`
**Issue:** `sem <- struct{}{}` blocked the main goroutine unconditionally. If all 10 semaphore slots were occupied, the monitor goroutine could not respond to context cancellation signals during shutdown.
**Fix:** Replaced the bare send with a `select` on both `sem` and `ctx.Done()`. When shutdown fires, the loop returns immediately instead of waiting for a semaphore slot.

### M6. Severity heuristic false-positives removed

**File:** `agent/loop.go:346-361`
**Issue:** `parseSeverityFromResponse` fell back to a keyword scan checking if the entire response contained "error" or "warning". Benign phrases like "No errors detected" or "Warning acknowledged" would incorrectly classify the response as severity `error` or `warning`.
**Fix:** Removed the heuristic fallback entirely. The parser now only matches structured `severity: <level>` markers and defaults to `info` otherwise. The structured format is what the monitoring prompt asks Claude to emit, so the heuristic was redundant.

### M9. Poller `Start` no longer allows concurrent polls

**File:** `connectors/poller.go`
**Issue:** Calling `Start` for a connection that was already polling cancelled the old context but immediately started the new goroutine without waiting for the old one to exit. Both goroutines could poll the same connection concurrently, producing duplicate log entries.
**Fix:** Introduced a `pollerEntry` struct with a `done` channel (closed when the goroutine exits). `Start` now waits on `<-entry.done` after cancelling the old context, ensuring the previous goroutine has fully exited before launching the replacement.

### M10. ListenerManager no longer holds mutex during `Close()`

**File:** `connectors/listener.go:50-53`
**Issue:** `Start` called `entry.listener.Close()` (which may block for up to a 5-second drain timeout) while holding `m.mu`, blocking all other listener operations.
**Fix:** The lock is now released before calling `cancel()` and `Close()` on the old listener, then re-acquired for the new entry insertion. This matches the existing pattern in the `Stop()` method.

### M14. DashboardPage API calls parallelized

**File:** `pages/DashboardPage.vue`
**Issue:** Five independent API calls (`getAppAgentConfig`, `listConnectionsByApp`, `fetchLogs`, `getAppStats`, `getMonitoringStatus`) were made sequentially, creating a visible waterfall delay on page load.
**Fix:** All five calls now run concurrently via `Promise.allSettled()`. Each result is inspected individually — fulfilled values are assigned, rejected calls are collected into the error banner. Same error-handling behavior, substantially faster page load.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `agent/monitor.go` | Fix | Semaphore acquire selects on `ctx.Done()` for clean shutdown |
| `agent/loop.go` | Fix | Removed heuristic keyword fallback from `parseSeverityFromResponse` |
| `connectors/poller.go` | Fix | `Start` waits for old goroutine to exit via `done` channel before launching new one |
| `connectors/listener.go` | Fix | `Start` releases mutex before calling `Close()` on existing listener |
| `pages/DashboardPage.vue` | Fix | Sequential API calls replaced with `Promise.allSettled()` |
| `handlers/org_members_test.go` | New | 9 integration tests for member CRUD + authorization |
| `handlers/organizations_test.go` | Fix | `envForUser` helper expanded with org member routes |

---

## 0.42.7 — Medium-Severity Fixes (2026-04-14)

Four medium-severity issues from the v0.42.5 code assessment resolved. These were the "must-fix for 8.5+" items from the assessment's recommended priority list.

### M4. Prometheus `/metrics` endpoint moved behind auth

**File:** `router.go`
**Issue:** `/metrics` was mounted outside the protected route group, exposing operational data (goroutine counts, request latencies, error rates) to unauthenticated callers.
**Fix:** Moved the `/metrics` handler into its own route group with the `Auth(jwks)` middleware applied. Prometheus scrapers now need a valid JWT.

### M8. ILIKE wildcard escaping — already resolved

**File:** `agent/tools_logs.go:15-20`
**Issue:** The assessment flagged `SearchLogsByUser` ILIKE queries as vulnerable to wildcard injection (`%`, `_`). On review, the only caller (`tools_logs.go`) already escapes via `escapeLike()` which replaces `\` → `\\`, `%` → `\%`, `_` → `\_`, and the SQL query uses `ESCAPE '\'`. The HTTP `ListLogs` handler does not use ILIKE at all. No fix required.

### M12. Dead navigation links removed

**Files:** `PublicNav.vue`, `PublicFooter.vue`
**Issue:** Links to `/security`, `/terms`, and `/privacy` had no corresponding routes — users clicking these landed on the 404 page.
**Fix:** Removed the three dead links. `/security` was in both desktop and mobile nav sections of `PublicNav`. `/terms`, `/privacy`, and `/security` were in `PublicFooter`. Links to existing routes (`/features`, `/pricing`) are retained.

### M16. App store no longer treats all errors as "needs onboarding"

**File:** `frontend/src/stores/app.ts`
**Issue:** The catch block in `init()` set `needsOnboarding = true` for any error — including network failures, 500s, or auth issues. A transient server error would incorrectly route users to the onboarding flow.
**Fix:** The catch now inspects the Axios error's response status. Only a 404 (no org found) triggers onboarding. All other errors re-throw, propagating to `App.vue`'s top-level catch which redirects to `/login` so the user can retry.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `router.go` | Fix | `/metrics` moved behind JWT auth middleware |
| `frontend/src/stores/app.ts` | Fix | `init()` catch differentiates 404 from other errors |
| `frontend/src/components/public/PublicNav.vue` | Fix | Removed dead `/security` link (desktop + mobile) |
| `frontend/src/components/public/PublicFooter.vue` | Fix | Removed dead `/terms`, `/privacy`, `/security` links |

---

## 0.42.6 — Post-Assessment Hardening (2026-04-14)

Third code assessment (v0.42.5) identified 7 high-severity issues across backend and frontend. All seven are resolved in this release.

### H1. Multi-org handlers now route through `UserQueries()` for RLS

**Files:** `handlers/applications.go`, `handlers/org_members.go`, `handlers/organizations.go`
**Issue:** Several handlers introduced during the multi-org phases (`CreateApplication`, `InviteMember`, `UpdateMemberRole`, `RemoveMember`, `CreateNewOrganization`, `Onboard`) used `s.Queries` directly instead of `UserQueries()`. This bypassed the `SET LOCAL app.current_user_id` session variable that RLS policies depend on for defense-in-depth enforcement. Handler-level authorization (`resolveOrgAndRole`) protected the endpoints functionally, but the DB layer had no independent access control.
**Fix:** All write operations now go through `UserQueries()`. For `CreateNewOrganization` and `Onboard` (which manage their own transactions), `set_config('app.current_user_id', ...)` is now called explicitly within the transaction.

### H2. WebSocket origin restricted to CORS allowlist

**File:** `handlers/chat.go`
**Issue:** `OriginPatterns: []string{"*"}` accepted WebSocket connections from any origin. Combined with the JWT token in the URL query string, a malicious page could potentially establish a WebSocket connection if the token were leaked via referrer or browser history.
**Fix:** New `wsOriginPatterns()` helper reads `CORS_ALLOWED_ORIGINS` (same env var as the CORS middleware) and falls back to localhost dev origins. The wildcard is eliminated.

### H3. Non-functional reports stack removed

**Files:** Removed: `handlers/reports.go`, `ReportsPage.vue`, `stores/reports.ts`, `api/reports.ts`, `types/report.ts`, `components/reports/ReportCard.vue`, `components/reports/ReportList.vue`, `components/reports/ReportDetail.vue`. Modified: `router.go`, `testhelpers_test.go`, `router/index.ts`, `AppSidebar.vue`.
**Issue:** `ListReports` and `GetReport` were stubs returning empty arrays unconditionally. The entire reports feature (backend handlers, frontend store, API client, page, 3 components, type definition, sidebar link, route) was wired end-to-end but served no data — the underlying `reports` DB table and investigation-to-report pipeline don't exist yet. `GetReport` also returned an array instead of an object, which is semantically wrong for a single-resource GET.
**Fix:** Removed the entire dead stack. Marketing references to "reports" (PricingPage, FeaturesPage) are retained as they describe future capability. The feature will be rebuilt from scratch when the data model is implemented.

### H4. 401 interceptor logout wrapped in try/catch

**File:** `frontend/src/api/client.ts`
**Issue:** The 401 response interceptor did `await auth.logout()` without error handling. If Supabase `signOut` failed (expired session, network error), the `window.location.href = '/login'` redirect never executed, leaving the user stuck on a broken authenticated page.
**Fix:** Wrapped `auth.logout()` in try/catch so the redirect to `/login` fires unconditionally.

### H5. WebSocket `send()` guarded on `readyState`

**File:** `frontend/src/composables/useWebSocket.ts`
**Issue:** `send()` checked for `ws` being null but not for `readyState !== OPEN`. Calling `send()` on a socket in `CONNECTING` state throws `DOMException`. While the chat page disables the send button when status isn't `open`, any other consumer of `useWebSocket` would crash.
**Fix:** Added `ws.readyState === WebSocket.OPEN` guard before calling `ws.send()`.

### H6. Wizard step `payloadExample` moved into `<script setup>`

**Files:** `StepWebhookSetup.vue`, `StepOTLPSetup.vue`
**Issue:** Both wizard step components defined `payloadExample` in a second `<script lang="ts">` block (non-setup). Variables from non-setup script blocks are not reliably available in the `<script setup>` template scope — the constant rendered as `undefined` in some Vue compiler versions.
**Fix:** Moved the constant definitions into the `<script setup>` block and removed the orphaned `<script>` blocks.

### H7. Test router synced with production router

**File:** `handlers/testhelpers_test.go`
**Issue:** The test router was missing 14 routes that exist in the production router: OTLP ingestion (`POST /v1/logs`), GitHub install + callback, all 8 notification endpoints, GitHub repo management, and the models endpoint. Any future tests targeting these routes would silently get 404/405 from the test router instead of exercising the handlers.
**Fix:** Added all missing routes to match `router.go` exactly (minus auth middleware, which the test harness replaces with a user-ID injector).

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `handlers/applications.go` | Fix | `CreateApplication` routes through `UserQueries()` |
| `handlers/org_members.go` | Fix | `InviteMember`, `UpdateMemberRole`, `RemoveMember` route through `UserQueries()` |
| `handlers/organizations.go` | Fix | `CreateNewOrganization` and `Onboard` set RLS session variable in their transactions |
| `handlers/chat.go` | Fix | WebSocket origin restricted; `wsOriginPatterns()` helper added |
| `handlers/reports.go` | Remove | Non-functional stub handlers deleted |
| `handlers/testhelpers_test.go` | Fix | Test router synced: added 14 missing routes, removed reports routes |
| `router.go` | Fix | Reports routes removed |
| `frontend/src/api/client.ts` | Fix | 401 interceptor logout wrapped in try/catch |
| `frontend/src/composables/useWebSocket.ts` | Fix | `send()` guards on `readyState === OPEN` |
| `frontend/src/components/connections/wizard/steps/StepWebhookSetup.vue` | Fix | `payloadExample` moved into `<script setup>` |
| `frontend/src/components/connections/wizard/steps/StepOTLPSetup.vue` | Fix | `payloadExample` moved into `<script setup>` |
| `frontend/src/router/index.ts` | Fix | Reports route removed |
| `frontend/src/components/common/AppSidebar.vue` | Fix | Reports sidebar link removed |
| `frontend/src/pages/ReportsPage.vue` | Remove | Dead page |
| `frontend/src/stores/reports.ts` | Remove | Dead store |
| `frontend/src/api/reports.ts` | Remove | Dead API module |
| `frontend/src/types/report.ts` | Remove | Dead type definition |
| `frontend/src/components/reports/*` | Remove | Dead components (ReportCard, ReportList, ReportDetail) |

---

## 0.42.5 — Critical & High-Severity Fixes (2026-04-14)

Second code assessment (v0.42.4) uncovered 2 critical and 6 high-severity issues. All eight are resolved in this release. The multi-org feature is now fully functional for users with multiple organizations.

### C1. OTLP ingestion endpoint fixed

**File:** `backend/internal/db/queries/log_buffer.sql:36-38`
**Issue:** `GetConnectionByWebhookToken` hard-filtered on `type = 'webhook_logs'`, so OTLP connections (`type = 'otlp'`) could never match. Every OTLP ingest request returned 401 "invalid token" regardless of the token's validity — the entire OTLP path was non-functional.
**Fix:** Changed the SQL filter to `type IN ('webhook_logs', 'otlp')`.

### C2. Multi-org operations now respect the active org context

**Files:** `backend/internal/api/handlers/org_members.go`, `organizations.go`, `applications.go`, `frontend/src/api/client.ts`
**Issue:** All org-scoped backend operations (`resolveOrgAndRole`, `GetOrganization`, `ListApplications`, `CreateApplication`) called `GetOrganizationByUser`, which always returns the user's **earliest** (primary) org. Users with multiple orgs could not manage, view, or modify any org other than their primary — the frontend multi-org switcher was cosmetic only.
**Fix:** Introduced an `X-Org-ID` header contract. The frontend Axios interceptor reads the active org ID from `localStorage` (already persisted by the app store's `selectOrg()`) and attaches it to every request. The backend's new `resolveOrgForUser` helper reads this header, validates the caller's membership via `GetOrgMembership`, and returns the target org. Falls back to the primary org when the header is absent (backward-compatible).

### H1. WebSocket `app_id` now authorization-checked

**File:** `backend/internal/api/handlers/chat.go:57-72`
**Issue:** The `app_id` query parameter on the WebSocket endpoint was accepted without any ownership check. An authenticated user who knew another org's app UUID could scope the agent's `search_codebase` tool to that org's GitHub repos — a cross-tenant data access vulnerability.
**Fix:** Added a `GetApplicationByOrgUser` check before accepting the `app_id`. Returns a WebSocket error if the app doesn't belong to the caller's org.

### H2. Sole-owner removal guard now checks owner count

**Files:** `backend/internal/db/queries/organizations.sql`, `backend/internal/api/handlers/org_members.go:290-299`
**Issue:** The guard preventing removal of the last owner used `CountOrgMembers` (all roles). An org with 5 members but only 1 owner would pass the check, allowing the sole owner to remove themselves and leaving the org permanently unmanageable.
**Fix:** Added a `CountOrgOwners` SQL query (`WHERE role = 'owner'`). The guard now checks whether the target is an owner and blocks removal when the owner count would drop to zero — regardless of who initiates the removal.

### H3. `org_members` table now has RLS enabled

**File:** `backend/migrations/027_org_members_rls.{up,down}.sql`
**Issue:** Migration 026 created the `org_members` table but never enabled Row Level Security. Handler-level checks (`resolveOrgAndRole`) provided application-level protection, but any query through `UserQueries()` could read memberships for any org without DB-level enforcement.
**Fix:** New migration 027 enables RLS and creates a policy that restricts access to orgs the caller belongs to, matching the pattern used on all other tenant-scoped tables.

### H4. `InviteMember` Content-Type header now set correctly

**File:** `backend/internal/api/handlers/org_members.go:146-147`
**Issue:** `WriteHeader(201)` was called before `Header().Set("Content-Type", "application/json")`. In Go's `net/http`, `WriteHeader` flushes headers — subsequent `Set` calls are silently ignored. The response body was JSON but the Content-Type defaulted to `text/plain`.
**Fix:** Swapped the two lines — header set before status write.

### H5. UTF-8 safe conversation title truncation

**File:** `backend/internal/api/handlers/chat.go:152-155`
**Issue:** `title[:50]` sliced by byte index. On multi-byte UTF-8 content (CJK characters, emoji), this could split a rune mid-sequence, producing invalid UTF-8 or a runtime panic.
**Fix:** Converted to `[]rune` before truncation: `string(titleRunes[:50]) + "..."`.

### H6. Org mutations now route through `UserQueries()` for RLS

**File:** `backend/internal/api/handlers/organizations.go:162-172,205-215`
**Issue:** `UpdateOrganization` and `DeleteOrganization` called `s.Queries` directly (shared pool, no RLS session variable). The handler-level `resolveOrgAndRole` check was the only protection — no defense-in-depth at the DB layer.
**Fix:** Both operations now use `s.UserQueries(ctx, userID)` to acquire a transaction with `app.current_user_id` set, matching the pattern used by all other write operations.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `backend/internal/db/queries/log_buffer.sql` | Fix | OTLP token lookup accepts `'otlp'` type |
| `backend/internal/db/queries/organizations.sql` | Add | `CountOrgOwners` query |
| `backend/internal/db/organizations.sql.go` | Regen | sqlc regenerated |
| `backend/internal/db/log_buffer.sql.go` | Regen | sqlc regenerated |
| `backend/internal/api/handlers/org_members.go` | Fix | `resolveOrgForUser` with `X-Org-ID` header; owner-count guard; Content-Type order |
| `backend/internal/api/handlers/organizations.go` | Fix | `GetOrganization` uses `resolveOrgAndRole`; mutations via `UserQueries()` |
| `backend/internal/api/handlers/applications.go` | Fix | `ListApplications`/`CreateApplication` use `resolveOrgAndRole` |
| `backend/internal/api/handlers/chat.go` | Fix | `app_id` auth check; UTF-8 safe truncation |
| `backend/migrations/027_org_members_rls.up.sql` | Add | RLS policy on `org_members` |
| `backend/migrations/027_org_members_rls.down.sql` | Add | Rollback for 027 |
| `frontend/src/api/client.ts` | Fix | `X-Org-ID` header in Axios interceptor |

---

## 0.42.4 — File Length Refactoring (2026-04-13)

Three components that exceeded or approached the 500-line maintainability threshold have been split into smaller, focused modules. No behaviour changes — purely structural.

### AgentNebula.vue: 654 → 47 lines

The 3D particle nebula component was the only file exceeding the 500-line hard limit. The bulk was three self-contained particle layer factories (each with inline GLSL shaders) and the Three.js scene lifecycle.

**Split into:**
- `agentNebulaShaders.ts` (71 lines) — the Ashima 3D Simplex Noise GLSL constant, previously copy-pasted into three shader strings
- `useAgentNebula.ts` (419 lines) — layer factories (`createPrimaryCloud`, `createWispTendrils`, `createCoreMotes`), scene initialisation, animation loop, resize handling, and disposal. Exports `initNebula()` which returns a `{ dispose }` handle
- `AgentNebula.vue` (47 lines) — slim wrapper: template, refs, `onMounted` → `initNebula()`, `onBeforeUnmount` → `dispose()`

### AppHeader.vue: 335 → 118 lines

The header managed three independent dropdown state machines (org, app, profile) with their own toggle/select/close logic and templates. Each dropdown is now a self-contained sub-component:

- `OrgDropdown.vue` (97 lines) — org list with role badges, create org action
- `AppDropdown.vue` (83 lines) — app list with selection, create app action
- `ProfileDropdown.vue` (90 lines) — user email, org link, settings, logout
- `AppHeader.vue` (118 lines) — layout shell, breadcrumb separators, outside-click coordination via `expose`d refs

### github.go: 479 → 236 + 252 lines

The monolithic GitHub handler file was approaching the 500-line threshold. The two concerns — installation flow and repository management — are now in separate files:

- `github_install.go` (236 lines) — `InstallGitHub` and `GitHubCallback`
- `github_repos.go` (252 lines) — `ListGitHubRepos`, `UpdateGitHubRepos`, `TestGitHubConnection`

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `frontend/src/components/connections/AgentNebula.vue` | Refactor | Slim wrapper (654 → 47 lines) |
| `frontend/src/components/connections/useAgentNebula.ts` | New | Scene lifecycle + layer factories |
| `frontend/src/components/connections/agentNebulaShaders.ts` | New | GLSL noise constant |
| `frontend/src/components/common/AppHeader.vue` | Refactor | Layout shell (335 → 118 lines) |
| `frontend/src/components/common/OrgDropdown.vue` | New | Org selector dropdown |
| `frontend/src/components/common/AppDropdown.vue` | New | App selector dropdown |
| `frontend/src/components/common/ProfileDropdown.vue` | New | Profile/logout dropdown |
| `backend/internal/api/handlers/github.go` | Deleted | Replaced by split files |
| `backend/internal/api/handlers/github_install.go` | New | Install + callback handlers |
| `backend/internal/api/handlers/github_repos.go` | New | Repo management handlers |

---

## 0.42.3 — Medium-Severity Fixes (2026-04-13)

Final sweep of the v0.42 code assessment. Six medium-severity issues across agent lifecycle, WebSocket observability, type consistency, and migration safety documentation.

### M1. Agent cancel field zeroed on stop

**File:** `backend/internal/agent/agent.go:111-117`
**Issue:** `Stop()` called `a.cancel()` but never set `a.cancel = nil`. On restart via `Start()`, the nil check at line 83 always found a non-nil cancel, leading to fragile restart logic where the old cancel function was called a second time (harmless but confusing).
**Fix:** `a.cancel = nil` is now set immediately after `a.cancel()` in `Stop()`.

### M2. Down migration data loss documented

**File:** `backend/migrations/026_org_members.down.sql`
**Issue:** Rolling back migration 026 picks only the earliest org membership per user via `DISTINCT ON`. Users with multiple org memberships permanently lose all but one. This is inherent to the schema change but was undocumented.
**Fix:** Added a prominent warning comment at the top of the down migration documenting the data loss and advising to export `org_members` before rolling back.

### M3. WebSocket error now captured and logged

**File:** `frontend/src/composables/useWebSocket.ts`
**Issue:** `onerror` silently set status to `'closed'` without capturing the error event or logging anything. WebSocket failures were completely invisible to both the UI and dev console.
**Fix:** Added `error` ref exposed from the composable. The error handler now stores the event and logs via `console.warn`. Consumers can display the error state if needed.

### M4. Malformed WebSocket messages logged

**File:** `frontend/src/composables/useAgent.ts:93-95`
**Issue:** The JSON parse `catch` block silently ignored malformed server messages. If the backend sent invalid JSON, it disappeared without trace.
**Fix:** The catch block now logs the parse error via `console.warn` for debuggability.

### M5. Dead `"database"` connection type guard removed

**File:** `backend/internal/agent/tools_db.go:39-42`
**Issue:** The type check accepted both `"database"` and `"postgres"`, but `connections_validate.go` only allows `"postgres"` during creation. The `"database"` type can never exist in the database — the check was dead legacy code.
**Fix:** Removed the `"database"` branch. The guard now only checks for `"postgres"`.

### M6. Reports handler made consistent

**File:** `backend/internal/api/handlers/reports.go`
**Issue:** `ListReports` returned `200 []` while `GetReport` returned `501 Not Implemented`. Inconsistent contract — clients assumed reports were supported but empty for list, yet unimplemented for detail.
**Fix:** Both endpoints now return `200 []` (empty array). This matches the frontend's expectation that reports are a supported but currently empty feature. When the reports system is built, both handlers will be replaced together.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `backend/internal/agent/agent.go` | Fix | Zero `cancel` in `Stop()` |
| `backend/migrations/026_org_members.down.sql` | Doc | Data loss warning |
| `frontend/src/composables/useWebSocket.ts` | Fix | Error ref + logging |
| `frontend/src/composables/useAgent.ts` | Fix | Log malformed messages |
| `backend/internal/agent/tools_db.go` | Fix | Remove dead type guard |
| `backend/internal/api/handlers/reports.go` | Fix | Consistent empty response |

---

## 0.42.2 — High-Severity Fixes (2026-04-13)

Continuation of the v0.42 code assessment. Fixes six high-severity issues across backend error handling, frontend state management, routing, and type safety.

### H1. Logs multi-source pagination clarified

**File:** `backend/internal/api/handlers/logs.go:254-271`
**Issue:** When `source=all`, the total reported is the sum of raw + agent counts. The assessment flagged this as incorrect, but on closer inspection the total *is* the true dataset size — the pagination math is correct. Deep pages may return fewer than `limit` items because the merge window (offset+limit per source) can be exhausted before covering all interleaved entries. This is an acceptable approximation that avoids an expensive UNION count query.
**Fix:** Added an explanatory comment documenting the design trade-off. No behaviour change needed.

### H2. Chat title update error logging

**File:** `backend/internal/api/handlers/chat.go:153-160`
**Issue:** `UpdateConversationTitleByUser` silently discarded errors. If the DB call failed, the conversation title remained blank with no logging or observability.
**Fix:** The return value is now checked and logged via `slog.Error` on failure.

### H3. Logs pagination race condition

**File:** `frontend/src/stores/logs.ts:38-50`
**Issue:** `nextPage()` and `prevPage()` modified `offset.value` and fired `fetchLogs()` without awaiting. Rapid clicks triggered concurrent requests with inconsistent offset state.
**Fix:** Both functions now `await fetchLogs()` and early-return if `loading.value` is already true.

### H4. Async logout awaited on 401

**File:** `frontend/src/api/client.ts:28-32`
**Issue:** The 401 interceptor called `auth.logout()` (async) without awaiting, then immediately redirected. The logout could fail to complete before navigation.
**Fix:** Interceptor callback is now `async`, and `auth.logout()` is awaited before the redirect.

### H5. Router guard waits for app store initialisation

**Files:** `frontend/src/stores/app.ts`, `frontend/src/router/index.ts`
**Issue:** The `beforeEach` guard checked `app.needsOnboarding` without confirming the app store had initialised. If the guard fired before `app.init()` completed, stale default state could incorrectly redirect to onboarding.
**Fix:** Added `initialized` ref to the app store (set `true` at the end of `init()`). The router guard now returns early if `!app.initialized`, matching the existing `!auth.initialized` pattern.

### H6. Notification type safety tightened

**File:** `frontend/src/types/notification.ts`
**Issue:** `UpdatePreferencesPayload.severity_threshold` was typed as `string` instead of the `'info' | 'warning' | 'error' | 'critical'` union. `CreateChannelPayload.config` and `UpdateChannelPayload.config` used `Record<string, unknown>` instead of the proper `EmailConfig | SlackConfig | DiscordConfig` union. Both allowed invalid values to pass type checking.
**Fix:** Aligned payload types with their model counterparts. Introduced `ChannelConfig` type alias for the config union.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `backend/internal/api/handlers/logs.go` | Fix | Documented merge-pagination trade-off |
| `backend/internal/api/handlers/chat.go` | Fix | Log title update errors |
| `frontend/src/stores/logs.ts` | Fix | Await pagination, guard against concurrent fetches |
| `frontend/src/api/client.ts` | Fix | Await logout before redirect on 401 |
| `frontend/src/stores/app.ts` | Fix | Added `initialized` flag |
| `frontend/src/router/index.ts` | Fix | Guard on `app.initialized` before onboarding redirect |
| `frontend/src/types/notification.ts` | Fix | Strict types for payloads |

---

## 0.42.1 — Post-Assessment Hardening (2026-04-13)

Comprehensive code assessment of the v0.42.0 multi-org implementation surfaced four critical issues. All four are fixed in this patch — no feature changes, no schema changes beyond an added index.

### C1. Resource leak in codebase tool

**File:** `backend/internal/agent/tools_codebase.go`
**Issue:** The GitHub codebase connector created via `codebase.New()` was never closed. Unlike `tools_db.go` which calls `defer pg.Close()`, the codebase tool leaked HTTP connections on every `search_codebase` invocation.
**Fix:** Added `defer connector.Close()` immediately after connector creation.

### C2. Dead stats query removed

**File:** `backend/internal/db/queries/stats.sql`
**Issue:** `GetDashboardStats` filtered by `user_id` without org scoping. With multi-org, a user in multiple orgs would see aggregated stats across all of them — potential cross-org data leakage. Investigation revealed this query was dead code: only `GetAppDashboardStats` (which is app-scoped and protected by `authorizeApp`) is called from any handler.
**Fix:** Removed the unused `GetDashboardStats` query entirely. Regenerated sqlc.

### C3. Missing `org_members` index

**File:** `backend/migrations/026_org_members.up.sql`
**Issue:** Only `idx_org_members_org_id` was created. RLS policies across 8 tables and queries like `GetOrganizationByUser`, `ListOrganizationsByUser`, `GetOrgMembership`, and `HasOrgMembership` all filter by `user_id`. While the composite PK `(user_id, org_id)` covers exact lookups, a dedicated index ensures optimal performance for user-only scans as the table grows.
**Fix:** Added `CREATE INDEX idx_org_members_user_id ON org_members(user_id)` alongside the existing org_id index.

### C4. Inconsistent API response unwrapping

**Files:** `frontend/src/api/*.ts`, `frontend/src/stores/*.ts`, `frontend/src/pages/AgentChatPage.vue`
**Issue:** API modules were split into two conventions — some returned raw `AxiosResponse<T>` (connections, conversations, reports, schedules, logs), others unwrapped and returned `T` directly (organizations, applications, notifications). Consumers had to handle responses differently, creating type confusion.
**Fix:** Standardised all API functions to unwrap `{ data }` internally and return `Promise<T>`. Updated all store consumers and page components to receive data directly instead of destructuring `{ data }` from the response.

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `backend/internal/agent/tools_codebase.go` | Fix | Added `defer connector.Close()` |
| `backend/internal/db/queries/stats.sql` | Fix | Removed dead `GetDashboardStats` query |
| `backend/internal/db/stats.sql.go` | Regen | Regenerated without dead query |
| `backend/migrations/026_org_members.up.sql` | Fix | Added `user_id` index on `org_members` |
| `frontend/src/api/connections.ts` | Fix | Async/await with data unwrap |
| `frontend/src/api/conversations.ts` | Fix | Async/await with data unwrap |
| `frontend/src/api/reports.ts` | Fix | Async/await with data unwrap |
| `frontend/src/api/schedules.ts` | Fix | Async/await with data unwrap |
| `frontend/src/api/logs.ts` | Fix | Async/await with data unwrap |
| `frontend/src/api/github.ts` | Fix | Async/await with data unwrap |
| `frontend/src/stores/connections.ts` | Fix | Consume unwrapped API responses |
| `frontend/src/stores/logs.ts` | Fix | Consume unwrapped API responses |
| `frontend/src/stores/reports.ts` | Fix | Consume unwrapped API responses |
| `frontend/src/stores/schedules.ts` | Fix | Consume unwrapped API responses |
| `frontend/src/pages/AgentChatPage.vue` | Fix | Consume unwrapped API response |

---

## 0.42.0 — Multi-Org Support & Team Management (2026-04-13)

Heimdall now supports multiple organisations per user, role-based team management, and a dedicated org-level navigation context. The previous 1:1 relationship between users and organisations has been replaced with a many-to-many membership model with roles, and the frontend has been restructured into two navigation contexts — org-level and app-level — connected by an org switcher in the header.

### The problem with single-org

The original data model hard-wired each user to exactly one organisation via a `users.org_id` foreign key. This made multi-tenancy impossible: a consultant working across client orgs, a developer with personal and work projects, or a team lead overseeing multiple product orgs all had to use separate accounts. There was also no concept of team membership — every user in an org had identical permissions.

### Database: `org_members` join table

A new `org_members` table replaces `users.org_id` with a many-to-many relationship:

| Column | Type | Purpose |
|--------|------|---------|
| `user_id` | uuid (PK, FK) | References `users.id` |
| `org_id` | uuid (PK, FK) | References `organizations.id` |
| `role` | `org_member_role` enum | `'owner'`, `'admin'`, `'member'` |
| `created_at` | timestamptz | When membership was created |

Migration `026_org_members` handles the full transition:
1. Creates the `org_member_role` enum and `org_members` table
2. Migrates all existing `users.org_id` relationships as `owner` memberships
3. Rewrites 8 RLS policies across `organizations`, `applications`, `app_agent_config`, `monitoring_state`, `notification_channels`, `notification_preferences`, `notification_log`, and `investigation_schedules` — all changed from `JOIN users u ON u.org_id` to `JOIN org_members om ON om.org_id`
4. Drops the `users.org_id` column

The down migration is fully reversible.

### Backend API — 9 new endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/orgs` | any | List all orgs the user belongs to (with roles) |
| `POST` | `/api/orgs` | any | Create a new organisation (caller becomes owner) |
| `PUT` | `/api/org` | admin+ | Update org name and/or slug |
| `DELETE` | `/api/org` | owner | Delete org (requires typing slug to confirm) |
| `GET` | `/api/org/members` | any member | List all members with email and role |
| `POST` | `/api/org/members/invite` | admin+ | Add user by email (must have existing account) |
| `PUT` | `/api/org/members/{userId}/role` | owner | Change a member's role |
| `DELETE` | `/api/org/members/{userId}` | admin+ | Remove member (sole-owner guard) |
| `GET` | `/api/org` (updated) | any member | Now includes caller's `role` in the response |

Role-based authorization uses a `resolveOrgAndRole` helper that fetches the caller's org and membership in one call, paired with a `hasMinRole` comparison function. This is implemented as handler-level helpers rather than middleware — the authorization logic is visible in each handler rather than hidden in route configuration.

The invite flow (V1) requires the target user to already have a Heimdall account. Future iterations can add pending invite tables, email notifications, and invite links.

### Frontend — two-context navigation

The UI now operates in two navigation contexts:

**Org context** (`/org/*` routes) — OrgSidebar shows Projects, Team, Billing, Settings. The header hides the app selector.

**App context** (all other routes) — AppSidebar shows Dashboard, Activity, Connections, etc. The header shows the full breadcrumb with app selector.

Context switching is driven by the route path: `route.path.startsWith('/org')` determines which sidebar renders. The sidebar swap uses a `<Transition>` crossfade (150ms, `out-in` mode) so the switch feels deliberate rather than jarring.

### Org-level pages

**Org Overview** (`/org`) — Projects grid showing all applications in the organisation. Includes search (client-side filter by name), sort (by name, date, or status), grid/list toggle (persisted to localStorage), and a "New project" button that opens the existing app creation wizard. Each app card shows status badge, connection count, schedule count, and created date. Clicking a card calls `appStore.selectApp()` and navigates to `/dashboard`, switching to app context.

**Team** (`/org/team`) — Full team management page. Members are listed with avatar initials, email, joined date, and role badge. Owners see inline role dropdowns on other members. Admins+ see an invite section with email input, role selector, and role description cards. Remove flow uses a confirmation modal. All controls are permission-gated in the UI and enforced server-side.

**Org Settings** (`/org/settings`) — Editable name and slug (admin+), copyable org ID, created date. Owner-only danger zone with delete confirmation modal that requires typing the org slug. Successful deletion resets the app store and redirects to onboarding.

**Billing** (`/org/billing`) — Placeholder for future billing integration.

### Multi-org switching

The header org breadcrumb is now a dropdown listing all organisations the user belongs to, each showing name and role badge (green for owner, amber for admin, neutral for member). Clicking a different org calls `appStore.selectOrg()` which updates the active org, refetches apps for that org, and navigates to `/org`. A "+ New organisation" option at the bottom opens a creation modal with auto-generated slug.

The app store's `init()` method now fetches all orgs via `GET /api/orgs`, restores the last active org from localStorage (`heimdall_current_org`), and falls back to the first org if the stored ID is no longer valid.

### Old settings page removed

`SettingsPage.vue` and its sub-components (`ProfileSection`, `OrganisationSection`, `ApplicationsSection`) have been deleted. Their functionality is now distributed across the org-level pages. `/settings` redirects to `/org/settings` for bookmark preservation. `DeleteAppModal.vue` is retained as a reusable component.

### Other polish

- **Context-aware logo link** — logo navigates to `/org` in org context, `/dashboard` in app context
- **Profile dropdown** — now includes "Organisation" link to `/org` overview
- **Mobile nav** — auto-closes on context switch via `watch(isOrgContext)`

### Files changed

| File | Kind | Summary |
|------|------|---------|
| `backend/migrations/026_org_members.up.sql` | New | Join table, data migration, 8 RLS policy rewrites |
| `backend/migrations/026_org_members.down.sql` | New | Full rollback |
| `backend/internal/db/queries/organizations.sql` | Edit | 8 new membership queries |
| `backend/internal/db/queries/users.sql` | Edit | `GetUserByEmail`, `HasOrgMembership`, removed `SetUserOrg` |
| `backend/internal/db/queries/applications.sql` | Edit | Auth join via `org_members` |
| `backend/internal/db/models.go` | Regen | `OrgMember` struct, `OrgMemberRole` enum |
| `backend/internal/db/organizations.sql.go` | Regen | Membership CRUD functions |
| `backend/internal/db/users.sql.go` | Regen | Updated user queries |
| `backend/internal/db/applications.sql.go` | Regen | Updated app auth query |
| `backend/internal/api/handlers/org_members.go` | New | Membership handlers + role helpers |
| `backend/internal/api/handlers/organizations.go` | Edit | Updated + new org handlers |
| `backend/internal/api/handlers/testhelpers_test.go` | Edit | Test routes + org_members setup |
| `backend/internal/api/router.go` | Edit | 7 new routes |
| `backend/internal/agent/monitor.go` | Edit | Removed `pgtype` wrapping |
| `frontend/src/types/organization.ts` | Edit | `OrgMember`, `OrgMemberRole`, `OrganizationWithRole` |
| `frontend/src/api/organizations.ts` | Edit | 7 new API client functions |
| `frontend/src/stores/app.ts` | Edit | Multi-org state, `selectOrg`, `addOrg` |
| `frontend/src/router/index.ts` | Edit | 4 org routes, `/settings` redirect |
| `frontend/src/layouts/DefaultLayout.vue` | Edit | Context-aware sidebar with crossfade |
| `frontend/src/components/common/AppHeader.vue` | Edit | Org switcher dropdown, context-aware logo |
| `frontend/src/components/common/OrgSidebar.vue` | New | Org-level navigation sidebar |
| `frontend/src/components/org/AppCard.vue` | New | Grid-view app card |
| `frontend/src/components/org/AppListRow.vue` | New | List-view app row |
| `frontend/src/components/org/CreateOrgModal.vue` | New | New org creation with auto-slug |
| `frontend/src/pages/org/OrgOverviewPage.vue` | New | Projects grid with search/sort/toggle |
| `frontend/src/pages/org/OrgTeamPage.vue` | New | Team management with invite/role/remove |
| `frontend/src/pages/org/OrgSettingsPage.vue` | New | Org settings + danger zone |
| `frontend/src/pages/org/OrgBillingPage.vue` | New | Billing placeholder |
| `frontend/src/pages/SettingsPage.vue` | Delete | Replaced by org pages |
| `frontend/src/components/settings/ProfileSection.vue` | Delete | Profile in header dropdown |
| `frontend/src/components/settings/OrganisationSection.vue` | Delete | Replaced by OrgSettingsPage |
| `frontend/src/components/settings/ApplicationsSection.vue` | Delete | Replaced by OrgOverviewPage |

### Verification

| Check | Result |
|-------|--------|
| `go build ./...` | Clean |
| `go vet ./...` | Clean |
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| `vitest run` | 52/52 tests pass |

---

## 0.41.1 — Unified Connections Page: Three-Lane Layout (2026-04-13)

The three separate Infrastructure pages (Ingestion, Enrichment, Outbound) have been consolidated into a single **Connections** page with a three-lane visual layout. Each lane groups connections by their data-flow role — what flows in, what the agent queries, and where it sends results — all rendered above the shared 3D agent nebula.

### The problem with three separate pages

Splitting connections across three tabs fragmented the visual metaphor. Most deployments have 3–8 total connections; distributing them across three pages meant each page showed 1–2 lonely bubbles above an identical nebula. The connection wizard already categorised connectors at creation time — repeating that taxonomy as top-level navigation added clicks without adding clarity.

### The approach — one page, three visual lanes

All connections live on a single page. A three-column grid groups them by category, each with a header showing the lane name, a directional icon, and a sublabel:

| Lane | Icon | Sublabel | What belongs here |
|------|------|----------|-------------------|
| **Ingestion** | ↓ arrow | Data flowing in | Supabase, Webhook, Syslog, OTLP, platform log sources |
| **Enrichment** | ↔ bidirectional | Agent tools | PostgreSQL, GitHub, MySQL — things the agent queries during investigations |
| **Outbound** | ↑ arrow | Alerts flowing out | Slack, Telegram, Linear, Trajan — where the agent delivers results |

Connections within each lane are arranged in a consistent **2-column grid** for clean alignment regardless of count.

### Category derivation

A new `typeToCategory()` utility in `flows.ts` maps each connection type string to its visual category (`'ingestion' | 'enrichment' | 'outbound'`). This is a pure frontend concern — no database schema changes, no new API fields. The category is derived from the connection type at render time.

### Curved arrow flow lines

The flow lines connecting bubbles to the nebula have been redesigned:

- **Geometry**: Quadratic beziers (`Q`, one control point) replaced with **cubic beziers** (`C`, two control points). CP1 drops the line straight down from the bubble for a clean vertical departure; CP2 sweeps it horizontally into the nebula centre for a smooth arrival. The curvature scales with horizontal distance — bubbles directly above get gentle arcs, side-column bubbles get dramatic S-curves.

- **Arrowheads**: SVG `<marker>` elements render small triangular arrowheads that auto-orient along the path:
  - Ingestion: arrow at nebula end (data flows in)
  - Outbound: arrow at bubble end (data flows out)
  - Enrichment: arrows on both ends (bidirectional)

- **Animation per category**:
  - Ingestion: `flow-down` — dashes animate top-to-bottom (2.5s linear)
  - Outbound: `flow-up` — dashes animate bottom-to-top (reversed)
  - Enrichment: `flow-bidir` — dashes pulse back and forth (3s ease-in-out), with shorter dash pattern (`2 6` vs `4 8`)

### Outbound connectors added to wizard

Four new outbound connector types added to the connection wizard (all coming soon):

| Connector | Description |
|-----------|-------------|
| **Slack** | Send alerts and reports to Slack channels |
| **Telegram** | Send alerts to Telegram chats and groups |
| **Linear** | Create issues from incidents automatically |
| **Trajan** | Create tickets in Trajan project management |

The wizard's `PlatformGrid` now has a fourth section — "Outbound channels" — with a send icon and descriptive copy explaining the purpose.

### Connector logos

SVG brand marks added to `ConnectorLogo.vue` for all four outbound types: Slack (hash mark, sourced from Simple Icons), Telegram (paper plane), Linear (official mark), and Trajan (custom "T" in rounded rectangle).

### Navigation changes

- **Sidebar**: Three items (Ingestion / Enrichment / Outbound) under "Infrastructure" replaced with single **Connections** item
- **Router**: `/enrichment` and `/outbound` now redirect to `/connections` (preserves existing bookmarks)
- **Deleted**: `EnrichmentPage.vue` and `OutboundPage.vue` stub pages removed

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/common/AppSidebar.vue` | Edit | Three nav items → one "Connections" item |
| `frontend/src/router/index.ts` | Edit | Enrichment/Outbound routes → redirects |
| `frontend/src/pages/EnrichmentPage.vue` | Delete | Stub removed |
| `frontend/src/pages/OutboundPage.vue` | Delete | Stub removed |
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Three-lane grid layout with category headers |
| `frontend/src/components/connections/ConnectionBubble.vue` | Edit | Removed max-width for grid-cell fill |
| `frontend/src/components/connections/FlowLines.vue` | Edit | Cubic S-curves, SVG arrowheads, per-category animation |
| `frontend/src/components/connections/wizard/flows.ts` | Edit | `ConnectionCategory` type, `typeToCategory()`, `'outbound'` section, 4 outbound connectors |
| `frontend/src/components/connections/wizard/PlatformGrid.vue` | Edit | New "Outbound channels" section |
| `frontend/src/components/icons/ConnectorLogo.vue` | Edit | Slack, Telegram, Linear, Trajan SVG logos |

### Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| `vitest run` | 52/52 tests pass |

---

## 0.41.0 — Neutral Canvas Colour Rebalance (2026-04-13)

The UI has been rebalanced from a green-tinted-everything aesthetic to a neutral dark canvas where phosphor green appears only on intentional interactive and brand elements. The same bones, substantially less green wash.

### The problem

The v0.35.0 "Technical Retro-Futurism" refresh applied green tint to every visual layer — backgrounds, all three text tiers, borders, scrollbars, text selection, surface textures, scanline overlays, and fade-in animations. The cumulative effect was wearing green-tinted glasses: the brand colour lost all punch because there was nothing neutral to contrast against.

Sister platform Trajan demonstrates the correct approach — orange appears only on badges, active tabs, and small interactive highlights while surfaces, borders, and text are all neutral grays. The accent colour punctuates; it doesn't permeate.

### The approach — neutral canvas, surgical green

Every ambient layer (backgrounds, text, borders, textures) shifted to pure cool neutrals. Green is reserved for elements where it communicates something — action buttons, active states, status indicators, brand decoration. Green coverage drops from ~80% of visual surface area to ~5–10%.

### Design token changes

**Backgrounds** — removed green cast from all four surface tokens:

| Token | Before | After |
|-------|--------|-------|
| `--bg-primary` | `#06080a` (green-black) | `#08090a` (pure neutral) |
| `--bg-surface` | `rgba(12, 16, 14, 0.55)` | `rgba(15, 16, 18, 0.55)` |
| `--bg-surface-hover` | `rgba(14, 20, 17, 0.75)` | `rgba(20, 21, 24, 0.75)` |
| `--bg-elevated` | `#0c100e` (green-black) | `#0e0f11` (neutral) |

**Text** — all three tiers shifted from green-tinted to cool slate:

| Token | Before | After |
|-------|--------|-------|
| `--text-primary` | `#e8ede9` (green white) | `#e2e4e8` (cool neutral) |
| `--text-secondary` | `#8a9e92` (muted green) | `#8a8f96` (cool slate) |
| `--text-muted` | `#4e6556` (obvious green) | `#4a4f56` (dark slate) |

**Borders** — structural lines are now colourless:

| Token | Before | After |
|-------|--------|-------|
| `--border` | `rgba(74, 122, 92, 0.18)` (green) | `rgba(140, 145, 155, 0.14)` (neutral gray) |
| `--border-hover` | `rgba(74, 122, 92, 0.32)` (green) | `rgba(140, 145, 155, 0.26)` (neutral gray) |

**Accent** — kept green, slightly tightened `--accent-subtle` (0.12 → 0.10) and `--accent-border` (0.25 → 0.22) to avoid over-bleed on the now-neutral canvas. All other accent, action, and status tokens unchanged.

### Global style changes

**Scrollbar thumb** — green → neutral gray at all three states (default, hover, active). Removed the green glow on active state.

**Text selection** — `rgba(74, 122, 92, 0.3)` → `rgba(140, 145, 155, 0.25)`.

**Surface grid texture** — green gridlines → neutral gray, opacity reduced from 0.06 to 0.04 for subtlety.

**Scanline overlay** — green banding → neutral gray (same 2% opacity).

**Fade-in animation** — removed the `border-color` transition that flashed green on every list item entry. Animation now handles only opacity and transform.

### AgentNebula shader rebalance

The 3D particle nebula on the Connections page had hardcoded GLSL colour values across three particle layers (ambient cloud, wisp tendrils, core motes). All three were rebalanced:

- **Base colours** shifted from green-dominant (`vec3(0.10, 0.20, 0.14)`) to neutral slate (`vec3(0.10, 0.11, 0.13)`)
- **Green accent mix weights** reduced (0.50 → 0.35) to prevent green from dominating the palette
- **Teal and warm amber** given proportionally more weight for colour variety
- **Dormant state** desaturates to neutral grey instead of green-grey

The nebula now reads as a moody, multi-tonal cloud that occasionally catches green highlights — rather than a green fog.

### What stays green

These elements retain phosphor green because it's meaningful there:

| Element | Why green |
|---------|----------|
| `--action` / `--action-hover` | Primary CTA — "do this" |
| `--accent` on active nav items | "You are here" |
| `--status-ok` | "This is healthy" |
| Chrome brackets (`.chrome-brackets`) | Brand decoration — HUD corners |
| Brand wordmark glow (`.brand-glow`) | Logo identity |
| FlowLines on Connections page | Animated data-flow visualisation |
| Glow-pulse animations | Atmospheric brand effect |

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/assets/styles/main.css` | Edit | All token + global style changes above |
| `frontend/src/components/connections/AgentNebula.vue` | Edit | GLSL shader palette rebalance (3 layers) |

### Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| `vitest run` | 52/52 tests pass |

---

## 0.40.0 — Connection Wizard: Platform-First Redesign (2026-04-13)

The "New Connection" wizard has been redesigned from a flat technical list into a structured, purpose-driven layout. The previous grid grouped connectors by implementation type (Log Sources / Databases / Generic) — categories that require the user to already know what protocol their platform uses. The new layout asks "where does your stuff run?" and communicates the *consequence* of each connection type.

### The problem with the old layout

The old wizard opened with three categories: **Log Sources** (Supabase, Webhook, Syslog, OTLP, Datadog), **Databases** (PostgreSQL, MySQL), and **Generic** (GitHub). This grouping was technically accurate but failed users in two ways:

1. **Platform-blind.** A developer on Fly.io had to know that Fly.io uses syslog drains, then find "Syslog" in the Log Sources section. The wizard forced users to translate from "where my stuff runs" to "what protocol Heimdall supports."

2. **No consequence labelling.** Adding PostgreSQL doesn't make logs flow — it gives the agent a queryable tool for investigations. Adding Supabase *does* make logs flow. The old grid treated both as equivalent items with no indication that one feeds the monitoring loop and the other doesn't.

### The new three-section layout

```
WHERE DO YOUR LOGS COME FROM?
Pick your platform — we'll handle the wiring.

  [Supabase]  [Fly.io]  [Vercel]   [Render]
  [Railway]   [Heroku]  [AWS]      [DigitalOcean]

──────────────── or ────────────────

CONNECT DIRECTLY
Already know the protocol? Skip the platform.

  [Webhook HTTP]  [Syslog TCP/TLS]  [OpenTelemetry OTLP]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 AGENT INVESTIGATION TOOLS
These don't send logs — they let the agent query
your systems during investigations and chat.

  [PostgreSQL]  [GitHub]  [MySQL]
```

**Section 1 — Platform log sources.** The primary question: "where does your app run?" Each platform card shows the underlying transport in its subtitle ("via Syslog drain", "via Log Drain", "via CloudWatch + OTLP"). Currently only Supabase is live; the remaining 7 platforms are "Coming soon" placeholders. Each maps to an existing connector type — when enabled, they'll start the appropriate flow (syslog, webhook, or OTLP) with platform-specific setup instructions. No backend work required.

**Section 2 — Direct protocols.** The escape hatch for engineers who already know the transport they want. Webhook (HTTP endpoint), Syslog (TCP/TLS listener), and OpenTelemetry (OTLP HTTP). All three are live and use the same wizard steps as before.

**Section 3 — Agent investigation tools.** Visually separated with a stronger border and a search icon. The subtitle "These don't send logs — they let the agent query your systems during investigations and chat" prevents the most common onboarding confusion: adding PostgreSQL alone and expecting monitoring to start. PostgreSQL and GitHub are live; MySQL is coming soon.

### Visual separation between sections

The three sections use intentionally different visual weights:

| Transition | Visual | Why |
|-----------|--------|-----|
| Sections 1 → 2 | Soft "or" divider (horizontal line with centred text) | Both sections achieve the same outcome (logs flowing in) — different paths based on user expertise |
| Sections 2 → 3 | Strong `border-t` + search icon in header | Fundamentally different category with different consequences — user should notice the break |

### Platform → connector mapping

Each coming-soon platform stores a `connectorType` that maps to an existing backend connector:

| Platform | Underlying connector | Transport |
|----------|---------------------|-----------|
| Fly.io | `syslog` | TCP/TLS drain |
| Vercel | `webhook_logs` | HTTP log drain |
| Render | `syslog` | TCP/TLS drain |
| Railway | `webhook_logs` | HTTP log drain |
| Heroku | `syslog` | Syslog drain |
| AWS | `otlp` | CloudWatch → OTLP |
| DigitalOcean | `syslog` | Log forwarding |

When a platform becomes available, it only needs `available: true` and wizard steps — the backend already handles the underlying protocol.

### Card redesign

Cards switched from a horizontal layout with 2-letter abbreviation badges to a vertical, centre-aligned layout with native brand logos (via `ConnectorLogo`) and a short subtitle instead of the longer description. This fits the denser 4-column grid needed for the platform section.

### Brand logos added

Official SVG marks from Simple Icons for: **Fly.io** (bird mark), **Vercel** (triangle), **Render** (stylised R), **Railway** (rail mark), **Heroku** (H mark), **AWS** (smile arrow), **DigitalOcean** (droplet). All mono-colour, `currentColor`-based.

### Data model change

The `PlatformFlow` interface in `flows.ts` was updated:

| Field | Before | After |
|-------|--------|-------|
| `category` | `'log_source' \| 'database' \| 'generic'` | **Removed** |
| `section` | — | **New:** `'platform_log' \| 'direct_protocol' \| 'agent_tool'` |
| `subtitle` | — | **New:** Short label shown on card (e.g. "via Syslog drain") |

The old `datadog` entry was removed (never wired up, doesn't fit the new taxonomy). Can be re-added as `platform_log` when built.

### What didn't change

The wizard state machine (`ConnectionWizard.vue`), all 8 step components (`StepName`, `StepSupabaseAuth`, `StepSupabaseTables`, `StepPostgresConfig`, `StepWebhookSetup`, `StepGitHubInstall`, `StepSyslogConfig`, `StepOTLPSetup`, `StepTest`), the step indicator (`WizardStepIndicator.vue`), the edit form (`ConnectionForm.vue`), and the entire backend are untouched. The redesign lives entirely in 4 frontend files.

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/connections/wizard/flows.ts` | Edit | `section` + `subtitle` replace `category`, 7 new platform entries, removed `datadog` |
| `frontend/src/components/icons/ConnectorLogo.vue` | Edit | Added 7 platform logos (Fly.io, Vercel, Render, Railway, Heroku, AWS, DigitalOcean) |
| `frontend/src/components/connections/wizard/PlatformGrid.vue` | Rewrite | 3-section layout with headers, subtitles, "or" divider, strong separator |
| `frontend/src/components/connections/wizard/PlatformCard.vue` | Rewrite | ConnectorLogo + subtitle, vertical centre layout, `cursor-default` for coming-soon |

### Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| `vitest run` | 52/52 tests pass |
| Supabase flow | select → Name → Auth → Tables → Test → created |
| Webhook flow | select → Name → Setup → created |
| Syslog flow | select → Name → Config → Test → created |
| OTLP flow | select → Name → Setup → created |
| PostgreSQL flow | select → Name → Config → Test → created |
| GitHub flow | select → Name → Install → redirect |
| Coming-soon platforms | Render with logo, name, subtitle, "Soon" badge; click does nothing; default cursor |

---

## 0.39.0 — Ingestion Page Redesign: 3D Agent Nebula (2026-04-13)

The Connections page has been rebuilt from the ground up around a new visual metaphor: data sources feed into the Heimdall agent. The previous dual-view layout (Blueprint grid + flat List, toggled via a segmented control) is replaced by a single unified view — clickable connection bubbles with native brand logos arranged above an ephemeral 3D particle nebula representing the agent, connected by animated SVG flow lines.

### Why this redesign

The old Blueprint view grouped connections by type in a left/centre/right column layout with a static eye icon as the "hub." It communicated structure but not the mental model that matters: **data flows from your infrastructure into an intelligent, living agent**. The new design makes the agent tangible — a breathing particle cloud that processes the data flowing into it from each connection above.

### The new layout

```
┌─────────────────────────────────────────────────┐
│  CONNECTIONS              [+ NEW CONNECTION]     │
├─────────────────────────────────────────────────┤
│                                                 │
│   [🔷 GitHub]  [⚡ Supabase]  [🐘 PostgreSQL]   │  ← Connection bubbles
│        \            |            /              │    with native brand logos
│         ╲           │           ╱               │
│          ▼          ▼          ▼                │  ← Animated flow lines
│                                                 │
│            ░░░░░░░░░░░░░░░░░░░░                │
│            ░░ PARTICLE NEBULA ░░                │  ← 3D agent abstraction
│            ░░░░░░░░░░░░░░░░░░░░                │    (Three.js + GLSL)
│                                                 │
└─────────────────────────────────────────────────┘
```

### Connection bubbles

Each connection is a clickable bubble component displaying:
- **Native brand SVG** — official logos from Simple Icons for Supabase, PostgreSQL, GitHub, OpenTelemetry, Datadog, MySQL; custom icons for Webhook and Syslog
- **Connection name** and type label
- **Status indicator** — green glow when active, red for error, grey for inactive
- **Staggered mount animation** — bubbles fade in sequentially via CSS custom property delay

Clicking a bubble opens the new **Connection Detail Modal** — a full-info overlay showing type-specific configuration (host, port, database, polling tables, etc.), metadata (created/updated dates, connection ID), masked sensitive fields with reveal toggles, and action buttons (Ping, Edit, Delete with two-step confirmation, Manage Repos for GitHub).

### Agent nebula

A 3-layer volumetric particle cloud rendered with raw Three.js and custom GLSL shaders, adapted from the Elephantasm reference architecture:

| Layer | Particles | Role |
|-------|-----------|------|
| Primary Cloud | 4,000 | Main body — 3-octave simplex noise displacement, Gaussian distribution |
| Wisp Tendrils | 1,000 | Atmospheric edge — radial drift with tangential noise, uniform shell |
| Core Motes | 400 | Bright heartbeat — high-frequency micro-jitter, tight Gaussian core |

The colour palette is tuned to Heimdall's retro-futurism theme: dark feldgrau base with phosphor-green, teal, and warm amber mood tints that cycle through irrational sine frequencies — the nebula never repeats within ~7 minutes. A global brightness pulse adds a slow breathing rhythm to the colour intensity.

Key properties:
- `THREE.Points` with custom `ShaderMaterial` and `AdditiveBlending`
- Ashima 3D Simplex noise (GLSL, inlined)
- `dormant` prop dims the nebula when no connections exist
- `prefers-reduced-motion` renders one frame then stops
- Full cleanup in `onBeforeUnmount` (geometries, materials, renderer disposed — no WebGL context leaks)

### Flow lines

Animated SVG dashed paths connect each bubble's bottom-centre to the nebula's top-centre via quadratic bezier curves. Two overlapping `<path>` elements per line: a faint static base (structural reference) and an animated flow with glow filter (continuous top-to-bottom dash movement at 2.5s period, staggered by 300ms per line). Hidden on mobile. Recomputed via `ResizeObserver` on layout changes.

### Empty state

When no connections exist, the nebula renders in dormant mode — dimmer, desaturated, tighter breathing — with an overlaid "No connections yet" message and "+ New Connection" CTA. The breathing dormant nebula communicates "the system is here, waiting to be activated."

### Bundle optimisation

Three.js is split into a separate `vendor-three` chunk via `manualChunks` in the Vite config. It loads lazily only when the user navigates to the Connections page.

| Chunk | Size (gzip) |
|-------|-------------|
| `ConnectionsPage` | 21 KB |
| `vendor-three` (lazy) | 122 KB |
| `index` (main bundle) | 115 KB — unchanged |

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/package.json` | Edit | Added `three`, `@types/three` |
| `frontend/vite.config.js` | Edit | Added `manualChunks` for three.js, `chunkSizeWarningLimit` |
| `frontend/vite.config.ts` | Edit | Synced same build config |
| `frontend/src/components/connections/AgentNebula.vue` | **New** | 3D particle nebula component |
| `frontend/src/components/icons/ConnectorLogo.vue` | **New** | Brand logo SVGs by type |
| `frontend/src/components/connections/ConnectionBubble.vue` | **New** | Clickable connection bubble |
| `frontend/src/components/connections/ConnectionDetailModal.vue` | **New** | Full-info detail modal |
| `frontend/src/components/connections/FlowLines.vue` | **New** | SVG animated connector lines |
| `frontend/src/pages/ConnectionsPage.vue` | Rewrite | Unified ingestion view replacing Blueprint/List |
| `frontend/src/components/connections/BlueprintView.vue` | **Deleted** | Replaced |
| `frontend/src/components/connections/BlueprintZone.vue` | **Deleted** | Replaced |
| `frontend/src/components/connections/BlueprintNode.vue` | **Deleted** | Replaced |
| `frontend/src/components/connections/ViewToggle.vue` | **Deleted** | No longer two view modes |
| `frontend/src/components/connections/ConnectionList.vue` | **Deleted** | Replaced |
| `frontend/src/components/connections/ConnectionCard.vue` | **Deleted** | Replaced |

### Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean — three.js in separate lazy chunk |
| `vitest run` | 52/52 tests pass |
| Existing flows | Wizard, edit form, test modal, GitHub repo selector, error banners all preserved |
| `prefers-reduced-motion` | Nebula freezes, flow lines static, bubble animations instant |

---

## 0.38.0 — Navigation Restructure & Infrastructure Triptych (2026-04-13)

The sidebar and header navigation have been restructured around clearer mental models. The previous layout grouped items by implementation origin (what was built when), but as the platform matures the navigation should reflect how users think about the system — not how it was assembled. This release reorganises the sidebar into four intentional sections, introduces an infrastructure triptych (Ingestion / Enrichment / Outbound), and promotes Agent Chat from a sidebar link to a persistent header icon.

### The three changes

**1. Notifications moves from AGENT → INTELLIGENCE**

Notifications are outputs of the agent's analytical work — assessment results, escalations, threshold breaches. They belong alongside Reports under Intelligence, not next to Configuration and Schedules under Agent. The Agent section is now purely about *configuring and scheduling* the agent's behaviour; Intelligence is where you *consume its output*.

**2. Agent Chat promoted to header icon**

Chat was a sidebar navigation item under Agent, which buried the most interactive surface in a list of configuration pages. This release removes it from the sidebar and adds a persistent icon button in the top-right header bar, immediately left of the Profile avatar. The icon is a hexagon with a centre dot — a symmetrical, technical shape that evokes a node or hub without resorting to speech-bubble clichés. It links to the existing `/agent/chat` route with no URL change. On hover, the icon highlights to the accent green and the border picks up the accent glow, consistent with the retro-futurism palette.

**3. Infrastructure triptych: Ingestion / Enrichment / Outbound**

The single "Connections" item under Infrastructure has been expanded into three items that map to a data-flow mental model:

| Item | Purpose | Status |
|------|---------|--------|
| **Ingestion** | Log sources and data pipelines flowing into Heimdall — webhooks, syslog, OTLP, API pollers | Live (renamed from Connections, same `/connections` route) |
| **Enrichment** | Context sources that help the agent understand the application — GitHub repos/codebases, manual system descriptions, architecture documentation | Placeholder (`/enrichment`, "Coming soon") |
| **Outbound** | Delivery channels for alerts and reports — Slack, Telegram, and future third-party integrations (Linear, Trajan, etc.) | Placeholder (`/outbound`, "Coming soon") |

The distinction matters for onboarding: Ingestion answers "where does data come from?", Enrichment answers "what context does the agent need?", and Outbound answers "where should findings go?". Keeping them separate makes each concept self-explanatory without requiring users to understand a single overloaded "Connections" page.

**Ingestion** reuses the existing Connections page and route (`/connections`, route name `connections`) — only the sidebar label changed, so no URLs break and no backend changes are needed.

**Enrichment** and **Outbound** are new Vue pages (`EnrichmentPage.vue`, `OutboundPage.vue`) with matching routes. Both render a page header with a descriptive subtitle and a "Coming soon" placeholder, following the same page-header pattern used by Reports and other pages.

### Resulting sidebar structure

```
OVERVIEW
  Dashboard
  Activity

INFRASTRUCTURE
  Ingestion          ← renamed from "Connections"
  Enrichment         ← new placeholder
  Outbound           ← new placeholder

AGENT
  Configuration
  Schedules
                     ← Chat removed (now in header)
                     ← Notifications removed (moved to Intelligence)

INTELLIGENCE
  Reports
  Notifications      ← moved from Agent
```

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/common/AppSidebar.vue` | Edit | Restructured sections array: renamed Connections → Ingestion, added Enrichment + Outbound, removed Chat, moved Notifications to Intelligence |
| `frontend/src/components/common/AppHeader.vue` | Edit | Added agent chat hexagon icon button in header right section, left of Profile avatar |
| `frontend/src/router/index.ts` | Edit | Added `/enrichment` and `/outbound` routes with lazy-loaded page components |
| `frontend/src/pages/EnrichmentPage.vue` | **New** | Placeholder page with header and "Coming soon" |
| `frontend/src/pages/OutboundPage.vue` | **New** | Placeholder page with header and "Coming soon" |

### Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean — both new pages appear as separate chunks |
| Existing routes | Unchanged — `/connections`, `/agent/chat`, `/notifications`, `/reports` all resolve to the same components |

---

## 0.37.0 — Activity Feed App-Scoping (2026-04-13)

The Activity feed was the last major surface in Heimdall that ignored the application selector — it merged logs and agent observations from every app into one org-wide timeline. Every other per-app surface (Connections, Dashboard, Agent Config, Scheduled Investigations) already filtered by the selected application, so switching apps left the Activity page showing stale cross-app noise. This release adds end-to-end app-scoping: a new `app_id` column on both log tables, app-filtered SQL queries, write-path population across all 8 ingestion code paths, and a frontend watcher that re-fetches when the user switches applications.

### Why denormalise instead of joining?

`log_buffer` already has a path to `app_id` through `connection_id → connections.app_id`, but joining on every Activity page load adds latency and complexity for a column that never changes after insert. `agent_log` is worse — its only app reference was buried in the JSONB `detail` payload, which can't be indexed efficiently. Adding a direct `app_id UUID` column to both tables trades a small write-time cost (one extra parameter on every INSERT) for a single `WHERE app_id = $1` clause on reads — no joins, no JSONB extraction, and the partial indexes serve exactly the queries the handler issues.

### Phase 1 — Database migration (`025_add_app_id_to_logs`)

**`log_buffer.app_id`** — nullable UUID with FK to `applications(id) ON DELETE CASCADE`. The backfill joins on `connections.app_id` to populate existing rows. At current production scale (0 rows in log_buffer due to the 48-hour retention pruner) the backfill is instantaneous, but the migration is written to handle arbitrary row counts safely.

**`agent_log.app_id`** — same column definition. The backfill extracts `(detail->>'app_id')::uuid` from the JSONB payload where present. Production had 32 agent_log rows, all successfully backfilled.

**Partial indexes** on both tables: `(app_id, timestamp DESC) WHERE app_id IS NOT NULL`. The `WHERE` clause keeps the index lean — historical rows with NULL `app_id` (pre-migration interactive chat entries, org-level audit events) are excluded from the index since they'll never match an app-scoped query. PostgreSQL's planner recognises that `app_id = $1` implies `app_id IS NOT NULL` and uses the partial index automatically.

**Down migration** drops indexes before columns (correct dependency order) with `IF EXISTS` guards.

### Phase 2 — Backend queries & handler

**Six new sqlc queries** mirror the existing `*ByUser` variants with an added `AND app_id = $N` clause:

| Query | Table | Filters |
|-------|-------|---------|
| `ListLogsByApp` | log_buffer | user_id + app_id |
| `ListLogsByAppAndSeverity` | log_buffer | user_id + app_id + severity |
| `CountLogsByApp` | log_buffer | user_id + app_id |
| `CountLogsByAppAndSeverity` | log_buffer | user_id + app_id + severity |
| `ListAgentLogByApp` | agent_log | user_id + app_id |
| `CountAgentLogByApp` | agent_log | user_id + app_id |

All six queries are served by the Phase 1 partial indexes — the `(app_id, timestamp DESC)` ordering matches the `ORDER BY ... DESC LIMIT N OFFSET M` pattern exactly, so the planner can satisfy the query with an index-only scan (plus a recheck against the user_id filter, which is always selective given single-user orgs).

**`ListLogs` handler** (`logs.go`) gains an optional `app_id` query parameter:

1. **Parse** — UUID validation, returns 400 on malformed input
2. **Authorise** — `GetApplicationByOrgUser(appID, userID)` confirms the app belongs to the user's org, returns 404 on mismatch
3. **Route** — a switch statement selects the correct query variant: `connectionID` takes priority (forces `source=raw`), then `appID+severity`, then `appID` alone, then the existing `severity` and `user-only` defaults
4. **Count** — the count query mirrors the list query exactly, so pagination totals are always accurate for the active filter combination

Backwards compatible: omitting `app_id` produces identical behaviour to before.

### Phase 3 — Write-path population

Every code path that INSERTs into `log_buffer` or `agent_log` now populates `app_id`. This is the most diffuse phase — 15 files touched — but each change is mechanical: accept an `appID` parameter, convert to `pgtype.UUID`, pass as the new query field.

**`InsertLogEntry` callers (log_buffer):**

| Caller | Source of app_id |
|--------|-----------------|
| Webhook handler (`webhooks.go`) | `conn.AppID` from `GetConnectionByWebhookToken` |
| OTLP handler (`otlp.go`) | `conn.AppID` from `GetConnectionByWebhookToken` |
| Supabase poller (`supabase.go`) | Stored `appID` field from constructor |
| Fly.io poller (`flyio.go`) | Stored `appID` field from constructor |
| Vercel poller (`vercel.go`) | Stored `appID` field from constructor |
| Railway poller (`railway.go`) | Stored `appID` field from constructor |
| MongoDB poller (`mongodb.go`) | Stored `appID` field from constructor |
| Syslog listener (`syslog.go`) | Stored `appID` field from constructor |

**Factory & resume paths.** `StartPoller` gains an `appID uuid.UUID` parameter, threaded through from `conn.AppID` at both create-time (connection handlers) and resume-time (`main.go` startup loop via `ListActiveConnectionsByType`). `NewSyslog` follows the same pattern. The validation handler passes `conn.AppID` for constructor compatibility but never inserts real data (test connections are read-only probes).

**`InsertAgentLog` callers (agent_log) — via `emit.go`:**

The emit layer's signature changes from `(ctx, userID, conversationID, ...)` to `(ctx, userID, appID *uuid.UUID, conversationID, ...)`. The `*uuid.UUID` type is the key design choice — `nil` means "no app context" (org-level events, interactive chat without an app selected), and the conversion to `pgtype.UUID` happens at the emit boundary so the rest of the agent layer works with native Go types.

| Caller | Source of app_id |
|--------|-----------------|
| Interactive chat (`loop.go`) | `appIDPtr` — derived from the `appID uuid.UUID` parameter; nil when `uuid.Nil` |
| Monitoring assessment (`monitor.go`) | `&app.ID` from `ListActiveApplicationsRow` |
| Monitoring tool calls (`loop.go` RunMonitoring) | `&monAppID` from `AppAgentConfig.AppID` |
| Scheduled investigation (`scheduler.go`) | `&s.AppID` from `InvestigationSchedule` |
| App created/deleted audit (`applications.go`) | `nil` — org-level events have no single app context |

### Phase 4 — Frontend

**API client** (`api/logs.ts`) — `app_id?: string` added to the `listLogs` params interface. Optional and backwards compatible.

**Logs store** (`stores/logs.ts`) — every `fetchLogs` call now includes `app_id: appStore.currentAppId ?? undefined`. The `?? undefined` conversion is load-bearing: `currentAppId` can be `null` (no app selected), and passing `null` to Axios would serialize as the literal string `"null"` in the query string. Converting to `undefined` causes Axios to omit the parameter entirely, which is the correct "unfiltered" signal. `useAppStore()` is called inside `fetchLogs` (not at store definition time) to avoid Pinia circular-dependency issues — this matches the lazy-access pattern already used by other stores.

**Activity page** (`ActivityPage.vue`) — a `watch()` on `appStore.currentAppId` resets pagination and re-fetches logs. This is the same reactive pattern used by ConnectionsPage, SchedulesPage, NotificationsPage, and DashboardPage — switching apps in the header breadcrumb immediately refreshes the Activity feed to show only that app's data.

### Phase 5 — Verification

| Check | Result |
|-------|--------|
| `go build ./...` | Clean |
| `go vet ./...` | Clean |
| `vue-tsc --noEmit` | Clean |
| Backend unit tests | All pass |
| Frontend tests (52) | All pass |
| `TestListLogs_AppScoped` | Inserts one scoped + one unscoped row; filtered query returns exactly 1 |
| `TestListLogs_AppScoped_InvalidAppID` | Returns 400 |
| `TestListLogs_AppScoped_WrongOrg` | Returns 404 |
| `fetchLogs omits app_id when no app selected` | Confirms undefined is not serialised |

### Design decisions & trade-offs

**Nullable `app_id`.** Interactive chat entries created before this release (and future chat sessions opened without an app selected) have `app_id = NULL`. When filtering by app, these rows are excluded — which is correct, since they don't belong to any specific application. When no `app_id` filter is active, all rows (including NULLs) are returned, preserving the org-wide view.

**`source=all` merge pagination.** When fetching from both sources (raw + agent), the handler fetches `offset + limit` rows from each source, merges by timestamp, then applies the original offset/limit to the merged set. The `total` is the sum of both source counts. At high offsets with uneven distribution, the last pages could show fewer items than expected — bounded by the `maxOffset = 10000` cap. This is an acceptable trade-off for avoiding a materialised view or UNION query.

**No `ListAgentLogByAppAndType` query.** The existing `ListAgentLogByUserAndType` has no app-scoped counterpart. The UI doesn't currently combine entry-type filtering with app filtering, so the query isn't needed yet — it can be added as a single SQL + sqlc-generate step if the need arises.

### Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/migrations/025_add_app_id_to_logs.up.sql` | **New** | Add `app_id` columns, backfill, partial indexes |
| `backend/migrations/025_add_app_id_to_logs.down.sql` | **New** | Drop indexes and columns |
| `backend/internal/db/queries/log_buffer.sql` | Edit | 4 new app-scoped queries + `app_id` on INSERT |
| `backend/internal/db/queries/agent_log.sql` | Edit | 2 new app-scoped queries + `app_id` on INSERT |
| `backend/internal/db/log_buffer.sql.go` | Regen | sqlc-generated |
| `backend/internal/db/agent_log.sql.go` | Regen | sqlc-generated |
| `backend/internal/db/models.go` | Regen | `AppID pgtype.UUID` on LogBuffer and AgentLog |
| `backend/internal/db/monitoring.sql.go` | Regen | sqlc-generated |
| `backend/internal/api/handlers/logs.go` | Edit | `app_id` query param, auth check, query routing |
| `backend/internal/api/handlers/logs_test.go` | Edit | 3 new app-scoping test cases |
| `backend/internal/api/handlers/webhooks.go` | Edit | Pass `conn.AppID` to `InsertLogEntry` |
| `backend/internal/api/handlers/otlp.go` | Edit | Pass `conn.AppID` to `InsertLogEntry` |
| `backend/internal/api/handlers/applications.go` | Edit | Pass `nil` app_id on org-level audit emits |
| `backend/internal/api/handlers/connections.go` | Edit | Pass `conn.AppID` to `StartPoller`/`NewSyslog` |
| `backend/internal/api/handlers/connections_validate.go` | Edit | Pass `uuid.Nil` to connector constructors |
| `backend/internal/api/handlers/connections_test_handler.go` | Edit | Pass `conn.AppID` to connector constructors |
| `backend/internal/agent/emit.go` | Edit | `appID *uuid.UUID` parameter on all emit functions |
| `backend/internal/agent/loop.go` | Edit | Derive `appIDPtr` from `appID`, pass through emit calls |
| `backend/internal/agent/monitor.go` | Edit | Pass `&app.ID` to `EmitLogWithSeverity` |
| `backend/internal/agent/scheduler.go` | Edit | Pass `&s.AppID` to `EmitLogWithSeverity` |
| `backend/internal/connectors/factory.go` | Edit | `appID uuid.UUID` parameter on `StartPoller` |
| `backend/internal/connectors/logs/supabase.go` | Edit | `appID` field + constructor param + INSERT |
| `backend/internal/connectors/logs/supabase_test.go` | Edit | Constructor call updated |
| `backend/internal/connectors/logs/flyio.go` | Edit | `appID` field + constructor param + INSERT |
| `backend/internal/connectors/logs/vercel.go` | Edit | `appID` field + constructor param + INSERT |
| `backend/internal/connectors/logs/railway.go` | Edit | `appID` field + constructor param + INSERT |
| `backend/internal/connectors/logs/mongodb.go` | Edit | `appID` field + constructor param + INSERT |
| `backend/internal/connectors/logs/syslog.go` | Edit | `appID` field + constructor param + INSERT |
| `backend/internal/connectors/logs/syslog_test.go` | Edit | Constructor call updated |
| `backend/cmd/heimdall/main.go` | Edit | Pass `conn.AppID` on poller/syslog resume at startup |
| `frontend/src/api/logs.ts` | Edit | `app_id?: string` param |
| `frontend/src/stores/logs.ts` | Edit | Include `currentAppId` in every fetch |
| `frontend/src/pages/ActivityPage.vue` | Edit | Watch `currentAppId`, reset + re-fetch |

---

## 0.36.0 — Top Header Nav Bar & Sidebar Slim-Down (2026-04-13)

The authenticated layout has been restructured from sidebar-first to header-first, inspired by the Supabase dashboard pattern. A full-width top header bar now handles identity context — which organisation, which application, who's logged in — while the sidebar is stripped back to pure page navigation. The "HEIMDALL" wordmark is restyled to match the public homepage's spaced-out monospace treatment, the green status dot and "Status: Active" label are removed, and the app selector moves from a sidebar dropdown into a breadcrumb trail in the header.

### AppHeader — new component

**`AppHeader.vue`** is a 48px (`h-12`) fixed header bar sitting above the sidebar and content area. It follows the Supabase breadcrumb pattern: logo on the left, then org and app context selectors separated by `/` dividers, then a profile avatar on the right.

**Left section — logo + breadcrumbs.** The Heimdall radial-eye icon is inlined as SVG (not an `<img>` tag) so it can use `currentColor` with the `text-accent` class, keeping it consistent with the design token system. Next to it, "H E I M D A L L" is rendered in `font-mono text-xs font-semibold uppercase tracking-[0.25em]` — the same letter-spacing treatment as `PublicNav.vue` on the public homepage. On small screens the wordmark hides (`hidden sm:inline`) to save horizontal space, leaving just the icon.

**Organisation breadcrumb.** A building icon + the org name + chevron. Clicking opens a dropdown showing the current organisation as a highlighted row. Currently single-org per user, but the dropdown is structurally ready for multi-org — when that feature lands, the list just needs populating. The dropdown closes on outside click or when another dropdown opens (all three are mutually exclusive).

**Application breadcrumb.** A database icon + the current app name + chevron. The dropdown lists all applications with the active one highlighted in `accent-bright` on `accent-subtle` background. A border-separated footer row shows "+ New application" which opens the `AppWizard` modal (moved from the sidebar). Selecting an app calls `app.selectApp()`, persists to localStorage, and closes the dropdown.

**Right section — profile.** A 28px circular avatar showing the first letter of the user's email, styled with `border-border bg-bg-surface`. The dropdown shows the email, a "Settings" link (navigates to the settings route), and a "Sign out" button that calls `auth.logout()` + `app.reset()` before redirecting to login.

**Interaction model.** All three dropdowns are mutually exclusive — opening one closes the other two. A document-level click listener dismisses any open dropdown when clicking outside its container ref. Enter/leave transitions use the same `opacity + translate-y` animation as `BaseSelect.vue` for visual consistency.

### AppSidebar — slim-down

The sidebar loses 108 lines (203 → 95) and three of its five sections:

| Removed section | Lines | New location |
|----------------|-------|-------------|
| Brand header (green dot, "Heimdall" text, "Status: Active") | ~10 | Removed — branding in header; status visible on Dashboard |
| App selector (`BaseSelect` + "Application" label) | ~10 | Header app breadcrumb |
| User footer (Settings, org name, email, Sign Out) | ~20 | Header profile dropdown |
| `AppWizard` mount point | ~2 | Header component |
| Related imports and logic (`useAuthStore`, `useAppStore`, `BaseSelect`, `AppWizard`, `useRouter`, `handleSelectApp`, `handleLogout`, `closeWizard`, `appOptions`, `NEW_APP_SENTINEL`) | ~55 | Logic moved to `AppHeader.vue` |

What remains is a clean navigation-only panel: four labelled sections (Overview, Infrastructure, Agent, Intelligence) with active-indicator bars and phosphor glow — unchanged from before. The width narrows from `w-60` (240px) to `w-56` (224px) since the wider app selector dropdown no longer dictates minimum width.

### DefaultLayout — structural change

The layout container changes from `min-h-screen flex` (horizontal, full-page scroll) to `h-screen flex-col overflow-hidden` (vertical, content-area scroll):

```
Before:                           After:
┌──────────┬──────────┐           ┌───────────────────────┐
│ Sidebar  │          │           │     AppHeader (48px)  │
│ (full    │ Content  │           ├────────┬──────────────┤
│  height) │ (whole   │           │Sidebar │              │
│          │  page    │           │(nav    │ Content      │
│          │  scrolls)│           │ only)  │ (scrolls)    │
└──────────┴──────────┘           └────────┴──────────────┘
```

**`AppHeader`** renders as the first child with `flex-shrink-0`. Below it, a `flex-1 min-h-0` row contains the sidebar and content. The content `<main>` gains `overflow-y-auto` so it scrolls independently — this is critical for the Agent Chat page which uses `h-full flex-col` and needs a fixed-height parent.

**Benefits:**
- Header and sidebar stay pinned without `position: fixed` or z-index gymnastics.
- macOS elastic scroll bounce is eliminated — the viewport is locked.
- Agent Chat's flex layout works correctly with a deterministic parent height.

**Mobile adjustment:** The hamburger button moves from `top-4` to `top-14` so it sits below the header rather than overlapping it. The slide-over sidebar overlay is unchanged.

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/common/AppHeader.vue` | **New** | Top nav bar — logo, org breadcrumb, app breadcrumb, profile dropdown (~260 lines) |
| `frontend/src/components/common/AppSidebar.vue` | Edit | Stripped to navigation-only (203 → 95 lines) |
| `frontend/src/layouts/DefaultLayout.vue` | Edit | Header-first layout, `h-screen` viewport, content-area scroll |

---

## 0.35.0 — Technical Retro-Futurism UI Refresh (2026-04-13)

Heimdall's visual identity has been muted since v0.14.4's feldgrau colour theme — the near-black backgrounds and 12%-saturation accent produced a cohesive but flat visual field where nothing truly "popped." This release evolves the aesthetic from techno-brutalist to **technical retro-futurism**: richer phosphor-green accents, mission-control glow effects, and atmospheric textures that make the interface feel like a powered-on control panel rather than a static dark page. No layout, font, or structural changes — the same bones, substantially more visual authority.

### Design tokens — colour palette evolution

**Accent saturation: ~12% → ~25%.** The core identity shift. Feldgrau (`#4d5d53`) becomes phosphor-feldgrau (`#4a7a5c`) — still grey-green, still military, but with enough chroma to register as a deliberate colour signal. The accent now reads as "operational green" rather than "grey that happens to lean green."

**Backgrounds: deeper tier separation.** The four background tiers (`--bg-primary` through `--bg-elevated`) shift to a cool blue-black undertone (`#06080a` base) with wider luminance steps between tiers. Cards and modals now visibly separate from the page canvas.

**Borders: 14% → 18% opacity.** The most impactful single-token change. Structural lines — card edges, dividers, panel borders — are now clearly legible. The grid of panels reads as an instrument rack rather than floating elements.

**Status colours: richer signal.** Warning amber shifts from `#c4a84a` to `#d4a832`, critical red from `#c45a4a` to `#d44a3a`, info blue from `#4a8aae` to `#4a92c4`. Each status colour also gains a matching `--status-*-glow` token at 20% opacity for box-shadow halos on active indicators.

**Action buttons: brighter CTA.** `--action` moves from `#3b8a5a` to `#38a85c` with a brighter hover state (`#42be68`) to maintain separation from the now-richer accent range.

**Text: crisper contrast.** Primary text brightens slightly (`#e8ede9`), secondary text picks up a green undertone (`#8a9e92`), and muted text becomes more saturated (`#4e6556`).

**New token: `--accent-glow`** (`rgba(74,122,92,0.15)`) — the phosphor halo colour used across all glow effects below.

### Glow & luminance effects

**`.glow-pulse`** — a 3-second breathing glow cycle that replaces `animate-pulse` on status indicators. The slower cadence (vs the standard 2s) suits a monitoring interface — systems breathe slowly. Status-specific variants `.glow-pulse-warn` and `.glow-pulse-critical` use their matching status glow tokens.

**`.brand-glow`** — a `text-shadow` that makes the "HEIMDALL" wordmark in the sidebar appear to emit faint phosphor light. Applied alongside the existing pulsing dot, now also using `glow-pulse`.

**`.glow-active` upgrade** — the existing class jumps from `box-shadow: 0 0 15px rgba(77,93,83,0.06)` (barely perceptible) to a visible double-layer phosphor halo using the new `--accent-glow` token.

**Focus ring enhancement** — `:focus-visible` gains a `box-shadow: 0 0 8px var(--accent-glow)` behind the existing outline, making focused elements glow.

**Scrollbar active state** — dragging the scrollbar now shows a brighter thumb with a subtle glow shadow.

### Animation enrichment

**`@keyframes fadeIn` — border flash.** Cards now briefly flash their borders at full accent colour on entry, then fade to the structural border colour — like instruments coming online on a control panel.

**`@keyframes reveal` — phosphor flash.** Agent chat messages "flare" green via `text-shadow` as they appear (0–70% of the animation), then settle to normal — simulating a CRT character being written to screen.

### Atmospheric textures

**`.surface-grid`** — a CSS-only 24×24px coordinate grid via `linear-gradient` at 6% opacity. Applied to the four Dashboard overview cards where there's enough whitespace for the texture to read as atmosphere rather than noise. Intentionally *not* applied to data-dense surfaces like the Activity feed — an earlier iteration used `var(--border)` opacity on the feed and it competed with the log entries.

**`.surface-scanlines`** — a `::after` pseudo-element overlay with 2% opacity horizontal banding (4px pitch), evoking CRT scan lines. Applied to the Agent Chat window. Uses `pointer-events: none` and `border-radius: inherit` to avoid interfering with content or breaking container rounding.

**`.chrome-brackets`** — `::before`/`::after` corner bracket decorations (12×12px, 60% opacity) using the accent colour. Applied to the four Dashboard overview cards. These are the visual grammar of HUD overlays and targeting reticles — two corner marks that frame each panel as a "viewport."

### Component application

**AppSidebar** — brand dot uses `glow-pulse` instead of `animate-pulse`; wordmark gains `.brand-glow`; section labels ("OVERVIEW", "INFRASTRUCTURE", etc.) shift from `text-text-muted` to `text-accent` so they carry the phosphor-green cast; active nav indicator bar changes to `bg-accent-bright` with an inline glow shadow.

**StatusBadge** — each state (active, error, warning) gains a `box-shadow` halo using the matching `--status-*-glow` token. Status dots use `glow-pulse`, `glow-pulse-critical`, or `glow-pulse-warn` instead of static fills.

**LogEntry** — agent entries gain a glow shadow on hover (`hover:shadow-[0_0_12px_var(--accent-glow)]`); `transition-colors` upgraded to `transition-all` to animate the shadow.

**DashboardPage** — four overview cards gain `.chrome-brackets`; monitoring continuous-mode dot uses `glow-pulse`; Recent Activity container no longer uses grid texture (removed after testing — too dense for data rows).

**AgentChatPage** — WebSocket connection status dots use status-specific glow-pulse variants; tool-progress indicator dot uses `glow-pulse`.

**ChatWindow** — container gains `.surface-scanlines`; "Processing" thinking indicator gains `.brand-glow` text-shadow.

### Brand guidelines update

`docs/brand-guidelines.md` updated to reflect the evolved aesthetic: identity description changed from "techno-brutalist" to "technical retro-futurism"; colour philosophy section renamed from "Feldgrau" to "Phosphor-Feldgrau" with saturation evolution explanation; all colour tables updated with new hex values; design principles table replaces "Brutalist restraint" with "Information emits light" (data glows, chrome doesn't) and "Atmospheric depth" (subtle textures create immersion through accumulation).

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/assets/styles/main.css` | Edit | Token update, glow effects, atmospheric textures (~120 lines) |
| `frontend/src/components/common/AppSidebar.vue` | Edit | Brand glow, section labels, nav indicator |
| `frontend/src/components/common/StatusBadge.vue` | Edit | Status glow shadows, pulse animations |
| `frontend/src/components/log/LogEntry.vue` | Edit | Hover glow, transition-all |
| `frontend/src/components/log/LogFeed.vue` | Edit | Reverted grid texture (too noisy on data rows) |
| `frontend/src/pages/DashboardPage.vue` | Edit | Chrome brackets, glow-pulse on monitoring dot |
| `frontend/src/pages/AgentChatPage.vue` | Edit | Status glow-pulse variants |
| `frontend/src/components/agent/ChatWindow.vue` | Edit | Scanlines overlay, brand-glow on thinking state |
| `docs/brand-guidelines.md` | Edit | Palette, principles, identity description |

---

## 0.34.1 — Dev-Mode Background Job Kill-Switch (2026-04-13)

Local development connects to the production database for convenience, but the five background systems — monitoring loop, investigation scheduler, log pruner, pollers, and syslog listeners — were firing against production data from every local `make dev-backend` run. This meant duplicate monitoring cycles, spurious scheduled investigations, and unnecessary API calls competing with the production instance. This release adds a single `DISABLE_BACKGROUND_JOBS` env var that suppresses all background work while keeping the API server, WebSocket chat, and log ingestion fully operational.

### Backend — config flag

**`Config.DisableBackgroundJobs`** (line 29 in `config.go`) — a new boolean field, set to `true` when the `DISABLE_BACKGROUND_JOBS` environment variable is present with any non-empty value. No value parsing or case-sensitivity — if the var exists, background jobs are off.

### Backend — agent gating

**`Agent.Start()` early return** (line 78 in `agent.go`) — when the flag is set, `Start()` logs an info-level message and returns immediately without spawning the three background goroutines (monitor, pruner, scheduler). The `Agent` struct is still fully constructed and registered with the router, so interactive WebSocket chat continues to work — only the autonomous background loops are suppressed.

### Backend — connector gating

**`main.go` conditional resume** (line 97) — `resumePollers()` and `resumeSyslogListeners()` are now wrapped in a `!cfg.DisableBackgroundJobs` guard. These functions restart polling goroutines (Supabase, Fly.io, Vercel, Railway, MongoDB) and syslog TCP listeners for all active connections on startup — skipping them prevents local dev from duplicating the production connector workload.

### Environment files

**`.env`** — `DISABLE_BACKGROUND_JOBS=true` added and active. Since `make dev-backend` loads this file, local runs automatically skip background jobs with no manual step required.

**`.env.example`** — documented the variable (commented out) so new contributors understand the option.

### What still runs locally

| System | Status | Why |
|--------|--------|-----|
| API routes (REST) | Active | Needed for frontend development |
| WebSocket chat | Active | Interactive agent queries are user-initiated, not autonomous |
| Webhook log ingestion | Active | Needed for testing ingestion flows |
| Monitoring loop | **Disabled** | Would duplicate production monitoring cycles |
| Investigation scheduler | **Disabled** | Would fire scheduled prompts against production |
| Log buffer pruner | **Disabled** | Would delete production log_buffer rows |
| Pollers (Supabase, Fly, etc.) | **Disabled** | Would duplicate production polling and hit rate limits |
| Syslog listeners | **Disabled** | Would bind ports and ingest duplicates |

### Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/config/config.go` | Edit | Added `DisableBackgroundJobs bool` field + env var read |
| `backend/internal/agent/agent.go` | Edit | Early return in `Start()` when flag is set |
| `backend/cmd/heimdall/main.go` | Edit | Gated `resumePollers` / `resumeSyslogListeners` |
| `.env` | Edit | Added `DISABLE_BACKGROUND_JOBS=true` |
| `.env.example` | Edit | Documented the variable (commented out) |

---

## 0.34.0 — Activity Detail Modal & Supabase Poller Tuning (2026-04-13)

The Activity feed previously truncated long entries at ~200 characters, hiding critical detail — agent assessments, root-cause analysis, and structured payload data were cut off with no way to expand them. This release adds a click-to-expand detail modal and reduces the Supabase poller frequency to avoid Management API rate limiting.

### Frontend — Activity detail modal

**`LogDetailModal.vue`** — clicking any entry in the Activity feed now opens a full-screen modal showing the complete, untruncated content. The modal follows the existing design language (backdrop blur, `bg-bg-surface` panel, Escape-to-close) established by `ConnectionTestModal` and `DeleteAppModal`.

**Modal sections:**
- **Header** — source badge (Agent / Monitor / Heartbeat / Raw) and severity badge (critical / warning / info), matching the colour coding used in the feed
- **Metadata grid** — timestamp, source type label, connection name (resolved from the connections store to a human-readable name rather than a raw UUID), and entry ID
- **Summary** — the full untruncated summary text, rendered with `whitespace-pre-wrap` for multi-line agent assessments
- **Detail / Payload** — iterates over the `detail` JSON object's top-level keys, rendering each as a labelled block. Scalar values display inline; nested objects are pretty-printed as JSON with scroll overflow. For raw logs this shows the full ingested payload; for agent logs it shows structured fields like `app_name`, `assessment`, `auto_severity`, `schedule_name`, etc.

**`LogEntry.vue` now clickable** — added `cursor-pointer`, a hover state for agent-sourced entries (`hover:bg-accent-subtle/80`), and a `@click` → `emit('select', entry)` event. The visual affordance signals interactivity without adding explicit buttons.

**`LogFeed.vue` wiring** — a `selectedEntry` ref tracks which entry is expanded. The `@select` event from `LogEntry` opens the modal; `@close` from the modal clears it. The `connections` prop (already passed from `ActivityPage`) is forwarded to the modal for name resolution.

### Backend — Supabase poller interval increase

**Default interval: 30s → 120s, minimum: 15s → 60s** — the Supabase Management API enforces a rate limit of ~60 requests per minute per project, shared across all consumers (Heimdall, the dashboard UI, CLI, CI scripts). With 6 log tables polled per cycle, the previous 30-second default produced 12 API calls per minute. Under concurrent dashboard usage, this consistently exhausted the quota and left the poller in a perpetual 429 loop — the Production Supabase connection for Elephantasm had been rate-limited for 2+ days with zero logs ingested.

The new 120-second default produces ~3 calls per minute (5% of the quota), leaving ample headroom. The 60-second minimum prevents users from configuring intervals that would re-trigger the problem. Existing connections with `poll_interval_secs` below the new minimum are clamped to the default at startup — the `NewSupabase()` constructor enforces the floor on every instantiation, not just on creation.

### Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/log/LogDetailModal.vue` | New | Activity detail modal (148 lines) |
| `frontend/src/components/log/LogEntry.vue` | Edit | Added click handler, cursor-pointer, hover state |
| `frontend/src/components/log/LogFeed.vue` | Edit | Wired modal via `selectedEntry` ref, forwarded connections |
| `backend/internal/connectors/logs/supabase.go` | Edit | Default interval 30→120s, minimum 15→60s |
| `backend/internal/connectors/logs/supabase_test.go` | Edit | Updated test fixtures and assertions to match new intervals |

---

## 0.33.1 — Code Quality & Structural Cleanup (2026-04-12)

A code assessment flagged eight issues across the codebase. This release addresses six of them — the structural and correctness fixes that carry zero behavioural change but reduce maintenance risk and improve observability.

### Backend — connections handler refactor

**`connections.go` split into three files** — the handler file had grown to 693 lines by mixing CRUD, validation, and connectivity testing. It's now three cohesive files:

| File | Lines | Responsibility |
|------|-------|----------------|
| `connections.go` | 404 | CRUD handlers (List, Get, Create, Update, Delete) |
| `connections_validate.go` | 80 | `validateConnectorConfig()` + input validation helpers |
| `connections_test_handler.go` | 190 | `TestConnection` handler (real I/O with timeouts) |

**Unified validation** — `CreateConnection` and `UpdateConnection` both inlined ~60 lines of identical connector validation (6 types + syslog TLS cert injection). Both now call a single `s.validateConnectorConfig(type, config)` method. The method lives on `*Server` because it needs `s.Config` for syslog TLS certs.

**Silenced DB errors now logged** — four `_ = queries.UpdateConnectionStatus(...)` calls in the syslog listener error paths (Create and Update) were silently discarding database errors. Replaced with `slog.Error` calls including `connection_id` for traceability.

### Backend — missing FK index

**Migration 024** — `notification_log.agent_log_id` had a foreign key constraint (`REFERENCES agent_log(id) ON DELETE SET NULL`) but no supporting index. Migration 021 added FK indexes for `notification_log(channel_id)` and `agent_log(conversation_id)` but missed this one. Without the index, every `DELETE` on `agent_log` triggered a sequential scan on `notification_log` — increasingly expensive during log retention cleanup. `IF NOT EXISTS` / `IF EXISTS` guards for idempotency.

### Frontend — NotificationsPage decomposition

**541 → 102 lines** — the page mixed preferences, channels (full CRUD + test), and notification history in a single component with 15 refs and 8 API imports. Now decomposed into three sub-components:

| Component | Lines | Responsibility |
|-----------|-------|----------------|
| `NotificationPreferences.vue` | 151 | Display/edit form, save logic |
| `NotificationChannels.vue` | 294 | Channel list, add/edit/delete, test button |
| `NotificationHistory.vue` | 58 | Read-only log with severity/status indicators |

The page remains the data-fetching shell — `Promise.all` parallel fetch preserved, app-switch state reset via `defineExpose` + template refs.

### Frontend — shared `extractApiError` utility

**2 duplicate definitions removed, 13 catch blocks fixed** — the same error-extraction logic existed independently in `stores/schedules.ts` and `pages/SchedulesPage.vue`, with inline variants across 15+ files. A single `extractApiError(e, fallback)` utility now handles all cases, checking `response.data.error` → `response.data.message` → `error.message` → fallback.

Every `catch (e: any)` in the frontend (11 occurrences across stores, pages, and components) was replaced with `catch (e: unknown)` + the shared utility — restoring TypeScript's type safety at error boundaries.

### Frontend — store encapsulation

**`ConnectionsPage.vue` no longer mutates store state directly** — the page was setting `store.loading`, `store.error`, and `store.connections` inline instead of calling a store action. A new `fetchConnectionsByApp(appId)` action on the connections store encapsulates the loading/error/fetch cycle.

### Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/api/handlers/connections.go` | Edit | Extracted validation, logged silent errors (693→404 lines) |
| `backend/internal/api/handlers/connections_validate.go` | New | `validateConnectorConfig` + validation helpers (80 lines) |
| `backend/internal/api/handlers/connections_test_handler.go` | New | `TestConnection` handler (190 lines) |
| `backend/migrations/024_notification_log_agent_log_idx.up.sql` | New | FK index on `notification_log(agent_log_id)` |
| `backend/migrations/024_notification_log_agent_log_idx.down.sql` | New | Rollback |
| `frontend/src/utils/apiError.ts` | New | Shared `extractApiError` utility |
| `frontend/src/stores/connections.ts` | Edit | +`fetchConnectionsByApp` action, `catch (e: unknown)` |
| `frontend/src/stores/logs.ts` | Edit | `catch (e: unknown)` + `extractApiError` |
| `frontend/src/stores/schedules.ts` | Edit | Removed local duplicate, imported shared utility |
| `frontend/src/pages/NotificationsPage.vue` | Edit | Decomposed (541→102 lines) |
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Delegates to store action, `catch (e: unknown)` |
| `frontend/src/pages/LoginPage.vue` | Edit | `catch (e: unknown)` + `extractApiError` |
| `frontend/src/pages/OnboardingPage.vue` | Edit | `catch (e: unknown)` + `extractApiError` |
| `frontend/src/pages/SchedulesPage.vue` | Edit | Removed local duplicate, imported shared utility |
| `frontend/src/components/notifications/NotificationPreferences.vue` | New | Preferences sub-component |
| `frontend/src/components/notifications/NotificationChannels.vue` | New | Channels sub-component |
| `frontend/src/components/notifications/NotificationHistory.vue` | New | History sub-component |
| `frontend/src/components/connections/ConnectionTestModal.vue` | Edit | `catch (e: unknown)` + `extractApiError` |
| `frontend/src/components/connections/ConnectionForm.vue` | Edit | `catch (e: unknown)` + `extractApiError` |
| `frontend/src/components/connections/GitHubRepoSelector.vue` | Edit | `catch (e: unknown)` + `extractApiError` |

---

## 0.33.0 — Expanded Model Catalogue & Picker (2026-04-12)

The model selection experience has been completely rethought. Previously, Agent Configuration showed 9 models in a native `<select>` dropdown with names crammed into single-line labels. This release expands the catalogue to 22 curated models across 10 vendors and 4 tiers, replaces the dropdown with a rich combobox picker, and adds server-side validation so typos are caught at the API boundary.

### Backend — enriched catalogue, validation, refresh tooling

**`ModelOption` struct enrichment** — four new fields: `vendor` (who made the model), `tier` (flagship/balanced/economy/specialist), `strengths` (short tags like "reasoning", "coding"), and `description` (one sentence). All additive — the API response shape is backwards-compatible.

**Expanded catalogue** — 3 Anthropic-direct + 19 OpenRouter = 22 models, covering Anthropic, OpenAI, Google, xAI, Qwen, Z.ai (GLM), MiniMax, Xiaomi, Moonshot, and Mistral. Each entry has a `tool_use_verified_at` date comment. DeepSeek models were proposed but dropped — not available on OpenRouter as of April 2026.

**`agent.ResolveModel(id)` helper** — scans both catalogues. Returns the full `ModelOption` and true if found.

**Server-side model validation** — `UpdateAppAgentConfig` now rejects unknown model IDs (`400 "unknown model: <id>"`) and provider/model mismatches (`400 "model <id> belongs to provider <x>, not <y>"`). Legacy config rows are unaffected on read — validation only fires on save.

**Catalogue refresh tool** — new `backend/cmd/refresh-models/` CLI. `--diff` compares curated prices/context against live OpenRouter data. `--suggest` lists uncurated models by vendor. Not wired into CI — developer-operated.

### Frontend — ModelPicker combobox

**`ModelPicker.vue`** replaces the native `<select>` with a custom combobox:
- **Search** — substring match across model name, ID, vendor, and strength tags
- **Tier filter chips** — All / Flagship / Balanced / Economy / Specialist
- **Vendor grouping** — collapsible section headers, sticky during scroll
- **Model cards** — three lines per entry: name + tier badge + provider badge, description, pricing + strength tags
- **Keyboard** — arrow keys navigate, Enter selects, Escape closes
- **Layout safety** — trigger width fixed (long names truncate), panel capped at `min(70vh, 32rem)`, no page overflow

**Display-mode enrichment** — the read-only config row now shows the model's human name and tier badge instead of just the raw ID.

### Tests

**Backend:** `TestCatalogueConsistency` (22 models × all-fields-populated), `TestCatalogueSize`, `TestDefaultModelInCatalogue`, `TestResolveModel` (4 cases), 4 new handler validation tests. Build-tagged `tool_use_vet_test.go` for live API vetting.

**Frontend:** 15 vitest cases for ModelPicker — rendering, search filtering, tier filtering, selection, keyboard navigation, empty state, overflow.

### Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/agent/models.go` | Edit | Enriched struct, vendor/tier constants, 9→22 models, `ResolveModel` helper |
| `backend/internal/agent/models_test.go` | New | Catalogue consistency + ResolveModel tests |
| `backend/internal/agent/tool_use_vet_test.go` | New | Build-tagged tool-use vetting test |
| `backend/internal/api/handlers/applications.go` | Edit | Model + provider/model validation |
| `backend/internal/api/handlers/applications_test.go` | Edit | +4 validation tests |
| `backend/cmd/refresh-models/main.go` | New | Catalogue drift detector CLI |
| `backend/cmd/refresh-models/README.md` | New | Usage notes |
| `frontend/src/types/models.ts` | Edit | +4 enriched fields |
| `frontend/src/components/agent/ModelPicker.vue` | New | Combobox picker |
| `frontend/src/components/agent/ModelCard.vue` | New | Per-model card |
| `frontend/src/components/agent/__tests__/ModelPicker.test.ts` | New | 15 vitest cases |
| `frontend/src/pages/AgentConfigPage.vue` | Edit | Swapped `<select>` for `<ModelPicker>`, enriched display row |
| `docs/archive/openrouter-models-2026-04-12.json` | New | Frozen API snapshot |

---

## 0.32.0 — Multi-App Setup & Settings (2026-04-12)

Heimdall now treats applications as first-class citizens that users can create and manage without touching the API directly. Previously the sidebar app selector was read-only after onboarding — adding or removing apps required backend calls. This release ships the full create/delete lifecycle across four phases: a backend surface, a frontend wizard, a Settings page, and a polish pass.

### Backend — delete handler, counts endpoint, activity telemetry

**`DELETE /api/apps/{appId}`** — new handler wrapping the existing `DeleteApplication` sqlc query. Enforces three things:

1. **Org-scope authorization** via `GetApplicationByOrgUser` — cross-org attempts get `404` (existence-hiding, matching every other `/api/apps/{appId}/*` route).
2. **Last-app guard** — `CountApplicationsByOrg` checks the org has more than one app. If not, returns `409 Conflict` with `{"error": "cannot delete last application", "code": "last_app"}`. The `code` field gives the frontend a stable key to render specific helper text instead of string-matching on the message.
3. **Cascade delete** — the handler stays thin because every child table (`connections`, `app_agent_config`, `monitoring_state`, `notification_*`, `investigation_schedules`) was already declared `ON DELETE CASCADE` in migrations 014–019 and 023.

**`GET /api/apps?include=counts`** — opt-in enriched shape. When the query param is set, the handler routes to `ListApplicationsByOrgWithCounts`, which augments each row with `connection_count` and `schedule_count` via two correlated subqueries. The default shape (no param) is unchanged — sidebar selector and other lean consumers keep their cheap path. Only the Settings page opts in.

**`jsonErrorWithCode` helper** — new function in `helpers.go` alongside the existing `jsonError`. Writes `{"error": "...", "code": "..."}` so the frontend can key off a stable machine-readable identifier for any guarded failure that needs specific UI rendering.

**Activity-feed telemetry** — both `CreateApplication` and `DeleteApplication` now fire-and-forget `application_created` / `application_deleted` entries into `agent_log` via `agent.EmitLog`. The delete event is emitted *after* the query succeeds so we never record a deletion that didn't happen. Entries are user-scoped (not app-scoped) with the app identity in the `detail` JSONB column, so `application_deleted` audit rows survive the cascade that destroys their own app — the whole point of an audit trail.

### Backend — new sqlc queries

Two new queries in `backend/internal/db/queries/applications.sql`:

- **`ListApplicationsByOrgWithCounts`** — `SELECT a.*, (SELECT COUNT(*) FROM connections WHERE app_id = a.id) AS connection_count, (SELECT COUNT(*) FROM investigation_schedules WHERE app_id = a.id) AS schedule_count` — cost linear in `#apps`, never fanning out across joins.
- **`CountApplicationsByOrg`** — `SELECT COUNT(*) FROM applications WHERE org_id = $1` — O(1) on an indexed `org_id`, used by the last-app guard without pulling every row back to Go.

No migration. Every schema dependency was already live.

### Backend tests

Six new handler integration tests in `applications_test.go`:

| Test | Guards |
|------|--------|
| `TestDeleteApplication` | Happy path: two apps → delete one → 204, list returns one |
| `TestDeleteApplication_LastAppGuard` | Only app → 409 with `code: "last_app"`, app still exists |
| `TestDeleteApplication_WrongOrg` | Cross-org delete → 404, foreign app untouched |
| `TestDeleteApplication_CascadesToConnections` | Create app + connection → delete app → connection row gone via FK cascade. **Regression tripwire** — catches future migrations that accidentally drop CASCADE |
| `TestListApplications_WithCounts` | `?include=counts` returns `connection_count` and `schedule_count` |
| `TestListApplications_DefaultShapeUnchanged` | Default call must not include count fields |

All tests run against real Postgres (skip cleanly without `DATABASE_URL`).

### Frontend — App Wizard

A new `AppWizard` component (`frontend/src/components/app-wizard/`) accessible from the sidebar app selector's "+ New application" sentinel row. Three visible steps:

1. **Details** — name (required) and optional description. The app is created eagerly at the end of this step via `app.createApp(name, { select: false })` — the `select: false` flag means the draft app doesn't mutate `currentAppId`, so the user's previously-active app stays live if they discard.
2. **Connector** — binary fork: "Add a connector" (hands off to the embedded `ConnectionWizard`) or "Skip for now" (advances to confirmation). The `ConnectionWizard` was refactored to accept an optional `appId` prop (`ConnectionWizard.vue`), with a `targetAppId` computed that resolves `props.appId ?? appStore.currentAppId`. Existing callers pass nothing and inherit zero behaviour change.
3. **Confirm** — success screen with "Go to Dashboard" button that commits the selection (`app.selectApp`) and routes away. A toast surfaces "Application created" via the existing `useToast` composable.

**Discard semantics:** if a draft exists, closing shows a confirmation overlay. On confirm, `app.deleteApp(draftAppId)` rolls back the eager-created row (best-effort — if the API call fails, the wizard still closes and the orphan shows up in Settings). Closing on step 1 (no draft) is a clean no-op.

**Nested-modal handling:** when the connector sub-flow is active, the AppWizard shell hides via `v-else` and `ConnectionWizard` takes over the screen — no stacked backdrops. Keyboard handling: AppWizard's Esc listener explicitly bails during `stepId === 'connector'` so the two wizards don't fight for the same key.

**Sidebar integration** (`AppSidebar.vue`): the app selector appends a `{ value: '__new_app__', label: '+ New application' }` sentinel. `handleSelectApp` intercepts the sentinel *before* calling `selectApp`, so the dropdown visually snaps back to the current app on the next render. `BaseSelect` is one-way (reads `modelValue` from props), so there's no intermediate state flicker.

**Store changes** (`stores/app.ts`):
- `createApp(name, { select })` — new optional flag. Defaults to `true` (existing behaviour preserved); AppWizard passes `false`.
- `deleteApp(appId)` — calls `DELETE /api/apps/{appId}`, filters the row out of `applications`, reconciles `currentAppId` by falling back to the first remaining app (or clearing it entirely if the list is empty).

### Frontend — Settings page

New `/settings` route (`frontend/src/pages/SettingsPage.vue`) linked from the sidebar footer, containing three sections:

**Profile** (`ProfileSection.vue`) — email and joined date from the existing Supabase session (no extra fetch), plus a duplicated "Sign out" button (users looking for account actions naturally visit Settings first).

**Organisation** (`OrganisationSection.vue`) — org name, slug, and created date. Member count deliberately skipped: no members table exists yet, and displaying a hardcoded "1" would be misleading. Inline comment explains the deferral.

**Applications** (`ApplicationsSection.vue`) — the interaction core. Fetches `GET /api/apps?include=counts` on mount into a **local ref** (not the app store — the store's `applications` list is the lean variant every other consumer uses). Per-row:

- Name, status badge, formatted created date.
- **Click-through counts** — "N connections" and "M schedules" are buttons that select the target app and route to the corresponding per-app page. Selecting first ensures the destination renders data for the row the user clicked, not whatever was selected before.
- **Delete button** — disabled when `rows.length <= 1`, with an explicit visible helper line: *"Your organisation must have at least one application. Create another first."* Not a tooltip — tooltips are invisible on touch.
- **"+ New application"** button opens the same `AppWizard` from the sidebar. On wizard close the section refetches unconditionally.

**DeleteAppModal** (`DeleteAppModal.vue`) — two-stage destructive action:

- **Stage 1 (confirm):** itemised list of what will cascade-delete (N connections, M schedules, all logs, all reports, all conversation history). User must type the exact app name (case-sensitive, no trimming) before the Delete button enables. Auto-focused input via `useTemplateRef` + `nextTick`. Errors surface inline without closing the modal.
- **Stage 2 (deleted):** calm success screen with a single "Close" button. The user sees explicit feedback that the destructive action landed. Parent refetches on close.

### Frontend — keyboard & toast polish

- **Enter advances** in both `StepAppDetails` (emits a semantic `submit` event with its own validity check) and `DeleteAppModal` (fires `handleDelete`, which guards on `canDelete && !deleting`).
- **Toast on app creation** via the existing `useToast` composable. The wizard captures `committedName` before clearing state so there's no stale-closure hazard. Delete flow intentionally does *not* toast — the `DeleteAppModal` has its own stage-2 acknowledgement, which is more explicit.

### Frontend tests

Six new vitest store tests in `stores/__tests__/app.test.ts`:

| Test | Guards |
|------|--------|
| `createApp selects the new app by default` | Pre-wizard behaviour preserved |
| `createApp with { select: false } leaves currentAppId untouched` | Eager-create contract |
| `deleteApp removes the row from the list` | Happy path |
| `deleteApp falls back to the first remaining app when current is deleted` | Sidebar never points at a dead row |
| `deleteApp clears currentAppId when no apps remain` | Defined state on edge case |
| `deleteApp propagates backend errors without touching local state` | No half-applied mutation on failure |

### Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/api/handlers/applications.go` | Edit | +`DeleteApplication`, `ListApplications` branches on `?include=counts`, +`EmitLog` telemetry |
| `backend/internal/api/handlers/helpers.go` | Edit | +`jsonErrorWithCode` |
| `backend/internal/api/router.go` | Edit | +`r.Delete("/", s.DeleteApplication)` |
| `backend/internal/api/handlers/testhelpers_test.go` | Edit | +mirrored DELETE route in test router |
| `backend/internal/api/handlers/applications_test.go` | Edit | +6 integration tests |
| `backend/internal/db/queries/applications.sql` | Edit | +`ListApplicationsByOrgWithCounts`, +`CountApplicationsByOrg` |
| `backend/internal/db/applications.sql.go` | **Regen** | sqlc regenerated |
| `frontend/src/types/organization.ts` | Edit | +`ApplicationWithCounts` interface |
| `frontend/src/api/applications.ts` | Edit | +`deleteApplication`, +`listApplicationsWithCounts` |
| `frontend/src/stores/app.ts` | Edit | `createApp` gains `{ select }` option, +`deleteApp`, -unused imports |
| `frontend/src/stores/__tests__/app.test.ts` | **New** | 6 store tests |
| `frontend/src/components/app-wizard/AppWizard.vue` | **New** | Root wizard, step orchestration, discard-rollback, toast |
| `frontend/src/components/app-wizard/steps/StepAppDetails.vue` | **New** | Step 1 — name + description |
| `frontend/src/components/app-wizard/steps/StepConnectorChoice.vue` | **New** | Step 2 — add/skip fork |
| `frontend/src/components/app-wizard/steps/StepConfirm.vue` | **New** | Step 3 — success screen |
| `frontend/src/pages/SettingsPage.vue` | **New** | Page shell with three sections |
| `frontend/src/components/settings/ProfileSection.vue` | **New** | Email, joined, sign-out |
| `frontend/src/components/settings/OrganisationSection.vue` | **New** | Org name, slug, created |
| `frontend/src/components/settings/ApplicationsSection.vue` | **New** | App list + counts + delete + wizard |
| `frontend/src/components/settings/DeleteAppModal.vue` | **New** | Two-stage typed-to-confirm |
| `frontend/src/components/common/AppSidebar.vue` | Edit | +sentinel option, +wizard mount, +Settings footer link |
| `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Edit | +optional `appId` prop, `targetAppId` computed, -unused `isFirstStep` |
| `frontend/src/router/index.ts` | Edit | +`/settings` route |
| `docs/vision.md` | Edit | +Applications section, +Settings platform section |

### Validation

- `go vet ./...` — clean
- `go build ./...` — clean
- `go test -count=1 ./internal/api/handlers ./internal/agent` — all pass uncached against real Postgres
- `vue-tsc -b --noEmit` — clean
- `npm run test -- --run` — 35/35 pass across 6 suites
- `eslint` on all touched files — zero issues
- `npm run build` — clean, 990ms. `SettingsPage` chunk: 13.37 kB (3.99 kB gzipped)

### Deployment notes

**No migration required.** Every schema dependency (CASCADE FKs, `applications` table, `investigation_schedules.app_id`) was already live from migrations 014–023. This release is a pure code deployment.

**Existing applications** continue to work unchanged. The sidebar selector gains the "+ New application" sentinel row; the `/settings` route is additive. No breaking changes to any API shape — the default `GET /api/apps` response is unchanged, and the enriched `?include=counts` variant is opt-in.

---

## 0.31.0 — Scheduled Investigations (2026-04-11)

A third agent operating mode ships in this release. Alongside **interactive** (WebSocket chat) and **monitoring** (classifier-flagged log escalation), the agent can now run **scheduled investigations** — stored prompts that fire on a cron schedule and write their output to the Activity feed. This is the first agent mode that can *proactively probe* pull-only connectors (Postgres, GitHub) without a user in the loop.

The Phase 3 alpha (v0.31.0-alpha, API-only, integer interval) shipped internally on the same day. This release (v0.31.0) layers real cron parsing and a frontend UI on top.

### Backend — cron parser swap

- **Dependency:** `github.com/robfig/cron/v3@v3.0.1`. De-facto Go cron parser, two years of stable releases, no transitive deps. Five-field expressions (minute hour day month weekday) — matches what users expect from `crontab`.
- **`backend/internal/agent/scheduler.go`:** `shouldFire` now prefers `cron_expr` when set and falls back to `interval_secs` otherwise. A malformed `cron_expr` is logged and returns `false` — we'd rather have a stuck schedule the user can fix via the UI than an infinite error loop firing the same broken expression every tick.
- **New exported helper `agent.ParseCronExpression(expr)`** — a thin wrapper around the package-level `cronParser` so the HTTP handler can validate expressions at create/update time without reaching into the scheduler internals.
- **`backend/internal/db/queries/investigation_schedules.sql`:** `CreateSchedule` and `UpdateSchedule` now take `cron_expr` as a nullable `pgtype.Text` param alongside `interval_secs`. Regenerated by sqlc. No migration — the `cron_expr` column was added in migration 023 (Phase 3) specifically so Phase 4 would be non-migrating.
- **`backend/internal/api/handlers/investigation_schedules.go`:** The request body now accepts either `interval_secs` *or* `cron_expr` (exactly-one-of, 400 on both or neither). When `cron_expr` is set, the handler parses it via `agent.ParseCronExpression` and returns 400 with the parser's error text so users see "invalid cron_expr: expected 5 fields, got 3" instead of a generic failure. A new `toDBScheduleParams` helper zeroes `interval_secs` when `cron_expr` wins, so the stored row is never ambiguous even though `shouldFire` is defensively designed to tolerate both.
- **`maxCronExprLen = 200`** — sanity cap on the cron field length. A well-formed expression is tens of bytes; anything longer is either pathological or an attempt to abuse the DB.

### Backend tests

Six new scheduler unit tests cover the cron path:

- `TestShouldFire_CronDue` — past-next-slot fires
- `TestShouldFire_CronNotYetDue` — future-next-slot doesn't
- `TestShouldFire_CronPrefersCronOverInterval` — **precedence tripwire** for the "cron wins when set" rule
- `TestShouldFire_CronInvalidExpression` — malformed expressions return `false` without panic
- `TestShouldFire_CronNeverRun` — never-run schedules short-circuit before the parser
- `TestParseCronExpression_Valid` — sanity check on the exported wrapper

Four new handler test cases cover the validation surface:

- Both `interval_secs` and `cron_expr` set → 400
- Neither set → 400
- Invalid cron expression → 400 with parser error text
- Cron expression exceeding max length → 400

Plus two happy-path handler tests: `TestCreateSchedule_CronMode` and `TestUpdateSchedule_IntervalToCron` (verifies mode-flipping cleanly zeroes the inactive field).

### Frontend — new `/schedules` page

- **`frontend/src/pages/SchedulesPage.vue`** — full CRUD UI for schedules, app-scoped via `useAppStore().currentAppId`. List view with cards, empty state, loading skeleton, error banner.
- **`frontend/src/components/schedules/ScheduleCard.vue`** — per-schedule row showing name, cadence, status badge, last-run timestamp, prompt preview, last summary or last error, and Run-now / Edit / Delete buttons.
- **`frontend/src/components/schedules/ScheduleModal.vue`** — create/edit modal with three scheduling modes: **Preset** (a row of one-click cron preset buttons), **Custom cron** (raw text input with backend-side validation), and **Interval** (legacy seconds input kept so Phase-3 schedules remain editable).
- **`frontend/src/stores/schedules.ts`** — Pinia setup-style store with `fetchSchedules`, `createSchedule`, `updateSchedule`, `deleteSchedule`, and `runNow`. Every action takes `appId` explicitly — no hidden stale state on app switch.
- **`frontend/src/api/schedules.ts`** — axios client wrapping the 5 REST endpoints. Notable: the store uses `client.patch` for updates, which required adding `patch: vi.fn()` to the global test mock in `src/test/setup.ts`.
- **`frontend/src/utils/cron-presets.ts`** — preset catalogue + `humanizeCron` + `formatInterval` + `formatSchedule` (the single entry point for rendering a schedule's cadence regardless of mode). **No dependency on `cronstrue`** — 20KB for rendering five presets is overkill; raw cron is the right affordance for users who opted into advanced mode.
- **`frontend/src/types/schedule.ts`** — `InvestigationSchedule` and `ScheduleInput` types mirroring the Go structs.
- **Routing:** new `/schedules` route lazy-loaded by the router, added to `AppSidebar.vue` under the **Agent** section next to Configuration, Chat, and Notifications.
- **Frontend tests:** six new store tests (`src/stores/__tests__/schedules.test.ts`) covering fetch/create/update/delete/runNow plus the `runningId` mid-flight state used by the card's per-row button disabling.

### UX notes

- **Three scheduling modes in the modal, not two.** The plan originally specified two modes (Preset presets + Custom cron), but I added a third **Interval** mode specifically so Phase-3 interval-only schedules remain editable without being force-migrated to cron on save. Users who never touch the Advanced toggle never see it change their data shape.
- **Preset matching.** When editing a schedule whose `cron_expr` matches one of the built-in presets, the modal keeps the user in Preset mode (friendlier). Custom expressions drop into Custom mode pre-filled.
- **"Run now" uses a per-row `runningId`**, not a global spinner — the store tracks which schedule is mid-`runNow` call so a card can disable its own button without freezing the list.
- **Native `window.confirm` for delete.** A custom confirmation modal would be over-engineering for a destructive action users can trivially redo by recreating a schedule.

### Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/go.mod`, `backend/go.sum` | Edit | Added `github.com/robfig/cron/v3@v3.0.1` |
| `backend/internal/agent/scheduler.go` | Edit | `shouldFire` uses cron parser; `cronParser` + `ParseCronExpression` exported |
| `backend/internal/agent/scheduler_test.go` | Edit | +6 cron-path unit tests, `cronText` helper |
| `backend/internal/db/queries/investigation_schedules.sql` | Edit | Create/Update take `cron_expr` |
| `backend/internal/db/investigation_schedules.sql.go` | **Regen** | sqlc regenerated |
| `backend/internal/api/handlers/investigation_schedules.go` | Edit | `cron_expr` validation + `toDBScheduleParams` helper |
| `backend/internal/api/handlers/investigation_schedules_test.go` | Edit | +6 cases: cron happy/error + exactly-one-of + mode flip |
| `frontend/src/types/schedule.ts` | **New** | `InvestigationSchedule`, `ScheduleInput` |
| `frontend/src/api/schedules.ts` | **New** | 5 REST client functions |
| `frontend/src/stores/schedules.ts` | **New** | Pinia store with `runningId` |
| `frontend/src/stores/__tests__/schedules.test.ts` | **New** | 6 vitest store tests |
| `frontend/src/utils/cron-presets.ts` | **New** | Presets + humanizer + formatSchedule |
| `frontend/src/pages/SchedulesPage.vue` | **New** | List page + modal wiring |
| `frontend/src/components/schedules/ScheduleCard.vue` | **New** | Per-schedule card |
| `frontend/src/components/schedules/ScheduleModal.vue` | **New** | Create/edit modal, 3 modes |
| `frontend/src/router/index.ts` | Edit | `/schedules` route |
| `frontend/src/components/common/AppSidebar.vue` | Edit | "Schedules" nav entry under Agent |
| `frontend/src/test/setup.ts` | Edit | Added `patch: vi.fn()` to global mock |
| `docs/vision.md` | Edit | Agent modes list +Scheduled; Platform Sections +Scheduled Investigations |

### Validation

- `cd backend && go vet ./...` — clean
- `cd backend && go build ./...` — clean
- `cd backend && go test ./...` — all packages pass; new cron tests run green (integration tests skip cleanly without `DATABASE_URL`)
- `cd frontend && npm run test -- --run` — 29 tests across 5 files, all pass
- `cd frontend && npx vue-tsc -b --noEmit` — clean
- `cd frontend && npm run build` — builds in ~1s, new `SchedulesPage` chunk is 16.89 kB (4.99 kB gzipped)
- `cd frontend && npx eslint <Phase 4 files>` — zero lint errors across all new/edited files

**Not validated in this pass:** Live end-to-end cron fire against a real LLM. The validation gate in the plan says "Create a schedule via the UI with `*/5 * * * *`, observe it fire within 5 minutes." This requires a running backend with `DATABASE_URL` + Anthropic key + an app whose scheduler goroutine can actually reach Claude. Manual smoke test once this is on a staging environment.

### Deployment notes

**No migration needed.** The `cron_expr` column shipped in migration 023 (Phase 3) and is already live wherever Phase 3 was deployed. This is the payoff for leaving `cron_expr` in the Phase 3 schema even though it wasn't used yet — Phase 4 is a pure-code change.

**Existing Phase 3 schedules** (interval-only) keep working without migration. The `shouldFire` function falls back to `interval_secs` whenever `cron_expr` is `NULL`, and the frontend's Interval mode lets users edit them without forcing a switch to cron.

**Frontend test setup change.** `patch: vi.fn()` was added to the global axios mock in `src/test/setup.ts` because the schedules store is the first in the codebase to use PATCH. If any other store adds PATCH calls in the future, they'll inherit this mock for free.

---

## 0.30.3 — Activity Feed Rename & Log Retention (2026-04-11)

Two small but high-visibility fixes to the logs experience, shipped as one release.

### Activity feed rename

The page formerly known as **Agent Log** has been renamed to **Activity** and promoted out of the Agent sidebar section into Overview, next to Dashboard. The page itself has always been a *unified* feed — raw logs from connectors (webhooks, Supabase, OTLP, syslog) plus agent observations — but the old name implied it only showed agent output, and the nesting under an "Agent" section compounded the mental-model mismatch. Users were missing their own raw logs because they didn't think to look on a page named "Agent Log" under an "Agent" sidebar heading.

- **Route:** `/agent/log` → `/activity`. The old path is kept as a redirect entry in `frontend/src/router/index.ts` so existing bookmarks and in-product links continue to work. Safe to delete after a grace period.
- **Component:** `frontend/src/pages/AgentLogPage.vue` → `frontend/src/pages/ActivityPage.vue`. Header text changed from "Agent Log" → "Activity". Subheader ("Unified chronological feed of all system activity") kept as-is — it was already accurate.
- **Sidebar:** Item moved from the `Agent` section (alongside Configuration/Chat/Notifications) into `Overview` (alongside Dashboard). Label: `Log` → `Activity`.
- **Source filter polish:** In `frontend/src/components/log/LogFilters.vue`, the "Agent activity" dropdown option was renamed to "Agent observations" to avoid a name collision with the page-level "Activity" and to more accurately describe what those rows actually contain (the agent's analyses and assessments, not raw activity).
- **Vision doc:** The `### Agent Log` section and the `Platform Sections` list item in `docs/vision.md` were updated to match.

No data model changes, no backend API changes, no migrations. Pure frontend polish.

### Log retention pruning (`PruneExpiredLogs` was dead code)

The `PruneExpiredLogs` SQL function in `backend/internal/db/queries/log_buffer.sql:52-53` and its sqlc-generated Go wrapper in `backend/internal/db/log_buffer.sql.go:260` existed since at least v0.4.0 but **had zero callers anywhere in the backend**. A grep for `PruneExpiredLogs` turned up only the definition and the generated wrapper. The practical consequence: `log_buffer` had been growing unbounded in production since the webhook log connector shipped.

- **New file:** `backend/internal/agent/pruner.go` — a `(a *Agent) Prune(ctx)` goroutine that runs a startup tick immediately, then ticks every hour. Errors are logged and swallowed (best-effort, must never take down the agent). Exposed as a method so tests can exercise `pruneTick` directly without waiting on the ticker.
- **`backend/internal/agent/agent.go` `Start()` now spawns two goroutines** — `Monitor` and `Prune` — each incrementing `a.wg` so `Stop()` joins both on shutdown. This edit is also load-bearing for Phase 3 of the logs-feed-and-scheduled-investigations plan, which will add a third goroutine (`InvestigationScheduler`) in the same place.
- **Retention window bumped from 24h → 48h** per user request. One-character SQL edit in `log_buffer.sql`; sqlc regenerated cleanly.
- **New tests:** `backend/internal/agent/pruner_test.go` — three tests using the existing `stubDBTX` pattern (error-swallowing, clean cancel, startup-tick-runs). Deliberately scoped to loop behavior rather than real-DB deletion — the SQL correctness is verified by sqlc regeneration, and the agent package has no real-DB test harness to piggyback on (a gap Phase 3 will likely force us to close).

**Deployment caveat:** when this deploys, the startup `pruneTick` runs a `DELETE FROM log_buffer WHERE ingested_at < now() - interval '48 hours'` against production. Since the query has never been called in production history, the first run may delete a meaningful backlog in a single statement. Consider running the DELETE manually at a quiet hour first, or adding a `LIMIT` clause if the backlog is too large for a single transaction.

### Files changed

| File | Change |
|------|--------|
| `backend/internal/db/queries/log_buffer.sql` | Retention interval `'24 hours'` → `'48 hours'` |
| `backend/internal/db/log_buffer.sql.go` | Regenerated by sqlc |
| `backend/internal/agent/pruner.go` | New — `Prune` loop and `pruneTick` helper |
| `backend/internal/agent/agent.go` | `Start()` spawns Monitor + Prune goroutines |
| `backend/internal/agent/pruner_test.go` | New — three unit tests |
| `frontend/src/pages/AgentLogPage.vue` → `ActivityPage.vue` | Renamed, header text updated |
| `frontend/src/router/index.ts` | Route `/agent/log` → `/activity` with legacy redirect |
| `frontend/src/components/common/AppSidebar.vue` | Activity moved to Overview section |
| `frontend/src/components/log/LogFilters.vue` | "Agent activity" → "Agent observations" |
| `docs/vision.md` | `Agent Log` → `Activity Feed` / `Activity` |

---

## 0.30.2 — Production Schema Drift Fix (2026-04-10)

Three database migrations (020–022) were committed and deployed with the application code but never executed against the production Supabase database. This caused a 500 error when updating agent config via the model dropdown — the `UpsertAppAgentConfig` query references the `provider` column (added in migration 022), which didn't exist in production.

**Symptom.** Changing the model on the Agent Config page returned `500` with:

```
ERROR: column "provider" of relation "app_agent_config" does not exist (SQLSTATE 42703)
```

The `GET` endpoint was unaffected because `SELECT *` tolerates missing columns at the sqlc scan level, but `INSERT INTO ... (provider)` fails hard when PostgreSQL doesn't recognise the column name.

**Root cause.** Migrations 020–022 were added across the GitHub App (v0.30.0) and OpenRouter (v0.29.0) releases but `migrate-up` was never run against production after deploy.

**Fix.** Ran `migrate-up` against production. Three migrations applied:

| Migration | Description |
|-----------|-------------|
| 020 | `github_repos` — repo selection table for GitHub App |
| 021 | `rls_missing_tables` — RLS policies for new tables |
| 022 | `agent_config_provider` — adds `provider TEXT NOT NULL DEFAULT 'anthropic'` to `app_agent_config` |

**Prevention.** This class of bug (code deployed ahead of schema) should be caught by a pre-deploy migration check or a CI step that compares the deployed migration version against the database. See `docs/executing/schema-drift-check.md` for a proposed solution.

---

## 0.30.1 — GitHub Connection Flow Fixes (2026-04-10)

Three bugs prevented the GitHub App integration (v0.16.0, provisioned in v0.30.0) from working end-to-end. They shared a root cause: the connection wizard assumed it owns connection creation, but for GitHub the OAuth callback creates the connection server-side, leaving the wizard out of sync.

---

### Phase 1 — Fix callback redirect

**Problem.** After the user installs the GitHub App, GitHub redirects to `/api/github/callback` on the **backend** origin. The handler then issued a relative redirect to `/connections?github=installed`, which resolved to the backend server — not the frontend SPA. The user saw a blank page.

**Fix.** Added a `FRONTEND_URL` config field (`backend/internal/config/config.go`, default `http://localhost:5173`) and changed the redirect in `github.go:240` to use the full frontend URL:

```go
http.Redirect(w, r, s.Config.FrontendURL+"/connections?github=installed", http.StatusFound)
```

**Deployment requirement:** `FRONTEND_URL` must be set in Fly.io secrets before this fix is live in production.

---

### Phase 2 — Prevent wizard from creating empty GitHub connections

**Problem.** The wizard's `finish()` function called `createConnection()` for all flows, including GitHub. This created a connection with `config: {}` (no `installation_id`), which broke `ListGitHubRepos` and `TestGitHubConnection`. The Done button was also incorrectly enabled on GitHub's install step because the disabled condition (`!stepValid && isTestStep`) only gated test steps.

**Fix.** Two changes in `ConnectionWizard.vue`:

1. Added an early return in `finish()` for GitHub flows — emits `close` without calling `createConnection()`.
2. Simplified the Done button's `:disabled` from `!stepValid && isTestStep` to `!stepValid`. If a step says it's invalid, Done is disabled regardless of step type. No regression for other flows — their last steps either emit `valid: true` or are test steps that already controlled `stepValid`. Removed the now-dead `isTestStep` computed.

---

### Phase 3 — Detect existing installation and skip to repo selection

**Problem.** `StepGitHubInstall` always showed "Install GitHub App" even when a GitHub connection already existed. Users who deleted and re-added GitHub were forced through the full OAuth flow again, creating confusion and duplicate connections.

**Fix.** Two changes:

1. `ConnectionWizard.vue` — added an early intercept in `selectPlatform()`: when `flowId === 'github'`, checks `store.connections` for an existing `type === 'github'` connection. If found, emits `manage-repos` with the connection ID and `close` — the wizard closes and the repo selector opens immediately. Added `manage-repos` to the emit signature.
2. `ConnectionsPage.vue` — wired `@manage-repos` on the wizard component to call `closeWizard()` then `openRepoSelector(id)`. Both functions already existed.

---

### Files changed

| File | Change |
|------|--------|
| `backend/internal/config/config.go` | Added `FrontendURL` field (env: `FRONTEND_URL`, default `http://localhost:5173`) |
| `backend/internal/api/handlers/github.go` | Callback redirect uses `s.Config.FrontendURL` prefix |
| `frontend/src/components/connections/wizard/ConnectionWizard.vue` | GitHub early return in `finish()`, simplified Done button gating, existing-install intercept in `selectPlatform()`, `manage-repos` emit |
| `frontend/src/pages/ConnectionsPage.vue` | Wired `@manage-repos` handler on wizard |

---

## 0.30.0 — GitHub App Provisioning (2026-04-10)

The v0.16.0 GitHub App integration was fully implemented in code but never operational — the four required environment variables (`GITHUB_APP_ID`, `GITHUB_CLIENT_ID`, `GITHUB_PRIVATE_KEY`, `GITHUB_APP_SLUG`) were missing from both the local `.env` and Fly.io production secrets. This release provisions the GitHub App and deploys the credentials, making the codebase connector live for the first time.

---

### GitHub App creation

Created **Heimdall Monitoring Agent** (`heimdall-monitoring-agent`) as a GitHub App under the **@kaminocorp** organisation.

| Setting | Value |
|---------|-------|
| **Owner** | @kaminocorp |
| **App name** | Heimdall Monitoring Agent |
| **Slug** | `heimdall-monitoring-agent` |
| **Homepage URL** | `https://heimdall-backend.fly.dev` |
| **Callback URL** | `https://heimdall-backend.fly.dev/api/github/callback` |
| **Webhook** | Inactive (handler not yet implemented) |
| **Installation scope** | Any account |

**Permissions granted:**
- Repository → **Contents**: Read-only (needed for `search_code`, `read_file`, `list_tree`)
- Repository → **Metadata**: Read-only (auto-granted)

No organisation or account permissions. No event subscriptions.

---

### Credentials deployed

Generated an RSA private key via the GitHub App settings page and collected the App ID (`3334949`) and Client ID (`Iv23liymBQCp9EhNYmZW`).

**Local `.env`** — added four new variables:
```
GITHUB_APP_ID=3334949
GITHUB_CLIENT_ID=Iv23liymBQCp9EhNYmZW
GITHUB_APP_SLUG=heimdall-monitoring-agent
GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY----- ... -----END RSA PRIVATE KEY-----"
```

**Fly.io production** — set the same four secrets via `fly secrets set`. Both machines (`683732ef163d68`, `0807de1c140d98`) restarted successfully with rolling deployment. DNS verified.

---

### Fly.io secrets (as of 0.30.0)

```
ANTHROPIC_API_KEY     ✅ Set
DATABASE_URL          ✅ Set
SUPABASE_URL          ✅ Set
OPENROUTER_API_KEY    ✅ Set
GITHUB_APP_ID         ✅ Set (new)
GITHUB_CLIENT_ID      ✅ Set (new)
GITHUB_PRIVATE_KEY    ✅ Set (new)
GITHUB_APP_SLUG       ✅ Set (new)
```

---

### What this unlocks

The GitHub install flow is now operational end-to-end:

1. User navigates to **Connections → Add GitHub** in the Heimdall UI
2. Backend generates a state JWT and redirects to `github.com/apps/heimdall-monitoring-agent/installations/new`
3. User selects repositories and installs the app
4. GitHub redirects back to `/api/github/callback` with `installation_id`
5. A `github` connection is created in the database
6. User enables specific repos via the repo selector
7. Agent can now use the `search_codebase` tool during investigations (`search_code`, `read_file`, `list_tree` actions via the GitHub codebase connector)

---

### No code changes

This release is purely operational — no source files were modified. The GitHub App integration code shipped in v0.16.0 is unchanged. The only file touched in the repo is `.env` (gitignored).

---

## 0.29.1 — OpenRouter Post-Assessment Polish (2026-04-08)

Five low-severity findings from the `docs/plans/openrouter-assessment.md` review of the 0.29.0 OpenRouter integration. None were blocking for production — this release clears the follow-up queue so the next engineer touching the provider layer inherits a clean slate.

Findings are numbered L1–L6 to match the assessment doc. L5 (severity parsing heuristic) was deliberately deferred as monitor-in-production rather than code fix.

---

### L3 — Deduplicate `defaultModelID` (highest-value fix)

**What / where.** The string `"claude-sonnet-4-6"` was hardcoded in **four** places:

1. `backend/internal/agent/loop.go:18` — the `defaultModelID` const the agent loop uses when no config row exists.
2. `backend/internal/api/handlers/applications.go:80` — the `CreateApplication` default that seeds new app_agent_config rows.
3. `backend/internal/api/handlers/applications.go:144` — the `GetAppAgentConfig` fallback returned when the config row is missing.
4. `backend/internal/api/handlers/applications.go:177` — the `UpdateAppAgentConfig` request fallback when the client omits `model`.

**How.** Renamed the loop's private `defaultModelID` to exported `agent.DefaultModelID`, added a package-level doc comment, and swapped all three handler call sites to `agent.DefaultModelID`. The `handlers` package already depends on `internal/db` and `internal/config` but had not previously imported `internal/agent` — verified via `go build ./...` that the new import does not create a cycle (agent depends on db/config, handlers depend on agent/db/config, no back-edge).

**Why.** The real hazard was drift on the next Sonnet release: updating the const in one place would silently leave the three handler call sites pointing at an old model. Every newly-created app would get Sonnet 4.6 while the loop defaulted to 4.7 — the kind of inconsistency that's easy to write and hard to notice. One const, one source of truth.

---

### L2 — OpenRouter `MaxTokens` default parity with Anthropic provider

**What / where.** `backend/internal/agent/provider_openrouter.go` — the `orRequest.MaxTokens` field had `json:"max_tokens,omitempty"` and `buildOpenRouterRequest` passed `params.MaxTokens` through raw.

**How.** Dropped the `omitempty` tag (field is now always emitted) and added the same `if maxTokens == 0 { maxTokens = 4096 }` guard that `provider_anthropic.go:40–42` has. Both providers now substitute 4096 identically when the caller forgets to set it.

**Why.** The agent loop currently always passes `defaultMaxTokens = 4096`, so the divergence was unreachable in 0.29.0 — but it was a time bomb. A future refactor that added a new call site forgetting to set MaxTokens would cause Anthropic to substitute 4096 while OpenRouter quietly fell back to each model's wildly different provider default (Gemini's is much higher than Claude's). Silent behavioural divergence between providers for the same `ChatParams` is exactly the bug class the Phase 1 abstraction was supposed to prevent.

---

### L1 — OpenRouter `Content` field always emitted

**What / where.** `backend/internal/agent/provider_openrouter.go` — the `orMessage.Content` field had `json:"content,omitempty"`.

**How.** Dropped the `omitempty`. Content is now always present in the wire body — empty string for assistant messages that have only `tool_calls`, populated otherwise. Added a code comment above `orMessage` explaining the choice.

**Why.** The OpenAI chat completions spec is ambiguous on whether assistant messages with `tool_calls` may omit `content` or must emit `content: ""` (or `null`). OpenAI itself tolerates omission; OpenRouter proxies to many downstream providers, some of which are stricter. Emitting an explicit empty string is the wire shape that's spec-legal *and* maximally compatible. Zero behavioural impact on the existing provider set; immunity to one class of future-provider-rejects-our-body bug.

---

### L4 — End-to-end integration test for OpenRouter routing

**What / where.** `backend/internal/agent/loop_test.go` — added `TestRunLoop_RoutesThroughOpenRouter`.

**How.** Stands up an `httptest.Server` that returns an OpenAI-shaped response, constructs an `Agent` with `NewOpenRouterProviderWithURL` registered under `defaultProviderName` (so `RunLoop` — which runs with `appID=uuid.Nil` and can't consult `app_agent_config.provider` — routes through it), calls `RunLoop("hello")`, and asserts three things on the captured wire body:

1. **Routing.** The mock server was hit exactly once and returned "routed via openrouter" (proves the loop actually went through `OpenRouterProvider.ChatCompletion`, not `AnthropicProvider`).
2. **Wire shape.** `captured.Messages[0].Role == "system"` with non-empty content (proves the system prompt was translated to first-message form per the Phase 2 divergence, not left as a top-level Anthropic-style field).
3. **MaxTokens emission.** `captured.MaxTokens == 4096` (pins the L2 fix: the field must be on the wire).

**Why.** The pre-0.29.1 test suite covered the provider translator at the field level (`provider_openrouter_test.go`, four tests) and the agent loop's Anthropic path end-to-end (`loop_test.go`, three tests). What was missing was a test that proved the routing itself: that the `providerFor(appCfg.Provider)` indirection in `runConversationCore` actually reaches the right provider, and that the full loop → provider → HTTP → translator chain produces an OpenAI-shaped body. Phase 3 was a two-line change — that's exactly the kind of change that gets silently broken by an unrelated refactor months later. This test is the tripwire.

The test also doubles as a regression guard for L1 and L2: if either of those fixes ever gets reverted, the `captured.MaxTokens == 4096` assertion and the wire-body inspection will fail.

---

### L6 — Frontend `formProvider` derived from `formModel`

**What / where.** `frontend/src/pages/AgentConfigPage.vue`.

**How.** Converted `formProvider` from a `ref` with an imperative `onModelSelect` `@change` handler to a `computed` that derives the provider from the selected model's catalogue entry, falling back to `config.value?.provider ?? 'anthropic'` for legacy-model safety. Dropped the `@change="onModelSelect"` binding from the `<select>`, deleted the `onModelSelect` function, and removed the manual `formProvider.value = ...` assignment from `startEdit`. Net diff: -7 lines.

**Why.** Two-source-of-truth state is a small but real bug magnet. The model ID is the user's actual choice; the provider is a deterministic function of it (look up the model's `provider` in the catalogue). Modelling that as a `computed` instead of a manually-synced `ref` removes an entire class of "did we forget to resync after X" bugs. The fallback order is preserved exactly: selected model's provider → existing config's provider → `'anthropic'`. Save payload is unchanged.

---

### L5 — Severity parsing (deferred, not fixed)

Documented in the assessment as "monitor in production" rather than a code change. `parseSeverityFromResponse` in `loop.go:334` is a prose keyword scan that assumes the Heimdall system prompt's exact "severity: critical" phrasing. OpenRouter is the first time the agent will routinely run on non-Anthropic models that may not emit that marker. Tracked as future work: if non-Anthropic models produce noisy severity ratings in production, move extraction into a structured severity tool call. Shipping 0.29.1 without touching this — the failure mode is "severity is sometimes wrong", not "agent breaks".

---

### Verification

```
cd backend && go build ./...                                  # clean
cd backend && go vet ./...                                    # clean
cd backend && go test ./internal/agent/... ./internal/api/handlers/...  # PASS
cd frontend && npx vue-tsc --noEmit                           # clean
```

All pre-existing tests still pass; the new `TestRunLoop_RoutesThroughOpenRouter` is the only addition. Zero behaviour change for users on Anthropic; OpenRouter users get a more spec-compliant wire body and guaranteed MaxTokens emission.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/loop.go` | `defaultModelID` → exported `DefaultModelID` with doc comment |
| 2 | `backend/internal/agent/provider_anthropic.go` | Comment updated to reference new exported name |
| 3 | `backend/internal/agent/provider_openrouter.go` | `orMessage.Content` no longer `omitempty`; `orRequest.MaxTokens` no longer `omitempty`; `buildOpenRouterRequest` adds Anthropic-parity 4096 default guard |
| 4 | `backend/internal/agent/loop_test.go` | New `TestRunLoop_RoutesThroughOpenRouter`; `io` import added |
| 5 | `backend/internal/api/handlers/applications.go` | New `internal/agent` import; 3 call sites switched to `agent.DefaultModelID` |
| 6 | `frontend/src/pages/AgentConfigPage.vue` | `formProvider` ref → computed; `onModelSelect` handler deleted; `@change` binding removed; `startEdit` no longer manually syncs provider |

---

## 0.29.0 — OpenRouter Multi-Provider Support (2026-04-08)

Decouples the agent from `anthropic-sdk-go` and adds OpenRouter as a second LLM provider, exposing hundreds of models (GPT-4o, Gemini 2.5, Llama 4, direct Claude) through a single integration. End-users pick a model from a grouped dropdown; the backend routes through the right provider based on per-app config. Monitoring stays on the 0.27.0 rate-limiter contract regardless of provider. Shipped in eight phases across the backend, database, and frontend, then hardened twice — once post-integration (Phase 7) and once post-assessment (Phase 8).

Canonical completion docs live in `docs/completions/openrouter-phase-1.md` through `docs/completions/openrouter-phase-8.md`. This entry is the condensed per-phase summary.

---

### Phase 1 — Provider abstraction (pure refactor, zero behaviour change)

The agent loop was speaking `anthropic-sdk-go` types directly in `loop.go`, `tools.go`, and `agent.go`. Adding a second provider would have meant editing all three files and touching every tool registration site — exactly the kind of fan-out that turns a 200-line feature into a 1000-line PR.

Phase 1 introduced a neutral `Provider` interface with a single `ChatCompletion` method and domain types (`ChatParams`, `ChatMessage`, `ContentBlock`, `ToolCall`, `ToolResult`, `ToolDef`, `ChatResponse`, `StopReason`) that the agent loop speaks exclusively. `AnthropicProvider` became the only file in the package that imports `anthropic-sdk-go`. The mapping between neutral and SDK types is mechanical and one-to-one, and the stop-reason collapse (any unknown reason → `EndTurn`) matches the pre-refactor loop's fallthrough exactly.

Streaming was deferred despite being bundled in the original plan — the selection doc explicitly said "Not now, Option A." Shipping the abstraction as a pure refactor first kept Phase 1 easy to review and rollback, and let Phases 2–6 land on top without taking on any frontend risk.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/provider.go` | New — interface + neutral types |
| 2 | `backend/internal/agent/provider_anthropic.go` | New — only file importing `anthropic-sdk-go` |
| 3 | `backend/internal/agent/agent.go` | `client *anthropic.Client` → `providers map[string]Provider`; `providerFor(name)` resolver with fallback to `"anthropic"` |
| 4 | `backend/internal/agent/loop.go` | `RunConversation` / `RunMonitoring` build neutral types; SDK imports removed; `defaultModelID` extracted as const |
| 5 | `backend/internal/agent/tools.go` | `ToolRegistry()` returns `[]ToolDef`; SDK import removed |
| 6 | `backend/internal/agent/loop_test.go` | Tests use `NewAnthropicProviderWithClient` to inject an SDK client wired to the existing `httptest.Server` — same mock server pattern, zero scenario changes |

---

### Phase 2 — `OpenRouterProvider` (raw HTTP, blocking)

The second provider implementation, routed through OpenRouter's OpenAI-compatible `/chat/completions` endpoint. Raw HTTP rather than a second SDK — the surface area was one endpoint, and pulling in `openai-go` would have brought its own retry policies, transport config, and version churn for no benefit.

The translation between neutral types and OpenAI wire format is the riskiest part of the phase. Five concrete divergences from Anthropic, all handled in `buildOpenRouterRequest` / `translateOpenRouterResponse`:

1. **System prompt** becomes the first `role:"system"` message, not a separate top-level field.
2. **Tool schemas** get wrapped in a full JSON Schema `object` (with `type`, `properties`, `required`) inside `function.parameters`, not passed as the inner schema alone.
3. **Tool calls in responses** arrive as a `tool_calls[]` array on the assistant message, not as `tool_use` content blocks — the translator flattens both sides so the agent loop sees a uniform `[]ContentBlock`.
4. **Tool results** are one `role:"tool"` message per result (with its own `tool_call_id`), not a single user message holding multiple `tool_result` blocks.
5. **Stop reason:** `finish_reason:"tool_calls"` → `StopReasonToolUse`; every other value collapses to `StopReasonEndTurn`, matching Phase 1's Anthropic mapping.

Two subtleties worth calling out: OpenAI's `function.arguments` is a *JSON-encoded string*, not an inline object, so the translator round-trips it through `json.RawMessage` so the agent loop's `json.Unmarshal(tu.Input, ...)` works unchanged. And OpenAI has no equivalent for Anthropic's `is_error` flag on tool results, so failed tool responses get a `"ERROR: "` content prefix — pragmatic, and models generally pick up on the convention.

Error handling distinguishes 401 (bad key), 402 (out of credits), 429 (rate limited), and catch-all 5xx, each with a readable message. No silent fallback to Anthropic — the selection plan explicitly said failing loudly is better than "why is my GPT-4o app responding like Claude."

Four `httptest`-backed tests pin the wire format at the field level, with the multi-turn `TestOpenRouter_AssistantToolUseRoundTrip` doing the heavy lifting.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/provider_openrouter.go` | New — ~270 lines of raw HTTP client + private wire types |
| 2 | `backend/internal/agent/provider_openrouter_test.go` | New — 4 tests: request shape, tool call response, assistant tool-use round-trip, error mapping |
| 3 | `backend/internal/config/config.go` | Added `OpenRouterKey string` field loaded from `OPENROUTER_API_KEY`; unvalidated (optional) |
| 4 | `backend/internal/agent/agent.go` | `New()` registers `providers["openrouter"]` only when `cfg.OpenRouterKey != ""`; logs `openrouter provider enabled` at startup |

---

### Phase 3 — `provider` column + handler validation

Phase 2 built the engine; Phase 3 connected the wires. The `app_agent_config` table gained a `provider TEXT NOT NULL DEFAULT 'anthropic'` column, threaded through sqlc, the agent loop, and the API handler.

The actual unblocker is two lines in `loop.go` — replacing the hardcoded `providerFor(defaultProviderName)` with `providerFor(appCfg.Provider)` in `RunConversation` and `RunMonitoring`. That's the payoff of Phase 1's abstraction: the loop change is one string per call site.

Validation lives in the handler, not the database. No `CHECK (provider IN ('anthropic','openrouter'))` constraint — adding a third provider shouldn't require a migration, and the `OPENROUTER_API_KEY` gate is server-side state a SQL constraint can't express anyway. The handler rejects `openrouter` at the API boundary when the key isn't set (visible error), and `providerFor` silently falls back to Anthropic if somehow a bad row slips through (defence in depth). **Zero test edits required** — the empty-string default resolves to `"anthropic"` via the fallback, so every existing `AppAgentConfig{...}` literal still compiles and passes.

| # | File | Change |
|---|------|--------|
| 1 | `backend/migrations/022_agent_config_provider.up.sql` | New — `ALTER TABLE app_agent_config ADD COLUMN provider TEXT NOT NULL DEFAULT 'anthropic'` |
| 2 | `backend/migrations/022_agent_config_provider.down.sql` | New — symmetric `DROP COLUMN` |
| 3 | `backend/internal/db/queries/app_agent_config.sql` | `provider` added to `UpsertAppAgentConfig` INSERT + ON CONFLICT SET |
| 4 | `backend/internal/db/{models.go,app_agent_config.sql.go}` | Regenerated via `sqlc generate` |
| 5 | `backend/internal/agent/loop.go` | Two call sites switch from `defaultProviderName` to `appCfg.Provider` |
| 6 | `backend/internal/api/handlers/applications.go` | `UpdateAppAgentConfig` accepts `provider`; validates `anthropic` (always) / `openrouter` (only if key set) / rejects anything else |

---

### Phase 4 — `GET /api/models` curated catalogue

A thin read-only endpoint serving the model dropdown (Phase 5). Three Anthropic models are always returned; six OpenRouter models are appended only when `OPENROUTER_API_KEY` is set. No DB access, no external calls — the lists live in Go and are served straight out of memory.

**Anthropic (3):** `claude-sonnet-4-6`, `claude-haiku-4-5-20251001`, `claude-opus-4-6`, all 200k context.

**OpenRouter (6):** `anthropic/claude-sonnet-4`, `openai/gpt-4o`, `openai/gpt-4o-mini`, `google/gemini-2.5-pro` (1M context), `google/gemini-2.5-flash` (1M context), `meta-llama/llama-4-maverick`.

Curated, not proxied from OpenRouter's 500+-model catalogue, for three reasons: tool-use compatibility is per-model (many smaller open-source models don't handle multi-turn tool calls reliably), pricing is stable in a hardcoded list, and a dropdown of six is browsable where six hundred isn't.

This is the third place `OpenRouterKey != ""` gates behaviour (after `agent.New` registration in Phase 2 and `UpdateAppAgentConfig` validation in Phase 3), completing the "user never sees a UI option the backend would reject" principle.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/models.go` | New — `ModelOption` / `Pricing` types + `AnthropicModels` / `OpenRouterModels` slices |
| 2 | `backend/internal/api/handlers/models.go` | New — `GetAvailableModels` handler, 4-line concatenation + key gate |
| 3 | `backend/internal/api/router.go` | `r.Get("/models", ...)` registered inside the auth-protected `/api` group |

---

### Phase 5 — Frontend dropdown

Replaced the free-text `<input>` + `<datalist>` model picker on the Agent Configuration page with a native `<select>` grouped by provider via `<optgroup>`. Fetches from `GET /api/models` on mount (parallelised with the config fetch via `Promise.all`).

Native `<select>` rather than extending the project's custom `BaseSelect` component — the custom dropdown doesn't support `<optgroup>`, and rebuilding keyboard nav, focused-index tracking, and grouped-panel styling for one use case was ~80 lines of new code for no user win. The trade-off is that the open panel is OS-drawn (not Tailwind-styled), but the closed state matches the rest of the form via `appearance-none` + the standard border/focus classes. If the model count grows past ~15 entries, a custom grouped dropdown becomes worth it; six entries in two groups doesn't justify the build.

Option labels render as `{Name} — {context} · ${prompt}/${completion}`, e.g. `Claude Sonnet 4.6 — 200k · $3/$15`. The `formatContext` helper switches to `1M` for Gemini's million-token window.

The provider is auto-set from the model's `<optgroup>` membership — users only ever pick a model, and the provider follows. This eliminates the invalid combo (`provider:"anthropic"` + `model:"openai/gpt-4o"`) and matches the user mental model ("I want GPT-4o," not "I want to switch providers").

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/types/models.ts` | New — mirrors `ModelOption` struct |
| 2 | `frontend/src/api/models.ts` | New — axios wrapper for `GET /api/models` |
| 3 | `frontend/src/types/organization.ts` | `AppAgentConfig.provider: 'anthropic' \| 'openrouter'` added |
| 4 | `frontend/src/pages/AgentConfigPage.vue` | Dropdown rewrite — `formProvider` state, `anthropicModels` / `openrouterModels` computeds, `onModelSelect` auto-sync, `saveConfig` includes provider; display mode shows `via {provider}` in muted text |

---

### Phase 6 — Tool-progress streaming (scope cut)

The original plan called for full token-by-token LLM streaming — ~600 lines spanning the `Provider` interface (new streaming method), both provider implementations, the agent loop (`RunConversationStream`), the chat handler (channel consumer), the frontend composable (in-progress message slot), and the chat page (active tools indicator).

Phase 6 shipped the **loop-level** streaming only — the user sees which tool is running live, but the model's prose still arrives as a single block. That cuts the scope roughly in half (~150 lines) and delivers the bigger UX win (users see `searching logs · querying database` live during the wait instead of a blank "thinking…" spinner). Token streaming is tracked separately in `docs/executing/streaming-implementation.md`.

The architectural change is that `runConversationCore` is now a single shared loop body with an optional event sink:

```go
func (a *Agent) runConversationCore(
    ctx context.Context, userID, appID uuid.UUID,
    conversationID *uuid.UUID, history []Message, input string,
    events chan<- AgentEvent, // nil for blocking callers
) (string, error)
```

`RunConversation` (blocking, what `RunLoop` uses) passes `nil`. `RunConversationStream` passes the producer side of a buffered channel and wraps the core call in a goroutine. A nil-safe `emitEvent` helper keeps the blocking path byte-for-byte equivalent to pre-Phase-6. The alternative — duplicating the 100-line loop body — would have been the classic "fixed in one path, forgot the other" source of bugs, and the loop is exactly the kind of code that gets touched frequently.

The frontend indicator has one UX subtlety worth preserving: after the last tool returns but before the final synthesis arrives, `isThinking` is re-armed so the wait is always visually accounted for. Without it, state 3 ("agent is composing its reply after tools") looks like the chat broke.

**Monitoring stays blocking** — the 0.27.0 rate limiter contract depends on `RunMonitoring` calling the blocking `ChatCompletion` path as one atomic unit. The callout is loud in `loop.go:187` for future contributors.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/loop_stream.go` | New — `AgentEvent` type, `RunConversationStream`, `streamBufferSize = 16` |
| 2 | `backend/internal/agent/loop.go` | `runConversationCore` private helper; `emitEvent` nil-safe helper; `tool_start` / `tool_result` emit sites added |
| 3 | `backend/internal/api/handlers/chat.go` | Replaced blocking `RunConversation` + write with `for ev := range events` consumer; forwards `tool_start` / `tool_result` to WS; persists final message on `message` event |
| 4 | `frontend/src/composables/useAgent.ts` | `activeTools` ref; `tool_start` / `tool_result` handlers; `isThinking` re-arm logic; error + final-message handlers clear `activeTools` |
| 5 | `frontend/src/pages/AgentChatPage.vue` | Active-tools indicator above `<ChatWindow>` — muted accent bar, pulsing dot, tool list joined by ` · ` |
| 6 | `docs/executing/streaming-implementation.md` | New — canonical follow-up plan for token streaming |

---

### Phase 7 — Post-integration hardening

A code review after Phases 1–6 surfaced two issues worth fixing before the production push.

**Fix 1 — Stale `defaultModelID` fallback.** `loop.go` still carried `defaultModelID = "claude-sonnet-4-5"` from a pre-0.21 world, and `provider_anthropic.go` had a matching `anthropic.ModelClaudeSonnet4_5` safety net. Every other site in the codebase (applications.go, organizations.go, `agent/models.go`, monitor.go, migrations 002 and 015, every test fixture) already pointed at `claude-sonnet-4-6`. The fallback would only fire if both config reads failed — unlikely on a healthy system, but "unlikely" is the wrong bar when the failure mode is "request goes to a model not in the curated list, producing a confusing error with no visible clue." The provider's own fallback was also deleted as redundant defensive coding (the loop guarantees `params.Model` is populated); a comment documents the caller contract.

**Fix 2 — Producer goroutine leak on WebSocket disconnect (the real bug).** `RunConversationStream` spawned a goroutine that pushed events onto a buffered channel (`streamBufferSize = 16`). If the chat client disconnected mid-loop, `chat.go` returned without draining the channel. A chatty tool-use loop can exceed 16 events (10 iterations × 2–3 tools × 2 events per tool ≈ 40–60 worst case), so the buffer fills, and the next blind channel send in `emitEvent` blocks **forever** — context cancellation alone doesn't unblock a bare send. The goroutine holds onto the loop's stack, tool results, message history, and provider response until the process restarts. One leak per flaky mobile client × days between deploys = the slow memory growth that turns into an OOM at 3am two weeks after deploy.

Fix: make all channel sends ctx-aware by wrapping them in `select { case events <- ev: case <-ctx.Done(): }`. The chat handler's `r.Context()` is cancelled automatically by `net/http` when the handler returns, so the cancellation now propagates cleanly into the producer — any blocked `emitEvent` unblocks, the loop's in-flight `ChatCompletion` aborts (same ctx on the HTTP request), `defer close(ch)` fires, and the channel is collected. The blocking path is untouched because `emitEvent` short-circuits at the nil-events check before reaching the `select`.

Zero test edits required — neither fix changed externally observable behaviour on any existing code path.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/loop.go` | `defaultModelID` bumped; `emitEvent` signature gains `ctx`; three call sites updated; channel send wrapped in `select` |
| 2 | `backend/internal/agent/provider_anthropic.go` | Stale `ModelClaudeSonnet4_5` fallback removed; contract comment added |
| 3 | `backend/internal/agent/loop_stream.go` | Terminal `error` / `message` sends wrapped in the same `select` pattern |

---

### Phase 8 — Post-assessment hardening (the production-push gate)

A second, more formal production-readiness assessment after Phase 7 graded the integration at ~8.0/10 with the gap concentrated in raw-HTTP hygiene and API-boundary polish. Three concrete fixes; the assessment also surfaced two false alarms that turned out to already be correct on re-verification (worth documenting so future readers don't re-investigate).

**Fix 1 — Unbounded OpenRouter response body read.** The HTTP client had a 2-minute request timeout, but `io.ReadAll(resp.Body)` had no size cap. A broken or hostile upstream could return an arbitrarily large body and OOM the Fly VM — which is exactly the failure mode 0.28.0's memory upgrade was trying to avoid in the first place. Introduced `openRouterMaxBodyBytes = 10 << 20` (10 MiB — ~100× legitimate response size) and wrapped the read in `io.LimitReader`. If the cap is ever hit, `json.Unmarshal` fails on the truncated body and surfaces as a normal `openrouter: decode response` error — the right signal for "something is very wrong upstream."

**Fix 2 — Rune-unsafe `truncate` helper.** The function doc comment claimed "n runes" but the body did `s[:n]` (byte slicing). Multi-byte characters (emoji, non-Latin script, CJK stack trace lines) in error bodies would get sliced mid-codepoint, producing invalid UTF-8 that `json.Marshal` silently replaces with U+FFFD — invisible mangling. Phase 7 flagged this and punted; Phase 8 fixed it with a zero-alloc fast path (`len(s) <= n` returns immediately, since UTF-8 guarantees byte length ≥ rune count) and a `[]rune`-sliced slow path only when needed.

**Fix 3 — Case-sensitive provider validation.** `req.Provider` switch in `applications.go` rejected `"OpenRouter"`, `"ANTHROPIC"`, and whitespace-padded variants. No correctness impact, but a rough API edge. Normalised with `strings.ToLower(strings.TrimSpace(...))` before the existing empty-check and switch. The `TrimSpace` also defends against stray-newline bugs from clients that read the value out of a file.

**False alarms (documented in the completion doc, not fixed):** The assessment flagged `orMessage.Content` as missing `omitempty` and the chat handler as sending-before-persisting. Both turned out to already be correct on re-read — `Content` has `omitempty`, and `chat.go:230` persists before `chat.go:233` sends. The lesson: treat a code review's suggestions as hypotheses, not findings, until verified against the actual lines.

Post-Phase-8 score: ~8.7/10. Ready to push.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/provider_openrouter.go` | `openRouterMaxBodyBytes` const; `io.LimitReader`-wrapped body read; rune-safe `truncate` with zero-alloc fast path |
| 2 | `backend/internal/api/handlers/applications.go` | `strings` import; `strings.ToLower(strings.TrimSpace(req.Provider))` normalisation before the validation switch |
| 3 | `docs/completions/openrouter-phase-8.md` | New — completion doc covering the three fixes, the two false alarms, and the pre-push checklist |

---

### Summary

| Phase | Category | Item | Impact |
|---|---|---|---|
| 1 | Refactor | Provider abstraction — neutral types + interface | Decouples agent loop from anthropic-sdk-go; load-bearing for everything downstream |
| 2 | Feature | `OpenRouterProvider` (raw HTTP, blocking) | Second provider implementation; ~270 lines; 4 wire-format tests |
| 2 | Feature | `OPENROUTER_API_KEY` config load + conditional registration | Providers map grows only when key is set |
| 3 | Migration | `022_agent_config_provider` — `provider` column | Per-app provider selection reaches the DB |
| 3 | Feature | `UpdateAppAgentConfig` provider validation | API-boundary rejection when key not set; no silent fallback |
| 4 | Feature | `GET /api/models` curated catalogue | 3 Anthropic + 6 OpenRouter models, key-gated |
| 5 | Feature | Grouped model dropdown with auto-set provider | End-users can select OpenRouter models from the UI |
| 6 | Feature | Tool-progress streaming (scope cut from token streaming) | Active-tools indicator replaces blank "thinking…" during multi-tool loops |
| 6 | Refactor | `runConversationCore` with nil-safe event sink | Shared loop body for blocking + streaming callers; no duplication |
| 7 | Fix | Stale `defaultModelID` `claude-sonnet-4-5` → `claude-sonnet-4-6` | Dead fallback path no longer points at a model outside the curated list |
| 7 | Fix | Producer goroutine leak on WebSocket disconnect | Ctx-aware `emitEvent`; eliminates slow memory growth on flaky clients |
| 8 | Fix | Unbounded OpenRouter response body read | `io.LimitReader` (10 MiB cap); bounds OOM risk on broken upstream |
| 8 | Fix | Rune-unsafe `truncate` helper | Zero-alloc fast path; no more invalid UTF-8 in multi-byte error bodies |
| 8 | Fix | Case-sensitive provider validation | `strings.ToLower(strings.TrimSpace(...))` normalisation at the API boundary |

**What stayed the same across all eight phases:**

- The 0.27.0 monitoring rate limiter (`a.limiter.Wait(ctx)`) — still wraps each `RunMonitoring` invocation as one atomic blocking unit. Monitoring never moved to streaming, by design.
- The 10-iteration cap on the tool-use loop.
- Tool dispatch via `Agent.Dispatch`, keyed on tool name.
- Conversation persistence via `chat.go` → `persistMessages`. Phase 6 moved the call site but not the ordering (persist → then WS write).
- `EmitLog` calls for `tool_call` / `tool_result` / `observation` agent_log entries. Every emit site is preserved with identical arguments.
- All existing tests — zero edits across all eight phases on the Go side; frontend tests untouched (23/23 green throughout).
- The Anthropic provider's behaviour. Every fix and every feature addition preserved the "Anthropic path is byte-equivalent to before" invariant.

**Known follow-ups not addressed in 0.29.0:**

1. **Token-by-token streaming** — tracked in `docs/executing/streaming-implementation.md`. The Phase 7 ctx-aware emit pattern and the Phase 8 `io.LimitReader` both carry through transparently when this lands.
2. **`claude-sonnet-4-6` literal duplicated across 7+ sites** — flagged in Phases 3, 4, and 7. Extract `agent.DefaultModelID` the next time the ID changes.
3. **Live OpenRouter pricing verification** — the six entries in `agent/models.go` came from the selection plan as a working baseline. Cross-check against openrouter.ai/models before announcing OpenRouter support publicly.
4. **End-to-end smoke test** against staging with a real `OPENROUTER_API_KEY`. The Phase 3 manual recipe is the canonical validation — wire-format tests cover translation but not a full agent-loop round-trip through a real model.

None of these block the production push.

---

## 0.28.0 — Fly.io VM Memory Upgrade (2026-04-06)

### Infra — VM memory doubled from 1 GB to 2 GB

The Lumber ONNX classifier (added in 0.20.3, hardened in 0.27.0) loads the quantized `mdbr-leaf-mt` model into memory at startup. The model alone occupies ~250–400 MB in the ONNX Runtime, leaving only ~400–620 MB headroom on the previous 1 GB machine — before the Go runtime, pgxpool, HTTP server, and concurrent classification workloads. Under the monitoring loop's `maxConcurrentApps: 10` concurrency, transient allocations during batch classification could push the machine to OOM.

2 GB gives comfortable headroom for the model baseline plus concurrent classification spikes, with no change to machine type (shared CPU remains appropriate — ONNX inference is serialised by the mutex added in Lumber v0.10.6).

**Cost impact:** $5.92/month → $11.11/month (+$5.19/month).

| # | File | Change |
|---|------|--------|
| 1 | `backend/fly.toml` | `memory = '1gb'` → `'2gb'`; `memory_mb = 1024` → `2048` |

---

## 0.27.1 — Go 1.25 Build Image (2026-04-06)

### Fix — Dockerfile build image updated to Go 1.25

`fly deploy` failed at `go mod download` with:

```
go: go.mod requires go >= 1.25.0 (running go 1.24.13; GOTOOLCHAIN=local)
```

When `golang.org/x/time v0.15.0` was added as a direct dependency in 0.27.0, `go get` automatically bumped the `go` directive in `go.mod` to `1.25.0` — the minimum version declared by that module. The Dockerfile build stage was still pinned to `golang:1.24`, which predates that requirement.

**Fix:** Build image updated from `golang:1.24` to `golang:1.25`.

| # | File | Change |
|---|------|--------|
| 1 | `backend/Dockerfile` | `FROM golang:1.24 AS build` → `FROM golang:1.25 AS build` |

---

## 0.27.0 — Lumber Classifier Hardening (2026-04-06)

Six issues found in a post-integration audit of the Lumber ONNX classifier pipeline: a thread-safety risk in the shared model instance, dead code in the severity gate, a silently divergent confidence threshold, an unbounded Claude escalation risk on classifier failure, invisible tokenizer truncation on large payloads, and unnecessary agent_log noise on every monitoring tick.

### Fix — Lumber upgraded v0.9.0 → v0.10.6 (thread-safety)

`ONNXEmbedder` in v0.9.0 had no mutex protecting ONNX inference calls. The underlying `DynamicAdvancedSession` is not thread-safe. With `maxConcurrentApps: 10` all sharing the same `*lumber.Lumber` instance and potentially calling `ClassifyBatch` concurrently during a monitoring tick, this was a real risk of memory corruption or crash under load. v0.10.6 adds a `sync.Mutex` around all inference calls.

| # | File | Change |
|---|------|--------|
| 1 | `backend/go.mod` | `github.com/kaminocorp/lumber v0.9.0` → `v0.10.6` |
| 2 | `backend/go.sum` | Updated |

---

### Fix — Dead DATA/SCHEDULED branches removed from severity gate

Lumber's current taxonomy has six root types: `ERROR`, `REQUEST`, `DEPLOY`, `SYSTEM`, `ACCESS`, `PERFORMANCE`. There is no `DATA` or `SCHEDULED` type — logs that would semantically match them are returned as `UNCLASSIFIED` by the model. The `DATA` and `SCHEDULED` switch cases in `ShouldEscalate` would never execute; the intended category-level nuance (e.g. treating `DATA.query_executed` as safe) was silently lost. Escalation still happened, but via the `UNCLASSIFIED → all` default rather than the intended branch.

**Fix:** Both switch cases removed. A comment documents the six actual root types. Any type not explicitly handled falls through to `default: return true`. Tests updated: the previously "safe" DATA/SCHEDULED categories are moved to the escalated list.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/severity_gate.go` | `DATA` and `SCHEDULED` cases removed; taxonomy comment added |
| 2 | `backend/internal/agent/severity_gate_test.go` | Safe DATA/SCHEDULED cases moved to escalated list |

---

### Fix — Confidence threshold ownership consolidated

`NewLumberClassifier` passed `lumber.WithConfidenceThreshold(0.5)` to the Lumber engine, which relabels low-confidence events internally. `classifier_lumber.go` then independently re-checked `event.Confidence < 0.5` before escalating. Two separate places owning the same policy: changing one without the other (e.g. tuning the threshold to 0.3) would silently diverge escalation behaviour.

**Fix:** Removed `lumber.WithConfidenceThreshold(0.5)` from `NewLumberClassifier`. The explicit `< 0.5` check in `Classify` is now the single source of truth; a comment documents the ownership decision.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/classifier_lumber.go` | `WithConfidenceThreshold` removed; ownership comment added |

---

### Fix — Rate limiter added on Claude invocations

The default `CLASSIFIER_MODE` is `fallback`. If the Lumber model fails to load (wrong path, corrupted file, ORT version mismatch), `PassthroughClassifier` silently takes over and escalates every log to Claude. With `maxConcurrentApps: 10`, `logBatchLimit: 200`, and a 15-second tick, this could drive significant unintended API cost with no signal beyond a startup warning log.

**Fix:** Added `golang.org/x/time/rate` as a direct dependency. A `*rate.Limiter` field was added to `Agent`, initialised at 30 sustained invocations per minute (1 per 2 seconds), burst of 5. `monitorApp` calls `limiter.Wait(ctx)` immediately before `RunMonitoring`; if the context is cancelled while waiting, the function returns without calling Claude. The rate is permissive for normal operation but bounds the damage on classifier failure.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/agent.go` | `*rate.Limiter` field; `monitorLLMRate` var; `golang.org/x/time/rate` imported |
| 2 | `backend/internal/agent/monitor.go` | `a.limiter.Wait(ctx)` before `RunMonitoring` |
| 3 | `backend/go.mod` | `golang.org/x/time` promoted to direct dependency |

---

### Fix — Payload size guard before classification

`ExtractText` falls back to the raw JSON string when no known message field is present. The underlying BERT-style tokenizer silently truncates at ~512 tokens, so very large JSON payloads degraded classification quality unpredictably with no visible indication.

**Fix:** Added `maxClassifyChars = 1000` constant. The raw JSON fallback path now truncates to 1000 UTF-8 characters before returning to `ClassifyBatch`, making the tokenizer boundary explicit and auditable. Extracted message strings (from known fields) are not truncated.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/extract.go` | `maxClassifyChars = 1000`; raw JSON fallback truncated |

---

### Fix — Heartbeat suppressed when no flagged logs

`EmitLogWithSeverity` was called for every app on every 15-second monitoring tick, writing an `agent_log` entry regardless of whether anything was flagged. In a multi-app deployment running continuous monitoring, this generated significant log volume in the UI with no signal value — a heartbeat entry for every clean tick.

**Fix:** The `EmitLogWithSeverity` call is now inside the `if len(flagged) > 0` block. Agent log entries are only written when the monitoring cycle has something to report. Go `slog` output to stdout continues unconditionally.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/monitor.go` | Heartbeat emit moved inside `flagged > 0` block |

---

### Summary

| # | Category | Item | Impact |
|---|----------|------|--------|
| 1 | Fix | Lumber v0.10.6 — mutex on ONNX inference | Thread-safety under concurrent load |
| 2 | Fix | Dead DATA/SCHEDULED gate branches removed | Correct escalation behaviour + clean code |
| 3 | Fix | Confidence threshold owned in one place | Policy divergence on threshold tuning prevented |
| 4 | Fix | Rate limiter on Claude calls (30/min) | API cost bounded on classifier failure |
| 5 | Fix | 1000-char cap on raw JSON before classification | Predictable tokenizer behaviour |
| 6 | Fix | Heartbeat only emitted when logs are flagged | Reduced agent_log noise in UI |

---

## 0.26.1 — Postgres Connector Interface Fix (2026-04-06)

### Bug fix — Postgres connector did not satisfy the `Connector` interface

`connectors/database/postgres.go` had two interface mismatches against the `Connector` interface defined in `connectors/connector.go`:

1. **`Close` signature mismatch** — `Postgres.Close(ctx context.Context) error` does not match the required `Close() error`. The `Registry.Remove` call site at `registry.go:38` calls `c.Close()` polymorphically via the interface, which would fail to compile if `Postgres` were ever stored in the registry.
2. **`Health` method missing** — `Connector` requires `Health(ctx context.Context) error`. `Postgres` had no such method, making it impossible to use `Postgres` anywhere a `Connector` or `QueryConnector` is expected.

Neither issue surfaced as a compile error today because `Postgres` is only ever used via its concrete type at the two call sites (`tools_db.go`, `connections.go`) — never stored as a `Connector`. That would change the moment it is passed to the registry or any other polymorphic context.

**Fix:**
- `Close(ctx context.Context) error` → `Close() error`. The underlying `pgx.Conn.Close` requires a context internally (to send a graceful termination to the server); `context.Background()` is used so teardown is never cancelled by an expired caller context.
- `Health(ctx context.Context) error` added — delegates to `conn.Ping(ctx)`.
- Two call sites updated from `pg.Close(ctx)` to `pg.Close()`.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/database/postgres.go` | `Close` signature fixed; `Health` method added |
| 2 | `backend/internal/agent/tools_db.go` | `defer pg.Close(ctx)` → `defer pg.Close()` |
| 3 | `backend/internal/api/handlers/connections.go` | `defer pg.Close(ctx)` → `defer pg.Close()` |

---

## 0.26.0 — Observability & Deployment Hygiene (2026-04-06)

Four items from the post-0.25.0 incomplete-features audit: a health check endpoint for Docker and K8s probes, a Prometheus metrics endpoint instrumenting the monitor loop and notification pipeline, JSON-structured log output for aggregators, and the missing notification env vars documented in `.env.example`.

### Feature — Health check endpoint

No `/health` route existed, so the Docker Compose backend service could not report a healthy status and K8s readiness/liveness probes had nothing to hit.

**Implementation:** New `GET /health` handler at `handlers/health.go`. Calls `pool.Ping` — returns `{"status":"ok"}` (200) if the database is reachable, `{"status":"error","db":"unreachable"}` (503) if not. Registered at the top level in `router.go`, outside the `/api` prefix and the JWT middleware group, so probes never need auth. `docker-compose.yml` backend service now includes a `healthcheck` using `curl -sf http://localhost:8080/health`.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/health.go` | New — `Health` handler with DB ping |
| 2 | `backend/internal/api/router.go` | `GET /health` registered at top level |
| 3 | `docker-compose.yml` | `healthcheck` block added to backend service |

---

### Feature — Prometheus metrics endpoint

No metrics existed, leaving the monitor loop, classifier pipeline, and notification dispatcher as black boxes in production.

**Implementation:** New `internal/metrics/metrics.go` package defines three metrics as package-level `promauto` vars, registered against the default Prometheus registry at init time. Instrumented at the two most valuable call sites in the codebase. `GET /metrics` exposed via `promhttp.Handler()` at the top level, unauthenticated (standard for Prometheus scraping).

**Metrics:**

| Metric | Type | Labels | What it measures |
|--------|------|--------|-----------------|
| `heimdall_monitor_tick_duration_seconds` | Histogram | — | Wall time of each full monitoring tick across all active apps |
| `heimdall_logs_classified_total` | Counter | `result` (safe \| flagged) | Log entries processed by the classifier pipeline |
| `heimdall_notifications_total` | Counter | `status` (sent \| failed), `channel_type` (email \| slack \| discord) | Notification dispatch outcomes by channel |

**Instrumentation call sites:**

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/metrics/metrics.go` | New — three metric definitions |
| 2 | `backend/internal/api/router.go` | `GET /metrics` via `promhttp.Handler()` |
| 3 | `backend/internal/agent/monitor.go` | `MonitorTickDuration` timed with deferred closure; `LogsClassifiedTotal` incremented after each `Classify` call |
| 4 | `backend/internal/notifications/notifier.go` | `NotificationsTotal` incremented at each dispatch outcome in `dispatchToChannel` |

---

### Feature — Structured JSON log output

`slog` wrote text lines to stdout only, making log ingestion into Datadog, Loki, or CloudWatch require fragile parsing.

**Implementation:** New `LOG_FORMAT` env var in config (default: `text`). When set to `json`, `main.go` initialises the default slog logger with `slog.NewJSONHandler` before any other startup code runs, so all subsequent log output is structured JSON. No change to log callsites.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/config/config.go` | `LogFormat string` field, loaded from `LOG_FORMAT` env var |
| 2 | `backend/cmd/heimdall/main.go` | JSON handler initialised if `cfg.LogFormat == "json"` |

---

### Fix — Notification env vars documented

`RESEND_API_KEY` and `NOTIFICATION_FROM_EMAIL` were loaded by config and required for email notifications, but absent from `.env.example`. Operators deploying email channels had no indication these were needed, causing silent send failures.

**Fix:** Both vars added to `.env.example` under the `# Optional` block alongside `LOG_FORMAT`.

| # | File | Change |
|---|------|--------|
| 1 | `.env.example` | `RESEND_API_KEY`, `NOTIFICATION_FROM_EMAIL`, `LOG_FORMAT` added |

---

### Summary

| # | Category | Item | Impact |
|---|----------|------|--------|
| 1 | Feature | `GET /health` with DB ping + Docker healthcheck | Deployment hygiene |
| 2 | Feature | `GET /metrics` — monitor tick duration, logs classified, notifications sent/failed | Production visibility |
| 3 | Feature | `LOG_FORMAT=json` structured slog output | Log aggregator compatibility |
| 4 | Fix | Notification env vars added to `.env.example` | Operator experience |

---

## 0.25.0 — Code Assessment Cleanup (2026-04-06)

Four maintenance items from the April 6 code assessment: dead code removed across backend and frontend, poller initialisation deduplicated into a shared factory, stub packages deleted, and a silenced `json.Marshal` error made explicit.

### Refactor — Dead handlers and frontend modules removed

Four backend handlers existed but were not registered in `router.go` — superseded by per-app equivalents (`GetAppAgentConfig`, `GetAppDashboardStats`) when the multi-app model was introduced in 0.10.0 but never deleted. Three frontend modules called the corresponding dead endpoints, and the `AgentConfig` type they referenced used a stale `'scheduled'` mode value instead of the live `'periodic'`.

**Fix:** Deleted all dead backend handlers, their frontend API modules, the orphaned Pinia store, and the stale type. The test file for the dead store was also removed — an orphaned test that passed gives false confidence about unreachable code.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/agent.go` | Deleted — `GetAgentConfig`, `UpdateAgentConfig`, `RunAgent` unregistered |
| 2 | `backend/internal/api/handlers/stats.go` | Deleted — `GetDashboardStats` unregistered |
| 3 | `frontend/src/api/agent.ts` | Deleted — called dead `/agent/config` endpoint |
| 4 | `frontend/src/stores/agent.ts` | Deleted — consumed dead API module, never imported by any page |
| 5 | `frontend/src/stores/__tests__/agent.test.ts` | Deleted — test for deleted store |
| 6 | `frontend/src/api/stats.ts` | Deleted — called dead `/stats` endpoint |
| 7 | `frontend/src/types/agent.ts` | Removed `AgentConfig` interface; WebSocket/chat types retained |

---

### Refactor — Poller factory extracted to eliminate duplication

Poller startup logic was duplicated between `cmd/heimdall/main.go` (`resumePollers`) and `handlers/connections.go` (`startPoller`) — the same switch over five connector types with identical `New* → poller.Start` patterns. Adding a new poll-based connector required changes in both places.

**Fix:** Extracted a `connectors.StartPoller` factory function. Both call sites now delegate to it. `resumePollers` simplified from a slice of structs-with-closures to a plain loop over type name strings.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/factory.go` | New — `StartPoller(poller, connType, config, connID, userID)` with single switch |
| 2 | `backend/internal/api/handlers/connections.go` | Removed `startPoller`; two call sites replaced with `connectors.StartPoller` |
| 3 | `backend/cmd/heimdall/main.go` | `resumePollers` rewritten to loop over type strings and call `connectors.StartPoller` |

---

### Refactor — Stub packages deleted

Three packages contained only TODO no-ops with no callers anywhere in the codebase. They were placeholders for future features that had not been developed.

**Fix:** Deleted all three packages in full. The real webhook ingestion path is `handlers/webhook_parsers.go` (HTTP handler), not the `StreamConnector` pattern the stub represented. The `reports` and `memory` packages had no callers at any layer.

| # | Package | Files deleted | Content |
|---|---------|--------------|---------|
| 1 | `internal/connectors/logs` | `webhook.go`, `webhook_test.go` | `Stream()` no-op; stub test asserting non-nil constructor |
| 2 | `internal/reports` | `generator.go`, `templates.go` | `Generate()` returning `nil, nil`; unused template struct |
| 3 | `internal/memory` | `client.go`, `memory.go`, `types.go` | `RecordEvent`, `QueryMemories`, `GetLessons` all no-ops |

---

### Bug fix — Silenced `json.Marshal` error in syslog TLS injection

In `CreateConnection` and `UpdateConnection`, after injecting server-level TLS cert/key into the syslog config map, the result was re-marshalled with `config, _ = json.Marshal(cfgMap)` — silently discarding any error. While `json.Marshal` on a `map[string]interface{}` with string values won't realistically fail, this pattern deviates from the codebase's otherwise consistent error handling and would mask any future regression.

**Fix:** Replaced the blank identifier with an explicit error check. Returns 500 on failure in both handlers.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/connections.go` | `CreateConnection` — `config, _ =` → `config, marshalErr =` with `jsonError` return |
| 2 | `backend/internal/api/handlers/connections.go` | `UpdateConnection` — same fix |

---

### Summary

| # | Category | Item | Severity |
|---|----------|------|----------|
| 1 | Dead code | Unregistered handlers + stale frontend modules | Medium |
| 2 | Maintainability | Duplicated poller init switch across two files | Medium |
| 3 | Dead code | Three stub packages with no callers | Low |
| 4 | Correctness | Silenced `json.Marshal` error in syslog config path | Low |

---

## 0.24.5 — Poller, Parser & Shutdown Hardening (2026-04-06)

Four fixes addressing silent data loss in pollers, payload misrouting in webhook parsers, missing failure feedback for syslog connections, and unbounded shutdown duration.

### Bug fix — Pollers silently drop entries with unparseable timestamps

Fly.io, Railway, and MongoDB pollers used `ts, _ := time.Parse(...)`, discarding the error. If an API returned an unexpected timestamp format, the parsed zero-value `time.Time{}` would never pass `ts.After(cursor)`, causing the entry to be permanently skipped — silent data loss with no log output.

**Fix:** Check the parse error. On failure, log a warning with the raw timestamp and fall back to `time.Now()` so the entry is still ingested.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/logs/flyio.go` | `time.Parse` error → `slog.Warn` + `time.Now()` fallback |
| 2 | `backend/internal/connectors/logs/railway.go` | Same pattern |
| 3 | `backend/internal/connectors/logs/mongodb.go` | Same pattern |

### Bug fix — Webhook format detection matches false positives

`isFirehosePayload` and `isPubSubPayload` used `strings.Contains` to detect formats — checking for `"requestId"` + `"records"` (Firehose) and `"message"` + `"subscription"` (Pub/Sub). The string `"message"` is extremely common in JSON payloads, so a Heimdall native payload with both `"message"` and `"subscription"` keys would be misrouted to the Pub/Sub parser, corrupting the log entry.

**Fix:** Replaced string matching with structural JSON unmarshaling. Each detector now unmarshals into the expected envelope struct and checks that the discriminating fields are non-empty:

- **Firehose:** requires `requestId` (non-empty string) and `records` (non-empty array)
- **Pub/Sub:** requires `subscription` (non-empty string) and `message.data` (non-empty string)

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/webhook_parsers.go` | `isFirehosePayload` and `isPubSubPayload` now accept `[]byte`, unmarshal into typed structs, check field values |

### Bug fix — Syslog listener failure returns success status

When a syslog listener failed to initialize or bind its port in `CreateConnection` or `UpdateConnection`, the error was logged but the HTTP response still returned the connection with `status: "inactive"`. The user had no way to know the listener wasn't running.

**Fix:** On listener failure, update the connection status to `"error"` in the database and reflect it in the response body. The HTTP status code remains 201 (the connection was created), but `"status": "error"` clearly signals the problem.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/connections.go` | `CreateConnection` and `UpdateConnection` set `conn.Status = "error"` and call `UpdateConnectionStatus` on listener failure |

### Resilience — Shutdown timeout for connectors

`poller.StopAll()`, `listener.StopAll()`, and `ag.Stop()` all block until their goroutines finish. If a poller's upstream API hangs or a listener's TCP drain takes too long, shutdown blocks indefinitely — preventing clean deploys.

**Fix:** Wrapped the connector shutdown sequence in a goroutine with a 10-second deadline. If connectors don't stop in time, a warning is logged and the process proceeds to exit.

| # | File | Change |
|---|------|--------|
| 1 | `backend/cmd/heimdall/main.go` | Connector stop calls wrapped in goroutine with `select` + `time.After(10s)` |

### Summary

| # | Category | Issue | Severity |
|---|----------|-------|----------|
| 1 | Data loss | Poller timestamp parse errors silently drop entries | High |
| 2 | Correctness | Webhook format detection false positives via string matching | Medium-High |
| 3 | UX | Syslog listener failure returns success status | Medium |
| 4 | Resilience | No shutdown timeout for pollers/listeners/agent | Medium |

---

## 0.24.4 — SDK Shutdown Safety (2026-04-06)

Pre-production code assessment found data-loss bugs in the JS and Python SDKs during shutdown scenarios. The Go SDK was already correct.

### Bug fix — JS SDK drops in-flight sends on shutdown

The `flush()` method is async but was called in fire-and-forget contexts — the timer callback (`setTimeout`) and the batchSize trigger in `log()` both dropped the returned Promise. If `shutdown()` was called while a timer-initiated send was in-flight, it returned immediately without waiting, and the HTTP request was abandoned.

**Root cause:** The SDK had no way to track fire-and-forget flush operations. A `flushing` flag was declared (line 37) but never used — suggesting concurrent flush protection was planned but not completed.

**Fix:** Replaced the unused `flushing` flag with an `inflightSends` Set that tracks all active send Promises. Every `flush()` call registers its send Promise in the set and removes it on completion. `shutdown()` now awaits both its own flush and all tracked in-flight sends via `Promise.all()`.

| # | File | Change |
|---|------|--------|
| 1 | `packages/sdk-js/src/index.ts` | Replaced `flushing` flag with `inflightSends` Set. `flush()` tracks sends. `shutdown()` awaits all in-flight sends. |

### Bug fix — Python SDK loses data on process exit

Two compounding issues caused data loss:

1. **`_closed` check outside lock (race condition)** — `log()` checked `self._closed` at line 83 without holding the lock, then acquired the lock at line 92 to append. If `shutdown()` executed between these two lines, entries appended after shutdown's flush were never sent.

2. **Daemon threads killed on exit** — `_flush_locked()` spawned send threads with `daemon=True`. Daemon threads are terminated immediately when the main thread exits, killing any in-flight HTTP requests. Combined with issue 1, this meant even successfully-queued entries could be lost.

**Fix:**
- Moved `_closed` check inside the lock, eliminating the race window between check and append.
- Changed send threads from `daemon=True` to non-daemon. Non-daemon threads keep the process alive until they complete, ensuring in-flight sends finish.
- Added `_send_threads` tracking list with cleanup of completed threads in `_flush_locked()`. `shutdown()` now joins all in-flight send threads (with a 30-second timeout per thread) before returning.

| # | File | Change |
|---|------|--------|
| 1 | `packages/sdk-python/heimdall_sdk/client.py` | `_closed` check moved inside lock. Send threads changed to non-daemon. Added `_send_threads` tracking. `shutdown()` joins all threads. `_flush_locked()` cleans up completed threads. |

### Summary

| # | SDK | Issue | Severity |
|---|-----|-------|----------|
| 1 | JS | In-flight sends dropped on shutdown (fire-and-forget flush) | High |
| 2 | Python | `_closed` race condition between check and lock acquisition | High |
| 3 | Python | Daemon send threads killed on process exit | High |

---

## 0.24.3 — UpdateConnection Validation (2026-04-06)

Pre-production code assessment found that `UpdateConnection` had no config validation, no syslog TLS injection, and no webhook token preservation — all of which were present in `CreateConnection`. A user updating any connection could break it silently.

### Bug fix — UpdateConnection skips all config validation

`CreateConnection` validated configs eagerly for all 7 connector types (Supabase, Fly.io, Vercel, Railway, MongoDB, syslog, webhook/OTLP) before inserting into the database. `UpdateConnection` skipped all validation entirely — invalid configs were written to the DB, the working poller/listener was stopped, and the replacement failed to start, leaving the connection in a broken state with no active connector.

**Fix:** Added the same config validation block from `CreateConnection` to `UpdateConnection`. All 7 connector types are now validated before the database update.

### Bug fix — UpdateConnection loses syslog TLS certs

`CreateConnection` injected server-level TLS cert/key (from `SYSLOG_TLS_CERT` / `SYSLOG_TLS_KEY` env vars) into syslog connections that didn't specify their own. `UpdateConnection` skipped this injection, so updating a syslog connection that relied on server-level certs would lose TLS configuration.

**Fix:** Added the same TLS cert/key injection logic to `UpdateConnection`.

### Bug fix — UpdateConnection loses webhook tokens

`CreateConnection` auto-generated a `webhook_token` for `webhook_logs` and `otlp` connections. `UpdateConnection` didn't preserve the existing token — if the update payload omitted the token field, it was overwritten with an empty config, breaking all active integrations using that token.

**Fix:** `UpdateConnection` now fetches the existing connection config before updating. If the new config omits `webhook_token`, the existing token is preserved.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/connections.go` | Added config validation for all 7 types, syslog TLS injection, and webhook token preservation to `UpdateConnection` |

### Summary

| # | Category | Issue | Severity |
|---|----------|-------|----------|
| 1 | Functional | UpdateConnection skips config validation for all types | Critical |
| 2 | Data loss | UpdateConnection loses syslog TLS certs on update | Critical |
| 3 | Data loss | UpdateConnection loses webhook/OTLP tokens on update | Critical |

---

## 0.24.2 — Production Hardening Pass (2026-04-06)

Pre-production review of the full ingestion pipeline (Phases 1–4) identified 8 issues across the syslog listener, API pollers, OTLP handler, and SDKs. All fixed in a single pass — no architectural changes, all additive.

### Security — Syslog DoS prevention

The syslog TCP/TLS listener accepted unbounded connections with no read timeout. A malicious actor (or misbehaving client) could exhaust goroutines by opening thousands of idle connections.

- **Connection semaphore** — Added a cap of 500 concurrent TCP connections per listener (`syslogMaxConnections`). Connections beyond the limit are rejected immediately with a warning log.
- **Read deadline** — Each connection now has a 5-minute read deadline (`syslogReadTimeout`), reset after each successful message. Idle clients are evicted automatically.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/logs/syslog.go` | Added `connSem` channel semaphore, `syslogMaxConnections` (500), `syslogReadTimeout` (5m). `handleConnection` sets/resets `conn.SetReadDeadline`. Accept loop enforces semaphore with non-blocking select. |

### Bug fix — Poller timestamp cursor loses entries with identical timestamps

All 4 API pollers (Fly.io, Vercel, Railway, MongoDB) used `!ts.After(cursor)` to skip already-seen entries. Two log entries with the same timestamp caused the second to be permanently skipped — silent data loss.

**Root cause:** The cursor was set to the exact `maxTS` of the batch. On the next poll, `!ts.After(cursor)` evaluates to `true` for entries at exactly that timestamp, skipping them.

**Fix:** After each poll cycle, advance the cursor by 1 nanosecond past the last seen timestamp (`maxTS.Add(time.Nanosecond)`), ensuring entries at the boundary are never re-skipped.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/logs/flyio.go` | `maxTS = maxTS.Add(time.Nanosecond)` after insert loop |
| 2 | `backend/internal/connectors/logs/vercel.go` | Same cursor advancement pattern |
| 3 | `backend/internal/connectors/logs/railway.go` | Same cursor advancement pattern |
| 4 | `backend/internal/connectors/logs/mongodb.go` | Same cursor advancement pattern |

### Bug fix — Poller DB insert errors silently advance cursor

When `InsertLogEntry` failed in any poller, the error was logged but the loop `continue`d, allowing the cursor to advance past the failed entries. Those entries were permanently lost — the next poll would never see them again.

**Fix:** On insert failure, return an error immediately. The cursor only advances for successfully inserted entries. The poller framework will retry the batch on the next poll cycle.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/logs/flyio.go` | `return count, maxTS, fmt.Errorf(...)` on insert error |
| 2 | `backend/internal/connectors/logs/vercel.go` | `return fmt.Errorf(...)` on insert error |
| 3 | `backend/internal/connectors/logs/railway.go` | `return fmt.Errorf(...)` on insert error |
| 4 | `backend/internal/connectors/logs/mongodb.go` | `return fmt.Errorf(...)` on insert error |

### Bug fix — TestConnection auto-passed for new poller types

The `TestConnection` handler's `default` case auto-passed any connection type not explicitly handled. The 4 new API pollers (Fly.io, Vercel, Railway, MongoDB) all fell into this default — users could create connections with invalid tokens or wrong project IDs and receive a "Connection established" success message.

**Fix:** Added explicit `Connect()`-based test cases for all 4 poller types, matching the existing Supabase pattern. Each test validates credentials against the external API with a 10-second timeout.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/connections.go` | Added `case "flyio"`, `case "vercel"`, `case "railway"`, `case "mongodb"` in `TestConnection` switch |

### Resilience — Poller HTTP 429 rate-limit handling

None of the API pollers checked for HTTP 429 (Too Many Requests). A rate-limited poller would log a generic API error and retry on the regular schedule, potentially escalating to a permanent ban on some platforms.

**Fix:** Added explicit 429 detection in all 4 pollers' HTTP response handling. The error message identifies the rate limit clearly in logs, and the poller framework's existing retry interval naturally provides backoff.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/logs/flyio.go` | 429 check in `Poll` and `pollMachineLogs` |
| 2 | `backend/internal/connectors/logs/vercel.go` | 429 check in `apiRequest` |
| 3 | `backend/internal/connectors/logs/railway.go` | 429 check in `graphQL` |
| 4 | `backend/internal/connectors/logs/mongodb.go` | 429 check in `apiRequest` |

### Bug fix — Go SDK loses in-flight logs on shutdown

`flushLocked()` spawned `go c.send(entries)` without tracking the goroutine. `Shutdown()` called `Flush()` (which spawned the goroutine) then returned immediately — in-flight HTTP requests could be killed by process exit.

**Fix:** Added `sync.WaitGroup` to track all send goroutines. `Shutdown()` now calls `wg.Wait()` after flushing, ensuring all in-flight sends complete before returning.

| # | File | Change |
|---|------|--------|
| 1 | `packages/sdk-go/heimdall.go` | Added `wg sync.WaitGroup`. `flushLocked` wraps `go c.send(entries)` with `wg.Add(1)` / `defer wg.Done()`. `Shutdown` calls `wg.Wait()`. |

### Bug fix — OTLP handler returns 200 on partial insert failure

The OTLP handler `continue`d past insert errors and returned HTTP 200 even when only some records were inserted. Clients had no way to know records were lost.

**Fix:** Track insert errors separately. Return HTTP 207 (Multi-Status) with `accepted` and `rejected` counts when some records fail. Return HTTP 500 only when all records fail (existing behaviour). HTTP 200 only when all records succeed.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/otlp.go` | Added `insertErrors` counter. Returns 207 with `{accepted, rejected}` on partial failure. |

### Bug fix — Python SDK shutdown race condition

`shutdown()` set `self._closed = True` without holding the lock, creating a race window where a concurrent `log()` call could interleave with the shutdown flush.

**Fix:** `shutdown()` now acquires the lock before setting `_closed` and calls `_flush_locked()` directly within the lock, eliminating the race.

| # | File | Change |
|---|------|--------|
| 1 | `packages/sdk-python/heimdall_sdk/client.py` | `shutdown()` acquires `_lock` before setting `_closed` and flushing |

### Summary

| # | Category | Issue | Severity |
|---|----------|-------|----------|
| 1 | Security | Syslog unbounded connections + no read timeout | High |
| 2 | Data loss | Poller cursor skips entries with identical timestamps | High |
| 3 | Data loss | Poller insert errors silently advance cursor | High |
| 4 | UX | TestConnection auto-passes invalid poller configs | High |
| 5 | Resilience | No HTTP 429 rate-limit handling in pollers | Medium |
| 6 | Data loss | Go SDK loses in-flight logs on shutdown | Medium |
| 7 | Correctness | OTLP handler masks partial insert failures | Medium |
| 8 | Correctness | Python SDK shutdown race condition | Low |

---

## 0.24.1 — Ingestion Hardening (2026-04-06)

Post-implementation review of the ingestion pipeline (Phases 1–3) surfaced and fixed 5 issues before production push.

### Security

- **GraphQL injection in Railway connector** — `Connect()` concatenated `ProjectID` directly into a GraphQL query string. Switched to parameterized variables (`$id: String!`), matching the pattern already used by `Poll()`.

### Bug fixes

- **MongoDB missing `cluster_name` validation** — `NewMongoDB` accepted empty `cluster_name` despite using it for source type (`"mongodb/<cluster>"`) and hostname discovery. Added a required check in the constructor.

### Cleanup

- **Removed tracked `__pycache__` files** — 4 Python bytecode files were committed to the repo. Removed from git index and added `**/__pycache__/` and `*.pyc` to `.gitignore`.
- **Removed dead code `parseOTLPTimestamp`** — Function in `otlp.go` was defined but never called. Removed along with the unused `strconv` import.
- **Documented `resolveAnyValue` `BoolValue` limitation** — OTLP `BoolValue: false` is indistinguishable from an unset field due to Go's zero-value semantics. Added a comment documenting the trade-off.

### Phase ordering review

Confirmed Phases 1–3 were implemented in the correct order with no do/undo conflicts. Each phase built additively on the last — no prior work was reverted or overwritten.

| # | File | Change |
|---|------|--------|
| 1 | `.gitignore` | Added `**/__pycache__/` and `*.pyc` |
| 2 | `backend/internal/connectors/logs/railway.go` | `Connect()` uses parameterized `$id` variable |
| 3 | `backend/internal/connectors/logs/mongodb.go` | Added `cluster_name is required` validation |
| 4 | `backend/internal/api/handlers/otlp.go` | Removed `parseOTLPTimestamp`, removed `strconv` import, documented `BoolValue` limitation |

---

## 0.24.0 — Webhook Parsers, API Pollers & Python/Go SDKs (2026-04-06)

Ingestion Phase 3 — four items completing the ingestion roadmap. Together with Phases 1–2, Heimdall now covers ~80% of early users' infrastructure.

### Webhook payload parsers

The webhook endpoint (`POST /api/webhooks/logs`) previously accepted only Heimdall's native JSON format. Added automatic format detection via `parseWebhookPayload` that inspects `Content-Type` and payload structure, then normalises to the internal format before insertion.

| Format | Detection | Source type |
|--------|-----------|-------------|
| Vercel NDJSON | `Content-Type: application/x-ndjson` or multi-line JSON structure | `vercel/<source>` |
| AWS Kinesis Firehose | `requestId` + `records` fields, base64-decoded records | `firehose` |
| GCP Pub/Sub | `message` + `subscription` fields, base64-decoded data | `pubsub` |
| Heimdall native | Default fallback (single object or JSON array) | As provided |

Shared `normalizeSeverity` function maps common severity strings from any platform to Heimdall's 5-level system.

### API pollers

Four new poll-based connectors, all following the established Supabase pattern (`PollConnector` interface, cursor-based dedup, rate-limit-aware intervals):

| Platform | API | Auth | Min interval | Source type |
|----------|-----|------|-------------|-------------|
| Fly.io | Machines API (`api.machines.dev`) | Bearer token | 15s | `flyio/<app>` |
| Vercel | REST API (`api.vercel.com`) | Bearer token | 30s | `vercel/deployment` |
| Railway | GraphQL API (`backboard.railway.app`) | Bearer token | 30s | `railway/<project>` |
| MongoDB Atlas | Admin API v2 (`cloud.mongodb.com`) | HTTP Basic | 60s | `mongodb/<cluster>` |

Refactored `resumePollers` in `main.go` from Supabase-only to a generic loop over all 5 poller types. Added `startPoller` helper in `connections.go` to consolidate poller initialization.

### Python SDK

`heimdall-sdk` — zero-dependency Python package using `urllib.request` (stdlib). Thread-safe batching with `threading.Lock`, background flush via `threading.Timer`, exponential backoff retry. Same API shape as the JS SDK. Python 3.9+.

### Go SDK

`github.com/hejijunhao/heimdall/sdk-go` — zero-dependency Go module using `net/http`. `sync.Mutex` for thread safety, `time.AfterFunc` for flush timer, goroutine-based async send. Custom `*http.Client` injectable via `Options.HTTPClient`.

### Test coverage

| Component | Tests |
|-----------|-------|
| Webhook parsers | 12 |
| Go SDK | 9 |
| Python SDK | 9 |

### Files created

| # | File | Purpose |
|---|------|---------|
| 1 | `backend/internal/api/handlers/webhook_parsers.go` | Format detection, Vercel/Firehose/Pub/Sub/native parsers, severity normalisation |
| 2 | `backend/internal/api/handlers/webhook_parsers_test.go` | 12 parser tests |
| 3 | `backend/internal/connectors/logs/flyio.go` | Fly.io Machines API poller |
| 4 | `backend/internal/connectors/logs/vercel.go` | Vercel REST API poller |
| 5 | `backend/internal/connectors/logs/railway.go` | Railway GraphQL API poller |
| 6 | `backend/internal/connectors/logs/mongodb.go` | MongoDB Atlas Admin API poller |
| 7 | `packages/sdk-python/heimdall_sdk/client.py` | Python SDK client |
| 8 | `packages/sdk-python/heimdall_sdk/__init__.py` | Package exports |
| 9 | `packages/sdk-python/tests/test_client.py` | 9 Python SDK tests |
| 10 | `packages/sdk-python/pyproject.toml` | Python package config |
| 11 | `packages/sdk-go/heimdall.go` | Go SDK client |
| 12 | `packages/sdk-go/heimdall_test.go` | 9 Go SDK tests |
| 13 | `packages/sdk-go/go.mod` | Go module definition |

### Files modified

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/webhooks.go` | Refactored to use `parseWebhookPayload` with size-limited body reading |
| 2 | `backend/internal/api/handlers/connections.go` | 4 new types in validation map, config validation, `startPoller` helper |
| 3 | `backend/cmd/heimdall/main.go` | Generic `resumePollers` for all 5 poller types |

---

## 0.23.0 — OTLP HTTP Receiver & JS SDK (2026-04-05)

Ingestion Phase 2 — two new ingestion paths covering OpenTelemetry-instrumented applications, serverless environments, and any Node.js app.

### OTLP HTTP receiver

New public endpoint `POST /api/v1/logs` accepting the OpenTelemetry Protocol `ExportLogsServiceRequest` JSON format. Uses the same bearer token auth as webhook ingestion (`GetConnectionByWebhookToken`). Flattens the nested OTel structure (`resourceLogs → scopeLogs → logRecords`) into individual `log_buffer` entries.

| OTLP field | Heimdall payload field |
|------------|----------------------|
| `resource.attributes` | `resource` (flattened map) |
| `scope.name` | `scope` |
| `logRecords[].timeUnixNano` | `time_unix_nano` |
| `logRecords[].severityText/Number` | `severity_text`, `severity_number` + mapped Heimdall severity |
| `logRecords[].body` | `body` (resolved AnyValue) |
| `logRecords[].attributes` | `attributes` (flattened map) |
| `logRecords[].traceId/spanId` | `trace_id`, `span_id` |

Source type is `"otlp"` by default, or `"otlp/<service.name>"` when the resource carries that attribute.

OTLP severity mapping: 1–8 → `debug`, 9–12 → `info`, 13–16 → `warning`, 17–20 → `error`, 21–24 → `critical`. Falls back to `severityText` string matching when the number is 0.

New connection type `"otlp"` added to `validConnectionTypes`. OTLP connections auto-generate a webhook token on creation.

### JS/TS SDK

`@heimdall/sdk` — zero-dependency TypeScript package using the global `fetch` API (Node 18+, Bun, Deno, Cloudflare Workers, browsers). Dual ESM/CJS output via tsup.

Batching: entries accumulate in-memory, flushed when buffer reaches `batchSize` (default 25) or `flushInterval` fires (default 5000ms). Retry: 4xx errors are permanent (no retry), 5xx and network errors retry with exponential backoff up to `maxRetries` (default 3).

API surface: `log(severity, sourceType, payload)` plus severity shorthands (`debug`, `info`, `warn`, `error`, `critical`), `flush()`, `shutdown()`, `pending`.

### Test coverage

| Component | Tests |
|-----------|-------|
| OTLP handler | 7 |
| JS SDK | 10 |

### Platforms unlocked

Neon (OTLP), Heroku Fir (OTLP), OTel Collector (OTLP), Fluent Bit / Vector (OTLP), any Node.js app (SDK), serverless — Lambda, Edge, Workers (SDK).

### Files created

| # | File | Purpose |
|---|------|---------|
| 1 | `backend/internal/api/handlers/otlp.go` | OTLP handler — parsing, flattening, severity mapping, insertion |
| 2 | `backend/internal/api/handlers/otlp_test.go` | 7 OTLP unit tests |
| 3 | `packages/sdk-js/src/index.ts` | Heimdall class — batching, retry, severity methods |
| 4 | `packages/sdk-js/src/index.test.ts` | 10 SDK unit tests |
| 5 | `packages/sdk-js/package.json` | Package config (tsup build, vitest) |
| 6 | `packages/sdk-js/tsconfig.json` | TypeScript config |
| 7 | `frontend/src/components/connections/wizard/steps/StepOTLPSetup.vue` | Wizard step — endpoint format, payload example |

### Files modified

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/router.go` | Added `POST /api/v1/logs` route |
| 2 | `backend/internal/api/handlers/connections.go` | Added `"otlp"` to valid types, auto-generate token for OTLP connections |
| 3 | `frontend/src/components/connections/wizard/flows.ts` | OTLP flow: Name → Setup (auto-valid) |

---

## 0.22.0 — Syslog TLS Listener (2026-04-05)

Ingestion Phase 1 — a production-ready TCP/TLS syslog listener that accepts RFC 5424 and RFC 3164 messages over persistent TCP connections and inserts them into `log_buffer`.

### Architecture

Syslog is a long-running TCP server, not a timer-driven poller. This required a new concurrency primitive: the **ListenerManager** — analogous to `Poller` but for persistent network listener goroutines. Each syslog connection binds a port and accepts inbound TCP connections; each client connection is handled in its own goroutine with line-by-line parsing via `bufio.Scanner`.

No external dependencies — uses Go's standard library (`net`, `crypto/tls`, `bufio`, `regexp`) for TCP/TLS listening and syslog parsing.

### Protocol support

Parser tries formats in order: RFC 5424 → RFC 3164 → raw fallback.

| Format | Pattern | Fields extracted |
|--------|---------|-----------------|
| RFC 5424 | `<PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID MSG` | facility, severity, timestamp, hostname, app name, proc ID, msg ID, message |
| RFC 3164 | `<PRI>TIMESTAMP HOSTNAME MSG` | facility, severity, BSD timestamp, hostname, message |
| Fallback | Any unstructured line | message (severity defaults to Informational) |

Severity mapping: 0–2 → `critical`, 3 → `error`, 4 → `warning`, 5–6 → `info`, 7 → `debug`.

### TLS configuration

TLS cert/key can be provided per-connection (config JSONB `tls_cert`/`tls_key`) or server-level (`SYSLOG_TLS_CERT`/`SYSLOG_TLS_KEY` env vars, auto-injected into connections that don't specify their own). Plaintext TCP mode (`protocol: "tcp"`) available for development.

### Connection lifecycle

Create → listener binds port. Listen → accepts TCP connections in a loop. Update → old listener stopped, new one started. Delete → listener stopped. Server restart → `resumeSyslogListeners()` restarts all active syslog connections.

### Platforms unlocked

Render (syslog log streams), Heroku Cedar (syslog drain), DigitalOcean (rsyslog forwarding), any Linux server (rsyslog / syslog-ng).

### Test coverage

| Component | Tests |
|-----------|-------|
| Syslog connector | 9 |

### Files created

| # | File | Purpose |
|---|------|---------|
| 1 | `backend/internal/connectors/logs/syslog.go` | Syslog listener — config, TCP/TLS server, RFC parsing, log insertion, graceful shutdown |
| 2 | `backend/internal/connectors/logs/syslog_test.go` | 9 unit tests |
| 3 | `backend/internal/connectors/listener.go` | `ListenerManager` — start/stop/stopAll for persistent listener goroutines |
| 4 | `frontend/src/components/connections/wizard/steps/StepSyslogConfig.vue` | Wizard step — port + protocol config |

### Files modified

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/config/config.go` | Added `SyslogTLSCert` and `SyslogTLSKey` fields |
| 2 | `backend/internal/api/handlers/server.go` | Added `Listener *connectors.ListenerManager` to Server struct |
| 3 | `backend/internal/api/router.go` | Updated `NewRouter` to accept `ListenerManager` |
| 4 | `backend/internal/api/handlers/connections.go` | Syslog validation, TLS injection, listener lifecycle on create/update/delete |
| 5 | `backend/cmd/heimdall/main.go` | Create `ListenerManager`, `resumeSyslogListeners` on boot, stop on shutdown |
| 6 | `frontend/src/components/connections/wizard/flows.ts` | Syslog flow: Name → Config → Test |
| 7 | `frontend/src/components/connections/ConnectionForm.vue` | Fixed syslog edit fields (listener, not remote target) |

---

## 0.21.0 — Kamino Design System Alignment (2026-04-05)

Systematic sizing and spacing uplift across the entire frontend, aligning Heimdall with the Kamino product family design system (Elephantasm). The UI previously felt undersized and structurally faint — buttons were thin, card borders nearly invisible, text dipped to 10px, and spacing was uniformly tight. No new features; every change is a design token update or Tailwind class adjustment.

### Motivation

A gap analysis against the Elephantasm design system revealed Heimdall was consistently one notch smaller across every dimension: font sizes, button padding, card padding, section spacing, modal padding, nav click targets, and — most critically — border visibility. The compound effect made the interface feel flimsy despite the strong brutalist aesthetic underneath. This release closes those gaps while preserving Heimdall's feldgrau identity, monochrome palette, and typographic character.

### Design token changes

| Token | Before | After | Effect |
|-------|--------|-------|--------|
| `--border` | `rgba(77, 93, 83, 0.06)` | `rgba(77, 93, 83, 0.14)` | Card/panel/input borders now visible at rest — the single highest-impact change |
| `--border-hover` | `rgba(77, 93, 83, 0.14)` | `rgba(77, 93, 83, 0.25)` | Stronger hover feedback on interactive edges |
| Global `:focus-visible` outline | `1px solid var(--accent)` | `2px solid var(--accent)` | More prominent keyboard focus indicator |

### Typography — 12px floor

Eliminated all `text-[10px]` (10px) and `text-[11px]` (11px) usage across 30 files. The minimum font size is now `text-xs` (12px / 0.75rem). Affected elements: sidebar section labels, card headers, status badges, log entry timestamps and severity tags, blueprint node labels, wizard step indicators, form helper text, notification history metadata, pricing badges, and the public footer.

### Spacing & sizing changes

| Element | Before | After |
|---------|--------|-------|
| **Page content padding** | `p-6 lg:p-8` | `px-6 py-8 lg:px-8 lg:py-12` — more vertical breathing room on desktop |
| **Card/panel padding** | `p-5` (20px) | `p-6` (24px) — across all dashboard cards, config panels, forms, report cards, connection cards |
| **Primary button padding** | `px-4 py-2` | `px-5 py-2.5` — taller, more confident CTAs (~40px effective height) |
| **Secondary button padding** | `px-4 py-2` | `px-5 py-2.5` — consistent with primary |
| **Wizard/modal button padding** | `px-4 py-1.5` | `px-5 py-2` — no more undersized dialog buttons |
| **Modal header padding** | `px-5 py-4` | `px-6 py-4` |
| **Modal body padding** | `px-5 py-5` | `px-6 py-6` |
| **Modal footer padding** | `px-5 py-3` | `px-6 py-3` |
| **Sidebar brand header** | `px-5 py-5` | `px-6 py-6` |
| **Sidebar user footer** | `px-5 py-4` | `px-6 py-5` |
| **Sidebar nav items** | `px-2 py-1.5` | `px-3 py-2` — better click targets |
| **Form field spacing** | `space-y-5` | `space-y-6` — all major forms |
| **Dashboard card grid** | `gap-4 mb-8` | `gap-5 mb-10` |

### Input focus states

Strengthened focus treatment on all text inputs, selects, and textareas across 8 files:

| Property | Before | After |
|----------|--------|-------|
| Border on focus | `focus:border-accent/50` (50% opacity) | `focus:border-accent` (full accent colour) |
| Focus ring | `focus:ring-accent/20` (20% opacity) | `focus:ring-accent/30` (30% opacity) |

### Files changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/assets/styles/main.css` | `--border`, `--border-hover` opacity bump; `:focus-visible` outline width 1px → 2px |
| 2 | `frontend/src/layouts/DefaultLayout.vue` | Page content padding uplift |
| 3 | `frontend/src/components/common/AppSidebar.vue` | Nav item padding, brand header, user footer, `text-[10px]`/`text-[11px]` → `text-xs` |
| 4 | `frontend/src/pages/DashboardPage.vue` | Card `p-5` → `p-6`, grid gap, `text-[10px]` → `text-xs` |
| 5 | `frontend/src/pages/AgentConfigPage.vue` | Card/form padding, button padding, focus states, form spacing, helper text |
| 6 | `frontend/src/pages/NotificationsPage.vue` | Card/form padding, button padding, focus states, form spacing, metadata text sizes |
| 7 | `frontend/src/pages/ConnectionsPage.vue` | Card padding, button padding |
| 8 | `frontend/src/pages/AgentLogPage.vue` | Skeleton card padding |
| 9 | `frontend/src/pages/AgentChatPage.vue` | (inherits token changes) |
| 10 | `frontend/src/pages/ReportsPage.vue` | Skeleton card padding |
| 11 | `frontend/src/pages/LoginPage.vue` | Focus states, form spacing |
| 12 | `frontend/src/pages/OnboardingPage.vue` | Helper text, form spacing |
| 13 | `frontend/src/pages/NotFoundPage.vue` | Button padding |
| 14 | `frontend/src/pages/public/PricingPage.vue` | Badge text size |
| 15 | `frontend/src/components/agent/ChatInput.vue` | Button padding, focus state |
| 16 | `frontend/src/components/agent/ChatMessage.vue` | Role label text size |
| 17 | `frontend/src/components/agent/ChatWindow.vue` | Thinking indicator text size |
| 18 | `frontend/src/components/common/StatusBadge.vue` | Badge text size |
| 19 | `frontend/src/components/common/CopyableField.vue` | Copy button text size |
| 20 | `frontend/src/components/reports/ReportCard.vue` | Card padding, severity badge text size |
| 21 | `frontend/src/components/log/LogEntry.vue` | Severity/source badge text sizes |
| 22 | `frontend/src/components/connections/ConnectionCard.vue` | Card padding, testing badge text size |
| 23 | `frontend/src/components/connections/ConnectionForm.vue` | Card/form padding, button padding, focus states, section header text |
| 24 | `frontend/src/components/connections/ConnectionTestModal.vue` | Modal padding, button padding, detail label text size |
| 25 | `frontend/src/components/connections/GitHubRepoSelector.vue` | Card padding, button padding, branch label text size |
| 26 | `frontend/src/components/connections/BlueprintView.vue` | Hub label text size |
| 27 | `frontend/src/components/connections/BlueprintZone.vue` | Zone header text size |
| 28 | `frontend/src/components/connections/BlueprintNode.vue` | Icon badge, type label, action button text sizes |
| 29 | `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Modal padding, button padding, discard dialog padding |
| 30 | `frontend/src/components/connections/wizard/WizardStepIndicator.vue` | Step label text size |
| 31 | `frontend/src/components/connections/wizard/PlatformGrid.vue` | Category label text size, group spacing |
| 32 | `frontend/src/components/connections/wizard/steps/StepName.vue` | Focus state |
| 33 | `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue` | Focus states |
| 34 | `frontend/src/components/connections/wizard/steps/StepSupabaseAuth.vue` | Focus states |
| 35 | `frontend/src/components/connections/wizard/steps/StepSupabaseTables.vue` | Form spacing |
| 36 | `frontend/src/components/connections/wizard/steps/StepWebhookSetup.vue` | Section header text size |
| 37 | `frontend/src/components/connections/wizard/steps/StepGitHubInstall.vue` | Button padding |
| 38 | `frontend/src/components/public/PublicFooter.vue` | Footer text `text-[11px]` → `text-xs` |
| 39 | `docs/brand-guidelines.md` | Updated border token values, sizing/spacing specifications, design principles |

---

## 0.20.5 — Stale-Asset Reload on Deploy (2026-04-02)

After a deployment, users who already had the site open would hit a blank page on their next navigation. The browser's cached `index.html` referenced code-split chunk filenames from the previous build (e.g. `DashboardPage-BkYIRdWl.js`). Those files no longer exist on the server, and Nginx's `try_files` SPA fallback served `index.html` (text/html) in their place, causing the browser to reject them with a MIME type error.

### Root cause

Vue Router lazy-loads page components via dynamic `import()`. When the target `.js` chunk has been replaced by a new build, the import fails with `Failed to fetch dynamically imported module`. No error handler existed on the router, so the failure surfaced as an unhandled promise rejection caught only by the global `window.unhandledrejection` listener in `main.ts` — which logged the error and showed a generic toast, but left the user stuck.

### Fix

Added a `router.onError` handler that detects dynamic import failures and performs a full page reload via `window.location.assign(to.fullPath)`. The reload fetches the current `index.html` with correct chunk references, and the user lands on the intended page seamlessly.

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/router/index.ts` | Added `router.onError` handler — detects `Failed to fetch dynamically imported module` and `Importing a module script failed` errors, reloads to the target route's `fullPath` |

---

## 0.20.4 — Public Site Header Overlap Fix (2026-04-02)

The fixed navigation bar (`position: fixed`, 76px tall) was removed from document flow but no corresponding space was reserved in the page layout. On tall viewports the hero's vertical centering masked the overlap, but on shorter screens (laptops, tablets, zoomed browsers) the top of page content was clipped behind the navbar.

### Root cause

`PublicLayout.vue` placed `<PublicNav />` and `<RouterView />` as flex siblings. Because the nav is `position: fixed`, it occupies no flow height — every page's content started at `top: 0`, directly under the navbar. The Features and Pricing pages partially compensated with `pt-20` (80px), leaving only 4px of clearance. The Landing page hero used `-mt-24` to nudge the heading upward for visual balance, which pulled it further behind the nav on shorter viewports.

### Fix

Moved the nav offset into `PublicLayout.vue` so it applies globally, then removed per-page workarounds.

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/layouts/PublicLayout.vue` | Wrapped `<RouterView>` in a content div with `pt-[76px]` to reserve space below the fixed nav |
| 2 | `frontend/src/pages/public/LandingPage.vue` | Removed `-mt-24` on the hero text container — no longer needed with correct layout offset |
| 3 | `frontend/src/pages/public/FeaturesPage.vue` | `pt-20` → `pt-8` — layout handles the nav offset, page keeps only section spacing |
| 4 | `frontend/src/pages/public/PricingPage.vue` | `pt-20` → `pt-8` — same |

---

## 0.20.3 — Lumber v0.9.0 Upgrade (2026-04-02)

Upgraded the Lumber log classifier dependency from a pinned commit pseudo-version (`v0.0.0-20260304033652-4f6b6e878057`) to the first tagged release (`v0.9.0`).

The previous version pulled a raw commit hash from `github.com/kaminocorp/lumber` with no semver tag, making builds dependent on an unreleased snapshot. Lumber v0.9.0 is a proper GitHub release with a stable API surface.

### Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/go.mod` | `github.com/kaminocorp/lumber` upgraded to `v0.9.0` |
| 2 | `backend/go.sum` | Updated checksums for new version |

No code changes required — Lumber's public API (`lumber.New`, `ClassifyBatch`, `Close`) is unchanged.

---

## 0.20.2 — Blueprint View & Wizard Guard (2026-03-25)

The Connections page had two UX gaps: the creation wizard silently discarded in-progress data on any accidental close, and all connections were shown as identical rectangles in a flat grid with no sense of infrastructure topology.

### Wizard discard confirmation

Closing the connection wizard (backdrop click, Escape, or X button) now checks for unsaved progress — platform selection, name, config fields, or a partially-created connection. If dirty, an inline overlay asks "Discard changes?" before proceeding. The confirmation renders inside the wizard modal itself (absolute-positioned over the body) to avoid z-index stacking issues and keep the user's in-progress state visible behind the semi-transparent backdrop.

**Changed:** `ConnectionWizard.vue` — added `isDirty` computed, `requestClose()` gatekeeper, `showDiscardConfirm` overlay.

### Architectural blueprint visualization

Replaced the card grid with a visual infrastructure diagram. Heimdall sits at the center as a glowing hub node, with three categorized zones radiating outward:

- **Log Sources** (left) — Supabase, Webhook, Syslog, Datadog connections near a server icon
- **Databases** (right) — PostgreSQL, MySQL connections near a database cylinder icon
- **Integrations** (bottom-left) — GitHub and other generic connections near a code brackets icon

SVG dashed bezier lines connect the hub to each zone, computed dynamically via `ResizeObserver` + `getBoundingClientRect()` and drawn in with a `stroke-dashoffset` animation on mount. Empty zones show dashed-border prompts ("+ Add a log source") that open the wizard on click.

A **Blueprint / List toggle** in the page header lets users switch between the new diagram and the original card grid. Preference persists to `localStorage`.

Category mapping imports directly from `flows.ts` — adding a new connector type automatically places it in the correct zone.

**Responsive:** Mobile stacks vertically (hub → zones), SVG lines hidden.

**New files:** `ViewToggle.vue`, `BlueprintNode.vue`, `BlueprintZone.vue`, `BlueprintView.vue`.
**Changed:** `ConnectionsPage.vue` — view toggle, conditional rendering, `@add` event wiring.

---

## 0.20.1 — Action Button Color (2026-03-22)

Primary action buttons (+ New Connection, Authenticate, Save, Send, Continue, etc.) used the same muted feldgrau accent (`#4d5d53`) as ambient UI elements, making them blend in rather than stand out as calls to action.

Added a new `--action` design token (`#3b8a5a`) — same green hue family but ~3× the saturation — and applied it to all 14 CTA buttons across 11 files. The existing `accent` palette is unchanged and continues to serve borders, badges, focus rings, and surface tints.

### Design Tokens

| Token | Value | Purpose |
|-------|-------|---------|
| `--action` | `#3b8a5a` | Primary CTA background |
| `--action-hover` | `#449e66` | CTA hover state |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/assets/styles/main.css` | New `--action` / `--action-hover` tokens + Tailwind `@theme` registration |
| 2 | `frontend/src/pages/ConnectionsPage.vue` | + New Connection button |
| 3 | `frontend/src/pages/LoginPage.vue` | Authenticate button |
| 4 | `frontend/src/pages/OnboardingPage.vue` | Get Started button |
| 5 | `frontend/src/pages/AgentConfigPage.vue` | Save Configuration button |
| 6 | `frontend/src/pages/NotificationsPage.vue` | Save + Add/Update Channel buttons |
| 7 | `frontend/src/components/agent/ChatInput.vue` | Send button |
| 8 | `frontend/src/components/connections/ConnectionForm.vue` | Add Connection + Install GitHub App buttons |
| 9 | `frontend/src/components/connections/GitHubRepoSelector.vue` | Save Selection button |
| 10 | `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Done + Continue buttons |
| 11 | `frontend/src/components/connections/wizard/steps/StepGitHubInstall.vue` | Install GitHub App button |

---

## 0.20.0 — Security & Production Hardening (2026-03-22)

Five rounds of hardening (code assessment → polish → production hardening × 2 → final polish) bringing the codebase from "works" to "production-ready". Covers security fixes, correctness bugs, resource leak prevention, and operational robustness. No new features — every change is a fix or improvement to existing code.

### Security Fixes

**SQL injection in Supabase connector (critical)** — `pollTable` interpolated user-controlled table names directly into a SQL string sent to the Supabase Management API. Added an allowlist of the 6 known Supabase log tables, validated at both parse time (`NewSupabase`) and poll time (defense-in-depth).

**XSS in email notifications (critical)** — `FormatEmailHTML()` interpolated LLM-generated text and user-controlled values directly into HTML. All 6 interpolated values now use `html.EscapeString()`.

**`search_logs` tool ignored the query parameter (critical)** — The agent's `search_logs` tool declared a required `query` parameter but never read it. Every "search" was actually "list recent logs", making investigation fundamentally broken. Added `SearchLogsByUser` and `SearchLogsByUserAndSeverity` SQL queries with `ILIKE` matching, plus a `LIKE` wildcard escape helper (`escapeLike`) to prevent `%` and `_` from being interpreted as wildcards.

**Wildcard CORS policy** — `Access-Control-Allow-Origin: *` allowed any website to make authenticated API calls. CORS is now origin-checked against an allowlist from the `CORS_ALLOWED_ORIGINS` env var (falls back to localhost in dev). All CORS headers are scoped to matching origins only.

**Missing RLS on 6 tables** — `organizations`, `applications`, `monitoring_state`, `notification_channels`, `notification_preferences`, and `notification_log` had no Row Level Security policies. Migration 021 enables RLS with `app_current_user_id()` policies, idempotent `DROP POLICY IF EXISTS` guards, and optimised joins. Two missing FK indexes added (`notification_log.channel_id`, `agent_log.conversation_id`).

### Correctness Fixes

**Interactive chat used wrong config source** — `RunConversation` loaded the global `agent_config` singleton instead of the per-app `app_agent_configs` row. Per-app model selection and system prompt overrides configured via the UI were ignored during chat.

**Log pagination broken for combined sources** — `source=all` fetched `limit` rows from each source with the same `offset`, producing inconsistent pages. Now fetches `offset + limit` from each, merges, then applies offset/limit to the merged result. Offset capped at 10,000 to prevent O(offset) memory growth.

**Pagination null vs empty array** — When offset exceeded results, Go serialised a nil slice as `null` instead of `[]`, crashing the frontend. Both the handler and store now guarantee `[]`.

**Pagination counts ignored filters** — `CountLogsByUser` counted all logs regardless of severity/connection filters. Added `CountLogsByUserAndSeverity` and `CountLogsByUserAndConnection` filtered count queries.

**Cursor desynchronisation** — The Supabase connector advanced its cursor based on all rows from the API response, including those whose `InsertLogEntry` failed. Failed rows were permanently skipped — silent data loss. Cursor now only advances for successfully inserted rows.

**Partial-create on Supabase poller failure** — If `NewSupabase()` failed after the DB insert was committed, the client received HTTP 400 but the connection persisted in the database. Config is now validated eagerly *before* the DB insert.

**`UpdateConnection` didn't restart poller** — Editing a Supabase connection's config left the old polling goroutine running. The poller is now stopped and restarted on update.

**`Poll()` always returned nil** — The `PollConnector` interface defines an error return, but the Supabase implementation always returned nil. Now returns the first error encountered across table polls.

### Resource & Lifecycle Fixes

**`os.Exit(1)` in server goroutine** — `ListenAndServe` error triggered immediate process exit, skipping all deferred cleanup (pool, classifier, poller). Replaced with an error channel; the main goroutine selects on both signal and error channels.

**Transaction commit used cancellable context** — If the HTTP client disconnected before `defer done()`, `tx.Commit(ctx)` failed with `context canceled`, silently rolling back successful writes. Now uses `context.WithoutCancel(ctx)` for the commit. Commit failures are logged via `slog.Error`.

**No poll timeout** — `conn.Poll()` received a context with no deadline. A hung upstream could block the goroutine forever. Each poll now runs with `context.WithTimeout(ctx, 2×interval)` (min 30s).

**No graceful drain on shutdown** — `StopAll()` cancelled contexts but returned immediately. Added `sync.WaitGroup` so in-flight polls complete before shutdown proceeds.

**Poller panics on zero interval** — `time.NewTicker(0)` panics. Added `minPollInterval` (5s) clamp.

**Response body size limit** — `io.ReadAll(resp.Body)` on the Supabase API response had no cap. Wrapped with `io.LimitReader` at 10 MB.

**Rate limit blocked goroutine** — When `X-RateLimit-Remaining: 0`, the code slept for the entire reset window (potentially hours). Replaced with log-and-continue; the poller naturally retries on its next tick.

**Rate limit held HTTP connection** — Body was read *after* the backoff sleep. Reordered to read body → check status → sleep, releasing the TCP connection before waiting.

**HTTP idle connections accumulated** — `Close()` was a no-op. Now calls `httpClient.CloseIdleConnections()`.

**Timer leaks** — `setInterval` in `StepTest`, `ConnectionTestModal`, and `ConnectionCard` was never cleared on unmount. `setTimeout` in `CopyableField` had the same issue. All now have `onBeforeUnmount` cleanup.

### Frontend Fixes

**Orphaned connection on wizard abandon** — If the user reached the test step and then closed the wizard, the connection persisted in the database. `handleClose()` now deletes the connection (best-effort) before emitting `close`.

**One-way prop sync in wizard steps** — `StepName`, `StepSupabaseAuth`, `StepSupabaseTables`, and `StepPostgresConfig` copied props on mount but never reacted to parent resets. Added inward watchers.

**Port validation gap** — `StepPostgresConfig` reported itself as valid without checking the port range. Invalid ports silently defaulted to 5432. Now validates 1–65535 before enabling Continue.

**Delete confirmation** — `ConnectionCard` delete was a single click with no confirmation. Added two-click pattern with 3-second auto-reset.

**Action button visibility** — Hidden on touch/keyboard devices. Added `focus-within:opacity-100`.

**Clipboard fallback** — `CopyableField` used `navigator.clipboard.writeText()` which requires HTTPS. Added `document.execCommand('copy')` fallback.

**Escape key handling** — Added to `ConnectionWizard` and `ConnectionTestModal`.

### Cleanup

- **Structured logging** — 7 `log.Printf` calls in connections handler replaced with `slog.Error` with structured attributes.
- **Dead code** — Removed unused `statusColor` computed in `ConnectionTestModal`.
- **Duplicate table list** — Extracted `supabaseLogTables` to `flows.ts`, imported by both `StepSupabaseTables` and `ConnectionForm`.
- **`splitCSV`** — Hand-rolled 12-line CSV parser in CORS middleware replaced with `strings.Split` + `TrimSpace`.
- **Missing `sb.Close()`** — `TestConnection` handler's Supabase case never closed the connector.
- **Input validation** — Added allowlists for connection `type`, `direction`, and `status` fields.

### Deployment Notes

**Environment variable required:** `CORS_ALLOWED_ORIGINS` must be set in production (comma-separated frontend origins). If unset, only localhost is allowed.

**Migration required:** `make migrate-up` to apply migration 021 (RLS policies + FK indexes).

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/logs/supabase.go` | Table allowlist, body size limit, cursor desync fix, rate limit non-blocking, body read reorder, close idle connections, return firstErr |
| 2 | `backend/internal/connectors/logs/supabase_test.go` | `TestNewSupabase_InvalidTableName` |
| 3 | `backend/internal/connectors/poller.go` | Poll timeout, interval clamp, WaitGroup drain |
| 4 | `backend/internal/api/handlers/connections.go` | Eager validation, poller restart on update, input allowlists, structured logging, `defer sb.Close()` |
| 5 | `backend/internal/api/handlers/userqueries.go` | `context.WithoutCancel` for commit, commit error logging, documentation |
| 6 | `backend/internal/api/handlers/logs.go` | Merged-source pagination, offset cap, filtered counts, null-safe response |
| 7 | `backend/internal/api/middleware/cors.go` | Origin allowlist, scoped headers, `strings.Split` |
| 8 | `backend/internal/agent/loop.go` | Per-app config in `RunConversation` |
| 9 | `backend/internal/agent/tools_logs.go` | Read `query` param, `escapeLike`, dispatch to search queries |
| 10 | `backend/internal/notifications/format.go` | `html.EscapeString` on all interpolated values |
| 11 | `backend/internal/db/queries/log_buffer.sql` | `SearchLogsByUser`, `SearchLogsByUserAndSeverity`, filtered counts, `ESCAPE '\'` |
| 12 | `backend/internal/db/log_buffer.sql.go` | Auto-generated by sqlc |
| 13 | `backend/cmd/heimdall/main.go` | Error channel replaces `os.Exit(1)` |
| 14 | `backend/migrations/021_rls_missing_tables.up.sql` | RLS policies for 6 tables, 2 FK indexes, idempotent guards |
| 15 | `backend/migrations/021_rls_missing_tables.down.sql` | Reverse migration |
| 16 | `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Orphan cleanup, Escape key |
| 17 | `frontend/src/components/connections/wizard/flows.ts` | Exported `supabaseLogTables` |
| 18 | `frontend/src/components/connections/wizard/steps/StepTest.vue` | Timer cleanup |
| 19 | `frontend/src/components/connections/wizard/steps/StepName.vue` | Inward prop sync |
| 20 | `frontend/src/components/connections/wizard/steps/StepSupabaseAuth.vue` | Inward prop sync |
| 21 | `frontend/src/components/connections/wizard/steps/StepSupabaseTables.vue` | Import shared table list, inward prop sync |
| 22 | `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue` | Port validation, inward prop sync |
| 23 | `frontend/src/components/connections/ConnectionTestModal.vue` | Timer cleanup, Escape key, remove dead code |
| 24 | `frontend/src/components/connections/ConnectionCard.vue` | Delete confirmation, timer cleanup, focus-within visibility |
| 25 | `frontend/src/components/connections/ConnectionForm.vue` | Import shared table list, edit-mode table selection fix |
| 26 | `frontend/src/components/common/CopyableField.vue` | Clipboard fallback, timer cleanup |

---

## 0.19.0 — Connection Wizard (2026-03-22)

The dropdown-based connection creation form is replaced with a guided multi-step wizard. Users now pick a platform from a visual grid, then walk through platform-specific steps (name → auth → config → test). This is a pure frontend change — the backend and API are identical.

### Wizard Architecture

The wizard uses a declarative flow system. Each platform defines its steps as data in `flows.ts`:

| Platform | Steps | Connector Type |
|----------|-------|----------------|
| Supabase | Name → Auth → Tables → Test | `supabase` |
| PostgreSQL | Name → Config → Test | `postgres` |
| Webhook | Name → Setup | `webhook_logs` |
| GitHub | Name → Install | `github` |

Coming-soon platforms (Datadog, Syslog, MySQL) appear in the grid with `opacity-40` and a "Soon" badge. Adding a new connector type = adding a flow definition + step components, zero wizard shell changes.

### Wizard Shell

Full-screen modal at `wizard/ConnectionWizard.vue` orchestrating the creation flow:

- **Platform selection** — `PlatformGrid` shows available connectors grouped by category (Log Sources, Databases, Generic)
- **Step progression** — Dynamic `<component :is="...">` renders the current step. Each step emits `valid` to control the Continue button.
- **Create-before-test** — For flows with a test step, the connection is created on the server before advancing to test. `StepTest` then calls the real `POST /connections/{id}/test` endpoint.
- **Transitions** — Steps slide left/right with opacity fade (200ms ease-out in, 150ms ease-in out)
- **Local state** — `WizardState` is a local `reactive()` object, not Pinia. Ephemeral, self-cleaning on modal close.

### Step Components

**Shared across all flows:**

| Component | Purpose |
|-----------|---------|
| `StepName` | Connection name input, validates non-empty |
| `StepTest` | Live connection test with elapsed timer, reuses `ConnectionTestModal` visual pattern |

**Supabase-specific:**

| Component | Purpose |
|-----------|---------|
| `StepSupabaseAuth` | Project reference + PAT inputs, info callout about Management API |
| `StepSupabaseTables` | Checkbox group for 6 log tables (defaults: `postgres_logs`, `auth_logs`), poll interval presets |

**Existing connector types:**

| Component | Purpose |
|-----------|---------|
| `StepPostgresConfig` | Host, port, database, username, password, SSL mode (2-column grid for host/port) |
| `StepWebhookSetup` | Endpoint URL format + payload example. Uses `CopyableField` for copy-to-clipboard. |
| `StepGitHubInstall` | GitHub App install button redirecting to GitHub OAuth |

### Supporting Components

**CopyableField** (`common/CopyableField.vue`) — Monospace code display with clipboard button. Used by `StepWebhookSetup` for webhook URLs and tokens.

**WizardStepIndicator** — Dot-line progress bar (`● ─── ○ ─── ○`). Active: `bg-accent`, completed: `bg-accent/60`, future: `bg-border`.

**PlatformGrid + PlatformCard** — Card grid grouped by category. Each card shows a 2-letter icon badge, name, and description.

### ConnectionsPage Integration

- **"+ New Connection"** button opens the wizard modal
- **`ConnectionForm`** preserved for editing existing connections (inline, not modal)
- Step components are lazy-loaded via `defineAsyncComponent` to keep the initial bundle small

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/components/connections/wizard/flows.ts` | Flow definitions and types |
| 2 | `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Wizard shell / orchestrator |
| 3 | `frontend/src/components/connections/wizard/WizardStepIndicator.vue` | Dot-line progress bar |
| 4 | `frontend/src/components/connections/wizard/PlatformGrid.vue` | Categorised platform card grid |
| 5 | `frontend/src/components/connections/wizard/PlatformCard.vue` | Individual platform card |
| 6 | `frontend/src/components/connections/wizard/steps/StepName.vue` | Name input (shared) |
| 7 | `frontend/src/components/connections/wizard/steps/StepTest.vue` | Live connection test (shared) |
| 8 | `frontend/src/components/connections/wizard/steps/StepSupabaseAuth.vue` | Supabase auth fields |
| 9 | `frontend/src/components/connections/wizard/steps/StepSupabaseTables.vue` | Supabase table selection |
| 10 | `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue` | PostgreSQL config fields |
| 11 | `frontend/src/components/connections/wizard/steps/StepWebhookSetup.vue` | Webhook setup info |
| 12 | `frontend/src/components/connections/wizard/steps/StepGitHubInstall.vue` | GitHub App install |
| 13 | `frontend/src/components/common/CopyableField.vue` | Copy-to-clipboard field |
| 14 | `frontend/src/pages/ConnectionsPage.vue` | Wire wizard modal, preserve form for editing |

---

## 0.18.0 — Supabase Connector (2026-03-22)

Heimdall can now ingest logs from Supabase projects. Users create a Supabase connection with a Personal Access Token and project reference, and Heimdall polls the Supabase Management API on a configurable interval (15–60 seconds), inserting logs into `log_buffer` where the existing monitoring loop classifies and escalates them. No Supabase plan restrictions — works with the free tier.

### Backend — Polling Connector

**`PollConnector` interface** (`connectors/connector.go`) — New interface alongside `StreamConnector` (push-based) and `QueryConnector` (on-demand). Polling is fundamentally pull-based — the connector initiates HTTP requests on a timer. `Poll(ctx, queries)` takes `*db.Queries` so it can call `InsertLogEntry` directly.

**Supabase connector** (`connectors/logs/supabase.go`) — Calls `GET /v1/projects/{ref}/analytics/endpoints/logs.all` with a SQL query per table. Supports 6 log tables: `postgres_logs`, `auth_logs`, `edge_logs`, `function_logs`, `storage_logs`, `realtime_logs`.

| Behaviour | Detail |
|-----------|--------|
| **Config** | `project_ref` (required), `access_token` (required), `poll_tables` (defaults to `["postgres_logs"]`), `poll_interval_secs` (defaults to 30, min 15) |
| **Connect/Health** | Validates PAT by running `SELECT 1` against the analytics endpoint |
| **Poll** | For each table: query since cursor → parse → insert into `log_buffer` → advance cursor. Continues polling other tables if one fails. |
| **Cursors** | In-memory per-table, initialised to `now() - 5 minutes` on startup |
| **Severity** | Derived from metadata fields (`error_severity`, `severity`, `level`) |

**Polling loop manager** (`connectors/poller.go`) — Manages one goroutine per active poll-based connection. `Start()` launches a goroutine on a ticker, firing once immediately then on each interval. `Stop()` cancels a single connection. `StopAll()` cancels all (called on shutdown).

### Backend — Server Wiring

- `Server` struct gains a `Poller` field, passed through `NewServer()` and `NewRouter()`
- `main.go` creates the `Poller`, calls `resumePollers()` on startup (queries `ListActiveConnectionsByType("supabase")` to restart polling for existing connections), and calls `poller.StopAll()` during shutdown
- `CreateConnection` starts the poller for new Supabase connections
- `DeleteConnection` stops the poller before deleting
- New sqlc query: `ListActiveConnectionsByType`

### Frontend — Connection Form & Card

**ConnectionForm** — Added `supabase` type with config fields (project reference, PAT, poll interval select) and a checkbox group for poll tables. Selecting Supabase auto-locks direction to `one_way`. Helper text explains the Management API and PAT generation.

**ConnectionCard** — Human-readable type labels (`supabase` → "Supabase"). Supabase cards show "Polling N tables" instead of the direction.

### Tests — 24 New Tests

**Backend (16 tests):**

| File | Tests | Coverage |
|------|-------|----------|
| `connectors/logs/supabase_test.go` | 12 | Config parsing, defaults, validation, Connect success/401/404, severity derivation, cursor advancement, Close, Health, rate limit 429 |
| `connectors/poller_test.go` | 4 | Start/stop, StopAll, replace existing, stop non-existent |

Uses `httptest.NewServer` to mock the Supabase API via an injectable `apiBase` field.

**Frontend (8 tests):**

| File | Tests | Coverage |
|------|-------|----------|
| `ConnectionForm.test.ts` | 8 | Supabase type renders, config fields, 6 table checkboxes, defaults, auto direction, submit payload, helper text, edit mode |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/connectors/connector.go` | Added `PollConnector` interface |
| 2 | `backend/internal/connectors/logs/supabase.go` | Supabase Management API polling connector |
| 3 | `backend/internal/connectors/logs/supabase_test.go` | 12 unit tests |
| 4 | `backend/internal/connectors/poller.go` | Goroutine-per-connection polling loop manager |
| 5 | `backend/internal/connectors/poller_test.go` | 4 unit tests |
| 6 | `backend/internal/api/handlers/server.go` | Added `Poller` field, updated `NewServer()` |
| 7 | `backend/internal/api/router.go` | Updated `NewRouter()` to accept `*connectors.Poller` |
| 8 | `backend/internal/api/handlers/connections.go` | Supabase test/create/delete handling |
| 9 | `backend/internal/api/handlers/testhelpers_test.go` | Added `Poller` to test Server struct |
| 10 | `backend/cmd/heimdall/main.go` | Create poller, resume on startup, shutdown |
| 11 | `backend/internal/db/queries/connections.sql` | Added `ListActiveConnectionsByType` query |
| 12 | `backend/internal/db/connections.sql.go` | Auto-generated by sqlc |
| 13 | `frontend/src/components/connections/ConnectionForm.vue` | Supabase type, config fields, poll_tables checkbox group, helper text, direction lock |
| 14 | `frontend/src/components/connections/ConnectionCard.vue` | Type labels, Supabase subtitle |
| 15 | `frontend/src/components/connections/__tests__/ConnectionForm.test.ts` | 8 component tests |

---

## 0.17.3 — Custom Dropdown Component (2026-03-21)

Every dropdown in the app used native HTML `<select>` elements. While the trigger could be styled with Tailwind, the dropdown panel itself is rendered by the operating system — meaning a jarring white menu appeared over the near-black techno-brutalist UI. This patch replaces all 10 native selects with a single reusable `BaseSelect` component that matches the design system end-to-end.

### BaseSelect Component

New component at `components/common/BaseSelect.vue` providing a fully custom dropdown:

- **Visual design** — dark `bg-bg-elevated` background, `border-border` borders, `font-mono` text, accent highlight on the selected option (`text-accent-bright bg-accent-subtle`), hover state (`bg-bg-surface-hover`), and a chevron indicator that rotates on open
- **Keyboard navigation** — Arrow Up/Down to move focus, Enter/Space to select, Escape to close
- **Click outside to close** — document-level click listener, cleaned up on unmount
- **Smooth transitions** — fade + slide animation on open/close via Vue `<Transition>`
- **Two sizes** — `default` for form fields (matching `px-3 py-2 text-sm`) and `sm` for compact contexts like filters and the sidebar (matching `px-3 py-1.5 text-xs`)
- **Disabled state** — reduces opacity and blocks interaction, matching existing input disabled styling

### Replacements

All 10 native `<select>` elements across 5 files were replaced:

| File | Selects | Context |
|------|---------|---------|
| `AppSidebar.vue` | 1 | Application switcher in the sidebar |
| `ConnectionForm.vue` | 3 | Connection type, direction, and dynamic config fields (SSL mode, protocol) |
| `LogFilters.vue` | 3 | Source, severity, and connection filter dropdowns |
| `NotificationsPage.vue` | 2 | Severity threshold and notification channel type |
| `AgentConfigPage.vue` | 1 | Monitoring mode selector (continuous/periodic/off) |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/components/common/BaseSelect.vue` | New reusable dropdown component |
| 2 | `frontend/src/components/common/AppSidebar.vue` | Native select → `BaseSelect` for app switcher |
| 3 | `frontend/src/components/connections/ConnectionForm.vue` | 3 native selects → `BaseSelect` (type, direction, config fields) |
| 4 | `frontend/src/components/log/LogFilters.vue` | 3 native selects → `BaseSelect` (source, severity, connection) |
| 5 | `frontend/src/pages/NotificationsPage.vue` | 2 native selects → `BaseSelect` (threshold, channel type) |
| 6 | `frontend/src/pages/AgentConfigPage.vue` | 1 native select → `BaseSelect` (monitoring mode) |

---

## 0.17.2 — SPA Routing Fix (2026-03-21)

Refreshing the browser on any sub-route (e.g. `/dashboard`, `/connections`, `/chat`) returned a Vercel 404 page. Navigating to root or opening a fresh tab worked because Vercel serves `index.html` for `/` automatically — but it had no instruction to do the same for deeper paths.

### Symptom

A page refresh on any route other than `/` produced:

```
404: NOT_FOUND
Code: NOT_FOUND
```

Closing the tab and reopening the app from root worked normally, because Vue Router handled all subsequent navigation client-side.

### Root Cause

The `vercel.json` configuration had rewrites for `/api/*` and `/ws/*` (proxying to the Fly.dev backend), but no **SPA fallback** for all other paths. When Vercel received a request for `/dashboard`, it looked for a matching file or directory, found nothing, and returned 404.

This is the standard SPA hosting problem: client-side routing relies on the History API to change the URL without a server round-trip, but a hard refresh or direct navigation sends a real HTTP request that the server must resolve to `index.html`.

### Fix

Added a catch-all rewrite as the **last rule** in `vercel.json`:

```json
{ "source": "/(.*)", "destination": "/index.html" }
```

Order is critical — Vercel evaluates rewrites top-to-bottom. The `/api/*` and `/ws/*` rules match first and proxy to the backend. Static assets (JS, CSS, images) are served from the build output before rewrites are consulted. Only truly unmatched paths (i.e. frontend routes) fall through to the catch-all, which serves `index.html` and lets Vue Router resolve the route client-side.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/vercel.json` | Added SPA catch-all rewrite `/(.*) → /index.html` |

---

## 0.17.1 — Connection Test Modal & Dashboard Fix (2026-03-21)

Connection testing was a black box — the UI showed a brief banner with a generic message and no detail on *why* a test failed. This patch adds a modal that shows the test in real time and surfaces the actual error, plus fixes a dashboard crash for new apps with no logs.

### Connection Test Modal

After creating, editing, or pinging a connection, a modal now overlays the page showing:

- **Connection metadata** — name, type, host, port, database, user, SSL mode — so you can immediately verify what's being tested
- **Live test status** — pulsing indicator with elapsed timer while the test runs
- **Result** — green success or red failure with the **full error message** from the backend

Previously, a failed Postgres connection test returned a generic `"Failed to connect to database"`. The backend now includes the underlying error (e.g., `hostname resolving error: lookup https on [fdaa::3]:53: no such host`), making misconfigurations immediately diagnosable without tailing server logs.

### Dashboard Null Guard

The dashboard crashed with `Cannot read properties of null (reading 'slice')` when the logs API returned `null` instead of an empty array (happens for newly onboarded apps with zero logs). Fixed at both layers:

- **Store** (`logs.ts`): `entries.value = data.data ?? []` — prevents null from entering the store
- **Consumer** (`DashboardPage.vue`): `logsStore.entries?.slice(0, 8) ?? []` — defensive guard in the computed property

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/connections.go` | Include actual error in test failure response via `fmt.Sprintf` |
| 2 | `frontend/src/components/connections/ConnectionTestModal.vue` | New modal component — test phases, connection metadata, elapsed timer |
| 3 | `frontend/src/pages/ConnectionsPage.vue` | Wire modal into create/edit/ping flows |
| 4 | `frontend/src/pages/DashboardPage.vue` | Null guard on `recentEntries` computed |
| 5 | `frontend/src/stores/logs.ts` | Null coalesce on API response `data.data` |

---

## 0.17.0 — RLS Session Variable Fix (2026-03-15)

The persistent `GET /api/logs` 500 that v0.16.1 made diagnosable is now fixed. The root cause was a PostgreSQL protocol incompatibility in `UserQueries` — the `SET LOCAL` statement doesn't support parameterised values (`$1`) under the extended query protocol that pgx uses by default.

### Symptom

After logging in, the dashboard showed "Server error — please try again" with a 500 on `GET /api/logs?source=all&limit=50&offset=0`. The v0.16.1 error logging surfaced the actual error in Fly.io logs:

```
ERROR database error error="ERROR: syntax error at or near \"$1\" (SQLSTATE 42601)"
```

Every endpoint that called `UserQueries()` was broken — `/api/logs`, `/api/conversations`, `/api/connections` CRUD, `/api/auth/me`, and `/ws/chat`. The app-scoped endpoints (`/api/apps/{id}/connections`, `/agent/config`, `/stats`, `/monitoring/status`) were unaffected because they use `s.Queries` directly with `authorizeApp()`, bypassing `UserQueries()` entirely.

### Root Cause

`UserQueries` (`userqueries.go`) opens a transaction and sets a PostgreSQL session variable for RLS policy evaluation:

```go
// Before (broken)
tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", userID.String())
```

PostgreSQL's `SET` is a **utility statement**, not a DML statement. It doesn't go through the parser's parameter-binding stage. When pgx sends this via the **extended query protocol** (its default), PostgreSQL receives the literal text `SET LOCAL app.current_user_id = $1` with a separate parameter value — but the `SET` parser doesn't know how to bind `$1`, so it throws `SQLSTATE 42601` (syntax error).

This likely started failing when the `DATABASE_URL` began routing through **Supavisor** (Supabase's connection pooler). Direct PostgreSQL connections can fall back to the simple query protocol for utility statements, but Supavisor enforces the extended protocol consistently.

### Fix

Replaced `SET LOCAL` with PostgreSQL's `set_config()` function — a regular SQL function that fully supports parameter binding in the extended protocol:

```go
// After (fixed)
tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID.String())
```

`set_config(name, value, is_local)` is the function-based equivalent of `SET LOCAL`. The third argument `true` scopes the setting to the current transaction, identical to `SET LOCAL` semantics. Because it's a standard function call (not a utility statement), pgx can bind `$1` normally.

### Why `set_config` over `SET LOCAL`

| | `SET LOCAL ... = $1` | `set_config($1, $2, true)` |
|---|---|---|
| Protocol | Utility statement — no param binding | Regular function — full param binding |
| pgx compatibility | Fails under extended protocol | Works under all protocols |
| Connection pooler safety | Breaks through Supavisor | Works through any pooler |
| Injection risk | Forces string interpolation as workaround | Native parameterisation, no interpolation needed |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/userqueries.go` | `SET LOCAL` → `set_config()` with parameterised binding |

---

## 0.16.1 — Server-Side Error Logging (2026-03-15)

A `GET /api/logs` 500 surfaced in production with no server-side trace — the `jsonError` helper was sending generic messages to clients but silently discarding the actual Go `err`. This patch closes that observability gap across all handlers.

### Why

Every `jsonError(w, "failed to ...", 500)` call swallowed the real error. The logging middleware only captured method, path, status, and duration — no error details, no query params. When the `/api/logs` endpoint returned 500 after a fresh onboarding, the Fly.io logs showed `status=500` but nothing about *why*. Diagnosing required reading source code and guessing.

### Changes

**New helper — `jsonServerError(w, message, err)`** (`helpers.go`)

Logs the actual error via `slog.Error` before sending the generic JSON response to the client. Separates the two concerns: safe client messages vs. full internal diagnostics for operators.

**44 replacements across 9 handler files**

Every `jsonError(w, "...", http.StatusInternalServerError)` call that had an `err` in scope was replaced with `jsonServerError(w, "...", err)`. Non-500 errors (400, 401, 404, 409) are unchanged.

| File | Replacements |
|------|-------------|
| `logs.go` | 5 |
| `connections.go` | 8 |
| `github.go` | 11 |
| `applications.go` | 7 |
| `organizations.go` | 6 |
| `notifications.go` | 7 |
| `conversations.go` | 2 |
| `stats.go` | 2 |
| `auth.go` | 1 |

**Logging middleware upgrade** (`middleware/logging.go`)

- 500+ responses now log at `ERROR` level (was `INFO` for all statuses)
- Query parameters are included for any 4xx/5xx response

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/helpers.go` | Added `jsonServerError` helper |
| 2 | `backend/internal/api/handlers/logs.go` | 5 error paths → `jsonServerError` |
| 3 | `backend/internal/api/handlers/connections.go` | 8 error paths → `jsonServerError` |
| 4 | `backend/internal/api/handlers/github.go` | 11 error paths → `jsonServerError` |
| 5 | `backend/internal/api/handlers/applications.go` | 7 error paths → `jsonServerError` |
| 6 | `backend/internal/api/handlers/organizations.go` | 6 error paths → `jsonServerError` |
| 7 | `backend/internal/api/handlers/notifications.go` | 7 error paths → `jsonServerError` |
| 8 | `backend/internal/api/handlers/conversations.go` | 2 error paths → `jsonServerError` |
| 9 | `backend/internal/api/handlers/stats.go` | 2 error paths → `jsonServerError` |
| 10 | `backend/internal/api/handlers/auth.go` | 1 error path → `jsonServerError` |
| 11 | `backend/internal/api/middleware/logging.go` | 500s → `slog.Error`; query params on 4xx/5xx |

---

## 0.16.0 — GitHub App Integration (2026-03-11)

Heimdall can now read your code. Connect a GitHub organization via a first-party GitHub App, select which repositories the agent can access, and Heimdall gains a `search_codebase` tool — code search, file reading, and tree listing — available in both interactive chat and the monitoring loop. When the agent investigates an anomaly, it can now trace errors back to the source.

### Why

Heimdall could search logs and query databases, but had no way to look at the code behind the systems it monitors. When the monitoring agent flagged an error spike, it could describe *what* happened but not *why* — it couldn't inspect the handler that returned 500s, the config that changed, or the migration that ran. GitHub App integration closes that gap: the agent can now correlate runtime behavior with source code.

### Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│  GitHub App Installation Flow                                     │
│                                                                   │
│  Frontend                    Backend                   GitHub     │
│  ────────                    ───────                   ──────     │
│  "Install GitHub App" ──→ GET /github/install                     │
│                           (generate state JWT) ──→ redirect to    │
│                                                   github.com/apps │
│                           ←── GET /github/callback ←── redirect   │
│                           (validate state JWT,                    │
│                            verify installation,                   │
│                            create connection)                     │
│  /connections?github=installed ←── 302 redirect                   │
│  (auto-open repo selector)                                        │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│  Agent Tool: search_codebase                                      │
│                                                                   │
│  Agent Loop ──→ Dispatch("search_codebase", {action, query, ...}) │
│                   │                                               │
│                   ├─ ListEnabledGitHubReposByApp(appID)           │
│                   ├─ Create codebase.GitHub connector              │
│                   ├─ Connect() → InstallationTokenFor() [cached]  │
│                   └─ Query() ──→ GitHub API                       │
│                        ├─ search_code  → GET /search/code         │
│                        ├─ read_file    → GET /repos/.../contents  │
│                        └─ list_tree    → GET /repos/.../git/trees │
└──────────────────────────────────────────────────────────────────┘
```

**Key design decisions:**

- **GitHub App, not PATs.** Organization-scoped installation with fine-grained repo permissions. Tokens are short-lived (1 hour) and cached in-memory with a 5-minute refresh margin. No long-lived secrets stored per-user.
- **State JWT for OAuth callback.** The install flow redirects through GitHub and back. The callback authenticates via a signed RS256 JWT (15-minute expiry, nonce) embedded in the `state` parameter — no session cookie required.
- **Nil-safe client.** If `GITHUB_APP_ID` is unset, the client is `nil` and the entire integration is disabled. All consumers check for nil before use, matching the existing `notifier` pattern.
- **Connector, not direct API.** The GitHub connector implements `QueryConnector`, the same interface as the Postgres connector. The agent tool doesn't know it's talking to GitHub — it marshals an action and calls `Query()`.

### Database (Migration 020)

One new table:

- **`github_repos`** — Per-connection repository tracking. Links a `connection_id` to a GitHub `repo_id` with an `enabled` toggle. Unique index on `(connection_id, repo_id)` for upsert semantics. RLS policy scopes access via the parent connection's `user_id`.

4 sqlc queries: list by connection, upsert, delete, and list enabled repos by app (JOIN through connections for the agent tool).

### Backend — GitHub Client (`internal/github/client.go`)

Core GitHub App authentication:

- **RSA private key** loading from env var (raw PEM) or file path
- **JWT generation** — RS256, `iss` = App ID, 10-minute expiry per GitHub spec
- **Installation token cache** — `sync.RWMutex`-protected map, keyed by installation ID, auto-refreshes 5 minutes before expiry
- **Authenticated API requests** — `APIRequest()` with `context.Context`, `X-GitHub-Api-Version` header, 10-second timeout

### Backend — GitHub Connector (`internal/connectors/codebase/github.go`)

Implements `QueryConnector` with three actions:

| Action | GitHub API | Guardrails |
|--------|-----------|------------|
| `search_code` | `GET /search/code` | 20 results max, 20 `repo:` qualifiers max, `url.Values` encoding |
| `read_file` | `GET /repos/.../contents` | Skip >1MB, truncate >50KB, binary detection, per-segment path escaping |
| `list_tree` | `GET /repos/.../git/trees?recursive=1` | 5,000 entries max, 10-second timeout |

All API calls route through `ghClient.APIRequest()` for consistent headers, timeouts, and `io.LimitReader` (5MB cap).

### Backend — Agent Tool Integration

- **`tools.go`** — `search_codebase` added to `ToolRegistry()` and `Dispatch()`. Breaking change: `Dispatch` signature now includes `appID uuid.UUID` for app-scoped tool access.
- **`tools_codebase.go`** — Loads enabled repos from DB, creates connector, executes query. Returns JSON results following the error-as-tool-result pattern.
- **`loop.go`** — `RunConversation` accepts `appID`; monitoring mode passes `appConfig.AppID`.
- **`prompt.go`** — Both system prompts updated to mention `search_codebase`.

### Backend — API Endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/api/github/install?app_id={id}` | JWT | Returns GitHub App install URL with signed state |
| `GET` | `/api/github/callback` | State JWT | Handles post-install redirect, creates/updates connection |
| `GET` | `/api/connections/{id}/github/repos` | JWT | Lists repos from GitHub API, merged with DB enabled state |
| `PUT` | `/api/connections/{id}/github/repos` | JWT | Upserts repo enabled/disabled state |

`TestConnection` extended to handle `type="github"` — verifies installation token validity.

### Frontend

- **`ConnectionForm.vue`** — GitHub type shows "Install GitHub App" button instead of config fields. Redirects browser to GitHub install URL.
- **`GitHubRepoSelector.vue`** — Fetches repos, shows toggle checkboxes with branch badges, scrollable list (max 320px), save/cancel.
- **`ConnectionCard.vue`** — "Repos" button for GitHub connections.
- **`ConnectionsPage.vue`** — Handles `?github=installed` redirect (success banner, auto-opens repo selector).
- **`useWebSocket.ts`** — Added `appId` to WebSocket options, passed as `?app_id=` query param.
- **`useAgent.ts`** — Passes `currentAppId` from Pinia store to WebSocket connection.

### Production Hardening (3 rounds, 20 fixes)

| Round | Fixes | Key Items |
|-------|-------|-----------|
| 1 (7.5→8.5) | 7 | RLS policy on `github_repos`, configurable app slug, error propagation on duplicate check, user-scoped queries for repo operations, dedicated HTTP client, context propagation, pagination bound |
| 2 (8.5→9) | 8 | **Callback route moved to public group** (was blocked by JWT middleware), search query double-encoding fix, `context.Context` on all GitHub API methods, DB error returns 500 (not silent continue), `io.LimitReader` on request body, unified connector HTTP client, per-segment path escaping |
| 3 (9→9.5) | 5 | Callback `UserQueries()` consistency, search query length bound (20 repos), `io.ReadAll` error check, dispatch routing tests for `search_codebase` |

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GITHUB_APP_ID` | No | — | GitHub App ID (numeric). If empty, integration is disabled. |
| `GITHUB_PRIVATE_KEY` | No | — | RSA private key (PEM string or file path) |
| `GITHUB_CLIENT_ID` | No | — | GitHub App client ID |
| `GITHUB_APP_SLUG` | No | `heimdall-agent` | GitHub App URL slug |
| `GITHUB_WEBHOOK_SECRET` | No | — | Webhook secret (reserved for future use) |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/github/client.go` | New — GitHub App client: JWT generation, installation token cache, authenticated API requests |
| 2 | `backend/internal/connectors/codebase/github.go` | Rewrite — QueryConnector with search_code, read_file, list_tree via GitHub API |
| 3 | `backend/internal/connectors/codebase/github_test.go` | Updated — nil client error test |
| 4 | `backend/internal/api/handlers/github.go` | New — InstallGitHub, GitHubCallback, ListGitHubRepos, UpdateGitHubRepos, TestGitHubConnection |
| 5 | `backend/internal/api/handlers/server.go` | Add `GitHub *github.Client` field to Server struct |
| 6 | `backend/internal/api/handlers/connections.go` | TestConnection handles `type="github"` |
| 7 | `backend/internal/api/handlers/chat.go` | Parse `app_id` from WebSocket query param, pass to RunConversation |
| 8 | `backend/internal/api/router.go` | Register GitHub routes; callback in public group, install in protected group |
| 9 | `backend/internal/agent/agent.go` | Add `githubClient` field; updated New() signature |
| 10 | `backend/internal/agent/tools.go` | Add `search_codebase` to ToolRegistry and Dispatch; appID parameter |
| 11 | `backend/internal/agent/tools_codebase.go` | New — toolSearchCodebase implementation |
| 12 | `backend/internal/agent/loop.go` | RunConversation accepts appID; monitoring passes appConfig.AppID |
| 13 | `backend/internal/agent/prompt.go` | Both system prompts mention search_codebase |
| 14 | `backend/internal/agent/tools_test.go` | Add SearchCodebase dispatch tests; updated existing tests with appID |
| 15 | `backend/internal/config/config.go` | Add 5 GitHub env vars |
| 16 | `backend/cmd/heimdall/main.go` | Conditional GitHub client init; pass to agent and router |
| 17 | `backend/migrations/020_github_repos.up.sql` | New — github_repos table with RLS |
| 18 | `backend/migrations/020_github_repos.down.sql` | New — drop table |
| 19 | `backend/internal/db/queries/github_repos.sql` | New — 4 queries |
| 20 | `backend/internal/db/github_repos.sql.go` | Regenerated — sqlc |
| 21 | `backend/internal/db/models.go` | Regenerated — GithubRepo model |
| 22 | `frontend/src/types/github.ts` | New — GitHubRepo interface |
| 23 | `frontend/src/api/github.ts` | New — getGitHubInstallURL, listGitHubRepos, updateGitHubRepos |
| 24 | `frontend/src/components/connections/GitHubRepoSelector.vue` | New — repo toggle list with save/cancel |
| 25 | `frontend/src/components/connections/ConnectionForm.vue` | GitHub type shows install button instead of config fields |
| 26 | `frontend/src/components/connections/ConnectionCard.vue` | "Repos" button for GitHub connections |
| 27 | `frontend/src/components/connections/ConnectionList.vue` | manage-repos event passthrough |
| 28 | `frontend/src/pages/ConnectionsPage.vue` | GitHub installed banner, auto-open repo selector |
| 29 | `frontend/src/composables/useWebSocket.ts` | Add appId to WebSocket options |
| 30 | `frontend/src/composables/useAgent.ts` | Pass currentAppId to WebSocket |

---

## 0.15.1 — Notifications Build Fix (2026-03-11)

Fixed a `vue-tsc` build failure in `NotificationsPage.vue` caused by inline `as` type assertions in the template. Vue's template compiler doesn't support TypeScript cast syntax — expressions like `(ch.config as { recipients: string[] }).recipients.join(', ')` produce parse errors during `vue-tsc -b`.

Extracted a `channelConfigSummary()` helper in the `<script setup>` block that performs the same type narrowing, replacing the two `<template v-if/v-else>` branches with a single `{{ channelConfigSummary(ch) }}` interpolation.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/pages/NotificationsPage.vue` | Add `channelConfigSummary()` helper; simplify channel config display in template |

---

## 0.15.0 — Notifications & Escalation (2026-03-11)

Heimdall can now alert you when it finds something. When the monitoring agent assesses flagged logs as `warning`, `error`, or `critical`, Heimdall dispatches notifications through configured channels — Email (via Resend), Slack (incoming webhook), or Discord (webhook). A monitoring agent that can't reach anyone is a smoke detector with no alarm.

### Why

Phase 8 gave Heimdall the ability to see — the monitor loop classifies logs, escalates to Claude, and writes assessments to the agent log. But findings stayed locked inside the dashboard. Users had to check Heimdall to learn something was wrong, which defeats the purpose of autonomous monitoring. Phase 9 closes the loop: Heimdall now speaks.

### Architecture

```
Monitor loop emits agent_log entry (severity ≥ threshold)
   ↓
┌─────────────────────────────────────┐
│  Notification Dispatcher            │  ← In-process, async goroutine
│  1. Load app notification prefs     │
│  2. Apply severity threshold filter │
│  3. Apply cooldown (dedup window)   │
│  4. Fan out to enabled channels     │
└─────────────┬───────────────────────┘
              │
     ┌────────┼────────────┐
     │        │            │
   Email    Slack       Discord
  (Resend   (Webhook)   (Webhook)
   API)
     │        │            │
     └────────┼────────────┘
              │
       Write to notification_log
       (delivery tracking)
```

**Key design decisions:**

- **Fire-and-forget.** Notification dispatch runs in a goroutine with `context.WithoutCancel()` — a failed or slow notification never blocks the monitor loop or cursor advance. Matches the existing `EmitLog` pattern.
- **Per-app cooldown.** A single cooldown window (default 15 minutes) suppresses all channels for an app. Prevents a noisy app from flooding every channel every 15 seconds.
- **Single retry.** One retry on failure, then mark as `failed`. No exponential backoff — monitoring is continuous, so the next cycle will re-notify if the issue persists.
- **Channel interface.** Same factory pattern as the `Classifier` interface. Each channel type is isolated, testable, and swappable.

### Database (Migrations 017–019)

Three new tables:

- **`notification_channels`** — Per-app, multiple channels. Type (`email`/`slack`/`discord`), name, JSONB config (recipients for email, webhook URL for Slack/Discord), enabled flag.
- **`notification_preferences`** — Per-app 1:1. Master enabled toggle, severity threshold (`info`/`warning`/`error`/`critical`), cooldown minutes. Defaults: disabled, warning threshold, 15-minute cooldown.
- **`notification_log`** — Every notification attempt tracked. Status (`pending`/`sent`/`failed`), error message, FK to `agent_log` for correlation, FK to `notification_channels` for channel metadata.

12 sqlc queries across 3 files: channel CRUD + enabled-only listing, preference get/upsert, log insert/update/list/last-sent.

### Backend — Notification Package

New `internal/notifications/` package:

- **`notifier.go`** — `Channel` interface, `NewChannel` factory, `Dispatcher` orchestrator (preference check → severity filter → cooldown → fan-out → delivery logging).
- **`slack.go`** — Slack Block Kit payload: header with severity emoji, section with summary, code block with full assessment, context with timestamp. Text truncated to safe limits (2000/2900 chars) with UTF-8-aware slicing.
- **`discord.go`** — Discord embed: severity-mapped colour (red/orange/yellow/blue), summary + assessment in description, timestamp footer. Truncated to 1800 chars.
- **`email.go`** — Resend API (`POST https://api.resend.com/emails`). HTML template with inline styles, severity badge, assessment block. Requires `RESEND_API_KEY` and `NOTIFICATION_FROM_EMAIL` env vars (optional — email channel only works if configured).
- **`format.go`** — Shared helpers: `SeverityEmoji()`, `SeverityColor()`, `FormatEmailHTML()`.

### Backend — Integration

- **`agent.go`** — Agent struct gains `notifier *notifications.Dispatcher` field. Passed from `main.go` at startup. Nil-safe — tests and non-notification environments skip dispatch.
- **`emit.go`** — `emitLog` and `EmitLogWithSeverity` now return `uuid.UUID` (the inserted `agent_log.id`). Returns `uuid.Nil` on error. Existing callers unaffected.
- **`monitor.go`** — After `EmitLogWithSeverity`, calls `go a.notifier.Notify(...)` in a goroutine with the agent log ID, app name, severity, summary, and assessment.
- **`config.go`** — Two new optional env vars: `RESEND_API_KEY`, `NOTIFICATION_FROM_EMAIL`.

### Backend — API Endpoints

8 new endpoints under `/apps/{appId}/notifications`, all using `authorizeApp()`:

| Method | Path | Handler |
|--------|------|---------|
| GET | `/notifications/preferences` | Returns preferences (sensible defaults if none set) |
| PUT | `/notifications/preferences` | Validates severity enum, cooldown 1–1440 |
| GET | `/notifications/channels` | Lists all channels for app |
| POST | `/notifications/channels` | Validates type, config shape per type |
| PUT | `/notifications/channels/{channelId}` | Verifies channel belongs to app |
| DELETE | `/notifications/channels/{channelId}` | Verifies channel belongs to app |
| POST | `/notifications/channels/{channelId}/test` | Sends synthetic test notification |
| GET | `/notifications/history` | Paginated notification log with channel metadata |

Webhook URL validation uses proper `url.Parse` + domain/path checks (not prefix matching) to prevent spoofing.

### Frontend

- **`NotificationsPage.vue`** — Three sections: preferences form (enable toggle, severity dropdown, cooldown input), channels list (add/edit/delete/test), recent notification history table with status badges. Watches `currentAppId` for app switches.
- **`api/notifications.ts`** — 8 API client functions matching all endpoints.
- **`types/notification.ts`** — TypeScript interfaces for preferences, channels, channel configs, log entries.
- **Router** — `/notifications` route added, lazy-loaded.
- **Sidebar** — "Notifications" nav item added under Agent section.

### Production Hardening (Step 9)

8 fixes applied during review:

1. **Context cancellation** — Notification goroutines used the monitor context, which cancelled before sends completed. Fixed with `context.WithoutCancel()`.
2. **UTF-8 truncation** — Slack/Discord text truncation could slice mid-codepoint. Fixed with `utf8.Valid` boundary checking.
3. **Silent DB errors** — `UpdateNotificationLogStatus` errors were swallowed. Now logged.
4. **JSON null vs empty array** — Empty channel/history lists returned `null` instead of `[]`. Fixed with slice initialisation.
5. **HTTP client timeouts** — Webhook client had no timeout. Added 10-second timeout.
6. **Truncate panic** — `truncateText` panicked if max < 4. Added bounds check.
7. **Webhook URL spoofing** — Prefix-based URL validation could be bypassed (`hooks.slack.com.evil.com`). Replaced with `url.Parse` + explicit host matching.
8. **Response body errors** — Missing error checks on `io.ReadAll` for error response bodies. Fixed with `io.LimitReader`.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/migrations/017_notification_channels.up.sql` | New — channels table + index |
| 2 | `backend/migrations/017_notification_channels.down.sql` | New — drop table |
| 3 | `backend/migrations/018_notification_preferences.up.sql` | New — preferences table |
| 4 | `backend/migrations/018_notification_preferences.down.sql` | New — drop table |
| 5 | `backend/migrations/019_notification_log.up.sql` | New — log table + indices |
| 6 | `backend/migrations/019_notification_log.down.sql` | New — drop table |
| 7 | `backend/internal/db/queries/notification_channels.sql` | New — 6 queries |
| 8 | `backend/internal/db/queries/notification_preferences.sql` | New — 2 queries |
| 9 | `backend/internal/db/queries/notification_log.sql` | New — 4 queries |
| 10 | `backend/internal/db/notification_channels.sql.go` | Regenerated — sqlc |
| 11 | `backend/internal/db/notification_preferences.sql.go` | Regenerated — sqlc |
| 12 | `backend/internal/db/notification_log.sql.go` | Regenerated — sqlc |
| 13 | `backend/internal/db/models.go` | Regenerated — 3 new model structs |
| 14 | `backend/internal/notifications/notifier.go` | New — Channel interface, factory, Dispatcher |
| 15 | `backend/internal/notifications/slack.go` | New — Slack Block Kit webhook |
| 16 | `backend/internal/notifications/discord.go` | New — Discord embed webhook |
| 17 | `backend/internal/notifications/email.go` | New — Resend API email |
| 18 | `backend/internal/notifications/format.go` | New — shared formatting helpers |
| 19 | `backend/internal/agent/agent.go` | Add notifier field to Agent struct |
| 20 | `backend/internal/agent/emit.go` | Return uuid.UUID from emitLog/EmitLogWithSeverity |
| 21 | `backend/internal/agent/monitor.go` | Call notifier.Notify after emit |
| 22 | `backend/internal/config/config.go` | Add Resend env vars |
| 23 | `backend/internal/api/handlers/notifications.go` | New — 8 handlers with validation |
| 24 | `backend/internal/api/router.go` | Register notification routes |
| 25 | `backend/cmd/heimdall/main.go` | Create dispatcher, pass to agent |
| 26 | `frontend/src/types/notification.ts` | New — TypeScript interfaces |
| 27 | `frontend/src/api/notifications.ts` | New — API client functions |
| 28 | `frontend/src/pages/NotificationsPage.vue` | New — preferences, channels, history |
| 29 | `frontend/src/router/index.ts` | Add /notifications route |
| 30 | `frontend/src/components/common/AppSidebar.vue` | Add Notifications nav item |

---

## 0.14.5 — Logo & Favicon (2026-03-11)

Created the official Heimdall logo — a six-pointed forked starburst with a centre eye dot. Hybrid of Concept B's hexagonal symmetry and a bold split-ray graphic style.

### Why

Heimdall had no brand mark — just a placeholder `.ico` and concept explorations in `docs/logos/`. Needed a minimal, distinctive icon that reads at favicon scale and reinforces the surveillance/all-seeing-eye identity.

### Design

- **6 forked rays** at 60° intervals (hexagonal symmetry, Bifrost connection). Each ray is two diverging prongs with angled tips — the outer edge extends further, creating directional tension.
- **Centre dot** (r=5.5) acts as the pupil — the all-seeing eye motif.
- **Void ring** between the dot and prong starts provides breathing room and reads clearly at small sizes.
- **12 prongs total**, all generated from the same base polygon with rotation transforms.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `docs/logos/heimdall-logo.svg` | New — monochrome logo with `currentColor` fill (CSS-tintable) |
| 2 | `docs/logos/heimdall-logo-preview.svg` | New — preview on dark background with labels |
| 3 | `frontend/public/favicon.svg` | New — white on black, square (browser applies its own clipping) |
| 4 | `frontend/index.html` | SVG favicon as primary, `.ico` as fallback |

---

## 0.14.4 — Feldgrau Colour Theme (2026-03-11)

Shifted the entire colour palette from vivid sage green to feldgrau — a desaturated, military grey-green inspired by German field uniforms.

### Why

The original accent (`#5a9e6a`) read as "forest / nature" — too organic for a surveillance-themed monitoring tool. Feldgrau (`#4d5d53`) drops saturation from ~40% to ~12%, producing a steely grey-green that reinforces the techno-brutalist, command-terminal aesthetic.

### Changes

- **Design tokens** — All 15 CSS custom properties in `:root` updated: accent, accent-hover, accent-bright, accent-subtle, accent-border, border, border-hover, status-ok. Background and text tokens shifted from green undertones to neutral grey-green.
- **Scrollbar, selection, focus ring, glow** — Hardcoded `rgba(90, 158, 106, …)` values in `main.css` replaced with `rgba(77, 93, 83, …)`.
- **HeroMesh pixel renderer** — Scan glow RGB blend target updated from `(90, 158, 106)` to `(77, 93, 83)`. Anomaly halo additive tints rebalanced for the lower-saturation palette.
- **LoginPage grid pattern** — Background grid lines updated to feldgrau RGBA.

### Colour Mapping

| Token | Before | After |
|-------|--------|-------|
| `--accent` | `#5a9e6a` | `#4d5d53` |
| `--accent-hover` | `#4a8c5a` | `#5a6e62` |
| `--accent-bright` | `#7ab889` | `#6e8578` |
| `--bg-primary` | `#060806` | `#070808` |
| `--text-secondary` | `#8a9a8a` | `#8a938e` |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/assets/styles/main.css` | All design tokens + hardcoded RGBA values shifted to feldgrau |
| 2 | `frontend/src/components/public/HeroMesh.vue` | Scan glow and halo RGB values updated |
| 3 | `frontend/src/pages/LoginPage.vue` | Grid pattern RGBA updated |

---

## 0.14.3 — Public Layout, Heading & Mesh Refinement (2026-03-10)

Extracted a shared public layout, fixed the scrollbar-induced nav shift, refined the hero heading, and rebuilt the mesh renderer for ultra-high density.

### Why

The nav bar shifted horizontally when navigating between pages with and without scrollbars. Each public page independently imported `PublicNav` and `PublicFooter`, meaning any change required touching three files. The hero heading ("The all-seeing eye") was poetic but vague — didn't communicate what the product does. The mesh needed higher density and better text legibility.

### Shared Public Layout

- **`PublicLayout.vue`** — New layout wrapper renders `PublicNav`, a `<RouterView />` slot, and `PublicFooter`. All public pages now get identical nav/footer from one source.
- **Nested routes** — Public routes restructured as children of a parent layout route in `router/index.ts`. `meta: { public: true }` lives on the parent only.
- **Auth guard fix** — Changed `to.meta?.public` to `to.matched.some(r => r.meta.public)` so child routes inherit the parent's public flag. Without this, the guard would redirect unauthenticated users to login on all public pages.
- **Stripped nav/footer** — Removed `PublicNav` and `PublicFooter` imports and rendering from `LandingPage`, `FeaturesPage`, and `PricingPage`.

### Scrollbar Layout Shift Fix

- **`scrollbar-gutter: stable`** on `html` — Reserves scrollbar gutter space on all pages, even when content doesn't overflow. Eliminates the ~15px nav shift when navigating between non-scrolling (landing) and scrolling (Platform, Pricing) pages.

### Hero Heading

- **"Autonomous system surveillance."** — Replaced "The all-seeing eye." with a blunt, techno-brutalist statement that explicitly describes what Heimdall does. Two lines: "Autonomous system" (white) / "surveillance." (accent green).

### Mesh Renderer Rebuild

- **107,520 dots** (420 × 256 grid, up from 13,500) — Ultra-fine density where individual dots are imperceptible; the surface reads as a woven material.
- **ImageData pixel writing** — Replaced Canvas `arc()` draw calls with direct RGBA writes to an `ImageData` buffer, `putImageData` once per frame. Single draw call regardless of point count — necessary for 108k points at 60fps.
- **Single-pixel dots** — Each point is one pixel at DPR resolution. At this density, the grid structure itself creates the surface texture.
- **Camera repositioned** — Mesh pushed to the lower portion of the hero (`CAMERA_Y` raised to 140). Top gradient extended (opaque to 20%, transparent at 65%) so heading and subtitle sit on clean dark background.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/layouts/PublicLayout.vue` | New — shared layout with nav + RouterView + footer |
| 2 | `frontend/src/router/index.ts` | Nested public routes under layout; `to.matched.some()` auth guard fix |
| 3 | `frontend/src/assets/styles/main.css` | Added `scrollbar-gutter: stable` to html |
| 4 | `frontend/src/components/public/HeroMesh.vue` | Rebuilt: 420×256 grid, ImageData renderer, repositioned camera |
| 5 | `frontend/src/pages/public/LandingPage.vue` | New heading, gradient tuning, removed nav/footer |
| 6 | `frontend/src/pages/public/FeaturesPage.vue` | Removed nav/footer (provided by layout) |
| 7 | `frontend/src/pages/public/PricingPage.vue` | Removed nav/footer (provided by layout) |

---

## 0.14.2 — Hero Mesh Animation (2026-03-10)

Added an animated 3D wireframe mesh to the landing page hero section, themed around real-time log monitoring. Shortened the headline to a punchier tagline.

### Why

The landing page had no visual hook — just text on a flat dark background. The mesh gives the page a distinctive, high-end feel while reinforcing what Heimdall does: watching data streams and detecting anomalies.

### Animated Mesh (`HeroMesh.vue`)

- **3D wireframe surface** — 150 x 90 dot grid (13,500 points) with thin connecting lines between adjacent dots, perspective-projected onto a Canvas 2D context.
- **Data stream flow** — Base sine waves travel right-to-left, evoking a real-time log timeline.
- **Anomaly peaks** — 5 Gaussian peaks that drift slowly across the surface and pulse in amplitude. Represent incidents rising above the noise floor.
- **Scan sweep** — A green band (`#5a9e6a`) sweeps continuously across the mesh. Flat areas get a faint tint; anomaly peaks glow brightly with an outer halo when the sweep passes — Heimdall detecting something.
- **Depth-aware rendering** — Per-frame depth range calculation drives alpha fade, dot sizing, and line opacity. Anomaly peaks get physically larger dots.
- **Performance** — Pre-allocated 2D point array (zero per-frame allocations), DPR-capped at 2x, `prefers-reduced-motion` respected.

### Hero Copy

- **Headline** — Changed from three-line "Watches everything / Investigates automatically / Reports what matters" to **"The all-seeing eye."** (two lines, references Heimdall's Norse mythology origin).
- **Subtitle** — Condensed to two sentences: "Autonomous AI monitoring for production systems. Watches 24/7. Investigates anomalies. Reports what matters."
- **Gradient overlays** — Vertical and horizontal gradient fades blend the mesh edges into the dark background for text legibility.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/components/public/HeroMesh.vue` | New — animated 3D wireframe mesh canvas component |
| 2 | `frontend/src/pages/public/LandingPage.vue` | Integrated mesh background, shortened headline, gradient overlays |

---

## 0.14.1 — Public Site Header & Footer Redesign (2026-03-10)

Redesigned the public website header and footer to match the Elephantasm design language — full-width, compact, typographic.

### Header

- **Full-width layout** — Removed `max-w-6xl` container; nav now stretches edge-to-edge.
- **Typographic brand** — Replaced SVG hexagon logo with spaced-out `H E I M D A L L` wordmark.
- **True-centered nav** — Navigation links use absolute positioning to center in the viewport independent of brand/actions widths.
- **Three nav items** — Platform (was Features), Pricing, Security.
- **Outlined CTA** — "Get Started" button changed from filled green to outlined accent border with hover fill. GitHub button gets matching outlined treatment with icon + label.
- **Uppercase throughout** — All nav text uses uppercase + wide tracking.

### Footer

- **Single-line, full-width** — Collapsed from two-variant component (inline vs. full with logo and columns) into one compact bar.
- **Three-zone layout** — Left: "A Kamino Corporation product." / Center: copyright + middot-separated links (Terms, Privacy, Platform, Pricing, Security, GitHub, X icon) / Right: italic quote.
- **X icon** — Replaced text "X" with the X/Twitter SVG logo.
- **Added Terms & Privacy** — Placeholder links at `/terms` and `/privacy`.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/components/public/PublicNav.vue` | Full-width, typographic brand, centered nav, outlined buttons, uppercase, Security link |
| 2 | `frontend/src/components/public/PublicFooter.vue` | Single-line three-zone footer, removed variant system, X icon, Terms/Privacy links |
| 3 | `frontend/src/pages/public/LandingPage.vue` | Larger hero text (8xl), uppercase, wider container |
| 4 | `frontend/src/pages/public/FeaturesPage.vue` | Updated `pt-` offset for taller nav |
| 5 | `frontend/src/pages/public/PricingPage.vue` | Updated `pt-` offset for taller nav |

---

## 0.14.0 — Dockerfile Model Fix (2026-03-09)

Fixed deployment failure caused by the Lumber ONNX model Dockerfile stage pointing at a deleted HuggingFace repo. Also added missing model files that could cause silent runtime failures.

### Why

`fly deploy` failed with `curl: (22) The requested URL returned error: 401` during the Docker build. The `2_Dense/model.safetensors` download URL pointed at `Snowflake/mdbr-leaf-mt`, a HuggingFace repo that no longer exists (404). The Lumber library's own Makefile uses `MongoDB/mdbr-leaf-mt` as the correct source — both repos (`onnx-community/mdbr-leaf-mt-ONNX` and `MongoDB/mdbr-leaf-mt`) are public and require no authentication.

### Changes

- **Fixed broken model URL (P0)** — Replaced `Snowflake/mdbr-leaf-mt` with `MongoDB/mdbr-leaf-mt` for the `2_Dense/model.safetensors` projection layer download. The Snowflake repo has been deleted.
- **Added missing model files (P1)** — Dockerfile was missing `model_quantized.onnx_data` (external data tensor), `tokenizer_config.json`, and `2_Dense/config.json`, all of which the Lumber Makefile downloads. Their absence could cause silent classifier failures at runtime, falling back to PassthroughClassifier.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/Dockerfile` | Fixed `2_Dense` URL from Snowflake → MongoDB; added 3 missing model file downloads |

---

## 0.13.0 — Phase 8 Hardening (2026-03-07)

Four rounds of review fixes across the full Phase 8 implementation. 21 issues identified and resolved, including 6 P0s covering security, data integrity, and resource management.

### Why

Phase 8 introduced the largest architectural change since scaffolding — multi-app data model, Lumber classifier, autonomous monitor loop, and 15 new API endpoints. Each review round stress-tested a different layer: authorization correctness, production resilience, API contract consistency, and edge-case safety.

### Round 1 — Authorization & Bounds (5 fixes)

- **Per-app authorization gap (P0)** — All `/api/apps/{appId}/*` handlers parsed `appId` from the URL without verifying the app belonged to the user's org. Added `authorizeApp` helper backed by `GetApplicationByOrgUser` query (JOINs applications → users on `org_id`). Every per-app endpoint now goes through this check.
- **Non-transactional onboarding (P1)** — `POST /api/onboard` ran 4 sequential queries (create org, link user, create app, upsert config). Partial failure left orphaned state. Wrapped in `pool.Begin()` + `Queries.WithTx()`.
- **Verbose monitoring prompt (P1)** — Replaced Variant A with token-optimized Variant B. Enforces `Severity: <level>` output format for reliable parsing.
- **Unbounded flagged log payload (P1)** — No limit on payload size or batch count sent to Claude. Added `maxPayloadChars = 2000` (per log) and `maxFlaggedForLLM = 50` (per cycle). Rune-safe truncation.
- **Semaphore blocking tick loop (P2)** — One hung Claude API call could block the entire tick cycle. Added `monitorAppTimeout = 2 minutes` per-app context deadline.

### Round 2 — Production Readiness (3 fixes)

- **CreateConnection missing app authorization (P0)** — Accepted `app_id` in request body without verifying it belonged to the user's org. A user could inject logs into a foreign org's app. Added `GetApplicationByOrgUser` check.
- **TestConnection error information leakage (P1)** — Raw Postgres error strings (containing hostnames, IPs) returned to clients. Replaced with generic messages; raw errors logged server-side only.
- **App.vue init failure (P1)** — No try-catch around `auth.init()` / `app.init()`. If either threw, the app froze on "Initializing..." forever. Added fallback redirect to login.

### Round 3 — API Contract Cleanup (9 fixes)

- **GetDashboardStats auth bypass (P0)** — Discarded `ok` from `UserIDFromContext`, could proceed without authenticated user. Added 401 guard.
- **ConnectionsPage wrong data source (P0)** — Called legacy `store.fetchConnections()` (user-scoped) instead of `listConnectionsByApp(appId)`. Page always showed wrong app's connections.
- **Legacy routes removed (P1)** — Deleted superseded endpoints: `GET /api/stats`, `GET/PUT /api/agent/config`, `POST /api/agent/run`. Removed associated test file `agent_test.go`.
- **UpdateConnectionStatus error silenced (P1)** — Silent `_ =` discard after test. Replaced with `log.Printf`.
- **Onboarding idempotency (P1)** — `POST /api/onboard` could be called multiple times. Added `org_id` check; returns 409 Conflict if already onboarded.
- **Mode enum validation (P2)** — `UpdateAppAgentConfig` accepted any string for `mode`. Added switch validation: only `continuous`, `periodic`, `off` accepted; returns 400 otherwise.
- **String building inefficiency (P2)** — `formatFlaggedLogs` used byte append. Replaced with `strings.Builder`.
- **Dashboard error accumulation (P2)** — Multiple sequential API calls each overwrote a single error variable. Changed to error array with joined display.

### Round 4 — Edge-Case Safety (4 fixes)

- **emitLog drops entire log on marshal failure (P0)** — If `json.Marshal(detail)` failed, the function returned early — permanently losing the summary, severity, and entry type. Fixed to write with `nil` detail instead.
- **UTF-8 truncation corruption (P0)** — 5 truncation sites used byte slicing (`summary[:200]`, `payload[:2000]`), which splits multi-byte characters. All converted to rune-safe truncation: `string([]rune(s)[:n])` with `utf8.RuneCountInString()` length checks. Affects `loop.go` (3 sites) and `monitor.go` (2 sites).
- **Agent.Start() double-start goroutine leak (P0)** — Calling `Start()` twice overwrote the `cancel` function, orphaning the first goroutine. Added guard: if `cancel != nil`, call `Stop()` first.

### Deferred to Phase 9

| Priority | Item | Location |
|----------|------|----------|
| P1 | Org slug format validation (regex + max length) | `handlers/organizations.go` |
| P1 | `SystemPromptOverride` max-length | `handlers/applications.go` |
| P1 | `UpdateConnectionStatus` SQL lacks `user_id` scope | `queries/connections.sql` |
| P1 | Missing indexes: `applications(status)`, `connections(app_id, status)` | New migration |
| P1 | Logs store doesn't filter by `app_id` | `frontend/src/stores/logs.ts` |
| P2 | Classifier thread-safety for concurrent ONNX inference | `classifier_lumber.go` |
| P2 | Dead code: `store.fetchConnections()` | `stores/connections.ts` |
| P2 | Hardcoded model names in frontend datalist | `AgentConfigPage.vue` |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/agent.go` | Double-start guard in `Start()` |
| 2 | `backend/internal/agent/emit.go` | Continue with nil detail on marshal failure |
| 3 | `backend/internal/agent/loop.go` | Rune-safe truncation at 3 sites |
| 4 | `backend/internal/agent/monitor.go` | Rune-safe truncation (2 sites), `strings.Builder`, payload/batch caps, per-app timeout, monitoring prompt Variant B |
| 5 | `backend/internal/agent/prompt.go` | Token-optimized monitoring prompt |
| 6 | `backend/internal/api/handlers/applications.go` | `authorizeApp` helper, mode enum validation |
| 7 | `backend/internal/api/handlers/connections.go` | App authorization on create, generic error messages, status update logging |
| 8 | `backend/internal/api/handlers/organizations.go` | Transactional onboarding, idempotency guard |
| 9 | `backend/internal/api/handlers/stats.go` | Auth check fix |
| 10 | `backend/internal/api/router.go` | Legacy routes removed |
| 11 | `backend/internal/db/queries/applications.sql` | `GetApplicationByOrgUser` query |
| 12 | `frontend/src/App.vue` | Init error handling with login fallback |
| 13 | `frontend/src/pages/ConnectionsPage.vue` | App-scoped connection fetching |
| 14 | `frontend/src/pages/DashboardPage.vue` | Error accumulation fix |

---

## 0.12.0 — Multi-App UI & API (2026-03-07)

Surfaced the multi-app data model in the API and frontend. Added org onboarding flow, application selector, per-app agent configuration, monitoring status dashboard, and new agent log entry badges. Created handler tests for the new organizational and application endpoints.

### Why

The 0.10.0 data model and 0.11.0 monitoring loop had no user-facing surface. Users couldn't create organizations, switch between apps, or see monitoring status. This release wires the entire multi-app model to the API layer and frontend, making it operational end-to-end.

### Backend — API Endpoints

15 new endpoints organized into three route groups:

| Method | Route | Purpose |
|--------|-------|---------|
| `GET` | `/api/org` | Get authenticated user's organization |
| `POST` | `/api/onboard` | Create org + first app + config in one transaction |
| `GET` | `/api/apps` | List apps in user's org |
| `POST` | `/api/apps` | Create app (auto-creates default agent config) |
| `GET` | `/api/apps/{appId}` | Get single app (with org authorization) |
| `GET` | `/api/apps/{appId}/connections` | List connections for app |
| `GET` | `/api/apps/{appId}/agent/config` | Get per-app agent config |
| `PUT` | `/api/apps/{appId}/agent/config` | Update per-app agent config |
| `GET` | `/api/apps/{appId}/monitoring/status` | Monitoring mode, interval, last check, running state |
| `GET` | `/api/apps/{appId}/stats` | Per-app dashboard stats (log count, connections) |

Additional query changes:
- `ListConnectionsByApp(app_id)` — connections scoped to application
- `GetAppDashboardStats(app_id)` — per-app stats replacing old user-scoped stats
- `CreateConnection` updated to require `app_id`
- `GetFirstUserInOrg(org_id)` — resolves a user for agent_log attribution in monitoring

### Frontend — Onboarding Flow

New `OnboardingPage.vue` — linear flow that creates an organization and first application in a single step:
1. Organization name → auto-generates slug on blur
2. Organization slug (editable)
3. First application name
4. Submits to `POST /api/onboard`; redirects to dashboard

Router guard detects `needsOnboarding` (user has no `org_id`) and redirects unauthenticated or un-onboarded users appropriately. `App.vue` calls `app.init()` post-authentication to load org/app state.

### Frontend — Application Management

- **`useAppStore` (Pinia)** — central store for org, applications list, and `currentAppId` (persisted to `localStorage`). Provides `init()`, `onboard()`, `createApp()`, and computed `currentApp`.
- **AppSidebar** — application selector dropdown in sidebar with org name in footer. Switching apps triggers reactive data re-fetch across all pages.
- **DashboardPage** — 4-column grid: Monitoring card (mode, status dot, last check time, logs processed/flagged ratio), Agent card, Connections card, Ingestion card. Watches `currentAppId` for re-fetch.
- **AgentConfigPage** — per-app config: model selector (datalist), mode toggle (continuous/periodic/off), interval presets (30s, 1m, 5m, 15m) + custom seconds input, system prompt override.
- **ConnectionsPage** — app-scoped list via `listConnectionsByApp`. Form injects `currentAppId` on create.
- **LogEntry** — new badges: `Monitor` (amber) for `monitoring` entries, `Heartbeat` (green) for `heartbeat` entries.

### Frontend — Types & API Layer

| File | Contents |
|------|----------|
| `types/organization.ts` | `Organization`, `Application`, `AppAgentConfig`, `MonitoringStatus`, `OnboardingPayload`, `OnboardingResponse` |
| `types/connection.ts` | Updated `Connection` with `app_id`; `CreateConnectionPayload` includes `app_id` |
| `api/organizations.ts` | `getOrganization()`, `onboard()` |
| `api/applications.ts` | `listApplications()`, `createApplication()`, `getAppAgentConfig()`, `updateAppAgentConfig()`, `getMonitoringStatus()`, `getAppStats()`, `listConnectionsByApp()` |

### Test Coverage

Test infrastructure overhauled: `testSetup` creates full org → app → config hierarchy. `testEnv` struct extended with `OrgID`, `AppID`. New `createTestConnection` helper.

| File | Tests | Coverage |
|------|-------|---------|
| `organizations_test.go` (new) | 4 | GetOrganization, Onboard, Onboard_DuplicateSlug, Onboard_MissingFields |
| `applications_test.go` (new) | 11 | CRUD, per-app config, monitoring status, stats, connections by app |
| `connections_test.go` (updated) | 6 | All creates include `app_id`, new `TestCreateConnection_MissingAppID` |
| `logs_test.go` (updated) | 3 | Connection creation includes `app_id` |
| `webhooks_test.go` (updated) | 2 | Connection creation includes `app_id` |

**Total handler tests: 29** (was 18, +11 new)

### Files Created

| # | File | Purpose |
|---|------|---------|
| 1 | `backend/internal/api/handlers/organizations.go` | Org + onboarding handlers |
| 2 | `backend/internal/api/handlers/organizations_test.go` | 4 org handler tests |
| 3 | `backend/internal/api/handlers/applications.go` | 11 per-app endpoint handlers |
| 4 | `backend/internal/api/handlers/applications_test.go` | 11 app handler tests |
| 5 | `frontend/src/pages/OnboardingPage.vue` | Onboarding form |
| 6 | `frontend/src/stores/app.ts` | App/org Pinia store |
| 7 | `frontend/src/api/organizations.ts` | Org API client |
| 8 | `frontend/src/api/applications.ts` | App API client |
| 9 | `frontend/src/types/organization.ts` | Org/app/config types |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/router.go` | New route groups: `/api/org`, `/api/onboard`, `/api/apps/{appId}/*` |
| 2 | `backend/internal/api/handlers/connections.go` | `CreateConnection` requires `app_id` |
| 3 | `backend/internal/api/handlers/stats.go` | Per-app stats query |
| 4 | `backend/internal/db/queries/connections.sql` | `ListConnectionsByApp` query |
| 5 | `backend/internal/db/queries/stats.sql` | `GetAppDashboardStats` query |
| 6 | `backend/internal/db/queries/users.sql` | `GetFirstUserInOrg` query |
| 7 | `frontend/src/App.vue` | Calls `app.init()` post-auth |
| 8 | `frontend/src/router/index.ts` | Onboarding route + guard rewrite |
| 9 | `frontend/src/layouts/DefaultLayout.vue` | Skip sidebar for onboarding |
| 10 | `frontend/src/components/common/AppSidebar.vue` | App selector + org footer |
| 11 | `frontend/src/pages/DashboardPage.vue` | 4-column grid, monitoring card |
| 12 | `frontend/src/pages/AgentConfigPage.vue` | Per-app config editing |
| 13 | `frontend/src/pages/ConnectionsPage.vue` | App-scoped connections |
| 14 | `frontend/src/components/connections/ConnectionForm.vue` | Removed `app_id` from form (injected by page) |
| 15 | `frontend/src/components/log/LogEntry.vue` | Monitor + Heartbeat badges |
| 16 | `frontend/src/types/connection.ts` | `app_id` on Connection and CreateConnectionPayload |
| 17 | `backend/internal/api/handlers/testhelpers_test.go` | Full org→app→config test setup |
| 18 | `backend/internal/api/handlers/connections_test.go` | All tests use `app_id` |
| 19 | `backend/internal/api/handlers/webhooks_test.go` | Connection creation with `app_id` |
| 20 | `backend/internal/api/handlers/logs_test.go` | Connection creation with `app_id` |

---

## 0.11.0 — Monitoring Mode (2026-03-07)

Implemented the autonomous monitoring pipeline — the core Phase 8 deliverable. Heimdall now watches systems 24/7 without user interaction: a background goroutine polls for new logs, classifies them through a deterministic Lumber ONNX pipeline, and escalates only flagged entries to Claude for assessment.

### Why

This is the central vision feature. Before 0.11.0, Heimdall only responded to user-initiated chat. Now it runs continuously, processing logs in the background, filtering noise through deterministic classification, and surfacing only what matters — autonomously.

### Classification Pipeline — Lumber Integration

Deterministic log classification using the Lumber ONNX model (in-process, no microservice). The pipeline runs before any LLM call, filtering safe logs so Claude only sees anomalies.

```
LogBuffer → ExtractText → Lumber.ClassifyBatch → ShouldEscalate → flagged | safe
```

**Text extraction** (`extract.go`) — converts arbitrary JSON log payloads to classifiable text:
1. Priority field scan: `message` → `msg` → `error` → `text` → `log` → `body`
2. Prepends `level`/`severity` field if present (e.g., `"ERROR: connection refused"`)
3. Falls back to raw JSON string if no known field found

**Classifier interface** (`classifier.go`) — two implementations:
- `LumberClassifier` — wraps Lumber ONNX with 0.5 confidence threshold. Returns 42-category taxonomy across 8 types (ERROR, REQUEST, DEPLOY, SYSTEM, ACCESS, DATA, SCHEDULED, PERFORMANCE) + UNCLASSIFIED fallback.
- `PassthroughClassifier` — escalates all logs when Lumber is unavailable.

**Three-way startup mode** via `CLASSIFIER_MODE` env var:
- `on` — require Lumber, fatal if model unavailable
- `off` — always use PassthroughClassifier
- `fallback` (default) — try Lumber, fall back to Passthrough if model loading fails

**Severity gate** (`severity_gate.go`) — pure function `ShouldEscalate(event)` with hardcoded rules per taxonomy type:

| Type | Escalate | Safe |
|------|----------|------|
| ERROR | All | — |
| PERFORMANCE | All | — |
| REQUEST | server_error, slow_request | success, redirect, client_error |
| DEPLOY | All | — |
| SYSTEM | resource_alert, config_change | health_check, process_lifecycle, scaling_event |
| ACCESS | login_failure, auth_failure, permission_change, api_key_event | login_success, session_expired |
| DATA | migration | query_executed, replication |
| SCHEDULED | cron_failed | cron_started, cron_completed |
| UNCLASSIFIED | Always | — |
| Unknown | Always (fail-safe) | — |

**Dockerfile changes** — Alpine → Debian bookworm-slim (ONNX requires glibc). 3-stage build: compile, download models from HuggingFace/GitHub, runtime. Models at `/opt/lumber/models`, library at `/usr/local/lib/libonnxruntime.so`.

### Monitor Loop

**Lifecycle** — `Agent.Start(ctx)` launches a background goroutine running `Monitor(ctx)`. `Agent.Stop()` cancels the context and waits via `sync.WaitGroup`. Wired into `main.go` startup/shutdown sequence: start agent after creation, stop before closing classifier and pool.

**Tick cycle** (every 15 seconds):
1. `ListActiveApplications` — 3-way JOIN returning apps with `status='active'`, `mode!='off'`, and at least one active connection
2. For each app (up to 10 concurrent via semaphore, 2-minute timeout each):
   - Resolve org user for log attribution (`GetFirstUserInOrg`)
   - Get cursor position (`GetMonitoringState`); first run initializes to `now()` and skips
   - Fetch up to 200 logs since cursor (`ListLogsSinceForApp`, ASC order)
   - Classify through Lumber pipeline
   - Emit heartbeat always (`logs_processed`, `safe`, `flagged` counts)
   - If flagged > 0: cap at 50 logs, truncate payloads to 2000 chars, format for LLM, call `RunMonitoring`
   - Emit monitoring entry with assessment text and severity
   - Advance cursor to last log's `ingested_at`

**Constants:**

| Constant | Value | Purpose |
|----------|-------|---------|
| `monitorTickInterval` | 15s | Global polling rate |
| `maxConcurrentApps` | 10 | Semaphore bound for concurrent app processing |
| `logBatchLimit` | 200 | Max logs fetched per app per cycle |
| `maxFlaggedForLLM` | 50 | Max flagged logs sent to Claude per cycle |
| `maxPayloadChars` | 2000 | Per-log payload truncation limit |
| `monitorAppTimeout` | 2m | Per-app processing deadline |

### RunMonitoring — Agent Method

New `RunMonitoring(ctx, userID, appConfig, flaggedLogs) → (assessment, severity)`:
- Uses monitoring-specific system prompt (Variant B — token-optimized, requests explicit `Severity: <level>` format)
- Loads per-app model from `app_agent_config` (falls back to `claude-sonnet-4-6`)
- Sessionless — no conversation history, no persistence
- Full tool-use loop (search_logs, query_database) capped at 10 iterations
- Error resilient: returns degraded assessment with `"error"` severity on API failure (never crashes the monitor)

**Severity parsing** from agent response:
1. Explicit marker: `Severity: critical|error|warning|info`
2. Heuristic keyword scan (priority order: critical > error > warning)
3. Default: `info`

### Agent Log Entries

Two new entry types emitted by the monitor:
- **`heartbeat`** — every cycle, every app. Contains `logs_processed`, `safe`, `flagged` counts. Lightweight status indicator.
- **`monitoring`** — only when flagged logs exist. Contains full assessment text, severity, flagged count. Triggers frontend amber badge.

### EmitLog Refactor

Refactored emit layer to support severity:
- `EmitLog(ctx, userID, conversationID, entryType, summary, detail)` — existing, severity defaults to empty
- `EmitLogWithSeverity(...)` — new, accepts explicit severity string
- Private `emitLog` delegate handles both. Fire-and-forget: errors logged but never propagated.

### Test Coverage

| File | Tests | Coverage |
|------|-------|---------|
| `extract_test.go` | 12 | Message field priority, level prepending, Supabase-style payloads, non-JSON fallback, batch extraction |
| `severity_gate_test.go` | 32 | All 8 taxonomy types × escalate/safe sub-cases, UNCLASSIFIED, unknown types |
| `classifier_test.go` | 7 | PassthroughClassifier (3 pure), LumberClassifier (4 integration, gated by `LUMBER_MODEL_DIR`) |
| `monitor_test.go` | 16 | Format (single/multi/metadata), severity parsing (explicit/heuristic/fallback/priority), scheduling (continuous/periodic/no-state), RunMonitoring (simple/tool-use/max-iterations), lifecycle (start/stop/context), error handling (DB failures, no panic) |

**Total agent tests: 63** (was 6 before Phase 8)

### Files Created

| # | File | Purpose |
|---|------|---------|
| 1 | `backend/internal/agent/classifier.go` | `Classifier` interface, `ClassifiedLog` struct, `PassthroughClassifier` |
| 2 | `backend/internal/agent/classifier_lumber.go` | `LumberClassifier` wrapping ONNX model |
| 3 | `backend/internal/agent/extract.go` | `ExtractText` — JSON payload → classifiable string |
| 4 | `backend/internal/agent/severity_gate.go` | `ShouldEscalate` — deterministic escalation rules |
| 5 | `backend/internal/agent/extract_test.go` | 12 extraction tests |
| 6 | `backend/internal/agent/severity_gate_test.go` | 32 severity gate tests |
| 7 | `backend/internal/agent/classifier_test.go` | 7 classifier tests |
| 8 | `backend/internal/agent/monitor_test.go` | 16 monitor loop tests |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/agent.go` | Added `classifier` field, `cancel`/`wg` for lifecycle, `Start()`/`Stop()` methods |
| 2 | `backend/internal/agent/monitor.go` | Full rewrite: tick loop, classification pipeline, monitorApp flow |
| 3 | `backend/internal/agent/loop.go` | Added `RunMonitoring` method, `parseSeverityFromResponse` |
| 4 | `backend/internal/agent/emit.go` | Refactored: `EmitLogWithSeverity`, private delegate |
| 5 | `backend/internal/agent/prompt.go` | Added `monitoringSystemPrompt` (Variant B), `BuildMonitoringPrompt()` |
| 6 | `backend/internal/config/config.go` | Added `ClassifierMode`, `LumberModelDir` config fields |
| 7 | `backend/cmd/heimdall/main.go` | Classifier init (3-way mode switch), `ag.Start()`, shutdown sequence |
| 8 | `backend/Dockerfile` | Alpine → Debian, 3-stage build, ONNX model download |
| 9 | `backend/go.mod` / `backend/go.sum` | Lumber dependency |

---

## 0.10.0 — Multi-App Data Model (2026-03-07)

Introduced the foundational data model for multi-application monitoring. Replaces the flat user-scoped model with an organizational hierarchy: **User → Organization → Application → Connection**. Adds per-application agent configuration and monitoring state tracking.

### Why

Heimdall previously assumed a single user with a flat set of connections. To support monitoring mode (Phase 8), the system needs to know *which application* each connection belongs to, configure the agent independently per app, and track monitoring progress per app. This release lays all the schema and query groundwork for that.

### Migration 014 — Organizations & Applications

- **`organizations`** table — `id`, `name`, `slug` (unique), timestamps. Represents a team or company.
- **`users.org_id`** — nullable FK to `organizations`. Null means the user hasn't completed onboarding.
- **`applications`** table — `id`, `org_id` FK (CASCADE), `name`, `status` (default `'active'`), timestamps. Each app is a distinct monitored system.
- **`connections.app_id`** — `NOT NULL` FK to `applications` (CASCADE). Connections now belong to apps, not directly to users.
- Existing `connections` and `log_buffer` rows wiped (test data) to allow the `NOT NULL` constraint.

### Migration 015 — Per-Application Agent Config

- **`app_agent_config`** table — `app_id` PK (1:1 with applications), `model` (default `claude-sonnet-4-6`), `mode` (continuous/periodic/off), `schedule_interval_secs` (default 60), `system_prompt_override` (nullable), timestamps.
- RLS policy `app_agent_config_org` — users can only access configs for apps within their org, enforced via `app_current_user_id()` subquery.
- Uses `schedule_interval_secs` (integer) rather than cron — simpler to validate and sufficient for interval-based monitoring.

### Migration 016 — Monitoring State

- **`monitoring_state`** table — `app_id` PK (1:1 with applications), `last_monitored_at` (cursor position), `updated_at`.
- Skip-on-resume semantics: when monitoring is re-enabled after being off, cursor resets to `now()` — no backfill of missed logs.

### sqlc Queries

Five new query files covering the full data access layer for the new model:

| File | Queries |
|------|---------|
| `queries/organizations.sql` | `CreateOrganization`, `GetOrganization`, `GetOrganizationBySlug`, `GetOrganizationByUser`, `UpdateOrganization` |
| `queries/applications.sql` | `CreateApplication`, `GetApplication`, `ListApplicationsByOrg`, `UpdateApplication`, `DeleteApplication` |
| `queries/app_agent_config.sql` | `GetAppAgentConfig`, `UpsertAppAgentConfig` |
| `queries/monitoring.sql` | `GetMonitoringState`, `UpsertMonitoringState`, `ResetMonitoringCursor`, `ListActiveApplications`, `ListLogsSinceForApp` |
| `queries/users.sql` | Updated `GetUser` to include `org_id`; added `SetUserOrg` |

Key query design:
- **`ListActiveApplications`** — three-way join (applications + app_agent_config + connections EXISTS) returning only apps with active status, monitoring enabled, and at least one active connection.
- **`ListLogsSinceForApp`** — fetches logs for an app's connections since the cursor timestamp, ordered ASC with configurable LIMIT for batched processing.

### Files Created

| # | File | Purpose |
|---|------|---------|
| 1 | `backend/migrations/014_organizations_applications.up.sql` | Orgs, apps, and connection re-parenting |
| 2 | `backend/migrations/014_organizations_applications.down.sql` | Reverse migration |
| 3 | `backend/migrations/015_app_agent_config.up.sql` | Per-app agent config table + RLS |
| 4 | `backend/migrations/015_app_agent_config.down.sql` | Reverse migration |
| 5 | `backend/migrations/016_monitoring_state.up.sql` | Monitoring cursor table |
| 6 | `backend/migrations/016_monitoring_state.down.sql` | Reverse migration |
| 7 | `backend/internal/db/queries/organizations.sql` | Org CRUD queries |
| 8 | `backend/internal/db/queries/applications.sql` | App CRUD queries |
| 9 | `backend/internal/db/queries/app_agent_config.sql` | Agent config queries |
| 10 | `backend/internal/db/queries/monitoring.sql` | Monitoring state + log fetch queries |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/db/queries/users.sql` | `GetUser` returns `org_id`; added `SetUserOrg` |
| 2 | `backend/internal/db/models.go` | Regenerated — new `Organization`, `Application`, `AppAgentConfig`, `MonitoringState` models |
| 3 | `backend/internal/db/users.sql.go` | Regenerated — `GetUser` includes `OrgID`, new `SetUserOrg` |
| 4 | `backend/internal/db/connections.sql.go` | Regenerated — `Connection` model includes `AppID` |
| 5 | `backend/internal/db/organizations.sql.go` | New generated file — 5 methods |
| 6 | `backend/internal/db/applications.sql.go` | New generated file — 5 methods |
| 7 | `backend/internal/db/app_agent_config.sql.go` | New generated file — 2 methods |
| 8 | `backend/internal/db/monitoring.sql.go` | New generated file — 5 methods |

---

## 0.9.1 — UI Polish & Test Coverage (2026-03-06)

Two-track release: UI polish across the frontend and test coverage for both backend and frontend.

### Why

The app was functional but rough around the edges — hardcoded dashboard values, no edit mode on agent config, generic loading spinners, no error feedback, and zero test coverage. This release addresses all of that.

### 7a — UI Polish

- **Toast notifications** — Global notification system via module-level singleton. Any code (including non-component modules like API interceptors) can trigger toasts. Renders via `<Teleport>` with enter/leave transitions.
- **Agent config editing** — Config page now supports editing model, mode, schedule, and system prompt. Model field uses `<input>` + `<datalist>` so new model IDs work without code changes.
- **Dashboard enhancement** — New `GET /api/stats` endpoint returns `log_count_24h`, `connection_count`, and `active_connections`. Dashboard now shows real data with a 3-column grid and hourly ingestion rate.
- **Loading skeletons** — Replaced `LoadingSpinner` with layout-mimicking skeleton loaders on all pages (connections, agent log, config, reports).
- **Error boundaries** — Global error handler in `main.ts` + axios response interceptor: 401 → auto-logout + redirect, 5xx → error toast, network failure → "Connection lost" toast. Dynamic `import()` avoids circular dependencies.
- **Responsive audit** — Checked all pages at 375px and 768px. Fixed header stacking on agent config/chat pages and made connection card action buttons always-visible on touch devices.

### 7b — Test Coverage

- **Go test infrastructure** — Tests run against real Supabase DB with per-test user creation and `t.Cleanup` cascade delete. Auth bypass via exported `ContextWithUserID()`.
- **Handler tests** — Black-box tests (`handlers_test` package) covering connections CRUD, agent config get/update, logs listing, webhook ingestion (valid + invalid token), auth me endpoint, and dashboard stats.
- **Agent loop tests** — Uses `httptest.Server` with `option.WithBaseURL` to intercept Anthropic API calls. Tests simple response, max iterations, and tool error flows. `stubDBTX` makes DB operations fail gracefully.
- **Frontend test infrastructure** — Vitest + happy-dom + Vue Test Utils. Fresh Pinia instance and global axios mock per test via setup file.
- **Store tests** — 17 tests across all four Pinia stores: connections (CRUD + testingId lifecycle), logs (fetch + pagination + source filter), agent (config fetch/update), auth (init/login/logout/isAuthenticated).

### Files Created

| # | File | Purpose |
|---|------|---------|
| 1 | `frontend/src/composables/useToast.ts` | Toast singleton: `show()` / `dismiss()` API |
| 2 | `frontend/src/components/common/ToastContainer.vue` | Fixed bottom-right toast renderer |
| 3 | `frontend/src/components/common/SkeletonBlock.vue` | Configurable skeleton loader with pulse animation |
| 4 | `frontend/src/api/stats.ts` | `getDashboardStats()` API function |
| 5 | `backend/internal/db/queries/stats.sql` | `GetDashboardStats` query |
| 6 | `backend/internal/db/stats.sql.go` | sqlc generated code |
| 7 | `backend/internal/api/handlers/stats.go` | `GET /api/stats` handler |
| 8 | `frontend/src/test/setup.ts` | Vitest setup: Pinia + axios mock |
| 9 | `frontend/src/stores/__tests__/connections.test.ts` | Connection store tests |
| 10 | `frontend/src/stores/__tests__/logs.test.ts` | Logs store tests |
| 11 | `frontend/src/stores/__tests__/agent.test.ts` | Agent store tests |
| 12 | `frontend/src/stores/__tests__/auth.test.ts` | Auth store tests |
| 13 | `backend/internal/api/handlers/testhelpers_test.go` | Go test setup + HTTP helper |
| 14 | `backend/internal/api/handlers/connections_test.go` | Connection handler tests |
| 15 | `backend/internal/api/handlers/agent_test.go` | Agent config handler tests |
| 16 | `backend/internal/api/handlers/logs_test.go` | Logs + stats handler tests |
| 17 | `backend/internal/api/handlers/webhooks_test.go` | Webhook handler tests |
| 18 | `backend/internal/api/handlers/auth_test.go` | Auth handler tests |
| 19 | `backend/internal/agent/loop_test.go` | Agent loop tests |
| 20 | `backend/internal/agent/tools_test.go` | Tool dispatch tests |

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/App.vue` | Mounted `<ToastContainer />` at app root |
| 2 | `frontend/src/main.ts` | Global error handler + unhandled rejection listener |
| 3 | `frontend/src/api/client.ts` | Axios response interceptor (401/5xx/network) |
| 4 | `frontend/src/stores/agent.ts` | Added `updateConfig()` action |
| 5 | `frontend/src/types/agent.ts` | Added `'off'` to mode union |
| 6 | `frontend/src/pages/DashboardPage.vue` | 3-column grid, real stats, error banner |
| 7 | `frontend/src/pages/AgentConfigPage.vue` | Edit mode, skeleton loader, responsive header |
| 8 | `frontend/src/pages/AgentChatPage.vue` | Responsive header |
| 9 | `frontend/src/pages/AgentLogPage.vue` | Skeleton loader |
| 10 | `frontend/src/pages/ConnectionsPage.vue` | Skeleton loader |
| 11 | `frontend/src/pages/ReportsPage.vue` | Skeleton loader |
| 12 | `frontend/src/components/connections/ConnectionCard.vue` | Touch-friendly action buttons |
| 13 | `backend/internal/api/router.go` | Added `/stats` route |
| 14 | `backend/internal/api/middleware/auth.go` | Exported `ContextWithUserID()` |
| 15 | `frontend/vite.config.ts` | Vitest test config |
| 16 | `frontend/tsconfig.app.json` | Added `vitest/globals` types |
| 17 | `frontend/package.json` | Test deps + scripts |
| 18 | `backend/go.mod` | Added `testify` |

---

## 0.9.0 — Public Website (2026-02-27)

Added the public marketing website as Vue routes inside the existing frontend app. Three pages — landing (`/`), features (`/features`), pricing (`/pricing`) — served without authentication alongside the existing dashboard. Dashboard moved from `/` to `/dashboard`.

### Why

Heimdall needed a public-facing presence to explain the product, show features, and present pricing — without spinning up a separate site. Keeping everything in a single Vue app means one Vercel deployment, shared design tokens, and no second build pipeline.

### Pages

- **`/` — Landing**: Full-screen hero with three-line headline, accent-coloured middle line, and two CTAs ("Start Monitoring" + "See How It Works"). No scroll — single viewport with inline footer pinned to bottom.
- **`/features` — Features**: Four sections — the problem (3-column grid), how it works (3-step sequence), core features (2×3 card grid with hover effects), and a CTA banner.
- **`/pricing` — Pricing**: Three-column pricing cards (Starter $0 / Pro $49 / Enterprise Custom). Pro tier highlighted with accent border and "Popular" badge. Feature checklists with check icons.

### Routing & Auth

Public routes use `meta: { public: true }` — the auth guard skips these entirely. The `DefaultLayout` was extended to bypass the sidebar/app-shell for public pages (same pattern as the login page). Post-login redirect updated from `/` to `/dashboard`.

### Design

Reuses the existing techno-brutalist design tokens from `main.css` — no new CSS variables or theme work needed. The public nav features the Bifrost Hexagram logo as inline SVG with desktop links, mobile hamburger menu, and active route highlighting.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/router/index.ts` | Added 3 public routes, moved dashboard to `/dashboard`, auth guard respects `meta.public` |
| 2 | `frontend/src/layouts/DefaultLayout.vue` | Layout bypass extended to all `route.meta?.public` pages |
| 3 | `frontend/src/pages/LoginPage.vue` | Post-login redirect → `/dashboard` |
| 4 | `frontend/src/components/common/AppSidebar.vue` | Dashboard link → `/dashboard` |

### Files Created

| # | File | Purpose |
|---|------|---------|
| 1 | `frontend/src/components/public/PublicNav.vue` | Marketing nav with logo, links, mobile menu |
| 2 | `frontend/src/components/public/PublicFooter.vue` | Footer with `inline` and `full` variants |
| 3 | `frontend/src/pages/public/LandingPage.vue` | Single-viewport landing page |
| 4 | `frontend/src/pages/public/FeaturesPage.vue` | Features page with 4 sections |
| 5 | `frontend/src/pages/public/PricingPage.vue` | Pricing page with 3 tiers |

### Open Items

1. GitHub link in nav needs real repo URL
2. Pricing tiers/prices/features are placeholders
3. Terms & Privacy pages not yet created

---

## 0.8.8 — Row Level Security (2026-02-26)

Added Row Level Security (RLS) policies to all user-scoped database tables. Every table that holds user data now has a `user_id` column and an RLS policy enforcing row-level isolation. The backend also injects the authenticated user's identity into each Postgres transaction via `SET LOCAL`, laying the groundwork for full RLS enforcement if the connection role ever changes from the current superuser.

### Why

Previously, data isolation was enforced entirely at the application layer — every query manually filtered by `user_id` via WHERE clauses. This works but is fragile: a single missed filter, a new query, or a raw SQL session could leak data across users. RLS provides defence-in-depth at the database level, ensuring Postgres itself enforces row ownership regardless of how queries are constructed.

### Phase 1 — Schema Gaps (migrations 011–012)

Two tables were missing `user_id` columns required for RLS:

- **`investigations`** — had no user scoping at all. Added `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE` with index.
- **`log_buffer`** — was scoped indirectly via `JOIN connections`. Added `user_id` column, backfilled from `connections.user_id`, then set `NOT NULL` with index.

All sqlc queries for both tables were updated to include `user_id` in their filters. The `log_buffer` queries were simplified from JOIN-based scoping to direct `WHERE user_id = $1` filtering. The webhook ingestion handler now writes `user_id` (from the connection record) when inserting log entries.

### Phase 2 — Per-Request User Context

Added a `UserQueries` helper that injects the authenticated user's identity into Postgres before executing queries:

1. Begins a transaction on the connection pool
2. Runs `SET LOCAL app.current_user_id = '<uuid>'` (scoped to the transaction)
3. Returns a `*db.Queries` wrapping that transaction + a cleanup function

All user-scoped HTTP handlers now call `UserQueries(ctx, userID)` instead of using the shared `Queries` instance directly. The WebSocket chat handler uses per-operation `UserQueries` calls (conversation load/create, message persist, title update) rather than a single long-lived transaction.

Non-user-scoped paths (agent config, webhook ingestion) continue using the shared `Queries` — they don't need user identity injection.

### Phase 3 — RLS Policies (migration 013)

A single migration that enables RLS on all six user-scoped tables:

| Table | Policy | Rule |
|-------|--------|------|
| `users` | `users_self` | `id = app_current_user_id()` |
| `connections` | `connections_owner` | `user_id = app_current_user_id()` |
| `conversations` | `conversations_owner` | `user_id = app_current_user_id()` |
| `agent_log` | `agent_log_owner` | `user_id = app_current_user_id()` |
| `log_buffer` | `log_buffer_owner` | `user_id = app_current_user_id()` |
| `investigations` | `investigations_owner` | `user_id = app_current_user_id()` |

A helper function `app_current_user_id()` safely reads `current_setting('app.current_user_id', true)` — returns `NULL` if unset, never errors.

`agent_config` is excluded — it's a system-wide single-row table with no user scoping.

### RLS Enforcement Model

The backend connects as the `postgres` superuser (table owner), which **bypasses RLS by default**. This is intentional — the backend retains full, unrestricted access. The policies protect against non-owner access paths: Supabase dashboard roles (`anon`, `authenticated`), PostgREST, and direct `psql` sessions with other roles. The `set_config` plumbing is in place so that if the backend ever migrates to a dedicated non-owner app role, RLS enforcement activates automatically.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/migrations/011_add_user_id_to_investigations.up.sql` | Add `user_id` column + index |
| 2 | `backend/migrations/011_add_user_id_to_investigations.down.sql` | Drop column + index |
| 3 | `backend/migrations/012_add_user_id_to_log_buffer.up.sql` | Add `user_id` column, backfill, set NOT NULL + index |
| 4 | `backend/migrations/012_add_user_id_to_log_buffer.down.sql` | Drop column + index |
| 5 | `backend/migrations/013_enable_rls.up.sql` | Helper function + RLS policies on 6 tables |
| 6 | `backend/migrations/013_enable_rls.down.sql` | Drop policies, disable RLS, drop function |
| 7 | `backend/internal/db/queries/investigations.sql` | All queries now filter by `user_id` |
| 8 | `backend/internal/db/queries/log_buffer.sql` | Replaced JOIN scoping with direct `user_id` filter, added `user_id` to INSERT |
| 9 | `backend/internal/db/models.go` | Regenerated — `Investigation` and `LogBuffer` structs include `UserID` |
| 10 | `backend/internal/db/investigations.sql.go` | Regenerated |
| 11 | `backend/internal/db/log_buffer.sql.go` | Regenerated |
| 12 | `backend/internal/api/handlers/server.go` | Added `Pool` field to Server struct |
| 13 | `backend/internal/api/handlers/userqueries.go` | New — `UserQueries()` helper |
| 14 | `backend/internal/api/handlers/connections.go` | All 6 handlers use `UserQueries` |
| 15 | `backend/internal/api/handlers/conversations.go` | Both handlers use `UserQueries` |
| 16 | `backend/internal/api/handlers/chat.go` | Per-operation `UserQueries` for WebSocket flow |
| 17 | `backend/internal/api/handlers/logs.go` | `ListLogs` uses `UserQueries` |
| 18 | `backend/internal/api/handlers/auth.go` | `Me` uses `UserQueries` |
| 19 | `backend/internal/api/handlers/webhooks.go` | Passes `conn.UserID` to `InsertLogEntry` |

---

## 0.8.7 — Connection Edit & Ping (2026-02-26)

Added inline editing and manual connectivity re-testing ("ping") to connection cards. Previously the only way to fix a misconfigured connection was to delete and recreate it, and there was no way to re-verify connectivity after infrastructure changes.

### Why

Users who entered wrong credentials or whose infrastructure changed (password rotation, firewall rules) had to delete and recreate connections from scratch. The connectivity test only ran once at creation time with no way to re-trigger it.

### Edit

- **Edit** button on each connection card (hover-reveal, alongside Delete).
- Opens the existing `ConnectionForm` pre-populated with the connection's current values.
- Type selector is locked during edit — changing type would invalidate config fields.
- Submit button reads **"Save Changes"** instead of "Add Connection".
- On save, calls `PUT /connections/:id` then automatically re-runs the connectivity test (same flow as create).

### Ping

- **Ping** button on each connection card — triggers `POST /connections/:id/test` on demand.
- Shows the pulsing "TESTING" badge while running, then updates to ACTIVE or ERROR.
- Error banner appears if the test fails, same as post-create behaviour.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/components/connections/ConnectionCard.vue` | Edit + Ping buttons, new emits |
| 2 | `frontend/src/components/connections/ConnectionList.vue` | Forward `edit` and `test` events |
| 3 | `frontend/src/components/connections/ConnectionForm.vue` | `initialValues` prop, edit mode, locked type, dynamic button label |
| 4 | `frontend/src/pages/ConnectionsPage.vue` | `editingConnection` ref, unified submit handler, ping handler |

No backend changes — the existing `PUT` and `POST .../test` endpoints already covered both flows.

---

## 0.8.6 — Connection Test on Create (2026-02-26)

Added automatic connectivity testing after creating a connection. The system now verifies credentials and reachability immediately, updating the connection status to `active` or `error` with a clear message — so users know right away whether their connection works.

### Why

Previously every new connection sat at `inactive` with no feedback. Users couldn't tell if credentials were wrong or a host was unreachable until the agent tried to use the connection later, at which point the error was buried in agent logs.

### New Endpoint

`POST /api/connections/{id}/test` — authenticated, user-scoped. Returns `{ "success": true/false, "message": "..." }` and updates the connection's status in the database.

| Type | Test behaviour |
|------|---------------|
| `postgres` | Builds connector, calls `Connect()` with 5s timeout, then `Close()` |
| `webhook_logs` / `syslog` / `github` | Auto-pass (no remote target to test yet) |

### Frontend UX

1. User creates a connection → card appears with a pulsing green **"TESTING"** badge
2. On success → badge transitions to **"ACTIVE"**
3. On failure → badge transitions to **"ERROR"** + an error banner shows the reason (e.g. *"Connection created but test failed: password authentication failed"*)

The user is never left guessing about connection state.

### Files Changed

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/connections.go` | `TestConnection` handler — fetch, test by type, update status |
| 2 | `backend/internal/api/router.go` | Register `POST /{id}/test` |
| 3 | `frontend/src/api/connections.ts` | `testConnection(id)` API call |
| 4 | `frontend/src/stores/connections.ts` | `testingId` ref + `testConnection` action |
| 5 | `frontend/src/pages/ConnectionsPage.vue` | Call test after create, show error on failure |
| 6 | `frontend/src/components/connections/ConnectionCard.vue` | `testing` prop → pulsing "TESTING" badge |
| 7 | `frontend/src/components/connections/ConnectionList.vue` | Pass `testingId` through to cards |

---

## 0.8.5 — Connection Config Fields (2026-02-26)

Added dynamic configuration fields to the connection form so users can provide actual credentials and connection details — database host, port, password, API tokens, etc. Previously the form only captured Name, Type, and Direction, sending an empty `config: {}` to the backend.

### Why

The backend's `config` JSONB column and the agent's `query_database` tool already supported full connection credentials, but there was no way to enter them through the UI. Without config data the agent couldn't connect to any user databases.

### Config Fields by Type

| Type | Fields |
|------|--------|
| **PostgreSQL** | Host, Port (default 5432), Database, Username, Password, SSL Mode (disable/require/verify-full) |
| **Webhook Logs** | None — info note explains the webhook token is auto-generated |
| **Syslog** | Host, Port (default 514), Protocol (UDP/TCP) |
| **GitHub** | Owner, Repository, Personal Access Token |

Fields render dynamically when the user switches connection type. Password and token fields use `type="password"` inputs. Default values (ports, SSL mode, protocol) are applied if the user doesn't override them.

### Bug Fix

Fixed a field name mismatch: the frontend sent `username` but the backend postgres connector expects `user` (`json:"user"` on `postgresConfig`). The form now sends `user` to match.

### Files Changed

1 file: `frontend/src/components/connections/ConnectionForm.vue`

---

## 0.8.4 — Disable Scale-to-Zero (2026-02-26)

Set `min_machines_running` from `0` to `1` in the Fly.io configuration to eliminate cold starts. Heimdall's backend now keeps at least one machine running at all times, so WebSocket connections and agent requests are served immediately without a spin-up delay.

### Why

Fly.io defaults to scale-to-zero when no traffic is flowing. For a monitoring agent that needs to be responsive on-demand (WebSocket chat, log ingestion webhooks), a cold start of several seconds is unacceptable — especially for WebSocket upgrades which can time out during machine boot.

### Files Changed

1 file: `backend/fly.toml` — `min_machines_running: 0 → 1`

---

## 0.8.3 — Auth Guard Race Condition Fix (2026-02-26)

Fixed a bug where unauthenticated users could land on the dashboard without being redirected to `/login`. The router navigation guard skips auth checks while the auth store is initializing — but once initialization completed, the guard never re-evaluated the already-resolved route, leaving unauthenticated users on protected pages.

### Root Cause

The `beforeEach` guard in `router/index.ts` returns early when `!auth.initialized`, allowing the initial navigation to proceed to any route. `App.vue` gates rendering behind `auth.init()`, but after init completes the route is already resolved — the guard doesn't re-fire because no new navigation occurs. The result: dashboard renders with `isAuthenticated === false`.

### Fix

Added `router.replace(router.currentRoute.value.fullPath)` in `App.vue` immediately after `auth.init()` resolves. This re-triggers the navigation guard with `initialized === true`, so the auth check runs and redirects sessionless visitors to `/login`. Using `replace` avoids a duplicate history entry.

### Files Changed

1 file: `frontend/src/App.vue`

---

## 0.8.2 — Missing Migration Fix (2026-02-26)

Applied migration `010_create_agent_log` which had been missing from the remote Supabase database. The migration was created in v0.7.0 (Phase 6) but never applied, causing 500 errors on `GET /api/logs` — the unified log endpoint queries both `log_buffer` and `agent_log`, and the missing table crashed every request.

### Database

- Applied `010_create_agent_log`: creates `agent_log` table for agent-emitted observations (`tool_call`, `tool_result`, `observation`). Indexes on `(user_id, created_at DESC)` and `(entry_type)`. Schema version now at **10**.

### Backend

- **Fixed WebSocket hijack failure** — the `statusWriter` in the logging middleware wrapped `http.ResponseWriter` but didn't implement `http.Hijacker`, preventing WebSocket upgrades. Added `Unwrap()` method so `coder/websocket` can reach the underlying connection. This was causing `"http.ResponseWriter does not implement http.Hijacker"` on every `/ws/chat` connection attempt.

### Root Cause (500 on /api/logs)

The `ListLogs` handler defaults `source` to `"all"`, which always queries `agent_log` via `ListAgentLogByUser`. With the table missing, the query failed and returned 500 before raw logs could be fetched — making the entire Agent Log page non-functional.

### Root Cause (WebSocket failure)

The `statusWriter` struct in `middleware/logging.go` embeds `http.ResponseWriter` to capture status codes, but Go's type promotion only surfaces the interface methods — not `http.Hijacker` from the concrete server type. The `Unwrap()` method lets the WebSocket library traverse the wrapper chain to find the real hijackable writer.

---

## 0.8.1 — Darker Background Tuning (2026-02-26)

Toned down the green tint on main-area backgrounds, pushing them closer to pure black. Sidebar unchanged. Purely cosmetic — 6 design tokens adjusted in `main.css`.

### Token Changes

| Token | Before | After |
|-------|--------|-------|
| `--bg-primary` | `#080c08` | `#060806` |
| `--bg-surface` | `rgba(14,24,14,0.5)` | `rgba(10,14,10,0.5)` |
| `--bg-surface-hover` | `rgba(14,24,14,0.7)` | `rgba(10,14,10,0.7)` |
| `--bg-elevated` | `#0e150e` | `#0a0e0a` |
| `--border` | `rgba(90,158,106,0.08)` | `rgba(90,158,106,0.06)` |
| `--border-hover` | `rgba(90,158,106,0.18)` | `rgba(90,158,106,0.14)` |

### Files Changed

1 file: `frontend/src/assets/styles/main.css`

---

## 0.8.0 — Techno-Brutalist Redesign (2026-02-26)

Full frontend redesign transforming Heimdall from a generic light-gray utility into a dark, military-grade AI monitoring interface. Green-black atmosphere, monospace-forward typography, structural borders, and "alive" interface effects. Zero new backend changes — purely frontend.

Spec: [`docs/executing/redesign-fe.md`](executing/redesign-fe.md)

### Design System (Phase 1)

- **Self-hosted fonts** — JetBrains Mono (display/UI, weights 400/500/700) and Inter (body text, weights 400/500/600) via `@fontsource`. No external CDN calls. ([phase8-foundation])
- **25 CSS design tokens** in `:root` — backgrounds (`#080c08` green-tinted near-black), text (warm whites with green undertone), accent (`#5a9e6a` muted forest sage), borders (green-tinted structural lines), status colors (desaturated camo register: olive-gold warnings, muted reds, steel blues). ([phase8-foundation])
- **Tailwind v4 `@theme` registration** — all tokens mapped to utility classes (`bg-bg-primary`, `text-accent`, `border-border`, `font-mono`). Dual-access: Tailwind utilities in templates, `var(--token)` in raw CSS. ([phase8-foundation])
- **Base styles** — dark background on `<html>` (prevents FOUC), antialiased rendering, green-tinted scrollbars, green selection highlight, accessible green focus rings. ([phase8-foundation])
- **`prefers-reduced-motion`** — blanket disable of all animations/transitions for users who opt out. ([phase8-foundation])

### Shell & Navigation (Phase 2)

- **Sidebar redesign** — branded header (pulsing green dot + "HEIMDALL" wordmark + status label), four navigation sections (OVERVIEW, INFRASTRUCTURE, AGENT, INTELLIGENCE) with uppercase monospace section labels. ([phase8-shell])
- **Active route highlighting** — reactive via `useRoute()`. Active item gets `bg-accent-subtle` background + green left-border accent bar. ([phase8-shell])
- **Mobile responsive** — sidebar collapses below `lg` breakpoint. Hamburger button triggers a `Teleport`-ed slide-over panel with `backdrop-blur-sm` overlay and CSS enter/leave transitions. ([phase8-shell])
- **Removed `AppHeader.vue`** — Heimdall branding moved into sidebar header. Main content area gains full vertical space. ([phase8-shell])

### Shared Components (Phase 3)

All 14 Vue components restyled to use design tokens exclusively. Zero references to Tailwind's default gray palette remain.

- **StatusBadge** — ghost-fill pill with colored dot indicator. States: active (green), inactive (muted), error (red), warning (yellow). `border-*/30 bg-*/10 text-*` pattern. ([phase8-components])
- **LoadingSpinner** — green accent spinner (`border-border` track, `border-t-accent` leading edge). ([phase8-components])
- **ConnectionCard** — dark surface card with hover border transition. Delete button hidden by default, fades in on hover (`group-hover:opacity-100`). ([phase8-components])
- **ConnectionForm** — dark inputs with green focus rings, accent primary button, outline cancel button. ([phase8-components])
- **ConnectionList** — 2-column responsive grid on `md+`. ([phase8-components])
- **LogEntry** — agent entries get green left-border accent + `bg-accent-subtle`. Severity badges as ghost-fill pills. ([phase8-components])
- **LogFilters** — dark monospace select dropdowns with `flex-wrap` for mobile. ([phase8-components])
- **LogFeed** — monospace pagination controls, muted entry count. ([phase8-components])
- **ChatMessage** — role labels ("OPERATOR" in green, "HEIMDALL" in muted) above bordered message blocks. User messages: accent-tinted. Agent messages: dark surface. ([phase8-components])
- **ChatWindow** — bordered container with scanning-line thinking indicator (CSS gradient sweep, 1.5s loop). ([phase8-components])
- **ChatInput** — dark elevated input, green "SEND" button, disabled state styling. ([phase8-components])
- **ReportCard** — left border colored by severity (red/yellow/blue). Ghost-fill severity badge. ([phase8-components])
- **ReportDetail** — monospace key-value grid with muted labels. ([phase8-components])

### Pages (Phase 4)

All 8 pages restyled with a consistent header pattern: uppercase monospace title + Inter subtitle + border divider.

- **LoginPage** — full-screen dark background with CSS grid overlay (`opacity-[0.03]`). Centered brand block + bordered login card. "AUTHENTICATE" button. Military-tech copy. ([phase8-pages])
- **DashboardPage** — **major enhancement** from 2-line placeholder to real system overview. Three data cards (System Status, Connections, Recent Activity) wired to existing stores. No new API endpoints. ([phase8-pages])
- **AgentChatPage** — connection status dot with human-readable labels (Connected/Connecting/Disconnected). Ghost-fill error banner. ([phase8-pages])
- **AgentLogPage** — consistent header, error styling. ([phase8-pages])
- **ConnectionsPage** — accent green "New Connection" button in header row. ([phase8-pages])
- **AgentConfigPage** — key-value pairs in bordered card with divider rows. Null values show em-dash / "Default". ([phase8-pages])
- **ReportsPage** — consistent header, muted empty state. ([phase8-pages])
- **NotFoundPage** — "Target not found" copy, outline return button. ([phase8-pages])

### Polish & Animation (Phase 5)

Five "alive" interface effects, all respecting `prefers-reduced-motion`:

- **Pulse dot** — `animate-pulse` circles on sidebar brand, dashboard status, login brand, chat connection status. ([phase8-polish])
- **Scanning line** — CSS gradient sweep on chat thinking indicator. ([phase8-polish])
- **Active glow** — `box-shadow: 0 0 15px rgba(90,158,106,0.06)` on active connection cards, report cards, and dashboard status card. Barely visible, felt rather than seen. ([phase8-polish])
- **Typing reveal** — `clip-path` animation (200ms) on agent chat messages. Paint-only operation, no layout thrashing. ([phase8-polish])
- **Staggered fade-in** — 250ms fade + 4px slide, staggered 30ms per item via CSS custom property `--stagger-index`. Applied to connection cards, log entries, report cards, dashboard cards, and activity rows. 12 items complete in 360ms (under 400ms cap). ([phase8-polish])

### Dependencies Added

| Package | Purpose |
|---------|---------|
| `@fontsource/jetbrains-mono` | Self-hosted JetBrains Mono |
| `@fontsource/inter` | Self-hosted Inter |

### Files Changed

38 files touched across 5 phases. 1 file deleted (`AppHeader.vue`). No backend changes.

[phase8-foundation]: completions/phase8-redesign-foundation.md
[phase8-shell]: completions/phase8-redesign-shell-nav.md
[phase8-components]: completions/phase8-redesign-components.md
[phase8-pages]: completions/phase8-redesign-pages.md
[phase8-polish]: completions/phase8-redesign-polish.md

---

## 0.7.1 — Supabase Auth Wiring (2026-02-26)

Replaced the placeholder login flow with real Supabase email/password authentication. The frontend now uses `@supabase/supabase-js` for sign-in, sign-up, session recovery, and automatic token refresh — no backend changes required.

### Auth Store (full rewrite)

- `init()` recovers session on page load via `supabase.auth.getSession()` and subscribes to `onAuthStateChange` for transparent token refresh. ([phase7])
- `login(email, password)` and `signup(email, password)` call Supabase auth methods directly. ([phase7])
- `token` is a computed from `session.access_token` — auto-updates on refresh, consumed by the existing axios interceptor and WebSocket `?token=` param with zero changes to either. ([phase7])

### Login Page (full rewrite)

- Real email/password form with loading state, error banner, and sign-in / sign-up toggle. ([phase7])
- Signup with email confirmation shows "Check your email" message. ([phase7])

### App & Router

- `App.vue` gates rendering behind `auth.initialized` — prevents flash of login page on reload while session recovery is in progress. ([phase7])
- Router guard skips redirect until auth is initialized; redirects authenticated users away from `/login` → `/`. ([phase7])

### Sidebar

- Displays authenticated user's email at the bottom of the nav. ([phase7])
- "Sign out" calls `supabase.auth.signOut()` and redirects to `/login`. ([phase7])

### New Files

- `frontend/src/lib/supabase.ts` — singleton Supabase client reading `VITE_SUPABASE_URL` and `VITE_SUPABASE_ANON_KEY` from env. ([phase7])
- `.env.example` updated with frontend Supabase env vars. ([phase7])

[phase7]: completions/phase7-supabase-auth-wiring.md

---

## 0.7.0 — Agent Log (2026-02-24)

Phase 6 — unified chronological feed combining raw ingested log entries with agent observations. The agent now emits structured log entries during its tool-use loop, and the log endpoint merges both data sources into a single timeline.

### Database

- Migration `010_create_agent_log`: new `agent_log` table for agent-emitted observations (`tool_call`, `tool_result`, `observation`). Indexes on `(user_id, created_at DESC)` and `(entry_type)`. ([phase6])
- `conversation_id` uses `ON DELETE SET NULL` — agent log entries survive conversation deletion, preserving the audit trail. ([phase6])
- sqlc queries: `InsertAgentLog`, `ListAgentLogByUser`, `ListAgentLogByUserAndType`, `CountAgentLogByUser`. ([phase6])

### Agent

- **`EmitLog` method** — fire-and-forget agent log emission. Prevents observability writes from degrading the agent's primary work. ([phase6])
- **Emit hooks in tool-use loop** — three emit points: `tool_call` on dispatch, `tool_result` on success/failure, `observation` on final response. ([phase6])
- `RunConversation` signature updated to accept `conversationID *uuid.UUID` — agent log entries emitted during chat link to the conversation. ([phase6])

### Backend

- **Unified `GET /api/logs`** — merges `log_buffer` and `agent_log` into a single feed. New `source` query param (`all`, `raw`, `agent`). When `connection_id` filter is set, `source` is forced to `raw`. ([phase6])
- Unified response shape with `source`, `source_type`, `summary`, and `detail` fields normalised across both entry types. ([phase6])

### Frontend

- **Source filter** — new dropdown in LogFilters (`All sources` / `Raw logs` / `Agent activity`). ([phase6])
- **Agent entry styling** — purple border and `AGENT` badge for agent-sourced entries. Human-readable labels: `tool_call` → "Tool Call", `tool_result` → "Tool Result", `observation` → "Observation". ([phase6])
- Updated `LogEntry` type, API client, Pinia store, and page wiring for `source` filtering. ([phase6])

[phase6]: completions/phase6-agent-log.md

---

## 0.6.1 — Post-Implementation Fixes (2026-02-24)

Review of Phases 1–5 identified error-handling gaps and repo hygiene issues. All fixes target resilience and cleanliness — no functional changes.

### Backend

- **`rand.Read` error handling** — `crypto/rand.Read` failure in webhook token generation now returns 500 instead of silently producing a zero-value token. ([phase5-fixes])
- **JSON marshal/unmarshal error handling** — malformed config JSON in `CreateConnection` now returns 400/500 instead of being silently swallowed. ([phase5-fixes])
- **WebSocket write error handling** — all five `wsjson.Write` calls in `HandleChat` now check return values; write failures terminate the handler cleanly instead of continuing to run the agent loop for a dead connection. ([phase5-fixes])

### Frontend

- **Missing `onerror` handler** — `useWebSocket` now handles WebSocket errors, transitioning status to `'closed'` instead of showing a stale "connecting" state. ([phase5-fixes])
- **Type safety bypass removed** — widened `ChatMessage.role` to include `'assistant'` and removed `as any` cast in `useAgent.ts`. ([phase5-fixes])

### Repo Hygiene

- **Removed `backend/heimdall` binary from git** — 19MB compiled binary was tracked since Phase 3, creating large deltas on every build. Added to `.gitignore` and removed from tracking. ([phase5-fixes])
- **Deleted duplicate `frontend/vite.config.js`** — identical to existing `vite.config.ts`. Vite prefers `.ts`, so the `.js` copy was dead weight. ([phase5-fixes])

[phase5-fixes]: completions/phase5-fixes.md

---

## 0.6.0 — Agent Chat (2026-02-24)

Phase 5 — WebSocket-based agent chat with multi-turn conversation context and persistence.

### Database

- Migration `009_add_user_id_to_conversations`: user-scopes the conversations table. ([phase5])
- Rewrote sqlc queries — all reads/writes now scoped by `user_id`. Added `UpdateConversationTitleByUser`. ([phase5])

### Auth

- Extracted `ValidateJWT(tokenStr, jwtSecret)` helper from HTTP middleware — shared by REST routes and WebSocket auth. ([phase5])
- WebSocket authentication via `?token=` query parameter — validates before upgrade, rejects with HTTP 401 if invalid. ([phase5])

### Agent

- Added `Message` domain type in `agent/message.go` — represents stored chat messages. ([phase5])
- Added `RunConversation(ctx, userID, history, input)` — converts stored messages to Claude params for multi-turn context. ([phase5])
- `RunLoop` now delegates to `RunConversation` with nil history — backward compatible. ([phase5])

### Backend

- Rewrote `HandleChat` WebSocket handler: JWT auth, conversation create/load, message loop with agent integration, persistence, status signaling. ([phase5])
- New `GET /api/conversations` — user-scoped list (summaries without messages). ([phase5])
- New `GET /api/conversations/:id` — full conversation with messages. ([phase5])

### Frontend

- `useWebSocket` accepts auth options — appends `token` and `conversation_id` as query params. ([phase5])
- `useAgent` handles four message types: `system`, `status`, `error`, and chat messages. Exposes `conversationId`, `isThinking`, `error`, `loadMessages`. ([phase5])
- New `api/conversations.ts` — `listConversations()` and `getConversation(id)`. ([phase5])
- `AgentChatPage` shows connection status indicator, loads conversation history on mount, displays error banner. ([phase5])
- `ChatWindow` shows animated thinking indicator, auto-scrolls on new messages. ([phase5])
- `ChatInput` disables during thinking and when disconnected. ([phase5])

[phase5]: completions/phase5-agent-chat.md

---

## 0.5.0 — Agent Loop (2026-02-23)

Phase 4 — Claude API tool-use integration, turning Heimdall from a log viewer into an AI monitoring agent.

### Dependencies

- Added `anthropic-sdk-go v1.26.0` — official Go SDK for Claude API. ([phase4])

### Agent

- Refactored `Agent` struct with real fields: `*db.Queries`, `*anthropic.Client`, `*config.Config`. ([phase4])
- Replaced custom `ToolDefinition`/`ToolParam` types with SDK-native `anthropic.ToolUnionParam`. ([phase4])
- Registered two tools: `search_logs` (severity filter, paginated results) and `query_database` (user-scoped, read-only). ([phase4])
- Implemented `RunLoop(ctx, userID, input)` — full Claude API tool-use loop with max 10 iterations. ([phase4])
- Tool errors returned as `isError` tool results — Claude handles failures gracefully. ([phase4])
- System prompt scoped to registered tools only — avoids wasted loop iterations on nonexistent tools. ([phase4])

### Postgres Connector

- Real implementation in `connectors/database/postgres.go`: parses config JSONB, connects with `default_transaction_read_only=on`, executes queries, returns `[]map[string]any`. ([phase4])
- Read-only enforcement at the PostgreSQL session level — prevents mutations regardless of SQL content. ([phase4])

### Backend

- Wired Agent into Server: `main.go` creates Agent → `NewRouter(cfg, pool, ag)` → `NewServer(cfg, pool, ag)`. ([phase4])
- `WriteTimeout` raised to 5 minutes — agent loop makes multiple Claude API calls that exceed the previous 30s limit. ([phase4])
- `GET /api/agent/config` now reads from DB with fallback defaults. ([phase4])
- `PUT /api/agent/config` now persists via `UpsertAgentConfig`. ([phase4])
- New `POST /api/agent/run` endpoint — JWT-protected, synchronous test harness for the agent loop. ([phase4])

[phase4]: completions/phase4-agent-loop.md

---

## 0.4.0 — Webhook Log Connector (2026-02-23)

Phase 3 — webhook-based log ingestion, user-scoped log queries, and frontend log display.

### Database

- Rewrote sqlc queries in `log_buffer.sql` — all reads now join through `connections` to scope by `user_id`. ([phase3])
- Added `LIMIT`/`OFFSET` pagination to all list queries. ([phase3])
- Added `CountLogsByUser` query for paginated response totals. ([phase3])
- Added `GetConnectionByWebhookToken` query for webhook auth. ([phase3])
- `InsertLogEntry` now returns the inserted row (`:one` instead of `:exec`). ([phase3])
- Ran `sqlc generate` — regenerated `log_buffer.sql.go`. ([phase3])

### Backend

- New `POST /api/webhooks/logs` endpoint — accepts log payloads (single or batch), authenticates via per-connection webhook token. ([phase3])
- Restructured `/api` router: public webhook route + `r.Group(...)` for JWT-protected routes — fixes Chi subrouter precedence. ([phase3])
- Migration `008_add_webhook_token_index`: partial functional index on `config->>'webhook_token'` for webhook auth lookups. ([phase3])
- Auto-generates `webhook_token` (32-byte hex) in connection `config` when creating `webhook_logs` type connections. ([phase3])
- Replaced stub `ListLogs` handler with real implementation — supports `severity`, `connection_id`, `limit`, `offset` query params. ([phase3])
- Paginated JSON response: `{ data, total, limit, offset }`. ([phase3])

### Frontend

- Updated API layer for paginated response shape (`PaginatedLogs` interface). ([phase3])
- Expanded logs store with pagination state (`total`, `limit`, `offset`), `error` handling, `nextPage`/`prevPage` actions. ([phase3])
- `LogFilters` now includes connection dropdown for filtering by source. ([phase3])
- `LogEntry` displays payload content (extracts `message` field or shows formatted JSON). ([phase3])
- `LogFeed` shows pagination controls and "X–Y of Z" summary. ([phase3])
- `AgentLogPage` wires connections store for filter dropdown, handles pagination and error display. ([phase3])

[phase3]: completions/phase3-webhook-log-connector.md

---

## 0.3.0 — Connections CRUD (2026-02-23)

Phase 2 — full CRUD for connections, scoped to the authenticated user.

### Database

- Migration `007_add_user_id_to_connections`: added `user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE` with index. ([phase2])
- Rewrote sqlc queries to scope all operations by `user_id` (`ListConnectionsByUser`, `GetConnectionByUser`, `DeleteConnectionByUser`). ([phase2])
- Ran `sqlc generate` — updated `connections.sql.go` and `models.go`. ([phase2])

### Backend

- Replaced 5 stub handlers with real implementations: List, Get, Create, Update, Delete. ([phase2])
- All handlers extract user UUID from JWT context; return 401 if absent. ([phase2])
- Create defaults: `direction` → `one_way`, `config` → `{}`, `status` → `inactive`. Returns 201. ([phase2])
- Delete returns 204 No Content. ([phase2])

### Frontend

- Expanded Pinia store with `createConnection`, `updateConnection`, `deleteConnection` actions and error state. ([phase2])
- `ConnectionForm` now includes direction field, emits typed `CreateConnectionPayload`, supports cancel. ([phase2])
- `ConnectionCard` shows direction, has Delete button. ([phase2])
- `ConnectionList` shows empty-state message, forwards delete events. ([phase2])
- `ConnectionsPage` has "New Connection" toggle, error banner, and full create/delete flow. ([phase2])

[phase2]: completions/phase2-connections-crud.md

---

## 0.2.4 — Auth Me Endpoint & DB Pool (2026-02-22)

Phase 1 (Auth) Task 4 — `/api/auth/me` endpoint and database connection pool.

### Server

- Created `pgxpool.Pool` in `main.go` from `DATABASE_URL` — first real database connection in the server. ([phase1-task4])
- `Server` struct now holds `*db.Queries`; `NewServer` accepts the pool and wraps it with `db.New(pool)`. ([phase1-task4])

### Endpoint

- Added `GET /api/auth/me` — returns the authenticated user's `id`, `email`, and `created_at` from `public.users`. ([phase1-task4])
- Uses `UserIDFromContext` (Task 2) to read the JWT subject and `GetUser` (Task 1) to query the database. ([phase1-task4])

### Dependencies

- Promoted `golang-jwt/jwt/v5`, `google/uuid`, `jackc/pgx/v5` from indirect to direct in `go.mod`. ([phase1-task4])
- Added `pgxpool` transitive deps (`jackc/puddle/v2`, `x/sync`). ([phase1-task4])

[phase1-task4]: completions/phase1-task4-auth-me.md

---

## 0.2.3 — Auth-Protected Routes (2026-02-22)

Phase 1 (Auth) Task 3 — apply JWT middleware to protected routes.

### Router

- Applied `middleware.Auth(cfg.SupabaseJWTSecret)` to the `/api` route group — all `/api/*` requests now require a valid Supabase JWT. ([phase1-task3])
- Removed `POST /api/auth/login` stub — Supabase Auth handles login directly. ([phase1-task3])
- Removed dead `Login` handler from `handlers/auth.go`. ([phase1-task3])
- `/ws/chat` remains unprotected — WebSocket auth deferred to Phase 5. ([phase1-task3])

[phase1-task3]: completions/phase1-task3-auth-routes.md

---

## 0.2.2 — JWT Verification Middleware (2026-02-22)

Phase 1 (Auth) Task 2 — backend JWT verification.

### Auth

- Replaced stub auth middleware with real Supabase JWT validation (`internal/api/middleware/auth.go`). ([phase1-task2])
- Validates HMAC-SHA256 signature, checks expiry, extracts user UUID from `sub` claim. ([phase1-task2])
- Added `Auth(jwtSecret) → middleware` constructor and `UserIDFromContext(ctx)` context helper. ([phase1-task2])

### Config

- Added `SupabaseJWTSecret` field to `Config`, loaded from `SUPABASE_JWT_SECRET` env var, required in `Validate()`. ([phase1-task2])

### Dependencies

- Added `github.com/golang-jwt/jwt/v5`. ([phase1-task2])

[phase1-task2]: completions/phase1-task2-jwt-middleware.md

---

## 0.2.1 — Users Table & Auth Groundwork (2026-02-22)

Phase 1 (Auth) Task 1 — database foundation for user identity.

### Database

- Added migration `006_create_users`: `public.users` table with FK to `auth.users(id) ON DELETE CASCADE`. ([phase1-task1])
- Added `handle_new_user()` trigger function (`SECURITY DEFINER`) + `on_auth_user_created` trigger — auto-syncs Supabase Auth sign-ups into `public.users`. ([phase1-task1])
- Applied migration against live Supabase instance.

### sqlc

- Added `GetUser` query (`internal/db/queries/users.sql`). ([phase1-task1])
- Ran `sqlc generate` — produced `User` struct in `models.go` and `GetUser` method in `users.sql.go`. ([phase1-task1])

[phase1-task1]: completions/phase1-task1-users-migration.md

---

## 0.2.0 — Supabase Database (2026-02-22)

Connected Heimdall to a live Supabase Postgres instance and ran all migrations.

### Database

- Switched from local Postgres to Supabase (direct connection, port 5432).
- Added `cmd/dbping` utility — standalone connection smoke test (`SELECT 1`).
- Ran all 5 migrations against Supabase: `connections`, `agent_config`, `investigations`, `conversations`, `log_buffer`.
- Installed `golang-migrate` CLI (`go install` with `postgres` tag).

### Config

- `DATABASE_URL` is the sole database env var — no separate host/port/user/password fields needed.

---

## 0.1.3 — Infrastructure & DevOps (2026-02-20)

Docker production readiness, database tooling, and developer onboarding.

### Docker

- Added multi-stage frontend Dockerfile: builds with Node, serves with nginx. ([fix-013])
- Added multi-stage backend Dockerfile: builds with Go, runs minimal Alpine binary. ([fix-013])
- Added nginx `try_files` SPA routing so all frontend routes resolve to `index.html`. ([fix-026])
- Fixed frontend port mapping in `docker-compose.yml` — was `5173:5173` (Vite dev), now `3000:80` (nginx). ([fix-026])
- Removed stale dev-mode volume mounts from frontend service. ([fix-026])
- Added Postgres 17 service to `docker-compose.yml` with persistent volume. ([fix-015])
- Added Postgres `pg_isready` healthcheck; backend depends on `service_healthy`. ([fix-026])
- Switched `Makefile` from deprecated `docker-compose` to `docker compose` (Compose v2). ([fix-014])

### Database

- Ran `sqlc generate` — produced type-safe Go code for all 5 query files (24 queries total). ([impl-sqlc])
- Generated: `db.go`, `models.go`, `connections.sql.go`, `agent_config.sql.go`, `investigations.sql.go`, `conversations.sql.go`, `log_buffer.sql.go`. ([impl-sqlc])
- Added index `idx_connections_status` on `connections.status` for filtered queries. ([fix-020])

### Docs & Config

- Created `.env.example` documenting all required and optional environment variables. ([fix-025])
- Added setup instructions, prerequisites, and migration guide to `README.md`. ([fix-025])

[fix-013]: completions/fix-013-add-dockerfiles.md
[fix-014]: completions/fix-014-docker-compose-v2.md
[fix-015]: completions/fix-015-add-postgres-service.md
[fix-020]: completions/fix-020-connections-status-index.md
[fix-025]: completions/fix-025-env-example-readme.md
[fix-026]: completions/fix-026-dockerfile-healthcheck.md
[impl-sqlc]: completions/impl-sqlc-generate.md

## 0.1.2 — Frontend Fixes (2026-02-20)

WebSocket data flow, routing guards, auth wiring, and component correctness.

### Composables & Stores

- Fixed `useAgent` — added `watch` on WebSocket data so agent replies are actually parsed and displayed. ([fix-004])
- Added axios request interceptor to attach `Authorization: Bearer` header from auth store. ([fix-022])
- Removed dead `useAuth` composable — logic already lived in `useAuthStore` Pinia store. ([fix-011])
- Extracted reports state into `useReportsStore` Pinia store for consistency with other pages. ([fix-024])

### Routing

- Added `beforeEach` auth guard — unauthenticated users redirect to `/login`. ([fix-021])
- Added catch-all `/:pathMatch(.*)*` route and `NotFoundPage.vue` for 404s. ([fix-021])

### Components

- Changed log severity class from `'error'` to `'critical'` to match the domain model type. ([fix-005])
- Updated `LogFilters.vue` dropdown value from `'error'` to `'critical'`. ([fix-005])
- Changed `ChatWindow.vue` `v-for` key from array index to `msg.id` (stable identity). ([fix-023])
- Added `id` field to `ChatMessage` type; assigned via `crypto.randomUUID()`. ([fix-023])

### Tooling

- Configured `typescript-eslint` parser so ESLint can lint `.ts` and `.vue` files. ([fix-012])

[fix-004]: completions/fix-004-useagent-websocket-data.md
[fix-005]: completions/fix-005-log-severity-mismatch.md
[fix-011]: completions/fix-011-remove-useauth-composable.md
[fix-012]: completions/fix-012-eslint-typescript-parser.md
[fix-021]: completions/fix-021-router-auth-guard-404.md
[fix-022]: completions/fix-022-axios-auth-interceptor.md
[fix-023]: completions/fix-023-chat-vfor-key.md
[fix-024]: completions/fix-024-reports-pinia-store.md

## 0.1.1 — Backend Fixes & Hardening (2026-02-20)

Dependency corrections, HTTP semantics, server hardening, and migration constraints.

### HTTP & Handlers

- Added `jsonError()` helper — replaces `http.Error` so JSON error responses keep `Content-Type: application/json`. ([fix-002])
- Introduced `Server` struct with `NewServer` constructor; converted all handlers from free functions to methods for dependency injection. ([fix-006])
- Changed `/api/auth/login` from `GET` to `POST`. ([fix-007])

### Agent & Connectors

- Registered 3 missing tools (`search_codebase`, `recall_similar_incidents`, `recall_lessons`) in `ToolRegistry()`. ([fix-003])
- Tool dispatch now returns an error for unknown tool names instead of silent empty success. ([fix-008])
- `Registry.Remove()` now calls `Close()` on the connector before deleting to prevent resource leaks. ([fix-010])

### Server

- Added `ReadTimeout`, `WriteTimeout`, `IdleTimeout` to `http.Server`. ([fix-017])
- Wrapped `ResponseWriter` in logging middleware to capture HTTP status codes. ([fix-018])

### Config

- Added `ElephantasmURL` and `ElephantasmKey` fields to `Config` struct. ([fix-016])
- Added `Config.Validate()` method; called on startup to fail fast on missing required env vars. ([fix-016])

### Migrations

- Enforced singleton constraint on `agent_config` (`id = 1` with `CHECK`). ([fix-009])
- Added `ON DELETE SET NULL` to `conversations.investigation_id` FK. ([fix-009])
- Added `ON DELETE CASCADE` to `log_buffer.connection_id` FK. ([fix-009])

### Dependencies

- Reclassified direct dependencies in `go.mod` (chi, websocket, etc. were incorrectly marked `// indirect`). ([fix-001])
- Removed 11 unused transitive entries from `go.mod` and `go.sum`. ([fix-001])

[fix-001]: completions/fix-001-go-mod-indirect.md
[fix-002]: completions/fix-002-http-error-content-type.md
[fix-003]: completions/fix-003-incomplete-tool-registry.md
[fix-006]: completions/fix-006-router-dependency-injection.md
[fix-007]: completions/fix-007-login-get-to-post.md
[fix-008]: completions/fix-008-unknown-tool-dispatch-error.md
[fix-009]: completions/fix-009-migration-fk-constraints.md
[fix-010]: completions/fix-010-registry-remove-close.md
[fix-016]: completions/fix-016-config-elephantasm-validate.md
[fix-017]: completions/fix-017-http-server-timeouts.md
[fix-018]: completions/fix-018-logging-status-code.md

---

## 0.1.0 — Scaffolding (2026-02-19)

Full monorepo scaffolding. Project structure, frontend, backend, database schema, and tooling — all initialised and compiling.

### Root

- Added `Makefile` with targets for dev, build, test, lint, sqlc, migrations, and Docker.
- Added `docker-compose.yml` (backend + frontend services, no local Postgres).
- Added `.gitignore` and `README.md`.

### Frontend (Vue 3 + Vite + TypeScript + Pinia + Tailwind v4)

- Initialised Vue 3 project with Vite, TypeScript, and Tailwind CSS v4.
- Installed runtime deps: `vue`, `vue-router`, `pinia`, `axios`.
- Created API client layer (`src/api/`) — one file per backend resource.
- Created 15 components across 5 domains: common, connections, agent chat, log feed, reports.
- Created 7 page views with Vue Router (lazy-loaded).
- Created 4 Pinia stores (auth, connections, agent, logs).
- Created 3 composables: `useWebSocket`, `useAgent`, `useAuth`.
- Created TypeScript types for all domain models + API envelope types.
- Created utility helpers (date formatting, app constants).
- Configured Vite dev proxy (`/api` → `:8080`, `/ws` → `ws://localhost:8080`).
- Verified: `vue-tsc` type-check passes with zero errors.

### Backend (Go + Chi v5 + pgx v5 + coder/websocket + anthropic-sdk-go)

- Initialised Go module (`github.com/crimson-sun/heimdall/backend`).
- Created HTTP server entry point with graceful shutdown.
- Created Chi router with all REST routes + WebSocket endpoint.
- Created 3 middleware layers: request logging, CORS, auth (passthrough placeholder).
- Created 6 handler files — all endpoints wired and responding (stubs).
- Created WebSocket chat handler with JSON read/write (placeholder echo).
- Created agent engine package: lifecycle, tool-use loop, monitoring goroutine, system prompt, 5 tool implementations (all stubs).
- Created connector layer: `Connector`, `StreamConnector`, `QueryConnector` interfaces + registry.
- Created 4 connector implementations: PostgreSQL, webhook logs, syslog, GitHub (all stubs with tests).
- Created Elephantasm memory client package (HTTP client, types, service layer).
- Created WebSocket hub package (hub, client, message types).
- Created report generation package (generator, templates).
- Verified: `go build ./...` and `go test ./...` both pass clean.

### Database

- Wrote 5 migration pairs (up/down) for: connections, agent_config, investigations, conversations, log_buffer.
- Wrote 5 sqlc query files covering all CRUD operations per the DB models spec.
- Created `sqlc.yaml` config (pgx/v5, uuid→uuid.UUID, jsonb→json.RawMessage, timestamptz→time.Time).
- Note: `sqlc generate` not yet run — requires a running Postgres instance.

### Docs

- Added `docs/overview-blueprint.md` — living architecture reference.
- Added `docs/completions/scaffolding.md` — detailed implementation notes.
