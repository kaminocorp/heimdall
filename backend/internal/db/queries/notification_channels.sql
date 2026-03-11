-- name: ListNotificationChannelsByApp :many
SELECT * FROM notification_channels
WHERE app_id = $1
ORDER BY created_at ASC;

-- name: GetNotificationChannel :one
SELECT * FROM notification_channels
WHERE id = $1;

-- name: CreateNotificationChannel :one
INSERT INTO notification_channels (app_id, type, name, config, enabled)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateNotificationChannel :one
UPDATE notification_channels
SET name = $2, config = $3, enabled = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteNotificationChannel :exec
DELETE FROM notification_channels WHERE id = $1;

-- name: ListEnabledChannelsByApp :many
SELECT * FROM notification_channels
WHERE app_id = $1 AND enabled = true;
