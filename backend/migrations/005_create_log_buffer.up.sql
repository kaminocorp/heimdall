CREATE TABLE log_buffer (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id   UUID NOT NULL REFERENCES connections(id),
    source_type     TEXT NOT NULL,
    severity        TEXT,
    payload         JSONB NOT NULL,
    ingested_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_log_buffer_ingested_at ON log_buffer (ingested_at DESC);
CREATE INDEX idx_log_buffer_connection ON log_buffer (connection_id, ingested_at DESC);
CREATE INDEX idx_log_buffer_severity ON log_buffer (severity, ingested_at DESC) WHERE severity IS NOT NULL;
