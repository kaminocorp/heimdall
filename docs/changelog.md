# Changelog

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
