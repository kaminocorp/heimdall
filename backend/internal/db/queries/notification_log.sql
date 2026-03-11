-- name: InsertNotificationLog :one
INSERT INTO notification_log (app_id, channel_id, agent_log_id, severity, summary, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateNotificationLogStatus :exec
UPDATE notification_log
SET status = $2, error_message = $3, sent_at = CASE WHEN $2 = 'sent' THEN now() ELSE sent_at END
WHERE id = $1;

-- name: ListNotificationLogByApp :many
SELECT nl.*, nc.type AS channel_type, nc.name AS channel_name
FROM notification_log nl
JOIN notification_channels nc ON nc.id = nl.channel_id
WHERE nl.app_id = $1
ORDER BY nl.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLastNotificationForApp :one
SELECT * FROM notification_log
WHERE app_id = $1 AND status = 'sent'
ORDER BY created_at DESC
LIMIT 1;
