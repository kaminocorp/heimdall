-- Enable RLS on org_members so that DB-level policies enforce tenant isolation
-- even if handler-level checks are bypassed.
ALTER TABLE org_members ENABLE ROW LEVEL SECURITY;

-- Users can see/modify memberships only for orgs they themselves belong to.
CREATE POLICY org_members_user_policy ON org_members
    FOR ALL USING (
        org_id IN (SELECT om.org_id FROM org_members om WHERE om.user_id = app_current_user_id())
    );
