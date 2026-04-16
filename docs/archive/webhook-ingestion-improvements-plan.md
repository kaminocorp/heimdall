# Webhook Ingestion — Implementation Plan

**Status:** Not started
**Owner:** TBD
**Assessment:** `webhook-ingestion-assessment.md`
**Prereqs:** None — all changes are localised to the ingestion path

---

## Principles

1. **Non-breaking first.** Phases 1–3 fix correctness and DX issues without
   changing the existing API contract. Existing integrations keep working.
2. **Breaking changes last.** Phase 4 introduces the new native format with a
   migration path. Old format remains supported behind a compatibility parser.
3. **Each phase is independently shippable.** No phase depends on a later one.
   Any phase can be deferred or skipped without blocking the others.

---

## Phase 0 — Fly.io wizard accuracy fix

**Gaps addressed:** #8 (wizard assumes enriched payload NATS doesn't produce)
**Breaking changes:** None — UI text only
**Estimated scope:** Small — single Vue component
**Dependency:** None — can ship immediately

### Task 0.1 — Rewrite Step 4 to make the Vector transform the primary path

**File:** `frontend/src/components/connections/wizard/steps/StepFlyioDrainSetup.vue`

The current Step 4 tells users the stock Fly Log Shipper image includes
structured `fly` metadata "by default" and presents the Vector transform as
a fallback for "custom configs." This is wrong — the NATS source emits flat
log lines without that metadata. Most users will need the transform.

**Changes:**

1. **Remove the misleading "no extra config needed" claim** (lines 107-109).
   Replace with a clear statement that the Fly Log Shipper's NATS source emits
   raw log lines that need reshaping for Heimdall.

2. **Make the VRL transform the default guidance.** Present a complete,
   ready-to-use Vector config snippet that:
   - Reads from the NATS source (already configured by the stock image)
   - Reshapes events into Heimdall's native format (`source_type`, `severity`,
     `payload`)
   - Sends to the HTTP sink pointed at Heimdall's webhook URL

   ```toml
   [transforms.reshape_for_heimdall]
   type = "remap"
   inputs = ["fly_log_source"]
   source = '''
     .source_type = "fly_app"
     .severity = downcase(.log.level) ?? "info"
     .payload = {
       "message": .message,
       "timestamp": .timestamp,
       "host": .host
     }
   '''
   ```

3. **Update the collapsible reference section.** The "Expected Payload Format"
   currently lists `fly.app.name`, `fly.region`, etc. — fields the NATS source
   never emits. Relabel this section as "Auto-detected format (HTTP log drain
   only)" and add a separate section showing the native format that the VRL
   transform produces.

4. **Add a brief note explaining the two Fly.io log paths:**
   - **NATS source** (what the Fly Log Shipper uses): flat log lines, requires
     the Vector transform shown above.
   - **HTTP log drain** (Fly's built-in drain API): enriched format with `fly`
     metadata, auto-detected by Heimdall. Not used by the Log Shipper.

### Task 0.2 — Add integration test for NATS-shaped Fly.io payloads

**File:** `backend/internal/api/handlers/webhooks_test.go`

Add a test that sends a raw NATS-sourced log line (flat, no `fly` metadata)
and verifies it is **not** auto-detected as Fly.io format. This codifies the
real-world failure Trajan encountered.

Then add a test for the same payload after it's been reshaped by the VRL
transform (native format with `source_type: "fly_app"`), verifying it's
accepted and stored correctly.

**Steps:**
1. Test: raw NATS payload `{"message": "...", "timestamp": "..."}` without
   `fly` object or `source_type` → falls through to native → 400 (missing
   `source_type`).
2. Test: reshaped payload `{"source_type": "fly_app", "severity": "info",
   "payload": {"message": "..."}}` → 201, stored correctly.

---

## Phase 1 — Correctness & safety fixes

**Gaps addressed:** #1 (partial failure), #5 (response leak)
**Breaking changes:** Response shape changes (callers should not depend on it)
**Estimated scope:** Small — handler-only changes, no new files

### Task 1.1 — Wrap batch inserts in a database transaction

**File:** `webhooks.go`

Replace the bare insert loop with a transaction that commits atomically.
All entries succeed together or none are persisted.

```go
// Target pattern:
tx, err := s.Pool.Begin(r.Context())
if err != nil { ... }
defer tx.Rollback(r.Context())

qtx := s.Queries.WithTx(tx)
for _, e := range entries {
    // validate entry...
    row, err := qtx.InsertLogEntry(r.Context(), params)
    if err != nil { ... }
    inserted = append(inserted, row)
}

if err := tx.Commit(r.Context()); err != nil { ... }
```

Validation must happen for ALL entries before any insert begins — fail fast on
the full batch, not mid-loop.

**Steps:**
1. Add a validation pass that checks all entries for `source_type` and `payload`
   before starting the transaction. Collect all validation errors with their
   indices.
2. If any entry fails validation, return 400 with structured errors (see Task 2.2)
   and insert nothing.
3. Wrap the insert loop in `Pool.Begin` / `Commit`. Use `Queries.WithTx(tx)`.
4. On insert error, the deferred `Rollback` ensures nothing is persisted.
5. Add test: send a batch of 5 where entry 3 is invalid. Assert 0 rows inserted.
6. Add test: send a batch of 5, all valid. Assert 5 rows inserted atomically.

### Task 1.2 — Strip internal fields from the 201 response

**File:** `webhooks.go`

Replace the raw `LogBuffer` row in the response with a minimal acknowledgement
struct. The caller needs confirmation, not internal database state.

```go
// New response struct
type webhookLogResponse struct {
    Accepted int    `json:"accepted"`
    Format   string `json:"format"`   // which parser handled the request
}
```

**Steps:**
1. Define `webhookLogResponse` struct.
2. Thread the detected format name through `parseWebhookPayload` (return it as
   a second value, or wrap entries in a result struct with a `Format` field).
3. Replace `json.NewEncoder(w).Encode(inserted)` with the new struct.
4. Update `webhooks_test.go` to assert on the new response shape.

---

## Phase 2 — Structured errors & format transparency

**Gaps addressed:** #4 (opaque errors), #3 (detection visibility)
**Breaking changes:** None — error response bodies become richer (additive)
**Estimated scope:** Small-medium — new error types, parser instrumentation

### Task 2.1 — Add format echo to every response

When `parseWebhookPayload` returns entries, it should also return the name of
the parser that handled the request. This value is:
- Included in the 201 response (`"format": "flyio_vector"`)
- Included in error responses (`"format_detected": "native"`)
- Logged server-side for debugging

**Format names:**
| Parser | Format string |
|--------|--------------|
| Heimdall native (single) | `native` |
| Heimdall native (array) | `native_batch` |
| Vercel NDJSON | `vercel_ndjson` |
| AWS Firehose | `aws_firehose` |
| GCP Pub/Sub | `gcp_pubsub` |
| Fly.io Vector | `flyio_vector` |

**Steps:**
1. Change `parseWebhookPayload` signature to return `([]webhookLogRequest, string, error)`
   where the string is the format name.
2. Each parser branch returns its format name.
3. Thread format through to response and error helpers.
4. Add test: send Fly.io payload, assert response contains `"format": "flyio_vector"`.

### Task 2.2 — Structured validation error responses

Replace string error messages with a JSON error envelope that tells the caller
exactly what went wrong.

```go
type webhookError struct {
    Error          string              `json:"error"`           // machine-readable code
    Message        string              `json:"message"`         // human-readable summary
    FormatDetected string              `json:"format_detected"` // which parser ran
    Details        []webhookFieldError `json:"details,omitempty"`
}

type webhookFieldError struct {
    Index int    `json:"index"`          // entry index in batch (-1 for top-level)
    Field string `json:"field"`          // field name
    Error string `json:"error"`          // what's wrong
}
```

**Error codes:**
| Code | Meaning |
|------|---------|
| `invalid_json` | Request body is not valid JSON |
| `empty_payload` | No log entries after parsing |
| `validation_failed` | One or more entries failed field validation |
| `parse_failed` | Format was detected but parsing failed |
| `unauthorized` | Missing or invalid bearer token |

**Steps:**
1. Define `webhookError` and `webhookFieldError` types.
2. Create a `jsonValidationError(w, format, details)` helper.
3. Replace all `jsonError(w, "...", 400)` calls with structured equivalents.
4. In the validation pass (from Task 1.1), collect per-entry errors with indices.
5. Add tests: missing `source_type` returns `validation_failed` with index.
6. Add tests: malformed JSON returns `invalid_json`.

### Task 2.3 — Server-side logging for format detection

Add a structured log line when a request is processed, recording:
- Connection ID
- Detected format
- Entry count
- Processing time

This aids debugging when an integrator reports unexpected behaviour. Use the
existing `slog` logger — no new dependencies.

**Steps:**
1. Add `slog.Info("webhook ingestion", "connection_id", conn.ID, "format", format, "entries", len(entries), "duration_ms", elapsed)` after successful processing.
2. On parse errors, log `slog.Warn(...)` with the detected format and error.

---

## Phase 3 — Explicit format selection & detection hardening

**Gaps addressed:** #7 (no explicit format path), #3 (precedence collisions)
**Breaking changes:** None — new routes, existing route unchanged
**Estimated scope:** Medium — new routes, detection refactor

### Task 3.1 — Add format-specific ingestion routes

Register new routes that bypass auto-detection entirely:

```
POST /api/webhooks/logs              ← auto-detect (existing, unchanged)
POST /api/webhooks/logs/flyio        ← Fly.io Vector format
POST /api/webhooks/logs/vercel       ← Vercel NDJSON
POST /api/webhooks/logs/firehose     ← AWS Kinesis Firehose
POST /api/webhooks/logs/pubsub       ← GCP Pub/Sub
```

**File:** `backend/internal/api/server.go` (route registration)

**Steps:**
1. Add route registrations for each format-specific path.
2. Each route calls a shared `ingestWithFormat(w, r, format)` handler that
   skips detection and calls the specific parser directly.
3. The existing `/api/webhooks/logs` route continues to auto-detect.
4. All format-specific routes share auth, validation, transaction, and response
   logic with the base handler (extract into shared helpers).
5. Update the wizard UI: connection detail modal should show the format-specific
   URL when the connection type is known (e.g., Fly.io connections show
   `/api/webhooks/logs/flyio`).
6. Add tests for each format-specific route.

### Task 3.2 — Support `X-Heimdall-Format` header as detection override

For callers who prefer the base URL but want deterministic parsing:

```
POST /api/webhooks/logs
X-Heimdall-Format: flyio
```

**Steps:**
1. In `IngestWebhookLogs`, check for `X-Heimdall-Format` header before calling
   `parseWebhookPayload`.
2. If present, validate it against the known format list and call the specific
   parser directly.
3. If the header value is unrecognised, return 400 with a clear error listing
   valid formats.
4. Add test: send native payload with `X-Heimdall-Format: flyio`, assert it's
   parsed as Fly.io (not auto-detected as native).

### Task 3.3 — Harden auto-detection to reduce false positives

Tighten the detection predicates to be more specific:

1. **NDJSON:** Require `Content-Type: application/x-ndjson` header. Remove the
   structural `isNDJSON` fallback — it's too broad. If Vercel doesn't set the
   content type, they should use the explicit route.

2. **Firehose:** In addition to `requestId` + `records`, also require
   `records` to be a non-empty array where items have a `data` field. This
   eliminates false positives from app logs that happen to have those field names.

3. **Fly.io:** Replace `strings.HasPrefix(probe.SourceType, "fly")` with an
   exact match against known Fly.io source types (`"fly_app"`, `"fly_io"`,
   `"fly_log_shipper"`). Keep the `fly` nested object check as-is — it's
   specific enough.

**Steps:**
1. Update `isNDJSON` to only return true when Content-Type header is `ndjson`.
2. Update `isFirehosePayload` to probe for `records[0].data` structure.
3. Update `isVectorFlyObject` to use exact match set instead of prefix.
4. Add negative tests: payload with `requestId` + `records` but no `.data`
   field should NOT match Firehose.
5. Add negative test: `source_type: "flywheel"` should NOT match Fly.io.

---

## Phase 4 — Native format v2 (breaking change, with migration path)

**Gaps addressed:** #2 (payload naming confusion)
**Breaking changes:** Yes — new field names for the native format
**Estimated scope:** Medium — schema change with backward compatibility layer

### Design: new native format

```json
{
  "source": "my-service",
  "level": "error",
  "message": "connection refused to db-primary",
  "attrs": {
    "host": "web-1",
    "request_id": "req-abc123",
    "duration_ms": 1250
  }
}
```

| Field | Required | Replaces | Rationale |
|-------|----------|----------|-----------|
| `source` | Yes | `source_type` | Shorter, more natural — what every dev expects |
| `level` | No | `severity` | Standard name across logging frameworks (syslog, structured logging, OpenTelemetry) |
| `message` | No | (buried in `payload`) | First-class field — the thing every developer searches for |
| `attrs` | No | `payload` | "Attributes" — no collision with "payload" as request body concept. Matches OpenTelemetry `Attributes` naming |

**Batch:**
```json
[
  { "source": "api", "level": "info", "message": "GET /health 200", "attrs": {} },
  { "source": "api", "level": "error", "message": "timeout", "attrs": { "upstream": "db" } }
]
```

### Task 4.1 — Define v2 request struct and dual-format parsing

**File:** `webhook_parsers.go`

```go
type webhookLogRequestV2 struct {
    Source  string          `json:"source"`
    Level   string          `json:"level"`
    Message string          `json:"message"`
    Attrs   json.RawMessage `json:"attrs"`
}
```

The native parser should accept **both** v1 and v2 field names. Detection logic:
- If the parsed object has `source` (v2 field), treat as v2.
- If it has `source_type` (v1 field), treat as v1.
- If both are present, prefer v2.

Internally, both are normalised to the same `webhookLogRequest` before insertion.
For v2, the stored `payload` in the database is constructed by merging `message`
into `attrs`:

```go
// v2 → internal normalisation
attrs["message"] = entry.Message  // promote message into the stored payload
entry.Payload = marshal(attrs)
entry.SourceType = entry.Source
entry.Severity = normalizeSeverity(entry.Level)
```

**Steps:**
1. Define `webhookLogRequestV2` struct.
2. In `parseNativePayload`, attempt v2 parsing first (check for `source` field).
3. Fall back to v1 parsing if `source` is absent.
4. Normalise v2 → v1 internal representation for the insert path.
5. Return format as `"native_v2"` or `"native"` accordingly.
6. Add tests for v2 single entry, v2 batch, mixed v1+v2 batch (should reject
   with clear error — don't mix versions in one batch).

### Task 4.2 — Update wizard UI and documentation

Update all user-facing surfaces to show v2 format as the primary/recommended
format:

1. **Connection detail modal:** Example payloads use v2 field names.
2. **Fly.io drain wizard:** The "Expected Payload Format" reference section
   shows both the auto-detected Fly format and the v2 native format.
3. **SDK examples:** If any SDK code references the webhook format, update to v2.

v1 format should still be shown as "Legacy format (still supported)" where
relevant, but v2 is the default in all new documentation.

**Steps:**
1. Update `StepFlyioDrainSetup.vue` example payloads.
2. Update `ConnectionDetailModal.vue` example payloads.
3. Grep for any hardcoded example payloads elsewhere and update.

### Task 4.3 — Deprecation headers for v1 native format

When a request is parsed using v1 native format, include a deprecation header
in the response:

```
Sunset: 2026-10-01
Deprecation: true
Link: <https://docs.heimdallwatch.com/api/webhook-format-v2>; rel="successor-version"
```

This follows the IETF Sunset Header RFC (RFC 8594) and gives integrators a
clear signal to migrate without breaking them.

**Steps:**
1. When format is `"native"`, add `Sunset` and `Deprecation` headers.
2. Include `"deprecated": true` in the response body.
3. Log a warning server-side for v1 usage to track migration progress.

---

## Phase 5 — Idempotency

**Gaps addressed:** #6 (no dedup mechanism)
**Breaking changes:** None — opt-in via header
**Estimated scope:** Medium — new table, cache logic

### Task 5.1 — Idempotency key support

Accept an optional `X-Idempotency-Key` header. When present:

1. Before processing, check if the key has been seen in the last 24 hours.
2. If seen, return the cached response (same status code and body) without
   re-inserting.
3. If not seen, process normally and cache the response keyed by
   `(connection_id, idempotency_key)`.

**Storage:** A new `webhook_idempotency` table:

```sql
CREATE TABLE webhook_idempotency (
    connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    idempotency_key TEXT NOT NULL,
    response_status INT NOT NULL,
    response_body JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (connection_id, idempotency_key)
);

-- Auto-expire after 24 hours
CREATE INDEX idx_webhook_idempotency_expiry ON webhook_idempotency (created_at);
```

**Steps:**
1. Create migration for `webhook_idempotency` table.
2. Add sqlc queries: `GetIdempotencyResult`, `InsertIdempotencyResult`.
3. In the handler, check for `X-Idempotency-Key` header early.
4. If key exists and is cached, return cached response immediately.
5. If key exists and is not cached, process and cache before responding.
6. If no key, process normally (current behaviour).
7. Add a periodic cleanup job or rely on `pg_cron` / TTL index to expire old keys.
8. Add tests: same key twice returns identical response, second call inserts 0 rows.

---

## Phase summary

| Phase | Gaps fixed | Breaking? | Dependency |
|-------|-----------|-----------|------------|
| 0 — Fly.io wizard fix | #8 | None (UI text) | None |
| 1 — Correctness & safety | #1, #5 | Response shape only | None |
| 2 — Structured errors | #4, #3 (visibility) | None | None |
| 3 — Explicit format routes | #7, #3 (hardening) | None | Phase 2 (format names) |
| 4 — Native format v2 | #2 | Yes (with compat) | Phase 2 (error structs) |
| 5 — Idempotency | #6 | None | Phase 1 (transactions) |

Phase 0 has no dependencies and can ship immediately. Phases 1 and 2 can be
implemented in parallel. Phase 3 depends on format name strings from Phase 2.
Phase 4 depends on error structs from Phase 2. Phase 5 depends on transactional
inserts from Phase 1.

```
Phase 0 (ship now, no deps)

Phase 1 ──────────────────► Phase 5
    │
    ├── Phase 2 ──► Phase 3
    │       │
    │       └────► Phase 4
    │
    (1 & 2 can run in parallel)
```
