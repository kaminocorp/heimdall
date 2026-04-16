# Webhook Ingestion — Phase 5: Idempotency

**Status:** Complete
**Date:** 2026-04-16
**Assessment:** `docs/executing/webhook-ingestion-assessment.md`
**Plan:** `docs/executing/webhook-ingestion-improvements-plan.md`
**Gaps addressed:** #6 (no dedup mechanism)

---

## Summary

Added opt-in idempotency support via an `X-Idempotency-Key` header.
When present, the response is cached for 24 hours. Subsequent requests
with the same key and connection return the cached response without
re-inserting log entries. Callers that don't send the header get the
same behavior as before — this is purely additive.

---

## Task 5.1a — Migration

**File:** `backend/migrations/032_webhook_idempotency.up.sql`

```sql
CREATE TABLE webhook_idempotency (
    connection_id   UUID        NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    idempotency_key TEXT        NOT NULL,
    response_status INT         NOT NULL,
    response_body   JSONB       NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (connection_id, idempotency_key)
);

CREATE INDEX idx_webhook_idempotency_expiry ON webhook_idempotency (created_at);
```

### Design decisions

- **Composite primary key `(connection_id, idempotency_key)`** rather than
  a global unique key. Two different connections can safely reuse the same
  idempotency key without collision. This matches the real-world pattern
  where each log shipper generates its own keys.

- **`ON DELETE CASCADE`** from `connections` — when a connection is deleted,
  its cached responses are cleaned up automatically.

- **Expiry index on `created_at`** enables efficient cleanup queries.
  Application-level pruning via `PruneExpiredIdempotencyKeys` (deletes
  rows older than 24 hours) rather than `pg_cron` — keeps infrastructure
  simple and avoids a Supabase extension dependency.

- **`response_body` as JSONB** rather than TEXT — allows future queries
  against cached response fields if needed, and JSONB is already the
  pattern used throughout the schema.

---

## Task 5.1b — sqlc queries

**File:** `backend/internal/db/queries/webhook_idempotency.sql`

| Query | Type | Purpose |
|-------|------|---------|
| `GetIdempotencyResult` | `:one` | Look up cached response by `(connection_id, idempotency_key)` within 24h window |
| `InsertIdempotencyResult` | `:exec` | Cache a response. Uses `ON CONFLICT DO NOTHING` to handle race conditions |
| `PruneExpiredIdempotencyKeys` | `:execrows` | Delete entries older than 24 hours (for periodic cleanup) |

The `ON CONFLICT DO NOTHING` on insert handles the edge case where two
concurrent requests with the same key both pass the cache-miss check —
the first insert wins, the second is silently dropped. Both requests
will have processed the entries (since they both missed the cache), but
the transaction ensures no partial state.

---

## Task 5.1c — Handler logic

**File:** `backend/internal/api/handlers/webhooks.go`

The idempotency check is added at the top of `ingestEntries`, before
validation or any database writes:

```
1. Read X-Idempotency-Key header
2. If present, query GetIdempotencyResult
   → Cache hit:  write cached status + body, set X-Idempotency-Replay: true, return
   → Cache miss: continue to normal processing
3. [Normal flow: validate, transaction, insert, commit]
4. If idempotency key was present, cache the response via InsertIdempotencyResult
5. Write response
```

### Key design choices

- **Check happens in `ingestEntries`**, not in each handler method. This
  ensures idempotency works for both auto-detect and format-specific
  routes without duplicating code.

- **`X-Idempotency-Replay: true` header** on cached responses lets
  callers distinguish between "your data was just ingested" and "your
  data was already ingested by a previous request." This follows the
  pattern used by Stripe and other APIs.

- **Only success responses (201) are cached.** If parsing or validation
  fails, the error is returned normally without caching — the caller
  should fix the payload and retry, not get a cached error forever.

- **Fire-and-forget cache insert.** The `InsertIdempotencyResult` error
  is discarded (`_ = s.Queries.InsertIdempotencyResult(...)`). If the
  cache write fails (e.g., connection pool exhausted), the log entries
  were already committed — the worst case is that a retry will insert
  duplicates, which is the same behavior as without idempotency.

---

## Test changes

**File:** `backend/internal/api/handlers/webhook_parsers_test.go`

| Test | What it verifies |
|------|-----------------|
| `TestWebhookResponseSerialization_ForCaching` | Response body round-trips through JSON marshal/unmarshal; `deprecated` omitted when false |
| `TestWebhookResponseSerialization_DeprecatedCaching` | v1 deprecated response round-trips with `deprecated: true` |

The full idempotency integration test (send same key twice, assert second
returns cached response with zero new rows) requires a running database
and is not included in the unit test suite. The existing
`TestIngestWebhookLogs` integration test verifies the handler still works
end-to-end; idempotency is opt-in and doesn't affect requests without
the header.

All 44 tests pass. `go vet` clean.

---

## Cleanup considerations

The `PruneExpiredIdempotencyKeys` query is generated but not yet wired
into a periodic job. Options for wiring:

1. **Add to the monitoring loop** — the background goroutine that runs
   every 15 seconds could call `PruneExpiredIdempotencyKeys` once per
   hour (check a timestamp, skip if <1h since last prune).
2. **Startup cleanup** — prune on server start, then periodically.
3. **API endpoint** — expose a `/api/admin/prune` endpoint for manual
   or cron-based cleanup.

The table will grow slowly (one row per idempotent request) and the
24-hour TTL in the `GetIdempotencyResult` query means stale rows are
never returned even if not pruned. Pruning is a space optimization,
not a correctness requirement.

---

## Files changed

| File | Change |
|------|--------|
| `migrations/032_webhook_idempotency.up.sql` | New table with composite PK and expiry index |
| `migrations/032_webhook_idempotency.down.sql` | Drop table and index |
| `internal/db/queries/webhook_idempotency.sql` | 3 sqlc queries |
| `internal/db/webhook_idempotency.sql.go` | Generated Go code (sqlc) |
| `internal/db/models.go` | Generated `WebhookIdempotency` model (sqlc) |
| `webhooks.go` | Idempotency check + cache in `ingestEntries` |
| `webhook_parsers_test.go` | 2 new response serialization tests |
