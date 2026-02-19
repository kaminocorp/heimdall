package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *Server) ListReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]any{})
}

func (s *Server) GetReport(w http.ResponseWriter, r *http.Request) {
	jsonError(w, "not implemented", http.StatusNotImplemented)
}
