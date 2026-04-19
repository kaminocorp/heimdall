package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
)

// OTLP JSON types — subset of the OpenTelemetry Protocol ExportLogsServiceRequest.
// See: https://opentelemetry.io/docs/specs/otlp/#otlphttp-request

type otlpExportRequest struct {
	ResourceLogs []otlpResourceLogs `json:"resourceLogs"`
}

type otlpResourceLogs struct {
	Resource  otlpResource    `json:"resource"`
	ScopeLogs []otlpScopeLogs `json:"scopeLogs"`
}

type otlpResource struct {
	Attributes []otlpKeyValue `json:"attributes"`
}

type otlpScopeLogs struct {
	Scope      otlpScope       `json:"scope"`
	LogRecords []otlpLogRecord `json:"logRecords"`
}

type otlpScope struct {
	Name string `json:"name"`
}

type otlpLogRecord struct {
	TimeUnixNano   string         `json:"timeUnixNano"`
	SeverityText   string         `json:"severityText"`
	SeverityNumber int            `json:"severityNumber"`
	Body           otlpAnyValue   `json:"body"`
	Attributes     []otlpKeyValue `json:"attributes"`
	TraceID        string         `json:"traceId"`
	SpanID         string         `json:"spanId"`
}

type otlpKeyValue struct {
	Key   string       `json:"key"`
	Value otlpAnyValue `json:"value"`
}

type otlpAnyValue struct {
	StringValue string          `json:"stringValue,omitempty"`
	IntValue    string          `json:"intValue,omitempty"`
	BoolValue   bool            `json:"boolValue,omitempty"`
	ArrayValue  json.RawMessage `json:"arrayValue,omitempty"`
	KvlistValue json.RawMessage `json:"kvlistValue,omitempty"`
}

const maxOTLPRequestBytes = 10 << 20 // 10 MB

// IngestOTLPLogs handles POST /v1/logs — the OTLP HTTP JSON endpoint.
// Auth uses the same bearer token mechanism as webhook ingestion.
//
// Phase 4: flattened OTLP records are converted into webhookLogRequest
// entries and run through the shared routeSources / insertFiltered pipeline.
// This unifies source filtering (service.name per record) and adds fan-out
// support for org-scoped OTLP connections — the same plumbing webhook
// ingestion uses.
func (s *Server) IngestOTLPLogs(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		jsonError(w, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	conn, err := s.Queries.GetConnectionByWebhookToken(r.Context(), token)
	if err != nil {
		jsonError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxOTLPRequestBytes))
	if err != nil {
		jsonError(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	var req otlpExportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		jsonError(w, "invalid OTLP JSON payload", http.StatusBadRequest)
		return
	}

	// Flatten the nested OTLP structure into webhookLogRequest entries. Each
	// log record becomes one entry whose SourceName is the enclosing
	// resource's service.name attribute — which is how OTLP identifies which
	// emitting service produced the record. Records without a service.name
	// fall back to the generic "otlp" source via sourceNameOf().
	entries := flattenOTLPRecords(req)
	totalRecords := len(entries)

	if totalRecords == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"accepted": 0})
		return
	}

	tx, err := s.Pool.Begin(r.Context())
	if err != nil {
		jsonServerError(w, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback(r.Context())

	qtx := s.Queries.WithTx(tx)

	routes, seenSources, err := routeSources(r.Context(), qtx, conn.ID, conn.AppID, entries)
	if err != nil {
		jsonServerError(w, "source filter pipeline failed", err)
		return
	}
	stats, inserted, err := insertFiltered(r.Context(), qtx, conn.ID, conn.UserID, entries, routes, seenSources)
	if err != nil {
		jsonServerError(w, "failed to insert log entry", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		jsonServerError(w, "failed to commit transaction", err)
		return
	}

	// Pipeline-page ingestion emits — see the parallel block in
	// ingestEntries (webhooks.go) for why this is post-commit.
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
	slog.Info("otlp ingestion",
		"connection_id", conn.ID,
		"scope", scope,
		"entries", totalRecords,
		"accepted", stats.Accepted,
		"filtered", stats.Filtered,
		"inserts", stats.Inserts,
		"sources", stats.Sources,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"accepted": stats.Accepted,
		"filtered": stats.Filtered,
	})
}

// flattenOTLPRecords turns a nested OTLP payload into a flat slice of
// webhookLogRequest entries, one per log record. SourceName is populated
// from the enclosing resource's service.name attribute.
func flattenOTLPRecords(req otlpExportRequest) []webhookLogRequest {
	var out []webhookLogRequest
	for _, rl := range req.ResourceLogs {
		serviceName := extractServiceName(rl.Resource.Attributes)
		resourceAttrs := flattenAttributes(rl.Resource.Attributes)

		// SourceType carries the human-readable "otlp/<service>" label used
		// in the activity feed; SourceName is the bare service identifier
		// used as the filter key.
		sourceType := "otlp"
		if serviceName != "" {
			sourceType = "otlp/" + serviceName
		}

		for _, sl := range rl.ScopeLogs {
			for _, lr := range sl.LogRecords {
				payload, err := json.Marshal(map[string]any{
					"time_unix_nano":  lr.TimeUnixNano,
					"severity_text":   lr.SeverityText,
					"severity_number": lr.SeverityNumber,
					"body":            resolveAnyValue(lr.Body),
					"attributes":      flattenAttributes(lr.Attributes),
					"resource":        resourceAttrs,
					"scope":           sl.Scope.Name,
					"trace_id":        lr.TraceID,
					"span_id":         lr.SpanID,
				})
				if err != nil {
					continue
				}

				// mapOTLPSeverity returns pgtype.Text; webhookLogRequest
				// wants a plain string that gets wrapped later by
				// insertFiltered. Unwrap here.
				sev := mapOTLPSeverity(lr.SeverityNumber, lr.SeverityText)
				var severityStr string
				if sev.Valid {
					severityStr = sev.String
				}

				out = append(out, webhookLogRequest{
					SourceType: sourceType,
					Severity:   severityStr,
					Payload:    payload,
					SourceName: serviceName, // may be empty → sourceNameOf falls back to SourceType
				})
			}
		}
	}
	return out
}

// extractServiceName finds the "service.name" attribute from OTel resource attributes.
func extractServiceName(attrs []otlpKeyValue) string {
	for _, a := range attrs {
		if a.Key == "service.name" && a.Value.StringValue != "" {
			return a.Value.StringValue
		}
	}
	return ""
}

// mapOTLPSeverity converts OTLP severity to a Heimdall severity string.
// OTLP severity numbers: 1-4=Trace, 5-8=Debug, 9-12=Info, 13-16=Warn, 17-20=Error, 21-24=Fatal.
func mapOTLPSeverity(num int, text string) pgtype.Text {
	var s string
	switch {
	case num >= 21:
		s = "critical"
	case num >= 17:
		s = "error"
	case num >= 13:
		s = "warning"
	case num >= 9:
		s = "info"
	case num >= 5:
		s = "debug"
	case num >= 1:
		s = "debug"
	default:
		// Fall back to text if number is unset.
		switch strings.ToUpper(text) {
		case "FATAL", "FATAL2", "FATAL3", "FATAL4":
			s = "critical"
		case "ERROR", "ERROR2", "ERROR3", "ERROR4":
			s = "error"
		case "WARN", "WARN2", "WARN3", "WARN4":
			s = "warning"
		case "INFO", "INFO2", "INFO3", "INFO4":
			s = "info"
		case "DEBUG", "DEBUG2", "DEBUG3", "DEBUG4":
			s = "debug"
		default:
			s = "info"
		}
	}
	return pgtype.Text{String: s, Valid: true}
}

// resolveAnyValue extracts the value from an OTLP AnyValue.
func resolveAnyValue(v otlpAnyValue) string {
	if v.StringValue != "" {
		return v.StringValue
	}
	if v.IntValue != "" {
		return v.IntValue
	}
	// BoolValue can't be distinguished from unset (Go zero value is false),
	// so we only emit "true" when explicitly set. This is an acceptable
	// trade-off — bool attributes are rare in OTLP log records.
	if v.BoolValue {
		return "true"
	}
	return ""
}

// flattenAttributes converts OTLP key-value attributes to a simple map.
func flattenAttributes(attrs []otlpKeyValue) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		m[a.Key] = resolveAnyValue(a.Value)
	}
	return m
}

// countLogRecords totals records across all scopes in an OTLP request.
// Retained from the pre-Phase-4 handler because otlp_test.go's
// TestCountLogRecords still exercises it — pure helper with no production
// consumer now that the new pipeline derives counts from len(entries), but
// cheap to keep.
func countLogRecords(req otlpExportRequest) int {
	var n int
	for _, rl := range req.ResourceLogs {
		for _, sl := range rl.ScopeLogs {
			n += len(sl.LogRecords)
		}
	}
	return n
}

