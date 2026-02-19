package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/crimson-sun/heimdall/backend/internal/api/handlers"
	"github.com/crimson-sun/heimdall/backend/internal/api/middleware"
	"github.com/crimson-sun/heimdall/backend/internal/config"
)

func NewRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logging)
	r.Use(middleware.CORS)

	r.Route("/api", func(r chi.Router) {
		r.Route("/connections", func(r chi.Router) {
			r.Get("/", handlers.ListConnections)
			r.Post("/", handlers.CreateConnection)
			r.Get("/{id}", handlers.GetConnection)
			r.Put("/{id}", handlers.UpdateConnection)
			r.Delete("/{id}", handlers.DeleteConnection)
		})

		r.Route("/agent", func(r chi.Router) {
			r.Get("/config", handlers.GetAgentConfig)
			r.Put("/config", handlers.UpdateAgentConfig)
		})

		r.Get("/logs", handlers.ListLogs)

		r.Route("/reports", func(r chi.Router) {
			r.Get("/", handlers.ListReports)
			r.Get("/{id}", handlers.GetReport)
		})

		r.Get("/auth/login", handlers.Login)
	})

	r.Get("/ws/chat", handlers.HandleChat)

	return r
}
