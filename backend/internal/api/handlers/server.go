package handlers

import "github.com/hejijunhao/heimdall/backend/internal/config"

// Server holds shared dependencies for all handlers.
type Server struct {
	Config *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{Config: cfg}
}
