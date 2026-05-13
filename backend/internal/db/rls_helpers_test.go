package db_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// rlsTestEnv wraps a single pgxpool.Pool used by every RLS test in this
// package. The pool authenticates as whatever DATABASE_URL specifies —
// in CI today that is `postgres`. Phase 6 flips it to `app_user` for
// runtime-shape tests; the §5.6d asAppUser fixture explicitly opens a
// SET LOCAL ROLE transaction so RLS regression tests fire correctly
// against `app_user` even when the underlying pool is `postgres`.
//
// Tests register cleanup via t.Cleanup so the package-level pool can
// stay alive for the suite's duration without leaking per-test state.
type rlsTestEnv struct {
	pool *pgxpool.Pool
}

// requireRLSEnv connects to DATABASE_URL and returns a shared env. Skips
// the test if DATABASE_URL is unset (the existing handler-package
// contract — keep CI green when the integration database is absent).
func requireRLSEnv(t *testing.T) *rlsTestEnv {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set; skipping RLS test")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return &rlsTestEnv{pool: pool}
}

// requireRoleExists skips the test with a clear message if `roleName` is
// not present in pg_roles. Bootstrap from
// docs/completions/rls-enforcement-phase-4.md must have run for the role
// split tests to make sense.
func requireRoleExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, roleName string) {
	t.Helper()

	var exists bool
	err := pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)", roleName).Scan(&exists)
	require.NoError(t, err)
	if !exists {
		t.Skipf("role %q not present; bootstrap from docs/completions/rls-enforcement-phase-4.md and apply migration 039", roleName)
	}
}

// asAppUser opens a transaction on `pool`, switches its current_role to
// `app_user` via SET LOCAL ROLE, and pins `app.current_user_id` to
// `userID` so RLS policies evaluate against that caller. The returned
// cleanup rolls back the transaction.
//
// SET LOCAL ROLE requires that the caller has membership in `app_user`.
// Migration 039 grants `app_user TO postgres` (and the PG17 WITH SET TRUE
// variant) for exactly this fixture. If migration 039 hasn't been applied,
// the test skips via requireRoleExists.
//
// This is the §5.6d opt-in fixture: tests that want to exercise RLS
// enforcement under `app_user` call this; tests that want plain superuser
// access don't. Production code never issues SET ROLE; this mechanism
// exists only to let CI assert enforcement before Phase 6 flips the env
// vars to authenticate as `app_user` directly.
func asAppUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (pgx.Tx, func()) {
	t.Helper()
	requireRoleExists(t, ctx, pool, "app_user")

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	if _, err := tx.Exec(ctx, "SET LOCAL ROLE app_user"); err != nil {
		tx.Rollback(ctx)
		t.Fatalf("SET LOCAL ROLE app_user failed: %v (migration 039 not applied?)", err)
	}
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID.String()); err != nil {
		tx.Rollback(ctx)
		t.Fatalf("setting app.current_user_id failed: %v", err)
	}

	cleanup := func() { tx.Rollback(context.Background()) }
	return tx, cleanup
}

// asCronUser is the cron-role analogue of asAppUser. Used by §5.6
// cron-role behaviour tests to assert grant boundaries from the role's
// own perspective.
func asCronUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (pgx.Tx, func()) {
	t.Helper()
	requireRoleExists(t, ctx, pool, "cron_user")

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	if _, err := tx.Exec(ctx, "SET LOCAL ROLE cron_user"); err != nil {
		tx.Rollback(ctx)
		t.Fatalf("SET LOCAL ROLE cron_user failed: %v (migration 039 not applied?)", err)
	}

	cleanup := func() { tx.Rollback(context.Background()) }
	return tx, cleanup
}

// tenant captures the IDs created for one synthetic tenant in a
// cross-tenant regression test. The fields are public so per-table tests
// can read whichever IDs they need.
type tenant struct {
	UserID       uuid.UUID
	OrgID        uuid.UUID
	AppID        uuid.UUID
	ConnectionID uuid.UUID
}

// makeTenant inserts a fresh user / org / org_membership / app /
// connection chain via the given pool (running as postgres / superuser
// for setup speed) and returns the resulting IDs. Cleanup deletes the
// org and the auth.users row, which cascades through the rest.
//
// The `label` parameter is purely cosmetic (it's appended to the email
// and the org slug for human readability when a test fails).
func makeTenant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) tenant {
	t.Helper()

	tn := tenant{
		UserID:       uuid.New(),
		OrgID:        uuid.New(),
		AppID:        uuid.New(),
		ConnectionID: uuid.New(),
	}
	email := fmt.Sprintf("rls-%s-%s@heimdall.test", label, tn.UserID.String()[:8])

	// auth.users → public.users (Supabase trigger replicates, fall through
	// with ON CONFLICT for plain Postgres).
	_, err := pool.Exec(ctx,
		`INSERT INTO auth.users (id, email, instance_id, aud, role, encrypted_password, confirmation_token, created_at, updated_at)
		 VALUES ($1, $2, '00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated', '', '', now(), now())`,
		tn.UserID, email)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		"INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email",
		tn.UserID, email)
	require.NoError(t, err)

	slug := fmt.Sprintf("rls-%s-%s", label, tn.OrgID.String()[:8])
	_, err = pool.Exec(ctx,
		"INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)",
		tn.OrgID, "RLS Test "+label, slug)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'owner')",
		tn.UserID, tn.OrgID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, 'active')",
		tn.AppID, tn.OrgID, "RLS App "+label)
	require.NoError(t, err)

	// A connection on the test app — webhook_logs is the simplest type and
	// has no required external state.
	_, err = pool.Exec(ctx,
		"INSERT INTO connections (id, org_id, user_id, name, type, config) VALUES ($1, $2, $3, $4, 'webhook_logs', '{}'::jsonb)",
		tn.ConnectionID, tn.OrgID, tn.UserID, "RLS Conn "+label)
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.Background()
		pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", tn.OrgID)
		pool.Exec(ctx, "DELETE FROM auth.users WHERE id = $1", tn.UserID)
	})

	return tn
}
