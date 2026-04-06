package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type webhookLogRequest struct {
	SourceType string          `json:"source_type"`
	Severity   string          `json:"severity"`
	Payload    json.RawMessage `json:"payload"`
}

const maxWebhookRequestBytes = 10 << 20 // 10 MB

func (s *Server) IngestWebhookLogs(w http.ResponseWriter, r *http.Request) {
	// Extract bearer token from Authorization header.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		jsonError(w, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Look up the connection by webhook token.
	conn, err := s.Queries.GetConnectionByWebhookToken(r.Context(), token)
	if err != nil {
		jsonError(w, "invalid webhook token", http.StatusUnauthorized)
		return
	}

	// Read body with size limit.
	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookRequestBytes))
	if err != nil {
		jsonError(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	// Detect format and parse into normalized entries.
	contentType := r.Header.Get("Content-Type")
	entries, err := parseWebhookPayload(body, contentType)
	if err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(entries) == 0 {
		jsonError(w, "empty log payload", http.StatusBadRequest)
		return
	}

	var inserted []db.LogBuffer
	for _, e := range entries {
		if e.SourceType == "" {
			jsonError(w, "source_type is required", http.StatusBadRequest)
			return
		}
		if len(e.Payload) == 0 || string(e.Payload) == "null" {
			jsonError(w, "payload is required", http.StatusBadRequest)
			return
		}

		var severity pgtype.Text
		if e.Severity != "" {
			severity = pgtype.Text{String: e.Severity, Valid: true}
		}

		row, err := s.Queries.InsertLogEntry(r.Context(), db.InsertLogEntryParams{
			ConnectionID: conn.ID,
			SourceType:   e.SourceType,
			Severity:     severity,
			Payload:      e.Payload,
			UserID:       conn.UserID,
		})
		if err != nil {
			jsonServerError(w, "failed to insert log entry", err)
			return
		}
		inserted = append(inserted, row)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if len(inserted) == 1 {
		json.NewEncoder(w).Encode(inserted[0])
	} else {
		json.NewEncoder(w).Encode(inserted)
	}
}
