CREATE INDEX idx_connections_webhook_token
    ON connections ((config->>'webhook_token'))
    WHERE type = 'webhook_logs';
