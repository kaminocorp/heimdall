//go:build toolvet

package agent

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestToolUseVetting runs a real agent loop against every model in the
// catalogue, verifying that each can handle multi-turn tool use. The test
// sends a task that requires the model to call search_logs at least once,
// interpret the result, and produce a final answer.
//
// This is an integration test that hits live APIs. It is gated behind the
// "toolvet" build tag so it never runs in CI by default.
//
// Run manually:
//
//	cd backend && go test -tags=toolvet -timeout=600s -v ./internal/agent/ -run TestToolUseVetting
//
// Required env vars: ANTHROPIC_API_KEY, OPENROUTER_API_KEY
// Cost: ~$0.10 per full run across all models.
func TestToolUseVetting(t *testing.T) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")

	if anthropicKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping tool-use vetting")
	}

	anthropicProvider := NewAnthropicProvider(anthropicKey)
	var openrouterProvider *OpenRouterProvider
	if openrouterKey != "" {
		openrouterProvider = NewOpenRouterProvider(openrouterKey)
	}

	// The tool definition: a fake search_logs that returns canned data.
	tools := []ToolDef{
		{
			Name:        "search_logs",
			Description: "Search application logs. Returns matching log entries.",
			Parameters: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query to match against log messages",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results to return",
				},
			},
			Required: []string{"query"},
		},
	}

	// Canned log data returned by the fake tool.
	cannedResult := `[
		{"timestamp": "2026-04-12T10:15:00Z", "level": "ERROR", "message": "connection pool exhausted: max_connections=100, active=100, waiting=23", "service": "api-gateway"},
		{"timestamp": "2026-04-12T10:15:01Z", "level": "WARN", "message": "request timeout after 30s: POST /api/payments", "service": "api-gateway"},
		{"timestamp": "2026-04-12T10:14:58Z", "level": "ERROR", "message": "connection pool exhausted: max_connections=100, active=100, waiting=19", "service": "api-gateway"}
	]`

	systemPrompt := "You are a monitoring agent. When asked to investigate, use search_logs to find relevant entries, then summarise your findings concisely."
	userMessage := "Investigate any recent errors in the api-gateway service. Search the logs and tell me what you find."

	// Build the full catalogue, gating OpenRouter entries on key availability.
	models := make([]ModelOption, 0, len(AnthropicModels)+len(OpenRouterModels))
	models = append(models, AnthropicModels...)
	if openrouterKey != "" {
		models = append(models, OpenRouterModels...)
	} else {
		t.Log("OPENROUTER_API_KEY not set, skipping OpenRouter models")
	}

	for _, model := range models {
		model := model // capture loop var
		t.Run(model.ID, func(t *testing.T) {
			// Resolve provider for this model.
			var provider Provider
			switch model.Provider {
			case "anthropic":
				provider = anthropicProvider
			case "openrouter":
				if openrouterProvider == nil {
					t.Skip("no OpenRouter provider available")
				}
				provider = openrouterProvider
			default:
				t.Fatalf("unknown provider %q for model %s", model.Provider, model.ID)
			}

			// Run up to 5 turns. The model should:
			// 1. Call search_logs (at least once)
			// 2. Get the canned result back
			// 3. Produce a final text answer
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			messages := []ChatMessage{
				{Role: "user", TextContent: userMessage},
			}

			toolCallCount := 0
			gotFinalAnswer := false
			maxTurns := 5

			for turn := 0; turn < maxTurns; turn++ {
				resp, err := provider.ChatCompletion(ctx, ChatParams{
					Model:        model.ID,
					SystemPrompt: systemPrompt,
					Messages:     messages,
					Tools:        tools,
					MaxTokens:    1024,
				})
				if err != nil {
					t.Fatalf("turn %d: ChatCompletion error: %v", turn, err)
				}

				if resp.StopReason == StopReasonToolUse {
					// Model wants to call tools. Process each tool_use block.
					var toolResults []ToolResult
					for _, block := range resp.Blocks {
						if block.Type == "tool_use" && block.ToolCall != nil {
							if block.ToolCall.Name != "search_logs" {
								t.Errorf("turn %d: unexpected tool call %q (expected search_logs)", turn, block.ToolCall.Name)
							}
							toolCallCount++
							toolResults = append(toolResults, ToolResult{
								ToolCallID: block.ToolCall.ID,
								Content:    cannedResult,
							})
						}
					}

					// Append assistant message with blocks, then tool results.
					messages = append(messages, ChatMessage{
						Role:   "assistant",
						Blocks: resp.Blocks,
					})
					messages = append(messages, ChatMessage{
						Role:        "user",
						ToolResults: toolResults,
					})
				} else {
					// End turn — model produced a final answer.
					gotFinalAnswer = true

					// Extract text for logging.
					var text string
					for _, block := range resp.Blocks {
						if block.Type == "text" {
							text = block.Text
							break
						}
					}
					t.Logf("model=%s tool_calls=%d answer_length=%d answer_preview=%s",
						model.ID, toolCallCount, len(text), truncateForLog(text, 120))
					break
				}
			}

			if toolCallCount == 0 {
				t.Errorf("model %s made 0 tool calls — expected at least 1 search_logs call", model.ID)
			}
			if !gotFinalAnswer {
				t.Errorf("model %s did not produce a final answer within %d turns", model.ID, maxTurns)
			}

			// Log token usage of the final turn for cost estimation.
			t.Logf("model=%s pricing=$%.2f/M in, $%.2f/M out",
				model.ID, model.Pricing.Prompt, model.Pricing.Completion)
		})
	}
}

// truncateForLog caps a string for test log output.
func truncateForLog(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// TestToolUseVettingCannedResultIsValidJSON is a quick sanity check that the
// canned data we feed to models during vetting is parseable JSON.
func TestToolUseVettingCannedResultIsValidJSON(t *testing.T) {
	canned := `[
		{"timestamp": "2026-04-12T10:15:00Z", "level": "ERROR", "message": "connection pool exhausted: max_connections=100, active=100, waiting=23", "service": "api-gateway"},
		{"timestamp": "2026-04-12T10:15:01Z", "level": "WARN", "message": "request timeout after 30s: POST /api/payments", "service": "api-gateway"},
		{"timestamp": "2026-04-12T10:14:58Z", "level": "ERROR", "message": "connection pool exhausted: max_connections=100, active=100, waiting=19", "service": "api-gateway"}
	]`
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(canned), &parsed); err != nil {
		t.Fatalf("canned log data is not valid JSON: %v", err)
	}
	if len(parsed) != 3 {
		t.Errorf("expected 3 log entries, got %d", len(parsed))
	}
}
