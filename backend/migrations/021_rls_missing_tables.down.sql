-- Reverse RLS policies and indexes added in 021_rls_missing_tables.up.sql

DROP INDEX IF EXISTS idx_agent_log_conversation_id;
DROP INDEX IF EXISTS idx_notification_log_channel_id;

DROP POLICY IF EXISTS notification_log_user_policy ON notification_log;
ALTER TABLE notification_log DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS notification_preferences_user_policy ON notification_preferences;
ALTER TABLE notification_preferences DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS notification_channels_user_policy ON notification_channels;
ALTER TABLE notification_channels DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS monitoring_state_user_policy ON monitoring_state;
ALTER TABLE monitoring_state DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS applications_user_policy ON applications;
ALTER TABLE applications DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS organizations_user_policy ON organizations;
ALTER TABLE organizations DISABLE ROW LEVEL SECURITY;
