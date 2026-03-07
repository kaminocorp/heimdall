package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func TestFormatFlaggedLogs(t *testing.T) {
	app := db.ListActiveApplicationsRow{
		ID:   uuid.New(),
		Name: "test-app",
	}

	flagged := []ClassifiedLog{
		{
			Log: db.LogBuffer{
				ID:         uuid.New(),
				Payload:    json.RawMessage(`{"message":"connection refused"}`),
				IngestedAt: time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
			},
			Type:       "ERROR",
			Category:   "connection_failure",
			Severity:   "error",
			Confidence: 0.95,
			Summary:    "connection refused",
		},
	}

	result := formatFlaggedLogs(app, flagged)

	assert.Contains(t, result, "Application: test-app")
	assert.Contains(t, result, "Flagged logs (1)")
	assert.Contains(t, result, "ERROR.connection_failure")
	assert.Contains(t, result, "confidence: 0.95")
	assert.Contains(t, result, "connection refused")
	assert.Contains(t, result, "2026-03-07T10:00:00Z")
}

func TestFormatFlaggedLogs_MultipleLogs(t *testing.T) {
	app := db.ListActiveApplicationsRow{
		ID:   uuid.New(),
		Name: "multi-app",
	}

	flagged := []ClassifiedLog{
		{
			Log:      db.LogBuffer{Payload: json.RawMessage(`{"message":"error 1"}`), IngestedAt: time.Now()},
			Type:     "ERROR",
			Category: "timeout",
			Severity: "error",
			Summary:  "error 1",
		},
		{
			Log:      db.LogBuffer{Payload: json.RawMessage(`{"message":"error 2"}`), IngestedAt: time.Now()},
			Type:     "PERFORMANCE",
			Category: "latency_spike",
			Severity: "warning",
			Summary:  "error 2",
		},
	}

	result := formatFlaggedLogs(app, flagged)
	assert.Contains(t, result, "Flagged logs (2)")
	assert.Contains(t, result, "--- Log 1 ---")
	assert.Contains(t, result, "--- Log 2 ---")
}

func TestFormatFlaggedLogs_TruncatesLargePayload(t *testing.T) {
	app := db.ListActiveApplicationsRow{
		ID:   uuid.New(),
		Name: "truncate-app",
	}

	// Build a payload larger than maxPayloadChars (2000).
	largePayload := make([]byte, 3000)
	for i := range largePayload {
		largePayload[i] = 'x'
	}

	flagged := []ClassifiedLog{
		{
			Log:      db.LogBuffer{Payload: largePayload, IngestedAt: time.Now()},
			Type:     "ERROR",
			Category: "large_payload",
			Severity: "error",
			Summary:  "large log entry",
		},
	}

	result := formatFlaggedLogs(app, flagged)
	assert.Contains(t, result, "... [truncated]")
	// The full 3000-char payload should NOT be in the output.
	assert.NotContains(t, result, string(largePayload))
}

func TestParseSeverityFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "explicit severity critical",
			text:     "Assessment: database down.\nSeverity: critical\nImmediate action required.",
			expected: "critical",
		},
		{
			name:     "explicit severity warning",
			text:     "Minor issue detected.\nSeverity: warning\nMonitor for recurrence.",
			expected: "warning",
		},
		{
			name:     "explicit severity info",
			text:     "False positive from classifier.\nSeverity: info\nNo action needed.",
			expected: "info",
		},
		{
			name:     "explicit severity error",
			text:     "Connection failures detected.\nSeverity: error\nInvestigate DB connectivity.",
			expected: "error",
		},
		{
			name:     "heuristic error keyword",
			text:     "Multiple error conditions detected in the application logs.",
			expected: "error",
		},
		{
			name:     "heuristic critical keyword",
			text:     "This is a critical failure in the database layer.",
			expected: "critical",
		},
		{
			name:     "heuristic warning keyword",
			text:     "This is a warning about elevated latency.",
			expected: "warning",
		},
		{
			name:     "no severity keywords defaults to info",
			text:     "Everything looks fine. Logs appear normal.",
			expected: "info",
		},
		{
			name:     "explicit takes priority over heuristic",
			text:     "There was an error in the logs.\nSeverity: info\nThis was a false positive.",
			expected: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseSeverityFromResponse(tt.text))
		})
	}
}

func TestShouldMonitor_ContinuousMode(t *testing.T) {
	agent := newToolTestAgent()
	ctx := context.Background()

	app := db.ListActiveApplicationsRow{
		ID:                   uuid.New(),
		Mode:                 "continuous",
		ScheduleIntervalSecs: 60,
	}

	// Continuous mode should always return true.
	assert.True(t, agent.shouldMonitor(ctx, app))
}

func TestShouldMonitor_PeriodicMode_NoState(t *testing.T) {
	agent := newToolTestAgent()
	ctx := context.Background()

	app := db.ListActiveApplicationsRow{
		ID:                   uuid.New(),
		Mode:                 "periodic",
		ScheduleIntervalSecs: 300,
	}

	// No monitoring state exists (DB stub returns error) → should monitor.
	assert.True(t, agent.shouldMonitor(ctx, app))
}

func TestRunMonitoring_SimpleResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(anthropicResponse("end_turn", []map[string]any{
			textBlock("Assessment: connection errors detected.\nSeverity: error\nInvestigate DB connectivity."),
		}))
	}))
	defer server.Close()

	agent := newTestAgent(t, server)
	ctx := context.Background()
	userID := uuid.New()
	appConfig := db.AppAgentConfig{
		AppID: uuid.New(),
		Model: "claude-sonnet-4-6",
	}

	assessment, severity := agent.RunMonitoring(ctx, userID, appConfig, "Flagged logs:\nERROR: connection refused")

	assert.Contains(t, assessment, "connection errors detected")
	assert.Equal(t, "error", severity)
}

func TestRunMonitoring_WithToolUse(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First call: request tool use.
			w.Write(anthropicResponse("tool_use", []map[string]any{
				toolUseBlock("toolu_mon1", "search_logs", map[string]any{"query": "connection refused"}),
			}))
			return
		}

		// Second call: return assessment.
		w.Write(anthropicResponse("end_turn", []map[string]any{
			textBlock("After investigation, this is a transient network issue.\nSeverity: warning"),
		}))
	}))
	defer server.Close()

	agent := newTestAgent(t, server)
	ctx := context.Background()
	userID := uuid.New()
	appConfig := db.AppAgentConfig{
		AppID: uuid.New(),
		Model: "claude-sonnet-4-6",
	}

	assessment, severity := agent.RunMonitoring(ctx, userID, appConfig, "Flagged: ERROR connection refused")

	assert.Contains(t, assessment, "transient network issue")
	assert.Equal(t, "warning", severity)
	assert.Equal(t, 2, callCount)
}

func TestRunMonitoring_MaxIterations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(anthropicResponse("tool_use", []map[string]any{
			toolUseBlock("toolu_loop", "search_logs", map[string]any{"query": "errors"}),
		}))
	}))
	defer server.Close()

	agent := newTestAgent(t, server)
	ctx := context.Background()
	userID := uuid.New()
	appConfig := db.AppAgentConfig{
		AppID: uuid.New(),
		Model: "claude-sonnet-4-6",
	}

	assessment, severity := agent.RunMonitoring(ctx, userID, appConfig, "Flagged: some error")

	assert.Contains(t, assessment, "exceeded max iterations")
	assert.Equal(t, "warning", severity)
}

func TestMonitorStartStop(t *testing.T) {
	agent := newToolTestAgent()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Start monitoring in a goroutine.
	done := make(chan struct{})
	go func() {
		agent.Monitor(ctx)
		close(done)
	}()

	// Cancel immediately — monitor should exit cleanly.
	cancel()

	select {
	case <-done:
		// Success — monitor exited.
	case <-time.After(2 * time.Second):
		t.Fatal("monitor did not exit within 2 seconds")
	}
}

func TestMonitorTick_NoApps(t *testing.T) {
	// With stubDBTX, ListActiveApplications will return an error.
	// monitorTick should log the error and return without panic.
	agent := newToolTestAgent()
	ctx := context.Background()
	sem := make(chan struct{}, maxConcurrentApps)

	// Should not panic.
	require.NotPanics(t, func() {
		agent.monitorTick(ctx, sem)
	})
}

func TestBuildMonitoringPrompt(t *testing.T) {
	t.Run("no override", func(t *testing.T) {
		prompt := BuildMonitoringPrompt("")
		assert.Contains(t, prompt, "autonomous monitoring agent")
		assert.Contains(t, prompt, "deterministic classifier")
		assert.Contains(t, prompt, "Severity: <info | warning | error | critical>")
		assert.Contains(t, prompt, "OUTPUT (strict format)")
	})

	t.Run("with override", func(t *testing.T) {
		prompt := BuildMonitoringPrompt("Focus on database errors specifically.")
		assert.Contains(t, prompt, "autonomous monitoring agent")
		assert.Contains(t, prompt, "Focus on database errors specifically.")
	})
}
