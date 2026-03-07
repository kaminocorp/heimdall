package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListApplications(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, "/api/apps", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var apps []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&apps))

	// testSetup creates one default app.
	assert.Len(t, apps, 1)
	assert.Equal(t, "Test App", apps[0]["name"])
	assert.Equal(t, env.AppID, apps[0]["id"])
}

func TestCreateApplication(t *testing.T) {
	env := testSetup(t)

	body := map[string]string{"name": "Second App"}
	rr := env.request(t, http.MethodPost, "/api/apps", body)
	require.Equal(t, http.StatusCreated, rr.Code)

	var app map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&app))
	assert.Equal(t, "Second App", app["name"])
	assert.Equal(t, "active", app["status"])
	assert.Equal(t, env.OrgID, app["org_id"])

	// Should now have 2 apps.
	rr = env.request(t, http.MethodGet, "/api/apps", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var apps []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&apps))
	assert.Len(t, apps, 2)
}

func TestCreateApplication_MissingName(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodPost, "/api/apps", map[string]string{"name": ""})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetApplication(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var app map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&app))
	assert.Equal(t, env.AppID, app["id"])
	assert.Equal(t, "Test App", app["name"])
}

func TestGetApplication_NotFound(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, "/api/apps/00000000-0000-0000-0000-000000000000", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetApplication_WrongOrg(t *testing.T) {
	env := testSetup(t)

	// Create a second org + app that does NOT belong to env's user.
	otherOrgID := uuid.New()
	otherOrgSlug := fmt.Sprintf("other-org-%s", otherOrgID.String()[:8])
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)",
		otherOrgID, "Other Org", otherOrgSlug,
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		env.Pool.Exec(t.Context(), "DELETE FROM organizations WHERE id = $1", otherOrgID)
	})

	otherAppID := uuid.New()
	_, err = env.Pool.Exec(t.Context(),
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		otherAppID, otherOrgID, "Other App", "active",
	)
	require.NoError(t, err)

	// Authenticated user should get 404 when trying to access other org's app.
	rr := env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s", otherAppID.String()), nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Same for agent config.
	rr = env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/agent/config", otherAppID.String()), nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Same for monitoring status.
	rr = env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/monitoring/status", otherAppID.String()), nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetAppAgentConfig(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))
	assert.Equal(t, env.AppID, cfg["app_id"])
	assert.Equal(t, "claude-sonnet-4-6", cfg["model"])
	assert.Equal(t, "off", cfg["mode"])
	assert.Equal(t, float64(60), cfg["schedule_interval_secs"])
}

func TestUpdateAppAgentConfig(t *testing.T) {
	env := testSetup(t)

	body := map[string]interface{}{
		"model":                  "claude-opus-4-6",
		"mode":                   "continuous",
		"schedule_interval_secs": 120,
		"system_prompt_override": "Custom monitoring prompt",
	}

	rr := env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), body)
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))
	assert.Equal(t, "claude-opus-4-6", cfg["model"])
	assert.Equal(t, "continuous", cfg["mode"])
	assert.Equal(t, float64(120), cfg["schedule_interval_secs"])
	assert.Equal(t, "Custom monitoring prompt", cfg["system_prompt_override"])

	// Verify persistence via GET.
	rr = env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var fetched map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&fetched))
	assert.Equal(t, "continuous", fetched["mode"])
	assert.Equal(t, "Custom monitoring prompt", fetched["system_prompt_override"])
}

func TestUpdateAppAgentConfig_Defaults(t *testing.T) {
	env := testSetup(t)

	// Empty body should use defaults.
	rr := env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), map[string]interface{}{})
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))
	assert.Equal(t, "claude-sonnet-4-6", cfg["model"])
	assert.Equal(t, "off", cfg["mode"])
	assert.Equal(t, float64(60), cfg["schedule_interval_secs"])
}

func TestUpdateAppAgentConfig_InvalidMode(t *testing.T) {
	env := testSetup(t)

	body := map[string]interface{}{
		"model": "claude-sonnet-4-6",
		"mode":  "banana",
	}

	rr := env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetMonitoringStatus(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/monitoring/status", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var status map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&status))
	assert.Equal(t, "off", status["mode"])
	assert.Equal(t, float64(60), status["schedule_interval_secs"])
	// No monitoring state yet, so last_monitored_at should be null.
	assert.Nil(t, status["last_monitored_at"])
}

func TestGetAppDashboardStats(t *testing.T) {
	env := testSetup(t)

	// Create a connection in the app.
	env.createTestConnection(t, "stats-conn")

	rr := env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/stats", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var stats map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&stats))

	_, hasLogCount := stats["log_count_24h"]
	assert.True(t, hasLogCount)

	connCount, hasConnCount := stats["connection_count"]
	assert.True(t, hasConnCount)
	assert.GreaterOrEqual(t, connCount, float64(1))

	_, hasActive := stats["active_connections"]
	assert.True(t, hasActive)
}

func TestListConnectionsByApp(t *testing.T) {
	env := testSetup(t)

	// Create two connections in the app.
	env.createTestConnection(t, "app-conn-a")
	env.createTestConnection(t, "app-conn-b")

	rr := env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/connections", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var conns []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conns))
	assert.Len(t, conns, 2)

	for _, c := range conns {
		assert.Equal(t, env.AppID, c["app_id"])
	}
}
