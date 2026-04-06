package handlers

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// parseWebhookPayload detects the payload format and normalizes it to a slice
// of webhookLogRequest entries. Supports:
//   - Heimdall native format (single object or array)
//   - Vercel NDJSON (newline-delimited JSON from Log Drains)
//   - AWS Kinesis Firehose HTTP delivery (base64-encoded records)
//   - GCP Pub/Sub push subscription (base64-encoded message data)
func parseWebhookPayload(raw []byte, contentType string) ([]webhookLogRequest, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, nil
	}

	// Try Vercel NDJSON: Content-Type is application/x-ndjson, or multiple lines each starting with {.
	if strings.Contains(contentType, "ndjson") || isNDJSON(trimmed) {
		return parseVercelNDJSON(raw)
	}

	// Try to parse as JSON first.
	var probe json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, err
	}

	// Detect AWS Firehose: structurally check for requestId + records fields.
	if isFirehosePayload(raw) {
		return parseFirehosePayload(raw)
	}

	// Detect GCP Pub/Sub: structurally check for message.data + subscription fields.
	if isPubSubPayload(raw) {
		return parsePubSubPayload(raw)
	}

	// Default: Heimdall native format.
	return parseNativePayload(raw)
}

// --- Heimdall native format ---

func parseNativePayload(raw []byte) ([]webhookLogRequest, error) {
	trimmed := strings.TrimSpace(string(raw))

	if len(trimmed) > 0 && trimmed[0] == '[' {
		var entries []webhookLogRequest
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, err
		}
		return entries, nil
	}

	var single webhookLogRequest
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, err
	}
	return []webhookLogRequest{single}, nil
}

// --- Vercel NDJSON ---

// Vercel Log Drain entry format.
type vercelLogEntry struct {
	ID            string `json:"id"`
	Message       string `json:"message"`
	Timestamp     int64  `json:"timestamp"`
	Source        string `json:"source"`        // "build", "lambda", "edge", "static"
	ProjectName   string `json:"projectName"`
	DeploymentURL string `json:"deploymentUrl"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	StatusCode    int    `json:"statusCode"`
	Level         string `json:"level"` // "info", "warning", "error"
	Proxy         *struct {
		Timestamp int64  `json:"timestamp"`
		Path      string `json:"path"`
		Host      string `json:"host"`
		Method    string `json:"method"`
		Scheme    string `json:"scheme"`
		StatusCode int   `json:"statusCode"`
	} `json:"proxy"`
}

func isNDJSON(s string) bool {
	lines := strings.SplitN(s, "\n", 3)
	if len(lines) < 2 {
		return false
	}
	// At least 2 lines, each starting with '{'.
	for _, line := range lines[:2] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "{") {
			return false
		}
	}
	return true
}

func parseVercelNDJSON(raw []byte) ([]webhookLogRequest, error) {
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	var entries []webhookLogRequest

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry vercelLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip malformed lines rather than failing the whole batch.
			continue
		}

		severity := mapVercelLevel(entry.Level)

		payload, _ := json.Marshal(map[string]any{
			"message":        entry.Message,
			"timestamp":      entry.Timestamp,
			"source":         entry.Source,
			"project_name":   entry.ProjectName,
			"deployment_url": entry.DeploymentURL,
			"host":           entry.Host,
			"path":           entry.Path,
			"status_code":    entry.StatusCode,
		})

		sourceType := "vercel"
		if entry.Source != "" {
			sourceType = "vercel/" + entry.Source
		}

		entries = append(entries, webhookLogRequest{
			SourceType: sourceType,
			Severity:   severity,
			Payload:    payload,
		})
	}

	return entries, nil
}

func mapVercelLevel(level string) string {
	switch strings.ToLower(level) {
	case "error":
		return "error"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}

// --- AWS Kinesis Firehose ---

type firehoseRequest struct {
	RequestID string           `json:"requestId"`
	Timestamp int64            `json:"timestamp"`
	Records   []firehoseRecord `json:"records"`
}

type firehoseRecord struct {
	Data string `json:"data"` // base64-encoded
}

func isFirehosePayload(raw []byte) bool {
	var probe struct {
		RequestID string          `json:"requestId"`
		Records   json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return probe.RequestID != "" && len(probe.Records) > 0
}

func parseFirehosePayload(raw []byte) ([]webhookLogRequest, error) {
	var req firehoseRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}

	var entries []webhookLogRequest
	for _, record := range req.Records {
		decoded, err := base64.StdEncoding.DecodeString(record.Data)
		if err != nil {
			continue
		}

		// Each record's data may be a single JSON object or NDJSON.
		lines := strings.Split(strings.TrimSpace(string(decoded)), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Try to parse as structured JSON.
			var obj map[string]any
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				// Treat as raw text.
				payload, _ := json.Marshal(map[string]any{"message": line})
				entries = append(entries, webhookLogRequest{
					SourceType: "firehose",
					Severity:   "info",
					Payload:    payload,
				})
				continue
			}

			payload, _ := json.Marshal(obj)
			severity := extractSeverityFromMap(obj)

			entries = append(entries, webhookLogRequest{
				SourceType: "firehose",
				Severity:   severity,
				Payload:    payload,
			})
		}
	}

	return entries, nil
}

// --- GCP Pub/Sub ---

type pubsubPushRequest struct {
	Message struct {
		Data       string            `json:"data"` // base64-encoded
		Attributes map[string]string `json:"attributes"`
		MessageID  string            `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func isPubSubPayload(raw []byte) bool {
	var probe struct {
		Message struct {
			Data string `json:"data"`
		} `json:"message"`
		Subscription string `json:"subscription"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	// Require both a non-empty subscription and message.data to distinguish from
	// Heimdall native payloads that happen to contain "message"/"subscription" keys.
	return probe.Subscription != "" && probe.Message.Data != ""
}

func parsePubSubPayload(raw []byte) ([]webhookLogRequest, error) {
	var req pubsubPushRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(req.Message.Data)
	if err != nil {
		// Data might not be base64 — try using it raw.
		decoded = []byte(req.Message.Data)
	}

	// The decoded data may be a JSON object, an array, or plain text.
	trimmed := strings.TrimSpace(string(decoded))

	// Try as JSON array of log entries.
	if strings.HasPrefix(trimmed, "[") {
		var entries []webhookLogRequest
		if err := json.Unmarshal(decoded, &entries); err == nil && len(entries) > 0 {
			for i := range entries {
				if entries[i].SourceType == "" {
					entries[i].SourceType = "pubsub"
				}
			}
			return entries, nil
		}
	}

	// Try as single JSON object.
	if strings.HasPrefix(trimmed, "{") {
		var obj map[string]any
		if err := json.Unmarshal(decoded, &obj); err == nil {
			payload, _ := json.Marshal(obj)
			severity := extractSeverityFromMap(obj)
			return []webhookLogRequest{{
				SourceType: "pubsub",
				Severity:   severity,
				Payload:    payload,
			}}, nil
		}
	}

	// Fallback: treat as raw text.
	payload, _ := json.Marshal(map[string]any{
		"message":     trimmed,
		"message_id":  req.Message.MessageID,
		"attributes":  req.Message.Attributes,
		"subscription": req.Subscription,
	})
	return []webhookLogRequest{{
		SourceType: "pubsub",
		Severity:   "info",
		Payload:    payload,
	}}, nil
}

// --- Shared helpers ---

// extractSeverityFromMap looks for common severity/level fields in a JSON object.
func extractSeverityFromMap(obj map[string]any) string {
	for _, key := range []string{"severity", "level", "log_level", "error_severity"} {
		if v, ok := obj[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return normalizeSeverity(s)
			}
		}
	}
	return "info"
}

func normalizeSeverity(s string) string {
	switch strings.ToLower(s) {
	case "emergency", "fatal", "critical", "crit":
		return "critical"
	case "error", "err":
		return "error"
	case "warning", "warn":
		return "warning"
	case "notice", "info", "informational":
		return "info"
	case "debug", "trace":
		return "debug"
	default:
		return "info"
	}
}
