-- Enable RLS on webhook_idempotency (defence-in-depth).
--
-- This table is only accessed by the webhook ingestion handler, which
-- connects as the postgres owner role (bypasses RLS). No policies are
-- needed — with RLS enabled and no GRANT, non-owner roles (anon,
-- authenticated) see zero rows even if PostgREST or a leaked key is
-- used to query the table directly.

ALTER TABLE webhook_idempotency ENABLE ROW LEVEL SECURITY;
