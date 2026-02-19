package handlers

import (
	"encoding/json"
	"net/http"
)

func GetAgentConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"model": "claude-sonnet-4-6",
		"mode":  "continuous",
	})
}

func UpdateAgentConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
