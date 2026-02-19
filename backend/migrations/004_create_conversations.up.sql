CREATE TABLE conversations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investigation_id  UUID REFERENCES investigations(id) ON DELETE SET NULL,
    title             TEXT,
    messages          JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_conversations_investigation ON conversations (investigation_id) WHERE investigation_id IS NOT NULL;
