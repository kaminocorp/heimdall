-- Enable RLS on tables that were missing it, consistent with the pattern
-- established for all other user-scoped tables (connections, log_buffer, etc.).
-- Uses DROP IF EXISTS before CREATE to make the migration idempotent.

-- organizations
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS organizations_user_policy ON organizations;
CREATE POLICY organizations_user_policy ON organizations
    FOR ALL USING (
        id IN (SELECT org_id FROM public.users WHERE id = app_current_user_id())
    );

-- applications
ALTER TABLE applications ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS applications_user_policy ON applications;
CREATE POLICY applications_user_policy ON applications
    FOR ALL USING (
        org_id IN (SELECT org_id FROM public.users WHERE id = app_current_user_id())
    );

-- monitoring_state
ALTER TABLE monitoring_state ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS monitoring_state_user_policy ON monitoring_state;
CREATE POLICY monitoring_state_user_policy ON monitoring_state
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- notification_channels
ALTER TABLE notification_channels ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS notification_channels_user_policy ON notification_channels;
CREATE POLICY notification_channels_user_policy ON notification_channels
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- notification_preferences
ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS notification_preferences_user_policy ON notification_preferences;
CREATE POLICY notification_preferences_user_policy ON notification_preferences
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- notification_log — use direct app_id column instead of joining through channel_id
-- for better performance and resilience to orphaned channel records.
ALTER TABLE notification_log ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS notification_log_user_policy ON notification_log;
CREATE POLICY notification_log_user_policy ON notification_log
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- Missing FK indexes for cascade performance.
CREATE INDEX IF NOT EXISTS idx_notification_log_channel_id ON notification_log(channel_id);
CREATE INDEX IF NOT EXISTS idx_agent_log_conversation_id ON agent_log(conversation_id);
