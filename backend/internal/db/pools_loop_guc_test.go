package db

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// These tests are in the internal `db` package (not `db_test`) so they can
// reach the unexported `Queries.db` field for raw SHOW commands inside the
// transaction the helpers build.
//
// The Phase-10 H3 invariant they pin: UserQueriesForLoop disables both
// statement_timeout and idle_in_transaction_session_timeout for the
// duration of the txn; UserQueries leaves both at the role/session
// default. A regression that swapped the loop-branch flag would compile
// silently and either (a) stall every short HTTP handler indefinitely
// (loop semantics applied to non-loop) or (b) abort the agent loop at
// Supabase's 8s default (non-loop semantics applied to loop). Neither
// shape is currently covered by behavioural tests; these close that gap.

// TestUserQueriesForLoop_DisablesTimeouts pins the positive assertion: the
// loop variant SETs both relevant timeouts to 0.
func TestUserQueriesForLoop_DisablesTimeouts(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set; skipping loop-tx GUC test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	pools := NewPools(pool, pool)
	q, _, done, err := pools.UserQueriesForLoop(ctx, uuid.New())
	require.NoError(t, err)
	defer done()

	var stmt, idle string
	require.NoError(t, q.db.QueryRow(ctx, "SHOW statement_timeout").Scan(&stmt))
	require.NoError(t, q.db.QueryRow(ctx, "SHOW idle_in_transaction_session_timeout").Scan(&idle))

	require.Equal(t, "0", stmt,
		"UserQueriesForLoop must SET LOCAL statement_timeout = 0 so a slow tool query inside a loop txn doesn't trip the per-role default")
	require.Equal(t, "0", idle,
		"UserQueriesForLoop must SET LOCAL idle_in_transaction_session_timeout = 0 so the gap between tool calls while Claude thinks doesn't abort the txn")
}

// TestUserQueries_DoesNotTouchTimeouts pins the no-bleed assertion: the
// non-loop variant must not SET LOCAL either timeout. Implemented by
// comparing values inside the txn against the session default observed on
// a fresh pool query — if the non-loop path issued SET LOCAL = 0 on a
// pool whose role default is non-zero (e.g. Supabase's 8s), the txn-side
// read would diverge.
//
// On a dev Postgres install where both defaults are already "0" this is
// a tautology (both readings are "0"); on Supabase it's a hard signal.
// Either way a regression in the opposite direction (UserQueries wrongly
// applying SET LOCAL) is caught.
func TestUserQueries_DoesNotTouchTimeouts(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set; skipping loop-tx GUC test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	var sessionStmt, sessionIdle string
	require.NoError(t, pool.QueryRow(ctx, "SHOW statement_timeout").Scan(&sessionStmt))
	require.NoError(t, pool.QueryRow(ctx, "SHOW idle_in_transaction_session_timeout").Scan(&sessionIdle))

	pools := NewPools(pool, pool)
	q, _, done, err := pools.UserQueries(ctx, uuid.New())
	require.NoError(t, err)
	defer done()

	var txStmt, txIdle string
	require.NoError(t, q.db.QueryRow(ctx, "SHOW statement_timeout").Scan(&txStmt))
	require.NoError(t, q.db.QueryRow(ctx, "SHOW idle_in_transaction_session_timeout").Scan(&txIdle))

	require.Equal(t, sessionStmt, txStmt,
		"UserQueries must not SET LOCAL statement_timeout — that's loop-only (parent plan §5.3b)")
	require.Equal(t, sessionIdle, txIdle,
		"UserQueries must not SET LOCAL idle_in_transaction_session_timeout — that's loop-only (parent plan §5.3b)")
}
