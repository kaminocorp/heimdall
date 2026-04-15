-- Org Members: replaces the 1:1 users.org_id column with a proper M:N
-- membership table (org_members) supporting roles.
--
-- Migration strategy:
--   1. Create the role enum and org_members table
--   2. Migrate existing users.org_id data into org_members rows (as 'owner')
--   3. Rewrite all RLS policies that join via users.org_id to join via org_members
--   4. Drop users.org_id column and its index

-- Step 1: Role enum and org_members table
CREATE TYPE org_member_role AS ENUM ('owner', 'admin', 'member');

CREATE TABLE org_members (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role       org_member_role NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, org_id)
);

CREATE INDEX idx_org_members_org_id ON org_members(org_id);
CREATE INDEX idx_org_members_user_id ON org_members(user_id);

-- Step 2: Migrate existing data — every user with an org becomes 'owner'
INSERT INTO org_members (user_id, org_id, role)
SELECT id, org_id, 'owner'
FROM users
WHERE org_id IS NOT NULL;

-- Step 3: Rewrite RLS policies
-- All existing policies that reference users.org_id are dropped and recreated
-- to join via org_members instead.

-- 3a. organizations (from 021)
DROP POLICY IF EXISTS organizations_user_policy ON organizations;
CREATE POLICY organizations_user_policy ON organizations
    FOR ALL USING (
        id IN (SELECT org_id FROM org_members WHERE user_id = app_current_user_id())
    );

-- 3b. applications (from 021)
DROP POLICY IF EXISTS applications_user_policy ON applications;
CREATE POLICY applications_user_policy ON applications
    FOR ALL USING (
        org_id IN (SELECT org_id FROM org_members WHERE user_id = app_current_user_id())
    );

-- 3c. app_agent_config (from 015)
DROP POLICY IF EXISTS app_agent_config_org ON app_agent_config;
CREATE POLICY app_agent_config_org ON app_agent_config
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN org_members om ON om.org_id = a.org_id
            WHERE om.user_id = app_current_user_id()
        )
    );

-- 3d. monitoring_state (from 021)
DROP POLICY IF EXISTS monitoring_state_user_policy ON monitoring_state;
CREATE POLICY monitoring_state_user_policy ON monitoring_state
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN org_members om ON om.org_id = a.org_id
            WHERE om.user_id = app_current_user_id()
        )
    );

-- 3e. notification_channels (from 021)
DROP POLICY IF EXISTS notification_channels_user_policy ON notification_channels;
CREATE POLICY notification_channels_user_policy ON notification_channels
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN org_members om ON om.org_id = a.org_id
            WHERE om.user_id = app_current_user_id()
        )
    );

-- 3f. notification_preferences (from 021)
DROP POLICY IF EXISTS notification_preferences_user_policy ON notification_preferences;
CREATE POLICY notification_preferences_user_policy ON notification_preferences
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN org_members om ON om.org_id = a.org_id
            WHERE om.user_id = app_current_user_id()
        )
    );

-- 3g. notification_log (from 021)
DROP POLICY IF EXISTS notification_log_user_policy ON notification_log;
CREATE POLICY notification_log_user_policy ON notification_log
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN org_members om ON om.org_id = a.org_id
            WHERE om.user_id = app_current_user_id()
        )
    );

-- 3h. investigation_schedules (from 023)
DROP POLICY IF EXISTS investigation_schedules_org ON investigation_schedules;
CREATE POLICY investigation_schedules_org ON investigation_schedules
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN org_members om ON om.org_id = a.org_id
            WHERE om.user_id = app_current_user_id()
        )
    );

-- Step 4: Drop the old column
DROP INDEX IF EXISTS idx_users_org_id;
ALTER TABLE users DROP COLUMN org_id;
