package handlers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// UserQueries begins a transaction, sets the Postgres session variable
// app.current_user_id for RLS policy evaluation, and returns a *db.Queries
// scoped to that transaction. The caller must defer done() to finalize the
// transaction and release the connection back to the pool.
//
// done() calls Commit. This is safe because PostgreSQL guarantees that if any
// statement within the transaction has failed, the transaction enters an aborted
// state and COMMIT automatically behaves as ROLLBACK. Each handler performs at
// most one write, so partial-commit is not possible.
func (s *Server) UserQueries(ctx context.Context, userID uuid.UUID) (queries *db.Queries, done func(), err error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID.String()); err != nil {
		tx.Rollback(ctx)
		return nil, nil, err
	}

	// Use context.WithoutCancel so that a client disconnect does not cancel
	// the commit — the write has already succeeded and must be persisted.
	commitCtx := context.WithoutCancel(ctx)
	return s.Queries.WithTx(tx), func() {
		if err := tx.Commit(commitCtx); err != nil {
			slog.Error("transaction commit failed", "user_id", userID, "err", err)
		}
	}, nil
}
