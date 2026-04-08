package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ToolRegistry returns the provider-agnostic tool definitions available to the
// agent. Each Provider implementation translates these to its native schema.
func ToolRegistry() []ToolDef {
	return []ToolDef{
		{
			Name:        "search_logs",
			Description: "Search recent logs ingested by Heimdall for patterns, keywords, or severity levels. Returns matching log entries in reverse chronological order.",
			Parameters: map[string]any{
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
		{
			Name:        "query_database",
			Description: "Execute a read-only SQL query against a user's connected PostgreSQL database. Use this to investigate database state, check table contents, or run diagnostic queries.",
			Parameters: map[string]any{
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
		{
			Name:        "search_codebase",
			Description: "Search or read code in connected GitHub repositories. Use to understand application code, find error sources, or investigate configuration.",
			Parameters: map[string]any{
				"action": map[string]any{
					"type":        "string",
					"enum":        []string{"search_code", "read_file", "list_tree"},
					"description": "The operation to perform",
				},
				"query": map[string]any{
					"type":        "string",
					"description": "Search query (required for search_code)",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "File or directory path (required for read_file/list_tree)",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "Repository in owner/repo format. If omitted, searches all connected repos.",
				},
				"ref": map[string]any{
					"type":        "string",
					"description": "Git ref (branch/tag/SHA). Defaults to the repo's default branch.",
				},
			},
			Required: []string{"action"},
		},
	}
}

// Dispatch routes a tool call to the appropriate implementation.
func (a *Agent) Dispatch(ctx context.Context, userID uuid.UUID, appID uuid.UUID, name string, input map[string]any) (string, error) {
	switch name {
	case "search_logs":
		return a.toolSearchLogs(ctx, userID, input)
	case "query_database":
		return a.toolQueryDatabase(ctx, userID, input)
	case "search_codebase":
		return a.toolSearchCodebase(ctx, userID, appID, input)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
