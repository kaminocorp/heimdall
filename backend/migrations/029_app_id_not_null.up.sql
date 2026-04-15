-- Enforce NOT NULL on app_id for log_buffer and agent_log.
--
-- Migration 025 added app_id as nullable and backfilled where possible.
-- Rows that still have NULL app_id are orphans — they are already invisible
-- to the Activity feed (queries filter by app_id = $1) so deleting them
-- closes the integrity gap without data-loss from the user's perspective.

-- Remove orphaned rows with no app association.
DELETE FROM log_buffer WHERE app_id IS NULL;
DELETE FROM agent_log WHERE app_id IS NULL;

-- Add NOT NULL constraint.
ALTER TABLE log_buffer ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE agent_log  ALTER COLUMN app_id SET NOT NULL;

-- Replace partial indexes (WHERE app_id IS NOT NULL) with full indexes
-- now that NULLs are impossible.
DROP INDEX IF EXISTS idx_log_buffer_app_id;
DROP INDEX IF EXISTS idx_agent_log_app_id;
CREATE INDEX idx_log_buffer_app_id ON log_buffer (app_id, ingested_at DESC);
CREATE INDEX idx_agent_log_app_id  ON agent_log  (app_id, created_at DESC);
