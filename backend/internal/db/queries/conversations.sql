-- name: CreateConversation :one
INSERT INTO conversations (user_id, investigation_id, title, messages)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetConversationByUser :one
SELECT * FROM conversations WHERE id = $1 AND user_id = $2;

-- name: ListConversationsByUser :many
SELECT * FROM conversations
WHERE user_id = $1
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateConversationMessagesByUser :exec
UPDATE conversations SET messages = $2, updated_at = now()
WHERE id = $1 AND user_id = $3;

-- name: UpdateConversationTitleByUser :exec
UPDATE conversations SET title = $2, updated_at = now()
WHERE id = $1 AND user_id = $3;
