package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/handlers"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/connectors"
	"github.com/hejijunhao/heimdall/backend/internal/github"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool, ag *agent.Agent, jwks *middleware.JWKSClient, gh *github.Client, poller *connectors.Poller, listener *connectors.ListenerManager) *chi.Mux {
	s := handlers.NewServer(cfg, pool, ag, jwks, gh, poller, listener)
	r := chi.NewRouter()

	r.Use(middleware.Logging)
	r.Use(middleware.CORS)

	r.Route("/api", func(r chi.Router) {
		// Public ingestion routes — no JWT required, larger body limits.
		// These are outside the global MaxBodySize so they can accept payloads up to their own limits.
		r.Post("/webhooks/logs", s.IngestWebhookLogs)
		r.Post("/webhooks/logs/{format}", s.IngestWebhookLogsWithFormat)
		r.Post("/v1/logs", s.IngestOTLPLogs)
		// GitHub callback is hit by browser redirect from GitHub — auth via state JWT, not session.
		r.Get("/github/callback", s.GitHubCallback)

		// Protected routes — require Supabase JWT, 1MB body limit.
		r.Group(func(r chi.Router) {
			r.Use(middleware.MaxBodySize(1 << 20)) // 1MB for JSON endpoints
			r.Use(middleware.Auth(jwks))

			// Organization & onboarding
			r.Get("/org", s.GetOrganization)
			r.Put("/org", s.UpdateOrganization)
			r.Delete("/org", s.DeleteOrganization)
			r.Post("/onboard", s.Onboard)

			// All orgs the user belongs to (multi-org switcher)
			r.Get("/orgs", s.ListUserOrganizations)
			r.Post("/orgs", s.CreateNewOrganization)

			// Org membership management
			r.Route("/org/members", func(r chi.Router) {
				r.Get("/", s.ListOrgMembers)
				r.Post("/invite", s.InviteMember)
				r.Put("/{userId}/role", s.UpdateMemberRole)
				r.Delete("/{userId}", s.RemoveMember)
			})

			// Applications
			r.Get("/apps", s.ListApplications)
			r.Post("/apps", s.CreateApplication)

			// Per-application routes
			r.Route("/apps/{appId}", func(r chi.Router) {
				r.Get("/", s.GetApplication)
				r.Delete("/", s.DeleteApplication)
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

				// Scheduled investigations (Phase 3)
				r.Get("/schedules", s.ListSchedules)
				r.Post("/schedules", s.CreateSchedule)
				r.Patch("/schedules/{id}", s.UpdateSchedule)
				r.Delete("/schedules/{id}", s.DeleteSchedule)
				r.Post("/schedules/{id}/run", s.RunScheduleNow)
			})

			// GitHub App integration (install requires JWT; callback is public above)
			r.Get("/github/install", s.InstallGitHub)

			// Connections (user-scoped, for create/update/delete/test)
			r.Route("/connections", func(r chi.Router) {
				r.Get("/", s.ListConnections)
				r.Post("/", s.CreateConnection)
				r.Get("/{id}", s.GetConnection)
				r.Put("/{id}", s.UpdateConnection)
				r.Delete("/{id}", s.DeleteConnection)
				r.Post("/{id}/test", s.TestConnection)
				// Source filtering. Generic across every multi-source
				// connector type: webhook_logs with Fly.io detection (Phase 1),
				// org-scoped webhooks (Phase 2), GitHub repos via the discover
				// endpoint (Phase 3). The legacy /github/repos routes were
				// retired when github_repos was migrated away in migration 036.
				r.Get("/{id}/sources", s.ListSourceFilters)
				r.Put("/{id}/sources", s.UpdateSourceFilters)
				r.Post("/{id}/sources", s.AddSourceFilter)
				// Source name passed as ?name=... — see DeleteSourceFilter for why.
				r.Delete("/{id}/sources", s.DeleteSourceFilter)
				// Connector-type-aware discovery — pulls the authoritative
				// source list from upstream (GitHub API today) and upserts
				// every result into connection_sources.
				r.Post("/{id}/sources/discover", s.DiscoverSources)
			})

			r.Get("/logs", s.ListLogs)

			r.Route("/conversations", func(r chi.Router) {
				r.Get("/", s.ListConversations)
				r.Get("/{id}", s.GetConversation)
			})

			r.Get("/auth/me", s.Me)

			// Available models for the agent-config dropdown.
			r.Get("/models", s.GetAvailableModels)
		})
	})

	r.Get("/health", s.Health)
	r.Get("/ws/chat", s.HandleChat)

	// Metrics behind auth so operational data isn't publicly exposed.
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwks))
		r.Handle("/metrics", promhttp.Handler())
	})

	return r
}
