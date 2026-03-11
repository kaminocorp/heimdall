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

-- Row Level Security: scope via parent connection's user_id.
ALTER TABLE github_repos ENABLE ROW LEVEL SECURITY;

CREATE POLICY github_repos_owner ON github_repos
  FOR ALL
  USING (
    connection_id IN (
      SELECT id FROM connections WHERE user_id = app_current_user_id()
    )
  );
