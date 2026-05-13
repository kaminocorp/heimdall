-- Reverse Migration A of the RLS role split (Phase 4).
--
-- Revokes every grant 039.up issued, in reverse order. Roles themselves are
-- NOT dropped — their lifecycle is environment-bootstrap-only, mirroring the
-- up migration's deliberate separation of role lifecycle from grant state.
--
-- Safe to apply even if the bootstrap never ran: REVOKE on a non-existent
-- grantee is a no-op. We still skip the body when roles are absent so the
-- log line stays clean.
--
-- Live-traffic guard (Phase 10 polish): if either runtime role currently
-- holds active connections, refuse to revoke unless an explicit override
-- is set. Revoking USAGE ON SCHEMA public from a role that's serving
-- traffic kills every in-flight query mid-statement and breaks subsequent
-- queries on the same connection until the pool reconnects. The guard
-- forces an operator to confirm they've drained traffic (or accept the
-- impact). Override with:
--
--     SET LOCAL heimdall.allow_revoke_with_traffic = 'yes';
--
-- before running this migration if you genuinely need to revoke against
-- a live system (e.g. emergency lockdown).

DO $$
DECLARE
    active_count int;
    override     text;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user')
       OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cron_user') THEN
        RAISE NOTICE 'app_user / cron_user not present; skipping revoke (nothing to undo)';
        RETURN;
    END IF;

    SELECT count(*) INTO active_count
    FROM pg_stat_activity
    WHERE usename IN ('app_user', 'cron_user') AND pid <> pg_backend_pid();

    IF active_count > 0 THEN
        BEGIN
            override := current_setting('heimdall.allow_revoke_with_traffic', true);
        EXCEPTION WHEN OTHERS THEN
            override := '';
        END;
        IF override IS DISTINCT FROM 'yes' THEN
            RAISE EXCEPTION
                'refusing to revoke runtime grants while % active connections still authenticate as app_user/cron_user; drain traffic or set heimdall.allow_revoke_with_traffic = ''yes'' to override',
                active_count;
        END IF;
        RAISE WARNING 'revoking runtime grants with % active app_user/cron_user connections (override engaged)', active_count;
    END IF;

    -- 5. Function execute grants
    EXECUTE 'REVOKE EXECUTE ON FUNCTION public.app_user_org_ids() FROM app_user, cron_user';
    EXECUTE 'REVOKE EXECUTE ON FUNCTION public.app_current_user_id() FROM app_user, cron_user';

    -- 4. Role membership of postgres (covers both PG17 WITH SET TRUE and pre-PG17 plain forms)
    EXECUTE 'REVOKE cron_user FROM postgres';
    EXECUTE 'REVOKE app_user FROM postgres';

    -- 3. cron_user — defaults, narrow writes, blanket SELECT, schema USAGE
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public REVOKE USAGE, SELECT ON SEQUENCES FROM cron_user';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public REVOKE SELECT ON TABLES FROM cron_user';
    EXECUTE 'REVOKE DELETE ON log_buffer FROM cron_user';
    EXECUTE 'REVOKE INSERT, UPDATE ON monitoring_state FROM cron_user';
    EXECUTE 'REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM cron_user';
    EXECUTE 'REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM cron_user';
    EXECUTE 'REVOKE USAGE ON SCHEMA public FROM cron_user';

    -- 2. app_user — defaults, then current grants, then schema USAGE
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public REVOKE USAGE, SELECT ON SEQUENCES FROM app_user';
    EXECUTE 'ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM app_user';
    EXECUTE 'REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM app_user';
    EXECUTE 'REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM app_user';
    EXECUTE 'REVOKE USAGE ON SCHEMA public FROM app_user';
END $$;
