package agent

import (
	"context"
	"encoding/json"
)

// StopReason indicates why the model stopped generating.
type StopReason int

const (
	StopReasonEndTurn StopReason = iota
	StopReasonToolUse
)

// ChatParams holds everything needed for a model invocation. It is the
// provider-agnostic input type that the agent loop builds before handing off
// to a Provider implementation.
type ChatParams struct {
	Model        string
	SystemPrompt string
	Messages     []ChatMessage
	Tools        []ToolDef
	MaxTokens    int
}

// ChatMessage is a provider-agnostic message in the conversation.
//
// Three shapes:
//   - Plain user message:        Role="user",      TextContent="..."
//   - Plain assistant message:   Role="assistant", TextContent="..."
//   - Rich assistant response:   Role="assistant", Blocks=[text + tool_use]
//   - Tool result reply:         Role="user",      ToolResults=[...]
type ChatMessage struct {
	Role        string
	TextContent string
	Blocks      []ContentBlock
	ToolResults []ToolResult
}

// ContentBlock is a single block in an assistant response.
type ContentBlock struct {
	Type     string // "text" or "tool_use"
	Text     string // populated when Type == "text"
	ToolCall *ToolCall // populated when Type == "tool_use"
}

// ToolCall represents a model's request to invoke a tool.
type ToolCall struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// ToolResult is the outcome of executing a tool, sent back to the model.
type ToolResult struct {
	ToolCallID string
	Content    string
	IsError    bool
}

// ToolDef is a provider-agnostic tool definition. Each Provider implementation
// translates this to its native schema (Anthropic's input_schema, OpenAI's
// function-calling format, etc.).
type ToolDef struct {
	Name        string
	Description string
	Parameters  map[string]any // JSON Schema "properties" object
	Required    []string
}

// ChatResponse is the full response from a non-streaming invocation.
type ChatResponse struct {
	StopReason StopReason
	Blocks     []ContentBlock
	Usage      Usage
}

// Usage tracks token consumption for the invocation.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Provider is the interface implemented by each model backend (Anthropic
// direct, OpenRouter, ...). The agent loop talks only to this interface — no
// SDK-specific types leak past Provider implementations.
//
// A future ChatCompletionStream method will be added when streaming for
// interactive chat is implemented; the interface is intentionally minimal
// today to keep the Phase 1 refactor scope tight.
type Provider interface {
	ChatCompletion(ctx context.Context, params ChatParams) (*ChatResponse, error)
}
