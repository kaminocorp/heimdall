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

	entries, err := parseWebhookPayload([]byte(raw), "application/json")
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

	entries, err := parseWebhookPayload([]byte(raw), "application/json")
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

	entries, err := parseWebhookPayload([]byte(raw), "application/json")
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

	entries, err := parseWebhookPayload([]byte(raw), "application/json")
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

	entries, err := parseWebhookPayload([]byte(raw), "application/json")
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
}

func TestParseWebhookPayload_Empty(t *testing.T) {
	entries, err := parseWebhookPayload([]byte(""), "application/json")
	require.NoError(t, err)
	assert.Nil(t, entries)
}
