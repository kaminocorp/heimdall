package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// TestRLSRegression_CrossTenant exercises the parent plan §5.6 RLS
// regression suite: for each critical table, insert a row owned by tenant
// B (under postgres / superuser) and assert that tenant A — connected via
// asAppUser — cannot see it.
//
// The asAppUser fixture pins app.current_user_id to tenant A's user_id
// and runs `SET LOCAL ROLE app_user`, so RLS evaluates exactly as it
// will after Phase 6's env-var flip. Until Phase 7 ships FORCE,
// `app_user` is non-superuser but the tables are in `rowsecurity = true`
// + (no FORCE) state, which is enough for these regression tests
// because `app_user` is not the table owner. (Owners bypass RLS without
// FORCE; non-owners do not.)
//
// Each subtest is independent so a per-table failure tells you which
// policy regressed. log_pipeline_events / webhook_idempotency are
// deliberately excluded — Phase 1 §5.2 surfaced that they have RLS
// enabled but no policies; Phase 7's Migration B adds the policies and
// adds the matching regression tests at that time.
func TestRLSRegression_CrossTenant(t *testing.T) {
	env := requireRLSEnv(t)
	ctx := context.Background()

	requireRoleExists(t, ctx, env.pool, "app_user")

	tenantA := makeTenant(t, ctx, env.pool, "a")
	tenantB := makeTenant(t, ctx, env.pool, "b")

	t.Run("connections", func(t *testing.T) {
		// makeTenant already inserted one connection per tenant. Read
		// tenant B's connection while authenticated as tenant A; expect
		// pgx.ErrNoRows.
		tx, cleanup := asAppUser(t, ctx, env.pool, tenantA.UserID)
		defer cleanup()

		var id uuid.UUID
		err := tx.QueryRow(ctx,
			"SELECT id FROM connections WHERE id = $1", tenantB.ConnectionID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"tenant A saw tenant B's connection (RLS regression on connections_org_member): err=%v", err)
	})

	t.Run("log_buffer", func(t *testing.T) {
		// Insert a log buffer row owned by tenant B (under postgres).
		logID := uuid.New()
		_, err := env.pool.Exec(ctx,
			`INSERT INTO log_buffer (id, user_id, app_id, connection_id, payload, ingested_at)
			 VALUES ($1, $2, $3, $4, '{"msg":"tenant-b-only"}'::jsonb, now())`,
			logID, tenantB.UserID, tenantB.AppID, tenantB.ConnectionID)
		require.NoError(t, err)
		t.Cleanup(func() { env.pool.Exec(context.Background(), "DELETE FROM log_buffer WHERE id = $1", logID) })

		tx, cleanup := asAppUser(t, ctx, env.pool, tenantA.UserID)
		defer cleanup()

		var id uuid.UUID
		err = tx.QueryRow(ctx, "SELECT id FROM log_buffer WHERE id = $1", logID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"tenant A saw tenant B's log_buffer row (RLS regression on log_buffer_owner): err=%v", err)
	})

	t.Run("conversations", func(t *testing.T) {
		convoID := uuid.New()
		_, err := env.pool.Exec(ctx,
			`INSERT INTO conversations (id, user_id, title, messages)
			 VALUES ($1, $2, 'tenant-b-only', '[]'::jsonb)`,
			convoID, tenantB.UserID)
		require.NoError(t, err)
		t.Cleanup(func() { env.pool.Exec(context.Background(), "DELETE FROM conversations WHERE id = $1", convoID) })

		tx, cleanup := asAppUser(t, ctx, env.pool, tenantA.UserID)
		defer cleanup()

		var id uuid.UUID
		err = tx.QueryRow(ctx, "SELECT id FROM conversations WHERE id = $1", convoID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"tenant A saw tenant B's conversation (RLS regression on conversations_owner): err=%v", err)
	})

	t.Run("connection_sources", func(t *testing.T) {
		// Insert a discovered source name on tenant B's connection.
		sourceID := uuid.New()
		_, err := env.pool.Exec(ctx,
			`INSERT INTO connection_sources (id, connection_id, source_name, first_seen_at, last_seen_at)
			 VALUES ($1, $2, $3, now(), now())`,
			sourceID, tenantB.ConnectionID, "tenant-b-source")
		require.NoError(t, err)
		t.Cleanup(func() { env.pool.Exec(context.Background(), "DELETE FROM connection_sources WHERE id = $1", sourceID) })

		tx, cleanup := asAppUser(t, ctx, env.pool, tenantA.UserID)
		defer cleanup()

		var id uuid.UUID
		err = tx.QueryRow(ctx, "SELECT id FROM connection_sources WHERE id = $1", sourceID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"tenant A saw tenant B's connection_source (RLS regression on connection_sources_org_member): err=%v", err)
	})

	t.Run("app_source_filters", func(t *testing.T) {
		filterID := uuid.New()
		_, err := env.pool.Exec(ctx,
			`INSERT INTO app_source_filters (id, app_id, connection_id, source_name, enabled)
			 VALUES ($1, $2, $3, 'tenant-b-source', true)`,
			filterID, tenantB.AppID, tenantB.ConnectionID)
		require.NoError(t, err)
		t.Cleanup(func() { env.pool.Exec(context.Background(), "DELETE FROM app_source_filters WHERE id = $1", filterID) })

		tx, cleanup := asAppUser(t, ctx, env.pool, tenantA.UserID)
		defer cleanup()

		var id uuid.UUID
		err = tx.QueryRow(ctx, "SELECT id FROM app_source_filters WHERE id = $1", filterID).Scan(&id)
		require.True(t, errors.Is(err, pgx.ErrNoRows),
			"tenant A saw tenant B's app_source_filter (RLS regression on app_source_filters_owner): err=%v", err)
	})
}
