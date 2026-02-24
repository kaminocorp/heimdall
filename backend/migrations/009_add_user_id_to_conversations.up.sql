ALTER TABLE conversations
    ADD COLUMN user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE;

CREATE INDEX idx_conversations_user_id ON conversations (user_id);
