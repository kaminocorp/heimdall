package agent

import "fmt"

// ToolDefinition describes a tool available to the agent.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]ToolParam
}

// ToolParam describes a parameter for a tool.
type ToolParam struct {
	Type        string
	Description string
	Required    bool
}

// ToolRegistry returns all tools available to the agent.
func ToolRegistry() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "query_database",
			Description: "Execute a read-only SQL query against the monitored database",
			Parameters: map[string]ToolParam{
				"sql": {Type: "string", Description: "The SQL query to execute", Required: true},
			},
		},
		{
			Name:        "search_logs",
			Description: "Search recent logs for patterns or keywords",
			Parameters: map[string]ToolParam{
				"query":     {Type: "string", Description: "Search query", Required: true},
				"timeframe": {Type: "string", Description: "Time window (e.g. '30m', '6h')", Required: false},
			},
		},
		{
			Name:        "search_codebase",
			Description: "Search the connected codebase via GitHub for relevant code",
			Parameters: map[string]ToolParam{
				"query": {Type: "string", Description: "Search query (filename, symbol, or keyword)", Required: true},
			},
		},
		{
			Name:        "recall_similar_incidents",
			Description: "Query long-term memory for similar past incidents",
			Parameters: map[string]ToolParam{
				"description": {Type: "string", Description: "Description of the current incident to match against", Required: true},
			},
		},
		{
			Name:        "recall_lessons",
			Description: "Query long-term memory for lessons learned on a topic",
			Parameters: map[string]ToolParam{
				"topic": {Type: "string", Description: "The topic to recall lessons about", Required: true},
			},
		},
	}
}

// Dispatch routes a tool call to the appropriate implementation.
func (a *Agent) Dispatch(name string, input map[string]any) (string, error) {
	switch name {
	case "query_database":
		return a.toolQueryDatabase(input)
	case "search_logs":
		return a.toolSearchLogs(input)
	case "search_codebase":
		return a.toolSearchCodebase(input)
	case "recall_similar_incidents":
		return a.toolRecallSimilarIncidents(input)
	case "recall_lessons":
		return a.toolRecallLessons(input)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
