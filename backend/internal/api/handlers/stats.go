package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
)

func (s *Server) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}
	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "internal error", err)
		return
	}
	defer done()

	stats, err := queries.GetDashboardStats(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "failed to fetch stats", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
