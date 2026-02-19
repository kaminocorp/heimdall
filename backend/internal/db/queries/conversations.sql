-- name: CreateConversation :one
INSERT INTO conversations (investigation_id, title, messages)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = $1;

-- name: ListConversations :many
SELECT * FROM conversations ORDER BY created_at DESC;

-- name: UpdateConversationMessages :exec
UPDATE conversations SET messages = $2, updated_at = now() WHERE id = $1;
