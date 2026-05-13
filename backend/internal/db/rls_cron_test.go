package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// TestCronRole_Behaviour exercises the four parent-plan §5.6 cron-role
// posture tests. Each subtest opens an asCronUser transaction (= postgres
// session with `SET LOCAL ROLE cron_user`) and verifies one structural
// property of the cron role's grant set.
//
// "42501" is the SQLSTATE for "insufficient_privilege" — the canonical
// permission-denied code. The tests assert the specific code rather than
// a substring of the human-readable message because Postgres translates
// messages and we want the assertion to survive locale changes.
func TestCronRole_Behaviour(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()

	requireRoleExists(t, ctx, env.pool, "cron_user")

	// One pre-built tenant so the enumerate test has at least one row to find.
	tn := makeTenant(t, ctx, env.pool, "cron")

	t.Run("EnumerateAcrossTenantsSucceeds", func(t *testing.T) {
		// Cross-tenant SELECT is the cron role's whole purpose. Should
		// return at least the tenant we just created.
		tx, cleanup := asCronUser(t, ctx, env.pool)
		defer cleanup()

		var count int
		err := tx.QueryRow(ctx,
			"SELECT count(*) FROM applications WHERE id = $1", tn.AppID).Scan(&count)
		require.NoError(t, err, "cron_user should be able to enumerate applications across tenants")
		require.Equal(t, 1, count, "cron_user could not see the just-created tenant's app")
	})

	t.Run("DropTableDenied", func(t *testing.T) {
		// DDL is structurally outside the cron role's grant set — only
		// `postgres` (via DIRECT_URL) can issue DDL.
		tx, cleanup := asCronUser(t, ctx, env.pool)
		defer cleanup()

		_, err := tx.Exec(ctx, "DROP TABLE IF EXISTS connections")
		require.Error(t, err, "cron_user must not be able to DROP TABLE")

		var pgErr *pgconn.PgError
		require.True(t, errors.As(err, &pgErr), "expected a PgError, got %T: %v", err, err)
		require.Equal(t, "42501", pgErr.Code,
			"expected SQLSTATE 42501 (insufficient_privilege), got %s: %s", pgErr.Code, pgErr.Message)
	})

	t.Run("AuthUsersReadDenied", func(t *testing.T) {
		// auth.* is a Supabase-internal schema. Neither runtime role
		// has USAGE on it; reads against auth.users must fail at the
		// schema/table level, not silently return an empty set.
		tx, cleanup := asCronUser(t, ctx, env.pool)
		defer cleanup()

		_, err := tx.Exec(ctx, "SELECT id FROM auth.users LIMIT 1")
		require.Error(t, err, "cron_user must not be able to read auth.users")

		var pgErr *pgconn.PgError
		require.True(t, errors.As(err, &pgErr), "expected a PgError, got %T: %v", err, err)
		require.Equal(t, "42501", pgErr.Code,
			"expected SQLSTATE 42501 (insufficient_privilege), got %s: %s", pgErr.Code, pgErr.Message)
	})

	t.Run("BlanketTenantTableInsertDenied", func(t *testing.T) {
		// log_buffer is a tenant table. The §4.2 grant set gives cron_user
		// SELECT on all tables and DELETE on log_buffer specifically (for
		// the pruner) — but no INSERT. An attempted INSERT must fail with
		// 42501 at the privilege layer, not silently succeed because of
		// BYPASSRLS. (BYPASSRLS doesn't add grants; it only suppresses
		// RLS policy evaluation.)
		tx, cleanup := asCronUser(t, ctx, env.pool)
		defer cleanup()

		_, err := tx.Exec(ctx,
			`INSERT INTO log_buffer (id, user_id, app_id, connection_id, payload, ingested_at)
			 VALUES ($1, $2, $3, $4, '{}'::jsonb, now())`,
			uuid.New(), tn.UserID, tn.AppID, tn.ConnectionID)
		require.Error(t, err, "cron_user must not be able to INSERT into log_buffer (tenant-write boundary)")

		var pgErr *pgconn.PgError
		require.True(t, errors.As(err, &pgErr), "expected a PgError, got %T: %v", err, err)
		require.Equal(t, "42501", pgErr.Code,
			"expected SQLSTATE 42501 (insufficient_privilege), got %s: %s", pgErr.Code, pgErr.Message)
	})
}
