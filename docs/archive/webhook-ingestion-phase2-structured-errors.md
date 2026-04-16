# Webhook Ingestion — Phase 2: Structured Errors & Format Transparency

**Status:** Complete
**Date:** 2026-04-16
**Assessment:** `docs/executing/webhook-ingestion-assessment.md`
**Plan:** `docs/executing/webhook-ingestion-improvements-plan.md`
**Gaps addressed:** #4 (opaque errors), #3 (detection visibility via format echo)

---

## Summary

Replaced all generic string errors in the webhook handler with structured
JSON error responses. Every error now includes a machine-readable code,
the detected format (when known), and per-entry field errors for batch
validation failures. Added server-side structured logging for ingestion
requests.

---

## Task 2.1 + 2.2 — Structured error responses with format echo

**File:** `backend/internal/api/handlers/webhooks.go`

### New types

```go
type webhookError struct {
    Error          string              `json:"error"`
    Message        string              `json:"message"`
    FormatDetected string              `json:"format_detected,omitempty"`
    Details        []webhookFieldError `json:"details,omitempty"`
}

type webhookFieldError struct {
    Index int    `json:"index"`
    Field string `json:"field"`
    Error string `json:"error"`
}
```

### Error codes

| Code | HTTP | When |
|------|------|------|
| `unauthorized` | 401 | Missing/invalid bearer token |
| `invalid_json` | 400 | Body cannot be read or is not valid JSON |
| `parse_failed` | 400 | Format detected but parsing failed (includes Go error message) |
| `empty_payload` | 400 | Parsing succeeded but produced zero entries |
| `validation_failed` | 400 | One or more entries missing `source_type` or `payload` |

### Key design decisions

- **`format_detected` is omitted (not `""`) when format is unknown.** This
  happens for auth errors and read failures — the body hasn't been parsed
  yet. Using `omitempty` keeps the response clean for callers that only
  check `error`.

- **`details` collects ALL validation errors, not just the first one.**
  The Phase 1 approach was fail-on-first with `fmt.Sprintf("entry %d: ...")`
  strings. Now the handler collects every invalid entry in the batch and
  returns them all at once. An automated caller fixing a batch of 50
  entries doesn't want to fix-and-retry 50 times.

- **`parse_failed` includes the Go error string.** For malformed JSON,
  this gives the caller the exact parse position (e.g. `"invalid character
  'x' looking for beginning of value"`). This is safe to expose — it
  contains no internal state, just JSON parsing diagnostics.

### What the caller sees

**Before (Phase 1):**
```json
{"error": "entry 2: source_type is required"}
```

**After (Phase 2):**
```json
{
  "error": "validation_failed",
  "message": "one or more entries failed validation",
  "format_detected": "native_batch",
  "details": [
    {"index": 2, "field": "source_type", "error": "required"},
    {"index": 4, "field": "payload", "error": "required"}
  ]
}
```

---

## Task 2.3 — Server-side structured logging

**File:** `backend/internal/api/handlers/webhooks.go`

Added three `slog` log points:

| Level | When | Fields |
|-------|------|--------|
| `slog.Info` | Successful ingestion (after commit) | `connection_id`, `format`, `entries`, `duration_ms` |
| `slog.Warn` | Parse error | `connection_id`, `error` |
| `slog.Warn` | Validation failure | `connection_id`, `format`, `errors` (count) |

### Why `slog` and not a custom logger

The codebase already uses `log/slog` throughout (`jsonServerError` in
`helpers.go` calls `slog.Error`). Keeping the same logger means webhook
logs appear alongside all other server logs with the same format and can
be filtered by the same tooling.

The `duration_ms` field on success logs enables latency monitoring — if
ingestion starts taking >100ms, something is wrong (likely DB contention
or large batches).

---

## Test changes

**File:** `backend/internal/api/handlers/webhook_parsers_test.go`

| Test | What it verifies |
|------|-----------------|
| `TestParseWebhookPayload_FormatEcho` (4 subtests) | Correct format string for native, native_batch, vercel_ndjson, flyio_vector |
| `TestParseWebhookPayload_InvalidJSON` | Malformed JSON returns error and empty format |
| `TestWebhookErrorSerialization` | `webhookError` with details serializes with correct field names, types, and indices |
| `TestWebhookErrorSerialization_NoDetails` | `omitempty` works: nil `details` and empty `format_detected` are omitted from JSON |

All 24 tests pass. `go vet` clean.

---

## Files changed

| File | Change |
|------|--------|
| `webhooks.go` | `webhookError`, `webhookFieldError` types; `jsonWebhookError` helper; all error paths use structured responses; validation collects all errors; `slog` logging on success/failure |
| `webhook_parsers_test.go` | 7 new tests (4 format echo subtests, 1 invalid JSON, 2 error serialization) |
