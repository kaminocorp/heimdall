package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// parseByFormat parses the payload using a specific parser, bypassing auto-detection.
// Used by format-specific routes (POST /api/webhooks/logs/{format}) and the
// X-Heimdall-Format header override.
func parseByFormat(raw []byte, format string) ([]webhookLogRequest, string, error) {
	switch format {
	case "flyio":
		entries, err := parseVectorFlyPayload(raw)
		return entries, "flyio_vector", err
	case "vercel":
		entries, err := parseVercelNDJSON(raw)
		return entries, "vercel_ndjson", err
	case "firehose":
		entries, err := parseFirehosePayload(raw)
		return entries, "aws_firehose", err
	case "pubsub":
		entries, err := parsePubSubPayload(raw)
		return entries, "gcp_pubsub", err
	case "native":
		entries, err := parseNativePayload(raw)
		if err != nil {
			return nil, "", err
		}
		return entries, nativeFormatString(raw, entries), nil
	default:
		return nil, "", fmt.Errorf("unknown format: %s", format)
	}
}

// parseWebhookPayload detects the payload format and normalizes it to a slice
// of webhookLogRequest entries. The returned format string identifies which
// parser handled the request (e.g. "native", "native_batch", "vercel_ndjson").
//
// Supports:
//   - Heimdall native format (single object or array)
//   - Vercel NDJSON (newline-delimited JSON from Log Drains)
//   - AWS Kinesis Firehose HTTP delivery (base64-encoded records)
//   - GCP Pub/Sub push subscription (base64-encoded message data)
//   - Fly.io Vector HTTP sink (Fly Log Shipper)
func parseWebhookPayload(raw []byte, contentType string) ([]webhookLogRequest, string, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, "", nil
	}

	// Vercel NDJSON: require explicit Content-Type header. Structural detection
	// (isNDJSON) was removed — it matched any multi-line JSON input, causing
	// false positives with Fly.io and other batch formats. Callers without the
	// ndjson content-type should use POST /api/webhooks/logs/vercel instead.
	if strings.Contains(contentType, "ndjson") {
		entries, err := parseVercelNDJSON(raw)
		return entries, "vercel_ndjson", err
	}

	// Try to parse as JSON first.
	var probe json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, "", err
	}

	// Detect AWS Firehose: structurally check for requestId + records fields.
	if isFirehosePayload(raw) {
		entries, err := parseFirehosePayload(raw)
		return entries, "aws_firehose", err
	}

	// Detect GCP Pub/Sub: structurally check for message.data + subscription fields.
	if isPubSubPayload(raw) {
		entries, err := parsePubSubPayload(raw)
		return entries, "gcp_pubsub", err
	}

	// Detect Vector HTTP sink (Fly Log Shipper): check for fly metadata or Vector source_type.
	if isVectorFlyPayload(raw) {
		entries, err := parseVectorFlyPayload(raw)
		return entries, "flyio_vector", err
	}

	// Default: Heimdall native format (v1 or v2).
	entries, err := parseNativePayload(raw)
	if err != nil {
		return nil, "", err
	}
	format := nativeFormatString(raw, entries)
	return entries, format, nil
}

// nativeFormatString determines the format string for a parsed native payload.
// It probes the raw input to detect v2 and checks entry count for batch suffix.
func nativeFormatString(raw []byte, entries []webhookLogRequest) string {
	trimmed := strings.TrimSpace(string(raw))
	// For batches, probe the first element; for singles, probe directly.
	var probeTarget []byte
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var items []json.RawMessage
		if json.Unmarshal(raw, &items) == nil && len(items) > 0 {
			probeTarget = items[0]
		}
	} else {
		probeTarget = raw
	}

	v2 := isNativeV2(probeTarget)
	batch := len(entries) > 1

	switch {
	case v2 && batch:
		return "native_v2_batch"
	case v2:
		return "native_v2"
	case batch:
		return "native_batch"
	default:
		return "native"
	}
}

// --- Heimdall native format (v1 + v2) ---

// webhookLogRequestV2 is the v2 native format with clearer field names.
// See the plan for rationale: source replaces source_type, level replaces
// severity, message is first-class, attrs replaces payload.
type webhookLogRequestV2 struct {
	Source  string          `json:"source"`
	Level   string          `json:"level"`
	Message string          `json:"message"`
	Attrs   json.RawMessage `json:"attrs"`
}

// nativeVersionProbe peeks at a single JSON object to determine whether it
// uses v1 field names (source_type) or v2 field names (source).
type nativeVersionProbe struct {
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
}

// parseNativePayload accepts both v1 and v2 native formats. It probes the
// first entry to determine the version, then parses the entire payload
// accordingly. Mixed v1+v2 batches are rejected.
func parseNativePayload(raw []byte) ([]webhookLogRequest, error) {
	trimmed := strings.TrimSpace(string(raw))

	if len(trimmed) > 0 && trimmed[0] == '[' {
		return parseNativeBatch(raw)
	}

	return parseNativeSingle(raw)
}

func parseNativeSingle(raw []byte) ([]webhookLogRequest, error) {
	if isNativeV2(raw) {
		entry, err := parseNativeV2Entry(raw)
		if err != nil {
			return nil, err
		}
		return []webhookLogRequest{entry}, nil
	}

	var single webhookLogRequest
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, err
	}
	return []webhookLogRequest{single}, nil
}

func parseNativeBatch(raw []byte) ([]webhookLogRequest, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	// Probe the first entry to determine the batch version.
	batchIsV2 := isNativeV2(items[0])

	var entries []webhookLogRequest
	for i, item := range items {
		itemIsV2 := isNativeV2(item)
		if itemIsV2 != batchIsV2 {
			return nil, fmt.Errorf("entry %d: mixed v1/v2 formats in batch — all entries must use the same format", i)
		}

		if itemIsV2 {
			entry, err := parseNativeV2Entry(item)
			if err != nil {
				return nil, fmt.Errorf("entry %d: %w", i, err)
			}
			entries = append(entries, entry)
		} else {
			var entry webhookLogRequest
			if err := json.Unmarshal(item, &entry); err != nil {
				return nil, fmt.Errorf("entry %d: %w", i, err)
			}
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// isNativeV2 returns true if the JSON object uses v2 field names.
// If both "source" and "source_type" are present, v2 takes precedence.
func isNativeV2(raw []byte) bool {
	var probe nativeVersionProbe
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return probe.Source != ""
}

// parseNativeV2Entry normalises a v2 entry into the internal v1 representation.
func parseNativeV2Entry(raw []byte) (webhookLogRequest, error) {
	var v2 webhookLogRequestV2
	if err := json.Unmarshal(raw, &v2); err != nil {
		return webhookLogRequest{}, err
	}

	if v2.Source == "" {
		return webhookLogRequest{}, fmt.Errorf("source is required")
	}

	// Build the stored payload by merging message into attrs.
	attrs := make(map[string]interface{})
	if len(v2.Attrs) > 0 && string(v2.Attrs) != "null" {
		if err := json.Unmarshal(v2.Attrs, &attrs); err != nil {
			return webhookLogRequest{}, fmt.Errorf("attrs must be a JSON object: %w", err)
		}
	}
	if v2.Message != "" {
		attrs["message"] = v2.Message
	}

	payload, err := json.Marshal(attrs)
	if err != nil {
		return webhookLogRequest{}, err
	}

	severity := "info"
	if v2.Level != "" {
		severity = normalizeSeverity(v2.Level)
	}

	return webhookLogRequest{
		SourceType: v2.Source,
		Severity:   severity,
		Payload:    payload,
	}, nil
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
		RequestID string `json:"requestId"`
		Records   []struct {
			Data string `json:"data"`
		} `json:"records"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	// Require requestId, at least one record, and the first record must have a
	// non-empty data field. This eliminates false positives from app logs that
	// happen to contain "requestId" + "records" fields.
	return probe.RequestID != "" && len(probe.Records) > 0 && probe.Records[0].Data != ""
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

// --- Vector HTTP sink (Fly Log Shipper) ---

// isVectorFlyPayload detects payloads sent by Vector's HTTP sink, specifically
// from the Fly Log Shipper. Detection works on both single objects and arrays.
//
// The Fly Log Shipper adds a nested "fly" object with app metadata. For generic
// Vector payloads, we also match if source_type starts with "fly".
func isVectorFlyPayload(raw []byte) bool {
	trimmed := strings.TrimSpace(string(raw))

	// For arrays, inspect the first element.
	if strings.HasPrefix(trimmed, "[") {
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
			return false
		}
		return isVectorFlyObject(arr[0])
	}

	return isVectorFlyObject(raw)
}

// knownFlySourceTypes is the set of source_type values that indicate a Fly.io
// origin. Uses exact matching instead of prefix to avoid false positives from
// unrelated source types like "flywheel" or "flutter".
var knownFlySourceTypes = map[string]bool{
	"fly_app":          true,
	"fly_io":           true,
	"fly_log_shipper":  true,
	"fly_app_logs":     true,
}

func isVectorFlyObject(raw []byte) bool {
	var probe struct {
		Fly        *json.RawMessage `json:"fly"`
		SourceType string           `json:"source_type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	// Match if there's a "fly" nested object (Fly Log Shipper) or
	// source_type is a known Fly.io value.
	return probe.Fly != nil || knownFlySourceTypes[probe.SourceType]
}

// vectorFlyEntry represents a log line from the Fly Log Shipper (Vector HTTP sink).
type vectorFlyEntry struct {
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
	Host       string `json:"host"`
	SourceType string `json:"source_type"`
	Fly        *struct {
		App struct {
			Name string `json:"name"`
		} `json:"app"`
		Machine struct {
			ID string `json:"id"`
		} `json:"machine"`
		Region string `json:"region"`
	} `json:"fly"`
	Log *struct {
		Level string `json:"level"`
	} `json:"log"`
}

func parseVectorFlyPayload(raw []byte) ([]webhookLogRequest, error) {
	trimmed := strings.TrimSpace(string(raw))

	// Handle array payloads (Vector batch mode).
	if strings.HasPrefix(trimmed, "[") {
		var entries []json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, err
		}
		var results []webhookLogRequest
		for _, entryRaw := range entries {
			parsed, err := parseVectorFlyEntry(entryRaw)
			if err != nil {
				continue // skip malformed entries
			}
			results = append(results, parsed)
		}
		return results, nil
	}

	// Single object.
	parsed, err := parseVectorFlyEntry(raw)
	if err != nil {
		return nil, err
	}
	return []webhookLogRequest{parsed}, nil
}

func parseVectorFlyEntry(raw []byte) (webhookLogRequest, error) {
	var entry vectorFlyEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return webhookLogRequest{}, err
	}

	// Build source type from Fly metadata. The per-app source name (used as
	// the filter key in app_source_filters) is the bare app name — "trajan"
	// rather than "flyio/trajan" — so users see their Fly app names unchanged
	// in the source selector. Falls back to source_type when fly metadata is
	// absent (non-shipper Vector senders).
	sourceType := "flyio"
	sourceName := ""
	if entry.Fly != nil && entry.Fly.App.Name != "" {
		sourceType = "flyio/" + entry.Fly.App.Name
		sourceName = entry.Fly.App.Name
	} else if entry.SourceType != "" {
		sourceType = entry.SourceType
	}

	// Extract severity from log.level or fall back to generic extraction.
	severity := "info"
	if entry.Log != nil && entry.Log.Level != "" {
		severity = normalizeSeverity(entry.Log.Level)
	} else {
		// Try generic extraction from the raw JSON.
		var obj map[string]any
		if err := json.Unmarshal(raw, &obj); err == nil {
			severity = extractSeverityFromMap(obj)
		}
	}

	// Build enriched payload preserving all original fields plus structured metadata.
	payload := make(map[string]any)
	payload["message"] = entry.Message
	payload["timestamp"] = entry.Timestamp
	payload["host"] = entry.Host

	if entry.Fly != nil {
		payload["app_name"] = entry.Fly.App.Name
		payload["machine_id"] = entry.Fly.Machine.ID
		payload["region"] = entry.Fly.Region
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return webhookLogRequest{}, err
	}

	return webhookLogRequest{
		SourceType: sourceType,
		Severity:   severity,
		Payload:    payloadJSON,
		SourceName: sourceName,
	}, nil
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
