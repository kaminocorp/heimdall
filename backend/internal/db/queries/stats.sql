-- name: GetDashboardStats :one
SELECT
  (SELECT count(*) FROM log_buffer WHERE log_buffer.user_id = $1 AND ingested_at > now() - interval '24 hours')::bigint AS log_count_24h,
  (SELECT count(*) FROM connections WHERE connections.user_id = $1)::bigint AS connection_count,
  (SELECT count(*) FROM connections WHERE connections.user_id = $1 AND status = 'active')::bigint AS active_connections;
