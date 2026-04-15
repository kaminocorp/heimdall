package handlers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// UserQueries begins a transaction, sets the Postgres session variable
// app.current_user_id for RLS policy evaluation, and returns:
//   - queries: a *db.Queries scoped to the transaction
//   - commit:  call on the success path to persist writes
//   - done:    defer immediately — rolls back if commit was not called
//
// Read-only handlers can ignore commit (use _ for the return value) and just
// defer done(). Write handlers must call commit() explicitly at the end of
// the success path before encoding the response.
//
// This pattern prevents partial commits: if a handler returns early on error
// after a successful first write, done() rolls back instead of persisting.
func (s *Server) UserQueries(ctx context.Context, userID uuid.UUID) (queries *db.Queries, commit func() error, done func(), err error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID.String()); err != nil {
		tx.Rollback(ctx)
		return nil, nil, nil, err
	}

	// Use context.WithoutCancel so that a client disconnect does not cancel
	// the commit/rollback — we must finalize either way.
	finalCtx := context.WithoutCancel(ctx)
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

	return s.Queries.WithTx(tx), commitFn, doneFn, nil
}
