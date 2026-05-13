package handlers

import (
	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/connectors"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/github"
)

// Server holds shared dependencies for all handlers.
//
// Pools is the only DB entry point — handlers go through Pools.UserQueries
// (per-request, RLS-scoped) or Pools.CronQueries (cross-tenant, BYPASSRLS).
// The transitional Pool/Queries shims that lived here through Phase 7 were
// removed in Phase 9: every handler now routes through the chokepoint, so
// the pairing tripwire (rls_pairing_test.go) no longer needs an exception
// for raw pool access in this package.
type Server struct {
	Config   *config.Config
	Pools    *db.Pools
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

func NewServer(cfg *config.Config, pools *db.Pools, ag *agent.Agent, jwks *middleware.JWKSClient, gh *github.Client, poller *connectors.Poller, listener *connectors.ListenerManager) *Server {
	return &Server{
		Config:             cfg,
		Pools:              pools,
		Agent:              ag,
		JWKS:               jwks,
		GitHub:             gh,
		Poller:             poller,
		Listener:           listener,
		pipelineSSELimiter: newSSEConnLimiter(pipelineSSEPerUserCap),
		pipelineBootstrap:  newBootstrapCache(),
	}
}
