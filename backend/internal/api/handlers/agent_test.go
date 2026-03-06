package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveAgentConfig reads the current agent_config row (if any) so it can be restored later.
// Returns nil if no config row exists.
func saveAgentConfig(t *testing.T, env *testEnv) map[string]interface{} {
	t.Helper()
	rr := env.request(t, http.MethodGet, "/api/agent/config", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))
	return cfg
}

// restoreAgentConfig writes back the original config. If the original had no "id"
// field (i.e. it was the defaults map, not a DB row), we delete the row instead.
func restoreAgentConfig(t *testing.T, env *testEnv, original map[string]interface{}) {
	t.Helper()
	if original == nil {
		return
	}

	// If the original response had no "id", it was the defaults — delete the DB row.
	if _, hasID := original["id"]; !hasID {
		_, err := env.Pool.Exec(context.Background(), "DELETE FROM agent_config WHERE id = 1")
		if err != nil {
			t.Logf("warning: failed to clean up agent_config: %v", err)
		}
		return
	}

	// Otherwise restore the saved values.
	env.request(t, http.MethodPut, "/api/agent/config", map[string]interface{}{
		"model": original["model"],
		"mode":  original["mode"],
	})
}

func TestGetAgentConfig(t *testing.T) {
	env := testSetup(t)

	original := saveAgentConfig(t, env)
	t.Cleanup(func() { restoreAgentConfig(t, env, original) })

	rr := env.request(t, http.MethodGet, "/api/agent/config", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))

	// Whether it comes from DB or defaults, model and mode must be present.
	assert.NotEmpty(t, cfg["model"])
	assert.NotEmpty(t, cfg["mode"])
}

func TestUpdateAgentConfig(t *testing.T) {
	env := testSetup(t)

	original := saveAgentConfig(t, env)
	t.Cleanup(func() { restoreAgentConfig(t, env, original) })

	body := map[string]string{
		"model": "claude-sonnet-4-5-20250929",
		"mode":  "on_demand",
	}
	rr := env.request(t, http.MethodPut, "/api/agent/config", body)
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))

	assert.Equal(t, "claude-sonnet-4-5-20250929", cfg["model"])
	assert.Equal(t, "on_demand", cfg["mode"])

	// Verify the change persists via GET.
	rr = env.request(t, http.MethodGet, "/api/agent/config", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var fetched map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&fetched))
	assert.Equal(t, "on_demand", fetched["mode"])
}
