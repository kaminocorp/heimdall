package agent

import "context"

// RunLoop executes the agent's tool-use loop for a given input.
// It sends the context to Claude, processes tool calls, and returns
// the final text response.
func (a *Agent) RunLoop(ctx context.Context, input string) (string, error) {
	// TODO: implement Claude tool-use loop
	// 1. Build context (input + system state)
	// 2. Send to Claude with tools
	// 3. If tool_call: execute tool, append result, goto 2
	// 4. If text response: return
	return "Agent loop not yet implemented.", nil
}
