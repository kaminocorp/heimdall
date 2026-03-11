-- Phase 9, Step 3: Notification log (delivery tracking)
-- Records every notification attempt for observability, dedup, and debugging.

CREATE TABLE notification_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id          UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    channel_id      UUID NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    agent_log_id    UUID REFERENCES agent_log(id) ON DELETE SET NULL,
    severity        TEXT NOT NULL,
    summary         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, sent, failed
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_log_app_created ON notification_log(app_id, created_at DESC);
CREATE INDEX idx_notification_log_status ON notification_log(status);
