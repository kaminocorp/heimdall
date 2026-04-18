package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateConnection_OrgScoped verifies that POST /api/connections with no
// app_id creates an org-scoped connection (app_id nil on the returned row).
// Only multi-source types are allowed — other types must still pass app_id.
func TestCreateConnection_OrgScoped(t *testing.T) {
	env := testSetup(t)

	// webhook_logs supports org scoping.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "org-wide-drain",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	assert.Nil(t, conn["app_id"], "org-scoped connection has nil app_id")
	assert.NotEmpty(t, conn["org_id"], "org-scoped connection has org_id")
	assert.Equal(t, env.OrgID, conn["org_id"])

	// postgres is app-scoped only — must fail without app_id.
	rr = env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "must-fail",
		"type": "postgres",
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestListConnectionsByApp_IncludesOrgScoped verifies that the per-app
// connection list returns both app-scoped AND org-scoped connections from the
// same org, so every app in the org sees the shared drain.
func TestListConnectionsByApp_IncludesOrgScoped(t *testing.T) {
	env := testSetup(t)

	// App-scoped connection.
	_ = env.createTestConnection(t, "app-scoped")

	// Org-scoped connection.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "org-scoped",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	// GET /api/apps/{appId}/connections should return both.
	rr = env.request(t, http.MethodGet, fmt.Sprintf("/api/apps/%s/connections", env.AppID), nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var list []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&list))
	require.Len(t, list, 2)

	names := []string{list[0]["name"].(string), list[1]["name"].(string)}
	assert.ElementsMatch(t, []string{"app-scoped", "org-scoped"}, names)

	// The org-scoped row has nil app_id; the app-scoped row carries the env.AppID.
	for _, row := range list {
		if row["name"] == "org-scoped" {
			assert.Nil(t, row["app_id"])
		} else {
			assert.Equal(t, env.AppID, row["app_id"])
		}
	}
}

// TestIngestWebhookLogs_OrgFanOut verifies that ingestion into an org-scoped
// connection fans out one entry to every app whose filter has enabled=true
// for that source. Two apps, two different enabled sources, one batch with
// both sources → each app receives exactly one row.
func TestIngestWebhookLogs_OrgFanOut(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Create a second app in the same org.
	secondAppID := uuid.New()
	_, err := env.Pool.Exec(ctx,
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		secondAppID, env.OrgID, "Test App 2", "active",
	)
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx,
		"INSERT INTO app_agent_config (app_id, model, mode, schedule_interval_secs) VALUES ($1, $2, $3, $4)",
		secondAppID, "claude-sonnet-4-6", "off", 60,
	)
	require.NoError(t, err)

	// Create an org-scoped webhook_logs connection.
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "org-drain",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, err := uuid.Parse(conn["id"].(string))
	require.NoError(t, err)
	token := conn["config"].(map[string]any)["webhook_token"].(string)

	_, err = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)
	require.NoError(t, err)

	// App 1 enables "trajan", App 2 enables "hadrian". Both POSTs go through
	// the source-filter endpoint with ?app_id=... since the connection is
	// org-scoped.
	rr = env.request(t, http.MethodPost,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID.String(), env.AppID),
		map[string]string{"source_name": "trajan"})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	rr = env.request(t, http.MethodPost,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID.String(), secondAppID.String()),
		map[string]string{"source_name": "hadrian"})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	// Batch with one entry per app's enabled source.
	body := []byte(`[
		{"message":"from trajan","host":"a","source_type":"fly_io","fly":{"app":{"name":"trajan"},"machine":{"id":"m1"},"region":"lhr"},"log":{"level":"info"}},
		{"message":"from hadrian","host":"b","source_type":"fly_io","fly":{"app":{"name":"hadrian"},"machine":{"id":"m2"},"region":"lhr"},"log":{"level":"warn"}}
	]`)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var result map[string]any
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	// Both entries have a matching app — no drops. `filtered` is omitted
	// from the response when zero (omitempty), so assert absence rather
	// than explicit 0.
	assert.Equal(t, float64(2), result["accepted"])
	assert.NotContains(t, result, "filtered")

	// Verify each app received exactly one log_buffer row with the correct app_id.
	var count1, count2 int
	err = env.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM log_buffer WHERE app_id = $1 AND connection_id = $2",
		env.AppID, connID).Scan(&count1)
	require.NoError(t, err)
	assert.Equal(t, 1, count1, "app 1 received exactly one row")

	err = env.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM log_buffer WHERE app_id = $1 AND connection_id = $2",
		secondAppID, connID).Scan(&count2)
	require.NoError(t, err)
	assert.Equal(t, 1, count2, "app 2 received exactly one row")
}

// TestIngestWebhookLogs_OrgFanOutSharedSource verifies that when the SAME
// source name is enabled in multiple apps, ingestion duplicates the entry
// into every matching app (the deliberate write-time fan-out).
func TestIngestWebhookLogs_OrgFanOutSharedSource(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	secondAppID := uuid.New()
	_, err := env.Pool.Exec(ctx,
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		secondAppID, env.OrgID, "Test App 2", "active",
	)
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx,
		"INSERT INTO app_agent_config (app_id, model, mode, schedule_interval_secs) VALUES ($1, $2, $3, $4)",
		secondAppID, "claude-sonnet-4-6", "off", 60,
	)
	require.NoError(t, err)

	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "shared-drain",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)
	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID, _ := uuid.Parse(conn["id"].(string))
	token := conn["config"].(map[string]any)["webhook_token"].(string)
	_, _ = env.Pool.Exec(ctx,
		"UPDATE connections SET status = 'active' WHERE id = $1", connID)

	// Both apps enable "shared-app".
	for _, appID := range []string{env.AppID, secondAppID.String()} {
		rr := env.request(t, http.MethodPost,
			fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID.String(), appID),
			map[string]string{"source_name": "shared-app"})
		require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	}

	body := []byte(`{"message":"shared log","host":"h","source_type":"fly_io","fly":{"app":{"name":"shared-app"},"machine":{"id":"m"},"region":"lhr"},"log":{"level":"info"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	env.Router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// One row per matching app — 1 source × 2 apps = 2 inserts.
	var total int
	err = env.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM log_buffer WHERE connection_id = $1", connID).Scan(&total)
	require.NoError(t, err)
	assert.Equal(t, 2, total, "fan-out inserts one row per enabled app")
}

// TestResolveSourceFilterApp_CrossOrgRejected is the headline Phase 2
// security test: a user who belongs to two orgs cannot use an app from org B
// to edit a filter on a connection in org A, even though GetApplicationByOrgUser
// and GetConnectionByUser both return rows (the user is a member of both).
// The `app.OrgID != conn.OrgID` check in resolveSourceFilterApp is the boundary
// that prevents this.
func TestResolveSourceFilterApp_CrossOrgRejected(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// Build a second org that the test user is also a member of, with its
	// own app. env.OrgID is "org A" (membership created by testSetup);
	// this is "org B".
	orgBID := uuid.New()
	orgBSlug := fmt.Sprintf("test-org-b-%s", orgBID.String()[:8])
	_, err := env.Pool.Exec(ctx,
		"INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)",
		orgBID, "Test Org B", orgBSlug)
	require.NoError(t, err)
	t.Cleanup(func() {
		env.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", orgBID)
	})

	userUUID, err := uuid.Parse(env.UserID)
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx,
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'owner')",
		userUUID, orgBID)
	require.NoError(t, err)

	orgBAppID := uuid.New()
	_, err = env.Pool.Exec(ctx,
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		orgBAppID, orgBID, "Org B App", "active")
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx,
		"INSERT INTO app_agent_config (app_id, model, mode, schedule_interval_secs) VALUES ($1, $2, $3, $4)",
		orgBAppID, "claude-sonnet-4-6", "off", 60)
	require.NoError(t, err)

	// Create an org-scoped connection in org A (env.OrgID).
	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "org-a-drain",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID := conn["id"].(string)
	require.Equal(t, env.OrgID, conn["org_id"])

	// Baseline: same org works.
	rr = env.request(t, http.MethodGet,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID, env.AppID), nil)
	assert.Equal(t, http.StatusOK, rr.Code, "own-org request should succeed")

	// Attack: user is a member of both orgs, tries to pass orgB's app against
	// the orgA connection. Must be rejected with 403 — 400 or 200 would be
	// catastrophic.
	rr = env.request(t, http.MethodGet,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID, orgBAppID.String()), nil)
	assert.Equal(t, http.StatusForbidden, rr.Code,
		"cross-org app must be forbidden (got %d: %s)", rr.Code, rr.Body.String())

	// Same check on the write paths — PUT, POST, DELETE must all reject
	// cross-org access, not just GET.
	rr = env.request(t, http.MethodPost,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID, orgBAppID.String()),
		map[string]string{"source_name": "malicious"})
	assert.Equal(t, http.StatusForbidden, rr.Code,
		"POST cross-org must be forbidden")

	rr = env.request(t, http.MethodPut,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID, orgBAppID.String()),
		map[string]any{"sources": []map[string]any{{"source_name": "x", "enabled": true}}})
	assert.Equal(t, http.StatusForbidden, rr.Code,
		"PUT cross-org must be forbidden")

	rr = env.request(t, http.MethodDelete,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s&name=x", connID, orgBAppID.String()), nil)
	assert.Equal(t, http.StatusForbidden, rr.Code,
		"DELETE cross-org must be forbidden")

	// Confirm no filter rows were created in orgB's app as a side effect —
	// belt-and-braces against a handler that 403s but still writes.
	var leaked int
	err = env.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM app_source_filters WHERE app_id = $1",
		orgBAppID).Scan(&leaked)
	require.NoError(t, err)
	assert.Equal(t, 0, leaked, "no filter rows should leak into orgB's app")
}

// TestResolveSourceFilterApp_AppScopedMismatchRejected verifies that on an
// app-scoped connection, passing a different (but otherwise-valid) app_id
// returns 403 rather than silently operating on the connection's own app.
func TestResolveSourceFilterApp_AppScopedMismatchRejected(t *testing.T) {
	env := testSetup(t)
	ctx := context.Background()

	// App-scoped connection in env.AppID.
	connID := env.createTestConnection(t, "app-scoped")

	// Second app in the same org. The user can see it — they're an owner —
	// but its id must not override the connection's own app_id.
	secondAppID := uuid.New()
	_, err := env.Pool.Exec(ctx,
		"INSERT INTO applications (id, org_id, name, status) VALUES ($1, $2, $3, $4)",
		secondAppID, env.OrgID, "Second App", "active")
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx,
		"INSERT INTO app_agent_config (app_id, model, mode, schedule_interval_secs) VALUES ($1, $2, $3, $4)",
		secondAppID, "claude-sonnet-4-6", "off", 60)
	require.NoError(t, err)

	// Omitting ?app_id= works (connection's own app_id is authoritative).
	rr := env.request(t, http.MethodGet,
		fmt.Sprintf("/api/connections/%s/sources", connID), nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Passing a matching ?app_id= also works.
	rr = env.request(t, http.MethodGet,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID, env.AppID), nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Passing a mismatched ?app_id= is forbidden — not 400, since the app_id
	// is syntactically valid and the user can see both apps.
	rr = env.request(t, http.MethodGet,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID, secondAppID.String()), nil)
	assert.Equal(t, http.StatusForbidden, rr.Code, rr.Body.String())
}

// TestSourceFilters_OrgScopedRequireAppID verifies that org-scoped
// connections require ?app_id= on source-filter endpoints.
func TestSourceFilters_OrgScopedRequireAppID(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodPost, "/api/connections", map[string]any{
		"name": "org-drain",
		"type": "webhook_logs",
	})
	require.Equal(t, http.StatusCreated, rr.Code)
	var conn map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conn))
	connID := conn["id"].(string)

	// GET without ?app_id= → 400 with descriptive error.
	rr = env.request(t, http.MethodGet, "/api/connections/"+connID+"/sources", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// POST without ?app_id= → 400.
	rr = env.request(t, http.MethodPost, "/api/connections/"+connID+"/sources",
		map[string]string{"source_name": "x"})
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// With ?app_id= → works.
	rr = env.request(t, http.MethodGet,
		fmt.Sprintf("/api/connections/%s/sources?app_id=%s", connID, env.AppID), nil)
	assert.Equal(t, http.StatusOK, rr.Code)
}
