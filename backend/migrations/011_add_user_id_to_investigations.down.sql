DROP INDEX IF EXISTS idx_investigations_user_id;
ALTER TABLE investigations DROP COLUMN user_id;
