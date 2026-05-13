package handlers

import (
	"encoding/json"
	"net/http"
)

// Health checks whether the server is up and both database pools are
// reachable. Returns 200 {"status":"ok"} or 503 with a per-pool field
// indicating which pool is down. Pinging both is load-bearing post-Phase-6:
// the cron pool authenticates as a different role with potentially
// different connectivity (different DSN, different PgBouncer client tier),
// so a cron-pool outage left only the LB blind under the old single-ping
// shape.
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := s.Pools.App.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"db":     "app_pool_unreachable",
		})
		return
	}
	if err := s.Pools.Cron.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"db":     "cron_pool_unreachable",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
