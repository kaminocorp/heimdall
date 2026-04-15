-- Reverse 026_org_members: restore users.org_id and revert RLS policies.
--
-- WARNING: This rollback is lossy for multi-org users. Each user is assigned
-- their earliest org membership (by created_at). All other memberships are
-- permanently discarded. Only run this if you are certain no user has
-- meaningful data in secondary organisations, or after exporting the
-- org_members table for manual recovery.

-- Guard: abort if any user belongs to more than one org.
DO $$
BEGIN
    IF EXISTS (
        SELECT user_id FROM org_members GROUP BY user_id HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'Cannot rollback: one or more users have multi-org memberships. '
            'Export org_members before proceeding, or remove secondary memberships manually.';
    END IF;
END $$;

-- Step 1: Re-add the org_id column to users
ALTER TABLE users ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE SET NULL;
CREATE INDEX idx_users_org_id ON users(org_id);

-- Step 2: Migrate data back — pick the first org membership for each user
UPDATE users u
SET org_id = om.org_id
FROM (
    SELECT DISTINCT ON (user_id) user_id, org_id
    FROM org_members
    ORDER BY user_id, created_at ASC
) om
WHERE u.id = om.user_id;

-- Step 3: Revert RLS policies to the old users.org_id pattern

-- 3a. organizations
DROP POLICY IF EXISTS organizations_user_policy ON organizations;
CREATE POLICY organizations_user_policy ON organizations
    FOR ALL USING (
        id IN (SELECT org_id FROM public.users WHERE id = app_current_user_id())
    );

-- 3b. applications
DROP POLICY IF EXISTS applications_user_policy ON applications;
CREATE POLICY applications_user_policy ON applications
    FOR ALL USING (
        org_id IN (SELECT org_id FROM public.users WHERE id = app_current_user_id())
    );

-- 3c. app_agent_config
DROP POLICY IF EXISTS app_agent_config_org ON app_agent_config;
CREATE POLICY app_agent_config_org ON app_agent_config
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- 3d. monitoring_state
DROP POLICY IF EXISTS monitoring_state_user_policy ON monitoring_state;
CREATE POLICY monitoring_state_user_policy ON monitoring_state
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- 3e. notification_channels
DROP POLICY IF EXISTS notification_channels_user_policy ON notification_channels;
CREATE POLICY notification_channels_user_policy ON notification_channels
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- 3f. notification_preferences
DROP POLICY IF EXISTS notification_preferences_user_policy ON notification_preferences;
CREATE POLICY notification_preferences_user_policy ON notification_preferences
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- 3g. notification_log
DROP POLICY IF EXISTS notification_log_user_policy ON notification_log;
CREATE POLICY notification_log_user_policy ON notification_log
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- 3h. investigation_schedules
DROP POLICY IF EXISTS investigation_schedules_org ON investigation_schedules;
CREATE POLICY investigation_schedules_org ON investigation_schedules
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );

-- Step 4: Drop org_members table and enum
DROP TABLE IF EXISTS org_members;
DROP TYPE IF EXISTS org_member_role;
