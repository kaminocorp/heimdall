-- name: InsertAgentLog :one
INSERT INTO agent_log (user_id, entry_type, summary, detail, severity, conversation_id, app_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAgentLogByUser :many
SELECT * FROM agent_log
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAgentLogByUserAndType :many
SELECT * FROM agent_log
WHERE user_id = $1 AND entry_type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAgentLogByUser :one
SELECT count(*) FROM agent_log
WHERE user_id = $1;

-- name: ListAgentLogByApp :many
SELECT * FROM agent_log
WHERE user_id = $1 AND app_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAgentLogByApp :one
SELECT count(*) FROM agent_log
WHERE user_id = $1 AND app_id = $2;
