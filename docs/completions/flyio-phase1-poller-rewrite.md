# Fly.io Phase 1 Completion — App-Level Poller Rewrite

**Date:** 2026-04-15
**Proposal:** `docs/executing/flyio-log-integration.md`
**Scope:** Task 1 from the proposal — rewrite the Fly.io API poller

---

## What Was Done

Replaced the N+1 per-machine polling approach with a single app-level NDJSON endpoint call. This was the critical backend fix identified in the proposal.

## Files Changed

| File | Change |
|------|--------|
| `backend/internal/connectors/logs/flyio.go` | Full rewrite of `Poll()` and `Connect()` |
| `backend/internal/connectors/logs/flyio_test.go` | New — 15 unit tests |

No changes were needed to `factory.go`, `connections_validate.go`, or `connections_test_handler.go` — all three reference `NewFlyio` which kept the same signature.

## What Changed in `flyio.go`

### `Poll()` — N+1 to single endpoint

**Before:** Listed all machines via `GET /v1/apps/{app}/machines`, filtered to `started`/`running`, then called `GET /v1/apps/{app}/machines/{id}/logs?limit=200` per machine. A 10-machine app made 11 API calls per poll cycle.

**After:** Single call to `GET /v1/apps/{app}/logs?start_time={cursor}`. Returns NDJSON (one JSON object per line) across all machines, including stopped and crashed ones. Parsed with `bufio.Scanner` for streaming memory efficiency.

### `Connect()` — Validates via logs endpoint

**Before:** Validated by listing machines (`GET /v1/apps/{app}/machines`).

**After:** Validates by hitting the app-level logs endpoint with a 1-minute lookback. This proves both token validity and app existence in a single call, and is consistent with what `Poll()` uses.

### `pollMachineLogs()` — Removed

No longer needed. All per-machine logic was in this function; the app-level endpoint returns logs from all machines in one response.

### New: `logsEndpoint()` helper

Builds the URL with proper escaping of app name and `start_time` query parameter. Shared between `Connect()` and `Poll()`.

### New: `flyLogEntry` struct

Captures the full NDJSON schema: `timestamp`, `message`, `level`, `instance` (machine ID), `region`, `meta`. The stored payload now includes `instance` and `region` fields that weren't available in the per-machine approach.

## Why These Choices

| Decision | Rationale |
|----------|-----------|
| **NDJSON line-by-line parsing** | The app-level endpoint streams NDJSON, not a JSON array. `bufio.Scanner` is more memory-efficient than buffering the entire body. |
| **10 MiB body cap** | Prevents unbounded memory growth from unexpectedly large responses, matching the previous `10<<20` limit. |
| **1 MiB per-line cap** | Scanner buffer limit prevents a single malformed mega-line from consuming memory. |
| **Cursor +1ns advancement** | Preserved from the original code. Fly.io timestamps have nanosecond precision; without this, logs with identical timestamps would be re-ingested. |
| **`time.Now()` fallback for unparseable timestamps** | Matches original behavior. Logs with bad timestamps still get ingested rather than silently dropped. |
| **Kept `normalizeSev()` unchanged** | Already correctly maps Fly.io's `error`/`err`/`warn`/`warning`/`debug` levels. |

## Problems Fixed

| Issue from proposal | Resolution |
|---------------------|------------|
| **N+1 API calls per poll** | Single endpoint call regardless of machine count |
| **Misses stopped/crashed machine logs** | App-level endpoint includes all machines |
| **Per-machine endpoint is undocumented** | Now uses the semi-public app-level endpoint (`flyctl` depends on it) |
| **200-entry cap per machine** | No per-machine cap; NDJSON streams all available entries since cursor |
| **Rate limiting risk** | 1 call vs N+1 dramatically reduces 429 surface |

## Test Coverage

15 tests in `flyio_test.go` covering:

- **Constructor:** valid config, defaults, missing fields, invalid JSON, poll interval minimum
- **Connect:** success, 401 unauthorized, 404 not found
- **Poll:** request formation (URL path, auth header, start_time param), empty response, rate limit 429, server error 500, malformed NDJSON lines, old entries filtered by cursor
- **Helpers:** `logsEndpoint` URL building, app name escaping, `normalizeSev` for all level values
- **Lifecycle:** `Close()`, `Health()`

All tests use `httptest.NewServer` with injected `apiBase` — no real Fly.io API calls.

## What's Next (Phase 2+)

Per the proposal, remaining tasks:
1. **Redesign wizard flow** — mode selection (drain vs polling), dynamic `connectorType` switching
2. **Create wizard components** — `StepFlyioMode`, `StepFlyioAuth`, `StepFlyioDrainSetup`
3. **Vector webhook parser** — test and add if the Log Shipper format doesn't auto-detect
4. **End-to-end testing** — both drain and polling paths
