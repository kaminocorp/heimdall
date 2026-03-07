package agent

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func logEntry(payload string) db.LogBuffer {
	return db.LogBuffer{
		ID:      uuid.New(),
		Payload: json.RawMessage(payload),
	}
}

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		expected string
	}{
		{
			name:     "standard message field",
			payload:  `{"level":"error","message":"connection refused"}`,
			expected: "ERROR: connection refused",
		},
		{
			name:     "shorthand msg field",
			payload:  `{"msg":"request completed","status":200}`,
			expected: "request completed",
		},
		{
			name:     "error field",
			payload:  `{"error":"timeout after 30s"}`,
			expected: "timeout after 30s",
		},
		{
			name:     "level prepended",
			payload:  `{"level":"warn","message":"high latency"}`,
			expected: "WARN: high latency",
		},
		{
			name:     "supabase error_severity",
			payload:  `{"error_severity":"ERROR","msg":"deadlock detected"}`,
			expected: "ERROR: deadlock detected",
		},
		{
			name:     "no message field falls back to raw JSON",
			payload:  `{"path":"/api/users","status":500}`,
			expected: `{"path":"/api/users","status":500}`,
		},
		{
			name:     "non-JSON payload",
			payload:  `plain text log line`,
			expected: "plain text log line",
		},
		{
			name:     "empty object",
			payload:  `{}`,
			expected: "{}",
		},
		{
			name:     "message field priority over error field",
			payload:  `{"message":"primary msg","error":"secondary"}`,
			expected: "primary msg",
		},
		{
			name:     "severity field used when no level",
			payload:  `{"severity":"info","message":"startup complete"}`,
			expected: "INFO: startup complete",
		},
		{
			name:     "empty message field skipped",
			payload:  `{"message":"","error":"real error"}`,
			expected: "real error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractText(logEntry(tt.payload))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTexts(t *testing.T) {
	logs := []db.LogBuffer{
		logEntry(`{"message":"first"}`),
		logEntry(`{"msg":"second"}`),
		logEntry(`plain third`),
	}
	texts := ExtractTexts(logs)
	assert.Equal(t, []string{"first", "second", "plain third"}, texts)
}
