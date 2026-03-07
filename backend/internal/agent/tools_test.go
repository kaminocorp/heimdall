package agent

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// newToolTestAgent creates an Agent for testing tool dispatch logic.
// It uses a stubDBTX so DB-dependent tools will return errors, but the
// dispatch routing itself can be verified.
func newToolTestAgent() *Agent {
	queries := db.New(&stubDBTX{})
	return &Agent{
		queries:    queries,
		config:     &config.Config{AnthropicKey: "test-key"},
		classifier: &PassthroughClassifier{},
	}
}

func TestDispatch_UnknownTool(t *testing.T) {
	agent := newToolTestAgent()
	ctx := context.Background()
	userID := uuid.New()

	result, err := agent.Dispatch(ctx, userID, "nonexistent_tool", map[string]any{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool: nonexistent_tool")
	assert.Empty(t, result)
}

func TestDispatch_SearchLogs(t *testing.T) {
	agent := newToolTestAgent()
	ctx := context.Background()
	userID := uuid.New()

	// search_logs will be dispatched to toolSearchLogs, which queries
	// the DB via stubDBTX. The stub returns an error, confirming that
	// dispatch correctly routed to the search_logs implementation.
	result, err := agent.Dispatch(ctx, userID, "search_logs", map[string]any{
		"query": "error",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "search_logs:")
	assert.Empty(t, result)
}

func TestDispatch_QueryDatabase(t *testing.T) {
	agent := newToolTestAgent()
	ctx := context.Background()
	userID := uuid.New()

	connID := uuid.New()

	// query_database will be dispatched to toolQueryDatabase, which fetches
	// the connection from DB via stubDBTX. The stub returns an error,
	// confirming that dispatch correctly routed to the query_database
	// implementation.
	result, err := agent.Dispatch(ctx, userID, "query_database", map[string]any{
		"sql":           "SELECT 1",
		"connection_id": connID.String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "query_database:")
	assert.Empty(t, result)
}
