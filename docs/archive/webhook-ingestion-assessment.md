# Webhook Ingestion — Architecture Assessment

**Status:** Assessment complete, implementation plan in `webhook-ingestion-improvements-plan.md`
**Date:** 2026-04-16
**Scope:** `POST /api/webhooks/logs` endpoint — handler, parsers, response contract

---

## Context

The webhook ingestion endpoint is Heimdall's primary data intake surface. Every
log that enters the system from external infrastructure (Fly.io, Vercel, AWS,
GCP, custom apps) passes through this single endpoint. Its design directly
determines how easy it is for teams to integrate with Heimdall.

This assessment was triggered by the Trajan team asking "what's the exact JSON
payload structure the webhook expects?" — a question that should have been
trivially answerable from the API contract, but wasn't.

**Files under review:**
- `backend/internal/api/handlers/webhooks.go` — main handler
- `backend/internal/api/handlers/webhook_parsers.go` — format detection + parsing

---

## What works well

### Auto-detection of multiple formats
The `parseWebhookPayload` function detects and normalises five external formats
(Vercel NDJSON, AWS Firehose, GCP Pub/Sub, Fly.io Vector, Heimdall native)
through a single endpoint. This is the right instinct — "just point your log
shipper at us and it works" — and meaningfully reduces onboarding friction.

### Severity normalisation
The `normalizeSeverity` function maps the full range of common log level strings
(`err`, `crit`, `fatal`, `warn`, `notice`, `trace`, etc.) to five canonical
values. This means integrators don't need to pre-map their levels.

### Single + batch in the same endpoint
Both `{}` and `[{}, {}]` are accepted without any configuration toggle. This
removes a common integration stumbling block.

### Bearer token auth
Simple, stateless, widely understood. No HMAC signing, no API key rotation
ceremony — appropriate for the current scale.

---

## Gap 1 — Partial failure silently corrupts batch semantics

**Severity:** Critical (correctness)
**Location:** `webhooks.go:57-86`

The handler iterates over parsed entries, inserting each one individually. If
entry 6 of 10 fails validation (e.g., missing `source_type`), the handler
returns HTTP 400 — but entries 1–5 are already committed to the database. The
caller receives an error and has no way to know which entries were persisted.

Consequences:
- **Silent duplicates on retry.** The caller sees 400, retries the full batch,
  and entries 1–5 are inserted again.
- **Silent data loss.** If the caller treats 400 as terminal, entries 7–10 are
  never ingested.
- **Impossible to debug.** Neither the caller nor Heimdall's own logs indicate
  which entries made it.

```go
// Current: no transaction, no partial-success reporting
for _, e := range entries {
    // validate...
    row, err := s.Queries.InsertLogEntry(r.Context(), ...)
    if err != nil {
        jsonServerError(w, "failed to insert log entry", err)
        return  // entries already inserted are committed
    }
    inserted = append(inserted, row)
}
```

---

## Gap 2 — The `payload` field name creates persistent DX confusion

**Severity:** High (developer experience)
**Location:** `webhooks.go:13-17` — `webhookLogRequest` struct

The native format requires a field called `payload` inside the request body.
Every developer's mental model: "the payload is what I send in the request
body." This creates a recursive naming problem — the payload contains a field
called payload — that confuses every new integrator.

This is exactly what happened with Trajan: they asked whether the body should
be wrapped in a `"payload"` key, because the word means "the outer envelope"
to them, not "the inner data blob."

Additionally, the most important field in any log entry — the human-readable
message — is buried inside the arbitrary `payload` JSON blob with no
first-class status. Developers have to decide for themselves where to put it.

---

## Gap 3 — Detection order creates silent precedence collisions

**Severity:** High (correctness)
**Location:** `webhook_parsers.go:15-49`

The auto-detection chain runs in fixed order:
```
NDJSON → Firehose → Pub/Sub → Fly.io → Native
```

This creates collision risks:

1. **NDJSON vs Fly.io:** A batch of Fly.io entries sent as newline-delimited
   JSON (which Vector can do) hits the Vercel NDJSON parser first, because
   `isNDJSON` matches any multi-line input where each line starts with `{`.

2. **Firehose false positive:** Any native payload that happens to contain
   `requestId` + `records` fields gets misrouted to the Firehose parser. These
   are common field names in application logs.

3. **Fly.io over-matching:** `strings.HasPrefix(probe.SourceType, "fly")`
   (line 353) means a source type like `"flywheel"` or `"flutter"` would NOT
   match, but `"fly_anything"` would. More critically, any payload with a
   top-level `"fly"` key triggers Fly.io parsing — a plausible field name in
   aviation, travel, or deployment contexts.

The caller has **zero visibility** into which parser handled their request.
Misrouted payloads are silently parsed with the wrong schema, producing
corrupted log entries that are difficult to diagnose.

---

## Gap 4 — Error messages are opaque

**Severity:** High (developer experience)
**Location:** `webhooks.go:48-50`

```go
jsonError(w, "invalid request body", http.StatusBadRequest)
```

When a payload fails parsing, the caller receives a generic string with no
indication of:
- Which format was attempted
- Which field caused the failure
- What the expected structure is
- Which entry in a batch failed (if applicable)

For a batch of 50 entries where entry 23 has a typo in `source_type`, the
error gives no way to locate the problem.

---

## Gap 5 — Response leaks internal identifiers

**Severity:** Medium (security hygiene)
**Location:** `webhooks.go:88-90`

The 201 response returns full `LogBuffer` database rows including `user_id`,
`app_id`, and `connection_id` — internal UUIDs that are meaningless to the
caller and should not be exposed to external systems. A webhook caller is an
external integration, not an authenticated UI user.

---

## Gap 6 — No idempotency mechanism

**Severity:** Medium (production resilience)
**Location:** `webhooks.go` — no dedup logic present

There is no idempotency key or deduplication mechanism. Network timeouts,
load balancer retries, or client-side retry logic will produce duplicate log
entries. For a monitoring system, duplicates can:
- Inflate error counts and trigger false alerts
- Pollute the activity feed with repeated entries
- Skew the Lumber classifier's pattern recognition

---

## Gap 7 — No explicit format selection path

**Severity:** Medium (developer experience)
**Location:** Architecture-level — single endpoint, detection-only

Auto-detection is the only way to indicate format. There is no mechanism for
a caller to explicitly declare "this is Fly.io format" and bypass detection.
This means:
- Integrators can't be certain which parser will handle their data
- Format collisions (Gap 3) have no workaround
- Debugging requires reading Heimdall's source code to understand detection logic

---

## Gap 8 — Fly.io wizard assumes enriched payload that NATS source doesn't produce

**Severity:** High (developer experience / correctness)
**Location:** `frontend/src/components/connections/wizard/steps/StepFlyioDrainSetup.vue:106-109`
**Reported by:** Trajan team (2026-04-16, post-integration feedback)

The wizard's Step 4 ("Configure Vector") tells users:

> "If you're using the stock Log Shipper image, the `fly` metadata object is
> included by default — no extra config needed."

This is **incorrect**. The Fly Log Shipper consumes Fly's internal NATS log
stream, which emits flat, unstructured log lines. These lines do **not** include
the nested `fly.app.name`, `fly.machine.id`, `fly.region` metadata that
Heimdall's `isVectorFlyObject` detector looks for (`webhook_parsers.go:343-354`).

The enriched `fly` metadata object is only present when using Fly's **HTTP log
drain API** — a different integration path that the Log Shipper does not use.

Consequences (confirmed by Trajan's integration experience):
- **Auto-detection silently fails.** The NATS-sourced payload has no `fly`
  object and no `source_type` starting with `"fly"`, so `isVectorFlyPayload`
  returns false. The payload falls through to the native parser.
- **Native parser rejects it.** The flat NATS log line doesn't have
  `source_type` or `payload` fields, so the native parser returns 400.
- **Users are forced to write a Vector transform anyway.** The wizard presents
  the transform as a "custom config" afterthought, when in practice it's the
  **required path** for most Fly Log Shipper deployments.
- **The "Expected Payload Format" reference section** (lines 136-158) lists
  fields (`fly.app.name`, `fly.region`, etc.) that the NATS source never emits,
  reinforcing the incorrect assumption.

The wizard should:
1. Make the Vector transform the **primary/default** guidance, not a fallback
   for "custom configs."
2. Clarify that the stock Fly Log Shipper image does not produce the enriched
   format.
3. Provide a complete, ready-to-use VRL transform that reshapes NATS events
   into Heimdall's native format (with `source_type`, `severity`, `payload`).
4. Move the enriched-format reference into a collapsible section labelled as
   applicable only to HTTP log drain integrations.

---

## Summary matrix

| # | Gap | Severity | Category | Breaking change? |
|---|-----|----------|----------|-----------------|
| 1 | Partial failure corrupts batches | Critical | Correctness | No |
| 2 | `payload` field name confusion | High | DX | Yes (native format) |
| 3 | Detection precedence collisions | High | Correctness | No |
| 4 | Opaque error messages | High | DX | No (additive) |
| 5 | Response leaks internal IDs | Medium | Security | Yes (response shape) |
| 6 | No idempotency mechanism | Medium | Resilience | No (additive) |
| 7 | No explicit format selection | Medium | DX | No (additive) |
| 8 | Wizard assumes enriched Fly.io payload NATS doesn't produce | High | DX / Correctness | No (UI text) |
