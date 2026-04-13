-- Activity feed app-scoping: add app_id to log_buffer and agent_log
-- so the Activity page can filter by the currently selected application.

-- log_buffer: denormalised app_id for direct filtering.
ALTER TABLE log_buffer
  ADD COLUMN app_id UUID REFERENCES applications(id) ON DELETE CASCADE;

-- Backfill from the connection's app_id for existing rows.
UPDATE log_buffer lb
  SET app_id = c.app_id
  FROM connections c
  WHERE lb.connection_id = c.id;

-- agent_log: new app_id column.
ALTER TABLE agent_log
  ADD COLUMN app_id UUID REFERENCES applications(id) ON DELETE CASCADE;

-- Backfill agent_log where detail contains app_id.
UPDATE agent_log
  SET app_id = (detail->>'app_id')::uuid
  WHERE detail IS NOT NULL
    AND detail->>'app_id' IS NOT NULL;

-- Indexes for the new filter path.
-- Partial indexes exclude NULLs that will never match a filtered query.
CREATE INDEX idx_log_buffer_app_id ON log_buffer (app_id, ingested_at DESC)
  WHERE app_id IS NOT NULL;
CREATE INDEX idx_agent_log_app_id ON agent_log (app_id, created_at DESC)
  WHERE app_id IS NOT NULL;
