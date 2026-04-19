package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/connectors"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/github"
)

// Server holds shared dependencies for all handlers.
type Server struct {
	Config   *config.Config
	Pool     *pgxpool.Pool
	Queries  *db.Queries
	Agent    *agent.Agent
	JWKS     *middleware.JWKSClient
	GitHub   *github.Client
	Poller   *connectors.Poller
	Listener *connectors.ListenerManager

	// Phase 3 hardening — lazy-allocated so tests that build Server{} as
	// a struct literal (testhelpers_test.go) don't need to wire them.
	// The pipeline handlers check for nil and fall back to an uncapped /
	// uncached code path when these are unset.
	pipelineSSELimiter *sseConnLimiter
	pipelineBootstrap  *bootstrapCache
}

// SetPipelineBootstrapCacheForTest wires a fresh bootstrap cache onto a
// struct-literal Server. Used by the handler integration tests which
// construct Server{} directly (bypassing NewServer) and need the cache
// paths to exercise in isolation.
func (s *Server) SetPipelineBootstrapCacheForTest() {
	s.pipelineBootstrap = newBootstrapCache()
}

func NewServer(cfg *config.Config, pool *pgxpool.Pool, ag *agent.Agent, jwks *middleware.JWKSClient, gh *github.Client, poller *connectors.Poller, listener *connectors.ListenerManager) *Server {
	return &Server{
		Config:             cfg,
		Pool:               pool,
		Queries:            db.New(pool),
		Agent:              ag,
		JWKS:               jwks,
		GitHub:             gh,
		Poller:             poller,
		Listener:           listener,
		pipelineSSELimiter: newSSEConnLimiter(pipelineSSEPerUserCap),
		pipelineBootstrap:  newBootstrapCache(),
	}
}
