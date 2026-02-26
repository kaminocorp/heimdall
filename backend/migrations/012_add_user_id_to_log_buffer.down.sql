DROP INDEX IF EXISTS idx_log_buffer_user_id_ingested;
ALTER TABLE log_buffer DROP COLUMN user_id;
