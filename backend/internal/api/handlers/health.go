package handlers

import (
	"encoding/json"
	"net/http"
)

// Health checks whether the server is up and the database is reachable.
// Returns 200 {"status":"ok"} or 503 {"status":"error","db":"unreachable"}.
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := s.Pool.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"db":     "unreachable",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
