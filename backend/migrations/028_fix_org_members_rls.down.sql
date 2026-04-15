-- Revert to the original (broken) policy from migration 027.
-- This is intentionally reverting to the recursive version so that
-- re-applying 028.up re-fixes it.
DROP POLICY IF EXISTS org_members_user_policy ON org_members;
DROP FUNCTION IF EXISTS app_user_org_ids();

CREATE POLICY org_members_user_policy ON org_members
    FOR ALL USING (
        org_id IN (SELECT om.org_id FROM org_members om WHERE om.user_id = app_current_user_id())
    );
