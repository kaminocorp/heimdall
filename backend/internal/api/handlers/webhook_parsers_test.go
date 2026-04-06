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
	entries, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "app", entries[0].SourceType)
	assert.Equal(t, "error", entries[0].Severity)
}

func TestParseNativePayload_Batch(t *testing.T) {
	raw := `[{"source_type":"a","severity":"info","payload":{}},{"source_type":"b","severity":"warning","payload":{}}]`
	entries, err := parseWebhookPayload([]byte(raw), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "a", entries[0].SourceType)
	assert.Equal(t, "b", entries[1].SourceType)
}

func TestParseVercelNDJSON(t *testing.T) {
	lines := `{"message":"GET /api/hello","timestamp":1712345678000,"source":"lambda","projectName":"my-app","level":"info"}
{"message":"Error: timeout","timestamp":1712345679000,"source":"lambda","projectName":"my-app","level":"error"}`

	entries, err := parseWebhookPayload([]byte(lines), "application/x-ndjson")
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

func TestParseVercelNDJSON_AutoDetect(t *testing.T) {
	// Without ndjson content-type, should still detect by structure.
	lines := `{"message":"line1","source":"edge","level":"info"}
{"message":"line2","source":"edge","level":"warning"}`

	entries, err := parseWebhookPayload([]byte(lines), "application/json")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "vercel/edge", entries[0].SourceType)
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

	entries, err := parseWebhookPayload(raw, "application/json")
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

	entries, err := parseWebhookPayload(raw, "application/json")
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

	entries, err := parseWebhookPayload(raw, "application/json")
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

	entries, err := parseWebhookPayload(raw, "application/json")
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

func TestParseWebhookPayload_Empty(t *testing.T) {
	entries, err := parseWebhookPayload([]byte(""), "application/json")
	require.NoError(t, err)
	assert.Nil(t, entries)
}
