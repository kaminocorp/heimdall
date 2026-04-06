# Phase 3 Completion — Webhook Parsers, API Pollers, Python & Go SDKs

**Date:** 2026-04-06
**Roadmap ref:** `docs/executing/ingestion-roadmap.md` — Phase 3, items 5–8
**Status:** Complete

---

## What was built

Four items completing the ingestion roadmap:

1. **Webhook payload parsers** — Vercel NDJSON, AWS Kinesis Firehose, GCP Pub/Sub format detection and normalization
2. **API pollers** — Fly.io, Vercel, Railway, MongoDB Atlas (4 new poll-based connectors)
3. **Python SDK** — `heimdall-sdk` PyPI package with batching and retry
4. **Go SDK** — `github.com/hejijunhao/heimdall/sdk-go` module with batching and retry

---

## Item 5: Webhook Payload Parsers

### Problem

The webhook endpoint (`POST /api/webhooks/logs`) only accepted Heimdall's native JSON format. Platform log drains send different formats: Vercel sends NDJSON, AWS Firehose wraps records in a delivery envelope with base64-encoded data, and GCP Pub/Sub push subscriptions wrap messages in their own envelope. Users had to transform payloads before sending, which defeated the purpose of native integration.

### Solution

Added automatic format detection to the webhook handler. The `parseWebhookPayload` function inspects the `Content-Type` header and payload structure to determine the format, then normalizes it to `[]webhookLogRequest` before the existing insertion loop.

### Format detection order

1. **Vercel NDJSON** — detected by `Content-Type: application/x-ndjson` or multiple newline-separated JSON lines
2. **AWS Firehose** — detected by presence of `"requestId"` and `"records"` fields in the JSON
3. **GCP Pub/Sub** — detected by presence of `"message"` and `"subscription"` fields
4. **Heimdall native** — default fallback (single object or JSON array)

### Vercel NDJSON parsing

Each line is parsed as a `vercelLogEntry` with fields: `message`, `timestamp`, `source` (build/lambda/edge/static), `projectName`, `level`, `statusCode`, `host`, `path`. The `source_type` is set to `"vercel/<source>"` (e.g. `"vercel/lambda"`). The Vercel `level` field maps to Heimdall severity.

### AWS Firehose parsing

The Firehose delivery envelope contains `records[].data` as base64-encoded strings. Each record's decoded data may be a single JSON object, NDJSON, or plain text. The parser handles all three. Severity is extracted from common fields (`severity`, `level`, `log_level`). Source type is `"firehose"`.

### GCP Pub/Sub parsing

The Pub/Sub push message contains `message.data` as base64-encoded content. The parser decodes it and tries JSON array → single JSON object → plain text. Metadata (`messageId`, `attributes`, `subscription`) is included in the payload for plain text messages. Source type is `"pubsub"`.

### Shared severity normalization

A `normalizeSeverity` function maps common severity strings from any platform to Heimdall's 5-level system: `emergency/fatal/critical/crit` → `critical`, `error/err` → `error`, `warning/warn` → `warning`, `notice/info/informational` → `info`, `debug/trace` → `debug`.

### Files

| File | Purpose |
|------|---------|
| `backend/internal/api/handlers/webhook_parsers.go` | Format detection, Vercel/Firehose/Pub/Sub/native parsers, severity normalization |
| `backend/internal/api/handlers/webhook_parsers_test.go` | 12 tests — native single/batch, Vercel NDJSON (explicit + auto-detect), Firehose JSON + raw text, Pub/Sub JSON + plain text, severity normalization, severity extraction, empty payload |
| `backend/internal/api/handlers/webhooks.go` | Refactored to use `parseWebhookPayload` with size-limited body reading |

### Test coverage

| Test | Format |
|------|--------|
| `TestParseNativePayload_Single` | Heimdall native (single) |
| `TestParseNativePayload_Batch` | Heimdall native (array) |
| `TestParseVercelNDJSON` | Vercel with ndjson content-type |
| `TestParseVercelNDJSON_AutoDetect` | Vercel auto-detected by structure |
| `TestParseFirehosePayload` | Firehose with JSON records |
| `TestParseFirehosePayload_RawText` | Firehose with plain text records |
| `TestParsePubSubPayload_JSON` | Pub/Sub with JSON data |
| `TestParsePubSubPayload_PlainText` | Pub/Sub with plain text data |
| `TestNormalizeSeverity` | 12 severity string mappings |
| `TestExtractSeverityFromMap` | Severity field extraction from maps |
| `TestParseWebhookPayload_Empty` | Empty payload handling |

---

## Item 6: API Pollers

### Architecture

All four pollers follow the Supabase pattern: implement `PollConnector` (Connect, Health, Close, Poll), load config from the connection's JSONB, use cursor-based polling with rate limit awareness, and insert into `log_buffer` via `InsertLogEntry`.

### Fly.io Poller (`flyio.go`)

| Field | Detail |
|-------|--------|
| **API** | Fly.io Machines API (`api.machines.dev`) |
| **Auth** | Bearer token |
| **Config** | `app_name`, `api_token`, `poll_interval_secs` (min 15s, default 30s) |
| **Strategy** | Lists machines → fetches logs per running machine → cursor-based dedup by timestamp |
| **Source type** | `flyio/<app_name>` |
| **Connection type** | `"flyio"` |

### Vercel Poller (`vercel.go`)

| Field | Detail |
|-------|--------|
| **API** | Vercel REST API (`api.vercel.com`) |
| **Auth** | Bearer token |
| **Config** | `api_token`, `project_id`, `team_id` (optional), `poll_interval_secs` (min 30s, default 60s) |
| **Strategy** | Polls deployment list filtered by `since` timestamp, captures deployment events (READY/ERROR) with git metadata |
| **Source type** | `vercel/deployment` |
| **Connection type** | `"vercel"` |

### Railway Poller (`railway.go`)

| Field | Detail |
|-------|--------|
| **API** | Railway GraphQL API (`backboard.railway.app/graphql/v2`) |
| **Auth** | Bearer token |
| **Config** | `api_token`, `project_id`, `service_id` (optional), `environment_id` (optional), `poll_interval_secs` (min 30s, default 60s) |
| **Strategy** | GraphQL query for `deploymentLogs` with `since` cursor, handles GraphQL errors |
| **Source type** | `railway/<project_id>` |
| **Connection type** | `"railway"` |

### MongoDB Atlas Poller (`mongodb.go`)

| Field | Detail |
|-------|--------|
| **API** | MongoDB Atlas Admin API v2 (`cloud.mongodb.com/api/atlas/v2`) |
| **Auth** | HTTP Basic (public key + private key) |
| **Config** | `public_key`, `private_key`, `group_id` (project), `cluster_name`, `poll_interval_secs` (min 60s, default 120s) |
| **Strategy** | Discovers cluster hostname → polls `dbAccessHistory` endpoint with date range → flags auth failures as warnings, failed logins as errors |
| **Source type** | `mongodb/<cluster_name>` |
| **Connection type** | `"mongodb"` |

### Connection handler changes

- All four types added to `validConnectionTypes` map
- Config validation on create via each connector's constructor
- `startPoller` helper function consolidates poller initialization for all 5 poll-based types (Supabase + 4 new)
- `resumePollers` in `main.go` rewritten as a loop over all poller types — no more Supabase-only hardcoding
- Update and delete handlers automatically stop/restart pollers for all types

### Files

| File | Purpose |
|------|---------|
| `backend/internal/connectors/logs/flyio.go` | Fly.io Machines API poller |
| `backend/internal/connectors/logs/vercel.go` | Vercel REST API poller |
| `backend/internal/connectors/logs/railway.go` | Railway GraphQL API poller |
| `backend/internal/connectors/logs/mongodb.go` | MongoDB Atlas Admin API poller |
| `backend/internal/api/handlers/connections.go` | 4 new types in validation map, config validation, `startPoller` helper |
| `backend/cmd/heimdall/main.go` | Generic `resumePollers` for all 5 poller types |

---

## Item 7: Python SDK

### Design

Zero-dependency Python package using `urllib.request` (stdlib). Thread-safe batching with a background timer, exponential backoff retry, and the same API shape as the JS SDK.

### API

```python
from heimdall_sdk import Heimdall, HeimdallOptions

monitor = Heimdall(HeimdallOptions(
    endpoint="https://heimdall.example.com",
    token="whk_...",
))

monitor.info("user.signup", {"user_id": "123"})
monitor.error("payment.failed", {"order_id": "456"})
monitor.shutdown()
```

### Package structure

```
packages/sdk-python/
├── heimdall_sdk/
│   ├── __init__.py      Exports: Heimdall, HeimdallOptions, LogEntry
│   └── client.py        Client class, batching, retry, severity methods
├── tests/
│   └── test_client.py   9 tests
└── pyproject.toml       hatchling build, Python 3.9+
```

### Key differences from JS SDK

- Uses `threading.Timer` for flush scheduling (Python has no event loop by default)
- Flush sends via background `threading.Thread` to avoid blocking the caller
- Uses `urllib.request` (stdlib) — no `requests`/`httpx` dependency
- Thread-safe via `threading.Lock`

### Test coverage (9 tests)

`test_log_entry_to_dict`, `test_buffers_entries`, `test_severity_shorthands`, `test_flush_sends_batch`, `test_auto_flush_on_batch_size`, `test_shutdown_flushes`, `test_empty_flush_is_noop`, `test_endpoint_trailing_slash`, `test_default_payload`

---

## Item 8: Go SDK

### Design

Standalone Go module (`github.com/hejijunhao/heimdall/sdk-go`). Zero dependencies beyond stdlib. Uses `sync.Mutex` for thread safety, `time.AfterFunc` for flush timer, and goroutines for async send.

### API

```go
monitor := heimdall.New(heimdall.Options{
    Endpoint: "https://heimdall.example.com",
    Token:    "whk_...",
})
defer monitor.Shutdown()

monitor.Info("user.signup", map[string]any{"user_id": "123"})
monitor.Error("payment.failed", map[string]any{"order_id": "456"})
```

### Package structure

```
packages/sdk-go/
├── go.mod             Module: github.com/hejijunhao/heimdall/sdk-go
├── heimdall.go        Client, Options, LogEntry, batching, retry
└── heimdall_test.go   9 tests (using httptest for real HTTP)
```

### Key differences from other SDKs

- Tests use `net/http/httptest` for real HTTP server assertions (not mocks)
- Custom `*http.Client` injectable via `Options.HTTPClient`
- Goroutine-based async send with `sync.Mutex` guarding the buffer

### Test coverage (9 tests)

`TestNew_Defaults`, `TestNew_StripTrailingSlash`, `TestClient_BuffersEntries`, `TestClient_SeverityShorthands`, `TestClient_PayloadFormat`, `TestClient_FlushEmpty`, `TestClient_ShutdownFlushes`, `TestClient_NilPayload`, `TestClient_OnError4xx`

---

## Phase 3 status

| Item | Status |
|------|--------|
| Webhook payload parsers (Vercel, Firehose, Pub/Sub) | **Done** |
| API pollers (Fly.io, Vercel, Railway, MongoDB Atlas) | **Done** |
| Python SDK | **Done** |
| Go SDK | **Done** |
| **Phase 3** | **Complete** |

---

## Ingestion Roadmap — Final Status

| Phase | Items | Status |
|-------|-------|--------|
| Phase 1 | Syslog TLS listener, Supabase API poller | **Complete** |
| Phase 2 | OTLP HTTP receiver, JS/TS SDK | **Complete** |
| Phase 3 | Webhook parsers, API pollers, Python SDK, Go SDK | **Complete** |
| **All phases** | **All 8 items** | **Complete** |

### Cumulative platform coverage

| Platform | Ingestion path | Phase |
|----------|---------------|-------|
| Supabase | API Poller | 1 |
| Render | Syslog | 1 |
| Heroku (Cedar) | Syslog or Webhook | 1 |
| DigitalOcean | Syslog | 1 |
| Linux servers | Syslog | 1 |
| Neon | OTLP | 2 |
| Heroku Fir | OTLP | 2 |
| OTel-instrumented apps | OTLP | 2 |
| Any Node.js app | JS SDK | 2 |
| Serverless (Lambda, Edge, Workers) | JS SDK | 2 |
| **Vercel** | **API Poller + NDJSON webhook parser** | **3** |
| **Fly.io** | **API Poller** | **3** |
| **Railway** | **API Poller** | **3** |
| **MongoDB Atlas** | **API Poller** | **3** |
| **AWS (via Firehose)** | **Webhook parser** | **3** |
| **GCP (via Pub/Sub)** | **Webhook parser** | **3** |
| **Any Python app** | **Python SDK** | **3** |
| **Any Go app** | **Go SDK** | **3** |

### Coverage estimate: ~80%

Per the roadmap's original model, all three phases together provide approximately 80% coverage of early users' infrastructure.

### Total test count across all phases

| Component | Tests |
|-----------|-------|
| Syslog connector | 9 |
| OTLP handler | 7 |
| JS/TS SDK | 10 |
| Webhook parsers | 12 |
| Go SDK | 9 |
| Python SDK | 9 |
| **Total new tests** | **56** |
