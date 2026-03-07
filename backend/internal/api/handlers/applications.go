package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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
		jsonError(w, "failed to list applications", http.StatusInternalServerError)
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
		jsonError(w, "failed to create application", http.StatusInternalServerError)
		return
	}

	// Create default agent config
	_, err = s.Queries.UpsertAppAgentConfig(r.Context(), db.UpsertAppAgentConfigParams{
		AppID:                app.ID,
		Model:                "claude-sonnet-4-6",
		Mode:                 "off",
		ScheduleIntervalSecs: 60,
	})
	if err != nil {
		jsonError(w, "failed to create agent config", http.StatusInternalServerError)
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
			"model":                  "claude-sonnet-4-6",
			"mode":                   "off",
			"schedule_interval_secs": 60,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

type updateAppAgentConfigRequest struct {
	Model                string `json:"model"`
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
		req.Model = "claude-sonnet-4-6"
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

	cfg, err := s.Queries.UpsertAppAgentConfig(r.Context(), db.UpsertAppAgentConfigParams{
		AppID:                app.ID,
		Model:                req.Model,
		Mode:                 req.Mode,
		ScheduleIntervalSecs: req.ScheduleIntervalSecs,
		SystemPromptOverride: pgtype.Text{
			String: req.SystemPromptOverride,
			Valid:  req.SystemPromptOverride != "",
		},
	})
	if err != nil {
		jsonError(w, "failed to update agent config", http.StatusInternalServerError)
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
		jsonError(w, "failed to get stats", http.StatusInternalServerError)
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
		jsonError(w, "failed to list connections", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}
