package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestAppPool_IsActuallyAppUser (§5.6b) is the post-Phase-6 assertion
// that DATABASE_URL really does authenticate as `app_user`. Today (Phase
// 5) DATABASE_URL still points at `postgres`, so the test would
// trivially fail; gated on HEIMDALL_ROLE_SPLIT_LIVE=1 so it stays
// skipped until Phase 6 flips the env var.
//
// Phase 6's job is to set HEIMDALL_ROLE_SPLIT_LIVE=1 in CI as part of
// the env-var rollout. Once flipped, this test becomes the catch for
// "did the deploy environment actually pick up the new DATABASE_URL?"
// — a class of regression that's otherwise invisible (everything still
// "works" for as long as the postgres-owner bypass keeps RLS cosmetic).
func TestAppPool_IsActuallyAppUser(t *testing.T) {
	if !envFlagOn("HEIMDALL_ROLE_SPLIT_LIVE") {
		t.Skip("HEIMDALL_ROLE_SPLIT_LIVE not set; connectivity test runs after Phase 6 flips the env vars")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	var role string
	err = pool.QueryRow(ctx, "SELECT current_user").Scan(&role)
	require.NoError(t, err)
	require.Equal(t, "app_user", role,
		"DATABASE_URL must authenticate as app_user post-Phase-6; got %q", role)
}

// TestCronPool_IsActuallyCronUser (§5.6b) — same shape for the cron pool.
func TestCronPool_IsActuallyCronUser(t *testing.T) {
	if !envFlagOn("HEIMDALL_ROLE_SPLIT_LIVE") {
		t.Skip("HEIMDALL_ROLE_SPLIT_LIVE not set; connectivity test runs after Phase 6 flips the env vars")
	}

	cronURL := os.Getenv("CRON_DATABASE_URL")
	if cronURL == "" {
		t.Skip("CRON_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cronURL)
	require.NoError(t, err)
	defer pool.Close()

	var role string
	err = pool.QueryRow(ctx, "SELECT current_user").Scan(&role)
	require.NoError(t, err)
	require.Equal(t, "cron_user", role,
		"CRON_DATABASE_URL must authenticate as cron_user post-Phase-6; got %q", role)
}
