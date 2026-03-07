-- Phase 8, Step 2: Per-application monitoring cursor
-- Tracks where the monitor left off so we only process new logs.

CREATE TABLE monitoring_state (
    app_id             UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    last_monitored_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
