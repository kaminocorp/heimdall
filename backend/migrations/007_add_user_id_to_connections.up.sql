-- Scope connections to the owning user.
ALTER TABLE connections
    ADD COLUMN user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE;

CREATE INDEX idx_connections_user_id ON connections (user_id);
