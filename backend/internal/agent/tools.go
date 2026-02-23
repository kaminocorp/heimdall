package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/google/uuid"
)

// ToolRegistry returns the SDK-native tool definitions available to the agent.
func ToolRegistry() []anthropic.ToolUnionParam {
	return []anthropic.ToolUnionParam{
		{OfTool: &anthropic.ToolParam{
			Name:        "search_logs",
			Description: anthropic.String("Search recent logs ingested by Heimdall for patterns, keywords, or severity levels. Returns matching log entries in reverse chronological order."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query describing what to look for in the logs",
					},
					"severity": map[string]any{
						"type":        "string",
						"description": "Filter by severity level (e.g. critical, error, warning, info, debug)",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of log entries to return (default 20, max 200)",
					},
				},
				Required: []string{"query"},
			},
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "query_database",
			Description: anthropic.String("Execute a read-only SQL query against a user's connected PostgreSQL database. Use this to investigate database state, check table contents, or run diagnostic queries."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"sql": map[string]any{
						"type":        "string",
						"description": "The SQL query to execute (must be read-only — SELECT, EXPLAIN, etc.)",
					},
					"connection_id": map[string]any{
						"type":        "string",
						"description": "The UUID of the database connection to query",
					},
				},
				Required: []string{"sql", "connection_id"},
			},
		}},
	}
}

// Dispatch routes a tool call to the appropriate implementation.
func (a *Agent) Dispatch(ctx context.Context, userID uuid.UUID, name string, input map[string]any) (string, error) {
	switch name {
	case "search_logs":
		return a.toolSearchLogs(ctx, userID, input)
	case "query_database":
		return a.toolQueryDatabase(ctx, userID, input)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
