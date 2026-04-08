package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

const (
	maxIterations    = 10
	defaultModelID   = "claude-sonnet-4-6"
	defaultMaxTokens = 4096
)

// RunLoop executes the agent's tool-use loop for a single input with no prior history.
// It delegates to RunConversation with an empty history.
func (a *Agent) RunLoop(ctx context.Context, userID uuid.UUID, input string) (string, error) {
	return a.RunConversation(ctx, userID, uuid.Nil, nil, nil, input)
}

// RunConversation executes the agent's tool-use loop with full conversation history.
// This is the blocking entry point — callers that want progress events as the
// loop iterates (tool start/result) should use RunConversationStream instead.
// Internally both methods funnel through runConversationCore.
func (a *Agent) RunConversation(ctx context.Context, userID uuid.UUID, appID uuid.UUID, conversationID *uuid.UUID, history []Message, input string) (string, error) {
	return a.runConversationCore(ctx, userID, appID, conversationID, history, input, nil)
}

// emitEvent is a nil-safe, ctx-aware send used by runConversationCore. When
// events is nil (the blocking RunConversation path), all calls are no-ops,
// which keeps the loop body identical between blocking and streaming callers.
//
// The ctx branch is load-bearing: if the WebSocket consumer disconnects
// mid-loop, chat.go stops draining the channel and the producer would
// otherwise block here forever once the 16-slot buffer filled. A chatty loop
// (10 iterations × several tools × 2 events each) can exceed that buffer, so
// "small enough to never stall" isn't a safe assumption. Selecting on
// ctx.Done() drops the event rather than leaking the producer goroutine —
// the user's already gone, so dropped progress events are harmless.
func emitEvent(ctx context.Context, events chan<- AgentEvent, ev AgentEvent) {
	if events == nil {
		return
	}
	select {
	case events <- ev:
	case <-ctx.Done():
	}
}

// runConversationCore is the shared loop body. When events is non-nil, it
// emits AgentEvent values for tool start/result transitions; the channel is
// owned and closed by the caller (RunConversationStream).
//
// conversationID is optional — when provided, agent observations are emitted to agent_log.
// appID is optional — when provided, enables app-scoped tools like search_codebase.
func (a *Agent) runConversationCore(ctx context.Context, userID uuid.UUID, appID uuid.UUID, conversationID *uuid.UUID, history []Message, input string, events chan<- AgentEvent) (string, error) {
	// Load agent config: prefer per-app config when appID is provided,
	// fall back to global agent_config singleton.
	model := defaultModelID
	systemOverride := ""
	providerName := defaultProviderName

	if appID != uuid.Nil {
		appCfg, err := a.queries.GetAppAgentConfig(ctx, appID)
		if err == nil {
			if appCfg.Model != "" {
				model = appCfg.Model
			}
			if appCfg.SystemPromptOverride.Valid {
				systemOverride = appCfg.SystemPromptOverride.String
			}
			if appCfg.Provider != "" {
				providerName = appCfg.Provider
			}
		}
	} else {
		cfg, err := a.queries.GetAgentConfig(ctx)
		if err == nil {
			if cfg.Model != "" {
				model = cfg.Model
			}
			if cfg.SystemPromptOverride.Valid {
				systemOverride = cfg.SystemPromptOverride.String
			}
		}
	}

	provider := a.providerFor(providerName)
	sysPrompt := BuildSystemPrompt(systemOverride)
	tools := ToolRegistry()

	// Build provider-agnostic messages from history + new input.
	var messages []ChatMessage
	for _, msg := range history {
		switch msg.Role {
		case "user":
			messages = append(messages, ChatMessage{Role: "user", TextContent: msg.Content})
		case "assistant":
			messages = append(messages, ChatMessage{Role: "assistant", TextContent: msg.Content})
		}
	}
	messages = append(messages, ChatMessage{Role: "user", TextContent: input})

	for i := range maxIterations {
		slog.Info("agent loop iteration", "iteration", i+1, "user_id", userID)

		resp, err := provider.ChatCompletion(ctx, ChatParams{
			Model:        model,
			SystemPrompt: sysPrompt,
			Messages:     messages,
			Tools:        tools,
			MaxTokens:    defaultMaxTokens,
		})
		if err != nil {
			return "", fmt.Errorf("agent loop: provider: %w", err)
		}

		// If the model is done talking, extract the text response.
		if resp.StopReason == StopReasonEndTurn {
			text := extractText(resp)

			// Emit observation for the final agent response.
			summary := text
			if utf8.RuneCountInString(summary) > 200 {
				summary = string([]rune(summary)[:200]) + "..."
			}
			a.EmitLog(ctx, userID, conversationID, "observation", summary, nil)

			return text, nil
		}

		// If the model wants to use tools, process each tool call.
		if resp.StopReason == StopReasonToolUse {
			// Append the assistant's response unchanged — its blocks (text +
			// tool_use) get round-tripped back to the provider on the next call.
			messages = append(messages, ChatMessage{Role: "assistant", Blocks: resp.Blocks})

			// Execute each tool and collect results.
			var toolResults []ToolResult
			for _, block := range resp.Blocks {
				if block.Type != "tool_use" || block.ToolCall == nil {
					continue
				}
				tu := block.ToolCall

				var toolInput map[string]any
				if err := json.Unmarshal(tu.Input, &toolInput); err != nil {
					toolResults = append(toolResults, ToolResult{
						ToolCallID: tu.ID,
						Content:    fmt.Sprintf("failed to parse tool input: %v", err),
						IsError:    true,
					})
					continue
				}

				// Emit tool_call log entry.
				a.EmitLog(ctx, userID, conversationID, "tool_call",
					fmt.Sprintf("Called %s", tu.Name),
					map[string]any{"tool": tu.Name, "input": toolInput},
				)

				// Stream tool_start to the WS client (nil-safe in blocking mode).
				emitEvent(ctx, events, AgentEvent{Type: "tool_start", Tool: tu.Name})

				slog.Info("agent dispatching tool", "tool", tu.Name, "user_id", userID)
				result, err := a.Dispatch(ctx, userID, appID, tu.Name, toolInput)
				if err != nil {
					slog.Warn("agent tool error", "tool", tu.Name, "err", err)
					toolResults = append(toolResults, ToolResult{
						ToolCallID: tu.ID,
						Content:    fmt.Sprintf("Tool error: %v", err),
						IsError:    true,
					})

					// Emit tool_result error entry.
					a.EmitLog(ctx, userID, conversationID, "tool_result",
						fmt.Sprintf("%s failed: %v", tu.Name, err),
						map[string]any{"tool": tu.Name, "error": err.Error()},
					)
					emitEvent(ctx, events, AgentEvent{Type: "tool_result", Tool: tu.Name, IsError: true})
				} else {
					toolResults = append(toolResults, ToolResult{
						ToolCallID: tu.ID,
						Content:    result,
					})

					// Emit tool_result success entry.
					resultSummary := result
					if utf8.RuneCountInString(resultSummary) > 200 {
						resultSummary = string([]rune(resultSummary)[:200]) + "..."
					}
					a.EmitLog(ctx, userID, conversationID, "tool_result",
						fmt.Sprintf("%s returned results", tu.Name),
						map[string]any{"tool": tu.Name, "result_preview": resultSummary},
					)
					emitEvent(ctx, events, AgentEvent{Type: "tool_result", Tool: tu.Name})
				}
			}
			messages = append(messages, ChatMessage{Role: "user", ToolResults: toolResults})
			continue
		}

		// Unexpected stop reason — return whatever text we have.
		return extractText(resp), nil
	}

	return "", fmt.Errorf("agent loop: exceeded max iterations (%d)", maxIterations)
}

// RunMonitoring executes the agent's tool-use loop for monitoring mode.
// Unlike RunConversation, this is sessionless (no conversation persistence),
// uses the monitoring-specific system prompt, and loads per-app agent config.
// Returns (assessment text, severity string).
//
// IMPORTANT: This path MUST stay blocking (non-streaming). The 0.27.0 rate
// limiter wraps each invocation as a single atomic unit; streaming would
// muddy that contract. See docs/executing/openrouter-implementation.md
// (Phase 1, Task 1.4) for the full reasoning.
func (a *Agent) RunMonitoring(ctx context.Context, userID uuid.UUID, appConfig db.AppAgentConfig, flaggedLogs string) (string, string) {
	model := appConfig.Model
	if model == "" {
		model = defaultModelID
	}

	// Monitoring mode resolves the provider from app_agent_config.provider,
	// which Phase 3 added with a default of 'anthropic'. Empty/unknown values
	// fall back to the default via providerFor.
	provider := a.providerFor(appConfig.Provider)

	override := ""
	if appConfig.SystemPromptOverride.Valid {
		override = appConfig.SystemPromptOverride.String
	}
	sysPrompt := BuildMonitoringPrompt(override)
	tools := ToolRegistry()

	messages := []ChatMessage{
		{Role: "user", TextContent: flaggedLogs},
	}

	for i := range maxIterations {
		slog.Info("monitoring loop iteration", "iteration", i+1, "app_id", appConfig.AppID)

		resp, err := provider.ChatCompletion(ctx, ChatParams{
			Model:        model,
			SystemPrompt: sysPrompt,
			Messages:     messages,
			Tools:        tools,
			MaxTokens:    defaultMaxTokens,
		})
		if err != nil {
			slog.Error("monitoring loop: provider error", "err", err, "app_id", appConfig.AppID)
			return "Monitoring assessment failed: provider error", "error"
		}

		if resp.StopReason == StopReasonEndTurn {
			text := extractText(resp)
			severity := parseSeverityFromResponse(text)
			return text, severity
		}

		if resp.StopReason == StopReasonToolUse {
			messages = append(messages, ChatMessage{Role: "assistant", Blocks: resp.Blocks})

			var toolResults []ToolResult
			for _, block := range resp.Blocks {
				if block.Type != "tool_use" || block.ToolCall == nil {
					continue
				}
				tu := block.ToolCall

				var toolInput map[string]any
				if err := json.Unmarshal(tu.Input, &toolInput); err != nil {
					toolResults = append(toolResults, ToolResult{
						ToolCallID: tu.ID,
						Content:    fmt.Sprintf("failed to parse tool input: %v", err),
						IsError:    true,
					})
					continue
				}

				a.EmitLog(ctx, userID, nil, "tool_call",
					fmt.Sprintf("Monitoring: called %s", tu.Name),
					map[string]any{"tool": tu.Name, "input": toolInput, "app_id": appConfig.AppID},
				)

				slog.Info("monitoring dispatching tool", "tool", tu.Name, "app_id", appConfig.AppID)
				result, err := a.Dispatch(ctx, userID, appConfig.AppID, tu.Name, toolInput)
				if err != nil {
					slog.Warn("monitoring tool error", "tool", tu.Name, "err", err)
					toolResults = append(toolResults, ToolResult{
						ToolCallID: tu.ID,
						Content:    fmt.Sprintf("Tool error: %v", err),
						IsError:    true,
					})
					a.EmitLog(ctx, userID, nil, "tool_result",
						fmt.Sprintf("Monitoring: %s failed: %v", tu.Name, err),
						map[string]any{"tool": tu.Name, "error": err.Error(), "app_id": appConfig.AppID},
					)
				} else {
					toolResults = append(toolResults, ToolResult{
						ToolCallID: tu.ID,
						Content:    result,
					})
					resultSummary := result
					if utf8.RuneCountInString(resultSummary) > 200 {
						resultSummary = string([]rune(resultSummary)[:200]) + "..."
					}
					a.EmitLog(ctx, userID, nil, "tool_result",
						fmt.Sprintf("Monitoring: %s returned results", tu.Name),
						map[string]any{"tool": tu.Name, "result_preview": resultSummary, "app_id": appConfig.AppID},
					)
				}
			}
			messages = append(messages, ChatMessage{Role: "user", ToolResults: toolResults})
			continue
		}

		return extractText(resp), "info"
	}

	return "Monitoring assessment exceeded max iterations", "warning"
}

// parseSeverityFromResponse attempts to extract a severity keyword from
// the agent's response text. Defaults to "info" if not found.
func parseSeverityFromResponse(text string) string {
	lower := strings.ToLower(text)
	// Check for explicit severity markers in order of priority.
	for _, sev := range []string{"critical", "error", "warning", "info"} {
		if strings.Contains(lower, "severity: "+sev) || strings.Contains(lower, "severity:"+sev) {
			return sev
		}
	}
	// Fall back to heuristic keyword scan.
	for _, sev := range []string{"critical", "error", "warning"} {
		if strings.Contains(lower, sev) {
			return sev
		}
	}
	return "info"
}

// extractText concatenates all text blocks from a provider response.
func extractText(resp *ChatResponse) string {
	var text string
	for _, block := range resp.Blocks {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return text
}
