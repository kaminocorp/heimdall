CREATE TABLE investigations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_type    TEXT NOT NULL,
    trigger_source  TEXT,
    summary         TEXT NOT NULL,
    severity        TEXT NOT NULL DEFAULT 'info',
    status          TEXT NOT NULL DEFAULT 'open',
    context         JSONB NOT NULL,
    findings        JSONB,
    tool_trace      JSONB,
    resolution      TEXT,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ
);

CREATE INDEX idx_investigations_status ON investigations (status);
CREATE INDEX idx_investigations_severity ON investigations (severity);
CREATE INDEX idx_investigations_started_at ON investigations (started_at DESC);
