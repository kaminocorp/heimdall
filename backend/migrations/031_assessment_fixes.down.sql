-- Reverse M9: re-create the explicit duplicate index
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);

-- Reverse C1: restore the original partial index (webhook_logs only)
DROP INDEX IF EXISTS idx_connections_webhook_token;
CREATE INDEX idx_connections_webhook_token
    ON connections ((config->>'webhook_token'))
    WHERE type = 'webhook_logs';
