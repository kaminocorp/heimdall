DROP POLICY IF EXISTS users_self ON users;
DROP POLICY IF EXISTS connections_owner ON connections;
DROP POLICY IF EXISTS conversations_owner ON conversations;
DROP POLICY IF EXISTS agent_log_owner ON agent_log;
DROP POLICY IF EXISTS log_buffer_owner ON log_buffer;
DROP POLICY IF EXISTS investigations_owner ON investigations;

ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE connections DISABLE ROW LEVEL SECURITY;
ALTER TABLE conversations DISABLE ROW LEVEL SECURITY;
ALTER TABLE agent_log DISABLE ROW LEVEL SECURITY;
ALTER TABLE log_buffer DISABLE ROW LEVEL SECURITY;
ALTER TABLE investigations DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS app_current_user_id();
