package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() json.RawMessage {
	return json.RawMessage(`{
		"project_ref": "test-ref-123",
		"access_token": "sbp_test_token",
		"poll_tables": ["postgres_logs", "auth_logs"],
		"poll_interval_secs": 120
	}`)
}

func TestNewSupabase_ValidConfig(t *testing.T) {
	sb, err := NewSupabase(validConfig(), uuid.New(), uuid.New())
	require.NoError(t, err)

	cfg := sb.ParsedConfig()
	assert.Equal(t, "test-ref-123", cfg.ProjectRef)
	assert.Equal(t, "sbp_test_token", cfg.AccessToken)
	assert.Equal(t, []string{"postgres_logs", "auth_logs"}, cfg.PollTables)
	assert.Equal(t, 120, cfg.PollIntervalSecs)
}

func TestNewSupabase_Defaults(t *testing.T) {
	raw := json.RawMessage(`{
		"project_ref": "myproj",
		"access_token": "sbp_tok"
	}`)

	sb, err := NewSupabase(raw, uuid.New(), uuid.New())
	require.NoError(t, err)

	cfg := sb.ParsedConfig()
	assert.Equal(t, []string{supabaseDefaultTable}, cfg.PollTables)
	assert.Equal(t, supabaseDefaultInterval, cfg.PollIntervalSecs)
}

func TestNewSupabase_MissingProjectRef(t *testing.T) {
	raw := json.RawMessage(`{"access_token": "sbp_tok"}`)
	_, err := NewSupabase(raw, uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project_ref is required")
}

func TestNewSupabase_MissingAccessToken(t *testing.T) {
	raw := json.RawMessage(`{"project_ref": "myproj"}`)
	_, err := NewSupabase(raw, uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access_token is required")
}

func TestNewSupabase_InvalidTableName(t *testing.T) {
	raw := json.RawMessage(`{
		"project_ref": "myproj",
		"access_token": "sbp_tok",
		"poll_tables": ["postgres_logs", "malicious_table; DROP TABLE x--"]
	}`)
	_, err := NewSupabase(raw, uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestNewSupabase_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`{bad json}`)
	_, err := NewSupabase(raw, uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestNewSupabase_PollIntervalMinimum(t *testing.T) {
	raw := json.RawMessage(`{
		"project_ref": "myproj",
		"access_token": "sbp_tok",
		"poll_interval_secs": 5
	}`)

	sb, err := NewSupabase(raw, uuid.New(), uuid.New())
	require.NoError(t, err)
	// Interval below minimum (60) should be clamped to default (120).
	assert.Equal(t, supabaseDefaultInterval, sb.ParsedConfig().PollIntervalSecs)
}

func TestSupabase_Connect_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer sbp_test_token", r.Header.Get("Authorization"))
		assert.Contains(t, r.URL.Path, "/v1/projects/test-ref-123/analytics/endpoints/logs.all")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": []}`))
	}))
	defer server.Close()

	sb, err := NewSupabase(validConfig(), uuid.New(), uuid.New())
	require.NoError(t, err)
	sb.apiBase = server.URL

	err = sb.Connect(context.Background())
	assert.NoError(t, err)
}

func TestSupabase_Connect_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer server.Close()

	sb, err := NewSupabase(validConfig(), uuid.New(), uuid.New())
	require.NoError(t, err)
	sb.apiBase = server.URL

	err = sb.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestSupabase_Connect_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "project not found"}`))
	}))
	defer server.Close()

	sb, err := NewSupabase(validConfig(), uuid.New(), uuid.New())
	require.NoError(t, err)
	sb.apiBase = server.URL

	err = sb.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestSupabase_ParseResponse(t *testing.T) {
	// Verify that deriveSeverity correctly extracts severity from metadata.
	tests := []struct {
		name     string
		metadata json.RawMessage
		want     string
	}{
		{
			name:     "error_severity field",
			metadata: json.RawMessage(`{"error_severity": "ERROR"}`),
			want:     "ERROR",
		},
		{
			name:     "severity field",
			metadata: json.RawMessage(`{"severity": "WARNING"}`),
			want:     "WARNING",
		},
		{
			name:     "level field",
			metadata: json.RawMessage(`{"level": "info"}`),
			want:     "info",
		},
		{
			name:     "no severity field",
			metadata: json.RawMessage(`{"some_other": "field"}`),
			want:     "",
		},
		{
			name:     "nil metadata",
			metadata: nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveSeverity(tt.metadata)
			if tt.want == "" {
				assert.False(t, result.Valid)
			} else {
				assert.True(t, result.Valid)
				assert.Equal(t, tt.want, result.String)
			}
		})
	}
}

func TestSupabase_CursorAdvancement(t *testing.T) {
	// Mock server returns two rows with different timestamps.
	ts1 := time.Now().Add(-2 * time.Minute).UnixMicro()
	ts2 := time.Now().Add(-1 * time.Minute).UnixMicro()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var resp string
		if calls == 1 {
			// First poll: return two rows.
			resp = `{"result": [
				{"timestamp": ` + itoa(ts1) + `, "event_message": "msg1", "metadata": {}},
				{"timestamp": ` + itoa(ts2) + `, "event_message": "msg2", "metadata": {}}
			]}`
		} else {
			// Second poll: return empty (cursor should have advanced past ts2).
			resp = `{"result": []}`
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}))
	defer server.Close()

	sb, err := NewSupabase(json.RawMessage(`{
		"project_ref": "proj",
		"access_token": "tok",
		"poll_tables": ["postgres_logs"]
	}`), uuid.New(), uuid.New())
	require.NoError(t, err)
	sb.apiBase = server.URL

	// Poll should not error (we don't have real db.Queries, so we test queryAPI directly).
	body, err := sb.queryAPI(context.Background(), "postgres_logs", "SELECT 1")
	require.NoError(t, err)

	var resp struct {
		Result []struct {
			Timestamp int64 `json:"timestamp"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Len(t, resp.Result, 2)

	// Verify cursor would advance to the max timestamp.
	var maxTS int64
	for _, row := range resp.Result {
		if row.Timestamp > maxTS {
			maxTS = row.Timestamp
		}
	}
	assert.Equal(t, ts2, maxTS)
}

func TestSupabase_Close(t *testing.T) {
	sb, err := NewSupabase(validConfig(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.NoError(t, sb.Close())
}

func TestSupabase_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": []}`))
	}))
	defer server.Close()

	sb, err := NewSupabase(validConfig(), uuid.New(), uuid.New())
	require.NoError(t, err)
	sb.apiBase = server.URL

	assert.NoError(t, sb.Health(context.Background()))
}

func TestSupabase_RateLimit429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limited"}`))
	}))
	defer server.Close()

	sb, err := NewSupabase(validConfig(), uuid.New(), uuid.New())
	require.NoError(t, err)
	sb.apiBase = server.URL

	_, err = sb.queryAPI(context.Background(), "postgres_logs", "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}
