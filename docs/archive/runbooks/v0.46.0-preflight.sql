-- v0.46.0 migration pre-flight check
--
-- Run this BEFORE `make migrate-up` on any environment that has not yet applied
-- migration 035 (`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;`
-- returns a value < 35).
--
-- Why: migration 035 adds `connections.org_id NOT NULL`, backfilled from
-- `applications.org_id`. If any connection's parent application has a NULL
-- `org_id`, the backfill leaves the row with NULL and the subsequent
-- `ALTER COLUMN org_id SET NOT NULL` aborts mid-migration with
-- `column "org_id" contains null values`, leaving the schema in an inconsistent
-- state.
--
-- Usage: psql "$DATABASE_URL" -f docs/archive/runbooks/v0.46.0-preflight.sql
-- Success: "v0.46.0 pre-flight OK." notice and exit code 0.
-- Failure: exception raised with the orphan count, exit code 3.

\set ON_ERROR_STOP on

DO $$
DECLARE
    orphan_count INT;
BEGIN
    SELECT count(*) INTO orphan_count
    FROM connections c
    LEFT JOIN applications a ON c.app_id = a.id
    WHERE a.org_id IS NULL;

    IF orphan_count > 0 THEN
        RAISE EXCEPTION
            'v0.46.0 pre-flight FAILED: % connection(s) have a parent application with NULL org_id. Investigate and repair before running migrate-up. Diagnostic: SELECT c.id, c.app_id, a.org_id FROM connections c LEFT JOIN applications a ON c.app_id = a.id WHERE a.org_id IS NULL;',
            orphan_count;
    END IF;

    RAISE NOTICE 'v0.46.0 pre-flight OK. Safe to run migrate-up.';
END $$;
