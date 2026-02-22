package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hejijunhao/heimdall/backend/internal/api/handlers"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/config"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool) *chi.Mux {
	s := handlers.NewServer(cfg, pool)
	r := chi.NewRouter()

	r.Use(middleware.Logging)
	r.Use(middleware.CORS)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.SupabaseJWTSecret))

		r.Route("/connections", func(r chi.Router) {
			r.Get("/", s.ListConnections)
			r.Post("/", s.CreateConnection)
			r.Get("/{id}", s.GetConnection)
			r.Put("/{id}", s.UpdateConnection)
			r.Delete("/{id}", s.DeleteConnection)
		})

		r.Route("/agent", func(r chi.Router) {
			r.Get("/config", s.GetAgentConfig)
			r.Put("/config", s.UpdateAgentConfig)
		})

		r.Get("/logs", s.ListLogs)

		r.Get("/auth/me", s.Me)

		r.Route("/reports", func(r chi.Router) {
			r.Get("/", s.ListReports)
			r.Get("/{id}", s.GetReport)
		})
	})

	r.Get("/ws/chat", s.HandleChat)

	return r
}
