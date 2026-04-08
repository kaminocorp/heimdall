package handlers

import (
	"encoding/json"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
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
