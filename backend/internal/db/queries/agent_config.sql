-- name: GetAgentConfig :one
SELECT * FROM agent_config LIMIT 1;

-- name: UpsertAgentConfig :one
INSERT INTO agent_config (id, model, mode, schedule, system_prompt_override)
VALUES (1, $1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET model = EXCLUDED.model, mode = EXCLUDED.mode, schedule = EXCLUDED.schedule, system_prompt_override = EXCLUDED.system_prompt_override, updated_at = now()
RETURNING *;
