DROP INDEX IF EXISTS idx_conversations_user_id;
ALTER TABLE conversations DROP COLUMN user_id;
