package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: implement actual authentication
	json.NewEncoder(w).Encode(map[string]any{
		"token": "placeholder",
	})
}
