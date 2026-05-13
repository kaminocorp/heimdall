package handlers

import (
	"context"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// UserQueries is a thin forwarder onto Pools.UserQueries — the canonical
// per-tenant transaction helper now lives on the db.Pools struct so
// background subsystems (agent, connectors, notifications) can share the
// exact same handoff shape as HTTP handlers. Handlers continue to call
// s.UserQueries(...) for convenience.
//
// See db.Pools.UserQueries for the contract.
func (s *Server) UserQueries(ctx context.Context, userID uuid.UUID) (queries *db.Queries, commit func() error, done func(), err error) {
	return s.Pools.UserQueries(ctx, userID)
}
