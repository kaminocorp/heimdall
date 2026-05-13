-- Reverse the Phase 7 follow-up — drop lookup_user_for_invite.
--
-- DROP FUNCTION cascades the EXECUTE grant. No need to REVOKE
-- explicitly; the grant is gone with the function.

DROP FUNCTION IF EXISTS public.lookup_user_for_invite(TEXT);
