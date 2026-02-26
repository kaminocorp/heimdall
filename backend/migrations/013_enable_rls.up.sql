-- =====================================================
-- Row Level Security policies for all user-scoped tables.
-- Uses app.current_user_id set via set_config() per transaction.
--
-- The backend connects as the postgres (owner) role which
-- bypasses RLS by default. These policies protect against
-- non-owner access paths: Supabase anon/authenticated roles,
-- PostgREST, and direct psql with other roles.
-- =====================================================

-- Helper: safely read the session user (returns NULL if unset).
CREATE OR REPLACE FUNCTION app_current_user_id() RETURNS UUID AS $$
  SELECT nullif(current_setting('app.current_user_id', true), '')::UUID;
$$ LANGUAGE sql STABLE;

-- ── users ────────────────────────────────────────────
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_self ON users
  FOR ALL
  USING (id = app_current_user_id());

-- ── connections ──────────────────────────────────────
ALTER TABLE connections ENABLE ROW LEVEL SECURITY;

CREATE POLICY connections_owner ON connections
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── conversations ────────────────────────────────────
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;

CREATE POLICY conversations_owner ON conversations
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── agent_log ────────────────────────────────────────
ALTER TABLE agent_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY agent_log_owner ON agent_log
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── log_buffer ───────────────────────────────────────
ALTER TABLE log_buffer ENABLE ROW LEVEL SECURITY;

CREATE POLICY log_buffer_owner ON log_buffer
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── investigations ───────────────────────────────────
ALTER TABLE investigations ENABLE ROW LEVEL SECURITY;

CREATE POLICY investigations_owner ON investigations
  FOR ALL
  USING (user_id = app_current_user_id());
