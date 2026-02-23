DROP INDEX IF EXISTS idx_connections_user_id;
ALTER TABLE connections DROP COLUMN IF EXISTS user_id;
