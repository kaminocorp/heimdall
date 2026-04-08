package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterProvider implements Provider against OpenRouter's
// OpenAI-compatible /chat/completions endpoint.
//
// We use raw HTTP rather than pulling in a second SDK for two reasons:
//  1. The surface we need is small (one endpoint, blocking only for now)
//  2. Avoids transitive dependency churn from a larger OpenAI SDK
//
// Streaming (SSE) will be added in the streaming follow-up — see
// docs/completions/openrouter-phase-1.md "Streaming follow-up" section.
type OpenRouterProvider struct {
	apiKey string
	http   *http.Client
	url    string // base URL, overridable for tests
}

const (
	openRouterDefaultURL = "https://openrouter.ai/api/v1/chat/completions"
	openRouterTimeout    = 2 * time.Minute
	openRouterReferer    = "https://heimdall.app"
	openRouterTitle      = "Heimdall"
	// openRouterMaxBodyBytes caps the response body we'll buffer in memory.
	// Real chat/completions responses are tens of KB; 10 MB is ~100× safety
	// margin. Anything larger is almost certainly a broken or hostile upstream,
	// and we'd rather error out than OOM the VM (Fly 2 GB with Lumber loaded).
	openRouterMaxBodyBytes = 10 << 20 // 10 MiB
)

// NewOpenRouterProvider builds a provider against the public OpenRouter URL.
func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
	return &OpenRouterProvider{
		apiKey: apiKey,
		http:   &http.Client{Timeout: openRouterTimeout},
		url:    openRouterDefaultURL,
	}
}

// NewOpenRouterProviderWithURL is for tests that point at a httptest.Server.
func NewOpenRouterProviderWithURL(apiKey, url string) *OpenRouterProvider {
	return &OpenRouterProvider{
		apiKey: apiKey,
		http:   &http.Client{Timeout: openRouterTimeout},
		url:    url,
	}
}

// --- OpenAI/OpenRouter wire types -------------------------------------------
//
// These are kept private to this file. They are NOT shared with the neutral
// Provider types — they exist only to (de)serialize the OpenRouter HTTP body.

type orRequest struct {
	Model      string      `json:"model"`
	MaxTokens  int         `json:"max_tokens,omitempty"`
	Messages   []orMessage `json:"messages"`
	Tools      []orTool    `json:"tools,omitempty"`
	ToolChoice string      `json:"tool_choice,omitempty"`
	Stream     bool        `json:"stream"`
}

type orMessage struct {
	Role       string       `json:"role"`                  // "system" | "user" | "assistant" | "tool"
	Content    string       `json:"content,omitempty"`     // text body (or empty when tool_calls present)
	ToolCalls  []orToolCall `json:"tool_calls,omitempty"`  // assistant → tool requests
	ToolCallID string       `json:"tool_call_id,omitempty"` // role:"tool" replies
	Name       string       `json:"name,omitempty"`
}

type orToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"` // always "function"
	Function orFunctionCall `json:"function"`
}

type orFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON-encoded string per OpenAI spec
}

type orTool struct {
	Type     string     `json:"type"` // always "function"
	Function orFunction `json:"function"`
}

type orFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"` // full JSON Schema object
}

type orResponse struct {
	ID      string     `json:"id"`
	Choices []orChoice `json:"choices"`
	Usage   orUsage    `json:"usage"`
}

type orChoice struct {
	Index        int       `json:"index"`
	Message      orMessage `json:"message"`
	FinishReason string    `json:"finish_reason"`
}

type orUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ---------------------------------------------------------------------------

// ChatCompletion sends a blocking request to OpenRouter and translates the
// response back to *ChatResponse.
func (p *OpenRouterProvider) ChatCompletion(ctx context.Context, params ChatParams) (*ChatResponse, error) {
	body, err := buildOpenRouterRequest(params)
	if err != nil {
		return nil, fmt.Errorf("openrouter: build request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openrouter: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("HTTP-Referer", openRouterReferer)
	req.Header.Set("X-Title", openRouterTitle)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter: http: %w", err)
	}
	defer resp.Body.Close()

	// Bound the response body so a broken/hostile upstream can't OOM us.
	// A legitimate response is tens of KB; see openRouterMaxBodyBytes.
	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, openRouterMaxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("openrouter: read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Surface specific failure modes the agent loop / operators care about.
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("openrouter: rate limited (429): %s", truncate(string(respBytes), 500))
		case http.StatusPaymentRequired:
			return nil, fmt.Errorf("openrouter: insufficient credits (402): %s", truncate(string(respBytes), 500))
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("openrouter: unauthorized (401): check OPENROUTER_API_KEY")
		default:
			return nil, fmt.Errorf("openrouter: http %d: %s", resp.StatusCode, truncate(string(respBytes), 500))
		}
	}

	var orResp orResponse
	if err := json.Unmarshal(respBytes, &orResp); err != nil {
		return nil, fmt.Errorf("openrouter: decode response: %w", err)
	}

	return translateOpenRouterResponse(&orResp)
}

// buildOpenRouterRequest translates the neutral ChatParams into the OpenAI
// wire format. The translation is deterministic and one-pass.
func buildOpenRouterRequest(params ChatParams) ([]byte, error) {
	out := orRequest{
		Model:     params.Model,
		MaxTokens: params.MaxTokens,
		Stream:    false,
	}

	// System prompt becomes the first message (vs. Anthropic's separate field).
	if params.SystemPrompt != "" {
		out.Messages = append(out.Messages, orMessage{
			Role:    "system",
			Content: params.SystemPrompt,
		})
	}

	for _, m := range params.Messages {
		switch m.Role {
		case "user":
			if len(m.ToolResults) > 0 {
				// Each tool result becomes a separate role:"tool" message.
				for _, tr := range m.ToolResults {
					content := tr.Content
					if tr.IsError {
						// OpenAI/OpenRouter has no isError flag — prefix the
						// content so the model can see it failed.
						content = "ERROR: " + tr.Content
					}
					out.Messages = append(out.Messages, orMessage{
						Role:       "tool",
						ToolCallID: tr.ToolCallID,
						Content:    content,
					})
				}
			} else {
				out.Messages = append(out.Messages, orMessage{
					Role:    "user",
					Content: m.TextContent,
				})
			}
		case "assistant":
			if len(m.Blocks) > 0 {
				// Rich assistant message: collect text into content, tool_use
				// blocks into tool_calls.
				var textParts []byte
				var toolCalls []orToolCall
				for _, b := range m.Blocks {
					switch b.Type {
					case "text":
						if len(textParts) > 0 {
							textParts = append(textParts, '\n')
						}
						textParts = append(textParts, []byte(b.Text)...)
					case "tool_use":
						if b.ToolCall == nil {
							continue
						}
						// OpenAI spec: arguments is a *string* of JSON, not an object.
						args := string(b.ToolCall.Input)
						if args == "" {
							args = "{}"
						}
						toolCalls = append(toolCalls, orToolCall{
							ID:   b.ToolCall.ID,
							Type: "function",
							Function: orFunctionCall{
								Name:      b.ToolCall.Name,
								Arguments: args,
							},
						})
					}
				}
				out.Messages = append(out.Messages, orMessage{
					Role:      "assistant",
					Content:   string(textParts),
					ToolCalls: toolCalls,
				})
			} else {
				out.Messages = append(out.Messages, orMessage{
					Role:    "assistant",
					Content: m.TextContent,
				})
			}
		}
	}

	for _, t := range params.Tools {
		// OpenAI expects a full JSON Schema object under parameters, not just
		// the properties bag. Wrap it.
		schema := map[string]any{
			"type":       "object",
			"properties": t.Parameters,
		}
		if len(t.Required) > 0 {
			schema["required"] = t.Required
		}
		out.Tools = append(out.Tools, orTool{
			Type: "function",
			Function: orFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  schema,
			},
		})
	}

	return json.Marshal(out)
}

// translateOpenRouterResponse converts the wire response into *ChatResponse.
func translateOpenRouterResponse(resp *orResponse) (*ChatResponse, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openrouter: response has no choices")
	}
	choice := resp.Choices[0]

	out := &ChatResponse{
		Usage: Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	// Map finish_reason → StopReason. OpenAI uses "tool_calls" when the model
	// wants to invoke tools; everything else (stop, length, content_filter, ...)
	// collapses to EndTurn so the loop returns whatever text exists.
	if choice.FinishReason == "tool_calls" {
		out.StopReason = StopReasonToolUse
	} else {
		out.StopReason = StopReasonEndTurn
	}

	// Text content first (if any).
	if choice.Message.Content != "" {
		out.Blocks = append(out.Blocks, ContentBlock{
			Type: "text",
			Text: choice.Message.Content,
		})
	}

	// Then tool_use blocks. The agent loop iterates Blocks in order, and
	// expects tool_use to come after any text — same shape Anthropic returns.
	for _, tc := range choice.Message.ToolCalls {
		// arguments is a JSON-encoded string per OpenAI spec; pass it through
		// as RawMessage so the agent loop's json.Unmarshal works unchanged.
		args := json.RawMessage(tc.Function.Arguments)
		if len(args) == 0 {
			args = json.RawMessage("{}")
		}
		out.Blocks = append(out.Blocks, ContentBlock{
			Type: "tool_use",
			ToolCall: &ToolCall{
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: args,
			},
		})
	}

	return out, nil
}

// truncate caps a string at n runes for error messages. Rune-safe so we
// don't emit invalid UTF-8 when an upstream error body contains multi-byte
// characters (which would otherwise be silently mangled by json.Marshal).
func truncate(s string, n int) string {
	if len(s) <= n {
		// Fast path: byte length <= n implies rune count <= n.
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
