-- Phase 7 follow-up — SECURITY DEFINER helper for the org-member
-- invite-by-email flow.
--
-- Background. After migration 040 ships FORCE ROW LEVEL SECURITY +
-- the users_org_visible SELECT widening, the existing org_members
-- AddOrgMember handler still cannot find an invitee who is not yet a
-- member of the inviter's org: users_self covers only the caller's own
-- row, users_org_visible covers only org-mates, and a not-yet-invited
-- user is by definition neither. GetUserByEmail returns pgx.ErrNoRows
-- under FORCE, and the handler reports "user not found — they must
-- have a Heimdall account first" even when the user does exist. The
-- Phase 7 completion doc named this regression as the lone follow-up.
--
-- Fix shape. A SECURITY DEFINER function whose body is a single
-- email→user_id lookup, owned by postgres (so it executes with
-- BYPASSRLS and reads the unfiltered users table), with EXECUTE
-- granted to app_user. The function exposes ONE bit of information —
-- "does a user with this email exist, and what is their UUID?" —
-- which is the minimum viable surface for the invite flow. It does NOT
-- expose timestamps, full row data, or anything else; if the invite
-- flow later wants the canonical email-as-stored, it can be threaded
-- through here too, but the response body already echoes the
-- caller-supplied email and that's been correct enough so far.
--
-- search_path is pinned to `public` to defeat the canonical
-- DEFINER-function-with-mutable-search_path attack class (where a
-- caller's session-level search_path could redirect the lookup to a
-- function in a malicious schema). Stylistic mirror of the existing
-- app_user_org_ids() helper from migration 028.
--
-- The function returns RETURNS UUID (single value, NULL when no
-- match), not RETURNS SETOF UUID. Single-value return composes
-- cleanly with sqlc's nullable-uuid override (see backend/sqlc.yaml)
-- and yields a *uuid.UUID return type that the handler checks for nil
-- — matching the explicit "user not found" branch better than the
-- ErrNoRows-on-empty-set alternative.

CREATE OR REPLACE FUNCTION public.lookup_user_for_invite(target_email TEXT)
RETURNS UUID
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = public
AS $$
    SELECT u.id
    FROM public.users u
    WHERE u.email = target_email
    LIMIT 1;
$$;

-- Deny PUBLIC; grant only to app_user. cron_user has no business
-- looking up invitees (every cron path resolves owner via cron-pool
-- enumeration, never via email); leaving cron_user off the EXECUTE
-- grant keeps the function's reach narrow.
REVOKE ALL ON FUNCTION public.lookup_user_for_invite(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.lookup_user_for_invite(TEXT) TO app_user;
