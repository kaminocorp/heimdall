# Webhook Ingestion — Phase 3: Explicit Format Routes & Detection Hardening

**Status:** Complete
**Date:** 2026-04-16
**Assessment:** `docs/executing/webhook-ingestion-assessment.md`
**Plan:** `docs/executing/webhook-ingestion-improvements-plan.md`
**Gaps addressed:** #7 (no explicit format path), #3 (precedence collisions)

---

## Summary

Added format-specific ingestion routes and an `X-Heimdall-Format` header
override so callers can bypass auto-detection entirely. Hardened three
auto-detection predicates to reduce false positive rates. Refactored the
handler into shared helpers (`readWebhookRequest`, `ingestEntries`) so
both auto-detect and format-specific paths share all downstream logic.

---

## Task 3.1 — Format-specific ingestion routes

**Files:** `backend/internal/api/handlers/webhooks.go`, `backend/internal/api/router.go`, `backend/internal/api/handlers/testhelpers_test.go`

### New routes

```
POST /api/webhooks/logs              ← auto-detect (existing, unchanged)
POST /api/webhooks/logs/{format}     ← format-specific (new)
```

Where `{format}` is one of: `flyio`, `vercel`, `firehose`, `pubsub`, `native`.

### How it works

A single `IngestWebhookLogsWithFormat` handler reads the `{format}` URL
parameter, validates it against the `validFormats` set, and calls
`parseByFormat` — a new function in `webhook_parsers.go` that dispatches
directly to the correct parser without running the auto-detection chain.

Both handlers share the same downstream path via two extracted helpers:

| Helper | Responsibility |
|--------|---------------|
| `readWebhookRequest` | Bearer token auth + body reading. Returns a `connResult` with connection/user/app IDs. |
| `ingestEntries` | Validation, transaction, insert, logging, response. Called identically by both handlers. |

### Why a single parameterised route instead of 5 separate routes

The plan suggested registering `/flyio`, `/vercel`, `/firehose`, `/pubsub`
as separate routes. A single `/{format}` route with a validated format set
is equivalent but avoids route registration duplication and makes adding
new formats a one-line change (add to `validFormats` + add a `case` in
`parseByFormat`).

Invalid format slugs return 404 with a structured error listing valid
formats, so there's no ambiguity.

---

## Task 3.2 — `X-Heimdall-Format` header override

**File:** `backend/internal/api/handlers/webhooks.go`

The auto-detect handler (`IngestWebhookLogs`) now checks for an
`X-Heimdall-Format` header before calling `parseWebhookPayload`. If
present:

1. The value is lowercased and validated against `validFormats`.
2. If valid, `parseByFormat` is called directly — auto-detection is
   skipped entirely.
3. If invalid, a 400 error lists the valid format names.

This gives callers who prefer the base URL a way to get deterministic
parsing without changing their endpoint configuration.

---

## Task 3.3 — Detection hardening

**File:** `backend/internal/api/handlers/webhook_parsers.go`

### NDJSON — removed structural fallback

**Before:** `isNDJSON` matched any input with 2+ lines where each line
starts with `{`. This matched Fly.io Vector batch payloads, native JSON
arrays formatted with newlines, and any multi-line JSON output.

**After:** NDJSON detection requires `Content-Type: application/x-ndjson`
header only. The structural `isNDJSON` function is still present (not
deleted) but no longer called from `parseWebhookPayload`. Callers without
the correct content-type should use `POST /api/webhooks/logs/vercel` or
`X-Heimdall-Format: vercel`.

### Firehose — require `records[0].data`

**Before:** Matched any payload with non-empty `requestId` string and
non-empty `records` JSON. App logs with field names like
`{"requestId": "req-123", "records": [{"id": "r1"}]}` were false positives.

**After:** The probe struct now unmarshals `records` as
`[]struct{ Data string }` and requires `records[0].Data != ""`. This
ensures the first record has a `data` field — the base64-encoded payload
that is the defining characteristic of Firehose delivery.

### Fly.io — exact match set instead of prefix

**Before:** `strings.HasPrefix(probe.SourceType, "fly")` matched
`"flywheel"`, `"flutter_app"`, and any other source type starting with
`"fly"`.

**After:** `knownFlySourceTypes` is a `map[string]bool` with four exact
values: `"fly_app"`, `"fly_io"`, `"fly_log_shipper"`, `"fly_app_logs"`.
The `fly` nested object check is unchanged — it's specific enough.

---

## Test changes

**File:** `backend/internal/api/handlers/webhook_parsers_test.go`

### Modified tests

| Test | Change |
|------|--------|
| `TestParseVercelNDJSON_AutoDetect` | **Renamed** to `TestParseVercelNDJSON_RequiresContentType`. Now asserts that NDJSON without content-type **fails** (was asserting success). |

### New tests

| Test | What it verifies |
|------|-----------------|
| `TestParseVercelNDJSON_ViaExplicitFormat` | NDJSON parsed correctly via `parseByFormat("vercel")` |
| `TestIsVectorFlyPayload_Negative` (extended) | `"flywheel"` and `"flutter"` no longer match Fly.io detection |
| `TestIsFirehosePayload_Negative` | Payload with `requestId` + `records` but no `.data` field does NOT match |
| `TestParseByFormat` (3 subtests) | Explicit format dispatch for flyio, native single, native batch |
| `TestParseByFormat_Unknown` | Unknown format returns error |

All 31 tests pass. `go vet` clean.

---

## Files changed

| File | Change |
|------|--------|
| `webhooks.go` | Refactored into `readWebhookRequest` + `ingestEntries` helpers; added `IngestWebhookLogsWithFormat` handler; added `X-Heimdall-Format` header check; added `validFormats` set |
| `webhook_parsers.go` | Added `parseByFormat` function; removed NDJSON structural fallback from `parseWebhookPayload`; hardened `isFirehosePayload` (require `data` field); replaced Fly.io prefix match with `knownFlySourceTypes` exact set |
| `router.go` | Registered `POST /api/webhooks/logs/{format}` |
| `testhelpers_test.go` | Registered format-specific route in test router |
| `webhook_parsers_test.go` | Updated 1 test, added 7 new tests (including 3 subtests) |
