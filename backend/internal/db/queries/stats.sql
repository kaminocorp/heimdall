-- name: GetAppDashboardStats :one
SELECT
  (SELECT count(*) FROM log_buffer lb WHERE lb.app_id = $1 AND lb.ingested_at > now() - interval '24 hours')::bigint AS log_count_24h,
  (SELECT count(*) FROM connections c WHERE c.app_id = $1)::bigint AS connection_count,
  (SELECT count(*) FROM connections c2 WHERE c2.app_id = $1 AND c2.status = 'active')::bigint AS active_connections;
