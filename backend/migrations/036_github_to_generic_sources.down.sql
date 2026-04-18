-- Reverse Phase 3: recreate github_repos and re-seed from the generic
-- tables. repo_id is unrecoverable (discarded during the up-migration), so
-- we synthesise negative placeholder ids so the NOT NULL + unique-index
-- constraints still hold — any rollback that matters in production should
-- be preceded by a fresh GitHub sync that repopulates real repo_ids via
-- the API. default_branch falls back to 'main' which matches the original
-- column default.

CREATE TABLE github_repos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id   UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    repo_full_name  TEXT NOT NULL,
    repo_id         BIGINT NOT NULL,
    default_branch  TEXT NOT NULL DEFAULT 'main',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_github_repos_conn_repo ON github_repos(connection_id, repo_id);
CREATE INDEX idx_github_repos_connection_id ON github_repos(connection_id);

ALTER TABLE github_repos ENABLE ROW LEVEL SECURITY;

CREATE POLICY github_repos_owner ON github_repos
  FOR ALL
  USING (
    connection_id IN (
      SELECT id FROM connections WHERE user_id = app_current_user_id()
    )
  );

-- Re-seed from app_source_filters rows whose parent connection is a GitHub
-- connection. One row per (connection, source_name). Negative row_number
-- keeps repo_id unique within a connection — an honest placeholder that
-- won't collide with any real GitHub numeric id.
INSERT INTO github_repos (connection_id, repo_full_name, repo_id, default_branch, enabled, created_at)
SELECT
    asf.connection_id,
    asf.source_name,
    -(row_number() OVER (PARTITION BY asf.connection_id ORDER BY asf.source_name))::BIGINT,
    'main',
    asf.enabled,
    asf.created_at
FROM app_source_filters asf
JOIN connections c ON c.id = asf.connection_id
WHERE c.type = 'github';
