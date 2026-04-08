package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// stubDBTX implements db.DBTX and returns errors for all operations.
// This makes GetAgentConfig fail (triggering default model/prompt) and
// InsertAgentLog fail silently (fire-and-forget in EmitLog).
type stubDBTX struct{}

func (s *stubDBTX) Exec(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag(""), fmt.Errorf("stubDBTX: not implemented")
}

func (s *stubDBTX) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	return nil, fmt.Errorf("stubDBTX: not implemented")
}

func (s *stubDBTX) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	return &stubRow{}
}

// stubRow implements pgx.Row and always returns an error on Scan.
type stubRow struct{}

func (s *stubRow) Scan(_ ...interface{}) error {
	return fmt.Errorf("stubRow: no rows")
}

// newTestAgent creates an Agent wired to a mock HTTP server instead of
// the real Anthropic API. The caller controls responses via the server handler.
func newTestAgent(t *testing.T, server *httptest.Server) *Agent {
	t.Helper()
	client := anthropic.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)
	queries := db.New(&stubDBTX{})
	return &Agent{
		queries: queries,
		providers: map[string]Provider{
			defaultProviderName: NewAnthropicProviderWithClient(&client),
		},
		config:     &config.Config{AnthropicKey: "test-key"},
		classifier: &PassthroughClassifier{},
	}
}

// anthropicResponse builds a JSON response body that mimics the Anthropic
// Messages API format.
func anthropicResponse(stopReason string, content []map[string]any) []byte {
	resp := map[string]any{
		"id":          "msg_test_123",
		"type":        "message",
		"role":        "assistant",
		"model":       "claude-sonnet-4-5-20250514",
		"stop_reason": stopReason,
		"content":     content,
		"usage": map[string]any{
			"input_tokens":  10,
			"output_tokens": 20,
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

// textBlock creates a text content block for an Anthropic API response.
func textBlock(text string) map[string]any {
	return map[string]any{
		"type": "text",
		"text": text,
	}
}

// toolUseBlock creates a tool_use content block for an Anthropic API response.
func toolUseBlock(id, name string, input map[string]any) map[string]any {
	return map[string]any{
		"type":  "tool_use",
		"id":    id,
		"name":  name,
		"input": input,
	}
}

func TestRunLoop_SimpleResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(anthropicResponse("end_turn", []map[string]any{
			textBlock("Hello! I am Heimdall, your monitoring agent."),
		}))
	}))
	defer server.Close()

	agent := newTestAgent(t, server)
	ctx := context.Background()
	userID := uuid.New()

	result, err := agent.RunLoop(ctx, userID, "Hello")

	require.NoError(t, err)
	assert.Equal(t, "Hello! I am Heimdall, your monitoring agent.", result)
}

func TestRunLoop_MaxIterations(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		// Always return a tool_use response so the loop never terminates naturally.
		w.Write(anthropicResponse("tool_use", []map[string]any{
			toolUseBlock(
				fmt.Sprintf("toolu_%d", callCount.Load()),
				"search_logs",
				map[string]any{"query": "errors"},
			),
		}))
	}))
	defer server.Close()

	agent := newTestAgent(t, server)
	ctx := context.Background()
	userID := uuid.New()

	result, err := agent.RunLoop(ctx, userID, "Find all errors")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeded max iterations")
	assert.Empty(t, result)
	assert.Equal(t, int32(maxIterations), callCount.Load())
}

func TestRunLoop_ToolError(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		if callCount.Load() == 1 {
			// First call: Claude requests a tool use with an unknown tool.
			// We use "search_logs" which will fail because the stubDBTX
			// returns errors — the tool error is sent back as isError=true.
			w.Write(anthropicResponse("tool_use", []map[string]any{
				toolUseBlock("toolu_err1", "search_logs", map[string]any{"query": "errors"}),
			}))
			return
		}

		// Second call: Claude sees the tool error and produces a text response.
		w.Write(anthropicResponse("end_turn", []map[string]any{
			textBlock("I encountered an error while searching logs."),
		}))
	}))
	defer server.Close()

	agent := newTestAgent(t, server)
	ctx := context.Background()
	userID := uuid.New()

	result, err := agent.RunLoop(ctx, userID, "Search for errors")

	require.NoError(t, err)
	assert.Equal(t, "I encountered an error while searching logs.", result)
	// Verify we made exactly 2 API calls: tool_use + final response.
	assert.Equal(t, int32(2), callCount.Load())
}
