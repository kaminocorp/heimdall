-- Reverse Phase 8, Step 1

DROP INDEX IF EXISTS idx_connections_app_id;
ALTER TABLE connections DROP COLUMN IF EXISTS app_id;

DROP TABLE IF EXISTS applications;

DROP INDEX IF EXISTS idx_users_org_id;
ALTER TABLE users DROP COLUMN IF EXISTS org_id;

DROP TABLE IF EXISTS organizations;
