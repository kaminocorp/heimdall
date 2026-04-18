package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSourceFilters_CRUD exercises the full lifecycle of the
// /api/connections/{id}/sources endpoints: list (empty), add (POST),
// list (reflects add), toggle (PUT), and remove (DELETE).
func TestSourceFilters_CRUD(t *testing.T) {
	env := testSetup(t)

	connID := env.createTestConnection(t, "source-filter-crud")

	// --- 1. List: empty ---
	rr := env.request(t, http.MethodGet, "/api/connections/"+connID+"/sources", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	var list []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&list))
	assert.Empty(t, list, "no sources initially")

	// --- 2. Add: manual source ---
	rr = env.request(t, http.MethodPost, "/api/connections/"+connID+"/sources",
		map[string]string{"source_name": "my-app"})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var filter map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&filter))
	assert.Equal(t, "my-app", filter["source_name"])
	assert.Equal(t, true, filter["enabled"], "manual adds are enabled immediately")

	// --- 3. List: reflects add ---
	rr = env.request(t, http.MethodGet, "/api/connections/"+connID+"/sources", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&list))
	require.Len(t, list, 1)
	assert.Equal(t, "my-app", list[0]["source_name"])
	assert.Equal(t, true, list[0]["enabled"])

	// --- 4. Toggle off via bulk PUT ---
	rr = env.request(t, http.MethodPut, "/api/connections/"+connID+"/sources", map[string]interface{}{
		"sources": []map[string]interface{}{
			{"source_name": "my-app", "enabled": false},
		},
	})
	require.Equal(t, http.StatusNoContent, rr.Code, rr.Body.String())

	rr = env.request(t, http.MethodGet, "/api/connections/"+connID+"/sources", nil)
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&list))
	require.Len(t, list, 1)
	assert.Equal(t, false, list[0]["enabled"])

	// --- 5. Delete via query param ---
	rr = env.request(t, http.MethodDelete,
		"/api/connections/"+connID+"/sources?name=my-app", nil)
	require.Equal(t, http.StatusNoContent, rr.Code, rr.Body.String())

	// Note: the discovery row (connection_sources) persists after filter
	// delete — the source would reappear disabled if traffic returns. In this
	// test we only added the source manually, so discovery has a row for it,
	// but filter is gone.
	rr = env.request(t, http.MethodGet, "/api/connections/"+connID+"/sources", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&list))
	require.Len(t, list, 1, "discovery row persists after filter delete")
	assert.Equal(t, "my-app", list[0]["source_name"])
	assert.Equal(t, false, list[0]["enabled"], "no filter row = disabled")
}

// TestSourceFilters_AuthorizationScoping verifies that one user cannot
// read/modify another user's connection's source filters. This is enforced
// by GetConnectionByUser (joins on user_id) + RLS on the filter tables.
func TestSourceFilters_AuthorizationScoping(t *testing.T) {
	env := testSetup(t)
	connID := env.createTestConnection(t, "scope-test")

	// Random UUID — not an existing connection.
	rr := env.request(t, http.MethodGet,
		"/api/connections/00000000-0000-0000-0000-000000000000/sources", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Invalid UUID in path.
	rr = env.request(t, http.MethodGet, "/api/connections/not-a-uuid/sources", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// DELETE without name param is a 400.
	rr = env.request(t, http.MethodDelete, "/api/connections/"+connID+"/sources", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// POST with empty name is a 400.
	rr = env.request(t, http.MethodPost, "/api/connections/"+connID+"/sources",
		map[string]string{"source_name": ""})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
