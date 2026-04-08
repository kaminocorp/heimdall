package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenRouter_RequestShape verifies that ChatParams gets translated into
// the OpenAI/OpenRouter wire format correctly: system prompt as first message,
// tools wrapped in a JSON Schema "object", and required headers set.
func TestOpenRouter_RequestShape(t *testing.T) {
	var captured orRequest
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &captured))

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "test",
			"choices": [{
				"index": 0,
				"message": {"role":"assistant","content":"hello back"},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens":5,"completion_tokens":3,"total_tokens":8}
		}`))
	}))
	defer server.Close()

	provider := NewOpenRouterProviderWithURL("test-key", server.URL)
	resp, err := provider.ChatCompletion(context.Background(), ChatParams{
		Model:        "openai/gpt-4o",
		SystemPrompt: "you are heimdall",
		Messages: []ChatMessage{
			{Role: "user", TextContent: "hello"},
		},
		Tools:     ToolRegistry(),
		MaxTokens: 1024,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Headers
	assert.Equal(t, "Bearer test-key", capturedHeaders.Get("Authorization"))
	assert.Equal(t, "application/json", capturedHeaders.Get("Content-Type"))
	assert.NotEmpty(t, capturedHeaders.Get("HTTP-Referer"))
	assert.NotEmpty(t, capturedHeaders.Get("X-Title"))

	// Request body
	assert.Equal(t, "openai/gpt-4o", captured.Model)
	assert.Equal(t, 1024, captured.MaxTokens)
	assert.False(t, captured.Stream)

	// System prompt is first message, not a separate field.
	require.GreaterOrEqual(t, len(captured.Messages), 2)
	assert.Equal(t, "system", captured.Messages[0].Role)
	assert.Equal(t, "you are heimdall", captured.Messages[0].Content)
	assert.Equal(t, "user", captured.Messages[1].Role)
	assert.Equal(t, "hello", captured.Messages[1].Content)

	// Tools wrapped in a JSON Schema object.
	require.NotEmpty(t, captured.Tools)
	for _, tool := range captured.Tools {
		assert.Equal(t, "function", tool.Type)
		assert.NotEmpty(t, tool.Function.Name)
		assert.Equal(t, "object", tool.Function.Parameters["type"])
		assert.NotNil(t, tool.Function.Parameters["properties"])
	}

	// Response was decoded into ChatResponse.
	assert.Equal(t, StopReasonEndTurn, resp.StopReason)
	require.Len(t, resp.Blocks, 1)
	assert.Equal(t, "text", resp.Blocks[0].Type)
	assert.Equal(t, "hello back", resp.Blocks[0].Text)
	assert.Equal(t, 5, resp.Usage.InputTokens)
	assert.Equal(t, 3, resp.Usage.OutputTokens)
}

// TestOpenRouter_ToolCallResponse verifies that finish_reason="tool_calls"
// becomes StopReasonToolUse and tool_calls are translated into ContentBlocks
// with json.RawMessage Input that the agent loop's Unmarshal can read.
func TestOpenRouter_ToolCallResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id":"test",
			"choices":[{
				"index":0,
				"message":{
					"role":"assistant",
					"content":"",
					"tool_calls":[{
						"id":"call_123",
						"type":"function",
						"function":{"name":"search_logs","arguments":"{\"query\":\"errors\"}"}
					}]
				},
				"finish_reason":"tool_calls"
			}],
			"usage":{"prompt_tokens":10,"completion_tokens":4,"total_tokens":14}
		}`))
	}))
	defer server.Close()

	provider := NewOpenRouterProviderWithURL("test-key", server.URL)
	resp, err := provider.ChatCompletion(context.Background(), ChatParams{
		Model: "openai/gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", TextContent: "find errors"},
		},
		Tools: ToolRegistry(),
	})

	require.NoError(t, err)
	assert.Equal(t, StopReasonToolUse, resp.StopReason)
	require.Len(t, resp.Blocks, 1)
	assert.Equal(t, "tool_use", resp.Blocks[0].Type)
	require.NotNil(t, resp.Blocks[0].ToolCall)
	assert.Equal(t, "call_123", resp.Blocks[0].ToolCall.ID)
	assert.Equal(t, "search_logs", resp.Blocks[0].ToolCall.Name)

	// Input should round-trip through json.Unmarshal.
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(resp.Blocks[0].ToolCall.Input, &parsed))
	assert.Equal(t, "errors", parsed["query"])
}

// TestOpenRouter_AssistantToolUseRoundTrip verifies that an assistant message
// with Blocks (text + tool_use) gets translated into a single OpenAI assistant
// message with content + tool_calls — and that tool results come back as
// role:"tool" messages with the matching tool_call_id.
func TestOpenRouter_AssistantToolUseRoundTrip(t *testing.T) {
	var captured orRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &captured))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"t","choices":[{"index":0,"message":{"role":"assistant","content":"done"},"finish_reason":"stop"}],"usage":{}}`))
	}))
	defer server.Close()

	provider := NewOpenRouterProviderWithURL("k", server.URL)
	_, err := provider.ChatCompletion(context.Background(), ChatParams{
		Model: "openai/gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", TextContent: "search"},
			{Role: "assistant", Blocks: []ContentBlock{
				{Type: "text", Text: "let me look"},
				{Type: "tool_use", ToolCall: &ToolCall{
					ID:    "call_xyz",
					Name:  "search_logs",
					Input: json.RawMessage(`{"query":"errors"}`),
				}},
			}},
			{Role: "user", ToolResults: []ToolResult{
				{ToolCallID: "call_xyz", Content: "no results", IsError: false},
			}},
		},
	})
	require.NoError(t, err)

	// Find the assistant message in the captured request.
	var assistant *orMessage
	var toolReply *orMessage
	for i := range captured.Messages {
		switch captured.Messages[i].Role {
		case "assistant":
			assistant = &captured.Messages[i]
		case "tool":
			toolReply = &captured.Messages[i]
		}
	}
	require.NotNil(t, assistant, "no assistant message in request")
	assert.Equal(t, "let me look", assistant.Content)
	require.Len(t, assistant.ToolCalls, 1)
	assert.Equal(t, "call_xyz", assistant.ToolCalls[0].ID)
	assert.Equal(t, "function", assistant.ToolCalls[0].Type)
	assert.Equal(t, "search_logs", assistant.ToolCalls[0].Function.Name)
	// arguments must be a JSON string per OpenAI spec — round-trip it.
	var args map[string]any
	require.NoError(t, json.Unmarshal([]byte(assistant.ToolCalls[0].Function.Arguments), &args))
	assert.Equal(t, "errors", args["query"])

	// Tool result became role:"tool" with the matching tool_call_id.
	require.NotNil(t, toolReply, "no tool reply in request")
	assert.Equal(t, "call_xyz", toolReply.ToolCallID)
	assert.Equal(t, "no results", toolReply.Content)
}

// TestOpenRouter_ErrorMapping verifies that 401/402/429 surface as readable errors.
func TestOpenRouter_ErrorMapping(t *testing.T) {
	cases := []struct {
		name   string
		status int
		want   string
	}{
		{"unauthorized", 401, "unauthorized (401)"},
		{"insufficient credits", 402, "insufficient credits (402)"},
		{"rate limited", 429, "rate limited (429)"},
		{"generic 500", 500, "http 500"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				w.Write([]byte(`{"error":"boom"}`))
			}))
			defer server.Close()
			provider := NewOpenRouterProviderWithURL("k", server.URL)
			_, err := provider.ChatCompletion(context.Background(), ChatParams{
				Model:    "openai/gpt-4o",
				Messages: []ChatMessage{{Role: "user", TextContent: "hi"}},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}
