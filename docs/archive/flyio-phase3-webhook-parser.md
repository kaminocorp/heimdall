# Fly.io Phase 3 Completion — Vector/Fly Log Shipper Webhook Parser

**Date:** 2026-04-15
**Proposal:** `docs/executing/flyio-log-integration.md`
**Scope:** Task 4 from the proposal — Vector webhook parser for the drain path

---

## What Was Done

Added a dedicated webhook parser for the Fly Log Shipper (Vector HTTP sink) format. This completes the drain path end-to-end: when the Fly Log Shipper POSTs to Heimdall's webhook endpoint, the payload is now correctly detected, parsed, and stored with Fly-specific metadata (app name, machine ID, region).

## Files Modified

| File | Change |
|------|--------|
| `backend/internal/api/handlers/webhook_parsers.go` | Added `isVectorFlyPayload()`, `isVectorFlyObject()`, `parseVectorFlyPayload()`, `parseVectorFlyEntry()`, and `vectorFlyEntry` struct |
| `backend/internal/api/handlers/webhook_parsers_test.go` | Added 6 new tests for the Vector/Fly parser |

## How Detection Works

The parser sits in the detection chain between GCP Pub/Sub and the Heimdall native fallback:

```
parseWebhookPayload()
  1. Vercel NDJSON    → content-type or structural detection
  2. AWS Firehose     → requestId + records
  3. GCP Pub/Sub      → message.data + subscription
  4. Vector/Fly (NEW) → "fly" nested object or source_type starts with "fly"
  5. Heimdall native  → fallback (single object or array)
```

**Detection heuristic** (`isVectorFlyPayload`):
- For single objects: checks for a `fly` nested object OR `source_type` starting with `"fly"`
- For arrays: inspects the first element using the same logic
- This catches both the Fly Log Shipper's specific format (with `fly.app.name` metadata) and simpler Vector configurations that set `source_type: "fly_io"`

**Why this ordering is safe**: The detection runs after Firehose and Pub/Sub (which have very distinct structural markers). It won't false-positive on Heimdall native payloads because those use `source_type` values like `"app"` or `"my-service"` — not `"fly*"`. And native payloads that happen to have a `fly` key would need it to be a JSON object (not a string), which is extremely unlikely.

## Parser Details

### Input format (Fly Log Shipper)

The Fly Log Shipper sends JSON objects with Fly-specific metadata:

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

Can arrive as a single object or a JSON array (Vector batch mode).

### Output (webhookLogRequest)

```go
webhookLogRequest{
    SourceType: "flyio/my-fly-app",     // from fly.app.name
    Severity:   "info",                  // from log.level, normalized
    Payload:    {                        // enriched JSON
        "message": "request completed in 12ms",
        "timestamp": "2026-04-15T12:00:00.123Z",
        "host": "e784079c",
        "app_name": "my-fly-app",
        "machine_id": "e784079c",
        "region": "lhr"
    }
}
```

### Field extraction

| Field | Source | Fallback |
|-------|--------|----------|
| `SourceType` | `"flyio/" + fly.app.name` | `source_type` from Vector, then `"flyio"` |
| `Severity` | `log.level` → `normalizeSeverity()` | `extractSeverityFromMap()` (checks `severity`, `level`, `log_level`, `error_severity` keys) |
| `Payload.message` | `message` | — |
| `Payload.timestamp` | `timestamp` | — |
| `Payload.host` | `host` | — |
| `Payload.app_name` | `fly.app.name` | omitted if no fly metadata |
| `Payload.machine_id` | `fly.machine.id` | omitted if no fly metadata |
| `Payload.region` | `fly.region` | omitted if no fly metadata |

## Test Coverage

6 new tests in `webhook_parsers_test.go`:

| Test | Scenario |
|------|----------|
| `TestParseVectorFly_Single` | Single Fly Log Shipper entry with full metadata |
| `TestParseVectorFly_Batch` | Array of entries (Vector batch mode) |
| `TestParseVectorFly_SeverityMapping` | `log.level: "warn"` → `"warning"` |
| `TestParseVectorFly_DetectBySourceType` | No `fly` object, but `source_type: "fly_io"` triggers detection |
| `TestParseVectorFly_NoFlyMetadata` | Falls back to `source_type` and `extractSeverityFromMap()` |
| `TestIsVectorFlyPayload_Negative` | Verifies native/other formats don't false-positive |

All existing parser tests continue to pass (19 total in the file).

## Build Status

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — all passing (0 failures)

## End-to-End Drain Path (Complete)

With all three phases done, the drain path works as follows:

```
User selects "Fly.io → Log Drain" in wizard
  → Wizard creates webhook_logs connection (server generates token)
  → User deploys Fly Log Shipper with webhook URL + token
  → Log Shipper connects to Fly's internal NATS stream
  → Log Shipper POSTs to /api/webhooks/logs
  → Handler extracts bearer token → finds connection
  → parseWebhookPayload() → isVectorFlyPayload() → true
  → parseVectorFlyPayload() extracts message, severity, Fly metadata
  → InsertLogEntry() stores in log_buffer
  → Monitoring loop picks up entries for classification + escalation
```
