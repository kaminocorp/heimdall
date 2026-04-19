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

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

type webhookLogRequest struct {
	SourceType string          `json:"source_type"`
	Severity   string          `json:"severity"`
	Payload    json.RawMessage `json:"payload"`
	// SourceName is the filter-key extracted by the parser (e.g. fly.app.name).
	// Parsers only populate this when the natural filter key differs from
	// SourceType — ingestion falls back to SourceType via sourceNameOf() when
	// it's empty.
	SourceName string `json:"-"`
}

// sourceNameOf returns the filter key the ingestion handler uses for
// (connection_sources / app_source_filters) lookups. Parsers that extract a
// meaningful per-source identifier (Fly.io → fly.app.name) populate
// SourceName directly; others leave it empty and fall back to SourceType.
func sourceNameOf(e webhookLogRequest) string {
	if e.SourceName != "" {
		return e.SourceName
	}
	return e.SourceType
}

// applySourceNamePathOverride reads an optional `source_name_path` setting
// from the connection config and, if present, extracts a source name from
// each entry's payload via dotted-path traversal. Overrides SourceName on
// every entry regardless of what the parser set — this is a deliberate
// "user knows better" escape hatch. Parsers already extract the best
// guess; the config override exists for the cases they can't handle
// (AWS Firehose log groups, GCP Pub/Sub attributes, etc.).
//
// Syntax: `field.subfield.deeper`. No array indexing, no wildcards —
// pragmatic dotted-getter rather than full JSONPath. If a segment is
// missing or the final value isn't a string, SourceName is left unchanged
// so the pre-existing parser value / SourceType fallback still applies.
func applySourceNamePathOverride(entries []webhookLogRequest, configRaw json.RawMessage) {
	path := readSourceNamePath(configRaw)
	if path == "" {
		return
	}
	segments := strings.Split(path, ".")
	for i := range entries {
		if name := extractStringByPath(entries[i].Payload, segments); name != "" {
			entries[i].SourceName = name
		}
	}
}

// readSourceNamePath returns the `source_name_path` field from a connection
// config, or "" if absent / malformed. Uses a bespoke shallow unmarshal
// rather than a shared config struct because every connector type has a
// different config shape and we only want this one field.
func readSourceNamePath(configRaw json.RawMessage) string {
	if len(configRaw) == 0 {
		return ""
	}
	var probe struct {
		SourceNamePath string `json:"source_name_path"`
	}
	if err := json.Unmarshal(configRaw, &probe); err != nil {
		return ""
	}
	return strings.TrimSpace(probe.SourceNamePath)
}

// extractStringByPath walks a JSON object by key segments. Returns the
// string at the leaf if the path resolves cleanly, "" otherwise. Silently
// tolerates non-object intermediates and type mismatches — those become
// "path didn't resolve" rather than errors.
func extractStringByPath(payload json.RawMessage, segments []string) string {
	if len(payload) == 0 || len(segments) == 0 {
		return ""
	}
	var cur any
	if err := json.Unmarshal(payload, &cur); err != nil {
		return ""
	}
	for _, seg := range segments {
		obj, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = obj[seg]
	}
	s, _ := cur.(string)
	return s
}

// webhookLogResponse is the minimal acknowledgement returned on successful ingestion.
// It intentionally omits internal identifiers (user_id, app_id, connection_id).
//
// Accepted counts entries that passed source filtering and were inserted.
// Filtered counts entries dropped because their source was disabled in
// app_source_filters (or no filter row existed — drop-by-default).
type webhookLogResponse struct {
	Accepted   int    `json:"accepted"`
	Filtered   int    `json:"filtered,omitempty"`
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
		s.ingestEntries(w, r, conn, entries, format)
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

	s.ingestEntries(w, r, conn, entries, format)
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

	s.ingestEntries(w, r, conn, entries, parsedFormat)
}

// connResult holds the fields needed from the connection lookup to avoid
// passing the full db.Connection type through the shared helpers.
//
// AppID is nullable: app-scoped connections (Phase 1 default) carry the
// target app directly, whereas org-scoped connections (Phase 2) leave it
// nil and rely on fan-out via app_source_filters at ingestion time.
//
// Config carries the raw connection config so Phase 4's source_name_path
// override can consult it without a second DB round-trip.
type connResult struct {
	ID     uuid.UUID
	UserID uuid.UUID
	OrgID  uuid.UUID
	AppID  *uuid.UUID
	Config json.RawMessage
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

	return connResult{
		ID:     conn.ID,
		UserID: conn.UserID,
		OrgID:  conn.OrgID,
		AppID:  conn.AppID,
		Config: conn.Config,
	}, body, true
}

// ingestEntries validates, inserts (transactionally), and responds. This is the
// shared tail called by both auto-detect and format-specific handlers.
//
// If the caller sends an X-Idempotency-Key header, the response is cached for
// 24 hours. Subsequent requests with the same key return the cached response
// without re-inserting.
//
// App-scoped vs org-scoped (Phase 2):
//   - conn.AppID != nil   → app-scoped. Each entry is stored at most once
//     into the connection's app, filtered by app_source_filters for that app.
//   - conn.AppID == nil   → org-scoped. Each entry fans out to every app in
//     the connection's org that has enabled the entry's source_name. Fan-out
//     is N×M inserts for a batch of N entries and M apps per source — the
//     design deliberately pays this write cost to keep reads (activity feed
//     queries) simple: no filter join on the read path.
func (s *Server) ingestEntries(w http.ResponseWriter, r *http.Request, conn connResult, entries []webhookLogRequest, format string) {
	start := time.Now()
	connID := conn.ID
	userID := conn.UserID

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

	// Phase 4 source_name_path: if the connection's config specifies a
	// dotted JSON path, override SourceName for every entry whose parser
	// didn't already set one. This lets users ingest payloads whose
	// natural source identifier is buried in a nested field rather than
	// being the flat SourceType — e.g. AWS Firehose forwarded from multiple
	// CloudWatch log groups, where the log group name is inside the
	// payload.
	applySourceNamePathOverride(entries, conn.Config)

	// Run the shared filter pipeline: discover every source in the batch,
	// resolve which apps want each source (app-scoped → one app; org-scoped
	// → fan-out), then insert the matching entries. Whole thing is atomic
	// under this transaction.
	routes, seenSources, err := routeSources(r.Context(), qtx, connID, conn.AppID, entries)
	if err != nil {
		jsonServerError(w, "source filter pipeline failed", err)
		return
	}
	stats, inserted, err := insertFiltered(r.Context(), qtx, connID, userID, entries, routes, seenSources)
	if err != nil {
		jsonServerError(w, "failed to insert log entry", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		jsonServerError(w, "failed to commit transaction", err)
		return
	}

	// Post-commit: emit Pipeline-page ingestion events for every newly
	// inserted log row. Deliberately after commit so the FK from
	// log_pipeline_events.log_id resolves; pre-commit it would race with
	// concurrent reads. Fire-and-forget per call — writer logs its own
	// failures and never bubbles them back.
	if s.Agent != nil {
		if pw := s.Agent.Pipeline(); pw != nil {
			for _, ins := range inserted {
				pw.WriteIngestion(r.Context(), agent.IngestionInput{
					LogID:      ins.LogID,
					AppID:      ins.AppID,
					SourceType: ins.SourceType,
					Severity:   ins.Severity,
				})
			}
		}
	}

	scope := "app"
	if conn.AppID == nil {
		scope = "org"
	}
	elapsed := time.Since(start)
	slog.Info("webhook ingestion",
		"connection_id", connID,
		"scope", scope,
		"format", format,
		"entries", len(entries),
		"accepted", stats.Accepted,
		"filtered", stats.Filtered,
		"inserts", stats.Inserts,
		"sources", stats.Sources,
		"duration_ms", elapsed.Milliseconds(),
	)
	accepted := stats.Accepted
	filtered := stats.Filtered

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

	// Build the response body. Accepted is the post-filter count; filtered is
	// the number of entries whose source wasn't enabled. Existing callers that
	// only read `accepted` continue to work — filtered is purely additive and
	// omitted from the JSON when zero.
	resp := webhookLogResponse{
		Accepted:   accepted,
		Filtered:   filtered,
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
