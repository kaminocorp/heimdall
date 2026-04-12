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

func TestDeleteApplication(t *testing.T) {
	env := testSetup(t)

	// Create a second app so we're above the last-app guard.
	rr := env.request(t, http.MethodPost, "/api/apps", map[string]string{"name": "Second App"})
	require.Equal(t, http.StatusCreated, rr.Code)
	var created map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	secondID := created["id"].(string)

	// Delete the second app: should 204 and list should drop back to 1.
	rr = env.request(t, http.MethodDelete, fmt.Sprintf("/api/apps/%s", secondID), nil)
	require.Equal(t, http.StatusNoContent, rr.Code)

	rr = env.request(t, http.MethodGet, "/api/apps", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var apps []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&apps))
	assert.Len(t, apps, 1)
	assert.Equal(t, env.AppID, apps[0]["id"])
}

func TestDeleteApplication_LastAppGuard(t *testing.T) {
	env := testSetup(t)

	// testSetup creates exactly one app — deleting it should hit the guard.
	rr := env.request(t, http.MethodDelete, fmt.Sprintf("/api/apps/%s", env.AppID), nil)
	require.Equal(t, http.StatusConflict, rr.Code)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Equal(t, "last_app", body["code"])
	assert.Contains(t, body["error"], "last application")

	// The app must still exist after the failed delete.
	rr = env.request(t, http.MethodGet, "/api/apps", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var apps []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&apps))
	assert.Len(t, apps, 1)
}

func TestDeleteApplication_WrongOrg(t *testing.T) {
	env := testSetup(t)

	// Bootstrap a foreign org + app the authenticated user should not see.
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

	// Cross-org delete must 404 — we don't leak existence of resources the
	// caller can't see, matching the pattern in TestGetApplication_WrongOrg.
	rr := env.request(t, http.MethodDelete, fmt.Sprintf("/api/apps/%s", otherAppID.String()), nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// And the foreign app must still exist.
	var exists bool
	err = env.Pool.QueryRow(t.Context(), "SELECT EXISTS (SELECT 1 FROM applications WHERE id = $1)", otherAppID).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "foreign app should not be deleted by a cross-org request")
}

func TestDeleteApplication_CascadesToConnections(t *testing.T) {
	env := testSetup(t)

	// Create a second app + a connection inside it. We delete the *second*
	// app (not env.AppID) because env.AppID is the last-app-guarded default.
	rr := env.request(t, http.MethodPost, "/api/apps", map[string]string{"name": "Cascade Target"})
	require.Equal(t, http.StatusCreated, rr.Code)
	var created map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&created))
	targetID := created["id"].(string)

	rr = env.request(t, http.MethodPost, "/api/connections", map[string]string{
		"app_id": targetID,
		"name":   "cascade-conn",
		"type":   "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)
	var conn map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID := conn["id"].(string)

	// Delete the app — the connection row should vanish via FK cascade.
	rr = env.request(t, http.MethodDelete, fmt.Sprintf("/api/apps/%s", targetID), nil)
	require.Equal(t, http.StatusNoContent, rr.Code)

	// Codify the cascade assumption: if a future migration accidentally drops
	// ON DELETE CASCADE on connections.app_id, this assertion will fail and
	// we'll catch the regression in CI instead of in production.
	var exists bool
	err := env.Pool.QueryRow(t.Context(), "SELECT EXISTS (SELECT 1 FROM connections WHERE id = $1)", connID).Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists, "connection should be cascade-deleted with its parent app")
}

func TestListApplications_WithCounts(t *testing.T) {
	env := testSetup(t)

	// Seed two connections under env.AppID so the counts are non-trivial.
	env.createTestConnection(t, "counts-conn-a")
	env.createTestConnection(t, "counts-conn-b")

	rr := env.request(t, http.MethodGet, "/api/apps?include=counts", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var rows []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&rows))
	require.Len(t, rows, 1)

	row := rows[0]
	assert.Equal(t, env.AppID, row["id"])
	assert.Equal(t, float64(2), row["connection_count"])
	// Schedule count is 0 since we didn't seed any investigation_schedules.
	assert.Equal(t, float64(0), row["schedule_count"])
}

func TestListApplications_DefaultShapeUnchanged(t *testing.T) {
	env := testSetup(t)

	// Default call (no ?include=counts) must not surface the count fields —
	// this guards against accidental shape changes that would break the
	// sidebar selector and any other lean consumer.
	rr := env.request(t, http.MethodGet, "/api/apps", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var rows []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&rows))
	require.Len(t, rows, 1)

	_, hasConnCount := rows[0]["connection_count"]
	_, hasSchedCount := rows[0]["schedule_count"]
	assert.False(t, hasConnCount, "default list must not include connection_count")
	assert.False(t, hasSchedCount, "default list must not include schedule_count")
}

func TestUpdateAppAgentConfig_RejectsUnknownModel(t *testing.T) {
	env := testSetup(t)

	body := map[string]interface{}{
		"model":    "anthropic/claude-sonnet-99",
		"provider": "anthropic",
		"mode":     "off",
	}

	rr := env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Contains(t, resp["error"], "unknown model")
}

func TestUpdateAppAgentConfig_RejectsProviderModelMismatch(t *testing.T) {
	env := testSetup(t)

	// Claude Opus 4.6 (direct) is provider "anthropic" — requesting it via
	// "openrouter" should be rejected as a mismatch.
	body := map[string]interface{}{
		"model":    "claude-opus-4-6",
		"provider": "openrouter",
		"mode":     "off",
	}

	rr := env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), body)
	// This hits the "openrouter provider not enabled" check first (no key in
	// test config), which is correct — but let's test a mismatch that gets past
	// the provider gate. An anthropic-direct model sent with provider "anthropic"
	// but actually there's no mismatch. Let's test with a model that resolves
	// to openrouter but provider says anthropic — but we can't because the
	// provider gate blocks openrouter. So we test the symmetric case: an
	// openrouter-provider model with provider "anthropic". Since the model's
	// catalogue entry says provider "openrouter" but the request says "anthropic",
	// validation should catch it.
	//
	// But wait — the request provider is "anthropic" (passes provider gate),
	// and the model resolves to provider "openrouter" → mismatch → 400.
	body2 := map[string]interface{}{
		"model":    "openai/gpt-5.4",
		"provider": "anthropic",
		"mode":     "off",
	}

	rr = env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), body2)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Contains(t, resp["error"], "belongs to provider")
}

func TestUpdateAppAgentConfig_AcceptsFlagshipModel(t *testing.T) {
	env := testSetup(t)

	body := map[string]interface{}{
		"model":    "claude-opus-4-6",
		"provider": "anthropic",
		"mode":     "continuous",
	}

	rr := env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), body)
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))
	assert.Equal(t, "claude-opus-4-6", cfg["model"])
	assert.Equal(t, "anthropic", cfg["provider"])
}

func TestUpdateAppAgentConfig_AcceptsEconomyModel(t *testing.T) {
	env := testSetup(t)

	body := map[string]interface{}{
		"model":    "claude-haiku-4-5-20251001",
		"provider": "anthropic",
		"mode":     "off",
	}

	rr := env.request(t, http.MethodPut, fmt.Sprintf("/api/apps/%s/agent/config", env.AppID), body)
	require.Equal(t, http.StatusOK, rr.Code)

	var cfg map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))
	assert.Equal(t, "claude-haiku-4-5-20251001", cfg["model"])
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
