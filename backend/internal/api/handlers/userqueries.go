package handlers

import (
	"context"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// UserQueries begins a transaction, sets the Postgres session variable
// app.current_user_id for RLS policy evaluation, and returns a *db.Queries
// scoped to that transaction. The caller must defer done() to commit the
// transaction and release the connection back to the pool.
func (s *Server) UserQueries(ctx context.Context, userID uuid.UUID) (queries *db.Queries, done func(), err error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID.String()); err != nil {
		tx.Rollback(ctx)
		return nil, nil, err
	}

	return s.Queries.WithTx(tx), func() { tx.Commit(ctx) }, nil
}
