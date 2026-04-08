-- name: GetAppAgentConfig :one
SELECT * FROM app_agent_config WHERE app_id = $1;

-- name: UpsertAppAgentConfig :one
INSERT INTO app_agent_config (app_id, model, mode, schedule_interval_secs, system_prompt_override, provider)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (app_id) DO UPDATE
SET model = EXCLUDED.model,
    mode = EXCLUDED.mode,
    schedule_interval_secs = EXCLUDED.schedule_interval_secs,
    system_prompt_override = EXCLUDED.system_prompt_override,
    provider = EXCLUDED.provider,
    updated_at = now()
RETURNING *;
