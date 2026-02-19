-- name: GetAgentConfig :one
SELECT * FROM agent_config LIMIT 1;

-- name: UpsertAgentConfig :one
INSERT INTO agent_config (model, mode, schedule, system_prompt_override)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET model = $1, mode = $2, schedule = $3, system_prompt_override = $4, updated_at = now()
RETURNING *;
