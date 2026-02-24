package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/google/uuid"
)

const maxIterations = 10

// RunLoop executes the agent's tool-use loop for a single input with no prior history.
// It delegates to RunConversation with an empty history.
func (a *Agent) RunLoop(ctx context.Context, userID uuid.UUID, input string) (string, error) {
	return a.RunConversation(ctx, userID, nil, nil, input)
}

// RunConversation executes the agent's tool-use loop with full conversation history.
// Prior messages are converted to Claude message params so the agent has multi-turn context.
// conversationID is optional — when provided, agent observations are emitted to agent_log.
func (a *Agent) RunConversation(ctx context.Context, userID uuid.UUID, conversationID *uuid.UUID, history []Message, input string) (string, error) {
	// Load agent config from DB (fallback to defaults if no row).
	model := anthropic.ModelClaudeSonnet4_5
	systemOverride := ""

	cfg, err := a.queries.GetAgentConfig(ctx)
	if err == nil {
		if cfg.Model != "" {
			model = anthropic.Model(cfg.Model)
		}
		if cfg.SystemPromptOverride.Valid {
			systemOverride = cfg.SystemPromptOverride.String
		}
	}

	sysPrompt := BuildSystemPrompt(systemOverride)
	tools := ToolRegistry()

	// Build messages from conversation history + new input.
	var messages []anthropic.MessageParam
	for _, msg := range history {
		switch msg.Role {
		case "user":
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case "assistant":
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))

	for i := range maxIterations {
		slog.Info("agent loop iteration", "iteration", i+1, "user_id", userID)

		resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     model,
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: sysPrompt},
			},
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return "", fmt.Errorf("agent loop: claude api: %w", err)
		}

		// If the model is done talking, extract the text response.
		if resp.StopReason == anthropic.StopReasonEndTurn {
			text := extractText(resp)

			// Emit observation for the final agent response.
			summary := text
			if len(summary) > 200 {
				summary = summary[:200] + "..."
			}
			a.EmitLog(ctx, userID, conversationID, "observation", summary, nil)

			return text, nil
		}

		// If the model wants to use tools, process each tool call.
		if resp.StopReason == anthropic.StopReasonToolUse {
			// Append the assistant's response (contains tool_use blocks).
			var assistantBlocks []anthropic.ContentBlockParamUnion
			for _, block := range resp.Content {
				switch v := block.AsAny().(type) {
				case anthropic.TextBlock:
					assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(v.Text))
				case anthropic.ToolUseBlock:
					assistantBlocks = append(assistantBlocks, anthropic.ContentBlockParamUnion{
						OfToolUse: &anthropic.ToolUseBlockParam{
							ID:    v.ID,
							Name:  v.Name,
							Input: v.Input,
						},
					})
				}
			}
			messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))

			// Execute each tool and collect results.
			var toolResults []anthropic.ContentBlockParamUnion
			for _, block := range resp.Content {
				if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
					var toolInput map[string]any
					if err := json.Unmarshal(tu.Input, &toolInput); err != nil {
						toolResults = append(toolResults, anthropic.NewToolResultBlock(
							tu.ID, fmt.Sprintf("failed to parse tool input: %v", err), true,
						))
						continue
					}

					// Emit tool_call log entry.
					a.EmitLog(ctx, userID, conversationID, "tool_call",
						fmt.Sprintf("Called %s", tu.Name),
						map[string]any{"tool": tu.Name, "input": toolInput},
					)

					slog.Info("agent dispatching tool", "tool", tu.Name, "user_id", userID)
					result, err := a.Dispatch(ctx, userID, tu.Name, toolInput)
					if err != nil {
						slog.Warn("agent tool error", "tool", tu.Name, "err", err)
						toolResults = append(toolResults, anthropic.NewToolResultBlock(
							tu.ID, fmt.Sprintf("Tool error: %v", err), true,
						))

						// Emit tool_result error entry.
						a.EmitLog(ctx, userID, conversationID, "tool_result",
							fmt.Sprintf("%s failed: %v", tu.Name, err),
							map[string]any{"tool": tu.Name, "error": err.Error()},
						)
					} else {
						toolResults = append(toolResults, anthropic.NewToolResultBlock(
							tu.ID, result, false,
						))

						// Emit tool_result success entry.
						resultSummary := result
						if len(resultSummary) > 200 {
							resultSummary = resultSummary[:200] + "..."
						}
						a.EmitLog(ctx, userID, conversationID, "tool_result",
							fmt.Sprintf("%s returned results", tu.Name),
							map[string]any{"tool": tu.Name, "result_preview": resultSummary},
						)
					}
				}
			}
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
			continue
		}

		// Unexpected stop reason — return whatever text we have.
		return extractText(resp), nil
	}

	return "", fmt.Errorf("agent loop: exceeded max iterations (%d)", maxIterations)
}

// extractText concatenates all text blocks from a Claude response.
func extractText(resp *anthropic.Message) string {
	var text string
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			text += tb.Text
		}
	}
	return text
}
