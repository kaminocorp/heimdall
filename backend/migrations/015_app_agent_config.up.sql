-- Phase 8, Step 1: Per-application agent configuration
-- Replaces the global singleton agent_config table.

CREATE TABLE app_agent_config (
    app_id                 UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    model                  TEXT NOT NULL DEFAULT 'claude-sonnet-4-6',
    mode                   TEXT NOT NULL DEFAULT 'continuous',
    schedule_interval_secs INTEGER NOT NULL DEFAULT 60,
    system_prompt_override TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE app_agent_config ENABLE ROW LEVEL SECURITY;

-- RLS: users can access configs for apps in their org
CREATE POLICY app_agent_config_org ON app_agent_config
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );
