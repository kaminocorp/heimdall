package handlers

import (
	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server holds shared dependencies for all handlers.
type Server struct {
	Config  *config.Config
	Queries *db.Queries
	Agent   *agent.Agent
}

func NewServer(cfg *config.Config, pool *pgxpool.Pool, ag *agent.Agent) *Server {
	return &Server{
		Config:  cfg,
		Queries: db.New(pool),
		Agent:   ag,
	}
}
