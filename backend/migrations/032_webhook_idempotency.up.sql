-- Idempotency cache for webhook ingestion (Phase 5).
-- Allows callers to send an X-Idempotency-Key header to prevent duplicate
-- inserts on retry. Keyed by (connection_id, idempotency_key) with a 24-hour
-- TTL enforced by application-level cleanup.

CREATE TABLE webhook_idempotency (
    connection_id   UUID        NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    idempotency_key TEXT        NOT NULL,
    response_status INT         NOT NULL,
    response_body   JSONB       NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (connection_id, idempotency_key)
);

-- Index for efficient expiry cleanup (DELETE WHERE created_at < now() - interval '24 hours').
CREATE INDEX idx_webhook_idempotency_expiry ON webhook_idempotency (created_at);
