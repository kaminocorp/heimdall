ALTER TABLE log_buffer
    ADD COLUMN user_id UUID REFERENCES public.users(id) ON DELETE CASCADE;

UPDATE log_buffer lb
    SET user_id = c.user_id
    FROM connections c
    WHERE lb.connection_id = c.id;

ALTER TABLE log_buffer ALTER COLUMN user_id SET NOT NULL;

CREATE INDEX idx_log_buffer_user_id_ingested ON log_buffer(user_id, ingested_at DESC);
