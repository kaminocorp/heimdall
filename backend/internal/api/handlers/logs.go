package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *Server) ListLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]any{})
}
