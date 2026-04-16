package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type webhookLogRequest struct {
	SourceType string          `json:"source_type"`
	Severity   string          `json:"severity"`
	Payload    json.RawMessage `json:"payload"`
}

// webhookLogResponse is the minimal acknowledgement returned on successful ingestion.
// It intentionally omits internal identifiers (user_id, app_id, connection_id).
type webhookLogResponse struct {
	Accepted   int    `json:"accepted"`
	Format     string `json:"format"`
	Deprecated bool   `json:"deprecated,omitempty"`
}

// webhookError is a structured error response for the webhook ingestion endpoint.
// It provides machine-readable codes, the detected format, and per-entry field
// errors so that automated callers can programmatically handle failures.
type webhookError struct {
	Error          string              `json:"error"`
	Message        string              `json:"message"`
	FormatDetected string              `json:"format_detected,omitempty"`
	Details        []webhookFieldError `json:"details,omitempty"`
}

// webhookFieldError identifies a specific validation failure within a batch entry.
type webhookFieldError struct {
	Index int    `json:"index"`
	Field string `json:"field"`
	Error string `json:"error"`
}

// jsonWebhookError writes a structured webhook error response.
func jsonWebhookError(w http.ResponseWriter, code, message, format string, status int, details []webhookFieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(webhookError{
		Error:          code,
		Message:        message,
		FormatDetected: format,
		Details:        details,
	})
}

// validFormats is the set of format slugs accepted by format-specific routes
// and the X-Heimdall-Format header.
var validFormats = map[string]bool{
	"flyio":    true,
	"vercel":   true,
	"firehose": true,
	"pubsub":   true,
	"native":   true,
}

const maxWebhookRequestBytes = 10 << 20 // 10 MB

// IngestWebhookLogs handles POST /api/webhooks/logs — the auto-detecting ingestion endpoint.
// If the X-Heimdall-Format header is set, it bypasses auto-detection and uses the specified parser.
func (s *Server) IngestWebhookLogs(w http.ResponseWriter, r *http.Request) {
	conn, body, ok := s.readWebhookRequest(w, r)
	if !ok {
		return
	}

	// Check for explicit format override via header.
	if headerFormat := r.Header.Get("X-Heimdall-Format"); headerFormat != "" {
		headerFormat = strings.ToLower(strings.TrimSpace(headerFormat))
		if !validFormats[headerFormat] {
			jsonWebhookError(w, "invalid_format",
				"unknown format: \""+headerFormat+"\"; valid formats: flyio, vercel, firehose, pubsub, native",
				"", http.StatusBadRequest, nil)
			return
		}
		entries, format, err := parseByFormat(body, headerFormat)
		if err != nil {
			slog.Warn("webhook parse error", "connection_id", conn.ID, "format", headerFormat, "error", err)
			jsonWebhookError(w, "parse_failed", "invalid request body: "+err.Error(), headerFormat, http.StatusBadRequest, nil)
			return
		}
		s.ingestEntries(w, r, conn.ID, conn.UserID, conn.AppID, entries, format)
		return
	}

	// Auto-detect format.
	contentType := r.Header.Get("Content-Type")
	entries, format, err := parseWebhookPayload(body, contentType)
	if err != nil {
		slog.Warn("webhook parse error", "connection_id", conn.ID, "error", err)
		jsonWebhookError(w, "parse_failed", "invalid request body: "+err.Error(), format, http.StatusBadRequest, nil)
		return
	}

	s.ingestEntries(w, r, conn.ID, conn.UserID, conn.AppID, entries, format)
}

// IngestWebhookLogsWithFormat handles POST /api/webhooks/logs/{format} —
// format-specific routes that bypass auto-detection entirely.
func (s *Server) IngestWebhookLogsWithFormat(w http.ResponseWriter, r *http.Request) {
	format := chi.URLParam(r, "format")
	if !validFormats[format] {
		jsonWebhookError(w, "invalid_format",
			"unknown format: \""+format+"\"; valid formats: flyio, vercel, firehose, pubsub, native",
			"", http.StatusNotFound, nil)
		return
	}

	conn, body, ok := s.readWebhookRequest(w, r)
	if !ok {
		return
	}

	entries, parsedFormat, err := parseByFormat(body, format)
	if err != nil {
		slog.Warn("webhook parse error", "connection_id", conn.ID, "format", format, "error", err)
		jsonWebhookError(w, "parse_failed", "invalid request body: "+err.Error(), format, http.StatusBadRequest, nil)
		return
	}

	s.ingestEntries(w, r, conn.ID, conn.UserID, conn.AppID, entries, parsedFormat)
}

// connResult holds the fields needed from the connection lookup to avoid
// passing the full db.Connection type through the shared helpers.
type connResult struct {
	ID     uuid.UUID
	UserID uuid.UUID
	AppID  uuid.UUID
}

// readWebhookRequest handles auth and body reading — the common preamble shared
// by both auto-detect and format-specific handlers. Returns false if an error
// response was already written.
func (s *Server) readWebhookRequest(w http.ResponseWriter, r *http.Request) (connResult, []byte, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		jsonWebhookError(w, "unauthorized", "missing or invalid authorization header", "", http.StatusUnauthorized, nil)
		return connResult{}, nil, false
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	conn, err := s.Queries.GetConnectionByWebhookToken(r.Context(), token)
	if err != nil {
		jsonWebhookError(w, "unauthorized", "invalid webhook token", "", http.StatusUnauthorized, nil)
		return connResult{}, nil, false
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookRequestBytes))
	if err != nil {
		jsonWebhookError(w, "invalid_json", "failed to read request body", "", http.StatusBadRequest, nil)
		return connResult{}, nil, false
	}

	return connResult{ID: conn.ID, UserID: conn.UserID, AppID: conn.AppID}, body, true
}

// ingestEntries validates, inserts (transactionally), and responds. This is the
// shared tail called by both auto-detect and format-specific handlers.
//
// If the caller sends an X-Idempotency-Key header, the response is cached for
// 24 hours. Subsequent requests with the same key return the cached response
// without re-inserting.
func (s *Server) ingestEntries(w http.ResponseWriter, r *http.Request, connID, userID, appID uuid.UUID, entries []webhookLogRequest, format string) {
	start := time.Now()

	// --- Idempotency: check for cached response ---
	idempotencyKey := strings.TrimSpace(r.Header.Get("X-Idempotency-Key"))
	if idempotencyKey != "" {
		cached, err := s.Queries.GetIdempotencyResult(r.Context(), db.GetIdempotencyResultParams{
			ConnectionID:   connID,
			IdempotencyKey: idempotencyKey,
		})
		if err == nil {
			// Cache hit — return the stored response without re-inserting.
			slog.Info("webhook idempotency hit",
				"connection_id", connID,
				"idempotency_key", idempotencyKey,
			)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotency-Replay", "true")
			w.WriteHeader(int(cached.ResponseStatus))
			w.Write(cached.ResponseBody)
			return
		}
		// err != nil means no cached result (pgx.ErrNoRows) — proceed normally.
	}

	if len(entries) == 0 {
		jsonWebhookError(w, "empty_payload", "no log entries found after parsing", format, http.StatusBadRequest, nil)
		return
	}

	// --- Upfront validation pass: check ALL entries, collect all errors ---
	var fieldErrors []webhookFieldError
	for i, e := range entries {
		if e.SourceType == "" {
			fieldErrors = append(fieldErrors, webhookFieldError{Index: i, Field: "source_type", Error: "required"})
		}
		if len(e.Payload) == 0 || string(e.Payload) == "null" {
			fieldErrors = append(fieldErrors, webhookFieldError{Index: i, Field: "payload", Error: "required"})
		}
	}
	if len(fieldErrors) > 0 {
		slog.Warn("webhook validation failed", "connection_id", connID, "format", format, "errors", len(fieldErrors))
		jsonWebhookError(w, "validation_failed", "one or more entries failed validation", format, http.StatusBadRequest, fieldErrors)
		return
	}

	// --- Transactional insert: all entries succeed or none are persisted ---
	tx, err := s.Pool.Begin(r.Context())
	if err != nil {
		jsonServerError(w, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback(r.Context())

	qtx := s.Queries.WithTx(tx)
	for _, e := range entries {
		var severity pgtype.Text
		if e.Severity != "" {
			severity = pgtype.Text{String: e.Severity, Valid: true}
		}

		_, err := qtx.InsertLogEntry(r.Context(), db.InsertLogEntryParams{
			ConnectionID: connID,
			SourceType:   e.SourceType,
			Severity:     severity,
			Payload:      e.Payload,
			UserID:       userID,
			AppID:        appID,
		})
		if err != nil {
			jsonServerError(w, "failed to insert log entry", err)
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		jsonServerError(w, "failed to commit transaction", err)
		return
	}

	elapsed := time.Since(start)
	slog.Info("webhook ingestion",
		"connection_id", connID,
		"format", format,
		"entries", len(entries),
		"duration_ms", elapsed.Milliseconds(),
	)

	// Deprecation signal for v1 native format (RFC 8594 Sunset Header).
	deprecated := format == "native" || format == "native_batch"
	if deprecated {
		slog.Warn("webhook v1 native format used (deprecated)",
			"connection_id", connID,
			"entries", len(entries),
		)
		w.Header().Set("Sunset", "2026-10-01")
		w.Header().Set("Deprecation", "true")
		w.Header().Set("Link", `<https://docs.heimdallwatch.com/api/webhook-format-v2>; rel="successor-version"`)
	}

	// Build the response body.
	resp := webhookLogResponse{
		Accepted:   len(entries),
		Format:     format,
		Deprecated: deprecated,
	}
	respBody, _ := json.Marshal(resp)

	// --- Idempotency: cache the response for future replays ---
	if idempotencyKey != "" {
		_ = s.Queries.InsertIdempotencyResult(r.Context(), db.InsertIdempotencyResultParams{
			ConnectionID:   connID,
			IdempotencyKey: idempotencyKey,
			ResponseStatus: int32(http.StatusCreated),
			ResponseBody:   respBody,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}
