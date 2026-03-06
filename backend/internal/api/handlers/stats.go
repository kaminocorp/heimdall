package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
)

func (s *Server) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	queries, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer done()

	stats, err := queries.GetDashboardStats(r.Context(), userID)
	if err != nil {
		jsonError(w, "failed to fetch stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
