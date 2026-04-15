-- name: GetMonitoringState :one
SELECT * FROM monitoring_state WHERE app_id = $1;

-- name: UpsertMonitoringState :exec
INSERT INTO monitoring_state (app_id, last_monitored_at, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (app_id) DO UPDATE
SET last_monitored_at = $2, updated_at = now();

-- name: ResetMonitoringCursor :exec
INSERT INTO monitoring_state (app_id, last_monitored_at, updated_at)
VALUES ($1, now(), now())
ON CONFLICT (app_id) DO UPDATE
SET last_monitored_at = now(), updated_at = now();

-- name: ListActiveApplications :many
SELECT a.id, a.org_id, a.name, c.mode, c.schedule_interval_secs
FROM applications a
JOIN app_agent_config c ON c.app_id = a.id
WHERE a.status = 'active'
  AND c.mode != 'off'
  AND EXISTS (
      SELECT 1 FROM connections conn
      WHERE conn.app_id = a.id AND conn.status = 'active'
  );

-- name: ListLogsSinceForApp :many
SELECT * FROM log_buffer
WHERE app_id = $1 AND ingested_at > $2
ORDER BY ingested_at ASC
LIMIT $3;
