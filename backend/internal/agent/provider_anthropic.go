package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider implements Provider against the official anthropic-sdk-go.
// It is the only place in the agent package that imports the SDK directly —
// every other file in the package speaks the neutral provider.go types.
type AnthropicProvider struct {
	client *anthropic.Client
}

// NewAnthropicProvider creates a provider with a real client.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicProvider{client: &c}
}

// NewAnthropicProviderWithClient lets tests inject a client wired to a mock
// HTTP server (see loop_test.go). Not for production use.
func NewAnthropicProviderWithClient(client *anthropic.Client) *AnthropicProvider {
	return &AnthropicProvider{client: client}
}

// ChatCompletion translates ChatParams → anthropic.MessageNewParams, calls the
// SDK, and translates the response back to *ChatResponse.
func (p *AnthropicProvider) ChatCompletion(ctx context.Context, params ChatParams) (*ChatResponse, error) {
	// params.Model is always populated by the agent loop (defaultModelID kicks
	// in there when no per-app / global config row is found). We don't second-
	// guess it here — passing an empty string through would be a caller bug,
	// not something the provider should paper over with a different default.
	model := anthropic.Model(params.Model)

	maxTokens := int64(params.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 4096
	}

	messages := make([]anthropic.MessageParam, 0, len(params.Messages))
	for _, m := range params.Messages {
		switch m.Role {
		case "user":
			if len(m.ToolResults) > 0 {
				blocks := make([]anthropic.ContentBlockParamUnion, 0, len(m.ToolResults))
				for _, tr := range m.ToolResults {
					blocks = append(blocks, anthropic.NewToolResultBlock(tr.ToolCallID, tr.Content, tr.IsError))
				}
				messages = append(messages, anthropic.NewUserMessage(blocks...))
			} else {
				messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(m.TextContent)))
			}
		case "assistant":
			if len(m.Blocks) > 0 {
				blocks := make([]anthropic.ContentBlockParamUnion, 0, len(m.Blocks))
				for _, b := range m.Blocks {
					switch b.Type {
					case "text":
						blocks = append(blocks, anthropic.NewTextBlock(b.Text))
					case "tool_use":
						if b.ToolCall == nil {
							continue
						}
						blocks = append(blocks, anthropic.ContentBlockParamUnion{
							OfToolUse: &anthropic.ToolUseBlockParam{
								ID:    b.ToolCall.ID,
								Name:  b.ToolCall.Name,
								Input: b.ToolCall.Input,
							},
						})
					}
				}
				messages = append(messages, anthropic.NewAssistantMessage(blocks...))
			} else {
				messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.TextContent)))
			}
		}
	}

	tools := make([]anthropic.ToolUnionParam, 0, len(params.Tools))
	for _, t := range params.Tools {
		tools = append(tools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: t.Parameters,
					Required:   t.Required,
				},
			},
		})
	}

	resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: maxTokens,
		System: []anthropic.TextBlockParam{
			{Text: params.SystemPrompt},
		},
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic: %w", err)
	}

	out := &ChatResponse{
		Usage: Usage{
			InputTokens:  int(resp.Usage.InputTokens),
			OutputTokens: int(resp.Usage.OutputTokens),
		},
	}

	switch resp.StopReason {
	case anthropic.StopReasonEndTurn:
		out.StopReason = StopReasonEndTurn
	case anthropic.StopReasonToolUse:
		out.StopReason = StopReasonToolUse
	default:
		// Unknown stop reasons (max_tokens, stop_sequence, ...) are surfaced
		// as EndTurn so the loop returns whatever text was produced. The agent
		// loop never relied on distinguishing them.
		out.StopReason = StopReasonEndTurn
	}

	for _, block := range resp.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			out.Blocks = append(out.Blocks, ContentBlock{Type: "text", Text: v.Text})
		case anthropic.ToolUseBlock:
			out.Blocks = append(out.Blocks, ContentBlock{
				Type: "tool_use",
				ToolCall: &ToolCall{
					ID:    v.ID,
					Name:  v.Name,
					Input: v.Input,
				},
			})
		}
	}

	return out, nil
}
