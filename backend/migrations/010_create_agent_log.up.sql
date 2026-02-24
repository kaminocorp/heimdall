CREATE TABLE agent_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    entry_type      TEXT NOT NULL,
    summary         TEXT NOT NULL,
    detail          JSONB,
    severity        TEXT,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agent_log_user_created ON agent_log (user_id, created_at DESC);
CREATE INDEX idx_agent_log_entry_type ON agent_log (entry_type);
