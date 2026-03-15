package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// jsonServerError logs the internal error for debugging, then sends a generic message to the client.
func jsonServerError(w http.ResponseWriter, message string, err error) {
	slog.Error(message, "error", err)
	jsonError(w, message, http.StatusInternalServerError)
}
