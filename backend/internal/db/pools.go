package db

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pools holds the two runtime database pools: App for per-tenant work and
// Cron for cross-tenant enumeration. The split exists to support the
// Phase 6 env-var flip from a single `postgres` superuser pool to two
// non-superuser roles (`app_user` for tenant work, `cron_user` for
// enumeration). Until the flip, both fields point at the same pool —
// the topology is in place so call sites can be refactored to the
// handoff pattern without depending on the privilege change.
//
// The two-pool struct is the single chokepoint for "where does a query
// run?" decisions:
//
//   - HTTP handlers and per-tenant background work go through
//     UserQueries / UserQueriesForLoop on App. These mint a
//     transaction-scoped *Queries with app.current_user_id set so RLS
//     policies evaluate against the caller.
//   - Cross-tenant enumeration (monitor's ListActiveApplications,
//     scheduler's ListEnabledSchedules, the log_buffer pruner) goes
//     through CronQueries on Cron. No transaction, no GUC — the cron
//     role's grants are intentionally narrow (SELECT-only on tenant
//     tables, narrow INSERT/UPDATE on monitoring_state, DELETE on
//     log_buffer) so accidental tenant writes from this path fail loudly.
type Pools struct {
	App  *pgxpool.Pool
	Cron *pgxpool.Pool
}

// NewPools wires the two-pool struct. Phase 3 ships `app` and `cron`
// resolving to the same superuser handle; Phase 6 flips the env vars
// so they resolve to distinct roles. Either way, the call shape never
// changes.
func NewPools(app, cron *pgxpool.Pool) *Pools {
	return &Pools{App: app, Cron: cron}
}

// UserQueries begins a transaction on the app pool, sets the
// `app.current_user_id` Postgres session variable for RLS policy
// evaluation, and returns:
//
//   - queries: a *Queries bound to the transaction
//   - commit:  call on the success path to persist writes
//   - done:    defer immediately — rolls back if commit was not called
//
// Read-only callers can ignore commit (assign to _) and just defer done().
// Write callers must call commit() explicitly on the success path before
// encoding the response. This pattern prevents partial commits: if a
// caller returns early on error after a successful first write, done()
// rolls back instead of persisting.
//
// The commit/rollback uses context.WithoutCancel so a client disconnect
// can't strand a half-applied transaction — the txn finalises either
// way once done() runs.
func (p *Pools) UserQueries(ctx context.Context, userID uuid.UUID) (queries *Queries, commit func() error, done func(), err error) {
	return p.userTx(ctx, userID, false)
}

// UserQueriesForLoop is the agent-loop variant of UserQueries. It does
// everything UserQueries does, plus disables Postgres'
// `idle_in_transaction_session_timeout` for the duration of the
// transaction. The agent's tool-use loop holds one transaction across
// up to 10 LLM iterations (Option A from the parent plan), and any
// individual ChatCompletion can take 30+ seconds. Without this, the
// default timeout (often 60s, sometimes lower) would abort the
// transaction mid-loop while the model is thinking.
//
// The trade-off is that one connection from the App pool is pinned for
// the whole loop — typically tens of seconds, occasionally minutes.
// App-pool sizing accounts for this; if connection starvation surfaces
// during the Phase 6 bake, the follow-up is to add a third "loop"
// pool rather than reverting to per-iteration short transactions
// (which would weaken the snapshot guarantee the loop relies on).
func (p *Pools) UserQueriesForLoop(ctx context.Context, userID uuid.UUID) (queries *Queries, commit func() error, done func(), err error) {
	return p.userTx(ctx, userID, true)
}

// CronQueries returns a *Queries bound directly to the cron pool — no
// transaction, no GUC. Used by cross-tenant enumeration paths that
// don't need RLS scoping (the cron role's grants are the safety net).
//
// Callers that need transactional semantics on the cron pool can
// extract Pool.Cron and run their own Begin/Commit; the only path
// today that needs that is the log_buffer pruner, which uses a single
// DELETE statement and doesn't benefit from a wrapping txn.
func (p *Pools) CronQueries() *Queries {
	return New(p.Cron)
}

// WithUserQueries is the closure-shaped sibling of UserQueries: open a
// transaction, set the user-id GUC, run fn against the resulting *Queries,
// commit on nil error, rollback otherwise. Cleaner than the manual
// (commit, done) pattern when the caller has a single linear block of
// work and doesn't need the explicit commit point — typical of the
// background subsystems' per-tenant work.
func (p *Pools) WithUserQueries(ctx context.Context, userID uuid.UUID, fn func(*Queries) error) error {
	q, commit, done, err := p.UserQueries(ctx, userID)
	if err != nil {
		return err
	}
	defer done()
	if err := fn(q); err != nil {
		return err
	}
	return commit()
}

func (p *Pools) userTx(ctx context.Context, userID uuid.UUID, loop bool) (*Queries, func() error, func(), error) {
	// finalCtx is reused for every cleanup path so a client disconnect
	// (which cancels the caller's ctx) cannot strand a half-applied
	// transaction. Both error-path rollbacks below and the deferred
	// done() use it; only the per-statement Exec calls use the caller's
	// ctx so the user-facing request can still cancel the in-flight
	// query.
	finalCtx := context.WithoutCancel(ctx)

	tx, err := p.App.Begin(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID.String()); err != nil {
		tx.Rollback(finalCtx)
		return nil, nil, nil, err
	}

	if loop {
		// Agent-loop transactions span LLM round-trips. Disable both the
		// idle-in-txn timeout (so the gap between tool calls while Claude
		// thinks doesn't trip Postgres) and the per-statement timeout (so
		// the next tool query inside the same txn doesn't trip Supabase's
		// per-role 8s default — which would otherwise abort the next
		// query *after* a long ChatCompletion, not the LLM call itself).
		// SET LOCAL is scoped to the current transaction only.
		if _, err := tx.Exec(ctx, "SET LOCAL idle_in_transaction_session_timeout = 0"); err != nil {
			tx.Rollback(finalCtx)
			return nil, nil, nil, err
		}
		if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = 0"); err != nil {
			tx.Rollback(finalCtx)
			return nil, nil, nil, err
		}
	}

	// committed is captured by both closures and is intentionally not
	// goroutine-safe: the (commit, done) pair is owned by a single
	// caller in a sequential block. Any future fan-out of this pattern
	// across goroutines must add its own synchronisation; the current
	// call sites (chat loop, monitor tick, scheduler tick, webhook
	// ingest) are all single-owner sequential.
	committed := false

	commitFn := func() error {
		committed = true
		return tx.Commit(finalCtx)
	}

	doneFn := func() {
		if !committed {
			if err := tx.Rollback(finalCtx); err != nil {
				slog.Error("transaction rollback failed", "user_id", userID, "err", err)
			}
		}
	}

	return New(tx), commitFn, doneFn, nil
}
