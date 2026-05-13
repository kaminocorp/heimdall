package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// TestRLSForce_PostgresOwnerIsBound pins the load-bearing property of
// FORCE ROW LEVEL SECURITY (Phase 7, migration 040): even the table owner
// — `postgres` — is subject to the policy. Without FORCE, the owner
// bypasses RLS entirely; a runtime path that accidentally connected as
// `postgres` (forgetting to flip DATABASE_URL to `app_user`) would silently
// see every tenant's rows.
//
// This is the *behavioural* counterpart to TestCatalogDrift_ForceCoverage
// in rls_catalog_test.go. The catalog test pins `forcerowsecurity = true`
// in pg_class; this test pins what FORCE actually means at query time.
//
// Test shape: open a transaction *as postgres* (the table owner), pin
// `app.current_user_id` to tenant A's user, and assert tenant B's rows
// are invisible. The asAppUser fixture explicitly switches role to
// `app_user`; here we deliberately do NOT — running as postgres is the
// whole point.
func TestRLSForce_PostgresOwnerIsBound(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()

	tenantA := makeTenant(t, ctx, env.pool, "force-a")
	tenantB := makeTenant(t, ctx, env.pool, "force-b")

	// Sanity: assert we are actually running as postgres. If DATABASE_URL
	// has already been flipped to app_user (Phase 6 dev parity), this
	// test isn't exercising the FORCE-vs-owner case — skip rather than
	// produce a misleading green.
	var currentUser string
	require.NoError(t, env.pool.QueryRow(ctx, "SELECT current_user").Scan(&currentUser))
	if currentUser != "postgres" {
		t.Skipf("DATABASE_URL connects as %q, not postgres — the FORCE-vs-owner case requires the table-owner role", currentUser)
	}

	// Stage tenant B rows on tables that are RLS-enabled + FORCE-on.
	// connections is the canonical case (org-scoped policy via
	// connections_org_member); log_buffer is the high-write case
	// (per-row owner policy).

	t.Run("connections invisible to other tenant under FORCE", func(t *testing.T) {
		tx, err := env.pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		// We stay as postgres — no SET LOCAL ROLE — and pin the GUC to
		// tenant A's user. With FORCE, the connections_org_member policy
		// must hide tenant B's row from this txn.
		_, err = tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", tenantA.UserID.String())
		require.NoError(t, err)

		var id uuid.UUID
		err = tx.QueryRow(ctx,
			"SELECT id FROM connections WHERE id = $1", tenantB.ConnectionID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"FORCE should hide tenant B's connection from a postgres-role txn pinned to tenant A; got err=%v id=%v", err, id)
	})

	t.Run("log_buffer invisible to other tenant under FORCE", func(t *testing.T) {
		// Insert a tenant-B-owned log_buffer row from outside the txn —
		// running as postgres, before FORCE is in scope, so the insert
		// itself is unimpeded.
		logID := uuid.New()
		_, err := env.pool.Exec(ctx,
			`INSERT INTO log_buffer (id, user_id, app_id, connection_id, payload, ingested_at)
			 VALUES ($1, $2, $3, $4, '{"msg":"force-b"}'::jsonb, now())`,
			logID, tenantB.UserID, tenantB.AppID, tenantB.ConnectionID)
		require.NoError(t, err)
		t.Cleanup(func() { env.pool.Exec(context.Background(), "DELETE FROM log_buffer WHERE id = $1", logID) })

		tx, err := env.pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", tenantA.UserID.String())
		require.NoError(t, err)

		var id uuid.UUID
		err = tx.QueryRow(ctx, "SELECT id FROM log_buffer WHERE id = $1", logID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"FORCE should hide tenant B's log_buffer row from a postgres-role txn pinned to tenant A; got err=%v id=%v", err, id)
	})

	t.Run("conversations invisible to other tenant under FORCE", func(t *testing.T) {
		convoID := uuid.New()
		_, err := env.pool.Exec(ctx,
			`INSERT INTO conversations (id, user_id, title, messages)
			 VALUES ($1, $2, 'force-b', '[]'::jsonb)`,
			convoID, tenantB.UserID)
		require.NoError(t, err)
		t.Cleanup(func() { env.pool.Exec(context.Background(), "DELETE FROM conversations WHERE id = $1", convoID) })

		tx, err := env.pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", tenantA.UserID.String())
		require.NoError(t, err)

		var id uuid.UUID
		err = tx.QueryRow(ctx, "SELECT id FROM conversations WHERE id = $1", convoID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"FORCE should hide tenant B's conversation from a postgres-role txn pinned to tenant A; got err=%v id=%v", err, id)
	})
}
