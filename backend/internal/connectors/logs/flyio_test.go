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

func flyioValidConfig() json.RawMessage {
	return json.RawMessage(`{
		"app_name": "my-fly-app",
		"api_token": "fo1_test_token",
		"poll_interval_secs": 30
	}`)
}

// --- Constructor tests ---

func TestNewFlyio_ValidConfig(t *testing.T) {
	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	cfg := f.ParsedConfig()
	assert.Equal(t, "my-fly-app", cfg.AppName)
	assert.Equal(t, "fo1_test_token", cfg.APIToken)
	assert.Equal(t, 30, cfg.PollIntervalSecs)
}

func TestNewFlyio_Defaults(t *testing.T) {
	raw := json.RawMessage(`{
		"app_name": "app",
		"api_token": "tok"
	}`)
	f, err := NewFlyio(raw, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, flyDefaultInterval, f.ParsedConfig().PollIntervalSecs)
}

func TestNewFlyio_MissingAppName(t *testing.T) {
	raw := json.RawMessage(`{"api_token": "tok"}`)
	_, err := NewFlyio(raw, uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "app_name is required")
}

func TestNewFlyio_MissingAPIToken(t *testing.T) {
	raw := json.RawMessage(`{"app_name": "app"}`)
	_, err := NewFlyio(raw, uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api_token is required")
}

func TestNewFlyio_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`{bad json}`)
	_, err := NewFlyio(raw, uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestNewFlyio_PollIntervalMinimum(t *testing.T) {
	raw := json.RawMessage(`{
		"app_name": "app",
		"api_token": "tok",
		"poll_interval_secs": 5
	}`)
	f, err := NewFlyio(raw, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, flyDefaultInterval, f.ParsedConfig().PollIntervalSecs)
}

// --- Connect tests ---

func TestFlyio_Connect_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer fo1_test_token", r.Header.Get("Authorization"))
		assert.Contains(t, r.URL.Path, "/v1/apps/my-fly-app/logs")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	assert.NoError(t, f.Connect(context.Background()))
}

func TestFlyio_Connect_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	err = f.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API token")
}

func TestFlyio_Connect_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	err = f.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- Poll tests (NDJSON parsing) ---

func TestFlyio_Poll_RequestFormation(t *testing.T) {
	// Verify that Poll sends the correct request to the app-level logs endpoint.
	var receivedPath string
	var receivedAuth string
	var receivedStartTime string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedAuth = r.Header.Get("Authorization")
		receivedStartTime = r.URL.Query().Get("start_time")
		// Return empty response — no log lines to process.
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	err = f.Poll(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "/v1/apps/my-fly-app/logs", receivedPath)
	assert.Equal(t, "Bearer fo1_test_token", receivedAuth)
	assert.NotEmpty(t, receivedStartTime)
}

func TestFlyio_Poll_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body — no log lines.
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	// With no lines to parse, Poll should succeed without touching queries.
	err = f.Poll(context.Background(), nil)
	assert.NoError(t, err)
}

func TestFlyio_Poll_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	err = f.Poll(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestFlyio_Poll_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	err = f.Poll(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestFlyio_Poll_MalformedLines(t *testing.T) {
	// Only truly unparseable JSON — these should be skipped gracefully.
	ndjson := "not valid json\n\n<<<broken>>>\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(ndjson))
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	// All lines are malformed or empty — no DB insert is attempted.
	err = f.Poll(context.Background(), nil)
	assert.NoError(t, err)
}

func TestFlyio_Poll_OldEntriesSkipped(t *testing.T) {
	// All entries are older than cursor — nothing should be inserted, no panic.
	old := time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339Nano)

	ndjson := fmt.Sprintf(`{"timestamp":"%s","message":"old","level":"info"}`, old)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(ndjson))
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL
	f.cursor = time.Now().Add(-5 * time.Minute) // newer than the log entry

	// Old entry is filtered out by cursor, so queries is never touched.
	err = f.Poll(context.Background(), nil)
	assert.NoError(t, err)
}

// --- LogsEndpoint tests ---

func TestFlyio_LogsEndpoint(t *testing.T) {
	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	ts := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	endpoint := f.logsEndpoint(ts)

	assert.Contains(t, endpoint, "/v1/apps/my-fly-app/logs")
	assert.Contains(t, endpoint, "start_time=")
	assert.Contains(t, endpoint, "2026-04-15")
}

func TestFlyio_LogsEndpoint_EscapesAppName(t *testing.T) {
	raw := json.RawMessage(`{
		"app_name": "app with spaces",
		"api_token": "tok"
	}`)
	f, err := NewFlyio(raw, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	endpoint := f.logsEndpoint(time.Now())
	assert.Contains(t, endpoint, "app%20with%20spaces")
}

// --- normalizeSev tests ---

func TestNormalizeSev(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"error", "error"},
		{"err", "error"},
		{"warn", "warning"},
		{"warning", "warning"},
		{"debug", "debug"},
		{"info", "info"},
		{"", "info"},
		{"unknown", "info"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeSev(tt.input)
			assert.True(t, result.Valid)
			assert.Equal(t, tt.want, result.String)
		})
	}
}

// --- Close / Health tests ---

func TestFlyio_Close(t *testing.T) {
	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.NoError(t, f.Close())
}

func TestFlyio_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	f, err := NewFlyio(flyioValidConfig(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	f.apiBase = server.URL

	assert.NoError(t, f.Health(context.Background()))
}
