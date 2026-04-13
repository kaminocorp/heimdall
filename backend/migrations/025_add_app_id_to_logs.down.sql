DROP INDEX IF EXISTS idx_agent_log_app_id;
DROP INDEX IF EXISTS idx_log_buffer_app_id;
ALTER TABLE agent_log DROP COLUMN IF EXISTS app_id;
ALTER TABLE log_buffer DROP COLUMN IF EXISTS app_id;
