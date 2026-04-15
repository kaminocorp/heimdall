-- Fix infinite recursion in org_members RLS policy.
-- The policy from migration 027 queries org_members inside its own USING
-- clause, which PostgreSQL detects as infinite recursion. The fix uses a
-- SECURITY DEFINER function that bypasses RLS when resolving the caller's
-- org memberships.

DROP POLICY IF EXISTS org_members_user_policy ON org_members;

CREATE OR REPLACE FUNCTION app_user_org_ids()
  RETURNS SETOF UUID
  LANGUAGE sql STABLE
  SECURITY DEFINER
  SET search_path = public
AS $$
  SELECT org_id FROM org_members WHERE user_id = app_current_user_id();
$$;

CREATE POLICY org_members_user_policy ON org_members
    FOR ALL USING (
        org_id IN (SELECT app_user_org_ids())
    );
