package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// TestSourceFilters_GitHubEnableFlowBackendE2E exercises the code path that
// runs when a user enables a GitHub repo via the generic /sources endpoint.
// Post-Phase-3, that's the only way — there is no /github/repos endpoint —
// and the agent's search_codebase tool reads through
// ListEnabledGitHubReposByApp which now joins app_source_filters.
//
// Flow:
//  1. Create a GitHub connection directly (bypassing the OAuth callback,
//     which is unreachable in tests).
//  2. Seed a `connection_sources` row — simulating what the discover
//     endpoint would produce after hitting the GitHub API.
//  3. Enable it via PUT /sources.
//  4. Call ListEnabledGitHubReposByApp the way the agent does.
//  5. Verify the repo surfaces with its connection config intact.
func TestSourceFilters_GitHubEnableFlowBackendE2E(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Step 1: insert a github connection. validateConnectorConfig doesn't
	// cover github (no connector constructor), so we can create one through
	// the public API with an arbitrary installation_id. We do it via raw
	// SQL to avoid depending on any frontend token-generation path.
	appIDUUID, err := uuid.Parse(env.AppID)
	require.NoError(t, err)
	userIDUUID, err := uuid.Parse(env.UserID)
	require.NoError(t, err)
	orgIDUUID, err := uuid.Parse(env.OrgID)
	require.NoError(t, err)

	connID := uuid.New()
	_, err = env.Pool.Exec(ctx,
		`INSERT INTO connections (id, user_id, org_id, app_id, name, type, direction, config, status)
		 VALUES ($1, $2, $3, $4, 'test-github', 'github', 'two_way', '{"installation_id":123456}', 'active')`,
		connID, userIDUUID, orgIDUUID, appIDUUID,
	)
	require.NoError(t, err)

	// Step 2: seed a connection_sources row for "octocat/hello-world" —
	// standing in for what discoverGitHubRepos would upsert.
	_, err = env.Pool.Exec(ctx,
		`INSERT INTO connection_sources (connection_id, source_name) VALUES ($1, $2)`,
		connID, "octocat/hello-world",
	)
	require.NoError(t, err)

	// Step 3: enable via the generic bulk-upsert endpoint.
	rr := env.request(t, http.MethodPut, "/api/connections/"+connID.String()+"/sources", map[string]any{
		"sources": []map[string]any{
			{"source_name": "octocat/hello-world", "enabled": true},
		},
	})
	require.Equal(t, http.StatusNoContent, rr.Code, rr.Body.String())

	// Step 4: read through the same query the agent's search_codebase uses.
	rows, err := env.Queries.ListEnabledGitHubReposByApp(ctx, appIDUUID)
	require.NoError(t, err)
	require.Len(t, rows, 1, "enabled repo surfaces to the agent tool")

	assert.Equal(t, connID, rows[0].ConnectionID)
	assert.Equal(t, "octocat/hello-world", rows[0].RepoFullName)

	// Step 5: the connection config comes through unchanged — the agent
	// reads installation_id from here to obtain an installation token.
	var cfg map[string]any
	require.NoError(t, json.Unmarshal(rows[0].ConnectionConfig, &cfg))
	assert.Equal(t, float64(123456), cfg["installation_id"])
}

// TestSourceFilters_GitHubDisabledNotReturned confirms the query filters on
// enabled = true. Same setup as above but with enabled=false — the row
// should be absent from the agent's view.
func TestSourceFilters_GitHubDisabledNotReturned(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	appIDUUID, _ := uuid.Parse(env.AppID)
	userIDUUID, _ := uuid.Parse(env.UserID)
	orgIDUUID, _ := uuid.Parse(env.OrgID)

	connID := uuid.New()
	_, err := env.Pool.Exec(ctx,
		`INSERT INTO connections (id, user_id, org_id, app_id, name, type, direction, config, status)
		 VALUES ($1, $2, $3, $4, 'test-github', 'github', 'two_way', '{"installation_id":99}', 'active')`,
		connID, userIDUUID, orgIDUUID, appIDUUID,
	)
	require.NoError(t, err)

	// Insert directly into app_source_filters with enabled=false — simulates
	// a repo the user has explicitly opted out of.
	_, err = env.Pool.Exec(ctx,
		`INSERT INTO app_source_filters (app_id, connection_id, source_name, enabled)
		 VALUES ($1, $2, 'disabled/repo', false)`,
		appIDUUID, connID,
	)
	require.NoError(t, err)

	rows, err := env.Queries.ListEnabledGitHubReposByApp(ctx, appIDUUID)
	require.NoError(t, err)
	assert.Empty(t, rows, "disabled repos must not surface to the agent")
}

// TestDiscoverSources_RejectsNonGitHub confirms the discover endpoint is
// only meaningful for connector types with an explicit upstream list API.
// Webhook connections get 400 with a message pointing to the passive flow.
func TestDiscoverSources_RejectsNonGitHub(t *testing.T) {
	env := testSetup(t)

	connID := env.createTestConnection(t, "plain-webhook")

	rr := env.request(t, http.MethodPost, "/api/connections/"+connID+"/sources/discover", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	// Error mentions the connection type to help debugging.
	assert.Contains(t, body["error"], "webhook_logs")
}

// TestDiscoverSources_GitHubAppNotConfigured checks that a GitHub connection
// without a configured GitHub App returns a clear 502 — the endpoint
// doesn't pretend the sync succeeded when it can't reach GitHub.
func TestDiscoverSources_GitHubAppNotConfigured(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	appIDUUID, _ := uuid.Parse(env.AppID)
	userIDUUID, _ := uuid.Parse(env.UserID)
	orgIDUUID, _ := uuid.Parse(env.OrgID)

	connID := uuid.New()
	_, err := env.Pool.Exec(ctx,
		`INSERT INTO connections (id, user_id, org_id, app_id, name, type, direction, config, status)
		 VALUES ($1, $2, $3, $4, 'gh', 'github', 'two_way', '{"installation_id":123456}', 'active')`,
		connID, userIDUUID, orgIDUUID, appIDUUID,
	)
	require.NoError(t, err)

	// Test harness leaves s.GitHub nil (see testhelpers_test.go) — the
	// discover helper bails out early with a "not configured" message
	// rather than crashing. 503 Service Unavailable is correct here: the
	// installation is fine, our deployment is misconfigured. Post-Phase-5
	// this is distinguished from the 502 Bad Gateway case (upstream failure).
	rr := env.request(t, http.MethodPost,
		fmt.Sprintf("/api/connections/%s/sources/discover", connID.String()), nil)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Contains(t, body["error"], "not configured")
}

// TestSourceFilters_RoutesDroppedForGitHubRepos is a sanity check that the
// old /github/repos endpoints are gone — a GET returns 404 after Phase 3.
func TestSourceFilters_RoutesDroppedForGitHubRepos(t *testing.T) {
	env := testSetup(t)
	connID := env.createTestConnection(t, "webhook-one")

	// Even against a valid connection id, the old path is no longer
	// registered — chi returns 404 for the route, not the handler.
	rr := env.request(t, http.MethodGet, "/api/connections/"+connID+"/github/repos", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestListEnabledGitHubReposByApp_QuerySignature confirms that the sqlc row
// type the agent tool depends on still exposes the two fields the agent
// reads (RepoFullName, ConnectionConfig). A regression here would silently
// break search_codebase at runtime.
func TestListEnabledGitHubReposByApp_QuerySignature(t *testing.T) {
	// Purely a type-level sanity check — compile-time guarantee.
	var row db.ListEnabledGitHubReposByAppRow
	_ = row.RepoFullName
	_ = row.ConnectionConfig
	_ = row.ConnectionID
}
