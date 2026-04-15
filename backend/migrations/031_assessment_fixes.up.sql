-- C1: Webhook token index misses OTLP type
-- The partial index predicate only covered 'webhook_logs', but
-- GetConnectionByWebhookToken filters on type IN ('webhook_logs', 'otlp').
-- OTLP connections fell through to a sequential scan.
DROP INDEX IF EXISTS idx_connections_webhook_token;
CREATE INDEX idx_connections_webhook_token
    ON connections ((config->>'webhook_token'))
    WHERE type IN ('webhook_logs', 'otlp');

-- M9: Duplicate unique index on organizations.slug
-- Column definition already has UNIQUE constraint (implicit index).
-- The explicit index is redundant — double storage, double write overhead.
DROP INDEX IF EXISTS idx_organizations_slug;
