package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *Server) GetAgentConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"model": "claude-sonnet-4-6",
		"mode":  "continuous",
	})
}

func (s *Server) UpdateAgentConfig(w http.ResponseWriter, r *http.Request) {
	jsonError(w, "not implemented", http.StatusNotImplemented)
}
