# Phase 2 Completion — OTLP HTTP Receiver & JS/TS SDK

**Date:** 2026-04-05
**Roadmap ref:** `docs/executing/ingestion-roadmap.md` — Phase 2, items 3 & 4
**Status:** Complete

---

## What was built

Two new ingestion paths that together cover OpenTelemetry-instrumented applications, serverless environments, and any Node.js app:

1. **OTLP HTTP Receiver** — `POST /api/v1/logs` accepting OpenTelemetry Protocol JSON payloads
2. **JS/TS SDK** — `@heimdall/sdk` npm package with batching, retry, and a clean logging API

---

## Item 1: OTLP HTTP Receiver

### Architecture

The OTLP endpoint lives alongside the existing webhook endpoint as a public (pre-auth) route. Both use bearer token authentication via `GetConnectionByWebhookToken`. The OTLP handler parses the nested OTel structure (`resourceLogs → scopeLogs → logRecords`), flattens it, and inserts individual entries into `log_buffer` via the same `InsertLogEntry` path.

### OTLP payload mapping

The handler accepts the standard `ExportLogsServiceRequest` JSON format:

```
resourceLogs[].resource.attributes    → payload.resource (flattened to map)
resourceLogs[].scopeLogs[].scope.name → payload.scope
logRecords[].timeUnixNano             → payload.time_unix_nano
logRecords[].severityText             → payload.severity_text
logRecords[].severityNumber           → payload.severity_number + Heimdall severity mapping
logRecords[].body                     → payload.body (resolved AnyValue)
logRecords[].attributes               → payload.attributes (flattened to map)
logRecords[].traceId / spanId         → payload.trace_id / span_id
```

The `source_type` is `"otlp"` by default, or `"otlp/<service.name>"` when the resource has a `service.name` attribute.

### Severity mapping

OTLP uses numeric severity (1–24) with named ranges:

| OTLP range | OTLP name | Heimdall severity |
|------------|-----------|-------------------|
| 1–8 | Trace, Debug | `debug` |
| 9–12 | Info | `info` |
| 13–16 | Warn | `warning` |
| 17–20 | Error | `error` |
| 21–24 | Fatal | `critical` |

Falls back to `severityText` string matching when the number is 0.

### Connection type

New connection type `"otlp"` added to `validConnectionTypes`. OTLP connections auto-generate a webhook token on creation (same mechanism as `webhook_logs`), which the sender uses as the `Authorization: Bearer <token>` header.

### Files

| File | Purpose |
|------|---------|
| `backend/internal/api/handlers/otlp.go` | OTLP handler — request parsing, payload flattening, severity mapping, log insertion |
| `backend/internal/api/handlers/otlp_test.go` | 7 unit tests — severity by number, severity by text, service name extraction, attribute flattening, value resolution, record counting |
| `backend/internal/api/router.go` | Added `POST /api/v1/logs` route |
| `backend/internal/api/handlers/connections.go` | Added `"otlp"` to valid types, auto-generate token for OTLP connections |
| `frontend/src/components/connections/wizard/steps/StepOTLPSetup.vue` | Wizard step — endpoint format, payload example, compatible sources list |
| `frontend/src/components/connections/wizard/flows.ts` | OTLP flow: Name → Setup (auto-valid) |

### Test coverage

| Test | What it validates |
|------|-------------------|
| `TestMapOTLPSeverity_ByNumber` | All 5 OTLP severity ranges → Heimdall mapping |
| `TestMapOTLPSeverity_ByText` | Text fallback: FATAL, ERROR, WARN, INFO, DEBUG, unknown |
| `TestExtractServiceName` | Finds `service.name` from resource attributes |
| `TestExtractServiceName_Missing` | Returns empty when attribute is absent |
| `TestFlattenAttributes` | Converts OTel key-value list to simple map |
| `TestFlattenAttributes_Empty` | Returns nil for empty/nil input |
| `TestResolveAnyValue` | Extracts string, int, bool values from AnyValue |
| `TestCountLogRecords` | Counts across nested resourceLogs/scopeLogs |

---

## Item 2: JS/TS SDK

### Design

The SDK is a zero-dependency TypeScript package that posts to Heimdall's webhook endpoint (`POST /api/webhooks/logs`) with automatic batching and retry. It uses the global `fetch` API (available in Node 18+, Bun, Deno, Cloudflare Workers, and all browsers).

### API surface

```typescript
import { Heimdall } from '@heimdall/sdk'

const monitor = new Heimdall({
  endpoint: 'https://heimdall.example.com',
  token: 'whk_...',
  batchSize: 25,        // flush after N entries (default: 25)
  flushInterval: 5000,  // flush after N ms (default: 5000)
  maxRetries: 3,        // retry 5xx/network errors (default: 3)
  onError: (err, entries) => console.error(err),
})

// Severity-specific methods
monitor.info('user.signup', { userId: '123', plan: 'pro' })
monitor.error('payment.failed', { orderId: '456', error: err.message })
monitor.warn('rate.limit.near', { current: 95, max: 100 })
monitor.debug('cache.miss', { key: 'user:123' })
monitor.critical('db.connection.lost', { host: 'prod-db' })

// Generic method
monitor.log('info', 'custom.event', { data: 'value' })

// Graceful shutdown (flushes remaining buffer)
await monitor.shutdown()
```

### Batching strategy

- Entries accumulate in an in-memory buffer
- Flush triggers when **either** condition is met:
  - Buffer reaches `batchSize` (default 25)
  - `flushInterval` timer fires (default 5000ms)
- Flush sends the entire buffer as a JSON array in a single HTTP POST
- The timer is reset after each flush and only starts when the buffer is non-empty

### Retry strategy

- **4xx errors** (bad token, bad payload): no retry, calls `onError` immediately
- **5xx errors** (server error): retries up to `maxRetries` with exponential backoff (`retryDelay * 2^attempt`)
- **Network errors** (fetch failed): same retry behaviour as 5xx
- After all retries exhausted: calls `onError` with the error and the failed entries

### Package structure

```
packages/sdk-js/
  src/
    index.ts        — Heimdall class, types, batching, retry logic
    index.test.ts   — 10 unit tests
  package.json      — @heimdall/sdk, dual ESM/CJS output via tsup
  tsconfig.json     — ES2020 target, strict mode
  dist/             — Built output (ESM .js, CJS .cjs, .d.ts)
```

### Output formats

| File | Format | Size |
|------|--------|------|
| `dist/index.js` | ESM | 3.6 KB |
| `dist/index.cjs` | CommonJS | 4.6 KB |
| `dist/index.d.ts` | TypeScript declarations | 2.4 KB |

### Test coverage

| Test | What it validates |
|------|-------------------|
| `buffers entries until batchSize` | Buffer accumulates, auto-flushes at threshold |
| `flushes on timer` | Timer-based flush when batch isn't full |
| `sends correct payload format` | Correct URL, headers, JSON body structure |
| `strips trailing slash from endpoint` | URL normalization |
| `provides severity shorthands` | debug/info/warn/error/critical map correctly |
| `calls onError for 4xx without retry` | Client errors not retried |
| `retries on 5xx and calls onError after exhaustion` | Exponential backoff, retry count |
| `retries on network error` | fetch failures trigger retry |
| `flush() no-op when empty` | No HTTP call for empty buffer |
| `shutdown flushes remaining` | Graceful drain on process exit |

---

## Phase 2 status

| Item | Status |
|------|--------|
| OTLP HTTP receiver | **Done** |
| JS/TS SDK | **Done** |
| **Phase 2** | **Complete** |

### Platforms now unlocked (cumulative)

| Platform | Ingestion path | Phase |
|----------|---------------|-------|
| Render | Syslog | 1 |
| Heroku (Cedar) | Syslog or Webhook | 1 |
| DigitalOcean | Syslog | 1 |
| Linux servers | Syslog | 1 |
| Supabase | API Poller | 1 |
| **Neon** | **OTLP** | **2** |
| **Heroku Fir** | **OTLP** | **2** |
| **OTel-instrumented apps** | **OTLP** | **2** |
| **OTel Collector** | **OTLP** | **2** |
| **Fluent Bit / Vector** | **OTLP** | **2** |
| **Any Node.js app** | **SDK** | **2** |
| **Serverless (Lambda, Edge, Workers)** | **SDK** | **2** |
| **Any app with fetch** | **SDK** | **2** |

### Coverage estimate

Per the ingestion roadmap's coverage model:
- Phase 1: ~45% (webhook 30% + syslog 15%)
- Phase 2: +25% (OTLP 15% + SDK 10%)
- **Total: ~70%** of early users covered

---

## What's next (Phase 3)

Per `docs/executing/ingestion-roadmap.md`:

1. **Webhook payload parsers** — Vercel NDJSON, AWS Firehose envelope, GCP Pub/Sub envelope (unlocks platform-specific webhook formats without new connector types)
2. **Additional API pollers** — Fly.io, Vercel, Railway, MongoDB Atlas
3. **Python SDK** — Django, FastAPI coverage
4. **Go SDK** — Go service coverage
