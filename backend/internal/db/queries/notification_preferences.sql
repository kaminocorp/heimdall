-- name: GetNotificationPreferences :one
SELECT * FROM notification_preferences WHERE app_id = $1;

-- name: UpsertNotificationPreferences :one
INSERT INTO notification_preferences (app_id, enabled, severity_threshold, cooldown_minutes)
VALUES ($1, $2, $3, $4)
ON CONFLICT (app_id) DO UPDATE
SET enabled = EXCLUDED.enabled,
    severity_threshold = EXCLUDED.severity_threshold,
    cooldown_minutes = EXCLUDED.cooldown_minutes,
    updated_at = now()
RETURNING *;
