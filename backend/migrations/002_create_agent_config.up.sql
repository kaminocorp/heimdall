CREATE TABLE agent_config (
    id          INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    model       TEXT NOT NULL DEFAULT 'claude-sonnet-4-6',
    mode        TEXT NOT NULL DEFAULT 'continuous',
    schedule    TEXT,
    system_prompt_override TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
