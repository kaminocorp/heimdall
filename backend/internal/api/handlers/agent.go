package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

var agentConfigDefaults = map[string]any{
	"model": "claude-sonnet-4-5-20250929",
	"mode":  "continuous",
}

func (s *Server) GetAgentConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.Queries.GetAgentConfig(r.Context())
	if err != nil {
		// No row in DB yet — return defaults.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agentConfigDefaults)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

type updateAgentConfigRequest struct {
	Model                string `json:"model"`
	Mode                 string `json:"mode"`
	Schedule             string `json:"schedule"`
	SystemPromptOverride string `json:"system_prompt_override"`
}

func (s *Server) UpdateAgentConfig(w http.ResponseWriter, r *http.Request) {
	var req updateAgentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = "claude-sonnet-4-5-20250929"
	}
	if req.Mode == "" {
		req.Mode = "continuous"
	}

	cfg, err := s.Queries.UpsertAgentConfig(r.Context(), db.UpsertAgentConfigParams{
		Model: req.Model,
		Mode:  req.Mode,
		Schedule: pgtype.Text{
			String: req.Schedule,
			Valid:  req.Schedule != "",
		},
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

type runAgentRequest struct {
	Input string `json:"input"`
}

type runAgentResponse struct {
	Response string `json:"response"`
}

func (s *Server) RunAgent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	var req runAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Input == "" {
		jsonError(w, "input is required", http.StatusBadRequest)
		return
	}

	response, err := s.Agent.RunLoop(r.Context(), userID, req.Input)
	if err != nil {
		jsonError(w, "agent error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runAgentResponse{Response: response})
}
