-- Phase 9, Step 2: Notification preferences (per-application, 1:1)
-- Master switch, severity threshold, and cooldown window for notification dispatch.

CREATE TABLE notification_preferences (
    app_id              UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    enabled             BOOLEAN NOT NULL DEFAULT false,   -- master switch
    severity_threshold  TEXT NOT NULL DEFAULT 'warning',  -- minimum: info, warning, error, critical
    cooldown_minutes    INTEGER NOT NULL DEFAULT 15,      -- suppress duplicates for N minutes
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
