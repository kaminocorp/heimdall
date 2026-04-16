# Webhook Ingestion — Phase 4: Native Format v2

**Status:** Complete
**Date:** 2026-04-16
**Assessment:** `docs/executing/webhook-ingestion-assessment.md`
**Plan:** `docs/executing/webhook-ingestion-improvements-plan.md`
**Gaps addressed:** #2 (payload field name confusion), #8 (Fly.io wizard accuracy — partial)

---

## Summary

Introduced a v2 native webhook format with clearer field names (`source`,
`level`, `message`, `attrs`), replacing the confusing v1 names
(`source_type`, `severity`, `payload`). The v1 format remains fully
supported with deprecation signals. Updated all user-facing UI surfaces
to show v2 as the recommended format.

---

## Task 4.1 — Dual-format native parsing

**File:** `backend/internal/api/handlers/webhook_parsers.go`

### v2 format

```json
{
  "source": "my-service",
  "level": "error",
  "message": "connection refused to db-primary",
  "attrs": {
    "host": "web-1",
    "request_id": "req-abc123"
  }
}
```

| v2 field | v1 equivalent | Required | Rationale |
|----------|--------------|----------|-----------|
| `source` | `source_type` | Yes | Shorter, more natural |
| `level` | `severity` | No (defaults to "info") | Standard across logging frameworks |
| `message` | (buried in `payload`) | No | First-class — the thing devs search for |
| `attrs` | `payload` | No | No collision with "payload" as request body concept |

### Detection logic

A `nativeVersionProbe` struct reads both `source` and `source_type` from
the JSON object:

- **`source` present** → v2 (even if `source_type` is also present)
- **`source_type` present, no `source`** → v1
- **Neither present** → v1 (will fail validation in the handler)

### v2 → v1 normalisation

v2 entries are normalised to the internal `webhookLogRequest` before
insertion, so no database schema changes are needed:

```go
entry.SourceType = v2.Source
entry.Severity   = normalizeSeverity(v2.Level)  // defaults to "info"
entry.Payload    = marshal(merge(attrs, {"message": v2.Message}))
```

The `message` field is merged into `attrs` to become part of the stored
payload, ensuring it's searchable alongside other attributes.

### Batch handling

For batches, the first entry is probed to determine the version. All
entries must use the same version — mixed v1+v2 batches are rejected with
a clear error:

```
entry 1: mixed v1/v2 formats in batch — all entries must use the same format
```

### Format strings

| Payload | Format string |
|---------|--------------|
| v1 single | `native` |
| v1 batch | `native_batch` |
| v2 single | `native_v2` |
| v2 batch | `native_v2_batch` |

The `nativeFormatString` helper determines this by probing the raw input
and checking entry count.

---

## Task 4.2 — UI example payload updates

### StepWebhookSetup.vue

Updated the example payload from a near-v2 format (had `source` but used
`severity` and `metadata`) to the exact v2 spec:

```json
{
  "source": "my-app",
  "level": "error",
  "message": "Connection refused to db-primary",
  "attrs": {
    "host": "web-1",
    "request_id": "req-abc123"
  }
}
```

### StepFlyioDrainSetup.vue

**Step 4 (Configure Vector)** was rewritten:

- **Removed** the misleading claim that the stock Fly Log Shipper includes
  structured `fly` metadata by default (it doesn't — NATS emits flat
  log lines). This partially addresses Gap #8.
- **VRL transform** now outputs v2 native format (`source`, `level`,
  `message`, `attrs`) instead of the old `source_type` tag.
- **Clarified** that the enriched `fly` metadata auto-detection only
  applies to HTTP log drain integrations, not NATS-based Log Shipper
  deployments.

**Reference section** split into two sub-sections:
- **Native v2 (recommended)** — the format produced by the Vector
  transform, with `source`, `level`, `message`, `attrs` fields.
- **Auto-detected format (HTTP log drain only)** — the enriched Fly.io
  metadata format, shown in muted styling to indicate it's secondary.

---

## Task 4.3 — Deprecation headers for v1 native format

**File:** `backend/internal/api/handlers/webhooks.go`

When a request is parsed as v1 native format (`"native"` or
`"native_batch"`), the response includes:

**Headers (RFC 8594 Sunset Header):**
```
Sunset: 2026-10-01
Deprecation: true
Link: <https://docs.heimdallwatch.com/api/webhook-format-v2>; rel="successor-version"
```

**Response body:**
```json
{
  "accepted": 5,
  "format": "native",
  "deprecated": true
}
```

The `deprecated` field uses `omitempty` — it's absent for v2 and all
non-native formats.

A `slog.Warn` is logged for every v1 request to track migration progress.

---

## Test changes

**File:** `backend/internal/api/handlers/webhook_parsers_test.go`

| Test | What it verifies |
|------|-----------------|
| `TestParseNativeV2_Single` | v2 single entry: correct format, source→SourceType mapping, message merged into payload |
| `TestParseNativeV2_Batch` | v2 batch: correct format (`native_v2_batch`), both entries parsed |
| `TestParseNativeV2_NoAttrs` | attrs is optional, severity normalisation ("warn" → "warning") |
| `TestParseNativeV2_NoLevel` | level defaults to "info" when absent |
| `TestParseNativeV2_MissingSource` | No `source` → falls through to v1 (empty SourceType, caught by handler validation) |
| `TestParseNativeV2_MixedBatch` | Mixed v1+v2 batch → error containing "mixed v1/v2" |
| `TestParseNativeV2_ViaExplicitFormat` | v2 works through `parseByFormat("native")` |
| `TestParseNativeV1_StillWorks` | v1 format unchanged — backward compatibility |
| `TestParseNativeV2_BothSourceFields` | Both `source` and `source_type` present → v2 wins |

All 42 tests pass. `go vet` clean. Frontend `vue-tsc` clean.

---

## Files changed

| File | Change |
|------|--------|
| `webhook_parsers.go` | `webhookLogRequestV2`, `nativeVersionProbe`, `nativeFormatString` types; `parseNativeSingle`, `parseNativeBatch`, `isNativeV2`, `parseNativeV2Entry` functions; refactored `parseNativePayload` |
| `webhooks.go` | `Deprecated` field on `webhookLogResponse`; Sunset/Deprecation/Link headers for v1; `slog.Warn` for v1 usage |
| `webhook_parsers_test.go` | 9 new tests for v2 parsing, backward compat, mixed batch rejection |
| `StepWebhookSetup.vue` | Example payload updated to v2 format |
| `StepFlyioDrainSetup.vue` | Step 4 rewritten with v2 VRL transform; reference section split into v2/auto-detect |
