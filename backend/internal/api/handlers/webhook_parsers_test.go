package handlers

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNativePayload_Single(t *testing.T) {
	raw := `{"source_type":"app","severity":"error","payload":{"msg":"fail"}}`
	entries, format, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "app", entries[0].SourceType)
	assert.Equal(t, "error", entries[0].Severity)
	assert.Equal(t, "native", format)
}

func TestParseNativePayload_Batch(t *testing.T) {
	raw := `[{"source_type":"a","severity":"info","payload":{}},{"source_type":"b","severity":"warning","payload":{}}]`
	entries, format, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "a", entries[0].SourceType)
	assert.Equal(t, "b", entries[1].SourceType)
	assert.Equal(t, "native_batch", format)
}

func TestParseVercelNDJSON(t *testing.T) {
	lines := `{"message":"GET /api/hello","timestamp":1712345678000,"source":"lambda","projectName":"my-app","level":"info"}
{"message":"Error: timeout","timestamp":1712345679000,"source":"lambda","projectName":"my-app","level":"error"}`

	entries, _, err := parseWebhookPayload([]byte(lines), "application/x-ndjson")
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "vercel/lambda", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)

	assert.Equal(t, "vercel/lambda", entries[1].SourceType)
	assert.Equal(t, "error", entries[1].Severity)

	var payload map[string]any
	json.Unmarshal(entries[1].Payload, &payload)
	assert.Equal(t, "Error: timeout", payload["message"])
}

func TestParseVercelNDJSON_RequiresContentType(t *testing.T) {
	// Without ndjson content-type, NDJSON is NOT auto-detected (structural
	// detection removed to prevent false positives). Callers without the
	// correct content-type should use the explicit /vercel route instead.
	lines := `{"message":"line1","source":"edge","level":"info"}
{"message":"line2","source":"edge","level":"warning"}`

	_, _, err := parseWebhookPayload([]byte(lines), "application/json")
	assert.Error(t, err, "NDJSON without content-type should fail JSON parsing")
}

func TestParseVercelNDJSON_ViaExplicitFormat(t *testing.T) {
	// The explicit format path (parseByFormat) should still parse NDJSON correctly.
	lines := `{"message":"line1","source":"edge","level":"info"}
{"message":"line2","source":"edge","level":"warning"}`

	entries, format, err := parseByFormat([]byte(lines), "vercel")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "vercel/edge", entries[0].SourceType)
	assert.Equal(t, "vercel_ndjson", format)
}

func TestParseFirehosePayload(t *testing.T) {
	record1 := base64.StdEncoding.EncodeToString([]byte(`{"message":"hello","severity":"info"}`))
	record2 := base64.StdEncoding.EncodeToString([]byte(`{"message":"fail","level":"error"}`))

	payload := map[string]any{
		"requestId": "req-123",
		"timestamp": 1712345678000,
		"records": []map[string]any{
			{"data": record1},
			{"data": record2},
		},
	}
	raw, _ := json.Marshal(payload)

	entries, _, err := parseWebhookPayload(raw, "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "firehose", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)

	assert.Equal(t, "firehose", entries[1].SourceType)
	assert.Equal(t, "error", entries[1].Severity)
}

func TestParseFirehosePayload_RawText(t *testing.T) {
	record := base64.StdEncoding.EncodeToString([]byte("Just a plain log line"))
	payload := map[string]any{
		"requestId": "req-456",
		"timestamp": 1712345678000,
		"records":   []map[string]any{{"data": record}},
	}
	raw, _ := json.Marshal(payload)

	entries, _, err := parseWebhookPayload(raw, "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "firehose", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)
}

func TestParsePubSubPayload_JSON(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte(`{"message":"db error","severity":"error","request_id":"abc"}`))
	payload := map[string]any{
		"message": map[string]any{
			"data":       data,
			"messageId":  "msg-123",
			"attributes": map[string]string{"source": "cloud-sql"},
		},
		"subscription": "projects/my-project/subscriptions/heimdall",
	}
	raw, _ := json.Marshal(payload)

	entries, _, err := parseWebhookPayload(raw, "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "pubsub", entries[0].SourceType)
	assert.Equal(t, "error", entries[0].Severity)
}

func TestParsePubSubPayload_PlainText(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("plain log message"))
	payload := map[string]any{
		"message": map[string]any{
			"data":      data,
			"messageId": "msg-789",
		},
		"subscription": "projects/p/subscriptions/s",
	}
	raw, _ := json.Marshal(payload)

	entries, _, err := parseWebhookPayload(raw, "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "pubsub", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"FATAL", "critical"},
		{"critical", "critical"},
		{"crit", "critical"},
		{"ERROR", "error"},
		{"err", "error"},
		{"WARNING", "warning"},
		{"warn", "warning"},
		{"INFO", "info"},
		{"notice", "info"},
		{"DEBUG", "debug"},
		{"trace", "debug"},
		{"unknown", "info"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeSeverity(tt.input), "input=%q", tt.input)
	}
}

func TestExtractSeverityFromMap(t *testing.T) {
	assert.Equal(t, "error", extractSeverityFromMap(map[string]any{"severity": "error"}))
	assert.Equal(t, "warning", extractSeverityFromMap(map[string]any{"level": "warn"}))
	assert.Equal(t, "info", extractSeverityFromMap(map[string]any{"no_severity_key": "value"}))
}

func TestParseVectorFly_Single(t *testing.T) {
	raw := `{
		"message": "request completed in 12ms",
		"timestamp": "2026-04-15T12:00:00.123Z",
		"host": "e784079c",
		"source_type": "fly_io",
		"fly": {
			"app": {"name": "my-fly-app"},
			"machine": {"id": "e784079c"},
			"region": "lhr"
		},
		"log": {"level": "info"}
	}`

	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, "flyio/my-fly-app", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)

	var payload map[string]any
	json.Unmarshal(entries[0].Payload, &payload)
	assert.Equal(t, "request completed in 12ms", payload["message"])
	assert.Equal(t, "e784079c", payload["machine_id"])
	assert.Equal(t, "lhr", payload["region"])
	assert.Equal(t, "my-fly-app", payload["app_name"])
}

func TestParseVectorFly_Batch(t *testing.T) {
	raw := `[
		{
			"message": "line 1",
			"timestamp": "2026-04-15T12:00:00Z",
			"host": "abc",
			"source_type": "fly_io",
			"fly": {"app": {"name": "app1"}, "machine": {"id": "abc"}, "region": "iad"},
			"log": {"level": "info"}
		},
		{
			"message": "disk full",
			"timestamp": "2026-04-15T12:00:01Z",
			"host": "def",
			"source_type": "fly_io",
			"fly": {"app": {"name": "app1"}, "machine": {"id": "def"}, "region": "lhr"},
			"log": {"level": "error"}
		}
	]`

	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "flyio/app1", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)

	assert.Equal(t, "flyio/app1", entries[1].SourceType)
	assert.Equal(t, "error", entries[1].Severity)
}

func TestParseVectorFly_SeverityMapping(t *testing.T) {
	raw := `{
		"message": "crash",
		"timestamp": "2026-04-15T12:00:00Z",
		"host": "abc",
		"source_type": "fly_io",
		"fly": {"app": {"name": "app"}, "machine": {"id": "abc"}, "region": "lhr"},
		"log": {"level": "warn"}
	}`

	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "warning", entries[0].Severity)
}

func TestParseVectorFly_DetectBySourceType(t *testing.T) {
	// No "fly" nested object, but source_type starts with "fly".
	raw := `{
		"message": "generic vector log",
		"timestamp": "2026-04-15T12:00:00Z",
		"host": "machine-1",
		"source_type": "fly_io"
	}`

	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "fly_io", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)
}

func TestParseVectorFly_NoFlyMetadata(t *testing.T) {
	// source_type starts with "fly" but no fly metadata — uses source_type as-is.
	raw := `{
		"message": "log line",
		"timestamp": "2026-04-15T12:00:00Z",
		"host": "host-1",
		"source_type": "fly_app_logs",
		"level": "error"
	}`

	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "fly_app_logs", entries[0].SourceType)
	// Falls back to extractSeverityFromMap since no log.level
	assert.Equal(t, "error", entries[0].Severity)
}

func TestIsVectorFlyPayload_Negative(t *testing.T) {
	// Should NOT match for Heimdall native or other formats.
	assert.False(t, isVectorFlyPayload([]byte(`{"source_type":"app","severity":"error","payload":{}}`)))
	assert.False(t, isVectorFlyPayload([]byte(`{"message":"hello","level":"info"}`)))
	assert.False(t, isVectorFlyPayload([]byte(`[{"source_type":"app","severity":"info","payload":{}}]`)))

	// "flywheel" should NOT match — exact match set, not prefix.
	assert.False(t, isVectorFlyPayload([]byte(`{"source_type":"flywheel","message":"data"}`)))
	assert.False(t, isVectorFlyPayload([]byte(`{"source_type":"flutter","message":"data"}`)))
}

func TestIsFirehosePayload_Negative(t *testing.T) {
	// An app log that happens to have "requestId" and "records" but no data
	// field in records should NOT match as Firehose.
	raw := `{"requestId":"req-123","records":[{"id":"r1","message":"not firehose"}]}`
	assert.False(t, isFirehosePayload([]byte(raw)))

	// Empty records array should NOT match.
	raw2 := `{"requestId":"req-456","records":[]}`
	assert.False(t, isFirehosePayload([]byte(raw2)))
}

func TestParseByFormat(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		payload    string
		wantFormat string
		wantLen    int
	}{
		{
			name:       "flyio explicit",
			format:     "flyio",
			payload:    `{"message":"hi","fly":{"app":{"name":"x"},"machine":{"id":"m"},"region":"iad"},"log":{"level":"info"}}`,
			wantFormat: "flyio_vector",
			wantLen:    1,
		},
		{
			name:       "native explicit single",
			format:     "native",
			payload:    `{"source_type":"app","severity":"info","payload":{}}`,
			wantFormat: "native",
			wantLen:    1,
		},
		{
			name:       "native explicit batch",
			format:     "native",
			payload:    `[{"source_type":"a","payload":{}},{"source_type":"b","payload":{}}]`,
			wantFormat: "native_batch",
			wantLen:    2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, format, err := parseByFormat([]byte(tt.payload), tt.format)
			require.NoError(t, err)
			assert.Len(t, entries, tt.wantLen)
			assert.Equal(t, tt.wantFormat, format)
		})
	}
}

func TestParseByFormat_Unknown(t *testing.T) {
	_, _, err := parseByFormat([]byte(`{}`), "unknown_format")
	assert.Error(t, err)
}

func TestParseWebhookPayload_Empty(t *testing.T) {
	entries, _, err := parseWebhookPayload([]byte(""), "application/json")
	require.NoError(t, err)
	assert.Nil(t, entries)
}

func TestParseWebhookPayload_FormatEcho(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		contentType string
		wantFormat  string
	}{
		{
			name:        "native single",
			payload:     `{"source_type":"app","severity":"info","payload":{}}`,
			contentType: "application/json",
			wantFormat:  "native",
		},
		{
			name:        "native batch",
			payload:     `[{"source_type":"a","payload":{}},{"source_type":"b","payload":{}}]`,
			contentType: "application/json",
			wantFormat:  "native_batch",
		},
		{
			name:        "vercel ndjson by content-type",
			payload:     "{\"message\":\"a\",\"source\":\"lambda\"}\n{\"message\":\"b\",\"source\":\"edge\"}",
			contentType: "application/x-ndjson",
			wantFormat:  "vercel_ndjson",
		},
		{
			name:        "flyio vector",
			payload:     `{"message":"hi","fly":{"app":{"name":"x"},"machine":{"id":"m"},"region":"iad"},"log":{"level":"info"}}`,
			contentType: "application/json",
			wantFormat:  "flyio_vector",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, format, err := parseWebhookPayload([]byte(tt.payload), tt.contentType)
			require.NoError(t, err)
			assert.Equal(t, tt.wantFormat, format)
		})
	}
}

func TestParseWebhookPayload_InvalidJSON(t *testing.T) {
	_, format, err := parseWebhookPayload([]byte(`{not json`), "application/json")
	assert.Error(t, err)
	assert.Empty(t, format)
}

func TestWebhookErrorSerialization(t *testing.T) {
	// Verify the structured error type serializes correctly.
	e := webhookError{
		Error:          "validation_failed",
		Message:        "one or more entries failed validation",
		FormatDetected: "native_batch",
		Details: []webhookFieldError{
			{Index: 2, Field: "source_type", Error: "required"},
			{Index: 4, Field: "payload", Error: "required"},
		},
	}
	b, err := json.Marshal(e)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &decoded))

	assert.Equal(t, "validation_failed", decoded["error"])
	assert.Equal(t, "native_batch", decoded["format_detected"])

	details := decoded["details"].([]interface{})
	require.Len(t, details, 2)

	first := details[0].(map[string]interface{})
	assert.Equal(t, float64(2), first["index"])
	assert.Equal(t, "source_type", first["field"])
	assert.Equal(t, "required", first["error"])
}

func TestWebhookErrorSerialization_NoDetails(t *testing.T) {
	// When details is nil, the field should be omitted from JSON.
	e := webhookError{
		Error:   "empty_payload",
		Message: "no log entries found after parsing",
	}
	b, err := json.Marshal(e)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &decoded))

	assert.Equal(t, "empty_payload", decoded["error"])
	_, hasDetails := decoded["details"]
	assert.False(t, hasDetails, "details should be omitted when nil")
	_, hasFormat := decoded["format_detected"]
	assert.False(t, hasFormat, "format_detected should be omitted when empty")
}

// --- Native v2 format tests ---

func TestParseNativeV2_Single(t *testing.T) {
	raw := `{"source":"my-app","level":"error","message":"connection refused","attrs":{"host":"web-1","request_id":"req-123"}}`
	entries, format, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "native_v2", format)
	assert.Equal(t, "my-app", entries[0].SourceType)
	assert.Equal(t, "error", entries[0].Severity)

	// Verify message was merged into payload.
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(entries[0].Payload, &payload))
	assert.Equal(t, "connection refused", payload["message"])
	assert.Equal(t, "web-1", payload["host"])
	assert.Equal(t, "req-123", payload["request_id"])
}

func TestParseNativeV2_Batch(t *testing.T) {
	raw := `[
		{"source":"api","level":"info","message":"GET /health 200","attrs":{}},
		{"source":"api","level":"error","message":"timeout","attrs":{"upstream":"db"}}
	]`
	entries, format, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "native_v2_batch", format)
	assert.Equal(t, "api", entries[0].SourceType)
	assert.Equal(t, "info", entries[0].Severity)
	assert.Equal(t, "api", entries[1].SourceType)
	assert.Equal(t, "error", entries[1].Severity)
}

func TestParseNativeV2_NoAttrs(t *testing.T) {
	// attrs is optional — message-only entries should work.
	raw := `{"source":"worker","level":"warn","message":"job queue full"}`
	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "worker", entries[0].SourceType)
	assert.Equal(t, "warning", entries[0].Severity) // "warn" normalises to "warning"

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(entries[0].Payload, &payload))
	assert.Equal(t, "job queue full", payload["message"])
}

func TestParseNativeV2_NoLevel(t *testing.T) {
	// level is optional — defaults to "info".
	raw := `{"source":"cron","message":"daily backup started"}`
	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "info", entries[0].Severity)
}

func TestParseNativeV2_MissingSource(t *testing.T) {
	// source is required in v2 — but this payload has neither "source" nor
	// "source_type", so it falls through to v1 parsing (not detected as v2).
	// The v1 parser will return it with empty SourceType, and validation
	// in the handler will reject it.
	raw := `{"level":"error","message":"orphan log"}`
	entries, _, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Empty(t, entries[0].SourceType)
}

func TestParseNativeV2_MixedBatch(t *testing.T) {
	// Mixed v1+v2 entries in a single batch should be rejected.
	raw := `[
		{"source":"api","level":"info","message":"v2 entry"},
		{"source_type":"legacy","severity":"error","payload":{"msg":"v1 entry"}}
	]`
	_, _, err := parseWebhookPayload([]byte(raw), "application/json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mixed v1/v2")
}

func TestParseNativeV2_ViaExplicitFormat(t *testing.T) {
	raw := `{"source":"my-app","level":"info","message":"hello"}`
	entries, format, err := parseByFormat([]byte(raw), "native")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "native_v2", format)
	assert.Equal(t, "my-app", entries[0].SourceType)
}

func TestParseNativeV1_StillWorks(t *testing.T) {
	// v1 format should continue to work unchanged.
	raw := `{"source_type":"legacy","severity":"error","payload":{"msg":"old format"}}`
	entries, format, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "native", format)
	assert.Equal(t, "legacy", entries[0].SourceType)
	assert.Equal(t, "error", entries[0].Severity)
}

func TestParseNativeV2_BothSourceFields(t *testing.T) {
	// If both "source" (v2) and "source_type" (v1) are present, v2 takes precedence.
	raw := `{"source":"v2-app","source_type":"v1-app","level":"info","message":"dual fields"}`
	entries, format, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "native_v2", format)
	assert.Equal(t, "v2-app", entries[0].SourceType) // v2 wins
}

// --- Idempotency response caching tests ---

func TestWebhookResponseSerialization_ForCaching(t *testing.T) {
	// The idempotency cache stores the JSON-marshalled webhookLogResponse.
	// Verify the cached body round-trips correctly.
	resp := webhookLogResponse{
		Accepted:   5,
		Format:     "native_v2_batch",
		Deprecated: false,
	}
	body, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded webhookLogResponse
	require.NoError(t, json.Unmarshal(body, &decoded))
	assert.Equal(t, 5, decoded.Accepted)
	assert.Equal(t, "native_v2_batch", decoded.Format)
	assert.False(t, decoded.Deprecated)

	// Deprecated should be omitted when false.
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &raw))
	_, hasDeprecated := raw["deprecated"]
	assert.False(t, hasDeprecated, "deprecated should be omitted when false")
}

func TestWebhookResponseSerialization_DeprecatedCaching(t *testing.T) {
	// v1 responses include deprecated:true — verify it round-trips.
	resp := webhookLogResponse{
		Accepted:   1,
		Format:     "native",
		Deprecated: true,
	}
	body, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded webhookLogResponse
	require.NoError(t, json.Unmarshal(body, &decoded))
	assert.True(t, decoded.Deprecated)
}
