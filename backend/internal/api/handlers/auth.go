package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
)

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	queries, _, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	user, err := queries.GetUser(r.Context(), userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
