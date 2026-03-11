package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/handlers"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/config"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool, ag *agent.Agent, jwks *middleware.JWKSClient) *chi.Mux {
	s := handlers.NewServer(cfg, pool, ag, jwks)
	r := chi.NewRouter()

	r.Use(middleware.Logging)
	r.Use(middleware.CORS)

	r.Route("/api", func(r chi.Router) {
		// Public routes — token-based auth, no JWT required.
		r.Post("/webhooks/logs", s.IngestWebhookLogs)

		// Protected routes — require Supabase JWT.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwks))

			// Organization & onboarding
			r.Get("/org", s.GetOrganization)
			r.Post("/onboard", s.Onboard)

			// Applications
			r.Get("/apps", s.ListApplications)
			r.Post("/apps", s.CreateApplication)

			// Per-application routes
			r.Route("/apps/{appId}", func(r chi.Router) {
				r.Get("/", s.GetApplication)
				r.Get("/connections", s.ListConnectionsByApp)
				r.Get("/agent/config", s.GetAppAgentConfig)
				r.Put("/agent/config", s.UpdateAppAgentConfig)
				r.Get("/monitoring/status", s.GetMonitoringStatus)
				r.Get("/stats", s.GetAppDashboardStats)

				// Notification management
				r.Get("/notifications/preferences", s.GetNotificationPreferences)
				r.Put("/notifications/preferences", s.UpdateNotificationPreferences)
				r.Get("/notifications/channels", s.ListNotificationChannels)
				r.Post("/notifications/channels", s.CreateNotificationChannel)
				r.Put("/notifications/channels/{channelId}", s.UpdateNotificationChannel)
				r.Delete("/notifications/channels/{channelId}", s.DeleteNotificationChannel)
				r.Post("/notifications/channels/{channelId}/test", s.TestNotificationChannel)
				r.Get("/notifications/history", s.ListNotificationHistory)
			})

			// Connections (user-scoped, for create/update/delete/test)
			r.Route("/connections", func(r chi.Router) {
				r.Get("/", s.ListConnections)
				r.Post("/", s.CreateConnection)
				r.Get("/{id}", s.GetConnection)
				r.Put("/{id}", s.UpdateConnection)
				r.Delete("/{id}", s.DeleteConnection)
				r.Post("/{id}/test", s.TestConnection)
			})

			r.Get("/logs", s.ListLogs)

			r.Route("/conversations", func(r chi.Router) {
				r.Get("/", s.ListConversations)
				r.Get("/{id}", s.GetConversation)
			})

			r.Get("/auth/me", s.Me)

			r.Route("/reports", func(r chi.Router) {
				r.Get("/", s.ListReports)
				r.Get("/{id}", s.GetReport)
			})
		})
	})

	r.Get("/ws/chat", s.HandleChat)

	return r
}
