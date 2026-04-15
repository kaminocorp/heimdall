package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
)

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// jsonErrorWithCode is like jsonError but attaches a stable machine-readable
// `code` field alongside the human-readable message. Use this for guarded
// failures the frontend needs to distinguish from generic errors — e.g. the
// last-app delete guard, where the UI renders a specific helper text rather
// than a generic toast.
func jsonErrorWithCode(w http.ResponseWriter, message, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
		"code":  code,
	})
}

// isValidEmail performs a basic format check using Go's net/mail parser.
func isValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	// Reject display names like "Name <email@x.com>" — require bare address.
	return addr.Address == email && strings.Contains(email, ".")
}

// jsonServerError logs the internal error for debugging, then sends a generic message to the client.
func jsonServerError(w http.ResponseWriter, message string, err error) {
	slog.Error(message, "error", err)
	jsonError(w, message, http.StatusInternalServerError)
}
