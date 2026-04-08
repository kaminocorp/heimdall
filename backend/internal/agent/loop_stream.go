package agent

import (
	"context"

	"github.com/google/uuid"
)

// AgentEvent is a single progress event emitted by RunConversationStream as
// the tool-use loop iterates. The chat WebSocket handler relays these to the
// browser so the user can see what the agent is doing in real time.
//
// Today the channel emits coarse-grained tool transitions and a single final
// "message" event carrying the full assistant text. Token-by-token streaming
// is intentionally out of scope — see docs/executing/streaming-implementation.md
// for the follow-up plan to upgrade this to per-token deltas without changing
// the consumer contract on chat.go or useAgent.ts.
type AgentEvent struct {
	// Type is one of: "tool_start", "tool_result", "message", "error".
	Type string

	// Tool is populated for "tool_start" and "tool_result".
	Tool string

	// Content carries the full assistant response on "message" and the error
	// text on "error". Empty for tool events.
	Content string

	// IsError is set on "tool_result" when the tool call failed. The frontend
	// can render the indicator in an error state if it wants; today the chat
	// UI just clears the active-tools list either way.
	IsError bool
}

// streamBufferSize bounds the channel between the producer goroutine and the
// WebSocket consumer. 16 is comfortably above the typical event count for one
// loop iteration (one tool_start + one tool_result per tool call, max 10
// iterations × ~3 tool calls each). The producer would only block on a slow
// or stalled WS reader, which is the right back-pressure behaviour.
const streamBufferSize = 16

// RunConversationStream is the streaming counterpart to RunConversation. It
// returns a channel that emits AgentEvent values as the loop iterates and
// closes once the loop terminates (either with a final "message" event or an
// "error" event). The caller must drain the channel — abandoning it would
// leak the producer goroutine if the buffer fills.
//
// The underlying LLM call is still blocking. The streaming here is purely
// over the loop's progress: tool_start fires immediately before each
// Dispatch call, tool_result fires immediately after, and the final message
// event lands once the model returns StopReasonEndTurn. Token-by-token
// streaming inside a single ChatCompletion call is the next iteration —
// see docs/executing/streaming-implementation.md.
func (a *Agent) RunConversationStream(ctx context.Context, userID uuid.UUID, appID uuid.UUID, conversationID *uuid.UUID, history []Message, input string) <-chan AgentEvent {
	ch := make(chan AgentEvent, streamBufferSize)

	go func() {
		defer close(ch)

		text, err := a.runConversationCore(ctx, userID, appID, conversationID, history, input, ch)
		// Terminal sends are ctx-aware for the same reason emitEvent is: if
		// the WebSocket consumer has already walked away, the buffer may be
		// full and a blind send would pin this goroutine indefinitely.
		var final AgentEvent
		if err != nil {
			final = AgentEvent{Type: "error", Content: err.Error()}
		} else {
			final = AgentEvent{Type: "message", Content: text}
		}
		select {
		case ch <- final:
		case <-ctx.Done():
		}
	}()

	return ch
}
