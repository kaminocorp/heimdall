package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/connectors/codebase"
)

func (a *Agent) toolSearchCodebase(ctx context.Context, userID uuid.UUID, appID uuid.UUID, input map[string]any) (string, error) {
	if a.githubClient == nil {
		return "", fmt.Errorf("search_codebase: GitHub App not configured on this server")
	}

	if appID == uuid.Nil {
		return "", fmt.Errorf("search_codebase: no application context — connect via an app to use codebase search")
	}

	action, _ := input["action"].(string)
	if action == "" {
		return "", fmt.Errorf("search_codebase: action parameter is required")
	}

	// Find enabled GitHub repos for this app.
	repos, err := a.queries.ListEnabledGitHubReposByApp(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("search_codebase: failed to find GitHub repos: %w", err)
	}
	if len(repos) == 0 {
		return "", fmt.Errorf("search_codebase: no enabled GitHub repos found for this application. Connect a GitHub repo first.")
	}

	// Collect enabled repo names.
	repoNames := make([]string, len(repos))
	for i, r := range repos {
		repoNames[i] = r.RepoFullName
	}

	// Create connector using the first repo's connection config.
	// NOTE: All repos must belong to the same GitHub installation. If a user has
	// repos across multiple installations, only the first installation is used.
	// Multi-installation support would require grouping repos by ConnectionID.
	connector, err := codebase.New(repos[0].ConnectionConfig, a.githubClient, repoNames)
	if err != nil {
		return "", fmt.Errorf("search_codebase: %w", err)
	}

	// Connect to get installation token.
	if err := connector.Connect(ctx); err != nil {
		return "", fmt.Errorf("search_codebase: %w", err)
	}

	// Build action JSON from input params.
	actionJSON, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("search_codebase: marshal action: %w", err)
	}

	// Execute query.
	result, err := connector.Query(ctx, string(actionJSON))
	if err != nil {
		return "", fmt.Errorf("search_codebase: %w", err)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("search_codebase: marshal result: %w", err)
	}

	return string(resultJSON), nil
}
