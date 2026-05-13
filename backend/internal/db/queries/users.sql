-- name: GetUser :one
SELECT id, email, created_at FROM users WHERE id = $1;

-- name: GetFirstUserInOrg :one
SELECT om.user_id AS id FROM org_members om
WHERE om.org_id = $1
ORDER BY om.created_at ASC LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, created_at FROM users WHERE email = $1;

-- name: HasOrgMembership :one
-- Returns true if the user belongs to any organization.
-- Used by the onboarding idempotency guard (replaces old user.OrgID.Valid check).
SELECT EXISTS(SELECT 1 FROM org_members WHERE user_id = $1) AS has_org;

-- name: LookupUserIDForInvite :one
-- Calls the SECURITY DEFINER helper from migration 041. Bypasses RLS
-- specifically for the invite-by-email flow: under FORCE the regular
-- GetUserByEmail returns ErrNoRows for any user not yet in the
-- caller's orgs, which is wrong for invitation lookups (the target is
-- by definition not yet a member).
--
-- COALESCE collapses "no user has this email" into uuid.Nil so sqlc
-- generates a non-nullable uuid.UUID return — sqlc v1.30 doesn't infer
-- nullability through SECURITY DEFINER function calls, and pgx errors
-- on NULL→uuid.UUID scans. Caller must compare the result against
-- uuid.Nil to detect "not found".
SELECT COALESCE(public.lookup_user_for_invite($1), '00000000-0000-0000-0000-000000000000'::uuid)::uuid AS user_id;
