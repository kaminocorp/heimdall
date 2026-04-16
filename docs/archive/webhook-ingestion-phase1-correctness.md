# Webhook Ingestion — Phase 1: Correctness & Safety

**Status:** Complete
**Date:** 2026-04-16
**Assessment:** `docs/executing/webhook-ingestion-assessment.md`
**Plan:** `docs/executing/webhook-ingestion-improvements-plan.md`
**Gaps addressed:** #1 (partial failure corrupts batches), #5 (response leaks internal IDs)

---

## Summary

Two changes to the `POST /api/webhooks/logs` handler that make batch
ingestion atomic and stop exposing internal database identifiers to
external webhook callers.

---

## Task 1.1 — Transactional batch inserts with upfront validation

**File:** `backend/internal/api/handlers/webhooks.go`

### Problem

The handler iterated over parsed entries, inserting each one individually
without a transaction. If entry N failed validation, entries 1–(N-1) were
already committed. The caller received a 400 but had no way to know which
entries were persisted, leading to silent duplicates on retry or silent
data loss if the caller treated 400 as terminal.

### What changed

1. **Upfront validation pass.** All entries are checked for `source_type`
   and `payload` before any database work begins. If any entry fails, the
   handler returns 400 immediately with the entry index in the error
   message (e.g. `"entry 3: source_type is required"`). Nothing is
   written to the database.

2. **Transactional insert.** The insert loop now runs inside a
   `pgxpool.Pool.Begin` / `Commit` transaction using `Queries.WithTx(tx)`.
   If any insert fails, the deferred `tx.Rollback` ensures nothing is
   persisted. All entries succeed or none do.

### Why this approach

- **Validate-then-insert** (two passes) rather than validate-and-insert
  (single pass with rollback on validation failure) because it avoids
  starting a transaction at all when the input is bad. This is cheaper for
  the common "malformed payload" case and gives clearer error semantics.

- **Single transaction** rather than per-entry transactions because the
  webhook contract is "this batch is one unit of work." Partial success
  with no reporting mechanism is worse than all-or-nothing.

---

## Task 1.2 — Minimal response (no internal IDs)

**Files:** `backend/internal/api/handlers/webhooks.go`, `backend/internal/api/handlers/webhook_parsers.go`

### Problem

The 201 response returned full `LogBuffer` database rows including
`user_id`, `app_id`, and `connection_id` — internal UUIDs that are
meaningless to webhook callers and should not be exposed to external
systems.

### What changed

1. **New response struct.** Replaced the raw `[]db.LogBuffer` response
   with a `webhookLogResponse` struct containing only two fields:

   ```json
   { "accepted": 5, "format": "native_batch" }
   ```

   - `accepted` — the number of log entries successfully stored.
   - `format` — which parser handled the request, providing transparency
     into auto-detection (lays groundwork for Phase 2 format echo).

2. **Format propagation from parsers.** `parseWebhookPayload` now returns
   a `([]webhookLogRequest, string, error)` triple, where the string is
   the detected format name. Each parser branch returns its own format
   identifier:

   | Parser | Format string |
   |--------|--------------|
   | Native (single) | `native` |
   | Native (array) | `native_batch` |
   | Vercel NDJSON | `vercel_ndjson` |
   | AWS Firehose | `aws_firehose` |
   | GCP Pub/Sub | `gcp_pubsub` |
   | Fly.io Vector | `flyio_vector` |

### Why this approach

- **Minimal surface area.** A webhook caller is an external integration,
  not an authenticated UI user. It needs confirmation ("your data was
  accepted"), not internal state.

- **Format echo** is useful for debugging integration issues — if a
  caller's Fly.io payload is silently handled by the NDJSON parser, the
  `format` field makes this visible. This directly addresses a sub-aspect
  of Gap #3 (detection collisions) and sets up Phase 2's structured error
  responses.

---

## Test changes

**File:** `backend/internal/api/handlers/webhook_parsers_test.go`

All 17 parser unit tests updated for the new 3-return-value signature of
`parseWebhookPayload`. Two tests (`TestParseNativePayload_Single` and
`TestParseNativePayload_Batch`) now also assert on the returned format
string (`"native"` and `"native_batch"` respectively).

**File:** `backend/internal/api/handlers/webhooks_test.go`

The `TestIngestWebhookLogs` integration test updated to assert on the new
response shape (`accepted` and `format` fields) instead of the old
`id` field from the raw `LogBuffer` row.

---

## Breaking change note

The 201 response body shape has changed. Previously it returned a
`LogBuffer` object (single entry) or array (batch) with fields like
`id`, `user_id`, `app_id`, `connection_id`. Now it returns:

```json
{ "accepted": 1, "format": "native" }
```

This is technically a breaking change for any caller that parses the
response body. However, webhook callers are external log shippers — they
typically check the HTTP status code (201 vs 4xx) and ignore the body.
No known integrations depend on the response fields.
