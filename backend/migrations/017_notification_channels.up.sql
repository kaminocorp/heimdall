-- Phase 9, Step 1: Notification channels (per-application)
-- Each app can have multiple notification channels (email, Slack, Discord),
-- each with its own type-specific config stored as JSONB.

CREATE TABLE notification_channels (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id     UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,          -- 'email', 'slack', 'discord'
    name       TEXT NOT NULL,          -- user-friendly label, e.g. "Ops Slack"
    config     JSONB NOT NULL,         -- channel-specific config (recipients, webhook_url, etc.)
    enabled    BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_channels_app_id ON notification_channels(app_id);
