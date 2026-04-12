package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (s *Server) ListApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	org, err := s.Queries.GetOrganizationByUser(r.Context(), userID)
	if err != nil {
		jsonError(w, "organization not found", http.StatusNotFound)
		return
	}

	// Opt-in counts mode. The Settings page uses `?include=counts` to render
	// per-app connection/schedule summaries in a single request. Default
	// callers (sidebar selector, etc.) get the lean shape to keep the common
	// path cheap and avoid breaking existing frontends.
	if r.URL.Query().Get("include") == "counts" {
		rows, err := s.Queries.ListApplicationsByOrgWithCounts(r.Context(), org.ID)
		if err != nil {
			jsonServerError(w, "failed to list applications", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
		return
	}

	apps, err := s.Queries.ListApplicationsByOrg(r.Context(), org.ID)
	if err != nil {
		jsonServerError(w, "failed to list applications", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

type createApplicationRequest struct {
	Name string `json:"name"`
}

func (s *Server) CreateApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	org, err := s.Queries.GetOrganizationByUser(r.Context(), userID)
	if err != nil {
		jsonError(w, "organization not found", http.StatusNotFound)
		return
	}

	var req createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	app, err := s.Queries.CreateApplication(r.Context(), db.CreateApplicationParams{
		OrgID:  org.ID,
		Name:   req.Name,
		Status: "active",
	})
	if err != nil {
		jsonServerError(w, "failed to create application", err)
		return
	}

	// Create default agent config
	_, err = s.Queries.UpsertAppAgentConfig(r.Context(), db.UpsertAppAgentConfigParams{
		AppID:                app.ID,
		Model:                agent.DefaultModelID,
		Mode:                 "off",
		ScheduleIntervalSecs: 60,
		Provider:             "anthropic",
	})
	if err != nil {
		jsonServerError(w, "failed to create agent config", err)
		return
	}

	// Fire-and-forget activity-feed entry. The agent_log table is keyed by
	// user_id (not app_id), so we stash the app identity in the detail
	// JSONB column — the row survives app deletion and remains meaningful
	// as audit history. Nil-check matches the pattern used elsewhere: the
	// test harness leaves s.Agent nil to avoid pulling in the full agent
	// stack, and we don't want handler coverage to blow up for that reason.
	if s.Agent != nil {
		s.Agent.EmitLog(
			r.Context(), userID, nil,
			"application_created",
			fmt.Sprintf("Application %q created", app.Name),
			map[string]any{
				"app_id":   app.ID.String(),
				"app_name": app.Name,
			},
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// DeleteApplication removes an application and cascade-deletes every row
// that references it: connections, app_agent_config, monitoring_state,
// notification_*, investigation_schedules. The cascade relationships are
// defined in the FK constraints (migrations 014–019, 023), so this handler
// is intentionally thin — it only enforces authorization and the last-app
// guard. If a future migration drops a cascade, the handler-level cascade
// test will fail loudly rather than silently leaking orphaned rows.
func (s *Server) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "appId"))
	if err != nil {
		jsonError(w, "invalid application id", http.StatusBadRequest)
		return
	}

	// Org-scope authorization: resolve the app through the user→org→app
	// join so a caller can't delete an app in a different org. 404 instead of
	// 403 matches the pattern used by every other /api/apps/{appId}/* route —
	// we don't leak existence of resources the caller can't see.
	app, err := s.Queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "application not found", http.StatusNotFound)
		return
	}

	// Last-app guard. Deleting the only remaining app would leave the user in
	// a first-run state with an empty sidebar selector and no valid currentAppId.
	// We block it here rather than allowing-and-redirecting so the UX stays
	// predictable. The `code` field is what the frontend keys off to render
	// the specific "create another app first" helper line instead of a
	// generic toast.
	count, err := s.Queries.CountApplicationsByOrg(r.Context(), app.OrgID)
	if err != nil {
		jsonServerError(w, "failed to count applications", err)
		return
	}
	if count <= 1 {
		jsonErrorWithCode(w, "cannot delete last application", "last_app", http.StatusConflict)
		return
	}

	if err := s.Queries.DeleteApplication(r.Context(), app.ID); err != nil {
		jsonServerError(w, "failed to delete application", err)
		return
	}

	// Fire-and-forget activity-feed entry. Emitted *after* the delete
	// succeeds so we never record a deletion that didn't happen. The row
	// is user-scoped, not app-scoped, so it survives the cascade delete
	// of everything under the app — which is the whole point: the audit
	// entry must outlive the thing it audits.
	if s.Agent != nil {
		s.Agent.EmitLog(
			r.Context(), userID, nil,
			"application_deleted",
			fmt.Sprintf("Application %q deleted", app.Name),
			map[string]any{
				"app_id":   app.ID.String(),
				"app_name": app.Name,
			},
		)
	}

	w.WriteHeader(http.StatusNoContent)
}

// authorizeApp verifies the app belongs to the authenticated user's org.
// Returns the application on success, writes an HTTP error and returns nil on failure.
func (s *Server) authorizeApp(w http.ResponseWriter, r *http.Request) *db.Application {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return nil
	}

	appID, err := uuid.Parse(chi.URLParam(r, "appId"))
	if err != nil {
		jsonError(w, "invalid application id", http.StatusBadRequest)
		return nil
	}

	app, err := s.Queries.GetApplicationByOrgUser(r.Context(), db.GetApplicationByOrgUserParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "application not found", http.StatusNotFound)
		return nil
	}

	return &app
}

func (s *Server) GetApplication(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

func (s *Server) GetAppAgentConfig(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	cfg, err := s.Queries.GetAppAgentConfig(r.Context(), app.ID)
	if err != nil {
		// No config row yet — return defaults
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"app_id":                 app.ID,
			"model":                  agent.DefaultModelID,
			"mode":                   "off",
			"schedule_interval_secs": 60,
			"provider":               "anthropic",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

type updateAppAgentConfigRequest struct {
	Model                string `json:"model"`
	Provider             string `json:"provider"`
	Mode                 string `json:"mode"`
	ScheduleIntervalSecs int32  `json:"schedule_interval_secs"`
	SystemPromptOverride string `json:"system_prompt_override"`
}

func (s *Server) UpdateAppAgentConfig(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	var req updateAppAgentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = agent.DefaultModelID
	}
	if req.Mode == "" {
		req.Mode = "off"
	}
	switch req.Mode {
	case "continuous", "periodic", "off":
		// valid
	default:
		jsonError(w, "mode must be one of: continuous, periodic, off", http.StatusBadRequest)
		return
	}
	if req.ScheduleIntervalSecs <= 0 {
		req.ScheduleIntervalSecs = 60
	}

	// Provider validation. Empty defaults to "anthropic" for backwards compat
	// with clients that pre-date the dropdown. "openrouter" is only accepted
	// when OPENROUTER_API_KEY is configured on the platform — otherwise we
	// reject at the API boundary instead of silently falling back, so the
	// user sees a clear error rather than wondering why their model choice
	// is being ignored.
	// Normalise before validating so "OpenRouter", " openrouter ", etc. all
	// resolve to the canonical lower-case form rather than falling through to
	// the default "unknown provider" error.
	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	if req.Provider == "" {
		req.Provider = "anthropic"
	}
	switch req.Provider {
	case "anthropic":
		// always available
	case "openrouter":
		if s.Config.OpenRouterKey == "" {
			jsonError(w, "openrouter provider not enabled on this server", http.StatusBadRequest)
			return
		}
	default:
		jsonError(w, "provider must be one of: anthropic, openrouter", http.StatusBadRequest)
		return
	}

	// Model validation. The catalogue (models.go) is the single source of
	// truth — any ID not in it is rejected so typos surface immediately
	// instead of silently failing at the first agent invocation.
	resolved, ok := agent.ResolveModel(req.Model)
	if !ok {
		jsonError(w, fmt.Sprintf("unknown model: %s", req.Model), http.StatusBadRequest)
		return
	}
	// Cross-check: the model's provider must match the request's provider.
	// Catches mismatches like {provider: "anthropic", model: "openai/gpt-5.4"}.
	if resolved.Provider != req.Provider {
		jsonError(w, fmt.Sprintf("model %s belongs to provider %q, not %q", req.Model, resolved.Provider, req.Provider), http.StatusBadRequest)
		return
	}

	cfg, err := s.Queries.UpsertAppAgentConfig(r.Context(), db.UpsertAppAgentConfigParams{
		AppID:                app.ID,
		Model:                req.Model,
		Provider:             req.Provider,
		Mode:                 req.Mode,
		ScheduleIntervalSecs: req.ScheduleIntervalSecs,
		SystemPromptOverride: pgtype.Text{
			String: req.SystemPromptOverride,
			Valid:  req.SystemPromptOverride != "",
		},
	})
	if err != nil {
		jsonServerError(w, "failed to update agent config", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func (s *Server) GetMonitoringStatus(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	type monitoringStatus struct {
		Mode                 string  `json:"mode"`
		ScheduleIntervalSecs int32   `json:"schedule_interval_secs"`
		LastMonitoredAt      *string `json:"last_monitored_at"`
	}

	cfg, err := s.Queries.GetAppAgentConfig(r.Context(), app.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(monitoringStatus{Mode: "off", ScheduleIntervalSecs: 60})
		return
	}

	status := monitoringStatus{
		Mode:                 cfg.Mode,
		ScheduleIntervalSecs: cfg.ScheduleIntervalSecs,
	}

	state, err := s.Queries.GetMonitoringState(r.Context(), app.ID)
	if err == nil {
		ts := state.LastMonitoredAt.Format("2006-01-02T15:04:05Z07:00")
		status.LastMonitoredAt = &ts
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) GetAppDashboardStats(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	stats, err := s.Queries.GetAppDashboardStats(r.Context(), app.ID)
	if err != nil {
		jsonServerError(w, "failed to get stats", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) ListConnectionsByApp(w http.ResponseWriter, r *http.Request) {
	app := s.authorizeApp(w, r)
	if app == nil {
		return
	}

	connections, err := s.Queries.ListConnectionsByApp(r.Context(), app.ID)
	if err != nil {
		jsonServerError(w, "failed to list connections", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}
