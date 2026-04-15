-- Rollback: revert app_id to nullable and restore partial indexes.

ALTER TABLE log_buffer ALTER COLUMN app_id DROP NOT NULL;
ALTER TABLE agent_log  ALTER COLUMN app_id DROP NOT NULL;

-- Restore partial indexes matching migration 025.
DROP INDEX IF EXISTS idx_log_buffer_app_id;
DROP INDEX IF EXISTS idx_agent_log_app_id;
CREATE INDEX idx_log_buffer_app_id ON log_buffer (app_id, ingested_at DESC)
  WHERE app_id IS NOT NULL;
CREATE INDEX idx_agent_log_app_id ON agent_log (app_id, created_at DESC)
  WHERE app_id IS NOT NULL;
