-- name: ListGitHubReposByConnection :many
SELECT id, connection_id, repo_full_name, repo_id, default_branch, enabled, created_at
FROM github_repos
WHERE connection_id = $1
ORDER BY repo_full_name;

-- name: UpsertGitHubRepo :one
INSERT INTO github_repos (connection_id, repo_full_name, repo_id, default_branch, enabled)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (connection_id, repo_id)
DO UPDATE SET
    repo_full_name = EXCLUDED.repo_full_name,
    default_branch = EXCLUDED.default_branch,
    enabled = EXCLUDED.enabled
RETURNING *;

-- name: DeleteGitHubRepo :exec
DELETE FROM github_repos WHERE id = $1;

-- name: ListEnabledGitHubReposByApp :many
SELECT gr.id, gr.connection_id, gr.repo_full_name, gr.repo_id, gr.default_branch, gr.enabled, gr.created_at,
       c.config AS connection_config
FROM github_repos gr
JOIN connections c ON c.id = gr.connection_id
WHERE c.app_id = $1
  AND c.type = 'github'
  AND c.status = 'active'
  AND gr.enabled = true
ORDER BY gr.repo_full_name;
